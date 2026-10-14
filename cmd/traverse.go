package main

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"
)

type Link struct {
	title       string
	redirect    bool
	linkAddress string
	// the loopBound shows that this link causes a loop
	// the nextLinkAddress just points at the next thing in the list
	loopBound bool
	loopLinks []Link
	// this is where a sum type would be important.
	// This should only be present when the value of
	// the redirect member is true, but sadly Go does
	// not have the sum-type language feature.
	redirectAddress string
}

// function to recursively get to philosophy
// this mode argument should ideally be a sum type
func getToPhilosophy(id, mode string, links []Link, params url.Values, c chan Result, wg *sync.WaitGroup) {
	var node string
	if len(links) > 0 {
		lastLink := links[len(links)-1]
		node = strings.TrimPrefix(lastLink.linkAddress, "/wiki/")

		if lastLink.redirect {
			node = strings.TrimPrefix(lastLink.redirectAddress, "/wiki/")
		} else if lastLink.loopBound {
			newLastLink := links[len(links)-2]
			node = strings.TrimPrefix(newLastLink.linkAddress, "/wiki")
		}
	}

	// declare relevant variables to help us with the tail recursion
	var (
		err      error
		loopMode bool
	)

	// our terminator case
	if node == "Philosophy" {
		// return links, nil
		defer wg.Done()

		return
	}

	// update the map of url.Values with the new page key
	params.Set("page", node)

	// ============================ SLEEP ===========================//
	// sleep a bit, so as not to constantly bombard Wiki's server
	time.Sleep(time.Millisecond * 500)
	// ==============================================================//
	res, err := fetch(client, BASE_URL, params)

	if err != nil {
		c <- FetchErr{err: err}
	}

	// call the parse function here to help us extract the relevant link
	dupFinder := findDuplicate(links)
	parseRes, err := parser(res, dupFinder)

	// declare our new link variable
	var newLink Link

	if errors.Is(err, ErrLoopLink) {
		loopMode = true
	} else if err != nil {
		c <- ParsingErr{err: err}
	}

	nextNode := strings.TrimPrefix(parseRes.linkAddr, "/wiki/")

	// handle this error here just for consistency purposes, but it is currently impossible
	// that there would be any error because of what the parser does for us under the hood
	// the error is actualy more so a "logical" error
	if len(nextNode) == len(parseRes.linkAddr) {
		c <- ParsingErr{err: errors.New("invalid link returned by parser")}
	}

	// update the values of the newLink
	newLink.title = nextNode
	newLink.linkAddress = parseRes.linkAddr

	// in redirect mode, just update the link of the previous element
	if parseRes.redirect {
		links[len(links)-1].redirect = true
		links[len(links)-1].redirectAddress = parseRes.linkAddr
		// send the result over a channel for an encountered redirect
		c <- RedirectLink{key: id, address: parseRes.linkAddr}
	}

	// if the new link causes a loop, then we pop it from our array and add
	// it as a "child" to the previous element in the array
	if loopMode {
		links[len(links)-1].loopBound = true
		fmt.Println("links that we have gathered so far!\n", links)
		c <- LoopLink{key: id}

		// in strict mode, we just stop the recursion immediately and return
		// it seems kind of weird to have a base case embedded in here though?
		if mode == "strict" {
			defer wg.Done()

			return
		}

		getToPhilosophy(id, mode, links, params, c, wg)
	} else {
		if !parseRes.redirect {
			links = append(links, newLink)
			c <- NewLink{key: id, val: newLink}
		}

		getToPhilosophy(id, mode, links, params, c, wg)
	}
}

// ============================== HELPER FUNCTION ===================================//
type DupKind string

const (
	DupNotInPrev DupKind = "duplicateNotInPrev"
	DupInPrev    DupKind = "duplicateInPrev"
	NoDup        DupKind = "noDuplicate"
)

// update made to the "unnecessary" recusrive function
func findDuplicate(ls []Link) func(s string) DupKind {
	// the name is self-explanatory
	// this helps us to know if our accumulated links slice contains the current link address

	var helper func(s string) DupKind

	helper = func(currLinkAddr string) DupKind {
		idx := slices.IndexFunc(ls, func(l Link) bool {
			return l.linkAddress == currLinkAddr
		})

		if idx != -1 && idx == len(ls)-1 {
			return DupInPrev
		} else if idx != -1 && idx != len(ls)-1 {
			return DupNotInPrev
		} else {
			return NoDup
		}
	}

	return helper
}
