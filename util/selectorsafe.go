package util

import "strings"

func SelectorSafe(in string) string {
	// Prepend with a because browsers reject querySelectors that start with a number
	// Replace . with - for the same reason
	return strings.ReplaceAll("a"+in, ".", "-")
}
