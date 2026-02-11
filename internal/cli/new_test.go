package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/BobMali/adr-helper/internal/cli"
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

func TestNewCmd_SupersedesFlag_UpdatesSupersededADR(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	// Create an existing ADR with nygard format
	adrContent := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Context\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "--supersedes", "1", "Better Approach"})

	err := root.Execute()
	require.NoError(t, err)

	// Verify superseded ADR was updated
	updated, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"))
	require.NoError(t, err)
	assert.Contains(t, string(updated), "Superseded by [ADR-0002](0002-better-approach.md)")
	assert.NotContains(t, string(updated), "\nAccepted\n")
}

func TestNewCmd_SupersedesFlag_AddsReverseLink(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Context\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "-s", "1", "Better Approach"})

	err := root.Execute()
	require.NoError(t, err)

	// Verify new ADR has Proposed status AND supersedes link
	newContent, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0002-better-approach.md"))
	require.NoError(t, err)
	assert.Contains(t, string(newContent), "Proposed")
	assert.Contains(t, string(newContent), "Supersedes [ADR-0001](0001-use-go.md)")
}

func TestNewCmd_SupersedesMultiple(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adr1 := "# 1. First\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	adr3 := "# 3. Third\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-first.md"), []byte(adr1), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0003-third.md"), []byte(adr3), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "-s", "1", "-s", "3", "Better"})

	err := root.Execute()
	require.NoError(t, err)

	// Both superseded ADRs updated
	updated1, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0001-first.md"))
	require.NoError(t, err)
	assert.Contains(t, string(updated1), "Superseded by [ADR-0004](0004-better.md)")

	updated3, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0003-third.md"))
	require.NoError(t, err)
	assert.Contains(t, string(updated3), "Superseded by [ADR-0004](0004-better.md)")

	// New ADR has Proposed status AND both reverse links
	newContent, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0004-better.md"))
	require.NoError(t, err)
	assert.Contains(t, string(newContent), "Proposed")
	assert.Contains(t, string(newContent), "Supersedes [ADR-0001](0001-first.md)")
	assert.Contains(t, string(newContent), "Supersedes [ADR-0003](0003-third.md)")
}

func TestNewCmd_SupersedesNonExistentID_ReturnsError(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "-s", "99", "Better"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)

	// No new ADR file should have been created
	entries, _ := os.ReadDir(filepath.Join(tmpDir, "docs/adr"))
	for _, e := range entries {
		assert.NotRegexp(t, `^\d{4}-`, e.Name())
	}
}

func TestNewCmd_SupersedesMADRFull(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "madr-full")

	madrContent := "---\nstatus: \"accepted\"\ndate: 2024-01-01\n---\n\n# 1. Use Go\n\n## Context and Problem Statement\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(madrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "-s", "1", "Better"})

	err := root.Execute()
	require.NoError(t, err)

	// Superseded ADR frontmatter updated
	updated, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"))
	require.NoError(t, err)
	assert.Contains(t, string(updated), "status: \"superseded by [ADR-0002](0002-better.md)\"")

	// New ADR has proposed + supersedes in frontmatter
	newContent, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0002-better.md"))
	require.NoError(t, err)
	assert.Contains(t, string(newContent), "status: \"proposed, supersedes [ADR-0001](0001-use-go.md)\"")
}

func TestNewCmd_SupersedesNoStatusInSupersededADR_ReturnsError(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	// ADR without any status format
	noStatusContent := "# 1. Use Go\n\n## Context\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(noStatusContent), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "-s", "1", "Better"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
}

func TestNewCmd_SupersedesNoStatusInNewADR_ReturnsError(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "madr-minimal")

	adrContent := "# 1. First\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-first.md"), []byte(adrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "-s", "1", "Better"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
}

func TestNewCmd_SupersedesMixedFormats(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	// Superseded ADR uses MADR full format
	madrContent := "---\nstatus: \"accepted\"\ndate: 2024-01-01\n---\n\n# 1. Use Go\n\n## Context and Problem Statement\n\nSome context.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(madrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "-s", "1", "Better"})

	err := root.Execute()
	require.NoError(t, err)

	// Superseded MADR full ADR updated correctly
	updated, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"))
	require.NoError(t, err)
	assert.Contains(t, string(updated), "status: \"superseded by [ADR-0002](0002-better.md)\"")

	// New nygard ADR has Proposed status AND supersedes link
	newContent, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0002-better.md"))
	require.NoError(t, err)
	assert.Contains(t, string(newContent), "Proposed")
	assert.Contains(t, string(newContent), "Supersedes [ADR-0001](0001-use-go.md)")
}

func TestNewCmd_SupersedesPrintsMessages(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"new", "-s", "1", "Better"})

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Superseded")
	assert.Contains(t, output, "0001-use-go.md")
}

func TestNewCmd_SupersedesDuplicateIDs(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	adrContent := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "-s", "1", "-s", "1", "Better"})

	err := root.Execute()
	require.NoError(t, err)

	// Only one supersedes line in the new ADR, with Proposed status
	newContent, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0002-better.md"))
	require.NoError(t, err)
	assert.Contains(t, string(newContent), "Proposed")
	assert.Equal(t, 1, strings.Count(string(newContent), "Supersedes [ADR-0001]"))
}

func TestNewCmd_SupersedesZeroID_ReturnsError(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "-s", "0", "Better"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
}

func TestNewCmd_SupersedesAlreadySupersededADR(t *testing.T) {
	tmpDir := chdirTemp(t)
	initWorkspace(t, tmpDir, "docs/adr", "nygard")

	// ADR already superseded by something else
	adrContent := "# 1. Use Go\n\nDate: 2024-01-01\n\n## Status\n\nSuperseded by [ADR-0003](0003-old.md)\n\n## Context\n\nContext.\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"), []byte(adrContent), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"new", "-s", "1", "Better"})

	err := root.Execute()
	require.NoError(t, err)

	updated, err := os.ReadFile(filepath.Join(tmpDir, "docs/adr", "0001-use-go.md"))
	require.NoError(t, err)
	assert.Contains(t, string(updated), "Superseded by [ADR-0002](0002-better.md)")
	assert.NotContains(t, string(updated), "ADR-0003")
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
