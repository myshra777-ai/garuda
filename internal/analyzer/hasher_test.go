// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestComputePackageTreeHash_Determinism(t *testing.T) {
	tempDir := t.TempDir()

	file1 := `package core
type Service struct { ID string }
`
	file2 := `package core
func NewService() *Service { return &Service{} }
`
	if err := os.WriteFile(filepath.Join(tempDir, "b_service.go"), []byte(file1), 0644); err != nil {
		t.Fatalf("failed to write b_service.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "a_factory.go"), []byte(file2), 0644); err != nil {
		t.Fatalf("failed to write a_factory.go: %v", err)
	}

	// Compute hash 1
	hash1, err := ComputePackageTreeHash(tempDir)
	if err != nil {
		t.Fatalf("ComputePackageTreeHash failed: %v", err)
	}

	// Compute hash 2 (must be strictly identical)
	hash2, err := ComputePackageTreeHash(tempDir)
	if err != nil {
		t.Fatalf("ComputePackageTreeHash second run failed: %v", err)
	}

	if !bytes.Equal(hash1, hash2) {
		t.Fatalf("tree hash non-deterministic: hash1=%x, hash2=%x", hash1, hash2)
	}

	// Modify content -> hash must change
	modifiedFile1 := `package core
type Service struct { ID string; Name string }
`
	if err := os.WriteFile(filepath.Join(tempDir, "b_service.go"), []byte(modifiedFile1), 0644); err != nil {
		t.Fatalf("failed to write modified file: %v", err)
	}

	hash3, err := ComputePackageTreeHash(tempDir)
	if err != nil {
		t.Fatalf("ComputePackageTreeHash after modification failed: %v", err)
	}

	if bytes.Equal(hash1, hash3) {
		t.Fatalf("expected hash to change after source modification, but remained identical: %x", hash3)
	}
}

func TestComputeSignatureHash_Sensitivity(t *testing.T) {
	h1 := ComputeSignatureHash("STRUCT", "Account", "", "ID string, Balance int64")
	h2 := ComputeSignatureHash("STRUCT", "Account", "", "ID string, Balance int64")
	h3 := ComputeSignatureHash("STRUCT", "Account", "", "ID string, Balance float64")

	if !bytes.Equal(h1, h2) {
		t.Errorf("signature hash non-deterministic for identical signatures")
	}
	if bytes.Equal(h1, h3) {
		t.Errorf("signature hash failed to detect contract mutation")
	}
}
