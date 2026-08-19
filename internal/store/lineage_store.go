// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// ListLineageEdgesByTasks returns all lineage edges for a list of task IDs.
func (s *PostgresStore) ListLineageEdgesByTasks(ctx context.Context, tenantID uuid.UUID, taskIDs []uuid.UUID) ([]types.LineageEdge, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	query := `
		SELECT source_task_id, target_task_id, edge_type, handoff_id, depth
		FROM lineage_edges
		WHERE tenant_id = $1 AND source_task_id = ANY($2)
		ORDER BY depth ASC
	`
	rows, err := s.pool.Query(ctx, query, tenantID, taskIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list lineage edges: %w", err)
	}
	defer rows.Close()

	var edges []types.LineageEdge
	for rows.Next() {
		var e types.LineageEdge
		var handoffID *uuid.UUID
		err := rows.Scan(&e.SourceTaskID, &e.TargetTaskID, &e.EdgeType, &handoffID, &e.Depth)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lineage edge: %w", err)
		}
		if handoffID != nil {
			e.HandoffID = handoffID
		}
		edges = append(edges, e)
	}
	return edges, nil
}

// GetLineageDAG returns the full lineage DAG for a single task (recursive).
