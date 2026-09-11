// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/merkle"
	"github.com/myshra777-ai/garuda/internal/store"
)

// TestAnchor_Coexistence_V0AndV1 writes one v1 evaluation and one v0
// evaluation for the same tenant, persists both, and confirms that
// VerifyEvaluation selects the correct verification path for each.
//
// This is the ADR-0002 Commit 5b.4 integration test. It proves that a
// single database can hold rows from both Merkle schemes, that the
// verification path is chosen automatically from the stored proof,
// and that neither scheme interferes with the other.
func TestAnchor_Coexistence_V0AndV1(t *testing.T) {
	pool := anchorTestPool(t)
	defer pool.Close()

	tenantID := uuid.New()
	ctx := context.Background()

	// Clean any prior state for this tenant.
	_, _ = pool.Exec(ctx, `DELETE FROM policy_evaluations WHERE tenant_id = $1`, tenantID)
	_, _ = pool.Exec(ctx, `DELETE FROM merkle_epoch_leaves WHERE tenant_id = $1`, tenantID)
	_, _ = pool.Exec(ctx, `DELETE FROM merkle_roots WHERE tenant_id = $1`, tenantID)
	_, _ = pool.Exec(ctx, `DELETE FROM policies WHERE tenant_id = $1`, tenantID)
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM policy_evaluations WHERE tenant_id = $1`, tenantID)
		_, _ = pool.Exec(ctx, `DELETE FROM merkle_epoch_leaves WHERE tenant_id = $1`, tenantID)
		_, _ = pool.Exec(ctx, `DELETE FROM merkle_roots WHERE tenant_id = $1`, tenantID)
		_, _ = pool.Exec(ctx, `DELETE FROM policies WHERE tenant_id = $1`, tenantID)
	}()

	// policy_evaluations.policy_id has an FK to policies(id). Create
	// a minimal policy row so the constraint is satisfiable.
	policyID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO policies (
			id, tenant_id, statement,
			scope_domain, scope_system, actor, status,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, 'test', 'test', 'test-runner', 'active', NOW(), NOW())
	`, policyID, tenantID, "TEST POLICY:coexistence")
	if err != nil {
		t.Fatalf("insert policy: %v", err)
	}

	anchor := NewAnchor(pool)

	// ── Step 1: write a v1 evaluation via the production path ──
	v1ev := &Evaluation{
		ID:            uuid.New(),
		TenantID:      tenantID,
		WorkspaceID:   uuid.New(),
		PolicyID:      policyID,
		PolicyVersion: "v1",
		Decision:      DecisionWarn,
		Reason:        "v1 evaluation for coexistence test",
		SubjectKind:   "claim",
		SubjectID:     "subj-v1",
		Actor:         "test-runner",
	}
	height1, proof1Bytes, err := anchor.AppendEvaluation(ctx, v1ev)
	if err != nil {
		t.Fatalf("append v1: %v", err)
	}
	v1ev.MerkleBlockHeight = &height1

	if _, err := pool.Exec(ctx, `
		INSERT INTO policy_evaluations (
			id, tenant_id, workspace_id, policy_id, policy_version,
			decision, reason, evidence, evaluated_at,
			merkle_block_height, merkle_proof, verification_version,
			subject_kind, subject_id, actor
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,1,$12,$13,$14)
	`,
		v1ev.ID, v1ev.TenantID, v1ev.WorkspaceID, v1ev.PolicyID, v1ev.PolicyVersion,
		string(v1ev.Decision), v1ev.Reason, []byte("{}"), time.Now().UTC(),
		v1ev.MerkleBlockHeight, proof1Bytes,
		v1ev.SubjectKind, v1ev.SubjectID, v1ev.Actor,
	); err != nil {
		t.Fatalf("persist v1 evaluation: %v", err)
	}

	// Confirm the v1 proof parses as a V1Proof.
	v1Proof, err := store.UnmarshalV1Proof(proof1Bytes)
	if err != nil {
		t.Fatalf("v1 proof did not parse: %v", err)
	}
	if v1Proof.Version != 1 {
		t.Fatalf("v1 proof version = %d, want 1", v1Proof.Version)
	}

	// ── Step 2: construct a v0 evaluation by hand ──
	//
	// The v0 format is a JSON object with decision_hash, prev_root,
	// new_root, and block_height. No current code path produces it.
	// This test constructs it manually to simulate a row written by an
	// older version of Garuda.
	v0ev := &Evaluation{
		ID:            uuid.New(),
		TenantID:      tenantID,
		WorkspaceID:   uuid.New(),
		PolicyID:      policyID,
		PolicyVersion: "v1",
		Decision:      DecisionWarn,
		Reason:        "v0 evaluation for coexistence test",
		SubjectKind:   "claim",
		SubjectID:     "subj-v0",
		Actor:         "test-runner",
	}

	// After the v1 write, the tenant's merkle_roots row holds the v1
	// epoch root. A v0 row written next would chain from it.
	var currentRoot string
	if err := pool.QueryRow(ctx,
		`SELECT root_hash FROM merkle_roots WHERE tenant_id = $1`,
		tenantID,
	).Scan(&currentRoot); err != nil {
		t.Fatalf("read current root: %v", err)
	}

	// Compute the v0 leaf hash and chain update using the deprecated
	// encoder and ChainHash — exactly what the old code did.
	v0canonical, err := canonicalEvaluationBytes(v0ev)
	if err != nil {
		t.Fatalf("v0 canonical: %v", err)
	}
	v0leaf := sha256.Sum256(v0canonical)
	v0leafHex := hex.EncodeToString(v0leaf[:])
	v0newRoot := merkle.ChainHash(currentRoot, v0leafHex)
	v0blockHeight := height1 + 1

	v0Proof := map[string]any{
		"decision_hash": v0leafHex,
		"prev_root":     currentRoot,
		"new_root":      v0newRoot,
		"block_height":  v0blockHeight,
	}
	v0ProofBytes, err := json.Marshal(v0Proof)
	if err != nil {
		t.Fatalf("marshal v0 proof: %v", err)
	}

	// Sanity: the v0 proof must NOT parse as a V1Proof.
	if _, err := store.UnmarshalV1Proof(v0ProofBytes); err == nil {
		t.Fatal("v0 proof unexpectedly parsed as a v1 proof")
	}

	v0ev.MerkleBlockHeight = &v0blockHeight
	if _, err := pool.Exec(ctx, `
		INSERT INTO policy_evaluations (
			id, tenant_id, workspace_id, policy_id, policy_version,
			decision, reason, evidence, evaluated_at,
			merkle_block_height, merkle_proof, verification_version,
			subject_kind, subject_id, actor
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,0,$12,$13,$14)
	`,
		v0ev.ID, v0ev.TenantID, v0ev.WorkspaceID, v0ev.PolicyID, v0ev.PolicyVersion,
		string(v0ev.Decision), v0ev.Reason, []byte("{}"), time.Now().UTC(),
		v0ev.MerkleBlockHeight, v0ProofBytes,
		v0ev.SubjectKind, v0ev.SubjectID, v0ev.Actor,
	); err != nil {
		t.Fatalf("persist v0 evaluation: %v", err)
	}

	// ── Step 3: verify both — auto-detection must select the
	//            correct path from the stored proof alone ──
	ok, err := anchor.VerifyEvaluation(ctx, v1ev)
	if err != nil {
		t.Fatalf("verify v1: %v", err)
	}
	if !ok {
		t.Error("v1 evaluation did not verify")
	}

	ok, err = anchor.VerifyEvaluation(ctx, v0ev)
	if err != nil {
		t.Fatalf("verify v0: %v", err)
	}
	if !ok {
		t.Error("v0 evaluation did not verify")
	}

	// ── Step 4: tamper checks — a modified reason must invalidate
	//            both the v0 and v1 proofs ──
	v1tampered := *v1ev
	v1tampered.Reason = "tampered"
	ok, err = anchor.VerifyEvaluation(ctx, &v1tampered)
	if err != nil {
		t.Fatalf("verify tampered v1: %v", err)
	}
	if ok {
		t.Error("tampered v1 evaluation still verified")
	}

	v0tampered := *v0ev
	v0tampered.Reason = "tampered"
	ok, err = anchor.VerifyEvaluation(ctx, &v0tampered)
	if err != nil {
		t.Fatalf("verify tampered v0: %v", err)
	}
	if ok {
		t.Error("tampered v0 evaluation still verified")
	}
}
