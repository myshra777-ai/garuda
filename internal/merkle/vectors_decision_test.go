// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package merkle

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

// TestDecisionVectorsV1 asserts that CanonicalDecision reproduces every
// frozen vector byte-for-byte.
//
// A failure here means the decision encoding has changed. That is a
// breaking change: every decision hash ever recorded becomes
// unverifiable under the new rule. The fix is NOT to update the vector
// file — it is to revert the change, or to bump CanonicalVersionV1 and
// open a new ADR.
func TestDecisionVectorsV1(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "decision_vectors_v1.json"))
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}

	var file struct {
		Vectors []struct {
			Name        string   `json:"name"`
			ID          string   `json:"id"`
			Title       string   `json:"title"`
			Status      string   `json:"status"`
			ScopeDomain string   `json:"scope_domain"`
			ScopeSystem string   `json:"scope_system"`
			Owner       string   `json:"owner"`
			EvidenceIDs []string `json:"evidence_ids"`
			HashHex     string   `json:"hash_hex"`
		} `json:"vectors"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("parse vectors: %v", err)
	}

	for _, v := range file.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			id := uuid.MustParse(v.ID)
			got := hex.EncodeToString(CanonicalDecision(
				id, v.Title, v.Status, v.ScopeDomain, v.ScopeSystem, v.Owner, v.EvidenceIDs,
			))
			if got != v.HashHex {
				t.Errorf("decision hash mismatch for %s\n  got  = %s\n  want = %s",
					v.Name, got, v.HashHex)
			}
		})
	}
}
