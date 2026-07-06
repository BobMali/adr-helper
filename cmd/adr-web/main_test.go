package main_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// freePort returns an available localhost address.
func freePort(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	return fmt.Sprintf("127.0.0.1:%d", listener.Addr().(*net.TCPAddr).Port)
}

func TestBinaryCompiles(t *testing.T) {
	cmd := exec.Command("go", "build", "-o", "/dev/null", ".")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "binary should compile: %s", output)
}

func TestBinaryHealthCheck(t *testing.T) {
	// Find a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", ".", "--addr", addr)
	require.NoError(t, cmd.Start())
	defer cmd.Process.Kill()

	// Wait for server to start
	var resp *http.Response
	for i := 0; i < 50; i++ {
		resp, err = http.Get(fmt.Sprintf("http://%s/health", addr))
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	require.NoError(t, err, "server should start and respond")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestBinary_DiscoversScopesAtBoot(t *testing.T) {
	// Build the binary so we can run it with an arbitrary working directory.
	bin := filepath.Join(t.TempDir(), "adr-web-test")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	// Workspace: config + ADRs with one valid scope and one oversized (invalid) one.
	work := t.TempDir()
	adrDir := filepath.Join(work, "docs", "adr")
	require.NoError(t, os.MkdirAll(adrDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(work, ".adr.json"),
		[]byte(`{"version":"1","directory":"docs/adr","template":"nygard-scoped","templateFile":"template.md"}`), 0o644))
	oversized := strings.Repeat("a", 65) // exceeds MaxScopeLength
	require.NoError(t, os.WriteFile(filepath.Join(adrDir, "0001-valid.md"),
		[]byte("# 1. Valid\n\nScope: Backend\n\n## Status\n\nAccepted\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(adrDir, "0002-invalid.md"),
		[]byte("# 2. Invalid\n\nScope: "+oversized+"\n\n## Status\n\nAccepted\n"), 0o644))

	addr := freePort(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, "--addr", addr)
	cmd.Dir = work
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	require.NoError(t, cmd.Start())
	defer cmd.Process.Kill()

	// Wait for readiness and read the served scopes.
	var body []byte
	for i := 0; i < 50; i++ {
		resp, e := http.Get(fmt.Sprintf("http://%s/api/scopes", addr))
		if e == nil {
			body, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.Contains(t, string(body), "Backend", "discovered valid scope should be served")
	assert.NotContains(t, string(body), oversized, "invalid scope must not be served")

	// The invalid-scope warning is logged unconditionally at boot (guards the
	// hoisted-warning fix). Kill + Wait first so reading the stderr buffer is race-free.
	cmd.Process.Kill()
	_ = cmd.Wait()
	assert.Contains(t, stderr.String(), "skipped invalid scope")

	// Overlay-only: .adr.json must NOT have been rewritten at boot.
	data, err := os.ReadFile(filepath.Join(work, ".adr.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(data), "Backend", "boot discovery must not persist to disk")
}
