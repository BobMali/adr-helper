package adr

import (
	"regexp"
	"strings"
)

// metaFieldExtractor holds a precompiled matcher for one metadata field. Compiling
// once (package init) matters because ExtractMetaFields runs over every ADR on the
// List() hot path, once per field.
type metaFieldExtractor struct {
	key  string
	kind string         // "meta" (title-block line) or "frontmatter" (YAML key)
	re   *regexp.Regexp // capture group 1 holds the raw value
}

var metaFieldExtractors = buildMetaFieldExtractors()

func buildMetaFieldExtractors() []metaFieldExtractor {
	out := make([]metaFieldExtractor, 0, len(allMetaFieldDefs))
	for _, d := range allMetaFieldDefs {
		var re *regexp.Regexp
		switch d.Kind {
		case "meta":
			// Title-block line "Heading: value" (case-insensitive), matched on the
			// friendly Heading — the label the app writes via ReplaceMetaField.
			re = metaFieldPattern(d.Heading)
		case "frontmatter":
			// YAML "key: value" matched on the exact lowercase Key. Frontmatter is
			// single-line only: multi-line block ("- Alice") or flow ("[Alice, Bob]")
			// YAML lists are not supported (regex-based parse, no YAML dependency).
			re = regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(d.Key) + `:[ \t]*(.*)$`)
		default:
			continue
		}
		out = append(out, metaFieldExtractor{key: d.Key, kind: d.Kind, re: re})
	}
	return out
}

// ExtractMetaFields parses recognized metadata fields (see AllMetaFieldDefs) from an
// ADR's raw content, returning field key -> trimmed, comma-split values. Fields with
// no value are omitted; the result is nil when nothing is found. Title-block ("meta")
// fields are read from the body (frontmatter skipped); "frontmatter" fields from the
// YAML block. Unfilled template placeholders like "{list everyone…}" are dropped.
func ExtractMetaFields(content string) map[string][]string {
	body := bodyAfterFrontmatter(content)
	fm := extractFrontmatter(content)

	var result map[string][]string
	for _, ex := range metaFieldExtractors {
		var raw string
		switch ex.kind {
		case "meta":
			m := ex.re.FindStringSubmatch(body)
			if m == nil {
				continue
			}
			raw = m[1]
		case "frontmatter":
			if fm == "" {
				continue
			}
			m := ex.re.FindStringSubmatch(fm)
			if m == nil {
				continue
			}
			raw = stripQuotes(strings.TrimSpace(m[1]))
		}

		values := splitMetaValue(raw)
		if len(values) == 0 {
			continue
		}
		if result == nil {
			result = make(map[string][]string)
		}
		result[ex.key] = values
	}
	return result
}

// splitMetaValue trims a raw field value, drops it entirely if it is a single "{…}"
// placeholder (checked BEFORE splitting so a placeholder containing a comma isn't
// mis-split), then comma-splits, trimming each token and dropping empties and any
// token that is itself a "{…}" placeholder (e.g. a partially-edited "Alice, {TBD}").
func splitMetaValue(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || isPlaceholder(raw) {
		return nil
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		p := strings.TrimSpace(part)
		if p == "" || isPlaceholder(p) {
			continue
		}
		out = append(out, p)
	}
	return out
}

// isPlaceholder reports whether s is a single brace-wrapped template placeholder.
func isPlaceholder(s string) bool {
	return len(s) >= 2 && s[0] == '{' && s[len(s)-1] == '}'
}
