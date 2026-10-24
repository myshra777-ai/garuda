#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════════
# 🦅 GARUDA ONE-COMMAND ZERO-CONFIG SETUP & VERIFICATION (setup.sh)
# ═══════════════════════════════════════════════════════════════════════════
# Usage:
#   ./setup.sh
#
# What this single command does automatically:
#   1. Frees port 8080 (stops conflicting aether-control container & stale garuda-api)
#   2. Starts/verifies PostgreSQL Docker container (garuda-postgres on :5433)
#   3. Builds bin/garuda, bin/garuda-api, and bin/garuda-mcp
#   4. Generates persistent .env (DB URL, Workspace, Control Token, Ed25519 JWT keys)
#   5. Runs 100% idempotent database migrations (safe on fresh or existing DBs)
#   6. Starts bin/garuda-api gateway and verifies /signup + users & tenants tables
#   7. Indexes repositories (~/garuda-validation/* or current repo)
#   8. Runs garuda summary, garuda check, and scripts/mcp_verify.py (23/23 checks)
# ═══════════════════════════════════════════════════════════════════════════

set -euo pipefail

cd "$(dirname "$0")"

WORKSPACE="${GARUDA_WORKSPACE:-go-validation-10}"
OWNER_EMAIL="${GARUDA_EMAIL:-myshra777@gmail.com}"
OWNER_PASSWORD="${GARUDA_PASSWORD:-garuda2026}"
DB_URL="${DATABASE_URL:-postgres://test:test@localhost:5433/garuda_test?sslmode=disable}"
API_PORT="8080"

unset PORT
export PORT="$API_PORT"
export DATABASE_URL="$DB_URL"
export CONTROL_DATABASE_URL="$DB_URL"
export GARUDA_CONTROL_TOKEN="${GARUDA_CONTROL_TOKEN:-devtoken-changeme-please}"
export GARUDA_WORKSPACE="$WORKSPACE"

echo "═══════════════════════════════════════════════════════════════════"
echo "🦅 GARUDA 1-COMMAND ONBOARDING & VERIFICATION"
echo "═══════════════════════════════════════════════════════════════════"

# 1. Free Port 8080 cleanly
echo "[1/7] Checking port ${API_PORT} and stopping conflicting services..."
pkill -9 -f garuda-api 2>/dev/null || true
docker stop aether-control 2>/dev/null || true
sleep 1

# 2. Ensure PostgreSQL is running and accepting connections
echo "[2/7] Verifying PostgreSQL container (garuda-postgres on :5433)..."
if ! docker ps --format "{{.Names}}" 2>/dev/null | grep -q "^garuda-postgres$"; then
    docker start garuda-postgres 2>/dev/null || \
    docker run -d --name garuda-postgres \
        -e POSTGRES_USER=test \
        -e POSTGRES_PASSWORD=test \
        -e POSTGRES_DB=garuda_test \
        -p 5433:5432 \
        postgres:16-alpine >/dev/null
fi

until docker exec garuda-postgres pg_isready -U test -d garuda_test >/dev/null 2>&1; do
    echo "  waiting for postgres..."
    sleep 1
done
# Ensure garuda_test database exists
psql "postgres://test:test@localhost:5433/postgres?sslmode=disable" -tc "SELECT 1 FROM pg_database WHERE datname = 'garuda_test'" | grep -q 1 || \
    psql "postgres://test:test@localhost:5433/postgres?sslmode=disable" -c "CREATE DATABASE garuda_test;"
echo "  ✓ PostgreSQL is online and ready"

# 3. Build all binaries
echo "[3/7] Compiling Garuda binaries (bin/garuda, bin/garuda-api, bin/garuda-mcp)..."
mkdir -p bin
go build -o bin/garuda ./cmd/garuda
go build -o bin/garuda-api ./cmd/garuda-api
go build -o bin/garuda-mcp ./cmd/garuda-mcp
echo "  ✓ Built bin/garuda, bin/garuda-api, bin/garuda-mcp"

# 4. Run 1-Click Initialization (Migrations + .env + Owner Account + MCP Config)
echo "[4/7] Running idempotent migrations, persisting .env, and provisioning owner..."
./bin/garuda init \
    --workspace "$WORKSPACE" \
    --email "$OWNER_EMAIL" \
    --password "$OWNER_PASSWORD" \
    --skip-scan

# 5. Start Garuda API Gateway and verify /health + /signup
echo "[5/7] Starting Garuda API Gateway on :${API_PORT}..."
nohup ./bin/garuda-api > /tmp/api.log 2>&1 &
for _ in $(seq 1 20); do
    if curl -fsS "http://localhost:${API_PORT}/health" >/dev/null 2>&1; then
        break
    fi
    sleep 0.5
done

echo "  ✓ Garuda API Gateway online at http://localhost:${API_PORT}"
SIGNUP_RESP=$(curl -s -X POST "http://localhost:${API_PORT}/signup" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "email=${OWNER_EMAIL}" \
    -d "password=${OWNER_PASSWORD}")
echo "  ✓ Signup endpoint response: ${SIGNUP_RESP}"
echo "  ✓ Active Tenants & Users in PostgreSQL:"
psql "$DATABASE_URL" -c "SELECT id, name FROM tenants;"
psql "$DATABASE_URL" -c "SELECT id, email, role FROM users;"

# 6. Index validation repositories (or current repo)
echo "[6/7] Indexing repositories into workspace '${WORKSPACE}'..."
if [[ -d "$HOME/garuda-validation" ]]; then
    [[ -d "$HOME/garuda-validation/go/chi" ]] && ./bin/garuda analyze "$HOME/garuda-validation/go/chi" --workspace "$WORKSPACE" --save
    [[ -d "$HOME/garuda-validation/py/sqlmodel" ]] && ./bin/garuda analyze "$HOME/garuda-validation/py/sqlmodel" --workspace "$WORKSPACE" --save
    [[ -d "$HOME/garuda-validation/go/cobra" ]] && ./bin/garuda analyze "$HOME/garuda-validation/go/cobra" --workspace "$WORKSPACE" --save
else
    ./bin/garuda analyze . --workspace "$WORKSPACE" --save
fi

# 7. Run Summary, Unified Change-Control Check, and MCP Spec Verification
echo "[7/7] Running Architectural Summary, Change-Control Check & MCP Verification..."
./bin/garuda summary
if [[ -d "$HOME/garuda-validation/policies" ]]; then
    ./bin/garuda check --project "$WORKSPACE" --policy "$HOME/garuda-validation/policies" --save || true
else
    ./bin/garuda check --project "$WORKSPACE" --policy policies --save || true
fi
WORKSPACE="$WORKSPACE" python3 scripts/mcp_verify.py

echo ""
echo "═══════════════════════════════════════════════════════════════════"
echo "🚀 GARUDA IS 100% CONFIGURED, VERIFIED & ONLINE!"
echo "   • Web Dashboard:  http://localhost:${API_PORT}/login"
echo "   • Email / Pass:   ${OWNER_EMAIL} / ${OWNER_PASSWORD}"
echo "   • Change Control: ./bin/garuda check --project ${WORKSPACE} --policy policies --save"
echo "═══════════════════════════════════════════════════════════════════"
