// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package talon

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestBuildPreEditBriefingRequiresWorkspaceStore(t *testing.T) {
	_, err := BuildPreEditBriefing(context.Background(), nil, uuid.Nil, uuid.Nil, "test-workspace", "target", "")
	if err == nil {
		t.Errorf("error = <nil>, want missing workspace database")
	}
}

func TestRunGuardRequiresWorkspaceStore(t *testing.T) {
	_, err := RunGuard(context.Background(), nil, uuid.Nil, uuid.Nil, GuardOptions{RepoDir: "."})
	if err == nil {
		t.Errorf("error = <nil>, want missing workspace database")
	}
}

func TestCompileContractFailsClosedOnMalformedPolicy(t *testing.T) {
	tmp := t.TempDir()
	badYAML := filepath.Join(tmp, "broken-rule.yaml")
	if err := os.WriteFile(badYAML, []byte(`
id: broken-rule
version: v1
then: [invalid yaml syntax:
`), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := CompileContract(
		context.Background(),
		nil,
		uuid.Nil,
		uuid.Nil,
		"test-workspace",
		tmp,
	)
	if err == nil {
		t.Errorf("error = <nil>, want malformed policy error")
	}
}
