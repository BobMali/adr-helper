package main_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
