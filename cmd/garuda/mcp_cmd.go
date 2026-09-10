// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Model Context Protocol (MCP) server endpoints for AI coding agents",
}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Garuda MCP server over standard input/output streams",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			dbURL = "postgres://garuda:garudapassword@localhost:5432/garuda?sslmode=disable"
		}

		ctx := context.Background()
		pool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("database connection failed: %w", err)
		}
		defer pool.Close()

		fmt.Fprintln(os.Stderr, "🤖 Garuda MCP Intelligence Server active. Listening for agent queries...")

		// Simple JSON-RPC read-eval-print loop over stdin/stdout for MCP stdio transport
		decoder := json.NewDecoder(os.Stdin)
		encoder := json.NewEncoder(os.Stdout)

		for {
			var req struct {
				JSONRPC string          `json:"jsonrpc"`
				ID      any             `json:"id"`
				Method  string          `json:"method"`
				Params  json.RawMessage `json:"params"`
			}

			if err := decoder.Decode(&req); err != nil {
				break
			}

			var resp struct {
				JSONRPC string `json:"jsonrpc"`
				ID      any    `json:"id"`
				Result  any    `json:"result,omitempty"`
				Error   any    `json:"error,omitempty"`
			}
			resp.JSONRPC = "2.0"
			resp.ID = req.ID

			switch req.Method {
			case "tools/list":
				resp.Result = map[string]any{
					"tools": []map[string]any{
						{
							"name":        "garuda_check_drift",
							"description": "Evaluate workspace architecture constraints, documentation health, and return gate verdicts.",
							"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
						},
						{
							"name":        "garuda_query_claims",
							"description": "Query specific architectural rules and constraints for a code symbol.",
							"inputSchema": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"subject": map[string]any{"type": "string", "description": "Symbol name e.g. RefundHandler"},
								},
							},
						},
					},
				}
			case "tools/call":
				var callParams struct {
					Name      string          `json:"name"`
					Arguments json.RawMessage `json:"arguments"`
				}
				if err := json.Unmarshal(req.Params, &callParams); err != nil {
					resp.Error = map[string]any{"code": -32602, "message": err.Error()}
					break
				}

				res, err := mcp.DispatchTool(ctx, pool, callParams.Name, callParams.Arguments)
				if err != nil {
					resp.Error = map[string]any{"code": -32000, "message": err.Error()}
				} else {
					resp.Result = map[string]any{
						"content": []map[string]any{
							{
								"type": "text",
								"text": formatToolResult(res),
							},
						},
					}
				}
			default:
				resp.Error = map[string]any{"code": -32601, "message": fmt.Sprintf("method not found: %s", req.Method)}
			}

			_ = encoder.Encode(&resp)
		}

		return nil
	},
}

func formatToolResult(res any) string {
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", res)
	}
	stringVal := string(b)
	return stringVal
}

func init() {
	mcpCmd.AddCommand(mcpServeCmd)
	rootCmd.AddCommand(mcpCmd)
}
