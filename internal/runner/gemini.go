package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"

	"subagents-mcp/internal/agents"
	"subagents-mcp/internal/validate"
)

// GeminiRunner invokes the Gemini CLI in non-interactive mode.
type GeminiRunner struct {
	logger      *zap.Logger
	execCommand func(ctx context.Context, name string, arg ...string) *exec.Cmd
	models      map[string]struct{}
}

// NewGeminiRunner constructs a new GeminiRunner.
func NewGeminiRunner(logger *zap.Logger, supportedModels []string) *GeminiRunner {
	return &GeminiRunner{
		logger:      logger,
		execCommand: exec.CommandContext,
		models:      toModelSet(supportedModels),
	}
}

// Run executes the Gemini CLI with the supplied prompt and model.
func (g *GeminiRunner) Run(ctx context.Context, agent agents.Agent, task string, workdir string, model string) (string, error) {
	if task == "" {
		return "", errors.New("task is required")
	}
	resolvedWorkdir, err := validate.Dir(workdir)
	if err != nil {
		return "", fmt.Errorf("validate workdir: %w", err)
	}

	if !supportsModel(g.models, model) {
		return "", fmt.Errorf("model %q not supported by gemini runner", model)
	}

	prompt := buildAgentPrompt(agent, task)

	args := []string{
		"-p", prompt,
		"--output-format", "json",
	}
	if model != "" {
		args = append(args, "-m", model)
	}

	cmd := g.execCommand(ctx, "gemini", args...)
	cmd.Dir = resolvedWorkdir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err = cmd.Run()
	duration := time.Since(start)

	g.logger.Info("delegate task completed",
		zap.String("runner", "gemini"),
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
		if isGeminiUsageLimitMessage(combined) {
			return "", &ErrUsageLimitExceeded{
				RunnerName: "gemini",
				Message:    extractUsageLimitMessage(combined),
			}
		}
		return "", fmt.Errorf("gemini exec failed: %w; stderr: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	var parsed struct {
		Response string `json:"response"`
	}
	if json.Unmarshal([]byte(output), &parsed) == nil && parsed.Response != "" {
		return strings.TrimSpace(parsed.Response), nil
	}
	return output, nil
}
