## Run
Auto-select from all runners:
## Prerequisites
- Go 1.23+
- Codex CLI on PATH and authenticated (for `--runner codex`)
- GitHub Copilot CLI on PATH and authenticated (for `--runner copilot`)
Prefer Codex:

## Build
```bash
go build ./...
Prefer Copilot:

## Agents Directory
- Must be an absolute, existing directory.

Prefer Gemini:
```bash
./subagents --agents-dir /abs/path/to/agents --runner gemini
```
- Contains `*.yaml` files with `persona` and `description`; optional `model` selects a preferred model for that agent.
- Example:
  ```yaml
  persona: |
    You are a relentless documentation analyst who finds the smallest official
    excerpts needed to answer the question, cites sources, and keeps summaries
    short and precise.
  description: "Docs excerpt fetcher"
  model: "gpt-4o-mini" # optional
  ```

## Run
Codex runner (default):
```bash
./subagents --agents-dir /abs/path/to/agents
```

Copilot runner:
```bash
./subagents --agents-dir /abs/path/to/agents --runner copilot
```

Gemini runner:
```bash
./subagents --agents-dir /abs/path/to/agents --runner gemini
```

Runner config (models and priorities):
```bash
./subagents \
  --agents-dir /abs/path/to/agents \
  --runner codex \               # preferred runner, tried first if model supported
  --runner-config /abs/path/to/runner_config.yaml
```
Runner config YAML example:
```yaml
runners:
  - name: codex
    priority: 1
    models: ["gpt-4o", "gpt-4o-mini"]
  - name: copilot
    priority: 2
    models: ["gpt-4o", "claude-3-opus"]
```
Only known runners are instantiated; priorities order the fallback sequence after the preferred `--runner`.

## Runner Notes
- Codex: uses `codex --cd <workdir> --sandbox read-only --ask-for-approval never exec "<prompt>"`; stderr shows activity, stdout carries final message.
- Copilot: uses `copilot -p "<prompt>" --allow-all-tools --allow-all-paths --stream off` with `Cmd.Dir` set to the requested working directory.
- Gemini: uses `gemini -p "<prompt>" --output-format json` with `-m <model>` when provided, running from the requested working directory.
- Runner selection: leave `--runner` blank to try every configured runner in priority order. Supplying `--runner <name>` prefers that CLI first; if the requested agent model is unsupported, the server falls back to other runners ordered by `priority` in the runner config YAML.
- Usage limit fallback: if a runner returns a usage/quota limit error (e.g., "You've hit your usage limit"), the server automatically tries the next available runner. Configure multiple runners for redundancy.

## Path Guardrails
- `--agents-dir` and `working_directory` must be absolute, existing directories, and cannot be `/`; symlinks are resolved before validation.
