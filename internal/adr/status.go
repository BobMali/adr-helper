package adr

import (
	"fmt"
	"regexp"
	"strings"
)

// SupersedesLink holds the number and filename for a cross-reference.
type SupersedesLink struct {
	Number   int
	Filename string
}

func formatLink(link SupersedesLink) string {
	return fmt.Sprintf("[ADR-%04d](%s)", link.Number, link.Filename)
}

var statusSectionPattern = regexp.MustCompile(`(?m)^## Status[ \t]*$`)

// hasStatusSection checks for a ## Status heading.
func hasStatusSection(content string) bool {
	return statusSectionPattern.MatchString(content)
}

// hasFrontmatterStatus checks for status: within YAML frontmatter.
func hasFrontmatterStatus(content string) bool {
	fm := extractFrontmatter(content)
	if fm == "" {
		return false
	}
	return regexp.MustCompile(`(?m)^status:.*$`).MatchString(fm)
}

// extractFrontmatter returns the YAML block between first two --- lines, or "".
func extractFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return ""
	}
	// Find the second ---
	rest := content[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return ""
	}
	// Return content between first --- and second --- (excluding the --- markers)
	return rest[:idx]
}

// replaceStatusSectionContent replaces the text between ## Status and the next ## heading (or EOF).
func replaceStatusSectionContent(content, newContent string) string {
	// Find ## Status heading
	loc := statusSectionPattern.FindStringIndex(content)
	if loc == nil {
		return content
	}

	// Find the end of the heading line + blank line separator
	afterHeading := loc[1]

	// Find next ## heading after status section
	restAfterHeading := content[afterHeading:]
	nextHeadingPattern := regexp.MustCompile(`(?m)\n\n## `)
	nextLoc := nextHeadingPattern.FindStringIndex(restAfterHeading)

	if nextLoc != nil {
		// Replace content between ## Status heading and next ## heading
		return content[:afterHeading] + "\n\n" + newContent + content[afterHeading+nextLoc[0]:]
	}

	// Status is the last section — replace to end of file
	// Preserve trailing newline
	return content[:afterHeading] + "\n\n" + newContent + "\n"
}

// replaceFrontmatterStatus replaces the status: line in YAML frontmatter only.
func replaceFrontmatterStatus(content, newValue string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}

	// Find the second --- line
	rest := content[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return content
	}

	frontmatter := rest[:idx]
	afterFrontmatter := rest[idx:]

	// Replace status: line within frontmatter only
	statusLine := regexp.MustCompile(`(?m)^status:.*$`)
	newFrontmatter := statusLine.ReplaceAllString(frontmatter, "status: \""+newValue+"\"")

	return "---" + newFrontmatter + afterFrontmatter
}

// extractStatusSectionContent returns the text between ## Status heading and the next ## heading (or EOF).
// Returns empty string if the section has no content.
func extractStatusSectionContent(content string) string {
	loc := statusSectionPattern.FindStringIndex(content)
	if loc == nil {
		return ""
	}

	afterHeading := loc[1]
	restAfterHeading := content[afterHeading:]

	nextHeadingPattern := regexp.MustCompile(`(?m)\n\n## `)
	nextLoc := nextHeadingPattern.FindStringIndex(restAfterHeading)

	var sectionBody string
	if nextLoc != nil {
		sectionBody = restAfterHeading[:nextLoc[0]]
	} else {
		sectionBody = restAfterHeading
	}

	return strings.TrimSpace(sectionBody)
}

// appendToStatusSectionContent appends text below existing status section content.
// If the section is empty, it just sets the content to appendText.
func appendToStatusSectionContent(content, appendText string) string {
	existing := extractStatusSectionContent(content)
	if existing == "" {
		return replaceStatusSectionContent(content, appendText)
	}
	return replaceStatusSectionContent(content, existing+"\n\n"+appendText)
}

// getFrontmatterStatusValue extracts the current value of status: from YAML frontmatter.
// Strips surrounding quotes if present.
func getFrontmatterStatusValue(content string) string {
	fm := extractFrontmatter(content)
	if fm == "" {
		return ""
	}
	statusLine := regexp.MustCompile(`(?m)^status:\s*(.*)$`)
	matches := statusLine.FindStringSubmatch(fm)
	if len(matches) < 2 {
		return ""
	}
	value := strings.TrimSpace(matches[1])
	// Strip surrounding quotes
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}
	return value
}

// SetSupersededBy updates the content's status to "Superseded by [ADR-N](filename)".
func SetSupersededBy(content string, link SupersedesLink) (string, error) {
	statusText := "Superseded by " + formatLink(link)

	if hasStatusSection(content) {
		return replaceStatusSectionContent(content, statusText), nil
	}
	if hasFrontmatterStatus(content) {
		return replaceFrontmatterStatus(content, "superseded by "+formatLink(link)), nil
	}
	return "", fmt.Errorf("no status section found: expected ## Status heading or status: in YAML frontmatter")
}

// SetSupersedes updates the content's status with "Supersedes [ADR-N](filename)" entries.
// Supersedes links are appended below the existing status text, not replacing it.
func SetSupersedes(content string, links []SupersedesLink) (string, error) {
	if hasStatusSection(content) {
		var lines []string
		for _, link := range links {
			lines = append(lines, "Supersedes "+formatLink(link))
		}
		return appendToStatusSectionContent(content, strings.Join(lines, "\n")), nil
	}
	if hasFrontmatterStatus(content) {
		var refs []string
		for _, link := range links {
			refs = append(refs, formatLink(link))
		}
		currentStatus := getFrontmatterStatusValue(content)
		newValue := currentStatus + ", supersedes " + strings.Join(refs, ", ")
		return replaceFrontmatterStatus(content, newValue), nil
	}
	return "", fmt.Errorf("no status section found: expected ## Status heading or status: in YAML frontmatter")
}
