// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/api"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/engine"
	"github.com/myshra777-ai/garuda/internal/runtime"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/telemetry"
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

		dbURL := getDBURL()
		if err := store.Migrate(dbURL, "migrations"); err != nil {
			return fmt.Errorf("database migration failed: %w", err)
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

		// Create the default admin user on first boot. Prints the
		// generated password once to stderr.
		if err := server.BootstrapAdminIfEmpty(ctx); err != nil {
			slog.Error("failed to bootstrap admin user", "error", err)
			os.Exit(1)
		}

		// The dev daemon uses the same route table as the production
		// API server. Before this change, dev_cmd.go carried its own
		// manually-curated route list, which had drifted from
		// handler.go's RegisterRoutes. The agent orchestration routes
		// (checkpoint, handoff, resume, lineage) existed in production
		// and returned 404 in dev. Any future route added to
		// RegisterRoutes is now available in dev without a second edit.
		//
		// api.SetupRouter owns the middleware chain (rate limiting,
		// recovery, request ID) and calls RegisterRoutes internally.
		// The dev-only graph overrides are registered via the override
		// hook: gorilla/mux matches the most recently registered route
		// for a given path, so these replace the production handlers
		// for /graph and /api/v1/graph.
		router := api.SetupRouter(server, server.RateLimiter(), nil, func(r *mux.Router) {
			r.HandleFunc("/graph", server.RequireSession(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write([]byte(visualizerHTML))
			})).Methods(http.MethodGet)

			r.HandleFunc("/api/v1/graph", server.RequireSession(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.Header.Get("Accept"), "text/html") &&
					!strings.Contains(r.URL.Query().Get("format"), "json") {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					_, _ = w.Write([]byte(visualizerHTML))
					return
				}
				server.HandleGraph(w, r)
			})).Methods(http.MethodGet)
		})

		bindAddr := getBindAddr()

		httpServer := &http.Server{
			Addr:    bindAddr,
			Handler: router,
		}

		// ─────────────────────────────────────────────────────────────
		// Background Worker Loop (10-second Merkle epoch)
		// Iterates every workspace in the tenant, not just uuid-ws.
		// ─────────────────────────────────────────────────────────────
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			tenantID := getTenantID()
			verifier := runtime.NewVerificationEngine(pool)

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					rows, err := pool.Query(ctx, `
						SELECT id FROM workspaces WHERE tenant_id = $1
					`, tenantID)
					if err != nil {
						continue
					}

					var workspaceIDs []uuid.UUID
					for rows.Next() {
						var id uuid.UUID
						if err := rows.Scan(&id); err == nil {
							workspaceIDs = append(workspaceIDs, id)
						}
					}
					rows.Close()

					for _, wsID := range workspaceIDs {
						_, _ = verifier.RecomputeWorkspaceVerification(ctx, tenantID, wsID)
					}
					_, _ = pgStore.CreateUnifiedMerkleSnapshot(ctx, tenantID)
				}
			}
		}()

		// ─────────────────────────────────────────────────────────────
		// Background Aggregate Refresher
		// Computes telemetry_aggregates for today and yesterday on
		// startup, then on a 15-minute ticker. Read by the admin
		// dashboard. See internal/telemetry/refresh.go.
		// ─────────────────────────────────────────────────────────────
		telemetry.StartAggregateRefresher(ctx, pgStore)

		// ─────────────────────────────────────────────────────────────
		// Start HTTP API
		// ─────────────────────────────────────────────────────────────
		go func() {
			fmt.Printf("🚀 Garuda Unified Daemon running at http://%s\n", bindAddr)
			fmt.Printf("   • Login:               GET  http://%s/login\n", bindAddr)
			fmt.Printf("   • Dashboard:           GET  http://%s/dashboard?workspace=<name>\n", bindAddr)
			fmt.Printf("   • Dashboard Stats API: GET  http://%s/api/v1/dashboard/stats?workspace=<name>\n", bindAddr)
			fmt.Printf("   • Dashboard Search:    GET  http://%s/api/v1/dashboard/search?q=<query>\n", bindAddr)
			fmt.Printf("   • Interactive Graph:   GET  http://%s/graph\n", bindAddr)
			fmt.Printf("   • Graph JSON API:      GET  http://%s/api/v1/graph?format=json\n", bindAddr)
			fmt.Printf("   • Telemetry Ingestion: POST http://%s/api/v1/telemetry/spans\n", bindAddr)
			fmt.Printf("   • Runtime Coverage:    GET  http://%s/api/v1/runtime/coverage\n", bindAddr)
			fmt.Printf("   • Dual-Root Merkle:    GET  http://%s/api/v1/merkle/state\n", bindAddr)
			fmt.Printf("   • SSE Live Events:     GET  http://%s/api/v1/events\n", bindAddr)
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
    <title>Garuda Architecture Map</title>
    <script src="https://d3js.org/d3.v7.min.js"></script>
    <style>
        body { margin: 0; background: #F7F7F3; background-image: radial-gradient(#CECEC5 1px, transparent 1px); background-size: 20px 20px; color: #191B19; font-family: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; overflow: hidden; }
        #header { position: absolute; top: 16px; left: 16px; z-index: 10; background: rgba(255, 255, 255, 0.94); padding: 14px 20px; border-radius: 10px; border: 1px solid #E7E7E1; box-shadow: 0 4px 16px rgba(25,27,25,0.06); }
        h1 { margin: 0 0 4px 0; font-size: 16px; font-weight: 650; color: #191B19; letter-spacing: -0.01em; }
        .meta { font-size: 11.5px; color: #767A72; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
        svg { width: 100vw; height: 100vh; }
        .node circle { stroke-width: 2px; cursor: pointer; transition: all 0.15s; }
        .node text { font-size: 11px; fill: #191B19; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-weight: 600; pointer-events: none; }
        .link { stroke: #9FA39A; stroke-opacity: 0.55; stroke-width: 1.4px; }
        .link.CONTRADICTED { stroke: #991B1B; stroke-width: 2.2px; stroke-dasharray: 5,4; }
        .node.CONTRADICTED circle { fill: #991B1B !important; stroke: #F3B8B8; }
        .node.SUPPORTED circle { fill: #FFFFFF; stroke: #191B19; }
        .node.repository circle { fill: #191B19; stroke: #D5E66A; stroke-width: 3px; }
        #tooltip { position: absolute; display: none; background: #FFFFFF; border: 1px solid #E7E7E1; padding: 10px 14px; border-radius: 8px; font-size: 12px; color: #191B19; pointer-events: none; z-index: 100; box-shadow: 0 8px 24px rgba(25,27,25,0.12); }
    </style>
</head>
<body>
    <div id="header">
        <h1>Garuda Architecture Map</h1>
        <div class="meta">Focused Epistemic Topology · <a href="/dashboard" style="color:#191B19; font-weight:600;">← Back to Command Center</a></div>
    </div>
    <div id="tooltip"></div>
    <svg></svg>

    <script>
        const width = window.innerWidth, height = window.innerHeight;
        const svg = d3.select("svg");
        const g = svg.append("g");
        svg.call(d3.zoom().scaleExtent([0.15, 4]).on("zoom", (e) => g.attr("transform", e.transform)));

        const tooltip = d3.select("#tooltip");
        const params = new URLSearchParams(window.location.search);
        const workspace = params.get("workspace") || "";

        fetch("/api/v1/graph?format=json" + (workspace ? "&workspace=" + encodeURIComponent(workspace) : ""))
            .then(res => res.json())
            .then(data => {
                const nodes = data.nodes || [];
                const links = (data.edges || []).map(d => ({ ...d, source: d.from, target: d.to }));

                const simulation = d3.forceSimulation(nodes)
                    .force("link", d3.forceLink(links).id(d => d.id).distance(120))
                    .force("charge", d3.forceManyBody().strength(-320))
                    .force("center", d3.forceCenter(width / 2, height / 2))
                    .force("collision", d3.forceCollide().radius(28));

                const link = g.append("g")
                    .selectAll("line")
                    .data(links)
                    .enter().append("line")
                    .attr("class", d => "link " + (d.status || ""))
                    .attr("stroke", d => d.status === "CONTRADICTED" ? "#991B1B" : "#9FA39A");

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
                    .attr("r", d => d.kind === "repository" ? 13 : 9)
                    .on("mouseover", (e, d) => {
                        tooltip.style("display", "block")
                            .html("<strong>" + d.label + "</strong><br/>Kind: " + d.kind + "<br/>Status: " + d.status + (d.count ? "<br/>Connections: " + d.count : ""))
                            .style("left", (e.pageX + 12) + "px")
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

// getBindAddr returns the address the daemon should bind to.
//
// Default is 127.0.0.1:8080 — loopback only, not reachable from the
// network. To expose the daemon on all interfaces, set GARUDA_BIND
// to an explicit address such as 0.0.0.0:8080 or ":8080". This is a
// deliberate opt-in: the previous default was ":8080" (all
// interfaces), which made every workspace reachable on any network
// the host was connected to.
func getBindAddr() string {
	if v := strings.TrimSpace(os.Getenv("GARUDA_BIND")); v != "" {
		return v
	}
	if p := strings.TrimSpace(os.Getenv("PORT")); p != "" {
		return "127.0.0.1:" + p
	}
	return "127.0.0.1:8080"
}
