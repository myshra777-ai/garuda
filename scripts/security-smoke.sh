#!/usr/bin/env bash
# Black-box security smoke test for a running Garuda dashboard/API.
# Creates throwaway accounts on the server you point it at. Use a TEST instance, never production.
#   BASE=http://127.0.0.1:8080 scripts/security-smoke.sh
# Optional: GARUDA_BIN=./bin/garuda DATABASE_URL=... seeds data for the second tenant so leaks are visible.
set -u
BASE="${BASE:-http://127.0.0.1:8080}"
TMP="$(mktemp -d)"; trap 'rm -rf "$TMP"' EXIT
RUN="$(date +%s)$RANDOM"
PASS="Smoke-Test-${RUN}"
pass=0; fail=0
ok()   { echo "  PASS  $1"; pass=$((pass+1)); }
bad()  { echo "  FAIL  $1"; fail=$((fail+1)); }
jget() { python3 -c 'import sys,json
try: d=json.load(sys.stdin); print(d.get(sys.argv[1],""))
except Exception: print("")' "$1"; }
signup() { # email org ws jar [fullname]
  curl -s -m 20 -A "curl/8" -c "$4" -H 'Content-Type: application/json' -H 'Accept: application/json' \
    -d "{\"email\":\"$1\",\"password\":\"$PASS\",\"account_type\":\"tenant_owner\",\"org_name\":\"$2\",\"workspace_name\":\"$3\",\"full_name\":\"${5:-Same Name}\"}" "$BASE/signup"
}
get() { curl -s -m 20 -A "curl/8" -b "$1" -H 'Accept: application/json' -w '\n%{http_code}' "$BASE$2"; }

echo "Target: $BASE"
curl -s -m 5 "$BASE/health" >/dev/null || { echo "server not reachable"; exit 2; }

echo "[1] Create two separate organizations"
A=$(signup "a-$RUN@smoke.test" "OrgA-$RUN" "wsa$RUN" "$TMP/a.jar" "Same Name")
B=$(signup "b-$RUN@smoke.test" "OrgB-$RUN" "wsb$RUN" "$TMP/b.jar" "Same Name")
TA=$(echo "$A" | jget tenant_id); TB=$(echo "$B" | jget tenant_id)
[ -n "$TA" ] && [ -n "$TB" ] && [ "$TA" != "$TB" ] && ok "two tenants created ($TA / $TB)" || bad "tenant creation failed: $A | $B"
[ -n "$TA" ] && echo "  (same full name 'Same Name' with different emails was accepted: names are not unique, as intended)"

if [ -n "${GARUDA_BIN:-}" ] && [ -n "$TB" ]; then
  d="$TMP/seed"; mkdir -p "$d"; printf 'package main\nfunc Secret%s() {}\nfunc main() {}\n' "$RUN" > "$d/main.go"; printf 'module seed\ngo 1.21\n' > "$d/go.mod"
  GARUDA_TENANT_ID="$TB" "$GARUDA_BIN" analyze "$d" --workspace "wsb$RUN" --save >/dev/null 2>&1 && echo "  (seeded tenant B with symbol Secret$RUN)"
fi

echo "[2] Tenant A must not see tenant B"
R=$(get "$TMP/b.jar" "/api/v1/dashboard/stats?workspace=wsb$RUN"); code=$(echo "$R" | tail -1)
[ "$code" = "200" ] && ok "control: B can read its own workspace (HTTP 200)" || bad "control: B could not read its own workspace (HTTP $code), so the 404s below prove nothing"
R=$(get "$TMP/a.jar" "/api/v1/workspaces"); echo "$R" | grep -q "wsb$RUN" && bad "A's workspace list contains B's workspace" || ok "A's workspace list does not contain B's workspace"
for path in "/api/v1/dashboard/stats?workspace=wsb$RUN" "/api/v1/dashboard/decisions?workspace=wsb$RUN" "/api/v1/dashboard/search?q=Secret$RUN&workspace=wsb$RUN" "/api/v1/graph?workspace=wsb$RUN" "/api/v1/dashboard/stats?workspace=wsb$RUN&tenant_id=$TB" "/api/v1/dashboard/policies?workspace=wsb$RUN&tenant_id=$TB"; do
  R=$(get "$TMP/a.jar" "$path"); code=$(echo "$R" | tail -1)
  if echo "$R" | grep -q "Secret$RUN"; then bad "A read B's data via $path"
  else ok "A got no B data via $path (HTTP $code)"; fi
done
R=$(get "$TMP/a.jar" "/api/v1/dashboard/me"); echo "$R" | grep -q "$TB" && bad "A's profile mentions B's tenant" || ok "A's profile does not mention B's tenant"

echo "[3] Anonymous access"
for path in "/api/v1/workspaces" "/api/v1/dashboard/stats"; do
  code=$(curl -s -m 10 -o /dev/null -w '%{http_code}' -H 'Accept: application/json' "$BASE$path")
  case "$code" in 401|403|302|303) ok "anonymous $path -> $code";; *) bad "anonymous $path -> $code";; esac
done

echo "[4] Duplicate registration must not reveal account details"
D=$(curl -s -m 20 -A "curl/8" -H 'Content-Type: application/json' -H 'Accept: application/json' \
  -d "{\"email\":\"a-$RUN@smoke.test\",\"password\":\"wrong-guess-123\",\"account_type\":\"tenant_owner\"}" "$BASE/signup")
echo "$D" | grep -qi 'user_id\|"role"' && bad "duplicate signup response leaks account details: $D" || ok "duplicate signup does not leak user id or role"
U=$(curl -s -m 20 -A "curl/8" -H 'Content-Type: application/json' -H 'Accept: application/json' \
  -d "{\"email\":\"A-$RUN@SMOKE.TEST\",\"password\":\"$PASS\",\"account_type\":\"tenant_owner\"}" "$BASE/signup")
[ "$(echo "$U" | jget status)" = "created" ] && bad "same email in different letter case created a second account" || ok "email is case-insensitive unique"

echo "[5] Hardcoded owner email must not grant ownership of the original tenant"
H=$(signup "myshra777@gmail.com" "Intruder-$RUN" "intr$RUN" "$TMP/h.jar" "Intruder")
if [ "$(echo "$H" | jget status)" = "created" ]; then
  R=$(get "$TMP/h.jar" "/api/v1/workspaces"); n=$(echo "$R" | grep -o '"name"' | wc -l)
  echo "$R" | grep -qE '"(role)"[^,}]*owner' && echo "  (this account lists $n workspace entries)"
  python3 - "$R" <<'PY' && bad "a fresh sign-up with the hardcoded owner email can see other tenants' workspaces" || ok "hardcoded owner email gets only its own workspace"
import sys,json,re
t=sys.argv[1].rsplit("\n",1)[0]
try: d=json.loads(t)
except Exception: sys.exit(1)
ws=d.get("workspaces",d) if isinstance(d,dict) else d
names=[w.get("name","") for w in ws if isinstance(w,dict)]
sys.exit(0 if len([n for n in names if not n.startswith("intr")])>0 else 1)
PY
elif echo "$H" | grep -q already_exists; then
  echo "  SKIP  that address is already registered on this instance; run this check on a fresh database"
else
  ok "sign-up with the hardcoded owner email was refused (no account created)"
fi

echo; echo "Result: $pass passed, $fail failed"
[ "$fail" -eq 0 ]
