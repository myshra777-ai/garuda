# Changelog

All notable changes to Garuda will be documented in this file.

## [Unreleased]

### Added
- **Hygiene & Advisory CLI**: Introduced `garuda hygiene` command with typed advisory analysis core, contract hardening, and `ponytail` compatibility alias.
- **MCP Integration & Server**:
  - Consolidated MCP into a long-lived stdio server binary.
  - Added MCP tools: `garuda.briefing`, `garuda.handoff`, `garuda.resume`, `garuda.blast_radius`, `garuda.verify_policy_evaluation`, and graph structure exploration tools.
  - Added session tracking and telemetry per tool invocation in `mcp_sessions` and `mcp_tool_calls`.
- **Control Plane & Dashboard**:
  - Multi-tab Control Plane dashboard (Operations, Tenants, Decisions, Agents, Business metrics).
  - Added read-only DB role and background telemetry refresh jobs (`telemetry_aggregates`).
  - UI improvements including embedded static CSS/JS assets, workspace picker on sidebar, detail drawers, and revision chain tracking for decisions.
- **Multi-Tenancy & Workspace Scope**:
  - Schema additions for tenant memberships and user signup flow (`/signup`).
  - Session-scoped and workspace-scoped data isolation across storage, HTTP API, and CLI.
- **Multi-Language Parsing**:
  - Python AST parsing support and unified polyglot knowledge graph construction.
  - TypeScript parser support via Tree-sitter.
- **Merkle Verification System**:
  - RFC 6962 Merkle tree structure with golden test vectors.
  - Canonical evaluation encoders, epoch leaves schema, and epoch root hash writes.
- **Benchmarking**:
  - Expanded benchmark corpus to 20 fixtures (GAP-20) with verified deterministic output.

### Changed
- **CLI Flags & Command Usage**:
  - Renamed `--fail-on BLOCK` flag to `--fail-on-block` across CLI, playbooks, and capabilities specification.
  - Policy evaluations now require `--save` flag to persist execution writes.
  - Policy directory reconciliation made explicitly opt-in via `--reconcile`.
- **Architecture & Refactoring**:
  - Decoupled CLI handlers from direct SQL access in favor of a typed storage layer.
  - Modularized `cmd/garuda` into dedicated command handlers.
  - Updated Go toolchain to Go 1.25.13 and updated `tree-sitter` dependency.
  - Scoped repository unique constraints per workspace.
  - Replaced package inference heuristics (`inferRepositoryFromPackage`) with formal repository foreign keys.

### Fixed
- **Storage & Tenant Leakage**:
  - Fixed cross-tenant and cross-workspace leaks across knowledge readers, inspectors, stats queries, and evaluators.
  - Resolved race condition during concurrent workspace creation with identical names.
  - Guaranteed `workspace_id NOT NULL` constraints on core entity and claim tables.
- **API & Middleware**:
  - Corrected API middleware ordering so `WithRequestID` runs prior to `WithLogging`.
  - Unified rate limiting around `IPRateLimiter`.
  - Corrected surface error responses for failed handoffs to HTTP clients.
- **Graph & Verifier**:
  - Attributed `CALLS` graph edges to true callers and restored edge reading directly from claims.
  - Preserved entity rows during semantic scans.
  - Properly released MCP budget reservations on error paths during contradiction listing.

### Removed
- Removed legacy `cas` (Content Addressable Storage) package and `IngestBlocks` routines.
- Removed deprecated audit-store package, functions, and corresponding HTTP endpoints.
- Removed dead AST polyglot package (`internal/ast/polyglot.go`).
- Purged orphan workspaces, test tenant residual data, and obsolete deployment config files (`fly.toml`, `render.yaml`).

### Semantic Changes
- **Quarantine Store for Runtime Contradictions**: Mismatches between runtime observed claims and static code expectations (`runtime_vs_static`) are now explicitly isolated in the quarantine decision table.
- **Merkle Tree & Epoch Auditing**: Implemented epoch-based Merkle tree leaves and tenant-scoped genesis records enforcing cryptographic verifiability of evaluation outputs.
- **Database Schema Migrations**: Introduced schema migrations (`082`-`086`) enforcing non-nullable workspace IDs, MCP session tracking tables, agent checkpoint lineage, and `mcp_agent_watermarks`.