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

// CanonicalDecision encodes a decision into its canonical byte form and
// returns the 32-byte SHA-256 digest.
//
// This replaces the deprecated HashDecision, which relied on
// encoding/json map ordering — an implementation detail of Go's stdlib,
// not a specification. CanonicalDecision uses the versioned encoder in
// canonical.go, so two independent implementations in different
// languages will produce byte-identical output for the same logical
// decision.
//
// Field order is fixed (see fieldDecisionID..fieldDecisionEvidence in
// canonical.go). Evidence IDs are sorted before encoding, so the caller
// does not need to pre-sort them.
//
// See ADR-0002 D6.
func CanonicalDecision(
	decisionID uuid.UUID,
	title, status, scopeDomain, scopeSystem, owner string,
	evidenceIDs []string,
) []byte {
	enc := NewEncoder()
	enc.UUID(fieldDecisionID, decisionID)
	enc.String(fieldDecisionTitle, title)
	enc.String(fieldDecisionStatus, status)
	enc.String(fieldDecisionScopeDo, scopeDomain)
	enc.String(fieldDecisionScopeSys, scopeSystem)
	enc.String(fieldDecisionOwner, owner)
	enc.StringSlice(fieldDecisionEvidence, evidenceIDs)

	out, err := enc.Finish()
	if err != nil {
		// Same reasoning as GenesisRoot: an error here means a coding
		// bug, not a runtime condition. Panicking is correct because a
		// wrong decision hash silently recorded would corrupt the
		// audit log.
		panic("merkle: canonical decision encoding failed: " + err.Error())
	}

	h := sha256.Sum256(out)
	return h[:]
}

// CanonicalDecisionHex returns CanonicalDecision as a lowercase hex
// string. Convenience wrapper for callers that store roots as text.
func CanonicalDecisionHex(
	decisionID uuid.UUID,
	title, status, scopeDomain, scopeSystem, owner string,
	evidenceIDs []string,
) string {
	return hex.EncodeToString(CanonicalDecision(
		decisionID, title, status, scopeDomain, scopeSystem, owner, evidenceIDs,
	))
}
