package adr_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveConfig_WritesJSONFile(t *testing.T) {
	dir := t.TempDir()

	cfg := &adr.Config{
		Directory:    "docs/adr",
		Template:     "nygard",
		TemplateFile: "custom.md",
	}
	err := adr.SaveConfig(dir, cfg)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, adr.ConfigFileName))
	require.NoError(t, err)

	var got map[string]string
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, adr.ConfigVersion, got["version"])
	assert.Equal(t, "docs/adr", got["directory"])
	assert.Equal(t, "nygard", got["template"])
	assert.Equal(t, "custom.md", got["templateFile"])
}

func TestSaveConfig_IndentedJSON(t *testing.T) {
	dir := t.TempDir()

	cfg := &adr.Config{
		Directory: "adrs",
		Template:  "madr-minimal",
	}
	err := adr.SaveConfig(dir, cfg)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, adr.ConfigFileName))
	require.NoError(t, err)

	// Indented JSON contains newlines and spaces
	assert.Contains(t, string(data), "  \"version\"")
	// Trailing newline for POSIX compliance
	assert.True(t, data[len(data)-1] == '\n')
}

func TestSaveConfig_SetsVersionAutomatically(t *testing.T) {
	dir := t.TempDir()

	cfg := &adr.Config{
		Version:   "should-be-overwritten",
		Directory: "docs/adr",
		Template:  "nygard",
	}
	err := adr.SaveConfig(dir, cfg)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, adr.ConfigFileName))
	require.NoError(t, err)

	var got adr.Config
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, adr.ConfigVersion, got.Version)
}

func TestLoadConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()

	want := &adr.Config{
		Directory:    "docs/adr",
		Template:     "madr-full",
		TemplateFile: "decisions.md",
	}
	require.NoError(t, adr.SaveConfig(dir, want))

	got, err := adr.LoadConfig(dir)
	require.NoError(t, err)

	assert.Equal(t, adr.ConfigVersion, got.Version)
	assert.Equal(t, "docs/adr", got.Directory)
	assert.Equal(t, "madr-full", got.Template)
	assert.Equal(t, "decisions.md", got.TemplateFile)
}

func TestLoadConfig_MissingFile_ReturnsErrConfigNotFound(t *testing.T) {
	dir := t.TempDir()

	_, err := adr.LoadConfig(dir)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, adr.ErrConfigNotFound))
}

func TestLoadConfig_CorruptJSON_ReturnsErrConfigInvalid(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, adr.ConfigFileName), []byte("{corrupt"), 0o644))

	_, err := adr.LoadConfig(dir)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, adr.ErrConfigInvalid))
}

func TestLoadConfig_UnsupportedVersion_ReturnsErrConfigInvalid(t *testing.T) {
	dir := t.TempDir()
	data := []byte(`{"version": "99", "directory": "docs", "template": "nygard"}`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, adr.ConfigFileName), data, 0o644))

	_, err := adr.LoadConfig(dir)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, adr.ErrConfigInvalid))
}

func TestLoadConfig_MissingTemplateFile_DefaultsToTemplateMd(t *testing.T) {
	dir := t.TempDir()
	// Old-style config without templateFile field
	data := []byte(`{"version": "1", "directory": "docs", "template": "nygard"}`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, adr.ConfigFileName), data, 0o644))

	cfg, err := adr.LoadConfig(dir)
	require.NoError(t, err)
	assert.Equal(t, "template.md", cfg.TemplateFile)
}

func TestLoadConfig_EmptyDirectory_ReturnsErrConfigInvalid(t *testing.T) {
	dir := t.TempDir()
	data := []byte(`{"version": "1", "directory": "", "template": "nygard"}`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, adr.ConfigFileName), data, 0o644))

	_, err := adr.LoadConfig(dir)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, adr.ErrConfigInvalid))
}
