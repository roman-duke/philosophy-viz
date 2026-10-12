package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
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

// ======================================== errors =========================================//
var ErrDuplicateSearch = errors.New("error finding link duplicate")

// let's make a simple fetch request to a wikipedia article to get the content and then
// log it to console
func main() {
	// create the necessarsy connection pool
	client = createClient()

	// placed here temporarily just so that the visitedTitles slice is not empty
	visitedTitles := []Link{
		{
			title:       "Set Theory",
			linkAddress: "/wiki/Set_Theory",
		},
	}

	paths, err := getToPhilosophy("Set Theory", "relaxed", visitedTitles, PARAMS)

	if err != nil {
		log.Fatalf("\033[31merror fetching content: %v\033[0m", err)
	}

	titles := make([]string, len(paths))
	for i, link := range paths {
		titles[i] = link.title
	}

	fmt.Printf("The ugly path to philosophy --> \033[36m%v\033[0m\n\n\n", paths)
	fmt.Printf("The path to philosophy --> \033[36m%s\033[0m\n", strings.Join(titles, " → "))
}

type PhiloPath struct {
	title   string
	address string
	path    []Link
}

func visitPaths() {
	// create a list of paths to traverse to "Philosophy"
	paths := []PhiloPath{
		{
			title:       "Albert Einstein",
			address: "/wiki/Albert_Einstein",
		},
		{
			title:       "Alireza Firouzja",
			address: "/wiki/Alireza_Firouzja",
		},
	}

	// now create a channel that we can use to collect the results of the goroutines
	// Actually, I am not sure if we should use the CSP paradigm here or the default
	// concurrency paradgim??
	// c := make(chan )
}
