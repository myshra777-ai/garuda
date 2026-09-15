#!/usr/bin/env bash
# mcp_tool_smoke.sh — verifies the four graph tools respond correctly.
#
# Read-only. Sends JSON-RPC sessions to the local MCP binary and
# prints every response. No database writes.
#
# Usage:
#   cd ~/garuda
#   ./scripts/mcp_tool_smoke.sh

set -euo pipefail

cd "$(dirname "$0")/.."

if [[ ! -x bin/garuda-mcp ]]; then
  echo "bin/garuda-mcp not built. Run:" >&2
  echo "  go build -o bin/garuda-mcp ./cmd/garuda-mcp" >&2
  exit 1
fi

export DATABASE_URL="${DATABASE_URL:-postgres://test:test@localhost:5433/garuda_test?sslmode=disable}"
WORKSPACE="${WORKSPACE:-go-validation-10}"

call() {
  local tool="$1"
  local args="$2"
  echo "═══ $tool ═══"
  echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"$tool\",\"arguments\":$args}}" \
    | GARUDA_MCP_QUIET=1 timeout 5 ./bin/garuda-mcp 2>/dev/null \
    | python3 -m json.tool
  echo
}

# 1. find_entity — resolve a name to an id, ranked by inbound edges.
FIND_RESP=$(echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"garuda.find_entity\",\"arguments\":{\"workspace\":\"$WORKSPACE\",\"name_pattern\":\"^SQLModel$\",\"limit\":3}}}" \
  | GARUDA_MCP_QUIET=1 timeout 5 ./bin/garuda-mcp 2>/dev/null)

echo "═══ garuda.find_entity ═══"
echo "$FIND_RESP" | python3 -m json.tool

ENTITY_ID=$(echo "$FIND_RESP" | python3 -c '
import sys, json
d = json.load(sys.stdin)
payload = json.loads(d["result"]["content"][0]["text"])
if not payload["results"]:
    sys.exit("no entity matched name_pattern")
print(payload["results"][0]["id"])
')
echo
echo "resolved entity_id: $ENTITY_ID"
echo

# 2. subclasses of the resolved entity.
call "garuda.subclasses" "{\"workspace\":\"$WORKSPACE\",\"entity_id\":\"$ENTITY_ID\",\"limit\":5}"

# 3. neighbors of the resolved entity, defaults (no external stubs, one hop).
call "garuda.neighbors" "{\"workspace\":\"$WORKSPACE\",\"entity_id\":\"$ENTITY_ID\",\"limit\":10}"

# 4. implementers of a known interface in the workspace.
call "garuda.implementers" "{\"workspace\":\"$WORKSPACE\",\"name\":\"CmdTyper\",\"limit\":5}"

echo "═══ smoke test complete ═══"
