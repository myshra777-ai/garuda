// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/tenant"
)

// recordSessionStart inserts a row into mcp_sessions and stores the
// returned id on the server. Best-effort: an error is logged at WARN
// and the handshake proceeds.
func (s *MCPServer) recordSessionStart(ctx context.Context, clientName, clientVersion, agentID string) {
	if s.store == nil {
		return
	}
	var tenantID uuid.UUID = tenant.CanonicalID
	var workspaceID *uuid.UUID

	// Resolve workspace from env if set. A nil workspace is fine —
	// the session simply has no workspace binding.
	if wsName := os.Getenv("GARUDA_WORKSPACE"); wsName != "" {
		if id, err := store.ResolveWorkspaceID(ctx, s.store.Pool(), tenantID, wsName); err == nil && id != uuid.Nil {
			workspaceID = &id
		}
	}

	writeCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	var id uuid.UUID
	err := s.store.Pool().QueryRow(writeCtx, `
        INSERT INTO mcp_sessions (
            tenant_id, workspace_id, session_token, client_name, client_version, agent_id
        ) VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `, tenantID, workspaceID, s.sessionID, clientName, nullIfEmpty(clientVersion), agentID).Scan(&id)
	if err != nil {
		slog.Warn("mcp session insert failed; continuing without session tracking", "error", err)
		return
	}

	s.sessionMu.Lock()
	s.sessionDBID = id
	s.sessionMu.Unlock()
}

// recordToolCall inserts a row into mcp_tool_calls and updates the
// parent session. Best-effort. Never blocks the caller for more than
// 500ms.
func (s *MCPServer) recordToolCall(ctx context.Context, toolName string, duration time.Duration, toolErr error, args map[string]interface{}) {
	s.sessionMu.RLock()
	sessionID := s.sessionDBID
	s.sessionMu.RUnlock()

	if sessionID == uuid.Nil || s.store == nil {
		// No session row: either the insert failed or we're in the
		// orphaned window. Skip silently.
		return
	}

	status := "ok"
	var errMsg interface{}
	if toolErr != nil {
		status = "error"
		errMsg = toolErr.Error()
	}

	argsSummary, _ := json.Marshal(filterArgs(toolName, args))

	writeCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	tx, err := s.store.Pool().Begin(writeCtx)
	if err != nil {
		slog.Warn("mcp tool call tracking: begin failed", "error", err)
		return
	}
	defer tx.Rollback(writeCtx)

	if _, err := tx.Exec(writeCtx, `
        INSERT INTO mcp_tool_calls (
            session_id, tool_name, duration_ms, status, error_message, args_summary
        ) VALUES ($1, $2, $3, $4, $5, $6)
    `, sessionID, toolName, duration.Milliseconds(), status, errMsg, argsSummary); err != nil {
		slog.Warn("mcp tool call insert failed", "error", err)
		return
	}

	if _, err := tx.Exec(writeCtx, `
        UPDATE mcp_sessions
           SET last_activity_at = NOW(),
               close_reason    = CASE WHEN closed_at IS NOT NULL THEN 'reactivated' ELSE close_reason END,
               closed_at       = NULL,
               request_count   = request_count + 1,
               tool_call_count = tool_call_count + CASE WHEN $2 = 'ok' THEN 1 ELSE 0 END
         WHERE id = $1
    `, sessionID, status); err != nil {
		slog.Warn("mcp session update failed", "error", err)
		return
	}

	if err := tx.Commit(writeCtx); err != nil {
		slog.Warn("mcp tool call tracking: commit failed", "error", err)
	}
}

// closeSession marks the session closed with the given reason. The
// WHERE closed_at IS NULL guard means the first close wins — the
// signal handler and the EOF path can both call this without
// racing on the timestamp.
func (s *MCPServer) closeSession(ctx context.Context, reason string) {
	s.sessionMu.RLock()
	sessionID := s.sessionDBID
	s.sessionMu.RUnlock()

	if sessionID == uuid.Nil || s.store == nil {
		return
	}

	writeCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	if _, err := s.store.Pool().Exec(writeCtx, `
        UPDATE mcp_sessions
           SET closed_at = NOW(), close_reason = $2
         WHERE id = $1 AND closed_at IS NULL
    `, sessionID, reason); err != nil {
		slog.Warn("mcp session close failed", "error", err)
	}
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
