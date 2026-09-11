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

// GenesisV1Content names this genesis in a way that is auditable from
// source. It is a compile-time constant: any reader of this file can
// reproduce the exact string that every tenant's first block hashes.
//
// Changing this string is a chain fork. It requires:
//   - A new ADR documenting the reason.
//   - A migration that stamps existing roots with the old version.
//   - A new GenesisRootV2 function that coexists with V1 for verification
//     of historical proofs.
const GenesisV1Content = "GARUDA_CHAIN_GENESIS_V1"

// GenesisV1Date marks the day this Merkle layer was activated. It is
// included in the genesis hash so that a change to the constant is
// detectable, and so any auditor can see the epoch.
const GenesisV1Date = "2026-09-11"

// Genesis field IDs. These are local to the genesis encoding — the
// canonical encoder treats each call to NewEncoder as a fresh field
// namespace — but we use a distinct range to make source review easy.
const (
	fieldGenesisContent byte = 0x30
	fieldGenesisDate    byte = 0x31
	fieldGenesisTenant  byte = 0x32
)

// GenesisRoot returns the genesis Merkle root for a tenant.
//
// Properties required by ADR-0002 D2:
//
//   - Deterministic: same tenant → same root, forever.
//   - Tenant-scoped: different tenants → different roots. This is what
//     prevents cross-tenant chain confusion.
//   - Versioned: content and date are literal constants; a change to
//     either is a breaking change, not a refactor.
//   - Auditable: an auditor with this source file, without access to
//     Garuda's database, can recompute the root for any tenant ID.
//
// The returned slice is 32 bytes (SHA-256 digest). It is safe for the
// caller to retain.
func GenesisRoot(tenantID uuid.UUID) []byte {
	enc := NewEncoder()
	enc.String(fieldGenesisContent, GenesisV1Content)
	enc.String(fieldGenesisDate, GenesisV1Date)
	enc.UUID(fieldGenesisTenant, tenantID)

	out, err := enc.Finish()
	if err != nil {
		// The encoder cannot fail here: all fields are in strictly
		// increasing order and every length is well under uint32 max.
		// A failure means a developer introduced a bug. Panic is the
		// correct response — a wrong genesis silently propagated is far
		// worse than a startup crash.
		panic("merkle: genesis encoding failed: " + err.Error())
	}

	h := sha256.Sum256(out)
	return h[:]
}

// GenesisRootHex returns GenesisRoot as a lowercase hex string.
// Convenience wrapper for callers that store roots as text.
func GenesisRootHex(tenantID uuid.UUID) string {
	return hex.EncodeToString(GenesisRoot(tenantID))
}
