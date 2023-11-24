package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kylelemons/godebug/diff"
)

func PrettyJsonString(input string) string {
	var out bytes.Buffer
	json.Indent(&out, []byte(input), "", "  ")
	return out.String()
}

var Diff = diff.Diff

func DiffOnly(one, two string) string {
	input := diff.Diff(one, two)
	parts := strings.Split(input, "\n")
	relevant := []string{}
	for _, part := range parts {
		if strings.Index(part, "+") == 0 || strings.Index(part, "-") == 0 {
			if string(part[len(part)-1]) == "," {
				relevant = append(relevant, part[1:len(part)-1])
			} else {
				relevant = append(relevant, part[1:])
			}
		}
	}
	pairs := []string{}
	hold := ""
	for _, part := range relevant {
		if hold == "" {
			hold = part
		} else {
			sep := strings.Index(part, ":")
			pairs = append(pairs, fmt.Sprintf("%s -> %s", hold, part[sep+1:]))
			hold = ""
		}
	}
	if hold != "" {
		pairs = append(pairs, hold)
	}
	return strings.Join(pairs, " ")
}
