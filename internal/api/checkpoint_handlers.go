// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/telemetry"
	"github.com/myshra777-ai/garuda/internal/types"
)

// CheckpointRequest defines the payload for saving a checkpoint across CLI and API calls.
type CheckpointRequest struct {
	AgentID        string                 `json:"agent_id"`
	CheckpointName string                 `json:"checkpoint_name,omitempty"`
	Reason         string                 `json:"reason,omitempty"`
	TaskID         *uuid.UUID             `json:"task_id,omitempty"`
	CheckpointData json.RawMessage        `json:"checkpoint_data,omitempty"`
	Data           map[string]interface{} `json:"data,omitempty"`
	TTLSeconds     int                    `json:"ttl_seconds,omitempty"`
}

// HandleAgentCheckpoint saves or updates an agent's runtime state.
func (s *Server) HandleAgentCheckpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	start := time.Now()
	telemetry.RecordFeatureUsage("agent_checkpoint")

	tenantID, err := resolveTenantID(r, uuid.MustParse("00000000-0000-0000-0000-000000000001"))
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

	chkName := req.CheckpointName
	if chkName == "" {
		chkName = "manual_checkpoint"
	}

	// 1. Construct guaranteed fallback payload if CheckpointData is empty
	var finalData []byte
	if len(req.CheckpointData) > 0 && bytes.TrimSpace(req.CheckpointData) != nil && string(req.CheckpointData) != "null" {
		finalData = req.CheckpointData
	} else if req.Data != nil {
		req.Data["checkpoint_name"] = chkName
		if req.Reason != "" {
			req.Data["reason"] = req.Reason
		}
		marshaled, err := json.Marshal(req.Data)
		if err == nil {
			finalData = marshaled
		}
	}

	// Fallback to minimal JSON map if still empty
	if len(finalData) == 0 {
		reason := req.Reason
		if reason == "" {
			reason = "manual_cli_trigger"
		}

		fallbackMap := map[string]interface{}{
			"agent_id":        req.AgentID,
			"checkpoint_name": chkName,
			"reason":          reason,
			"timestamp":       time.Now().UTC().Format(time.RFC3339),
		}
		finalData, _ = json.Marshal(fallbackMap)
	}

	now := time.Now().UTC()

	// Derives deterministic UUID for named checkpoints to trigger ON CONFLICT (id) updates
	checkpointID := uuid.NewSHA1(tenantID, []byte(chkName))

	checkpoint := &types.Checkpoint{
		ID:             checkpointID,
		TenantID:       tenantID,
		AgentID:        req.AgentID,
		CheckpointName: chkName,
		TaskID:         req.TaskID,
		CheckpointData: finalData,
		Status:         "active",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if req.TTLSeconds > 0 {
		expiry := now.Add(time.Duration(req.TTLSeconds) * time.Second)
		checkpoint.ExpiresAt = &expiry
	} else {
		expiry := now.Add(30 * 24 * time.Hour) // 30-day default retention
		checkpoint.ExpiresAt = &expiry
	}

	// 2. Persist to storage and handle ON CONFLICT / unique constraint errors cleanly
	if err := s.store.SaveCheckpoint(r.Context(), checkpoint); err != nil {
		telemetry.RecordError("db_save_failed", err.Error())
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "on conflict") {
			s.RespondWithError(w, http.StatusConflict, "checkpoint conflict: record already exists")
			return
		}
		s.RespondWithError(w, http.StatusInternalServerError, "failed to save checkpoint: "+err.Error())
		return
	}

	telemetry.RecordAPILatency(time.Since(start))

	// Fetch current Merkle root for response attestation if available
	merkleRoot := "GENESIS"
	if latestSnap, err := s.store.GetLatestMerkleSnapshot(r.Context(), tenantID); err == nil && latestSnap != nil {
		merkleRoot = latestSnap.SnapshotHash
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":               checkpoint.ID.String(),
		"status":           checkpoint.Status,
		"agent_id":         checkpoint.AgentID,
		"checkpoint_name":  chkName,
		"merkle_root_hash": merkleRoot,
		"created_at":       checkpoint.CreatedAt.Format(time.RFC3339),
	})
}

// HandleGetAgentCheckpoint retrieves a saved checkpoint by ID.
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

	tenantID, err := resolveTenantID(r, uuid.MustParse("00000000-0000-0000-0000-000000000001"))
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

// HandleAgentResume loads a checkpoint and returns it for task resumption.
func (s *Server) HandleAgentResume(w http.ResponseWriter, r *http.Request) {
	s.HandleGetAgentCheckpoint(w, r)
}

// HandleAgentHandoff transfers a task checkpoint from one agent context to another.
func (s *Server) HandleAgentHandoff(w http.ResponseWriter, r *http.Request) {
	type HandoffRequest struct {
		CheckpointID uuid.UUID `json:"checkpoint_id"`
		FromAgentID  string    `json:"from_agent_id"`
		ToAgentID    string    `json:"to_agent_id"`
	}

	tenantID, err := resolveTenantID(r, uuid.MustParse("00000000-0000-0000-0000-000000000001"))
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
		CheckpointName: checkpoint.CheckpointName + "-handoff",
		TaskID:         checkpoint.TaskID,
		CheckpointData: checkpoint.CheckpointData,
		Status:         "active",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.store.SaveCheckpoint(r.Context(), newCheckpoint); err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, "failed to handoff checkpoint: "+err.Error())
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
