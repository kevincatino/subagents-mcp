package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"subagents-mcp/internal/agents"
	"subagents-mcp/internal/runner"
	"subagents-mcp/internal/validate"
)

type Handlers struct {
	repo   agents.Repository
	runner runner.AgentRunner
	logger *zap.Logger
}

func NewHandlers(repo agents.Repository, runner runner.AgentRunner, logger *zap.Logger) *Handlers {
	return &Handlers{repo: repo, runner: runner, logger: logger}
}

type listAgentsResult struct {
	Content []contentItem `json:"content"`
}

// agentSummary exposes only the public metadata for an agent.
type agentSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type delegateArgs struct {
	Agent            string `json:"agent"`
	Task             string `json:"task"`
	WorkingDirectory string `json:"working_directory"`
}

type delegateResult struct {
	Content []contentItem `json:"content"`
}

type contentItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

func (h *Handlers) ListAgents(ctx context.Context) (listAgentsResult, error) {
	agentsList, err := h.repo.ListAgents(ctx)
	if err != nil {
		return listAgentsResult{}, err
	}
	summaries := make([]agentSummary, 0, len(agentsList))
	for _, agent := range agentsList {
		summaries = append(summaries, agentSummary{
			Name:        agent.Name,
			Description: agent.Description,
		})
	}

	payload, err := json.Marshal(map[string]any{"agents": summaries})
	if err != nil {
		return listAgentsResult{}, fmt.Errorf("marshal agents: %w", err)
	}
	return listAgentsResult{Content: []contentItem{{Type: "text", Text: string(payload)}}}, nil
}

func (h *Handlers) DelegateTask(ctx context.Context, args delegateArgs) (delegateResult, error) {
	if args.Agent == "" {
		return delegateResult{}, fmt.Errorf("agent is required")
	}
	if args.Task == "" {
		return delegateResult{}, fmt.Errorf("task is required")
	}
	workdir, err := validate.Dir(args.WorkingDirectory)
	if err != nil {
		return delegateResult{}, fmt.Errorf("working_directory invalid: %w", err)
	}

	agentsList, err := h.repo.ListAgents(ctx)
	if err != nil {
		return delegateResult{}, err
	}

	var selected *agents.Agent
	for i := range agentsList {
		if agentsList[i].Name == args.Agent {
			selected = &agentsList[i]
			break
		}
	}
	if selected == nil {
		return delegateResult{}, fmt.Errorf("agent %q not found", args.Agent)
	}

	output, err := h.runner.Run(ctx, *selected, args.Task, workdir)
	if err != nil {
		return delegateResult{}, err
	}

	return delegateResult{Content: []contentItem{{Type: "text", Text: output}}}, nil
}

func decodeArgs[T any](raw json.RawMessage) (T, error) {
	var args T
	if len(raw) == 0 {
		return args, fmt.Errorf("arguments are required")
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return args, err
	}
	return args, nil
}
