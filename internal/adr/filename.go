package adr

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

var adrFilePattern = regexp.MustCompile(`^(\d{4,})-.*\.md$`)

// IsADRFilename reports whether name matches the ADR file naming convention
// (NNNN-*.md, four or more leading digits). name must be a bare filename, not a
// path: passing a value that may contain '/' or '\' can yield a misleading result.
func IsADRFilename(name string) bool {
	return adrFilePattern.MatchString(name)
}

// adrFile is an ADR markdown file discovered in a directory.
type adrFile struct {
	Number int
	Name   string
}

// listADRFiles returns the ADR files in dir (NNNN-*.md), parsed and in
// os.ReadDir order. Directories, non-ADR names, and files whose number can't be
// parsed are skipped. The raw os.ReadDir error is returned unwrapped so callers
// can wrap it or check os.IsNotExist.
//
// It is deliberately a thin "matches convention + parses to a number" primitive:
// keep all caller-specific filtering in the callers, not here.
func listADRFiles(dir string) ([]adrFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []adrFile
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
		files = append(files, adrFile{Number: n, Name: entry.Name()})
	}
	return files, nil
}

// FormatFilename returns the ADR filename for the given number and title.
func FormatFilename(number int, title string) (string, error) {
	slug, err := Slugify(title)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d-%s.md", number, slug), nil
}

// FindADRFile finds the ADR file with the given number in dir and returns its filename.
func FindADRFile(dir string, number int) (string, error) {
	files, err := listADRFiles(dir)
	if err != nil {
		return "", fmt.Errorf("reading directory %q: %w", dir, err)
	}

	for _, f := range files {
		if f.Number == number {
			return f.Name, nil
		}
	}
	return "", fmt.Errorf("ADR %04d: %w", number, ErrNotFound)
}

// NextNumber scans dir for existing ADR files and returns max+1, or 1 if none found.
func NextNumber(dir string) (int, error) {
	files, err := listADRFiles(dir)
	if err != nil {
		return 0, fmt.Errorf("reading directory %q: %w", dir, err)
	}

	max := 0
	for _, f := range files {
		if f.Number > max {
			max = f.Number
		}
	}
	return max + 1, nil
}
