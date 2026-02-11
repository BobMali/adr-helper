package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/spf13/cobra"
)

// validateTemplateFile checks that templateFile is a valid filename for the template.
func validateTemplateFile(name string) error {
	if name == "" {
		return fmt.Errorf("template file name must not be empty")
	}
	if strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("template file name must not contain path separators: %q", name)
	}
	if filepath.Ext(name) != ".md" {
		return fmt.Errorf("template file name must have .md extension: %q", name)
	}
	return nil
}

// NewInitCmd creates the init subcommand for initializing an ADR directory.
func NewInitCmd() *cobra.Command {
	var template string
	var templateFile string
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

			if err := validateTemplateFile(templateFile); err != nil {
				return err
			}

			content, err := adr.TemplateContent(template)
			if err != nil {
				return err
			}

			// Preflight: check for existing files before any writes
			configPath := filepath.Join(".", adr.ConfigFileName)
			if _, err := os.Stat(configPath); err == nil && !force {
				return fmt.Errorf("config already exists at %q, use --force to overwrite", configPath)
			}

			templatePath := filepath.Join(dir, templateFile)
			if _, err := os.Stat(templatePath); err == nil && !force {
				return fmt.Errorf("template already exists at %q, use --force to overwrite", templatePath)
			}

			// Mutations
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("creating directory %q: %w", dir, err)
			}

			if err := os.WriteFile(templatePath, []byte(content), 0o644); err != nil {
				return fmt.Errorf("writing template: %w", err)
			}

			cfg := &adr.Config{
				Directory:    dir,
				Template:     template,
				TemplateFile: templateFile,
			}
			if err := adr.SaveConfig(".", cfg); err != nil {
				return fmt.Errorf("writing config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Initialized ADR directory at %s with template: %s\n", dir, template)
			return nil
		},
	}

	cmd.Flags().StringVarP(&template, "template", "t", string(adr.TemplateNygard),
		fmt.Sprintf("template format (%v)", adr.ValidTemplateNames()))
	cmd.Flags().StringVar(&templateFile, "template-file", "template.md",
		"output filename for the template")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing files if they exist")

	return cmd
}
