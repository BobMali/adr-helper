package adr

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ Repository = (*FileRepository)(nil)

func TestFileRepository_List_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	repo := NewFileRepository(dir)
	adrs, err := repo.List(context.Background())

	require.NoError(t, err)
	assert.Empty(t, adrs)
}

func TestFileRepository_List_SingleADR(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0001-use-go.md", "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n")

	repo := NewFileRepository(dir)
	adrs, err := repo.List(context.Background())

	require.NoError(t, err)
	require.Len(t, adrs, 1)
	assert.Equal(t, 1, adrs[0].Number)
	assert.Equal(t, "Use Go", adrs[0].Title)
	assert.Equal(t, Accepted, adrs[0].Status)
}

func TestFileRepository_List_SortedByNumber(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0003-third.md", "# 3. Third\n\nDate: 2024-03-01\n\n## Status\n\nAccepted\n")
	writeFile(t, dir, "0001-first.md", "# 1. First\n\nDate: 2024-01-01\n\n## Status\n\nAccepted\n")
	writeFile(t, dir, "0002-second.md", "# 2. Second\n\nDate: 2024-02-01\n\n## Status\n\nProposed\n")

	repo := NewFileRepository(dir)
	adrs, err := repo.List(context.Background())

	require.NoError(t, err)
	require.Len(t, adrs, 3)
	assert.Equal(t, 1, adrs[0].Number)
	assert.Equal(t, 2, adrs[1].Number)
	assert.Equal(t, 3, adrs[2].Number)
}

func TestFileRepository_List_SkipsNonADRFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0001-use-go.md", "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n")
	writeFile(t, dir, "README.md", "# README\n")
	writeFile(t, dir, "notes.txt", "some notes")

	repo := NewFileRepository(dir)
	adrs, err := repo.List(context.Background())

	require.NoError(t, err)
	require.Len(t, adrs, 1)
	assert.Equal(t, "Use Go", adrs[0].Title)
}

func TestFileRepository_List_SkipsMalformedADR(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0001-good.md", "# 1. Good\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n")
	writeFile(t, dir, "0002-bad.md", "# 2. Bad\n\nDate: 2024-01-15\n\n## Status\n\ninvalid-status\n")

	repo := NewFileRepository(dir)
	adrs, err := repo.List(context.Background())

	require.NoError(t, err)
	require.Len(t, adrs, 1)
	assert.Equal(t, "Good", adrs[0].Title)
}

func TestFileRepository_List_NonexistentDir(t *testing.T) {
	repo := NewFileRepository("/nonexistent/path")
	_, err := repo.List(context.Background())

	assert.Error(t, err)
}

func TestFileRepository_Get(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0001-use-go.md", "# 1. Use Go\n\nDate: 2024-01-15\n\n## Status\n\nAccepted\n")

	repo := NewFileRepository(dir)
	record, err := repo.Get(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, 1, record.Number)
	assert.Equal(t, "Use Go", record.Title)
}

func TestFileRepository_Get_NotFound(t *testing.T) {
	dir := t.TempDir()

	repo := NewFileRepository(dir)
	_, err := repo.Get(context.Background(), 99)

	assert.Error(t, err)
}

func TestFileRepository_NextNumber(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0001-first.md", "# 1. First\n")
	writeFile(t, dir, "0002-second.md", "# 2. Second\n")

	repo := NewFileRepository(dir)
	next, err := repo.NextNumber(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 3, next)
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
	require.NoError(t, err)
}
