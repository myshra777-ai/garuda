// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// handleCreateTask implements `garuda.create_task` (BUG-13 / CLASS-05).
// Creates a persisted task UUID (and provisions optional owner/target agent rows)
// so callers can pass real UUIDs into `garuda.handoff` and `garuda.resume`.
func (s *MCPServer) handleCreateTask(args map[string]interface{}) (interface{}, error) {
	if ReadOnlyMode() {
		return ReadOnlyRejection("garuda.create_task"), nil
	}
	ctx := context.Background()
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		tenantID = defaultTenantID
		if rawTid, ok := args["tenant_id"].(string); ok && strings.TrimSpace(rawTid) != "" {
			if parsed, pErr := uuid.Parse(strings.TrimSpace(rawTid)); pErr == nil {
				tenantID = parsed
			}
		}
	}
	wsName := s.resolveMCPWorkspaceName(args, workspaceID)

	title, _ := args["title"].(string)
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("title is required for garuda.create_task")
	}
	description, _ := args["description"].(string)
	scopeDomain, _ := args["scope_domain"].(string)
	scopeSystem, _ := args["scope_system"].(string)
	priority := 1
	if v, ok := args["priority"].(float64); ok && v >= 0 {
		priority = int(v)
	}

	taskID := uuid.New()
	if rawTaskID, ok := args["task_id"].(string); ok && strings.TrimSpace(rawTaskID) != "" {
		if parsed, pErr := uuid.Parse(strings.TrimSpace(rawTaskID)); pErr == nil {
			taskID = parsed
		}
	}

	ownerStr, _ := args["owner_agent_id"].(string)
	if ownerStr == "" {
		ownerStr, _ = args["agent_id"].(string)
	}
	var ownerID uuid.UUID
	if strings.TrimSpace(ownerStr) != "" {
		ownerID, _ = resolveOrDeriveUUID(tenantID, "agent", ownerStr)
	} else {
		ownerID = uuid.New()
		ownerStr = "agent-" + ownerID.String()[:8]
	}

	targetStr, _ := args["target_agent_id"].(string)
	var targetID uuid.UUID
	if strings.TrimSpace(targetStr) != "" {
		targetID, _ = resolveOrDeriveUUID(tenantID, "agent", targetStr)
	} else {
		targetID = uuid.New()
		targetStr = "agent-" + targetID.String()[:8]
	}

	var descPtr *string
	if strings.TrimSpace(description) != "" {
		trimmed := strings.TrimSpace(description)
		descPtr = &trimmed
	}

	now := time.Now().UTC()
	if s.store != nil && s.store.Pool() != nil {
		_, _ = s.store.Pool().Exec(ctx, `
			INSERT INTO tenants (id, name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING
		`, tenantID, "tenant-"+tenantID.String()[:8])

		_, _ = s.store.Pool().Exec(ctx, `
			INSERT INTO agents (id, tenant_id, name, model_type, session_id, status)
			VALUES ($1, $2, $3, 'mcp-agent', $4, 'idle'),
			       ($5, $2, $6, 'mcp-agent', $7, 'idle')
			ON CONFLICT (id) DO UPDATE SET status = 'idle'
		`, ownerID, tenantID, ownerStr, "sess-"+ownerID.String()[:8], targetID, targetStr, "sess-"+targetID.String()[:8])

		taskObj := &types.Task{
			ID:           taskID,
			TenantID:     tenantID,
			Title:        title,
			Description:  descPtr,
			Status:       types.TaskStatus("in_progress"),
			Priority:     priority,
			OwnerAgentID: &ownerID,
			ScopeDomain:  scopeDomain,
			ScopeSystem:  scopeSystem,
			Version:      1,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if saveErr := s.store.SaveTask(ctx, taskObj); saveErr != nil {
			_, _ = s.store.Pool().Exec(ctx, `
				INSERT INTO tasks (id, tenant_id, title, description, status, priority, owner_agent_id)
				VALUES ($1, $2, $3, $4, 'in_progress', $5, $6)
				ON CONFLICT (id) DO UPDATE SET owner_agent_id = EXCLUDED.owner_agent_id, status = 'in_progress', updated_at = NOW()
			`, taskID, tenantID, title, descPtr, priority, ownerID)
		}
	}

	return map[string]interface{}{
		"status":          "created",
		"task_id":         taskID.String(),
		"title":           title,
		"description":     description,
		"workspace":       wsName,
		"tenant_id":       tenantID.String(),
		"owner_agent_id":  ownerID.String(),
		"source_agent_id": ownerID.String(),
		"target_agent_id": targetID.String(),
		"created_at":      now.Format(time.RFC3339),
		"next_step_hint":  fmt.Sprintf("Pass task_id=%q, source_agent_id=%q, target_agent_id=%q to garuda.handoff", taskID.String(), ownerID.String(), targetID.String()),
	}, nil
}
