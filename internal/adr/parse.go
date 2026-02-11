package adr

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Metadata holds parsed metadata extracted from an ADR's raw markdown content.
type Metadata struct {
	Number int
	Title  string
	Status string
	Date   string
}

var numberedHeadingPattern = regexp.MustCompile(`(?m)^# (\d+)\.\s+(.+)$`)
var plainHeadingPattern = regexp.MustCompile(`(?m)^# (.+)$`)
var bodyDatePattern = regexp.MustCompile(`(?mi)^[Dd]ate:\s*(.+)$`)
var frontmatterDatePattern = regexp.MustCompile(`(?m)^date:\s*(.+)$`)

// ExtractMetadata parses an ADR's raw markdown content and returns structured metadata.
func ExtractMetadata(content string) Metadata {
	var m Metadata

	// Number + Title from heading
	if matches := numberedHeadingPattern.FindStringSubmatch(content); len(matches) == 3 {
		// Safe to ignore error — regex guarantees digits
		n := 0
		for _, ch := range matches[1] {
			n = n*10 + int(ch-'0')
		}
		m.Number = n
		m.Title = strings.TrimSpace(matches[2])
	} else if matches := plainHeadingPattern.FindStringSubmatch(content); len(matches) == 2 {
		m.Title = strings.TrimSpace(matches[1])
	}

	// Status
	if hasStatusSection(content) {
		m.Status = extractStatusSectionContent(content)
	} else if hasFrontmatterStatus(content) {
		m.Status = getFrontmatterStatusValue(content)
	}

	// Date — prefer body "Date:" line, fall back to frontmatter "date:"
	body := bodyAfterFrontmatter(content)
	if matches := bodyDatePattern.FindStringSubmatch(body); len(matches) == 2 {
		m.Date = strings.TrimSpace(matches[1])
	} else {
		fm := extractFrontmatter(content)
		if matches := frontmatterDatePattern.FindStringSubmatch(fm); len(matches) == 2 {
			m.Date = stripQuotes(strings.TrimSpace(matches[1]))
		}
	}

	return m
}

// MetadataToADR converts raw Metadata strings into a typed ADR.
// Uses fallbackNumber when m.Number is 0 (e.g. MADR-style headings without number prefix).
// Returns error on invalid status; zero time.Time if date can't be parsed.
func MetadataToADR(m Metadata, fallbackNumber int) (ADR, error) {
	number := m.Number
	if number == 0 {
		number = fallbackNumber
	}

	var status Status
	// Extract first non-empty line for status parsing.
	// Status sections may contain additional lines (e.g. "Supersedes ..." references).
	statusLine := firstNonEmptyLine(m.Status)
	if statusLine == "" {
		status = Proposed
	} else {
		var ok bool
		status, ok = ParseStatus(statusLine)
		if !ok {
			return ADR{}, fmt.Errorf("invalid status %q", m.Status)
		}
	}

	var date time.Time
	if m.Date != "" {
		parsed, err := time.Parse("2006-01-02", m.Date)
		if err == nil {
			date = parsed
		}
	}

	return ADR{
		Number: number,
		Title:  m.Title,
		Status: status,
		Date:   date,
	}, nil
}

// firstNonEmptyLine returns the first non-blank line from s, trimmed.
func firstNonEmptyLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		t := strings.TrimSpace(line)
		if t != "" {
			return t
		}
	}
	return ""
}

// bodyAfterFrontmatter returns content after the YAML frontmatter block,
// or the full content if there is no frontmatter.
func bodyAfterFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}
	rest := content[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return content
	}
	// Skip past the closing --- line
	after := rest[idx+4:]
	if len(after) > 0 && after[0] == '\n' {
		after = after[1:]
	}
	return after
}

// stripQuotes removes surrounding double quotes if present.
func stripQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
