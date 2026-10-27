package core

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
var wikiLink = regexp.MustCompile(`^/wiki/[^:#?]+$`)

// relevant ANSI sequence codes
const (
	Bold    = "\033[1m"
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Cyan    = "\033[36m"
	Yellow  = "\033[33m"
	Magenta = "\033[35m"
)

// html void elements that do not have a closing tag
var voidElementsMap = map[atom.Atom]bool{
	atom.Area:   true,
	atom.Br:     true,
	atom.Base:   true,
	atom.Col:    true,
	atom.Embed:  true,
	atom.Hr:     true,
	atom.Img:    true,
	atom.Input:  true,
	atom.Link:   true,
	atom.Meta:   true,
	atom.Param:  true,
	atom.Source: true,
	atom.Track:  true,
	atom.Wbr:    true,
	atom.B:      true,
}

// ===================================== errors ============================================//
var ErrLoopLink = errors.New("link causes loop")

type ParseRes struct {
	linkAddr string
	redirect bool
	isLoop   bool
}

// this function receives the returned decoded json as input, then parses to build a DOM tree
// (via the HTML5 implementation) and then searches the tree to look for the "relevant" links
func parser(arg *WikiPageRes, findD func(s string) DupKind) (*ParseRes, error) {
	wikiHTMLContent := arg.Parse.Text

	tokenizer := html.NewTokenizer(strings.NewReader(wikiHTMLContent))

	// Now we look for the relevant "<p>" tags, as the "relevant" links are embedded
	// in them.

	// we keep a small queue of the parenthesis and only execute the body of the StartTagToken
	// bit if the queue is empty (we don't care about anything embedded in parenthesis)
	pQ := make([]string, 0)

	// we also keep a small queue of the tag too
	tQ := make([]string, 0)

	// our targetAtom starts out as a div; we then check if we actually got content or we got a
	// redirect. In the case where we got content, we switch the target atom to look for a "p" tag
	// and then "a" afterwards. If we instead got a redirectMsg, then we just look for the redirect
	// link
	targetAtom, wrapperAtom := atom.P, atom.P

	// this tells us if we should search for anchor tags within a "p" tag or within any tag in the case
	// of our redirect response
	redirectFlag := false

	for {
		tag, err := tokenize(targetAtom, wrapperAtom, &pQ, &tQ, redirectFlag, tokenizer)

		// pTag error handling
		pErr := tokenErrorHandler(err)
		if pErr != nil {
			return nil, pErr
		}

		// remember to switch the "targetAtom" bit so as not to scan for anchor tags in irrelevant
		// places
		if tag == nil {
			continue
		} else if tag.DataAtom == atom.Div {
			// if the returned tag is a div, then it means we have encountered a redirect. In that case,
			// switch to look for the anchor link, else continue the tokenization process normally.
			// else continue the tokenization process normally.
			redirectFlag = true
			targetAtom = atom.A
		} else if tag.DataAtom == atom.P && tag.Type == html.StartTagToken {
			// fmt.Println("we have found the start of our paragraph token", *tag)
			targetAtom = atom.A
		} else if tag.DataAtom == atom.P && tag.Type == html.EndTagToken && !redirectFlag {
			// switch back to look for p tags in the case that our redirect flag is false (the normal case)
			targetAtom = atom.P
		} else if tag.DataAtom == atom.A && tag.Type == html.StartTagToken {
			validLink, ok := getLink(*tag)

			if !ok {
				// continue with the scanning if the link is not valid
				continue
			}

			// checks if our element contains a dupicate and if the duplicate
			// has previously been discovered
			dup := findD(validLink)

			switch dup {
			case DupInPrev:
				continue

			case DupNotInPrev:
				return &ParseRes{linkAddr: validLink, redirect: redirectFlag, isLoop: true}, nil

			case NoDup:
				return &ParseRes{linkAddr: validLink, redirect: redirectFlag}, nil
			}
		}
	}
}

func getLink(token html.Token) (string, bool) {
	address := ""

	for _, attr := range token.Attr {
		// this is where we check the validity of the link
		// links that go back to already visited pages are also deemed "invalid"
		if attr.Key == "href" {
			// we trim any discovered "#"suffix in the link
			before, _, _ := strings.Cut(attr.Val, "#")
			if wikiLink.Match([]byte(before)) {
				address = attr.Val
				return address, true
			}
		}
	}

	return address, false
}

func tokenize(tAtom, twAtom atom.Atom, pQ, tags *[]string, r bool, tokenizer *html.Tokenizer) (*html.Token, error) {
	tokenType := tokenizer.Next()

	switch tokenType {
	case html.ErrorToken:
		fmt.Println("state of the parenthesis queue", pQ)
		fmt.Println("state of the tags queue", tags)
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
		if r && token.DataAtom == tAtom {
			return &token, nil
		} else if (token.DataAtom == tAtom || token.DataAtom == twAtom) && len(*tags) == 0 {
			return &token, nil
		} else if token.DataAtom == atom.Div {
			// if we encounter a div, check the value of the class attribute to see if we
			// got a redirect message
			for _, attr := range token.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "redirectMsg") {
					return &token, nil
				} else if attr.Key == "class" && strings.Contains(attr.Val, "mw-parser-output") {
					return nil, nil
				}
			}
		}

		// it might be best to just have a whitelist of tags that the anchor links can
		// be found in, but given that the currently observed case is just one (bold tags)
		// then that is basically an overkill at the moment
		if !voidElementsMap[token.DataAtom] {
			switch token.Type {
			case html.StartTagToken:
				*tags = append(*tags, token.Data)
			case html.EndTagToken:
				if len(*tags) > 0 {
					*tags = (*tags)[:len(*tags)-1]
				}
			}
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
