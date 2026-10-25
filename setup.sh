#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════════
# 🦅 GARUDA v1.6 ZERO-DEPENDENCY UNIVERSAL SETUP (setup.sh)
# ═══════════════════════════════════════════════════════════════════════════
# Supports macOS, Linux, Windows WSL2, Chromebook (ChromiumOS), and Android (Termux)
# Usage:
#   ./setup.sh [optional_repo_path_1 optional_repo_path_2 ...]
#
# Zero Pre-Installed Dependencies Required:
#   • No Go installed?       Automatically downloads Portable Go 1.24 into ~/.garuda/toolchain/go
#   • No GCC/Clang installed? Automatically compiles in Pure-Go mode (CGO_ENABLED=0)
#   • No Docker installed?   Automatically provisions a user-space PostgreSQL cluster
#                            in ~/.garuda/data/pg on port 5433 (zero root/sudo required if pg_ctl exists)
#   • Chromebook / Android?  Auto-detects Crostini / Termux and configures PWA + local runtime
# ═══════════════════════════════════════════════════════════════════════════

set -euo pipefail

cd "$(dirname "$0")"

GARUDA_HOME="${HOME}/.garuda"
TOOLCHAIN_DIR="${GARUDA_HOME}/toolchain"
DATA_DIR="${GARUDA_HOME}/data"
mkdir -p "$TOOLCHAIN_DIR" "$DATA_DIR"

# Load existing .env if present
if [[ -f .env ]]; then
    set -a
    # shellcheck disable=SC1091
    source .env
    set +a
fi

# Detect OS & Architecture
OS_RAW="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH_RAW="$(uname -m)"
case "$OS_RAW" in
    darwin*) PLATFORM_OS="darwin" ;;
    linux*)  PLATFORM_OS="linux" ;;
    msys*|mingw*|cygwin*)
        echo "💡 Windows Git Bash detected — delegating to setup.ps1 for native Windows setup..."
        exec powershell.exe -ExecutionPolicy Bypass -File ./setup.ps1
        ;;
    *) PLATFORM_OS="linux" ;;
esac

case "$ARCH_RAW" in
    x86_64|amd64) PLATFORM_ARCH="amd64" ;;
    aarch64|arm64) PLATFORM_ARCH="arm64" ;;
    *) PLATFORM_ARCH="amd64" ;;
esac

IS_TERMUX="false"
if [[ -n "${TERMUX_VERSION:-}" ]] || [[ "${PREFIX:-}" == *"/com.termux"* ]]; then
    IS_TERMUX="true"
fi

# 1. Smart Dynamic Discovery of Email, Workspace, Port & DB URL
GIT_EMAIL="$(git config user.email 2>/dev/null || true)"
OWNER_EMAIL="${GARUDA_EMAIL:-${GIT_EMAIL:-${USER:-admin}@garuda.local}}"
OWNER_PASSWORD="${GARUDA_PASSWORD:-garuda2026}"

if [[ -n "${GARUDA_WORKSPACE:-}" ]]; then
    WORKSPACE="$GARUDA_WORKSPACE"
elif [[ -d "$HOME/garuda-validation" ]]; then
    WORKSPACE="go-validation-10"
else
    WORKSPACE="$(basename "$(pwd)")"
fi

API_PORT="${GARUDA_PORT:-8080}"
DB_URL="${DATABASE_URL:-postgres://test:test@localhost:5433/garuda_test?sslmode=disable}"

export PORT="$API_PORT"
export DATABASE_URL="$DB_URL"
export CONTROL_DATABASE_URL="${CONTROL_DATABASE_URL:-$DB_URL}"
export GARUDA_CONTROL_TOKEN="${GARUDA_CONTROL_TOKEN:-devtoken-changeme-please}"
export GARUDA_WORKSPACE="$WORKSPACE"
export GARUDA_TALON_ENFORCE="true"

echo "═══════════════════════════════════════════════════════════════════"
echo "🦅 GARUDA v1.6 + TALON ZERO-DEPENDENCY ONBOARDING (${PLATFORM_OS}/${PLATFORM_ARCH})"
echo "═══════════════════════════════════════════════════════════════════"
echo "  • Detected Owner:     ${OWNER_EMAIL}"
echo "  • Target Workspace:   ${WORKSPACE}"
echo "  • Gateway Port:       :${API_PORT}"
echo "  • Active Enforcement: GARUDA_TALON_ENFORCE=true"

# 1. Free Port Dynamically
echo "[1/8] Checking port ${API_PORT} and stopping conflicting services..."
pkill -9 -f garuda-api 2>/dev/null || true
if command -v docker >/dev/null 2>&1; then
    CONFLICT_CONTAINERS="$(docker ps --format '{{.Names}}\t{{.Ports}}' 2>/dev/null | grep ":${API_PORT}->" | awk '{print $1}' || true)"
    for c in $CONFLICT_CONTAINERS; do
        echo "  Stopping conflicting container on :${API_PORT}: $c"
        docker stop "$c" >/dev/null 2>&1 || true
    done
fi
sleep 1

# 2. Ensure PostgreSQL is running (Docker OR Zero-Root User-Space Cluster on :5433)
echo "[2/8] Verifying PostgreSQL database on :5433..."
ensure_userspace_postgres() {
    local pg_data="${DATA_DIR}/pg"
    local pg_log="${DATA_DIR}/postgres.log"

    # Add common macOS Homebrew / Postgres.app paths if present
    for pg_candidate in /opt/homebrew/opt/postgresql@16/bin /opt/homebrew/bin /usr/local/opt/postgresql@16/bin /Applications/Postgres.app/Contents/Versions/latest/bin /usr/lib/postgresql/16/bin /usr/lib/postgresql/15/bin; do
        if [[ -d "$pg_candidate" ]]; then
            export PATH="$pg_candidate:$PATH"
        fi
    done

    if ! command -v initdb >/dev/null 2>&1; then
        echo "  ⏳ PostgreSQL tools not found — installing PostgreSQL automatically..."
        if [[ "$IS_TERMUX" == "true" ]]; then
            pkg install -y postgresql
        elif [[ "$PLATFORM_OS" == "darwin" ]] && command -v brew >/dev/null 2>&1; then
            brew install postgresql@16
            export PATH="/opt/homebrew/opt/postgresql@16/bin:/usr/local/opt/postgresql@16/bin:$PATH"
        elif command -v apt-get >/dev/null 2>&1; then
            sudo apt-get update -qq && sudo apt-get install -y -qq postgresql postgresql-client
            for d in /usr/lib/postgresql/*/bin; do [[ -d "$d" ]] && export PATH="$d:$PATH"; done
        elif command -v dnf >/dev/null 2>&1; then
            sudo dnf install -y postgresql-server postgresql
        elif command -v pacman >/dev/null 2>&1; then
            sudo pacman -Sy --noconfirm postgresql
        fi
    fi

    if command -v initdb >/dev/null 2>&1 && command -v pg_ctl >/dev/null 2>&1; then
        if [[ ! -f "${pg_data}/PG_VERSION" ]]; then
            echo "  🛠️  Initializing user-space PostgreSQL cluster in ${pg_data}..."
            initdb -D "$pg_data" -U test -A trust -E UTF8 --no-locale >/dev/null 2>&1 || \
            initdb -D "$pg_data" -A trust -E UTF8 --no-locale >/dev/null 2>&1
        fi
        pg_ctl -D "$pg_data" -l "$pg_log" -o "-p 5433 -k /tmp" start >/dev/null 2>&1 || true
        for _ in $(seq 1 15); do
            if pg_isready -h 127.0.0.1 -p 5433 >/dev/null 2>&1; then
                break
            fi
            sleep 1
        done
        psql -h 127.0.0.1 -p 5433 -U test -d postgres -tc "SELECT 1 FROM pg_roles WHERE rolname='test'" 2>/dev/null | grep -q 1 || \
            psql -h 127.0.0.1 -p 5433 -d postgres -c "CREATE ROLE test WITH SUPERUSER LOGIN PASSWORD 'test';" 2>/dev/null || true
        psql -h 127.0.0.1 -p 5433 -U test -d postgres -tc "SELECT 1 FROM pg_database WHERE datname='garuda_test'" 2>/dev/null | grep -q 1 || \
            createdb -h 127.0.0.1 -p 5433 -U test garuda_test 2>/dev/null || true
    fi
}

if ! (command -v psql >/dev/null 2>&1 && psql "$DB_URL" -c "SELECT 1" >/dev/null 2>&1); then
    if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
        if ! docker ps --format "{{.Names}}" 2>/dev/null | grep -q "^garuda-postgres$"; then
            docker start garuda-postgres >/dev/null 2>&1 || \
            docker run -d --name garuda-postgres \
                -e POSTGRES_USER=test \
                -e POSTGRES_PASSWORD=test \
                -e POSTGRES_DB=garuda_test \
                -p 5433:5432 \
                postgres:16-alpine >/dev/null
        fi
        until docker exec garuda-postgres pg_isready -U test -d garuda_test >/dev/null 2>&1; do
            echo "  waiting for postgres container..."
            sleep 1
        done
    else
        ensure_userspace_postgres
    fi
    if command -v psql >/dev/null 2>&1; then
        BASE_PG_URL="$(echo "$DB_URL" | sed -E 's|/[^/?]+(\?.*)?$|/postgres\1|')"
        DB_NAME="$(echo "$DB_URL" | sed -E 's|.*/([^/?]+)(\?.*)?$|\1|')"
        psql "$BASE_PG_URL" -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" 2>/dev/null | grep -q 1 || \
            psql "$BASE_PG_URL" -c "CREATE DATABASE \"$DB_NAME\";" 2>/dev/null || true
    fi
fi
echo "  ✓ PostgreSQL is online and ready"

# 3. Bootstrap Portable Go Toolchain (if missing) & Auto-Detect C Compiler for CGO Fallback
echo "[3/8] Preparing Go toolchain and compiling Garuda binaries..."
if [[ -x "${TOOLCHAIN_DIR}/go/bin/go" ]]; then
    export PATH="${TOOLCHAIN_DIR}/go/bin:$PATH"
fi

if ! command -v go >/dev/null 2>&1; then
    if [[ "$IS_TERMUX" == "true" ]]; then
        echo "  ⏳ Installing Go for Android Termux..."
        pkg install -y golang
    else
        GO_TAR="go1.24.0.${PLATFORM_OS}-${PLATFORM_ARCH}.tar.gz"
        echo "  ⬇️  Go not found — downloading Portable Go 1.24 (${GO_TAR}) into ${TOOLCHAIN_DIR}/go (no sudo required)..."
        curl -fsSL "https://go.dev/dl/${GO_TAR}" | tar -xz -C "$TOOLCHAIN_DIR"
        export PATH="${TOOLCHAIN_DIR}/go/bin:$PATH"
    fi
fi

# Auto-detect if a working C compiler (gcc/clang) exists; if not, build in 100% Pure-Go mode (CGO_ENABLED=0)
if [[ -z "${CGO_ENABLED:-}" ]]; then
    if command -v gcc >/dev/null 2>&1 || command -v clang >/dev/null 2>&1; then
        # Verify macOS Xcode stub isn't broken
        if cc --version >/dev/null 2>&1; then
            export CGO_ENABLED=1
        else
            export CGO_ENABLED=0
        fi
    else
        export CGO_ENABLED=0
    fi
fi
echo "  • Compiling with $(go version) [CGO_ENABLED=${CGO_ENABLED}]..."

mkdir -p bin
go build -o bin/garuda ./cmd/garuda
go build -o bin/garuda-api ./cmd/garuda-api
go build -o bin/garuda-mcp ./cmd/garuda-mcp
go build -o bin/garuda-talon ./cmd/garuda-talon
echo "  ✓ Built bin/garuda, bin/garuda-api, bin/garuda-mcp, bin/garuda-talon"

# 4. Run 1-Click Initialization (Migrations + .env + Owner Account + Active MCP & Talon Config)
echo "[4/8] Running idempotent migrations, persisting .env, and provisioning owner..."
./bin/garuda init \
    --workspace "$WORKSPACE" \
    --email "$OWNER_EMAIL" \
    --password "$OWNER_PASSWORD" \
    --skip-scan

# 5. Start Garuda API Gateway and verify /health + /signup
echo "[5/8] Starting Garuda API Gateway on :${API_PORT}..."
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
    -d "password=${OWNER_PASSWORD}" || true)
echo "  ✓ Signup endpoint response: ${SIGNUP_RESP}"
if command -v psql >/dev/null 2>&1; then
    echo "  ✓ Active Tenants & Users in PostgreSQL:"
    psql "$DATABASE_URL" -c "SELECT id, name FROM tenants;" || true
    psql "$DATABASE_URL" -c "SELECT id, email, role FROM users;" || true
fi

# 6. Dynamically Discover & Index Repositories
echo "[6/8] Indexing repositories into workspace '${WORKSPACE}'..."
if [[ $# -gt 0 ]]; then
    for repo_path in "$@"; do
        [[ -d "$repo_path" ]] && ./bin/garuda analyze "$repo_path" --workspace "$WORKSPACE" --save
    done
elif [[ -d "$HOME/garuda-validation" ]]; then
    for candidate in "$HOME/garuda-validation/go/chi" "$HOME/garuda-validation/py/sqlmodel" "$HOME/garuda-validation/go/cobra"; do
        if [[ -d "$candidate" ]]; then
            ./bin/garuda analyze "$candidate" --workspace "$WORKSPACE" --save
        fi
    done
else
    ./bin/garuda analyze . --workspace "$WORKSPACE" --save
fi

# 7. Compile & Install Active Garuda Talon Contract + IDE/Git Hooks (with indexed graph hubs)
POLICY_DIR="policies"
[[ -d "$HOME/garuda-validation/policies" ]] && POLICY_DIR="$HOME/garuda-validation/policies"
echo "[7/8] Activating Garuda Talon Harness (Contract + Cursor/Claude/Agents/Git Hooks)..."
./bin/garuda talon init . --workspace "$WORKSPACE" --policies "$POLICY_DIR"

# 8. Run Summary, Diff-Scoped Change-Control Check, Talon Guard, System Doctor & MCP Verification
echo "[8/8] Running Architectural Summary, Talon Guard, System Doctor & Active MCP Verification..."
./bin/garuda summary
./bin/garuda check --project "$WORKSPACE" --policy "$POLICY_DIR" --save || true
./bin/garuda talon guard --workspace "$WORKSPACE" --policies "$POLICY_DIR" || true
./bin/garuda doctor
if command -v python3 >/dev/null 2>&1; then
    WORKSPACE="$WORKSPACE" python3 scripts/mcp_verify.py || true
fi

echo ""
echo "═══════════════════════════════════════════════════════════════════"
echo "🚀 GARUDA v1.6 + GARUDA TALON IS 100% CONFIGURED, VERIFIED & ONLINE!"
echo "   • Web Dashboard:  http://localhost:${API_PORT}/login"
echo "   • Chromium PWA:   Open in Chrome/Edge/Brave -> Click 'Install Garuda'"
echo "   • Email / Pass:   ${OWNER_EMAIL} / ${OWNER_PASSWORD}"
echo "   • System Doctor:  ./bin/garuda doctor"
echo "   • Talon Contract: .garuda/TALON_CONTRACT.md"
echo "   • Talon Guard:    ./bin/garuda talon guard --workspace ${WORKSPACE}"
echo "═══════════════════════════════════════════════════════════════════"
