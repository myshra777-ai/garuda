// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// TestHandleHandoffAndResume_HappyPath drives the MCP-layer handlers
// end to end against a real database: handleHandoff creates a
// checkpoint and transfers task ownership, handleResume consumes the
// checkpoint and returns the state, and a second handleResume fails.
//
// This is the MCP-layer equivalent of TestHandoffAndResume in
// internal/store. The store test proves the transaction; this test
// proves the argument marshalling and response shape the MCP client
// actually receives.
//
// The second-resume assertion is the important one: it verifies the
// MCP layer preserves the store's Serializable transaction semantics
// — the checkpoint flips from active to restored on the first call,
// and the second call correctly reports not_found.
func TestHandleHandoffAndResume_HappyPath(t *testing.T) {
	srv := getTestMCPServer(t)

	tenantID := uuid.MustParse("d1d1d1d1-d1d1-4d1d-8d1d-d1d1d1d1d1d1")
	sourceID := uuid.MustParse("d2d2d2d2-d2d2-4d2d-8d2d-d2d2d2d2d2d2")
	targetID := uuid.MustParse("d3d3d3d3-d3d3-4d3d-8d3d-d3d3d3d3d3d3")
	taskID := uuid.MustParse("d4d4d4d4-d4d4-4d4d-8d4d-d4d4d4d4d4d4")

	cleanupTenantFixtures(t, srv, tenantID)
	t.Cleanup(func() { cleanupTenantFixtures(t, srv, tenantID) })

	seedTenantWorkspace(t, srv, tenantID, "mcp-happy-path-ws")

	ctx := context.Background()
	if _, err := srv.store.Pool().Exec(ctx, `
		INSERT INTO agents (id, tenant_id, name, model_type, session_id, status)
		VALUES ($1, $2, 'mcp-happy-source', 'test', 'mcp-sess-1', 'idle'),
		       ($3, $2, 'mcp-happy-target', 'test', 'mcp-sess-2', 'idle')
	`, sourceID, tenantID, targetID); err != nil {
		t.Fatalf("seed agents: %v", err)
	}
	if _, err := srv.store.Pool().Exec(ctx, `
		INSERT INTO tasks (id, tenant_id, title, status, owner_agent_id)
		VALUES ($1, $2, 'mcp-happy-task', 'pending', $3)
	`, taskID, tenantID, sourceID); err != nil {
		t.Fatalf("seed task: %v", err)
	}

	// ── handleHandoff ──
	result, err := srv.handleHandoff(map[string]interface{}{
		"tenant_id":       tenantID.String(),
		"workspace":       "mcp-happy-path-ws",
		"task_id":         taskID.String(),
		"source_agent_id": sourceID.String(),
		"target_agent_id": targetID.String(),
		"reason":          "mcp-happy-path",
		"checkpoint_data": map[string]interface{}{"marker": "mcp-happy-path"},
	})
	if err != nil {
		t.Fatalf("handleHandoff: %v", err)
	}
	out, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("handleHandoff returned %T, want map[string]interface{}", result)
	}
	if out["status"] != "completed" {
		t.Fatalf("handleHandoff status = %v, want completed (reason: %v)", out["status"], out["reason"])
	}
	handoffID, _ := out["handoff_id"].(string)
	checkpointID, _ := out["checkpoint_id"].(string)
	if handoffID == "" {
		t.Error("handoff_id is empty")
	}
	if checkpointID == "" {
		t.Fatal("checkpoint_id is empty; cannot test resume")
	}
	if _, err := uuid.Parse(checkpointID); err != nil {
		t.Fatalf("checkpoint_id %q is not a UUID: %v", checkpointID, err)
	}

	// ── first handleResume ──
	resumeResult, err := srv.handleResume(map[string]interface{}{
		"tenant_id":     tenantID.String(),
		"workspace":     "mcp-happy-path-ws",
		"agent_id":      targetID.String(),
		"checkpoint_id": checkpointID,
	})
	if err != nil {
		t.Fatalf("handleResume: %v", err)
	}
	rOut, ok := resumeResult.(map[string]interface{})
	if !ok {
		t.Fatalf("handleResume returned %T, want map[string]interface{}", resumeResult)
	}
	if rOut["status"] != "restored" {
		t.Fatalf("first resume status = %v, want restored (reason: %v)", rOut["status"], rOut["reason"])
	}
	state, ok := rOut["state"].(map[string]interface{})
	if !ok {
		t.Fatalf("resume state is %T, want map[string]interface{}", rOut["state"])
	}
	if state["marker"] != "mcp-happy-path" {
		t.Errorf("resume state marker = %v, want mcp-happy-path", state["marker"])
	}

	// ── second handleResume must report not_found ──
	secondResult, err := srv.handleResume(map[string]interface{}{
		"tenant_id":     tenantID.String(),
		"workspace":     "mcp-happy-path-ws",
		"agent_id":      targetID.String(),
		"checkpoint_id": checkpointID,
	})
	if err != nil {
		t.Fatalf("second handleResume: %v", err)
	}
	secondOut, ok := secondResult.(map[string]interface{})
	if !ok {
		t.Fatalf("second handleResume returned %T, want map[string]interface{}", secondResult)
	}
	if secondOut["status"] != "not_found" {
		t.Errorf("second resume status = %v, want not_found", secondOut["status"])
	}
}
