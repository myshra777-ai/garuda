package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/myshra777-ai/garuda/internal/api"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/engine"
	"github.com/myshra777-ai/garuda/internal/lineage"
	"github.com/myshra777-ai/garuda/internal/mcp"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/telemetry"
)

func main() {
	// 1. Initialize Telemetry with config and consent checks.
	telConfig := telemetry.LoadConfigFromEnv()

	if os.Getenv("GARUDA_MODE") == "passive" {
		telConfig.Mode = "passive"
	}

	if err := telemetry.InitTelemetry(telConfig); err != nil {
		slog.Warn("Telemetry init failed", "error", err)
	}
	slog.Info("Telemetry engine initialized", "enabled", telConfig.Enabled, "mode", telConfig.Mode, "tenant", os.Getenv("GARUDA_TENANT_ID"))

	// 2. Load Configuration from Environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
	}

	// 3. Database Migrations
	slog.Info("Running migrations...")
	if err := store.Migrate(dbURL, "migrations"); err != nil {
		slog.Error("Migration failed", "error", err)
		os.Exit(1)
	}
	slog.Info("Migrations completed successfully")

	// 4. Identity & Access Management (JWT)
	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		jwtIssuer = "garuda"
	}
	jwtAudience := os.Getenv("JWT_AUDIENCE")
	if jwtAudience == "" {
		jwtAudience = "garuda-api"
	}
	jwtExpiry := 15 * time.Minute

	jwtConfig, err := auth.NewJWTConfig(jwtIssuer, jwtAudience, jwtExpiry)
	if err != nil {
		slog.Error("Failed to initialize JWT config", "error", err)
		os.Exit(1)
	}
	slog.Info("JWT authentication initialized", "public_key", jwtConfig.GetPublicKeyHex())

	// 5. Core Storage & Engine Initializations
	dbStore, err := store.NewPostgresStore(dbURL)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer dbStore.Close()

	// Lineage Core Graph & Engine Instances
	_ = lineage.NewGraph(1000)
	lineageEngine := engine.NewLineageEngine(dbStore)
	contradictionEngine := engine.NewContradictionEngine(dbStore)

	// App Services
	authService := auth.NewAuthService(dbStore, jwtConfig)
	server := api.NewServer(dbStore, authService, jwtConfig, contradictionEngine, lineageEngine)

	// 6. Security & Traffic Controls
	rateLimiter := api.NewRateLimiter(100, time.Minute, 1000)

	// 7. Separate Muxing for Public vs. JWT-Protected Endpoints
	mainMux := http.NewServeMux()
	protectedMux := http.NewServeMux()

	// ------------------------------------------------------------------
	// A. PUBLIC / SYSTEM ENDPOINTS (Accessible without JWT)
	// ------------------------------------------------------------------
	mainMux.HandleFunc("GET /health", server.HandleHealth)
	mainMux.HandleFunc("GET /system/health", server.HandleHealth)
	mainMux.HandleFunc("GET /system/discover", server.HandleSystemDiscover)
	mainMux.HandleFunc("GET /system/bootstrap", server.HandleSystemBootstrap)
	mainMux.HandleFunc("POST /sandbox", server.HandleSandbox)

	// Documentation & Visual Dashboards
	mainMux.HandleFunc("GET /dashboard", server.HandleDashboard)
	mainMux.HandleFunc("GET /docs", server.HandleSwaggerUI)
	mainMux.HandleFunc("GET /openapi.yaml", server.HandleOpenAPISpec)
	mainMux.HandleFunc("GET /openapi.json", server.HandleOpenAPISpec)
	mainMux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// MCP Bridge & Debug Token
	mainMux.HandleFunc("POST /mcp/bridge", mcp.BridgeHandler("http://localhost:8080"))
	mainMux.HandleFunc("GET /debug/token", server.HandleDebugToken)

	// ------------------------------------------------------------------
	// B. PROTECTED API ROUTES (Require JWT Authorization)
	// ------------------------------------------------------------------
	// Decisions
	protectedMux.HandleFunc("POST /api/v1/decisions/submit", server.HandleProposeDecision)
	protectedMux.HandleFunc("POST /api/v1/decisions", server.HandleProposeDecision)
	protectedMux.HandleFunc("GET /api/v1/decisions/active", server.HandleDecisionsActiveAt)
	protectedMux.HandleFunc("GET /api/v1/decisions/{id}/history", server.HandleDecisionHistory)
	protectedMux.HandleFunc("GET /api/v1/decisions/{id}/lineage", server.HandleDecisionLineage)

	// Multi-Agent Execution & Checkpoints
	protectedMux.HandleFunc("POST /api/v1/agents/warmup", server.HandleAgentWarmup)
	protectedMux.HandleFunc("POST /api/v1/agents/checkpoint", server.HandleAgentCheckpoint)
	protectedMux.HandleFunc("GET /api/v1/agents/checkpoint/{id}", server.HandleGetAgentCheckpoint)
	protectedMux.HandleFunc("POST /api/v1/agents/resume", server.HandleAgentResume)
	protectedMux.HandleFunc("POST /api/v1/agents/handoff", server.HandleAgentHandoff)

	// Audit & Compliance
	protectedMux.HandleFunc("GET /api/v1/audit/export", server.HandleExportAuditLogs)
	protectedMux.HandleFunc("GET /api/v1/audit/verify/{id}", server.HandleVerifyAuditLog)
	protectedMux.HandleFunc("GET /api/v1/evidence/verify/{id}", server.HandleVerifyDecision)
	protectedMux.HandleFunc("GET /api/v1/evidence/snapshots", server.HandleListMerkleSnapshots)

	// Budget & Metering
	protectedMux.HandleFunc("GET /api/v1/budget", server.HandleGetBudget)
	protectedMux.HandleFunc("POST /api/v1/budget/consume", server.HandleConsumeBudget)

	// Router Pre-Flight Evaluation
	protectedMux.HandleFunc("POST /api/v1/router/evaluate", server.HandleEvaluateRoute)

	// Dashboard Stats Stream
	protectedMux.HandleFunc("GET /api/v1/dashboard/stats", server.HandleDashboardStats)
	protectedMux.HandleFunc("GET /api/v1/events", server.HandleLiveEvents)

	// ------------------------------------------------------------------
	// C. ROUTE INTEGRATION & MIDDLEWARE PIPELINE
	// ------------------------------------------------------------------
	// Wrap protected API tree with JWT Authentication
	protectedHandler := api.WithAuth(jwtConfig)(protectedMux)
	mainMux.Handle("/api/", protectedHandler)

	// Global Middleware Chain: Recovery -> Logging -> RequestID -> MerkleHeader -> RateLimit -> CORS -> Mux
	handler := api.WithRecovery(
		api.WithLogging(
			api.WithRequestID(
				server.WithMerkleHeader(
					api.WithRateLimit(rateLimiter)(
						api.WithCORS([]string{"*"})(mainMux),
					),
				),
			),
		),
	)

	// 8. HTTP Server Configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 9. Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		slog.Info("Garuda API Secure Gateway online", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Gateway server crashed unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for termination signal
	<-ctx.Done()
	slog.Info("Shutting down gateway gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := telemetry.ShutdownTelemetry(shutdownCtx); err != nil {
		slog.Error("Telemetry metric flush on shutdown failed", "error", err)
	} else {
		slog.Info("Telemetry pipeline drained cleanly")
	}

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful HTTP stack shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Gateway shutdown complete")
}
