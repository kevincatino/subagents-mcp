# Decisions

- MCP protocol version responded as `2024-11-05`; tool list limited to `list_agents` and `delegate_task`.
- Default runner is Codex CLI with read-only sandbox and `--ask-for-approval never` for non-interactive delegation; Copilot runner available via flag.
- Path guardrails enforced: agents directory and delegate working directory must be absolute, existing, non-root directories; symlinks are resolved.
