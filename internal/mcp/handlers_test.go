package mcp

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"

	"subagents-mcp/internal/agents"
)

type stubRepo struct {
	agents []agents.Agent
	err    error
}

func (s stubRepo) ListAgents(ctx context.Context) ([]agents.Agent, error) {
	return s.agents, s.err
}

type stubRunner struct {
	output string
	err    error
}

func (s stubRunner) Run(ctx context.Context, agent agents.Agent, task string, workdir string) (string, error) {
	return s.output, s.err
}

func TestListAgentsHandler(t *testing.T) {
	repo := stubRepo{agents: []agents.Agent{{Name: "a", Persona: "p", Description: "d"}}}
	h := NewHandlers(repo, stubRunner{}, zap.NewNop())

	result, err := h.ListAgents(context.Background())
	if err != nil {
		t.Fatalf("ListAgents error: %v", err)
	}
	if len(result.Agents) != 1 || result.Agents[0].Name != "a" {
		t.Fatalf("unexpected agents: %#v", result.Agents)
	}
}

func TestDelegateTaskHandlerValidates(t *testing.T) {
	repo := stubRepo{agents: []agents.Agent{{Name: "a", Persona: "p", Description: "d"}}}
	h := NewHandlers(repo, stubRunner{}, zap.NewNop())

	if _, err := h.DelegateTask(context.Background(), delegateArgs{}); err == nil {
		t.Fatal("expected error for missing agent and task")
	}
	if _, err := h.DelegateTask(context.Background(), delegateArgs{Agent: "missing", Task: "t", WorkingDirectory: "/tmp"}); err == nil {
		t.Fatal("expected error for missing agent match")
	}
}

func TestDelegateTaskHandlerHappyPath(t *testing.T) {
	repo := stubRepo{agents: []agents.Agent{{Name: "a", Persona: "p", Description: "d"}}}
	runner := stubRunner{output: "done"}
	h := NewHandlers(repo, runner, zap.NewNop())

	// ensure /tmp exists
	result, err := h.DelegateTask(context.Background(), delegateArgs{Agent: "a", Task: "t", WorkingDirectory: "/tmp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Content) != 1 || result.Content[0].Text != "done" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestDelegateTaskHandlerPropagatesRunnerError(t *testing.T) {
	repo := stubRepo{agents: []agents.Agent{{Name: "a", Persona: "p", Description: "d"}}}
	runner := stubRunner{err: errors.New("fail")}
	h := NewHandlers(repo, runner, zap.NewNop())

	if _, err := h.DelegateTask(context.Background(), delegateArgs{Agent: "a", Task: "t", WorkingDirectory: "/tmp"}); err == nil {
		t.Fatal("expected runner error")
	}
}
