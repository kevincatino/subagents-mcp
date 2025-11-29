package runner

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"go.uber.org/zap"

	"subagents-mcp/internal/agents"
)

func TestCopilotRunnerValidation(t *testing.T) {
	r := NewCopilotRunner(zap.NewNop())
	if _, err := r.Run(context.Background(), agents.Agent{Name: "a", Persona: "p", Description: "d"}, "", "/tmp"); err == nil {
		t.Fatal("expected error for empty task")
	}
	if _, err := r.Run(context.Background(), agents.Agent{Name: "a", Persona: "p", Description: "d"}, "task", "relative"); err == nil {
		t.Fatal("expected error for relative workdir")
	}
}

func TestCopilotRunnerBuildsCommand(t *testing.T) {
	logger := zap.NewNop()
	r := NewCopilotRunner(logger)

	dir := t.TempDir()
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("eval symlinks: %v", err)
	}

	var gotName string
	var gotArgs []string
	var cmdRef *exec.Cmd
	r.execCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		gotName = name
		gotArgs = append([]string(nil), arg...)
		cmdRef = exec.CommandContext(ctx, "echo", "ok")
		return cmdRef
	}

	out, err := r.Run(context.Background(), agents.Agent{Name: "agent", Persona: "p", Description: "d"}, "do something", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "ok" {
		t.Fatalf("expected ok, got %q", out)
	}
	if gotName != "copilot" {
		t.Fatalf("expected copilot command, got %s", gotName)
	}
	expected := []string{"-p", "p\n\nTask: do something", "--allow-all-tools", "--allow-all-paths", "--stream", "off"}
	if len(gotArgs) != len(expected) {
		t.Fatalf("expected args %v got %v", expected, gotArgs)
	}
	for i := range expected {
		if gotArgs[i] != expected[i] {
			t.Fatalf("arg %d mismatch: expected %s got %s", i, expected[i], gotArgs[i])
		}
	}
	if cmdRef == nil {
		t.Fatal("expected command reference")
	}
	if cmdRef.Dir != resolvedDir {
		t.Fatalf("expected workdir %s got %s", resolvedDir, cmdRef.Dir)
	}
}
