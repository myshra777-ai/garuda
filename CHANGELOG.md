# Changelog

All notable changes to Garuda are documented in this file.

## [v1.2.0] - 2026-10-24

### Added
- **Unified Change-Control Decision CLI (`garuda check`)**:
  - Added `cmd/garuda/check_cmd.go` implementing `garuda check [dir] --project <ws> --target <symbol> --policy <dir> --save --fail-on-block --json`.
  - Connects all five Garuda planes in a single command: (1) Software State & Transitive Blast Radius, (2) 3-State Epistemic Evidence (`SUPPORTED` / `UNVERIFIED` / `CONTRADICTED`), (3) Declarative Policy Evaluation, (4) Governance Verdict (`ALLOW` / `WARN` / `REVIEW` / `BLOCK`), and (5) RFC 6962 Merkle Ledger Seal.
- **1-Command Zero-Config Onboarding (`./setup.sh` & Upgraded `garuda init`)**:
  - Added top-level `./setup.sh` script and upgraded `cmd/garuda/init_cmd.go` (`garuda init`) to automatically start or create the `garuda-postgres` Docker container, compile `bin/garuda`, `bin/garuda-api`, and `bin/garuda-mcp`, persist `.env` settings and Ed25519 JWT keys, run idempotent migrations, provision the owner account, configure Cursor (`.cursor/mcp.json`) and Claude Desktop MCP, and verify `/health`, `/signup`, and MCP compliance in a single command.
- **Automatic `.env` Loader & Persistent JWT Key Management (`internal/config/env.go`)**:
  - Added `internal/config/env.go` (`LoadDotEnv`, `UpsertDotEnv`, `ResolveDBURL`, `ResolveAPIAddr`) invoked automatically at startup across `garuda`, `garuda-api`, and `garuda-mcp`.
  - Upgraded `internal/auth/jwt.go` (`NewJWTConfig`) to automatically persist generated Ed25519 `JWT_PRIVATE_KEY_HEX` and `JWT_PUBLIC_KEY_HEX` keys to `.env` so tokens remain valid across daemon restarts without manual copy-pasting.
- **Graph-Native Policy Engine Predicates (`internal/policy/evaluator.go` & `internal/policy/parser.go`)**:
  - `transitive_claim_exists`: Multi-hop recursive CTE (`WITH RECURSIVE call_path`) detecting hidden transitive dependencies up to `max_depth` hops (e.g., `Handler -> Helper -> Database`).
  - `layer_violation`: Enforces architectural layer boundaries between `from_layer` and `to_layer` package regex patterns across `CALLS` and `IMPORTS` edges.
  - `cycle_exists`: Package-level circular dependency detection (`WITH RECURSIVE cycle_search`, `A -> B -> C -> A`).
  - `blast_radius_exceeds`: Flags entities whose distinct inbound caller count exceeds `max_callers`.
  - Added starter policy `policies/architecture-layer-guardrails.yaml`.
- **Multi-Language Call-Graph (`CALLS`) Extraction Parity**:
  - **Python (`internal/ast/python/parser.go` & `integration.go`)**: Added `rePyCall` function and method body call-site extraction and resolution for `self.method()`, intra-module functions, and imported symbols.
  - **TypeScript (`internal/ast/typescript/parser.go` & `integration.go`)**: Added Tree-sitter `call_expression` traversal and Pass-2 `CALLS` edge resolution for `this.method()` and module functions, plus automatic `modulePath` extraction from `package.json` in `cmd/garuda/analyze.go`.
  - **Go (`internal/analyzer/workspace_analyzer.go`)**: Added intra-package direct function call (`*ast.Ident`) resolution via `info.Uses[fun]`, populated `ReceiverType` on extracted `KindMethod` entities, and added cross-package structural interface satisfaction (`deriveWorkspaceInterfaceImplementations`).
- **RFC 6962 Cryptographic Decision Proof Verification (`internal/store/decision_store.go`)**:
  - Added `VerifyDecisionCryptographic` recomputing `merkle.CanonicalDecision` leaf hashes and verifying stored `V1Proof` inclusion paths in `GET /api/v1/evidence/verify/{id}` (`internal/api/verification_handlers.go`).

### Changed
- **Idempotent Database Migration Engine (`internal/store/migrate.go` & `migrations/*.sql`)**:
  - Upgraded all 14 non-idempotent SQL migration files (`001`, `002`, `003`, `004`, `005`, `007`, `008`, `013`, `014`, `019`, `021`, `022`, `023`, `084`) with `DO $$ BEGIN CREATE TYPE ... EXCEPTION WHEN duplicate_object THEN NULL; END $$;` and `IF NOT EXISTS` guards.
  - Upgraded `store.Migrate` to automatically locate `migrations/` relative to the binary or working directory and gracefully record pre-existing schema objects (`SQLSTATE 42710`, `42P07`, `42701`, `42P16`) when a database was previously populated via manual `psql -f` loops.
  - Wired automatic `store.Migrate` execution into `garuda init`, `garuda dev`, `garuda analyze --save`, and `garuda check`.
- **CLI/API-Friendly `/signup` Endpoint & Canonical Workspace Access (`internal/api/signup_handlers.go` & `internal/store/workspace.go`)**:
  - Updated `HandleSignupPOST` to support both `application/x-www-form-urlencoded` and `application/json`, default `confirm = password` when `confirm` is omitted by CLI/curl callers, return structured JSON (`201 Created` / `200 OK`) for API/curl clients, and grant interactive users membership in the canonical tenant and CLI-analyzed workspaces.
  - Updated `CreateWorkspace` (`internal/store/workspace.go`) to automatically add existing tenant owners/members to `workspace_members` when new workspaces are created via CLI.
- **Smart Dynamic MCP Graph Resolution (`cmd/garuda-mcp/tools_graph.go` & `scripts/mcp_verify.py`)**:
  - Upgraded `resolveSubjectEntity` to support case-insensitive and qualified suffix matching (`%.Name`), and added `resolveTopClaimTarget` fallback in `handleSubclasses` and `handleImplementers` so MCP tools never fail with transport-level JSON-RPC errors when queried with sample symbol names from other repositories.
  - Updated `scripts/mcp_verify.py` and `scripts/mcp_tool_smoke.sh` to auto-load `.env` and dynamically discover workspace hubs and interfaces (`23/23 checks passed`).

### Fixed
- Fixed `internal/analyzer/workspace_analyzer.go` `Entity.ReceiverType` population on Go methods and resolved `e.Receiver` field reference in `deriveWorkspaceInterfaceImplementations`.
- Fixed `internal/policy/evaluator.go` ignoring `scope.package` and `scope.pattern` in `evalEntityExists` and `evalClaimExists`.
- Fixed `internal/runtime/verifier.go` (`RecomputeWorkspaceVerification`) so runtime edges observed between known entities that have no corresponding static `claims` edge are flagged as `CONTRADICTED` (shadow runtime dependencies) rather than only matching hardcoded string markers.
- Fixed VS Code extension MCP client (`vscode-extension/src/mcpClient.ts`) binary path lookup (`bin/garuda-mcp`) and JSON-RPC tool names (`garuda.detect_contradictions`, `garuda.blast_radius`, `garuda.governance.status`), and added `name`/`symbol` resolution to `handleBlastRadius` in `cmd/garuda-mcp/tools_governance.go`.

---

## [v1.0.0] - 2026-10-12

### Added
- **Hygiene & Advisory CLI**: Introduced `garuda hygiene` command with typed advisory analysis core, contract hardening, and `ponytail` compatibility alias.
- **MCP Integration & Server**:
  - Consolidated MCP into a long-lived stdio server binary (`bin/garuda-mcp`).
  - Added MCP tools: `garuda.briefing`, `garuda.handoff`, `garuda.resume`, `garuda.blast_radius`, `garuda.verify_policy_evaluation`, and graph structure exploration tools (`garuda.entities`, `garuda.neighbors`, `garuda.find_entity`, `garuda.subclasses`, `garuda.implementers`).
  - Added session tracking and telemetry per tool invocation in `mcp_sessions` and `mcp_tool_calls`.
- **Control Plane & Dashboard**:
  - Multi-tab Control Plane dashboard (Operations, Tenants, Decisions, Agents, Business metrics).
  - Added read-only DB role and background telemetry refresh jobs (`telemetry_aggregates`).
  - UI improvements including embedded static CSS/JS assets, workspace picker on sidebar, detail drawers, and revision chain tracking for decisions.
- **Multi-Tenancy & Workspace Scope**:
  - Schema additions for tenant memberships (`tenant_members`, `workspace_members`) and user signup flow (`/signup`).
  - Session-scoped and workspace-scoped data isolation across storage, HTTP API, and CLI.
- **Multi-Language Parsing**:
  - Python AST parsing support and unified polyglot knowledge graph construction.
  - TypeScript parser support via Tree-sitter.
- **Merkle Verification System**:
  - RFC 6962 Merkle tree structure with golden test vectors.
  - Canonical evaluation encoders, epoch leaves schema, and epoch root hash writes.
- **Benchmarking**:
  - Expanded benchmark corpus to 20 fixtures (GAP-20) with verified deterministic output.