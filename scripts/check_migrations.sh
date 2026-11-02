#!/usr/bin/env bash
# Verifies migrations/*.sql apply cleanly on an EMPTY database (cold boot) and are
# idempotent (second full pass). Any SQL error fails the script; nothing is masked.
#
# SAFETY: this script NEVER touches the database you point it at. It connects to the
# server in the URL, creates a brand-new scratch database, runs both passes there, and
# drops it. (Several migrations delete or drop data when re-run, so they must never be
# pointed at a database that holds real data.)
#
# usage: scripts/check_migrations.sh "postgres://user:pass@host:port/anydb?sslmode=disable"
#        (the user needs CREATEDB)
set -euo pipefail
URL="${1:-${DATABASE_URL:-}}"
[[ -n "$URL" ]] || { echo "usage: $0 <DATABASE_URL of a server where a scratch DB may be created>" >&2; exit 2; }
DIR="$(cd "$(dirname "$0")/.." && pwd)/migrations"

base="${URL%%\?*}"; query="${URL#"$base"}"; root="${base%/*}"
SCRATCH="garuda_migcheck_$(date +%s)_$RANDOM"
DB_URL="$root/$SCRATCH$query"
ADMIN_URL="$URL"

psql "$ADMIN_URL" -X -q -v ON_ERROR_STOP=1 -c "CREATE DATABASE \"$SCRATCH\"" >/dev/null
trap 'psql "$ADMIN_URL" -X -q -c "DROP DATABASE IF EXISTS \"'"$SCRATCH"'\"" >/dev/null 2>&1 || true' EXIT
echo "using scratch database $SCRATCH (dropped on exit); your database is not modified"

fail=0
for pass in 1 2; do
  label=$([[ $pass == 1 ]] && echo "cold-boot" || echo "re-run")
  n=0
  for f in $(ls "$DIR"/*.sql | LC_ALL=C sort); do
    n=$((n+1))
    if ! out=$(psql "$DB_URL" -X -q -v ON_ERROR_STOP=1 --single-transaction -f "$f" 2>&1 >/dev/null); then
      echo "[$label] FAIL $(basename "$f"): $(echo "$out" | grep -m1 ERROR || echo "$out" | head -1)"
      fail=$((fail+1))
    fi
  done
  echo "[$label] files=$n failures_so_far=$fail"
done
[[ $fail -eq 0 ]] && echo "migrations OK (cold boot + idempotent re-run)" || { echo "migrations FAILED: $fail" >&2; exit 1; }
