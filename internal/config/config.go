package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds user configurable options.
type Config struct {
	Prompt string `yaml:"prompt"`
}

// Load merges global and repo config files.
func Load(repoRoot string) (*Config, error) {
	cfg := &Config{}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	paths := []string{
		filepath.Join(home, ".config", "gh", "pr-inbox.yml"),
		filepath.Join(repoRoot, ".github", "pr-inbox.yml"),
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			if err := applyFile(cfg, path); err != nil {
				return nil, err
			}
		}
	}
	return cfg, nil
}

func applyFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, cfg)
}
