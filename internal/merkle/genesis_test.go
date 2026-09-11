// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package merkle

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/google/uuid"
)

func TestGenesisRoot_Deterministic(t *testing.T) {
	tenant := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	a := GenesisRoot(tenant)
	b := GenesisRoot(tenant)
	if !bytes.Equal(a, b) {
		t.Errorf("genesis root is not deterministic:\n  a=%x\n  b=%x", a, b)
	}
}

func TestGenesisRoot_TenantScoped(t *testing.T) {
	tenantA := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tenantB := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	rootA := GenesisRoot(tenantA)
	rootB := GenesisRoot(tenantB)

	if bytes.Equal(rootA, rootB) {
		t.Fatal("two tenants produced the same genesis root — this is the bug ADR-0002 D2 exists to prevent")
	}
}

func TestGenesisRoot_Length(t *testing.T) {
	tenant := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	root := GenesisRoot(tenant)
	if len(root) != 32 {
		t.Errorf("genesis root length = %d, want 32", len(root))
	}
}

func TestGenesisRoot_GenesisIsNotZero(t *testing.T) {
	tenant := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	root := GenesisRoot(tenant)
	zero := make([]byte, 32)
	if bytes.Equal(root, zero) {
		t.Error("genesis root is all zeros")
	}
}

// TestGenesisRoot_PrintVectors prints genesis roots for fixed tenant IDs.
// Run with -v and freeze the hex values into testdata/genesis_vectors_v1.json
// in a follow-up commit.
func TestGenesisRoot_PrintVectors(t *testing.T) {
	tenants := []struct {
		name string
		id   string
	}{
		{"nil_tenant", "00000000-0000-0000-0000-000000000000"},
		{"default_tenant", "00000000-0000-0000-0000-000000000001"},
		{"example_a", "11111111-1111-1111-1111-111111111111"},
		{"example_b", "22222222-2222-2222-2222-222222222222"},
	}
	for _, tc := range tenants {
		id := uuid.MustParse(tc.id)
		hexRoot := hex.EncodeToString(GenesisRoot(id))
		t.Logf("%-16s %s", tc.name, hexRoot)
	}
}
