// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCollectGitDeltasRejectsNonRepositoryWorkingTree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	_, _, err := CollectGitDeltas(t.TempDir(), false, "", nil)
	if err == nil {
		t.Fatal("expected Git discovery to fail outside a repository")
	}
}

func TestCollectGitDeltasIncludesUntrackedSourceInUnbornRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	dir := t.TempDir()
	cmd := exec.Command("git", "-C", dir, "init", "--quiet")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}

	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte("package main\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	mode, deltas, err := CollectGitDeltas(dir, false, "", nil)
	if err != nil {
		t.Fatalf("CollectGitDeltas: %v", err)
	}
	if mode != "working_tree" {
		t.Fatalf("mode = %q, want working_tree", mode)
	}
	if len(deltas) != 1 || deltas[0].FilePath != "main.go" || deltas[0].Status != "A" {
		t.Fatalf("deltas = %#v, want one added main.go delta", deltas)
	}
}
