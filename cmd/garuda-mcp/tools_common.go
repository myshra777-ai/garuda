// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/knowledge"
	"github.com/myshra777-ai/garuda/internal/policy"
	"github.com/myshra777-ai/garuda/internal/store"
)

var ErrApprovalRequiresHuman = errors.New("human approval requires GARUDA_MCP_APPROVER to be set in the server environment or CLI co-signing via 'garuda audit ... --decide' / 'garuda rulebook approve'")

func ResolveApprover(requestedBy string) (string, error) {
	envApprover := strings.TrimSpace(os.Getenv("GARUDA_MCP_APPROVER"))
	req := strings.TrimSpace(requestedBy)
	if envApprover == "" {
		return "", ErrApprovalRequiresHuman
	}
	if req != "" && !strings.EqualFold(req, envApprover) {
		return "", fmt.Errorf("approved_by %q does not match attested GARUDA_MCP_APPROVER %q", req, envApprover)
	}
	return envApprover, nil
}

func ReadOnlyMode() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("GARUDA_MCP_READ_ONLY"))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

func ReadOnlyRejection(tool string) map[string]any {
	return map[string]any{
		"status":    "rejected",
		"read_only": true,
		"tool":      tool,
		"reason":    "GARUDA_MCP_READ_ONLY is enabled; mutating MCP tools are disabled in this session",
	}
}

func IsMutatingTool(toolName string, args map[string]interface{}) bool {
	switch toolName {
	case "garuda.propose_decision",
		"garuda.decide_logic_variant",
		"garuda.reconcile_data_audit",
		"garuda.reconcile_financial_audit",
		"garuda.approve_base_rulebook",
		"garuda.synthesize_rulebook",
		"garuda.docs_ingest",
		"garuda.create_task",
		"garuda.handoff",
		"garuda.resume",
		"garuda.reset_audit_state",
		"garuda.create_workspace":
		return true
	case "garuda.criticality_tiers":
		if s, ok := args["surface_name"].(string); ok && strings.TrimSpace(s) != "" {
			return true
		}
		if p, ok := args["package_pattern"].(string); ok && strings.TrimSpace(p) != "" {
			return true
		}
		return false
	case "garuda.talon_verify_edit":
		if v, ok := args["save_to_ledger"].(bool); ok && v {
			return true
		}
		if v, ok := args["save_ledger"].(bool); ok && v {
			return true
		}
		return false
	default:
		return false
	}
}

// IsFixturePath returns true if a file path or package belongs to benchmark, test, or tutorial fixtures (B-01, B-14, B-18).
func IsFixturePath(pathOrPkg string) bool {
	p := strings.ToLower(filepath.ToSlash(strings.TrimSpace(pathOrPkg)))
	return strings.Contains(p, "test/benchmark/truth_fixtures") ||
		strings.Contains(p, "test/fixtures/") ||
		strings.Contains(p, "fixtures/") ||
		strings.Contains(p, "active_contradictions") ||
		strings.Contains(p, "/testdata/") ||
		strings.Contains(p, "garuda-bench") ||
		strings.Contains(p, "docs_src") ||
		strings.Contains(p, "conftest")
}

// IsThirdPartyValidationPath returns true if a path or package belongs to the external
// validation corpus (go/chi, go/cobra, py/sqlmodel) rather than first-party Garuda code.
func IsThirdPartyValidationPath(pathOrPkg string) bool {
	p := strings.ToLower(filepath.ToSlash(strings.TrimSpace(pathOrPkg)))
	return strings.Contains(p, "garuda-validation") ||
		strings.HasPrefix(p, "go/chi") ||
		strings.HasPrefix(p, "go/cobra") ||
		strings.HasPrefix(p, "py/sqlmodel") ||
		strings.HasPrefix(p, "sqlmodel") ||
		strings.Contains(p, "/vendor/") ||
		strings.Contains(p, "node_modules")
}

// ExcludeFixturesSQL returns the canonical SQL WHERE clause fragment that filters out non-production
// fixture, benchmark, tutorial, and conftest entities unless includeFixtures is true (B-01, B-07, B-08, B-14).
func ExcludeFixturesSQL(col string, includeFixtures bool) string {
	if includeFixtures {
		return ""
	}
	pkgCol := "package"
	if strings.HasPrefix(col, "e.") {
		pkgCol = "e.package"
	} else if strings.HasPrefix(col, "ef.") {
		pkgCol = "ef.package"
	}
	return fmt.Sprintf(
		" AND COALESCE(%s, '') NOT ILIKE '%%fixture%%' AND COALESCE(%s, '') NOT ILIKE '%%/testdata/%%' AND COALESCE(%s, '') NOT ILIKE '%%active_contradictions%%' AND COALESCE(%s, '') NOT ILIKE '%%garuda-bench%%' AND COALESCE(%s, '') NOT ILIKE '%%docs_src%%' AND COALESCE(%s, '') NOT ILIKE '%%conftest%%' AND COALESCE(%s, '') NOT ILIKE '%%conftest%%'",
		col, col, col, col, col, col, pkgCol,
	)
}

func IncludeFixturesRequested(args map[string]interface{}) bool {
	if v, ok := args["include_fixtures"].(bool); ok {
		return v
	}
	if sc, ok := args["scope"].(string); ok && strings.EqualFold(strings.TrimSpace(sc), "all") {
		return true
	}
	return false
}

func ParseCursorOffset(args map[string]interface{}) int {
	if raw, ok := args["cursor"].(string); ok && strings.TrimSpace(raw) != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil && n >= 0 {
			return n
		}
	}
	if rawNum, ok := args["cursor"].(float64); ok && rawNum >= 0 {
		return int(rawNum)
	}
	if rawOff, ok := args["offset"].(float64); ok && rawOff >= 0 {
		return int(rawOff)
	}
	return 0
}

func ExtractSourceSnippet(filePath string, lineStart int) string {
	clean := sanitizeMCPFilePath(filePath)
	if clean == "" || lineStart <= 0 {
		return ""
	}
	repoRoot := resolveMCPRepoRoot()
	candidates := []string{
		filepath.Join(repoRoot, clean),
		filePath,
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		candidates = append(candidates,
			filepath.Join(home, "garuda", clean),
			filepath.Join(home, "garuda-validation", clean),
		)
	}
	for _, cand := range candidates {
		f, err := os.Open(cand)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		startLn := lineStart - 1
		if startLn < 1 {
			startLn = 1
		}
		endLn := lineStart + 2
		cur := 0
		var lines []string
		for scanner.Scan() {
			cur++
			if cur >= startLn && cur <= endLn {
				lines = append(lines, fmt.Sprintf("%d: %s", cur, strings.TrimRight(scanner.Text(), "\r\n")))
			}
			if cur > endLn {
				break
			}
		}
		_ = f.Close()
		if len(lines) > 0 {
			return strings.Join(lines, "\n")
		}
	}
	return ""
}

// ComputePolicySetVersion uses the Single Policy Registry (`policy.LoadActivePolicies`)
// so policy.list, policy.evaluate, briefing, docs_graph, and snapshot_receipt always
// agree 100% on active and superseded policy IDs (fixes B-03, B-10, B-16).
func ComputePolicySetVersion(ctx context.Context, pool *pgxpool.Pool, tenantID uuid.UUID, policyDir string) (string, []string, []string) {
	if policyDir == "" {
		policyDir = resolveMCPPolicyDir("")
	}
	_, activeIDs, supersededIDs, _ := policy.LoadActivePolicies(ctx, pool, tenantID, policyDir)
	sum := sha256.Sum256([]byte("active:" + strings.Join(activeIDs, ",") + "|superseded:" + strings.Join(supersededIDs, ",")))
	ver := "polset-" + hex.EncodeToString(sum[:])[:12]
	return ver, activeIDs, supersededIDs
}

// ComputeSnapshotReceipt builds the unified provenance receipt attached to tool responses.
// For isolated empty workspaces (like audit-sandbox), it reports an isolated empty-workspace
// snapshot identity instead of leaking the default repo's git commit and dirty hash (fixes B-05).
func ComputeSnapshotReceipt(ctx context.Context, pool *pgxpool.Pool, tenantID, workspaceID uuid.UUID, wsName string) map[string]interface{} {
	polVer, activeIDs, supersededIDs := ComputePolicySetVersion(ctx, pool, tenantID, "")

	var repoCount, entCount int
	if pool != nil && workspaceID != uuid.Nil {
		_ = pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM repositories WHERE workspace_id = $1`, workspaceID).Scan(&repoCount)
		_ = pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM entities WHERE workspace_id = $1 AND kind != 'external'`+ExcludeFixturesSQL("file_path", false), workspaceID).Scan(&entCount)
	}

	// B-05: If a non-default workspace has 0 indexed repositories/entities, return an isolated empty receipt!
	if pool != nil && workspaceID != uuid.Nil && repoCount == 0 && entCount == 0 && !strings.EqualFold(wsName, "default") {
		emptySum := sha256.Sum256([]byte("isolated-empty-workspace:" + wsName + ":" + workspaceID.String() + ":" + polVer))
		emptyHash := hex.EncodeToString(emptySum[:])
		return map[string]interface{}{
			"workspace":                 wsName,
			"workspace_id":              workspaceID.String(),
			"workspace_isolated":        true,
			"indexed_repositories":      0,
			"indexed_entities":          0,
			"commit_sha":                "unindexed-empty-workspace",
			"commit_short":              "unindexed",
			"worktree_dirty":            false,
			"dirty_files_count":         0,
			"dirty_files_hash":          "",
			"graph_fingerprint":         "empty-workspace-" + emptyHash[:12],
			"receipt_hash":              emptyHash,
			"policy_set_version":        polVer,
			"drift_definition_version":  "v2.0-bidirectional-ast-doc",
			"scope_filter":              "production_only (include_fixtures=false)",
			"active_policies":           activeIDs,
			"superseded_policies":       supersededIDs,
			"active_policies_count":     len(activeIDs),
			"superseded_policies_count": len(supersededIDs),
			"head_matches_index":        true,
			"reproduce_command":         fmt.Sprintf("garuda check --workspace %s", wsName),
		}
	}

	repoDir := resolveMCPRepoRoot()
	activeCtx := analyzer.ResolveBuildContext(repoDir, "go", wsName, nil)
	var recordedCommit string
	var recordedDirty bool
	if pool != nil && workspaceID != uuid.Nil {
		if rec := store.GetLatestRecordedBuildContext(ctx, pool, tenantID, workspaceID); rec != nil {
			recordedCommit = rec.CommitShort
			recordedDirty = rec.WorktreeDirty
			if rec.GraphFingerprint != "" {
				activeCtx.GraphFingerprint = rec.GraphFingerprint
				activeCtx.ReceiptHash = activeCtx.ComputeReceiptHash()
			}
		}
	}
	headMatchesIndex := recordedCommit == "" || (recordedCommit == activeCtx.CommitShort && recordedDirty == activeCtx.WorktreeDirty)

	return map[string]interface{}{
		"workspace":                 wsName,
		"workspace_isolated":        true,
		"indexed_repositories":      repoCount,
		"indexed_entities":          entCount,
		"commit_sha":                activeCtx.CommitSHA,
		"commit_short":              activeCtx.CommitShort,
		"worktree_dirty":            activeCtx.WorktreeDirty,
		"dirty_files_count":         activeCtx.DirtyFilesCount,
		"dirty_files_hash":          activeCtx.DirtyFilesHash,
		"graph_fingerprint":         activeCtx.GraphFingerprint,
		"receipt_hash":              activeCtx.ReceiptHash,
		"policy_set_version":        polVer,
		"drift_definition_version":  "v2.0-bidirectional-ast-doc",
		"scope_filter":              "production_only (include_fixtures=false)",
		"active_policies":           activeIDs,
		"superseded_policies":       supersededIDs,
		"active_policies_count":     len(activeIDs),
		"superseded_policies_count": len(supersededIDs),
		"head_matches_index":        headMatchesIndex,
		"reproduce_command":         activeCtx.ReproduceCommand,
	}
}

// WorkspaceCountsSummary is the single source of truth for workspace entity,
// relationship, epistemic claim, and drift counts across briefing, entities,
// governance.status, hygiene_report, summary, and check_drift (fixes B-07 & B-17).
type WorkspaceCountsSummary struct {
	Repositories                 int `json:"repositories"`
	Entities                     int `json:"entities"`
	ExportedEntities             int `json:"exported_entities"`
	ExternalStubs                int `json:"external_stubs"`
	ASTRelationships             int `json:"ast_relationships"`
	DocumentClaims               int `json:"document_claims"`
	SupportedClaims              int `json:"supported_claims"`
	UnverifiedClaims             int `json:"unverified_claims"`
	ContradictedClaims           int `json:"contradicted_claims"`
	OpenContradictions           int `json:"open_contradictions"`
	DocToCodeDrift               int `json:"doc_to_code_drift"`
	CodeToDocDrift               int `json:"code_to_doc_drift"`
	TotalDriftFindings           int `json:"total_drift_findings"`
	UndocumentedExportedEntities int `json:"undocumented_exported_entities"`
	UndocumentedDriftFindings    int `json:"undocumented_drift_findings"`
}

// LoadWorkspaceCounts queries the canonical production-scoped workspace entity, AST edge,
// document claim, and bidirectional drift counts using the exact same SQL filter as handleEntities (fixes B-07).
func LoadWorkspaceCounts(ctx context.Context, pool *pgxpool.Pool, tenantID, workspaceID uuid.UUID) WorkspaceCountsSummary {
	var out WorkspaceCountsSummary
	if pool == nil || workspaceID == uuid.Nil {
		return out
	}
	_ = pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM repositories WHERE workspace_id = $1`, workspaceID).Scan(&out.Repositories)
	_ = pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE kind != 'external'`+ExcludeFixturesSQL("file_path", false)+`)::int,
			COUNT(*) FILTER (WHERE kind != 'external' AND is_exported = true`+ExcludeFixturesSQL("file_path", false)+`)::int,
			COUNT(*) FILTER (WHERE kind = 'external')::int
		FROM entities
		WHERE workspace_id = $1
	`, workspaceID).Scan(&out.Entities, &out.ExportedEntities, &out.ExternalStubs)
	_ = pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM claims WHERE workspace_id = $1`, workspaceID).Scan(&out.ASTRelationships)

	eval := knowledge.NewEvaluator(pool)
	stats, codeToDoc, err := eval.EvaluateWorkspace(ctx, tenantID, workspaceID)
	if err == nil && stats != nil {
		out.CodeToDocDrift = len(codeToDoc)
		out.UndocumentedExportedEntities = stats.UndocumentedExportedEntities
	}

	_ = pool.QueryRow(ctx, `
		SELECT
			COUNT(*)::int,
			COUNT(*) FILTER (WHERE UPPER(COALESCE(status, 'UNVERIFIED')) = 'SUPPORTED')::int,
			COUNT(*) FILTER (WHERE UPPER(COALESCE(status, 'UNVERIFIED')) = 'CONTRADICTED')::int,
			COUNT(*) FILTER (WHERE UPPER(COALESCE(status, 'UNVERIFIED')) NOT IN ('SUPPORTED', 'CONTRADICTED'))::int
		FROM document_claims
		WHERE workspace_id = $1
	`, workspaceID).Scan(&out.DocumentClaims, &out.SupportedClaims, &out.ContradictedClaims, &out.UnverifiedClaims)

	var runtimeContradictions int
	_ = pool.QueryRow(ctx, `
		SELECT COALESCE(COUNT(*)::int, 0)
		FROM claim_verifications
		WHERE workspace_id = $1 AND status = 'CONTRADICTED'
	`, workspaceID).Scan(&runtimeContradictions)
	out.OpenContradictions = out.ContradictedClaims + runtimeContradictions

	out.DocToCodeDrift = out.UnverifiedClaims + out.ContradictedClaims
	out.TotalDriftFindings = out.DocToCodeDrift + out.CodeToDocDrift
	out.UndocumentedDriftFindings = out.TotalDriftFindings
	if out.UndocumentedExportedEntities == 0 && out.ExportedEntities > out.SupportedClaims {
		out.UndocumentedExportedEntities = out.ExportedEntities - out.SupportedClaims
	}
	return out
}

type CanonicalRelRow struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Type      string `json:"type"`
	File      string `json:"file,omitempty"`
	Line      int    `json:"line,omitempty"`
	Snippet   string `json:"content_snippet,omitempty"`
	Direction string `json:"direction,omitempty"`
}

func QueryCanonicalEntityEdges(ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, entityIDStr, entityName string) (incoming []CanonicalRelRow, outgoing []CanonicalRelRow) {
	incoming = make([]CanonicalRelRow, 0)
	outgoing = make([]CanonicalRelRow, 0)
	if pool == nil || workspaceID == uuid.Nil {
		return incoming, outgoing
	}
	entUUID, err := uuid.Parse(strings.TrimSpace(entityIDStr))
	if err != nil {
		return incoming, outgoing
	}

	inRows, inErr := pool.Query(ctx, `
		SELECT COALESCE(e.name, c.from_entity_id::text), c.claim_type,
		       COALESCE(e.file_path, ''), COALESCE(e.line_start, 0)
		FROM claims c
		JOIN entities e ON e.id = c.from_entity_id
		WHERE c.workspace_id = $1 AND c.to_entity_id = $2 AND e.kind != 'external'
	`+ExcludeFixturesSQL("e.file_path", false)+`
		ORDER BY c.claim_type ASC, e.name ASC
	`, workspaceID, entUUID)
	if inErr == nil {
		for inRows.Next() {
			var fromName, ctype, fpath string
			var ln int
			if inRows.Scan(&fromName, &ctype, &fpath, &ln) == nil {
				cleanF := sanitizeMCPFilePath(fpath)
				incoming = append(incoming, CanonicalRelRow{
					From:      fromName,
					To:        entityName,
					Type:      ctype,
					File:      cleanF,
					Line:      ln,
					Snippet:   ExtractSourceSnippet(cleanF, ln),
					Direction: "incoming",
				})
			}
		}
		inRows.Close()
	}

	outRows, outErr := pool.Query(ctx, `
		SELECT COALESCE(e.name, c.to_entity_id::text), c.claim_type,
		       COALESCE(e.file_path, ''), COALESCE(e.line_start, 0)
		FROM claims c
		LEFT JOIN entities e ON e.id = c.to_entity_id
		WHERE c.workspace_id = $1 AND c.from_entity_id = $2
		ORDER BY c.claim_type ASC, e.name ASC
	`, workspaceID, entUUID)
	if outErr == nil {
		for outRows.Next() {
			var toName, ctype, fpath string
			var ln int
			if outRows.Scan(&toName, &ctype, &fpath, &ln) == nil {
				cleanF := sanitizeMCPFilePath(fpath)
				outgoing = append(outgoing, CanonicalRelRow{
					From:      entityName,
					To:        toName,
					Type:      ctype,
					File:      cleanF,
					Line:      ln,
					Snippet:   ExtractSourceSnippet(cleanF, ln),
					Direction: "outgoing",
				})
			}
		}
		outRows.Close()
	}

	return incoming, outgoing
}

// SanitizeGlobalJSONString strips host-specific home directory prefixes from
// serialized JSON payloads so no MCP response ever leaks local filesystem paths.
func SanitizeGlobalJSONString(raw string) string {
	if raw == "" {
		return ""
	}
	out := raw
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		slashHome := filepath.ToSlash(home)
		for _, sub := range []string{"/garuda-validation/", "/garuda/"} {
			out = strings.ReplaceAll(out, slashHome+sub, "")
		}
		out = strings.ReplaceAll(out, slashHome+"/", "")
	}
	for _, prefix := range []string{"/home/rohit/garuda-validation/", "/home/rohit/garuda/", "/home/rohit/", "/tmp/garuda/"} {
		out = strings.ReplaceAll(out, prefix, "")
	}
	return out
}
