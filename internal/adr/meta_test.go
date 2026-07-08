package adr_test

import (
	"testing"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/stretchr/testify/assert"
)

func TestExtractMetaFields_ScopeTitleBlock(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "single value",
			content: "# 1. Title\n\nScope: backend\n\n## Status\n\nAccepted\n",
			want:    []string{"backend"},
		},
		{
			name:    "multiple comma-separated values, trimmed",
			content: "# 1. Title\n\nScope: backend,  api , web\n",
			want:    []string{"backend", "api", "web"},
		},
		{
			name:    "case-insensitive label match",
			content: "# 1. Title\n\nscope: backend\n",
			want:    []string{"backend"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := adr.ExtractMetaFields(tt.content)
			assert.Equal(t, tt.want, meta["scope"])
		})
	}
}

func TestExtractMetaFields_FrontmatterFields(t *testing.T) {
	content := "---\n" +
		"status: accepted\n" +
		"decision-makers: Alice, Bob\n" +
		"consulted: \"Carol, Dave\"\n" +
		"informed: Eve\n" +
		"---\n\n# 1. Title\n"

	meta := adr.ExtractMetaFields(content)
	assert.Equal(t, []string{"Alice", "Bob"}, meta["decision-makers"])
	assert.Equal(t, []string{"Carol", "Dave"}, meta["consulted"], "quoted list should be unquoted then split")
	assert.Equal(t, []string{"Eve"}, meta["informed"])
}

func TestExtractMetaFields_SkipsWholePlaceholder(t *testing.T) {
	// The shipped MADR template placeholder must not become a facet value.
	content := "---\n" +
		"decision-makers: {list everyone involved in the decision}\n" +
		"---\n\n# 1. Title\n"

	meta := adr.ExtractMetaFields(content)
	_, ok := meta["decision-makers"]
	assert.False(t, ok, "unfilled {…} placeholder must be dropped")
}

func TestExtractMetaFields_PlaceholderWithCommaNotMisSplit(t *testing.T) {
	content := "---\n" +
		"consulted: {experts: sre, backend}\n" +
		"---\n\n# 1. Title\n"

	meta := adr.ExtractMetaFields(content)
	_, ok := meta["consulted"]
	assert.False(t, ok, "a {…} placeholder containing a comma must be dropped whole, not split")
}

func TestExtractMetaFields_DropsPostSplitPlaceholderFragment(t *testing.T) {
	content := "---\n" +
		"decision-makers: Alice, {TBD}\n" +
		"---\n\n# 1. Title\n"

	meta := adr.ExtractMetaFields(content)
	assert.Equal(t, []string{"Alice"}, meta["decision-makers"], "post-split {…} fragment must be dropped")
}

func TestExtractMetaFields_MissingFieldsOmitted(t *testing.T) {
	meta := adr.ExtractMetaFields("# 1. Plain\n\n## Status\n\nAccepted\n")
	assert.Empty(t, meta, "an ADR with no recognized metadata yields no keys")
}

func TestExtractMetaFields_FrontmatterAndBodyIsolation(t *testing.T) {
	// A body line "Scope:" is title-block metadata; a frontmatter "scope:" is NOT read
	// as scope (scope is a title-block field), and body prose containing a frontmatter
	// key must not be read as that frontmatter field.
	content := "---\n" +
		"decision-makers: Alice\n" +
		"---\n\n# 1. Title\n\nScope: backend\n\nSome prose mentioning decision-makers: not-a-field.\n"

	meta := adr.ExtractMetaFields(content)
	assert.Equal(t, []string{"backend"}, meta["scope"])
	assert.Equal(t, []string{"Alice"}, meta["decision-makers"], "frontmatter field read from frontmatter, not body prose")
}

func TestExtractMetaFields_HeadingDiffersFromKey(t *testing.T) {
	// Regression: frontmatter fields must be matched on their lowercase Key
	// (decision-makers), never on the friendly Heading ("Decision Makers").
	content := "---\ndecision-makers: Alice\n---\n\n# 1. Title\n"
	meta := adr.ExtractMetaFields(content)
	assert.Equal(t, []string{"Alice"}, meta["decision-makers"])
	assert.NotContains(t, meta, "Decision Makers")
}
