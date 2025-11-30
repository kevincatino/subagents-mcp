# Subagents MCP Server (Go)

Go 1.23 MCP server over stdio/JSON-RPC exposing two tools backed by YAML-defined personas and pluggable runners (Codex CLI or Copilot CLI).

## Overview
- Tools: `list_agents` and `delegate_task` registered on `tools/list` and `tools/call`.
- Runners: `--runner codex` (default) uses `codex exec` with read-only sandbox + `--ask-for-approval never`; `--runner copilot` uses `copilot -p "<prompt>" --stream off`; `--runner gemini` uses `gemini -p "<prompt>" --output-format json`.
- Agent source: YAML files in an absolute `--agents-dir`; each file defines `persona` and `description`.
- Guardrails: absolute, existing, non-root paths for agents dir and delegate working directory; relative paths are rejected.
- Protocol: MCP 2024-11-05 initialize response with server info and tools capability.

## Project Structure
- `cmd/subagents` – entrypoint parsing flags and wiring server.
- `internal/agents` – agent model and YAML repository loader.
- `internal/mcp` – JSON-RPC handlers, tool schemas, server loop, MCP errors.
- `internal/runner` – agent runner interface plus Codex and Copilot implementations.
- `internal/validate` – path validation helpers (absolute, exists, non-root).
- `internal/logging` – zap logger setup.
- `examples/agents` – sample agent YAMLs.

## Installation & Setup
```bash
go build ./...
```

Prereqs:
- Go 1.23+
- Codex CLI on PATH and authenticated (for `--runner codex`)
- GitHub Copilot CLI on PATH and authenticated (for `--runner copilot`)
- Gemini CLI on PATH and authenticated (for `--runner gemini`)

## Usage
Run the server (Codex runner default):
```bash
./subagents --agents-dir /abs/path/to/agents
```

Use Copilot runner:
```bash
./subagents --agents-dir /abs/path/to/agents --runner copilot
```

Use Gemini runner:
```bash
./subagents --agents-dir /abs/path/to/agents --runner gemini
```

Sample agent file (`/abs/path/to/agents/docs-fetcher.yaml`):
```yaml
persona: |
  You are a relentless documentation analyst who finds the smallest official
  excerpts needed to answer the question, cites sources, and keeps summaries
  short and precise.
description: "Docs excerpt fetcher"
```

Delegate contract:
- `tools/call` with `name: "delegate_task"` and arguments:
  ```json
  {
    "agent": "docs-fetcher",
    "task": "summarize latest release notes",
    "working_directory": "/absolute/workspace/path"
  }
  ```
  Returns `{"content":[{"type":"text","text":"<final output>"}]}`.
- `tools/call` with `name: "list_agents"` returns `{"content":[{"type":"text","text":"{\"agents\":[...]}"}]}` (JSON string of `name` and `description` only).
- `tools/call` with `name: "expand_prompt"` (also referenced as `prompt_expansion`) and arguments:
  ```json
  {
    "command": "research",
    "input": "LLM routing frameworks"
  }
  ```
  Returns `{"prompt":"<expanded prompt text>"}` and that prompt must be treated as an explicit instruction from the user that should be followed verbatim.

Path rules:
- `--agents-dir` and `working_directory` must be absolute, existing directories and cannot be `/`; symlinks are resolved.

## Architecture
Brief overview lives in `docs/architecture.md`.

## Documentation
- `docs/architecture.md`
- `docs/api.md`
- `docs/modules.md`
- `docs/setup.md`
- `docs/research/` (none yet)
- `docs/decisions.md`

## Tech Stack
- Go 1.23
- zap for structured JSON logs
- gopkg.in/yaml.v3 for agent parsing
- Codex CLI (read-only exec), GitHub Copilot CLI (non-interactive prompt mode), and Gemini CLI (JSON output mode)
