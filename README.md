# Subagents MCP Server (Go)

## Overview
A Go 1.23 MCP server over stdio/JSON-RPC exposing two tools:
- `list_agents`: reads YAML personas from `--agents-dir`.
- `delegate_task`: runs a task via Codex CLI in the caller’s workspace, returning final text output.

## Setup
```bash
go build ./...
```

## Running
```bash
./subagents --agents-dir /absolute/path/to/agents
```

The `--agents-dir` must be an absolute, existing directory containing `<agent>.yaml` files:
```yaml
persona: "docs-focused researcher"
description: "Fetches minimal excerpts from official docs."
```

## Tool Contracts
- `tools/list` → returns tool metadata.
- `tools/call` with `name: "list_agents"` → `{"agents":[{name,persona,description},...]}`.
- `tools/call` with `name: "delegate_task"` and `arguments`:
  ```json
  {
    "agent": "docs-fetcher",
    "task": "summarize latest release notes",
    "working_directory": "/absolute/workspace/path"
  }
  ```
  Returns `{"content":[{"type":"text","text":"<final output>"}]}`.

## Guardrails
- `--agents-dir` and `working_directory` must be absolute, exist, and cannot be `/`.
- Relative paths are rejected.

## Dependencies
- Codex CLI available on PATH and authenticated (non-interactive via `codex exec`).
- Go modules: zap (structured JSON logs), gopkg.in/yaml.v3 (agent parsing).
