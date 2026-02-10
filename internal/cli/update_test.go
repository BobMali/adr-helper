package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/malek/adr-helper/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpdateCmd_UseAndShort(t *testing.T) {
	cmd := cli.NewUpdateCmd()
	assert.Equal(t, "update <id> [status]", cmd.Use)
	assert.Contains(t, cmd.Short, "status")
}

func TestUpdateCmd_NygardSetsAccepted(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nProposed\n\n## Context\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"update", "1", "accepted"})

	err := root.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "## Status\n\nAccepted\n\n## Context")
	assert.NotContains(t, string(content), "Proposed")
	assert.Contains(t, buf.String(), "Updated")
	assert.Contains(t, buf.String(), "0001-use-go.md")
	assert.Contains(t, buf.String(), "accepted")
}

func TestUpdateCmd_NygardSetsRejected(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nProposed\n\n## Context\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"update", "1", "rejected"})

	err := root.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "## Status\n\nRejected\n\n## Context")
}

func TestUpdateCmd_PreservesOtherSections(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nProposed\n\n## Context\n\nContext text.\n\n## Decision\n\nDecision text.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"update", "1", "accepted"})

	err := root.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "## Context\n\nContext text.")
	assert.Contains(t, string(content), "## Decision\n\nDecision text.")
}

func TestUpdateCmd_MADRFullFrontmatter(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "madr-full")

	adrContent := "---\nstatus: \"proposed\"\ndate: 2024-01-01\n---\n\n# 1. Use Go\n\n## Context and Problem Statement\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"update", "1", "accepted"})

	err := root.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "status: \"accepted\"")
}

func TestUpdateCmd_NoArgs_ReturnsError(t *testing.T) {
	root := cli.NewRootCmd()
	root.SetArgs([]string{"update"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_InvalidID_ReturnsError(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"update", "abc", "accepted"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ADR ID")
}

func TestUpdateCmd_ZeroID_ReturnsError(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"update", "0", "accepted"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestUpdateCmd_NonExistentADR_ReturnsError(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"update", "99", "accepted"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_NoConfig_ReturnsError(t *testing.T) {
	chdirTemp(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"update", "1", "accepted"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
}

func TestUpdateCmd_NoStatusSection_ReturnsError(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	// ADR without any status section (like madr-minimal)
	noStatusContent := "# 1. Use Go\n\n## Context\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(noStatusContent), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"update", "1", "accepted"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no status section")
}

func TestUpdateCmd_CaseInsensitiveExactMatch(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\n## Status\n\nProposed\n\n## Context\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"update", "1", "ACCEPTED"})

	err := root.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "## Status\n\nAccepted")
}

func TestUpdateCmd_FuzzySuggestsClosest(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\n## Status\n\nProposed\n\n## Context\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"update", "1", "acceped"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "did you mean")
	assert.Contains(t, err.Error(), "accepted")
}

func TestUpdateCmd_FuzzyListsAllForGarbage(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\n## Status\n\nProposed\n\n## Context\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"update", "1", "foobar"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "valid statuses")
}

func TestUpdateCmd_InteractiveSelectsStatus(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\n## Status\n\nProposed\n\n## Context\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetIn(strings.NewReader("2\n"))
	root.SetArgs([]string{"update", "1"})

	err := root.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "## Status\n\nAccepted")

	assert.Contains(t, buf.String(), "Select a status")
}

func TestUpdateCmd_InteractiveInvalidChoice(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\n## Status\n\nProposed\n\n## Context\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetIn(strings.NewReader("9\n"))
	root.SetArgs([]string{"update", "1"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid choice")
}

func TestUpdateCmd_InteractiveNonNumeric(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\n## Status\n\nProposed\n\n## Context\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetIn(strings.NewReader("abc\n"))
	root.SetArgs([]string{"update", "1"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid choice")
}

func TestUpdateCmd_InteractiveEmptyInput(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\n## Status\n\nProposed\n\n## Context\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetIn(strings.NewReader("\n"))
	root.SetArgs([]string{"update", "1"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid choice")
}

func TestUpdateCmd_PreservesSupersedes(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	// ADR 1: original
	adr1 := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Context\n\nOriginal context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adr1), 0o644))

	// Create ADR 2 via `adr new -s 1 "Better"`
	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "-s", "1", "Better"})
	require.NoError(t, root.Execute())

	// Now update ADR 2's status to accepted
	root = cli.NewRootCmd()
	root.SetArgs([]string{"update", "2", "accepted"})
	require.NoError(t, root.Execute())

	content, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0002-better.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "Accepted")
	assert.Contains(t, string(content), "Supersedes [ADR-0001]")
	assert.NotContains(t, string(content), "Proposed")
}
