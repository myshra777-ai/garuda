// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Model Context Protocol (MCP) — see bin/garuda-mcp for the server binary",
	Long: `Garuda exposes its semantic graph over the Model Context Protocol (MCP).

The MCP server is a separate binary: bin/garuda-mcp. It is the same
process every MCP client (Cursor, Claude Desktop, Codex, GitHub
Copilot) launches over stdio.

The previous implementation exposed an MCP surface through this CLI
subcommand. That implementation was incomplete — it did not respond to
the "initialize" method that all MCP clients use as their handshake —
and has been removed. Any config written by an older version of
"garuda init" should be replaced; re-run "garuda init" to update it.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		self, _ := os.Executable()
		fmt.Fprintf(os.Stderr, "The MCP server is a separate binary.\n")
		fmt.Fprintf(os.Stderr, "Launch it directly: bin/garuda-mcp\n")
		fmt.Fprintf(os.Stderr, "(You are running: %s)\n", self)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
