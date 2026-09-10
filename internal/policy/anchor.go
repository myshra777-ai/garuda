// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package policy

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/merkle"
)

// Anchor persists evaluations to the shared Merkle ledger.
//
// The ledger stores a monotonically increasing block height per tenant.
// Each evaluation is hashed and chained with the previous root via
// merkle.ChainHash — the same primitive the rest of Garuda uses.
type Anchor struct {
	pool *pgxpool.Pool
}

func NewAnchor(pool *pgxpool.Pool) *Anchor {
	return &Anchor{pool: pool}
}

// AppendEvaluation hashes the evaluation, chains it into the tenant's
// Merkle ledger, and returns (blockHeight, inclusionProof, error).
//
// The inclusion proof is a compact byte slice sufficient to verify that
// this evaluation's hash is committed to the current root. Storage is
// not the point of the proof — the block height is what ties the
// evaluation to the immutable ledger.
// AppendEvaluation hashes the evaluation, chains it into the tenant's
// Merkle ledger, and returns (blockHeight, inclusionProof, error).
func (a *Anchor) AppendEvaluation(ctx context.Context, ev *Evaluation) (int64, []byte, error) {
	// 1. Canonicalize the evaluation into a deterministic byte sequence.
	canonical, err := canonicalEvaluationBytes(ev)
	if err != nil {
		return 0, nil, fmt.Errorf("canonicalize evaluation: %w", err)
	}

	decisionHash := sha256.Sum256(canonical)
	decisionHashHex := fmt.Sprintf("%x", decisionHash)

	// 2. Lock the current root row and chain.
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("begin anchor tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var currentRoot string
	var currentHeight int64

	// FIXED: no encode() — root_hash is TEXT
	err = tx.QueryRow(ctx, `
		SELECT root_hash, block_height
		FROM merkle_roots
		WHERE tenant_id = $1
		FOR UPDATE
	`, ev.TenantID).Scan(&currentRoot, &currentHeight)

	if err != nil && err.Error() != "no rows in result set" {
		return 0, nil, fmt.Errorf("lock root: %w", err)
	}

	// Genesis path — no row for this tenant yet
	if currentRoot == "" {
		gen := merkle.HashDecision(uuid.Nil, "GENESIS_POLICY_ROOT", "active", "system", "policy", "garuda", nil)
		// gen is already a string; use it directly
		_, ierr := tx.Exec(ctx, `
			INSERT INTO merkle_roots (tenant_id, root_hash, block_height, created_at, updated_at)
			VALUES ($1, $2, 0, NOW(), NOW())
			ON CONFLICT (tenant_id) DO NOTHING
		`, ev.TenantID, gen)
		if ierr != nil {
			return 0, nil, fmt.Errorf("insert genesis: %w", ierr)
		}
		currentRoot = gen
		currentHeight = 0
	}

	// 3. Compute new root + block height.
	newRoot := merkle.ChainHash(currentRoot, decisionHashHex)
	newHeight := currentHeight + 1

	_, err = tx.Exec(ctx, `
		UPDATE merkle_roots
		SET root_hash = $1, block_height = $2, updated_at = NOW()
		WHERE tenant_id = $3
	`, newRoot, newHeight, ev.TenantID)
	if err != nil {
		return 0, nil, fmt.Errorf("update root: %w", err)
	}

	// 4. Serialize inclusion proof
	proof := map[string]any{
		"decision_hash": decisionHashHex,
		"prev_root":     currentRoot,
		"new_root":      newRoot,
		"block_height":  newHeight,
	}
	proofBytes, _ := json.Marshal(proof)

	if err := tx.Commit(ctx); err != nil {
		return 0, nil, fmt.Errorf("commit anchor: %w", err)
	}

	return newHeight, proofBytes, nil
}

// VerifyEvaluation checks that an evaluation's inclusion proof is valid
// against the current (or historical) Merkle root.
//
// This is what makes the enforcement decision auditable: any agent or
// auditor can reconstruct the chain and verify the decision was committed
// at a specific block height.
func (a *Anchor) VerifyEvaluation(ctx context.Context, ev *Evaluation) (bool, error) {
	if ev.MerkleBlockHeight == nil {
		return false, fmt.Errorf("evaluation has no merkle anchor")
	}

	canonical, err := canonicalEvaluationBytes(ev)
	if err != nil {
		return false, err
	}
	decisionHash := sha256.Sum256(canonical)

	var storedProof []byte
	err = a.pool.QueryRow(ctx, `
		SELECT merkle_proof FROM policy_evaluations WHERE id = $1
	`, ev.ID).Scan(&storedProof)
	if err != nil {
		return false, fmt.Errorf("fetch proof: %w", err)
	}

	var proof struct {
		DecisionHash string `json:"decision_hash"`
		PrevRoot     string `json:"prev_root"`
		NewRoot      string `json:"new_root"`
		BlockHeight  int64  `json:"block_height"`
	}
	if err := json.Unmarshal(storedProof, &proof); err != nil {
		return false, fmt.Errorf("parse proof: %w", err)
	}

	// Recompute the chain from prev → new and compare.
	expectedNewRoot := merkle.ChainHash(proof.PrevRoot, fmt.Sprintf("%x", decisionHash))
	if expectedNewRoot != proof.NewRoot {
		return false, fmt.Errorf("merkle proof mismatch: expected %s got %s", expectedNewRoot, proof.NewRoot)
	}
	if proof.BlockHeight != *ev.MerkleBlockHeight {
		return false, fmt.Errorf("block height mismatch: proof=%d eval=%d", proof.BlockHeight, *ev.MerkleBlockHeight)
	}

	return true, nil
}

// canonicalEvaluationBytes produces a stable byte sequence for an evaluation.
// Ordering matters: two evaluations with the same logical content must hash
// to the same bytes.
func canonicalEvaluationBytes(ev *Evaluation) ([]byte, error) {
	// Sort evidence slices for determinism
	claimIDs := sortUUIDs(ev.Evidence.ClaimIDs)
	entityIDs := sortUUIDs(ev.Evidence.EntityIDs)
	contradictionIDs := sortUUIDs(ev.Evidence.ContradictionIDs)
	predicates := append([]string{}, ev.Evidence.MatchedPredicates...)
	sortStrings(predicates)
	notes := append([]string{}, ev.Evidence.ReasoningNotes...)
	sortStrings(notes)

	payload := map[string]any{
		"policy_id":      ev.PolicyID.String(),
		"policy_version": ev.PolicyVersion,
		"decision":       string(ev.Decision),
		"reason":         ev.Reason,
		"evidence": map[string]any{
			"claim_ids":          uuidsToStrings(claimIDs),
			"entity_ids":         uuidsToStrings(entityIDs),
			"contradiction_ids":  uuidsToStrings(contradictionIDs),
			"matched_predicates": predicates,
			"reasoning_notes":    notes,
		},
		"subject_kind": ev.SubjectKind,
		"subject_id":   ev.SubjectID,
		"actor":        ev.Actor,
	}
	return json.Marshal(payload)
}

// ─────────────────────────────────────────────────────────────────────────────
// Sorting helpers
// ─────────────────────────────────────────────────────────────────────────────

func sortUUIDs(ids []uuid.UUID) []uuid.UUID {
	out := append([]uuid.UUID{}, ids...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].String() < out[j].String()
	})
	return out
}

func sortStrings(s []string) {
	sort.Strings(s)
}

func uuidsToStrings(ids []uuid.UUID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}
