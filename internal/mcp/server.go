// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Server represents the Garuda Model Context Protocol server instance.
type Server struct {
	pool      *pgxpool.Pool
	tenantID  uuid.UUID
	workspace uuid.UUID
}

// NewServer initializes a new MCP server with pool, tenantID, and workspace.
func NewServer(pool *pgxpool.Pool, tenantID uuid.UUID, workspace uuid.UUID) *Server {
	return &Server{
		pool:      pool,
		tenantID:  tenantID,
		workspace: workspace,
	}
}

// ExecuteTool routes an incoming tool execution request and returns a JSON string result.
func (s *Server) ExecuteTool(ctx context.Context, toolName string, args any) (string, error) {
	var rawArgs json.RawMessage
	if args != nil {
		switch v := args.(type) {
		case json.RawMessage:
			rawArgs = v
		case []byte:
			rawArgs = v
		case string:
			rawArgs = json.RawMessage(v)
		default:
			b, err := json.Marshal(args)
			if err != nil {
				return "", err
			}
			rawArgs = b
		}
	}

	res, err := DispatchTool(ctx, s.pool, toolName, rawArgs)
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// DispatchTool routes an incoming MCP tool execution request to the appropriate Garuda handler.
func DispatchTool(ctx context.Context, pool *pgxpool.Pool, toolName string, arguments json.RawMessage) (any, error) {
	switch toolName {
	case "garuda_check_drift":
		return HandleCheckDrift(ctx, pool, arguments)
	case "garuda_query_claims":
		return HandleQueryClaims(ctx, pool, arguments)
	default:
		return nil, fmt.Errorf("unknown mcp tool: %s", toolName)
	}
}
