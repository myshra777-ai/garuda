// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package talon

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestBuildPreEditBriefingRequiresWorkspaceStore(t *testing.T) {
	_, err := BuildPreEditBriefing(context.Background(), nil, uuid.Nil, uuid.Nil, "default", "HandlePayment", "")
	if err == nil || !strings.Contains(err.Error(), "workspace database is unavailable") {
		t.Fatalf("error = %v, want missing workspace database", err)
	}
}

func TestRunGuardRequiresWorkspaceStore(t *testing.T) {
	_, err := RunGuard(context.Background(), nil, uuid.Nil, uuid.Nil, GuardOptions{RepoDir: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "workspace database is unavailable") {
		t.Fatalf("error = %v, want missing workspace database", err)
	}
}

func TestCompileContractFailsClosedOnMalformedPolicy(t *testing.T) {
	policyDir := filepath.Join(t.TempDir(), "policies")
	if err := os.MkdirAll(policyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	broken := filepath.Join(policyDir, "broken.yaml")
	if err := os.WriteFile(broken, []byte("id: [unterminated\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := CompileContract(context.Background(), nil, uuid.Nil, uuid.Nil, "default", policyDir)
	if err == nil || !strings.Contains(err.Error(), "parse workspace policy directory") {
		t.Fatalf("error = %v, want malformed policy error", err)
	}
}
