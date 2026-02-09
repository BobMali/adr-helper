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
)

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
