package util

import (
	"net/url"
	"strings"
)

func RootPath(in *url.URL) string {
	pathParts := strings.Split(in.Path, "/")
	l := len(pathParts)
	if l == 0 {
		return ""
	}
	if l > 1 && pathParts[0] == "" {
		return pathParts[1]
	}
	return pathParts[0]
}
