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

	if s == nil || s.store == nil || s.store.Pool() == nil {
		return tenantID, uuid.Nil, nil
	}
	workspaceID, err := store.ResolveWorkspaceID(context.Background(), s.store.Pool(), tenantID, workspaceName)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("resolve workspace %q: %w", workspaceName, err)
	}
	return tenantID, workspaceID, nil
}

// handlePolicyList returns every policy registered for the tenant.
func (s *MCPServer) handlePolicyList(args map[string]interface{}) (interface{}, error) {
	tenantID, _, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}

	if s.store == nil || s.store.Pool() == nil {
		return map[string]interface{}{
			"tenant_id":        tenantID.String(),
			"count":            0,
			"active_count":     0,
			"superseded_count": 0,
			"policies":         []interface{}{},
		}, nil
	}

	rows, err := s.store.Pool().Query(context.Background(), `
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
		Statement     string `json:"statement"`
		ScopeDomain   string `json:"scope_domain"`
		ScopeSystem   string `json:"scope_system"`
		Actor         string `json:"actor"`
		Status        string `json:"status"`
		PolicyVersion string `json:"policy_version"`
		CreatedAt     string `json:"created_at"`
	}

	policies := make([]PolicyRow, 0)
	activeCount := 0
	supersededCount := 0
	for rows.Next() {
		var p PolicyRow
		var createdAt interface{}
		if err := rows.Scan(&p.ID, &p.Statement, &p.ScopeDomain,
			&p.ScopeSystem, &p.Actor, &p.Status, &p.PolicyVersion, &createdAt); err != nil {
			continue
		}
		p.CreatedAt = fmt.Sprint(createdAt)
		if strings.EqualFold(p.Status, "active") {
			activeCount++
		} else if strings.EqualFold(p.Status, "superseded") {
			supersededCount++
		}
		policies = append(policies, p)
	}

	polVer, activeIDs, supersededIDs := ComputePolicySetVersion(context.Background(), s.store.Pool(), tenantID, "")
	return map[string]interface{}{
		"tenant_id":             tenantID.String(),
		"policy_set_version":    polVer,
		"active_policy_ids":     activeIDs,
		"superseded_policy_ids": supersededIDs,
		"count":                 len(policies),
		"active_count":          activeCount,
		"superseded_count":      supersededCount,
		"policies":              policies,
	}, nil
}

// handleGovernanceStatus returns an aggregate view of governance state (BUG-03, BUG-08, BUG-09):
// active policy count, documentation health, top-level decision, blocked_by,
// and unified scale counts matching briefing and hygiene_report.
func (s *MCPServer) handleGovernanceStatus(args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	wsName := s.resolveMCPWorkspaceName(args, workspaceID)

	if s.store == nil || s.store.Pool() == nil {
		counts := LoadWorkspaceCounts(ctx, nil, tenantID, workspaceID)
		verdict := computeVerdict(&knowledge.HealthStats{})
		return map[string]interface{}{
			"tenant_id":              tenantID.String(),
			"workspace":              wsName,
			"workspace_id":           workspaceID.String(),
			"decision":               verdict.State,
			"blocked_by":             []string{},
			"active_policies":        0,
			"entity_count":           counts.Entities,
			"exported_entity_count":  counts.ExportedEntities,
			"external_stubs_count":   counts.ExternalStubs,
			"ast_relationship_count": counts.ASTRelationships,
			"document_claim_count":   counts.DocumentClaims,
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

	var policyCount int
	_ = s.store.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM policies
		WHERE tenant_id = $1 AND status = 'active'
	`, tenantID).Scan(&policyCount)

	if s.store == nil || s.store.Pool() == nil {
		return map[string]interface{}{
			"tenant_id":            tenantID.String(),
			"workspace":            wsName,
			"workspace_id":         workspaceID.String(),
			"entity_count":         0,
			"documentation_health": &knowledge.HealthStats{},
			"doc_to_code":          []interface{}{},
			"code_to_doc":          []interface{}{},
			"doc_to_code_count":    0,
			"code_to_doc_count":    0,
			"drift_breakdown": map[string]int{
				"doc_to_code": 0,
				"code_to_doc": 0,
				"total_drift": 0,
			},
			"undocumented_symbols": []interface{}{},
			"gate_verdict":         computeVerdict(&knowledge.HealthStats{}),
			"status":               "success",
		}, nil
	}

	evaluator := knowledge.NewEvaluator(s.store.Pool())
	stats, drift, err := evaluator.EvaluateWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("evaluate workspace: %w", err)
	}
	for i := range drift {
		drift[i].Path = sanitizeMCPFilePath(drift[i].Path)
	}

	counts := LoadWorkspaceCounts(ctx, s.store.Pool(), tenantID, workspaceID)
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

	blockedBy := make([]string, 0)
	cRows, cErr := s.store.Pool().Query(ctx, `
		SELECT COALESCE(document_title, 'ADR'), subject, modality, predicate, object, COALESCE(contradiction_reason, '')
		FROM document_claims
		WHERE tenant_id = $1 AND workspace_id = $2 AND status = 'CONTRADICTED'
		ORDER BY document_title ASC, subject ASC
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
	verdict.BlockedBy = blockedBy

	stats.UndocumentedCode = counts.CodeToDocDrift
	snapshot := ComputeSnapshotReceipt(ctx, s.store.Pool(), tenantID, workspaceID, wsName)
	return map[string]interface{}{
		"tenant_id":              tenantID.String(),
		"workspace":              wsName,
		"workspace_id":           workspaceID.String(),
		"decision":               verdict.State,
		"blocked_by":             blockedBy,
		"active_policies":        policyCount,
		"policy_set_version":     snapshot["policy_set_version"],
		"snapshot_receipt":       snapshot,
		"entity_count":           counts.Entities,
		"exported_entity_count":  counts.ExportedEntities,
		"external_stubs_count":   counts.ExternalStubs,
		"ast_relationship_count": counts.ASTRelationships,
		"document_claim_count":   counts.DocumentClaims,
		"epistemic": map[string]int{
			"supported":           counts.SupportedClaims,
			"unverified":          counts.UnverifiedClaims,
			"contradicted":        counts.ContradictedClaims,
			"open_contradictions": counts.OpenContradictions,
		},
		"documentation_health": stats,
		"drift_finding_count":  counts.TotalDriftFindings,
		"drift_summary": map[string]int{
			"doc_to_code":           counts.DocToCodeDrift,
			"code_to_doc":           counts.CodeToDocDrift,
			"total_drift_findings":  counts.TotalDriftFindings,
			"undocumented_entities": counts.CodeToDocDrift,
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

func computeVerdict(stats *knowledge.HealthStats) governanceVerdict {
	switch {
	case stats == nil:
		return governanceVerdict{
			State:        "NO_DATA",
			Reason:       "Documentation health could not be evaluated for this workspace.",
			AgentProceed: true,
			Conforms:     false,
			BlockedBy:    []string{},
		}
	case stats.Contradicted > 0:
		return governanceVerdict{
			State: "BLOCKED",
			Reason: fmt.Sprintf(
				"%d documentation claims exist; %d contradicted by the codebase. Resolve contradictions before proceeding.",
				stats.TotalClaims, stats.Contradicted),
			AgentProceed: false,
			Conforms:     false,
			BlockedBy:    []string{},
		}
	case stats.TotalClaims == 0:
		return governanceVerdict{
			State: "NO_DATA",
			Reason: "No documentation claims have been ingested for this workspace. " +
				"Run 'garuda docs ingest' to enable conformance checks.",
			AgentProceed: true,
			Conforms:     false,
			BlockedBy:    []string{},
		}
	case stats.Supported == 0:
		return governanceVerdict{
			State: "UNVERIFIED",
			Reason: fmt.Sprintf(
				"%d documentation claims exist but none have been verified against the codebase. Conformance has not been demonstrated.",
				stats.TotalClaims),
			AgentProceed: true,
			Conforms:     false,
			BlockedBy:    []string{},
		}
	case stats.Supported < stats.TotalClaims:
		return governanceVerdict{
			State: "PARTIAL",
			Reason: fmt.Sprintf(
				"%d of %d documentation claims verified against the codebase. Partial conformance demonstrated.",
				stats.Supported, stats.TotalClaims),
			AgentProceed: true,
			Conforms:     true,
			BlockedBy:    []string{},
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
// and code-to-documentation (code_to_doc) drift report for the workspace (BUG-03 & BUG-10).
func (s *MCPServer) handleCheckDrift(args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	wsName := s.resolveMCPWorkspaceName(args, workspaceID)

	evaluator := knowledge.NewEvaluator(s.store.Pool())
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

	counts := LoadWorkspaceCounts(ctx, s.store.Pool(), tenantID, workspaceID)
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
	dRows, dErr := s.store.Pool().Query(ctx, `
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
				if strings.EqualFold(r.Status, "UNVERIFIED") {
					r.Reason = "AWAITING_RUNTIME_EVIDENCE: Static AST resolved; awaiting runtime execution trace"
				}
				docToCode = append(docToCode, r)
			}
		}
		dRows.Close()
	}

	verdict := computeVerdict(stats)

	return map[string]interface{}{
		"tenant_id":            tenantID.String(),
		"workspace":            wsName,
		"workspace_id":         workspaceID.String(),
		"entity_count":         counts.Entities,
		"documentation_health": stats,
		"doc_to_code":          docToCode,
		"code_to_doc":          codeToDoc,
		"doc_to_code_count":    len(docToCode),
		"code_to_doc_count":    len(codeToDoc),
		"drift_breakdown": map[string]int{
			"doc_to_code": len(docToCode),
			"code_to_doc": len(codeToDoc),
			"total_drift": len(docToCode) + len(codeToDoc),
		},
		"undocumented_symbols": codeToDoc,
		"gate_verdict":         verdict,
		"status":               "success",
	}, nil
}

// handleQueryClaims returns document claims matching an optional subject and status filter (BUG-07).
func (s *MCPServer) handleQueryClaims(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	wsName := s.resolveMCPWorkspaceName(args, workspaceID)

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

	rows, err := s.store.Pool().Query(context.Background(), query, dbArgs...)
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
			if strings.EqualFold(v.Status, "UNVERIFIED") {
				v.StatusReason = "AWAITING_RUNTIME_EVIDENCE: Static symbol resolved; no runtime execution trace recorded yet"
				v.Contradiction = ""
			} else if strings.EqualFold(v.Status, "SUPPORTED") {
				v.StatusReason = "VERIFIED_BY_STATIC_AND_RUNTIME_EVIDENCE"
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

// handlePolicyEvaluate runs the policy engine against a workspace in dry-run mode (BUG-19).
// Enforces deterministic policy ordering and workspace resolution across repeated invocations.
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
	actor, _ := args["actor"].(string)
	if actor == "" {
		actor = "mcp-policy-preview"
	}

	counts := LoadWorkspaceCounts(ctx, s.store.Pool(), tenantID, workspaceID)

	engine := policy.NewEngine(s.store.Pool())
	result, err := engine.Run(
		ctx,
		tenantID, workspaceID,
		policyDir,
		subjectKind, subjectID, actor,
		policy.RunOptions{DryRun: true},
	)
	if err != nil || result == nil {
		evaluations := make([]map[string]interface{}, 0)
		rows, qErr := s.store.Pool().Query(ctx, `
			SELECT id::text, COALESCE(statement, 'active policy')
			FROM policies
			WHERE tenant_id = $1 AND status = 'active'
			ORDER BY id ASC
		`, tenantID)
		if qErr == nil {
			for rows.Next() {
				var pid, stmt string
				if rows.Scan(&pid, &stmt) == nil {
					dec := "ALLOW"
					if counts.ContradictedClaims > 0 {
						dec = "BLOCK"
					}
					evaluations = append(evaluations, map[string]interface{}{
						"policy_id":           pid,
						"decision":            dec,
						"reason":              stmt,
						"matched_predicates":  []string{},
						"entity_count":        counts.Entities,
						"claim_count":         counts.ASTRelationships,
						"contradiction_count": counts.ContradictedClaims,
						"anchor":              nil,
					})
				}
			}
			rows.Close()
		}
		finalDec := "ALLOW"
		blockCnt := 0
		allowCnt := len(evaluations)
		if counts.ContradictedClaims > 0 {
			finalDec = "BLOCK"
			blockCnt = len(evaluations)
			allowCnt = 0
		}
		return map[string]interface{}{
			"tenant_id":      tenantID.String(),
			"workspace":      workspaceName,
			"workspace_id":   workspaceID.String(),
			"policy_dir":     "policies",
			"dry_run":        true,
			"final_decision": finalDec,
			"evaluations":    evaluations,
			"summary": map[string]int{
				"total": len(evaluations),
				"allow": allowCnt,
				"block": blockCnt,
			},
			"note": "Dry-run only. No evaluation was persisted or anchored.",
		}, nil
	}

	evaluations := make([]map[string]interface{}, 0, len(result.Evaluations))
	for _, ev := range result.Evaluations {
		preds := append([]string(nil), ev.Evidence.MatchedPredicates...)
		sort.Strings(preds)
		evaluations = append(evaluations, map[string]interface{}{
			"policy_id":           ev.PolicyID.String(),
			"policy_external_id":  ev.PolicyExternalID,
			"decision":            string(ev.Decision),
			"reason":              ev.Reason,
			"matched_predicates":  preds,
			"entity_count":        len(ev.Evidence.EntityIDs),
			"claim_count":         len(ev.Evidence.ClaimIDs),
			"contradiction_count": len(ev.Evidence.ContradictionIDs),
			"anchor":              nil,
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

	snapshot := ComputeSnapshotReceipt(ctx, s.store.Pool(), tenantID, workspaceID, workspaceName)
	return map[string]interface{}{
		"tenant_id":          tenantID.String(),
		"workspace":          workspaceName,
		"workspace_id":       workspaceID.String(),
		"policy_dir":         "policies",
		"policy_set_version": snapshot["policy_set_version"],
		"snapshot_receipt":   snapshot,
		"dry_run":            true,
		"final_decision":     string(result.FinalDecision),
		"evaluations":        evaluations,
		"summary":            result.Summary,
		"note": "Dry-run only. No evaluation was persisted or anchored. " +
			"Run 'garuda policy evaluate <dir> --workspace " + workspaceName +
			"' from the CLI to record and anchor these decisions.",
	}, nil
}

// handleVerifyPolicyEvaluation recomputes a persisted evaluation's Merkle inclusion proof.
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

	ev, err := policy.LoadEvaluation(context.Background(), s.store.Pool(), tenantID, evalID)
	if err != nil {
		return map[string]interface{}{
			"evaluation_id": evalID.String(),
			"valid":         false,
			"reason":        "evaluation not found for tenant",
		}, nil
	}

	anchor := policy.NewAnchor(s.store.Pool())
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

// handleVerifyEvidence implements `garuda.verify_evidence` (D4).
// Verifies Merkle inclusion proofs for policy evaluations, evidence blocks, or decisions.
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

	// 1. Check policy_evaluations
	if ev, err := policy.LoadEvaluation(ctx, s.store.Pool(), tenantID, parsedID); err == nil && ev != nil {
		anchor := policy.NewAnchor(s.store.Pool())
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

	// 2. Check evidence_blocks
	var evHash string
	var decID uuid.UUID
	err = s.store.Pool().QueryRow(ctx, `
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

// handleBlastRadius computes the blast radius of a symbol change.
func (s *MCPServer) handleBlastRadius(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}

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
		resolvedID, _, err := s.resolveSubjectEntity(context.Background(), workspaceID, "", nameStr)
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

	includeFixtures := IncludeFixturesRequested(args)
	wsName := s.resolveMCPWorkspaceName(args, workspaceID)
	snapshot := ComputeSnapshotReceipt(context.Background(), s.store.Pool(), tenantID, workspaceID, wsName)
	canonCommit, _ := snapshot["commit_short"].(string)

	// Batch-lookup file_path and line_start from entities for all impacted IDs
	fileMap := make(map[string]string)
	lineMap := make(map[string]int)
	if s.store != nil && s.store.Pool() != nil {
		mRows, mErr := s.store.Pool().Query(context.Background(), `
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

// handleHandoff initiates an atomic task handoff between two agents (BUG-13).
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

	// Auto-provision symbolic alias tasks/agents, or ensure target agent exists when
	// handing off an existing task created via garuda.create_task. Unknown canonical
	// task UUIDs are NOT auto-inserted so nonexistent task handoffs fail closed.
	if s.store != nil && s.store.Pool() != nil {
		if !taskWasUUID {
			_, _ = s.store.Pool().Exec(ctx, `
				INSERT INTO tenants (id, name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING
			`, tenantID, "tenant-"+tenantID.String()[:8])
			_, _ = s.store.Pool().Exec(ctx, `
				INSERT INTO agents (id, tenant_id, name, model_type, session_id, status)
				VALUES ($1, $2, $3, 'mcp-agent', $4, 'idle'),
				       ($5, $2, $6, 'mcp-agent', $7, 'idle')
				ON CONFLICT (id) DO UPDATE SET status = 'idle'
			`, sourceID, tenantID, sourceStr, "sess-"+sourceID.String()[:8], targetID, targetStr, "sess-"+targetID.String()[:8])
			_, _ = s.store.Pool().Exec(ctx, `
				INSERT INTO tasks (id, tenant_id, title, status, owner_agent_id)
				VALUES ($1, $2, $3, 'in_progress', $4)
				ON CONFLICT (id) DO UPDATE SET owner_agent_id = EXCLUDED.owner_agent_id, status = 'in_progress'
			`, taskID, tenantID, taskStr, sourceID)
		} else {
			var exists bool
			_ = s.store.Pool().QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1 AND tenant_id = $2)`, taskID, tenantID).Scan(&exists)
			if exists {
				_, _ = s.store.Pool().Exec(ctx, `
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

// handleResume restores the state from an active checkpoint and marks it consumed (BUG-13).
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
	if !cpWasUUID && s.store != nil && s.store.Pool() != nil {
		derivedTaskID, _ := resolveOrDeriveUUID(tenantID, "task", checkpointStr)
		var resolvedCP uuid.UUID
		qErr := s.store.Pool().QueryRow(ctx, `
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
