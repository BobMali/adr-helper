package adr

import (
	"embed"
	"fmt"
)

// TemplateName identifies a supported ADR template format.
type TemplateName string

const (
	TemplateNygard       TemplateName = "nygard"
	TemplateNygardScoped TemplateName = "nygard-scoped"
	TemplateMADRMinimal  TemplateName = "madr-minimal"
	TemplateMADRFull     TemplateName = "madr-full"
)

//go:embed templates/*.md
var templateFS embed.FS

var templateFiles = map[TemplateName]string{
	TemplateNygard:       "templates/nygard.md",
	TemplateNygardScoped: "templates/nygard-scoped.md",
	TemplateMADRMinimal:  "templates/madr-minimal.md",
	TemplateMADRFull:     "templates/madr-full.md",
}

// TemplateSectionDef describes a user-editable section within an ADR template.
type TemplateSectionDef struct {
	Key         string `json:"key"`
	Heading     string `json:"heading"`
	Kind        string `json:"kind"` // "h2", "h3", "meta" (title-block line), or "frontmatter" (YAML key)
	Optional    bool   `json:"optional"`
	Placeholder string `json:"placeholder"`
	// Vocabulary indicates the field is filled from a project-managed list of
	// selectable values (Config.Scopes) rather than free text.
	Vocabulary bool `json:"vocabulary,omitempty"`
}

var templateSections = map[TemplateName][]TemplateSectionDef{
	TemplateNygard: {
		{Key: "context", Heading: "Context", Kind: "h2", Optional: false, Placeholder: "What is the issue that we're seeing that is motivating this decision or change?"},
		{Key: "decision", Heading: "Decision", Kind: "h2", Optional: false, Placeholder: "What is the change that we're proposing and/or doing?"},
		{Key: "consequences", Heading: "Consequences", Kind: "h2", Optional: false, Placeholder: "What becomes easier or more difficult to do because of this change?"},
	},
	TemplateNygardScoped: {
		{Key: "scope", Heading: "Scope", Kind: "meta", Optional: false, Vocabulary: true, Placeholder: "Which part(s) of the system this decision applies to"},
		{Key: "context", Heading: "Context", Kind: "h2", Optional: false, Placeholder: "What is the issue that we're seeing that is motivating this decision or change?"},
		{Key: "decision", Heading: "Decision", Kind: "h2", Optional: false, Placeholder: "What is the change that we're proposing and/or doing?"},
		{Key: "consequences", Heading: "Consequences", Kind: "h2", Optional: false, Placeholder: "What becomes easier or more difficult to do because of this change?"},
	},
	TemplateMADRMinimal: {
		{Key: "context-and-problem-statement", Heading: "Context and Problem Statement", Kind: "h2", Optional: false, Placeholder: "Describe the context and problem statement, e.g., in free form using two to three sentences or in the form of an illustrative story."},
		{Key: "considered-options", Heading: "Considered Options", Kind: "h2", Optional: false, Placeholder: "* Option 1\n* Option 2\n* Option 3"},
		{Key: "decision-outcome", Heading: "Decision Outcome", Kind: "h2", Optional: false, Placeholder: "Chosen option: \"{title of option}\", because {justification}."},
		{Key: "consequences", Heading: "Consequences", Kind: "h3", Optional: true, Placeholder: "* Good, because {positive consequence}\n* Bad, because {negative consequence}"},
	},
	TemplateMADRFull: {
		// Frontmatter metadata fields. Key is the exact lowercase YAML key as it
		// appears in the file; getFrontmatterField matches on Key, not Heading.
		// Excluded from the create form (see TemplateSections) — filter facets only.
		{Key: "decision-makers", Heading: "Decision Makers", Kind: "frontmatter", Optional: true, Placeholder: "list everyone involved in the decision"},
		{Key: "consulted", Heading: "Consulted", Kind: "frontmatter", Optional: true, Placeholder: "list everyone whose opinions are sought"},
		{Key: "informed", Heading: "Informed", Kind: "frontmatter", Optional: true, Placeholder: "list everyone who is kept up-to-date on progress"},
		{Key: "context-and-problem-statement", Heading: "Context and Problem Statement", Kind: "h2", Optional: false, Placeholder: "Describe the context and problem statement, e.g., in free form using two to three sentences or in the form of an illustrative story."},
		{Key: "decision-drivers", Heading: "Decision Drivers", Kind: "h2", Optional: true, Placeholder: "* Decision driver 1, e.g., a force, facing concern, …\n* Decision driver 2, e.g., a force, facing concern, …"},
		{Key: "considered-options", Heading: "Considered Options", Kind: "h2", Optional: false, Placeholder: "* Option 1\n* Option 2\n* Option 3"},
		{Key: "decision-outcome", Heading: "Decision Outcome", Kind: "h2", Optional: false, Placeholder: "Chosen option: \"{title of option}\", because {justification}."},
		{Key: "consequences", Heading: "Consequences", Kind: "h3", Optional: true, Placeholder: "* Good, because {positive consequence}\n* Bad, because {negative consequence}"},
		{Key: "confirmation", Heading: "Confirmation", Kind: "h3", Optional: true, Placeholder: "Describe how the implementation of/compliance with the ADR can/will be confirmed."},
		{Key: "pros-and-cons-of-the-options", Heading: "Pros and Cons of the Options", Kind: "h2", Optional: true, Placeholder: "### Option 1\n\n* Good, because {argument a}\n* Bad, because {argument d}"},
		{Key: "more-information", Heading: "More Information", Kind: "h2", Optional: true, Placeholder: "Provide additional evidence/confidence for the decision outcome here."},
	},
}

// TemplateSections returns the ordered list of user-editable section definitions
// for the named template. The Status section is excluded (server-managed), and so
// are "frontmatter" fields — the create form only knows how to write "meta"
// title-block lines and "h2"/"h3" body sections, so a frontmatter def leaking
// through would be mishandled (see handleCreateADR). Frontmatter fields are surfaced
// only as filter facets via AllMetaFieldDefs.
func TemplateSections(name string) ([]TemplateSectionDef, error) {
	tn := TemplateName(name)
	sections, ok := templateSections[tn]
	if !ok {
		return nil, fmt.Errorf("unknown template %q", name)
	}
	result := make([]TemplateSectionDef, 0, len(sections))
	for _, s := range sections {
		if s.Kind == "frontmatter" {
			continue
		}
		result = append(result, s)
	}
	return result, nil
}

// allMetaFieldDefs is the deduped union of every metadata field (Kind "meta" or
// "frontmatter") across all templates, in ValidTemplateNames order. It is the single
// source for metadata extraction and the filter-facet API. Built once: templateSections
// is a map (nondeterministic iteration), so we iterate the fixed template-name order to
// keep facet/JSON ordering stable.
var allMetaFieldDefs = buildAllMetaFieldDefs()

func buildAllMetaFieldDefs() []TemplateSectionDef {
	var result []TemplateSectionDef
	seen := make(map[string]bool)
	for _, name := range ValidTemplateNames() {
		for _, s := range templateSections[TemplateName(name)] {
			if s.Kind != "meta" && s.Kind != "frontmatter" {
				continue
			}
			if seen[s.Key] {
				continue
			}
			seen[s.Key] = true
			result = append(result, s)
		}
	}
	return result
}

// AllMetaFieldDefs returns the deduped union of metadata field definitions across all
// templates (Kind "meta" or "frontmatter"), in a stable order. These are the fields
// carried in ADR.Meta and exposed as filter facets.
func AllMetaFieldDefs() []TemplateSectionDef {
	result := make([]TemplateSectionDef, len(allMetaFieldDefs))
	copy(result, allMetaFieldDefs)
	return result
}

// TemplateContent returns the content of the named template.
func TemplateContent(name string) (string, error) {
	tn := TemplateName(name)
	path, ok := templateFiles[tn]
	if !ok {
		return "", fmt.Errorf("unknown template %q, valid templates: %v", name, ValidTemplateNames())
	}
	data, err := templateFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading template %q: %w", name, err)
	}
	return string(data), nil
}

// ValidTemplateNames returns the list of supported template names.
func ValidTemplateNames() []string {
	return []string{
		string(TemplateNygard),
		string(TemplateNygardScoped),
		string(TemplateMADRMinimal),
		string(TemplateMADRFull),
	}
}
