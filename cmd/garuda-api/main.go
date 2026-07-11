package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/techtaytor/garuda/internal/api"
	"github.com/techtaytor/garuda/internal/auth"
	"github.com/techtaytor/garuda/internal/engine"
	"github.com/techtaytor/garuda/internal/lineage"
	"github.com/techtaytor/garuda/internal/store"
	"github.com/techtaytor/garuda/internal/telemetry"
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
	_ = engine.NewLineageEngine(dbStore)
	_ = engine.NewContradictionEngine(dbStore)

	// App Services
	authService := auth.NewAuthService(dbStore, jwtConfig)
	server := api.NewServer(dbStore, authService, jwtConfig)

	// 6. Security & Traffic Controls
	rateLimiter := api.NewRateLimiter(100, time.Minute, 1000)

	// 7. Routing & Multiplexing
	mux := http.NewServeMux()

	// Unprotected System Endpoints
	mux.HandleFunc("GET /health", server.HandleHealth)
	mux.HandleFunc("GET /debug/token", server.HandleDebugToken)

	// Protected V4 Telemetry-Instrumented Core Endpoints
	mux.HandleFunc("POST /api/v1/decisions/submit", server.HandleSubmitDecision)
	mux.HandleFunc("POST /api/v1/agents/warmup", server.HandleWarmup)

	// 8. Middleware Pipeline Construction
	// Order of execution: Recovery -> Logging -> RequestID -> Auth Authentication -> Rate Limiting -> CORS -> Handler
	handler := api.WithRecovery(
		api.WithLogging(
			api.WithRequestID(
				api.WithAuth(jwtConfig)(
					api.WithRateLimit(rateLimiter)(
						api.WithCORS([]string{"*"})(mux),
					),
				),
			),
		),
	)

	// 9. HTTP Server Configuration
	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 10. Graceful Shutdown & Intercept Mechanisms
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

	// CRITICAL: Flush remaining telemetry batches before stopping the HTTP network stack
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
