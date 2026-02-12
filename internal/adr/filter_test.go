package adr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterByQuery_EmptyQueryReturnsAll(t *testing.T) {
	records := []ADR{
		{Number: 1, Title: "Use Go"},
		{Number: 2, Title: "Use Chi Router"},
	}
	result := FilterByQuery(records, "")
	assert.Equal(t, records, result)
}

func TestFilterByQuery_WhitespaceQueryReturnsAll(t *testing.T) {
	records := []ADR{
		{Number: 1, Title: "Use Go"},
	}
	result := FilterByQuery(records, "   ")
	assert.Equal(t, records, result)
}

func TestFilterByQuery_CaseInsensitiveTitleMatch(t *testing.T) {
	records := []ADR{
		{Number: 1, Title: "Use Go"},
		{Number: 2, Title: "Use Chi Router"},
	}
	result := FilterByQuery(records, "chi")
	assert.Len(t, result, 1)
	assert.Equal(t, "Use Chi Router", result[0].Title)
}

func TestFilterByQuery_MultipleTitleMatches(t *testing.T) {
	records := []ADR{
		{Number: 1, Title: "Use Go for CLI"},
		{Number: 2, Title: "Use Chi Router"},
		{Number: 3, Title: "Go modules layout"},
	}
	result := FilterByQuery(records, "go")
	assert.Len(t, result, 2)
	assert.Equal(t, "Use Go for CLI", result[0].Title)
	assert.Equal(t, "Go modules layout", result[1].Title)
}

func TestFilterByQuery_ExactNumberMatch(t *testing.T) {
	records := []ADR{
		{Number: 12, Title: "Use PostgreSQL"},
		{Number: 123, Title: "Use Redis"},
		{Number: 312, Title: "Use Kafka"},
	}
	result := FilterByQuery(records, "12")
	assert.Len(t, result, 1)
	assert.Equal(t, 12, result[0].Number)
}

func TestFilterByQuery_MixedAlphanumericMatchesTitleOnly(t *testing.T) {
	records := []ADR{
		{Number: 12, Title: "Something"},
		{Number: 2, Title: "Rule 12a applies"},
	}
	result := FilterByQuery(records, "12a")
	assert.Len(t, result, 1)
	assert.Equal(t, "Rule 12a applies", result[0].Title)
}

func TestFilterByQuery_NoMatchesReturnsEmptySlice(t *testing.T) {
	records := []ADR{
		{Number: 1, Title: "Use Go"},
	}
	result := FilterByQuery(records, "zzz")
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestFilterByQuery_DoesNotMutateInput(t *testing.T) {
	records := []ADR{
		{Number: 1, Title: "Use Go"},
		{Number: 2, Title: "Use Chi"},
	}
	original := make([]ADR, len(records))
	copy(original, records)

	FilterByQuery(records, "chi")

	assert.Equal(t, original, records)
}

func TestFilterByQuery_NilInputReturnsNil(t *testing.T) {
	result := FilterByQuery(nil, "anything")
	assert.Nil(t, result)
}

func TestFilterByQuery_EmptyInputReturnsEmpty(t *testing.T) {
	result := FilterByQuery([]ADR{}, "anything")
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestFilterByQuery_NoDuplicateWhenBothTitleAndNumberMatch(t *testing.T) {
	records := []ADR{
		{Number: 12, Title: "ADR 12 about caching"},
	}
	result := FilterByQuery(records, "12")
	assert.Len(t, result, 1)
}
