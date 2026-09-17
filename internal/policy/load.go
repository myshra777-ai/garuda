// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package policy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LoadEvaluation loads one policy evaluation by id, scoped to a
// tenant. Returns an error if the evaluation does not exist or does
// not belong to the tenant.
//
// This is the shared loader for both the CLI (`garuda policy verify`)
// and the MCP server (`garuda.verify_policy_evaluation`). Promoted
// from cmd/garuda during the MCP surface expansion so that "how to
// load an evaluation by id" has one definition, not two.
func LoadEvaluation(ctx context.Context, pool *pgxpool.Pool, tenantID, evalID uuid.UUID) (*Evaluation, error) {
	var ev Evaluation
	var evidenceJSON []byte
	var decision, reason string
	err := pool.QueryRow(ctx, `
		SELECT id, tenant_id, workspace_id, policy_id, policy_version,
		       decision, reason, evidence, evaluated_at,
		       merkle_block_height, merkle_proof,
		       subject_kind, subject_id, actor
		FROM policy_evaluations WHERE id = $1 AND tenant_id = $2
	`, evalID, tenantID).Scan(
		&ev.ID, &ev.TenantID, &ev.WorkspaceID, &ev.PolicyID, &ev.PolicyVersion,
		&decision, &reason, &evidenceJSON, &ev.EvaluatedAt,
		&ev.MerkleBlockHeight, &ev.MerkleProof,
		&ev.SubjectKind, &ev.SubjectID, &ev.Actor,
	)
	if err != nil {
		return nil, fmt.Errorf("evaluation not found: %w", err)
	}
	ev.Decision = Decision(decision)
	ev.Reason = reason
	_ = json.Unmarshal(evidenceJSON, &ev.Evidence)
	return &ev, nil
}
