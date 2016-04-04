package utils

import (
	"regexp"
	"strings"
)

// validate the request path is a v1 path
func ValidateV1(path string) bool {
	var validPath = regexp.MustCompile("^/(v1)/*.*")
	m := validPath.FindStringSubmatch(path)
	if m == nil {
		return false
	}
	return true
}

func SingleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}