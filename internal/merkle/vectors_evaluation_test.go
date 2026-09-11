// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package merkle

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

// TestEvaluationVectorsV1 asserts that CanonicalEvaluation reproduces
// every frozen vector byte-for-byte.
//
// A failure here means the evaluation encoding has changed. That is a
// breaking change: every evaluation hash ever recorded becomes
// unverifiable under the new rule. The fix is NOT to update the vector
// file — it is to revert the change, or to bump CanonicalVersionV1 and
// open a new ADR.
func TestEvaluationVectorsV1(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "evaluation_vectors_v1.json"))
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}

	var file struct {
		Vectors []struct {
			Name  string `json:"name"`
			Input struct {
				PolicyID          string   `json:"policy_id"`
				PolicyVersion     string   `json:"policy_version"`
				Decision          string   `json:"decision"`
				Reason            string   `json:"reason"`
				ClaimIDs          []string `json:"claim_ids"`
				EntityIDs         []string `json:"entity_ids"`
				ContradictionIDs  []string `json:"contradiction_ids"`
				MatchedPredicates []string `json:"matched_predicates"`
				ReasoningNotes    []string `json:"reasoning_notes"`
				SnapshotID        *string  `json:"snapshot_id"`
				SubjectKind       string   `json:"subject_kind"`
				SubjectID         string   `json:"subject_id"`
				Actor             string   `json:"actor"`
			} `json:"input"`
			HashHex string `json:"hash_hex"`
		} `json:"vectors"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("parse vectors: %v", err)
	}

	parseUUIDs := func(ss []string) []uuid.UUID {
		out := make([]uuid.UUID, len(ss))
		for i, s := range ss {
			out[i] = uuid.MustParse(s)
		}
		return out
	}

	for _, v := range file.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			var snapPtr *uuid.UUID
			if v.Input.SnapshotID != nil {
				s := uuid.MustParse(*v.Input.SnapshotID)
				snapPtr = &s
			}

			in := EvaluationInput{
				PolicyID:          uuid.MustParse(v.Input.PolicyID),
				PolicyVersion:     v.Input.PolicyVersion,
				Decision:          v.Input.Decision,
				Reason:            v.Input.Reason,
				ClaimIDs:          parseUUIDs(v.Input.ClaimIDs),
				EntityIDs:         parseUUIDs(v.Input.EntityIDs),
				ContradictionIDs:  parseUUIDs(v.Input.ContradictionIDs),
				MatchedPredicates: v.Input.MatchedPredicates,
				ReasoningNotes:    v.Input.ReasoningNotes,
				SnapshotID:        snapPtr,
				SubjectKind:       v.Input.SubjectKind,
				SubjectID:         v.Input.SubjectID,
				Actor:             v.Input.Actor,
			}

			got := hex.EncodeToString(CanonicalEvaluation(in))
			if got != v.HashHex {
				t.Errorf("evaluation hash mismatch for %q\n  got  = %s\n  want = %s",
					v.Name, got, v.HashHex)
			}
		})
	}
}
