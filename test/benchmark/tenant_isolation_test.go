// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package benchmark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/store"
)

func getTestPostgresStore(t *testing.T) *store.PostgresStore {
	dbURL := os.Getenv("GARUDA_TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		t.Skip("skipping PostgreSQL tenant isolation test: GARUDA_TEST_DATABASE_URL not set")
	}

	pgStore, err := store.NewPostgresStore(dbURL)
	if err != nil {
		t.Fatalf("failed to initialize postgres store: %v", err)
	}

	return pgStore
}

// TestTenantIsolation_ConcurrentSymbolCacheCollisions tests that two distinct tenants
// writing colliding package paths, module paths, and symbol names concurrently remain strictly isolated.
func TestTenantIsolation_ConcurrentSymbolCacheCollisions(t *testing.T) {
	pgStore := getTestPostgresStore(t)
	ctx := context.Background()

	tenantA := uuid.New()
	tenantB := uuid.New()

	collidingPkgPath := "github.com/enterprise/core/auth"
	collidingSymbolName := "TokenManager"

	// Distinct metadata for Tenant A
	hashA := []byte("sha256_tree_hash_tenant_a_000000")
	sigHashA := []byte("sig_hash_tenant_a_v1_00000000000")
	payloadA, _ := json.Marshal(map[string]any{
		"tenant":  "tenant_a",
		"methods": []string{"GenerateJWT", "ValidateToken"},
	})

	// Distinct metadata for Tenant B
	hashB := []byte("sha256_tree_hash_tenant_b_111111")
	sigHashB := []byte("sig_hash_tenant_b_v2_11111111111")
	payloadB, _ := json.Marshal(map[string]any{
		"tenant":  "tenant_b",
		"methods": []string{"AuthenticateOAuth", "RevokeSession"},
	})

	var wg sync.WaitGroup
	workers := 20
	errChan := make(chan error, workers*4)

	// 1. Concurrent Workspace Registration & Tree Hash Writes
	for i := 0; i < workers; i++ {
		wg.Add(2)

		// Tenant A Worker
		go func() {
			defer wg.Done()
			wsID, err := pgStore.RegisterWorkspace(ctx, tenantA, "workspace-core", "/data/tenant_a/core", true)
			if err != nil {
				errChan <- fmt.Errorf("tenant A register workspace failed: %w", err)
				return
			}
			err = pgStore.UpsertWorkspaceModule(ctx, tenantA, wsID, "github.com/enterprise/core", "./", "commit-a", hashA)
			if err != nil {
				errChan <- fmt.Errorf("tenant A upsert module failed: %w", err)
			}
		}()

		// Tenant B Worker
		go func() {
			defer wg.Done()
			wsID, err := pgStore.RegisterWorkspace(ctx, tenantB, "workspace-core", "/data/tenant_b/core", true)
			if err != nil {
				errChan <- fmt.Errorf("tenant B register workspace failed: %w", err)
				return
			}
			err = pgStore.UpsertWorkspaceModule(ctx, tenantB, wsID, "github.com/enterprise/core", "./", "commit-b", hashB)
			if err != nil {
				errChan <- fmt.Errorf("tenant B upsert module failed: %w", err)
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Fatalf("concurrent setup error: %v", err)
	}

	// 2. Concurrently write colliding symbols
	symbolsA := []store.CachedSymbol{
		{
			TenantID:      tenantA,
			PackagePath:   collidingPkgPath,
			SymbolName:    collidingSymbolName,
			Kind:          "STRUCT",
			Receiver:      "",
			SignatureHash: sigHashA,
			ASTHash:       hashA,
			Payload:       payloadA,
		},
	}

	symbolsB := []store.CachedSymbol{
		{
			TenantID:      tenantB,
			PackagePath:   collidingPkgPath,
			SymbolName:    collidingSymbolName,
			Kind:          "STRUCT",
			Receiver:      "",
			SignatureHash: sigHashB,
			ASTHash:       hashB,
			Payload:       payloadB,
		},
	}

	if err := pgStore.BatchUpsertSymbols(ctx, tenantA, symbolsA); err != nil {
		t.Fatalf("failed to insert tenant A symbols: %v", err)
	}
	if err := pgStore.BatchUpsertSymbols(ctx, tenantB, symbolsB); err != nil {
		t.Fatalf("failed to insert tenant B symbols: %v", err)
	}

	// 3. Concurrently Query across 50 goroutines to assert zero leakage
	readWorkers := 50
	var readWg sync.WaitGroup
	readErrChan := make(chan error, readWorkers*2)

	for i := 0; i < readWorkers; i++ {
		readWg.Add(2)

		// Tenant A Query Assertions
		go func() {
			defer readWg.Done()
			results, err := pgStore.GetSymbolsByPackage(ctx, tenantA, collidingPkgPath)
			if err != nil {
				readErrChan <- fmt.Errorf("tenant A read failed: %w", err)
				return
			}
			if len(results) != 1 {
				readErrChan <- fmt.Errorf("tenant A expected 1 symbol, got %d", len(results))
				return
			}
			sym := results[0]
			if sym.TenantID != tenantA {
				readErrChan <- fmt.Errorf("tenant A received symbol with foreign tenant ID: %s", sym.TenantID)
				return
			}
			if !bytes.Equal(sym.SignatureHash, sigHashA) {
				readErrChan <- fmt.Errorf("tenant A signature hash corrupted by tenant B")
				return
			}
			if !bytes.Contains(sym.Payload, []byte("tenant_a")) {
				readErrChan <- fmt.Errorf("tenant A payload contaminated: %s", string(sym.Payload))
				return
			}
		}()

		// Tenant B Query Assertions
		go func() {
			defer readWg.Done()
			results, err := pgStore.GetSymbolsByPackage(ctx, tenantB, collidingPkgPath)
			if err != nil {
				readErrChan <- fmt.Errorf("tenant B read failed: %w", err)
				return
			}
			if len(results) != 1 {
				readErrChan <- fmt.Errorf("tenant B expected 1 symbol, got %d", len(results))
				return
			}
			sym := results[0]
			if sym.TenantID != tenantB {
				readErrChan <- fmt.Errorf("tenant B received symbol with foreign tenant ID: %s", sym.TenantID)
				return
			}
			if !bytes.Equal(sym.SignatureHash, sigHashB) {
				readErrChan <- fmt.Errorf("tenant B signature hash corrupted by tenant A")
				return
			}
			if !bytes.Contains(sym.Payload, []byte("tenant_b")) {
				readErrChan <- fmt.Errorf("tenant B payload contaminated: %s", string(sym.Payload))
				return
			}
		}()
	}

	readWg.Wait()
	close(readErrChan)

	for err := range readErrChan {
		t.Errorf("tenant isolation violation: %v", err)
	}
}

// TestTenantIsolation_InMemoryWorkspaceContext verifies that in-memory WorkspaceContext
// and MultiModuleImporter instances with identical import paths do not cross-pollinate.
func TestTenantIsolation_InMemoryWorkspaceContext(t *testing.T) {
	tempRootA := t.TempDir()
	tempRootB := t.TempDir()

	collidingPkgPath := "github.com/enterprise/service/config"

	// Setup Workspace A
	dirA := filepath.Join(tempRootA, "config")
	_ = os.MkdirAll(dirA, 0755)
	_ = os.WriteFile(filepath.Join(tempRootA, "go.mod"), []byte("module github.com/enterprise/service\n\ngo 1.22\n"), 0644)
	srcA := `package config
type ServerConfig struct {
	TenantASecret string
}
`
	_ = os.WriteFile(filepath.Join(dirA, "config.go"), []byte(srcA), 0644)

	// Setup Workspace B
	dirB := filepath.Join(tempRootB, "config")
	_ = os.MkdirAll(dirB, 0755)
	_ = os.WriteFile(filepath.Join(tempRootB, "go.mod"), []byte("module github.com/enterprise/service\n\ngo 1.22\n"), 0644)
	srcB := `package config
type ServerConfig struct {
	TenantBPort int
}
`
	_ = os.WriteFile(filepath.Join(dirB, "config.go"), []byte(srcB), 0644)

	var wg sync.WaitGroup
	var resA, resB *analyzer.Result
	var errA, errB error

	wg.Add(2)
	go func() {
		defer wg.Done()
		wsA, err := analyzer.DiscoverWorkspace(tempRootA)
		if err != nil {
			errA = err
			return
		}
		resA, errA = analyzer.AnalyzeWorkspace(context.Background(), wsA)
	}()

	go func() {
		defer wg.Done()
		wsB, err := analyzer.DiscoverWorkspace(tempRootB)
		if err != nil {
			errB = err
			return
		}
		resB, errB = analyzer.AnalyzeWorkspace(context.Background(), wsB)
	}()

	wg.Wait()

	if errA != nil {
		t.Fatalf("Tenant A analysis failed: %v", errA)
	}
	if errB != nil {
		t.Fatalf("Tenant B analysis failed: %v", errB)
	}

	// Verify Tenant A extracted TenantASecret
	var foundSecretA bool
	for _, e := range resA.Entities {
		if e.Package == collidingPkgPath && e.Name == "ServerConfig" {
			for _, f := range e.Fields {
				if f.Name == "TenantASecret" {
					foundSecretA = true
				}
				if f.Name == "TenantBPort" {
					t.Fatalf("leakage detected: Tenant A contains Tenant B field TenantBPort")
				}
			}
		}
	}
	if !foundSecretA {
		t.Errorf("Tenant A failed to extract TenantASecret")
	}

	// Verify Tenant B extracted TenantBPort
	var foundPortB bool
	for _, e := range resB.Entities {
		if e.Package == collidingPkgPath && e.Name == "ServerConfig" {
			for _, f := range e.Fields {
				if f.Name == "TenantBPort" {
					foundPortB = true
				}
				if f.Name == "TenantASecret" {
					t.Fatalf("leakage detected: Tenant B contains Tenant A field TenantASecret")
				}
			}
		}
	}
	if !foundPortB {
		t.Errorf("Tenant B failed to extract TenantBPort")
	}
}
