#!/bin/bash
# Garuda Phase 0 Status Checker
# Run this from the root of your garuda repository.

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   Garuda Phase 0 Status Check         ${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

missing=0
partial=0

# Function to check file existence
check_file() {
    if [ -f "$1" ]; then
        echo -e "${GREEN}✅ $1${NC}"
        return 0
    else
        echo -e "${RED}❌ $1${NC}"
        missing=$((missing+1))
        return 1
    fi
}

# Function to check file and grep for a pattern
check_file_content() {
    if [ -f "$1" ]; then
        if grep -q "$2" "$1" 2>/dev/null; then
            echo -e "${GREEN}✅ $1 (contains '$2')${NC}"
        else
            echo -e "${YELLOW}⚠️ $1 (missing expected pattern: '$2')${NC}"
            partial=$((partial+1))
        fi
    else
        echo -e "${RED}❌ $1${NC}"
        missing=$((missing+1))
    fi
}

echo -e "${BLUE}--- Phase 0: Core Types & Registry ---${NC}"
check_file "internal/types/types.go"
check_file_content "internal/types/types.go" "type Decision struct"
check_file_content "internal/types/types.go" "type EvidenceHash"
check_file_content "internal/types/types.go" "type Contradiction struct"
check_file "internal/registry/registry.go"
check_file_content "internal/registry/registry.go" "type DecisionStore interface"
check_file_content "internal/registry/registry.go" "SaveDecision"
echo ""

echo -e "${BLUE}--- Phase 0: Storage Layer ---${NC}"
check_file "internal/store/postgres.go"
check_file_content "internal/store/postgres.go" "IngestEvidence"
check_file_content "internal/store/postgres.go" "type PostgresStore struct"
check_file "internal/store/decision_store.go"
check_file_content "internal/store/decision_store.go" "SaveDecision"
check_file_content "internal/store/decision_store.go" "GetDecision"
check_file_content "internal/store/decision_store.go" "GetDecisionRevisions"
check_file "internal/store/refcount.go"
check_file_content "internal/store/refcount.go" "EmitReferenceChange"
check_file "internal/store/flusher.go"
check_file_content "internal/store/flusher.go" "StartRefCountFlusher"
check_file "internal/store/redis.go"
check_file_content "internal/store/redis.go" "EnsureStreamGroup"
check_file "internal/store/migrate.go"
check_file "internal/store/integration_test.go"
echo ""

echo -e "${BLUE}--- Phase 0: API Layer (partial) ---${NC}"
check_file "internal/api/handler.go"
check_file_content "internal/api/handler.go" "type Server struct"
check_file "internal/api/middleware.go"
check_file_content "internal/api/middleware.go" "WithAuth"
check_file "internal/api/decision_handlers.go"
check_file_content "internal/api/decision_handlers.go" "HandleProposeDecision"
check_file_content "internal/api/decision_handlers.go" "HandleGetDecisionRevisions"
check_file "internal/api/api_test.go"
echo ""

echo -e "${BLUE}--- Phase 0: Auth Layer ---${NC}"
check_file "internal/auth/jwt.go"
check_file_content "internal/auth/jwt.go" "NewJWTConfig"
check_file "internal/auth/user.go"
check_file_content "internal/auth/user.go" "HashPassword"
check_file "internal/auth/service.go"
check_file_content "internal/auth/service.go" "NewAuthService"
check_file "internal/auth/jwt_test.go"
echo ""

echo -e "${BLUE}--- Phase 0: Engine & Lineage ---${NC}"
check_file "internal/engine/contradiction.go"
check_file_content "internal/engine/contradiction.go" "ValidateDecision"
check_file "internal/engine/lineage.go"
check_file_content "internal/engine/lineage.go" "RegisterDecision"
check_file "internal/engine/contradiction_test.go"
check_file "internal/lineage/graph.go"
check_file_content "internal/lineage/graph.go" "AddEdge"
check_file "internal/lineage/graph_test.go"
echo ""

echo -e "${BLUE}--- Phase 0: Migrations ---${NC}"
check_file "migrations/001_cas_blocks.sql"
check_file "migrations/002_task_manifests.sql"
check_file "migrations/003_truth_graph_core.sql"
check_file "migrations/004_facts_assumptions.sql"
check_file "migrations/005_decision_revisions.sql"
echo ""

echo -e "${BLUE}--- Phase 0: Command entrypoints ---${NC}"
check_file "cmd/garuda-api/main.go"
check_file_content "cmd/garuda-api/main.go" "NewServer"
check_file "cmd/migrate/main.go"
check_file_content "cmd/migrate/main.go" "store.Migrate"
echo ""

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   Summary                              ${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✅ Files present: $(( $(find . -type f -name "*.go" -o -name "*.sql" | wc -l) - missing ))${NC}"
echo -e "${RED}❌ Missing files: $missing${NC}"
echo -e "${YELLOW}⚠️ Partial/Incomplete: $partial${NC}"

if [ $missing -eq 0 ] && [ $partial -eq 0 ]; then
    echo -e "\n${GREEN}🎉 Phase 0 is fully complete. You are ready to start Phase 1.${NC}"
else
    echo -e "\n${YELLOW}⚠️ Please add the missing files or correct the partial ones before proceeding.${NC}"
fi