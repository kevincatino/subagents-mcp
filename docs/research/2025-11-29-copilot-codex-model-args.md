---
date: 2025-11-29T21:17:10+0000
researcher: context7-doc-fetcher
git_commit: 69d0a1fa53b83a8ba637a00a4c268084612a9abf
branch: main
repository: subagents-mcp
topic: "Passing explicit model flags to Copilot and Codex CLIs in non-interactive runs"
tags: [research,codebase,external-integration,copilot,codex,runner]
status: complete
last_updated: 2025-11-29
last_updated_by: context7-doc-fetcher
---

# Research: Passing model flags to Copilot and Codex CLIs

**Date**: 2025-11-29T21:17:10+0000  
**Researcher**: context7-doc-fetcher  
**Git Commit**: 69d0a1fa53b83a8ba637a00a4c268084612a9abf  
**Branch**: main

## Question

How should we pass the requested model to both the GitHub Copilot CLI and the OpenAI Codex CLI for non-interactive `delegate_task` runs, and does the current `AgentRunner` implementation already inject those flags?

## External documentation

1. **GitHub Copilot CLI** can pin the runtime model with `copilot --model <model-id> "<prompt>"` in non-interactive mode and `/model <model-id>` inside sessions; the CLI also supports `~/.copilot/config` entries such as `{"model":"claude-sonnet-4.5"}` for a persisted default. (Sources: https://context7.com/github/copilot-cli/llms.txt)
2. **Codex CLI**’s `exec` command accepts `codex exec --model <model-id> "<prompt>"` (or the top-level `codex --model <model-id>` flag) to override whatever default model is configured, including structured `--json` or `--last` invocations. (Sources: https://github.com/openai/codex/blob/main/docs/exec.md and https://github.com/openai/codex/blob/main/docs/config.md)

These flags are the documented, non-interactive way to guarantee the exact model ID (e.g., `claude-haiku-4.5`, `gpt-5.1-codex-max`) is used for the run rather than relying on defaults.

## Codebase verification

- `internal/runner/copilot.go:33-79` builds `copilot` arguments with `-p <prompt>`, `--allow-all-tools`, `--allow-all-paths`, and `--stream off`, but it never appends `--model <model>` even though `CopilotRunner.Run` receives the requested `model` string and logs it. As a result, Copilot always runs with its default model rather than honoring the agent’s model.
- `internal/runner/codex.go:33-78` follows the same pattern: it enforces guardrails, validates the model against its configured support set, constructs `codex exec` args, and omits any `--model` flag before executing the CLI. Consequently, the Codex runner also ignores the desired model at the CLI level despite performing selector-side filtering.
- `internal/runner/selector.go:97-114` ensures only runners that advertise support for the requested model are considered, but once a candidate runner is selected it never forwards the model to the CLI command itself.

## Implications & next steps

1. The documented CLI hooks (`copilot --model` and `codex exec --model`) are the right surface for pinning the model in non-interactive runs, so we should append those flags to `args` before invoking the binaries.
2. Implement both runners (Copilot and Codex) to only add `--model <model>` when the model string is non-empty; keep the existing selector guardrails for validation.
3. Once the flags are in place, the runners will actually execute the requested models rather than just logging the preference.
