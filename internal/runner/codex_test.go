package runner

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"testing"

	"go.uber.org/zap"

	"subagents-mcp/internal/agents"
)

func TestCodexRunnerValidation(t *testing.T) {
	logger := zap.NewNop()
	r := NewCodexRunner(logger, []string{"gpt-4o"})

	if _, err := r.Run(context.Background(), agents.Agent{Name: "a", Persona: "p", Description: "d"}, "", "/tmp", "gpt-4o"); err == nil {
		t.Fatal("expected error for empty task")
	}

	if _, err := r.Run(context.Background(), agents.Agent{Name: "a", Persona: "p", Description: "d"}, "task", "relative", "gpt-4o"); err == nil {
		t.Fatal("expected error for relative path")
	}

	if _, err := r.Run(context.Background(), agents.Agent{Name: "a", Persona: "p", Description: "d"}, "task", "/tmp", "other"); err == nil {
		t.Fatal("expected error for unsupported model")
	}
}

func TestCodexRunnerBuildsCommand(t *testing.T) {
	logger := zap.NewNop()
	r := NewCodexRunner(logger, nil)

	dir := t.TempDir()
	var gotName string
	var gotArgs []string
	r.execCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		gotName = name
		gotArgs = append([]string(nil), arg...)
		return exec.CommandContext(ctx, "echo", "ok")
	}

	out, err := r.Run(context.Background(), agents.Agent{Name: "agent", Persona: "p", Description: "d"}, "do something", dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "ok" {
		t.Fatalf("expected trimmed stdout 'ok', got %q", out)
	}
	if gotName != "codex" {
		t.Fatalf("expected command name codex, got %s", gotName)
	}
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("eval symlinks: %v", err)
	}
	expected := []string{"--cd", resolvedDir, "--sandbox", "read-only", "--ask-for-approval", "never", "exec", "--skip-git-repo-check", "p\n\nTask: do something"}
	if len(gotArgs) != len(expected) {
		t.Fatalf("expected args %v got %v", expected, gotArgs)
	}
	for i := range expected {
		if gotArgs[i] != expected[i] {
			t.Fatalf("arg %d mismatch: expected %s got %s", i, expected[i], gotArgs[i])
		}
	}
}

func TestCodexRunnerIncludesModelFlag(t *testing.T) {
	logger := zap.NewNop()
	r := NewCodexRunner(logger, []string{"gpt-5"})

	dir := t.TempDir()
	var gotArgs []string
	r.execCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		gotArgs = append([]string(nil), arg...)
		return exec.CommandContext(ctx, "echo", "ok")
	}

	if _, err := r.Run(context.Background(), agents.Agent{Name: "agent", Persona: "p", Description: "d"}, "task", dir, "gpt-5"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for i := 0; i < len(gotArgs)-3; i++ {
		if gotArgs[i] == "exec" && gotArgs[i+1] == "--skip-git-repo-check" && gotArgs[i+2] == "--model" && gotArgs[i+3] == "gpt-5" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected --model flag after exec, got args: %v", gotArgs)
	}
}

func TestCodexRunner_UsageLimitDetection(t *testing.T) {
	logger := zap.NewNop()
	r := NewCodexRunner(logger, nil)

	dir := t.TempDir()
	r.execCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		// Simulate a command that exits with error and outputs usage limit message
		cmd := exec.CommandContext(ctx, "sh", "-c", `echo '{"type":"error","message":"You'\''ve hit your usage limit. Upgrade to Pro"}' >&2; exit 1`)
		return cmd
	}

	_, err := r.Run(context.Background(), agents.Agent{Name: "agent", Persona: "p", Description: "d"}, "do something", dir, "")
	if err == nil {
		t.Fatal("expected error")
	}

	if !IsUsageLimitError(err) {
		t.Fatalf("expected ErrUsageLimitExceeded, got: %v", err)
	}

	var usageErr *ErrUsageLimitExceeded
	if !errors.As(err, &usageErr) {
		t.Fatal("expected error to be ErrUsageLimitExceeded")
	}
	if usageErr.RunnerName != "codex" {
		t.Errorf("expected runner name 'codex', got %q", usageErr.RunnerName)
	}
}

func TestCodexRunner_OtherErrorNotUsageLimit(t *testing.T) {
	logger := zap.NewNop()
	r := NewCodexRunner(logger, nil)

	dir := t.TempDir()
	r.execCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		// Simulate a generic error without usage limit message
		cmd := exec.CommandContext(ctx, "sh", "-c", `echo 'network timeout' >&2; exit 1`)
		return cmd
	}

	_, err := r.Run(context.Background(), agents.Agent{Name: "agent", Persona: "p", Description: "d"}, "do something", dir, "")
	if err == nil {
		t.Fatal("expected error")
	}

	if IsUsageLimitError(err) {
		t.Fatalf("expected generic error, not ErrUsageLimitExceeded: %v", err)
	}
}
