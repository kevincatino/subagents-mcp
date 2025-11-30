# Modules

- `cmd/subagents/main.go` – flag parsing (`--agents-dir`, `--runner`, optional `--runner-config`), logger init, wiring repository, runner selector, and server.
- `internal/agents` – `Agent` model validation and YAML repository loader for persona files (persona, description, optional model).
- `internal/mcp` – JSON-RPC request handling, initialize response, tool schemas, tool dispatch, and MCP error helpers.
- `internal/mcp/handlers.go` – implementations of `list_agents` and `delegate_task`.
- `internal/runner` – `AgentRunner` interface plus Codex, Copilot, and Gemini runner adapters, prompt builder, runner config loader, and model-aware selector that orders runners by priority.
- `internal/validate` – path validation (absolute, existing, non-root, symlink-resolved).
- `internal/logging` – zap production logger configuration.
- `examples/agents` – sample agent YAML(s) for local testing.
