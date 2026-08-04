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
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/telemetry"
)

func main() {
	// 1. Initialize Telemetry (Garuda v4 Roadmap Specification)
	telConfig := telemetry.LoadConfigFromEnv()
	telemetry.InitTelemetry(telConfig)
	slog.Info("Telemetry engine initialized", "enabled", telConfig.Enabled, "tenant", os.Getenv("GARUDA_TENANT_ID"))

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
	// A. UNPROTECTED / PUBLIC ROUTES (Accessible via Browser & Web UI)
	// ------------------------------------------------------------------
	mainMux.HandleFunc("GET /health", server.HandleHealth)
	mainMux.HandleFunc("GET /dashboard", server.HandleDashboard)
	mainMux.HandleFunc("GET /api/v1/events", server.HandleLiveEvents) // The Web Dashboard's SSE live-update script hit /api/v1/events, which requires JWT auth.
	mainMux.HandleFunc("GET /debug/token", server.HandleDebugToken)
	mainMux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// ------------------------------------------------------------------
	// B. PROTECTED API ROUTES (Require JWT Authorization)
	// ------------------------------------------------------------------
	protectedMux.HandleFunc("POST /api/v1/decisions/submit", server.HandleProposeDecision)
	protectedMux.HandleFunc("GET /api/v1/decisions/{id}/lineage", server.HandleDecisionLineage)
	protectedMux.HandleFunc("POST /api/v1/agents/warmup", server.HandleAgentWarmup)

	// Agent checkpoint endpoints
	protectedMux.HandleFunc("POST /api/v1/agents/checkpoint", server.HandleAgentCheckpoint)
	protectedMux.HandleFunc("GET /api/v1/agents/checkpoint/{id}", server.HandleGetAgentCheckpoint)
	protectedMux.HandleFunc("POST /api/v1/agents/resume", server.HandleAgentResume)
	protectedMux.HandleFunc("POST /api/v1/agents/handoff", server.HandleAgentHandoff)

	// Budget & Metering endpoints
	protectedMux.HandleFunc("GET /api/v1/budget", server.HandleGetBudget)
	protectedMux.HandleFunc("POST /api/v1/budget/consume", server.HandleConsumeBudget)

	// Evidence & Verification endpoints
	protectedMux.HandleFunc("GET /api/v1/evidence/verify/{id}", server.HandleVerifyDecision)
	protectedMux.HandleFunc("GET /api/v1/evidence/snapshots", server.HandleListMerkleSnapshots)

	// Temporal / Bitemporal routes
	protectedMux.HandleFunc("GET /api/v1/decisions/active", server.HandleDecisionsActiveAt)
	protectedMux.HandleFunc("GET /api/v1/decisions/{id}/history", server.HandleDecisionHistory)

	// ------------------------------------------------------------------
	// C. ROUTE INTEGRATION & MIDDLEWARE PIPELINE
	// ------------------------------------------------------------------
	// Wrap protected API endpoints with the JWT Auth Middleware
	protectedHandler := api.WithAuth(jwtConfig)(protectedMux)

	// Mount protected API tree onto main multiplexer under /api/
	mainMux.Handle("/api/", protectedHandler)

	// Construct global middleware chain: Recovery -> Logging -> RequestID -> RateLimit -> CORS -> Mux
	handler := api.WithRecovery(
		api.WithLogging(
			api.WithRequestID(
				api.WithRateLimit(rateLimiter)(
					api.WithCORS([]string{"*"})(mainMux),
				),
			),
		),
	)

	// 8. HTTP Server Configuration
	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 9. Graceful Shutdown & Intercept Mechanisms
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		slog.Info("Garuda API Secure Gateway online", "addr", ":8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Gateway server crashed unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for termination signal
	<-ctx.Done()
	slog.Info("Shutting down gateway gracefully...")

	// Allocate a 30-second window to flush processes and metrics safely
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Flush remaining telemetry batches before stopping network listeners
	if err := telemetry.ShutdownTelemetry(shutdownCtx); err != nil {
		slog.Error("Telemetry metric flush on shutdown failed", "error", err)
	} else {
		slog.Info("Telemetry pipeline drained cleanly")
	}

	// Terminate active network listeners
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful HTTP stack shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Gateway shutdown complete")
}
