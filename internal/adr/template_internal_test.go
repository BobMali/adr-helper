package adr

import "testing"

// TestMetaFieldKeys_NoConflictingDuplicates guards against a future template
// introducing the same metadata-field Key with a different Kind or Heading.
// AllMetaFieldDefs dedupes by Key and would silently pick an arbitrary winner, so
// this white-box test inspects the raw registry before dedup and fails loudly.
func TestMetaFieldKeys_NoConflictingDuplicates(t *testing.T) {
	seen := make(map[string]TemplateSectionDef)
	for _, name := range ValidTemplateNames() {
		for _, s := range templateSections[TemplateName(name)] {
			if s.Kind != "meta" && s.Kind != "frontmatter" {
				continue
			}
			if prev, ok := seen[s.Key]; ok {
				if prev.Kind != s.Kind {
					t.Errorf("metadata key %q used with differing Kind: %q vs %q", s.Key, prev.Kind, s.Kind)
				}
				if prev.Heading != s.Heading {
					t.Errorf("metadata key %q used with differing Heading: %q vs %q", s.Key, prev.Heading, s.Heading)
				}
			}
			seen[s.Key] = s
		}
	}
}
