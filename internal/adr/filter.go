package adr

import (
	"strconv"
	"strings"
)

// FilterByQuery filters records by a search query. It matches case-insensitively
// against the title, and for all-digit queries it also matches an exact ADR number.
// An empty or whitespace-only query returns the input unchanged.
func FilterByQuery(records []ADR, query string) []ADR {
	query = strings.TrimSpace(query)
	if query == "" {
		return records
	}
	if records == nil {
		return nil
	}

	lowerQuery := strings.ToLower(query)
	numQuery, isNum := 0, isAllDigits(query)
	if isNum {
		numQuery, _ = strconv.Atoi(query)
	}

	result := []ADR{}
	for _, r := range records {
		if strings.Contains(strings.ToLower(r.Title), lowerQuery) {
			result = append(result, r)
			continue
		}
		if isNum && r.Number == numQuery {
			result = append(result, r)
		}
	}
	return result
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
