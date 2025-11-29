package agents

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAMLRepository loads agents from YAML files in a directory.
type YAMLRepository struct {
	baseDir string
}

func NewYAMLRepository(baseDir string) *YAMLRepository {
	return &YAMLRepository{baseDir: baseDir}
}

func (r *YAMLRepository) ListAgents(ctx context.Context) ([]Agent, error) {
	entries, err := os.ReadDir(r.baseDir)
	if err != nil {
		return nil, fmt.Errorf("read agents dir: %w", err)
	}

	var agentsList []Agent
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".yaml")
		content, err := os.ReadFile(filepath.Join(r.baseDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}

		var raw struct {
			Persona     string `yaml:"persona"`
			Description string `yaml:"description"`
			Model       string `yaml:"model"`
		}
		if err := yaml.Unmarshal(content, &raw); err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}

		agent := Agent{
			Name:        name,
			Persona:     strings.TrimSpace(raw.Persona),
			Description: strings.TrimSpace(raw.Description),
			Model:       strings.TrimSpace(raw.Model),
		}
		if err := agent.Validate(); err != nil {
			return nil, fmt.Errorf("validate %s: %w", entry.Name(), err)
		}
		agentsList = append(agentsList, agent)
	}

	return agentsList, nil
}
