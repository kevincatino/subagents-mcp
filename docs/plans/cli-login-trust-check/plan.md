# Plan: Add CLI Login & Trust Verification Script

## Objective
Ship a portable shell script (under `scripts/`) that exercises every supported runner CLI (Codex, GitHub Copilot, Gemini) so maintainers can quickly confirm two pre-flight requirements before starting the MCP server:
1. Each CLI binary is installed and authenticated.
2. The current working directory is already whitelisted/trusted so future non-interactive runs will not block on approvals.

## Current State & Constraints
- The MCP server simply shells out to `codex`, `copilot`, or `gemini` (docs/setup.md, internal/runner/*.go); authentication and trust configuration must be handled ahead of time.
- Codex exposes an `account/read` JSON-RPC method and trusts directories listed in `~/.codex/config.toml` via `[projects."<abs-path>"] trust_level = "trusted"` (context7 Codex docs).
- GitHub Copilot CLI can run non-interactively with `--allow-all-paths`/`--allow-tool`, but does not surface a single “auth status” flag; `/user show` within the interactive session is the documented confirmation step (context7 Copilot docs).
- Gemini CLI relies on cached OAuth/API-key credentials and `settings.json` keys such as `security.folderTrust.enabled` + `includeDirectories` (context7 Gemini docs).
- No helper script exists today; `scripts/test-delegate-task.sh` assumes CLIs are already configured.

## Proposed Implementation
1. **Script scaffolding**
   - Create `scripts/verify-runner-clis.sh` (POSIX bash, `set -euo pipefail`).
   - Accept optional `--workdir <path>` (default: `pwd`) and normalize to an absolute path via `realpath`.
   - Emit structured sections per CLI (prefixed with emoji/headers) to make manual review easy.

2. **Shared helpers**
   - `require_cmd <name>`: fail fast if `command -v` is missing.
   - `print_status <cli> <message>`: consistent messaging.
   - `resolve_workdir`: ensure directory exists + display what will be tested (reuse `validate.Dir` semantics by shelling out to `python - <<'PY'` or re-implement path checks in bash).

3. **Codex checks**
   - **Auth**: invoke `codex rpc account/read` (or comparable command) and parse JSON via `python -m json.tool`/`jq` fallback. Success criterion: `result.account` present; otherwise tell the user to run `codex login`.
   - **Trust**: read `~/.codex/config.toml` (if missing, warn). Use an inline Python/TOML snippet (`tomllib` in Python 3.11+) to look for `[projects."<resolved workdir>"] trust_level == "trusted"`. If absent, show the TOML block the user should add.
   - **Dry-run**: fire `codex --cd "$WORKDIR" --sandbox read-only --ask-for-approval never exec "echo trust-check"` to ensure the CLI can enter the workspace without pausing. Bubble stderr if Codex refuses due to trust/auth so the user sees the exact prompt.

4. **Copilot checks**
   - **Auth**: launch a no-op non-interactive run (`copilot -p "echo trust-check" --allow-all-tools --allow-all-paths --stream off`) from the target directory. A successful exit implies the logged-in user/token works; failures will include “Please run /login” or GitHub auth errors per docs.
   - **Trust**: Copilot only honors per-path allowlists through its CLI flags, so the script should report whether `--allow-all-paths` was needed (i.e., document that the MCP runner uses the same flag set). If a safer check is desired, optionally open an interactive session (`copilot -p ""`) and instruct the user to run `/user show` + `/allow` commands manually; note this in output, as there is no non-interactive status command.

5. **Gemini checks**
   - **Auth**: run `gemini -p "echo trust-check" --output-format json` inside the workdir. Capture JSON response and verify it parses; failures surface auth issues (“Login with Google” prompt or `GEMINI_API_KEY` errors).
   - **Trust**: inspect `~/.config/gemini/settings.json` (path from docs) for `security.folderTrust.enabled` and `includeDirectories` entries; warn if folder trust is disabled or the workdir is missing. Provide the JSON snippet to append when needed.

6. **User guidance & exit codes**
   - Exit success only if all CLIs both executed the dry-run command and passed trust file checks.
   - On failures, print actionable remediation steps (e.g., `codex login`, sample TOML/JSON entries, instructions to run `/user show` inside Copilot).
   - Offer `--skip <cli>` flag (optional) for environments that only install a subset of runners.

7. **Documentation updates**
   - Reference the new script in `README.md` and `docs/setup.md` under the prerequisites section so operators know how to verify their environment.
   - Add a short note in `docs/modules.md` (runner section) pointing to the script for manual verification.

## Validation
- Manual test each CLI happy-path on a machine with all three CLIs configured; capture sample output to ensure JSON parsing succeeds.
- Exercise failure paths by temporarily moving credential files or editing trust configs to confirm the script surfaces actionable errors.
- Run `shellcheck` locally (if available) to catch scripting issues.

## Risks / Open Questions
- **Copilot auth status**: there is no documented headless command that reports the signed-in user; the best practical check is running a harmless command and interpreting failure messages. Confirm this limitation with the team or explore whether `copilot --status` exists in newer builds.
- **Config file locations**: paths such as `~/.codex/config.toml` or `~/.config/gemini/settings.json` may vary across OSes; the script should either detect common variants or accept env overrides (e.g., `CODEX_CONFIG`, `GEMINI_CONFIG`).
- **Local Python/TOML availability**: parsing TOML/JSON via Python assumes macOS ships with Python 3.11+ (for `tomllib`). If unavailable, vendor a small Go helper or fallback to `python3 -c 'import toml'` with a dependency.
- **CLI prompts breaking automation**: If any CLI prints interactive login URLs instead of failing fast, the script should detect a stalled subprocess (via `timeout` or background read) to avoid hanging the checks.
