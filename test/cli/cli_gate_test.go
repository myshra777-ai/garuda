// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCLIGate_AnalyzeAndDiff_FailOnBreaking(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Build temporary garuda binary
	binPath := filepath.Join(tempDir, "garuda")
	buildCmd := exec.Command("go", "build", "-o", binPath, "github.com/myshra777-ai/garuda/cmd/garuda")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build garuda binary: %v\nOutput: %s", err, string(out))
	}

	// 2. Setup V1 Repository
	v1Dir := filepath.Join(tempDir, "v1")
	_ = os.MkdirAll(v1Dir, 0755)
	_ = os.WriteFile(filepath.Join(v1Dir, "go.mod"), []byte("module github.com/test/service\n\ngo 1.22\n"), 0644)
	v1Src := `package service

type AuthService interface {
	ValidateToken(token string) bool
}
`
	_ = os.WriteFile(filepath.Join(v1Dir, "auth.go"), []byte(v1Src), 0644)

	// 3. Setup V2 Non-Breaking Repository (Added method)
	v2NonBreakingDir := filepath.Join(tempDir, "v2_clean")
	_ = os.MkdirAll(v2NonBreakingDir, 0755)
	_ = os.WriteFile(filepath.Join(v2NonBreakingDir, "go.mod"), []byte("module github.com/test/service\n\ngo 1.22\n"), 0644)
	v2NonBreakingSrc := `package service

type AuthService interface {
	ValidateToken(token string) bool
}

type TokenMetrics struct {
	Count int64
}
`
	_ = os.WriteFile(filepath.Join(v2NonBreakingDir, "auth.go"), []byte(v2NonBreakingSrc), 0644)

	// 4. Setup V2 Breaking Repository (Mutated interface method)
	v2BreakingDir := filepath.Join(tempDir, "v2_break")
	_ = os.MkdirAll(v2BreakingDir, 0755)
	_ = os.WriteFile(filepath.Join(v2BreakingDir, "go.mod"), []byte("module github.com/test/service\n\ngo 1.22\n"), 0644)
	v2BreakingSrc := `package service

type AuthService interface {
	ValidateToken(token string, contextID string) bool
}
`
	_ = os.WriteFile(filepath.Join(v2BreakingDir, "auth.go"), []byte(v2BreakingSrc), 0644)

	// 5. Generate Snapshots
	v1Snapshot := filepath.Join(tempDir, "v1.json")
	cmd1 := exec.Command(binPath, "analyze", v1Dir, "-o", v1Snapshot)
	if out, err := cmd1.CombinedOutput(); err != nil {
		t.Fatalf("v1 analyze failed: %v\nOutput: %s", err, string(out))
	}

	v2CleanSnapshot := filepath.Join(tempDir, "v2_clean.json")
	cmd2 := exec.Command(binPath, "analyze", v2NonBreakingDir, "-o", v2CleanSnapshot)
	if out, err := cmd2.CombinedOutput(); err != nil {
		t.Fatalf("v2 clean analyze failed: %v\nOutput: %s", err, string(out))
	}

	v2BreakSnapshot := filepath.Join(tempDir, "v2_break.json")
	cmd3 := exec.Command(binPath, "analyze", v2BreakingDir, "-o", v2BreakSnapshot)
	if out, err := cmd3.CombinedOutput(); err != nil {
		t.Fatalf("v2 break analyze failed: %v\nOutput: %s", err, string(out))
	}

	// 6. Test Diff Non-Breaking: must exit with code 0
	cleanDiffCmd := exec.Command(binPath, "diff", v1Snapshot, v2CleanSnapshot, "--fail-on-breaking")
	if out, err := cleanDiffCmd.CombinedOutput(); err != nil {
		t.Fatalf("expected clean diff to exit with code 0, got error: %v\nOutput: %s", err, string(out))
	}

	// 7. Test Diff Breaking: must exit with non-zero code (code 1)
	breakDiffCmd := exec.Command(binPath, "diff", v1Snapshot, v2BreakSnapshot, "--fail-on-breaking")
	out, err := breakDiffCmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected breaking diff to fail with code 1, but exited cleanly with output: %s", string(out))
	}
}
