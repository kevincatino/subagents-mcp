# Setup

## Prerequisites
- Go 1.23+
- Codex CLI on PATH and authenticated (for `--runner codex`)
- GitHub Copilot CLI on PATH and authenticated (for `--runner copilot`)

## Build
```bash
go build ./...
```

## Agents Directory
- Must be an absolute, existing directory.
- Contains `*.yaml` files with `persona` and `description`.
- Example:
  ```yaml
  persona: |
    You are a relentless documentation analyst who finds the smallest official
    excerpts needed to answer the question, cites sources, and keeps summaries
    short and precise.
  description: "Docs excerpt fetcher"
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

## Runner Notes
- Codex: uses `codex --cd <workdir> --sandbox read-only --ask-for-approval never exec "<prompt>"`; stderr shows activity, stdout carries final message.
- Copilot: uses `copilot -p "<prompt>" --allow-all-tools --allow-all-paths --stream off` with `Cmd.Dir` set to the requested working directory.

## Path Guardrails
- `--agents-dir` and `working_directory` must be absolute, existing directories, and cannot be `/`; symlinks are resolved before validation.
