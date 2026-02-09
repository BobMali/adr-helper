package main_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBinaryCompiles(t *testing.T) {
	cmd := exec.Command("go", "build", "-o", "/dev/null", ".")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "binary should compile: %s", output)
}

func TestBinaryHelp(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--help")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "help should succeed: %s", output)

	assert.Contains(t, string(output), "adr")
	assert.Contains(t, string(output), "Architecture Decision Records")
}
