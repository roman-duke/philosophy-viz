package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type LangLink struct {
	Lang     string `json:"lang"`
	Url      string `json:"url"`
	Langname string `json:"langname"`
	Autonym  string `json:"autonym"`
	Title    string `json:"title"`
}

// define the shape of the Wikipedia page response
type WikiPageRes struct {
	Parse struct {
		Title     string     `json:"title"`
		Text      string     `json:"text"`
		LangLinks []LangLink `json:"langlinks"`
	}
}

// define the shape of the Wikipedia query
// (search for titles) response
type SearchSchema struct {
	Title     string `json:"title"`
	PageId    int    `json:"pageid"`
	Size      int    `json:"size"`
	Wordcount int    `json:"wordcount"`
	Snippet   string `json:"snippet"`
	TimeStamp string `json:"timestamp"`
}

type WikiQueryRes struct {
	Query struct {
		SearchArr []SearchSchema `json:"search"`
	}
}

// the type of the result to expect
type PathsMap map[string][]Link

func createClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

// given that this does something that is stateful, we have to pass a context to it
// so that we can cancel on cleanup or something. I think that is the idiomatic way of
// doing it in Go??
func fetch[T any](c *http.Client, urlString string, params url.Values) (*T, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, urlString, nil)

	if err != nil {
		return nil, err
	}

	// use the default methods to add the necessary query params
	req.URL.RawQuery = url.Values.Encode(params)

	// set the necessary User Agent header so that the request isn't blocked by Wikipedia
	req.Header.Set("User-Agent", "Path to Philosophy Crawler (https://philosophy-viz)")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %v", resp.Status)
	}

	// Decode: JSON bytes -> struct
	var wikiRes T
	err = json.NewDecoder(resp.Body).Decode(&wikiRes)

	if err != nil {
		return nil, err
	}

	return &wikiRes, nil
}
