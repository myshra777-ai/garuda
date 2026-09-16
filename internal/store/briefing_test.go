// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestGetBriefing_FirstCallAndSecondCall(t *testing.T) {
	ctx := context.Background()
	store := getTestPostgresStore(t)

	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	workspaceID := uuid.MustParse("c71c76f5-5dad-4eb8-b35f-fc9b01e0fff7") // go-validation-10
	agentID := "test-briefing-agent"
	sessionID := "test-session"

	// Cleanup: remove the watermark this test creates.
	t.Cleanup(func() {
		_, _ = store.Pool().Exec(ctx, `DELETE FROM mcp_agent_watermarks WHERE tenant_id = $1 AND workspace_id = $2 AND agent_id = $3`, tenantID, workspaceID, agentID)
	})
	_, _ = store.Pool().Exec(ctx, `DELETE FROM mcp_agent_watermarks WHERE tenant_id = $1 AND workspace_id = $2 AND agent_id = $3`, tenantID, workspaceID, agentID)

	// First call: no prior briefing.
	first, err := store.GetBriefing(ctx, tenantID, workspaceID, "go-validation-10", agentID, sessionID)
	if err != nil {
		t.Fatalf("first GetBriefing: %v", err)
	}
	if first == nil {
		t.Fatal("first GetBriefing returned nil")
	}
	if first.NewSince.HasPriorBriefing {
		t.Error("first briefing should have HasPriorBriefing = false")
	}
	if first.Workspace != "go-validation-10" {
		t.Errorf("expected workspace name go-validation-10, got %q", first.Workspace)
	}
	if first.Trust.Status != "VERIFIED" {
		t.Errorf("expected trust status VERIFIED, got %q", first.Trust.Status)
	}

	// Second call: prior briefing exists.
	second, err := store.GetBriefing(ctx, tenantID, workspaceID, "go-validation-10", agentID, sessionID)
	if err != nil {
		t.Fatalf("second GetBriefing: %v", err)
	}
	if !second.NewSince.HasPriorBriefing {
		t.Error("second briefing should have HasPriorBriefing = true")
	}
}
