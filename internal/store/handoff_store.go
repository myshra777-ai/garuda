package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/myshra777-ai/garuda/internal/types"
)

// ============================================================
// Handoff Request & Response Types
// ============================================================

type HandoffRequest struct {
	TenantID       uuid.UUID   `json:"tenant_id"`
	TaskID         uuid.UUID   `json:"task_id"`
	SourceAgentID  uuid.UUID   `json:"source_agent_id"`
	TargetAgentID  uuid.UUID   `json:"target_agent_id"`
	Reason         string      `json:"reason,omitempty"`
	CheckpointData interface{} `json:"checkpoint_data"` // serializable context
}

type HandoffResponse struct {
	HandoffID    uuid.UUID `json:"handoff_id"`
	CheckpointID uuid.UUID `json:"checkpoint_id"`
	TaskID       uuid.UUID `json:"task_id"`
	Status       string    `json:"status"`
}

// ============================================================
// Handoff Store Methods
// ============================================================

// ExecuteHandoffTransaction performs an atomic, crash‑safe handoff.
func (s *PostgresStore) ExecuteHandoffTransaction(ctx context.Context, req *HandoffRequest) (*HandoffResponse, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Lock and validate source agent
	sourceAgent, err := s.lockAgent(ctx, tx, req.SourceAgentID, req.TenantID)
	if err != nil {
		return nil, err
	}
	if sourceAgent.Status == "transitioning" {
		return nil, errors.New("source agent is already in transition")
	}

	// 2. Lock and validate target agent
	targetAgent, err := s.lockAgent(ctx, tx, req.TargetAgentID, req.TenantID)
	if err != nil {
		return nil, err
	}
	if targetAgent.Status == "offline" {
		return nil, errors.New("target agent is offline")
	}

	// 3. Lock and validate task
	task, err := s.lockTask(ctx, tx, req.TaskID, req.TenantID)
	if err != nil {
		return nil, err
	}
	if task.OwnerAgentID == nil || *task.OwnerAgentID != req.SourceAgentID {
		return nil, errors.New("source agent does not own this task")
	}

	// 4. Mark source agent as transitioning (prevents concurrent handoffs)
	if err := s.updateAgentStatus(ctx, tx, req.SourceAgentID, "transitioning"); err != nil {
		return nil, err
	}

	// 5. Serialize checkpoint data with CAS deduplication
	checkpointDataJSON, err := json.Marshal(req.CheckpointData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal checkpoint data: %w", err)
	}
	stateHash := computeStateHash(req.TenantID, req.SourceAgentID, checkpointDataJSON)

	// 6. Deduplicate checkpoint using evidence_blocks (CAS)
	if err := s.storeCheckpointInCAS(ctx, tx, stateHash, checkpointDataJSON); err != nil {
		return nil, err
	}

	// 7. Create checkpoint record
	checkpointID := uuid.New()
	checkpointQuery := `
		INSERT INTO agent_checkpoints (id, tenant_id, task_id, agent_id, checkpoint_data, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'active', NOW(), NOW())
	`
	if _, err := tx.Exec(ctx, checkpointQuery,
		checkpointID, req.TenantID, req.TaskID, req.SourceAgentID.String(),
		checkpointDataJSON,
	); err != nil {
		return nil, fmt.Errorf("failed to create checkpoint: %w", err)
	}

	// 8. Create handoff record (status = 'in_progress')
	handoffID := uuid.New()
	handoffQuery := `
		INSERT INTO handoffs (id, tenant_id, task_id, source_agent_id, target_agent_id, checkpoint_id, reason, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'in_progress', NOW())
	`
	if _, err := tx.Exec(ctx, handoffQuery,
		handoffID, req.TenantID, req.TaskID, req.SourceAgentID,
		req.TargetAgentID, checkpointID, req.Reason,
	); err != nil {
		return nil, fmt.Errorf("failed to create handoff record: %w", err)
	}

	// 9. Update task ownership to target agent
	taskUpdateQuery := `
		UPDATE tasks
		SET owner_agent_id = $1, updated_at = NOW(), version = version + 1
		WHERE id = $2 AND tenant_id = $3
	`
	if _, err := tx.Exec(ctx, taskUpdateQuery, req.TargetAgentID, req.TaskID, req.TenantID); err != nil {
		return nil, fmt.Errorf("failed to update task ownership: %w", err)
	}

	// 10. Add lineage edge (handoff type)
	edgeQuery := `
		INSERT INTO lineage_edges (id, tenant_id, source_task_id, target_task_id, edge_type, handoff_id, created_at)
		VALUES ($1, $2, $3, $4, 'handoff', $5, NOW())
	`
	edgeID := uuid.New()
	if _, err := tx.Exec(ctx, edgeQuery,
		edgeID, req.TenantID, req.TaskID, req.TaskID, handoffID,
	); err != nil {
		return nil, fmt.Errorf("failed to create lineage edge: %w", err)
	}

	// 11. Update source agent status to 'paused'
	if err := s.updateAgentStatus(ctx, tx, req.SourceAgentID, "paused"); err != nil {
		return nil, err
	}

	// 12. Update target agent status to 'working'
	if err := s.updateAgentStatus(ctx, tx, req.TargetAgentID, "working"); err != nil {
		return nil, err
	}

	// 13. Mark handoff as completed
	completeQuery := `
		UPDATE handoffs SET status = 'completed', completed_at = NOW()
		WHERE id = $1
	`
	if _, err := tx.Exec(ctx, completeQuery, handoffID); err != nil {
		return nil, fmt.Errorf("failed to complete handoff: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit handoff transaction: %w", err)
	}

	return &HandoffResponse{
		HandoffID:    handoffID,
		CheckpointID: checkpointID,
		TaskID:       req.TaskID,
		Status:       "completed",
	}, nil
}

// ============================================================
// Resume Agent from Checkpoint
// ============================================================

func (s *PostgresStore) ResumeAgent(ctx context.Context, tenantID, agentID, checkpointID uuid.UUID) (interface{}, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Lock agent
	_, err = s.lockAgent(ctx, tx, agentID, tenantID)
	if err != nil {
		return nil, err
	}

	// 2. Fetch checkpoint
	var checkpointData []byte
	checkpointQuery := `
		SELECT checkpoint_data
		FROM agent_checkpoints
		WHERE id = $1 AND tenant_id = $2 AND status = 'active'
		FOR UPDATE
	`
	err = tx.QueryRow(ctx, checkpointQuery, checkpointID, tenantID).Scan(&checkpointData)
	if err != nil {
		return nil, fmt.Errorf("checkpoint not found: %w", err)
	}

	// 3. Mark checkpoint as restored
	restoreQuery := `
		UPDATE agent_checkpoints SET status = 'restored', updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`
	if _, err := tx.Exec(ctx, restoreQuery, checkpointID, tenantID); err != nil {
		return nil, err
	}

	// 4. Update agent status to 'working'
	if err := s.updateAgentStatus(ctx, tx, agentID, "working"); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// 5. Return restored data
	var restored interface{}
	if err := json.Unmarshal(checkpointData, &restored); err != nil {
		return nil, err
	}
	return restored, nil
}

// ============================================================
// Lineage DAG Query with Cycle & Depth Protection
// ============================================================

func (s *PostgresStore) GetLineageDAG(ctx context.Context, tenantID, taskID uuid.UUID) ([]types.LineageEdge, error) {
	query := `
		WITH RECURSIVE lineage_dag AS (
			-- Anchor: start from the given task
			SELECT 
				e.source_task_id,
				e.target_task_id,
				e.edge_type,
				e.handoff_id,
				1 AS depth,
				ARRAY[e.source_task_id] AS path
			FROM lineage_edges e
			WHERE e.source_task_id = $1 AND e.tenant_id = $2

			UNION ALL

			-- Recursive: traverse downstream
			SELECT 
				e.source_task_id,
				e.target_task_id,
				e.edge_type,
				e.handoff_id,
				d.depth + 1,
				d.path || e.source_task_id
			FROM lineage_edges e
			JOIN lineage_dag d ON e.source_task_id = d.target_task_id
			WHERE e.tenant_id = $2
			  AND NOT (e.source_task_id = ANY(d.path)) -- Cycle Guard
			  AND d.depth < 25                          -- Depth Guard
		)
		SELECT source_task_id, target_task_id, edge_type, handoff_id, depth
		FROM lineage_dag
		ORDER BY depth
	`

	rows, err := s.pool.Query(ctx, query, taskID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query lineage DAG: %w", err)
	}
	defer rows.Close()
	var edges []types.LineageEdge
	for rows.Next() {
		var e types.LineageEdge
		var handoffID *uuid.UUID

		if err := rows.Scan(&e.SourceTaskID, &e.TargetTaskID, &e.EdgeType, &handoffID, &e.Depth); err != nil {
			return nil, err
		}
		if handoffID != nil {
			e.HandoffID = handoffID
		}
		edges = append(edges, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return edges, nil
}

// ============================================================
// Private Helpers
// ============================================================

func (s *PostgresStore) lockAgent(ctx context.Context, tx pgx.Tx, agentID, tenantID uuid.UUID) (*types.Agent, error) {
	query := `
		SELECT id, tenant_id, name, model_type, session_id, status, current_task_id, last_heartbeat, metadata
		FROM agents
		WHERE id = $1 AND tenant_id = $2
		FOR UPDATE
	`
	var a types.Agent
	err := tx.QueryRow(ctx, query, agentID, tenantID).Scan(
		&a.ID, &a.TenantID, &a.Name, &a.ModelType, &a.SessionID,
		&a.Status, &a.CurrentTaskID, &a.LastHeartbeat, &a.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("agent %s not found: %w", agentID, err)
	}
	return &a, nil
}

func (s *PostgresStore) lockTask(ctx context.Context, tx pgx.Tx, taskID, tenantID uuid.UUID) (*types.Task, error) {
	query := `
		SELECT id, tenant_id, title, description, status, priority,
		       owner_agent_id, parent_task_id, scope_domain, scope_system,
		       version, created_at, updated_at, completed_at
		FROM tasks
		WHERE id = $1 AND tenant_id = $2
		FOR UPDATE
	`
	var t types.Task
	err := tx.QueryRow(ctx, query, taskID, tenantID).Scan(
		&t.ID, &t.TenantID, &t.Title, &t.Description, &t.Status,
		&t.Priority, &t.OwnerAgentID, &t.ParentTaskID, &t.ScopeDomain,
		&t.ScopeSystem, &t.Version, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt,
	)
	return &t, err
}

func (s *PostgresStore) updateAgentStatus(ctx context.Context, tx pgx.Tx, agentID uuid.UUID, status string) error {
	query := `
		UPDATE agents SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := tx.Exec(ctx, query, status, agentID)
	return err
}

func (s *PostgresStore) storeCheckpointInCAS(ctx context.Context, tx pgx.Tx, stateHash string, data []byte) error {
	query := `
		INSERT INTO evidence_blocks (block_hash, content, ref_count, created_at)
		VALUES ($1, $2, 1, NOW())
		ON CONFLICT (block_hash) DO UPDATE SET ref_count = evidence_blocks.ref_count + 1
	`
	_, err := tx.Exec(ctx, query, stateHash, data)
	return err
}

func computeStateHash(tenantID, agentID uuid.UUID, data []byte) string {
	h := sha256.New()
	h.Write([]byte(tenantID.String()))
	h.Write([]byte(agentID.String()))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// ListHandoffsByScope returns handoffs linked to tasks in a given scope.
func (s *PostgresStore) ListHandoffsByScope(ctx context.Context, tenantID uuid.UUID, domain, system string) ([]*types.HandoffRecord, error) {
	query := `
		SELECT h.id, h.tenant_id, h.task_id, h.source_agent_id, h.target_agent_id,
		       h.reason, h.status, h.created_at, h.completed_at
		FROM handoffs h
		JOIN tasks t ON h.task_id = t.id
		WHERE h.tenant_id = $1 AND t.scope_domain = $2 AND t.scope_system = $3
		ORDER BY h.created_at DESC
	`
	rows, err := s.pool.Query(ctx, query, tenantID, domain, system)
	if err != nil {
		return nil, fmt.Errorf("failed to list handoffs: %w", err)
	}
	defer rows.Close()

	var handoffs []*types.HandoffRecord
	for rows.Next() {
		var h types.HandoffRecord
		err := rows.Scan(&h.ID, &h.TenantID, &h.TaskID, &h.SourceAgentID, &h.TargetAgentID,
			&h.Reason, &h.Status, &h.CreatedAt, &h.CompletedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan handoff: %w", err)
		}
		handoffs = append(handoffs, &h)
	}
	return handoffs, nil
}
