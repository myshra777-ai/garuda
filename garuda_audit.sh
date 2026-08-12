#!/bin/bash
# garuda_audit.sh
# Comprehensive audit after Copilot auto-fixes

set -e

PROJECT_ROOT=$(pwd)
REPORT_FILE="garuda_audit_$(date +%Y%m%d_%H%M%S).txt"
ERRORS=0

echo "🔍 Garuda Hardening & Regression Audit"
echo "======================================="
echo "Project root: $PROJECT_ROOT"
echo "Output will be written to: $REPORT_FILE"
echo ""

exec > >(tee -a "$REPORT_FILE") 2>&1

# ---------- Colors ----------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# ---------- 1. Build & Test Status ----------
echo "## 1. Build & Test Status"
echo "--------------------------"

echo -n "▶ go build ./... "
if go build ./... 2>&1; then
    echo -e "${GREEN}✅ PASS${NC}"
else
    echo -e "${RED}❌ FAIL${NC}"
    ERRORS=$((ERRORS+1))
fi

echo -n "▶ go vet ./... "
if go vet ./... 2>&1; then
    echo -e "${GREEN}✅ PASS${NC}"
else
    echo -e "${RED}❌ FAIL${NC}"
    ERRORS=$((ERRORS+1))
fi

echo -n "▶ go test ./... (short) "
if go test ./... -short 2>&1 | grep -q "^FAIL"; then
    echo -e "${RED}❌ FAIL${NC}"
    ERRORS=$((ERRORS+1))
else
    echo -e "${GREEN}✅ PASS${NC}"
fi

echo ""

# ---------- 2. Critical Hardening Checks ----------
echo "## 2. Critical Hardening Checks"
echo "--------------------------------"

# Check 1: Immutable revisions (should NOT have ON CONFLICT on decisions)
echo -n "▶ 1. Mutable decisions (ON CONFLICT) ... "
if grep -r "ON CONFLICT.*decisions" internal/store/*.go 2>/dev/null | grep -v "decision_revisions" | grep -v "migration" > /dev/null; then
    echo -e "${RED}❌ FAIL${NC} — ON CONFLICT found in decisions table (mutable)"
    ERRORS=$((ERRORS+1))
else
    echo -e "${GREEN}✅ PASS${NC} — no ON CONFLICT on decisions"
fi

# Check 2: Decision revisions present
echo -n "▶ 2. Decision revisions table exists ... "
if grep -q "CREATE TABLE.*decision_revisions" migrations/*.sql 2>/dev/null; then
    echo -e "${GREEN}✅ PASS${NC}"
else
    echo -e "${RED}❌ FAIL${NC} — decision_revisions table missing"
    ERRORS=$((ERRORS+1))
fi

# Check 3: Actor derived from context (not from client)
echo -n "▶ 3. Actor derived from auth context ... "
if grep -r "req\.Actor" internal/api/*.go 2>/dev/null | grep -v "omitempty" > /dev/null; then
    echo -e "${RED}❌ FAIL${NC} — client-provided actor found"
    ERRORS=$((ERRORS+1))
else
    echo -e "${GREEN}✅ PASS${NC} — actor from context only"
fi

# Check 4: Debug token guarded
echo -n "▶ 4. Debug token production guard ... "
if grep -r "HandleDebugToken" internal/api/*.go 2>/dev/null | grep -q "GARUDA_ENV"; then
    echo -e "${GREEN}✅ PASS${NC}"
else
    echo -e "${YELLOW}⚠️ WARN${NC} — debug token may not be guarded"
fi

# Check 5: Safe error responses (not leaking internals)
echo -n "▶ 5. Safe error responses ... "
if grep -r "http.Error.*err\.Error()" internal/api/*.go 2>/dev/null > /dev/null; then
    echo -e "${RED}❌ FAIL${NC} — internal errors leaked to client"
    ERRORS=$((ERRORS+1))
else
    echo -e "${GREEN}✅ PASS${NC} — no error leakage"
fi

# Check 6: Merkle/hash chain atomicity
echo -n "▶ 6. Merkle update in same transaction ... "
if grep -A20 "BEGIN" internal/store/revision_store.go 2>/dev/null | grep -q "UPDATE merkle_roots"; then
    echo -e "${GREEN}✅ PASS${NC}"
else
    echo -e "${RED}❌ FAIL${NC} — Merkle update not in transaction"
    ERRORS=$((ERRORS+1))
fi

# Check 7: ON DELETE CASCADE removed from revisions
echo -n "▶ 7. ON DELETE CASCADE removed ... "
if grep -r "ON DELETE CASCADE" internal/store/*.go 2>/dev/null | grep -v "evidence" | grep -v "migration" | grep -v "contradictions" > /dev/null; then
    echo -e "${RED}❌ FAIL${NC} — CASCADE still found"
    ERRORS=$((ERRORS+1))
else
    echo -e "${GREEN}✅ PASS${NC}"
fi

# Check 8: Request ID middleware
echo -n "▶ 8. Request ID middleware ... "
if grep -r "request_id" internal/api/middleware.go 2>/dev/null > /dev/null; then
    echo -e "${GREEN}✅ PASS${NC}"
else
    echo -e "${YELLOW}⚠️ WARN${NC} — request ID middleware may be missing"
fi

echo ""

# ---------- 3. Reverted Weak Code Detection ----------
echo "## 3. Reverted Weak Code Detection"
echo "-----------------------------------"

echo "▶ Looking for old patterns that may have been reintroduced:"

# Old mutable decision pattern
if grep -r "SaveDecision" internal/store/*.go 2>/dev/null | grep -v "revision" > /dev/null; then
    echo -e "  ${YELLOW}⚠️ SaveDecision found — may be old mutable pattern${NC}"
fi

# Hardcoded policy checks
if grep -r "do not change schema" internal/engine/*.go 2>/dev/null > /dev/null; then
    echo -e "  ${YELLOW}⚠️ Hardcoded policy string found${NC}"
fi

# Check for old contradiction engine (string matching)
if grep -r "postgres.*mysql" internal/engine/*.go 2>/dev/null > /dev/null; then
    echo -e "  ${YELLOW}⚠️ Heuristic contradiction engine may still be present${NC}"
fi

# Check for consensus stub
if grep -r "AST_VERIFIED_DIGEST" internal/engine/*.go 2>/dev/null > /dev/null; then
    echo -e "  ${YELLOW}⚠️ Consensus stub still present${NC}"
fi

echo ""

# ---------- 4. Git Diff Summary ----------
echo "## 4. Git Diff Summary"
echo "----------------------"

echo "▶ Recent changes (last 5 commits):"
git log --oneline -5 2>/dev/null || echo "Not a git repo"

echo ""
echo "▶ Uncommitted changes:"
git status --porcelain 2>/dev/null || echo "Not a git repo"

echo ""
echo "▶ Diff summary (since last known good state):"
echo "   (If you have a reference branch, use: git diff main..refactor)"
echo "   Showing staged + unstaged changes:"
git diff --stat HEAD 2>/dev/null || echo "No diff"

echo ""

# ---------- 5. Database Schema Check ----------
echo "## 5. Database Schema Check"
echo "----------------------------"

if command -v psql &>/dev/null && psql "$DATABASE_URL" -c "SELECT 1" &>/dev/null 2>&1; then
    echo "✅ Database reachable"
    
    # Check if decision_revisions table exists
    if psql "$DATABASE_URL" -t -c "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='decision_revisions')" 2>/dev/null | grep -q t; then
        echo -e "  decision_revisions: ${GREEN}✅ EXISTS${NC}"
    else
        echo -e "  decision_revisions: ${RED}❌ MISSING${NC}"
        ERRORS=$((ERRORS+1))
    fi
    
    # Check if ON DELETE CASCADE is present on revisions
    if psql "$DATABASE_URL" -t -c "SELECT 1 FROM information_schema.constraint_column_usage WHERE table_name='decision_revisions' AND constraint_name LIKE '%cascade%'" 2>/dev/null | grep -q 1; then
        echo -e "  ON DELETE CASCADE: ${RED}❌ FOUND${NC} (should be RESTRICT)"
        ERRORS=$((ERRORS+1))
    else
        echo -e "  ON DELETE CASCADE: ${GREEN}✅ not present${NC}"
    fi
    
    # Check merkle_roots table
    if psql "$DATABASE_URL" -t -c "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='merkle_roots')" 2>/dev/null | grep -q t; then
        echo -e "  merkle_roots: ${GREEN}✅ EXISTS${NC}"
    else
        echo -e "  merkle_roots: ${YELLOW}⚠️ MISSING${NC}"
    fi
    
else
    echo -e "${YELLOW}⚠️ Database not reachable (set DATABASE_URL)${NC}"
fi

echo ""

# ---------- 6. Summary ----------
echo "## 6. Summary of Findings"
echo "--------------------------"

if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✅ All critical checks passed.${NC}"
    echo "   Garuda appears to be in a hardened state."
else
    echo -e "${RED}❌ $ERRORS critical issue(s) found.${NC}"
    echo "   Please review the audit report above."
fi

echo ""
echo "📄 Audit report saved to: $REPORT_FILE"
echo ""

# Exit with error code if any issues found
exit $ERRORS