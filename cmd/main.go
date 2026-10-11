package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var BASE_URL = "https://en.wikipedia.org/w/api.php"
var PARAMS = url.Values{
	"action": {"parse"},
	"format": {"json"},
	// "page":          {"Albert Einstein"},
	"formatversion": {"2"},
}

type LangLink struct {
	Lang     string `json:"lang"`
	Url      string `json:"url"`
	Langname string `json:"langname"`
	Autonym  string `json:"autonym"`
	Title    string `json:"title"`
}

// define the shape of the Wikipedia response
type WikiPageRes struct {
	Parse struct {
		Title     string     `json:"title"`
		Text      string     `json:"text"`
		LangLinks []LangLink `json:"langlinks"`
	}
}

var client *http.Client

// let's make a simple fetch request to a wikipedia article to get the content and then
// log it to console
func main() {
	// create the necessarsy connection pool
	client = createClient()

	visitedTitles := make([]Link, 0)

	paths, err := getToPhilosophy("Albert Einstein", "relaxed", visitedTitles, PARAMS)

	if err != nil {
		log.Fatalf("\033[31merror fetching content: %v\033[0m", err)
	}

	fmt.Printf("The path to philosophy --> \033[36m%v\033[0m", paths)
}

type Link struct {
	title       string
	redirect    bool
	linkAddress string
	// the loopBound shows that this link causes a loop
	// the nextLinkAddress just points at the next thing in the list
	loopBound bool
	loopLinks []Link
	// this is where a sum type would be important.
	// This should only be present when the value of
	// the redirect member is true, but sadly Go does
	// not have the sum-type language feature.
	redirectAddress string
}

// TODO: If the redirect is something that has been visited before, then go back to the previous
// page and search for the next link that has not previously been visited.

// Unexpected complexity arisen (redirect links being loop links)

// function to recursively get to philosophy
// this mode argument should ideally be a sum type
func getToPhilosophy(node string, mode string, links []Link, params url.Values) ([]Link, error) {

	// declare relevant variables to help us with the tail recursion
	var (
		ts           []Link
		err          error
	)

	// our terminator case
	if node == "Philosophy" {
		return links, nil
	}

	fmt.Printf("%s%sour node for this iteration: (%s)!%s\n", Bold, Red, node, Reset)

	// update the map of url.Values with the new page key
	params.Set("page", node)

	// ============================ SLEEP ===========================//
	// sleep a bit, so as not to constantly bombard Wiki's server
	time.Sleep(time.Millisecond * 20)
	// ==============================================================//
	res, err := fetch(client, BASE_URL, params)

	if err != nil {
		return links, err
	}

	// call the parse function here to help us extract the relevant link
	parseRes, err := parser(res, mode, links)

	// declare our new link variable
	var newLink Link

	if errors.Is(err, ErrLoopLink) {
		newLink.loopBound = true
	} else if err != nil {
		return links, err
	}

	nextNode := strings.TrimPrefix(parseRes.linkAddr, "/wiki/")

	fmt.Println("Here is our next node:", nextNode)
	fmt.Println("length of our new node", len(nextNode))

	// handle this error here just for consistency purposes, but it is currently impossible
	// that there would be any error because of what the parser does for us under the hood
	// the error is actualy more so a "logical" error
	if len(nextNode) == len(parseRes.linkAddr) {
		return links, errors.New("invalid link returned by parser")
	}

	// update the values of the newLink
	newLink.title = nextNode
	newLink.linkAddress = parseRes.linkAddr

	// in redirect mode, just update the link of the previous element
	if parseRes.redirect {
		links[len(links)-1].redirect = true
		links[len(links)-1].redirectAddress = parseRes.linkAddr
	}

	// if the new link causes a loop, then we pop it from our array and add
	// it as a "child" to the previous element in the array
	if newLink.loopBound {
		lastElem := links[len(links)-1]
		links = links[:len(links)-1]
		links[len(links)-1].loopLinks = append(links[len(links)-1].loopLinks, lastElem)

		// the call would be next node of the last "valid" link
		n := links[len(links)-1].title
		ts, err = getToPhilosophy(n, mode, links, params)
	} else {
		if !parseRes.redirect {
			links = append(links, newLink)
		}

		ts, err = getToPhilosophy(nextNode, mode, links, params)
	}

	return ts, err
}

func createClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

// given that this does something that is stateful, we have to pass a context to it
// so that we can cancel on cleanup or something. I think that is the idiomatic way of
// doing it in Go??
func fetch(client *http.Client, urlString string, params url.Values) (*WikiPageRes, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, urlString, nil)

	// use the default methods to add the necessary query params
	req.URL.RawQuery = url.Values.Encode(params)

	log.Printf("The url to fetch from: %s", req.URL.String())

	if err != nil {
		return nil, err
	}

	// set the necessary User Agent header so that the request isn't blocked by Wikipedia
	req.Header.Set("User-Agent", "Path to Philosophy Crawler (https://philosophy-viz)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %v", resp.Status)
	}

	// Decode: JSON bytes -> struct
	var wikiRes WikiPageRes
	err = json.NewDecoder(resp.Body).Decode(&wikiRes)

	if err != nil {
		return nil, err
	}

	return &wikiRes, nil
}
