package adr

import (
	"fmt"
	"sort"
	"strings"
)

// SortADRs sorts records IN PLACE by field ("number", "title", "status", or "date"),
// ascending unless desc is true. Unlike FilterByQuery/FilterByMetaField, it does not
// return a new slice. It returns an error for an unknown field (and leaves records
// untouched in that case). Ties break by ascending Number, so the result is deterministic
// regardless of the input order.
func SortADRs(records []ADR, field string, desc bool) error {
	less, err := sortLess(records, field, desc)
	if err != nil {
		return err
	}
	sort.SliceStable(records, less)
	return nil
}

// sortLess builds the comparator for the given field, validating the field name BEFORE any
// sorting happens (sort.Interface.Less can't surface an error, and validating inside it
// would risk a half-mutated slice on an unknown field).
func sortLess(records []ADR, field string, desc bool) (func(i, j int) bool, error) {
	switch field {
	case "number":
		// Number is unique, so no tiebreaker is needed.
		return func(i, j int) bool {
			return applyDir(records[i].Number < records[j].Number, records[i].Number > records[j].Number, desc)
		}, nil
	case "title":
		// Case-insensitive byte comparison. This intentionally differs from the web's
		// locale-aware localeCompare — matching that would need golang.org/x/text/collate,
		// which the project deliberately avoids.
		return func(i, j int) bool {
			a, b := strings.ToLower(records[i].Title), strings.ToLower(records[j].Title)
			if a == b {
				return records[i].Number < records[j].Number
			}
			return applyDir(a < b, a > b, desc)
		}, nil
	case "status":
		return func(i, j int) bool {
			a, b := records[i].Status.LifecycleOrder(), records[j].Status.LifecycleOrder()
			if a == b {
				return records[i].Number < records[j].Number
			}
			return applyDir(a < b, a > b, desc)
		}, nil
	case "date":
		return func(i, j int) bool {
			return dateLess(records[i], records[j], desc)
		}, nil
	default:
		return nil, fmt.Errorf("invalid sort field %q: expected number, title, status, or date", field)
	}
}

// dateLess orders by Date. Undated ADRs (zero time.Time) always sort last, independent of
// direction; dated ADRs flip with desc. Equal dates break by ascending Number.
func dateLess(a, b ADR, desc bool) bool {
	az, bz := a.Date.IsZero(), b.Date.IsZero()
	if az || bz {
		if az && bz {
			return a.Number < b.Number
		}
		return bz // b undated → a comes first; a undated → a comes last
	}
	if a.Date.Equal(b.Date) {
		return a.Number < b.Number
	}
	return applyDir(a.Date.Before(b.Date), a.Date.After(b.Date), desc)
}

// applyDir returns asc for ascending order and desc-flipped otherwise.
func applyDir(asc, gt, desc bool) bool {
	if desc {
		return gt
	}
	return asc
}
