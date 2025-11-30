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
	models      map[string]struct{}
}

func NewCodexRunner(logger *zap.Logger, supportedModels []string) *CodexRunner {
	return &CodexRunner{
		logger:      logger,
		execCommand: exec.CommandContext,
		models:      toModelSet(supportedModels),
	}
}

func (c *CodexRunner) Run(ctx context.Context, agent agents.Agent, task string, workdir string, model string) (string, error) {
	if task == "" {
		return "", errors.New("task is required")
	}
	resolvedWorkdir, err := validate.Dir(workdir)
	if err != nil {
		return "", fmt.Errorf("validate workdir: %w", err)
	}

	if !supportsModel(c.models, model) {
		return "", fmt.Errorf("model %q not supported by codex runner", model)
	}

	prompt := buildAgentPrompt(agent, task)

	args := []string{
		"--cd", resolvedWorkdir,
		"--sandbox", "read-only",
		"--ask-for-approval", "never",
		"exec",
		"--skip-git-repo-check",
	}
	if model != "" {
		args = append(args, "--model", model)
	}
	args = append(args, prompt)

	cmd := c.execCommand(ctx, "codex", args...)
	cmd.Dir = resolvedWorkdir
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
		zap.String("model", model),
		zap.Duration("duration", duration),
		zap.ByteString("stderr", stderr.Bytes()),
		zap.Error(err),
	)

	if err != nil {
		combined := stderr.String() + stdout.String()
		if isCodexUsageLimitMessage(combined) {
			return "", &ErrUsageLimitExceeded{
				RunnerName: "codex",
				Message:    extractUsageLimitMessage(combined),
			}
		}
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

func toModelSet(models []string) map[string]struct{} {
	if len(models) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(models))
	for _, m := range models {
		set[m] = struct{}{}
	}
	return set
}

func supportsModel(set map[string]struct{}, model string) bool {
	if set == nil || model == "" {
		return true
	}
	_, ok := set[model]
	return ok
}
