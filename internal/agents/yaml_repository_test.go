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
		write(t, filepath.Join(dir, "alpha.yaml"), "persona: alpha\ndescription: first agent\n")
		write(t, filepath.Join(dir, "beta.yaml"), "persona: beta persona\ndescription: second agent\n")

		repo := NewYAMLRepository(dir)
		agents, err := repo.ListAgents(context.Background())
		if err != nil {
			t.Fatalf("ListAgents error: %v", err)
		}
		if len(agents) != 2 {
			t.Fatalf("expected 2 agents, got %d", len(agents))
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
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
