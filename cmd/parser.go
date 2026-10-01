package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"golang.org/x/net/html"
)

// RULES FOR FINDING THE "IMPORTANT LINKS"
// - Skip anything inside parentheses
// - Skip italicized links
// - Skip anything not in the main body
// - Skip links to the current page (self-links and same page-anchors)
// - Skip links that lead to "cycles"
// - Skip red links and external links
// - Skip non-article links - pronunciation/IPA help and namespaces like File:, Help:, Wikpedia:, etc

// this function receives the returned decoded json as input, then parses to build a DOM tree
// (via the HTML5 implementation) and then searches the tree to look for the "relevant" links
func parser(arg *WikiPageRes) (string, error) {
	wikiHTMLContent := arg.Parse.Text

	tokenizer := html.NewTokenizer(strings.NewReader(wikiHTMLContent))

	count := 0

	// Now we look for the relevant "<p>" tags
	for {
		tokenType := tokenizer.Next()

		switch tokenType {
		case html.ErrorToken:
			if tokenizer.Err() == io.EOF {
				log.Printf("\033[31mend of file reached\033[0m")
				return "", io.EOF
			}

			return "", tokenizer.Err()

		case html.StartTagToken, html.SelfClosingTagToken:
			if tokenName := tokenizer.Token().Data; tokenName == "p" {
				fmt.Println("Found a <p> tag")
				fmt.Printf("the paragraph token has been found %d times\n", count+1)
				count++
			}
		}
	}
}
