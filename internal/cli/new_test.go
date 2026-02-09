package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/malek/adr-helper/internal/adr"
	"github.com/malek/adr-helper/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initWorkspace creates a .adr.json and ADR directory with a template file.
func initWorkspace(t *testing.T, tmpDir, dir, template string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, dir), 0o755))

	content, err := adr.TemplateContent(template)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, dir, "template.md"), []byte(content), 0o644))

	cfg := &adr.Config{
		Directory:    dir,
		Template:     template,
		TemplateFile: "template.md",
	}
	require.NoError(t, adr.SaveConfig(tmpDir, cfg))
}

func TestNewNewCmd_UseAndShort(t *testing.T) {
	cmd := cli.NewNewCmd()
	assert.Equal(t, "new <title>", cmd.Use)
	assert.Contains(t, cmd.Short, "new")
}

func TestNewCmd_CreatesFirstADR(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "Use Go for CLI"})

	err := root.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go-for-cli.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "# 1. Use Go for CLI")
}

func TestNewCmd_IncrementsFromExisting(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	// Create an existing ADR
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-first.md"), []byte(""), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "Second Decision"})

	err := root.Execute()
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(tmpDir, "docs/adr", "0002-second-decision.md"))
	assert.NoError(t, err)
}

func TestNewCmd_NonSequentialNumbers(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-x.md"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0005-x.md"), []byte(""), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "Next One"})

	err := root.Execute()
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(tmpDir, "docs/adr", "0006-next-one.md"))
	assert.NoError(t, err)
}

func TestNewCmd_UsesTemplateFromConfig(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := "docs/adr"
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, dir), 0o755))

	content, err := adr.TemplateContent("nygard")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, dir, "custom.md"), []byte(content), 0o644))

	cfg := &adr.Config{
		Directory:    dir,
		Template:     "nygard",
		TemplateFile: "custom.md",
	}
	require.NoError(t, adr.SaveConfig(tmpDir, cfg))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "Test"})

	err = root.Execute()
	require.NoError(t, err)

	result, err := os.ReadFile(filepath.Join(tmpDir, dir, "0001-test.md"))
	require.NoError(t, err)
	assert.Contains(t, string(result), "# 1. Test")
}

func TestNewCmd_RendersNygardTemplate(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "Use Go"})

	err := root.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"))
	require.NoError(t, err)
	s := string(content)

	assert.Contains(t, s, "# 1. Use Go")
	assert.Contains(t, s, "Date: ")
	assert.Contains(t, s, "## Status")
	assert.Contains(t, s, "## Context")
	assert.Contains(t, s, "## Decision")
	assert.Contains(t, s, "## Consequences")
}

func TestNewCmd_RendersMADRTemplate(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "madr-full")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "Use Chi"})

	err := root.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0001-use-chi.md"))
	require.NoError(t, err)
	s := string(content)

	assert.Contains(t, s, "# 1. Use Chi")
	assert.Contains(t, s, "date: ")
	assert.Contains(t, s, "## Context and Problem Statement")
}

func TestNewCmd_NoArgs_ReturnsError(t *testing.T) {
	chdirTemp(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
}

func TestNewCmd_EmptyTitle_ReturnsError(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "!!!"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slug")
}

func TestNewCmd_NoConfig_ReturnsError(t *testing.T) {
	chdirTemp(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "Test"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
}

func TestNewCmd_MissingTemplateFile_ReturnsError(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := "docs/adr"
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, dir), 0o755))

	// Config pointing to non-existent template
	cfg := &adr.Config{
		Directory:    dir,
		Template:     "nygard",
		TemplateFile: "missing.md",
	}
	require.NoError(t, adr.SaveConfig(tmpDir, cfg))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "Test"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
}

func TestNewCmd_MissingDirectory_ReturnsError(t *testing.T) {
	tmpDir := chdirTemp(t)

	// Config pointing to non-existent directory
	cfg := &adr.Config{
		Directory:    "nonexistent",
		Template:     "nygard",
		TemplateFile: "template.md",
	}
	require.NoError(t, adr.SaveConfig(tmpDir, cfg))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "Test"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
}

func TestNewCmd_PrintsSuccessMessage(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")
	buf := new(bytes.Buffer)

	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"new", "Use Go"})

	err := root.Execute()
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "Created")
	assert.Contains(t, buf.String(), "0001-use-go.md")
}
