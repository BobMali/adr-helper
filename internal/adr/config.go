package adr

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ConfigFileName = ".adr.json"
	ConfigVersion  = "1"
)

var (
	ErrConfigNotFound = errors.New("config not found")
	ErrConfigInvalid  = errors.New("config invalid")
)

// Config represents the ADR project configuration stored in .adr.json.
type Config struct {
	Version   string `json:"version"`
	Directory string `json:"directory"`
	Template  string `json:"template"`
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

	return &cfg, nil
}
