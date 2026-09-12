// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/knowledge"
	"github.com/myshra777-ai/garuda/internal/policy"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/tenant"
)

// defaultTenantID is the fallback tenant used when neither the request
// nor the environment specifies one. It matches the default used by
// the CLI (getTenantID in cmd/garuda/root.go).
var defaultTenantID = tenant.CanonicalID

// resolveTenantAndWorkspace extracts tenant_id and workspace from the
// request arguments, with fallback to environment variables and then
// to the most recently updated workspace for the tenant.
//
// Precedence:
//
//	tenant_id  →  GARUDA_TENANT_ID  →  defaultTenantID
//	workspace  →  GARUDA_WORKSPACE  →  most recently updated
//
// Callers that supply an invalid tenant UUID receive an error, not a
// silent fallback. Invalid input is a bug in the caller, not a case to
// paper over. A workspace name that does not resolve to a row is also
// an error.
func (s *MCPServer) resolveTenantAndWorkspace(args map[string]interface{}) (uuid.UUID, uuid.UUID, error) {
	tenantID := defaultTenantID

	if raw, ok := args["tenant_id"].(string); ok && raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("invalid tenant_id %q: %w", raw, err)
		}
		tenantID = parsed
	} else if envTid := os.Getenv("GARUDA_TENANT_ID"); envTid != "" {
		parsed, err := uuid.Parse(envTid)
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("invalid GARUDA_TENANT_ID %q: %w", envTid, err)
		}
		tenantID = parsed
	}

	// Workspace name comes from args, then env, then empty. Empty
	// resolves to the most recently updated workspace for the tenant.
	// The previous literal "default" matched no workspace row unless
	// `garuda init` had seeded it.
	workspaceName := ""
	if raw, ok := args["workspace"].(string); ok && raw != "" {
		workspaceName = raw
	} else if envWs := os.Getenv("GARUDA_WORKSPACE"); envWs != "" {
		workspaceName = envWs
	}

	workspaceID, err := store.ResolveWorkspaceID(context.Background(), s.store.Pool(), tenantID, workspaceName)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("resolve workspace %q: %w", workspaceName, err)
	}
	return tenantID, workspaceID, nil
}

// handlePolicyList returns every policy registered for the tenant.
//
// Read-only. Does not evaluate, does not anchor, does not touch the
// Merkle log.
func (s *MCPServer) handlePolicyList(args map[string]interface{}) (interface{}, error) {
	tenantID, _, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}

	rows, err := s.store.Pool().Query(context.Background(), `
		SELECT id, statement, scope_domain, scope_system, actor, status,
		       policy_version, created_at
		FROM policies
		WHERE tenant_id = $1
		ORDER BY created_at DESC
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
	for rows.Next() {
		var p PolicyRow
		var createdAt interface{}
		if err := rows.Scan(&p.ID, &p.Statement, &p.ScopeDomain,
			&p.ScopeSystem, &p.Actor, &p.Status, &p.PolicyVersion, &createdAt); err != nil {
			continue
		}
		p.CreatedAt = fmt.Sprint(createdAt)
		policies = append(policies, p)
	}

	return map[string]interface{}{
		"tenant_id": tenantID.String(),
		"count":     len(policies),
		"policies":  policies,
	}, nil
}

// handleGovernanceStatus returns an aggregate view of governance state
// for the workspace: active policy count, documentation health, and
// contradiction count.
//
// Read-only. The evaluator runs queries but writes nothing.
func (s *MCPServer) handleGovernanceStatus(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}

	// Active policy count.
	var policyCount int
	_ = s.store.Pool().QueryRow(context.Background(), `
		SELECT COUNT(*) FROM policies
		WHERE tenant_id = $1 AND status = 'active'
	`, tenantID).Scan(&policyCount)

	// Documentation health and drift findings.
	evaluator := knowledge.NewEvaluator(s.store.Pool())
	stats, drift, err := evaluator.EvaluateWorkspace(context.Background(), tenantID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("evaluate workspace: %w", err)
	}

	verdict := computeVerdict(stats)

	return map[string]interface{}{
		"tenant_id":              tenantID.String(),
		"workspace":              workspaceID.String(),
		"active_policies":        policyCount,
		"documentation_health":   stats,
		"drift_finding_count":    len(drift),
		"drift_findings_preview": previewDrift(drift, 10),
		"gate_verdict":           verdict,
	}, nil
}

// governanceVerdict is a structured, honest assessment of whether the
// workspace currently demonstrates conformance to its declared intent.
//
// The distinction between AgentProceed and Conforms matters:
//
//	UNVERIFIED → AgentProceed=true,  Conforms=false
//	  An unverified workspace is not a blocked one. But neither is it
//	  a proven-conformant one. The agent may proceed; it must not
//	  claim the state has been verified.
//
//	NO_DATA → AgentProceed=true, Conforms=false
//	  No documentation has been ingested. We have no basis for
//	  judgement in either direction.
//
//	BLOCKED → AgentProceed=false, Conforms=false
//	  Contradictions exist. Do not proceed without resolving them.
//
//	PARTIAL → AgentProceed=true,  Conforms=true
//	  At least one claim has been verified, and no contradictions were
//	  found. Conformance is demonstrated but not complete.
//
//	CONFORMS → AgentProceed=true,  Conforms=true
//	  Every claim verified, no contradictions.
//
// The previous implementation collapsed all of these into a single
// "PASSED" or "BLOCKED" string. That produced a false claim of
// conformance whenever zero claims had been verified — precisely the
// case this function is designed to name honestly.
type governanceVerdict struct {
	State        string `json:"state"`
	Reason       string `json:"reason"`
	AgentProceed bool   `json:"agent_should_proceed"`
	Conforms     bool   `json:"conformance_demonstrated"`
}

// computeVerdict returns the honest governance verdict for a workspace
// given its current documentation health.
//
// Rules, in priority order:
//
//  1. Any contradiction → BLOCKED.
//  2. Zero claims ingested → NO_DATA.
//  3. Zero supported claims → UNVERIFIED.
//  4. Some but not all supported → PARTIAL.
//  5. All supported → CONFORMS.
//
// Rule 1 takes precedence because a contradiction invalidates the
// workspace as a working surface for agents regardless of how many
// claims have been verified.
func computeVerdict(stats *knowledge.HealthStats) governanceVerdict {
	switch {
	case stats == nil:
		return governanceVerdict{
			State:        "NO_DATA",
			Reason:       "Documentation health could not be evaluated for this workspace.",
			AgentProceed: true,
			Conforms:     false,
		}
	case stats.Contradicted > 0:
		return governanceVerdict{
			State: "BLOCKED",
			Reason: fmt.Sprintf(
				"%d documentation claims exist; %d contradicted by the codebase. Resolve contradictions before proceeding.",
				stats.TotalClaims, stats.Contradicted),
			AgentProceed: false,
			Conforms:     false,
		}
	case stats.TotalClaims == 0:
		return governanceVerdict{
			State: "NO_DATA",
			Reason: "No documentation claims have been ingested for this workspace. " +
				"Run 'garuda docs ingest' to enable conformance checks.",
			AgentProceed: true,
			Conforms:     false,
		}
	case stats.Supported == 0:
		return governanceVerdict{
			State: "UNVERIFIED",
			Reason: fmt.Sprintf(
				"%d documentation claims exist but none have been verified against the codebase. Conformance has not been demonstrated.",
				stats.TotalClaims),
			AgentProceed: true,
			Conforms:     false,
		}
	case stats.Supported < stats.TotalClaims:
		return governanceVerdict{
			State: "PARTIAL",
			Reason: fmt.Sprintf(
				"%d of %d documentation claims verified against the codebase. Partial conformance demonstrated.",
				stats.Supported, stats.TotalClaims),
			AgentProceed: true,
			Conforms:     true,
		}
	default:
		return governanceVerdict{
			State: "CONFORMS",
			Reason: fmt.Sprintf(
				"All %d documentation claims verified against the codebase. No contradictions detected.",
				stats.TotalClaims),
			AgentProceed: true,
			Conforms:     true,
		}
	}
}

// handleCheckDrift returns the documentation-to-code drift report for
// the workspace.
//
// Read-only. Equivalent to what HandleCheckDrift in
// internal/mcp/knowledge_tool.go returns, but resolves tenant and
// workspace from the caller's args instead of hardcoding defaults.
func (s *MCPServer) handleCheckDrift(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}

	evaluator := knowledge.NewEvaluator(s.store.Pool())
	stats, drift, err := evaluator.EvaluateWorkspace(context.Background(), tenantID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("evaluate workspace: %w", err)
	}

	verdict := computeVerdict(stats)

	return map[string]interface{}{
		"tenant_id":            tenantID.String(),
		"workspace":            workspaceID.String(),
		"documentation_health": stats,
		"undocumented_symbols": drift,
		"gate_verdict":         verdict,
		"status":               "success",
	}, nil
}

// handleQueryClaims returns document claims matching an optional
// subject filter.
//
// Read-only.
func (s *MCPServer) handleQueryClaims(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}

	subject, _ := args["subject"].(string)

	query := `
		SELECT document_title, section_title, subject, modality,
		       predicate, object, status, contradiction_reason
		FROM document_claims
		WHERE tenant_id = $1 AND workspace_id = $2
	`
	dbArgs := []interface{}{tenantID, workspaceID}

	if subject != "" {
		query += " AND (subject ILIKE $3 OR object ILIKE $3)"
		dbArgs = append(dbArgs, "%"+subject+"%")
	}
	query += " ORDER BY line_start ASC"

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
		Contradiction string `json:"contradiction_reason,omitempty"`
	}

	results := make([]ClaimView, 0)
	for rows.Next() {
		var v ClaimView
		if err := rows.Scan(&v.Document, &v.Section, &v.Subject, &v.Modality,
			&v.Predicate, &v.Object, &v.Status, &v.Contradiction); err == nil {
			results = append(results, v)
		}
	}

	return map[string]interface{}{
		"tenant_id": tenantID.String(),
		"workspace": workspaceID.String(),
		"subject":   subject,
		"count":     len(results),
		"claims":    results,
	}, nil
}

// previewDrift returns at most n findings, for inclusion in summary
// responses where the full list would be noise.
func previewDrift(findings []knowledge.CodeDriftFinding, n int) []knowledge.CodeDriftFinding {
	if len(findings) <= n {
		return findings
	}
	return findings[:n]
}

// handlePolicyEvaluate runs the policy engine against a workspace in
// dry-run mode. Decisions are computed and returned but are NOT
// persisted or Merkle-anchored. A separate human-driven CLI invocation
// (garuda policy evaluate) is required to anchor a governance decision
// to the ledger.
//
// Rationale: policy evaluation is a governance act. If an agent could
// trigger arbitrary anchored evaluations, the audit log fills with
// exploratory noise. Preview is safe for agents; commit requires a
// human.
func (s *MCPServer) handlePolicyEvaluate(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}

	// Keep the workspace name for display. resolveTenantAndWorkspace
	// returns the UUID; the name is what a human reads in logs and
	// responses, and it may be empty if the resolver picked the most
	// recently updated workspace.
	workspaceName, _ := args["workspace"].(string)
	if workspaceName == "" {
		workspaceName = os.Getenv("GARUDA_WORKSPACE")
	}
	if workspaceName == "" {
		workspaceName = workspaceID.String()
	}

	policyDir, _ := args["policy_dir"].(string)
	if policyDir == "" {
		policyDir = os.Getenv("GARUDA_POLICY_DIR")
	}
	if policyDir == "" {
		policyDir = "./policies"
	}

	subjectKind, _ := args["subject_kind"].(string)
	subjectID, _ := args["subject_id"].(string)
	actor, _ := args["actor"].(string)
	if actor == "" {
		actor = "mcp-policy-preview"
	}

	engine := policy.NewEngine(s.store.Pool())
	result, err := engine.Run(
		context.Background(),
		tenantID, workspaceID,
		policyDir,
		subjectKind, subjectID, actor,
		policy.RunOptions{DryRun: true},
	)
	if err != nil {
		return nil, fmt.Errorf("policy evaluation failed: %w", err)
	}

	// Shape the response with the fields that matter for an agent
	// deciding whether to proceed.
	evaluations := make([]map[string]interface{}, 0, len(result.Evaluations))
	for _, ev := range result.Evaluations {
		evaluations = append(evaluations, map[string]interface{}{
			"policy_id":           ev.PolicyID.String(),
			"decision":            string(ev.Decision),
			"reason":              ev.Reason,
			"matched_predicates":  ev.Evidence.MatchedPredicates,
			"entity_count":        len(ev.Evidence.EntityIDs),
			"claim_count":         len(ev.Evidence.ClaimIDs),
			"contradiction_count": len(ev.Evidence.ContradictionIDs),
			"anchor":              nil, // always nil in dry run
		})
	}

	return map[string]interface{}{
		"tenant_id":      tenantID.String(),
		"workspace":      workspaceName,
		"policy_dir":     policyDir,
		"dry_run":        true,
		"final_decision": string(result.FinalDecision),
		"evaluations":    evaluations,
		"summary":        result.Summary,
		"note": "Dry-run only. No evaluation was persisted or anchored. " +
			"Run 'garuda policy evaluate <dir> --workspace " + workspaceName +
			"' from the CLI to record and anchor these decisions.",
	}, nil
}
