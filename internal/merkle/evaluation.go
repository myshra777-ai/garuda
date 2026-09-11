// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package merkle

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
)

// EvaluationInput is the canonical-form input for hashing a policy
// evaluation.
//
// It carries only the fields that belong in the hash. Fields that are
// set after hashing (MerkleBlockHeight, MerkleProof) or that describe
// the container rather than the content (ID, TenantID, WorkspaceID,
// EvaluatedAt) are deliberately excluded.
//
// See the field-ID comment in canonical.go for the exclusion rationale.
type EvaluationInput struct {
	PolicyID          uuid.UUID
	PolicyVersion     string
	Decision          string
	Reason            string
	ClaimIDs          []uuid.UUID
	EntityIDs         []uuid.UUID
	ContradictionIDs  []uuid.UUID
	MatchedPredicates []string
	ReasoningNotes    []string
	SnapshotID        *uuid.UUID
	SubjectKind       string
	SubjectID         string
	Actor             string
}

// CanonicalEvaluation encodes a policy evaluation into canonical bytes
// and returns the 32-byte SHA-256 digest.
//
// Replaces the ad-hoc json.Marshal(map[string]any) construction that
// lived in internal/policy/anchor.go's canonicalEvaluationBytes. That
// path depended on Go's map ordering, which is an implementation detail
// of encoding/json, not a specification.
//
// Slice fields (ClaimIDs, EntityIDs, ContradictionIDs, MatchedPredicates,
// ReasoningNotes) are sorted before encoding, so the caller does not
// need to pre-sort.
//
// SnapshotID is optional. If nil, the field is omitted from the
// encoding entirely. If present, it is written at field ID 0x29. The
// two cases produce different hashes — that is correct.
//
// See ADR-0002 D6 for the analogous decision-hash rule.
func CanonicalEvaluation(in EvaluationInput) []byte {
	enc := NewEncoder()
	enc.UUID(fieldEvalPolicyID, in.PolicyID)
	enc.String(fieldEvalPolicyVersion, in.PolicyVersion)
	enc.String(fieldEvalDecision, in.Decision)
	enc.String(fieldEvalReason, in.Reason)
	enc.UUIDSlice(fieldEvalClaimIDs, in.ClaimIDs)
	enc.UUIDSlice(fieldEvalEntityIDs, in.EntityIDs)
	enc.UUIDSlice(fieldEvalContradictionIDs, in.ContradictionIDs)
	enc.StringSlice(fieldEvalMatchedPredicates, in.MatchedPredicates)
	enc.StringSlice(fieldEvalReasoningNotes, in.ReasoningNotes)
	if in.SnapshotID != nil {
		enc.UUID(fieldEvalSnapshotID, *in.SnapshotID)
	}
	enc.String(fieldEvalSubjectKind, in.SubjectKind)
	enc.String(fieldEvalSubjectID, in.SubjectID)
	enc.String(fieldEvalActor, in.Actor)

	out, err := enc.Finish()
	if err != nil {
		// Same reasoning as GenesisRoot and CanonicalDecision: an error
		// here means a coding bug, not a runtime condition. A wrong
		// evaluation hash silently recorded would corrupt the audit log.
		panic("merkle: evaluation encoding failed: " + err.Error())
	}

	h := sha256.Sum256(out)
	return h[:]
}

// CanonicalEvaluationHex returns CanonicalEvaluation as lowercase hex.
// Convenience wrapper for callers that store hashes as text.
func CanonicalEvaluationHex(in EvaluationInput) string {
	return hex.EncodeToString(CanonicalEvaluation(in))
}
