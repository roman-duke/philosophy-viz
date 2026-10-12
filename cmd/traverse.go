package main

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
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
func getToPhilosophy(node string, mode string, links []Link, params url.Values) ([]Link, error) {

	// declare relevant variables to help us with the tail recursion
	var (
		ts       []Link
		err      error
		loopMode bool
	)

	// our terminator case
	if node == "Philosophy" {
		return links, nil
	}

	// fmt.Printf("%s%sour node for this iteration: (%s)!%s\n", Bold, Red, node, Reset)

	fmt.Printf("%s%sHere is our current list of titles so far:\n %v%s\n", Bold, Red, links, Reset)

	// update the map of url.Values with the new page key
	params.Set("page", node)

	// ============================ SLEEP ===========================//
	// sleep a bit, so as not to constantly bombard Wiki's server
	time.Sleep(time.Millisecond * 500)
	// ==============================================================//
	res, err := fetch(client, BASE_URL, params)

	if err != nil {
		return links, err
	}

	// call the parse function here to help us extract the relevant link
	dupFinder := findDuplicate(links)
	parseRes, err := parser(res, dupFinder)

	// declare our new link variable
	var newLink Link

	if errors.Is(err, ErrLoopLink) {
		loopMode = true
	} else if err != nil {
		return links, err
	}

	nextNode := strings.TrimPrefix(parseRes.linkAddr, "/wiki/")

	fmt.Println("Here is our next node:", nextNode)
	fmt.Println("length of our new node", len(nextNode))

	// handle this error here just for consistency purposes, but it is currently impossible
	// that there would be any error because of what the parser does for us under the hood
	// the error is actualy more so a "logical" error
	if len(nextNode) == len(parseRes.linkAddr) {
		return links, errors.New("invalid link returned by parser")
	}

	// update the values of the newLink
	newLink.title = nextNode
	newLink.linkAddress = parseRes.linkAddr

	// in redirect mode, just update the link of the previous element
	if parseRes.redirect {
		links[len(links)-1].redirect = true
		links[len(links)-1].redirectAddress = parseRes.linkAddr
	}

	// if the new link causes a loop, then we pop it from our array and add
	// it as a "child" to the previous element in the array
	if loopMode {
		lastElem := links[len(links)-1]
		lastElem.loopBound = true

		links = links[:len(links)-1]
		links[len(links)-1].loopLinks = append(links[len(links)-1].loopLinks, lastElem)

		// the call would be next node of the last "valid" link
		n := links[len(links)-1].title

		// in strict mode, we just stop the recursion immediately and return
		// it seems kind of weird to have a base case embedded in here though?
		if mode == "strict" {
			return links, err
		}

		ts, err = getToPhilosophy(n, mode, links, params)
	} else {
		if !parseRes.redirect {
			links = append(links, newLink)
		}

		ts, err = getToPhilosophy(nextNode, mode, links, params)
	}

	return ts, err
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
