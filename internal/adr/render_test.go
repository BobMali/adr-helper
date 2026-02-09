package adr_test

import (
	"strings"
	"testing"
	"time"

	"github.com/malek/adr-helper/internal/adr"
	"github.com/stretchr/testify/assert"
)

func TestRenderTemplate_NygardFormat(t *testing.T) {
	record := &adr.ADR{Number: 1, Title: "Use Go", Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)}
	content := "# Title\n\nDate:\n\n## Status\n\nProposed\n"

	result := adr.RenderTemplate(content, record)

	assert.Contains(t, result, "# 1. Use Go")
	assert.Contains(t, result, "Date: 2024-01-15")
}

func TestRenderTemplate_MADRMinimalFormat(t *testing.T) {
	record := &adr.ADR{Number: 3, Title: "Use Vue", Date: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)}
	content := "# {short title}\n\n## Context and Problem Statement\n"

	result := adr.RenderTemplate(content, record)

	assert.Contains(t, result, "# 3. Use Vue")
	assert.Contains(t, result, "## Context and Problem Statement")
}

func TestRenderTemplate_MADRFullFrontmatter(t *testing.T) {
	record := &adr.ADR{Number: 2, Title: "Use Chi", Date: time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)}
	content := "---\nstatus: \"{proposed}\"\ndate: {YYYY-MM-DD when the decision was last updated}\n---\n\n# {short title}\n\n## Context\n"

	result := adr.RenderTemplate(content, record)

	assert.Contains(t, result, "# 2. Use Chi")
	assert.Contains(t, result, "date: 2024-03-20")
}

func TestRenderTemplate_OnlyReplacesFirstHeading(t *testing.T) {
	record := &adr.ADR{Number: 1, Title: "Test", Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	content := "# Title\n\n# Second Heading\n"

	result := adr.RenderTemplate(content, record)

	assert.Equal(t, 1, strings.Count(result, "# 1. Test"))
	assert.Contains(t, result, "# Second Heading")
}

func TestRenderTemplate_NygardSetsDefaultStatus(t *testing.T) {
	record := &adr.ADR{Number: 1, Title: "Use Go", Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Status: adr.Proposed}
	content := "# Title\n\nDate:\n\n## Status\n\nWhat is the status, such as proposed, accepted, rejected, deprecated, superseded, etc.?\n\n## Context\n\nSome context.\n"

	result := adr.RenderTemplate(content, record)

	assert.Contains(t, result, "## Status\n\nProposed\n\n## Context")
	assert.NotContains(t, result, "What is the status")
}

func TestRenderTemplate_MADRFullSetsDefaultStatus(t *testing.T) {
	record := &adr.ADR{Number: 2, Title: "Use Chi", Date: time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC), Status: adr.Proposed}
	content := "---\nstatus: \"{proposed | rejected | accepted | deprecated | … | superseded by ADR-0123\"\ndate: {YYYY-MM-DD when the decision was last updated}\n---\n\n# {short title}\n\n## Context\n"

	result := adr.RenderTemplate(content, record)

	assert.Contains(t, result, "status: \"proposed\"")
	assert.NotContains(t, result, "proposed | rejected")
}

func TestRenderTemplate_MADRMinimalNoStatusUnchanged(t *testing.T) {
	record := &adr.ADR{Number: 3, Title: "Use Vue", Date: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), Status: adr.Proposed}
	content := "# {short title}\n\n## Context and Problem Statement\n\nSome problem.\n"

	result := adr.RenderTemplate(content, record)

	// No status section or frontmatter status — content should not have "Proposed" injected
	assert.NotContains(t, result, "Proposed")
	assert.NotContains(t, result, "proposed")
	assert.Contains(t, result, "## Context and Problem Statement")
}

func TestRenderTemplate_PreservesBody(t *testing.T) {
	record := &adr.ADR{Number: 1, Title: "Test", Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	content := "# Title\n\nDate:\n\n## Status\n\n## Context\n\n## Decision\n\n## Consequences\n"

	result := adr.RenderTemplate(content, record)

	assert.Contains(t, result, "## Status")
	assert.Contains(t, result, "## Context")
	assert.Contains(t, result, "## Decision")
	assert.Contains(t, result, "## Consequences")
}
