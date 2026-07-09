package adr_test

import (
	"testing"
	"time"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func date(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func numbers(records []adr.ADR) []int {
	out := make([]int, len(records))
	for i, r := range records {
		out[i] = r.Number
	}
	return out
}

func TestSortADRs_Number(t *testing.T) {
	recs := []adr.ADR{{Number: 3}, {Number: 1}, {Number: 2}}
	require.NoError(t, adr.SortADRs(recs, "number", false))
	assert.Equal(t, []int{1, 2, 3}, numbers(recs))

	require.NoError(t, adr.SortADRs(recs, "number", true))
	assert.Equal(t, []int{3, 2, 1}, numbers(recs))
}

func TestSortADRs_Title(t *testing.T) {
	recs := []adr.ADR{
		{Number: 1, Title: "banana"},
		{Number: 2, Title: "Apple"},
		{Number: 3, Title: "cherry"},
	}
	require.NoError(t, adr.SortADRs(recs, "title", false))
	assert.Equal(t, []int{2, 1, 3}, numbers(recs), "case-insensitive ascending")

	require.NoError(t, adr.SortADRs(recs, "title", true))
	assert.Equal(t, []int{3, 1, 2}, numbers(recs))
}

func TestSortADRs_Status_UsesLifecycleOrder(t *testing.T) {
	recs := []adr.ADR{
		{Number: 1, Status: adr.Rejected},
		{Number: 2, Status: adr.Proposed},
		{Number: 3, Status: adr.Deprecated},
		{Number: 4, Status: adr.Accepted},
		{Number: 5, Status: adr.Superseded},
	}
	require.NoError(t, adr.SortADRs(recs, "status", false))
	// Lifecycle: proposed, accepted, deprecated, superseded, rejected — NOT iota order.
	assert.Equal(t, []int{2, 4, 3, 5, 1}, numbers(recs))
	// Rejected must land after Deprecated and Superseded (unlike the enum).
	assert.Equal(t, adr.Rejected, recs[len(recs)-1].Status)
}

func TestSortADRs_Date_UndatedAlwaysLast(t *testing.T) {
	build := func() []adr.ADR {
		return []adr.ADR{
			{Number: 1, Date: date("2025-03-01")},
			{Number: 2, Date: date("")}, // undated
			{Number: 3, Date: date("2025-01-01")},
			{Number: 4, Date: date("2025-02-01")},
		}
	}

	asc := build()
	require.NoError(t, adr.SortADRs(asc, "date", false))
	assert.Equal(t, []int{3, 4, 1, 2}, numbers(asc), "ascending oldest→newest, undated last")

	desc := build()
	require.NoError(t, adr.SortADRs(desc, "date", true))
	assert.Equal(t, []int{1, 4, 3, 2}, numbers(desc), "descending newest→oldest, undated STILL last")
}

func TestSortADRs_TieBreaksByNumber_IndependentOfInputOrder(t *testing.T) {
	// Same date, deliberately out of number order on input — the explicit Number tiebreaker
	// (not reliance on caller ordering) must produce ascending numbers.
	recs := []adr.ADR{
		{Number: 5, Date: date("2025-01-01")},
		{Number: 2, Date: date("2025-01-01")},
		{Number: 9, Date: date("2025-01-01")},
	}
	require.NoError(t, adr.SortADRs(recs, "date", false))
	assert.Equal(t, []int{2, 5, 9}, numbers(recs))

	// Two undated entries also tie-break by Number.
	undated := []adr.ADR{{Number: 7}, {Number: 3}}
	require.NoError(t, adr.SortADRs(undated, "date", true))
	assert.Equal(t, []int{3, 7}, numbers(undated))
}

func TestSortADRs_UnknownField_ErrorsAndLeavesSliceUnmodified(t *testing.T) {
	recs := []adr.ADR{{Number: 3}, {Number: 1}, {Number: 2}}
	err := adr.SortADRs(recs, "bogus", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sort field")
	assert.Equal(t, []int{3, 1, 2}, numbers(recs), "slice must be untouched on error")
}
