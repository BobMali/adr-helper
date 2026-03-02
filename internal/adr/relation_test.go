package adr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddRelation_Nygard_InsertsNewSection(t *testing.T) {
	content := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Context\n\nSome context.\n"
	link := ADRLink{Number: 3, Filename: "0003-use-chi.md"}

	result, err := AddRelation(content, link)
	require.NoError(t, err)
	assert.Contains(t, result, "## Relations\n\nRelates to [ADR-0003](0003-use-chi.md)  \n\n## Context")
	assert.Contains(t, result, "## Status\n\nAccepted\n\n## Relations")
}

func TestAddRelation_Nygard_AppendsToExisting(t *testing.T) {
	content := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Relations\n\nRelates to [ADR-0002](0002-use-rust.md)  \n\n## Context\n\nSome context.\n"
	link := ADRLink{Number: 3, Filename: "0003-use-chi.md"}

	result, err := AddRelation(content, link)
	require.NoError(t, err)
	assert.Contains(t, result, "Relates to [ADR-0002](0002-use-rust.md)  \nRelates to [ADR-0003](0003-use-chi.md)  \n\n## Context")
}

func TestAddRelation_Nygard_Idempotent(t *testing.T) {
	content := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Relations\n\nRelates to [ADR-0003](0003-use-chi.md)  \n\n## Context\n\nSome context.\n"
	link := ADRLink{Number: 3, Filename: "0003-use-chi.md"}

	result, err := AddRelation(content, link)
	require.NoError(t, err)
	assert.Equal(t, content, result)
}

func TestAddRelation_Nygard_PreservesAllSections(t *testing.T) {
	content := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Context\n\nContext text.\n\n## Decision\n\nDecision text.\n\n## Consequences\n\nConsequences text.\n"
	link := ADRLink{Number: 3, Filename: "0003-use-chi.md"}

	result, err := AddRelation(content, link)
	require.NoError(t, err)
	assert.Contains(t, result, "## Status\n\nAccepted\n\n## Relations")
	assert.Contains(t, result, "## Context\n\nContext text.")
	assert.Contains(t, result, "## Decision\n\nDecision text.")
	assert.Contains(t, result, "## Consequences\n\nConsequences text.")
}

func TestAddRelation_MADRFrontmatter_InsertsBeforeFirstHeading(t *testing.T) {
	content := "---\nstatus: \"accepted\"\ndate: 2024-01-01\n---\n\n# 1. Use Go\n\n## Context and Problem Statement\n\nSome context.\n"
	link := ADRLink{Number: 3, Filename: "0003-use-chi.md"}

	result, err := AddRelation(content, link)
	require.NoError(t, err)
	assert.Contains(t, result, "## Relations\n\nRelates to [ADR-0003](0003-use-chi.md)  \n\n## Context and Problem Statement")
}

func TestAddRelation_MADRFrontmatter_AppendsToExisting(t *testing.T) {
	content := "---\nstatus: \"accepted\"\ndate: 2024-01-01\n---\n\n# 1. Use Go\n\n## Relations\n\nRelates to [ADR-0002](0002-use-rust.md)  \n\n## Context and Problem Statement\n\nSome context.\n"
	link := ADRLink{Number: 3, Filename: "0003-use-chi.md"}

	result, err := AddRelation(content, link)
	require.NoError(t, err)
	assert.Contains(t, result, "Relates to [ADR-0002](0002-use-rust.md)  \nRelates to [ADR-0003](0003-use-chi.md)  \n\n## Context and Problem Statement")
}

func TestAddRelation_NoRecognizedFormat_ReturnsError(t *testing.T) {
	content := "Just some random text without any structure.\n"
	link := ADRLink{Number: 3, Filename: "0003-use-chi.md"}

	_, err := AddRelation(content, link)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recognized ADR format")
}

func TestAddRelation_StatusIsLastSection(t *testing.T) {
	content := "# 1. Use Go\n\n## Status\n\nAccepted\n"
	link := ADRLink{Number: 3, Filename: "0003-use-chi.md"}

	result, err := AddRelation(content, link)
	require.NoError(t, err)
	assert.Contains(t, result, "## Status\n\nAccepted\n\n## Relations\n\nRelates to [ADR-0003](0003-use-chi.md)  \n")
}
