package adr

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
