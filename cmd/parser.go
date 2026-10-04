package main

import (
	"errors"
	"fmt"
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
var wikiLink = regexp.MustCompile(`^/wiki/[\w%-]+$`)

// relevant ANSI sequence codes
const (
	Red   = "\033[31m"
	Reset = "\033[0m"
	Cyan  = "\033[36m"
)

// this function receives the returned decoded json as input, then parses to build a DOM tree
// (via the HTML5 implementation) and then searches the tree to look for the "relevant" links
func parser(arg *WikiPageRes) (string, error) {
	wikiHTMLContent := arg.Parse.Text

	tokenizer := html.NewTokenizer(strings.NewReader(wikiHTMLContent))

	// Now we look for the relevant "<p>" tags, as the "relevant" links are embedded
	// in them.

	// we keep a small queue of the parenthesis and only execute the body of the StartTagToken
	// bit if the queue is empty (we don't care about anything embedded in parenthesis)
	pQ := make([]string, 0)
	targetAtom := atom.P

	for {
		tag, err := tokenize(targetAtom, &pQ, tokenizer)

		// pTag error handling
		pErr := tokenErrorHandler(err)
		if pErr != nil {
			return "", pErr
		}

		// remember to switch the "targetAtom" bit because there may be cases where
		// the link is not found in the first paragraph
		if tag != nil && targetAtom == atom.P {
			fmt.Println("we have found our paragraph token", *tag)
			targetAtom = atom.A
		} else if tag != nil && targetAtom == atom.A {
			fmt.Println("we have found our token", *tag)

			validLink, ok := getLink(*tag)

			if !ok {
				// continue with the scanning if the link is not valid
				continue
			}

			return validLink, nil
		}
	}
}

func getLink(token html.Token) (string, bool) {
	address := ""

	for _, attr := range token.Attr {
		// this is where we check the validity of the link
		if attr.Key == "href" && wikiLink.Match([]byte(attr.Val)) {
			address = attr.Val
			break
		}
	}

	return address, true
}

func tokenize(tAtom atom.Atom, pQ *[]string, tokenizer *html.Tokenizer) (*html.Token, error) {
	tokenType := tokenizer.Next()

	switch tokenType {
	case html.ErrorToken:
		if tokenizer.Err() == io.EOF {
			return nil, io.EOF
		}

		return nil, tokenizer.Err()

	case html.StartTagToken:
		if len(*pQ) != 0 {
			return nil, nil
		}

		token := tokenizer.Token()

		if token.DataAtom == tAtom {
			return &token, nil
		}

	case html.TextToken:
		textT := tokenizer.Token()

		// unlike the leetcode problem, we are guaranteed to have the parenthesis formatted
		// in the correct order, hence we just pop or append based on the kind of parenthesis
		// we encounter
		switch textT.Data {
		case "(":
			*pQ = append(*pQ, "(")
		case ")":
			*pQ = (*pQ)[:len(*pQ)-1]
		}

		return nil, nil
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
