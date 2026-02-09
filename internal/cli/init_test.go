package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/malek/adr-helper/internal/adr"
	"github.com/malek/adr-helper/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chdirTemp changes the working directory to a temp dir and restores it on cleanup.
func chdirTemp(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(origDir) })
	require.NoError(t, os.Chdir(tmpDir))
	return tmpDir
}

func TestNewInitCmd_UseAndShort(t *testing.T) {
	cmd := cli.NewInitCmd()
	assert.Equal(t, "init [path]", cmd.Use)
	assert.Contains(t, cmd.Short, "init")
}

func TestInitCmd_CreatesDirectoryAndWritesDefaultTemplate(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "my-adrs")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", dir})

	err := root.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "template.md"))
	require.NoError(t, err)

	assert.Contains(t, string(content), "## Status")
	assert.Contains(t, string(content), "## Context")
	assert.Contains(t, string(content), "## Decision")
	assert.Contains(t, string(content), "## Consequences")
}

func TestInitCmd_NoArgs_WritesTemplateInCurrentDir(t *testing.T) {
	tmpDir := chdirTemp(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init"})

	err := root.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "template.md"))
	require.NoError(t, err)

	assert.Contains(t, string(content), "## Status")
}

func TestInitCmd_TemplateFlagMADRMinimal(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "adrs")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", dir, "--template", "madr-minimal"})

	err := root.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "template.md"))
	require.NoError(t, err)

	assert.Contains(t, string(content), "## Context and Problem Statement")
	assert.Contains(t, string(content), "## Considered Options")
	assert.Contains(t, string(content), "## Decision Outcome")
}

func TestInitCmd_TemplateFlagMADRFull(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "adrs")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", dir, "--template", "madr-full"})

	err := root.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "template.md"))
	require.NoError(t, err)

	assert.Contains(t, string(content), "## Decision Drivers")
	assert.Contains(t, string(content), "## Pros and Cons of the Options")
	assert.Contains(t, string(content), "## More Information")
}

func TestInitCmd_ShorthandTemplateFlag(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "adrs")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", dir, "-t", "nygard"})

	err := root.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "template.md"))
	require.NoError(t, err)

	assert.Contains(t, string(content), "## Status")
}

func TestInitCmd_InvalidTemplate_ReturnsError(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "adrs")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", dir, "--template", "invalid"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestInitCmd_ExistingTemplate_RefusesToOverwrite(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "adrs")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "template.md"), []byte("existing"), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", dir})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Verify original content preserved
	content, err := os.ReadFile(filepath.Join(dir, "template.md"))
	require.NoError(t, err)
	assert.Equal(t, "existing", string(content))
}

func TestInitCmd_ForceOverwritesExistingTemplate(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "adrs")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "template.md"), []byte("existing"), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", dir, "--force"})

	err := root.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "template.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "## Status")
}

func TestInitCmd_PrintsSuccessMessage(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "adrs")
	buf := new(bytes.Buffer)

	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"init", dir, "-t", "madr-minimal"})

	err := root.Execute()
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "Initialized")
	assert.Contains(t, buf.String(), "madr-minimal")
}

// --- Config file tests ---

func TestInitCmd_WritesConfigFile(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "docs", "adr")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", dir})

	err := root.Execute()
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(tmpDir, adr.ConfigFileName))
	require.NoError(t, err)

	var cfg adr.Config
	require.NoError(t, json.Unmarshal(data, &cfg))

	assert.Equal(t, adr.ConfigVersion, cfg.Version)
	assert.Equal(t, dir, cfg.Directory)
	assert.Equal(t, "nygard", cfg.Template) // default template
	assert.Equal(t, "template.md", cfg.TemplateFile)
}

func TestInitCmd_WritesConfigWithCustomTemplate(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "docs", "adr")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", dir, "-t", "madr-full"})

	err := root.Execute()
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(tmpDir, adr.ConfigFileName))
	require.NoError(t, err)

	var cfg adr.Config
	require.NoError(t, json.Unmarshal(data, &cfg))
	assert.Equal(t, "madr-full", cfg.Template)
}

func TestInitCmd_NoArgs_WritesConfigWithDotDirectory(t *testing.T) {
	tmpDir := chdirTemp(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init"})

	err := root.Execute()
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(tmpDir, adr.ConfigFileName))
	require.NoError(t, err)

	var cfg adr.Config
	require.NoError(t, json.Unmarshal(data, &cfg))
	assert.Equal(t, ".", cfg.Directory)
}

func TestInitCmd_ExistingConfig_RefusesToOverwrite(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "adrs")

	// Pre-create config file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, adr.ConfigFileName), []byte(`{"version":"1"}`), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", dir})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestInitCmd_ExistingConfig_ForceOverwrites(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "adrs")

	// Pre-create config file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, adr.ConfigFileName), []byte(`{"version":"1"}`), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", dir, "--force"})

	err := root.Execute()
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(tmpDir, adr.ConfigFileName))
	require.NoError(t, err)

	var cfg adr.Config
	require.NoError(t, json.Unmarshal(data, &cfg))
	assert.Equal(t, dir, cfg.Directory)
	assert.Equal(t, "nygard", cfg.Template)
	assert.Equal(t, "template.md", cfg.TemplateFile)
}

func TestInitCmd_ExistingBothFiles_ForceOverwritesBoth(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "adrs")
	require.NoError(t, os.MkdirAll(dir, 0o755))

	// Pre-create both files
	require.NoError(t, os.WriteFile(filepath.Join(dir, "template.md"), []byte("old"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, adr.ConfigFileName), []byte(`{"version":"1"}`), 0o644))

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", dir, "--force", "-t", "madr-minimal"})

	err := root.Execute()
	require.NoError(t, err)

	// template.md overwritten
	content, err := os.ReadFile(filepath.Join(dir, "template.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "## Context and Problem Statement")

	// .adr.json overwritten
	data, err := os.ReadFile(filepath.Join(tmpDir, adr.ConfigFileName))
	require.NoError(t, err)
	var cfg adr.Config
	require.NoError(t, json.Unmarshal(data, &cfg))
	assert.Equal(t, "madr-minimal", cfg.Template)
	assert.Equal(t, "template.md", cfg.TemplateFile)
}

// --- Template file flag tests ---

func TestInitCmd_DefaultTemplateFile_SavedToConfig(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "adrs")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", dir})

	err := root.Execute()
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(tmpDir, adr.ConfigFileName))
	require.NoError(t, err)

	var cfg adr.Config
	require.NoError(t, json.Unmarshal(data, &cfg))
	assert.Equal(t, "template.md", cfg.TemplateFile)
}

func TestInitCmd_CustomTemplateFile_WritesNamedFile(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "adrs")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", dir, "--template-file", "custom.md"})

	err := root.Execute()
	require.NoError(t, err)

	// custom.md should exist with template content
	content, err := os.ReadFile(filepath.Join(dir, "custom.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "## Status")

	// template.md should NOT exist
	_, err = os.Stat(filepath.Join(dir, "template.md"))
	assert.True(t, os.IsNotExist(err))
}

func TestInitCmd_CustomTemplateFile_SavedToConfig(t *testing.T) {
	tmpDir := chdirTemp(t)
	dir := filepath.Join(tmpDir, "adrs")

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", dir, "--template-file", "decisions.md"})

	err := root.Execute()
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(tmpDir, adr.ConfigFileName))
	require.NoError(t, err)

	var cfg adr.Config
	require.NoError(t, json.Unmarshal(data, &cfg))
	assert.Equal(t, "decisions.md", cfg.TemplateFile)
}

func TestInitCmd_TemplateFileWithPath_ReturnsError(t *testing.T) {
	chdirTemp(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", "adrs", "--template-file", "../bad.md"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path separator")
}

func TestInitCmd_EmptyTemplateFile_ReturnsError(t *testing.T) {
	chdirTemp(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", "adrs", "--template-file", ""})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must not be empty")
}

func TestInitCmd_TemplateFileNonMdExtension_ReturnsError(t *testing.T) {
	chdirTemp(t)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"init", "adrs", "--template-file", "template.txt"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ".md")
}
