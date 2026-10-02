package main

import (
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

// this function receives the returned decoded json as input, then parses to build a DOM tree
// (via the HTML5 implementation) and then searches the tree to look for the "relevant" links
func parser(arg *WikiPageRes) (string, error) {
	wikiHTMLContent := arg.Parse.Text

	tokenizer := html.NewTokenizer(strings.NewReader(wikiHTMLContent))

	// Now we look for the relevant "<p>" tags, as the "relevant" links are embedded
	// in them.
	for {
		tokenType := tokenizer.Next()

		switch tokenType {
		case html.ErrorToken:
			if tokenizer.Err() == io.EOF {
				log.Printf("\033[31mend of file reached\033[0m")
				return "", io.EOF
			}

			return "", tokenizer.Err()

		case html.StartTagToken:
			token := tokenizer.Token()
			// remember to bring this back and use the regular string comparison to see the performance
			// degradation (this current approach is faster as it's just an integer comparison)

			if token.DataAtom == atom.P {
				// check for all the children of this node (and then identify the "anchor" links)

				// we keep a small queue of the parenthesis and only execute the body of the StartTagToken
				// bit if the queue is empty
				pQ := make([]string, 0)

				for {
					innerType := tokenizer.Next()

					switch innerType {
					case html.ErrorToken:
						if tokenizer.Err() == io.EOF {
							log.Printf("\033[31mend of file reached\033[0m")
							return "", io.EOF
						}

						return "", tokenizer.Err()

					case html.TextToken:
						textT := tokenizer.Token()

						// unlike the leetcode problem, we are guaranteed to have the parenthesis formatted
						// in the correct order. hence we just pop or append based on the kind of parenthesis
						// we encounter
						switch textT.Data {
						case "(":
							pQ = append(pQ, "(")

						case ")":
							pQ = pQ[:len(pQ)-1]
						}

					case html.StartTagToken:
						if len(pQ) != 0 {
							continue
						}

						innerToken := tokenizer.Token()

						if innerToken.DataAtom == atom.A {
							validLink, ok := getLink(innerToken)

							if !ok {
								// just continue scanning if the link is not valid
								continue
							}

							return validLink, nil
						}
					}
				}
			}
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
