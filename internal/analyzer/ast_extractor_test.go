// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractRelationships(t *testing.T) {
	tmpDir := t.TempDir()
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
	err := os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte(source), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := Extract(tmpDir)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	foundEntities := make(map[string]bool)
	for _, e := range result.Entities {
		foundEntities[e.Name] = true
	}

	for _, expected := range []string{"User", "Greet", "main"} {
		if !foundEntities[expected] {
			t.Errorf("Expected entity %s not found", expected)
		}
	}

	foundCalls := false
	foundImports := false
	for _, r := range result.Relationships {
		if r.From == "main" && r.To == "Greet" && r.Type == string(RelCalls) {
			foundCalls = true
		}
		if r.To == "fmt" && r.Type == string(RelImports) {
			foundImports = true
		}
	}

	if !foundCalls {
		t.Errorf("Expected CALLS relationship from main to Greet not found")
	}
	if !foundImports {
		t.Errorf("Expected IMPORTS relationship for fmt not found")
	}
}
