# Changelog

## [1.6.0] - 2026-10-25 — Universal Zero-Dependency Onboarding (Windows, macOS, Linux, Chromebook & Android)

### Added
- **Windows PowerShell 1-Command Installers (`install.ps1` & `setup.ps1`)**: Zero-dependency Windows 10/11 bootstrapper that automatically downloads Portable Zero-Admin PostgreSQL 16 (`~/.garuda/runtime/postgres`) and Portable Go 1.24 (`~/.garuda/toolchain/go`), compiles with `CGO_ENABLED=0`, runs migrations, configures `%APPDATA%\Claude\claude_desktop_config.json` and `.cursor\mcp.json`, and launches the dashboard.
- **Universal 1-Command Installer (`install.sh` & upgraded `setup.sh`)**: Zero-dependency installer for macOS (Apple Silicon & Intel), Linux, Chromebook (ChromiumOS Crostini), and Android (Termux) with automatic Portable Go toolchain provisioning, user-space PostgreSQL cluster initialization (`~/.garuda/data/pg`), and automatic `CGO_ENABLED=0` fallback when `gcc`/`clang` is absent.
- **Pure-Go (`!cgo`) TypeScript AST Fallback (`internal/ast/typescript/parser_nocgo.go` & `internal/ast/treesitter/bridge_nocgo.go`)**: Enables 100% Pure-Go compilation (`CGO_ENABLED=0`) and cross-OS binary builds without requiring a C compiler (`gcc`, `clang`, MinGW, or Xcode Command Line Tools).
- **Garuda System Doctor (`garuda doctor [--fix]`)**: Built-in CLI diagnostic and auto-remediation command (`cmd/garuda/doctor_cmd.go`) verifying OS/arch, `.env`, PostgreSQL connectivity, migrations, binaries, Talon hooks, and API gateway health.
- **Chromium, ChromeOS, Android & Desktop PWA Support (`/manifest.json`, `/sw.js`, `/icon.svg`)**: 1-click installable Progressive Web App manifest, Service Worker, and mobile/tablet responsive CSS for `/dashboard`.
- **Dedicated Installation & Troubleshooting Guide (`INSTALL.md`)**: Step-by-step 1-command onboarding for non-technical users (PMs, founders) and engineers, plus a 12-scenario installation error resolution reference.

All notable changes to Garuda are documented in this file.

## [v1.5.0] - 2026-10-25

### Added — Garuda Talon Active Agent Harness & 4 Pre-Flight Core Bridge Primitives
- **Garuda Talon Active Execution Harness (`internal/talon`, `cmd/garuda/talon_cmd.go`, `cmd/garuda-talon`)**:
  - `garuda talon init` (automatically triggered by `garuda init` and `./setup.sh`): Compiles live workspace ADRs (`document_claims`), YAML policies (`policies/*.yaml`), and top high-blast-radius hubs into `.garuda/TALON_CONTRACT.md`, `.cursor/rules/garuda-talon.mdc` (`alwaysApply: true`), `.cursor/mcp.json` (`GARUDA_TALON_ENFORCE=true`), `CLAUDE.md`, `.claude/settings.json` (`PostToolUse` hook running `garuda talon guard --hook`), `AGENTS.md`, `.windsurfrules`, and `.git/hooks/pre-commit` + `commit-msg`.
  - `garuda talon brief [target]`: Emits deterministic pre-flight architectural context, transitive blast radius, and applicable ADR rules before an agent modifies a symbol or file.
  - `garuda talon guard [--staged | --diff <rev> | --files <csv> | --hook]`: Runs sub-second incremental AST sync on dirty working-tree files, maps git diff hunks to touched AST symbols, evaluates YAML policies and ADR claims in net-new diff scope, and emits a deterministic **Self-Repair Prompt** on `BLOCK`.
  - `garuda talon run -- <agent-cmd>`: Supervisor loop that executes an autonomous coding agent, intercepts working-tree changes, runs `talon guard`, and automatically loops the agent with the self-repair prompt until `ALLOW` is reached and Merkle-sealed.
- **Active MCP Server Enforcement (`cmd/garuda-mcp/main.go`, `cmd/garuda-mcp/tools_talon.go`)**:
  - Added MCP `prompts/list` & `prompts/get` (`garuda.architectural_contract`) and `resources/list` & `resources/read` (`garuda://workspace/contract`, `garuda://workspace/adr-guardrails`).
  - Added `garuda.talon_pre_edit` and `garuda.talon_verify_edit` MCP tools (expanding the registry to **23 MCP tools**).
  - Added active session-level contract injection (`GARUDA_TALON_ENFORCE=true` default): the first tool call in any MCP session automatically injects the workspace's active ADR `MUST`/`MUST_NOT` rules and YAML policy guardrails into the response payload so agents are bound even if they skip `garuda.briefing`.
- **Git-Diff Native Targeting (`internal/diff/diff.go`, `cmd/garuda/check_cmd.go`)**:
  - Added `ParseUnifiedDiff`, `CollectGitDeltas`, and `ResolveTouchedEntities` to map `git diff --unified=0` hunks (`--staged`, `--diff <rev>`, `--files`) directly to affected AST entities (`[line_start, line_end]`) and compute union transitive blast radius across all modified symbols.
- **Net-New vs. Baseline Violation Filtering (`internal/policy/types.go`, `evaluator.go`, `engine.go`)**:
  - Added `DiffScopeFilter` (`TouchedEntityIDs`, `TouchedPackages`, `OnlyNetNew`) to `policy.RunOptions` and `Evaluator.EvaluateWithScope`, partitioning violations touching the active diff (`net_new_violations_count`) from pre-existing legacy debt (`baseline_violations_count`).
- **Incremental Working-Tree Re-Index (`cmd/garuda/analyze.go`)**:
  - Added `--incremental` and `--files` flags with shared in-memory package caching so dirty working-tree edits are synchronized into PostgreSQL before `garuda check --diff` or `garuda talon guard` evaluates them.
- **Python & TypeScript Cross-File Alias & Instance Call Resolution (`internal/ast/python`, `internal/ast/typescript`)**:
  - Upgraded Python AST parser and integration to resolve `import X as Y`, `from X import Y as Z`, and class instance calls (`self.repo.fetch()`, `local_repo.fetch()`) across files at `0.95` static confidence.
  - Upgraded TypeScript Tree-sitter parser and integration to resolve named import aliases (`import { X as Y }`), namespace imports (`import * as NS`), `new Class()` instantiations, and instance/field method calls across files at `0.95` static confidence.

## [v1.4.0] - 2026-10-24
- Redesigned `/dashboard` as a 6-surface Software Intelligence Instrument (`Briefing`, `Architecture Map`, `Change Review Workbench`, `Policies`, `Contradictions & Evidence`, `Governance & Audit Ledger`) with global `⌘K` palette, 1-hop/2-hop neighborhood graph, and synchronized 4-tab Inspector Drawer.
- Fixed Benchmark 024 (`TestBenchmark_024_GoWork_CrossModule_ResolutionAndBlastRadius`) to assign `1.0` exact compiler-verified confidence to cross-module `IMPLEMENTS` edges.

## [v1.3.0] - 2026-10-24

### Security & Trust Hardening (GPT-5 Astra Due-Diligence Audit Remediation)
- **Cryptographic Ed25519 JWT Enforcement in `AuthMiddleware` (`internal/api/auth_middleware.go`)**:
  - Wired `s.jwtConfig.ValidateToken(token)` directly into `Server.AuthMiddleware` across all `/api/v1` routes.
  - Forged, malformed, or expired Bearer tokens now strictly return `401 Unauthorized` instead of passing a presence-only check.
  - Request `tenant_id` and `actor` identity are extracted directly from the cryptographically verified JWT claims (`auth.ContextWithActorAndTenant`) rather than defaulting to `tenant.CanonicalID` or trusting unverified `X-Garuda-Actor` headers.
  - Added `TestAuthMiddleware_EnforcesRealEd25519JWTValidation` in `internal/api/router_test.go`.
- **Edge-Level Epistemic Verification for `SUPPORTED` Claims (`internal/runtime/verifier.go`)**:
  - Hardened Statement 3 of `RecomputeWorkspaceVerification` so a static claim (`from_entity_id -> to_entity_id`) is promoted to `SUPPORTED` only when runtime evidence confirms **both** the caller (`from_entity_id`) and the specific target (`to_entity_id`) via a correlated `runtime_edges` record or co-execution in the same `trace_id`. Observing only the caller no longer auto-promotes all of its outgoing static claims.
- **Unified Runtime Span Ingestion & Correlation (`internal/api/handler.go` & `internal/api/runtime_handlers.go`)**:
  - Wired `POST /api/v1/telemetry/spans` (and `/api/v1/runtime/spans`) to `HandleIngestRuntimeSpans`, ensuring ingested OpenTelemetry spans automatically run `EntityCorrelator.Correlate` for both caller and callee symbols, populate `runtime_observations` and `runtime_edges`, and write synchronous `CONTRADICTED` or `SUPPORTED` records into `claim_verifications`.
  - Updated `scripts/runtime_contradiction_test.py` to authenticate with a real Ed25519 JWT or session cookie.
- **Fail-Closed Policy Parsing, Evaluation & Merkle Anchoring (`internal/policy/parser.go` & `internal/policy/engine.go`)**:
  - Updated `ParseDirectory` (`internal/policy/parser.go`) to fail closed and return an immediate error if any `.yaml`/`.yml` file in the policy directory fails YAML unmarshaling or schema/predicate validation, instead of silently skipping malformed policy files.
  - Updated `Engine.Run` (`internal/policy/engine.go`) so any predicate evaluation error or non-dry-run persistence/Merkle anchoring failure aborts the run and returns an error rather than continuing with `ALLOW`.
  - Added `TestParseDirectory_FailsClosedOnMalformedPolicy` in `internal/policy/engine_run_test.go`.
- **Go Type-Checking Completeness & `vendor/` Resolution (`internal/analyzer/workspace_analyzer.go`, `importer.go`, `model.go`)**:
  - Added `TypeErrors int` and `AnalysisComplete bool` to `analyzer.Stats` and bounded `TypeDiagnostics []string` to `analyzer.Result` instead of silently discarding `go/types` errors.
  - Upgraded `MultiModuleImporter.ImportFrom` (`internal/analyzer/importer.go`) to resolve packages from module-local `vendor/` directories before falling back to `importer.Default()`.
- **Production Operations & Probe Alignment (`internal/api/handler.go`, `deploy/k8s/garuda-stack.yaml`, `deploy/compose/docker-compose.prod.yml`)**:
  - Registered `/health`, `/healthz`, `/readyz`, and `/api/v1/system/health` (with database ping verification on `/readyz` and `/api/v1/system/health`).
  - Hardened `deploy/k8s/garuda-stack.yaml` with non-root `securityContext`, read-only root filesystem, `NetworkPolicy`, and secret references (`JWT_PRIVATE_KEY_HEX`, `JWT_PUBLIC_KEY_HEX`, `DATABASE_URL`).
  - Hardened `deploy/compose/docker-compose.prod.yml` by removing public `0.0.0.0:5432` and `0.0.0.0:6379` port bindings for PostgreSQL and Redis and requiring explicit secrets (`:?must be set in production`).

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

## [v1.1.0] - 2026-09-20
- Initial public release of the Garuda Epistemic Architecture & Governance Engine (`garuda`, `garuda-api`, `garuda-mcp`), Dual-Root RFC 6962 Merkle Ledger, 20-tool MCP server, and D3.js Epistemic Truth Graph visualizer.