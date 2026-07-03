package adr_test

import (
	"testing"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateContent_Nygard(t *testing.T) {
	content, err := adr.TemplateContent("nygard")
	require.NoError(t, err)

	assert.Contains(t, content, "## Status")
	assert.Contains(t, content, "## Context")
	assert.Contains(t, content, "## Decision")
	assert.Contains(t, content, "## Consequences")
}

func TestTemplateContent_MADRMinimal(t *testing.T) {
	content, err := adr.TemplateContent("madr-minimal")
	require.NoError(t, err)

	assert.Contains(t, content, "## Context and Problem Statement")
	assert.Contains(t, content, "## Considered Options")
	assert.Contains(t, content, "## Decision Outcome")
}

func TestTemplateContent_MADRFull(t *testing.T) {
	content, err := adr.TemplateContent("madr-full")
	require.NoError(t, err)

	assert.Contains(t, content, "## Decision Drivers")
	assert.Contains(t, content, "## Pros and Cons of the Options")
	assert.Contains(t, content, "## More Information")
}

func TestTemplateContent_InvalidName(t *testing.T) {
	_, err := adr.TemplateContent("invalid")
	assert.Error(t, err)
}

// --- TemplateSections ---

func TestTemplateSections_NygardHas3Sections(t *testing.T) {
	sections, err := adr.TemplateSections("nygard")
	require.NoError(t, err)
	assert.Len(t, sections, 3)
}

func TestTemplateSections_MADRMinimalHas4Sections(t *testing.T) {
	sections, err := adr.TemplateSections("madr-minimal")
	require.NoError(t, err)
	assert.Len(t, sections, 4)
}

func TestTemplateSections_MADRFullHas8Sections(t *testing.T) {
	sections, err := adr.TemplateSections("madr-full")
	require.NoError(t, err)
	assert.Len(t, sections, 8)
}

func TestTemplateSections_NoStatusSection(t *testing.T) {
	for _, name := range adr.ValidTemplateNames() {
		sections, err := adr.TemplateSections(name)
		require.NoError(t, err)
		for _, s := range sections {
			assert.NotEqual(t, "status", s.Key, "template %s should not have a Status section", name)
			assert.NotEqual(t, "Status", s.Heading, "template %s should not have a Status heading", name)
		}
	}
}

func TestTemplateSections_UniqueKeys(t *testing.T) {
	for _, name := range adr.ValidTemplateNames() {
		sections, err := adr.TemplateSections(name)
		require.NoError(t, err)
		seen := make(map[string]bool)
		for _, s := range sections {
			assert.False(t, seen[s.Key], "duplicate key %q in template %s", s.Key, name)
			seen[s.Key] = true
		}
	}
}

func TestTemplateSections_AllHaveNonEmptyFields(t *testing.T) {
	for _, name := range adr.ValidTemplateNames() {
		sections, err := adr.TemplateSections(name)
		require.NoError(t, err)
		for _, s := range sections {
			assert.NotEmpty(t, s.Key)
			assert.NotEmpty(t, s.Heading)
			assert.NotEmpty(t, s.Kind)
			assert.NotEmpty(t, s.Placeholder)
		}
	}
}

func TestTemplateSections_InvalidTemplate(t *testing.T) {
	_, err := adr.TemplateSections("invalid")
	assert.Error(t, err)
}

func TestValidTemplateNames(t *testing.T) {
	names := adr.ValidTemplateNames()
	assert.Contains(t, names, "nygard")
	assert.Contains(t, names, "nygard-scoped")
	assert.Contains(t, names, "madr-minimal")
	assert.Contains(t, names, "madr-full")
	assert.Len(t, names, 4)
}

func TestTemplateContent_NygardScoped_HasScopeLine(t *testing.T) {
	content, err := adr.TemplateContent("nygard-scoped")
	require.NoError(t, err)
	assert.Contains(t, content, "Scope:")
	assert.Contains(t, content, "## Decision")
}

func TestValidTemplateNames_IncludesNygardScoped(t *testing.T) {
	assert.Contains(t, adr.ValidTemplateNames(), "nygard-scoped")
}

func TestTemplateSections_NygardScoped_ScopeFirstAndVocabulary(t *testing.T) {
	sections, err := adr.TemplateSections("nygard-scoped")
	require.NoError(t, err)
	require.Len(t, sections, 4)

	scope := sections[0]
	assert.Equal(t, "scope", scope.Key)
	assert.Equal(t, "Scope", scope.Heading)
	assert.Equal(t, "meta", scope.Kind)
	assert.False(t, scope.Optional, "scope is required in the form")
	assert.True(t, scope.Vocabulary)

	assert.Equal(t, "context", sections[1].Key)
	assert.Equal(t, "decision", sections[2].Key)
	assert.Equal(t, "consequences", sections[3].Key)
}
