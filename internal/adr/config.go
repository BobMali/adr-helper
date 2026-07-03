package adr

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	ConfigFileName = ".adr.json"
	ConfigVersion  = "1"

	// MaxScopeLength is the maximum number of runes allowed in a scope value.
	MaxScopeLength = 64
)

var (
	ErrConfigNotFound = errors.New("config not found")
	ErrConfigInvalid  = errors.New("config invalid")
	// ErrInvalidScope is returned when a scope value fails validation.
	ErrInvalidScope = errors.New("invalid scope")
)

// Config represents the ADR project configuration stored in .adr.json.
type Config struct {
	Version      string   `json:"version"`
	Directory    string   `json:"directory"`
	Template     string   `json:"template"`
	TemplateFile string   `json:"templateFile"`
	Scopes       []string `json:"scopes,omitempty"`
}

// HasScope reports whether value matches an existing scope, case-insensitively.
// It returns the canonical stored spelling and true when found.
func (c *Config) HasScope(value string) (string, bool) {
	needle := strings.ToLower(strings.TrimSpace(value))
	for _, s := range c.Scopes {
		if strings.ToLower(s) == needle {
			return s, true
		}
	}
	return "", false
}

// AddScope validates value and appends it to the vocabulary unless an equal
// scope already exists (case-insensitive), in which case it is a no-op. It
// returns a copy of the updated scope list. Invalid values return ErrInvalidScope.
func (c *Config) AddScope(value string) ([]string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, fmt.Errorf("scope must not be empty: %w", ErrInvalidScope)
	}
	if strings.ContainsRune(trimmed, ',') {
		return nil, fmt.Errorf("scope %q must not contain commas: %w", trimmed, ErrInvalidScope)
	}
	for _, r := range trimmed {
		if r < 0x20 || r == 0x7f {
			return nil, fmt.Errorf("scope must not contain control characters: %w", ErrInvalidScope)
		}
	}
	if utf8.RuneCountInString(trimmed) > MaxScopeLength {
		return nil, fmt.Errorf("scope must be at most %d characters: %w", MaxScopeLength, ErrInvalidScope)
	}

	if _, ok := c.HasScope(trimmed); !ok {
		c.Scopes = append(c.Scopes, trimmed)
	}
	return append([]string(nil), c.Scopes...), nil
}

// SaveConfig writes the config as indented JSON to dir/.adr.json.
func SaveConfig(dir string, cfg *Config) error {
	out := *cfg
	out.Version = ConfigVersion

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	data = append(data, '\n')

	return os.WriteFile(filepath.Join(dir, ConfigFileName), data, 0o644)
}

// LoadConfig reads and validates the config from dir/.adr.json.
func LoadConfig(dir string) (*Config, error) {
	path := filepath.Join(dir, ConfigFileName)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("reading %s: %w", path, ErrConfigNotFound)
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, ErrConfigInvalid)
	}

	if cfg.Version != ConfigVersion {
		return nil, fmt.Errorf("unsupported config version %q: %w", cfg.Version, ErrConfigInvalid)
	}
	if cfg.Directory == "" {
		return nil, fmt.Errorf("directory must not be empty: %w", ErrConfigInvalid)
	}
	if cfg.TemplateFile == "" {
		cfg.TemplateFile = "template.md"
	}

	return &cfg, nil
}
