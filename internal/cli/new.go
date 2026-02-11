package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/spf13/cobra"
)

// NewNewCmd creates the new subcommand for creating a new ADR.
func NewNewCmd() *cobra.Command {
	var supersedes []int

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

			if len(supersedes) > 0 {
				// Validate and deduplicate IDs
				ids, err := deduplicateIDs(supersedes)
				if err != nil {
					return err
				}

				// Resolve all superseded ADR files — fail early
				type mutation struct {
					path    string
					content string
				}
				var mutations []mutation
				var links []adr.SupersedesLink

				for _, id := range ids {
					oldFilename, err := adr.FindADRFile(cfg.Directory, id)
					if err != nil {
						return fmt.Errorf("cannot supersede ADR %04d: %w", id, err)
					}
					links = append(links, adr.SupersedesLink{Number: id, Filename: oldFilename})

					// Read and compute new content
					oldPath := filepath.Join(cfg.Directory, oldFilename)
					oldContent, err := os.ReadFile(oldPath)
					if err != nil {
						return fmt.Errorf("reading ADR %04d: %w", id, err)
					}

					newLink := adr.SupersedesLink{Number: number, Filename: filename}
					updatedContent, err := adr.SetSupersededBy(string(oldContent), newLink)
					if err != nil {
						return fmt.Errorf("updating ADR %04d: %w", id, err)
					}
					mutations = append(mutations, mutation{path: oldPath, content: updatedContent})
				}

				// Compute new ADR content with supersedes links
				rendered, err = adr.SetSupersedes(rendered, links)
				if err != nil {
					return fmt.Errorf("setting supersedes in new ADR: %w", err)
				}

				// All computation succeeded — now write files
				filePath := filepath.Join(cfg.Directory, filename)
				if err := os.WriteFile(filePath, []byte(rendered), 0o644); err != nil {
					return fmt.Errorf("writing ADR: %w", err)
				}

				for _, m := range mutations {
					if err := os.WriteFile(m.path, []byte(m.content), 0o644); err != nil {
						return fmt.Errorf("writing updated ADR: %w", err)
					}
					fmt.Fprintf(cmd.OutOrStdout(), "Superseded %s\n", m.path)
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", filePath)
				return nil
			}

			filePath := filepath.Join(cfg.Directory, filename)
			if err := os.WriteFile(filePath, []byte(rendered), 0o644); err != nil {
				return fmt.Errorf("writing ADR: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", filePath)
			return nil
		},
	}

	cmd.Flags().IntSliceVarP(&supersedes, "supersedes", "s", nil,
		"ID of ADR(s) that this new ADR supersedes")
	return cmd
}

// deduplicateIDs validates and deduplicates a slice of ADR IDs.
func deduplicateIDs(ids []int) ([]int, error) {
	seen := make(map[int]bool)
	var result []int
	for _, id := range ids {
		if id <= 0 {
			return nil, fmt.Errorf("invalid ADR ID %d: must be positive", id)
		}
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	sort.Ints(result)
	return result, nil
}
