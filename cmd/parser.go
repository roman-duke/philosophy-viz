package main

import (
	"errors"
	// "slices"
	// "fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// RULES FOR FINDING THE "IMPORTANT LINKS"
// - 1. Skip anything inside parentheses
// - 2. Skip italicized links
// - 3. Skip anything not in the main body ---> completed
// - 4. Skip links to the current page (self-links and same page-anchors) ---->
// - 5. Skip links that lead to "cycles"
// - 6. Skip red links and external links
// - 7. Skip non-article links - pronunciation/IPA help and namespaces like File:, Help:, Wikpedia:, etc

// This already guarantees that rules 4, 6 and 7
var wikiLink = regexp.MustCompile(`^/wiki/[^:#?]+$`)

// relevant ANSI sequence codes
const (
	Red   = "\033[31m"
	Reset = "\033[0m"
	Cyan  = "\033[36m"
)

// this function receives the returned decoded json as input, then parses to build a DOM tree
// (via the HTML5 implementation) and then searches the tree to look for the "relevant" links
func parser(arg *WikiPageRes, prevTitle string) (string, error) {
	wikiHTMLContent := arg.Parse.Text

	tokenizer := html.NewTokenizer(strings.NewReader(wikiHTMLContent))

	// Now we look for the relevant "<p>" tags, as the "relevant" links are embedded
	// in them.

	// we keep a small queue of the parenthesis and only execute the body of the StartTagToken
	// bit if the queue is empty (we don't care about anything embedded in parenthesis)
	pQ := make([]string, 0)

	// we also keep a small queue of the tag too
	tQ := make([]string, 0)
	targetAtom, wrapperAtom := atom.P, atom.P

	for {
		tag, err := tokenize(targetAtom, wrapperAtom, &pQ, &tQ, tokenizer)

		// pTag error handling
		pErr := tokenErrorHandler(err)
		if pErr != nil {
			return "", pErr
		}

		// remember to switch the "targetAtom" bit because there may be cases where
		// the link is not found in the first paragraph
		if tag != nil && tag.DataAtom == atom.P && tag.Type == html.StartTagToken {
			// fmt.Println("we have found the start of our paragraph token", *tag)
			targetAtom = atom.A
		}

		if tag != nil && tag.DataAtom == atom.P && tag.Type == html.EndTagToken {
			// fmt.Println("we have found the end of our paragraph token", *tag)
			targetAtom = atom.P
		} else if tag != nil && tag.DataAtom == atom.A && tag.Type == html.StartTagToken {
			// fmt.Println("we have found our anchor token", *tag)

			validLink, ok := getLink(*tag, prevTitle)

			if !ok {
				// continue with the scanning if the link is not valid
				continue
			}

			return validLink, nil
		}
	}
}

func getLink(token html.Token, blacklistTitle string) (string, bool) {
	address := ""

	for _, attr := range token.Attr {
		// this is where we check the validity of the link
		// links that go back to already visited pages are also deemed "invalid"
		if attr.Key == "href" && wikiLink.Match([]byte(attr.Val)) && strings.TrimPrefix(attr.Val, "/wiki/") != blacklistTitle {
			address = attr.Val
			return address, true
		}
	}

	return address, false
}

func tokenize(tAtom, twAtom atom.Atom, pQ, tags *[]string, tokenizer *html.Tokenizer) (*html.Token, error) {
	tokenType := tokenizer.Next()

	switch tokenType {
	case html.ErrorToken:
		if tokenizer.Err() == io.EOF {
			return nil, io.EOF
		}

		return nil, tokenizer.Err()

	case html.StartTagToken, html.EndTagToken:
		if len(*pQ) != 0 {
			return nil, nil
		}

		token := tokenizer.Token()

		// we only return the token if we encountered our "target" token
		// our main focus is "p" tags in the "main body" of the article
		if token.DataAtom == tAtom || token.DataAtom == twAtom && len(*tags) == 0 {
			return &token, nil
		}

		switch token.Type {
		case html.StartTagToken:
			*tags = append(*tags, token.Data)
		case html.EndTagToken:
			*tags = (*tags)[:len(*tags)-1]
		}

	case html.TextToken:
		if tAtom == atom.A {
			textT := tokenizer.Token()

			// unlike the leetcode problem, we are guaranteed to have the parenthesis formatted
			// in the correct order, hence we just pop or append based on the kind of parenthesis
			// we encounter
			trimmedText := strings.TrimSpace(textT.Data)

			if strings.Contains(trimmedText, "(") {
				*pQ = append(*pQ, "(")
			}

			if strings.Contains(trimmedText, ")") {
				*pQ = (*pQ)[:len(*pQ)-1]
			}
		}
	}

	return nil, nil
}

func tokenErrorHandler(err error) error {
	if errors.Is(err, io.EOF) {
		log.Printf("\033[31mend of file reached\033[0m")
		return io.EOF
	} else if err != nil {
		return err
	}

	return nil
}
