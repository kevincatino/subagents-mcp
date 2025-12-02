package runner

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"

	"subagents-mcp/internal/agents"
)

type fakeRunner struct {
	name   string
	called bool
	model  string
	output string
	runErr error
}

func (f *fakeRunner) Run(ctx context.Context, agent agents.Agent, task string, workdir string, model string) (string, error) {
	f.called = true
	f.model = model
	return f.output, f.runErr
}

func TestSelectorPrefersFlagAndFallsBack(t *testing.T) {
	origFactories := runnerFactories
	defer func() { runnerFactories = origFactories }()

	codex := &fakeRunner{name: "codex", output: "codex-out"}
	copilot := &fakeRunner{name: "copilot", output: "copilot-out"}

	runnerFactories = map[string]func(*zap.Logger, []string) AgentRunner{
		"codex": func(_ *zap.Logger, _ []string) AgentRunner { return codex },
		"copilot": func(_ *zap.Logger, _ []string) AgentRunner {
			return copilot
		},
	}

	cfg := Config{
		Runners: []RunnerConfig{
			{Name: "codex", Priority: 2, Models: []string{"gpt-4o"}},
			{Name: "copilot", Priority: 1, Models: []string{"claude"}},
		},
	}

	selector, err := NewSelector(zap.NewNop(), cfg, "codex")
	if err != nil {
		t.Fatalf("NewSelector error: %v", err)
	}

	out, err := selector.Run(context.Background(), agents.Agent{Name: "a", Persona: "p", Description: "d"}, "task", "/tmp", "gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "codex-out" {
		t.Fatalf("expected codex output, got %q", out)
	}
	if !codex.called {
		t.Fatal("expected codex runner to be called")
	}
	if copilot.called {
		t.Fatal("expected copilot runner not to be called")
	}

	codex.called = false
	copilot.called = false

	out, err = selector.Run(context.Background(), agents.Agent{Name: "a", Persona: "p", Description: "d"}, "task", "/tmp", "claude")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "copilot-out" {
		t.Fatalf("expected copilot output, got %q", out)
	}
	if !copilot.called {
		t.Fatal("expected copilot to be called for fallback")
	}
	if codex.called {
		t.Fatal("expected codex to be skipped for unsupported model")
	}
}

func TestSelector_NoPreferredUsesConfigOrder(t *testing.T) {
	origFactories := runnerFactories
	defer func() { runnerFactories = origFactories }()

	codex := &fakeRunner{name: "codex", output: "codex-out"}
	copilot := &fakeRunner{name: "copilot", output: "copilot-out"}

	runnerFactories = map[string]func(*zap.Logger, []string) AgentRunner{
		"codex":   func(_ *zap.Logger, _ []string) AgentRunner { return codex },
		"copilot": func(_ *zap.Logger, _ []string) AgentRunner { return copilot },
	}

	cfg := Config{
		Runners: []RunnerConfig{
			{Name: "codex", Priority: 2, Models: []string{"gpt-4o"}},
			{Name: "copilot", Priority: 1, Models: []string{"gpt-4o"}},
		},
	}

	selector, err := NewSelector(zap.NewNop(), cfg, "")
	if err != nil {
		t.Fatalf("NewSelector error: %v", err)
	}

	out, err := selector.Run(context.Background(), agents.Agent{Name: "a", Persona: "p", Description: "d"}, "task", "/tmp", "gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "copilot-out" {
		t.Fatalf("expected copilot output first, got %q", out)
	}
	if !copilot.called {
		t.Fatal("expected copilot to be called first when no preferred runner")
	}
	if codex.called {
		t.Fatal("expected codex not to be called after copilot success")
	}
}

func TestSelector_NoPreferredNoConfigDefaultsToAll(t *testing.T) {
	origFactories := runnerFactories
	defer func() { runnerFactories = origFactories }()

	codex := &fakeRunner{
		name:   "codex",
		runErr: &ErrUsageLimitExceeded{RunnerName: "codex", Message: "quota"},
	}
	copilot := &fakeRunner{
		name:   "copilot",
		runErr: &ErrUsageLimitExceeded{RunnerName: "copilot", Message: "quota"},
	}
	gemini := &fakeRunner{name: "gemini", output: "gemini-out"}

	runnerFactories = map[string]func(*zap.Logger, []string) AgentRunner{
		"codex":   func(_ *zap.Logger, _ []string) AgentRunner { return codex },
		"copilot": func(_ *zap.Logger, _ []string) AgentRunner { return copilot },
		"gemini":  func(_ *zap.Logger, _ []string) AgentRunner { return gemini },
	}

	selector, err := NewSelector(zap.NewNop(), Config{}, "")
	if err != nil {
		t.Fatalf("NewSelector error: %v", err)
	}

	out, err := selector.Run(context.Background(), agents.Agent{Name: "a", Persona: "p", Description: "d"}, "task", "/tmp", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "gemini-out" {
		t.Fatalf("expected gemini output, got %q", out)
	}
	if !codex.called || !copilot.called || !gemini.called {
		t.Fatalf("expected all runners to be attempted, got codex=%v copilot=%v gemini=%v", codex.called, copilot.called, gemini.called)
	}
}

func TestSelectorErrorsWhenUnsupported(t *testing.T) {
	origFactories := runnerFactories
	defer func() { runnerFactories = origFactories }()

	codex := &fakeRunner{name: "codex", output: "codex-out"}
	runnerFactories = map[string]func(*zap.Logger, []string) AgentRunner{
		"codex": func(_ *zap.Logger, _ []string) AgentRunner { return codex },
	}

	cfg := Config{
		Runners: []RunnerConfig{
			{Name: "codex", Priority: 1, Models: []string{"gpt-4o"}},
		},
	}

	selector, err := NewSelector(zap.NewNop(), cfg, "codex")
	if err != nil {
		t.Fatalf("NewSelector error: %v", err)
	}

	if _, err := selector.Run(context.Background(), agents.Agent{Name: "a", Persona: "p", Description: "d"}, "task", "/tmp", "claude"); err == nil {
		t.Fatal("expected error when no runner supports model")
	}
	if codex.called {
		t.Fatal("expected codex not to be invoked for unsupported model")
	}
}

func TestSelector_FallbackOnUsageLimit(t *testing.T) {
	origFactories := runnerFactories
	defer func() { runnerFactories = origFactories }()

	codex := &fakeRunner{
		name:   "codex",
		runErr: &ErrUsageLimitExceeded{RunnerName: "codex", Message: "quota exceeded"},
	}
	copilot := &fakeRunner{name: "copilot", output: "copilot-out"}

	runnerFactories = map[string]func(*zap.Logger, []string) AgentRunner{
		"codex":   func(_ *zap.Logger, _ []string) AgentRunner { return codex },
		"copilot": func(_ *zap.Logger, _ []string) AgentRunner { return copilot },
	}

	cfg := Config{
		Runners: []RunnerConfig{
			{Name: "codex", Priority: 1, Models: []string{"gpt-4o"}},
			{Name: "copilot", Priority: 2, Models: []string{"gpt-4o"}},
		},
	}

	selector, err := NewSelector(zap.NewNop(), cfg, "codex")
	if err != nil {
		t.Fatalf("NewSelector error: %v", err)
	}

	out, err := selector.Run(context.Background(), agents.Agent{Name: "a", Persona: "p", Description: "d"}, "task", "/tmp", "gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "copilot-out" {
		t.Fatalf("expected copilot output, got %q", out)
	}
	if !codex.called {
		t.Fatal("expected codex to be called first")
	}
	if !copilot.called {
		t.Fatal("expected copilot to be called as fallback")
	}
}

func TestSelector_AllRunnersExhausted(t *testing.T) {
	origFactories := runnerFactories
	defer func() { runnerFactories = origFactories }()

	codex := &fakeRunner{
		name:   "codex",
		runErr: &ErrUsageLimitExceeded{RunnerName: "codex", Message: "quota exceeded"},
	}
	copilot := &fakeRunner{
		name:   "copilot",
		runErr: &ErrUsageLimitExceeded{RunnerName: "copilot", Message: "rate limited"},
	}

	runnerFactories = map[string]func(*zap.Logger, []string) AgentRunner{
		"codex":   func(_ *zap.Logger, _ []string) AgentRunner { return codex },
		"copilot": func(_ *zap.Logger, _ []string) AgentRunner { return copilot },
	}

	cfg := Config{
		Runners: []RunnerConfig{
			{Name: "codex", Priority: 1, Models: []string{"gpt-4o"}},
			{Name: "copilot", Priority: 2, Models: []string{"gpt-4o"}},
		},
	}

	selector, err := NewSelector(zap.NewNop(), cfg, "codex")
	if err != nil {
		t.Fatalf("NewSelector error: %v", err)
	}

	_, err = selector.Run(context.Background(), agents.Agent{Name: "a", Persona: "p", Description: "d"}, "task", "/tmp", "gpt-4o")
	if err == nil {
		t.Fatal("expected error when all runners exhausted")
	}
	if !codex.called || !copilot.called {
		t.Fatal("expected both runners to be tried")
	}
	// The error should wrap the last usage limit error
	var usageErr *ErrUsageLimitExceeded
	if !errors.As(err, &usageErr) {
		t.Fatalf("expected error chain to contain ErrUsageLimitExceeded, got: %v", err)
	}
}

func TestSelector_NonUsageLimitErrorNoFallback(t *testing.T) {
	origFactories := runnerFactories
	defer func() { runnerFactories = origFactories }()

	codex := &fakeRunner{
		name:   "codex",
		runErr: errors.New("network timeout"),
	}
	copilot := &fakeRunner{name: "copilot", output: "copilot-out"}

	runnerFactories = map[string]func(*zap.Logger, []string) AgentRunner{
		"codex":   func(_ *zap.Logger, _ []string) AgentRunner { return codex },
		"copilot": func(_ *zap.Logger, _ []string) AgentRunner { return copilot },
	}

	cfg := Config{
		Runners: []RunnerConfig{
			{Name: "codex", Priority: 1, Models: []string{"gpt-4o"}},
			{Name: "copilot", Priority: 2, Models: []string{"gpt-4o"}},
		},
	}

	selector, err := NewSelector(zap.NewNop(), cfg, "codex")
	if err != nil {
		t.Fatalf("NewSelector error: %v", err)
	}

	_, err = selector.Run(context.Background(), agents.Agent{Name: "a", Persona: "p", Description: "d"}, "task", "/tmp", "gpt-4o")
	if err == nil {
		t.Fatal("expected error")
	}
	if !codex.called {
		t.Fatal("expected codex to be called")
	}
	if copilot.called {
		t.Fatal("expected copilot NOT to be called for non-usage-limit error")
	}
}
