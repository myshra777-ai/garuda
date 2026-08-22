// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ComputePackageTreeHash reads all non-test .go files in a directory in sorted order
// and computes a combined SHA-256 tree hash representing the package state.
func ComputePackageTreeHash(dirPath string) ([]byte, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package directory %s: %w", dirPath, err)
	}

	var goFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
			goFiles = append(goFiles, entry.Name())
		}
	}

	if len(goFiles) == 0 {
		// Pure test directories or empty packages yield a deterministic empty hash
		emptyHash := sha256.Sum256([]byte{})
		return emptyHash[:], nil
	}

	// Deterministic alphabetical ordering
	sort.Strings(goFiles)

	hasher := sha256.New()
	for _, fileName := range goFiles {
		filePath := filepath.Join(dirPath, fileName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}
		hasher.Write(fmt.Appendf(nil, "%s:", fileName))
		hasher.Write(content)
	}

	return hasher.Sum(nil), nil
}

// ComputeSignatureHash creates a cryptographic digest of an entity's public contract
// (fields, method signatures, return types) to detect breaking vs non-breaking changes.
func ComputeSignatureHash(kind, name, receiver, signature string) []byte {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s|%s|%s|%s", kind, name, receiver, signature)))
	return h.Sum(nil)
}
