package adr

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	nonAlphanumHyphen = regexp.MustCompile(`[^a-z0-9-]`)
	multipleHyphens   = regexp.MustCompile(`-{2,}`)
)

// Slugify converts a title into a URL-friendly slug.
// Returns an error if the resulting slug is empty.
func Slugify(title string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(title))
	s = strings.ReplaceAll(s, " ", "-")
	s = nonAlphanumHyphen.ReplaceAllString(s, "")
	s = multipleHyphens.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")

	if s == "" {
		return "", fmt.Errorf("title %q produces an empty slug", title)
	}
	return s, nil
}
