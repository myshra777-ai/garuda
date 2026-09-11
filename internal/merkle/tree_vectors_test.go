// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package merkle

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestTreeVectorsV1 asserts every frozen tree vector reproduces.
// A failure means the tree construction has changed, which invalidates
// every historical Merkle proof. The fix is to revert, not to update
// the vector file.
func TestTreeVectorsV1(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "tree_vectors_v1.json"))
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}

	var file struct {
		Trees []struct {
			Name   string   `json:"name"`
			Leaves []string `json:"leaves"`
			Root   string   `json:"root"`
			Proofs []struct {
				Index int `json:"index"`
				Path  []struct {
					Position string `json:"position"`
					Hash     string `json:"hash"`
				} `json:"path"`
			} `json:"proofs"`
		} `json:"trees"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("parse vectors: %v", err)
	}

	for _, tree := range file.Trees {
		t.Run(tree.Name, func(t *testing.T) {
			leaves := make([][]byte, len(tree.Leaves))
			for i, s := range tree.Leaves {
				leaves[i] = []byte(s)
			}

			// Root
			gotRoot := BuildRoot(leaves)
			wantRoot, err := hex.DecodeString(tree.Root)
			if err != nil {
				t.Fatalf("decode expected root: %v", err)
			}
			if !bytes.Equal(gotRoot, wantRoot) {
				t.Errorf("root mismatch:\n  got  = %s\n  want = %s",
					hex.EncodeToString(gotRoot), tree.Root)
			}

			// Proofs
			for _, p := range tree.Proofs {
				gotProof, err := BuildProof(leaves, p.Index)
				if err != nil {
					t.Errorf("index %d: BuildProof: %v", p.Index, err)
					continue
				}
				if len(gotProof) != len(p.Path) {
					t.Errorf("index %d: proof length %d, want %d",
						p.Index, len(gotProof), len(p.Path))
					continue
				}
				for i, want := range p.Path {
					got := gotProof[i]
					var wantPos byte
					switch want.Position {
					case "L":
						wantPos = SiblingLeft
					case "R":
						wantPos = SiblingRight
					}
					if got.Position != wantPos {
						t.Errorf("index %d step %d: position %c, want %c",
							p.Index, i, got.Position, wantPos)
					}
					wantHash, _ := hex.DecodeString(want.Hash)
					if !bytes.Equal(got.Hash, wantHash) {
						t.Errorf("index %d step %d: hash mismatch\n  got  = %s\n  want = %s",
							p.Index, i, hex.EncodeToString(got.Hash), want.Hash)
					}
				}

				// Also assert the proof verifies against the frozen root
				if !VerifyInclusion(leaves[p.Index], gotProof, wantRoot) {
					t.Errorf("index %d: proof does not verify against frozen root", p.Index)
				}
			}
		})
	}
}
