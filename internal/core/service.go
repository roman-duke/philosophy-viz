package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
)

var BASE_URL = "https://en.wikipedia.org/w/api.php"

// ======================================== errors =========================================//
var ErrDuplicateSearch = errors.New("error finding link duplicate")

//==========================================================================================//

// =================================== PLAN ========================================//
// 1. Communication: each goroutine should communicate its visited links via a channel
// only the main goroutine is responsible for writing to the map. incremental updates
// should be sent over time ✅
// 2. Shared cache - there should be a cache for fetched content and a separate one for
// visited links ✅
// 3. All goroutines have read access to the "shared memory"/"cache". This is how they
// would check for already visited/traversed paths (not sure I would be able to do this)

// ============================ quick "fake" type union =====================================//
type TraverseMode string

const (
	Relaxed TraverseMode = "relaxed"
	Strict  TraverseMode = "strict"
)

//===========================================================================================//

type Service struct {
	client *http.Client
	store  *Store
}

func NewService(st *Store) *Service {
	// create the client.
	c := createClient()

	return &Service{client: c, store: st}
}

// TOOD: thread the context to the fetch function to enable us cancel any in-flight request made to
// wikipedia
func (s *Service) VisitPaths(ctx context.Context, titles []string, mode TraverseMode) <-chan StreamMessage {
	// we basically construct the map of the initial paths based off the supplied articles
	pMap := make(PathsMap)
	for _, val := range titles {
		pMap[val] = []Link{{title: val, linkAddress: fmt.Sprintf("/wiki/%s", val)}}
	}

	out := make(chan StreamMessage)

	go func() {
		defer close(out)

		// now create a channel that we can use to collect the results of the goroutines
		// We shall use the CSP paradigm here to allow the separate goroutines communicate
		// their result over time
		c := make(chan Result)

		// also create a wait group to be used for handling lifecyle/completion
		var wg sync.WaitGroup

		// spin up the necessary goroutines
		for k, v := range pMap {
			wg.Add(1)

			var PARAMS = url.Values{
				"action":        {"parse"},
				"format":        {"json"},
				"formatversion": {"2"},
			}

			go s.getToPhilosophy(k, mode, v, PARAMS, c, &wg, ctx)
		}

		// anonymous goroutine that is resposible for closing the channel after all goroutines
		// have run till completion
		go func() {
			wg.Wait()
			close(c)
		}()

		for msg := range c {
			switch result := msg.(type) {
			case NewLink:
				// append to the already existing key
				pMap[result.key] = append(pMap[result.key], result.val)

			case RedirectLink:
				links, ok := pMap[result.key]
				if !ok {
					// TODO: unhandled error
				}

				links[len(links)-1].redirect = true
				links[len(links)-1].redirectAddress = result.address

			case LoopLink:
				links, ok := pMap[result.key]
				if !ok {
					// TODO: unhandled error
				}

				lastLink := links[len(links)-1]
				lastLink.loopBound = true

				links = links[:len(links)-1]
				links[len(links)-1].loopLinks = append(links[len(links)-1].loopLinks, lastLink)

			case ParsingErr:
				log.Printf("\033[31m parsing error occurred in the %s goroutine: %v\033[0m\n", result.key, result.err)

			case FetchErr:
				log.Printf("\033[31m fetch error occurred in the %s goroutine: %v\033[0m\n", result.key, result.err)

			default:
				panic(fmt.Sprintf("unhandled Result variant: %T", result))
			}

			select {
			case out <- streamMessageMapper(msg):
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

// method to fuzzy find Wikipedia titles
func (s *Service) FuzzyFindTitles(c context.Context, keyword string) (WikiQueryRes, error) {
	var PARAMS = url.Values{
		"action": {"query"},
		"list":   {"search"},
		// we specify the edit distance to be of length 2
		"srsearch": {fmt.Sprintf("intitle:%s~2", keyword)},
		"utf8":     {"1"},
		"format":   {"json"},
	}

	res, err := fetch[WikiQueryRes](c, s.client, BASE_URL, PARAMS)

	if err != nil {
		return WikiQueryRes{}, err
	}

	return *res, nil
}

// helper function to map the result to a client-friendly message
type EventType string

const (
	Node     EventType = "node"
	Redirect EventType = "redirect"
	Loop     EventType = "loop"
	Err      EventType = "error"
)

type StreamMessage struct {
	EventType EventType
	Data      []byte
}

func streamMessageMapper(v Result) StreamMessage {
	switch x := v.(type) {
	case NewLink:
		jsonData, _ := json.Marshal(nodeDTO{Key: x.key, Value: linkDTO{Title: x.val.title, LinkAddress: x.val.linkAddress, Redirect: x.val.redirect, RedirectAddress: x.val.redirectAddress, LoopBound: x.val.loopBound}})
		return StreamMessage{EventType: Node, Data: jsonData}

	case RedirectLink:
		jsonData, _ := json.Marshal(redirectDTO{Key: x.key, Address: x.address})
		return StreamMessage{EventType: Redirect, Data: jsonData}

	case LoopLink:
		jsonData, _ := json.Marshal(cycleDTO{Key: x.key})
		return StreamMessage{EventType: Loop, Data: jsonData}

	case ParsingErr:
		jsonData, _ := json.Marshal(errDTO{Key: x.key, Message: x.err.Error()})
		return StreamMessage{EventType: Err, Data: jsonData}

	case FetchErr:
		jsonData, _ := json.Marshal(errDTO{Key: x.key, Message: x.err.Error()})
		return StreamMessage{EventType: Err, Data: jsonData}

	default:
		panic(fmt.Sprintf("unhandled Result variant: %T", v))
	}
}

// a discriminated union is needed for this because there are different states that
// our program could be in:
// 1. the first case is "new" - this tells us that a new link has been found and should
// be appended to the already discovered list of links
// 2. the second case is "redirect" - this tells us that the previously existing link
// was a redirect and hence should be updated as such
// 3. the third case is "loop" - this tells us that the current link causes a loop
// and hence should be popped from the list of links and put as a child of the new last
// element in the list of links
type Result interface {
	isResult() // "sealed" marker
}

type NewLink struct {
	key string
	val Link
}

type RedirectLink struct {
	key string
	// val Link
	address string
}

type LoopLink struct {
	key string
}

type FetchErr struct {
	key string
	err error
}

type ParsingErr struct {
	key string
	err error
}

func (NewLink) isResult()      {}
func (LoopLink) isResult()     {}
func (RedirectLink) isResult() {}
func (ParsingErr) isResult()   {}
func (FetchErr) isResult()     {}

// ============================== RELEVANT DTO's =====================================//
type linkDTO struct {
	Title           string    `json:"title"`
	LinkAddress     string    `json:"linkAddress"`
	Redirect        bool      `json:"redirect"`
	RedirectAddress string    `json:"redirectAddress"`
	LoopBound       bool      `json:"loopBound"`
}

type nodeDTO struct {
	Key   string  `json:"key"`
	Value linkDTO `json:"link"`
}

type cycleDTO struct {
	Key string `json:"key"`
}

type redirectDTO struct {
	Key     string `json:"key"`
	Address string `json:"address"`
}

type errDTO struct {
	Key     string `json:"key"`
	Message string `json:"message"`
}
