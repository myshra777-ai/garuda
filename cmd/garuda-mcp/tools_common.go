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
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
		"garuda.reset_audit_state":
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

// WorkspaceCountsSummary is the single source of truth for workspace entity,
// relationship, and epistemic claim counts across briefing, governance.status,
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
}

// LoadWorkspaceCounts queries the canonical workspace entity, AST edge, and
// document claim counts using a single consistent filter definition:
//   - Entities:           kind != 'external'
//   - ExportedEntities:   kind != 'external' AND is_exported = true
//   - ExternalStubs:      kind = 'external'
//   - ASTRelationships:   COUNT(*) FROM claims WHERE workspace_id = $1
//   - DocumentClaims:     COUNT(*) FROM document_claims WHERE workspace_id = $1
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
	return out
}
