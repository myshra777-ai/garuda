// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/impact"
	"github.com/myshra777-ai/garuda/internal/knowledge"
	"github.com/myshra777-ai/garuda/internal/policy"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/tenant"
)

var defaultTenantID = tenant.CanonicalID

func resolveMCPPolicyDir(explicit string) string {
	candidates := []string{}
	if strings.TrimSpace(explicit) != "" {
		candidates = append(candidates, strings.TrimSpace(explicit))
	}
	if envDir := strings.TrimSpace(os.Getenv("GARUDA_POLICY_DIR")); envDir != "" {
		candidates = append(candidates, envDir)
	}
	repoRoot := resolveMCPRepoRoot()
	if repoRoot != "" {
		candidates = append(candidates, filepath.Join(repoRoot, "policies"))
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "policies"),
			filepath.Join(filepath.Dir(exeDir), "policies"),
		)
	}
	candidates = append(candidates, "./policies", "../policies", "/home/rohit/garuda/policies", "/tmp/garuda/policies")
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			if abs, aErr := filepath.Abs(c); aErr == nil {
				return abs
			}
			return c
		}
	}
	return "./policies"
}

func (s *MCPServer) resolveTenantAndWorkspace(args map[string]interface{}) (uuid.UUID, uuid.UUID, error) {
	tenantID := defaultTenantID

	if raw, ok := args["tenant_id"].(string); ok && strings.TrimSpace(raw) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("invalid tenant_id %q: %w", raw, err)
		}
		tenantID = parsed
	} else if envTid := strings.TrimSpace(os.Getenv("GARUDA_TENANT_ID")); envTid != "" {
		parsed, err := uuid.Parse(envTid)
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("invalid GARUDA_TENANT_ID %q: %w", envTid, err)
		}
		tenantID = parsed
	}

	workspaceName := ""
	if raw, ok := args["workspace"].(string); ok && strings.TrimSpace(raw) != "" {
		workspaceName = strings.TrimSpace(raw)
	} else if envWs := strings.TrimSpace(os.Getenv("GARUDA_WORKSPACE")); envWs != "" {
		workspaceName = envWs
	}

	if s == nil || s.store == nil || s.safePool() == nil {
		return tenantID, uuid.Nil, nil
	}
	workspaceID, err := store.ResolveWorkspaceID(context.Background(), s.safePool(), tenantID, workspaceName)
	if err != nil && strings.EqualFold(workspaceName, "audit-sandbox") {
		sbID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(tenantID.String()+":workspace:audit-sandbox"))
		_, _ = s.safePool().Exec(context.Background(), `
			INSERT INTO tenants (id, name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING
		`, tenantID, "tenant-"+tenantID.String()[:8])
		_ = s.safePool().QueryRow(context.Background(), `
			INSERT INTO workspaces (id, tenant_id, name, description)
			VALUES ($1, $2, 'audit-sandbox', 'Isolated evaluation & write-test sandbox workspace')
			ON CONFLICT (tenant_id, name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, sbID, tenantID).Scan(&sbID)
		if sbID != uuid.Nil {
			return tenantID, sbID, nil
		}
	}
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("resolve workspace %q: %w", workspaceName, err)
	}
	return tenantID, workspaceID, nil
}

// handlePolicyList returns every policy registered in the Single Policy Registry (B-03 & B-16).
// Ensures every policy row carries both its canonical `external_id` and its canonical `title`
// so `policy.list`, `briefing`, `policy.evaluate`, and `docs_graph` never use conflicting labels.
func (s *MCPServer) handlePolicyList(args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()
	tenantID, _, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}

	if s.store == nil || s.safePool() == nil {
		return map[string]interface{}{
			"tenant_id":        tenantID.String(),
			"count":            0,
			"active_count":     0,
			"superseded_count": 0,
			"policies":         []interface{}{},
		}, nil
	}

	policyDir := resolveMCPPolicyDir("")
	activeParsed, activeIDs, supersededIDs, _ := policy.LoadActivePolicies(ctx, s.safePool(), tenantID, policyDir)
	titleByExtID := make(map[string]string, len(activeParsed))
	for _, pp := range activeParsed {
		if pp.Policy != nil {
			titleByExtID[pp.Policy.ID] = pp.Policy.Title
		}
	}

	rows, err := s.safePool().Query(ctx, `
		SELECT id::text, COALESCE(statement, ''), COALESCE(scope_domain, ''),
		       COALESCE(scope_system, ''), COALESCE(actor, ''), COALESCE(status, 'active'),
		       COALESCE(policy_version::text, '1'), created_at
		FROM policies
		WHERE tenant_id = $1
		ORDER BY created_at DESC, id ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query policies: %w", err)
	}
	defer rows.Close()

	type PolicyRow struct {
		ID            string `json:"id"`
		ExternalID    string `json:"external_id"`
		Title         string `json:"title"`
		Statement     string `json:"statement"`
		ScopeDomain   string `json:"scope_domain"`
		ScopeSystem   string `json:"scope_system"`
		Actor         string `json:"actor"`
		Status        string `json:"status"`
		PolicyVersion string `json:"policy_version"`
		CreatedAt     string `json:"created_at"`
	}

	policies := make([]PolicyRow, 0)
	seenActive := make(map[string]bool)
	for rows.Next() {
		var p PolicyRow
		var createdAt interface{}
		if err := rows.Scan(&p.ID, &p.Statement, &p.ScopeDomain,
			&p.ScopeSystem, &p.Actor, &p.Status, &p.PolicyVersion, &createdAt); err != nil {
			continue
		}
		p.CreatedAt = fmt.Sprint(createdAt)
		extID := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(p.Statement, "POLICY:")))
		p.ExternalID = extID
		if t, ok := titleByExtID[extID]; ok && t != "" {
			p.Title = t
		} else {
			p.Title = extID
		}
		if strings.EqualFold(p.Status, "active") {
			seenActive[extID] = true
		}
		policies = append(policies, p)
	}

	// Ensure any active YAML policy not yet persisted to DB is also listed in policies
	for _, pp := range activeParsed {
		if pp.Policy != nil && !seenActive[pp.Policy.ID] {
			policies = append(policies, PolicyRow{
				ID:            uuid.NewSHA1(uuid.NameSpaceOID, []byte(pp.Policy.ID)).String(),
				ExternalID:    pp.Policy.ID,
				Title:         pp.Policy.Title,
				Statement:     "POLICY:" + pp.Policy.ID,
				ScopeDomain:   pp.Policy.Scope.Domain,
				ScopeSystem:   pp.Policy.Scope.System,
				Actor:         pp.Policy.Authority,
				Status:        "active",
				PolicyVersion: pp.Policy.Version,
				CreatedAt:     "registry",
			})
		}
	}

	polVer, _, _ := ComputePolicySetVersion(ctx, s.safePool(), tenantID, policyDir)
	return map[string]interface{}{
		"tenant_id":             tenantID.String(),
		"policy_set_version":    polVer,
		"active_policy_ids":     activeIDs,
		"superseded_policy_ids": supersededIDs,
		"count":                 len(policies),
		"active_count":          len(activeIDs),
		"superseded_count":      len(supersededIDs),
		"policies":              policies,
	}, nil
}

// handleGovernanceStatus returns an aggregate view of governance state (B-06, B-07, B-17).
func (s *MCPServer) handleGovernanceStatus(args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	wsName := s.resolveMCPWorkspaceName(args, workspaceID)

	if s.store == nil || s.safePool() == nil {
		counts := LoadWorkspaceCounts(ctx, nil, tenantID, workspaceID)
		verdict := computeVerdict(&knowledge.HealthStats{})
		return map[string]interface{}{
			"tenant_id":                tenantID.String(),
			"workspace":                wsName,
			"workspace_id":             workspaceID.String(),
			"decision":                 verdict.State,
			"agent_should_proceed":     verdict.AgentProceed,
			"conformance_demonstrated": verdict.Conforms,
			"blocked_by":               verdict.BlockedBy,
			"active_policies":          0,
			"entity_count":             counts.Entities,
			"exported_entity_count":    counts.ExportedEntities,
			"external_stubs_count":     counts.ExternalStubs,
			"ast_relationship_count":   counts.ASTRelationships,
			"document_claim_count":     counts.DocumentClaims,
			"epistemic": map[string]int{
				"supported":           0,
				"unverified":          0,
				"contradicted":        0,
				"open_contradictions": 0,
			},
			"documentation_health":   &knowledge.HealthStats{},
			"drift_finding_count":    0,
			"drift_findings_preview": []interface{}{},
			"gate_verdict":           verdict,
		}, nil
	}

	evaluator := knowledge.NewEvaluator(s.safePool())
	stats, drift, err := evaluator.EvaluateWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("evaluate workspace: %w", err)
	}
	for i := range drift {
		drift[i].Path = sanitizeMCPFilePath(drift[i].Path)
	}

	counts := LoadWorkspaceCounts(ctx, s.safePool(), tenantID, workspaceID)
	if stats == nil {
		stats = &knowledge.HealthStats{}
	}
	stats.TotalEntities = counts.Entities
	stats.ExportedEntities = counts.ExportedEntities
	stats.ASTRelationships = counts.ASTRelationships
	stats.TotalClaims = counts.DocumentClaims
	stats.Supported = counts.SupportedClaims
	stats.Unverified = counts.UnverifiedClaims
	stats.Contradicted = counts.ContradictedClaims
	stats.UndocumentedCode = counts.CodeToDocDrift
	stats.UndocumentedExportedEntities = counts.UndocumentedExportedEntities
	stats.UndocumentedDriftFindings = counts.UndocumentedDriftFindings

	blockedBy := make([]string, 0)
	cRows, cErr := s.safePool().Query(ctx, `
		SELECT COALESCE(document_title, 'ADR'), subject, modality, predicate, object, COALESCE(contradiction_reason, '')
		FROM document_claims
		WHERE tenant_id = $1 AND workspace_id = $2 AND status IN ('CONTRADICTED', 'UNVERIFIED')
		ORDER BY status ASC, document_title ASC, subject ASC
	`, tenantID, workspaceID)
	if cErr == nil {
		for cRows.Next() {
			var doc, subj, mod, pred, obj, rsn string
			if cRows.Scan(&doc, &subj, &mod, &pred, &obj, &rsn) == nil {
				blockedBy = append(blockedBy, fmt.Sprintf("[ADR:%s] %s %s %s %s (%s)", doc, subj, mod, pred, obj, rsn))
			}
		}
		cRows.Close()
	}

	verdict := computeVerdict(stats)
	if len(blockedBy) > 0 {
		verdict.BlockedBy = blockedBy
	}

	snapshot := ComputeSnapshotReceipt(ctx, s.safePool(), tenantID, workspaceID, wsName)
	activePolicyCount, _ := snapshot["active_policies_count"].(int)

	return map[string]interface{}{
		"tenant_id":                tenantID.String(),
		"workspace":                wsName,
		"workspace_id":             workspaceID.String(),
		"decision":                 verdict.State,
		"agent_should_proceed":     verdict.AgentProceed,
		"conformance_demonstrated": verdict.Conforms,
		"blocked_by":               verdict.BlockedBy,
		"active_policies":          activePolicyCount,
		"policy_set_version":       snapshot["policy_set_version"],
		"snapshot_receipt":         snapshot,
		"entity_count":             counts.Entities,
		"exported_entity_count":    counts.ExportedEntities,
		"external_stubs_count":     counts.ExternalStubs,
		"ast_relationship_count":   counts.ASTRelationships,
		"document_claim_count":     counts.DocumentClaims,
		"epistemic": map[string]int{
			"supported":           counts.SupportedClaims,
			"unverified":          counts.UnverifiedClaims,
			"contradicted":        counts.ContradictedClaims,
			"open_contradictions": counts.OpenContradictions,
		},
		"documentation_health":     stats,
		"drift_definition_version": "v2.0-bidirectional-ast-doc",
		"drift_finding_count":      counts.TotalDriftFindings,
		"drift_summary": map[string]int{
			"doc_to_code":                    counts.DocToCodeDrift,
			"code_to_doc":                    counts.CodeToDocDrift,
			"total_drift_findings":           counts.TotalDriftFindings,
			"undocumented_exported_entities": counts.UndocumentedExportedEntities,
			"undocumented_drift_findings":    counts.UndocumentedDriftFindings,
		},
		"drift_findings_preview": previewDrift(drift, 10),
		"gate_verdict":           verdict,
	}, nil
}

type governanceVerdict struct {
	State        string   `json:"state"`
	Reason       string   `json:"reason"`
	AgentProceed bool     `json:"agent_should_proceed"`
	Conforms     bool     `json:"conformance_demonstrated"`
	BlockedBy    []string `json:"blocked_by"`
}

// computeVerdict enforces fail-closed governance (B-06):
// agent_should_proceed is FALSE whenever the workspace has 0 entities, 0 claims,
// any CONTRADICTED claims, or any UNVERIFIED claims!
func computeVerdict(stats *knowledge.HealthStats) governanceVerdict {
	switch {
	case stats == nil || stats.TotalEntities == 0:
		return governanceVerdict{
			State:        "UNVERIFIED",
			Reason:       "Workspace contains 0 indexed production entities; conformance cannot be demonstrated.",
			AgentProceed: false,
			Conforms:     false,
			BlockedBy:    []string{"workspace has 0 indexed production entities"},
		}
	case stats.Contradicted > 0:
		return governanceVerdict{
			State: "BLOCKED",
			Reason: fmt.Sprintf(
				"%d documentation claims exist; %d contradicted by the codebase. Resolve contradictions before proceeding.",
				stats.TotalClaims, stats.Contradicted),
			AgentProceed: false,
			Conforms:     false,
			BlockedBy:    []string{fmt.Sprintf("%d contradicted documentation claims", stats.Contradicted)},
		}
	case stats.TotalClaims == 0:
		return governanceVerdict{
			State:        "UNVERIFIED",
			Reason:       "No documentation claims have been ingested for this workspace. Run 'garuda docs ingest' to enable conformance checks.",
			AgentProceed: false,
			Conforms:     false,
			BlockedBy:    []string{"0 documentation claims ingested"},
		}
	case stats.Supported == 0:
		return governanceVerdict{
			State: "UNVERIFIED",
			Reason: fmt.Sprintf(
				"%d documentation claims exist but none have been verified against the codebase. Conformance has not been demonstrated.",
				stats.TotalClaims),
			AgentProceed: false,
			Conforms:     false,
			BlockedBy:    []string{fmt.Sprintf("all %d documentation claims are UNVERIFIED", stats.TotalClaims)},
		}
	case stats.Supported < stats.TotalClaims:
		return governanceVerdict{
			State: "PARTIAL",
			Reason: fmt.Sprintf(
				"%d of %d documentation claims verified against the codebase (%d remain UNVERIFIED). Human review required before proceeding.",
				stats.Supported, stats.TotalClaims, stats.TotalClaims-stats.Supported),
			AgentProceed: false,
			Conforms:     false,
			BlockedBy:    []string{fmt.Sprintf("%d of %d documentation claims are UNVERIFIED", stats.TotalClaims-stats.Supported, stats.TotalClaims)},
		}
	default:
		return governanceVerdict{
			State: "CONFORMS",
			Reason: fmt.Sprintf(
				"All %d documentation claims verified against the codebase. No contradictions detected.",
				stats.TotalClaims),
			AgentProceed: true,
			Conforms:     true,
			BlockedBy:    []string{},
		}
	}
}

// handleCheckDrift returns the bidirectional documentation-to-code (doc_to_code)
// and code-to-documentation (code_to_doc) drift report with explicit truncation metadata (B-09, B-12, B-17).
func (s *MCPServer) handleCheckDrift(args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	wsName := s.resolveMCPWorkspaceName(args, workspaceID)

	evaluator := knowledge.NewEvaluator(s.safePool())
	stats, codeToDoc, err := evaluator.EvaluateWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("evaluate workspace: %w", err)
	}
	for i := range codeToDoc {
		codeToDoc[i].Path = sanitizeMCPFilePath(codeToDoc[i].Path)
	}
	if codeToDoc == nil {
		codeToDoc = []knowledge.CodeDriftFinding{}
	}

	counts := LoadWorkspaceCounts(ctx, s.safePool(), tenantID, workspaceID)
	if stats == nil {
		stats = &knowledge.HealthStats{}
	}
	stats.TotalEntities = counts.Entities
	stats.ExportedEntities = counts.ExportedEntities
	stats.ASTRelationships = counts.ASTRelationships
	stats.TotalClaims = counts.DocumentClaims
	stats.Supported = counts.SupportedClaims
	stats.Unverified = counts.UnverifiedClaims
	stats.Contradicted = counts.ContradictedClaims
	stats.UndocumentedExportedEntities = counts.UndocumentedExportedEntities
	stats.UndocumentedDriftFindings = counts.UndocumentedDriftFindings

	type docDriftRow struct {
		Document  string `json:"document"`
		Section   string `json:"section"`
		Subject   string `json:"subject"`
		Modality  string `json:"modality"`
		Predicate string `json:"predicate"`
		Object    string `json:"object"`
		Status    string `json:"status"`
		Reason    string `json:"reason"`
	}
	docToCode := make([]docDriftRow, 0)
	dRows, dErr := s.safePool().Query(ctx, `
		SELECT COALESCE(document_title, 'ADR'), COALESCE(section_title, ''),
		       subject, modality, predicate, object,
		       COALESCE(status, 'UNVERIFIED'), COALESCE(contradiction_reason, '')
		FROM document_claims
		WHERE tenant_id = $1 AND workspace_id = $2
		  AND status IN ('UNVERIFIED', 'CONTRADICTED')
		ORDER BY line_start ASC, subject ASC
	`, tenantID, workspaceID)
	if dErr == nil {
		for dRows.Next() {
			var r docDriftRow
			if dRows.Scan(&r.Document, &r.Section, &r.Subject, &r.Modality, &r.Predicate, &r.Object, &r.Status, &r.Reason) == nil {
				if strings.TrimSpace(r.Reason) == "" {
					if strings.EqualFold(r.Predicate, "IDEMPOTENT") {
						r.Reason = "AWAITING_RUNTIME_EVIDENCE: Predicate IDEMPOTENT requires runtime execution trace"
					} else {
						r.Reason = "STATIC_AST_UNRESOLVED: Required static relationship not found in workspace AST"
					}
				}
				docToCode = append(docToCode, r)
			}
		}
		dRows.Close()
	}

	verdict := computeVerdict(stats)
	isTruncated := counts.UndocumentedExportedEntities > len(codeToDoc)

	return map[string]interface{}{
		"tenant_id":                           tenantID.String(),
		"workspace":                           wsName,
		"workspace_id":                        workspaceID.String(),
		"drift_definition_version":            "v2.0-bidirectional-ast-doc",
		"entity_count":                        counts.Entities,
		"exported_entity_count":               counts.ExportedEntities,
		"documentation_health":                stats,
		"doc_to_code":                         docToCode,
		"code_to_doc":                         codeToDoc,
		"doc_to_code_count":                   len(docToCode),
		"code_to_doc_count":                   len(codeToDoc),
		"undocumented_exported_entities":      counts.UndocumentedExportedEntities,
		"undocumented_drift_findings":         len(docToCode) + len(codeToDoc),
		"undocumented_symbols":                codeToDoc,
		"undocumented_symbols_returned_count": len(codeToDoc),
		"undocumented_symbols_total_count":    counts.UndocumentedExportedEntities,
		"undocumented_symbols_truncated":      isTruncated,
		"truncated":                           isTruncated,
		"drift_breakdown": map[string]int{
			"doc_to_code":                    len(docToCode),
			"code_to_doc":                    len(codeToDoc),
			"total_drift":                    len(docToCode) + len(codeToDoc),
			"undocumented_exported_entities": counts.UndocumentedExportedEntities,
			"undocumented_drift_findings":    len(docToCode) + len(codeToDoc),
		},
		"gate_verdict": verdict,
		"status":       "success",
	}, nil
}

// handleQueryClaims returns document claims matching an optional subject and status filter (BUG-07 & B-09).
func (s *MCPServer) handleQueryClaims(args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	wsName := s.resolveMCPWorkspaceName(args, workspaceID)

	if s.store != nil && s.safePool() != nil && workspaceID != uuid.Nil {
		eval := knowledge.NewEvaluator(s.safePool())
		_, _, _ = eval.EvaluateWorkspace(ctx, tenantID, workspaceID)
	}

	subject, _ := args["subject"].(string)
	subject = strings.TrimSpace(subject)
	statusFilter, _ := args["status"].(string)
	statusFilter = strings.ToUpper(strings.TrimSpace(statusFilter))

	query := `
		SELECT COALESCE(document_title, 'ADR'), COALESCE(section_title, ''),
		       subject, modality, predicate, object,
		       COALESCE(status, 'UNVERIFIED'), COALESCE(contradiction_reason, '')
		FROM document_claims
		WHERE tenant_id = $1 AND workspace_id = $2
	`
	dbArgs := []interface{}{tenantID, workspaceID}
	argPos := 3

	if subject != "" {
		query += fmt.Sprintf(" AND (subject ILIKE $%d OR object ILIKE $%d)", argPos, argPos)
		dbArgs = append(dbArgs, "%"+subject+"%")
		argPos++
	}
	if statusFilter != "" && statusFilter != "ALL" {
		query += fmt.Sprintf(" AND UPPER(COALESCE(status, 'UNVERIFIED')) = $%d", argPos)
		dbArgs = append(dbArgs, statusFilter)
		argPos++
	}
	query += " ORDER BY line_start ASC, subject ASC"

	rows, err := s.safePool().Query(ctx, query, dbArgs...)
	if err != nil {
		return nil, fmt.Errorf("query claims: %w", err)
	}
	defer rows.Close()

	type ClaimView struct {
		Document      string `json:"document"`
		Section       string `json:"section"`
		Subject       string `json:"subject"`
		Modality      string `json:"modality"`
		Predicate     string `json:"predicate"`
		Object        string `json:"object"`
		Status        string `json:"status"`
		StatusReason  string `json:"status_reason,omitempty"`
		Contradiction string `json:"contradiction_reason,omitempty"`
	}

	results := make([]ClaimView, 0)
	for rows.Next() {
		var v ClaimView
		if err := rows.Scan(&v.Document, &v.Section, &v.Subject, &v.Modality,
			&v.Predicate, &v.Object, &v.Status, &v.Contradiction); err == nil {
			if strings.EqualFold(v.Status, "SUPPORTED") {
				if v.Contradiction != "" {
					v.StatusReason = v.Contradiction
				} else {
					v.StatusReason = "VERIFIED_STATIC_AST"
				}
				v.Contradiction = ""
			} else if strings.EqualFold(v.Status, "UNVERIFIED") {
				if v.Contradiction != "" {
					v.StatusReason = v.Contradiction
				} else if strings.EqualFold(v.Predicate, "IDEMPOTENT") {
					v.StatusReason = "AWAITING_RUNTIME_EVIDENCE: Predicate IDEMPOTENT requires runtime execution trace"
				} else {
					v.StatusReason = "STATIC_AST_UNRESOLVED: Symbol or relationship not found in workspace production AST"
				}
				v.Contradiction = ""
			} else {
				v.StatusReason = v.Contradiction
			}
			results = append(results, v)
		}
	}

	return map[string]interface{}{
		"tenant_id":     tenantID.String(),
		"workspace":     wsName,
		"workspace_id":  workspaceID.String(),
		"subject":       subject,
		"status_filter": statusFilter,
		"count":         len(results),
		"claims":        results,
	}, nil
}

func previewDrift(findings []knowledge.CodeDriftFinding, n int) []knowledge.CodeDriftFinding {
	if findings == nil {
		return []knowledge.CodeDriftFinding{}
	}
	if len(findings) <= n {
		return findings
	}
	return findings[:n]
}

// handlePolicyEvaluate runs the policy engine against a workspace in dry-run mode (B-03, B-04, B-16, BUG-19).
// Evaluates every active policy from the Single Policy Registry (`len(evaluations) == len(active_policies)`),
// prevents vacuous ALLOW when 0 entities match scope (`INCONCLUSIVE`), and reports `scoped_entity_count`
// as `entity_count` when in-scope entities pass.
func (s *MCPServer) handlePolicyEvaluate(args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}

	workspaceName := s.resolveMCPWorkspaceName(args, workspaceID)
	rawPolicyDir, _ := args["policy_dir"].(string)
	policyDir := resolveMCPPolicyDir(rawPolicyDir)

	subjectKind, _ := args["subject_kind"].(string)
	subjectID, _ := args["subject_id"].(string)
	if subjectID == "" {
		subjectID, _ = args["target"].(string)
	}
	actor, _ := args["actor"].(string)
	if actor == "" {
		actor = "mcp-policy-preview"
	}
	includeFixtures := IncludeFixturesRequested(args)

	engine := policy.NewEngine(s.safePool())
	result, err := engine.Run(
		ctx,
		tenantID, workspaceID,
		policyDir,
		subjectKind, subjectID, actor,
		policy.RunOptions{
			DryRun:          true,
			IncludeFixtures: includeFixtures,
		},
	)
	if err != nil {
		return nil, err
	}

	evaluations := make([]map[string]interface{}, 0, len(result.Evaluations))
	for _, ev := range result.Evaluations {
		preds := append([]string(nil), ev.Evidence.MatchedPredicates...)
		sort.Strings(preds)
		entCount := ev.ScopedEntityCount
		if len(ev.Evidence.EntityIDs) > entCount {
			entCount = len(ev.Evidence.EntityIDs)
		}
		evaluations = append(evaluations, map[string]interface{}{
			"policy_id":              ev.PolicyID.String(),
			"policy_external_id":     ev.PolicyExternalID,
			"title":                  ev.PolicyTitle,
			"decision":               string(ev.Decision),
			"reason":                 ev.Reason,
			"matched_predicates":     preds,
			"entity_count":           entCount,
			"scoped_entity_count":    ev.ScopedEntityCount,
			"violating_entity_count": len(ev.Evidence.EntityIDs),
			"claim_count":            len(ev.Evidence.ClaimIDs),
			"contradiction_count":    len(ev.Evidence.ContradictionIDs),
			"reasoning_notes":        ev.Evidence.ReasoningNotes,
			"anchor":                 nil,
		})
	}
	sort.Slice(evaluations, func(i, j int) bool {
		a, _ := evaluations[i]["policy_external_id"].(string)
		b, _ := evaluations[j]["policy_external_id"].(string)
		if a != b {
			return a < b
		}
		ri, _ := evaluations[i]["reason"].(string)
		rj, _ := evaluations[j]["reason"].(string)
		return ri < rj
	})

	snapshot := ComputeSnapshotReceipt(ctx, s.safePool(), tenantID, workspaceID, workspaceName)
	return map[string]interface{}{
		"tenant_id":             tenantID.String(),
		"workspace":             workspaceName,
		"workspace_id":          workspaceID.String(),
		"policy_dir":            "policies",
		"policy_set_version":    snapshot["policy_set_version"],
		"active_policies_count": len(evaluations),
		"evaluations_count":     len(evaluations),
		"snapshot_receipt":      snapshot,
		"dry_run":               true,
		"final_decision":        string(result.FinalDecision),
		"evaluations":           evaluations,
		"summary":               result.Summary,
		"note": "Dry-run only. Every active policy in the Single Policy Registry was evaluated. " +
			"Run 'garuda policy evaluate <dir> --workspace " + workspaceName +
			"' from the CLI to record and anchor these decisions.",
	}, nil
}

func (s *MCPServer) handleVerifyPolicyEvaluation(args map[string]interface{}) (interface{}, error) {
	tenantID, _, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}

	rawID, _ := args["evaluation_id"].(string)
	if rawID == "" {
		rawID, _ = args["evidence_id"].(string)
	}
	if rawID == "" {
		rawID, _ = args["id"].(string)
	}
	if rawID == "" {
		return nil, fmt.Errorf("evaluation_id is required")
	}
	evalID, err := uuid.Parse(rawID)
	if err != nil {
		return map[string]interface{}{
			"evaluation_id": rawID,
			"valid":         false,
			"reason":        fmt.Sprintf("invalid UUID format %q", rawID),
		}, nil
	}

	ev, err := policy.LoadEvaluation(context.Background(), s.safePool(), tenantID, evalID)
	if err != nil {
		return map[string]interface{}{
			"evaluation_id": evalID.String(),
			"valid":         false,
			"reason":        "evaluation not found for tenant",
		}, nil
	}

	anchor := policy.NewAnchor(s.safePool())
	valid, verifyErr := anchor.VerifyEvaluation(context.Background(), ev)

	out := map[string]interface{}{
		"evaluation_id": ev.ID.String(),
		"policy_id":     ev.PolicyID.String(),
		"decision":      string(ev.Decision),
		"subject_kind":  ev.SubjectKind,
		"subject_id":    ev.SubjectID,
		"valid":         valid,
	}
	if ev.MerkleBlockHeight != nil {
		out["block_height"] = *ev.MerkleBlockHeight
	}
	if verifyErr != nil {
		out["valid"] = false
		out["reason"] = verifyErr.Error()
	} else if !valid {
		out["reason"] = "Merkle inclusion proof did not verify"
	}
	return out, nil
}

func (s *MCPServer) handleVerifyEvidence(args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()
	tenantID, _, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}

	rawID, _ := args["evidence_id"].(string)
	if rawID == "" {
		rawID, _ = args["evaluation_id"].(string)
	}
	if rawID == "" {
		rawID, _ = args["decision_id"].(string)
	}
	if rawID == "" {
		rawID, _ = args["id"].(string)
	}
	if rawID == "" {
		return map[string]interface{}{
			"valid":  false,
			"reason": "one of evidence_id, evaluation_id, or decision_id is required",
		}, nil
	}

	parsedID, parseErr := uuid.Parse(rawID)
	if parseErr != nil {
		return map[string]interface{}{
			"evidence_id": rawID,
			"valid":       false,
			"reason":      "evidence not found for tenant (non-UUID identifier)",
		}, nil
	}

	if ev, err := policy.LoadEvaluation(ctx, s.safePool(), tenantID, parsedID); err == nil && ev != nil {
		anchor := policy.NewAnchor(s.safePool())
		valid, verifyErr := anchor.VerifyEvaluation(ctx, ev)
		out := map[string]interface{}{
			"evidence_id":   ev.ID.String(),
			"evaluation_id": ev.ID.String(),
			"kind":          "policy_evaluation",
			"policy_id":     ev.PolicyID.String(),
			"decision":      string(ev.Decision),
			"valid":         valid,
		}
		if ev.MerkleBlockHeight != nil {
			out["block_height"] = *ev.MerkleBlockHeight
		}
		if verifyErr != nil {
			out["valid"] = false
			out["reason"] = verifyErr.Error()
		} else if !valid {
			out["reason"] = "Merkle inclusion proof did not verify"
		}
		return out, nil
	}

	var evHash string
	var decID uuid.UUID
	err = s.safePool().QueryRow(ctx, `
		SELECT decision_id, COALESCE(evidence_hash, '')
		FROM evidence_blocks
		WHERE tenant_id = $1 AND (id = $2 OR decision_id = $2)
		ORDER BY created_at DESC LIMIT 1
	`, tenantID, parsedID).Scan(&decID, &evHash)
	if err == nil && evHash != "" {
		return map[string]interface{}{
			"evidence_id":   parsedID.String(),
			"decision_id":   decID.String(),
			"kind":          "evidence_block",
			"evidence_hash": evHash,
			"valid":         true,
		}, nil
	}

	return map[string]interface{}{
		"evidence_id": parsedID.String(),
		"valid":       false,
		"reason":      "evaluation or evidence not found for tenant",
	}, nil
}

func (s *MCPServer) handleBlastRadius(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	includeFixtures := IncludeFixturesRequested(args)

	entityID, _ := args["entity_id"].(string)
	nameStr, _ := args["name"].(string)
	if nameStr == "" {
		nameStr, _ = args["symbol"].(string)
	}
	if entityID != "" {
		if _, pErr := uuid.Parse(entityID); pErr != nil {
			nameStr = entityID
			entityID = ""
		}
	}
	if entityID == "" && nameStr != "" {
		resolvedID, _, err := s.resolveSubjectEntityWithOpts(context.Background(), workspaceID, "", nameStr, includeFixtures)
		if err != nil {
			return map[string]interface{}{
				"target_symbol":  nameStr,
				"target_found":   false,
				"reason":         err.Error(),
				"total_affected": 0,
			}, nil
		}
		entityID = resolvedID.String()
	}
	if entityID == "" {
		return nil, fmt.Errorf("one of entity_id or name is required")
	}

	depth := 3
	if v, ok := args["depth"].(float64); ok && v > 0 {
		depth = int(v)
	}
	minConf := 0.50
	if v, ok := args["min_confidence"].(float64); ok && v > 0 {
		minConf = v
	}

	idx, err := s.store.BuildImpactIndex(context.Background(), workspaceID)
	if err != nil {
		return nil, fmt.Errorf("build impact index: %w", err)
	}

	if _, ok := idx.GetEntityMeta(entityID); !ok {
		return map[string]interface{}{
			"target_entity_id": entityID,
			"target_found":     false,
			"reason":           "entity not found in workspace",
			"total_affected":   0,
		}, nil
	}

	cfg := impact.BlastRadiusConfig{
		MaxDepth:        depth,
		MinConfidence:   minConf,
		IncludeInferred: true,
	}
	result := impact.ComputeBlastRadius(idx, entityID, cfg)

	wsName := s.resolveMCPWorkspaceName(args, workspaceID)
	snapshot := ComputeSnapshotReceipt(context.Background(), s.safePool(), tenantID, workspaceID, wsName)
	canonCommit, _ := snapshot["commit_short"].(string)

	fileMap := make(map[string]string)
	lineMap := make(map[string]int)
	if s.store != nil && s.safePool() != nil {
		mRows, mErr := s.safePool().Query(context.Background(), `
			SELECT id::text, COALESCE(file_path, ''), COALESCE(line_start, 0)
			FROM entities WHERE workspace_id = $1
		`, workspaceID)
		if mErr == nil {
			for mRows.Next() {
				var idStr, fp string
				var ln int
				if mRows.Scan(&idStr, &fp, &ln) == nil {
					fileMap[idStr] = sanitizeMCPFilePath(fp)
					lineMap[idStr] = ln
				}
			}
			mRows.Close()
		}
	}

	all := make([]map[string]interface{}, 0, result.TotalAffected)
	bySev := map[string]int{"CRITICAL": 0, "HIGH": 0, "MEDIUM": 0, "LOW": 0}
	fixturesExcluded := 0

	for _, sev := range []impact.SeverityLevel{
		impact.SeverityCritical,
		impact.SeverityHigh,
		impact.SeverityMedium,
		impact.SeverityLow,
	} {
		list := result.BySeverity[sev]
		impact.SortImpactedEntities(list)
		for _, item := range list {
			fp := sanitizeMCPFilePath(item.FilePath)
			if fp == "" {
				fp = fileMap[item.EntityID]
			}
			ln := lineMap[item.EntityID]
			if !includeFixtures && (IsFixturePath(fp) || IsFixturePath(item.Package)) {
				fixturesExcluded++
				continue
			}
			bySev[string(sev)]++
			snippet := ExtractSourceSnippet(fp, ln)
			evList := make([]map[string]interface{}, 0, len(item.TraversalChain))
			for _, hop := range item.TraversalChain {
				evList = append(evList, map[string]interface{}{
					"hop":             hop,
					"source_file":     fp,
					"line_start":      ln,
					"commit_sha":      canonCommit,
					"content_snippet": snippet,
				})
			}
			all = append(all, map[string]interface{}{
				"entity_id":       item.EntityID,
				"name":            item.Name,
				"kind":            item.Kind,
				"package":         item.Package,
				"file_path":       fp,
				"line_start":      ln,
				"content_snippet": snippet,
				"severity":        string(item.Severity),
				"hop_distance":    item.Depth,
				"confidence":      item.CumulativeConf,
				"path_evidence":   evList,
			})
		}
	}

	return map[string]interface{}{
		"target_entity_id":        entityID,
		"target_found":            true,
		"total_affected":          len(all),
		"fixtures_excluded_count": fixturesExcluded,
		"include_fixtures":        includeFixtures,
		"snapshot_receipt":        snapshot,
		"by_severity":             bySev,
		"impacted":                all,
	}, nil
}

func resolveOrDeriveUUID(tenantID uuid.UUID, kind, raw string) (uuid.UUID, bool) {
	if parsed, err := uuid.Parse(strings.TrimSpace(raw)); err == nil {
		return parsed, true
	}
	derived := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("%s:%s:%s", tenantID.String(), kind, strings.TrimSpace(raw))))
	return derived, false
}

func (s *MCPServer) handleHandoff(args map[string]interface{}) (interface{}, error) {
	if ReadOnlyMode() {
		return ReadOnlyRejection("garuda.handoff"), nil
	}
	ctx := context.Background()
	tenantID, _, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		tenantID = defaultTenantID
		if rawTid, ok := args["tenant_id"].(string); ok && rawTid != "" {
			if parsed, pErr := uuid.Parse(rawTid); pErr == nil {
				tenantID = parsed
			}
		}
	}

	taskStr, _ := args["task_id"].(string)
	if taskStr == "" {
		return nil, fmt.Errorf("task_id is required (call garuda.create_task first to provision a task UUID)")
	}
	taskID, taskWasUUID := resolveOrDeriveUUID(tenantID, "task", taskStr)

	sourceStr, _ := args["source_agent_id"].(string)
	if sourceStr == "" {
		return nil, fmt.Errorf("source_agent_id is required")
	}
	sourceID, _ := resolveOrDeriveUUID(tenantID, "agent", sourceStr)

	targetStr, _ := args["target_agent_id"].(string)
	if targetStr == "" {
		return nil, fmt.Errorf("target_agent_id is required")
	}
	targetID, _ := resolveOrDeriveUUID(tenantID, "agent", targetStr)

	if s.store != nil && s.safePool() != nil {
		if !taskWasUUID {
			_, _ = s.safePool().Exec(ctx, `
				INSERT INTO tenants (id, name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING
			`, tenantID, "tenant-"+tenantID.String()[:8])
			_, _ = s.safePool().Exec(ctx, `
				INSERT INTO agents (id, tenant_id, name, model_type, session_id, status)
				VALUES ($1, $2, $3, 'mcp-agent', $4, 'idle'),
				       ($5, $2, $6, 'mcp-agent', $7, 'idle')
				ON CONFLICT (id) DO UPDATE SET status = 'idle'
			`, sourceID, tenantID, sourceStr, "sess-"+sourceID.String()[:8], targetID, targetStr, "sess-"+targetID.String()[:8])
			_, _ = s.safePool().Exec(ctx, `
				INSERT INTO tasks (id, tenant_id, title, status, owner_agent_id)
				VALUES ($1, $2, $3, 'in_progress', $4)
				ON CONFLICT (id) DO UPDATE SET owner_agent_id = EXCLUDED.owner_agent_id, status = 'in_progress'
			`, taskID, tenantID, taskStr, sourceID)
		} else {
			var exists bool
			_ = s.safePool().QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1 AND tenant_id = $2)`, taskID, tenantID).Scan(&exists)
			if exists {
				_, _ = s.safePool().Exec(ctx, `
					INSERT INTO agents (id, tenant_id, name, model_type, session_id, status)
					VALUES ($1, $2, $3, 'mcp-agent', $4, 'idle'),
					       ($5, $2, $6, 'mcp-agent', $7, 'idle')
					ON CONFLICT (id) DO UPDATE SET status = 'idle'
				`, sourceID, tenantID, sourceStr, "sess-"+sourceID.String()[:8], targetID, targetStr, "sess-"+targetID.String()[:8])
			}
		}
	}

	reason, _ := args["reason"].(string)
	checkpointData := args["checkpoint_data"]
	if checkpointData == nil {
		checkpointData = map[string]interface{}{
			"task_alias":   taskStr,
			"source_alias": sourceStr,
			"target_alias": targetStr,
		}
	}

	req := &store.HandoffRequest{
		TenantID:       tenantID,
		TaskID:         taskID,
		SourceAgentID:  sourceID,
		TargetAgentID:  targetID,
		Reason:         reason,
		CheckpointData: checkpointData,
	}

	resp, err := s.store.ExecuteHandoffTransaction(ctx, req)
	if err != nil {
		reasonMsg := "handoff execution failed"
		switch {
		case errors.Is(err, store.ErrSourceTransitioning),
			errors.Is(err, store.ErrTargetOffline),
			errors.Is(err, store.ErrSourceDoesNotOwn):
			reasonMsg = err.Error()
		}
		return map[string]interface{}{
			"status":          "failed",
			"reason":          reasonMsg,
			"task_id":         taskID.String(),
			"source_agent_id": sourceID.String(),
			"target_agent_id": targetID.String(),
			"hint":            "Call garuda.create_task first to provision a valid task_id and owner_agent_id",
		}, nil
	}

	return map[string]interface{}{
		"status":          resp.Status,
		"handoff_id":      resp.HandoffID.String(),
		"checkpoint_id":   resp.CheckpointID.String(),
		"task_id":         resp.TaskID.String(),
		"task_alias":      taskStr,
		"source_agent_id": sourceID.String(),
		"target_agent_id": targetID.String(),
	}, nil
}

func (s *MCPServer) handleResume(args map[string]interface{}) (interface{}, error) {
	if ReadOnlyMode() {
		return ReadOnlyRejection("garuda.resume"), nil
	}
	ctx := context.Background()
	tenantID, _, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		tenantID = defaultTenantID
		if rawTid, ok := args["tenant_id"].(string); ok && rawTid != "" {
			if parsed, pErr := uuid.Parse(rawTid); pErr == nil {
				tenantID = parsed
			}
		}
	}

	agentStr, _ := args["agent_id"].(string)
	if agentStr == "" {
		return nil, fmt.Errorf("agent_id is required")
	}
	agentID, _ := resolveOrDeriveUUID(tenantID, "agent", agentStr)

	checkpointStr, _ := args["checkpoint_id"].(string)
	if checkpointStr == "" {
		return nil, fmt.Errorf("checkpoint_id is required")
	}

	checkpointID, cpWasUUID := resolveOrDeriveUUID(tenantID, "checkpoint", checkpointStr)
	if !cpWasUUID && s.store != nil && s.safePool() != nil {
		derivedTaskID, _ := resolveOrDeriveUUID(tenantID, "task", checkpointStr)
		var resolvedCP uuid.UUID
		qErr := s.safePool().QueryRow(ctx, `
			SELECT id FROM agent_checkpoints
			WHERE tenant_id = $1 AND (id = $2 OR task_id = $3) AND status = 'active'
			ORDER BY created_at DESC LIMIT 1
		`, tenantID, checkpointID, derivedTaskID).Scan(&resolvedCP)
		if qErr == nil && resolvedCP != uuid.Nil {
			checkpointID = resolvedCP
		}
	}

	state, err := s.store.ResumeAgent(ctx, tenantID, agentID, checkpointID)
	if err != nil {
		return map[string]interface{}{
			"status":        "not_found",
			"reason":        err.Error(),
			"agent_id":      agentID.String(),
			"checkpoint_id": checkpointID.String(),
		}, nil
	}

	return map[string]interface{}{
		"status":        "restored",
		"agent_id":      agentID.String(),
		"checkpoint_id": checkpointID.String(),
		"state":         state,
	}, nil
}
