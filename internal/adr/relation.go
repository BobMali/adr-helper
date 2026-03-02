package adr

import (
	"fmt"
	"regexp"
	"strings"
)

var relationsSectionPattern = regexp.MustCompile(`(?m)^## Relations[s]?[ \t]*$`)

// hasRelationsSection checks for a ## Relations heading.
func hasRelationsSection(content string) bool {
	return relationsSectionPattern.MatchString(content)
}

// extractRelationsSectionContent returns the text between ## Relations heading and the next ## heading (or EOF).
func extractRelationsSectionContent(content string) string {
	loc := relationsSectionPattern.FindStringIndex(content)
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

	return strings.Trim(sectionBody, "\n")
}

// insertRelationsSection inserts a new ## Relations section with the given initial line.
// For nygard format: inserts after ## Status (before the next ## heading).
// For MADR frontmatter format: inserts before the first ## heading after the title.
func insertRelationsSection(content, initialLine string) (string, error) {
	section := "## Relations\n\n" + initialLine

	if hasStatusSection(content) {
		// Find ## Status
		loc := statusSectionPattern.FindStringIndex(content)
		afterStatus := loc[1]
		rest := content[afterStatus:]

		// Find the next ## heading after Status
		nextHeadingPattern := regexp.MustCompile(`(?m)\n\n## `)
		nextLoc := nextHeadingPattern.FindStringIndex(rest)

		if nextLoc != nil {
			// Insert between Status section and next heading
			insertPoint := afterStatus + nextLoc[0]
			return content[:insertPoint] + "\n\n" + section + content[insertPoint:], nil
		}
		// Status is the last section — append at end
		return strings.TrimRight(content, "\n") + "\n\n" + section + "\n", nil
	}

	if hasFrontmatterStatus(content) {
		// Find the first ## heading after the title line (# ...)
		titlePattern := regexp.MustCompile(`(?m)^# .+$`)
		titleLoc := titlePattern.FindStringIndex(content)
		if titleLoc == nil {
			return "", fmt.Errorf("no recognized ADR format: missing title")
		}
		afterTitle := content[titleLoc[1]:]
		headingPattern := regexp.MustCompile(`(?m)\n\n## `)
		headingLoc := headingPattern.FindStringIndex(afterTitle)
		if headingLoc != nil {
			insertPoint := titleLoc[1] + headingLoc[0]
			return content[:insertPoint] + "\n\n" + section + content[insertPoint:], nil
		}
		// No ## heading after title — append at end
		return strings.TrimRight(content, "\n") + "\n\n" + section + "\n", nil
	}

	return "", fmt.Errorf("no recognized ADR format: expected ## Status heading or status: in YAML frontmatter")
}

// appendToRelationsSection appends a line to the existing ## Relations section.
func appendToRelationsSection(content, line string) string {
	existing := extractRelationsSectionContent(content)
	newContent := existing + "\n" + line

	// Replace the Relations section content
	loc := relationsSectionPattern.FindStringIndex(content)
	if loc == nil {
		return content
	}

	afterHeading := loc[1]
	rest := content[afterHeading:]

	nextHeadingPattern := regexp.MustCompile(`(?m)\n\n## `)
	nextLoc := nextHeadingPattern.FindStringIndex(rest)

	if nextLoc != nil {
		return content[:afterHeading] + "\n\n" + newContent + content[afterHeading+nextLoc[0]:]
	}
	return content[:afterHeading] + "\n\n" + newContent + "\n"
}

// AddRelation adds a "Relates to [ADR-NNNN](filename)" line to the ## Relations section.
// If no ## Relations section exists, one is inserted.
// Idempotent: skips if the link already exists.
func AddRelation(content string, link ADRLink) (string, error) {
	line := "Relates to " + formatADRLink(link) + "  "

	if hasRelationsSection(content) {
		existing := extractRelationsSectionContent(content)
		if strings.Contains(existing, formatADRLink(link)) {
			return content, nil
		}
		return appendToRelationsSection(content, line), nil
	}

	return insertRelationsSection(content, line)
}
