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

// define the shape of the Wikipedia response
type WikiPageRes struct {
	Parse struct {
		Title     string     `json:"title"`
		Text      string     `json:"text"`
		LangLinks []LangLink `json:"langlinks"`
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
func (s *Service) fetch(urlString string, params url.Values) (*WikiPageRes, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, urlString, nil)

	// use the default methods to add the necessary query params
	req.URL.RawQuery = url.Values.Encode(params)

	if err != nil {
		return nil, err
	}

	// set the necessary User Agent header so that the request isn't blocked by Wikipedia
	req.Header.Set("User-Agent", "Path to Philosophy Crawler (https://philosophy-viz)")

	resp, err := s.client.Do(req)
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
