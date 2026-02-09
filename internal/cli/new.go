package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/malek/adr-helper/internal/adr"
	"github.com/spf13/cobra"
)

// NewNewCmd creates the new subcommand for creating a new ADR.
func NewNewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new <title>",
		Short: "Create a new ADR",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := args[0]

			cfg, err := adr.LoadConfig(".")
			if err != nil {
				return err
			}

			if _, err := os.Stat(cfg.Directory); err != nil {
				return fmt.Errorf("ADR directory %q not found: %w", cfg.Directory, err)
			}

			templatePath := filepath.Join(cfg.Directory, cfg.TemplateFile)
			templateContent, err := os.ReadFile(templatePath)
			if err != nil {
				return fmt.Errorf("reading template %q: %w", templatePath, err)
			}

			number, err := adr.NextNumber(cfg.Directory)
			if err != nil {
				return err
			}

			filename, err := adr.FormatFilename(number, title)
			if err != nil {
				return err
			}

			record := adr.New(number, title)
			rendered := adr.RenderTemplate(string(templateContent), record)

			filePath := filepath.Join(cfg.Directory, filename)
			if err := os.WriteFile(filePath, []byte(rendered), 0o644); err != nil {
				return fmt.Errorf("writing ADR: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", filePath)
			return nil
		},
	}
	return cmd
}
