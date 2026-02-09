package adr_test

import (
	"testing"

	"github.com/malek/adr-helper/internal/adr"
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

func TestValidTemplateNames(t *testing.T) {
	names := adr.ValidTemplateNames()
	assert.Contains(t, names, "nygard")
	assert.Contains(t, names, "madr-minimal")
	assert.Contains(t, names, "madr-full")
	assert.Len(t, names, 3)
}
