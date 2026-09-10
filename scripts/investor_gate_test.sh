#!/usr/bin/env bash
set -e

echo "===================================================="
echo "🦅 GARUDA INVESTOR READINESS GATE & PRE-FLIGHT TEST"
echo "===================================================="

# 1. Ensure test workspace has a valid Go module for workspace discovery
if [ ! -f "test_workspace/go.mod" ]; then
    echo "[0/5] Initializing sandbox Go module..."
    cd test_workspace && go mod init github.com/myshra777-ai/garuda-sandbox && cd ..
fi

# 2. Compile binary
echo "[1/5] Compiling Garuda binary..."
go build -o bin/garuda ./cmd/garuda

# 3. Run static analysis
echo "[2/5] Executing static and polyglot analysis..."
./bin/garuda analyze .
./bin/garuda analyze test_workspace

# 4. Ingest specs using the document ingestion pipeline
echo "[3/5] Ingesting architectural specifications & ADRs..."
./bin/garuda docs ingest test_workspace/docs/architecture_roadmap.md

# 5. Run verification & drift detection
echo "[4/5] Running documentation health & drift gates..."
./bin/garuda docs verify

echo ""
echo "===================================================="
echo "✨ ALL INVESTOR VALIDATION GATES PASSED SUCCESSFULLY"
echo "===================================================="
echo "Launch Mission Control for your pitch demo:"
echo "  ./bin/garuda dashboard web"
echo "===================================================="
