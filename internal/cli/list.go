package cli

import (
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

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all ADRs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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
	return cmd
}
