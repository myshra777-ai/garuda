# CHANGELOG

All notable changes to Garuda since release `v0.2.0` are documented below.

## [Unreleased]

### Added
- **Hygiene Engine & CLI**:
  - Introduced `garuda hygiene` CLI command (with ponytail compatibility alias) powered by a typed advisory analysis core.
  - Added CI gating and report-only hygiene capabilities.
- **MCP (Model Context Protocol) Server & Tools**:
  - Consolidated MCP functionality into a single long-lived stdio server process.
  - Added governance, graph structure, and workflow MCP tools: `garuda.briefing`, `garuda.verify_policy_evaluation`, `garuda.blast_radius`, `garuda.handoff`, and `garuda.resume`.
  - Added session recording, tool call tracking, and telemetry in dedicated `mcp_sessions` and `mcp_tool_calls` tables.
- **Multi-Tenancy & Workspace Scope**:
  - Added tenant, user membership, and signup (`/signup`) schema and endpoints.
  - Added session-scoped workspace resolution and workspace picker UI on dashboard sidebar.
  - Enforced `workspace_id NOT NULL` constraint across core database tables.
- **Control Plane & Admin Dashboard**:
  - Added read-only admin control plane (`/admin`) featuring Business, Tenants, Operations, Agents, and Decisions tabs.
  - Added in-process `pgx` tracer metric queries for DB size, query percentiles, stalled sessions, and error rates.
  - Added a dedicated read-only database role for the Control Plane.
- **Merkle Ledger & Cryptographic Verification**:
  - Implemented RFC 6962-compliant Merkle tree engine with canonical byte encoding and golden vectors (ADR-0002).
  - Added epoch leaves schema, `EpochRoot`, canonical evaluation encoders, and tenant-scoped genesis blocks.
  - Added parent decision revision tracking (`decision.parent_id`).
- **Multi-Language Parsing (Experimental)**:
  - Added Python AST parser and tree-sitter TypeScript parser for polyglot semantic graph generation.

### Changed
- **CLI & Flag Standardization**:
  - Modularized `cmd/garuda` into decoupled command handlers and typed store layers.
  - Updated CLI flag `--fail-on BLOCK` to `--fail-on-block`.
- **Policy Enforcement**:
  - Gated policy evaluation persistence behind explicit `--save` flag.
  - Made active directory policy reconciliation opt-in via `--reconcile`.
- **Documentation & Compliance**:
  - Refreshed project documentation, specifications, and capabilities matrix to align with verified features.
  - Added SPDX license headers across new and existing source files.

### Fixed
- **Isolation & Security**:
  - Fixed cross-tenant and cross-workspace leaks across knowledge readers, evaluators, workspace resolvers, and CLI commands.
  - Addressed concurrent workspace creation race conditions in `CreateWorkspace`.
- **MCP & Protocol Handling**:
  - Suppressed response emission to JSON-RPC notifications in the MCP server.
  - Resolved workspace target resolution errors and fixed budget reservation release on MCP error paths.
- **Runtime Ingestion & Verification**:
  - Corrected `ON CONFLICT` target constraints during runtime ingestion.
  - Fixed `lockTask` error handling and corrected Merkle root text storage formatting.
  - Attributed `CALLS` graph edges to real calling entities.

### Removed
- Removed legacy audit-store functions, dead CAS code, and associated HTTP endpoints.
- Removed obsolete `fly.toml` and `render.yaml` deployment configs.
- Removed unneeded global "default" workspace fallbacks in CLI/HTTP paths.

### Semantic Changes
- **Quarantine Engine**: `runtime_vs_static` contradictions are now routed directly into the quarantine store (`quarantine_decisions`).
- **Graph Evidence Attribution**: Evidence is now systematically attached to all graph edges upon insertion.
- **Entity Scan Retention**: Semantic scans now preserve existing entity rows instead of discarding unreferenced entity metadata.