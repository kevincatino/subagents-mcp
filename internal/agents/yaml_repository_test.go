package agents

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestYAMLRepository_ListAgents(t *testing.T) {
	t.Run("loads valid agents", func(t *testing.T) {
		dir := t.TempDir()
		write(t, filepath.Join(dir, "alpha.yaml"), "persona: alpha\ndescription: first agent\nmodel: gpt-4o\n")
		write(t, filepath.Join(dir, "beta.yaml"), "persona: beta persona\ndescription: second agent\n")

		repo := NewYAMLRepository(dir)
		agents, err := repo.ListAgents(context.Background())
		if err != nil {
			t.Fatalf("ListAgents error: %v", err)
		}
		if len(agents) != 2 {
			t.Fatalf("expected 2 agents, got %d", len(agents))
		}
		if agents[0].Model != "gpt-4o" {
			t.Fatalf("expected model to be set, got %q", agents[0].Model)
		}
		if agents[1].Model != "" {
			t.Fatalf("expected empty model, got %q", agents[1].Model)
		}
	})

	t.Run("errors on missing required fields", func(t *testing.T) {
		dir := t.TempDir()
		write(t, filepath.Join(dir, "alpha.yaml"), "persona: \ndescription: missing persona\n")

		repo := NewYAMLRepository(dir)
		if _, err := repo.ListAgents(context.Background()); err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		dir := t.TempDir()
		write(t, filepath.Join(dir, "alpha.yaml"), "persona:  alpha  \ndescription:  desc  \nmodel:  gpt-4o  \n")

		repo := NewYAMLRepository(dir)
		agents, err := repo.ListAgents(context.Background())
		if err != nil {
			t.Fatalf("ListAgents error: %v", err)
		}
		if agents[0].Persona != "alpha" {
			t.Fatalf("expected trimmed persona, got %q", agents[0].Persona)
		}
		if agents[0].Description != "desc" {
			t.Fatalf("expected trimmed description, got %q", agents[0].Description)
		}
		if agents[0].Model != "gpt-4o" {
			t.Fatalf("expected trimmed model, got %q", agents[0].Model)
		}
	})
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
