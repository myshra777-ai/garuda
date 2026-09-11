// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestAnalyzeDirectory_SinglePackage validates the full workspace analyzer
// against a minimal single-module Go program.
//
// NOTE: call edges produced by the workspace analyzer carry the PACKAGE as
// source, not the calling function. This is a known granularity limit vs. the
// callgraph extractor used by the benchmark. Assertions here match the
// workspace analyzer's current contract, not an aspiration.
func TestAnalyzeDirectory_SinglePackage(t *testing.T) {
	tmpDir := t.TempDir()

	goMod := "module example.com/test\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	source := `package test

import "fmt"

type User struct {
	Name string
}

func (u *User) Greet() {
	fmt.Println(u.Name)
}

func main() {
	u := &User{Name: "Alice"}
	u.Greet()
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte(source), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := AnalyzeDirectory(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeDirectory failed: %v", err)
	}

	// --- Entities ---
	foundEntities := make(map[string]bool)
	for _, e := range result.Entities {
		foundEntities[e.Name] = true
	}
	for _, expected := range []string{"User", "Greet", "main"} {
		if !foundEntities[expected] {
			t.Errorf("expected entity %q not found", expected)
		}
	}

	// --- Relationships ---
	var hasImportsFmt, hasCallsGreet, hasCallsPrintln bool
	for i, r := range result.Relationships {
		switch {
		case r.Type == string(RelImports) && r.To == "fmt":
			hasImportsFmt = true
		case r.Type == string(RelCalls) && r.To == "example.com/test.Greet":
			hasCallsGreet = true
		case r.Type == string(RelCalls) && r.To == "fmt.Println":
			hasCallsPrintln = true
		}

		// Honesty invariant applies to every edge the analyzer emits.
		if r.Confidence <= 0 {
			t.Errorf("relationship %d (%s -> %s) has non-positive confidence: %f",
				i, r.From, r.To, r.Confidence)
		}
		if r.ResolutionMethod == "" {
			t.Errorf("relationship %d (%s -> %s) has empty resolution_method",
				i, r.From, r.To)
		}
		if r.ResolutionStatus == "" {
			t.Errorf("relationship %d (%s -> %s) has empty resolution_status",
				i, r.From, r.To)
		}
	}

	if !hasImportsFmt {
		t.Errorf("expected IMPORTS edge to fmt (got %d relationships)", len(result.Relationships))
	}
	if !hasCallsGreet {
		t.Errorf("expected CALLS edge to example.com/test.Greet (got %d relationships)", len(result.Relationships))
	}
	if !hasCallsPrintln {
		t.Errorf("expected CALLS edge to fmt.Println (got %d relationships)", len(result.Relationships))
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
