// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/budget"
	"github.com/myshra777-ai/garuda/internal/config"
	"github.com/myshra777-ai/garuda/internal/engine"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/tenant"
	"github.com/myshra777-ai/garuda/internal/types"
)

// MCPServer holds the process-scoped state of one MCP session.
type MCPServer struct {
	store               *store.PostgresStore
	lineageEngine       *engine.LineageEngine
	contradictionEngine *engine.ContradictionEngine
	authService         *auth.AuthService
	jwtConfig           *auth.JWTConfig

	sessionID    string
	startedAt    time.Time
	instanceHash string
	clientName   string
	sessionDBID  uuid.UUID
	sessionMu    sync.RWMutex
	oriented     bool
}

func main() {
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--help", "-h", "help":
			fmt.Println("Garuda MCP Server v1.9.6 (JSON-RPC 2.0 over stdio — 40 governance & AST tools)")
			fmt.Println("Usage: garuda-mcp [--version | --help]")
			return
		case "--version", "-v", "version":
			fmt.Println("Garuda MCP Server v1.9.6")
			return
		}
	}

	_ = os.Setenv("GARUDA_MCP_QUIET", "1")
	if os.Getenv("GARUDA_MCP_DEBUG") == "" {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	}

	enc := json.NewEncoder(os.Stdout)

	config.LoadDotEnv()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" && os.Getenv("GARUDA_ENV") != "production" {
		dbURL = config.ResolveDBURL()
	}
	if dbURL == "" {
		_ = enc.Encode(MCPResponse{
			JSONRPC: "2.0",
			Error: &MCPError{
				Code:    -32000,
				Message: "DATABASE_URL environment variable is required",
			},
		})
		slog.Error("DATABASE_URL is not set")
		os.Exit(1)
	}

	dbStore, err := store.NewPostgresStore(dbURL)
	if err != nil {
		_ = enc.Encode(MCPResponse{
			JSONRPC: "2.0",
			Error: &MCPError{
				Code:    -32000,
				Message: fmt.Sprintf("database connection failed: %v", err),
			},
		})
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbStore.Close()

	jwtConfig, _ := auth.NewJWTConfig("garuda-mcp", "garuda-api", 15*time.Minute)
	authService := auth.NewAuthService(dbStore, jwtConfig)
	lineageEngine := engine.NewLineageEngine(dbStore)
	contradictionEngine := engine.NewContradictionEngine(dbStore)

	server := &MCPServer{
		store:               dbStore,
		lineageEngine:       lineageEngine,
		contradictionEngine: contradictionEngine,
		authService:         authService,
		jwtConfig:           jwtConfig,
		sessionID:           uuid.New().String(),
		startedAt:           time.Now().UTC(),
		instanceHash:        computeInstanceHash(),
	}

	slog.Info("Garuda MCP Server ready",
		"session_id", server.sessionID,
		"pid", os.Getpid(),
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		slog.Info("Received shutdown signal", "session_id", server.sessionID)
		server.closeSession(context.Background(), "graceful")
		os.Exit(0)
	}()

	dec := json.NewDecoder(os.Stdin)
	requestCount := 0

	for {
		var req MCPRequest
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				slog.Info("stdin closed, shutting down",
					"session_id", server.sessionID,
					"requests_handled", requestCount,
				)
				server.closeSession(context.Background(), "graceful")
				return
			}
			slog.Warn("Failed to parse MCP request",
				"session_id", server.sessionID,
				"error", err,
			)
			_ = enc.Encode(MCPResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error: &MCPError{
					Code:    -32700,
					Message: fmt.Sprintf("Parse error: %v", err),
				},
			})
			continue
		}

		requestCount++

		if req.ID == nil {
			continue
		}

		resp := server.handleRequest(req)

		if err := enc.Encode(resp); err != nil {
			slog.Error("Failed to write MCP response",
				"session_id", server.sessionID,
				"error", err,
			)
			return
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// MCP JSON-RPC types
// ─────────────────────────────────────────────────────────────────────────────

type MCPRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Tool registry (Formatted cleanly for BUG-20 / CLASS-01)
// ─────────────────────────────────────────────────────────────────────────────

var allTools = []map[string]interface{}{
	{
		"name":        "garuda.briefing",
		"description": "Session-start briefing: workspace state, policy verdict, epistemic counts, trust anchor, scale, top hubs, active policies, open contradictions, and what changed since this agent's last briefing. Read-only. Call this first.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name. Defaults to the server's active workspace.",
				},
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Caller-supplied agent identifier. Defaults to the MCP client name from the handshake.",
				},
			},
		},
	},
	{
		"name":        "garuda.query",
		"description": "Query the Garuda knowledge graph (entities, document claims, and decisions) with natural language keywords. Rejects raw SQL statements.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The natural language query",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Optional workspace name",
				},
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional tenant ID",
				},
			},
			"required": []string{"query"},
		},
	},
	{
		"name":        "garuda.get_lineage",
		"description": "Get the full governance and AST lineage of a decision UUID, actor name (e.g. 'cli-operator'), or symbol name.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"decision_id": map[string]interface{}{
					"type":        "string",
					"description": "Decision UUID, actor identifier, or symbol name",
				},
				"symbol": map[string]interface{}{
					"type":        "string",
					"description": "Optional symbol name (alias for decision_id)",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Optional entity/symbol name",
				},
				"entity_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional entity UUID",
				},
				"actor": map[string]interface{}{
					"type":        "string",
					"description": "Optional actor filter",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Optional workspace name",
				},
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant ID",
				},
			},
		},
	},
	{
		"name":        "garuda.detect_contradictions",
		"description": "Detect unresolved contradictions across decisions and ADR document claims in the tenant knowledge graph.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant ID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Optional workspace name",
				},
				"scope": map[string]interface{}{
					"type":        "string",
					"description": "Optional scope filter",
				},
				"status": map[string]interface{}{
					"type":        "string",
					"description": "Optional status filter: open, resolved, or all (default open)",
				},
			},
		},
	},
	{
		"name":        "garuda.get_impact",
		"description": "Find what breaks if a decision UUID or code symbol (e.g. 'AppendLeafAndSealTx') is changed. Computes transitive blast radius and child decision impact.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"decision_id": map[string]interface{}{
					"type":        "string",
					"description": "Decision UUID, entity UUID, or symbol name",
				},
				"symbol": map[string]interface{}{
					"type":        "string",
					"description": "Target symbol name (e.g. 'AppendLeafAndSealTx')",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Alias for symbol name",
				},
				"entity_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional entity UUID or symbol name",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Optional workspace name",
				},
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant ID",
				},
			},
		},
	},
	{
		"name":        "garuda.blast_radius",
		"description": "Compute the blast radius of a symbol change: which entities consume the target entity, directly or transitively, up to a depth. Read-only. Returns a critical-first list and per-severity counts.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"entity_id": map[string]interface{}{
					"type":        "string",
					"description": "Target entity UUID or symbol name",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Target symbol name (alternative to entity_id)",
				},
				"symbol": map[string]interface{}{
					"type":        "string",
					"description": "Target symbol name (alternative to entity_id)",
				},
				"depth": map[string]interface{}{
					"type":        "integer",
					"description": "Max traversal depth. Default 3.",
				},
				"min_confidence": map[string]interface{}{
					"type":        "number",
					"description": "Minimum edge confidence. Default 0.50.",
				},
			},
		},
	},
	{
		"name":        "garuda.propose_decision",
		"description": "Propose a new architectural decision (creates a draft).",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "Decision title",
				},
				"statement": map[string]interface{}{
					"type":        "string",
					"description": "Normative decision statement or rationale",
				},
				"rationale": map[string]interface{}{
					"type":        "string",
					"description": "Alias for statement",
				},
				"scope_domain": map[string]interface{}{
					"type":        "string",
					"description": "Scope domain",
				},
				"scope_system": map[string]interface{}{
					"type":        "string",
					"description": "Scope system",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Optional workspace name",
				},
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant ID",
				},
			},
			"required": []string{"title"},
		},
	},
	{
		"name":        "garuda.policy.list",
		"description": "List policies registered for a tenant. Read-only.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID; falls back to GARUDA_TENANT_ID env or default tenant",
				},
			},
		},
	},
	{
		"name":        "garuda.governance.status",
		"description": "Aggregate governance state: top-level decision, active policies, documentation health, blocked_by list, and unified entity/claim counts. Read-only.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name; falls back to GARUDA_WORKSPACE env or 'default'",
				},
			},
		},
	},
	{
		"name":        "garuda.check_drift",
		"description": "Bidirectional documentation-to-code (doc_to_code) and code-to-documentation (code_to_doc) drift report. Read-only.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
			},
		},
	},
	{
		"name":        "garuda.query_claims",
		"description": "Query document claims matching a subject and/or epistemic status (SUPPORTED, UNVERIFIED, CONTRADICTED). Read-only.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"subject": map[string]interface{}{
					"type":        "string",
					"description": "Filter by subject or object (substring)",
				},
				"status": map[string]interface{}{
					"type":        "string",
					"description": "Optional status filter: SUPPORTED, UNVERIFIED, CONTRADICTED, or ALL",
				},
			},
		},
	},
	{
		"name":        "garuda.policy.evaluate",
		"description": "Dry-run the policy engine against a workspace and return deterministic policy decisions. Does NOT persist or Merkle-anchor. Use 'garuda policy evaluate' from the CLI to anchor a real governance decision.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"policy_dir": map[string]interface{}{
					"type":        "string",
					"description": "Directory containing policy YAML. Defaults to GARUDA_POLICY_DIR or ./policies",
				},
				"subject_kind": map[string]interface{}{
					"type":        "string",
					"description": "Optional subject kind",
				},
				"subject_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional subject ID",
				},
				"actor": map[string]interface{}{
					"type":        "string",
					"description": "Optional actor identifier",
				},
			},
		},
	},
	{
		"name":        "garuda.verify_policy_evaluation",
		"description": "Verify the Merkle inclusion proof for a persisted policy evaluation. Returns valid=true if the proof recomputes to the anchored root. Read-only; no side effects.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"evaluation_id": map[string]interface{}{
					"type":        "string",
					"description": "Policy evaluation UUID returned by garuda policy evaluate --save",
				},
			},
			"required": []string{"evaluation_id"},
		},
	},
	{
		"name":        "garuda.verify_evidence",
		"description": "Verify the RFC 6962 Merkle inclusion proof or SHA-256 evidence digest for a policy evaluation, evidence block, or decision UUID. Read-only; no side effects.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"evidence_id": map[string]interface{}{
					"type":        "string",
					"description": "Evidence block or policy evaluation UUID",
				},
				"evaluation_id": map[string]interface{}{
					"type":        "string",
					"description": "Alias for evidence_id (policy evaluation UUID)",
				},
				"decision_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional decision UUID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Optional workspace name",
				},
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
			},
		},
	},
	{
		"name":        "garuda.entities",
		"description": "List semantic entities in a workspace. Filterable by package and kind. Read-only.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"package": map[string]interface{}{
					"type":        "string",
					"description": "Optional package name substring",
				},
				"kind": map[string]interface{}{
					"type":        "string",
					"description": "Optional entity kind (struct, interface, function, method)",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Max rows, default 100, cap 1000",
				},
			},
		},
	},
	{
		"name":        "garuda.inspect",
		"description": "Inspect one entity with its signature and incoming and outgoing relationships. Read-only.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"entity_id": map[string]interface{}{
					"type":        "string",
					"description": "Entity UUID or symbol name",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Optional symbol name",
				},
			},
			"required": []string{"entity_id"},
		},
	},
	{
		"name":        "garuda.subclasses",
		"description": "List entities that inherit from or embed the subject. Uses INHERITS and EMBEDS claims. Read-only.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"entity_id": map[string]interface{}{
					"type":        "string",
					"description": "Subject entity UUID",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Subject entity name (alternative to entity_id)",
				},
				"language": map[string]interface{}{
					"type":        "string",
					"description": "Optional language filter (go, python, typescript)",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Max rows, default 50, cap 500",
				},
			},
		},
	},
	{
		"name":        "garuda.implementers",
		"description": "List entities that implement the subject interface. Uses IMPLEMENTS claims. Read-only.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"entity_id": map[string]interface{}{
					"type":        "string",
					"description": "Subject interface UUID",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Subject interface name (alternative to entity_id)",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Max rows, default 50, cap 500",
				},
			},
		},
	},
	{
		"name":        "garuda.neighbors",
		"description": "List every entity connected to the subject in one hop, both incoming and outgoing directions. External stubs excluded by default. Read-only.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"entity_id": map[string]interface{}{
					"type":        "string",
					"description": "Subject entity UUID or symbol name",
				},
				"edge_kinds": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Optional filter, e.g. ['CALLS','IMPORTS']",
				},
				"include_external": map[string]interface{}{
					"type":        "boolean",
					"description": "Include external stubs (default false)",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Max rows per direction, default 100, cap 500",
				},
			},
			"required": []string{"entity_id"},
		},
	},
	{
		"name":        "garuda.create_task",
		"description": "Create a new task and provision source/target agent UUIDs in the workspace so garuda.handoff and garuda.resume can operate on persisted UUIDs.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"title": map[string]interface{}{
					"type":        "string",
					"description": "Task title",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Optional task description",
				},
				"priority": map[string]interface{}{
					"type":        "number",
					"description": "Task priority (default 1)",
				},
				"owner_agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional owner/source agent UUID or alias",
				},
				"target_agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional target agent UUID or alias",
				},
				"scope_domain": map[string]interface{}{
					"type":        "string",
					"description": "Optional scope domain",
				},
				"scope_system": map[string]interface{}{
					"type":        "string",
					"description": "Optional scope system",
				},
			},
			"required": []string{"title"},
		},
	},
	{
		"name":        "garuda.handoff",
		"description": "Initiate an atomic task handoff between two agents. Creates a checkpoint of the source agent's state, transfers task ownership, and transitions both agents' statuses. Call garuda.create_task first to provision UUIDs.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"task_id": map[string]interface{}{
					"type":        "string",
					"description": "Task UUID returned by garuda.create_task (or symbolic task identifier)",
				},
				"source_agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Source agent UUID or symbolic identifier",
				},
				"target_agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Target agent UUID or symbolic identifier",
				},
				"reason": map[string]interface{}{
					"type":        "string",
					"description": "Optional human-readable reason",
				},
				"checkpoint_data": map[string]interface{}{
					"type":        "object",
					"description": "Optional serializable state to carry across the handoff",
				},
			},
			"required": []string{"task_id", "source_agent_id", "target_agent_id"},
		},
	},
	{
		"name":        "garuda.resume",
		"description": "Restore the state from an active checkpoint created by garuda.handoff, and mark the checkpoint as consumed. A second resume of the same checkpoint returns {status: 'not_found'}.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Agent UUID or symbolic identifier performing the resume",
				},
				"checkpoint_id": map[string]interface{}{
					"type":        "string",
					"description": "Checkpoint UUID or task identifier returned by garuda.handoff",
				},
			},
			"required": []string{"agent_id", "checkpoint_id"},
		},
	},
	{
		"name":        "garuda.find_entity",
		"description": "Find entities by exact name, regex name_pattern, kind, package, or file path. Returns status='not_found' with empty results when a filter matches 0 entities. Read-only.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Exact or case-insensitive symbol name to match",
				},
				"name_pattern": map[string]interface{}{
					"type":        "string",
					"description": "Regex against entity name",
				},
				"kind": map[string]interface{}{
					"type":        "string",
					"description": "Entity kind (struct, interface, function, method, class, package)",
				},
				"package": map[string]interface{}{
					"type":        "string",
					"description": "Package name substring",
				},
				"file_path_contains": map[string]interface{}{
					"type":        "string",
					"description": "File path substring",
				},
				"exported_only": map[string]interface{}{
					"type":        "boolean",
					"description": "Only exported entities (default false)",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Max rows, default 50, cap 500",
				},
			},
		},
	},
	{
		"name":        "garuda.talon_pre_edit",
		"description": "MANDATORY pre-edit check: call before modifying any file or symbol. Returns transitive blast radius, callers, applicable ADR MUST/MUST_NOT claims, and active policy guardrails.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"target": map[string]interface{}{
					"type":        "string",
					"description": "Target symbol name or file path about to be edited",
				},
			},
		},
	},
	{
		"name":        "garuda.talon_verify_edit",
		"description": "MANDATORY post-edit verification gate: call immediately after modifying code. Runs in dry-run mode (side_effects=[]) by default; set save_to_ledger=true to append a Merkle ledger block (side_effects=['merkle_ledger_append']).",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tenant_id": map[string]interface{}{
					"type":        "string",
					"description": "Tenant UUID",
				},
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"repo_dir": map[string]interface{}{
					"type":        "string",
					"description": "Repository root directory (default .)",
				},
				"staged": map[string]interface{}{
					"type":        "boolean",
					"description": "Check staged git index changes",
				},
				"diff": map[string]interface{}{
					"type":        "string",
					"description": "Git revision range to check",
				},
				"files": map[string]interface{}{
					"type":        "string",
					"description": "Comma-separated list of modified files",
				},
				"save_to_ledger": map[string]interface{}{
					"type":        "boolean",
					"description": "Explicitly anchor evaluation to RFC 6962 Merkle ledger (default false)",
				},
				"save_ledger": map[string]interface{}{
					"type":        "boolean",
					"description": "Alias for save_to_ledger (default false)",
				},
			},
		},
	},
	{
		"name":        "garuda.disambiguate_logic",
		"description": "Compare competing undocumented formulas across teams with 0.0-1.0 team weights summing to 1.00, surface criticality tiers (HIGH_SEVERE, MEDIUM, LOW), and explicit mode ('live_ast' vs 'fixture').",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"team": map[string]interface{}{
					"type":        "string",
					"description": "Active team context (ops, marketing, management)",
				},
				"concept": map[string]interface{}{
					"type":        "string",
					"description": "Business concept or formula to disambiguate",
				},
			},
		},
	},
	{
		"name":        "garuda.audit_data_discrepancy",
		"description": "Audit cross-team business logic and metric number discrepancies in the Data & Logic Audit Ledger by tracing AST and SQL lineage before production deployment.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"metric": map[string]interface{}{
					"type":        "string",
					"description": "Metric to reconcile",
				},
				"team_a_name": map[string]interface{}{
					"type":        "string",
					"description": "Team A display name",
				},
				"team_a_cr": map[string]interface{}{
					"type":        "number",
					"description": "Team A reported metric value",
				},
				"team_a_crore": map[string]interface{}{
					"type":        "number",
					"description": "Alias for team_a_cr",
				},
				"team_b_name": map[string]interface{}{
					"type":        "string",
					"description": "Team B display name",
				},
				"team_b_cr": map[string]interface{}{
					"type":        "number",
					"description": "Team B reported metric value",
				},
				"team_b_crore": map[string]interface{}{
					"type":        "number",
					"description": "Alias for team_b_cr",
				},
			},
		},
	},
	{
		"name":        "garuda.reconcile_data_audit",
		"description": "Record a reconciliation proposal for a flagged cross-team data discrepancy in the Data & Logic Audit Ledger. Requires GARUDA_MCP_APPROVER or CLI co-signing for human approval.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"audit_id": map[string]interface{}{
					"type":        "string",
					"description": "Data audit item ID",
				},
				"approved_by": map[string]interface{}{
					"type":        "string",
					"description": "Approving Lead Auditor or Architect (must match GARUDA_MCP_APPROVER)",
				},
			},
		},
	},
	{
		"name":        "garuda.audit_financial_discrepancy",
		"description": "Legacy alias for garuda.audit_data_discrepancy: audit cross-team business metric number discrepancies in the Data & Logic Audit Ledger.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"metric": map[string]interface{}{
					"type":        "string",
					"description": "Business metric to reconcile",
				},
				"team_a_name": map[string]interface{}{
					"type":        "string",
					"description": "Team A display name",
				},
				"team_a_cr": map[string]interface{}{
					"type":        "number",
					"description": "Team A reported metric value",
				},
				"team_a_crore": map[string]interface{}{
					"type":        "number",
					"description": "Alias for team_a_cr",
				},
				"team_b_name": map[string]interface{}{
					"type":        "string",
					"description": "Team B display name",
				},
				"team_b_cr": map[string]interface{}{
					"type":        "number",
					"description": "Team B reported metric value",
				},
				"team_b_crore": map[string]interface{}{
					"type":        "number",
					"description": "Alias for team_b_cr",
				},
			},
		},
	},
	{
		"name":        "garuda.synthesize_rulebook",
		"description": "Synthesize unorganized repository architecture, surface criticality tiers, and multi-team formula weights into a structured draft Base Rulebook for human stakeholder review and approval.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"output": map[string]interface{}{
					"type":        "string",
					"description": "Optional output file path (e.g. /tmp/smoke_rulebook.md)",
				},
				"output_path": map[string]interface{}{
					"type":        "string",
					"description": "Alias for output file path",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Alias for output file path",
				},
			},
		},
	},
	{
		"name":        "garuda.decide_logic_variant",
		"description": "Propose or record a decision on a competing multi-team formula variant in PostgreSQL. Requires GARUDA_MCP_APPROVER or CLI co-signing for APPROVED_BY_HUMAN.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"team": map[string]interface{}{
					"type":        "string",
					"description": "Active team role (ops, marketing, management)",
				},
				"variant_id": map[string]interface{}{
					"type":        "string",
					"description": "Variant ID (e.g. loss_ratio_ops_v1)",
				},
				"decision_action": map[string]interface{}{
					"type":        "string",
					"description": "APPROVED_BY_HUMAN, APPROVE_VARIANT, PROPOSE_VARIANT, FREEZE_SEVERE, SET_CUSTOM_WEIGHT, or RESET_TO_PENDING",
				},
				"custom_weight": map[string]interface{}{
					"type":        "number",
					"description": "Optional custom weight (0.0 to 1.0)",
				},
				"approved_by": map[string]interface{}{
					"type":        "string",
					"description": "Optional stakeholder identifier (must match GARUDA_MCP_APPROVER for direct approval)",
				},
			},
		},
	},
	{
		"name":        "garuda.reconcile_financial_audit",
		"description": "Legacy alias for garuda.reconcile_data_audit: record reconciliation for a flagged cross-team metric discrepancy in the Data & Logic Audit Ledger.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"audit_id": map[string]interface{}{
					"type":        "string",
					"description": "Audit item ID",
				},
				"approved_by": map[string]interface{}{
					"type":        "string",
					"description": "Approving Auditor or Architect",
				},
			},
		},
	},
	{
		"name":        "garuda.approve_base_rulebook",
		"description": "Propose or approve the synthesized Base Rulebook with an RFC 6962 SHA-256 digest. Requires GARUDA_MCP_APPROVER or CLI co-signing for human approval.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"approved_by": map[string]interface{}{
					"type":        "string",
					"description": "Approver name or role",
				},
				"output": map[string]interface{}{
					"type":        "string",
					"description": "Optional output file path",
				},
				"output_path": map[string]interface{}{
					"type":        "string",
					"description": "Alias for output file path",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Alias for output file path",
				},
			},
		},
	},
	{
		"name":        "garuda.criticality_tiers",
		"description": "List or register 3-Tier Surface Criticality classifications (HIGH_SEVERE freeze, MEDIUM trade-off, LOW UI box alignment).",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"tier_filter": map[string]interface{}{
					"type":        "string",
					"description": "Optional filter: HIGH_SEVERE, MEDIUM, LOW, or ALL",
				},
				"surface_name": map[string]interface{}{
					"type":        "string",
					"description": "Optional custom surface name to register",
				},
				"package_pattern": map[string]interface{}{
					"type":        "string",
					"description": "Optional package pattern for custom surface",
				},
				"criticality_tier": map[string]interface{}{
					"type":        "string",
					"description": "Optional tier for custom surface (HIGH_SEVERE, MEDIUM, LOW)",
				},
			},
		},
	},
	{
		"name":        "garuda.docs_graph",
		"description": "Retrieve the 4-Tier Document-to-Code & Guardrails Lineage Matrix (Documents -> Normative Rules -> Active Policies -> AST Symbols).",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"doc_type": map[string]interface{}{
					"type":        "string",
					"description": "Optional filter: ADR, SPEC, POLICY, or ALL",
				},
			},
		},
	},
	{
		"name":        "garuda.docs_ingest",
		"description": "Ingest Markdown ADRs or design specifications into document_claims and correlate them against workspace AST symbols.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "File or directory path to ingest (default docs/adr)",
				},
			},
		},
	},
	{
		"name":        "garuda.doctor_diagnose",
		"description": "Run the Garuda System Doctor diagnostic suite across platform, PostgreSQL, Talon contract, and API gateway health.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
			},
		},
	},
	{
		"name":        "garuda.build_matrix_check",
		"description": "Verify the codebase across the 4-target Cross-Platform Build Matrix (linux/amd64, darwin/arm64, windows/amd64, cgo/purego) and return the deterministic BuildContext receipt.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"repo_dir": map[string]interface{}{
					"type":        "string",
					"description": "Repository root path",
				},
			},
		},
	},
	{
		"name":        "garuda.hygiene_report",
		"description": "Run the static code hygiene analyzer on the workspace graph to report advisory unreferenced candidates, duplicate symbol names, and stdlib alternatives with structured file:line locations.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{
					"type":        "string",
					"description": "Workspace name",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Max findings per category (default 25)",
				},
			},
		},
	},
	{
		"name":        "garuda.reset_audit_state",
		"description": "Reset smoke-test or agent-proposed audit state in the workspace back to clean baseline.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{"type": "string", "description": "Workspace name"},
				"tenant_id": map[string]interface{}{"type": "string", "description": "Optional tenant UUID"},
			},
		},
	},
	{
		"name":        "garuda.create_workspace",
		"description": "Provision an isolated sandbox workspace (e.g. 'audit-sandbox') so mutating governance or audit checks can run safely without affecting the default workspace.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{"type": "string", "description": "New sandbox workspace name (e.g. 'audit-sandbox')"},
				"name":      map[string]interface{}{"type": "string", "description": "Alias for workspace name"},
				"tenant_id": map[string]interface{}{"type": "string", "description": "Optional tenant UUID"},
			},
		},
	},
	{
		"name":        "garuda.verify_receipt",
		"description": "Recompute receipt_hash, graph_fingerprint, and policy_set_version from the current AST and git commit to cryptographically verify a pre-edit or guard receipt.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace":    map[string]interface{}{"type": "string", "description": "Workspace name"},
				"receipt_hash": map[string]interface{}{"type": "string", "description": "Optional receipt_hash to verify against recomputed digest"},
				"tenant_id":    map[string]interface{}{"type": "string", "description": "Optional tenant UUID"},
			},
		},
	},
	{
		"name":        "garuda.diff_snapshots",
		"description": "Compare the indexed graph snapshot receipt against current HEAD and list active vs superseded policies, commit sync state, and drift deltas.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{"type": "string", "description": "Workspace name"},
				"tenant_id": map[string]interface{}{"type": "string", "description": "Optional tenant UUID"},
			},
		},
	},
	{
		"name":        "garuda.diff_impact",
		"description": "Compute diff-level blast radius in one call: returns per-symbol impact with file, line_start, and source snippets for every symbol touched in the git diff.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{"type": "string", "description": "Workspace name"},
				"repo_dir":  map[string]interface{}{"type": "string", "description": "Optional repository directory"},
				"staged":    map[string]interface{}{"type": "boolean", "description": "Inspect staged changes"},
				"diff":      map[string]interface{}{"type": "string", "description": "Git revision range (e.g. 'HEAD')"},
				"files":     map[string]interface{}{"type": "string", "description": "Optional comma-separated file list"},
				"tenant_id": map[string]interface{}{"type": "string", "description": "Optional tenant UUID"},
			},
		},
	},
	{
		"name":        "garuda.symbol_test_coverage",
		"description": "Show which unit tests (*_test.go) and runtime traces exercise a symbol or package, explaining exported handler verification status.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{"type": "string", "description": "Workspace name"},
				"symbol":    map[string]interface{}{"type": "string", "description": "Target symbol name"},
				"name":      map[string]interface{}{"type": "string", "description": "Alias for symbol name"},
				"package":   map[string]interface{}{"type": "string", "description": "Optional package filter"},
				"tenant_id": map[string]interface{}{"type": "string", "description": "Optional tenant UUID"},
			},
		},
	},
	{
		"name":        "garuda.snapshot",
		"description": "Return the current workspace snapshot ID, commit SHA, dirty-tree status, policy-set version, and freshness verdict (fresh, stale, dirty) in one lightweight call.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{"type": "string", "description": "Workspace name (defaults to default)"},
			},
		},
	},
	{
		"name":        "garuda.explain_claim",
		"description": "Return the compiler AST or policy derivation path and source snippet for a claim_id or symbol name.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{"type": "string", "description": "Workspace name"},
				"claim_id":  map[string]interface{}{"type": "string", "description": "Claim UUID or symbol name to explain"},
			},
		},
	},
	{
		"name":        "garuda.snippet",
		"description": "Return hash-verified source lines around a symbol or file_path:line_start against the active snapshot.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace":  map[string]interface{}{"type": "string", "description": "Workspace name"},
				"entity_id":  map[string]interface{}{"type": "string", "description": "Entity UUID or symbol name"},
				"name":       map[string]interface{}{"type": "string", "description": "Symbol name"},
				"file_path":  map[string]interface{}{"type": "string", "description": "Optional repository-relative file path"},
				"line_start": map[string]interface{}{"type": "number", "description": "Optional 1-indexed line number"},
			},
		},
	},
	{
		"name":        "garuda.summary",
		"description": "Return a compact token-saving workspace summary (counts, active policies, and snapshot ID).",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{"type": "string", "description": "Workspace name"},
				"scope":     map[string]interface{}{"type": "string", "description": "Scope filter (default: production_only)"},
			},
		},
	},
	{
		"name":        "garuda.owners",
		"description": "Return the declared owner and team for an entity from the ownership mapping, or UNASSIGNED if undeclared.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{"type": "string", "description": "Workspace name"},
				"entity_id": map[string]interface{}{"type": "string", "description": "Entity UUID or symbol name"},
				"name":      map[string]interface{}{"type": "string", "description": "Symbol name"},
			},
		},
	},
	{
		"name":        "garuda.verify_claim",
		"description": "Verify a structured edge (from_symbol, predicate, to_symbol) or prose claim against the indexed graph and return SUPPORTED, CONTRADICTED, or NOT_IN_GRAPH with a cryptographic receipt.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace":   map[string]interface{}{"type": "string", "description": "Workspace name"},
				"claim":       map[string]interface{}{"type": "string", "description": "Claim text (e.g. 'RefundHandler CALLS IdempotencyStore'), symbol, or claim_id"},
				"from_symbol": map[string]interface{}{"type": "string", "description": "Structured caller/source symbol name"},
				"predicate":   map[string]interface{}{"type": "string", "description": "Edge predicate (CALLS, IMPORTS, IMPLEMENTS, INHERITS, REFERENCES, WRITES, READS)"},
				"to_symbol":   map[string]interface{}{"type": "string", "description": "Structured callee/target symbol name"},
			},
		},
	},
	{
		"name":        "garuda.test_coverage",
		"description": "Alias for garuda.symbol_test_coverage: inspect which unit tests and runtime traces exercise a symbol.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{"type": "string", "description": "Workspace name"},
				"symbol":    map[string]interface{}{"type": "string", "description": "Symbol name"},
				"package":   map[string]interface{}{"type": "string", "description": "Package filter"},
			},
		},
	},
	{
		"name":        "garuda.audit_export",
		"description": "Export workspace decisions, policy evaluations, and Merkle block anchors as structured JSON or CSV for compliance filing.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace": map[string]interface{}{"type": "string", "description": "Workspace name"},
				"format":    map[string]interface{}{"type": "string", "description": "Export format: 'json' or 'csv' (default 'json')"},
				"tenant_id": map[string]interface{}{"type": "string", "description": "Optional tenant UUID"},
			},
		},
	},
	{
		"name":        "garuda.preflight_bundle",
		"description": "Single-call token-saving pre-edit bundle combining snapshot freshness, talon_pre_edit briefing, test coverage, and declared ownership for a symbol or file.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workspace":        map[string]interface{}{"type": "string", "description": "Workspace name"},
				"target":           map[string]interface{}{"type": "string", "description": "Target symbol or file path"},
				"symbol":           map[string]interface{}{"type": "string", "description": "Alias for target symbol"},
				"include_fixtures": map[string]interface{}{"type": "boolean", "description": "Include test/fixtures scope (default false)"},
				"tenant_id":        map[string]interface{}{"type": "string", "description": "Optional tenant UUID"},
			},
		},
	},
}

var toolNames = func() map[string]bool {
	m := make(map[string]bool, len(allTools))
	for _, t := range allTools {
		if name, ok := t["name"].(string); ok {
			m[name] = true
		}
	}
	return m
}()

// ─────────────────────────────────────────────────────────────────────────────
// Request dispatch
// ─────────────────────────────────────────────────────────────────────────────

func (s *MCPServer) handleRequest(req MCPRequest) MCPResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	case "prompts/list":
		return s.handlePromptsList(req)
	case "prompts/get":
		return s.handlePromptsGet(req)
	case "resources/list":
		return s.handleResourcesList(req)
	case "resources/read":
		return s.handleResourcesRead(req)
	default:
		return s.errorResponse(req.ID, -32601, "Method not found: "+req.Method)
	}
}

func (s *MCPServer) handleInitialize(req MCPRequest) MCPResponse {
	if req.Params != nil {
		if ci, ok := req.Params["clientInfo"].(map[string]interface{}); ok {
			if name, ok := ci["name"].(string); ok && name != "" {
				s.clientName = name
			}
		}
	}

	clientVersion := ""
	if req.Params != nil {
		if ci, ok := req.Params["clientInfo"].(map[string]interface{}); ok {
			if v, ok := ci["version"].(string); ok {
				clientVersion = v
			}
		}
	}
	agentID := os.Getenv("GARUDA_AGENT")
	if agentID == "" {
		agentID = "mcp-agent"
	}
	s.recordSessionStart(context.Background(), s.clientName, clientVersion, agentID)

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2025-06-18",
			"capabilities": map[string]interface{}{
				"tools":     map[string]interface{}{},
				"prompts":   map[string]interface{}{},
				"resources": map[string]interface{}{},
			},
			"instructions": "Garuda Talon Active Enforcement is enabled. Call garuda.talon_pre_edit before modifying code and garuda.talon_verify_edit immediately after editing to verify 0 net-new policy or ADR violations.",
			"serverInfo": map[string]interface{}{
				"name":    "garuda-mcp",
				"version": "1.9.6",
			},
		},
	}
}

func (s *MCPServer) handleToolsList(req MCPRequest) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": allTools,
		},
	}
}

func (s *MCPServer) checkAndConsumeBudget(ctx context.Context, tenantID uuid.UUID, agentID, toolName string, payload interface{}) (func(), error) {
	budgetState, err := s.store.GetTenantBudget(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	estimatedTokens := budget.EstimateTokens("mcp_"+toolName, payload)
	if budgetState.Status == "exhausted" || budgetState.TokenBalance < int64(estimatedTokens) {
		return nil, fmt.Errorf("insufficient budget: need %d, have %d", estimatedTokens, budgetState.TokenBalance)
	}

	commit := func() {
		req := types.BudgetConsumptionRequest{
			AgentID:        agentID,
			TokensUsed:     estimatedTokens,
			ExecutionsUsed: 1,
			Operation:      "mcp_" + toolName,
		}
		_, _ = s.store.ConsumeBudgetDeduct(ctx, tenantID, req)
	}

	return commit, nil
}

func (s *MCPServer) handleToolsCall(req MCPRequest) MCPResponse {
	toolName, ok := req.Params["name"].(string)
	if !ok {
		return s.errorResponse(req.ID, -32602, "Invalid tool name")
	}

	if !toolNames[toolName] {
		return s.errorResponse(req.ID, -32601, "Tool not found: "+toolName)
	}

	args, _ := req.Params["arguments"].(map[string]interface{})
	if args == nil {
		args = map[string]interface{}{}
	}

	agentID, _ := args["agent_id"].(string)
	if agentID == "" {
		agentID = "mcp-agent"
	}

	// Enforce GARUDA_MCP_READ_ONLY at the dispatch gate (BUG-02)
	if ReadOnlyMode() && IsMutatingTool(toolName, args) {
		rej := ReadOnlyRejection(toolName)
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": mustJSON(rej),
					},
				},
			},
		}
	}

	start := time.Now()
	var result interface{}
	var err error

	switch toolName {
	case "garuda.query":
		result, err = s.handleQuery(args)
	case "garuda.briefing":
		result, err = s.handleBriefing(args)
	case "garuda.get_lineage":
		result, err = s.handleGetLineage(args)
	case "garuda.detect_contradictions":
		result, err = s.handleDetectContradictions(args)
	case "garuda.get_impact":
		result, err = s.handleGetImpact(args)
	case "garuda.blast_radius":
		result, err = s.handleBlastRadius(args)
	case "garuda.propose_decision":
		result, err = s.handleProposeDecision(args)
	case "garuda.policy.list":
		result, err = s.handlePolicyList(args)
	case "garuda.governance.status":
		result, err = s.handleGovernanceStatus(args)
	case "garuda.check_drift":
		result, err = s.handleCheckDrift(args)
	case "garuda.query_claims":
		result, err = s.handleQueryClaims(args)
	case "garuda.policy.evaluate":
		result, err = s.handlePolicyEvaluate(args)
	case "garuda.verify_policy_evaluation":
		result, err = s.handleVerifyPolicyEvaluation(args)
	case "garuda.verify_evidence":
		result, err = s.handleVerifyEvidence(args)
	case "garuda.entities":
		result, err = s.handleEntities(args)
	case "garuda.inspect":
		result, err = s.handleInspect(args)
	case "garuda.subclasses":
		result, err = s.handleSubclasses(args)
	case "garuda.implementers":
		result, err = s.handleImplementers(args)
	case "garuda.neighbors":
		result, err = s.handleNeighbors(args)
	case "garuda.create_task":
		result, err = s.handleCreateTask(args)
	case "garuda.handoff":
		result, err = s.handleHandoff(args)
	case "garuda.resume":
		result, err = s.handleResume(args)
	case "garuda.find_entity":
		result, err = s.handleFindEntity(args)
	case "garuda.talon_pre_edit":
		result, err = s.handleTalonPreEdit(args)
	case "garuda.talon_verify_edit":
		result, err = s.handleTalonVerifyEdit(args)
	case "garuda.disambiguate_logic":
		result, err = s.handleDisambiguateLogic(args)
	case "garuda.audit_data_discrepancy", "garuda.audit_financial_discrepancy":
		result, err = s.handleAuditDataDiscrepancy(args)
	case "garuda.synthesize_rulebook":
		result, err = s.handleSynthesizeRulebook(args)
	case "garuda.decide_logic_variant":
		result, err = s.handleDecideLogicVariant(args)
	case "garuda.reconcile_data_audit", "garuda.reconcile_financial_audit":
		result, err = s.handleReconcileDataAudit(args)
	case "garuda.approve_base_rulebook":
		result, err = s.handleApproveBaseRulebook(args)
	case "garuda.criticality_tiers":
		result, err = s.handleCriticalityTiers(args)
	case "garuda.docs_graph":
		result, err = s.handleDocsGraph(args)
	case "garuda.docs_ingest":
		result, err = s.handleDocsIngest(args)
	case "garuda.doctor_diagnose":
		result, err = s.handleDoctorDiagnose(args)
	case "garuda.build_matrix_check":
		result, err = s.handleBuildMatrixCheck(args)
	case "garuda.hygiene_report":
		result, err = s.handleHygieneReport(args)
	case "garuda.reset_audit_state":
		result, err = s.handleResetAuditState(args)
	case "garuda.create_workspace":
		result, err = s.handleCreateWorkspace(args)
	case "garuda.verify_receipt":
		result, err = s.handleVerifyReceipt(args)
	case "garuda.diff_snapshots":
		result, err = s.handleDiffSnapshots(args)
	case "garuda.diff_impact":
		result, err = s.handleDiffImpact(args)
	case "garuda.symbol_test_coverage":
		result, err = s.handleSymbolTestCoverage(args)
	case "garuda.snapshot":
		result, err = s.handleSnapshot(args)
	case "garuda.explain_claim":
		result, err = s.handleExplainClaim(args)
	case "garuda.snippet":
		result, err = s.handleSnippet(args)
	case "garuda.summary":
		result, err = s.handleSummary(args)
	case "garuda.owners":
		result, err = s.handleOwners(args)
	case "garuda.verify_claim":
		result, err = s.handleVerifyClaim(args)
	case "garuda.test_coverage":
		result, err = s.handleSymbolTestCoverage(args)
	case "garuda.audit_export":
		result, err = s.handleAuditExport(args)
	case "garuda.preflight_bundle":
		result, err = s.handlePreflightBundle(args)
	default:
		return s.errorResponse(req.ID, -32601, "Tool not found: "+toolName)
	}

	duration := time.Since(start)

	s.emitToolInvocation(
		context.Background(),
		toolName,
		agentID,
		0,
		duration,
		err == nil,
	)
	s.recordToolCall(context.Background(), toolName, duration, err, args)

	if err != nil {
		if strings.Contains(err.Error(), "resolve workspace") {
			reqWs, _ := args["workspace"].(string)
			avail := []string{"default", "audit-sandbox"}
			if s.store != nil && s.safePool() != nil {
				if rows, qErr := s.safePool().Query(context.Background(), `SELECT name FROM workspaces ORDER BY name ASC LIMIT 25`); qErr == nil {
					var dbWs []string
					for rows.Next() {
						var wn string
						if rows.Scan(&wn) == nil && wn != "" {
							dbWs = append(dbWs, wn)
						}
					}
					rows.Close()
					if len(dbWs) > 0 {
						avail = dbWs
					}
				}
			}
			notFoundPayload := map[string]interface{}{
				"status":               "workspace_not_found",
				"error":                "workspace_not_found",
				"workspace":            reqWs,
				"available_workspaces": avail,
				"reason":               fmt.Sprintf("workspace %q was not found; available workspaces in this tenant: %s", reqWs, strings.Join(avail, ", ")),
			}
			return MCPResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": mustJSON(notFoundPayload),
						},
					},
				},
			}
		}
		return s.errorResponse(req.ID, -32000, err.Error())
	}

	result = s.injectActiveContract(toolName, result)

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": mustJSON(result),
				},
			},
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Tool handlers
// ─────────────────────────────────────────────────────────────────────────────

func (s *MCPServer) handleProposeDecision(args map[string]interface{}) (interface{}, error) {
	if ReadOnlyMode() {
		return ReadOnlyRejection("garuda.propose_decision"), nil
	}
	tenantID, _ := s.resolveTenant(args)
	agentID, _ := args["agent_id"].(string)
	if agentID == "" {
		agentID = "mcp-agent"
	}

	commitBudget, err := s.checkAndConsumeBudget(context.Background(), tenantID, agentID, "propose_decision", args)
	if err != nil {
		return nil, err
	}

	result, err := s.processProposeDecisionLogic(args)
	if err != nil {
		return nil, err
	}

	commitBudget()

	return result, nil
}

func (s *MCPServer) processProposeDecisionLogic(args map[string]interface{}) (interface{}, error) {
	title, ok := args["title"].(string)
	if !ok || title == "" {
		return nil, fmt.Errorf("title is required")
	}
	statement, _ := args["statement"].(string)
	if statement == "" {
		statement, _ = args["rationale"].(string)
	}
	scopeDomain, _ := args["scope_domain"].(string)
	scopeSystem, _ := args["scope_system"].(string)
	tenantID, _ := s.resolveTenant(args)

	if tenantID == uuid.Nil {
		tenantID = tenant.CanonicalID
	}

	owner := s.mcpCallerIdentity()
	now := time.Now().UTC()
	decision := &types.Decision{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Title:       title,
		Statement:   statement,
		Status:      types.StatusDraft,
		Scope:       types.Scope{Domain: scopeDomain, System: scopeSystem},
		ScopeDomain: scopeDomain,
		ScopeSystem: scopeSystem,
		Owner:       owner,
		Confidence:  0.8,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.store.SaveDecision(context.Background(), decision); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":           decision.ID.String(),
		"tenant_id":    decision.TenantID.String(),
		"title":        decision.Title,
		"statement":    decision.Statement,
		"scope_domain": scopeDomain,
		"scope_system": scopeSystem,
		"owner":        owner,
		"status":       decision.Status.String(),
	}, nil
}

var rawSQLGuardRe = regexp.MustCompile(`(?is)^\s*(SELECT|INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|GRANT|REVOKE)\s+|\b(SELECT\s+.+\s+FROM|UNION\s+ALL\s+SELECT|UNION\s+SELECT|DROP\s+TABLE|DELETE\s+FROM|INSERT\s+INTO|UPDATE\s+\w+\s+SET|;\s*(SELECT|DROP|DELETE|INSERT|UPDATE)|--\s*$)`)

// handleQuery implements `garuda.query` (BUG-04 / CLASS-04).
// Explicitly rejects raw SQL statements with {status: "rejected", reason: "..."}
// instead of silently returning {count: 0}.
func (s *MCPServer) handleQuery(args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()
	query, ok := args["query"].(string)
	if !ok || strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("query is required")
	}

	if rawSQLGuardRe.MatchString(query) {
		return map[string]interface{}{
			"status":   "rejected",
			"rejected": true,
			"query":    query,
			"reason":   "garuda.query does not accept raw SQL; use garuda.entities or garuda.find_entity",
			"count":    0,
			"results":  []interface{}{},
		}, nil
	}

	tenantID, _ := s.resolveTenant(args)
	commitBudget, err := s.checkAndConsumeBudget(ctx, tenantID, "mcp-agent", "query", args)
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0)
	qLower := strings.ToLower(strings.TrimSpace(query))

	// 1. Search decisions
	if decisions, dErr := s.store.GetDecisionsByScope(ctx, tenantID, "", ""); dErr == nil {
		for _, d := range decisions {
			if strings.Contains(strings.ToLower(d.Title), qLower) || strings.Contains(strings.ToLower(d.Statement), qLower) {
				results = append(results, map[string]interface{}{
					"kind":       "decision",
					"id":         d.ID.String(),
					"title":      d.Title,
					"statement":  d.Statement,
					"status":     d.Status.String(),
					"scope":      d.Scope,
					"confidence": d.Confidence,
				})
			}
		}
	}

	// 2. Search workspace entities and document_claims
	if _, workspaceID, wErr := s.resolveTenantAndWorkspace(args); wErr == nil && workspaceID != uuid.Nil {
		eRows, eErr := s.safePool().Query(ctx, `
			SELECT id::text, name, kind, COALESCE(package, ''), COALESCE(file_path, ''), COALESCE(line_start, 0)
			FROM entities
			WHERE workspace_id = $1 AND kind != 'external'
			  AND (name ILIKE '%' || $2 || '%' OR COALESCE(package, '') ILIKE '%' || $2 || '%')
			ORDER BY is_exported DESC, name ASC
			LIMIT 20
		`, workspaceID, query)
		if eErr == nil {
			for eRows.Next() {
				var eid, ename, ekind, epkg, efile string
				var eline int
				if eRows.Scan(&eid, &ename, &ekind, &epkg, &efile, &eline) == nil {
					results = append(results, map[string]interface{}{
						"kind":        "entity",
						"id":          eid,
						"name":        ename,
						"entity_kind": ekind,
						"package":     epkg,
						"file":        sanitizeMCPFilePath(efile),
						"line_start":  eline,
					})
				}
			}
			eRows.Close()
		}

		cRows, cErr := s.safePool().Query(ctx, `
			SELECT COALESCE(document_title, 'ADR'), subject, modality, predicate, object, COALESCE(status, 'UNVERIFIED')
			FROM document_claims
			WHERE tenant_id = $1 AND workspace_id = $2
			  AND (subject ILIKE '%' || $3 || '%' OR object ILIKE '%' || $3 || '%' OR COALESCE(document_title, '') ILIKE '%' || $3 || '%')
			LIMIT 20
		`, tenantID, workspaceID, query)
		if cErr == nil {
			for cRows.Next() {
				var doc, subj, mod, pred, obj, st string
				if cRows.Scan(&doc, &subj, &mod, &pred, &obj, &st) == nil {
					results = append(results, map[string]interface{}{
						"kind":      "document_claim",
						"document":  doc,
						"subject":   subj,
						"modality":  mod,
						"predicate": pred,
						"object":    obj,
						"status":    st,
					})
				}
			}
			cRows.Close()
		}
	}

	commitBudget()

	return map[string]interface{}{
		"status":  "ok",
		"query":   query,
		"count":   len(results),
		"results": results,
	}, nil
}

// handleGetLineage implements `garuda.get_lineage` (BUG-06).
// Accepts `decision_id`, `symbol`, `name`, `entity_id`, or `actor`.
func (s *MCPServer) handleGetLineage(args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()
	tenantID, _ := s.resolveTenant(args)
	commitBudget, err := s.checkAndConsumeBudget(ctx, tenantID, "mcp-agent", "get_lineage", args)
	if err != nil {
		return nil, err
	}

	targetStr := ""
	for _, k := range []string{"decision_id", "symbol", "name", "entity_id", "actor"} {
		if v, ok := args[k].(string); ok && strings.TrimSpace(v) != "" {
			targetStr = strings.TrimSpace(v)
			break
		}
	}
	if targetStr == "" {
		return nil, fmt.Errorf("one of decision_id, symbol, name, entity_id, or actor is required")
	}

	if decisionID, pErr := uuid.Parse(targetStr); pErr == nil {
		if decision, dErr := s.store.GetDecision(ctx, tenantID, decisionID); dErr == nil && decision != nil {
			children, _ := s.store.ListDecisionsByParent(ctx, tenantID, decisionID)
			var parent *types.Decision
			if decision.ParentID != nil && *decision.ParentID != uuid.Nil {
				p, _ := s.store.GetDecision(ctx, tenantID, *decision.ParentID)
				parent = p
			}
			commitBudget()
			return map[string]interface{}{
				"decision_id": decisionID.String(),
				"decision":    decision,
				"children":    children,
				"parent":      parent,
			}, nil
		}
	}

	matchedDecisions := make([]map[string]interface{}, 0)
	dRows, dErr := s.safePool().Query(ctx, `
		SELECT id::text, title, COALESCE(statement, ''), COALESCE(owner, ''), COALESCE(status::text, 'draft'), created_at
		FROM decisions
		WHERE tenant_id = $1
		  AND (COALESCE(owner, '') ILIKE '%' || $2 || '%' OR title ILIKE '%' || $2 || '%')
		ORDER BY created_at DESC
		LIMIT 25
	`, tenantID, targetStr)
	if dErr == nil {
		for dRows.Next() {
			var did, dtitle, dstmt, downer, dst string
			var createdAt interface{}
			if dRows.Scan(&did, &dtitle, &dstmt, &downer, &dst, &createdAt) == nil {
				matchedDecisions = append(matchedDecisions, map[string]interface{}{
					"id":         did,
					"title":      dtitle,
					"statement":  dstmt,
					"owner":      downer,
					"status":     dst,
					"created_at": fmt.Sprint(createdAt),
				})
			}
		}
		dRows.Close()
	}

	matchedEvaluations := make([]map[string]interface{}, 0)
	matchedEntities := make([]map[string]interface{}, 0)
	if _, workspaceID, wErr := s.resolveTenantAndWorkspace(args); wErr == nil && workspaceID != uuid.Nil {
		evRows, evErr := s.safePool().Query(ctx, `
			SELECT id::text, policy_id::text, COALESCE(decision, ''), COALESCE(actor, ''),
			       COALESCE(subject_kind, ''), COALESCE(subject_id, ''), COALESCE(merkle_block_height, 0), evaluated_at
			FROM policy_evaluations
			WHERE tenant_id = $1 AND workspace_id = $2
			  AND (COALESCE(actor, '') ILIKE '%' || $3 || '%' OR COALESCE(subject_id, '') ILIKE '%' || $3 || '%')
			ORDER BY evaluated_at DESC
			LIMIT 25
		`, tenantID, workspaceID, targetStr)
		if evErr == nil {
			for evRows.Next() {
				var evid, polid, dec, act, skind, sid string
				var bh int64
				var evAt interface{}
				if evRows.Scan(&evid, &polid, &dec, &act, &skind, &sid, &bh, &evAt) == nil {
					matchedEvaluations = append(matchedEvaluations, map[string]interface{}{
						"evaluation_id":       evid,
						"policy_id":           polid,
						"decision":            dec,
						"actor":               act,
						"subject_kind":        skind,
						"subject_id":          sid,
						"merkle_block_height": bh,
						"evaluated_at":        fmt.Sprint(evAt),
					})
				}
			}
			evRows.Close()
		}

		entRows, entErr := s.safePool().Query(ctx, `
			SELECT id::text, name, kind, COALESCE(package, ''), COALESCE(file_path, ''), COALESCE(line_start, 0)
			FROM entities
			WHERE workspace_id = $1 AND kind != 'external'
			  AND (name ILIKE $2 OR name ILIKE '%.' || $2)
			ORDER BY is_exported DESC, name ASC
			LIMIT 10
		`, workspaceID, targetStr)
		if entErr == nil {
			for entRows.Next() {
				var eid, ename, ekind, epkg, efile string
				var eline int
				if entRows.Scan(&eid, &ename, &ekind, &epkg, &efile, &eline) == nil {
					matchedEntities = append(matchedEntities, map[string]interface{}{
						"id":         eid,
						"name":       ename,
						"kind":       ekind,
						"package":    epkg,
						"file":       sanitizeMCPFilePath(efile),
						"line_start": eline,
					})
				}
			}
			entRows.Close()
		}
	}

	commitBudget()

	return map[string]interface{}{
		"target":             targetStr,
		"decisions_count":    len(matchedDecisions),
		"decisions":          matchedDecisions,
		"evaluations_count":  len(matchedEvaluations),
		"policy_evaluations": matchedEvaluations,
		"entities_count":     len(matchedEntities),
		"entities":           matchedEntities,
		"lineage_count":      len(matchedDecisions) + len(matchedEvaluations) + len(matchedEntities),
	}, nil
}

func (s *MCPServer) handleDetectContradictions(args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()
	tenantID, err := s.resolveTenant(args)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tenant: %w", err)
	}

	commitBudget, err := s.checkAndConsumeBudget(ctx, tenantID, "mcp-agent", "detect_contradictions", args)
	if err != nil {
		return nil, err
	}

	statusFilter, _ := args["status"].(string)
	statusFilter = strings.ToLower(strings.TrimSpace(statusFilter))
	resolvedFlag := statusFilter == "resolved"

	contradictions, err := s.store.ListContradictions(ctx, tenantID, resolvedFlag)
	if err != nil {
		commitBudget()
		return nil, fmt.Errorf("failed to query contradictions: %w", err)
	}

	adrContradictions := make([]map[string]interface{}, 0)
	if _, workspaceID, wErr := s.resolveTenantAndWorkspace(args); wErr == nil && workspaceID != uuid.Nil && !resolvedFlag {
		rows, qErr := s.safePool().Query(ctx, `
			SELECT id::text, COALESCE(document_title, 'ADR'), subject, modality, predicate, object, COALESCE(contradiction_reason, '')
			FROM document_claims
			WHERE tenant_id = $1 AND workspace_id = $2 AND status = 'CONTRADICTED'
		`, tenantID, workspaceID)
		if qErr == nil {
			for rows.Next() {
				var cid, doc, subj, mod, pred, obj, rsn string
				if rows.Scan(&cid, &doc, &subj, &mod, &pred, &obj, &rsn) == nil {
					adrContradictions = append(adrContradictions, map[string]interface{}{
						"id":        cid,
						"kind":      "adr_vs_ast",
						"document":  doc,
						"subject":   subj,
						"modality":  mod,
						"predicate": pred,
						"object":    obj,
						"reason":    rsn,
					})
				}
			}
			rows.Close()
		}
	}

	commitBudget()

	return map[string]interface{}{
		"tenant_id":          tenantID.String(),
		"status_filter":      map[bool]string{true: statusFilter, false: "open"}[statusFilter != ""],
		"count":              len(contradictions) + len(adrContradictions),
		"contradictions":     contradictions,
		"adr_contradictions": adrContradictions,
	}, nil
}

// handleGetImpact implements `garuda.get_impact` (BUG-06 / CLASS-05).
// Accepts `decision_id`, `symbol`, `name`, or `entity_id` and delegates symbol/entity
// lookups directly to handleBlastRadius so B4 and B6 return identical blast-radius results.
func (s *MCPServer) handleGetImpact(args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()
	tenantID, _ := s.resolveTenant(args)
	commitBudget, err := s.checkAndConsumeBudget(ctx, tenantID, "mcp-agent", "get_impact", args)
	if err != nil {
		return nil, err
	}

	targetStr := ""
	for _, k := range []string{"decision_id", "symbol", "name", "entity_id"} {
		if v, ok := args[k].(string); ok && strings.TrimSpace(v) != "" {
			targetStr = strings.TrimSpace(v)
			break
		}
	}
	if targetStr == "" {
		return nil, fmt.Errorf("one of decision_id, symbol, name, or entity_id is required")
	}

	// 1. If targetStr is a valid UUID and matches a Decision row, return decision children impact
	if decisionID, pErr := uuid.Parse(targetStr); pErr == nil {
		if dec, dErr := s.store.GetDecision(ctx, tenantID, decisionID); dErr == nil && dec != nil {
			children, _ := s.store.ListDecisionsByParent(ctx, tenantID, decisionID)
			commitBudget()
			return map[string]interface{}{
				"decision_id":     decisionID.String(),
				"impact_count":    len(children),
				"impact_children": children,
			}, nil
		}
	}

	// 2. Otherwise resolve as an AST entity UUID or symbol name (e.g. "AppendLeafAndSealTx")
	brArgs := map[string]interface{}{
		"tenant_id": tenantID.String(),
	}
	if ws, ok := args["workspace"].(string); ok && ws != "" {
		brArgs["workspace"] = ws
	}
	if _, pErr := uuid.Parse(targetStr); pErr == nil {
		brArgs["entity_id"] = targetStr
	} else {
		brArgs["name"] = targetStr
	}

	brResult, brErr := s.handleBlastRadius(brArgs)
	commitBudget()
	if brErr != nil {
		return nil, brErr
	}
	if m, ok := brResult.(map[string]interface{}); ok {
		m["decision_id"] = targetStr
		m["symbol"] = targetStr
		m["impact_count"] = m["total_affected"]
		return m, nil
	}
	return brResult, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func (s *MCPServer) resolveTenant(args map[string]interface{}) (uuid.UUID, error) {
	if tenantStr, ok := args["tenant_id"].(string); ok && tenantStr != "" {
		return uuid.Parse(tenantStr)
	}
	return tenant.CanonicalID, nil
}

func (s *MCPServer) errorResponse(id interface{}, code int, message string) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}
}

func mustJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return SanitizeGlobalJSONString(string(b))
}
