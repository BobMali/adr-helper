package adr

import (
	"embed"
	"fmt"
)

// TemplateName identifies a supported ADR template format.
type TemplateName string

const (
	TemplateNygard      TemplateName = "nygard"
	TemplateMADRMinimal TemplateName = "madr-minimal"
	TemplateMADRFull    TemplateName = "madr-full"
)

//go:embed templates/*.md
var templateFS embed.FS

var templateFiles = map[TemplateName]string{
	TemplateNygard:      "templates/nygard.md",
	TemplateMADRMinimal: "templates/madr-minimal.md",
	TemplateMADRFull:    "templates/madr-full.md",
}

// TemplateSectionDef describes a user-editable section within an ADR template.
type TemplateSectionDef struct {
	Key         string `json:"key"`
	Heading     string `json:"heading"`
	Kind        string `json:"kind"` // "h2" or "h3"
	Optional    bool   `json:"optional"`
	Placeholder string `json:"placeholder"`
}

var templateSections = map[TemplateName][]TemplateSectionDef{
	TemplateNygard: {
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
// for the named template. The Status section is excluded (server-managed).
func TemplateSections(name string) ([]TemplateSectionDef, error) {
	tn := TemplateName(name)
	sections, ok := templateSections[tn]
	if !ok {
		return nil, fmt.Errorf("unknown template %q", name)
	}
	// Return a copy to prevent mutation
	result := make([]TemplateSectionDef, len(sections))
	copy(result, sections)
	return result, nil
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
		string(TemplateMADRMinimal),
		string(TemplateMADRFull),
	}
}
