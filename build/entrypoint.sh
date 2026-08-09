#!/bin/sh
set -e

# Wait for PostgreSQL to be ready (if using local DB)
if [ -n "$DATABASE_URL" ]; then
    echo "Waiting for database..."
    until pg_isready -d "$DATABASE_URL" 2>/dev/null || psql "$DATABASE_URL" -c "SELECT 1" 2>/dev/null; do
        sleep 1
    done
    echo "Database is ready!"
fi

# Run migrations (if --migrate flag is passed)
if [ "$1" = "--migrate" ]; then
    echo "Running migrations..."
    ./garuda-api --migrate || true
    exit 0
fi

# Start the API
exec ./garuda-api
