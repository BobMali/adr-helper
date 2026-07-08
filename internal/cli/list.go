package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type listJSON struct {
	Number int        `json:"number"`
	Title  string     `json:"title"`
	Status adr.Status `json:"status"`
	Date   string     `json:"date"`
}

// NewListCmd creates the list subcommand for displaying all ADRs.
func NewListCmd() *cobra.Command {
	var plain bool
	var jsonOutput bool
	var search string
	var count bool
	var scopes []string
	var scopeMatch string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all ADRs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			matchAll, err := parseScopeMatch(scopeMatch)
			if err != nil {
				return err
			}

			cfg, err := adr.LoadConfig(".")
			if err != nil {
				return err
			}

			repo := adr.NewFileRepository(cfg.Directory)
			records, err := repo.List(cmd.Context())
			if err != nil {
				return err
			}

			if search != "" {
				records = adr.FilterByQuery(records, search)
			}

			if len(scopes) > 0 {
				records = adr.FilterByMetaField(records, "scope", scopes, matchAll)
			}

			if count {
				counts := adr.CountByStatus(records)
				if jsonOutput {
					return json.NewEncoder(cmd.OutOrStdout()).Encode(counts)
				}

				noColor := plain || os.Getenv("NO_COLOR") != ""
				greenStyle := color.New(color.FgGreen)
				yellowStyle := color.New(color.FgYellow)
				redStyle := color.New(color.FgRed)
				if noColor {
					greenStyle.DisableColor()
					yellowStyle.DisableColor()
					redStyle.DisableColor()
				} else {
					greenStyle.EnableColor()
					yellowStyle.EnableColor()
					redStyle.EnableColor()
				}

				// Compute column width from the longest visible status name.
				colWidth := len("Status")
				for _, s := range adr.AllStatuses() {
					if n := len(s.String()); n > colWidth {
						colWidth = n
					}
				}
				colWidth += 3 // padding between columns

				var buf bytes.Buffer
				fmt.Fprintf(&buf, "%-*s%s\n", colWidth, "Status", "Count")
				for _, s := range adr.AllStatuses() {
					label := statusColor(s.String(), greenStyle, yellowStyle, redStyle)
					pad := colWidth - len(s.String())
					fmt.Fprintf(&buf, "%s%*s%d\n", label, pad, "", counts.ByStatus[s])
				}
				fmt.Fprintln(&buf)
				fmt.Fprintf(&buf, "%-*s%d\n", colWidth, "Total", counts.Total)
				_, err := buf.WriteTo(cmd.OutOrStdout())
				return err
			}

			if jsonOutput {
				result := make([]listJSON, len(records))
				for i, r := range records {
					dateStr := ""
					if !r.Date.IsZero() {
						dateStr = r.Date.Format("2006-01-02")
					}
					result[i] = listJSON{
						Number: r.Number,
						Title:  r.Title,
						Status: r.Status,
						Date:   dateStr,
					}
				}
				return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
			}

			noColor := plain || os.Getenv("NO_COLOR") != ""

			greenStyle := color.New(color.FgGreen)
			yellowStyle := color.New(color.FgYellow)
			redStyle := color.New(color.FgRed)
			if noColor {
				greenStyle.DisableColor()
				yellowStyle.DisableColor()
				redStyle.DisableColor()
			} else {
				greenStyle.EnableColor()
				yellowStyle.EnableColor()
				redStyle.EnableColor()
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tDate\tTitle\tStatus")
			for _, r := range records {
				dateStr := ""
				if !r.Date.IsZero() {
					dateStr = r.Date.Format("2006-01-02")
				}
				statusStr := statusColor(r.Status.String(), greenStyle, yellowStyle, redStyle)
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", r.Number, dateStr, r.Title, statusStr)
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVar(&plain, "plain", false, "disable colored output")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON array")
	cmd.Flags().StringVarP(&search, "search", "s", "", "filter ADRs by title or number")
	cmd.Flags().BoolVar(&count, "count", false, "show status counts instead of listing ADRs")
	cmd.Flags().StringSliceVar(&scopes, "scope", nil, "filter ADRs by scope (repeatable or comma-separated)")
	cmd.Flags().StringVar(&scopeMatch, "scope-match", "any", "how multiple --scope values combine: any (union) or all (intersection)")
	return cmd
}

// parseScopeMatch validates the --scope-match value and reports whether ALL selected
// scopes must be present (true) versus ANY (false, the default).
func parseScopeMatch(mode string) (bool, error) {
	switch mode {
	case "any":
		return false, nil
	case "all":
		return true, nil
	default:
		return false, fmt.Errorf("invalid --scope-match %q: expected \"any\" or \"all\"", mode)
	}
}
