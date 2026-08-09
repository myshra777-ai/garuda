#!/bin/sh
set -e

# Wait for PgBouncer to be ready
echo "Waiting for PgBouncer..."
until pg_isready -h pgbouncer -p 6432 -U garuda -d garuda > /dev/null 2>&1; do
  sleep 1
done
echo "PgBouncer is ready"

# Run migrations
./garuda-api --migrate || true

# Start the API
exec ./garuda-api