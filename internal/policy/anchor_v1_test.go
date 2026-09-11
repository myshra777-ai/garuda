// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/store"
)

func anchorTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("GARUDA_TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("GARUDA_TEST_DATABASE_URL not set; skipping anchor v1 integration test")
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	return pool
}

func TestAnchor_AppendAndVerifyV1(t *testing.T) {
	pool := anchorTestPool(t)
	defer pool.Close()

	tenantID := uuid.New()
	ctx := context.Background()

	// Clean any prior state.
	_, _ = pool.Exec(ctx, `DELETE FROM merkle_epoch_leaves WHERE tenant_id = $1`, tenantID)
	_, _ = pool.Exec(ctx, `DELETE FROM merkle_roots WHERE tenant_id = $1`, tenantID)
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM merkle_epoch_leaves WHERE tenant_id = $1`, tenantID)
		_, _ = pool.Exec(ctx, `DELETE FROM merkle_roots WHERE tenant_id = $1`, tenantID)
	}()

	// policy_evaluations.policy_id has an FK to policies(id). Create a
	// minimal policy row first so the constraint is satisfiable.
	policyID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO policies (
			id, tenant_id, statement,
			scope_domain, scope_system, actor, status,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, 'test', 'test', 'test-runner', 'active', NOW(), NOW())
	`, policyID, tenantID, "TEST POLICY:anchor-v1")
	if err != nil {
		t.Fatalf("insert policy: %v", err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM policies WHERE id = $1`, policyID)
	}()

	ev := &Evaluation{
		ID:            uuid.New(),
		TenantID:      tenantID,
		WorkspaceID:   uuid.New(),
		PolicyID:      policyID,
		PolicyVersion: "v1",
		Decision:      DecisionWarn,
		Reason:        "test reason",
		SubjectKind:   "claim",
		SubjectID:     "subj-1",
		Actor:         "test-runner",
	}

	anchor := NewAnchor(pool)
	height, proofBytes, err := anchor.AppendEvaluation(ctx, ev)
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if height <= 0 {
		t.Fatalf("height = %d, want positive", height)
	}
	if len(proofBytes) == 0 {
		t.Fatal("proof bytes empty")
	}

	// Persist the row so VerifyEvaluation can find the proof.
	ev.MerkleBlockHeight = &height
	if _, err := pool.Exec(ctx, `
		INSERT INTO policy_evaluations (
			id, tenant_id, workspace_id, policy_id, policy_version,
			decision, reason, evidence, evaluated_at,
			merkle_block_height, merkle_proof, verification_version,
			subject_kind, subject_id, actor
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,1,$12,$13,$14)
	`,
		ev.ID, ev.TenantID, ev.WorkspaceID, ev.PolicyID, ev.PolicyVersion,
		string(ev.Decision), ev.Reason, []byte("{}"), ev.EvaluatedAt,
		ev.MerkleBlockHeight, proofBytes,
		ev.SubjectKind, ev.SubjectID, ev.Actor,
	); err != nil {
		t.Fatalf("persist eval: %v", err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM policy_evaluations WHERE id = $1`, ev.ID)
	}()

	ok, err := anchor.VerifyEvaluation(ctx, ev)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !ok {
		t.Fatal("v1 proof did not verify")
	}

	// Prove the proof parses back as a v1 proof.
	v1, err := store.UnmarshalV1Proof(proofBytes)
	if err != nil {
		t.Fatalf("proof is not v1: %v", err)
	}
	if v1.Version != 1 {
		t.Errorf("proof version = %d, want 1", v1.Version)
	}
	if v1.EpochHeight != height {
		t.Errorf("proof height = %d, want %d", v1.EpochHeight, height)
	}
}
