// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// seedLeaf converts a UTF-8 seed string to a leaf hash. This is the
// convention used by the vector file. If this ever changes, every
// vector in testdata/epoch_vectors_v1.json will mismatch — which is
// the intended detection.
func seedLeaf(seed string) []byte {
	h := sha256.Sum256([]byte(seed))
	return h[:]
}

// TestEpochVectorsV1 asserts that EpochRoot reproduces every frozen
// vector byte-for-byte.
//
// A failure here means the epoch encoding has changed. That is a
// breaking change: every epoch root ever recorded becomes unverifiable
// under the new rule. The fix is NOT to update the vector file — it is
// to revert the change, or to bump CanonicalVersionV1 and open a new
// ADR.
func TestEpochVectorsV1(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "epoch_vectors_v1.json"))
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}

	var file struct {
		Vectors []struct {
			Name         string   `json:"name"`
			StaticSeeds  []string `json:"static_seeds"`
			RuntimeSeeds []string `json:"runtime_seeds"`
			ParentSeed   string   `json:"parent_seed"`
			Height       uint64   `json:"height"`
			RootHex      string   `json:"root_hex"`
		} `json:"vectors"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("parse vectors: %v", err)
	}

	for _, v := range file.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			static := make([][]byte, len(v.StaticSeeds))
			for i, s := range v.StaticSeeds {
				static[i] = seedLeaf(s)
			}
			runtime := make([][]byte, len(v.RuntimeSeeds))
			for i, s := range v.RuntimeSeeds {
				runtime[i] = seedLeaf(s)
			}
			parent := seedLeaf(v.ParentSeed)

			got := hex.EncodeToString(EpochRoot(static, runtime, parent, v.Height))
			if got != v.RootHex {
				t.Errorf("epoch root mismatch for %q\n  got  = %s\n  want = %s",
					v.Name, got, v.RootHex)
			}
		})
	}
}
