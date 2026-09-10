#!/usr/bin/env bash
set -e

echo "===================================================="
echo "🦅 PREPARING GARUDA ENTERPRISE PITCH DEMO"
echo "===================================================="

# 1. Compile latest binary with polyglot and document ingestion
echo "[1/4] Building Garuda Enterprise binary..."
go build -o bin/garuda ./cmd/garuda

# 2. Analyze the multi-repository microservices benchmark corpus
echo "[2/4] Indexing multi-repository microservices mesh (Auth + Gateway)..."
./bin/garuda analyze test/benchmark/truth_fixtures/016-multi-repo

# 3. Ingest real architectural decision records (ADRs)
echo "[3/4] Ingesting enterprise ADR specifications & compliance rules..."
./bin/garuda docs ingest docs/adr/0001-payment-idempotency.md

# 4. Verify cross-repository dependencies and drift
echo "[4/4] Executing cross-repo verification and compliance gates..."
./bin/garuda docs verify

echo ""
echo "===================================================="
echo "🚀 PITCH DEMO ENVIRONMENT READY"
echo "===================================================="
echo "To launch the live Company Graph Dashboard, run:"
echo "  ./bin/garuda dashboard web"
echo "===================================================="
