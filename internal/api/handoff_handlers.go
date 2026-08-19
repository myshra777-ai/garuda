// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/telemetry"
	"github.com/myshra777-ai/garuda/internal/types"
)

// HandleHandoff processes POST /api/v1/agents/handoff
func (s *Server) HandleHandoff(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := getRequestID(r)

	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok {
		tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}

	var req store.HandoffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid json payload", requestID)
		return
	}

	// Validate required fields
	if req.TaskID == uuid.Nil || req.SourceAgentID == uuid.Nil || req.TargetAgentID == uuid.Nil {
		s.RespondWithError(w, http.StatusUnprocessableEntity, "task_id, source_agent_id, and target_agent_id are required", requestID)
		return
	}

	// Bind tenant ID from authenticated context
	req.TenantID = tenantID

	modelName, modelProvider := getModelInfo(r)

	// Execute atomic, crash-safe handoff transaction (optional on store)
	resp := (*store.HandoffResponse)(nil)
	var err error

	if ps, ok := s.store.(interface {
		ExecuteHandoffTransaction(ctx context.Context, req *store.HandoffRequest) (*store.HandoffResponse, error)
	}); ok {
		resp, err = ps.ExecuteHandoffTransaction(r.Context(), &req)
		if err != nil {
			telemetry.RecordHandoffWithModel(
				modelName,
				modelProvider,
				float64(time.Since(start).Milliseconds()),
				false,
			)

			slog.Error("handoff execution failed",
				"error", err,
				"request_id", requestID,
				"task_id", req.TaskID,
				"tenant_id", tenantID,
			)
			s.RespondWithError(w, http.StatusConflict, "handoff execution failed", requestID)
			return
		}
	} else {
		telemetry.RecordHandoffWithModel(
			modelName,
			modelProvider,
			float64(time.Since(start).Milliseconds()),
			false,
		)
		s.RespondWithError(w, http.StatusNotImplemented, "store does not support handoff operations", requestID)
		return
	}

	success := err == nil
	telemetry.RecordHandoffWithModel(
		modelName,
		modelProvider,
		float64(time.Since(start).Milliseconds()),
		success,
	)

	// Broadcast live handoff telemetry event via SSE
	if s.sseBroker != nil {
		s.sseBroker.Publish(telemetry.EventType("agent_handoff"), map[string]interface{}{
			"handoff_id":      resp.HandoffID,
			"checkpoint_id":   resp.CheckpointID,
			"task_id":         resp.TaskID,
			"source_agent_id": req.SourceAgentID,
			"target_agent_id": req.TargetAgentID,
			"reason":          req.Reason,
			"status":          resp.Status,
			"request_id":      requestID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// HandleResume processes POST /api/v1/agents/resume
func (s *Server) HandleResume(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)

	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok {
		tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}

	var req struct {
		AgentID      uuid.UUID `json:"agent_id"`
		CheckpointID uuid.UUID `json:"checkpoint_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid json payload", requestID)
		return
	}

	if req.AgentID == uuid.Nil || req.CheckpointID == uuid.Nil {
		s.RespondWithError(w, http.StatusUnprocessableEntity, "agent_id and checkpoint_id are required", requestID)
		return
	}

	// Resume agent via optional store method
	var restoredState interface{}
	if rs, ok := s.store.(interface {
		ResumeAgent(ctx context.Context, tenantID, agentID, checkpointID uuid.UUID) (interface{}, error)
	}); ok {
		var err error
		restoredState, err = rs.ResumeAgent(r.Context(), tenantID, req.AgentID, req.CheckpointID)
		if err != nil {
			slog.Error("resume agent failed",
				"error", err,
				"request_id", requestID,
				"agent_id", req.AgentID,
				"checkpoint_id", req.CheckpointID,
			)
			s.RespondWithError(w, http.StatusNotFound, "failed to resume agent execution state", requestID)
			return
		}
	} else {
		s.RespondWithError(w, http.StatusNotImplemented, "store does not support resume operations", requestID)
		return
	}

	// Broadcast resume event
	if s.sseBroker != nil {
		s.sseBroker.Publish(telemetry.EventType("agent_resumed"), map[string]interface{}{
			"agent_id":      req.AgentID,
			"checkpoint_id": req.CheckpointID,
			"status":        "working",
			"request_id":    requestID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id":      req.AgentID,
		"checkpoint_id": req.CheckpointID,
		"status":        "working",
		"state":         restoredState,
		"request_id":    requestID,
	})
}

// HandleGetLineage processes GET /api/v1/tasks/{task_id}/lineage
func (s *Server) HandleGetLineage(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)

	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok {
		tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}

	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["task_id"])
	if err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid task_id UUID", requestID)
		return
	}

	// Optional lineage query
	var edges []types.LineageEdge
	if ls, ok := s.store.(interface {
		GetLineageDAG(ctx context.Context, tenantID, taskID uuid.UUID) ([]types.LineageEdge, error)
	}); ok {
		var err error
		edges, err = ls.GetLineageDAG(r.Context(), tenantID, taskID)
		if err != nil {
			slog.Error("failed to query lineage DAG",
				"error", err,
				"request_id", requestID,
				"task_id", taskID,
				"tenant_id", tenantID,
			)
			s.RespondWithError(w, http.StatusInternalServerError, "failed to query lineage DAG", requestID)
			return
		}
	} else {
		s.RespondWithError(w, http.StatusNotImplemented, "store does not support lineage queries", requestID)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"task_id":    taskID,
		"edges":      edges,
		"total":      len(edges),
		"request_id": requestID,
	})
}
