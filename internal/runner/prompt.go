package runner

import (
	"fmt"
	"strings"

	"subagents-mcp/internal/agents"
)

// buildAgentPrompt injects the agent persona ahead of the task so the runner
// has full context on the delegate's role.
func buildAgentPrompt(agent agents.Agent, task string) string {
	persona := strings.TrimSpace(agent.Persona)
	trimmedTask := strings.TrimSpace(task)

	switch {
	case persona == "":
		return trimmedTask
	case trimmedTask == "":
		return persona
	default:
		return fmt.Sprintf("%s\n\nTask: %s", persona, trimmedTask)
	}
}
