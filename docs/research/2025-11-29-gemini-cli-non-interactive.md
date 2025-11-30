---
date: 2025-11-29T21:48:28-03:00
researcher: context7-doc-fetcher
git_commit: 75de5cadf3cd711e550f1e00172d3aa6c022b204
branch: main
repository: subagents-mcp
topic: "how to use gemini-cli in non interactive mode and specify which model to use via cli arguments"
tags: [research, external-integration, gemini-cli]
status: complete
last_updated: 2025-11-29
last_updated_by: context7-doc-fetcher
---

# Research: how to use gemini-cli in non interactive mode and specify which model to use via cli arguments

**Date**: 2025-11-29T21:48:28-03:00  
**Researcher**: context7-doc-fetcher  
**Git Commit**: 75de5cadf3cd711e550f1e00172d3aa6c022b204  
**Branch**: main

## Summary
Gemini CLI supports fully scripted workflows by accepting prompts directly on the command line (or via piping) and emitting structured JSON output. The same invocation lets you pin a specific model by passing `-m <alias|model-id>`, so you can control both the prompt and the underlying generative model from automation or CI scripts.

## Non-interactive invocation
- Use `-p/--prompt` to send the user prompt inline, e.g., `gemini -p "Describe how to test CLI automation"`; this keeps the process headless and suitable for shell scripts.  
- Alternatively, pipe the prompt through standard input (`echo "Explain this code" | gemini`) when the prompt originates from another command or file.  
- Add `--output-format json` (or `stream-json`) to receive a machine-readable payload containing both the response and metadata, which simplifies downstream parsing for automation tools.  
Sources: Google Gemini CLI docs (`docs/cli/index.md`, `docs/cli/headless.md`).

## Model selection via CLI flag
- Supply `-m <model>` to fix the CLI to a named alias or explicit model identifier (for example, `gemini -m gemini-2.5-flash`). Aliases configured in `settings.json` are honored, so teams can define human-friendly shortcuts for their preferred models.  
Sources: Google Gemini CLI docs (`docs/cli/headless.md`).

## Automation tips
- Combine `-p` with `--output-format json` to script multi-step flows: the CLI stays non-interactive while providing structured responses that can be piped into parsers or logged with telemetry.  
- When reusing prompts inside scripts, store them in files and pipe them (`cat prompt.txt | gemini --output-format json -m gemini-2.5-flash`) to keep command lines readable.  

## Sources
- Google Gemini CLI docs — `docs/cli/index.md` (prompt flag documentation).  
- Google Gemini CLI docs — `docs/cli/headless.md` (headless mode, piping, output formats, and `-m` flag examples).
