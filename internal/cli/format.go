package cli

import (
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/malek/adr-helper/internal/adr"
)

// FormatOptions controls how FormatADR renders content.
type FormatOptions struct {
	NoColor bool
}

var linkPattern = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

// FormatADR formats ADR markdown content with terminal colors.
func FormatADR(content string, opts FormatOptions) string {
	h1Style := color.New(color.FgCyan, color.Bold)
	h2Style := color.New(color.Bold)
	h3Style := color.New(color.Bold)
	dimStyle := color.New(color.Faint)
	greenStyle := color.New(color.FgGreen)
	yellowStyle := color.New(color.FgYellow)
	redStyle := color.New(color.FgRed)
	linkLabelStyle := color.New(color.FgBlue, color.Underline)
	linkTargetStyle := color.New(color.Faint)

	styles := []*color.Color{
		h1Style, h2Style, h3Style, dimStyle,
		greenStyle, yellowStyle, redStyle,
		linkLabelStyle, linkTargetStyle,
	}

	if opts.NoColor {
		for _, s := range styles {
			s.DisableColor()
		}
	} else {
		for _, s := range styles {
			s.EnableColor()
		}
	}

	lines := strings.Split(content, "\n")
	var result strings.Builder
	inFrontmatter := false
	frontmatterCount := 0
	afterStatusHeading := false

	for i, line := range lines {
		if i > 0 {
			result.WriteByte('\n')
		}

		trimmed := strings.TrimSpace(line)

		// Frontmatter delimiter
		if trimmed == "---" {
			frontmatterCount++
			if frontmatterCount <= 2 {
				inFrontmatter = frontmatterCount == 1
			}
			result.WriteString(dimStyle.Sprint(line))
			continue
		}

		// Inside frontmatter
		if inFrontmatter {
			result.WriteString(dimStyle.Sprint(line))
			continue
		}

		// H1
		if strings.HasPrefix(trimmed, "# ") {
			afterStatusHeading = false
			result.WriteString(h1Style.Sprint(line))
			continue
		}

		// H2
		if strings.HasPrefix(trimmed, "## ") {
			heading := strings.TrimPrefix(trimmed, "## ")
			afterStatusHeading = strings.EqualFold(heading, "Status")
			result.WriteString(h2Style.Sprint(line))
			continue
		}

		// H3
		if strings.HasPrefix(trimmed, "### ") {
			result.WriteString(h3Style.Sprint(line))
			continue
		}

		// Date line
		if strings.HasPrefix(trimmed, "Date:") || strings.HasPrefix(trimmed, "date:") {
			result.WriteString(dimStyle.Sprint(line))
			continue
		}

		// Status value line
		if afterStatusHeading && trimmed != "" {
			result.WriteString(statusColor(trimmed, greenStyle, yellowStyle, redStyle))
			continue
		}

		// Bullet lines
		if isBulletLine(trimmed) {
			result.WriteString(formatBullet(line, opts, linkLabelStyle, linkTargetStyle))
			continue
		}

		// Default: process inline links
		result.WriteString(formatInlineLinks(line, linkLabelStyle, linkTargetStyle))
	}

	return result.String()
}

func statusColor(text string, green, yellow, red *color.Color) string {
	status, ok := adr.ParseStatus(text)
	if !ok {
		return text
	}
	switch status.Category() {
	case adr.StatusCategoryActive:
		return green.Sprint(text)
	case adr.StatusCategoryPending:
		return yellow.Sprint(text)
	case adr.StatusCategoryInactive:
		return red.Sprint(text)
	}
	return text
}

func isBulletLine(trimmed string) bool {
	return strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ")
}

func formatBullet(line string, opts FormatOptions, linkLabel, linkTarget *color.Color) string {
	indent := len(line) - len(strings.TrimLeft(line, " \t"))
	trimmed := strings.TrimLeft(line, " \t")
	// Replace "- " or "* " with "• "
	rest := trimmed[2:]
	formatted := line[:indent] + "• " + formatInlineLinks(rest, linkLabel, linkTarget)
	return formatted
}

func formatInlineLinks(line string, labelStyle, targetStyle *color.Color) string {
	return linkPattern.ReplaceAllStringFunc(line, func(match string) string {
		subs := linkPattern.FindStringSubmatch(match)
		if len(subs) != 3 {
			return match
		}
		return labelStyle.Sprint(subs[1]) + targetStyle.Sprint("("+subs[2]+")")
	})
}
