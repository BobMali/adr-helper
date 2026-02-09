package adr

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

var adrFilePattern = regexp.MustCompile(`^(\d{4,})-.*\.md$`)

// FormatFilename returns the ADR filename for the given number and title.
func FormatFilename(number int, title string) (string, error) {
	slug, err := Slugify(title)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d-%s.md", number, slug), nil
}

// NextNumber scans dir for existing ADR files and returns max+1, or 1 if none found.
func NextNumber(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("reading directory %q: %w", dir, err)
	}

	max := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := adrFilePattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}
		n, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}
		if n > max {
			max = n
		}
	}
	return max + 1, nil
}
