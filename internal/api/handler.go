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
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/engine"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/telemetry"
	"github.com/myshra777-ai/garuda/internal/tenant"
	"github.com/myshra777-ai/garuda/internal/topology"
	"github.com/myshra777-ai/garuda/internal/types"
)

type Server struct {
	store               store.GraphStore
	authService         *auth.AuthService
	jwtConfig           *auth.JWTConfig
	contradictionEngine *engine.ContradictionEngine
	lineageEngine       *engine.LineageEngine
	topologyGenerator   *topology.Generator
	topologyExecutor    *topology.Executor
	sseBroker           *telemetry.SSEBroker
	rateLimiter         *RateLimiter
	sessions            *SessionStore
	controlPool         *pgxpool.Pool
}

func NewServer(
	s store.GraphStore,
	authSvc *auth.AuthService,
	jwtCfg *auth.JWTConfig,
	contraEngine *engine.ContradictionEngine,
	linEngine *engine.LineageEngine,
	topoGen *topology.Generator,
	topoExec *topology.Executor,
) *Server {
	rps := 100.0
	burst := 200
	if val := os.Getenv("GARUDA_RATE_LIMIT_RPS"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil && parsed > 0 {
			rps = parsed
		}
	}
	if val := os.Getenv("GARUDA_RATE_LIMIT_BURST"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			burst = parsed
		}
	}

	return &Server{
		store:               s,
		authService:         authSvc,
		jwtConfig:           jwtCfg,
		contradictionEngine: contraEngine,
		lineageEngine:       linEngine,
		topologyGenerator:   topoGen,
		topologyExecutor:    topoExec,
		sseBroker:           telemetry.NewSSEBroker(),
		rateLimiter:         NewRateLimiter(rps, burst),
		sessions:            NewSessionStore(SessionTTL),
	}
}

func (s *Server) SetControlPool(pool *pgxpool.Pool) {
	s.controlPool = pool
}

func (s *Server) RateLimiter() *RateLimiter {
	return s.rateLimiter
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	TraceID string `json:"trace_id,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode json response", "error", err)
	}
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, status, ErrorResponse{
		Error: message,
		Code:  status,
	})
}

type CreateEntityRequest struct {
	Type       string         `json:"type"`
	Name       string         `json:"name"`
	Properties map[string]any `json:"properties"`
}

type CreateClaimRequest struct {
	SubjectID  uuid.UUID      `json:"subject_id"`
	Predicate  string         `json:"predicate"`
	ObjectID   uuid.UUID      `json:"object_id"`
	Confidence float64        `json:"confidence"`
	Source     string         `json:"source"`
	Metadata   map[string]any `json:"metadata"`
}

type ResolveContradictionRequest struct {
	Strategy string `json:"strategy"`
}

type GeneratePlansRequest struct {
	Goal        string         `json:"goal"`
	Constraints map[string]any `json:"constraints"`
}

// SetupRouter builds the HTTP middleware and route stack.
//
// The optional overrides callback runs after RegisterRoutes and before
// the middleware chain is wrapped around the router. Because
// gorilla/mux evaluates routes in registration order, an override
// must be registered on the existing route via mux's route
// replacement — or, simpler, registered BEFORE RegisterRoutes so it
// wins first-match. SetupRouter calls overrides before RegisterRoutes
// for that reason: any path registered in overrides shadows the
// default registered in RegisterRoutes.
//
// Pass nil when no overrides are needed (production API server, most
// tests). cmd/garuda/dev_cmd.go passes a callback that replaces
// /graph and /api/v1/graph with the inline D3 visualizer.
func SetupRouter(server *Server, rateLimiter *RateLimiter, mcpHandler http.Handler, overrides func(*mux.Router)) http.Handler {
	r := mux.NewRouter()

	if overrides != nil {
		overrides(r)
	}

	server.RegisterRoutes(r, mcpHandler)

	var handler http.Handler = r
	if rateLimiter != nil {
		handler = RateLimitMiddleware(rateLimiter)(handler)
	}
	handler = RecoveryMiddleware(handler)
	handler = RequestIDMiddleware(handler)

	return handler
}

func (s *Server) RegisterRoutes(r *mux.Router, mcpHandler http.Handler) {
	// Liveness and readiness probes (supports /health, /healthz, /readyz, and /api/v1/system/health)
	r.HandleFunc("/health", s.handleHealth).Methods(http.MethodGet)
	r.HandleFunc("/healthz", s.handleHealth).Methods(http.MethodGet)
	r.HandleFunc("/readyz", s.handleReady).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/system/health", s.handleReady).Methods(http.MethodGet)

	// Browser login & signup routes (public)
	r.HandleFunc("/login", s.HandleLoginPage).Methods(http.MethodGet)
	r.HandleFunc("/login", s.HandleLoginSubmit).Methods(http.MethodPost)
	r.HandleFunc("/signup", s.HandleSignupSubmit).Methods(http.MethodPost)
	r.HandleFunc("/logout", s.HandleLogout).Methods(http.MethodPost)

	// Token-based Auth API (public)
	r.HandleFunc("/api/v1/auth/login", s.handleLogin).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/auth/signup", s.HandleSignupSubmit).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/auth/refresh", s.handleRefresh).Methods(http.MethodPost)

	// Browser UI routes — gated by session cookie, redirect to /login
	r.HandleFunc("/", s.RequireSession(s.HandleAdminDashboard)).Methods(http.MethodGet)
	r.HandleFunc("/dashboard", s.RequireSession(s.HandleDashboardPage)).Methods(http.MethodGet)
	r.HandleFunc("/admin", s.RequireSession(s.RequireOwner(s.HandleAdminDashboard))).Methods(http.MethodGet)
	r.HandleFunc("/admin/users", s.RequireSession(s.RequireOwner(s.HandleCreateUser))).Methods(http.MethodPost)
	r.HandleFunc("/admin/users/role", s.RequireSession(s.RequireOwner(s.HandleUpdateUserRole))).Methods(http.MethodPost)
	r.HandleFunc("/admin/users/delete", s.RequireSession(s.RequireOwner(s.HandleDeleteUser))).Methods(http.MethodPost)
	r.HandleFunc("/graph", s.RequireSession(s.HandleGraph)).Methods(http.MethodGet)

	// Control Plane — operator console across all tenants on this DB.
	// Gated by GARUDA_CONTROL_TOKEN, NOT by tenant session cookies.
	// Returns 404 when the token or pool is unset.
	r.HandleFunc("/control", s.RequireControlToken(s.HandleControlOverview)).Methods(http.MethodGet)
	r.HandleFunc("/control/tenants/{id}", s.RequireControlToken(s.HandleControlTenantDetail)).Methods(http.MethodGet)
	r.HandleFunc("/control/workspaces/{id}", s.RequireControlToken(s.HandleControlWorkspaceDetail)).Methods(http.MethodGet)
	r.HandleFunc("/control/system", s.RequireControlToken(s.HandleControlSystem)).Methods(http.MethodGet)
	r.HandleFunc("/control/audit", s.RequireControlToken(s.HandleControlAudit)).Methods(http.MethodGet)
	r.HandleFunc("/control/Analytics", s.RequireControlToken(s.HandleControlAnalytics)).Methods(http.MethodGet)
	r.HandleFunc("/control/analytics", s.RequireControlToken(s.HandleControlAnalytics)).Methods(http.MethodGet)
	r.HandleFunc("/control/logs", s.RequireControlToken(s.HandleControlLogs)).Methods(http.MethodGet)
	r.HandleFunc("/control/Billing", s.RequireControlToken(s.HandleControlBilling)).Methods(http.MethodGet)
	r.HandleFunc("/control/billing", s.RequireControlToken(s.HandleControlBilling)).Methods(http.MethodGet)
	r.HandleFunc("/control/Support", s.RequireControlToken(s.HandleControlSupport)).Methods(http.MethodGet)
	r.HandleFunc("/control/support", s.RequireControlToken(s.HandleControlSupport)).Methods(http.MethodGet)

	// Protected API v1 Subrouter (cryptographically verifies Ed25519 JWT or valid session cookie)
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(s.AuthMiddleware)

	// Dashboard & Graph API
	api.HandleFunc("/dashboard/stats", s.HandleDashboardStats).Methods(http.MethodGet)
	api.HandleFunc("/dashboard/search", s.HandleDashboardSearch).Methods(http.MethodGet)
	api.HandleFunc("/graph", s.HandleGraph).Methods(http.MethodGet)
	api.HandleFunc("/events", s.HandleSSE).Methods(http.MethodGet)

	// Runtime & Telemetry Ingestion (wired to full entity correlator + runtime_edges + claim_verifications)
	api.HandleFunc("/telemetry/spans", s.HandleIngestRuntimeSpans).Methods(http.MethodPost)
	api.HandleFunc("/runtime/spans", s.HandleIngestRuntimeSpans).Methods(http.MethodPost)
	api.HandleFunc("/runtime/coverage", s.HandleRuntimeCoverage).Methods(http.MethodGet)
	api.HandleFunc("/merkle/state", s.HandleMerkleState).Methods(http.MethodGet)

	// Architectural Decision Records (MADR-compatible)
	api.HandleFunc("/decisions", s.HandleCreateDecision).Methods(http.MethodPost)
	api.HandleFunc("/decisions", s.HandleListDecisions).Methods(http.MethodGet)
	api.HandleFunc("/decisions/{id}", s.HandleGetDecision).Methods(http.MethodGet)
	api.HandleFunc("/decisions/{id}/supersede", s.HandleSupersedeDecision).Methods(http.MethodPost)

	// Core Entity & Claim Graph Operations
	api.HandleFunc("/entities", s.handleCreateEntity).Methods(http.MethodPost)
	api.HandleFunc("/entities/{id}", s.handleGetEntity).Methods(http.MethodGet)
	api.HandleFunc("/claims", s.handleCreateClaim).Methods(http.MethodPost)
	api.HandleFunc("/claims/{id}", s.handleGetClaim).Methods(http.MethodGet)

	// Contradictions
	api.HandleFunc("/contradictions", s.handleListContradictions).Methods(http.MethodGet)
	api.HandleFunc("/contradictions/{id}/resolve", s.handleResolveContradiction).Methods(http.MethodPost)

	// Agent Orchestration & Handoff
	api.HandleFunc("/agent/checkpoint", s.HandleAgentCheckpoint).Methods(http.MethodPost)
	api.HandleFunc("/agent/handoff", s.HandleCreateHandoff).Methods(http.MethodPost)
	api.HandleFunc("/agent/handoff/{session_id}", s.HandleGetHandoff).Methods(http.MethodGet)
	api.HandleFunc("/agent/resume", s.HandleResumeAgent).Methods(http.MethodPost)

	// Lineage & Topology
	api.HandleFunc("/lineage/status", s.HandleLineageStatus).Methods(http.MethodGet)
	api.HandleFunc("/lineage/{id}", s.handleGetLineage).Methods(http.MethodGet)
	api.HandleFunc("/topology/plans", s.handleGeneratePlans).Methods(http.MethodPost)
	api.HandleFunc("/topology/plans/{id}/execute", s.handleExecutePlan).Methods(http.MethodPost)

	if mcpHandler != nil {
		r.PathPrefix("/mcp").Handler(mcpHandler)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if pgStore, ok := s.store.(*store.PostgresStore); ok && pgStore != nil && pgStore.Pool() != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pgStore.Pool().Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status":    "unready",
				"reason":    "database ping failed",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":    "healthy",
		"readiness": "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if s.authService == nil {
		writeJSONError(w, "auth service not configured", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	identifier := req.Email
	if identifier == "" {
		identifier = req.Username
	}

	accessToken, refreshToken, err := s.authService.Authenticate(r.Context(), identifier, req.Password)
	if err != nil {
		writeJSONError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if s.authService == nil {
		writeJSONError(w, "auth service not configured", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	accessToken, err := s.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeJSONError(w, "invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"access_token": accessToken,
	})
}

func (s *Server) handleCreateEntity(w http.ResponseWriter, r *http.Request) {
	var req CreateEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok {
		tenantID = tenant.CanonicalID
	}

	entity := &types.Entity{
		ID:         uuid.New(),
		TenantID:   tenantID,
		Type:       req.Type,
		Name:       req.Name,
		Properties: req.Properties,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	if err := s.store.CreateEntity(r.Context(), entity); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, entity)
}

func (s *Server) handleGetEntity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeJSONError(w, "invalid entity id", http.StatusBadRequest)
		return
	}

	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok {
		tenantID = tenant.CanonicalID
	}

	entity, err := s.store.GetEntity(r.Context(), tenantID, id)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, entity)
}

func (s *Server) handleCreateClaim(w http.ResponseWriter, r *http.Request) {
	var req CreateClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok {
		tenantID = tenant.CanonicalID
	}

	claim := &types.Claim{
		ID:         uuid.New(),
		TenantID:   tenantID,
		SubjectID:  req.SubjectID,
		Predicate:  req.Predicate,
		ObjectID:   req.ObjectID,
		Confidence: req.Confidence,
		Status:     types.ClaimStatusActive,
		Source:     req.Source,
		Metadata:   req.Metadata,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	if err := s.store.CreateClaim(r.Context(), claim); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.contradictionEngine != nil {
		_, _ = s.contradictionEngine.DetectContradictions(r.Context(), tenantID, claim)
	}

	writeJSON(w, http.StatusCreated, claim)
}

func (s *Server) handleGetClaim(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeJSONError(w, "invalid claim id", http.StatusBadRequest)
		return
	}

	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok {
		tenantID = tenant.CanonicalID
	}

	claim, err := s.store.GetClaim(r.Context(), tenantID, id)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, claim)
}

func (s *Server) handleListContradictions(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok {
		tenantID = tenant.CanonicalID
	}

	contradictions, err := s.store.ListActiveContradictions(r.Context(), tenantID)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, contradictions)
}

func (s *Server) handleResolveContradiction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeJSONError(w, "invalid contradiction id", http.StatusBadRequest)
		return
	}

	var req ResolveContradictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok {
		tenantID = tenant.CanonicalID
	}

	if err := s.contradictionEngine.ResolveContradiction(r.Context(), tenantID, id, req.Strategy); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

func (s *Server) handleGetLineage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeJSONError(w, "invalid entity id", http.StatusBadRequest)
		return
	}

	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok {
		tenantID = tenant.CanonicalID
	}

	lineage, err := s.lineageEngine.GetEntityLineage(r.Context(), tenantID, id, 5)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, lineage)
}

func (s *Server) handleGeneratePlans(w http.ResponseWriter, r *http.Request) {
	var req GeneratePlansRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok {
		tenantID = tenant.CanonicalID
	}

	plans, err := s.topologyGenerator.GeneratePlans(r.Context(), tenantID, req.Goal, req.Constraints)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, plans)
}

func (s *Server) handleExecutePlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		writeJSONError(w, "invalid plan id", http.StatusBadRequest)
		return
	}

	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok {
		tenantID = tenant.CanonicalID
	}

	result, err := s.topologyExecutor.ExecutePlan(r.Context(), tenantID, id)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, result)
}
