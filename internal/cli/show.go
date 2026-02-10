package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/malek/adr-helper/internal/adr"
	"github.com/spf13/cobra"
)

// NewShowCmd creates the show subcommand for displaying an ADR in the terminal.
func NewShowCmd() *cobra.Command {
	var plain bool

	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Display an ADR in the terminal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ADR ID %q: must be a number", args[0])
			}
			if id <= 0 {
				return fmt.Errorf("invalid ADR ID %d: must be positive", id)
			}

			cfg, err := adr.LoadConfig(".")
			if err != nil {
				return err
			}

			filename, err := adr.FindADRFile(cfg.Directory, id)
			if err != nil {
				return err
			}

			content, err := os.ReadFile(filepath.Join(cfg.Directory, filename))
			if err != nil {
				return fmt.Errorf("reading ADR: %w", err)
			}

			noColor := plain || os.Getenv("NO_COLOR") != ""
			formatted := FormatADR(string(content), FormatOptions{NoColor: noColor})
			fmt.Fprint(cmd.OutOrStdout(), formatted)
			return nil
		},
	}

	cmd.Flags().BoolVar(&plain, "plain", false, "disable colored output")
	return cmd
}
