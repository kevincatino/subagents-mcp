package runner

import (
	"context"
	"fmt"
	"sort"

	"go.uber.org/zap"

	"subagents-mcp/internal/agents"
)

// runnerFactories can be overridden in tests to stub runner construction.
var runnerFactories = map[string]func(*zap.Logger, []string) AgentRunner{
	"codex": func(logger *zap.Logger, models []string) AgentRunner {
		return NewCodexRunner(logger, models)
	},
	"copilot": func(logger *zap.Logger, models []string) AgentRunner {
		return NewCopilotRunner(logger, models)
	},
	"gemini": func(logger *zap.Logger, models []string) AgentRunner {
		return NewGeminiRunner(logger, models)
	},
}

var defaultRunnerOrder = []string{"codex", "copilot", "gemini"}

type namedRunner struct {
	name     string
	priority int
	models   map[string]struct{}
	runner   AgentRunner
}

// Selector chooses a concrete runner based on model support and priority.
type Selector struct {
	logger    *zap.Logger
	preferred *namedRunner
	fallbacks []namedRunner
}

// NewSelector builds a model-aware runner selector using a preferred runner name
// and optional configuration. The preferred runner is always tried first when
// available; fallbacks are sorted by priority (then name).

func NewSelector(logger *zap.Logger, cfg Config, preferred string) (*Selector, error) {
	entries := make([]namedRunner, 0, len(cfg.Runners))
	for _, rc := range cfg.Runners {
		build, ok := runnerFactories[rc.Name]
		if !ok {
			logger.Warn("ignoring unknown runner in config", zap.String("runner", rc.Name))
			continue
		}
		entries = append(entries, namedRunner{
			name:     rc.Name,
			priority: rc.Priority,
			models:   toModelSet(rc.Models),
			runner:   build(logger, rc.Models),
		})
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].priority == entries[j].priority {
			return entries[i].name < entries[j].name
		}
		return entries[i].priority < entries[j].priority
	})

	if preferred == "" {
		if len(entries) == 0 {
			entries = defaultRunnerEntries(logger)
		}
		return &Selector{
			logger:    logger,
			preferred: nil,
			fallbacks: entries,
		}, nil
	}

	constructor, ok := runnerFactories[preferred]
	if !ok {
		return nil, fmt.Errorf("invalid runner %q", preferred)
	}

	var preferredRunner *namedRunner
	for i := range entries {
		if entries[i].name == preferred {
			preferredRunner = &entries[i]
			break
		}
	}

	if preferredRunner == nil {
		preferredRunner = &namedRunner{
			name:   preferred,
			runner: constructor(logger, nil),
			models: nil,
		}
	}

	fallbacks := make([]namedRunner, 0, len(entries))
	for _, entry := range entries {
		if entry.name == preferredRunner.name {
			continue
		}
		fallbacks = append(fallbacks, entry)
	}

	return &Selector{
		logger:    logger,
		preferred: preferredRunner,
		fallbacks: fallbacks,
	}, nil
}

func defaultRunnerEntries(logger *zap.Logger) []namedRunner {
	entries := make([]namedRunner, 0, len(defaultRunnerOrder))
	for idx, name := range defaultRunnerOrder {
		build, ok := runnerFactories[name]
		if !ok {
			continue
		}
		entries = append(entries, namedRunner{
			name:     name,
			priority: idx + 1,
			models:   nil,
			runner:   build(logger, nil),
		})
	}
	return entries
}

func (s *Selector) Run(ctx context.Context, agent agents.Agent, task string, workdir string, model string) (string, error) {
	candidates := make([]namedRunner, 0, 1+len(s.fallbacks))
	if s.preferred != nil {
		candidates = append(candidates, *s.preferred)
	}
	candidates = append(candidates, s.fallbacks...)

	var lastUsageLimitErr error
	for _, candidate := range candidates {
		if !supportsModel(candidate.models, model) {
			continue
		}
		output, err := candidate.runner.Run(ctx, agent, task, workdir, model)
		if err == nil {
			return output, nil
		}
		if IsUsageLimitError(err) {
			s.logger.Warn("runner hit usage limit, trying next",
				zap.String("runner", candidate.name),
				zap.Error(err))
			lastUsageLimitErr = err
			continue
		}
		// Non-usage-limit error: fail immediately
		return "", err
	}

	if lastUsageLimitErr != nil {
		return "", fmt.Errorf("all runners exhausted due to usage limits: %w", lastUsageLimitErr)
	}
	if model == "" {
		return "", fmt.Errorf("no runner available")
	}
	return "", fmt.Errorf("no runner supports model %q", model)
}
