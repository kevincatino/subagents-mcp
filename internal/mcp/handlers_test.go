package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"subagents-mcp/internal/agents"

	"go.uber.org/zap"
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

func (s stubRunner) Run(ctx context.Context, agent agents.Agent, task string, workdir string, model string) (string, error) {
	return s.output, s.err
}

func TestListAgentsHandler(t *testing.T) {
	repo := stubRepo{agents: []agents.Agent{{Name: "a", Persona: "p", Description: "d"}}}
	h := NewHandlers(repo, stubRunner{}, zap.NewNop())

	result, err := h.ListAgents(context.Background())
	if err != nil {
		t.Fatalf("ListAgents error: %v", err)
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected single content item, got %d", len(result.Content))
	}
	if result.Content[0].Type != "text" {
		t.Fatalf("expected text content, got %q", result.Content[0].Type)
	}
	var payload struct {
		Agents []struct {
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Persona     *string `json:"persona,omitempty"`
		} `json:"agents"`
	}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if len(payload.Agents) != 1 || payload.Agents[0].Name != "a" || payload.Agents[0].Description != "d" {
		t.Fatalf("unexpected agents: %#v", payload.Agents)
	}
	if payload.Agents[0].Persona != nil {
		t.Fatalf("persona should be omitted: %#v", payload.Agents[0])
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
