// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package merkle

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
)

func leafBytes(s string) []byte { return []byte(s) }

// TestTree_RoundTrip asserts that for a given leaf set and every index,
// the inclusion proof verifies against the root.
func TestTree_RoundTrip(t *testing.T) {
	cases := []struct {
		name   string
		leaves []string
	}{
		{"1-leaf", []string{"a"}},
		{"2-leaf", []string{"a", "b"}},
		{"3-leaf", []string{"a", "b", "c"}},
		{"4-leaf", []string{"a", "b", "c", "d"}},
		{"5-leaf", []string{"a", "b", "c", "d", "e"}},
		{"7-leaf", []string{"a", "b", "c", "d", "e", "f", "g"}},
		{"8-leaf", []string{"a", "b", "c", "d", "e", "f", "g", "h"}},
		{"15-leaf", []string{
			"a", "b", "c", "d", "e", "f", "g", "h",
			"i", "j", "k", "l", "m", "n", "o",
		}},
		{"16-leaf", []string{
			"a", "b", "c", "d", "e", "f", "g", "h",
			"i", "j", "k", "l", "m", "n", "o", "p",
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			leaves := make([][]byte, len(tc.leaves))
			for i, s := range tc.leaves {
				leaves[i] = leafBytes(s)
			}
			root := BuildRoot(leaves)

			for i := range leaves {
				proof, err := BuildProof(leaves, i)
				if err != nil {
					t.Fatalf("BuildProof(%d): %v", i, err)
				}
				if !VerifyInclusion(leaves[i], proof, root) {
					t.Errorf("index %d: proof does not verify against root\n  leaf: %q\n  root: %x",
						i, tc.leaves[i], root)
				}
			}
		})
	}
}

// TestTree_TamperDetection asserts that tampering with any component
// causes verification to fail.
func TestTree_TamperDetection(t *testing.T) {
	leaves := [][]byte{
		leafBytes("alpha"),
		leafBytes("beta"),
		leafBytes("gamma"),
		leafBytes("delta"),
	}
	root := BuildRoot(leaves)
	proof, err := BuildProof(leaves, 1)
	if err != nil {
		t.Fatalf("BuildProof: %v", err)
	}

	// Original verifies
	if !VerifyInclusion(leaves[1], proof, root) {
		t.Fatal("original proof does not verify")
	}

	t.Run("tampered_leaf", func(t *testing.T) {
		if VerifyInclusion(leafBytes("BETA"), proof, root) {
			t.Error("tampered leaf verified")
		}
	})

	t.Run("tampered_proof_hash", func(t *testing.T) {
		bad := make([]ProofNode, len(proof))
		copy(bad, proof)
		bad[0].Hash = bytes.Repeat([]byte{0xFF}, 32)
		if VerifyInclusion(leaves[1], bad, root) {
			t.Error("tampered proof verified")
		}
	})

	t.Run("tampered_root", func(t *testing.T) {
		badRoot := bytes.Repeat([]byte{0xAA}, 32)
		if VerifyInclusion(leaves[1], proof, badRoot) {
			t.Error("tampered root verified")
		}
	})

	t.Run("truncated_proof", func(t *testing.T) {
		if len(proof) > 0 && VerifyInclusion(leaves[1], proof[:len(proof)-1], root) {
			t.Error("truncated proof verified")
		}
	})

	t.Run("extended_proof", func(t *testing.T) {
		extended := append([]ProofNode{}, proof...)
		extended = append(extended, ProofNode{Position: SiblingLeft, Hash: bytes.Repeat([]byte{0x00}, 32)})
		if VerifyInclusion(leaves[1], extended, root) {
			t.Error("extended proof verified")
		}
	})
}

// TestTree_OddDuplicationIsNotAmbiguous asserts that the classic
// CVE-2012-2459 case does not apply: [A,B,C] and [A,B,C,C] must
// produce different roots.
func TestTree_OddDuplicationIsNotAmbiguous(t *testing.T) {
	abc := [][]byte{leafBytes("a"), leafBytes("b"), leafBytes("c")}
	abcc := [][]byte{leafBytes("a"), leafBytes("b"), leafBytes("c"), leafBytes("c")}

	rootABC := BuildRoot(abc)
	rootABCC := BuildRoot(abcc)

	if bytes.Equal(rootABC, rootABCC) {
		t.Errorf("CVE-2012-2459 style collision:\n  root(A,B,C)    = %x\n  root(A,B,C,C) = %x",
			rootABC, rootABCC)
	}
}

// TestTree_EmptyTree asserts the defined empty-tree root.
func TestTree_EmptyTree(t *testing.T) {
	root := BuildRoot(nil)
	if len(root) != 32 {
		t.Fatalf("empty root length = %d, want 32", len(root))
	}
	// Same input twice → same output
	root2 := BuildRoot([][]byte{})
	if !bytes.Equal(root, root2) {
		t.Error("empty tree root is not deterministic")
	}
}

// TestTree_ProofErrors asserts error paths.
func TestTree_ProofErrors(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		_, err := BuildProof(nil, 0)
		if err != ErrEmptyTree {
			t.Errorf("expected ErrEmptyTree, got %v", err)
		}
	})
	t.Run("index_negative", func(t *testing.T) {
		_, err := BuildProof([][]byte{leafBytes("a")}, -1)
		if err == nil {
			t.Error("expected error for negative index")
		}
	})
	t.Run("index_out_of_range", func(t *testing.T) {
		_, err := BuildProof([][]byte{leafBytes("a")}, 1)
		if err == nil {
			t.Error("expected error for out-of-range index")
		}
	})
}

// TestTree_DomainSeparation asserts that a leaf's bytes can never
// collide with an internal node's bytes even under adversarial input.
func TestTree_DomainSeparation(t *testing.T) {
	// Craft leaf data whose commitment would equal an internal node
	// of some other pair, if domain separation were absent.
	// With prefixes, this is impossible by construction.
	leafA := []byte("a")
	leafB := []byte("b")
	internal := InternalCommitment(LeafCommitment(leafA), LeafCommitment(leafB))

	// Adversarial leaf whose bytes equal the internal commitment
	adversarial := internal
	leafCommitment := LeafCommitment(adversarial)

	if bytes.Equal(leafCommitment, internal) {
		t.Error("leaf commitment equals internal commitment — domain separation failed")
	}
}

// TestTree_GoldenVectors prints hex roots and proofs for a fixed set of
// trees. Run with -v to capture values for freezing into testdata.
func TestTree_GoldenVectors(t *testing.T) {
	cases := []struct {
		name   string
		leaves []string
	}{
		{"empty", nil},
		{"1-leaf", []string{"a"}},
		{"2-leaf", []string{"a", "b"}},
		{"3-leaf", []string{"a", "b", "c"}},
		{"4-leaf", []string{"a", "b", "c", "d"}},
		{"7-leaf", []string{"a", "b", "c", "d", "e", "f", "g"}},
	}

	for _, tc := range cases {
		leaves := make([][]byte, len(tc.leaves))
		for i, s := range tc.leaves {
			leaves[i] = leafBytes(s)
		}
		root := BuildRoot(leaves)
		t.Logf("%s:\n  root = %s", tc.name, hex.EncodeToString(root))

		for i := range leaves {
			proof, _ := BuildProof(leaves, i)
			var proofHex []string
			for _, pn := range proof {
				proofHex = append(proofHex, fmt.Sprintf("%c:%s", pn.Position, hex.EncodeToString(pn.Hash)))
			}
			t.Logf("  proof[%d] = %v", i, proofHex)
		}
	}
}
