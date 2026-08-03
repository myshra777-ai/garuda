package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/telemetry"
	"github.com/myshra777-ai/garuda/internal/types"
)

// CheckpointRequest defines the payload for saving a checkpoint.
type CheckpointRequest struct {
	AgentID        string          `json:"agent_id"`
	TaskID         *uuid.UUID      `json:"task_id,omitempty"`
	CheckpointData json.RawMessage `json:"checkpoint_data"`
	TTLSeconds     int             `json:"ttl_seconds,omitempty"`
}

// HandleAgentCheckpoint saves an agent's runtime state.
func (s *Server) HandleAgentCheckpoint(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	telemetry.RecordFeatureUsage("agent_checkpoint")

	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id is required")
		return
	}

	var req CheckpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		telemetry.RecordError("malformed_json", err.Error())
		s.RespondWithError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if req.AgentID == "" {
		s.RespondWithError(w, http.StatusUnprocessableEntity, "agent_id is required")
		return
	}

	now := time.Now().UTC()
	checkpoint := &types.Checkpoint{
		ID:             uuid.New(),
		TenantID:       tenantID,
		AgentID:        req.AgentID,
		TaskID:         req.TaskID,
		CheckpointData: req.CheckpointData,
		Status:         "active",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if req.TTLSeconds > 0 {
		expiry := now.Add(time.Duration(req.TTLSeconds) * time.Second)
		checkpoint.ExpiresAt = &expiry
	}

	if err := s.store.SaveCheckpoint(r.Context(), checkpoint); err != nil {
		telemetry.RecordError("db_save_failed", err.Error())
		s.RespondWithError(w, http.StatusInternalServerError, "failed to save checkpoint")
		return
	}

	telemetry.RecordAPILatency(time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         checkpoint.ID.String(),
		"status":     checkpoint.Status,
		"agent_id":   checkpoint.AgentID,
		"created_at": checkpoint.CreatedAt,
	})
}

// HandleGetAgentCheckpoint retrieves a saved checkpoint.
func (s *Server) HandleGetAgentCheckpoint(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 1 {
		s.RespondWithError(w, http.StatusBadRequest, "invalid request path")
		return
	}
	checkpointIDStr := pathParts[len(pathParts)-1]

	checkpointID, err := uuid.Parse(checkpointIDStr)
	if err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid checkpoint ID")
		return
	}

	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id is required")
		return
	}

	checkpoint, err := s.store.GetCheckpoint(r.Context(), tenantID, checkpointID)
	if err != nil {
		s.RespondWithError(w, http.StatusNotFound, "checkpoint not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(checkpoint)
}

// HandleAgentResume loads a checkpoint and returns it for resumption.
func (s *Server) HandleAgentResume(w http.ResponseWriter, r *http.Request) {
	s.HandleGetAgentCheckpoint(w, r)
}

// HandleAgentHandoff transfers a task from one agent to another.
func (s *Server) HandleAgentHandoff(w http.ResponseWriter, r *http.Request) {
	type HandoffRequest struct {
		CheckpointID uuid.UUID `json:"checkpoint_id"`
		FromAgentID  string    `json:"from_agent_id"`
		ToAgentID    string    `json:"to_agent_id"`
	}

	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id is required")
		return
	}

	var req HandoffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if req.CheckpointID == uuid.Nil || req.ToAgentID == "" {
		s.RespondWithError(w, http.StatusUnprocessableEntity, "checkpoint_id and to_agent_id are required")
		return
	}

	checkpoint, err := s.store.GetCheckpoint(r.Context(), tenantID, req.CheckpointID)
	if err != nil {
		s.RespondWithError(w, http.StatusNotFound, "checkpoint not found")
		return
	}

	now := time.Now().UTC()
	newCheckpoint := &types.Checkpoint{
		ID:             uuid.New(),
		TenantID:       tenantID,
		AgentID:        req.ToAgentID,
		TaskID:         checkpoint.TaskID,
		CheckpointData: checkpoint.CheckpointData,
		Status:         "active",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.store.SaveCheckpoint(r.Context(), newCheckpoint); err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, "failed to handoff checkpoint")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"new_checkpoint_id": newCheckpoint.ID.String(),
		"from_agent":        checkpoint.AgentID,
		"to_agent":          req.ToAgentID,
	})
}
