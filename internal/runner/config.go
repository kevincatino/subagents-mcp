package runner

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config describes available runners, their priorities, and supported models.
type Config struct {
	Runners []RunnerConfig `yaml:"runners"`
}

// RunnerConfig represents a single runner entry loaded from YAML.
type RunnerConfig struct {
	Name     string   `yaml:"name"`
	Priority int      `yaml:"priority"`
	Models   []string `yaml:"models"`
}

// LoadConfig reads runner configuration from a YAML file and validates it.
func LoadConfig(path string) (Config, error) {
	if path == "" {
		return Config{}, fmt.Errorf("config path is required")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read runner config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse runner config: %w", err)
	}

	if err := cfg.validateAndNormalize(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c *Config) validateAndNormalize() error {
	for i := range c.Runners {
		c.Runners[i].Name = strings.TrimSpace(c.Runners[i].Name)
		if c.Runners[i].Name == "" {
			return fmt.Errorf("runner name is required")
		}
		if c.Runners[i].Priority <= 0 {
			return fmt.Errorf("runner %q priority must be greater than zero", c.Runners[i].Name)
		}

		var models []string
		for _, m := range c.Runners[i].Models {
			trimmed := strings.TrimSpace(m)
			if trimmed != "" {
				models = append(models, trimmed)
			}
		}
		c.Runners[i].Models = models
	}
	return nil
}
