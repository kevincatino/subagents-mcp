package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("loads and trims models", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		content := `
runners:
  - name: codex
    priority: 1
    models: [" gpt-4o ", "gpt-4o-mini", " "]
  - name: copilot
    priority: 2
    models:
      - claude
      -  anthropic/opus
`
		writeFile(t, path, content)

		cfg, err := LoadConfig(path)
		if err != nil {
			t.Fatalf("LoadConfig error: %v", err)
		}
		if len(cfg.Runners) != 2 {
			t.Fatalf("expected 2 runners, got %d", len(cfg.Runners))
		}
		if cfg.Runners[0].Models[0] != "gpt-4o" || cfg.Runners[0].Models[1] != "gpt-4o-mini" {
			t.Fatalf("unexpected models: %+v", cfg.Runners[0].Models)
		}
		if cfg.Runners[1].Models[1] != "anthropic/opus" {
			t.Fatalf("expected trimmed model, got %q", cfg.Runners[1].Models[1])
		}
	})

	t.Run("errors on missing name", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		writeFile(t, path, "runners:\n  - priority: 1\n")

		if _, err := LoadConfig(path); err == nil {
			t.Fatal("expected error for missing name")
		}
	})

	t.Run("errors on non-positive priority", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		writeFile(t, path, "runners:\n  - name: codex\n    priority: 0\n")

		if _, err := LoadConfig(path); err == nil {
			t.Fatal("expected error for invalid priority")
		}
	})
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
