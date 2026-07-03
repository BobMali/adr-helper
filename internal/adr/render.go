package adr

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	headingPattern   = regexp.MustCompile(`(?m)^# .*$`)
	dateUpperPattern = regexp.MustCompile(`(?m)^Date:.*$`)
	dateLowerPattern = regexp.MustCompile(`(?m)^date:.*$`)
	metaValueNewline = regexp.MustCompile(`[\r\n]+`)
)

// ReplaceMetaField replaces the value of the first title-block metadata line
// matching "Label:" (case-insensitive) with "Label: value", emitting the label
// in the canonical spelling passed by the caller. Any newlines in value are
// collapsed to single spaces to preserve the single-line invariant of a
// title-block field. Returns (result, found); found is false when no matching
// line exists (e.g. a template without that field).
func ReplaceMetaField(content, label, value string) (string, bool) {
	sanitized := strings.TrimSpace(metaValueNewline.ReplaceAllString(value, " "))
	pattern := regexp.MustCompile(`(?mi)^` + regexp.QuoteMeta(label) + `:.*$`)

	found := false
	result := pattern.ReplaceAllStringFunc(content, func(match string) string {
		if found {
			return match
		}
		found = true
		return label + ": " + sanitized
	})
	return result, found
}

// ReplaceSectionContent replaces the body text under the first matching
// heading (## or ###) with newBody. Returns (result, found).
// The heading match is case-insensitive on the heading text.
func ReplaceSectionContent(content, heading, newBody string) (string, bool) {
	lines := strings.Split(content, "\n")
	lowerHeading := strings.ToLower(strings.TrimSpace(heading))

	// Find the heading line
	headingIdx := -1
	headingLevel := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		level, text := parseHeadingLine(trimmed)
		if level > 0 && strings.ToLower(text) == lowerHeading {
			headingIdx = i
			headingLevel = level
			break
		}
	}

	if headingIdx < 0 {
		return content, false
	}

	// Find the end of this section: next heading of equal or higher level (lower number), or EOF
	bodyStart := headingIdx + 1
	bodyEnd := len(lines)
	for i := bodyStart; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		level, _ := parseHeadingLine(trimmed)
		if level > 0 && level <= headingLevel {
			bodyEnd = i
			break
		}
	}

	// Build the result: heading line + blank line + newBody + blank line + rest
	var b strings.Builder
	for i := 0; i <= headingIdx; i++ {
		b.WriteString(lines[i])
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(newBody)
	b.WriteString("\n")
	if bodyEnd < len(lines) {
		b.WriteString("\n")
		for i := bodyEnd; i < len(lines); i++ {
			b.WriteString(lines[i])
			if i < len(lines)-1 {
				b.WriteString("\n")
			}
		}
	}

	return b.String(), true
}

// parseHeadingLine returns (level, text) for a markdown heading line.
// e.g. "## Context" -> (2, "Context"), "### Foo" -> (3, "Foo").
// Returns (0, "") if not a heading.
func parseHeadingLine(line string) (int, string) {
	if !strings.HasPrefix(line, "#") {
		return 0, ""
	}
	level := 0
	for _, ch := range line {
		if ch == '#' {
			level++
		} else {
			break
		}
	}
	if level > 6 {
		return 0, ""
	}
	text := strings.TrimSpace(line[level:])
	return level, text
}

// RenderTemplate replaces the first top-level heading and date lines in template content
// with values from the given ADR record.
func RenderTemplate(content string, record *ADR) string {
	heading := fmt.Sprintf("# %d. %s", record.Number, record.Title)
	dateStr := record.Date.Format("2006-01-02")

	// Replace only the first top-level heading
	replaced := false
	result := headingPattern.ReplaceAllStringFunc(content, func(match string) string {
		if !replaced {
			replaced = true
			return heading
		}
		return match
	})

	// Replace Date: and date: lines
	result = dateUpperPattern.ReplaceAllString(result, "Date: "+dateStr)
	result = dateLowerPattern.ReplaceAllString(result, "date: "+dateStr)

	// Set default status from the ADR record
	if hasStatusSection(result) {
		result = replaceStatusSectionContent(result, record.Status.String())
	} else if hasFrontmatterStatus(result) {
		result = replaceFrontmatterStatus(result, strings.ToLower(record.Status.String()))
	}

	return result
}
