package api

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/engine"
	"github.com/myshra777-ai/garuda/internal/types"
)

// APIError represents a sanitized JSON error frame.
type APIError struct {
	Code      int       `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// CheckpointRecord tracks an agent execution checkpoint snapshot.
type CheckpointRecord struct {
	ID        string                 `json:"id"`
	Status    string                 `json:"status,omitempty"`
	State     map[string]interface{} `json:"state,omitempty"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// AgentCheckpointRequest defines the incoming payload for saving a checkpoint.
type AgentCheckpointRequest struct {
	ID     string                 `json:"id"`
	Status string                 `json:"status,omitempty"`
	State  map[string]interface{} `json:"state,omitempty"`
}

// Server holds dependencies for network API handlers.
type Server struct {
	store               types.DecisionStore
	authService         *auth.AuthService
	jwtConfig           *auth.JWTConfig
	contradictionEngine *engine.ContradictionEngine
	lineageEngine       *engine.LineageEngine
	checkpointMu        sync.RWMutex
	checkpointStore     map[string]CheckpointRecord
}

// NewServer creates a new API server instance with initialized state.
func NewServer(
	store types.DecisionStore,
	authService *auth.AuthService,
	jwtConfig *auth.JWTConfig,
	contradictionEngine *engine.ContradictionEngine,
	lineageEngine *engine.LineageEngine,
) *Server {
	return &Server{
		store:               store,
		authService:         authService,
		jwtConfig:           jwtConfig,
		contradictionEngine: contradictionEngine,
		lineageEngine:       lineageEngine,
		checkpointStore:     make(map[string]CheckpointRecord),
	}
}

// RespondWithError writes a standard JSON error response to the client.
func (s *Server) RespondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(APIError{
		Code:      code,
		Message:   msg,
		Timestamp: time.Now().UTC(),
	})
}

// HandleHealth returns gateway health status.
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// HandleDebugToken generates a tenant-scoped JWT for testing (disabled in production).
func (s *Server) HandleDebugToken(w http.ResponseWriter, r *http.Request) {
	// Guard: Never expose debug endpoint in production environments
	if os.Getenv("GARUDA_ENV") == "production" {
		s.RespondWithError(w, http.StatusNotFound, "debug endpoint disabled in production")
		return
	}

	if r.Method != http.MethodGet {
		s.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if s.jwtConfig == nil {
		s.RespondWithError(w, http.StatusInternalServerError, "jwt config unavailable")
		return
	}

	actor := r.URL.Query().Get("actor")
	if actor == "" {
		actor = "debug-user"
	}

	tenantIDStr := r.URL.Query().Get("tenant_id")
	var tenantID uuid.UUID
	if tenantIDStr != "" {
		var err error
		tenantID, err = uuid.Parse(tenantIDStr)
		if err != nil {
			s.RespondWithError(w, http.StatusBadRequest, "invalid tenant_id format")
			return
		}
	} else {
		// Deterministic tenant UUID derived from actor name
		tenantID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(actor))
	}

	token, err := s.jwtConfig.GenerateToken(actor, tenantID)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"token":     token,
		"tenant_id": tenantID.String(),
		"actor":     actor,
	})
}

// HandleAgentCheckpoint stores an agent execution state snapshot.
// REMOVE THESE DUPLICATE METHODS FROM internal/api/handler.go:

// func (s *Server) HandleAgentCheckpoint(w http.ResponseWriter, r *http.Request) { ... }
// func (s *Server) HandleGetAgentCheckpoint(w http.ResponseWriter, r *http.Request) { ... }
// HandleGetAgentCheckpoint retrieves a stored agent execution state snapshot.
