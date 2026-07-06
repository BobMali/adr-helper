package cli_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/BobMali/adr-helper/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScopeCmd_UseAndShort(t *testing.T) {
	cmd := cli.NewScopeCmd()
	assert.Equal(t, "scope", cmd.Use)
	assert.Contains(t, cmd.Short, "scope")
}

func TestScopeCmd_BareNamespace_PrintsHelp(t *testing.T) {
	tmpDir := chdirTemp(t)
	initScopedWorkspace(t, tmpDir, nil)

	root := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"scope"})

	require.NoError(t, root.Execute())
	out := buf.String()
	assert.Contains(t, out, "add")
	assert.Contains(t, out, "list")
}

func TestScopeCmd_Add_NewScope_Persisted(t *testing.T) {
	tmpDir := chdirTemp(t)
	initScopedWorkspace(t, tmpDir, nil)

	root := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"scope", "add", "Backend"})

	require.NoError(t, root.Execute())
	assert.Contains(t, buf.String(), `Added scope "Backend"`)

	cfg, err := adr.LoadConfig(tmpDir)
	require.NoError(t, err)
	_, ok := cfg.HasScope("backend")
	assert.True(t, ok, "scope should be persisted to .adr.json")
}

func TestScopeCmd_Add_Multiple_AllPersisted(t *testing.T) {
	tmpDir := chdirTemp(t)
	initScopedWorkspace(t, tmpDir, nil)

	root := cli.NewRootCmd()
	root.SetArgs([]string{"scope", "add", "Backend", "Frontend"})
	require.NoError(t, root.Execute())

	cfg, err := adr.LoadConfig(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, []string{"Backend", "Frontend"}, cfg.Scopes)
}

func TestScopeCmd_Add_ExactDuplicate_Errors(t *testing.T) {
	tmpDir := chdirTemp(t)
	initScopedWorkspace(t, tmpDir, []string{"Backend"})

	root := cli.NewRootCmd()
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{"scope", "add", "Backend"})

	err := root.Execute()
	require.Error(t, err)
	assert.Equal(t, `scope "Backend" already exists`, err.Error())

	cfg, err := adr.LoadConfig(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, []string{"Backend"}, cfg.Scopes, "config must be unchanged")
}

func TestScopeCmd_Add_DifferentCaseDuplicate_ReportsCanonical(t *testing.T) {
	tmpDir := chdirTemp(t)
	initScopedWorkspace(t, tmpDir, []string{"Backend"})

	root := cli.NewRootCmd()
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{"scope", "add", "backend"})

	err := root.Execute()
	require.Error(t, err)
	assert.Equal(t, `scope "backend" already exists as "Backend"`, err.Error())

	cfg, err := adr.LoadConfig(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, []string{"Backend"}, cfg.Scopes, "config must be unchanged")
}

func TestScopeCmd_Add_Multiple_OneDuplicate_Atomic(t *testing.T) {
	tmpDir := chdirTemp(t)
	initScopedWorkspace(t, tmpDir, []string{"Backend"})

	root := cli.NewRootCmd()
	root.SilenceErrors = true
	root.SilenceUsage = true
	// "New" is valid and would be added first, but "backend" duplicates the
	// existing scope — the whole batch must fail and persist nothing.
	root.SetArgs([]string{"scope", "add", "New", "backend"})

	err := root.Execute()
	require.Error(t, err)

	cfg, err := adr.LoadConfig(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, []string{"Backend"}, cfg.Scopes, "no scope from a failed batch may persist")
}

func TestScopeCmd_Add_WithinBatchDuplicate_Atomic(t *testing.T) {
	tmpDir := chdirTemp(t)
	initScopedWorkspace(t, tmpDir, nil)

	root := cli.NewRootCmd()
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{"scope", "add", "Alpha", "alpha"})

	err := root.Execute()
	require.Error(t, err)

	cfg, err := adr.LoadConfig(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, cfg.Scopes, "a within-batch duplicate must persist nothing")
}

func TestScopeCmd_Add_Invalid_WrapsErrInvalidScope(t *testing.T) {
	tmpDir := chdirTemp(t)
	initScopedWorkspace(t, tmpDir, nil)

	root := cli.NewRootCmd()
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{"scope", "add", "a,b"})

	err := root.Execute()
	require.Error(t, err)
	assert.True(t, errors.Is(err, adr.ErrInvalidScope))

	cfg, err := adr.LoadConfig(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, cfg.Scopes)
}

func TestScopeCmd_Add_NoConfig_Errors(t *testing.T) {
	chdirTemp(t) // no workspace initialised

	root := cli.NewRootCmd()
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{"scope", "add", "Backend"})

	err := root.Execute()
	require.Error(t, err)
	assert.True(t, errors.Is(err, adr.ErrConfigNotFound))
}

func TestScopeCmd_List_PrintsScopes(t *testing.T) {
	tmpDir := chdirTemp(t)
	initScopedWorkspace(t, tmpDir, []string{"Backend", "Frontend"})

	root := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"scope", "list"})

	require.NoError(t, root.Execute())
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Equal(t, []string{"Backend", "Frontend"}, lines)
}

func TestScopeCmd_List_Empty_PrintsNoScopes(t *testing.T) {
	tmpDir := chdirTemp(t)
	initScopedWorkspace(t, tmpDir, nil)

	root := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"scope", "list"})

	require.NoError(t, root.Execute())
	assert.Contains(t, buf.String(), "No scopes defined")
}

func TestScopeCmd_Discover_AddsScopesFromADRs(t *testing.T) {
	tmpDir := chdirTemp(t)
	initScopedWorkspace(t, tmpDir, []string{"Existing"})
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-first.md"),
		[]byte("# 1. First\n\nScope: Backend, Frontend\n\n## Status\n\nAccepted\n"), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"scope", "discover"})
	require.NoError(t, root.Execute())

	assert.Contains(t, buf.String(), `Added scope "Backend"`)
	assert.Contains(t, buf.String(), `Added scope "Frontend"`)

	cfg, err := adr.LoadConfig(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, []string{"Existing", "Backend", "Frontend"}, cfg.Scopes)
}

func TestScopeCmd_Discover_NothingNew(t *testing.T) {
	tmpDir := chdirTemp(t)
	initScopedWorkspace(t, tmpDir, []string{"Backend"})
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs/adr", "0001-first.md"),
		[]byte("# 1. First\n\nScope: backend\n\n## Status\n\nAccepted\n"), 0o644))

	buf := new(bytes.Buffer)
	root := cli.NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"scope", "discover"})
	require.NoError(t, root.Execute())

	assert.Contains(t, buf.String(), "No new scopes discovered")

	cfg, err := adr.LoadConfig(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, []string{"Backend"}, cfg.Scopes)
}
