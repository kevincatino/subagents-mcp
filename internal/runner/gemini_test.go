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

func TestGeminiRunnerValidation(t *testing.T) {
	r := NewGeminiRunner(zap.NewNop(), []string{"gemini-1.5"})
	agent := agents.Agent{Name: "agent", Persona: "persona", Description: "desc"}

	if _, err := r.Run(context.Background(), agent, "", "/tmp", "gemini-1.5"); err == nil {
		t.Fatal("expected error for empty task")
	}
	if _, err := r.Run(context.Background(), agent, "task", "relative", "gemini-1.5"); err == nil {
		t.Fatal("expected error for relative workdir")
	}
	if _, err := r.Run(context.Background(), agent, "task", "/tmp", "other"); err == nil {
		t.Fatal("expected error for unsupported model")
	}
}

func TestGeminiRunnerBuildsCommand(t *testing.T) {
	logger := zap.NewNop()
	r := NewGeminiRunner(logger, nil)

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

	out, err := r.Run(context.Background(), agents.Agent{Name: "agent", Persona: "p", Description: "d"}, "do something", dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "ok" {
		t.Fatalf("expected ok, got %q", out)
	}
	if gotName != "gemini" {
		t.Fatalf("expected gemini command, got %s", gotName)
	}
	expected := []string{"-p", "p\n\nTask: do something", "--output-format", "json"}
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

func TestGeminiRunnerIncludesModelFlag(t *testing.T) {
	logger := zap.NewNop()
	r := NewGeminiRunner(logger, []string{"gemini-2.5"})

	dir := t.TempDir()
	var gotArgs []string
	r.execCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		gotArgs = append([]string(nil), arg...)
		return exec.CommandContext(ctx, "echo", "ok")
	}

	if _, err := r.Run(context.Background(), agents.Agent{Name: "agent", Persona: "p", Description: "d"}, "task", dir, "gemini-2.5"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPrefix := []string{"-p", "p\n\nTask: task", "--output-format", "json", "-m", "gemini-2.5"}
	if len(gotArgs) < len(expectedPrefix) {
		t.Fatalf("expected at least %d args, got %d", len(expectedPrefix), len(gotArgs))
	}
	for i := range expectedPrefix {
		if gotArgs[i] != expectedPrefix[i] {
			t.Fatalf("expected arg %d to be %s, got %s", i, expectedPrefix[i], gotArgs[i])
		}
	}
}

func TestGeminiRunnerUsageLimitDetection(t *testing.T) {
	logger := zap.NewNop()
	r := NewGeminiRunner(logger, nil)

	dir := t.TempDir()
	r.execCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, "sh", "-c", `echo 'quota exceeded' >&2; exit 1`)
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
	if usageErr.RunnerName != "gemini" {
		t.Errorf("expected runner name 'gemini', got %q", usageErr.RunnerName)
	}
}

func TestGeminiRunnerOtherErrorNotUsageLimit(t *testing.T) {
	logger := zap.NewNop()
	r := NewGeminiRunner(logger, nil)

	dir := t.TempDir()
	r.execCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, "sh", "-c", `echo 'authentication failed' >&2; exit 1`)
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
