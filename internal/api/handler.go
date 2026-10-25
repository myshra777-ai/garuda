// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"sync"
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
	ID        string                 `json:"id"`
	Workspace string                 `json:"workspace,omitempty"`
	AgentID   string                 `json:"agent_id,omitempty"`
	Status    string                 `json:"status,omitempty"`
	State     map[string]interface{} `json:"state,omitempty"`
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
	sseCount            chan struct{}
	checkpointMu        sync.RWMutex
	checkpointStore     map[string]CheckpointRecord
	sessions            *SessionStore
	rateLimiter         *IPRateLimiter

	// controlPool is a separate database connection scoped to the
	// garuda_control_ro role. Read-only. Used by Control Plane
	// handlers only. Nil when CONTROL_DATABASE_URL is unset, in
	// which case the Control Plane returns 404 for every request.
	controlPool *pgxpool.Pool

	// processStart is the wall-clock time the API server began
	// accepting requests. Used by the Control Plane Operations tab
	// to render uptime. Set once in NewServer; never updated.
	processStart time.Time

	// errorRate tracks 5xx-over-total responses in a rolling 60-minute
	// window. Consumed by the Control Plane Operations tab.
	errorRate *rateCounter
}

// SetControlPool attaches the read-only Control Plane pool. Called by
// the two entry points (cmd/garuda-api, cmd/garuda/dev) after
// NewServer, when CONTROL_DATABASE_URL is set.
func (s *Server) SetControlPool(pool *pgxpool.Pool) {
	s.controlPool = pool
}

// NewServer creates a new API server instance with initialized state and SSE broker.
func NewServer(
	store types.DecisionStore,
	authService *auth.AuthService,
	jwtConfig *auth.JWTConfig,
	contradictionEngine *engine.ContradictionEngine,
	lineageEngine *engine.LineageEngine,
	topologyGenerator *topology.Generator,
	topologyExecutor *topology.Executor,
	rateLimiters ...*IPRateLimiter,
) *Server {
	sseBroker := telemetry.NewSSEBroker()
	rateLimiter := (*IPRateLimiter)(nil)
	if len(rateLimiters) > 0 {
		rateLimiter = rateLimiters[0]
	}
	if rateLimiter == nil {
		rateLimiter = NewIPRateLimiter(readRateLimitPerMinute(), readRateLimitBurst())
	}

	return &Server{
		store:               store,
		authService:         authService,
		jwtConfig:           jwtConfig,
		contradictionEngine: contradictionEngine,
		lineageEngine:       lineageEngine,
		topologyGenerator:   topologyGenerator,
		topologyExecutor:    topologyExecutor,
		sseBroker:           sseBroker,
		sseCount:            make(chan struct{}, maxSSEConnections),
		checkpointStore:     make(map[string]CheckpointRecord),
		sessions:            NewSessionStore(SessionTTL),
		rateLimiter:         rateLimiter,
		processStart:        time.Now().UTC(),
		errorRate:           &rateCounter{},
	}
}

// RegisterRoutes sets up HTTP routing and subrouter middleware hierarchy.
func (s *Server) RegisterRoutes(r *mux.Router) {
	// =========================================================================
	// PUBLIC ROUTES (no authentication)
	// =========================================================================
	r.HandleFunc("/health", s.HandleHealth).Methods(http.MethodGet)
	r.HandleFunc("/healthz", s.HandleHealth).Methods(http.MethodGet)
	r.HandleFunc("/readyz", s.HandleReady).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/system/health", s.HandleReady).Methods(http.MethodGet)
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}).Methods(http.MethodGet)
	r.HandleFunc("/icon.svg", s.HandlePWAIcon).Methods(http.MethodGet)
	r.HandleFunc("/manifest.json", s.HandlePWAManifest).Methods(http.MethodGet)
	r.HandleFunc("/sw.js", s.HandlePWAServiceWorker).Methods(http.MethodGet)
	r.HandleFunc("/login", s.HandleLoginGET).Methods(http.MethodGet)
	r.HandleFunc("/login", s.HandleLoginPOST).Methods(http.MethodPost)
	r.HandleFunc("/signup", s.HandleSignupGET).Methods(http.MethodGet)
	r.HandleFunc("/signup", s.HandleSignupPOST).Methods(http.MethodPost)
	r.HandleFunc("/logout", s.HandleLogout).Methods(http.MethodPost)

	// =========================================================================
	// SESSION-PROTECTED ROUTES (browser)
	// =========================================================================
	sess := r.NewRoute().Subrouter()
	sess.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.RequireSession(next.ServeHTTP)(w, r)
		})
	})

	sess.HandleFunc("/dashboard", s.HandleDashboard).Methods(http.MethodGet)
	sess.HandleFunc("/api/v1/dashboard/stats", s.HandleDashboardStats).Methods(http.MethodGet)
	sess.HandleFunc("/api/v1/dashboard/search", s.HandleDashboardSearch).Methods(http.MethodGet)
	sess.HandleFunc("/api/v1/dashboard/sessions", s.HandleDashboardSessions).Methods(http.MethodGet)
	sess.HandleFunc("/api/v1/dashboard/decisions", s.HandleDashboardDecisions).Methods(http.MethodGet)
	sess.HandleFunc("/api/v1/dashboard/decisions/{id}", s.HandleDashboardDecisionDetail).Methods(http.MethodGet)
	sess.HandleFunc("/api/v1/dashboard/policies", s.HandleDashboardPolicies).Methods(http.MethodGet)
	sess.HandleFunc("/api/v1/dashboard/policies/verify", s.HandleDashboardPolicyVerify).Methods(http.MethodGet)
	sess.HandleFunc("/api/v1/graph", s.HandleGraph).Methods(http.MethodGet)
	sess.HandleFunc("/api/v1/events", s.HandleLiveEvents).Methods(http.MethodGet)
	sess.HandleFunc("/api/v1/workspaces", s.HandleListWorkspaces).Methods(http.MethodGet)
	sess.HandleFunc("/system/discover", s.HandleSystemDiscover).Methods(http.MethodGet)
	sess.HandleFunc("/admin", s.HandleAdminDashboard).Methods(http.MethodGet)
	// =========================================================================
	// JWT-PROTECTED API SUBROUTER (CLI / MCP tokens)
	// =========================================================================
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(s.AuthMiddleware)

	api.HandleFunc("/decisions/propose", s.HandleProposeDecision).Methods(http.MethodPost)
	api.HandleFunc("/evidence/verify/{id}", s.HandleVerifyDecision).Methods(http.MethodGet)
	api.HandleFunc("/agent/warmup", s.HandleAgentWarmup).Methods(http.MethodGet)
	api.HandleFunc("/agents/checkpoint", s.HandleAgentCheckpoint).Methods(http.MethodPost)
	api.HandleFunc("/agents/handoff", s.HandleHandoff).Methods(http.MethodPost)
	api.HandleFunc("/agents/resume", s.HandleResume).Methods(http.MethodPost)
	api.HandleFunc("/tasks/{task_id}/lineage", s.HandleGetLineage).Methods(http.MethodGet)
	api.HandleFunc("/router/evaluate", s.HandleEvaluateRoute).Methods(http.MethodPost, http.MethodOptions)

	api.Handle("/telemetry/stream", s.sseBroker).Methods(http.MethodGet)
	api.HandleFunc("/telemetry/spans", s.HandleIngestRuntimeSpans).Methods(http.MethodPost)
	api.HandleFunc("/runtime/spans", s.HandleIngestRuntimeSpans).Methods(http.MethodPost)
	api.HandleFunc("/runtime/coverage", s.HandleGetRuntimeCoverage).Methods(http.MethodGet)
	api.HandleFunc("/merkle/state", s.HandleGetMerkleState).Methods(http.MethodGet)

	// Debug endpoint (only available outside production)
	r.HandleFunc("/debug/token", s.HandleDebugToken).Methods(http.MethodGet)

	// ── Control Plane (/ _ / control) ──
	r.HandleFunc("/_/control/login", s.HandleControlPlaneLoginGET).Methods(http.MethodGet)
	r.HandleFunc("/_/control/login", s.HandleControlPlaneLoginPOST).Methods(http.MethodPost)
	r.HandleFunc("/_/control/logout", s.HandleControlPlaneLogout).Methods(http.MethodPost)

	cp := r.PathPrefix("/_/control").Subrouter()
	cp.Use(s.controlPlaneAuth)
	cp.HandleFunc("", s.HandleControlPlane).Methods(http.MethodGet)
}

// HandleHealth returns gateway health status.
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// HandleReady verifies database connectivity when backed by PostgresStore.
func (s *Server) HandleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}
	if pgStore, ok := s.store.(*store.PostgresStore); ok && pgStore != nil && pgStore.Pool() != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pgStore.Pool().Ping(ctx); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status": "unready",
				"reason": "database ping failed",
			})
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "readiness": "ready"})
}

// HandleDebugToken generates a tenant-scoped JWT for testing.
func (s *Server) HandleDebugToken(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("GARUDA_ENV") == "production" {
		s.RespondWithError(w, http.StatusNotFound, "debug endpoint disabled in production", "")
		return
	}

	if r.Method != http.MethodGet {
		s.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	if s.jwtConfig == nil {
		s.RespondWithError(w, http.StatusInternalServerError, "jwt config unavailable", "")
		return
	}

	actor := r.URL.Query().Get("actor")
	if actor == "" {
		actor = "debug-user"
	}

	// Align default tenant with the system dashboard tenant ID
	tenantID := tenant.CanonicalID
	if tenantIDStr := r.URL.Query().Get("tenant_id"); tenantIDStr != "" {
		if parsed, err := uuid.Parse(tenantIDStr); err == nil {
			tenantID = parsed
		}
	}

	token, err := s.jwtConfig.GenerateToken(actor, tenantID)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, "failed to generate token", "")
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

func readRateLimitPerMinute() int {
	if v := os.Getenv("GARUDA_RATE_LIMIT_PER_MINUTE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultRateLimitPerMinute
}

func readRateLimitBurst() int {
	if v := os.Getenv("GARUDA_RATE_LIMIT_BURST"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultRateLimitBurst
}

func (s *Server) RateLimiter() *IPRateLimiter {
	return s.rateLimiter
}

// HandlePWAIcon serves a crisp SVG vector icon for Progressive Web App installation.
func (s *Server) HandlePWAIcon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128"><rect width="128" height="128" rx="24" fill="#090d16"/><path d="M64 20L104 42V86L64 108L24 86V42L64 20Z" fill="none" stroke="#38bdf8" stroke-width="6"/><circle cx="64" cy="64" r="16" fill="#10b981"/><path d="M64 20V48M104 42L78 56M104 86L78 72M64 108V80M24 86L50 72M24 42L50 56" stroke="#38bdf8" stroke-width="4"/></svg>`))
}

// HandlePWAManifest serves the Web App Manifest so Chrome/Edge/Chromebook/Android/macOS/Windows can install Garuda as a standalone app.
func (s *Server) HandlePWAManifest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/manifest+json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{
  "name": "Garuda v1.6 — Software Intelligence & Change Control",
  "short_name": "Garuda",
  "description": "Deterministic Architecture Change-Control Gate, Active AI Agent Harness (Talon), and 6-Surface Software Intelligence Instrument",
  "start_url": "/dashboard",
  "display": "standalone",
  "background_color": "#090d16",
  "theme_color": "#090d16",
  "icons": [
    {
      "src": "/icon.svg",
      "sizes": "any",
      "type": "image/svg+xml",
      "purpose": "any maskable"
    }
  ]
}`))
}

// HandlePWAServiceWorker serves a lightweight service worker for Chromium/Android PWA compliance.
func (s *Server) HandlePWAServiceWorker(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate', (e) => e.waitUntil(self.clients.claim()));
self.addEventListener('fetch', () => {});
`))
}
