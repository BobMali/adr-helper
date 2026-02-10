package cli

import (
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestFormatADR_PlainPassthrough(t *testing.T) {
	input := "Hello\nWorld\n"
	got := FormatADR(input, FormatOptions{NoColor: true})
	assert.Equal(t, input, got)
}

func TestFormatADR_H1_ColorEnabled(t *testing.T) {
	orig := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = orig }()

	input := "# 1. Use Go\n"
	got := FormatADR(input, FormatOptions{NoColor: false})
	assert.Contains(t, got, "\x1b[")
	assert.Contains(t, got, "1. Use Go")
}

func TestFormatADR_H1_PlainMode(t *testing.T) {
	input := "# 1. Use Go\n"
	got := FormatADR(input, FormatOptions{NoColor: true})
	assert.NotContains(t, got, "\x1b[")
	assert.Contains(t, got, "# 1. Use Go")
}

func TestFormatADR_H2_Bold(t *testing.T) {
	orig := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = orig }()

	input := "## Status\n"
	got := FormatADR(input, FormatOptions{NoColor: false})
	assert.Contains(t, got, "\x1b[")
	assert.Contains(t, got, "Status")
}

func TestFormatADR_FrontmatterDimmed(t *testing.T) {
	orig := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = orig }()

	input := "---\nstatus: \"proposed\"\ndate: 2024-01-01\n---\n"
	got := FormatADR(input, FormatOptions{NoColor: false})
	// Each frontmatter line should have ANSI (dim)
	for _, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		assert.Contains(t, line, "\x1b[", "frontmatter line should be styled: %q", line)
	}
}

func TestFormatADR_DateLineDimmed(t *testing.T) {
	orig := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = orig }()

	input := "Date: 2024-01-01\n"
	got := FormatADR(input, FormatOptions{NoColor: false})
	assert.Contains(t, got, "\x1b[")
	assert.Contains(t, got, "Date: 2024-01-01")
}

func TestFormatADR_StatusColors(t *testing.T) {
	orig := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = orig }()

	tests := []struct {
		statusLine string
		desc       string
	}{
		{"Accepted", "active status should be colored"},
		{"Proposed", "pending status should be colored"},
		{"Rejected", "inactive status should be colored"},
		{"Superseded by ADR-0005", "superseded should be colored"},
	}

	for _, tt := range tests {
		t.Run(tt.statusLine, func(t *testing.T) {
			input := "## Status\n\n" + tt.statusLine + "\n"
			got := FormatADR(input, FormatOptions{NoColor: false})
			lines := strings.Split(got, "\n")
			// Find the status line (not the heading, not blank)
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || strings.HasPrefix(trimmed, "#") {
					continue
				}
				// Remove ANSI to check content
				if strings.Contains(stripANSI(line), tt.statusLine) {
					assert.Contains(t, line, "\x1b[", tt.desc)
				}
			}
		})
	}
}

func TestFormatADR_BulletReplacement(t *testing.T) {
	input := "- First item\n* Second item\n  - Nested item\n"
	got := FormatADR(input, FormatOptions{NoColor: true})
	assert.Contains(t, got, "• First item")
	assert.Contains(t, got, "• Second item")
	assert.Contains(t, got, "  • Nested item")
	assert.NotContains(t, got, "- First")
	assert.NotContains(t, got, "* Second")
}

func TestFormatADR_LinkFormatting(t *testing.T) {
	orig := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = orig }()

	input := "See [ADR-0001](0001-use-go.md) for details.\n"
	got := FormatADR(input, FormatOptions{NoColor: false})
	assert.Contains(t, got, "ADR-0001")
	assert.Contains(t, got, "0001-use-go.md")
}

func TestFormatADR_FullNygardPlain(t *testing.T) {
	input := `# 1. Use Go

Date: 2024-01-01

## Status

Accepted

## Context

We need a programming language.

## Decision

We will use Go.

## Consequences

- Fast compilation
- Good concurrency
`
	got := FormatADR(input, FormatOptions{NoColor: true})
	assert.Contains(t, got, "# 1. Use Go")
	assert.Contains(t, got, "Date: 2024-01-01")
	assert.Contains(t, got, "## Status")
	assert.Contains(t, got, "Accepted")
	assert.Contains(t, got, "## Context")
	assert.Contains(t, got, "## Decision")
	assert.Contains(t, got, "## Consequences")
	assert.Contains(t, got, "• Fast compilation")
	assert.Contains(t, got, "• Good concurrency")
}

func TestFormatADR_FullMADRFrontmatterPlain(t *testing.T) {
	input := `---
status: "proposed"
date: 2024-01-01
---

# 1. Use Go

## Context and Problem Statement

We need a language.

## Decision Drivers

- Performance
- Simplicity

## Considered Options

- Go
- Rust

## Decision Outcome

Chosen option: Go
`
	got := FormatADR(input, FormatOptions{NoColor: true})
	assert.Contains(t, got, "---")
	assert.Contains(t, got, "status: \"proposed\"")
	assert.Contains(t, got, "# 1. Use Go")
	assert.Contains(t, got, "## Decision Drivers")
	assert.Contains(t, got, "• Performance")
	assert.Contains(t, got, "• Go")
}

// stripANSI removes ANSI escape codes from a string.
func stripANSI(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Skip to 'm'
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			i = j + 1
			continue
		}
		result.WriteByte(s[i])
		i++
	}
	return result.String()
}
