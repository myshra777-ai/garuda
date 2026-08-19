// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"os"
	"testing"
)

func TestExtractRelationships(t *testing.T) {
	content := `
package test

import "fmt"

type User struct {
    Name string
}

func (u *User) Greet() {
    fmt.Println("Hello")
}

func main() {
    u := User{}
    u.Greet()
}
`
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.go"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpDir+"/go.mod", []byte("module test\n\ngo 1.21"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Extract(tmpDir)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Expected entities
	expectedEntities := map[string]bool{
		"test.User":  true,
		"test.main":  true,
		"test.Greet": true,
		"test":       true, // package
		tmpFile:      true, // file
		"test.fmt":   true, // external import
	}

	for _, e := range result.Entities {
		if _, ok := expectedEntities[e.ID]; !ok {
			t.Logf("Unexpected entity: %s (kind: %s)", e.ID, e.Kind)
		}
	}

	// Check CALLS relationship from main to Greet
	foundCalls := false
	for _, rel := range result.Relationships {
		if rel.Type == string(RelCalls) && rel.From == "test.main" && rel.To == "test.Greet" {
			foundCalls = true
		}
	}
	if !foundCalls {
		t.Error("Expected CALLS relationship from main to Greet not found")
	}

	// Check IMPORTS relationship from package to fmt
	foundImports := false
	for _, rel := range result.Relationships {
		if rel.Type == string(RelImports) && rel.From == "test" && rel.To == "fmt" {
			foundImports = true
		}
	}
	if !foundImports {
		t.Error("Expected IMPORTS relationship from test to fmt not found")
	}
}
