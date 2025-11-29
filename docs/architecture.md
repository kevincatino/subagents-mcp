# Architecture

## Overview
The server exposes MCP 2024-11-05 over stdio/JSON-RPC with two tools: `list_agents` and `delegate_task`. It wires a YAML-backed agent repository to a pluggable runner (Codex CLI or Copilot CLI) and returns tool results as MCP content items.

## Components
- Entrypoint (`cmd/subagents/main.go`): parses flags `--agents-dir` (required, absolute) and `--runner` (`codex` default, `copilot` optional); constructs logger, repository, runner, and server.
- Validation (`internal/validate`): ensures paths are absolute, existing directories, not `/`, and resolves symlinks.
- Agents (`internal/agents`): `Agent` model validation plus YAML repository that loads `*.yaml` personas (`persona`, `description`) from `--agents-dir`.
- MCP layer (`internal/mcp`): JSON-RPC request decoding, initialize handshake, tools list, and tool dispatch to handlers; uses MCP error codes for protocol issues.
- Handlers (`internal/mcp/handlers.go`): implement `list_agents` (returns JSON string of name/description) and `delegate_task` (validates args, ensures agent exists, runs via runner).
- Runners (`internal/runner`): `AgentRunner` interface with Codex and Copilot implementations that inject agent persona into the task prompt and execute in the provided working directory.
- Logging (`internal/logging`): zap production JSON logger.

## Control Flow
1. Client sends `initialize`; server responds with protocol version `2024-11-05`, tools capability, and server info.
2. `tools/list` returns tool metadata with JSON Schemas.
3. `tools/call` routes to handlers:
   - `list_agents`: reads YAML personas and returns JSON payload of agents.
   - `delegate_task`: validates agent name and working directory, builds persona+task prompt, invokes selected runner, returns final stdout text.

## Runners and Guardrails
- Codex runner: `codex --cd <workdir> --sandbox read-only --ask-for-approval never exec "<prompt>"`; activity streams to stderr, final message to stdout.
- Copilot runner: `copilot -p "<prompt>" --allow-all-tools --allow-all-paths --stream off` executed in the working directory.
- Guardrails: reject empty/relative/root paths; symlinks resolved; working directory must exist.
