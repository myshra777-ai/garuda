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
	"strings"
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
		
		// Serve Interactive Visualizer HTML on /graph or when Accept header contains text/html
		mux.HandleFunc("/graph", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(visualizerHTML))
		})

		mux.HandleFunc("/api/v1/graph", func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get("Accept"), "text/html") && !strings.Contains(r.URL.Query().Get("format"), "json") {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write([]byte(visualizerHTML))
				return
			}
			server.HandleGraph(w, r)
		})

		httpServer := &http.Server{
			Addr:    ":8080",
			Handler: mux,
		}

		// Background Worker Loop (10-second Merkle epoch)
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

		// Start HTTP API
		go func() {
			fmt.Println("🚀 Garuda Unified Daemon running at http://localhost:8080")
			fmt.Println("   • Interactive Graph:   GET  http://localhost:8080/graph")
			fmt.Println("   • Telemetry Ingestion: POST http://localhost:8080/api/v1/telemetry/spans")
			fmt.Println("   • Runtime Coverage:    GET  http://localhost:8080/api/v1/runtime/coverage")
			fmt.Println("   • Dual-Root Merkle:    GET  http://localhost:8080/api/v1/merkle/state")
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

const visualizerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Garuda Epistemic Truth Graph</title>
    <script src="https://d3js.org/d3.v7.min.js"></script>
    <style>
        body { margin: 0; background: #0f172a; color: #f8fafc; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; overflow: hidden; }
        #header { position: absolute; top: 16px; left: 16px; z-index: 10; background: rgba(15, 23, 42, 0.85); padding: 12px 20px; border-radius: 8px; border: 1px solid #334155; backdrop-filter: blur(8px); }
        h1 { margin: 0 0 4px 0; font-size: 18px; font-weight: 600; color: #38bdf8; }
        .meta { font-size: 12px; color: #94a3b8; }
        svg { width: 100vw; height: 100vh; }
        .node circle { stroke-width: 2px; cursor: pointer; transition: all 0.2s; }
        .node text { font-size: 11px; fill: #cbd5e1; pointer-events: none; }
        .link { stroke-opacity: 0.6; stroke-width: 1.5px; }
        .link.CONTRADICTED { stroke: #ef4444; stroke-width: 2.5px; stroke-dasharray: 4; }
        .node.CONTRADICTED circle { fill: #ef4444 !important; stroke: #fca5a5; }
        .node.SUPPORTED circle { fill: #10b981; stroke: #6ee7b7; }
        .node.repository circle { fill: #6366f1; stroke: #a5b4fc; }
        #tooltip { position: absolute; display: none; background: #1e293b; border: 1px solid #475569; padding: 10px 14px; border-radius: 6px; font-size: 12px; pointer-events: none; z-index: 100; box-shadow: 0 10px 15px -3px rgba(0,0,0,0.5); }
    </style>
</head>
<body>
    <div id="header">
        <h1>🦅 Garuda Epistemic Truth Graph</h1>
        <div class="meta">Dual-Root Merkle Engine · Real-time AST & Runtime Verification</div>
    </div>
    <div id="tooltip"></div>
    <svg></svg>

    <script>
        const width = window.innerWidth, height = window.innerHeight;
        const svg = d3.select("svg");
        const g = svg.append("g");
        svg.call(d3.zoom().scaleExtent([0.1, 4]).on("zoom", (e) => g.attr("transform", e.transform)));

        const tooltip = d3.select("#tooltip");

        fetch("/api/v1/graph?format=json")
            .then(res => res.json())
            .then(data => {
                const nodes = data.nodes || [];
                const links = (data.edges || []).map(d => ({ ...d, source: d.from, target: d.to }));

                const simulation = d3.forceSimulation(nodes)
                    .force("link", d3.forceLink(links).id(d => d.id).distance(120))
                    .force("charge", d3.forceManyBody().strength(-300))
                    .force("center", d3.forceCenter(width / 2, height / 2))
                    .force("collision", d3.forceCollide().radius(25));

                const link = g.append("g")
                    .selectAll("line")
                    .data(links)
                    .enter().append("line")
                    .attr("class", d => "link " + (d.status || ""))
                    .attr("stroke", d => d.status === "CONTRADICTED" ? "#ef4444" : "#475569");

                const node = g.append("g")
                    .selectAll("g")
                    .data(nodes)
                    .enter().append("g")
                    .attr("class", d => "node " + (d.status || "") + " " + (d.kind || ""))
                    .call(d3.drag()
                        .on("start", (e, d) => { if (!e.active) simulation.alphaTarget(0.3).restart(); d.fx = d.x; d.fy = d.y; })
                        .on("drag", (e, d) => { d.fx = e.x; d.fy = e.y; })
                        .on("end", (e, d) => { if (!e.active) simulation.alphaTarget(0); d.fx = null; d.fy = null; }));

                node.append("circle")
                    .attr("r", d => d.kind === "repository" ? 12 : 8)
                    .on("mouseover", (e, d) => {
                        tooltip.style("display", "block")
                            .html("<strong>" + d.label + "</strong><br/>Kind: " + d.kind + "<br/>Status: " + d.status + (d.count ? "<br/>Entities: " + d.count : ""))
                            .style("left", (e.pageX + 10) + "px")
                            .style("top", (e.pageY - 10) + "px");
                    })
                    .on("mouseout", () => tooltip.style("display", "none"));

                node.append("text")
                    .attr("dx", 14)
                    .attr("dy", 4)
                    .text(d => d.label);

                simulation.on("tick", () => {
                    link.attr("x1", d => d.source.x).attr("y1", d => d.source.y)
                        .attr("x2", d => d.target.x).attr("y2", d => d.target.y);
                    node.attr("transform", d => "translate(" + d.x + "," + d.y + ")");
                });
            });
    </script>
</body>
</html>`

func init() {
	rootCmd.AddCommand(devCmd)
}
