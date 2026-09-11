// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package merkle

import (
	"crypto/sha256"
	"encoding/hex"
)

// EpochRoot computes the root of one Merkle epoch.
//
// Design (ADR-0002 D5):
//
//	static_root  = BuildRoot(staticLeaves)
//	runtime_root = BuildRoot(runtimeLeaves)
//	epoch_root   = SHA256(version_byte
//	                     || static_root
//	                     || runtime_root
//	                     || parent_epoch_root
//	                     || block_height_u64_be)
//
// Every leaf is committed to via its own tier's Merkle tree. The two
// tiers are combined by hashing, and the parent epoch root ties this
// epoch to the previous one. The block height prevents an attacker
// from re-using an old epoch root at a different height.
//
// Empty tier handling: if a tier has no leaves, BuildRoot returns the
// defined empty-tree root from tree.go. An epoch with no static leaves
// is still valid as long as the runtime tier or the parent chain
// carries meaning.
//
// All inputs are already-hashed leaf values (32 bytes each). The
// function does not re-hash them at the top level — BuildRoot applies
// LeafCommitment to each. This matches how decision hashes and runtime
// verification hashes are produced upstream.
//
// Field IDs for the epoch encoding are defined once in canonical.go
// (0x10 through 0x13) and written here in strictly increasing order
// as the encoder requires.
func EpochRoot(staticLeaves, runtimeLeaves [][]byte, parentRoot []byte, blockHeight uint64) []byte {
	staticRoot := BuildRoot(staticLeaves)
	runtimeRoot := BuildRoot(runtimeLeaves)

	enc := NewEncoder()
	enc.Bytes(fieldEpochStaticRoot, staticRoot)
	enc.Bytes(fieldEpochRuntimeRoot, runtimeRoot)
	enc.Bytes(fieldEpochParentRoot, parentRoot)
	enc.Uint64(fieldEpochBlockHeight, blockHeight)

	out, err := enc.Finish()
	if err != nil {
		// Same reasoning as GenesisRoot and CanonicalDecision: an error
		// here means a coding bug, not a runtime condition. Panicking
		// is correct because a wrong epoch root silently recorded would
		// corrupt the audit log.
		panic("merkle: epoch root encoding failed: " + err.Error())
	}

	h := sha256.Sum256(out)
	return h[:]
}

// EpochRootHex returns EpochRoot as lowercase hex. Convenience wrapper
// for callers that store roots as text.
func EpochRootHex(staticLeaves, runtimeLeaves [][]byte, parentRoot []byte, blockHeight uint64) string {
	return hex.EncodeToString(EpochRoot(staticLeaves, runtimeLeaves, parentRoot, blockHeight))
}

// EncodeLeavesHex converts a slice of raw leaf hashes into the hex
// string form used by the merkle_epoch_leaves table.
//
// Preserving slice order is the caller's responsibility. The storage
// layer records each leaf at an explicit index; reconstruction reads
// them back in index order. This function does not sort.
func EncodeLeavesHex(leaves [][]byte) []string {
	out := make([]string, len(leaves))
	for i, l := range leaves {
		out[i] = hex.EncodeToString(l)
	}
	return out
}

// DecodeLeavesHex is the inverse of EncodeLeavesHex.
func DecodeLeavesHex(hexes []string) ([][]byte, error) {
	out := make([][]byte, len(hexes))
	for i, h := range hexes {
		b, err := hex.DecodeString(h)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}
