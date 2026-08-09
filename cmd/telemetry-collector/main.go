package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TelemetryEvent struct {
	InstanceHash  string `json:"instance_hash"`
	SessionID     string `json:"session_id"`
	Mode          string `json:"mode"`
	GarudaVersion string `json:"garuda_version"`
	AgentRuntime  string `json:"agent_runtime"`
	ModelProvider string `json:"model_provider"`
	ModelName     string `json:"model_name"`
	TokensSaved   int64  `json:"tokens_saved"`
}

type TelemetryBatch struct {
	Version string           `json:"version"`
	Events  []TelemetryEvent `json:"events"`
	Count   int              `json:"count"`
}

type CollectorServer struct {
	pool *pgxpool.Pool
	mu   sync.RWMutex
}

func (c *CollectorServer) setPool(pool *pgxpool.Pool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pool = pool
}

func (c *CollectorServer) getPool() *pgxpool.Pool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.pool
}

func main() {
	port := os.Getenv("TELEMETRY_PORT")
	if port == "" {
		port = "8081"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	collector := &CollectorServer{}

	mux := http.NewServeMux()

	// Health check endpoint (always available immediately for Docker healthcheck)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Telemetry ingest endpoint
	mux.HandleFunc("/v1/telemetry/ping", collector.handleIngest)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 1. Start HTTP server FIRST so Docker healthchecks pass immediately
	go func() {
		slog.Info("Telemetry Collector HTTP Server listening", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Collector HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// 2. Connect to database in background / retry loop
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go initDatabaseWithRetry(ctx, dbURL, collector)

	// 3. Graceful Shutdown
	<-ctx.Done()
	slog.Info("Shutting down Telemetry Collector gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful HTTP server shutdown failed", "error", err)
	}

	if pool := collector.getPool(); pool != nil {
		pool.Close()
		slog.Info("Database pool closed cleanly")
	}

	slog.Info("Telemetry Collector stopped")
}

func initDatabaseWithRetry(ctx context.Context, dbURL string, collector *CollectorServer) {
	backoff := 1 * time.Second
	maxBackoff := 15 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		slog.Info("Attempting database connection...")
		pool, err := pgxpool.New(ctx, dbURL)
		if err == nil {
			// Ensure schema exists
			createTableQuery := `
			CREATE TABLE IF NOT EXISTS telemetry_events (
				id BIGSERIAL PRIMARY KEY,
				instance_hash TEXT,
				session_id TEXT,
				mode TEXT,
				garuda_version TEXT,
				agent_runtime TEXT,
				model_provider TEXT,
				model_name TEXT,
				tokens_saved BIGINT,
				created_at TIMESTAMPTZ DEFAULT NOW()
			);`

			if _, err := pool.Exec(ctx, createTableQuery); err == nil {
				slog.Info("Database pool connected and schema validated successfully")
				collector.setPool(pool)
				return
			} else {
				slog.Warn("Failed to execute telemetry schema creation", "error", err)
				pool.Close()
			}
		} else {
			slog.Warn("Failed to create database pool", "error", err)
		}

		slog.Info("Retrying database connection", "retry_in", backoff)
		time.Sleep(backoff)
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

func (c *CollectorServer) handleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pool := c.getPool()
	if pool == nil {
		slog.Warn("Ingest request rejected: database connection initializing")
		http.Error(w, "database initializing", http.StatusServiceUnavailable)
		return
	}

	var batch TelemetryBatch
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	tx, err := pool.Begin(ctx)
	if err != nil {
		slog.Error("Failed to start database transaction", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	insertQuery := `
		INSERT INTO telemetry_events (
			instance_hash, session_id, mode, garuda_version, agent_runtime,
			model_provider, model_name, tokens_saved
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	for _, evt := range batch.Events {
		_, err := tx.Exec(ctx, insertQuery,
			evt.InstanceHash, evt.SessionID, evt.Mode, evt.GarudaVersion, evt.AgentRuntime,
			evt.ModelProvider, evt.ModelName, evt.TokensSaved,
		)
		if err != nil {
			slog.Error("Failed to insert telemetry event", "error", err)
			http.Error(w, "failed to record telemetry events", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("Failed to commit telemetry transaction", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
