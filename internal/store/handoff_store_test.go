// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestHandoffAndResume exercises the full two-agent handoff and
// resume path end to end against the test database.
//
// Before this test existed, the path was dead at four layers:
//
//  1. ExecuteHandoffTransaction inserted into evidence_blocks with
//     columns that do not exist. Every call failed and rolled back.
//  2. handoffs.checkpoint_id referenced a table (checkpoints) that no
//     code writes to. The FK was unsatisfiable.
//  3. The dev daemon did not register the /api/v1/agents/handoff
//     route. The CLI returned 404 before reaching the store.
//  4. Task.Description was a string; the column is nullable. lockTask
//     could not scan a NULL description.
//
// All four were fixed on 2026-09-16. This test is the regression guard.
//
// The test inserts its own fixtures with deterministic UUIDs, runs
// the transaction, asserts the full state, then cleans up.
func TestHandoffAndResume(t *testing.T) {
	ctx := context.Background()
	store := getTestPostgresStore(t)

	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	sourceAgentID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	targetAgentID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	taskID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")

	// cleanup removes every fixture this test inserts. Called once
	// before the inserts and again in t.Cleanup. The pre-test call
	// makes the test idempotent: a prior run that failed mid-execution
	// cannot poison the next run. Delete in reverse dependency order.
	cleanup := func() {
		_, _ = store.Pool().Exec(ctx, `DELETE FROM lineage_edges WHERE tenant_id = $1 AND (source_task_id = $2 OR target_task_id = $2)`, tenantID, taskID)
		_, _ = store.Pool().Exec(ctx, `DELETE FROM handoffs WHERE tenant_id = $1 AND task_id = $2`, tenantID, taskID)
		_, _ = store.Pool().Exec(ctx, `DELETE FROM agent_checkpoints WHERE tenant_id = $1 AND task_id = $2`, tenantID, taskID)
		_, _ = store.Pool().Exec(ctx, `DELETE FROM tasks WHERE tenant_id = $1 AND id = $2`, tenantID, taskID)
		_, _ = store.Pool().Exec(ctx, `DELETE FROM agents WHERE tenant_id = $1 AND id IN ($2, $3)`, tenantID, sourceAgentID, targetAgentID)
	}

	cleanup()
	t.Cleanup(cleanup)

	// Set up fixtures.
	_, err := store.Pool().Exec(ctx, `
		INSERT INTO agents (id, tenant_id, name, model_type, session_id, status)
		VALUES ($1, $2, 'test-source-agent', 'test', 'test-session-source', 'idle')
	`, sourceAgentID, tenantID)
	if err != nil {
		t.Fatalf("insert source agent: %v", err)
	}

	_, err = store.Pool().Exec(ctx, `
		INSERT INTO agents (id, tenant_id, name, model_type, session_id, status)
		VALUES ($1, $2, 'test-target-agent', 'test', 'test-session-target', 'idle')
	`, targetAgentID, tenantID)
	if err != nil {
		t.Fatalf("insert target agent: %v", err)
	}

	_, err = store.Pool().Exec(ctx, `
		INSERT INTO tasks (id, tenant_id, title, status, owner_agent_id)
		VALUES ($1, $2, 'handoff-store-test-task', 'pending', $3)
	`, taskID, tenantID, sourceAgentID)
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}

	// Execute the handoff.
	resp, err := store.ExecuteHandoffTransaction(ctx, &HandoffRequest{
		TenantID:       tenantID,
		TaskID:         taskID,
		SourceAgentID:  sourceAgentID,
		TargetAgentID:  targetAgentID,
		Reason:         "handoff_store_test",
		CheckpointData: map[string]string{"test": "handoff_store_test"},
	})
	if err != nil {
		t.Fatalf("ExecuteHandoffTransaction: %v", err)
	}
	if resp == nil {
		t.Fatal("ExecuteHandoffTransaction returned nil response")
	}
	if resp.Status != "completed" {
		t.Errorf("expected handoff status 'completed', got %q", resp.Status)
	}
	if resp.HandoffID == uuid.Nil {
		t.Error("expected non-nil HandoffID")
	}
	if resp.CheckpointID == uuid.Nil {
		t.Error("expected non-nil CheckpointID")
	}

	// Assert the checkpoint was created as active.
	var checkpointStatus string
	err = store.Pool().QueryRow(ctx, `
		SELECT status FROM agent_checkpoints WHERE tenant_id = $1 AND id = $2
	`, tenantID, resp.CheckpointID).Scan(&checkpointStatus)
	if err != nil {
		t.Fatalf("query checkpoint: %v", err)
	}
	if checkpointStatus != "active" {
		t.Errorf("expected checkpoint status 'active', got %q", checkpointStatus)
	}

	// Assert the handoff record was written as completed.
	var handoffStatus string
	err = store.Pool().QueryRow(ctx, `
		SELECT status FROM handoffs WHERE tenant_id = $1 AND id = $2
	`, tenantID, resp.HandoffID).Scan(&handoffStatus)
	if err != nil {
		t.Fatalf("query handoff: %v", err)
	}
	if handoffStatus != "completed" {
		t.Errorf("expected handoff status 'completed', got %q", handoffStatus)
	}

	// Assert agent status transitions.
	var sourceStatus, targetStatus string
	err = store.Pool().QueryRow(ctx, `
		SELECT status FROM agents WHERE tenant_id = $1 AND id = $2
	`, tenantID, sourceAgentID).Scan(&sourceStatus)
	if err != nil {
		t.Fatalf("query source agent: %v", err)
	}
	if sourceStatus != "paused" {
		t.Errorf("expected source agent status 'paused', got %q", sourceStatus)
	}

	err = store.Pool().QueryRow(ctx, `
		SELECT status FROM agents WHERE tenant_id = $1 AND id = $2
	`, tenantID, targetAgentID).Scan(&targetStatus)
	if err != nil {
		t.Fatalf("query target agent: %v", err)
	}
	if targetStatus != "working" {
		t.Errorf("expected target agent status 'working', got %q", targetStatus)
	}

	// First resume: the checkpoint is consumed, the state is returned.
	state, err := store.ResumeAgent(ctx, tenantID, targetAgentID, resp.CheckpointID)
	if err != nil {
		t.Fatalf("first ResumeAgent: %v", err)
	}
	if state == nil {
		t.Error("first ResumeAgent returned nil state")
	}

	// Assert the checkpoint moved to restored.
	err = store.Pool().QueryRow(ctx, `
		SELECT status FROM agent_checkpoints WHERE tenant_id = $1 AND id = $2
	`, tenantID, resp.CheckpointID).Scan(&checkpointStatus)
	if err != nil {
		t.Fatalf("query checkpoint after resume: %v", err)
	}
	if checkpointStatus != "restored" {
		t.Errorf("expected checkpoint status 'restored' after resume, got %q", checkpointStatus)
	}

	// Second resume: the checkpoint is no longer active. Must fail.
	_, err = store.ResumeAgent(ctx, tenantID, targetAgentID, resp.CheckpointID)
	if err == nil {
		t.Fatal("expected second ResumeAgent to fail; it succeeded")
	}
	if !strings.Contains(err.Error(), "checkpoint not found") {
		t.Errorf("expected 'checkpoint not found' error, got %v", err)
	}
}

// TestTwoHandoffsInSameTenant is the regression guard for a constraint
// collision that made the second handoff in any tenant fail. The
// checkpoint INSERT inside ExecuteHandoffTransaction did not set
// checkpoint_name, so it took the schema default 'manual_checkpoint'.
// The unique constraint on (tenant_id, checkpoint_name) fires on the
// second handoff in the same tenant, regardless of which task is being
// handed off. The Serializable transaction rolls back; the handoff
// never happens.
//
// TestHandoffAndResume did not catch this because it cleans up its
// single task's checkpoint row before every run, so the second
// invocation never collides with the first.
func TestTwoHandoffsInSameTenant(t *testing.T) {
	ctx := context.Background()
	store := getTestPostgresStore(t)

	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	sourceAgentID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaab")
	targetAgentID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbc")
	taskAID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccd")
	taskBID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-ccccccccccce")

	cleanup := func() {
		_, _ = store.Pool().Exec(ctx, `DELETE FROM lineage_edges WHERE tenant_id = $1 AND (source_task_id IN ($2, $3) OR target_task_id IN ($2, $3))`, tenantID, taskAID, taskBID)
		_, _ = store.Pool().Exec(ctx, `DELETE FROM handoffs WHERE tenant_id = $1 AND task_id IN ($2, $3)`, tenantID, taskAID, taskBID)
		_, _ = store.Pool().Exec(ctx, `DELETE FROM agent_checkpoints WHERE tenant_id = $1 AND task_id IN ($2, $3)`, tenantID, taskAID, taskBID)
		_, _ = store.Pool().Exec(ctx, `DELETE FROM tasks WHERE tenant_id = $1 AND id IN ($2, $3)`, tenantID, taskAID, taskBID)
		_, _ = store.Pool().Exec(ctx, `DELETE FROM agents WHERE tenant_id = $1 AND id IN ($2, $3)`, tenantID, sourceAgentID, targetAgentID)
	}
	cleanup()
	t.Cleanup(cleanup)

	_, err := store.Pool().Exec(ctx, `
		INSERT INTO agents (id, tenant_id, name, model_type, session_id, status)
		VALUES ($1, $2, 'test-source-agent-2', 'test', 'test-session-source-2', 'idle')
	`, sourceAgentID, tenantID)
	if err != nil {
		t.Fatalf("insert source agent: %v", err)
	}

	_, err = store.Pool().Exec(ctx, `
		INSERT INTO agents (id, tenant_id, name, model_type, session_id, status)
		VALUES ($1, $2, 'test-target-agent-2', 'test', 'test-session-target-2', 'idle')
	`, targetAgentID, tenantID)
	if err != nil {
		t.Fatalf("insert target agent: %v", err)
	}

	for i, taskID := range []uuid.UUID{taskAID, taskBID} {
		_, err := store.Pool().Exec(ctx, `
			INSERT INTO tasks (id, tenant_id, title, status, owner_agent_id)
			VALUES ($1, $2, $3, 'pending', $4)
		`, taskID, tenantID, fmt.Sprintf("handoff-collision-test-task-%d", i), sourceAgentID)
		if err != nil {
			t.Fatalf("insert task %d: %v", i, err)
		}
	}

	// First handoff in the tenant. Succeeds even with the bug: no
	// checkpoint with the default name exists yet.
	respA, err := store.ExecuteHandoffTransaction(ctx, &HandoffRequest{
		TenantID:       tenantID,
		TaskID:         taskAID,
		SourceAgentID:  sourceAgentID,
		TargetAgentID:  targetAgentID,
		Reason:         "first handoff in tenant",
		CheckpointData: map[string]string{"task": "A"},
	})
	if err != nil {
		t.Fatalf("first handoff: %v", err)
	}
	if respA.Status != "completed" {
		t.Errorf("first handoff status = %q, want completed", respA.Status)
	}

	// Second handoff, same tenant, same agents, different task. Before
	// the fix, the checkpoint INSERT collides with A's row on
	// (tenant_id, checkpoint_name='manual_checkpoint').
	respB, err := store.ExecuteHandoffTransaction(ctx, &HandoffRequest{
		TenantID:       tenantID,
		TaskID:         taskBID,
		SourceAgentID:  sourceAgentID,
		TargetAgentID:  targetAgentID,
		Reason:         "second handoff in tenant",
		CheckpointData: map[string]string{"task": "B"},
	})
	if err != nil {
		t.Fatalf("second handoff: %v", err)
	}
	if respB.Status != "completed" {
		t.Errorf("second handoff status = %q, want completed", respB.Status)
	}
	if respA.CheckpointID == respB.CheckpointID {
		t.Error("handoff A and handoff B share a checkpoint ID")
	}
}
