# Plan: Add Gemini CLI runner to MCP server

## Objective
Enable the MCP server to delegate tasks through the Gemini CLI in exactly the same JSON-RPC tooling layer as the existing Codex/Copilot runners, letting callers pick Gemini via the `--runner` flag and pin the `-m` model override from agent metadata.

## Background & constraints
- The server exposes `delegate_task` and `list_agents` via `internal/mcp`, uses `runner.AgentRunner` implementations under `internal/runner`, and wires runners through `cmd/subagents/main.go` (runner flag defaults to `codex`).  
- Each runner needs to construct a prompt from `internal/runner/prompt.go`, launch an external CLI via `os/exec`, and stream the output back to the MCP response.  
- `runner.Selector` currently supports `codex` and `copilot` plus a YAML `runner-config` for priorities/models; new Gemini runner must plug into that selector.  
- Research notes in `docs/research/2025-11-29-gemini-cli-non-interactive.md` confirm Gemini accepts `-p/--prompt`, `--output-format json|stream-json`, and `-m <model|alias>` for non-interactive flows.

## Scope
1. Add a `GeminiRunner` that:
   - Validates the working directory and agent/task data like other runners.
   - Builds the prompt string via `buildAgentPrompt`.
   - Executes the `gemini` binary with `-p <prompt>` (or `--prompt`) plus automation-friendly flags (`--output-format json` or `stream-json`, `-m <model>` when provided, and any sandbox/credentials flags needed).  
   - Detects usage/quota-limit output (extend `internal/runner/errors.go` with Gemini-specific patterns) so the selector can fall back.  
   - Mirrors log/metrics instrumentation already emitted by `CodexRunner`.

2. Wire Gemini into the runner framework:
   - Extend `runnerFactories` (and any tests) so `"gemini"` maps to the new runner and accepts supported models from config.
   - Update `cmd/subagents/main.go` to document `--runner gemini` and adjust validation/messages if necessary.
   - Ensure `runner.Selector` sees Gemini entries via YAML config (repositories should list `name: gemini`, `priority`, `models`), and verify `supportsModel` behavior works the same way.

3. Docs & docs/plans coverage:
   - Update `README.md` (and `docs/setup.md` if needed) to explain the new runner, the `gemini` binary requirement, and the `-m` flag plus automation suggestions (non-interactive prompt, JSON output). Reference the research note as the source for command-line details.  
   - Optionally document how to add Gemini to `docs/plans` or `docs/research` (this plan already tracks the work).

4. Testing & validation:
   - Add unit tests in `internal/runner` for `GeminiRunner` command construction (mocking `execCommand`) and usage limit detection.
   - Extend `selector_test.go` (and others if needed) to ensure Gemini runner appears in candidate list, respects config models, and yields clear errors when the binary is missing or the model is unsupported.  
   - Run `go test ./...` locally once the runner is implemented to catch integration issues with new dependencies.

## Unknowns / follow-ups
- Confirm whether Gemini CLI needs authentication/environment setup similar to Codex/Copilot (e.g., `GOOGLE_APPLICATION_CREDENTIALS`), and document any prerequisites.  
- Decide whether `GeminiRunner` should use `--output-format json` vs `stream-json` based on downstream parser expectations; the research note recommends JSON for automation, so start with that.  
- Determine model alias mapping strategy (do we rely purely on agent `model` strings, or should we let the `settings.json` alias mapping be configurable?).  

## Status
- [x] Added `GeminiRunner`, selector wiring, usage-limit detection, and unit tests.
- [x] Updated README, `docs/setup.md`, and `docs/modules.md` to describe the new runner and prerequisites.
