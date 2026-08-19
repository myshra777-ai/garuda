// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// generateFingerprint creates a SHA‑256 hash of the canonical representation
func generateFingerprint(result *Result) string {
	// Sort entities and relationships for deterministic output
	sort.Slice(result.Entities, func(i, j int) bool {
		return result.Entities[i].ID < result.Entities[j].ID
	})
	sort.Slice(result.Relationships, func(i, j int) bool {
		return result.Relationships[i].From < result.Relationships[j].From
	})

	// Build a simplified canonical model
	canonical := struct {
		Entities      []Entity       `json:"entities"`
		Relationships []Relationship `json:"relationships"`
		Stats         Stats          `json:"stats"`
	}{
		Entities:      result.Entities,
		Relationships: result.Relationships,
		Stats:         result.Stats,
	}

	data, _ := json.Marshal(canonical)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
