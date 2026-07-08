package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BobMali/adr-helper/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewListCmd_UseAndShort(t *testing.T) {
	cmd := cli.NewListCmd()
	assert.Equal(t, "list", cmd.Use)
	assert.Contains(t, cmd.Short, "List")
}

func TestListCmd_DisplaysADRs(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	// Create two ADR files
	adr1 := "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	adr2 := "# 2. Use Chi Router\n\nDate: 2024-02-01\n\n## Status\n\nProposed\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adr1), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0002-use-chi-router.md"), []byte(adr2), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--plain"})

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "1")
	assert.Contains(t, output, "2024-01-15")
	assert.Contains(t, output, "Use Go")
	assert.Contains(t, output, "Accepted")
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "2024-02-01")
	assert.Contains(t, output, "Use Chi Router")
	assert.Contains(t, output, "Proposed")
}

func TestListCmd_EmptyDirectory(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--plain"})

	err := root.Execute()
	require.NoError(t, err)

	// Should have a header but no data rows
	output := buf.String()
	assert.Contains(t, output, "ID")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 1, len(lines), "expected only header line")
}

func TestListCmd_NoConfig_ReturnsError(t *testing.T) {
	chdirTemp(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"list"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
}

func TestListCmd_JSON_OutputsValidJSONArray(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adr1 := "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adr1), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--json"})

	err := root.Execute()
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 1)
	assert.Equal(t, float64(1), result[0]["number"])
	assert.Equal(t, "Use Go", result[0]["title"])
	assert.Equal(t, "Accepted", result[0]["status"])
	assert.Equal(t, "2024-01-15", result[0]["date"])
}

func TestListCmd_JSON_EmptyDirectory_OutputsEmptyArray(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--json"})

	err := root.Execute()
	require.NoError(t, err)

	output := strings.TrimSpace(buf.String())
	assert.Equal(t, "[]", output)
}

func TestListCmd_PlainSuppressesANSI(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adr1 := "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adr1), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--plain"})

	err := root.Execute()
	require.NoError(t, err)

	assert.NotContains(t, buf.String(), "\x1b[")
}

func TestListCmd_NoColorEnvSuppressesANSI(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adr1 := "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adr1), 0o644))

	t.Setenv("NO_COLOR", "1")

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list"})

	err := root.Execute()
	require.NoError(t, err)

	assert.NotContains(t, buf.String(), "\x1b[")
}

func TestListCmd_JSON_TakesPrecedenceOverPlain(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adr1 := "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adr1), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--json", "--plain"})

	err := root.Execute()
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Len(t, result, 1)
}

func TestListCmd_SearchByTitle(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adr1 := "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	adr2 := "# 2. Use Chi Router\n\nDate: 2024-02-01\n\n## Status\n\nProposed\n\n## Context\n\nContext.\n"
	adr3 := "# 3. Use PostgreSQL\n\nDate: 2024-03-10\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adr1), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0002-use-chi-router.md"), []byte(adr2), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0003-use-postgresql.md"), []byte(adr3), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--plain", "-s", "Chi"})

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Use Chi Router")
	assert.NotContains(t, output, "Use Go")
	assert.NotContains(t, output, "Use PostgreSQL")
}

func TestListCmd_SearchByNumber(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adr1 := "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	adr2 := "# 2. Use Chi Router\n\nDate: 2024-02-01\n\n## Status\n\nProposed\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adr1), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0002-use-chi-router.md"), []byte(adr2), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--plain", "-s", "2"})

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Use Chi Router")
	assert.NotContains(t, output, "Use Go")
}

func TestListCmd_SearchNoResults(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adr1 := "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adr1), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--plain", "--search", "nonexistent"})

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "ID")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 1, len(lines), "expected only header line when no results match")
}

func TestListCmd_SearchJSON(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adr1 := "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	adr2 := "# 2. Use Chi Router\n\nDate: 2024-02-01\n\n## Status\n\nProposed\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adr1), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0002-use-chi-router.md"), []byte(adr2), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--json", "--search", "Go"})

	err := root.Execute()
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 1)
	assert.Equal(t, "Use Go", result[0]["title"])
}

func TestListCmd_SearchCaseInsensitive(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adr1 := "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	adr2 := "# 2. Use Chi Router\n\nDate: 2024-02-01\n\n## Status\n\nProposed\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adr1), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0002-use-chi-router.md"), []byte(adr2), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--plain", "-s", "chi router"})

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Use Chi Router")
	assert.NotContains(t, output, "Use Go")
}

func TestListCmd_Count_ShowsStatusCounts(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adr1 := "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	adr2 := "# 2. Use Chi Router\n\nDate: 2024-02-01\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	adr3 := "# 3. Use PostgreSQL\n\nDate: 2024-03-10\n\n## Status\n\nProposed\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adr1), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0002-use-chi-router.md"), []byte(adr2), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0003-use-postgresql.md"), []byte(adr3), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--count", "--plain"})

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Proposed")
	assert.Contains(t, output, "Accepted")
	assert.Contains(t, output, "Total")
	// Check specific counts appear: 2 Accepted, 1 Proposed
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Accepted") {
			assert.Contains(t, line, "2")
		}
		if strings.Contains(line, "Proposed") {
			assert.Contains(t, line, "1")
		}
		if strings.Contains(line, "Total") {
			assert.Contains(t, line, "3")
		}
	}
}

func TestListCmd_Count_EmptyDirectory(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--count", "--plain"})

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Total")
	// Total should be 0
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Total") {
			assert.Contains(t, line, "0")
		}
	}
}

func TestListCmd_Count_JSON(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adr1 := "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	adr2 := "# 2. Use Chi Router\n\nDate: 2024-02-01\n\n## Status\n\nProposed\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adr1), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0002-use-chi-router.md"), []byte(adr2), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--count", "--json"})

	err := root.Execute()
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, float64(2), result["total"])

	byStatus, ok := result["byStatus"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(1), byStatus["Accepted"])
	assert.Equal(t, float64(1), byStatus["Proposed"])
	assert.Equal(t, float64(0), byStatus["Rejected"])
}

func TestListCmd_Count_WithSearch(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adr1 := "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	adr2 := "# 2. Use Chi Router\n\nDate: 2024-02-01\n\n## Status\n\nProposed\n\n## Context\n\nContext.\n"
	adr3 := "# 3. Use Go Modules\n\nDate: 2024-03-10\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adr1), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0002-use-chi-router.md"), []byte(adr2), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0003-use-go-modules.md"), []byte(adr3), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--count", "--plain", "-s", "Go"})

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Only 2 ADRs match "Go": #1 and #3, both Accepted
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Total") {
			assert.Contains(t, line, "2")
		}
		if strings.Contains(line, "Accepted") {
			assert.Contains(t, line, "2")
		}
		if strings.Contains(line, "Proposed") {
			assert.Contains(t, line, "0")
		}
	}
}

func TestListCmd_Count_PlainSuppressesANSI(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adr1 := "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adr1), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--count", "--plain"})

	err := root.Execute()
	require.NoError(t, err)

	assert.NotContains(t, buf.String(), "\x1b[")
}

func TestListCmd_RejectsExtraArgs(t *testing.T) {
	chdirTemp(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"list", "foo"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
}

// --- scope filtering ---

func writeScopedADRs(t *testing.T, dir string) {
	t.Helper()
	a := "# 1. Alpha\n\nScope: backend, api\n\n## Status\n\nAccepted\n"
	b := "# 2. Beta\n\nScope: web\n\n## Status\n\nProposed\n"
	c := "# 3. Gamma\n\nScope: backend\n\n## Status\n\nAccepted\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "0001-alpha.md"), []byte(a), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "0002-beta.md"), []byte(b), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "0003-gamma.md"), []byte(c), 0o644))
}

func runList(t *testing.T, args ...string) string {
	t.Helper()
	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs(append([]string{"list"}, args...))
	require.NoError(t, root.Execute())
	return buf.String()
}

func TestListCmd_ScopeFilter_Any(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard-scoped")
	writeScopedADRs(t, filepath.Join(tmpDir, "docs/adr"))

	out := runList(t, "--plain", "--scope", "web")
	assert.Contains(t, out, "Beta")
	assert.NotContains(t, out, "Alpha")
	assert.NotContains(t, out, "Gamma")
}

func TestListCmd_ScopeFilter_MultipleAnyUnion(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard-scoped")
	writeScopedADRs(t, filepath.Join(tmpDir, "docs/adr"))

	out := runList(t, "--plain", "--scope", "web,backend")
	assert.Contains(t, out, "Alpha")
	assert.Contains(t, out, "Beta")
	assert.Contains(t, out, "Gamma")
}

func TestListCmd_ScopeFilter_All(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard-scoped")
	writeScopedADRs(t, filepath.Join(tmpDir, "docs/adr"))

	out := runList(t, "--plain", "--scope", "backend", "--scope", "api", "--scope-match", "all")
	assert.Contains(t, out, "Alpha")
	assert.NotContains(t, out, "Gamma") // has backend but not api
	assert.NotContains(t, out, "Beta")
}

func TestListCmd_ScopeFilter_CaseInsensitiveAndLenientUnknown(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard-scoped")
	writeScopedADRs(t, filepath.Join(tmpDir, "docs/adr"))

	out := runList(t, "--plain", "--scope", "BACKEND")
	assert.Contains(t, out, "Alpha")
	assert.Contains(t, out, "Gamma")

	// Unknown scope is not an error — it simply matches nothing.
	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"list", "--plain", "--scope", "does-not-exist"})
	require.NoError(t, root.Execute())
	assert.NotContains(t, buf.String(), "Alpha")
}

func TestListCmd_ScopeFilter_JSONReflectsFilter(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard-scoped")
	writeScopedADRs(t, filepath.Join(tmpDir, "docs/adr"))

	out := runList(t, "--json", "--scope", "web")
	var rows []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &rows))
	require.Len(t, rows, 1)
	assert.Equal(t, "Beta", rows[0]["title"])
}

func TestListCmd_ScopeFilter_CountReflectsFilter(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard-scoped")
	writeScopedADRs(t, filepath.Join(tmpDir, "docs/adr"))

	out := runList(t, "--plain", "--count", "--scope", "backend")
	assert.Contains(t, out, "Total")
	assert.Contains(t, out, "2") // Alpha + Gamma
}

func TestListCmd_ScopeMatch_Invalid(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard-scoped")
	writeScopedADRs(t, filepath.Join(tmpDir, "docs/adr"))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"list", "--scope", "backend", "--scope-match", "sometimes"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scope-match")
}
