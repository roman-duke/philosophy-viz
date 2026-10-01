package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var BASE_URL = "https://en.wikipedia.org/w/api.php"
var PARAMS = url.Values{
	"action": {"parse"},
	"format": {"json"},
	"page":   {"Pet door"},
}

// let's make a simple fetch request to a wikipedia article to get the content and then
// log it to console
func main() {
	// create the necessarsy connection pool
	client := createClient()

	res, err := fetch(client, BASE_URL)
	if err != nil {
		log.Fatalf("\033[31merror fetching content: %v\033[0m", err)
	}

	// convert the response to a string for parsing by the parsing function
	// parser()

	err = os.WriteFile("dump.json", res, 0644)
	if err != nil {
		log.Fatalf("\033[31merror while writing to file: %v\033[0m", err)
	}
}

func createClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

func fetch(client *http.Client, urlString string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, urlString, nil)

	// use the default methods to add the necessary query params
	req.URL.RawQuery = url.Values.Encode(PARAMS)

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

	return io.ReadAll(io.LimitReader(resp.Body, 1<<15))
}
