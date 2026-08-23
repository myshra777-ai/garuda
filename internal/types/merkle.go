// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"time"

	"github.com/google/uuid"
)

// MerkleProof contains inclusion proof path to cryptographically verify a decision against the root.
type MerkleProof struct {
	DecisionID  uuid.UUID `json:"decision_id"`
	LeafHash    string    `json:"leaf_hash"`
	ParentHash  string    `json:"parent_hash"`
	RootHash    string    `json:"root_hash"`
	BlockHeight int64     `json:"block_height"`
	TenantID    uuid.UUID `json:"tenant_id"`
	ProofPath   []string  `json:"proof_path,omitempty"`
	IsVerified  bool      `json:"is_verified"`
	CreatedAt   time.Time `json:"created_at"`
}

// MerkleRoot represents current root hash and block height state for a tenant.
type MerkleRoot struct {
	TenantID    uuid.UUID `json:"tenant_id"`
	RootHash    string    `json:"root_hash"`
	BlockHeight int64     `json:"block_height"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EvidenceBlock represents a single hashed block in an immutable evidence chain.
type EvidenceBlock struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	DecisionID   uuid.UUID `json:"decision_id"`
	PrevHash     string    `json:"prev_hash"`
	EvidenceHash string    `json:"evidence_hash"`
	Payload      any       `json:"payload"`
	CreatedAt    time.Time `json:"created_at"`
}

// MerkleSnapshot represents a historical snapshot of a tenant's Merkle root.
type MerkleSnapshot struct {
	ID                      uuid.UUID  `json:"id"`
	TenantID                uuid.UUID  `json:"tenant_id"`
	ParentSnapshotID        *uuid.UUID `json:"parent_snapshot_id,omitempty"`
	BlockHeight             int64      `json:"block_height"`
	SnapshotHash            string     `json:"snapshot_hash"`
	RootHash                string     `json:"root_hash,omitempty"`
	StaticRootHash          string     `json:"static_root_hash"`
	RuntimeRootHash         string     `json:"runtime_root_hash"`
	RuntimeLeafCount        int64      `json:"runtime_leaf_count"`
	VerifiedClaimsCount     int64      `json:"verified_claims_count"`
	ContradictedClaimsCount int64      `json:"contradicted_claims_count"`
	EpochTimestamp          int64      `json:"epoch_timestamp,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
}
