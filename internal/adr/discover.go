package adr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExtractScope returns the value of the first title-block "Scope:" line
// (case-insensitive), trimmed. YAML frontmatter is skipped so that only a body
// title-block line — the form the app itself writes via ReplaceMetaField — is
// discovered. The bool is false when no Scope line exists.
func ExtractScope(content string) (string, bool) {
	m := metaFieldPattern("Scope").FindStringSubmatch(bodyAfterFrontmatter(content))
	if m == nil {
		return "", false
	}
	return strings.TrimSpace(m[1]), true
}

// DiscoverScopes scans dir for ADR files and returns every scope token found in
// their "Scope:" lines, comma-split and trimmed. The result is RAW: it preserves
// file+line order, keeps duplicates, and is NOT validated — callers must pass it
// through Config.MergeScopes before use. A missing directory yields (nil, nil)
// rather than an error; unreadable individual files are skipped.
func DiscoverScopes(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading directory %q: %w", dir, err)
	}

	var tokens []string
	for _, entry := range entries {
		if entry.IsDir() || !IsADRFilename(entry.Name()) {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue // best-effort: skip files we can't read
		}
		value, ok := ExtractScope(string(content))
		if !ok || value == "" {
			continue
		}
		for _, part := range strings.Split(value, ",") {
			if p := strings.TrimSpace(part); p != "" {
				tokens = append(tokens, p)
			}
		}
	}
	return tokens, nil
}

// MergeScopes adds each value to the vocabulary, reusing AddScope/HasScope for
// validation and case-insensitive dedup. It returns the canonical spellings
// newly added and the values skipped for failing validation (deduped
// case-insensitively). Existing scopes and case-insensitive duplicates are
// silently ignored. Only in-memory state is mutated; the caller persists.
func (c *Config) MergeScopes(values []string) (added, invalid []string) {
	seenInvalid := make(map[string]bool)
	for _, v := range values {
		if _, ok := c.HasScope(v); ok {
			continue // already present (case-insensitive) — silent
		}
		if _, err := c.AddScope(v); err != nil {
			key := strings.ToLower(strings.TrimSpace(v))
			if !seenInvalid[key] {
				seenInvalid[key] = true
				invalid = append(invalid, strings.TrimSpace(v))
			}
			continue
		}
		added = append(added, strings.TrimSpace(v))
	}
	return added, invalid
}

// DiscoverAndMergeScopes scans cfg.Directory for scopes and merges them into cfg
// in memory (no persistence). It is the single entry point used by `adr init`,
// `adr scope discover`, and the web server at boot, so the discovery behavior
// can't drift between call sites.
func DiscoverAndMergeScopes(cfg *Config) (added, invalid []string, err error) {
	raw, err := DiscoverScopes(cfg.Directory)
	if err != nil {
		return nil, nil, err
	}
	added, invalid = cfg.MergeScopes(raw)
	return added, invalid, nil
}
