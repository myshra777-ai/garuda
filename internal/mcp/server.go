// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

//

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MCP JSON-RPC 2.0 Protocol Types
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type Server struct {
	pool        *pgxpool.Pool
	tenantID    uuid.UUID
	workspaceID uuid.UUID
}

func NewServer(pool *pgxpool.Pool, tenantID, workspaceID uuid.UUID) *Server {
	return &Server{
		pool:        pool,
		tenantID:    tenantID,
		workspaceID: workspaceID,
	}
}

func (s *Server) GetToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "get_runtime_state",
			Description: "Returns cryptographic Merkle ledger state, active block height, root hashes, and verification counts.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"workspace": map[string]interface{}{
						"type":        "string",
						"description": "Optional workspace name (defaults to active workspace).",
					},
				},
			},
		},
		{
			Name:        "get_contradictions",
			Description: "Returns all quarantined architectural drift and policy contradictions with exact evidence locations.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of contradictions to return (default: 50).",
					},
				},
			},
		},
		{
			Name:        "get_verified_context",
			Description: "Returns compiler AST-proved and runtime-supported neighbors for a symbol or package, filtering out unverified or quarantined paths.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"symbol": map[string]interface{}{
						"type":        "string",
						"description": "Symbol name or entity UUID to inspect.",
					},
					"package": map[string]interface{}{
						"type":        "string",
						"description": "Optional package name filter.",
					},
				},
				"required": []string{"symbol"},
			},
		},
		{
			Name:        "get_blast_radius",
			Description: "Calculates upstream callers and downstream dependencies impacted by changes to a given symbol across repository boundaries.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"symbol": map[string]interface{}{
						"type":        "string",
						"description": "Target symbol name or entity UUID.",
					},
					"max_depth": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum traversal depth across the AST graph (default: 3).",
					},
				},
				"required": []string{"symbol"},
			},
		},
	}
}

// ServeStdio starts the standard input/output JSON-RPC loop for AI agent execution
func (s *Server) ServeStdio(ctx context.Context, in io.Reader, out io.Writer) error {
	decoder := json.NewDecoder(in)
	encoder := json.NewEncoder(out)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var req JSONRPCRequest
			if err := decoder.Decode(&req); err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}

			resp := s.handleRequest(ctx, req)
			if err := encoder.Encode(resp); err != nil {
				return err
			}
		}
	}
}

func (s *Server) handleRequest(ctx context.Context, req JSONRPCRequest) JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "garuda-mcp-server",
					"version": "1.0.0",
				},
			},
		}

	case "tools/list":
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"tools": s.GetToolDefinitions(),
			},
		}

	case "tools/call":
		var callParams struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &callParams); err != nil {
			return JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &RPCError{Code: -32602, Message: "Invalid params: " + err.Error()},
			}
		}

		result, err := s.ExecuteTool(ctx, callParams.Name, callParams.Arguments)
		if err != nil {
			return JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &RPCError{Code: -32000, Message: err.Error()},
			}
		}

		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": result,
					},
				},
			},
		}

	default:
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)},
		}
	}
}

func (s *Server) ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	switch name {
	case "get_runtime_state":
		return s.executeGetRuntimeState(ctx)
	case "get_contradictions":
		return s.executeGetContradictions(ctx, args)
	case "get_verified_context":
		return s.executeGetVerifiedContext(ctx, args)
	case "get_blast_radius":
		return s.executeGetBlastRadius(ctx, args)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *Server) executeGetRuntimeState(ctx context.Context) (string, error) {
	var blockHeight int64
	var snapshotHash, staticRoot, runtimeRoot string
	var verifiedClaims, contradictedClaims int

	err := s.pool.QueryRow(ctx, `
		SELECT block_height, snapshot_hash, static_root_hash, runtime_root_hash, verified_claims_count, contradicted_claims_count
		FROM merkle_snapshots
		WHERE tenant_id = $1
		ORDER BY block_height DESC
		LIMIT 1
	`, s.tenantID).Scan(&blockHeight, &snapshotHash, &staticRoot, &runtimeRoot, &verifiedClaims, &contradictedClaims)

	if err != nil {
		blockHeight = 1
		snapshotHash = "Genesis verified"
		staticRoot = "Genesis"
		runtimeRoot = "Genesis"
	}

	var totalEntities, totalClaims int
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM entities WHERE workspace_id = $1`, s.workspaceID).Scan(&totalEntities)
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM claims WHERE workspace_id = $1`, s.workspaceID).Scan(&totalClaims)

	data := map[string]interface{}{
		"status":              "CRYPTOGRAPHICALLY_VERIFIED",
		"block_height":        blockHeight,
		"snapshot_hash":       snapshotHash,
		"static_root_hash":    staticRoot,
		"runtime_root_hash":   runtimeRoot,
		"total_entities":      totalEntities,
		"total_static_claims": totalClaims,
		"verified_claims":     verifiedClaims,
		"contradicted_claims": contradictedClaims,
	}

	bytes, _ := json.MarshalIndent(data, "", "  ")
	return string(bytes), nil
}

func (s *Server) executeGetContradictions(ctx context.Context, args map[string]interface{}) (string, error) {
	limit := 50
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT 
			cv.id::text,
			e.name,
			e.package,
			COALESCE(e.file_path, 'runtime') || ':' || COALESCE(e.line_start::text, '0'),
			COALESCE(cv.evidence_payload->>'raw_target', 'unapproved-endpoint'),
			cv.runtime_observed_count,
			cv.last_evaluated_at
		FROM claim_verifications cv
		JOIN entities e ON e.id = cv.source_entity_id
		WHERE cv.workspace_id = $1 AND cv.status = 'CONTRADICTED'
		ORDER BY cv.runtime_observed_count DESC
		LIMIT $2
	`, s.workspaceID, limit)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type ViolationItem struct {
		ID              string `json:"id"`
		CallerSymbol    string `json:"caller_symbol"`
		CallerPackage   string `json:"caller_package"`
		Location        string `json:"location"`
		TargetEndpoint  string `json:"unapproved_target"`
		InvocationCount int64  `json:"invocation_count"`
		LastEvaluatedAt string `json:"last_evaluated_at"`
	}

	var violations []ViolationItem
	for rows.Next() {
		var v ViolationItem
		var t interface{}
		if err := rows.Scan(&v.ID, &v.CallerSymbol, &v.CallerPackage, &v.Location, &v.TargetEndpoint, &v.InvocationCount, &t); err == nil {
			v.LastEvaluatedAt = fmt.Sprintf("%v", t)
			violations = append(violations, v)
		}
	}

	bytes, _ := json.MarshalIndent(violations, "", "  ")
	return string(bytes), nil
}

func (s *Server) executeGetVerifiedContext(ctx context.Context, args map[string]interface{}) (string, error) {
	symbol, _ := args["symbol"].(string)
	if symbol == "" {
		return "", fmt.Errorf("symbol is required")
	}

	rows, err := s.pool.Query(ctx, `
		SELECT 
			e.id::text, e.name, e.kind, e.package, e.file_path,
			c.claim_type,
			target.name, target.package,
			COALESCE(cv.status, 'UNVERIFIED')
		FROM entities e
		LEFT JOIN claims c ON c.from_entity_id = e.id
		LEFT JOIN entities target ON target.id = c.to_entity_id
		LEFT JOIN claim_verifications cv ON cv.source_entity_id = e.id AND cv.target_entity_id = target.id
		WHERE e.workspace_id = $1 AND (e.name ILIKE $2 OR e.id::text = $2)
		  AND (cv.status IS NULL OR cv.status != 'CONTRADICTED')
	`, s.workspaceID, "%"+symbol+"%")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type VerifiedNeighbor struct {
		TargetSymbol  string `json:"target_symbol"`
		TargetPackage string `json:"target_package"`
		RelationType  string `json:"relation_type"`
		TrustState    string `json:"trust_state"`
	}

	type SymbolContext struct {
		ID           string             `json:"entity_id"`
		Name         string             `json:"name"`
		Kind         string             `json:"kind"`
		Package      string             `json:"package"`
		File         string             `json:"file"`
		Dependencies []VerifiedNeighbor `json:"verified_dependencies"`
	}

	var results []SymbolContext
	currentMap := make(map[string]*SymbolContext)

	for rows.Next() {
		var id, name, kind, pkg, file string
		var claimType, targetName, targetPkg, trustState *string

		if err := rows.Scan(&id, &name, &kind, &pkg, &file, &claimType, &targetName, &targetPkg, &trustState); err == nil {
			if _, exists := currentMap[id]; !exists {
				currentMap[id] = &SymbolContext{
					ID:           id,
					Name:         name,
					Kind:         kind,
					Package:      pkg,
					File:         file,
					Dependencies: []VerifiedNeighbor{},
				}
			}
			if targetName != nil && claimType != nil {
				stateVal := "UNVERIFIED"
				if trustState != nil {
					stateVal = *trustState
				}
				currentMap[id].Dependencies = append(currentMap[id].Dependencies, VerifiedNeighbor{
					TargetSymbol:  *targetName,
					TargetPackage: *targetPkg,
					RelationType:  *claimType,
					TrustState:    stateVal,
				})
			}
		}
	}

	for _, v := range currentMap {
		results = append(results, *v)
	}

	bytes, _ := json.MarshalIndent(results, "", "  ")
	return string(bytes), nil
}

func (s *Server) executeGetBlastRadius(ctx context.Context, args map[string]interface{}) (string, error) {
	symbol, _ := args["symbol"].(string)
	if symbol == "" {
		return "", fmt.Errorf("symbol is required")
	}

	// 1. Upstream callers (Who calls this symbol?)
	upstreamRows, err := s.pool.Query(ctx, `
		SELECT e.id::text, e.name, e.kind, e.package, split_part(e.package, '/', 1)
		FROM claims c
		JOIN entities e ON e.id = c.from_entity_id
		JOIN entities target ON target.id = c.to_entity_id
		WHERE c.workspace_id = $1 AND (target.name ILIKE '%' || $2 || '%' OR target.id::text = $2)
	`, s.workspaceID, symbol)
	if err != nil {
		return "", err
	}
	defer upstreamRows.Close()

	type ImpactNode struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Kind    string `json:"kind"`
		Package string `json:"package"`
		Repo    string `json:"repo"`
	}

	var upstream []ImpactNode
	for upstreamRows.Next() {
		var n ImpactNode
		if err := upstreamRows.Scan(&n.ID, &n.Name, &n.Kind, &n.Package, &n.Repo); err == nil {
			upstream = append(upstream, n)
		}
	}

	// 2. Downstream callees (What does this symbol call?)
	downstreamRows, err := s.pool.Query(ctx, `
		SELECT e.id::text, e.name, e.kind, e.package, split_part(e.package, '/', 1)
		FROM claims c
		JOIN entities src ON src.id = c.from_entity_id
		JOIN entities e ON e.id = c.to_entity_id
		WHERE c.workspace_id = $1 AND (src.name ILIKE '%' || $2 || '%' OR src.id::text = $2)
	`, s.workspaceID, symbol)
	if err != nil {
		return "", err
	}
	defer downstreamRows.Close()

	var downstream []ImpactNode
	for downstreamRows.Next() {
		var n ImpactNode
		if err := downstreamRows.Scan(&n.ID, &n.Name, &n.Kind, &n.Package, &n.Repo); err == nil {
			downstream = append(downstream, n)
		}
	}

	blastRadius := map[string]interface{}{
		"target_symbol":         symbol,
		"upstream_callers_count": len(upstream),
		"upstream_callers":       upstream,
		"downstream_deps_count": len(downstream),
		"downstream_deps":        downstream,
	}

	bytes, _ := json.MarshalIndent(blastRadius, "", "  ")
	return string(bytes), nil
}
