package runner

import (
	"context"

	"subagents-mcp/internal/agents"
)

// AgentRunner executes tasks for a given agent.
type AgentRunner interface {
	Run(ctx context.Context, agent agents.Agent, task string, workdir string, model string) (string, error)
}
