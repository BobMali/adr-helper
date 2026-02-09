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
