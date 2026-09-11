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
	"github.com/myshra777-ai/garuda/internal/store"
)

// Anchor persists evaluations to the shared Merkle ledger.
//
// Every new evaluation is written as a v1 Merkle entry: a canonical
// leaf hash, an epoch root computed over one leaf per epoch, and a
// self-contained JSON inclusion proof stored alongside the row.
//
// v0 evaluations — those written before this commit — remain
// verifiable through the deprecated ChainHash path. VerifyEvaluation
// branches on the proof's version field.
type Anchor struct {
	pool *pgxpool.Pool
}

func NewAnchor(pool *pgxpool.Pool) *Anchor {
	return &Anchor{pool: pool}
}

// AppendEvaluation anchors a policy evaluation in the tenant's v1
// Merkle log and returns the epoch height plus the serialized proof.
//
// The proof is a JSON V1Proof: leaf hash, tier roots, parent root,
// epoch root, epoch height, and the inclusion proof path. It is
// self-contained — verification requires no database access.
func (a *Anchor) AppendEvaluation(ctx context.Context, ev *Evaluation) (int64, []byte, error) {
	leafHash := merkle.CanonicalEvaluation(evaluationInput(ev))

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("begin anchor tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	res, err := store.AppendLeafAndSealTx(ctx, tx, ev.TenantID, store.TierStatic, leafHash)
	if err != nil {
		return 0, nil, fmt.Errorf("anchor leaf: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, nil, fmt.Errorf("commit anchor: %w", err)
	}

	proofBytes, err := store.EncodeV1Proof(res).Marshal()
	if err != nil {
		return 0, nil, fmt.Errorf("marshal proof: %w", err)
	}

	return res.EpochHeight, proofBytes, nil
}

// evaluationInput constructs the canonical input for hashing an
// evaluation. See merkle.EvaluationInput for the field exclusion
// rationale.
func evaluationInput(ev *Evaluation) merkle.EvaluationInput {
	return merkle.EvaluationInput{
		PolicyID:          ev.PolicyID,
		PolicyVersion:     ev.PolicyVersion,
		Decision:          string(ev.Decision),
		Reason:            ev.Reason,
		ClaimIDs:          ev.Evidence.ClaimIDs,
		EntityIDs:         ev.Evidence.EntityIDs,
		ContradictionIDs:  ev.Evidence.ContradictionIDs,
		MatchedPredicates: ev.Evidence.MatchedPredicates,
		ReasoningNotes:    ev.Evidence.ReasoningNotes,
		SnapshotID:        ev.Evidence.SnapshotID,
		SubjectKind:       ev.SubjectKind,
		SubjectID:         ev.SubjectID,
		Actor:             ev.Actor,
	}
}

// VerifyEvaluation checks that an evaluation's inclusion proof is valid.
//
// Version 1 proofs use store.VerifyV1Proof (pure function, no chain
// replay, O(log n) hashes). Version 0 or absent proofs use the
// deprecated ChainHash path retained for historical rows.
func (a *Anchor) VerifyEvaluation(ctx context.Context, ev *Evaluation) (bool, error) {
	if ev.MerkleBlockHeight == nil {
		return false, fmt.Errorf("evaluation has no merkle anchor")
	}

	var storedProof []byte
	err := a.pool.QueryRow(ctx, `
		SELECT merkle_proof FROM policy_evaluations WHERE id = $1
	`, ev.ID).Scan(&storedProof)
	if err != nil {
		return false, fmt.Errorf("fetch proof: %w", err)
	}

	// Try v1 first. If the JSON parses as a V1Proof, the version is 1
	// and verification is a pure function of the proof.
	if v1, err := store.UnmarshalV1Proof(storedProof); err == nil {
		leafHash := merkle.CanonicalEvaluation(evaluationInput(ev))
		ok, err := store.VerifyV1Proof(leafHash, v1)
		if err != nil {
			return false, fmt.Errorf("v1 verify: %w", err)
		}
		if v1.EpochHeight != *ev.MerkleBlockHeight {
			return false, fmt.Errorf("block height mismatch: proof=%d eval=%d", v1.EpochHeight, *ev.MerkleBlockHeight)
		}
		return ok, nil
	}

	// Fall back to v0.
	return a.verifyV0(ev, storedProof)
}

// verifyV0 is the deprecated verification path for evaluations written
// before the v1 write path was adopted.
func (a *Anchor) verifyV0(ev *Evaluation, storedProof []byte) (bool, error) {
	canonical, err := canonicalEvaluationBytes(ev)
	if err != nil {
		return false, err
	}
	decisionHash := sha256.Sum256(canonical)

	var proof struct {
		DecisionHash string `json:"decision_hash"`
		PrevRoot     string `json:"prev_root"`
		NewRoot      string `json:"new_root"`
		BlockHeight  int64  `json:"block_height"`
	}
	if err := json.Unmarshal(storedProof, &proof); err != nil {
		return false, fmt.Errorf("parse proof: %w", err)
	}

	expectedNewRoot := merkle.ChainHash(proof.PrevRoot, fmt.Sprintf("%x", decisionHash))
	if expectedNewRoot != proof.NewRoot {
		return false, nil
	}
	if proof.BlockHeight != *ev.MerkleBlockHeight {
		return false, fmt.Errorf("block height mismatch: proof=%d eval=%d", proof.BlockHeight, *ev.MerkleBlockHeight)
	}
	return true, nil
}

// canonicalEvaluationBytes is the deprecated v0 encoder. Retained only
// for verifying historical proofs.
//
// Deprecated: use merkle.CanonicalEvaluation via evaluationInput.
func canonicalEvaluationBytes(ev *Evaluation) ([]byte, error) {
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
