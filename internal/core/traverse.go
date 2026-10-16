package core

import (
	"errors"
	// "fmt"
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
func (s *Service) getToPhilosophy(id string, mode TraverseMode, links []Link, params url.Values, c chan Result, wg *sync.WaitGroup) {
	var node string
	if len(links) > 0 {
		lastLink := links[len(links)-1]
		node = strings.TrimPrefix(lastLink.linkAddress, "/wiki/")

		if lastLink.redirect {
			node = strings.TrimPrefix(lastLink.redirectAddress, "/wiki/")
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
	res, err := s.fetch(BASE_URL, params)

	if err != nil {
		c <- FetchErr{err: err}
		defer wg.Done()

		return
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

		defer wg.Done()
		return
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
		lastElem := links[len(links)-1]
		lastElem.loopBound = true

		links = links[:len(links)-1]
		links[len(links)-1].loopLinks = append(links[len(links)-1].loopLinks, lastElem)

		c <- LoopLink{key: id}

		// in strict mode, we just stop the recursion immediately and return
		// it seems kind of weird to have a base case embedded in here though?
		if mode == "strict" {
			defer wg.Done()

			return
		}

		s.getToPhilosophy(id, mode, links, params, c, wg)
	} else {
		if !parseRes.redirect {
			links = append(links, newLink)
			c <- NewLink{key: id, val: newLink}
		}

		s.getToPhilosophy(id, mode, links, params, c, wg)
	}
}

// ============================== HELPER FUNCTION ===================================//
type DupKind string

const (
	DupNotInPrev DupKind = "duplicateNotInPrev"
	DupInPrev    DupKind = "duplicateInPrev"
	NoDup        DupKind = "noDuplicate"
)

// have a look at this "unnecessary" recursive function
func findDuplicate(ls []Link) func(s string) DupKind {
	// the name is self-explanatory
	// this helps us to know if our accumulated links slice contains the current link address
	// using this function in the main call site would lead to our implementation having a
	// time complexity of O(n^2 x m)-ish (we are ignoring the internal checks which happen too)
	// anyways, the sample size for this project (the list of links is going to be pretty small)

	var helper func(s string) DupKind

	// keep a copy of the main loop links
	copyLs := ls

	helper = func(currLinkAddr string) DupKind {
		idx := slices.IndexFunc(ls, func(l Link) bool {
			// first we check the link address of the node and then we check its children
			// (the loop links) if any exists
			// let's use recursion (for no just justified reason, loll)
			isEqualRoot := l.linkAddress == currLinkAddr
			var isInnerEq DupKind

			if len(l.loopLinks) != 0 {
				// update the links bit so that the next iteration runs on the loopLinks
				ls = l.loopLinks
				isInnerEq = helper(currLinkAddr)

				// point the loopLinks bit back to the original copy again
				ls = copyLs
			} else {
				return isEqualRoot
			}

			return isEqualRoot || (isInnerEq == DupInPrev || isInnerEq == DupNotInPrev)
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
