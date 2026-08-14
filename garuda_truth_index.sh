#!/bin/bash
# garuda_truth_index.sh
# Comprehensive snapshot of Garuda's codebase, schema, and state

set -e

PROJECT_ROOT=$(pwd)
REPORT_FILE="garuda_truth_index_$(date +%Y%m%d_%H%M%S).md"
DATABASE_URL="${DATABASE_URL:-postgres://test:test@localhost:5433/garuda_test?sslmode=disable}"

echo "📊 Generating Garuda Truth Index..."
echo "Project root: $PROJECT_ROOT"
echo "Output: $REPORT_FILE"
echo ""

exec > >(tee -a "$REPORT_FILE") 2>&1

# ============================================================
# 1. Environment & Build Status
# ============================================================
echo "# 🧠 Garuda Truth Index"
echo ""
echo "**Generated:** $(date)"
echo "**Project:** $PROJECT_ROOT"
echo "**Go Version:** $(go version 2>/dev/null || echo 'N/A')"
echo ""
echo "## 1. Build & Test Status"
echo ""
echo -n "- Build: "
go build ./... 2>&1 >/dev/null && echo "✅ PASS" || echo "❌ FAIL"
echo -n "- Tests (short): "
go test ./... -short 2>&1 >/dev/null && echo "✅ PASS" || echo "❌ FAIL"
echo ""

# ============================================================
# 2. Database Schema
# ============================================================
echo "## 2. Database Schema"
echo ""

if command -v psql &>/dev/null && psql "$DATABASE_URL" -c "SELECT 1" &>/dev/null 2>&1; then
    echo "✅ Database reachable at $DATABASE_URL"
    echo ""
    
    # List all tables
    echo "### Tables"
    psql "$DATABASE_URL" -t -c "SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename;" 2>/dev/null | grep -v '^$' | sed 's/^/  - /'
    echo ""
    
    # For each key table, show schema
    for table in decisions decision_revisions evidence_store merkle_roots audit_events workspaces repositories harvested_decisions contradictions; do
        if psql "$DATABASE_URL" -t -c "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='$table')" 2>/dev/null | grep -q t; then
            echo "### Table: \`$table\`"
            echo '```sql'
            psql "$DATABASE_URL" -c "\d $table" 2>/dev/null
            echo '```'
            echo ""
        fi
    done
    
    # Migration status
    echo "### Migration Status"
    echo '```sql'
    psql "$DATABASE_URL" -c "SELECT version, applied_at FROM migrations ORDER BY version;" 2>/dev/null || echo "No migrations table found"
    echo '```'
    echo ""
else
    echo "⚠️ Database not reachable (set DATABASE_URL)"
    echo ""
fi

# ============================================================
# 3. Go Core Types
# ============================================================
echo "## 3. Go Core Types"
echo ""

echo "### Decision"
grep -A30 "type Decision struct" internal/types/types.go 2>/dev/null || echo "Not found"
echo ""

echo "### DecisionRevision"
grep -A20 "type DecisionRevision struct" internal/types/types.go 2>/dev/null || echo "Not found"
echo ""

echo "### Contradiction"
grep -A15 "type Contradiction struct" internal/types/types.go 2>/dev/null || echo "Not found"
echo ""

echo "### Evidence"
grep -A10 "type Evidence struct" internal/types/types.go 2>/dev/null || echo "Not found"
echo ""

echo "### Scope"
grep -A10 "type Scope struct" internal/types/types.go 2>/dev/null || echo "Not found"
echo ""

echo "### Analyze Result"
grep -A30 "type Result struct" internal/analyzer/model.go 2>/dev/null || echo "Not found"
echo ""

# ============================================================
# 4. Store Methods
# ============================================================
echo "## 4. Store Methods"
echo ""

echo "### Decision Store Methods"
grep -E "^func \(s \*PostgresStore\)" internal/store/decision_store.go 2>/dev/null | sed 's/^/  - /' || echo "  - None found"
echo ""

echo "### Revision Store Methods"
grep -E "^func \(s \*PostgresStore\)" internal/store/revision_store.go 2>/dev/null | sed 's/^/  - /' || echo "  - None found"
echo ""

echo "### Contradiction Store Methods"
grep -E "^func \(s \*PostgresStore\)" internal/store/contradiction_store.go 2>/dev/null | sed 's/^/  - /' || echo "  - None found"
echo ""

echo "### Workspace Store Methods"
grep -E "^func \(s \*PostgresStore\)" internal/store/workspace.go 2>/dev/null | sed 's/^/  - /' || echo "  - None found"
echo ""

echo "### Analysis Store Methods"
grep -E "^func \(s \*PostgresStore\)" internal/store/analysis.go 2>/dev/null | sed 's/^/  - /' || echo "  - None found"
echo ""

# ============================================================
# 5. API Handlers
# ============================================================
echo "## 5. API Handlers"
echo ""

echo "### Registered Routes"
grep -E "mux\.HandleFunc" cmd/garuda-api/main.go 2>/dev/null | sed 's/^/  - /' || echo "  - None found"
echo ""

echo "### Handler Functions"
grep -E "^func \(s \*Server\) Handle" internal/api/*.go 2>/dev/null | sed 's/^/  - /' || echo "  - None found"
echo ""

# ============================================================
# 6. CLI Commands
# ============================================================
echo "## 6. CLI Commands"
echo ""

echo "### Root Commands"
grep -E "^\s+rootCmd\.AddCommand" cmd/garuda/main.go 2>/dev/null | sed 's/^/  - /' || echo "  - None found"
echo ""

echo "### Command Definitions"
grep -E "^var \w+Cmd = &cobra\.Command" cmd/garuda/main.go 2>/dev/null | sed 's/^/  - /' || echo "  - None found"
echo ""

# ============================================================
# 7. Migration Files
# ============================================================
echo "## 7. Migration Files"
echo ""

echo "### Applied Migrations"
ls -1 migrations/*.sql 2>/dev/null | sed 's/^/  - /' || echo "  - No migrations"
echo ""

echo "### Migration Status (from DB)"
if psql "$DATABASE_URL" -c "SELECT version FROM migrations ORDER BY version;" 2>/dev/null; then
    echo "  - Applied: $(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM migrations;" 2>/dev/null | tr -d ' ')"
    echo "  - Pending: $(ls -1 migrations/*.sql 2>/dev/null | wc -l) total - $(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM migrations;" 2>/dev/null | tr -d ' ')"
else
    echo "  - Migration table not found"
fi
echo ""

# ============================================================
# 8. Summary / Key Findings
# ============================================================
echo "## 8. Key Findings"
echo ""

# Check for critical columns
echo "### Critical Schema Checks"
if psql "$DATABASE_URL" -c "SELECT 1 FROM information_schema.columns WHERE table_name='decisions' AND column_name='tenant_id'" 2>/dev/null | grep -q 1; then
    echo "  - decisions.tenant_id: ✅ EXISTS"
else
    echo "  - decisions.tenant_id: ❌ MISSING"
fi

if psql "$DATABASE_URL" -c "SELECT 1 FROM information_schema.columns WHERE table_name='decision_revisions' AND column_name='decision_hash'" 2>/dev/null | grep -q 1; then
    echo "  - decision_revisions.decision_hash: ✅ EXISTS"
else
    echo "  - decision_revisions.decision_hash: ❌ MISSING"
fi

if psql "$DATABASE_URL" -c "SELECT 1 FROM information_schema.columns WHERE table_name='evidence_store' AND column_name='tenant_id'" 2>/dev/null | grep -q 1; then
    echo "  - evidence_store.tenant_id: ✅ EXISTS"
else
    echo "  - evidence_store.tenant_id: ❌ MISSING"
fi

if psql "$DATABASE_URL" -c "SELECT 1 FROM information_schema.tables WHERE table_name='merkle_roots'" 2>/dev/null | grep -q t; then
    echo "  - merkle_roots table: ✅ EXISTS"
else
    echo "  - merkle_roots table: ❌ MISSING"
fi
echo ""

# ============================================================
# 9. Git Status
# ============================================================
echo "## 9. Git Status"
echo ""
git log --oneline -5 2>/dev/null | sed 's/^/  - /' || echo "  - Not a git repo"
echo ""

echo "---"
echo "✅ Truth Index generated: $REPORT_FILE"
echo ""
echo "📌 Next: Review this file to identify schema/code mismatches."