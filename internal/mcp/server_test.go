package mcp

import (
	"context"
	"testing"

	"subagents-mcp/internal/agents"

	"go.uber.org/zap"
)

type initStubRepo struct {
	agents []agents.Agent
}

func (s initStubRepo) ListAgents(ctx context.Context) ([]agents.Agent, error) {
	return s.agents, nil
}

type initStubRunner struct{}

func (initStubRunner) Run(ctx context.Context, agent agents.Agent, task string, workdir string) (string, error) {
	return "", nil
}

func TestHandleInitialize(t *testing.T) {
	s := NewServer(zap.NewNop(), initStubRepo{}, initStubRunner{})

	resp, ok := s.handle(context.Background(), Request{JSONRPC: "2.0", ID: 1, Method: "initialize"})
	if !ok {
		t.Fatalf("expected response, got ok=%v", ok)
	}

	result, ok := resp.Result.(InitializeResult)
	if !ok {
		t.Fatalf("unexpected result type: %#v", resp.Result)
	}
	if result.ProtocolVersion == "" {
		t.Fatalf("missing protocolVersion in result: %#v", result)
	}
}

func TestNotificationsInitializedIgnored(t *testing.T) {
	s := NewServer(zap.NewNop(), initStubRepo{}, initStubRunner{})

	_, ok := s.handle(context.Background(), Request{JSONRPC: "2.0", Method: "notifications/initialized"})
	if ok {
		t.Fatalf("expected notification to be ignored")
	}
}
