---
date: 2025-12-01T12:51:00-05:00
researcher: GitHub Copilot
git_commit: 0910aec213e5c1eaba560ce0d0f976eb25457042
branch: main
repository: subagents-mcp
topic: "Cursor agent CLI non-interactive usage"
tags: [research, external, cursor, cli]
status: complete
last_updated: 2025-12-01
last_updated_by: GitHub Copilot
---

## Summary
Cursor’s `cursor-agent` binary can run entirely headless by pairing print mode (`-p`/`--print` or by writing the prompt inline) with automation-friendly flags. Scripts typically supply the prompt, desired model, and output format in a single command while authenticating via `CURSOR_API_KEY`. Output control (`--output-format text|json|stream-json`, `--stream-partial-output`) keeps CI logs readable or machine-parseable. Model selection works either through the documented `--model <name>` flag (e.g., `gpt-5`, `sonnet-4`) or via the `/model <name>` slash command inside sessions when you need to retarget providers interactively.

## Methodology
- Queried the Context7 documentation fetcher for "Cursor Agent CLI" automation guidance.
- Used the sanctioned web-search agent to pull the latest public docs on headless usage, authentication, output formats, and model selection.
- Cross-referenced overlapping sections (overview, headless mode, output format reference, authentication reference) to reconcile flag availability and slash-command behavior.

## Detailed Findings

### Non-interactive / headless execution
- Official headless guide recommends `cursor-agent -p "<prompt>" --model "gpt-5" --output-format text` for scripts so the CLI never opens an interactive TUI. `-p`/`--print` triggers immediate response printing; adding `--force` applies suggested edits automatically (otherwise they remain advisory). [Cursor CLI overview](https://cursor.com/docs/cli/overview); [Headless CLI](https://cursor.com/docs/cli/headless).
- Prompts may reference local files (`cursor-agent -p "Review ./main.go"`), and the agent will load them as context. Optional streaming flags (`--output-format stream-json --stream-partial-output`) provide NDJSON event streams for dashboards or long-running tasks. [Headless CLI](https://cursor.com/docs/cli/headless).

### Key arguments & options
- `-p/--prompt` (alias `--print` in headless docs) supplies the task text directly, enabling batch scripts/CI to run without human input. [Headless CLI](https://cursor.com/docs/cli/headless).
- `--model <name>` pins the backend model; documented examples: `gpt-5`, `sonnet-4`, `opus-4.1`, `grok`. The same names can be issued via `/model <name>` inside a session. [CLI overview](https://cursor.com/cli); [Context7 doc excerpt](https://cursor.com/docs/cli/overview).
- `--output-format text|json|stream-json` controls the response structure (text for humans, single JSON blob for parsing, or NDJSON event stream). `--stream-partial-output` streams incremental assistant tokens when paired with `stream-json`. [Output format reference](https://cursor.com/docs/cli/reference/output-format).
- `--force` lets CLI edits write to disk automatically (otherwise they are proposed). [Headless CLI](https://cursor.com/docs/cli/headless).
- Additional automation switches mentioned: `--allow-tools/--allow-all-tools` (mirroring other CLIs) and environment-detection that auto-enables print mode when stdout is not a TTY; confirm via `cursor-agent --help` in your environment for the latest list.

### Authentication & environment variables
- Browser login flow: `cursor-agent login`, `cursor-agent status`, `cursor-agent logout`. [Authentication reference](https://cursor.com/docs/cli/reference/authentication).
- Non-interactive auth: export `CURSOR_API_KEY` before invoking the CLI or pass `--api-key <value>` per call. This avoids login prompts during CI/CD runs. [Authentication reference](https://cursor.com/docs/cli/reference/authentication).
- Troubleshooting tips include re-running login if “Not authenticated” appears and using `--endpoint`/`--insecure` for custom installations.

### Model selection specifics
- Slash commands allow on-the-fly switching: `/model auto`, `/model gpt-5`, `/model sonnet-4`, `/model opus-4.1`, `/model grok`. This is useful when already in an interactive session or when a script sends pre-run commands. [Cursor CLI page](https://cursor.com/cli).
- Context7’s excerpt highlights a dedicated `--model` flag for headless usage; treat this flag as authoritative when available and fall back to issuing `/model ...` followed by the real prompt if running an older binary that lacks the flag.

## Open Questions / Follow-ups
- Confirm your installed `cursor-agent --help` output to ensure the `--model` flag (and any other automation-specific flags like `--allow-tools`) exists in the shipped version.
- Monitor Cursor release notes for expanding model name support or additional output formats.
