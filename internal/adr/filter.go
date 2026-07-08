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

// FilterByMetaField filters records by a metadata field's values, matching
// case-insensitively against each record's Meta[key]. With matchAll=false (Any/union)
// a record is kept if its field shares at least one value; with matchAll=true
// (All/intersection) it is kept only if its field contains every requested value.
// An empty values slice is a no-op (returns the input unchanged). Records lacking the
// field never match a non-empty filter.
func FilterByMetaField(records []ADR, key string, values []string, matchAll bool) []ADR {
	// Normalize requested values (lowercased, empties dropped).
	wanted := make([]string, 0, len(values))
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			wanted = append(wanted, strings.ToLower(v))
		}
	}
	if len(wanted) == 0 {
		return records
	}
	if records == nil {
		return nil
	}

	result := []ADR{}
	for _, r := range records {
		present := make(map[string]bool, len(r.Meta[key]))
		for _, v := range r.Meta[key] {
			present[strings.ToLower(strings.TrimSpace(v))] = true
		}
		if metaMatches(present, wanted, matchAll) {
			result = append(result, r)
		}
	}
	return result
}

// metaMatches reports whether the present value set satisfies the wanted values under
// the given mode. Any: at least one wanted value present. All: every wanted value present.
func metaMatches(present map[string]bool, wanted []string, matchAll bool) bool {
	if matchAll {
		for _, w := range wanted {
			if !present[w] {
				return false
			}
		}
		return true
	}
	for _, w := range wanted {
		if present[w] {
			return true
		}
	}
	return false
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
