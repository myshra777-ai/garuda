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
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/budget"
	"github.com/myshra777-ai/garuda/internal/engine"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/types"
)

// MCPServer holds the process-scoped state of one MCP session.
//
// A session begins when the host (Claude Desktop, Cursor, or any
// other MCP client) spawns this process and ends when the process
// exits. The session ID is generated once at startup and is included
// in every log line and, in future commits, in every telemetry event.
//
// Store, engines, and auth service are all constructed once. The
// stdio loop can handle hundreds of tool calls without reconnecting
// to the database.
type MCPServer struct {
	store               *store.PostgresStore
	lineageEngine       *engine.LineageEngine
	contradictionEngine *engine.ContradictionEngine
	authService         *auth.AuthService
	jwtConfig           *auth.JWTConfig

	// sessionID is generated at startup and does not change for the
	// lifetime of the process.
	sessionID string
	startedAt time.Time
}

func main() {
	// All logging goes to stderr. stdout is the JSON-RPC channel and
	// must contain nothing but line-delimited MCP messages.
	slog.Info("Garuda MCP Server starting")

	enc := json.NewEncoder(os.Stdout)

	// DATABASE_URL is required. There is no fallback: a missing value
	// in production should fail loudly, not connect to a local test
	// database.
	dbURL := os.Getenv("DATABASE_URL")
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
	}

	slog.Info("Garuda MCP Server ready",
		"session_id", server.sessionID,
		"pid", os.Getpid(),
	)

	// Graceful shutdown on SIGINT/SIGTERM.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		slog.Info("Received shutdown signal", "session_id", server.sessionID)
		os.Exit(0)
	}()

	// Line-delimited JSON-RPC over stdio. The loop handles one request
	// at a time and exits when stdin closes (EOF), which the MCP host
	// signals by closing the pipe.
	//
	// A parse error does not terminate the process. The host may send
	// another well-formed request; we respond with a JSON-RPC parse
	// error and continue reading.
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
				return
			}
			// Malformed input. Respond with a parse error and keep the
			// loop alive; the host may recover.
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
	default:
		return s.errorResponse(req.ID, -32601, "Method not found: "+req.Method)
	}
}

func (s *MCPServer) handleInitialize(req MCPRequest) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "0.1.0",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "garuda-mcp",
				"version": "1.0.0",
			},
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Tool listing
// ─────────────────────────────────────────────────────────────────────────────

func (s *MCPServer) handleToolsList(req MCPRequest) MCPResponse {
	tools := []map[string]interface{}{
		{
			"name":        "garuda.query",
			"description": "Query the Garuda knowledge graph with natural language",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query":     map[string]interface{}{"type": "string", "description": "The natural language query"},
					"tenant_id": map[string]interface{}{"type": "string", "description": "Optional tenant ID"},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "garuda.get_lineage",
			"description": "Get the full lineage of a decision",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"decision_id": map[string]interface{}{"type": "string", "description": "The decision UUID"},
					"tenant_id":   map[string]interface{}{"type": "string", "description": "Tenant ID"},
				},
				"required": []string{"decision_id"},
			},
		},
		{
			"name":        "garuda.detect_contradictions",
			"description": "Detect unresolved contradictions in the tenant knowledge graph",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"tenant_id": map[string]interface{}{"type": "string", "description": "Tenant ID"},
					"scope":     map[string]interface{}{"type": "string", "description": "Optional scope filter"},
				},
			},
		},
		{
			"name":        "garuda.get_impact",
			"description": "Find what breaks if a decision is changed",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"decision_id": map[string]interface{}{"type": "string", "description": "The decision UUID"},
					"tenant_id":   map[string]interface{}{"type": "string", "description": "Tenant ID"},
				},
				"required": []string{"decision_id"},
			},
		},
		{
			"name":        "garuda.propose_decision",
			"description": "Propose a new decision (creates a draft)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title":        map[string]interface{}{"type": "string", "description": "Decision title"},
					"scope_domain": map[string]interface{}{"type": "string", "description": "Scope domain"},
					"scope_system": map[string]interface{}{"type": "string", "description": "Scope system"},
					"tenant_id":    map[string]interface{}{"type": "string", "description": "Tenant ID"},
				},
				"required": []string{"title"},
			},
		},
		// ── Governance tools (Phase 2.2a) ──
		{
			"name":        "garuda.policy.list",
			"description": "List policies registered for a tenant. Read-only.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"tenant_id": map[string]interface{}{"type": "string", "description": "Tenant UUID; falls back to GARUDA_TENANT_ID env or default tenant"},
				},
			},
		},
		{
			"name":        "garuda.governance.status",
			"description": "Aggregate governance state: active policies, documentation health, contradiction count. Read-only.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"tenant_id": map[string]interface{}{"type": "string", "description": "Tenant UUID"},
					"workspace": map[string]interface{}{"type": "string", "description": "Workspace name; falls back to GARUDA_WORKSPACE env or 'default'"},
				},
			},
		},
		{
			"name":        "garuda.check_drift",
			"description": "Documentation-to-code drift report. Read-only.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"tenant_id": map[string]interface{}{"type": "string", "description": "Tenant UUID"},
					"workspace": map[string]interface{}{"type": "string", "description": "Workspace name"},
				},
			},
		},
		{
			"name":        "garuda.query_claims",
			"description": "Query document claims matching a subject. Read-only.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"tenant_id": map[string]interface{}{"type": "string", "description": "Tenant UUID"},
					"workspace": map[string]interface{}{"type": "string", "description": "Workspace name"},
					"subject":   map[string]interface{}{"type": "string", "description": "Filter by subject or object (substring)"},
				},
			},
		},
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Budget enforcement
// ─────────────────────────────────────────────────────────────────────────────

// checkAndConsumeBudget performs a pre-flight budget check and returns
// a commit closure. The caller invokes commit only after business
// logic has succeeded, so failed operations consume no tokens.
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

// ─────────────────────────────────────────────────────────────────────────────
// Tool dispatch
// ─────────────────────────────────────────────────────────────────────────────

func (s *MCPServer) handleToolsCall(req MCPRequest) MCPResponse {
	toolName, ok := req.Params["name"].(string)
	if !ok {
		return s.errorResponse(req.ID, -32602, "Invalid tool name")
	}

	args, _ := req.Params["arguments"].(map[string]interface{})

	var result interface{}
	var err error

	// Handlers control their own pre-flight and post-commit budget lifecycles.
	switch toolName {
	case "garuda.query":
		result, err = s.handleQuery(args)
	case "garuda.get_lineage":
		result, err = s.handleGetLineage(args)
	case "garuda.detect_contradictions":
		result, err = s.handleDetectContradictions(args)
	case "garuda.get_impact":
		result, err = s.handleGetImpact(args)
	case "garuda.propose_decision":
		result, err = s.handleProposeDecision(args)

	// Governance tools (Phase 2.2a). All read-only. None anchor to
	// the Merkle ledger — only the decision proposals do.
	case "garuda.policy.list":
		result, err = s.handlePolicyList(args)
	case "garuda.governance.status":
		result, err = s.handleGovernanceStatus(args)
	case "garuda.check_drift":
		result, err = s.handleCheckDrift(args)
	case "garuda.query_claims":
		result, err = s.handleQueryClaims(args)

	default:
		return s.errorResponse(req.ID, -32601, "Tool not found: "+toolName)
	}

	if err != nil {
		return s.errorResponse(req.ID, -32000, err.Error())
	}

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
	scopeDomain, _ := args["scope_domain"].(string)
	scopeSystem, _ := args["scope_system"].(string)
	tenantID, _ := s.resolveTenant(args)

	if tenantID == uuid.Nil {
		tenantID = uuid.NewSHA1(uuid.NameSpaceOID, []byte("default-tenant"))
	}

	now := time.Now().UTC()
	decision := &types.Decision{
		ID:         uuid.New(),
		TenantID:   tenantID,
		Title:      title,
		Status:     types.StatusDraft,
		Scope:      types.Scope{Domain: scopeDomain, System: scopeSystem},
		Owner:      "mcp-agent",
		Confidence: 0.8,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.store.SaveDecision(context.Background(), decision); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":        decision.ID.String(),
		"tenant_id": decision.TenantID.String(),
		"title":     decision.Title,
		"status":    decision.Status.String(),
	}, nil
}

func (s *MCPServer) handleQuery(args map[string]interface{}) (interface{}, error) {
	tenantID, _ := s.resolveTenant(args)
	commitBudget, err := s.checkAndConsumeBudget(context.Background(), tenantID, "mcp-agent", "query", args)
	if err != nil {
		return nil, err
	}

	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query is required")
	}

	decisions, err := s.store.GetDecisionsByScope(context.Background(), tenantID, "", "")
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0)
	for _, d := range decisions {
		if strings.Contains(strings.ToLower(d.Title), strings.ToLower(query)) {
			results = append(results, map[string]interface{}{
				"id":         d.ID.String(),
				"title":      d.Title,
				"status":     d.Status.String(),
				"scope":      d.Scope,
				"confidence": d.Confidence,
			})
		}
	}

	commitBudget()

	return map[string]interface{}{
		"query":   query,
		"count":   len(results),
		"results": results,
	}, nil
}

func (s *MCPServer) handleGetLineage(args map[string]interface{}) (interface{}, error) {
	tenantID, _ := s.resolveTenant(args)
	commitBudget, err := s.checkAndConsumeBudget(context.Background(), tenantID, "mcp-agent", "get_lineage", args)
	if err != nil {
		return nil, err
	}

	decisionIDStr, ok := args["decision_id"].(string)
	if !ok || decisionIDStr == "" {
		return nil, fmt.Errorf("decision_id is required")
	}
	decisionID, err := uuid.Parse(decisionIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid decision_id")
	}

	decision, err := s.store.GetDecision(context.Background(), tenantID, decisionID)
	if err != nil {
		return nil, err
	}

	children, _ := s.store.ListDecisionsByParent(context.Background(), tenantID, decisionID)

	var parent *types.Decision
	if decision.ParentID != nil && *decision.ParentID != uuid.Nil {
		p, _ := s.store.GetDecision(context.Background(), tenantID, *decision.ParentID)
		parent = p
	}

	commitBudget()

	return map[string]interface{}{
		"decision": decision,
		"children": children,
		"parent":   parent,
	}, nil
}

func (s *MCPServer) handleDetectContradictions(args map[string]interface{}) (interface{}, error) {
	tenantID, err := s.resolveTenant(args)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tenant: %w", err)
	}

	commitBudget, err := s.checkAndConsumeBudget(context.Background(), tenantID, "mcp-agent", "detect_contradictions", args)
	if err != nil {
		return nil, err
	}

	contradictions, err := s.store.ListContradictions(context.Background(), tenantID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to query contradictions: %w", err)
	}

	commitBudget()

	return map[string]interface{}{
		"tenant_id":      tenantID.String(),
		"count":          len(contradictions),
		"contradictions": contradictions,
	}, nil
}

func (s *MCPServer) handleGetImpact(args map[string]interface{}) (interface{}, error) {
	tenantID, _ := s.resolveTenant(args)
	commitBudget, err := s.checkAndConsumeBudget(context.Background(), tenantID, "mcp-agent", "get_impact", args)
	if err != nil {
		return nil, err
	}

	decisionIDStr, ok := args["decision_id"].(string)
	if !ok || decisionIDStr == "" {
		return nil, fmt.Errorf("decision_id is required")
	}
	decisionID, err := uuid.Parse(decisionIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid decision_id")
	}

	children, _ := s.store.ListDecisionsByParent(context.Background(), tenantID, decisionID)

	commitBudget()

	return map[string]interface{}{
		"decision_id":     decisionID.String(),
		"impact_count":    len(children),
		"impact_children": children,
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func (s *MCPServer) resolveTenant(args map[string]interface{}) (uuid.UUID, error) {
	if tenantStr, ok := args["tenant_id"].(string); ok && tenantStr != "" {
		return uuid.Parse(tenantStr)
	}
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("default-tenant")), nil
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
	return string(b)
}
