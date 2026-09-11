// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/myshra777-ai/garuda/internal/merkle"
)

// Tier constants matching the merkle_epoch_leaves schema CHECK constraint.
const (
	TierStatic  = 0
	TierRuntime = 1
)

// V1WriteResult is the output of AppendLeafAndSeal. It is both the
// return value for the caller and the source material for the JSON
// proof that gets persisted alongside the leaf.
type V1WriteResult struct {
	EpochHeight   int64
	Tier          int
	LeafIndex     int
	LeafHash      []byte
	EpochRoot     []byte
	ProofPath     []merkle.ProofNode
	TierRoot      []byte
	OtherTierRoot []byte
	ParentRoot    []byte
}

// AppendLeafAndSeal writes one leaf as its own epoch. It is the atomic
// write primitive for v1 Merkle entries.
//
// Every call:
//  1. Locks the tenant's merkle_roots row.
//  2. Increments block_height.
//  3. Builds a single-leaf tree for the given tier, and an empty tree
//     for the other tier.
//  4. Computes the epoch root via merkle.EpochRootFromTiers.
//  5. Persists the leaf to merkle_epoch_leaves.
//  6. Updates merkle_roots with the new epoch root and version=1.
//
// Batching multiple leaves into one epoch is a schema-compatible future
// upgrade: merkle_epoch_leaves already supports it, and the proof format
// does not change. This function intentionally writes one leaf per
// epoch so that every current call site can adopt v1 without changing
// its write cadence.
//
// The caller is responsible for supplying a 32-byte leafHash. If
// len(leafHash) != 32, the call returns an error without touching the
// database.
// AppendLeafAndSeal opens a transaction and delegates to the internal
// helper. Use this when the leaf is the only thing being written.
// When the leaf must be committed atomically with another row (e.g., a
// decision), use appendLeafAndSealTx directly with an outer transaction.
func (s *PostgresStore) AppendLeafAndSeal(ctx context.Context, tenantID uuid.UUID, tier int, leafHash []byte) (*V1WriteResult, error) {
	if len(leafHash) != 32 {
		return nil, fmt.Errorf("append leaf: hash must be 32 bytes, got %d", len(leafHash))
	}
	if tier != TierStatic && tier != TierRuntime {
		return nil, fmt.Errorf("append leaf: invalid tier %d", tier)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin v1 write tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	res, err := appendLeafAndSealTx(ctx, tx, tenantID, tier, leafHash)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit v1 write: %w", err)
	}
	return res, nil
}

// appendLeafAndSealTx is the transaction-scoped core of AppendLeafAndSeal.
// It does not begin or commit — the caller owns the transaction. Used by
// SaveDecision and other writers that must anchor a leaf atomically
// alongside their own row insert.
//
// On return, the caller may inspect the leaf row (in merkle_epoch_leaves),
// the updated merkle_roots row, and res. All are visible within the
// caller's transaction and committed or rolled back with it.
func appendLeafAndSealTx(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, tier int, leafHash []byte) (*V1WriteResult, error) {
	if len(leafHash) != 32 {
		return nil, fmt.Errorf("append leaf: hash must be 32 bytes, got %d", len(leafHash))
	}
	if tier != TierStatic && tier != TierRuntime {
		return nil, fmt.Errorf("append leaf: invalid tier %d", tier)
	}

	var currentRoot string
	var currentHeight int64
	err := tx.QueryRow(ctx, `
		SELECT root_hash, block_height
		FROM merkle_roots
		WHERE tenant_id = $1
		FOR UPDATE
	`, tenantID).Scan(&currentRoot, &currentHeight)

	if errors.Is(err, pgx.ErrNoRows) {
		gen := merkle.GenesisRootHex(tenantID)
		_, err = tx.Exec(ctx, `
			INSERT INTO merkle_roots (tenant_id, root_hash, block_height, verification_version, created_at, updated_at)
			VALUES ($1, $2, 0, 1, NOW(), NOW())
		`, tenantID, gen)
		if err != nil {
			return nil, fmt.Errorf("insert v1 genesis: %w", err)
		}
		currentRoot = gen
		currentHeight = 0
	} else if err != nil {
		return nil, fmt.Errorf("lock merkle root: %w", err)
	}

	parentBytes, err := hex.DecodeString(currentRoot)
	if err != nil || len(parentBytes) != 32 {
		return nil, fmt.Errorf("parent root is not 32-byte hex: %v", err)
	}

	newHeight := currentHeight + 1

	var staticRoot, runtimeRoot []byte
	if tier == TierStatic {
		staticRoot = merkle.BuildRoot([][]byte{leafHash})
		runtimeRoot = merkle.BuildRoot(nil)
	} else {
		staticRoot = merkle.BuildRoot(nil)
		runtimeRoot = merkle.BuildRoot([][]byte{leafHash})
	}
	proofPath, err := merkle.BuildProof([][]byte{leafHash}, 0)
	if err != nil {
		return nil, fmt.Errorf("build v1 proof: %w", err)
	}

	epochRoot := merkle.EpochRootFromTiers(staticRoot, runtimeRoot, parentBytes, uint64(newHeight))

	leafHex := hex.EncodeToString(leafHash)
	_, err = tx.Exec(ctx, `
		INSERT INTO merkle_epoch_leaves (tenant_id, epoch_height, tier, leaf_index, leaf_hash)
		VALUES ($1, $2, $3, 0, $4)
	`, tenantID, newHeight, tier, leafHex)
	if err != nil {
		return nil, fmt.Errorf("insert epoch leaf: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE merkle_roots
		SET root_hash = $1, block_height = $2, updated_at = NOW(), verification_version = 1
		WHERE tenant_id = $3
	`, hex.EncodeToString(epochRoot), newHeight, tenantID)
	if err != nil {
		return nil, fmt.Errorf("update merkle root: %w", err)
	}

	return &V1WriteResult{
		EpochHeight:   newHeight,
		Tier:          tier,
		LeafIndex:     0,
		LeafHash:      leafHash,
		EpochRoot:     epochRoot,
		ProofPath:     proofPath,
		TierRoot:      staticRoot,
		OtherTierRoot: runtimeRoot,
		ParentRoot:    parentBytes,
	}, nil
}

// V1Proof is the JSON-serializable form of a V1WriteResult.
// It is what gets stored in merkle_proof columns and what a verifier
// outside Garuda's process would receive.
//
// All hash fields are lowercase hex. The proof path position is
// encoded as "L" or "R" for language neutrality.
type V1Proof struct {
	Version       int           `json:"version"`
	EpochHeight   int64         `json:"epoch_height"`
	Tier          int           `json:"tier"`
	LeafIndex     int           `json:"leaf_index"`
	LeafHash      string        `json:"leaf_hash"`
	ProofPath     []V1ProofNode `json:"proof_path"`
	TierRoot      string        `json:"tier_root"`
	OtherTierRoot string        `json:"other_tier_root"`
	ParentRoot    string        `json:"parent_root"`
	EpochRoot     string        `json:"epoch_root"`
}

// V1ProofNode is one step in the inclusion proof path.
type V1ProofNode struct {
	Position string `json:"position"` // "L" or "R"
	Hash     string `json:"hash"`     // hex, 32 bytes
}

// EncodeV1Proof converts a V1WriteResult into its serializable form.
func EncodeV1Proof(r *V1WriteResult) *V1Proof {
	if r == nil {
		return nil
	}
	path := make([]V1ProofNode, len(r.ProofPath))
	for i, pn := range r.ProofPath {
		pos := "L"
		if pn.Position == merkle.SiblingRight {
			pos = "R"
		}
		path[i] = V1ProofNode{
			Position: pos,
			Hash:     hex.EncodeToString(pn.Hash),
		}
	}
	return &V1Proof{
		Version:       1,
		EpochHeight:   r.EpochHeight,
		Tier:          r.Tier,
		LeafIndex:     r.LeafIndex,
		LeafHash:      hex.EncodeToString(r.LeafHash),
		ProofPath:     path,
		TierRoot:      hex.EncodeToString(r.TierRoot),
		OtherTierRoot: hex.EncodeToString(r.OtherTierRoot),
		ParentRoot:    hex.EncodeToString(r.ParentRoot),
		EpochRoot:     hex.EncodeToString(r.EpochRoot),
	}
}

// MarshalJSON produces compact JSON for storage in merkle_proof columns.
func (p *V1Proof) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// UnmarshalV1Proof parses the JSON form back into a V1Proof.
func UnmarshalV1Proof(data []byte) (*V1Proof, error) {
	var p V1Proof
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	if p.Version != 1 {
		return nil, fmt.Errorf("unsupported proof version %d", p.Version)
	}
	return &p, nil
}

// VerifyV1Proof checks that a leaf is included in an epoch.
//
// Two-step verification:
//  1. VerifyInclusion(leafHash, proofPath, tierRoot) must be true.
//  2. EpochRootFromTiers(tierRoot, otherTierRoot, parentRoot, height)
//     must equal the claimed epochRoot.
//
// Both steps are pure functions of the proof's own data. No database
// access. No chain replay. O(log n) hashes for step 1, O(1) for step 2.
func VerifyV1Proof(leafHash []byte, proof *V1Proof) (bool, error) {
	if proof == nil {
		return false, fmt.Errorf("nil proof")
	}
	if proof.Version != 1 {
		return false, fmt.Errorf("unsupported proof version %d", proof.Version)
	}

	// Convert hex fields to bytes.
	tierRoot, err := hex.DecodeString(proof.TierRoot)
	if err != nil || len(tierRoot) != 32 {
		return false, fmt.Errorf("tier_root: %v", err)
	}
	otherTierRoot, err := hex.DecodeString(proof.OtherTierRoot)
	if err != nil || len(otherTierRoot) != 32 {
		return false, fmt.Errorf("other_tier_root: %v", err)
	}
	parentRoot, err := hex.DecodeString(proof.ParentRoot)
	if err != nil || len(parentRoot) != 32 {
		return false, fmt.Errorf("parent_root: %v", err)
	}
	epochRoot, err := hex.DecodeString(proof.EpochRoot)
	if err != nil || len(epochRoot) != 32 {
		return false, fmt.Errorf("epoch_root: %v", err)
	}

	// Also verify the caller's leafHash matches what the proof claims.
	claimedLeaf, err := hex.DecodeString(proof.LeafHash)
	if err != nil || !bytes.Equal(claimedLeaf, leafHash) {
		return false, nil
	}

	// Step 1: leaf is in its tier.
	proofNodes := make([]merkle.ProofNode, len(proof.ProofPath))
	for i, pn := range proof.ProofPath {
		h, err := hex.DecodeString(pn.Hash)
		if err != nil || len(h) != 32 {
			return false, fmt.Errorf("proof path[%d] hash: %v", i, err)
		}
		var pos byte
		switch pn.Position {
		case "L":
			pos = merkle.SiblingLeft
		case "R":
			pos = merkle.SiblingRight
		default:
			return false, fmt.Errorf("proof path[%d] unknown position %q", i, pn.Position)
		}
		proofNodes[i] = merkle.ProofNode{Position: pos, Hash: h}
	}
	if !merkle.VerifyInclusion(leafHash, proofNodes, tierRoot) {
		return false, nil
	}

	// Step 2: tier roots combine to the claimed epoch root.
	var staticRoot, runtimeRoot []byte
	if proof.Tier == TierStatic {
		staticRoot = tierRoot
		runtimeRoot = otherTierRoot
	} else {
		staticRoot = otherTierRoot
		runtimeRoot = tierRoot
	}
	recomputed := merkle.EpochRootFromTiers(staticRoot, runtimeRoot, parentRoot, uint64(proof.EpochHeight))
	if !bytes.Equal(recomputed, epochRoot) {
		return false, nil
	}

	return true, nil
}

// TimeFromEpochHeight is a helper for callers that want to log the
// write time. The height is a monotonic counter, not a timestamp.
// This function exists to make that distinction explicit.
func TimeFromEpochHeight(_ int64) time.Time {
	return time.Time{}
}
