package adr_test

import (
	"testing"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateStatus_Nygard(t *testing.T) {
	content := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nProposed\n\n## Context\n\nSome context.\n"

	result, err := adr.UpdateStatus(content, "accepted")
	require.NoError(t, err)
	assert.Contains(t, result, "## Status\n\nAccepted\n\n## Context")
	assert.NotContains(t, result, "Proposed")
}

func TestUpdateStatus_NygardCaseInsensitive(t *testing.T) {
	content := "# 1. Use Go\n\n## Status\n\nProposed\n\n## Context\n\nSome context.\n"

	result, err := adr.UpdateStatus(content, "REJECTED")
	require.NoError(t, err)
	assert.Contains(t, result, "## Status\n\nRejected\n\n## Context")
}

func TestUpdateStatus_Frontmatter(t *testing.T) {
	content := "---\nstatus: \"proposed\"\ndate: 2024-01-01\n---\n\n# 1. Use Go\n\n## Context and Problem Statement\n\nSome context.\n"

	result, err := adr.UpdateStatus(content, "accepted")
	require.NoError(t, err)
	assert.Contains(t, result, "status: \"accepted\"")
	assert.Contains(t, result, "Some context.")
}

func TestUpdateStatus_NoStatusSection_ReturnsError(t *testing.T) {
	content := "# 1. Use Go\n\n## Context\n\nSome context.\n"

	_, err := adr.UpdateStatus(content, "accepted")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no status section found")
}

func TestUpdateStatus_PreservesOtherSections(t *testing.T) {
	content := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nProposed\n\n## Context\n\nContext text.\n\n## Decision\n\nDecision text.\n\n## Consequences\n\nConsequences text.\n"

	result, err := adr.UpdateStatus(content, "deprecated")
	require.NoError(t, err)
	assert.Contains(t, result, "## Status\n\nDeprecated\n\n## Context")
	assert.Contains(t, result, "## Context\n\nContext text.")
	assert.Contains(t, result, "## Decision\n\nDecision text.")
	assert.Contains(t, result, "## Consequences\n\nConsequences text.")
}

func TestUpdateStatus_NygardPreservesSupersedes(t *testing.T) {
	content := "# 2. Better\n\nDate: 2024-01-01\n\n## Status\n\nProposed\n\nSupersedes [ADR-0001](0001-old.md)\n\n## Context\n\nSome context.\n"

	result, err := adr.UpdateStatus(content, "accepted")
	require.NoError(t, err)
	assert.Contains(t, result, "Accepted")
	assert.Contains(t, result, "Supersedes [ADR-0001](0001-old.md)")
	assert.NotContains(t, result, "Proposed")
}

func TestUpdateStatus_NygardPreservesMultipleSupersedes(t *testing.T) {
	content := "# 6. Better\n\nDate: 2024-01-01\n\n## Status\n\nProposed\n\nSupersedes [ADR-0001](0001-first.md)\nSupersedes [ADR-0003](0003-third.md)\n\n## Context\n\nSome context.\n"

	result, err := adr.UpdateStatus(content, "accepted")
	require.NoError(t, err)
	assert.Contains(t, result, "Accepted")
	assert.Contains(t, result, "Supersedes [ADR-0001](0001-first.md)")
	assert.Contains(t, result, "Supersedes [ADR-0003](0003-third.md)")
	assert.NotContains(t, result, "Proposed")
}

func TestUpdateStatus_FrontmatterPreservesSupersedes(t *testing.T) {
	content := "---\nstatus: \"proposed, supersedes [ADR-0001](0001-first.md)\"\ndate: 2024-01-01\n---\n\n# 2. Better\n\n## Context and Problem Statement\n\nSome context.\n"

	result, err := adr.UpdateStatus(content, "accepted")
	require.NoError(t, err)
	assert.Contains(t, result, `status: "accepted, supersedes [ADR-0001](0001-first.md)"`)
}

func TestSetSupersededBy_NygardFormat(t *testing.T) {
	content := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Context\n\nSome context.\n"
	link := adr.SupersedesLink{Number: 6, Filename: "0006-new.md"}

	result, err := adr.SetSupersededBy(content, link)
	require.NoError(t, err)
	assert.Contains(t, result, "## Status\n\nSuperseded by [ADR-0006](0006-new.md)\n\n## Context")
	assert.Contains(t, result, "Some context.")
}

func TestSetSupersededBy_NygardWithPlaceholder(t *testing.T) {
	content := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nWhat is the status, such as proposed, accepted, rejected, deprecated, superseded, etc.?\n\n## Context\n\nSome context.\n"
	link := adr.SupersedesLink{Number: 6, Filename: "0006-new.md"}

	result, err := adr.SetSupersededBy(content, link)
	require.NoError(t, err)
	assert.Contains(t, result, "## Status\n\nSuperseded by [ADR-0006](0006-new.md)\n\n## Context")
}

func TestSetSupersededBy_NygardStatusAtEndOfFile(t *testing.T) {
	content := "# 1. Use Go\n\n## Status\n\nAccepted\n"
	link := adr.SupersedesLink{Number: 6, Filename: "0006-new.md"}

	result, err := adr.SetSupersededBy(content, link)
	require.NoError(t, err)
	assert.Contains(t, result, "## Status\n\nSuperseded by [ADR-0006](0006-new.md)\n")
	assert.NotContains(t, result, "Accepted")
}

func TestSetSupersededBy_NygardPreservesOtherSections(t *testing.T) {
	content := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Context\n\nContext text.\n\n## Decision\n\nDecision text.\n\n## Consequences\n\nConsequences text.\n"
	link := adr.SupersedesLink{Number: 6, Filename: "0006-new.md"}

	result, err := adr.SetSupersededBy(content, link)
	require.NoError(t, err)
	assert.Contains(t, result, "## Context\n\nContext text.")
	assert.Contains(t, result, "## Decision\n\nDecision text.")
	assert.Contains(t, result, "## Consequences\n\nConsequences text.")
}

func TestSetSupersededBy_MADRFullFrontmatter(t *testing.T) {
	content := "---\nstatus: \"accepted\"\ndate: 2024-01-01\n---\n\n# 1. Use Go\n\n## Context and Problem Statement\n\nSome context.\n"
	link := adr.SupersedesLink{Number: 6, Filename: "0006-new.md"}

	result, err := adr.SetSupersededBy(content, link)
	require.NoError(t, err)
	assert.Contains(t, result, "status: \"superseded by [ADR-0006](0006-new.md)\"")
	assert.Contains(t, result, "Some context.")
}

func TestSetSupersededBy_MADRFrontmatterScopedToBlock(t *testing.T) {
	content := "---\nstatus: \"accepted\"\ndate: 2024-01-01\n---\n\n# 1. Use Go\n\nstatus: this should not change\n"
	link := adr.SupersedesLink{Number: 6, Filename: "0006-new.md"}

	result, err := adr.SetSupersededBy(content, link)
	require.NoError(t, err)
	assert.Contains(t, result, "status: \"superseded by [ADR-0006](0006-new.md)\"")
	assert.Contains(t, result, "status: this should not change")
}

func TestSetSupersededBy_AlreadySuperseded(t *testing.T) {
	content := "# 1. Use Go\n\n## Status\n\nSuperseded by [ADR-0003](0003-old.md)\n\n## Context\n\nSome context.\n"
	link := adr.SupersedesLink{Number: 6, Filename: "0006-new.md"}

	result, err := adr.SetSupersededBy(content, link)
	require.NoError(t, err)
	assert.Contains(t, result, "Superseded by [ADR-0006](0006-new.md)")
	assert.NotContains(t, result, "ADR-0003")
}

func TestSetSupersededBy_NoStatusSection_ReturnsError(t *testing.T) {
	content := "# 1. Use Go\n\n## Context\n\nSome context.\n"
	link := adr.SupersedesLink{Number: 6, Filename: "0006-new.md"}

	_, err := adr.SetSupersededBy(content, link)
	assert.Error(t, err)
}

func TestSetSupersedes_NygardSingle(t *testing.T) {
	content := "# 2. Better\n\nDate: 2024-01-01\n\n## Status\n\nProposed\n\n## Context\n\nSome context.\n"
	links := []adr.SupersedesLink{{Number: 1, Filename: "0001-old.md"}}

	result, err := adr.SetSupersedes(content, links)
	require.NoError(t, err)
	assert.Contains(t, result, "## Status\n\nProposed\n\nSupersedes [ADR-0001](0001-old.md)\n\n## Context")
}

func TestSetSupersedes_NygardMultiple(t *testing.T) {
	content := "# 6. Better\n\nDate: 2024-01-01\n\n## Status\n\nProposed\n\n## Context\n\nSome context.\n"
	links := []adr.SupersedesLink{
		{Number: 1, Filename: "0001-first.md"},
		{Number: 5, Filename: "0005-fifth.md"},
	}

	result, err := adr.SetSupersedes(content, links)
	require.NoError(t, err)
	assert.Contains(t, result, "## Status\n\nProposed\n\nSupersedes [ADR-0001](0001-first.md)\nSupersedes [ADR-0005](0005-fifth.md)\n\n## Context")
}

func TestSetSupersedes_MADRFullFrontmatter(t *testing.T) {
	content := "---\nstatus: \"proposed\"\ndate: 2024-01-01\n---\n\n# 6. Better\n\n## Context and Problem Statement\n\nSome context.\n"
	links := []adr.SupersedesLink{
		{Number: 1, Filename: "0001-first.md"},
		{Number: 5, Filename: "0005-fifth.md"},
	}

	result, err := adr.SetSupersedes(content, links)
	require.NoError(t, err)
	assert.Contains(t, result, "status: \"proposed, supersedes [ADR-0001](0001-first.md), [ADR-0005](0005-fifth.md)\"")
}

func TestSetSupersedes_NoStatusSection_ReturnsError(t *testing.T) {
	content := "# 2. Better\n\n## Context\n\nSome context.\n"
	links := []adr.SupersedesLink{{Number: 1, Filename: "0001-old.md"}}

	_, err := adr.SetSupersedes(content, links)
	assert.Error(t, err)
}

func TestSetSupersedes_PreservesRestOfContent(t *testing.T) {
	content := "# 6. Better\n\nDate: 2024-01-01\n\n## Status\n\nProposed\n\n## Context\n\nContext text.\n\n## Decision\n\nDecision text.\n"
	links := []adr.SupersedesLink{{Number: 1, Filename: "0001-old.md"}}

	result, err := adr.SetSupersedes(content, links)
	require.NoError(t, err)
	assert.Contains(t, result, "Proposed\n\nSupersedes [ADR-0001](0001-old.md)")
	assert.Contains(t, result, "## Context\n\nContext text.")
	assert.Contains(t, result, "## Decision\n\nDecision text.")
}

func TestSetSupersedes_EmptyStatusSection(t *testing.T) {
	content := "# 2. Better\n\n## Status\n\n## Context\n\nSome context.\n"
	links := []adr.SupersedesLink{{Number: 1, Filename: "0001-old.md"}}

	result, err := adr.SetSupersedes(content, links)
	require.NoError(t, err)
	assert.Contains(t, result, "## Status\n\nSupersedes [ADR-0001](0001-old.md)\n\n## Context")
}
