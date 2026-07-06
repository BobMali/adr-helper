package cli

import (
	"fmt"
	"strings"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/spf13/cobra"
)

// NewScopeCmd creates the scope command namespace for managing the controlled
// scope vocabulary stored in .adr.json.
func NewScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scope",
		Short: "Manage the scope vocabulary",
		Long:  "Manage the controlled scope vocabulary used by scoped templates.",
	}
	cmd.AddCommand(newScopeAddCmd())
	cmd.AddCommand(newScopeListCmd())
	return cmd
}

// newScopeAddCmd creates the "scope add" subcommand. It adds one or more scopes
// to the vocabulary atomically: if any value is a duplicate (case-insensitive)
// or invalid, it fails without persisting anything.
func newScopeAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <value> [value...]",
		Short: "Add one or more scopes to the vocabulary",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := adr.LoadConfig(".")
			if err != nil {
				return err
			}

			for _, v := range args {
				trimmed := strings.TrimSpace(v)
				if canonical, ok := cfg.HasScope(v); ok {
					if trimmed == canonical {
						return fmt.Errorf("scope %q already exists", canonical)
					}
					return fmt.Errorf("scope %q already exists as %q", trimmed, canonical)
				}
				if _, err := cfg.AddScope(v); err != nil {
					return err
				}
			}

			if err := adr.SaveConfig(".", cfg); err != nil {
				return err
			}

			for _, v := range args {
				fmt.Fprintf(cmd.OutOrStdout(), "Added scope %q\n", strings.TrimSpace(v))
			}
			return nil
		},
	}
}

// newScopeListCmd creates the "scope list" subcommand, printing the vocabulary
// one scope per line.
func newScopeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the scope vocabulary",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := adr.LoadConfig(".")
			if err != nil {
				return err
			}

			if len(cfg.Scopes) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No scopes defined")
				return nil
			}

			for _, s := range cfg.Scopes {
				fmt.Fprintln(cmd.OutOrStdout(), s)
			}
			return nil
		},
	}
}
