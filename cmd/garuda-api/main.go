// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/api"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/engine"
	"github.com/myshra777-ai/garuda/internal/lineage"
	"github.com/myshra777-ai/garuda/internal/mcp"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/telemetry"
	"github.com/myshra777-ai/garuda/internal/topology"
)

func main() {
	// 1. Initialize Telemetry
	telConfig := telemetry.LoadConfigFromEnv()
	if os.Getenv("GARUDA_MODE") == "passive" {
		telConfig.Mode = "passive"
	}
	if err := telemetry.InitTelemetry(telConfig); err != nil {
		slog.Warn("Telemetry init failed", "error", err)
	}
	slog.Info("Telemetry engine initialized", "enabled", telConfig.Enabled, "mode", telConfig.Mode)

	// 2. Load Configuration
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

	// 4. JWT Config
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

	// 5. Core Storage & Engines
	dbStore, err := store.NewPostgresStore(dbURL)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer dbStore.Close()

	// 5b. Error log — wrap the default slog handler so every Warn and
	// Error record is also persisted to errors_log. Must run after
	// the store pool exists (the writer needs it) and before any
	// request is served.
	innerHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	errHandler := api.NewErrorLogHandler(innerHandler, 1024)
	slog.SetDefault(slog.New(errHandler))
	go api.StartErrorLogWriter(context.Background(), errHandler, dbStore.Pool())

	_ = lineage.NewGraph(1000)
	lineageEngine := engine.NewLineageEngine(dbStore)
	contradictionEngine := engine.NewContradictionEngine(dbStore)

	shield := engine.NewPreFlightShield(contradictionEngine)
	topologyExecutor := topology.NewExecutor(dbStore, shield)
	topologyGenerator := topology.NewGenerator(dbStore)

	authService := auth.NewAuthService(dbStore, jwtConfig)

	// 6. API Server
	//
	// NewServer constructs the rate limiter internally when none is
	// passed. The Control Plane pool is attached after construction,
	// via SetControlPool, because it depends on an optional env var
	// that NewServer should not read.
	server := api.NewServer(
		dbStore,
		authService,
		jwtConfig,
		contradictionEngine,
		lineageEngine,
		topologyGenerator,
		topologyExecutor,
	)

	// Control Plane read-only pool. Optional: if CONTROL_DATABASE_URL
	// is unset, the Control Plane returns 404 for every request and
	// the tenant dashboard is unaffected.
	if controlURL := os.Getenv("CONTROL_DATABASE_URL"); controlURL != "" {
		controlPool, err := pgxpool.New(context.Background(), controlURL)
		if err != nil {
			slog.Warn("Control Plane pool failed to open; Control Plane will return 404", "error", err)
		} else {
			server.SetControlPool(controlPool)
			defer controlPool.Close()
		}
	}

	// 7. Rate Limiter
	rateLimiter := server.RateLimiter()

	// 8. Build Router
	handler := api.SetupRouter(server, rateLimiter, mcp.BridgeHandler("http://localhost:8080"), nil)

	// 9. HTTP Server
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

	// 10. Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		slog.Info("Garuda API Secure Gateway online", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Gateway server crashed unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

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
