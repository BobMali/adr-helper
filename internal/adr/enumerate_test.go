package adr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListADRFiles(t *testing.T) {
	t.Run("empty dir", func(t *testing.T) {
		files, err := listADRFiles(t.TempDir())
		require.NoError(t, err)
		assert.Empty(t, files)
	})

	t.Run("missing dir returns a raw os.IsNotExist error", func(t *testing.T) {
		_, err := listADRFiles(filepath.Join(t.TempDir(), "nope"))
		require.Error(t, err)
		assert.True(t, os.IsNotExist(err), "raw error must satisfy os.IsNotExist")
	})

	t.Run("filters non-ADR entries and parses numbers in ReadDir order", func(t *testing.T) {
		dir := t.TempDir()
		for _, n := range []string{"0001-a.md", "0002-b.md", "README.md", "template.md"} {
			require.NoError(t, os.WriteFile(filepath.Join(dir, n), []byte(""), 0o644))
		}
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "0009-x.md"), 0o755)) // dir named like an ADR

		files, err := listADRFiles(dir)
		require.NoError(t, err)
		assert.Equal(t, []adrFile{
			{Number: 1, Name: "0001-a.md"},
			{Number: 2, Name: "0002-b.md"},
		}, files)
	})

	t.Run("skips numbers that overflow int", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "99999999999999999999-x.md"), []byte(""), 0o644)) // 20 nines
		require.NoError(t, os.WriteFile(filepath.Join(dir, "0001-ok.md"), []byte(""), 0o644))

		files, err := listADRFiles(dir)
		require.NoError(t, err)
		assert.Equal(t, []adrFile{{Number: 1, Name: "0001-ok.md"}}, files)
	})
}
