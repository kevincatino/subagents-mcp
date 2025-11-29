# Architecture

## Overview
The server exposes MCP 2024-11-05 over stdio/JSON-RPC with two tools: `list_agents` and `delegate_task`. It wires a YAML-backed agent repository (persona/description plus optional `model`) to a runner selector that prefers the CLI-specified runner and falls back based on configured model support/priority, returning tool results as MCP content items.

## Components
- Entrypoint (`cmd/subagents/main.go`): parses flags `--agents-dir` (required, absolute), `--runner` (`codex` default, `copilot` optional), and `--runner-config` (optional YAML describing priorities/models); constructs logger, repository, runner selector, and server.
- Validation (`internal/validate`): ensures paths are absolute, existing directories, not `/`, and resolves symlinks.
- Agents (`internal/agents`): `Agent` model validation plus YAML repository that loads `*.yaml` personas (`persona`, `description`, optional `model`) from `--agents-dir`.
- MCP layer (`internal/mcp`): JSON-RPC request decoding, initialize handshake, tools list, and tool dispatch to handlers; uses MCP error codes for protocol issues.
- Handlers (`internal/mcp/handlers.go`): implement `list_agents` (returns JSON string of name/description) and `delegate_task` (validates args, ensures agent exists, runs via runner selector with the agent’s `model`).
- Runners (`internal/runner`): `AgentRunner` interface with Codex and Copilot implementations that inject agent persona into the task prompt and execute in the provided working directory; a selector chooses a concrete runner based on model support and priority.
- Logging (`internal/logging`): zap production JSON logger.

## Control Flow
1. Client sends `initialize`; server responds with protocol version `2024-11-05`, tools capability, and server info.
2. `tools/list` returns tool metadata with JSON Schemas.
3. `tools/call` routes to handlers:
   - `list_agents`: reads YAML personas and returns JSON payload of agents.
   - `delegate_task`: validates agent name and working directory, builds persona+task prompt, invokes selected runner (preferred CLI runner if it supports the agent model; otherwise, fall back by config priority), returns final stdout text.

## Runners and Guardrails
- Codex runner: `codex --cd <workdir> --sandbox read-only --ask-for-approval never exec "<prompt>"`; activity streams to stderr, final message to stdout.
- Copilot runner: `copilot -p "<prompt>" --allow-all-tools --allow-all-paths --stream off` executed in the working directory.
- Guardrails: reject empty/relative/root paths; symlinks resolved; working directory must exist.
