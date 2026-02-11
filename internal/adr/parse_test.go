package adr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractMetadata_Nygard(t *testing.T) {
	content := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Context\n\nSome context.\n"

	meta := ExtractMetadata(content)

	assert.Equal(t, 1, meta.Number)
	assert.Equal(t, "Use Go", meta.Title)
	assert.Equal(t, "Accepted", meta.Status)
	assert.Equal(t, "2024-01-01", meta.Date)
}

func TestExtractMetadata_MADRFull(t *testing.T) {
	content := "---\nstatus: \"proposed\"\ndate: 2024-03-15\n---\n\n# 2. Use Chi Router\n\n## Context and Problem Statement\n\nNeed a router.\n"

	meta := ExtractMetadata(content)

	assert.Equal(t, 2, meta.Number)
	assert.Equal(t, "Use Chi Router", meta.Title)
	assert.Equal(t, "proposed", meta.Status)
	assert.Equal(t, "2024-03-15", meta.Date)
}

func TestExtractMetadata_MADRMinimal_NoStatus(t *testing.T) {
	content := "# 3. Use Vite\n\nDate: 2024-06-01\n\n## Context and Problem Statement\n\nNeed a bundler.\n"

	meta := ExtractMetadata(content)

	assert.Equal(t, 3, meta.Number)
	assert.Equal(t, "Use Vite", meta.Title)
	assert.Equal(t, "", meta.Status)
	assert.Equal(t, "2024-06-01", meta.Date)
}

func TestExtractMetadata_HeadingWithoutNumber(t *testing.T) {
	content := "# Some Title\n\nDate: 2024-02-01\n\n## Status\n\nProposed\n"

	meta := ExtractMetadata(content)

	assert.Equal(t, 0, meta.Number)
	assert.Equal(t, "Some Title", meta.Title)
	assert.Equal(t, "Proposed", meta.Status)
	assert.Equal(t, "2024-02-01", meta.Date)
}

func TestExtractMetadata_SupersededWithLink(t *testing.T) {
	content := "# 4. Old Approach\n\nDate: 2024-01-01\n\n## Status\n\nSuperseded by [ADR-0005](0005-foo.md)\n\n## Context\n\nOld context.\n"

	meta := ExtractMetadata(content)

	assert.Equal(t, 4, meta.Number)
	assert.Equal(t, "Old Approach", meta.Title)
	assert.Equal(t, "Superseded by [ADR-0005](0005-foo.md)", meta.Status)
	assert.Equal(t, "2024-01-01", meta.Date)
}

func TestExtractMetadata_FrontmatterDateWithQuotes(t *testing.T) {
	content := "---\nstatus: \"accepted\"\ndate: \"2024-07-01\"\n---\n\n# 5. Use TypeScript\n"

	meta := ExtractMetadata(content)

	assert.Equal(t, 5, meta.Number)
	assert.Equal(t, "Use TypeScript", meta.Title)
	assert.Equal(t, "accepted", meta.Status)
	assert.Equal(t, "2024-07-01", meta.Date)
}

func TestExtractMetadata_EmptyContent(t *testing.T) {
	meta := ExtractMetadata("")

	assert.Equal(t, 0, meta.Number)
	assert.Equal(t, "", meta.Title)
	assert.Equal(t, "", meta.Status)
	assert.Equal(t, "", meta.Date)
}

func TestExtractMetadata_DateInBody_PreferredOverFrontmatter(t *testing.T) {
	content := "---\nstatus: \"accepted\"\ndate: 2024-01-01\n---\n\n# 6. Mixed Format\n\nDate: 2024-06-15\n"

	meta := ExtractMetadata(content)

	assert.Equal(t, "2024-06-15", meta.Date)
}

func TestMetadataToADR_ValidMetadata(t *testing.T) {
	m := Metadata{Number: 1, Title: "Use Go", Status: "Accepted", Date: "2024-01-15"}

	got, err := MetadataToADR(m, 0)

	require.NoError(t, err)
	assert.Equal(t, 1, got.Number)
	assert.Equal(t, "Use Go", got.Title)
	assert.Equal(t, Accepted, got.Status)
	assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), got.Date)
}

func TestMetadataToADR_UseFallbackNumber(t *testing.T) {
	m := Metadata{Number: 0, Title: "Some Title", Status: "Proposed", Date: "2024-02-01"}

	got, err := MetadataToADR(m, 5)

	require.NoError(t, err)
	assert.Equal(t, 5, got.Number)
}

func TestMetadataToADR_InvalidStatus(t *testing.T) {
	m := Metadata{Number: 1, Title: "Bad Status", Status: "bogus", Date: "2024-01-01"}

	_, err := MetadataToADR(m, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestMetadataToADR_EmptyDate(t *testing.T) {
	m := Metadata{Number: 1, Title: "No Date", Status: "Accepted", Date: ""}

	got, err := MetadataToADR(m, 0)

	require.NoError(t, err)
	assert.True(t, got.Date.IsZero())
}

func TestMetadataToADR_SupersededByPrefix(t *testing.T) {
	m := Metadata{Number: 4, Title: "Old", Status: "Superseded by [ADR-0005](0005-foo.md)", Date: "2024-01-01"}

	got, err := MetadataToADR(m, 0)

	require.NoError(t, err)
	assert.Equal(t, Superseded, got.Status)
}

func TestMetadataToADR_MultiLineStatusSection(t *testing.T) {
	m := Metadata{
		Number: 5,
		Title:  "test5",
		Status: "Proposed\n\nSupersedes [ADR-0003](0003-test3.md)\nSupersedes [ADR-0004](0004-test4.md)",
		Date:   "2026-02-11",
	}
	got, err := MetadataToADR(m, 0)
	require.NoError(t, err)
	assert.Equal(t, Proposed, got.Status)
	assert.Equal(t, 5, got.Number)
}

func TestMetadataToADR_EmptyStatus(t *testing.T) {
	m := Metadata{Number: 1, Title: "No Status", Status: "", Date: "2024-01-01"}

	got, err := MetadataToADR(m, 0)

	require.NoError(t, err)
	assert.Equal(t, Proposed, got.Status)
}
