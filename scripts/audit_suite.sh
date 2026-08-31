#!/usr/bin/env bash
set -euo pipefail

echo "================================================="
echo "  🦅 GARUDA PRE-AUDIT HARDENING & VERIFICATION"
echo "================================================="

echo "1. Checking Go formatting & vet..."
go vet ./...
test -z "$(gofmt -l $(find cmd internal -name '*.go'))" || {
echo "❌ unformatted Go files detected. Run 'go fmt ./...'"
exit 1
}

echo "2. Running Race Detector across all packages..."
go test -race -v -count=1 ./internal/... ./cmd/...

echo "3. Testing Contract Type Exporter..."
go test -v ./internal/contract/...

echo "4. Testing Token Guard Middleware..."
go test -v ./internal/mcp/...

echo "================================================="
echo "✅ All invariants verified. Codebase ready for audit."
echo "================================================="