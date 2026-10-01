package main

import "strings"

// RULES FOR FINDING THE "IMPORTANT LINKS"
// - Skip anything inside parentheses
// - Skip italicized links
// - Skip anything not in the main body
// - Skip links to the current page (self-links and same page-anchors)
// - Skip links that lead to "cycles"
// - Skip red links and external links
// - Skip non-article links - pronunciation/IPA help and namespaces like File:, Help:, Wikpedia:, etc

// this function receives the returned json as input and then parses to look for the relevant links
func parser(content string) {
	// the aim of this is to find the "important links"
	for {
		paragraphIdx := getParagraphIdx(content)

	}
}

func getParagraphIdx(s string) int {
	return strings.Index(s, "<p>")
}

func searchForValidLink(substring string) (string, bool) {
	target := "<a href=\""

	for {
		linkIdx := strings.Index(substring, "<a href=")

		if linkIdx == -1 {
			return "", false
		}

		// extract out the actual link


		// look for the opening parenthesis
		parenIndex := strings.Index(substring[:linkIdx], "(")
		if parenIndex != -1 {
			substring = substring[linkIdx+len(target):]
			continue
		}

		// if
	}
}
