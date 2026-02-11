package adr_test

import (
	"testing"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlugify_BasicTitle(t *testing.T) {
	slug, err := adr.Slugify("Use Go for CLI")
	require.NoError(t, err)
	assert.Equal(t, "use-go-for-cli", slug)
}

func TestSlugify_SpecialCharacters(t *testing.T) {
	slug, err := adr.Slugify("Use Go! & TypeScript?")
	require.NoError(t, err)
	assert.Equal(t, "use-go-typescript", slug)
}

func TestSlugify_CollapseHyphens(t *testing.T) {
	slug, err := adr.Slugify("Use---Go")
	require.NoError(t, err)
	assert.Equal(t, "use-go", slug)
}

func TestSlugify_LeadingTrailingSpaces(t *testing.T) {
	slug, err := adr.Slugify("  foo  ")
	require.NoError(t, err)
	assert.Equal(t, "foo", slug)
}

func TestSlugify_Numbers(t *testing.T) {
	slug, err := adr.Slugify("ADR 123 Test")
	require.NoError(t, err)
	assert.Equal(t, "adr-123-test", slug)
}

func TestSlugify_PreservesHyphens(t *testing.T) {
	slug, err := adr.Slugify("Use-Go")
	require.NoError(t, err)
	assert.Equal(t, "use-go", slug)
}

func TestSlugify_EmptyString_ReturnsError(t *testing.T) {
	_, err := adr.Slugify("")
	assert.Error(t, err)
}

func TestSlugify_AllSpecialChars_ReturnsError(t *testing.T) {
	_, err := adr.Slugify("!!!")
	assert.Error(t, err)
}
