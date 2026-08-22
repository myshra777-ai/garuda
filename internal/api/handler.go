// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/engine"
	"github.com/myshra777-ai/garuda/internal/telemetry"
	"github.com/myshra777-ai/garuda/internal/topology"
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

// DashboardStatsResponse defines the structured JSON response for dashboard analytics.
type DashboardStatsResponse struct {
	TenantID             string  `json:"tenant_id"`
	TotalDecisions       int     `json:"total_decisions"`
	ActiveContradictions int     `json:"active_contradictions"`
	TokensSaved          int64   `json:"tokens_saved"`
	CostSaved            float64 `json:"cost_saved"`
	LatestMerkleHash     string  `json:"latest_merkle_hash"`
	ParentMerkleHash     string  `json:"parent_merkle_hash"`
	LatestBlockHeight    int64   `json:"latest_block_height"`
	Timestamp            string  `json:"timestamp"`
}

// EvaluateRouteRequest payload for testing or routing policy enforcement.
type EvaluateRouteRequest struct {
	TenantID    string `json:"tenant_id"`
	ScopeDomain string `json:"scope_domain"`
	ScopeSystem string `json:"scope_system"`
	Action      string `json:"action"`
}

// Server holds dependencies for network API handlers.
type Server struct {
	store               types.DecisionStore
	authService         *auth.AuthService
	jwtConfig           *auth.JWTConfig
	contradictionEngine *engine.ContradictionEngine
	lineageEngine       *engine.LineageEngine
	topologyGenerator   *topology.Generator
	topologyExecutor    *topology.Executor
	sseBroker           *telemetry.SSEBroker
	router              *mux.Router
	checkpointMu        sync.RWMutex
	checkpointStore     map[string]CheckpointRecord
}

// NewServer creates a new API server instance with initialized state, router, and SSE broker.
func NewServer(
	store types.DecisionStore,
	authService *auth.AuthService,
	jwtConfig *auth.JWTConfig,
	contradictionEngine *engine.ContradictionEngine,
	lineageEngine *engine.LineageEngine,
	topologyGenerator *topology.Generator,
	topologyExecutor *topology.Executor,
) *Server {
	sseBroker := telemetry.NewSSEBroker()

	s := &Server{
		store:               store,
		authService:         authService,
		jwtConfig:           jwtConfig,
		contradictionEngine: contradictionEngine,
		lineageEngine:       lineageEngine,
		topologyGenerator:   topologyGenerator,
		topologyExecutor:    topologyExecutor,
		sseBroker:           sseBroker,
		router:              mux.NewRouter(),
		checkpointStore:     make(map[string]CheckpointRecord),
	}

	s.RegisterRoutes(s.router)
	return s
}

// RegisterRoutes sets up HTTP routing and subrouter middleware hierarchy.
func (s *Server) RegisterRoutes(r *mux.Router) {
	// =========================================================================
	// PUBLIC ROUTES (No JWT Bearer Token Required)
	// =========================================================================
	r.HandleFunc("/health", s.HandleHealth).Methods(http.MethodGet)
	r.HandleFunc("/dashboard", s.HandleDashboard).Methods(http.MethodGet)
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}).Methods(http.MethodGet)
	r.HandleFunc("/debug/token", s.HandleDebugToken).Methods(http.MethodGet)

	// Public Dashboard Telemetry, Graph & Global Search Endpoints
	r.HandleFunc("/api/v1/dashboard/stats", s.HandleDashboardStats).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/dashboard/search", s.HandleDashboardSearch).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/graph", s.HandleGraph).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/events", s.HandleLiveEvents).Methods(http.MethodGet)

	// =========================================================================
	// PROTECTED API SUBROUTER (JWT Auth Required for Sensitive Actions)
	// =========================================================================
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(s.AuthMiddleware)

	// Live Telemetry SSE Stream Endpoint
	api.Handle("/telemetry/stream", s.sseBroker).Methods(http.MethodGet)

	// Pre-Flight Classification Router Evaluation
	api.HandleFunc("/router/evaluate", s.HandleEvaluateRoute).Methods(http.MethodPost, http.MethodOptions)

	// Multi-Agent Execution & Lineage
	api.HandleFunc("/agents/handoff", s.HandleHandoff).Methods(http.MethodPost)
	api.HandleFunc("/agents/resume", s.HandleResume).Methods(http.MethodPost)
	api.HandleFunc("/tasks/{task_id}/lineage", s.HandleGetLineage).Methods(http.MethodGet)

	// Audit Log & Compliance Endpoints
	api.HandleFunc("/audit/export", s.HandleExportAuditLogs).Methods(http.MethodGet)
	api.HandleFunc("/audit/verify/{id}", s.HandleVerifyAuditLog).Methods(http.MethodGet)
}

// ServeHTTP implements http.Handler for the Server struct.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// RespondWithError writes a standard JSON error response to the client.

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

// HandleVerifyAuditLog verifies a given audit event ID against the tenant's active Merkle root.
func (s *Server) HandleVerifyAuditLog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventIDStr := vars["id"]
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid audit event id format")
		return
	}

	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id is required")
		return
	}

	verification, err := s.store.VerifyAuditEvent(r.Context(), tenantID, eventID)
	if err != nil {
		s.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(verification)
}
