// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package version

import "fmt"

const (
	// Version is the single canonical release version across all Garuda binaries and docs.
	Version = "v1.9.9"

	// SemVer is the bare semantic version string used in MCP serverInfo and OpenAPI specs.
	SemVer = "1.9.9"

	// ReleaseDate is the canonical release date for v1.9.9.
	ReleaseDate = "2026-11-01"

	// MCPProtocolVersion is the JSON-RPC 2.0 MCP protocol version negotiated on initialize.
	MCPProtocolVersion = "2025-06-18"

	// MCPToolCount is the canonical count of registered tools in bin/garuda-mcp.
	MCPToolCount = 54

	// MCPMutatingToolCount is the canonical count of mutating tools gated by GARUDA_MCP_READ_ONLY.
	MCPMutatingToolCount = 13

	// MCPReadOnlyToolCount is the canonical count of read-only inspection tools.
	MCPReadOnlyToolCount = 41

	// DashboardSurfaceCount is the canonical count of working surfaces in /dashboard.
	DashboardSurfaceCount = 10

	// DashboardEndpointCount is the canonical count of /api/v1/dashboard/* endpoints.
	DashboardEndpointCount = 19

	// MigrationCount is the canonical count of idempotent PostgreSQL migrations (001..095).
	MigrationCount = 95

	// PolicyPredicateCount is the canonical count of graph-native YAML policy predicates.
	PolicyPredicateCount = 9
)

// MCPHelpBanner returns the canonical --help header for bin/garuda-mcp.
func MCPHelpBanner() string {
	return fmt.Sprintf(
		"Garuda MCP Server %s (JSON-RPC 2.0 over stdio — %d governance, audit & AST tools: %d read-only, %d mutating)",
		Version, MCPToolCount, MCPReadOnlyToolCount, MCPMutatingToolCount,
	)
}

// MCPVersionBanner returns the canonical --version string for bin/garuda-mcp.
func MCPVersionBanner() string {
	return fmt.Sprintf("Garuda MCP Server %s", Version)
}
