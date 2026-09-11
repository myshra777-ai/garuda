// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"bytes"
	"context"
	"crypto/sha256"
	"os"
	"testing"

	"github.com/google/uuid"
)

func testDB(t *testing.T) (*PostgresStore, func()) {
	t.Helper()
	dbURL := os.Getenv("GARUDA_TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("GARUDA_TEST_DATABASE_URL not set; skipping v1 storage integration test")
	}
	s, err := NewPostgresStore(dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	return s, func() { s.Close() }
}

func cleanupTenant(t *testing.T, s *PostgresStore, tenantID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `DELETE FROM merkle_epoch_leaves WHERE tenant_id = $1`, tenantID)
	_, _ = s.pool.Exec(ctx, `DELETE FROM merkle_roots WHERE tenant_id = $1`, tenantID)
}

func leafFromSeed(seed string) []byte {
	h := sha256.Sum256([]byte(seed))
	return h[:]
}

func TestAppendLeafAndSeal_SingleLeafVerifies(t *testing.T) {
	s, cleanup := testDB(t)
	defer cleanup()

	tenantID := uuid.New()
	defer cleanupTenant(t, s, tenantID)

	ctx := context.Background()
	leaf := leafFromSeed("decision-1")

	res, err := s.AppendLeafAndSeal(ctx, tenantID, TierStatic, leaf)
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if res.EpochHeight != 1 {
		t.Errorf("epoch height = %d, want 1", res.EpochHeight)
	}
	if res.Tier != TierStatic {
		t.Errorf("tier = %d, want %d", res.Tier, TierStatic)
	}

	proof := EncodeV1Proof(res)
	ok, err := VerifyV1Proof(leaf, proof)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !ok {
		t.Fatal("freshly written proof did not verify")
	}
}

func TestAppendLeafAndSeal_MultipleWrites(t *testing.T) {
	s, cleanup := testDB(t)
	defer cleanup()

	tenantID := uuid.New()
	defer cleanupTenant(t, s, tenantID)

	ctx := context.Background()

	type entry struct {
		leaf  []byte
		proof *V1Proof
	}
	var entries []entry
	for i := 0; i < 10; i++ {
		seed := "decision-" + string(rune('a'+i))
		leaf := leafFromSeed(seed)
		res, err := s.AppendLeafAndSeal(ctx, tenantID, TierStatic, leaf)
		if err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
		if res.EpochHeight != int64(i+1) {
			t.Errorf("epoch height = %d, want %d", res.EpochHeight, i+1)
		}
		entries = append(entries, entry{leaf: leaf, proof: EncodeV1Proof(res)})
	}

	// Every proof must still verify, independently.
	for i, e := range entries {
		ok, err := VerifyV1Proof(e.leaf, e.proof)
		if err != nil {
			t.Errorf("entry %d: %v", i, err)
		} else if !ok {
			t.Errorf("entry %d: proof did not verify", i)
		}
	}
}

func TestAppendLeafAndSeal_TamperLeafFails(t *testing.T) {
	s, cleanup := testDB(t)
	defer cleanup()

	tenantID := uuid.New()
	defer cleanupTenant(t, s, tenantID)

	ctx := context.Background()
	leaf := leafFromSeed("original")
	res, err := s.AppendLeafAndSeal(ctx, tenantID, TierStatic, leaf)
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	proof := EncodeV1Proof(res)

	tampered := leafFromSeed("tampered")
	ok, err := VerifyV1Proof(tampered, proof)
	if err != nil {
		t.Fatalf("verify err: %v", err)
	}
	if ok {
		t.Fatal("tampered leaf verified — this is a security failure")
	}
}

func TestAppendLeafAndSeal_TamperEpochRootFails(t *testing.T) {
	s, cleanup := testDB(t)
	defer cleanup()

	tenantID := uuid.New()
	defer cleanupTenant(t, s, tenantID)

	ctx := context.Background()
	leaf := leafFromSeed("original")
	res, err := s.AppendLeafAndSeal(ctx, tenantID, TierStatic, leaf)
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	proof := EncodeV1Proof(res)
	proof.EpochRoot = "0000000000000000000000000000000000000000000000000000000000000000"

	ok, err := VerifyV1Proof(leaf, proof)
	if err != nil {
		t.Fatalf("verify err: %v", err)
	}
	if ok {
		t.Fatal("tampered epoch root verified — this is a security failure")
	}
}

func TestAppendLeafAndSeal_TierSeparation(t *testing.T) {
	s, cleanup := testDB(t)
	defer cleanup()

	tenantID := uuid.New()
	defer cleanupTenant(t, s, tenantID)

	ctx := context.Background()
	staticLeaf := leafFromSeed("static-1")
	runtimeLeaf := leafFromSeed("runtime-1")

	staticRes, err := s.AppendLeafAndSeal(ctx, tenantID, TierStatic, staticLeaf)
	if err != nil {
		t.Fatalf("static: %v", err)
	}
	runtimeRes, err := s.AppendLeafAndSeal(ctx, tenantID, TierRuntime, runtimeLeaf)
	if err != nil {
		t.Fatalf("runtime: %v", err)
	}

	// Both roots must differ.
	if bytes.Equal(staticRes.EpochRoot, runtimeRes.EpochRoot) {
		t.Fatal("static and runtime epochs produced same root")
	}

	// Cross-tier verification must fail.
	if ok, _ := VerifyV1Proof(staticLeaf, EncodeV1Proof(runtimeRes)); ok {
		t.Fatal("static leaf verified against runtime proof")
	}
}

func TestAppendLeafAndSeal_InvalidInput(t *testing.T) {
	s, cleanup := testDB(t)
	defer cleanup()

	tenantID := uuid.New()
	ctx := context.Background()

	// Wrong hash length.
	if _, err := s.AppendLeafAndSeal(ctx, tenantID, TierStatic, []byte("short")); err == nil {
		t.Error("expected error for short hash")
	}

	// Invalid tier.
	if _, err := s.AppendLeafAndSeal(ctx, tenantID, 99, leafFromSeed("x")); err == nil {
		t.Error("expected error for invalid tier")
	}
}
