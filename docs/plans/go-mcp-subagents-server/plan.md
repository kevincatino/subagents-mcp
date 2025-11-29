# Go MCP Subagents Server Implementation Plan

## Overview
Build a Go 1.23 MCP server over stdio/JSON-RPC that exposes two tools (`list_agents`, `delegate_task`) backed by YAML-defined agent personas. The server validates absolute paths, uses zap for structured logs, and delegates tasks to Codex CLI in non-interactive mode while executing in the caller’s workspace directory and returning only the final text output.

## Current State Analysis
- Empty repository; greenfield implementation.
- Requirements: stdio transport, `--agents-dir` flag, YAML files (`<agent>.yaml`) with `persona` and `description`.
- Runner: Codex CLI non-interactive (`codex exec "<task>"`) with working dir via `--cd`, headless flags `--sandbox read-only --ask-for-approval never`, final text on stdout; assume user is already authenticated with Codex (no per-run API key injection).
- Guardrails: block relative paths; working directories must be absolute, exist, and not `/`.

## Desired End State
- Executable (e.g., `cmd/subagents/main.go`) parses `--agents-dir`, validates absolute existing directory.
- MCP server over stdio registers `list_agents` and `delegate_task` tools with schemas.
- Agent repository loads YAML personas/descriptions and surfaces them to `list_agents`.
- `delegate_task` routes to a pluggable `AgentRunner` interface; Codex runner implementation spawns `codex exec` in the provided workspace dir, capturing final stdout text.
- Structured JSON logging via zap across layers; errors surfaced with MCP-compliant error codes.
- Tests covering config validation, YAML parsing, runner command construction, and tool handlers.

### Key Discoveries:
- Codex non-interactive: `codex exec "<task>"` streams activity to stderr, final message to stdout; `--cd <dir>` sets working dir; headless flags `--sandbox read-only --ask-for-approval never`; assumes prior Codex authentication (no env key injection needed) (openai/codex docs/exec.md, docs/getting-started.md, docs/sandbox.md).
- YAML parsing can use `yaml.Unmarshal` from `gopkg.in/yaml.v3` / goccy/go-yaml (API parity) for simple structs.

## What We're NOT Doing
- No additional runners beyond Codex (others can plug in later).
- No network or transport beyond stdio JSON-RPC.
- No extra agent schema fields beyond `persona` and `description`.
- No filesystem writes outside validated absolute paths or root (`/`).

## Implementation Approach
- Clean architecture layering:
  - Domain: `Agent` entity, ports `AgentRepository`, `AgentRunner`, `MCPServer`.
  - Adapters: YAML-backed repository, Codex runner, stdio JSON-RPC server.
  - Interface layer: tool schemas and handlers mapping to domain interfaces.
  - Composition root in `cmd/subagents/main.go` wiring flags, logger, adapters, server.
- Defensive validation for `--agents-dir` and delegate `working_directory` (absolute, exists, not `/`, reject relative/symlink traversal as needed).

## Phase 1: CLI, Config, and Logging

### Overview
Set up entrypoint, flag parsing, logger, and path validation for `--agents-dir`.

### Changes Required:

#### 1. Entrypoint & Config
**File**: `cmd/subagents/main.go`  
**Changes**: Parse `--agents-dir`; ensure absolute/existing/non-root; initialize zap (JSON encoder); wire adapters and server start.

#### 2. Validation Helpers
**File**: `internal/validate/paths.go`  
**Changes**: Provide utilities to resolve absolute paths, reject relative, ensure existence, and block `/`.

### Success Criteria:

#### Automated Verification:
- [ ] Build succeeds: `go build ./...`
- [ ] Vet passes: `go vet ./...`
- [ ] Unit tests for validation/config: `go test ./...`

#### Manual Verification:
- [ ] Running `./subagents --agents-dir <abs path>` starts without errors and logs structured JSON.
- [ ] Passing a relative or root path is rejected with clear error.

---

## Phase 2: Agent Repository (YAML)

### Overview
Implement YAML-backed `AgentRepository` loading `<agent>.yaml` with `persona` and `description`.

### Changes Required:

#### 1. Agent Domain and Repository
**File**: `internal/agents/model.go`  
**Changes**: Define `Agent` struct (`Name`, `Persona`, `Description`), validation.

**File**: `internal/agents/yaml_repository.go`  
**Changes**: Load all `*.yaml` files from `--agents-dir`; parse with `yaml.Unmarshal`; return slice of agents; handle malformed files with detailed errors.

#### 2. Tests
**File**: `internal/agents/yaml_repository_test.go`  
**Changes**: Table tests covering valid load, missing fields, malformed YAML, empty directory.

### Success Criteria:

#### Automated Verification:
- [ ] Repo tests pass: `go test ./...`
- [ ] Vet passes: `go vet ./...`

#### Manual Verification:
- [ ] Sample agents directory with two YAML files returns both agents via repository API.
- [ ] Malformed YAML surfaces descriptive error.

---

## Phase 3: MCP Server and Tools

### Overview
Expose stdio JSON-RPC MCP server with `list_agents` and `delegate_task`.

### Changes Required:

#### 1. Tool Schemas and Server
**File**: `internal/mcp/server.go`  
**Changes**: Register tools with schemas; wire handlers; run over stdio; emit zap logs.

**File**: `internal/mcp/types.go`  
**Changes**: Define request/response DTOs for tool inputs/outputs and errors.

#### 2. Tool Handlers
**File**: `internal/mcp/handlers.go`  
**Changes**: `list_agents` pulls from repository; `delegate_task` validates args (agent exists, working_directory absolute/existing/non-root), then calls runner; return final text.

### Success Criteria:

#### Automated Verification:
- [ ] Handler unit tests: `go test ./...`
- [ ] JSON schemas match required fields (agent, task, working_directory).

#### Manual Verification:
- [ ] `tools/list` returns agents from sample directory.
- [ ] `tools/call` with valid agent and workspace returns Codex final text; invalid paths reject with clear MCP error.

---

## Phase 4: Agent Runner (Codex)

### Overview
Implement `AgentRunner` interface and Codex CLI adapter for non-interactive exec.

### Changes Required:

#### 1. Runner Interface and Adapter
**File**: `internal/runner/runner.go`  
**Changes**: Define `Run(ctx, agent, task, workdir) (string, error)` interface.

**File**: `internal/runner/codex.go`  
**Changes**: Build `codex exec "<task>" --cd <workdir> --sandbox read-only --ask-for-approval never`; assume Codex auth already configured; capture stdout (final text) and stderr (logs); enforce working dir guardrails; return stdout string.

#### 2. Tests
**File**: `internal/runner/codex_test.go`  
**Changes**: Command-building tests (no exec) and guardrail tests.

### Success Criteria:

#### Automated Verification:
- [ ] Runner unit tests: `go test ./...`
- [ ] Vet passes: `go vet ./...`

#### Manual Verification:
- [ ] With valid Codex authentication and workspace, delegate runs and returns final text only.
- [ ] stderr captures progress logs; stdout contains only final answer.
- [ ] Relative or root workdir is rejected.

---

## Phase 5: Observability and Polishing

### Overview
Ensure consistent logging, error handling, and developer ergonomics.

### Changes Required:

#### 1. Logging and Errors
**File**: `internal/logging/logger.go`  
**Changes**: zap production JSON config; structured fields (tool, agent, cwd, duration, error).

**File**: `internal/mcp/errors.go`  
**Changes**: Helpers for MCP error responses with integer codes and safe messages.

#### 2. Developer Docs
**File**: `README.md`  
**Changes**: Usage examples for `--agents-dir`, sample YAML, MCP tools contract, Codex runner env requirements.

### Success Criteria:

#### Automated Verification:
- [ ] `go test ./...` passes.
- [ ] `go vet ./...` passes.

#### Manual Verification:
- [ ] Logs are JSON with contextual fields.
- [ ] Error responses are informative without leaking sensitive data.

---

## Testing Strategy

### Unit Tests:
- Path validation (absolute, exists, not root).
- YAML repository parsing and error handling.
- Tool handlers (happy path and invalid inputs) with mocked repository/runner.
- Runner command construction.

### Integration Tests:
- Optional: lightweight in-process stdio server test with mock runner/repo to verify tool contract.

### Manual Testing Steps:
1. Create sample agents dir with two YAML files; run server with `--agents-dir <abs path>`.
2. Call `tools/list` via MCP client and verify agents.
3. Call `delegate_task` with valid workspace and task; observe final text on stdout.
4. Try relative and root working_directory to confirm rejection.

## Performance Considerations
- Small footprint; main cost is Codex subprocess. Stream stderr to avoid buffering; bound stdout capture.

## Migration Notes
- None (greenfield). Future: allow hot-reload of agent YAMLs if needed.

## Documentation Requirements
- Include in README: CLI usage, YAML schema, working dir guardrails, Codex requirements (assumed authenticated), `codex exec` flags, MCP tool schemas.
- External references to fetch/maintain:
  - Codex non-interactive exec flags (`codex exec`, `--cd`, `--sandbox read-only`, `--ask-for-approval never`, `-o/--output-last-message`, `--json`); authentication assumed preconfigured.
  - MCP tool/list/call shapes (JSON-RPC 2.0).
  - Go YAML parsing (`gopkg.in/yaml.v3`).

## References
- Codex CLI docs: `openai/codex` `docs/exec.md`, `docs/getting-started.md`, `docs/sandbox.md`.
- MCP spec: `modelcontextprotocol.io/specification/2024-11-05/server/tools`.
