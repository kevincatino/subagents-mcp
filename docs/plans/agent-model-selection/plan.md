# Agent Model Selection Plan

## Context
- Agents are defined via YAML files loaded by `internal/agents/yaml_repository.go`; schema currently requires `persona` and `description`, with the agent name derived from the filename.
- Runner choice is static at startup via the `--runner` flag in `cmd/subagents/main.go`, selecting either `CodexRunner` or `CopilotRunner` that implement `runner.AgentRunner` with signature `Run(ctx, agent, task, workdir)`.
- MCP `delegate_task` resolves the agent once, then invokes the injected runner; there is no model awareness or dynamic runner fallback today.
- Docs and examples (`docs/setup.md`, `docs/architecture.md`, `examples/agents/*.yaml`) mirror the minimal schema and single-runner behavior.

## Goals
- Allow an optional `model` field per agent YAML to request a specific model for that agent’s tasks.
- Extend the runner pipeline so each `AgentRunner` advertises the models it supports, sourced from a YAML config mapping runner name -> models + priority.
- Pass the selected model into `AgentRunner.Run` and ensure runner invocation uses it when applicable.
- Implement runner selection that prefers the CLI `--runner` as the first candidate only if it supports the agent’s requested model; otherwise fall back to other configured runners by priority.
- Keep backward compatibility when no model is specified or config is missing/incomplete, defaulting to the existing single-runner behavior.

## Non-Goals
- Adding new runner types beyond Codex and Copilot.
- Implementing per-request overrides for runner priority beyond the config/CLI default.
- Changing MCP tool schemas beyond what is required to carry the model through (if at all).

## Proposed Design
- **Agent schema**: Add `model` (optional, string) to `internal/agents/model.go`; ensure validation still passes when absent. Update YAML loader to parse and trim it, and include it in the returned `Agent`.
- **Runner config file**: Introduce a YAML config (e.g., `runner_config.yaml`) mapping runner names to `models` (string list) and `priority` (int). Example:
  ```yaml
  runners:
    - name: codex
      priority: 1
      models: ["gpt-4o", "gpt-4o-mini"]
    - name: copilot
      priority: 2
      models: ["gpt-4o", "gpt-4o-mini", "anthropic/claude-3-opus"]
  ```
  - Config loader should ignore unknown runner names but log/warn; only runners with implementations are instantiated.
  - Priority determines fallback order after the CLI-provided runner.
- **Runner interface change**: Update `runner.AgentRunner` to accept a model parameter, e.g., `Run(ctx, agent, task, workdir, model string)`. Implementations should receive the resolved model (or empty string) and, where relevant, pass it to the underlying CLI or validate support.
- **Runner capabilities**: Each runner gets a capabilities struct constructed from the config entry (models list). At startup, build a map of runner name -> {priority, models, instance}.
- **Selection algorithm**:
  - Identify the agent’s requested model (may be empty).
  - Start with the CLI `--runner` choice if it exists and supports the requested model (or if no model specified).
  - If unsuitable or errors because of unsupported model, iterate remaining configured runners by ascending priority (ties? use deterministic secondary sort such as name) and choose the first that supports the model (or, if no model, the first available).
  - Surface a clear error if no runner supports the requested model.
  - Keep existing guardrails for workdir validation and agent existence.
- **Wiring changes**:
  - CLI: add a flag or default path for the runner config file; load and validate before constructing runners.
  - MCP handlers: when delegating, use the agent’s `model` and the selection logic to pick the runner, then call `Run` with the model.
- **Documentation**:
  - Update `docs/setup.md`, `docs/architecture.md`, `docs/modules.md`, `docs/api.md` (if schema changes), and examples to show the optional `model` field and runner config.
  - Provide sample runner config file in `examples` or `docs`.

## Work Plan
1. **Schema & loader**: Add `model` to `Agent`; update YAML parsing/tests; keep validation compatible.
2. **Runner config**: Define config struct/types, loader, and validation (priority, non-empty names, optional models). Add tests and sample YAML.
3. **Runner interface & impls**: Extend `AgentRunner` signature; update Codex/Copilot runners to accept `model` and, if feasible, validate against their configured model list before executing.
4. **Selection orchestration**: Implement selection component that takes CLI-preferred runner, config priorities/models, and agent.model to return the runner instance; wire into MCP handler/server composition.
5. **CLI wiring**: Add flag(s) for runner config path and instantiate runners per config; preserve default single-runner path when config absent.
6. **Docs/examples**: Refresh docs and sample agent YAMLs plus new runner config sample.
7. **Validation & tests**: Expand unit tests for loader, selection logic, runner invocation signature, and MCP handler integration.

## Risks / Open Questions
- What is the default behavior when `model` is omitted? (Proposed: use CLI runner first; fall back per priority if CLI runner fails and a model is later required?).
- How should we handle ties/duplicate priorities in config? (Proposed: stable name sort as tiebreaker and warn on duplicates.)
- Should an unsupported-model mismatch be a selection-time error or should the runner return a structured error? (Proposed: selection-time error to avoid running the wrong model.)
- Do we need to thread model into MCP tool schemas for visibility, or keep it internal? (Currently leaning internal since agents are server-side assets.)

## Testing Strategy
- Unit tests for agent YAML loader including optional `model`.
- Unit tests for runner config parsing/validation (happy path, missing fields, duplicate priorities).
- Tests for selection logic across combinations: model match, model missing, unsupported model error, CLI preference honored.
- Updated runner tests to assert model argument is passed through.
- Handler tests to verify model-aware runner selection and error surfacing when no runner supports the model.
