package cli

import "github.com/spf13/cobra"

// NewRootCmd creates and returns the root Cobra command for adr-cli.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adr",
		Short: "A tool for managing Architecture Decision Records",
		Long:  "adr is a command-line tool for creating and managing Architecture Decision Records (ADRs).",
	}
	return cmd
}
