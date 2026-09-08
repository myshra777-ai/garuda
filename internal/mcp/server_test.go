// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package mcp_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/myshra777-ai/garuda/internal/mcp/guard"
)

func TestSessionGuard_ConcurrencySafety(t *testing.T) {
	cfg := guard.Config{
		MaxSessionDuration: 10 * time.Second,
		MaxTokens:          100000,
		Mode:               guard.ModeDevelopment,
	}
	sg := guard.NewSessionGuard(cfg)

	const goroutines = 50
	const iterations = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = sg.Intercept(context.Background(), "get_verified_context", false, 10)
			}
		}(i)
	}

	wg.Wait()

	used, limit, _ := sg.Usage()
	expected := int64(goroutines * iterations * 10)
	if used != expected {
		t.Fatalf("concurrency token discrepancy: got %d, expected %d", used, expected)
	}
	if limit != 100000 {
		t.Fatalf("unexpected limit: got %d", limit)
	}
}

func TestSessionGuard_ProductionWriteBlock(t *testing.T) {
	cfg := guard.Config{
		MaxSessionDuration: 10 * time.Second,
		MaxTokens:          100000,
		Mode:               guard.ModeProduction,
	}
	sg := guard.NewSessionGuard(cfg)

	// Read tool should pass
	if err := sg.Intercept(context.Background(), "get_blast_radius", false, 10); err != nil {
		t.Fatalf("expected read tool to pass on production, got: %v", err)
	}

	// Mutating tool must fail
	err := sg.Intercept(context.Background(), "apply_refactor_patch", true, 10)
	if err == nil {
		t.Fatalf("expected mutating tool to be blocked on production workspace")
	}
}
