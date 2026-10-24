#!/usr/bin/env bash
# mcp_tool_smoke.sh — verifies the four graph tools respond correctly.
# Dynamically discovers entities from the active workspace (no hardcoded names).

set -euo pipefail

cd "$(dirname "$0")/.."

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

if [[ ! -x bin/garuda-mcp ]]; then
  echo "Building bin/garuda-mcp..."
  go build -o bin/garuda-mcp ./cmd/garuda-mcp
fi

export DATABASE_URL="${DATABASE_URL:-postgres://test:test@localhost:5433/garuda_test?sslmode=disable}"
WORKSPACE="${WORKSPACE:-${GARUDA_WORKSPACE:-}}"

call() {
  local tool="$1"
  local args="$2"
  echo "═══ $tool ═══"
  echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"$tool\",\"arguments\":$args}}" \
    | GARUDA_MCP_QUIET=1 timeout 5 ./bin/garuda-mcp 2>/dev/null \
    | python3 -m json.tool
  echo
}

# 1. find_entity — resolve the top hub entity in the workspace dynamically.
FIND_RESP=$(echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"garuda.find_entity\",\"arguments\":{\"workspace\":\"$WORKSPACE\",\"limit\":3}}}" \
  | GARUDA_MCP_QUIET=1 timeout 5 ./bin/garuda-mcp 2>/dev/null)

echo "═══ garuda.find_entity ═══"
echo "$FIND_RESP" | python3 -m json.tool

ENTITY_ID=$(echo "$FIND_RESP" | python3 -c '
import sys, json
d = json.load(sys.stdin)
payload = json.loads(d["result"]["content"][0]["text"])
if not payload.get("results"):
    sys.exit("no entity found in workspace")
print(payload["results"][0]["id"])
')
echo
echo "resolved entity_id: $ENTITY_ID"
echo

# 2. subclasses of the resolved entity (or top inherited entity).
call "garuda.subclasses" "{\"workspace\":\"$WORKSPACE\",\"entity_id\":\"$ENTITY_ID\",\"limit\":5}"

# 3. neighbors of the resolved entity.
call "garuda.neighbors" "{\"workspace\":\"$WORKSPACE\",\"entity_id\":\"$ENTITY_ID\",\"limit\":10}"

# 4. implementers of top interface in the workspace.
call "garuda.implementers" "{\"workspace\":\"$WORKSPACE\",\"name\":\"Router\",\"limit\":5}"

echo "═══ smoke test complete ═══"
