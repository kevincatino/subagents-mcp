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

// CopilotRunner invokes the GitHub Copilot CLI in non-interactive, single-prompt mode.
type CopilotRunner struct {
	logger      *zap.Logger
	execCommand func(ctx context.Context, name string, arg ...string) *exec.Cmd
	models      map[string]struct{}
}

func NewCopilotRunner(logger *zap.Logger, supportedModels []string) *CopilotRunner {
	return &CopilotRunner{
		logger:      logger,
		execCommand: exec.CommandContext,
		models:      toModelSet(supportedModels),
	}
}

func (c *CopilotRunner) Run(ctx context.Context, agent agents.Agent, task string, workdir string, model string) (string, error) {
	if task == "" {
		return "", errors.New("task is required")
	}
	resolvedWorkdir, err := validate.Dir(workdir)
	if err != nil {
		return "", fmt.Errorf("validate workdir: %w", err)
	}

	if !supportsModel(c.models, model) {
		return "", fmt.Errorf("model %q not supported by copilot runner", model)
	}

	prompt := buildAgentPrompt(agent, task)

	args := []string{
		"-p", prompt,
		"--allow-all-tools",
		"--allow-all-paths",
		"--stream", "off",
	}
	if model != "" {
		args = append([]string{"--model", model}, args...)
	}

	cmd := c.execCommand(ctx, "copilot", args...)
	cmd.Dir = resolvedWorkdir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err = cmd.Run()
	duration := time.Since(start)

	c.logger.Info("delegate task completed",
		zap.String("runner", "copilot"),
		zap.String("agent", agent.Name),
		zap.String("workdir", resolvedWorkdir),
		zap.String("task", truncate(task, 200)),
		zap.String("model", model),
		zap.Duration("duration", duration),
		zap.ByteString("stderr", stderr.Bytes()),
		zap.Error(err),
	)

	if err != nil {
		return "", fmt.Errorf("copilot exec failed: %w; stderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}
