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

func TestListCmd_RejectsExtraArgs(t *testing.T) {
	chdirTemp(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"list", "foo"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
}
