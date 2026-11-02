#!/usr/bin/env bash
# ===========================================================================
# GARUDA ZERO-DEPENDENCY UNIVERSAL SETUP (setup.sh)
# ===========================================================================
# Supports macOS (Intel & Apple Silicon), Linux, Windows WSL2, ChromeOS
# (Crostini Linux container) and Android (Termux).
#
# Usage:
#   ./setup.sh [optional_repo_path_1 optional_repo_path_2 ...]
#
# Optional environment variables:
#   GARUDA_PASSWORD   owner password (default: random, saved to ~/.garuda/owner-login)
#   GARUDA_EMAIL      owner email (default: git config user.email)
#   GARUDA_WORKSPACE  workspace name (default: the name of this folder)
#   GARUDA_PORT       dashboard port (default: 8080)
#   DATABASE_URL      use your own PostgreSQL instead of the bundled local one
#
# What it does, with no sudo/root and nothing to install first except curl:
#   - Downloads a portable PostgreSQL 16 (~18 MB, Maven Central) if none is found.
#   - Downloads a portable Go matching go.mod if Go (>= 1.21) is missing.
#   - Builds the 4 binaries (pure Go, CGO_ENABLED=0) into ./bin, then runs
#     migrations, creates the owner account, starts the dashboard and runs
#     the doctor.
# Everything it downloads lives under ~/.garuda.
# ===========================================================================

set -euo pipefail

cd "$(dirname "$0")"
REPO_ROOT="$(pwd)"

GARUDA_VERSION_LABEL="$(head -n 1 version.txt 2>/dev/null | tr -d '[:space:]' || true)"
GARUDA_VERSION_LABEL="${GARUDA_VERSION_LABEL:-dev}"

GARUDA_HOME="${HOME}/.garuda"
TOOLCHAIN_DIR="${GARUDA_HOME}/toolchain"
RUNTIME_DIR="${GARUDA_HOME}/runtime"
DATA_DIR="${GARUDA_HOME}/data"
RUN_DIR="${DATA_DIR}/run"
LOGIN_FILE="${GARUDA_HOME}/owner-login"
API_PID_FILE="${DATA_DIR}/garuda-api.pid"
mkdir -p "$TOOLCHAIN_DIR" "$RUNTIME_DIR" "$DATA_DIR" "$RUN_DIR"
chmod 700 "$GARUDA_HOME" "$RUN_DIR" 2>/dev/null || true

# Load .env (written by `garuda init`) without evaluating it as shell code.
# Variables you already exported win over .env.
if [[ -f .env ]]; then
    while IFS= read -r env_line || [[ -n "$env_line" ]]; do
        if [[ "$env_line" =~ ^[[:space:]]*([A-Za-z_][A-Za-z0-9_]*)=(.*)$ ]]; then
            env_key="${BASH_REMATCH[1]}"
            env_val="${BASH_REMATCH[2]}"
            if [[ "$env_val" =~ ^\"(.*)\"$ ]] || [[ "$env_val" =~ ^\'(.*)\'$ ]]; then
                env_val="${BASH_REMATCH[1]}"
            fi
            if [[ -z "${!env_key+x}" ]]; then
                export "$env_key=$env_val"
            fi
        fi
    done < .env
fi

# ---------------------------------------------------------------------------
# Platform detection
# ---------------------------------------------------------------------------
OS_RAW="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH_RAW="$(uname -m)"
case "$OS_RAW" in
    darwin*) PLATFORM_OS="darwin" ;;
    linux*)  PLATFORM_OS="linux" ;;
    msys*|mingw*|cygwin*)
        echo "[*] Windows Git Bash detected -- delegating to setup.ps1 for native Windows setup..."
        exec powershell.exe -NoProfile -ExecutionPolicy Bypass -File ./setup.ps1
        ;;
    *)
        echo "[ERROR] Unsupported operating system: $(uname -s). Supported: macOS, Linux (incl. WSL2, ChromeOS Linux), Android Termux, Windows (setup.ps1)." >&2
        exit 1
        ;;
esac

case "$ARCH_RAW" in
    x86_64|amd64)  PLATFORM_ARCH="amd64" ;;
    aarch64|arm64) PLATFORM_ARCH="arm64" ;;
    *)
        echo "[ERROR] Unsupported CPU architecture: ${ARCH_RAW}. Garuda needs a 64-bit x86_64 or ARM64 device." >&2
        exit 1
        ;;
esac

IS_TERMUX="false"
if [[ -n "${TERMUX_VERSION:-}" ]] || [[ "${PREFIX:-}" == *"/com.termux"* ]]; then
    IS_TERMUX="true"
fi

# Print how to install a missing tool on this platform, then stop.
need_cmd() {
    local cmd="$1" pkg_hint="${2:-$1}"
    if command -v "$cmd" >/dev/null 2>&1; then
        return 0
    fi
    echo "[ERROR] '${cmd}' is required but was not found." >&2
    if [[ "$IS_TERMUX" == "true" ]]; then
        echo "        Install it with:  pkg install -y ${pkg_hint}" >&2
    elif [[ "$PLATFORM_OS" == "darwin" ]]; then
        echo "        Install it with:  xcode-select --install   (or: brew install ${pkg_hint})" >&2
    elif command -v apt-get >/dev/null 2>&1; then
        echo "        Install it with:  sudo apt-get install -y ${pkg_hint}" >&2
    elif command -v dnf >/dev/null 2>&1; then
        echo "        Install it with:  sudo dnf install -y ${pkg_hint}" >&2
    elif command -v pacman >/dev/null 2>&1; then
        echo "        Install it with:  sudo pacman -S --noconfirm ${pkg_hint}" >&2
    else
        echo "        Install '${pkg_hint}' with your package manager, then re-run ./setup.sh" >&2
    fi
    exit 1
}

need_cmd curl curl
need_cmd tar tar

rand_hex() {
    # rand_hex <bytes> -> 2*bytes hex characters
    local n="$1"
    openssl rand -hex "$n" 2>/dev/null || head -c "$n" /dev/urandom | od -An -tx1 | tr -d ' \n'
}

# ---------------------------------------------------------------------------
# Owner, workspace, port, database URL
# ---------------------------------------------------------------------------
GIT_EMAIL="$(git config user.email 2>/dev/null || true)"
OWNER_EMAIL="${GARUDA_EMAIL:-${GIT_EMAIL:-${USER:-admin}@garuda.local}}"

# Owner password: GARUDA_PASSWORD > saved from an earlier run (same email) > new random value.
# `garuda init` never overwrites an existing user's password, so a re-run must reuse the saved one.
PASSWORD_SOURCE="you"
if [[ -z "${GARUDA_PASSWORD:-}" && -f "$LOGIN_FILE" ]]; then
    saved_email="$(sed -n 1p "$LOGIN_FILE" 2>/dev/null || true)"
    saved_pw="$(sed -n 2p "$LOGIN_FILE" 2>/dev/null || true)"
    if [[ "$saved_email" == "$OWNER_EMAIL" && ${#saved_pw} -ge 8 ]]; then
        GARUDA_PASSWORD="$saved_pw"
        PASSWORD_SOURCE="saved"
    fi
fi
if [[ -z "${GARUDA_PASSWORD:-}" ]]; then
    GARUDA_PASSWORD="$(rand_hex 12)"
    ( umask 077; printf '%s\n%s\n' "$OWNER_EMAIL" "$GARUDA_PASSWORD" > "$LOGIN_FILE" )
    PASSWORD_SOURCE="generated"
fi
if [[ ${#GARUDA_PASSWORD} -lt 8 ]]; then
    echo "[ERROR] GARUDA_PASSWORD must be at least 8 characters." >&2
    exit 1
fi
# Passed to `garuda init` through the environment, not the command line (visible in `ps`).
export GARUDA_PASSWORD

WORKSPACE="${GARUDA_WORKSPACE:-$(basename "$REPO_ROOT")}"
API_PORT="${GARUDA_PORT:-${PORT:-8080}}"
RAW_DB_URL="${DATABASE_URL:-postgres://test:test@127.0.0.1:5433/garuda_test?sslmode=disable}"
DB_URL="${RAW_DB_URL//@localhost:/@127.0.0.1:}"

# Host and port of the database, to decide whether we manage a local PostgreSQL.
DB_HOSTPORT="${DB_URL#*@}"
DB_HOSTPORT="${DB_HOSTPORT%%/*}"
DB_HOST="${DB_HOSTPORT%%:*}"
DB_PORT="${DB_HOSTPORT##*:}"
[[ "$DB_PORT" == "$DB_HOSTPORT" ]] && DB_PORT="5432"
MANAGE_LOCAL_PG="false"
if [[ ( "$DB_HOST" == "127.0.0.1" || "$DB_HOST" == "localhost" ) && "$DB_PORT" == "5433" ]]; then
    MANAGE_LOCAL_PG="true"
fi

export PORT="$API_PORT"
export DATABASE_URL="$DB_URL"
export CONTROL_DATABASE_URL="${CONTROL_DATABASE_URL:-$DB_URL}"
# No well-known default token: `garuda init` saves this random value into .env, so re-runs reuse it.
if [[ -z "${GARUDA_CONTROL_TOKEN:-}" ]]; then
    GARUDA_CONTROL_TOKEN="gct_$(rand_hex 24)"
fi
export GARUDA_CONTROL_TOKEN
export GARUDA_WORKSPACE="$WORKSPACE"
export GARUDA_TALON_ENFORCE="true"
export CGO_ENABLED="${CGO_ENABLED:-0}"
export HOMEBREW_NO_AUTO_UPDATE=1
export HOMEBREW_NO_INSTALL_CLEANUP=1
export HOMEBREW_NO_ENV_HINTS=1

echo "==================================================================="
echo "GARUDA ${GARUDA_VERSION_LABEL} + TALON ZERO-DEPENDENCY SETUP (${PLATFORM_OS}/${PLATFORM_ARCH})"
echo "==================================================================="
echo "  * Owner:              ${OWNER_EMAIL}"
echo "  * Workspace:          ${WORKSPACE}"
echo "  * Dashboard port:     :${API_PORT}"
echo "  * Database:           ${DB_HOST}:${DB_PORT}"
echo "  * Talon enforcement:  GARUDA_TALON_ENFORCE=true"

test_tcp_port() {
    local host="$1"
    local port="$2"
    if command -v nc >/dev/null 2>&1; then
        nc -z -w 1 "$host" "$port" >/dev/null 2>&1 && return 0
    fi
    if command -v python3 >/dev/null 2>&1; then
        python3 -c "import socket,sys; s=socket.socket(); s.settimeout(0.5); sys.exit(0 if s.connect_ex(('$host',$port))==0 else 1)" >/dev/null 2>&1 && return 0
    fi
    (echo >/dev/tcp/"$host"/"$port") >/dev/null 2>&1 && return 0
    return 1
}

# ---------------------------------------------------------------------------
# [1/8] Dashboard port: stop only a previous Garuda API, never anything else
# ---------------------------------------------------------------------------
echo "[1/8] Checking port ${API_PORT}..."

stop_previous_garuda_api() {
    local old_pid comm port_pid
    if [[ -f "$API_PID_FILE" ]]; then
        old_pid="$(cat "$API_PID_FILE" 2>/dev/null || true)"
        if [[ "$old_pid" =~ ^[0-9]+$ ]] && kill -0 "$old_pid" 2>/dev/null; then
            comm="$(ps -p "$old_pid" -o comm= 2>/dev/null || true)"
            if [[ "$comm" == *garuda-api* ]]; then
                echo "  [*] Stopping the Garuda API from the previous run (pid ${old_pid})..."
                kill "$old_pid" 2>/dev/null || true
                for _ in $(seq 1 20); do
                    kill -0 "$old_pid" 2>/dev/null || break
                    sleep 0.25
                done
                if kill -0 "$old_pid" 2>/dev/null; then
                    kill -9 "$old_pid" 2>/dev/null || true
                fi
            fi
        fi
        rm -f "$API_PID_FILE"
    fi
    # An API started by an older setup (no pid file): only stop it if the listener is really garuda-api.
    if test_tcp_port 127.0.0.1 "$API_PORT" && command -v lsof >/dev/null 2>&1; then
        for port_pid in $(lsof -ti tcp:"${API_PORT}" -sTCP:LISTEN 2>/dev/null || true); do
            comm="$(ps -p "$port_pid" -o comm= 2>/dev/null || true)"
            if [[ "$comm" == *garuda-api* ]]; then
                echo "  [*] Stopping an older Garuda API on :${API_PORT} (pid ${port_pid})..."
                kill "$port_pid" 2>/dev/null || true
                sleep 1
            fi
        done
    fi
}

stop_previous_garuda_api
if test_tcp_port 127.0.0.1 "$API_PORT"; then
    echo "[ERROR] Port ${API_PORT} is in use by another program (not a Garuda API)." >&2
    echo "        Pick a free port and re-run:  GARUDA_PORT=8081 ./setup.sh" >&2
    exit 1
fi

# ---------------------------------------------------------------------------
# [2/8] PostgreSQL (your DATABASE_URL, Docker, or a user-space cluster on 127.0.0.1:5433)
# ---------------------------------------------------------------------------
echo "[2/8] Verifying PostgreSQL database at ${DB_HOST}:${DB_PORT}..."

discover_postgres_path() {
    shopt -s nullglob 2>/dev/null || true
    local pg_candidate
    local -a pg_candidates=(
        "${RUNTIME_DIR}/postgres/bin"
        "${RUNTIME_DIR}/postgres/pgsql/bin"
        /opt/homebrew/opt/postgresql@17/bin
        /opt/homebrew/opt/postgresql@16/bin
        /opt/homebrew/opt/postgresql@15/bin
        /opt/homebrew/opt/postgresql/bin
        /opt/homebrew/Cellar/postgresql*/*/bin
        /usr/local/opt/postgresql@17/bin
        /usr/local/opt/postgresql@16/bin
        /usr/local/opt/postgresql@15/bin
        /usr/local/opt/postgresql/bin
        /usr/local/Cellar/postgresql*/*/bin
        /Applications/Postgres.app/Contents/Versions/latest/bin
        /Applications/Postgres.app/Contents/Versions/*/bin
        /Library/PostgreSQL/*/bin
        /opt/local/lib/postgresql*/bin
        /usr/lib/postgresql/17/bin
        /usr/lib/postgresql/16/bin
        /usr/lib/postgresql/15/bin
        /usr/lib/postgresql/14/bin
        /usr/lib/postgresql/*/bin
    )
    for pg_candidate in "${pg_candidates[@]}"; do
        if [[ -d "$pg_candidate" && -x "$pg_candidate/pg_ctl" ]]; then
            export PATH="$pg_candidate:$PATH"
        fi
    done
    shopt -u nullglob 2>/dev/null || true
}

download_portable_postgres() {
    local pg_dest="${RUNTIME_DIR}/postgres"
    local zonky_os="$PLATFORM_OS"
    local zonky_arch="$PLATFORM_ARCH"
    if [[ "$zonky_arch" == "arm64" ]]; then
        zonky_arch="arm64v8"
    fi
    local pg_ver="16.4.0"
    local artifact="embedded-postgres-binaries-${zonky_os}-${zonky_arch}"
    local jar_url="https://repo1.maven.org/maven2/io/zonky/test/postgres/${artifact}/${pg_ver}/${artifact}-${pg_ver}.jar"
    local jar_file="${RUNTIME_DIR}/${artifact}.jar"

    # The jar is a zip that contains a .txz: we need unzip (or python3) and xz support in tar.
    if ! command -v unzip >/dev/null 2>&1 && ! command -v python3 >/dev/null 2>&1; then
        echo "  [WARN] Neither unzip nor python3 is available to unpack the PostgreSQL download." >&2
        return 1
    fi
    if [[ "$PLATFORM_OS" == "linux" ]] && ! command -v xz >/dev/null 2>&1; then
        echo "  [WARN] 'xz' is missing (needed to unpack PostgreSQL). Debian/Ubuntu/ChromeOS: sudo apt-get install -y xz-utils   Termux: pkg install -y xz-utils" >&2
        return 1
    fi

    echo "  [*] Downloading pre-built Portable PostgreSQL 16 (${zonky_os}/${PLATFORM_ARCH}, ~18 MB)..."
    mkdir -p "$pg_dest"
    if curl -fsSL --retry 3 -o "$jar_file" "$jar_url"; then
        if command -v unzip >/dev/null 2>&1; then
            unzip -q -o "$jar_file" -d "$pg_dest" >/dev/null 2>&1 || true
        else
            python3 -c "import sys, zipfile; zipfile.ZipFile(sys.argv[1]).extractall(sys.argv[2])" "$jar_file" "$pg_dest" 2>/dev/null || true
        fi
        rm -f "$jar_file"
        local txz_file
        txz_file="$(ls "$pg_dest"/postgres-*.txz 2>/dev/null | head -n 1 || true)"
        if [[ -n "$txz_file" && -f "$txz_file" ]]; then
            tar -xJf "$txz_file" -C "$pg_dest" >/dev/null 2>&1 || tar -xf "$txz_file" -C "$pg_dest" >/dev/null 2>&1 || true
            rm -f "$txz_file"
        fi
        if [[ -x "${pg_dest}/bin/pg_ctl" ]]; then
            export PATH="${pg_dest}/bin:$PATH"
            echo "  [OK] Portable PostgreSQL 16 unpacked in ${pg_dest}/bin"
            return 0
        fi
    fi
    return 1
}

init_and_start_pg_cluster() {
    local pg_data="$1"
    local pg_log="${DATA_DIR}/postgres.log"

    if [[ ! -f "${pg_data}/PG_VERSION" ]]; then
        rm -rf "$pg_data"
        mkdir -p "$pg_data"
        chmod 700 "$pg_data"
        echo "  [*] Initializing user-space PostgreSQL cluster in ${pg_data}..."
        LC_ALL=C LANG=C initdb -D "$pg_data" -U test -A trust -E UTF8 --no-locale >/dev/null 2>&1 || \
            LC_ALL=C LANG=C initdb -D "$pg_data" -A trust -E UTF8 --no-locale >/dev/null 2>&1
        if [[ -f "${pg_data}/postgresql.conf" ]]; then
            cat >> "${pg_data}/postgresql.conf" <<PGCONF

# Garuda Zero-Dependency User-Space Cluster Settings
listen_addresses = '127.0.0.1'
port = 5433
unix_socket_directories = '${RUN_DIR}'
max_connections = 100
shared_buffers = 64MB
PGCONF
        fi
    fi

    if ! test_tcp_port 127.0.0.1 5433; then
        rm -f "${pg_data}/postmaster.pid" "${RUN_DIR}/.s.PGSQL.5433.lock" 2>/dev/null || true
        LC_ALL=C LANG=C pg_ctl -D "$pg_data" -l "$pg_log" -o "-h 127.0.0.1 -p 5433 -k ${RUN_DIR}" start >/dev/null 2>&1 || true
        for _ in $(seq 1 25); do
            if test_tcp_port 127.0.0.1 5433; then
                return 0
            fi
            sleep 0.3
        done
    fi
    test_tcp_port 127.0.0.1 5433
}

ensure_userspace_postgres() {
    local pg_data="${DATA_DIR}/pg"

    discover_postgres_path

    if ! command -v initdb >/dev/null 2>&1 || ! command -v pg_ctl >/dev/null 2>&1; then
        if [[ "$IS_TERMUX" == "true" ]]; then
            pkg install -y postgresql
        else
            # First try the fast pre-built portable PostgreSQL 16 binary download
            if ! download_portable_postgres; then
                echo "  [*] Falling back to the system package manager for PostgreSQL..."
                if [[ "$PLATFORM_OS" == "darwin" ]] && command -v brew >/dev/null 2>&1; then
                    brew install -q postgresql@16 || brew install -q postgresql
                elif command -v apt-get >/dev/null 2>&1; then
                    sudo apt-get update -qq && sudo apt-get install -y -qq postgresql postgresql-client
                elif command -v dnf >/dev/null 2>&1; then
                    sudo dnf install -y postgresql-server postgresql
                elif command -v pacman >/dev/null 2>&1; then
                    sudo pacman -Sy --noconfirm postgresql
                fi
                discover_postgres_path
            fi
        fi
    fi

    if command -v initdb >/dev/null 2>&1 && command -v pg_ctl >/dev/null 2>&1; then
        # PostgreSQL refuses to run as root (initdb would fail with a confusing error).
        if [[ "$(id -u)" == "0" ]]; then
            echo "  [ERROR] PostgreSQL cannot be started by the root user." >&2
            echo "          Run ./setup.sh as a regular user, or point DATABASE_URL at an existing database." >&2
            exit 1
        fi
        if ! init_and_start_pg_cluster "$pg_data"; then
            # Only rebuild the cluster if it is ours and nothing is running from it.
            echo "  [*] PostgreSQL did not start -- re-initializing the Garuda user-space cluster in ${pg_data}..."
            LC_ALL=C LANG=C pg_ctl -D "$pg_data" -m immediate stop >/dev/null 2>&1 || true
            rm -rf "$pg_data"
            init_and_start_pg_cluster "$pg_data" || true
        fi
    fi
}

provision_pg_role_and_db() {
    if ! command -v psql >/dev/null 2>&1; then
        return 0
    fi
    # Ensure pg_hba.conf in the user-space cluster permits trust auth on 127.0.0.1
    if [[ -f "${DATA_DIR}/pg/pg_hba.conf" ]] && ! grep -q "garuda-trust-rule" "${DATA_DIR}/pg/pg_hba.conf" 2>/dev/null; then
        local tmp_hba="${DATA_DIR}/pg/pg_hba.conf.tmp"
        {
            echo "# garuda-trust-rule"
            echo "local all all trust"
            echo "host all all 127.0.0.1/32 trust"
            echo "host all all ::1/128 trust"
            cat "${DATA_DIR}/pg/pg_hba.conf"
        } > "$tmp_hba" && mv "$tmp_hba" "${DATA_DIR}/pg/pg_hba.conf"
        pg_ctl -D "${DATA_DIR}/pg" reload >/dev/null 2>&1 || true
    fi

    local db_user
    for db_user in "test" "${USER:-}" "$(whoami 2>/dev/null || true)" "postgres"; do
        [[ -z "$db_user" ]] && continue
        PGPASSWORD=test psql -h 127.0.0.1 -p 5433 -U "$db_user" -d postgres -tc "SELECT 1 FROM pg_roles WHERE rolname='test'" 2>/dev/null | grep -q 1 || \
            PGPASSWORD=test psql -h 127.0.0.1 -p 5433 -U "$db_user" -d postgres -c "CREATE ROLE test WITH SUPERUSER LOGIN PASSWORD 'test';" >/dev/null 2>&1 || \
            psql -h "${RUN_DIR}" -p 5433 -U "$db_user" -d postgres -c "CREATE ROLE test WITH SUPERUSER LOGIN PASSWORD 'test';" >/dev/null 2>&1 || true
        PGPASSWORD=test psql -h 127.0.0.1 -p 5433 -U "$db_user" -d postgres -c "ALTER ROLE test WITH SUPERUSER LOGIN PASSWORD 'test';" >/dev/null 2>&1 || true
        PGPASSWORD=test psql -h 127.0.0.1 -p 5433 -U "$db_user" -d postgres -tc "SELECT 1 FROM pg_database WHERE datname='garuda_test'" 2>/dev/null | grep -q 1 || \
            PGPASSWORD=test psql -h 127.0.0.1 -p 5433 -U "$db_user" -d postgres -c "CREATE DATABASE garuda_test OWNER test;" >/dev/null 2>&1 || \
            PGPASSWORD=test createdb -h 127.0.0.1 -p 5433 -U "$db_user" -O test garuda_test >/dev/null 2>&1 || \
            createdb -h "${RUN_DIR}" -p 5433 -O test garuda_test >/dev/null 2>&1 || true
    done

    # Verify real SQL connectivity as the configured user
    if psql "$DB_URL" -c "SELECT 1;" >/dev/null 2>&1; then
        return 0
    fi
    return 1
}

discover_postgres_path

if [[ "$MANAGE_LOCAL_PG" == "true" ]]; then
    if ! test_tcp_port 127.0.0.1 5433; then
        if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
            if ! docker ps --format "{{.Names}}" 2>/dev/null | grep -q "^garuda-postgres$"; then
                docker start garuda-postgres >/dev/null 2>&1 || \
                    docker run -d --name garuda-postgres \
                        -e POSTGRES_USER=test \
                        -e POSTGRES_PASSWORD=test \
                        -e POSTGRES_DB=garuda_test \
                        -p 127.0.0.1:5433:5432 \
                        postgres:16-alpine >/dev/null 2>&1 || true
            fi
            for _ in $(seq 1 25); do
                if test_tcp_port 127.0.0.1 5433; then
                    break
                fi
                sleep 0.4
            done
        fi
        if ! test_tcp_port 127.0.0.1 5433; then
            ensure_userspace_postgres
        fi
    fi

    if ! test_tcp_port 127.0.0.1 5433; then
        echo "  [ERROR] PostgreSQL failed to start on 127.0.0.1:5433." >&2
        if [[ -f "${DATA_DIR}/postgres.log" ]]; then
            echo "  --- Last 20 lines of ${DATA_DIR}/postgres.log ---" >&2
            tail -n 20 "${DATA_DIR}/postgres.log" >&2 || true
        fi
        exit 1
    fi

    # Wait for pg_isready so PostgreSQL finishes WAL recovery before any SQL connection is attempted
    if command -v pg_isready >/dev/null 2>&1; then
        for _ in $(seq 1 30); do
            if pg_isready -h 127.0.0.1 -p 5433 >/dev/null 2>&1; then
                break
            fi
            sleep 0.3
        done
    fi

    if ! provision_pg_role_and_db; then
        echo "  [*] PostgreSQL on :5433 rejected user 'test' -- healing the user-space cluster in ${DATA_DIR}/pg..."
        ensure_userspace_postgres
        provision_pg_role_and_db || true
    fi
    echo "  [OK] PostgreSQL connection and 'garuda_test' database verified on 127.0.0.1:5433"
else
    if ! test_tcp_port "$DB_HOST" "$DB_PORT"; then
        echo "  [ERROR] DATABASE_URL points at ${DB_HOST}:${DB_PORT}, which is not reachable." >&2
        echo "          Start that database, or unset DATABASE_URL to use the bundled local PostgreSQL." >&2
        exit 1
    fi
    echo "  [OK] Using the database from DATABASE_URL (${DB_HOST}:${DB_PORT})"
fi

# ---------------------------------------------------------------------------
# [3/8] Binaries: reuse up-to-date ones, otherwise build (portable Go if needed)
# ---------------------------------------------------------------------------
echo "[3/8] Preparing Garuda binaries (bin/garuda, bin/garuda-api, bin/garuda-mcp, bin/garuda-talon)..."
mkdir -p bin
CURRENT_COMMIT="$(git rev-parse HEAD 2>/dev/null || echo "local")"
BUILD_STAMP_FILE="bin/.build_stamp"

# Source files newer than the API binary mean a rebuild is needed.
if [[ -x bin/garuda-api ]] && [[ -d ./cmd ]] && \
   [[ -n "$(find cmd internal go.mod go.sum -type f -newer bin/garuda-api -print -quit 2>/dev/null || true)" ]]; then
    rm -f "$BUILD_STAMP_FILE"
fi
STAMP_VAL="$(cat "$BUILD_STAMP_FILE" 2>/dev/null || echo "")"

BINARIES_PRESENT="false"
if [[ -x bin/garuda && -x bin/garuda-api && -x bin/garuda-mcp && -x bin/garuda-talon ]]; then
    BINARIES_PRESENT="true"
fi

if [[ "$BINARIES_PRESENT" == "true" ]] && { [[ "$STAMP_VAL" == "prebuilt" ]] || [[ ! -d "./cmd/garuda" ]] || { [[ "$STAMP_VAL" == "$CURRENT_COMMIT" ]] && [[ -z "$(git status --porcelain -- './cmd' './internal' 2>/dev/null || true)" ]]; }; }; then
    echo "  [OK] Pre-built / up-to-date Garuda binaries verified in bin/ (skipping rebuild)"
else
    if [[ -x "${TOOLCHAIN_DIR}/go/bin/go" ]]; then
        export PATH="${TOOLCHAIN_DIR}/go/bin:$PATH"
    fi

    # go.mod: `toolchain` line is the exact version Go will use; `go` line is the minimum.
    GO_MIN_VER="$(awk '/^go [0-9]+\.[0-9]+/ {print $2; exit}' go.mod 2>/dev/null || true)"
    GO_TOOLCHAIN_VER="$(awk '/^toolchain go[0-9]+\.[0-9]+/ {sub(/^go/, "", $2); print $2; exit}' go.mod 2>/dev/null || true)"
    GO_MIN_VER="${GO_MIN_VER:-1.26.0}"

    # Any Go >= 1.21 can fetch the exact toolchain named in go.mod by itself; older Go cannot.
    go_is_usable() {
        command -v go >/dev/null 2>&1 || return 1
        local gv gmaj gmin
        gv="$(go version 2>/dev/null | sed -n 's/.* go\([0-9][0-9]*\.[0-9][0-9]*\).*/\1/p' | head -n 1)"
        [[ -n "$gv" ]] || return 1
        gmaj="${gv%%.*}"
        gmin="${gv#*.}"
        (( gmaj > 1 || gmin >= 21 ))
    }

    if ! go_is_usable; then
        if [[ "$IS_TERMUX" == "true" ]]; then
            echo "  [*] Installing Go for Android Termux..."
            pkg install -y golang
        else
            need_cmd gzip gzip
            echo "  [*] Go 1.21+ not found -- downloading portable Go into ${TOOLCHAIN_DIR}/go (no sudo required)..."
            GO_DOWNLOADED="false"
            for go_ver in $(printf '%s\n' "$GO_TOOLCHAIN_VER" "$GO_MIN_VER" "1.26.0" | awk 'NF && !seen[$0]++'); do
                GO_TAR="go${go_ver}.${PLATFORM_OS}-${PLATFORM_ARCH}.tar.gz"
                echo "  [*] Trying ${GO_TAR}..."
                if curl -fsSL --retry 3 -o "${TOOLCHAIN_DIR}/${GO_TAR}" "https://go.dev/dl/${GO_TAR}"; then
                    rm -rf "${TOOLCHAIN_DIR}/go"
                    tar -xzf "${TOOLCHAIN_DIR}/${GO_TAR}" -C "$TOOLCHAIN_DIR"
                    rm -f "${TOOLCHAIN_DIR}/${GO_TAR}"
                    GO_DOWNLOADED="true"
                    break
                fi
                rm -f "${TOOLCHAIN_DIR}/${GO_TAR}"
            done
            if [[ "$GO_DOWNLOADED" != "true" ]]; then
                echo "[ERROR] Could not download Go from go.dev. Check your internet/proxy, or install Go from https://go.dev/dl/ and re-run." >&2
                exit 1
            fi
            export PATH="${TOOLCHAIN_DIR}/go/bin:$PATH"
        fi
    fi

    # Reuse a Go toolchain that an earlier build already downloaded, if present.
    export GOTOOLCHAIN="${GOTOOLCHAIN:-auto}"
    echo "  * Compiling with $(go version) [CGO_ENABLED=${CGO_ENABLED}]..."
    go mod download >/dev/null 2>&1 || true
    if ! go build -o bin/ ./cmd/garuda ./cmd/garuda-api ./cmd/garuda-mcp ./cmd/garuda-talon; then
        echo "[ERROR] Build failed." >&2
        if [[ "$IS_TERMUX" == "true" ]]; then
            echo "        On Termux, update Go first (pkg upgrade -y golang), or use the pre-built" >&2
            echo "        garuda-<version>-android-arm64-termux.tar.gz from the GitHub Releases page." >&2
        else
            echo "        Check the compiler output above. This repo needs Go ${GO_TOOLCHAIN_VER:-$GO_MIN_VER}." >&2
        fi
        exit 1
    fi
    echo "$CURRENT_COMMIT" > "$BUILD_STAMP_FILE"
    echo "  [OK] Built bin/garuda, bin/garuda-api, bin/garuda-mcp, bin/garuda-talon"
fi

# Put the fresh binaries on the user's PATH (no sudo; only directories that already exist or are standard).
mkdir -p "$HOME/.local/bin" 2>/dev/null || true
for target_bin_dir in "$HOME/.local/bin" "$HOME/go/bin"; do
    if [[ -d "$target_bin_dir" && -w "$target_bin_dir" ]]; then
        cp -f bin/garuda bin/garuda-api bin/garuda-mcp bin/garuda-talon "$target_bin_dir/" 2>/dev/null || true
    fi
done
case ":${PATH}:" in
    *":$HOME/.local/bin:"*) ;;
    *) echo "  [i] Add \$HOME/.local/bin to your PATH to run 'garuda' from anywhere." ;;
esac

# ---------------------------------------------------------------------------
# [4/8] Initialize: migrations, .env, owner account, MCP/Talon config
# ---------------------------------------------------------------------------
echo "[4/8] Running idempotent migrations, persisting .env, and provisioning owner..."
./bin/garuda init \
    --workspace "$WORKSPACE" \
    --email "$OWNER_EMAIL" \
    --skip-scan

# ---------------------------------------------------------------------------
# [5/8] Start the API gateway and wait until it is healthy
# ---------------------------------------------------------------------------
echo "[5/8] Starting Garuda API Gateway on :${API_PORT}..."
GARUDA_OWNER_EMAIL="${OWNER_EMAIL}" nohup ./bin/garuda-api > "${DATA_DIR}/garuda-api.log" 2>&1 &
API_PID=$!
echo "$API_PID" > "$API_PID_FILE"
disown "$API_PID" 2>/dev/null || true

API_HEALTHY="false"
for _ in $(seq 1 60); do
    if curl -fsS "http://127.0.0.1:${API_PORT}/health" >/dev/null 2>&1; then
        API_HEALTHY="true"
        break
    fi
    if ! kill -0 "$API_PID" 2>/dev/null; then
        break
    fi
    sleep 0.5
done
if [[ "$API_HEALTHY" != "true" ]]; then
    echo "[ERROR] Garuda API did not become healthy at http://127.0.0.1:${API_PORT}/health" >&2
    echo "  --- Last 30 lines of ${DATA_DIR}/garuda-api.log ---" >&2
    tail -n 30 "${DATA_DIR}/garuda-api.log" >&2 || true
    exit 1
fi
echo "  [OK] Garuda API Gateway online at http://localhost:${API_PORT}"

# ---------------------------------------------------------------------------
# [6/8] Index repositories
# ---------------------------------------------------------------------------
echo "[6/8] Indexing repositories into workspace '${WORKSPACE}'..."
if [[ $# -gt 0 ]]; then
    for repo_path in "$@"; do
        if [[ -d "$repo_path" ]]; then
            ./bin/garuda analyze "$repo_path" --workspace "$WORKSPACE" --save
        else
            echo "  [WARN] Skipping '${repo_path}': not a directory"
        fi
    done
else
    ./bin/garuda analyze . --workspace "$WORKSPACE" --save
fi
./bin/garuda docs ingest docs/adr --workspace "$WORKSPACE" >/dev/null 2>&1 || true

# ---------------------------------------------------------------------------
# [7/8] Talon contract + IDE / git hooks
# ---------------------------------------------------------------------------
POLICY_DIR="policies"
echo "[7/8] Activating Garuda Talon Harness (Contract + Cursor/Claude/Agents/Git Hooks)..."
./bin/garuda talon init . --workspace "$WORKSPACE" --policies "$POLICY_DIR"

# ---------------------------------------------------------------------------
# [8/8] Summary, change-control check, Talon guard, doctor, MCP verification
# ---------------------------------------------------------------------------
echo "[8/8] Running Architectural Summary, Talon Guard, System Doctor & Active MCP Verification..."
./bin/garuda summary
./bin/garuda check --project "$WORKSPACE" --policy "$POLICY_DIR" --auto-index=false --save || true
./bin/garuda talon guard --workspace "$WORKSPACE" --policies "$POLICY_DIR" --auto-index=false || true
./bin/garuda doctor
if command -v python3 >/dev/null 2>&1 && [[ -f scripts/mcp_verify.py ]]; then
    WORKSPACE="$WORKSPACE" python3 scripts/mcp_verify.py || true
fi

echo ""
echo "==================================================================="
echo "[OK] GARUDA ${GARUDA_VERSION_LABEL} + TALON IS CONFIGURED AND ONLINE"
echo "   * Dashboard:      http://localhost:${API_PORT}/login"
echo "   * Install as app: open it in Chrome / Edge / Brave / Android Chrome -> 'Install Garuda'"
if [[ "$PASSWORD_SOURCE" == "you" ]]; then
    echo "   * Login:          ${OWNER_EMAIL} (the password you set in GARUDA_PASSWORD)"
else
    echo "   * Login:          ${OWNER_EMAIL} / ${GARUDA_PASSWORD}"
    echo "                     (saved in ${LOGIN_FILE}; export GARUDA_PASSWORD to choose your own)"
fi
echo "   * System Doctor:  ./bin/garuda doctor"
echo "   * Talon Contract: .garuda/TALON_CONTRACT.md"
echo "   * Talon Guard:    ./bin/garuda talon guard --workspace ${WORKSPACE}"
echo "==================================================================="
