package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"

	"subagents-mcp/internal/agents"
	"subagents-mcp/internal/validate"
)

// CodexRunner invokes the Codex CLI in non-interactive mode.
type CodexRunner struct {
	logger      *zap.Logger
	execCommand func(ctx context.Context, name string, arg ...string) *exec.Cmd
}

func NewCodexRunner(logger *zap.Logger) *CodexRunner {
	return &CodexRunner{
		logger:      logger,
		execCommand: exec.CommandContext,
	}
}

func (c *CodexRunner) Run(ctx context.Context, agent agents.Agent, task string, workdir string) (string, error) {
	if task == "" {
		return "", errors.New("task is required")
	}
	resolvedWorkdir, err := validate.Dir(workdir)
	if err != nil {
		return "", fmt.Errorf("validate workdir: %w", err)
	}

	args := []string{
		"--cd", resolvedWorkdir,
		"--sandbox", "read-only",
		"--ask-for-approval", "never",
		"exec",
		task,
	}

	cmd := c.execCommand(ctx, "codex", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err = cmd.Run()
	duration := time.Since(start)

	c.logger.Info("delegate task completed",
		zap.String("agent", agent.Name),
		zap.String("workdir", resolvedWorkdir),
		zap.String("task", truncate(task, 200)),
		zap.Duration("duration", duration),
		zap.ByteString("stderr", stderr.Bytes()),
		zap.Error(err),
	)

	if err != nil {
		return "", fmt.Errorf("codex exec failed: %w; stderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "..."
}
