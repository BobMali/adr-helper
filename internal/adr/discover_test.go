package adr_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractScope(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantValue string
		wantOK    bool
	}{
		{
			name:      "present with value",
			content:   "# 1. Title\n\nDate:\n\nScope: Backend, Frontend\n\n## Status\n\nAccepted\n",
			wantValue: "Backend, Frontend",
			wantOK:    true,
		},
		{
			name:      "present but empty",
			content:   "# 1. Title\n\nScope:\n\n## Status\n",
			wantValue: "",
			wantOK:    true,
		},
		{
			name:      "absent",
			content:   "# 1. Title\n\nDate:\n\n## Status\n\nAccepted\n",
			wantValue: "",
			wantOK:    false,
		},
		{
			name:      "lowercase label",
			content:   "# 1. Title\n\nscope: Backend\n",
			wantValue: "Backend",
			wantOK:    true,
		},
		{
			name:      "CRLF line endings",
			content:   "# 1. Title\r\n\r\nScope: Backend, Frontend\r\n\r\n## Status\r\n",
			wantValue: "Backend, Frontend",
			wantOK:    true,
		},
		{
			name:      "frontmatter scope is ignored, body scope wins",
			content:   "---\nscope: FromFrontmatter\n---\n\n# 1. Title\n\nScope: Backend\n",
			wantValue: "Backend",
			wantOK:    true,
		},
		{
			name:      "frontmatter-only scope is not discovered",
			content:   "---\nscope: FromFrontmatter\n---\n\n# 1. Title\n\n## Status\n",
			wantValue: "",
			wantOK:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := adr.ExtractScope(tt.content)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantValue, got)
		})
	}
}

// writeADR writes an ADR file with the given scope line value into dir.
func writeADR(t *testing.T, dir, name, scopeValue string) {
	t.Helper()
	body := "# 1. Title\n\nDate:\n\nScope: " + scopeValue + "\n\n## Status\n\nAccepted\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644))
}

func TestDiscoverScopes(t *testing.T) {
	dir := t.TempDir()
	writeADR(t, dir, "0001-first.md", "Backend, Frontend")
	writeADR(t, dir, "0002-second.md", "backend, API") // duplicate casing + new value
	writeADR(t, dir, "0003-empty.md", "")              // empty scope line
	require.NoError(t, os.WriteFile(filepath.Join(dir, "0004-none.md"),
		[]byte("# 4. No Scope\n\n## Status\n\nAccepted\n"), 0o644)) // no scope line
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.md"),
		[]byte("Scope: Ignored\n"), 0o644)) // non-ADR filename
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "0005-subdir.md"), 0o755)) // a dir, not a file

	got, err := adr.DiscoverScopes(dir)
	require.NoError(t, err)
	// Raw, order-preserving, duplicates preserved. Empty/none/non-ADR excluded.
	assert.Equal(t, []string{"Backend", "Frontend", "backend", "API"}, got)
}

func TestDiscoverScopes_MissingDirectory(t *testing.T) {
	got, err := adr.DiscoverScopes(filepath.Join(t.TempDir(), "does-not-exist"))
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestConfig_MergeScopes_DedupsAndCanonicalises(t *testing.T) {
	cfg := &adr.Config{Directory: "docs/adr", Scopes: []string{"Existing"}}

	added, invalid := cfg.MergeScopes([]string{"Backend", "Frontend", "backend", "Existing", "existing"})

	assert.Equal(t, []string{"Backend", "Frontend"}, added)
	assert.Empty(t, invalid)
	assert.Equal(t, []string{"Existing", "Backend", "Frontend"}, cfg.Scopes)
}

func TestConfig_MergeScopes_RoutesInvalidValues(t *testing.T) {
	cfg := &adr.Config{Directory: "docs/adr"}
	oversized := strings.Repeat("a", 65) // exceeds MaxScopeLength (64)

	added, invalid := cfg.MergeScopes([]string{oversized, "Good", oversized})

	assert.Equal(t, []string{"Good"}, added)
	assert.Equal(t, []string{oversized}, invalid, "invalid values are deduped")
	assert.Equal(t, []string{"Good"}, cfg.Scopes)
}

func TestDiscoverAndMergeScopes(t *testing.T) {
	dir := t.TempDir()
	writeADR(t, dir, "0001-first.md", "Backend, Frontend")
	writeADR(t, dir, "0002-second.md", "backend, API")
	cfg := &adr.Config{Directory: dir}

	added, invalid, err := adr.DiscoverAndMergeScopes(cfg)
	require.NoError(t, err)
	assert.Empty(t, invalid)
	assert.Equal(t, []string{"Backend", "Frontend", "API"}, added)
	assert.Equal(t, []string{"Backend", "Frontend", "API"}, cfg.Scopes)
}
