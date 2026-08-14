package canonical

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/myshra777-ai/garuda/internal/types"
)

// DecisionContent represents the immutable semantic content of a decision.
// It excludes metadata like actor, timestamp, revision_id.
type DecisionContent struct {
	Title       string               `json:"title"`
	Statement   string               `json:"statement"`
	Scope       types.Scope          `json:"scope"`
	Owner       string               `json:"owner"`
	Confidence  float64              `json:"confidence"`
	EvidenceIDs []types.EvidenceHash `json:"evidence_ids,omitempty"`
	ParentID    *types.DecisionID    `json:"parent_id,omitempty"`
	// Future: subject, predicate, object for semantic claims
}

// CanonicalJSON returns a stable JSON representation.
// Fields are sorted deterministically to ensure consistent hashing.
func (c DecisionContent) CanonicalJSON() ([]byte, error) {
	// Use a struct with field order that will be stable across marshaling
	// We marshal into a map with sorted keys for determinism.
	m := map[string]interface{}{
		"title":      c.Title,
		"statement":  c.Statement,
		"scope":      c.Scope,
		"owner":      c.Owner,
		"confidence": c.Confidence,
	}
	if len(c.EvidenceIDs) > 0 {
		m["evidence_ids"] = c.EvidenceIDs
	}
	if c.ParentID != nil {
		m["parent_id"] = c.ParentID
	}
	return json.Marshal(m)
}

// Hash returns SHA-256 of the canonical JSON.
func (c DecisionContent) Hash() ([32]byte, error) {
	data, err := c.CanonicalJSON()
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(data), nil
}

// MustHash panics if hashing fails (for tests/where error is impossible).
func (c DecisionContent) MustHash() [32]byte {
	h, err := c.Hash()
	if err != nil {
		panic(fmt.Sprintf("failed to hash decision content: %v", err))
	}
	return h
}
