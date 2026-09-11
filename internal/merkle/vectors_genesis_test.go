// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package merkle

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

// TestGenesisVectorsV1 asserts that GenesisRoot reproduces every frozen
// vector byte-for-byte.
//
// A failure here means the genesis encoding has changed. That is a
// breaking change: every tenant's chain start becomes unverifiable
// under the new rule. The fix is NOT to update the vector file — it is
// to revert the change, or to bump the genesis content constant and
// open a new ADR per the procedure described in genesis.go.
func TestGenesisVectorsV1(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "genesis_vectors_v1.json"))
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}

	var file struct {
		Vectors []struct {
			Name     string `json:"name"`
			TenantID string `json:"tenant_id"`
			RootHex  string `json:"root_hex"`
		} `json:"vectors"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("parse vectors: %v", err)
	}

	for _, v := range file.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			tenant := uuid.MustParse(v.TenantID)
			got := hex.EncodeToString(GenesisRoot(tenant))
			if got != v.RootHex {
				t.Errorf("genesis root mismatch for tenant %s\n  got  = %s\n  want = %s",
					v.TenantID, got, v.RootHex)
			}
		})
	}
}
