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

	visitedLinks := make([]string, 0)

	paths, err := getToPhilosophy("Alireza Firouzja", "", visitedLinks, PARAMS)

	if err != nil {
		log.Fatalf("\033[31merror fetching content: %v\033[0m", err)
	}

	fmt.Printf("The path to philosophy --> \033[36m%s\033[0m", paths)
}

// function to recursively get to philosophy
func getToPhilosophy(node, prevPage string, links []string, params url.Values) ([]string, error) {

	for node != "Philosophy" {
		fmt.Printf("%s%sour node for this iteration: (%s)!%s\n", Bold, Red, node, Reset)

		// update the map of url.Values with the new page key
		params.Set("page", node)

		res, err := fetch(client, BASE_URL, params)

		if err != nil {
			return links, err
		}

		// pass the pointer to the res body to the parser func to look for the valid link
		linkAddr, err := parser(res, prevPage)

		if err != nil {
			return links, err
		}

		nextNode := strings.TrimPrefix(linkAddr, "/wiki/")

		fmt.Println("Here is our next node:", nextNode)
		fmt.Println("length of our new node", len(nextNode))

		// handle this error here just for consistency purposes, but it is currently impossible
		// that there would be any error because of what the parser does for us under the hood
		// the error is actualy more so a "logical" error
		if len(nextNode) == len(linkAddr) {
			return links, errors.New("invalid link returned by parser")
		}

		// set prevPage to be current node (parser uses this internally)
		prevPage = node

		// sleep a bit, so as not to constantly bombard Wiki's server
		links = append(links, linkAddr)
		time.Sleep(time.Millisecond * 20)

		fmt.Printf("Here is our next node just before we run recursively call bro! -> %s%s(%s)%s\n", Bold, Cyan, nextNode, Reset)
		// missing return or break statemen here causes the previous function at the top of the stack to be stuck
		// actually, this is so retarded lolll. Damn, why do we have recursion and iteration combined when they
		// are "serving" the same purpose?
		getToPhilosophy(nextNode, prevPage, links, params)
	}

	return links, nil
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
