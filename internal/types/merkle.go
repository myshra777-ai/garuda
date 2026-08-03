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
