// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package benchmark

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

func TestBenchmark_024_GoWork_CrossModule_ResolutionAndBlastRadius(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Create go.work
	goWorkContent := `go 1.22

use (
	./contracts
	./services/payment
	./gateways/api
)
`
	if err := os.WriteFile(filepath.Join(tempDir, "go.work"), []byte(goWorkContent), 0644); err != nil {
		t.Fatalf("failed to create go.work: %v", err)
	}

	// 2. Module A: contracts (Provider)
	contractsDir := filepath.Join(tempDir, "contracts")
	_ = os.MkdirAll(contractsDir, 0755)
	_ = os.WriteFile(filepath.Join(contractsDir, "go.mod"), []byte("module github.com/garuda/contracts\n\ngo 1.22\n"), 0644)
	contractsSrc := `package contracts

type Transaction struct {
	ID     string
	Amount int64
}

type PaymentProcessor interface {
	Process(tx Transaction) (string, error)
}
`
	_ = os.WriteFile(filepath.Join(contractsDir, "contracts.go"), []byte(contractsSrc), 0644)

	// 3. Module B: services/payment (Implementer)
	paymentDir := filepath.Join(tempDir, "services", "payment")
	_ = os.MkdirAll(paymentDir, 0755)
	_ = os.WriteFile(filepath.Join(paymentDir, "go.mod"), []byte("module github.com/garuda/payment\n\ngo 1.22\nrequire github.com/garuda/contracts v0.0.0\n"), 0644)
	paymentSrc := `package payment

import "github.com/garuda/contracts"

type StripeGateway struct {
	APIKey string
}

func (s *StripeGateway) Process(tx contracts.Transaction) (string, error) {
	return "stripe_charge_ok", nil
}
`
	_ = os.WriteFile(filepath.Join(paymentDir, "stripe.go"), []byte(paymentSrc), 0644)

	// 4. Module C: gateways/api (Consumer)
	apiDir := filepath.Join(tempDir, "gateways", "api")
	_ = os.MkdirAll(apiDir, 0755)
	_ = os.WriteFile(filepath.Join(apiDir, "go.mod"), []byte("module github.com/garuda/api\n\ngo 1.22\nrequire (\n\tgithub.com/garuda/contracts v0.0.0\n\tgithub.com/garuda/payment v0.0.0\n)\n"), 0644)
	apiSrc := `package api

import (
	"github.com/garuda/contracts"
	"github.com/garuda/payment"
)

type CheckoutRouter struct {
	Gateway *payment.StripeGateway
}

func (r *CheckoutRouter) HandleCheckout(txID string, amt int64) (string, error) {
	tx := contracts.Transaction{ID: txID, Amount: amt}
	return r.Gateway.Process(tx)
}
`
	_ = os.WriteFile(filepath.Join(apiDir, "router.go"), []byte(apiSrc), 0644)

	// 5. Discover Workspace & Analyze Baseline (V1)
	ws, err := analyzer.DiscoverWorkspace(tempDir)
	if err != nil {
		t.Fatalf("DiscoverWorkspace failed: %v", err)
	}

	v1Result, err := analyzer.AnalyzeWorkspace(context.Background(), ws)
	if err != nil {
		t.Fatalf("AnalyzeWorkspace V1 failed: %v", err)
	}

	// 6. Ground-Truth Invariant Validations
	if v1Result.Stats.Packages != 3 {
		t.Errorf("expected 3 packages, got %d", v1Result.Stats.Packages)
	}
	if v1Result.Stats.Structs != 3 { // Transaction, StripeGateway, CheckoutRouter
		t.Errorf("expected 3 structs, got %d", v1Result.Stats.Structs)
	}
	if v1Result.Stats.Interfaces != 1 { // PaymentProcessor
		t.Errorf("expected 1 interface, got %d", v1Result.Stats.Interfaces)
	}

	// 7. Verify Cross-Module IMPLEMENTS Edge
	var foundImplements bool
	for _, rel := range v1Result.Relationships {
		if rel.Type == string(analyzer.RelImplements) &&
			strings.Contains(rel.From, "StripeGateway") &&
			strings.Contains(rel.To, "PaymentProcessor") {
			foundImplements = true
			if rel.Confidence != 1.0 {
				t.Errorf("expected 1.0 confidence for cross-module IMPLEMENTS edge, got %f", rel.Confidence)
			}
		}
	}
	if !foundImplements {
		t.Errorf("failed to discover cross-module IMPLEMENTS edge: StripeGateway -> PaymentProcessor")
	}

	// 8. Simulate Breaking Change in Module A (V2)
	v2ContractsSrc := `package contracts

type Transaction struct {
	ID     string
	Amount int64
}

type PaymentProcessor interface {
	Process(tx Transaction, authCode string) (string, error)
}
`
	_ = os.WriteFile(filepath.Join(contractsDir, "contracts.go"), []byte(v2ContractsSrc), 0644)

	v2Result, err := analyzer.AnalyzeWorkspace(context.Background(), ws)
	if err != nil {
		t.Fatalf("AnalyzeWorkspace V2 failed: %v", err)
	}

	// 9. Compute AST Semantic Diff
	diffReport := analyzer.Diff(v1Result, v2Result)
	if diffReport == nil {
		t.Fatalf("expected non-nil diff report")
	}

	if !diffReport.HasBreakingChanges() {
		t.Errorf("expected diff report to detect breaking changes on interface mutation")
	}

	// 10. Verify Reverse Cross-Module Blast Traversal
	// From broken interface (PaymentProcessor) -> Implementer (StripeGateway)
	var directConsumers []string
	targetInterface := "github.com/garuda/contracts.PaymentProcessor"

	for _, rel := range v1Result.Relationships {
		if rel.Type == string(analyzer.RelImplements) && rel.To == targetInterface {
			directConsumers = append(directConsumers, rel.From)
		}
	}

	if len(directConsumers) == 0 {
		t.Fatalf("expected at least 1 direct implementer in blast radius, got 0")
	}

	var foundStripeConsumer bool
	for _, c := range directConsumers {
		if strings.Contains(c, "StripeGateway") {
			foundStripeConsumer = true
		}
	}
	if !foundStripeConsumer {
		t.Errorf("failed to trace reverse blast radius to StripeGateway in Module B")
	}
}
