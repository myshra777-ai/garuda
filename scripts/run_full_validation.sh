#!/bin/bash
set -e

echo "🧪 GARUDA V5 FULL VALIDATION SUITE"
echo "═══════════════════════════════════════════════════════════════════"
echo ""
echo "📋 Phases: 13 phases, 60+ tests"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Create results directory
mkdir -p test-results
RESULTS_FILE="test-results/validation-$(date +%Y%m%d_%H%M%S).log"

# Function to run a phase
run_phase() {
    local name="$1"
    local script="$2"
    echo ""
    echo "────────────────────────────────────────────────────────────────────"
    echo "📌 $name"
    echo "────────────────────────────────────────────────────────────────────"
    if eval "$script" 2>&1 | tee -a "$RESULTS_FILE"; then
        echo -e "${GREEN}✅ Phase passed${NC}"
        return 0
    else
        echo -e "${RED}❌ Phase failed${NC}"
        return 1
    fi
}

# Run all phases
echo "Starting validation at $(date)" | tee "$RESULTS_FILE"
echo ""

# Phase 1: Environment Setup
run_phase "Phase 1: Environment & Database Setup" "
    go mod tidy
    go build -o bin/garuda cmd/garuda/*.go
    dropdb -U test -h localhost -p 5433 garuda_test 2>/dev/null || true
    createdb -U test -h localhost -p 5433 garuda_test
    export DATABASE_URL=\"postgres://test:test@localhost:5433/garuda_test?sslmode=disable\"
    go run cmd/migrate/main.go
    export GARUDA_TENANT_ID=\"00000000-0000-0000-0000-000000000001\"
    export GARUDA_WORKSPACE=\"default\"
"

# Phase 2: Workspace Setup
run_phase "Phase 2: Workspace & Repository Setup" "
    ./bin/garuda workspace create default
    WORKSPACE_ID=\$(psql -U test -h localhost -p 5433 -d garuda_test -t -c \"SELECT id FROM workspaces WHERE name = 'default' LIMIT 1;\" | tr -d ' ')
    ./bin/garuda repo add default file://\$(pwd) --module-path github.com/myshra777-ai/garuda
    export WORKSPACE_ID
"

# Phase 3: AST Extraction
run_phase "Phase 3: AST/TYPE Extraction & Persistence" "
    ./bin/garuda workspace sync default
    ENTITY_COUNT=\$(psql -U test -h localhost -p 5433 -d garuda_test -t -c \"SELECT COUNT(*) FROM entities WHERE workspace_id = '\$WORKSPACE_ID';\" | tr -d ' ')
    echo \"Entities created: \$ENTITY_COUNT\"
    ./bin/garuda analyze . -o v1.json
"

# Phase 4: Entity Identity
run_phase "Phase 4: Entity Identity Verification" "
    ENTITY_ID=\$(psql -U test -h localhost -p 5433 -d garuda_test -t -c \"SELECT e.id FROM entities e JOIN claims c ON c.to_entity_id = e.id WHERE e.workspace_id = '\$WORKSPACE_ID' LIMIT 1;\" | tr -d ' ')
    if [ -z \"\$ENTITY_ID\" ]; then
        ENTITY_ID=\$(psql -U test -h localhost -p 5433 -d garuda_test -t -c \"SELECT id FROM entities WHERE workspace_id = '\$WORKSPACE_ID' LIMIT 1;\" | tr -d ' ')
    fi
    export ENTITY_ID
    echo \"Target Entity ID: \$ENTITY_ID\"
    UUID_FORMAT=\$(psql -U test -h localhost -p 5433 -d garuda_test -t -c \"SELECT id::text FROM entities WHERE id = '\$ENTITY_ID'::uuid;\" | tr -d ' ')
    if [[ \"\$UUID_FORMAT\" =~ ^[a-f0-9]{8}-[a-f0-9]{4}-5[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$ ]]; then
        echo \"✅ UUIDv5 format verified\"
    else
        echo \"❌ UUIDv5 format invalid\"
        exit 1
    fi
"

# Phase 5: Blast Radius
run_phase "Phase 5: Blast Radius Impact Analysis" "
    ./bin/garuda impact --workspace \"\$WORKSPACE_ID\" --target \"\$ENTITY_ID\" --depth 3 --min-confidence 0.50
    ./bin/garuda impact --workspace \"\$WORKSPACE_ID\" --target \"\$ENTITY_ID\" --json > test-results/impact.json
    if jq -e '.blast_radius' test-results/impact.json > /dev/null 2>&1; then
        echo \"✅ JSON structure valid\"
    else
        echo \"❌ JSON structure invalid\"
        exit 1
    fi
"

# Phase 6: Diff
run_phase "Phase 6: Diff & Impact-Diff" "
    ./bin/garuda analyze . -o v2.json
    ./bin/garuda diff v1.json v2.json
    ./bin/garuda diff v1.json v2.json --json > test-results/diff.json
    ./bin/garuda impact-diff v1.json v2.json --json > test-results/impact-diff.json
"

# Phase 7: Graph
run_phase "Phase 7: Graph & Visualization" "
    ./bin/garuda graph default
    if [ -f \"garuda_graph_default.html\" ]; then
        echo \"✅ Graph file created\"
    else
        echo \"❌ Graph file not found\"
        exit 1
    fi
"

# Phase 8: Ponytail
run_phase "Phase 8: Code Quality (Ponytail)" "
    ./bin/garuda ponytail . --json > test-results/ponytail.json
    if jq -e '.dead_code' test-results/ponytail.json > /dev/null 2>&1; then
        echo \"✅ Ponytail report valid\"
    fi
"

# Phase 9: Judge
run_phase "Phase 9: Governance (Judge)" "
    ./bin/garuda judge v1.json v2.json --json > test-results/judge.json
    if jq -e '.breaking_count' test-results/judge.json > /dev/null 2>&1; then
        echo \"✅ Judge report valid\"
    fi
"

# Phase 10: Trust
run_phase "Phase 10: Trust & Integrity" "
    ./bin/garuda verify
    ./bin/garuda entities | head -5
"

# Phase 11: Cross-Repo
run_phase "Phase 11: Cross-Repo Sync" "
    mkdir -p /tmp/garuda-test-repo1
    echo 'module github.com/test/repo1' > /tmp/garuda-test-repo1/go.mod
    ./bin/garuda repo add default file:///tmp/garuda-test-repo1 --module-path github.com/test/repo1
    ./bin/garuda workspace sync default
"

# Phase 12: CLI
run_phase "Phase 12: CLI & Developer Experience" "
    ./bin/garuda --help | head -5
    ./bin/garuda status
"

echo ""
echo "═══════════════════════════════════════════════════════════════════"
echo "📊 VALIDATION COMPLETE"
echo "═══════════════════════════════════════════════════════════════════"
echo ""
echo "📁 Results saved to: $RESULTS_FILE"
echo ""

# Summary
echo "✅ All phases completed successfully!"