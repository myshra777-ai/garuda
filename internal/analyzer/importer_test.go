// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMultiModuleImporter_CrossModuleResolution(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Create go.work
	goWorkContent := `go 1.22

use (
	./contracts
	./services/billing
)
`
	if err := os.WriteFile(filepath.Join(tempDir, "go.work"), []byte(goWorkContent), 0644); err != nil {
		t.Fatalf("failed to create go.work: %v", err)
	}

	// 2. Create Module A (contracts)
	contractsDir := filepath.Join(tempDir, "contracts")
	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create contracts dir: %v", err)
	}
	contractsMod := `module github.com/example/contracts

go 1.22
`
	if err := os.WriteFile(filepath.Join(contractsDir, "go.mod"), []byte(contractsMod), 0644); err != nil {
		t.Fatalf("failed to write contracts go.mod: %v", err)
	}

	contractSrc := `package contracts

type CustomerID string

type PaymentGateway interface {
	Process(id CustomerID, amount int64) (string, error)
}

type Account struct {
	ID      CustomerID
	Balance int64
}
`
	if err := os.WriteFile(filepath.Join(contractsDir, "contracts.go"), []byte(contractSrc), 0644); err != nil {
		t.Fatalf("failed to write contracts.go: %v", err)
	}

	// 3. Create Module B (services/billing) importing Module A
	billingDir := filepath.Join(tempDir, "services", "billing")
	if err := os.MkdirAll(billingDir, 0755); err != nil {
		t.Fatalf("failed to create billing dir: %v", err)
	}
	billingMod := `module github.com/example/billing

go 1.22

require github.com/example/contracts v0.0.0
`
	if err := os.WriteFile(filepath.Join(billingDir, "go.mod"), []byte(billingMod), 0644); err != nil {
		t.Fatalf("failed to write billing go.mod: %v", err)
	}

	billingSrc := `package billing

import "github.com/example/contracts"

type BillingEngine struct {
	Gateway contracts.PaymentGateway
	Account contracts.Account
}

func (b *BillingEngine) Charge(amount int64) (string, error) {
	return b.Gateway.Process(b.Account.ID, amount)
}
`
	billingFile := filepath.Join(billingDir, "billing.go")
	if err := os.WriteFile(billingFile, []byte(billingSrc), 0644); err != nil {
		t.Fatalf("failed to write billing.go: %v", err)
	}

	// 4. Discover Workspace
	ws, err := DiscoverWorkspace(tempDir)
	if err != nil {
		t.Fatalf("DiscoverWorkspace failed: %v", err)
	}

	// 5. Instantiate MultiModuleImporter
	fset := token.NewFileSet()
	imp := NewMultiModuleImporter(fset, ws)

	// 6. Parse and type-check Module B using MultiModuleImporter
	parsedBillingFile, err := parser.ParseFile(fset, billingFile, billingSrc, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse billing.go: %v", err)
	}

	billingPkg := types.NewPackage("github.com/example/billing", "billing")
	conf := types.Config{
		Importer: imp,
	}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	checker := types.NewChecker(&conf, fset, billingPkg, info)
	if err := checker.Files([]*ast.File{parsedBillingFile}); err != nil {
		t.Fatalf("cross-module type check failed: %v", err)
	}

	// 7. Verify Resolved Types
	engineObj := billingPkg.Scope().Lookup("BillingEngine")
	if engineObj == nil {
		t.Fatalf("failed to resolve BillingEngine struct in billing package")
	}

	namedType, ok := engineObj.Type().(*types.Named)
	if !ok {
		t.Fatalf("BillingEngine is not a named type")
	}

	structType, ok := namedType.Underlying().(*types.Struct)
	if !ok {
		t.Fatalf("BillingEngine underlying is not a struct")
	}

	if structType.NumFields() != 2 {
		t.Fatalf("expected 2 fields on BillingEngine, got %d", structType.NumFields())
	}

	gatewayField := structType.Field(0)
	if gatewayField.Name() != "Gateway" {
		t.Errorf("expected first field to be Gateway, got %s", gatewayField.Name())
	}

	// Check that Gateway's underlying type resolved to contracts.PaymentGateway Interface
	gatewayType := gatewayField.Type()
	if !strings.Contains(gatewayType.String(), "github.com/example/contracts.PaymentGateway") {
		t.Errorf("Gateway field type mismatch: %s", gatewayType.String())
	}
}

func TestMultiModuleImporter_CycleDetection(t *testing.T) {
	tempDir := t.TempDir()

	ws := &WorkspaceContext{
		RootPath: tempDir,
		PackageRoots: map[string]string{
			"pkg/a": filepath.Join(tempDir, "a"),
		},
	}

	fset := token.NewFileSet()
	imp := NewMultiModuleImporter(fset, ws)

	// Simulate cycle condition
	imp.inProgress["pkg/a"] = true

	_, err := imp.Import("pkg/a")
	if err == nil {
		t.Fatalf("expected cycle detection error, got nil")
	}
	if !strings.Contains(err.Error(), "import cycle detected") {
		t.Errorf("unexpected error message: %v", err)
	}
}
