// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/api"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/engine"
	"github.com/myshra777-ai/garuda/internal/runtime"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/topology"
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start the unified Garuda daemon (API server + background verification worker)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			fmt.Println("\n🛑 Shutting down Garuda daemon...")
			cancel()
		}()

		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			dbURL = "postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
		}

		pool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer pool.Close()

		pgStore, err := store.NewPostgresStore(dbURL)
		if err != nil {
			return fmt.Errorf("failed to initialize postgres store: %w", err)
		}

		jwtConfig, err := auth.NewJWTConfig("garuda-dev", "garuda-api", 24*time.Hour)
		if err != nil {
			return fmt.Errorf("failed to initialize jwt config: %w", err)
		}

		authService := auth.NewAuthService(pgStore, jwtConfig)
		contraEngine := engine.NewContradictionEngine(pgStore)
		lineageEngine := engine.NewLineageEngine(pgStore)
		shield := engine.NewPreFlightShield(contraEngine)
		topoGen := topology.NewGenerator(pgStore)
		topoExec := topology.NewExecutor(pgStore, shield)

		server := api.NewServer(pgStore, authService, jwtConfig, contraEngine, lineageEngine, topoGen, topoExec)

		mux := http.NewServeMux()
		mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"healthy","service":"garuda-unified"}`))
		})
		mux.HandleFunc("/api/v1/telemetry/spans", server.HandleIngestRuntimeSpans)
		mux.HandleFunc("/api/v1/runtime/coverage", server.HandleGetRuntimeCoverage)
		mux.HandleFunc("/api/v1/merkle/state", server.HandleGetMerkleState)
		mux.HandleFunc("/api/v1/graph", server.HandleGraph)

		httpServer := &http.Server{
			Addr:    ":8080",
			Handler: mux,
		}

		// 1. Background Worker Loop (Automatic Merkle epoch computation every 10s)
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
			verifier := runtime.NewVerificationEngine(pool)

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					var workspaceID uuid.UUID
					err := pool.QueryRow(ctx, `SELECT id FROM workspaces WHERE name = 'uuid-ws' LIMIT 1`).Scan(&workspaceID)
					if err == nil {
						_, _ = verifier.RecomputeWorkspaceVerification(ctx, workspaceID, tenantID)
						_, _ = pgStore.CreateUnifiedMerkleSnapshot(ctx, tenantID)
					}
				}
			}
		}()

		// 2. Start HTTP API
		go func() {
			fmt.Println("🚀 Garuda Unified Daemon running at http://localhost:8080")
			fmt.Println("   • Telemetry Ingestion: POST http://localhost:8080/api/v1/telemetry/spans")
			fmt.Println("   • Runtime Coverage:    GET  http://localhost:8080/api/v1/runtime/coverage")
			fmt.Println("   • Dual-Root Merkle:    GET  http://localhost:8080/api/v1/merkle/state")
			fmt.Println("   • Graph Visualizer:    GET  http://localhost:8080/api/v1/graph")
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Printf("HTTP server error: %v\n", err)
			}
		}()

		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutdownCancel()
		return httpServer.Shutdown(shutdownCtx)
	},
}

func init() {
	rootCmd.AddCommand(devCmd)
}
