#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ -z "${GOMODCACHE:-}" ]]; then
  export GOMODCACHE="$ROOT_DIR/.gomodcache"
  export GOPATH="$ROOT_DIR/.gopath"
  export GOCACHE="$ROOT_DIR/.gocache"
fi

AGENTS_DIR="${1:-"$ROOT_DIR/examples/agents"}"

REQ='{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}'

printf '%s\n' "$REQ" | go run ../cmd/subagents --agents-dir "$AGENTS_DIR"
