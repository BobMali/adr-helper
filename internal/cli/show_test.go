package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/malek/adr-helper/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewShowCmd_UseAndShort(t *testing.T) {
	cmd := cli.NewShowCmd()
	assert.Equal(t, "show <id>", cmd.Use)
	assert.Contains(t, cmd.Short, "Display")
}

func TestShowCmd_DisplaysADRContent(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Context\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"show", "1", "--plain"})

	err := root.Execute()
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "1. Use Go")
	assert.Contains(t, buf.String(), "Accepted")
	assert.Contains(t, buf.String(), "Some context.")
}

func TestShowCmd_PlainSuppressesANSI(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\n## Status\n\nAccepted\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"show", "1", "--plain"})

	err := root.Execute()
	require.NoError(t, err)
	assert.NotContains(t, buf.String(), "\x1b[")
}

func TestShowCmd_NoColorEnvSuppressesANSI(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\n## Status\n\nAccepted\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	t.Setenv("NO_COLOR", "1")

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"show", "1"})

	err := root.Execute()
	require.NoError(t, err)
	assert.NotContains(t, buf.String(), "\x1b[")
}

func TestShowCmd_NonExistentADR_ReturnsError(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"show", "99"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestShowCmd_InvalidID_ReturnsError(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"show", "abc"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ADR ID")
}

func TestShowCmd_ZeroID_ReturnsError(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"show", "0"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestShowCmd_NoConfig_ReturnsError(t *testing.T) {
	chdirTemp(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"show", "1"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
}

func TestShowCmd_MADRFullWithFrontmatter(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "madr-full")

	adrContent := "---\nstatus: \"proposed\"\ndate: 2024-01-01\n---\n\n# 1. Use Go\n\n## Context and Problem Statement\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"show", "1", "--plain"})

	err := root.Execute()
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "---")
	assert.Contains(t, buf.String(), "status: \"proposed\"")
	assert.Contains(t, buf.String(), "1. Use Go")
	assert.Contains(t, buf.String(), "Some context.")
}
