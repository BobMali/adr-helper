package cli_test

import (
	"bytes"
	"testing"

	"github.com/malek/adr-helper/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCmd_ExecutesWithoutError(t *testing.T) {
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)
}

func TestNewRootCmd_HelpContainsExpectedText(t *testing.T) {
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "adr-cli")
	assert.Contains(t, output, "Architecture Decision Records")
}
