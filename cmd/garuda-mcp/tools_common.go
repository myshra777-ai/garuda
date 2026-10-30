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
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/knowledge"
	"github.com/myshra777-ai/garuda/internal/policy"
	"github.com/myshra777-ai/garuda/internal/store"
)

// ErrApprovalRequiresHuman is returned when an MCP tool attempts to record a
// human approval without cryptographic or out-of-band human attestation via
// GARUDA_MCP_APPROVER.
var ErrApprovalRequiresHuman = errors.New("human approval requires GARUDA_MCP_APPROVER to be set in the server environment or CLI co-signing via 'garuda audit ... --decide' / 'garuda rulebook approve'")

// ResolveApprover verifies whether a human approver credential is configured
// in the server environment (GARUDA_MCP_APPROVER). Unauthenticated AI agents
// cannot self-certify as "Human Stakeholder" (CLASS-03 / BUG-01).
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

// ReadOnlyMode reports whether GARUDA_MCP_READ_ONLY is enabled in the server
// environment (BUG-02). When true, all mutating MCP tools are rejected before
// any database or filesystem state is modified.
func ReadOnlyMode() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("GARUDA_MCP_READ_ONLY"))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// ReadOnlyRejection returns the canonical structured rejection payload when a
// mutating tool is invoked while GARUDA_MCP_READ_ONLY=true.
func ReadOnlyRejection(tool string) map[string]any {
	return map[string]any{
		"status":    "rejected",
		"read_only": true,
		"tool":      tool,
		"reason":    "GARUDA_MCP_READ_ONLY is enabled; mutating MCP tools are disabled in this session",
	}
}

// IsMutatingTool returns true if the given MCP tool mutates database or
// filesystem state. For garuda.talon_verify_edit, it is mutating only when
// save_to_ledger or save_ledger is explicitly true.
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

// IsFixturePath returns true if a file path or package belongs to benchmark or test fixtures.
func IsFixturePath(pathOrPkg string) bool {
	p := strings.ToLower(filepath.ToSlash(strings.TrimSpace(pathOrPkg)))
	return strings.Contains(p, "test/benchmark/truth_fixtures") ||
		strings.Contains(p, "test/fixtures/") ||
		strings.Contains(p, "/testdata/")
}

// IncludeFixturesRequested returns true only when include_fixtures=true is explicitly passed.
func IncludeFixturesRequested(args map[string]interface{}) bool {
	if v, ok := args["include_fixtures"].(bool); ok {
		return v
	}
	return false
}

// ParseCursorOffset extracts an integer pagination offset from `cursor` or `offset` arguments.
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

// ExtractSourceSnippet reads 1..5 lines of source code around lineStart from the repository.
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

// ComputePolicySetVersion computes a deterministic version hash and active/superseded lists
// from the Single Policy Registry (YAML policies + PostgreSQL supersession status).
func ComputePolicySetVersion(ctx context.Context, pool *pgxpool.Pool, tenantID uuid.UUID, policyDir string) (string, []string, []string) {
	if policyDir == "" {
		policyDir = resolveMCPPolicyDir("")
	}
	supersededMap := policy.LoadSupersededPolicyIDs(ctx, pool, tenantID)
	activeSet := make(map[string]bool)
	supersededSet := make(map[string]bool)
	for k := range supersededMap {
		supersededSet[k] = true
	}

	if parsed, err := policy.ParseDirectory(policyDir); err == nil {
		for _, pp := range parsed {
			pid := strings.ToLower(strings.TrimSpace(pp.Policy.ID))
			if pid == "" {
				continue
			}
			if supersededMap[pid] {
				supersededSet[pid] = true
			} else {
				activeSet[pid] = true
			}
		}
	}
	if pool != nil && tenantID != uuid.Nil {
		rows, err := pool.Query(ctx, `
			SELECT COALESCE(statement, ''), COALESCE(status, 'active')
			FROM policies
			WHERE tenant_id = $1
			ORDER BY created_at ASC
		`, tenantID)
		if err == nil {
			for rows.Next() {
				var stmt, st string
				if rows.Scan(&stmt, &st) == nil {
					pid := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(stmt, "POLICY:")))
					if pid == "" {
						continue
					}
					if strings.EqualFold(st, "superseded") || supersededMap[pid] {
						supersededSet[pid] = true
						delete(activeSet, pid)
					} else if strings.EqualFold(st, "active") {
						activeSet[pid] = true
					}
				}
			}
			rows.Close()
		}
	}

	activeIDs := make([]string, 0, len(activeSet))
	for k := range activeSet {
		activeIDs = append(activeIDs, k)
	}
	sort.Strings(activeIDs)

	supersededIDs := make([]string, 0, len(supersededSet))
	for k := range supersededSet {
		supersededIDs = append(supersededIDs, k)
	}
	sort.Strings(supersededIDs)

	sum := sha256.Sum256([]byte("active:" + strings.Join(activeIDs, ",") + "|superseded:" + strings.Join(supersededIDs, ",")))
	ver := "polset-" + hex.EncodeToString(sum[:])[:12]
	return ver, activeIDs, supersededIDs
}

// ComputeSnapshotReceipt builds the unified provenance receipt attached to tool responses.
func ComputeSnapshotReceipt(ctx context.Context, pool *pgxpool.Pool, tenantID, workspaceID uuid.UUID, wsName string) map[string]interface{} {
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
	polVer, activeIDs, supersededIDs := ComputePolicySetVersion(ctx, pool, tenantID, "")
	headMatchesIndex := recordedCommit == "" || (recordedCommit == activeCtx.CommitShort && recordedDirty == activeCtx.WorktreeDirty)

	return map[string]interface{}{
		"workspace":                 wsName,
		"commit_sha":                activeCtx.CommitSHA,
		"commit_short":              activeCtx.CommitShort,
		"worktree_dirty":            activeCtx.WorktreeDirty,
		"graph_fingerprint":         activeCtx.GraphFingerprint,
		"receipt_hash":              activeCtx.ReceiptHash,
		"policy_set_version":        polVer,
		"active_policies":           activeIDs,
		"superseded_policies":       supersededIDs,
		"active_policies_count":     len(activeIDs),
		"superseded_policies_count": len(supersededIDs),
		"head_matches_index":        headMatchesIndex,
		"reproduce_command":         activeCtx.ReproduceCommand,
	}
}

// WorkspaceCountsSummary is the single source of truth for workspace entity,
// relationship, epistemic claim, and drift counts across briefing, governance.status,
// hygiene_report, and check_drift (CLASS-02 / BUG-03 / BUG-08).
type WorkspaceCountsSummary struct {
	Repositories       int `json:"repositories"`
	Entities           int `json:"entities"`
	ExportedEntities   int `json:"exported_entities"`
	ExternalStubs      int `json:"external_stubs"`
	ASTRelationships   int `json:"ast_relationships"`
	DocumentClaims     int `json:"document_claims"`
	SupportedClaims    int `json:"supported_claims"`
	UnverifiedClaims   int `json:"unverified_claims"`
	ContradictedClaims int `json:"contradicted_claims"`
	OpenContradictions int `json:"open_contradictions"`
	DocToCodeDrift     int `json:"doc_to_code_drift"`
	CodeToDocDrift     int `json:"code_to_doc_drift"`
	TotalDriftFindings int `json:"total_drift_findings"`
}

// LoadWorkspaceCounts queries the canonical workspace entity, AST edge,
// document claim, and bidirectional drift counts using a single consistent definition.
func LoadWorkspaceCounts(ctx context.Context, pool *pgxpool.Pool, tenantID, workspaceID uuid.UUID) WorkspaceCountsSummary {
	var out WorkspaceCountsSummary
	if pool == nil || workspaceID == uuid.Nil {
		return out
	}
	_ = pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM repositories WHERE workspace_id = $1`, workspaceID).Scan(&out.Repositories)
	_ = pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE kind != 'external')::int,
			COUNT(*) FILTER (WHERE kind != 'external' AND is_exported = true)::int,
			COUNT(*) FILTER (WHERE kind = 'external')::int
		FROM entities
		WHERE workspace_id = $1
	`, workspaceID).Scan(&out.Entities, &out.ExportedEntities, &out.ExternalStubs)
	_ = pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM claims WHERE workspace_id = $1`, workspaceID).Scan(&out.ASTRelationships)
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
	eval := knowledge.NewEvaluator(pool)
	if _, codeToDoc, err := eval.EvaluateWorkspace(ctx, tenantID, workspaceID); err == nil {
		out.CodeToDocDrift = len(codeToDoc)
	}
	out.TotalDriftFindings = out.DocToCodeDrift + out.CodeToDocDrift
	return out
}

// CanonicalRelRow represents a normalized incoming or outgoing relationship edge.
type CanonicalRelRow struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Type      string `json:"type"`
	File      string `json:"file,omitempty"`
	Line      int    `json:"line,omitempty"`
	Snippet   string `json:"content_snippet,omitempty"`
	Direction string `json:"direction,omitempty"`
}

// QueryCanonicalEntityEdges queries incoming and outgoing edges for an entity ID
// using the exact same SQL join as find_entity and neighbors so inbound_edges and
// incoming_count never diverge (fixes Bug 5).
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

	// Incoming edges: claims where to_entity_id = entUUID
	inRows, inErr := pool.Query(ctx, `
		SELECT COALESCE(e.name, c.from_entity_id::text), c.claim_type,
		       COALESCE(e.file_path, ''), COALESCE(e.line_start, 0)
		FROM claims c
		LEFT JOIN entities e ON e.id = c.from_entity_id
		WHERE c.workspace_id = $1 AND c.to_entity_id = $2
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

	// Outgoing edges: claims where from_entity_id = entUUID
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

var homePathLeakRe = regexp.MustCompile(`/home/[^/"\s\\]+/(garuda-validation/|garuda/)?|/tmp/garuda/`)

// SanitizeGlobalJSONString strips developer home directory prefixes from serialized JSON payloads.
func SanitizeGlobalJSONString(s string) string {
	return homePathLeakRe.ReplaceAllString(s, "")
}
