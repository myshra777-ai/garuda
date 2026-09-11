// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package merkle

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/google/uuid"
)

func baseEvalInput() EvaluationInput {
	policyID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	return EvaluationInput{
		PolicyID:          policyID,
		PolicyVersion:     "v1",
		Decision:          "WARN",
		Reason:            "test reason",
		ClaimIDs:          []uuid.UUID{uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")},
		EntityIDs:         []uuid.UUID{uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")},
		ContradictionIDs:  []uuid.UUID{uuid.MustParse("cccccccc-0000-0000-0000-000000000001")},
		MatchedPredicates: []string{"predicate_a"},
		ReasoningNotes:    []string{"note_a"},
		SubjectKind:       "claim",
		SubjectID:         "subj-1",
		Actor:             "cli-operator",
	}
}

func TestCanonicalEvaluation_Deterministic(t *testing.T) {
	in := baseEvalInput()
	a := CanonicalEvaluation(in)
	b := CanonicalEvaluation(in)
	if !bytes.Equal(a, b) {
		t.Fatal("evaluation hash not deterministic")
	}
}

func TestCanonicalEvaluation_Length(t *testing.T) {
	h := CanonicalEvaluation(baseEvalInput())
	if len(h) != 32 {
		t.Fatalf("hash length = %d, want 32", len(h))
	}
}

func TestCanonicalEvaluation_FieldSensitivity(t *testing.T) {
	base := CanonicalEvaluation(baseEvalInput())

	cases := []struct {
		name string
		mut  func(*EvaluationInput)
	}{
		{"policy_id", func(in *EvaluationInput) { in.PolicyID = uuid.New() }},
		{"policy_version", func(in *EvaluationInput) { in.PolicyVersion = "v2" }},
		{"decision", func(in *EvaluationInput) { in.Decision = "BLOCK" }},
		{"reason", func(in *EvaluationInput) { in.Reason = "different" }},
		{"subject_kind", func(in *EvaluationInput) { in.SubjectKind = "entity" }},
		{"subject_id", func(in *EvaluationInput) { in.SubjectID = "subj-2" }},
		{"actor", func(in *EvaluationInput) { in.Actor = "someone-else" }},
		{"claim_ids", func(in *EvaluationInput) {
			in.ClaimIDs = []uuid.UUID{uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000002")}
		}},
		{"entity_ids", func(in *EvaluationInput) {
			in.EntityIDs = []uuid.UUID{uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")}
		}},
		{"contradiction_ids", func(in *EvaluationInput) {
			in.ContradictionIDs = []uuid.UUID{uuid.MustParse("cccccccc-0000-0000-0000-000000000002")}
		}},
		{"matched_predicates", func(in *EvaluationInput) {
			in.MatchedPredicates = []string{"predicate_b"}
		}},
		{"reasoning_notes", func(in *EvaluationInput) {
			in.ReasoningNotes = []string{"note_b"}
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := baseEvalInput()
			tc.mut(&in)
			h := CanonicalEvaluation(in)
			if bytes.Equal(base, h) {
				t.Errorf("changing %s did not change hash", tc.name)
			}
		})
	}
}

func TestCanonicalEvaluation_SnapshotIDOptionality(t *testing.T) {
	inA := baseEvalInput()
	inB := baseEvalInput()
	snapID := uuid.MustParse("99999999-0000-0000-0000-000000000001")
	inB.SnapshotID = &snapID

	a := CanonicalEvaluation(inA)
	b := CanonicalEvaluation(inB)
	if bytes.Equal(a, b) {
		t.Error("nil snapshot_id and set snapshot_id produced same hash")
	}

	// Two nils → same hash
	c := CanonicalEvaluation(inA)
	if !bytes.Equal(a, c) {
		t.Error("determinism broke for nil snapshot_id")
	}
}

func TestCanonicalEvaluation_EvidenceOrderIndependent(t *testing.T) {
	idsA := []uuid.UUID{
		uuid.MustParse("33333333-0000-0000-0000-000000000003"),
		uuid.MustParse("11111111-0000-0000-0000-000000000001"),
		uuid.MustParse("22222222-0000-0000-0000-000000000002"),
	}
	idsB := []uuid.UUID{
		uuid.MustParse("11111111-0000-0000-0000-000000000001"),
		uuid.MustParse("22222222-0000-0000-0000-000000000002"),
		uuid.MustParse("33333333-0000-0000-0000-000000000003"),
	}

	inA := baseEvalInput()
	inA.ClaimIDs = idsA
	inB := baseEvalInput()
	inB.ClaimIDs = idsB

	a := CanonicalEvaluation(inA)
	b := CanonicalEvaluation(inB)
	if !bytes.Equal(a, b) {
		t.Error("claim_id order affected hash")
	}
}

func TestCanonicalEvaluation_StringSliceOrderIndependent(t *testing.T) {
	inA := baseEvalInput()
	inA.MatchedPredicates = []string{"gamma", "alpha", "beta"}
	inB := baseEvalInput()
	inB.MatchedPredicates = []string{"alpha", "beta", "gamma"}

	a := CanonicalEvaluation(inA)
	b := CanonicalEvaluation(inB)
	if !bytes.Equal(a, b) {
		t.Error("matched_predicates order affected hash")
	}
}

func TestCanonicalEvaluation_LengthPrefixDisambiguates(t *testing.T) {
	inA := baseEvalInput()
	inA.Reason = "ab"
	inA.SubjectID = "c"

	inB := baseEvalInput()
	inB.Reason = "a"
	inB.SubjectID = "bc"

	a := CanonicalEvaluation(inA)
	b := CanonicalEvaluation(inB)
	if bytes.Equal(a, b) {
		t.Error("length prefixing failed: [\"ab\",\"c\"] collided with [\"a\",\"bc\"]")
	}
}

// TestCanonicalEvaluation_PrintVectors logs evaluation hashes for fixed
// inputs. Run with -v and freeze into
// testdata/evaluation_vectors_v1.json in a follow-up commit.
func TestCanonicalEvaluation_PrintVectors(t *testing.T) {
	policyID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	snapID := uuid.MustParse("99999999-0000-0000-0000-000000000001")
	claimA := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	claimB := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000002")
	entityA := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")

	cases := []struct {
		name string
		in   EvaluationInput
	}{
		{
			name: "empty",
			in:   EvaluationInput{},
		},
		{
			name: "minimal",
			in: EvaluationInput{
				PolicyID:  policyID,
				Decision:  "ALLOW",
				Reason:    "no predicate matched",
				SubjectID: "subj-1",
			},
		},
		{
			name: "full_no_snapshot",
			in: EvaluationInput{
				PolicyID:          policyID,
				PolicyVersion:     "v1",
				Decision:          "WARN",
				Reason:            "runtime evidence missing",
				ClaimIDs:          []uuid.UUID{claimA, claimB},
				EntityIDs:         []uuid.UUID{entityA},
				MatchedPredicates: []string{"verification_missing"},
				ReasoningNotes:    []string{"checked 5 entities"},
				SubjectKind:       "claim",
				SubjectID:         "subj-1",
				Actor:             "cli-operator",
			},
		},
		{
			name: "full_with_snapshot",
			in: EvaluationInput{
				PolicyID:          policyID,
				PolicyVersion:     "v1",
				Decision:          "BLOCK",
				Reason:            "contradiction detected",
				ClaimIDs:          []uuid.UUID{claimA},
				ContradictionIDs:  []uuid.UUID{claimB},
				MatchedPredicates: []string{"contradiction_exists"},
				SnapshotID:        &snapID,
				SubjectKind:       "claim",
				SubjectID:         "subj-1",
				Actor:             "cli-operator",
			},
		},
		{
			name: "unicode_reason",
			in: EvaluationInput{
				PolicyID:  policyID,
				Decision:  "WARN",
				Reason:    "日本語の理由",
				SubjectID: "subj-1",
			},
		},
	}

	for _, tc := range cases {
		hexHash := hex.EncodeToString(CanonicalEvaluation(tc.in))
		t.Logf("%-20s %s", tc.name, hexHash)
	}
}
