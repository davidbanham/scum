package util

import "strings"

func FirstFiveChars(input string) string {
	return strings.ToUpper(Truncate(5, input))
}

func FirstChar(input string) string {
	return strings.ToUpper(Truncate(1, input))
}
