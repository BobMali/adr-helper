package adr_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/malek/adr-helper/internal/adr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatFilename_Basic(t *testing.T) {
	name, err := adr.FormatFilename(1, "Use Go for CLI")
	require.NoError(t, err)
	assert.Equal(t, "0001-use-go-for-cli.md", name)
}

func TestFormatFilename_LargeNumber(t *testing.T) {
	name, err := adr.FormatFilename(42, "Test")
	require.NoError(t, err)
	assert.Equal(t, "0042-test.md", name)
}

func TestFormatFilename_FourDigits(t *testing.T) {
	name, err := adr.FormatFilename(1234, "Test")
	require.NoError(t, err)
	assert.Equal(t, "1234-test.md", name)
}

func TestNextNumber_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	num, err := adr.NextNumber(dir)
	require.NoError(t, err)
	assert.Equal(t, 1, num)
}

func TestNextNumber_SingleFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "0001-test.md"), []byte(""), 0o644))

	num, err := adr.NextNumber(dir)
	require.NoError(t, err)
	assert.Equal(t, 2, num)
}

func TestNextNumber_NonSequential(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "0001-x.md"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "0005-x.md"), []byte(""), 0o644))

	num, err := adr.NextNumber(dir)
	require.NoError(t, err)
	assert.Equal(t, 6, num)
}

func TestNextNumber_IgnoresNonADRFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "template.md"), []byte(""), 0o644))

	num, err := adr.NextNumber(dir)
	require.NoError(t, err)
	assert.Equal(t, 1, num)
}

func TestNextNumber_IgnoresDirectories(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "0001-foo.md"), 0o755))

	num, err := adr.NextNumber(dir)
	require.NoError(t, err)
	assert.Equal(t, 1, num)
}

func TestNextNumber_DirectoryNotFound(t *testing.T) {
	_, err := adr.NextNumber("/nonexistent/path/to/adrs")
	assert.Error(t, err)
}

func TestFindADRFile_FindsExistingFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "0001-use-go.md"), []byte(""), 0o644))

	filename, err := adr.FindADRFile(dir, 1)
	require.NoError(t, err)
	assert.Equal(t, "0001-use-go.md", filename)
}

func TestFindADRFile_FindsAmongMultiple(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "0001-first.md"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "0003-third.md"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "0005-fifth.md"), []byte(""), 0o644))

	filename, err := adr.FindADRFile(dir, 3)
	require.NoError(t, err)
	assert.Equal(t, "0003-third.md", filename)
}

func TestFindADRFile_NotFound_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "0001-first.md"), []byte(""), 0o644))

	_, err := adr.FindADRFile(dir, 99)
	assert.Error(t, err)
}

func TestFindADRFile_DirectoryNotFound_ReturnsError(t *testing.T) {
	_, err := adr.FindADRFile("/nonexistent/path", 1)
	assert.Error(t, err)
}

func TestFindADRFile_IgnoresTemplate(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "template.md"), []byte(""), 0o644))

	_, err := adr.FindADRFile(dir, 1)
	assert.Error(t, err)
}
