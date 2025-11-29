#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ -z "${GOMODCACHE:-}" ]]; then
  export GOMODCACHE="$ROOT_DIR/.gomodcache"
  export GOPATH="$ROOT_DIR/.gopath"
  export GOCACHE="$ROOT_DIR/.gocache"
fi

AGENTS_DIR="${1:-"$ROOT_DIR/examples/agents"}"
WORKDIR="${2:-"$ROOT_DIR"}"
AGENT_NAME="${3:-"docs-fetcher"}"
TASK="${4:-"say hello from codex"}"
RUNNER="${5:-"copilot"}"

REQ=$(cat <<EOF
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"delegate_task","arguments":{"agent":"$AGENT_NAME","task":"$TASK","working_directory":"$WORKDIR"}}}
EOF
)

printf '%s\n' "$REQ" | go run ../cmd/subagents --agents-dir "$AGENTS_DIR" --runner "$RUNNER"
