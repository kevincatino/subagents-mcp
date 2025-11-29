# Modules

- `cmd/subagents/main.go` – flag parsing (`--agents-dir`, `--runner`), logger init, wiring repository, runner, and server.
- `internal/agents` – `Agent` model validation and YAML repository loader for persona files.
- `internal/mcp` – JSON-RPC request handling, initialize response, tool schemas, tool dispatch, and MCP error helpers.
- `internal/mcp/handlers.go` – implementations of `list_agents` and `delegate_task`.
- `internal/runner` – `AgentRunner` interface plus Codex and Copilot runner adapters and prompt builder.
- `internal/validate` – path validation (absolute, existing, non-root, symlink-resolved).
- `internal/logging` – zap production logger configuration.
- `examples/agents` – sample agent YAML(s) for local testing.
