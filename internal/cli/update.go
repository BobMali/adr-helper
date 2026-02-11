package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/spf13/cobra"
)

// NewUpdateCmd creates the update subcommand for changing an ADR's status.
func NewUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id> [status]",
		Short: "Update the status of an existing ADR",
		Args:  cobra.RangeArgs(1, 2),
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

			filePath := filepath.Join(cfg.Directory, filename)
			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("reading ADR: %w", err)
			}

			var status string
			if len(args) == 2 {
				status, err = resolveStatus(args[1])
			} else {
				status, err = promptStatus(cmd)
			}
			if err != nil {
				return err
			}

			updated, err := adr.UpdateStatus(string(content), status)
			if err != nil {
				return err
			}

			if err := os.WriteFile(filePath, []byte(updated), 0o644); err != nil {
				return fmt.Errorf("writing ADR: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Updated %s status to %s\n", filename, status)
			return nil
		},
	}
	return cmd
}

// resolveStatus validates a status string with fuzzy matching.
func resolveStatus(input string) (string, error) {
	lower := strings.ToLower(input)
	allStatuses := adr.AllStatusStrings()

	// Exact match (case-insensitive)
	for _, s := range allStatuses {
		if lower == s {
			return s, nil
		}
	}

	// Fuzzy match — find closest
	best, bestDist := closestStatus(lower, allStatuses)
	if bestDist <= 3 {
		return "", fmt.Errorf("unknown status %q, did you mean %q?", input, best)
	}

	return "", fmt.Errorf("unknown status %q, valid statuses: %s", input, strings.Join(allStatuses, ", "))
}

// promptStatus displays an interactive menu and reads the user's choice.
func promptStatus(cmd *cobra.Command) (string, error) {
	allStatuses := adr.AllStatusStrings()

	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Select a status:")
	for i, s := range allStatuses {
		fmt.Fprintf(out, "  %d) %s\n", i+1, s)
	}
	fmt.Fprint(out, "Enter choice: ")

	scanner := bufio.NewScanner(cmd.InOrStdin())
	if !scanner.Scan() {
		return "", fmt.Errorf("invalid choice: no input")
	}

	line := strings.TrimSpace(scanner.Text())
	choice, err := strconv.Atoi(line)
	if err != nil || choice < 1 || choice > len(allStatuses) {
		return "", fmt.Errorf("invalid choice: %q", line)
	}

	return allStatuses[choice-1], nil
}

// levenshtein computes the Levenshtein distance between two strings.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(curr[j-1]+1, min(prev[j]+1, prev[j-1]+cost))
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}

// closestStatus finds the status with the smallest Levenshtein distance.
func closestStatus(input string, statuses []string) (string, int) {
	best := statuses[0]
	bestDist := levenshtein(input, statuses[0])
	for _, s := range statuses[1:] {
		d := levenshtein(input, s)
		if d < bestDist {
			best = s
			bestDist = d
		}
	}
	return best, bestDist
}
