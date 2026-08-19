// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package engine

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

type mockCheckpointStore struct{}

func (m *mockCheckpointStore) CreateCheckpoint(ctx context.Context, tenantID uuid.UUID, agentID string, name string, reason string, state json.RawMessage, merkleRoot string) (*types.AgentCheckpoint, error) {
	now := time.Now().UTC()

	var data types.CheckpointData
	if len(state) > 0 {
		_ = json.Unmarshal(state, &data)
	}

	return &types.AgentCheckpoint{
		ID:             uuid.New(),
		TenantID:       tenantID,
		AgentID:        agentID,
		CheckpointName: name,
		Status:         types.CheckpointStatusActive,
		CheckpointData: data,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func (m *mockCheckpointStore) GetLatestCheckpoint(ctx context.Context, tenantID uuid.UUID, agentID string) (*types.AgentCheckpoint, error) {
	return nil, nil
}

func TestDeadMansSwitchTrigger(t *testing.T) {
	mockStore := &mockCheckpointStore{}
	watchdog := NewWatchdogEngine(mockStore)
	ctx := context.Background()

	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	agentID := "test-loop-agent"

	// Simulate 9 consecutive identical tool calls (below threshold)
	for i := 0; i < 9; i++ {
		err := watchdog.RecordActivity(ctx, tenantID, agentID, "query_database", "dummy_root")
		if err != nil {
			t.Fatalf("unexpected error on call %d: %v", i+1, err)
		}
	}

	// 10th consecutive identical call triggers Dead-Man's Switch
	err := watchdog.RecordActivity(ctx, tenantID, agentID, "query_database", "dummy_root")
	if err == nil {
		t.Fatalf("expected Dead-Man's Switch error on 10th consecutive tool call, got nil")
	}
}
