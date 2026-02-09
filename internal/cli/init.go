package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/malek/adr-helper/internal/adr"
	"github.com/spf13/cobra"
)

// NewInitCmd creates the init subcommand for initializing an ADR directory.
func NewInitCmd() *cobra.Command {
	var template string
	var force bool

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a new ADR directory with a template (init)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			content, err := adr.TemplateContent(template)
			if err != nil {
				return err
			}

			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("creating directory %q: %w", dir, err)
			}

			path := filepath.Join(dir, "template.md")
			if _, err := os.Stat(path); err == nil && !force {
				return fmt.Errorf("template already exists at %q, use --force to overwrite", path)
			}
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return fmt.Errorf("writing template: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Initialized ADR directory at %s with template: %s\n", dir, template)
			return nil
		},
	}

	cmd.Flags().StringVarP(&template, "template", "t", string(adr.TemplateNygard),
		fmt.Sprintf("template format (%v)", adr.ValidTemplateNames()))
	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing template.md")

	return cmd
}
