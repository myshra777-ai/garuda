// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package merkle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func leaf(s string) []byte {
	h := sha256.Sum256([]byte(s))
	return h[:]
}

func TestEpochRoot_Deterministic(t *testing.T) {
	static := [][]byte{leaf("d1"), leaf("d2")}
	runtime := [][]byte{leaf("r1")}
	parent := leaf("genesis")

	a := EpochRoot(static, runtime, parent, 1)
	b := EpochRoot(static, runtime, parent, 1)
	if !bytes.Equal(a, b) {
		t.Fatalf("epoch root not deterministic")
	}
}

func TestEpochRoot_HeightSensitive(t *testing.T) {
	static := [][]byte{leaf("d1")}
	runtime := [][]byte{leaf("r1")}
	parent := leaf("genesis")

	a := EpochRoot(static, runtime, parent, 1)
	b := EpochRoot(static, runtime, parent, 2)
	if bytes.Equal(a, b) {
		t.Fatalf("different heights produced same root")
	}
}

func TestEpochRoot_ParentSensitive(t *testing.T) {
	static := [][]byte{leaf("d1")}
	runtime := [][]byte{leaf("r1")}

	a := EpochRoot(static, runtime, leaf("genesis_a"), 1)
	b := EpochRoot(static, runtime, leaf("genesis_b"), 1)
	if bytes.Equal(a, b) {
		t.Fatalf("different parents produced same root")
	}
}

func TestEpochRoot_TierSensitive(t *testing.T) {
	parent := leaf("genesis")

	// Same leaf data, different tier
	a := EpochRoot([][]byte{leaf("x")}, nil, parent, 1)
	b := EpochRoot(nil, [][]byte{leaf("x")}, parent, 1)
	if bytes.Equal(a, b) {
		t.Fatalf("static vs runtime tiers produced same root for same leaf")
	}
}

func TestEpochRoot_EmptyTierAllowed(t *testing.T) {
	// Empty static tier, non-empty runtime — valid
	root := EpochRoot(nil, [][]byte{leaf("r1")}, leaf("genesis"), 1)
	if len(root) != 32 {
		t.Fatalf("epoch root length = %d, want 32", len(root))
	}
}

func TestEpochRoot_TwoEmptyTiersStillValid(t *testing.T) {
	// Both tiers empty, but epoch still has a defined root because
	// empty-tree hashes are defined constants.
	root := EpochRoot(nil, nil, leaf("genesis"), 1)
	if len(root) != 32 {
		t.Fatalf("epoch root length = %d, want 32", len(root))
	}
}

func TestEpochRoot_LeafOrderMatters(t *testing.T) {
	parent := leaf("genesis")
	a := EpochRoot([][]byte{leaf("a"), leaf("b")}, nil, parent, 1)
	b := EpochRoot([][]byte{leaf("b"), leaf("a")}, nil, parent, 1)
	if bytes.Equal(a, b) {
		t.Fatalf("leaf order did not affect epoch root")
	}
}

func TestEncodeDecodeLeavesHex_RoundTrip(t *testing.T) {
	orig := [][]byte{leaf("a"), leaf("b"), leaf("c")}
	hexes := EncodeLeavesHex(orig)
	decoded, err := DecodeLeavesHex(hexes)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	for i := range orig {
		if !bytes.Equal(orig[i], decoded[i]) {
			t.Errorf("leaf %d mismatch", i)
		}
	}
}

// TestEpochRoot_PrintVectors prints epoch roots for fixed inputs.
// Run with -v and freeze into testdata/epoch_vectors_v1.json in a
// follow-up commit.
func TestEpochRoot_PrintVectors(t *testing.T) {
	type vec struct {
		name         string
		staticSeeds  []string
		runtimeSeeds []string
		parentSeed   string
		height       uint64
	}

	vectors := []vec{
		{
			name:       "empty_both_tiers",
			parentSeed: "genesis_default",
			height:     1,
		},
		{
			name:        "static_only",
			staticSeeds: []string{"d1", "d2", "d3"},
			parentSeed:  "genesis_default",
			height:      1,
		},
		{
			name:         "runtime_only",
			runtimeSeeds: []string{"r1", "r2"},
			parentSeed:   "genesis_default",
			height:       1,
		},
		{
			name:         "both_tiers",
			staticSeeds:  []string{"d1", "d2"},
			runtimeSeeds: []string{"r1", "r2", "r3"},
			parentSeed:   "genesis_default",
			height:       1,
		},
		{
			name:         "height_100",
			staticSeeds:  []string{"d1"},
			runtimeSeeds: []string{"r1"},
			parentSeed:   "genesis_default",
			height:       100,
		},
	}

	for _, v := range vectors {
		var static [][]byte
		for _, s := range v.staticSeeds {
			static = append(static, leaf(s))
		}
		var runtime [][]byte
		for _, s := range v.runtimeSeeds {
			runtime = append(runtime, leaf(s))
		}
		parent := leaf(v.parentSeed)
		root := EpochRoot(static, runtime, parent, v.height)
		t.Logf("%-20s height=%-4d %s", v.name, v.height, hex.EncodeToString(root))
	}
}
