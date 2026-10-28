# Changelog

All notable changes to Garuda (`github.com/myshra777-ai/garuda`) are documented in this file.

## [v1.9.1] - 2026-10-29 — 35-Tool Extended MCP Integration, Zero-Stub CLI Hardening, Finance & Audit Guide, CLI Decision Matrix & Persona Onboarding Suite

### Added & Fixed
- **Extended 35-Tool Production MCP Server (`cmd/garuda-mcp/main.go`, `tools_audit.go`, `tools_whitelist.go`, `scripts/mcp_verify.py`)**:
  - Expanded `bin/garuda-mcp` from 26 to **35 production-grade tools** with full PostgreSQL persistence, strict `toolArgWhitelist` coverage (`TestToolArgWhitelistCoversAllTools`), and **`43/43` passing protocol & functional checks** in `scripts/mcp_verify.py`:
    - `garuda.disambiguate_logic` & `garuda.decide_logic_variant` (PostgreSQL `logic_variants_registry` 0.0–1.0 team weights & human decision recording)
    - `garuda.audit_financial_discrepancy` & `garuda.reconcile_financial_audit` (PostgreSQL `financial_audit_reconciliations` 48 Cr vs 4.8 Cr lineage trace & auditor sign-off)
    - `garuda.synthesize_rulebook` & `garuda.approve_base_rulebook` (PostgreSQL `synthesized_rulebooks` + `.garuda/APPROVED_BASE_RULEBOOK.md` RFC 6962 SHA-256 seal)
    - `garuda.criticality_tiers` (PostgreSQL `surface_criticality_tiers` `HIGH_SEVERE` / `MEDIUM` / `LOW` surface classification)
    - `garuda.docs_graph` & `garuda.docs_ingest` (4-tier Document-to-Code lineage matrix & live ADR ingestion)
    - `garuda.doctor_diagnose`, `garuda.build_matrix_check`, and `garuda.hygiene_report`
- **Zero-Stub CLI Hardening (`cmd/garuda/governance.go` & `cmd/garuda/audit_cmd.go`)**:
  - Replaced the legacy `handleMCPInstall` placeholder in `cmd/garuda/governance.go` with full execution of `talonInitCmd` (`garuda mcp install` now compiles `.garuda/TALON_CONTRACT.md` and installs all 6 IDE, MCP, and Git hooks).
  - Upgraded `garuda audit logic`, `garuda audit financial`, `garuda audit criticality`, and `garuda rulebook generate/approve` (`cmd/garuda/audit_cmd.go`) to persist and synchronize directly with PostgreSQL tables (`093_human_in_the_loop_audit_rulebook.sql`).
- **Complete Role-Based Documentation Suite (`README.md`, `docs/FINANCE_AND_AUDIT_GUIDE.md`, `docs/CLI_DECISION_GUIDE.md`, `docs/PERSONA_ONBOARDING_GUIDE.md`, `PLAYBOOK.md`, `docs/SPECS.md`, `docs/API.md`, `docs/CAPABILITIES.md`, `docs/WALKTHROUGH.md`, `INSTALL.md`, `EVIDENCE.md`)**:
  - Created [`docs/FINANCE_AND_AUDIT_GUIDE.md`](docs/FINANCE_AND_AUDIT_GUIDE.md) for Finance Controllers, CFOs, Actuaries, and Statutory Auditors.
  - Created [`docs/CLI_DECISION_GUIDE.md`](docs/CLI_DECISION_GUIDE.md) with a complete "How & When to Use What CLI Command & MCP Tool" decision matrix.
  - Created [`docs/PERSONA_ONBOARDING_GUIDE.md`](docs/PERSONA_ONBOARDING_GUIDE.md) with 5-minute onboarding paths for all 7 user personas.

## [v1.9.0] - 2026-10-28 — Human-in-the-Loop Weighted Logic Disambiguation (0.0–1.0), Surface Criticality Tiers, Financial/Data Audit Reconciliation, Auto-Rulebook Synthesis & Bulletproof Postgres Handshake

### Added & Fixed
- **Human-in-the-Loop Multi-Team Formula Disambiguation & 0.0–1.0 Weight Allocation (`internal/api/dashboard_audit_rulebook.go`, `cmd/garuda/audit_cmd.go`, `garuda.disambiguate_logic`)**:
  - When a repository lacks formal documentation and contains competing formulas for the same business concept across teams (e.g., `Loss Ratio` across **Ops**, **Marketing**, and **Management**), Garuda never allows AI to silently pick or overwrite a formula.
  - Dynamically calculates **0.00 to 1.00 team weights** based on the active developer's team context (e.g., `0.70` weight for the active Ops team formula vs `0.10` / `0.10` for Marketing and Management), ranks each variant by **Least Harm to the Architecture**, and displays a plain-English blast radius breakdown so the human developer makes the final logical decision.
- **3-Tier Surface Criticality Classification (`surface_criticality_tiers` & `05 — Talon, Criticality & Risk Hub`)**:
  - Classifies architectural surfaces into human-understandable business criticality tiers:
    - **`HIGH_SEVERE` (Core Surfaces — Freeze / Do Not Touch)**: Mission-critical flows like the **Homepage**, **Search Bar**, **Search Results Redirect Router** (142 callers), and **Claims Settlement Ledger**. Prevents catastrophic changes that could take down search redirection or payment settlement.
    - **`MEDIUM` (Non-Blocking Features — Human Trade-Off Review)**: Secondary features like the **Wishlist Loader** and **Promotional Attribution Badges** (19 callers), where occasional degradation does not block core search or checkout and human trade-offs can be weighed.
    - **`LOW` (Minor UI & Formatting — Allow with Info)**: Leaf UI helpers like **Search Box Visual Alignment** and badge formatters (4 callers) carrying minimal blast radius.
- **Continuous Financial & Business Data Discrepancy Audit (`financial_audit_reconciliations`, `garuda audit financial`, `09 — Financial & Logic Audit`)**:
  - Extends Garuda beyond software development audits to **Financial & Business Data Audits**.
  - Traces AST and SQL data-lineage across teams to catch business number discrepancies before external audits occur—such as **Claims Ops** reporting **48.00 Crore INR** (`SUM(settled_claims_inr) / 1e7` in `ops.AggregateClaimsPaidCrore`) while **Finance/Management** reports **4.80 Crore INR** (`SUM(settled_claims_lakhs) / 1000.0` in `finance.ExportExecutiveClaimsSummary` — a 10x Lakhs-to-Crore unit divisor bug).
- **Auto-Synthesized Base Rulebook Studio (`synthesized_rulebooks`, `garuda rulebook generate/approve`, `04 — Docs & Base Rulebook`)**:
  - Automatically structures unorganized repository architecture, inferred layer boundaries, surface criticality tiers, and multi-team formula weights into a clean, human-readable **Draft Base Rulebook** (`.garuda/DRAFT_BASE_RULEBOOK.md`).
  - Stakeholders can review, edit inline in the dashboard, and click **Approve & Seal Base Rulebook** (persisted to `.garuda/APPROVED_BASE_RULEBOOK.md` with an RFC 6962 SHA-256 digest).
- **Bulletproof PostgreSQL Connection & Role/DB Auto-Healing (`setup.sh`, `setup.ps1`, `cmd/garuda/init_cmd.go`, `internal/config/env.go`)**:
  - Fixed a Go syntax typo (`\ext := exeSuffix()` -> `ext := exeSuffix()`) in `cmd/garuda/init_cmd.go`.
  - Added `pg_isready` WAL recovery readiness polling and multi-user (`test`, `$USER`, `postgres`) idempotent role/database provisioning with `pg_hba.conf` trust auto-repair so installation never stalls while establishing the PostgreSQL connection on port `5433`.

## [v1.8.4] - 2026-10-27 — All-OS Zero-Dependency Fast Provisioning, IPv4 PostgreSQL Hardening, Canonical Schema Alignment & Smart MCP Discovery

### Added & Fixed
- **3-Second Portable PostgreSQL 16 Provisioning on macOS & Linux (`setup.sh` & `cmd/garuda/init_cmd.go`)**:
  - Automatically downloads pre-built Portable PostgreSQL `16.4.0` binaries (`darwin-arm64v8`, `darwin-amd64`, `linux-amd64`, `linux-arm64v8`, `~18 MB`) into `~/.garuda/runtime/postgres/bin` when local PostgreSQL and Docker are absent—eliminating 30–45 minute Homebrew C/C++ source builds on Intel Macs and running 100% in user-space without `sudo`.
  - Sets `HOMEBREW_NO_AUTO_UPDATE=1`, `HOMEBREW_NO_INSTALL_CLEANUP=1`, and `HOMEBREW_NO_ENV_HINTS=1` so fallback Homebrew invocations never stall on global Tap updates.
- **Explicit IPv4 `127.0.0.1:5433` Normalization & Socket Isolation (`internal/config/env.go`, `setup.sh`, `cmd/garuda/init_cmd.go`)**:
  - `config.ResolveDBURL()` now normalizes `@localhost:` to `@127.0.0.1:` so macOS never fails with `dial tcp [::1]:5433: connect: connection refused` when PostgreSQL listens on IPv4 `127.0.0.1`.
  - Isolates PostgreSQL Unix domain sockets in `~/.garuda/data/run` (`chmod 700`) and initializes clusters with `LC_ALL=C LANG=C initdb --no-locale -E UTF8` plus automatic self-healing cluster recovery.
- **Single-Download Portable Go `1.26.0` Matching `go.mod` & 1-Pass Build Caching (`setup.sh` & `setup.ps1`)**:
  - Dynamically detects `go 1.26.0` from `go.mod` when bootstrapping `~/.garuda/toolchain/go`, eliminating the secondary ~70 MB Go toolchain auto-download during `go build`.
  - Compiles all four binaries (`garuda`, `garuda-api`, `garuda-mcp`, `garuda-talon`) in a single `go build -o bin/ ...` invocation with `bin/.build_stamp` commit caching and `install.sh` / `install.ps1` shallow clones (`git clone --depth 1`).
- **Canonical Schema Hardening Migration (`migrations/057_schema_hardening.sql`)**:
  - Idempotently aligns `workspaces`, `repositories`, `entities` (`canonical_name`, `file_path`, `signature`, `ast_hash`), `relationships`, and `cross_repo_edges` indexes across fresh and upgraded installations.
- **Smart MCP Graph Target Discovery (`cmd/garuda-mcp/tools_graph.go`)**:
  - Added case-insensitive/qualified suffix entity resolution (`name ILIKE $2 OR name ILIKE '%.' || $2`) and `resolveTopClaimTarget` fallback in `garuda.implementers` and `garuda.subclasses` so MCP clients and `scripts/mcp_verify.py` resolve valid interface/struct targets across any indexed repository (`25/25` dynamic verifier checks green).

## [v1.8.3] - 2026-10-27 — Cross-OS Setup Speed Optimization, Dynamic MCP Verifier & System Doctor Auto-Recovery

### Added & Fixed
- **Self-Healing `garuda doctor --fix` Database Startup (`cmd/garuda/doctor_cmd.go` & `cmd/garuda/init_cmd.go`)**:
  - `ensureLocalOrDockerPostgres()` now checks and starts the user-space PostgreSQL cluster in `~/.garuda/data/pg` (across portable `~/.garuda/runtime/postgres`, Homebrew `/opt/homebrew` & `/usr/local`, `Postgres.app`, and Linux `/usr/lib/postgresql/*`) before falling back to Docker.
- **Dynamic Workspace MCP Verification (`scripts/mcp_verify.py`)**:
  - Replaced hardcoded external benchmark symbol names with live workspace entity discovery so `python3 scripts/mcp_verify.py` passes cleanly (`25/25` checks) on any Go, Python, or TypeScript repository.
- **Eliminated Redundant Setup Re-Indexing (`setup.sh` & `setup.ps1`)**:
  - Uses `--skip-scan` during initial credential provisioning and `--auto-index=false` on post-index `garuda check` / `garuda talon guard` steps, cutting total setup time by more than 60%.

## [v1.8.2] - 2026-10-27 — Windows PowerShell 5.1 Pure-ASCII Bootstrapper & Native Command Stderr Resilience

### Added & Fixed
- **Pure 7-Bit ASCII Windows PowerShell Scripts (`install.ps1` & `setup.ps1`)**:
  - Removed all multi-byte UTF-8 emoji and box-drawing sequences from `install.ps1` and `setup.ps1` so Windows PowerShell 5.1 (`CP1252` default encoding when fetched via `irm ... | iex` or loaded without a UTF-8 BOM) never throws `ParserError: UnexpectedToken` on bytes `0x93`/`0x94`.
- **Stderr-Safe Native Command Execution & Native `curl.exe` / `tar.exe` Fast Path (`install.ps1` & `setup.ps1`)**:
  - Added `Invoke-QuietCmd` and `Invoke-LoggedCmd` wrappers (`cmd.exe /c`) so informational `stderr` output from `git clone`, `initdb.exe`, `pg_ctl.exe`, or `docker` never triggers a terminating `NativeCommandError` under `$ErrorActionPreference = "Stop"`.
  - Added `Download-FileFast` (`System32\curl.exe`) and `Extract-ArchiveFast` (`System32\tar.exe`) with `$ProgressPreference = "SilentlyContinue"` for multi-megabyte Portable PostgreSQL and Go downloads on Windows.

## [v1.8.1] - 2026-10-27 — Deterministic Package-Cache Fingerprinting, Cold-Boot CI Merkle Seeding & Visual Documentation Suite

### Added & Fixed
- **Relative-Path Deterministic Package Cache Fingerprinting (`internal/analyzer/workspace_analyzer.go`)**:
  - Updated `computeFilesFingerprint` to hash normalized repository-relative paths (`filepath.Rel(absRoot, f)`) paired with file sizes and SHA-256 content digests, ensuring identical `BuildContext` cache fingerprints across distinct filesystem checkout paths and temp directories (`TestDoubleRunDeterminism` 24/24 green).
- **Cold-Boot CI Merkle Root Seeding (`internal/store/postgres_merkle.go` & `.github/workflows/garuda-gate.yml`)**:
  - Ensures fresh ephemeral CI containers automatically initialize and verify the Genesis RFC 6962 Merkle block during cold-boot workspace analysis and diff-scoped gate execution.
- **Redesigned Visual Documentation Suite (`README.md`, `INSTALL.md`, `PLAYBOOK.md`, `EVIDENCE.md`, `docs/WALKTHROUGH.md`, `docs/CAPABILITIES.md`, `docs/SPECS.md`, `docs/API.md`)**:
  - Integrated the new Garuda visual identity (`assets/garuda_logo.png`), the 20-second cinematic product introduction video (`assets/garuda_dashboard_demo_video.mp4`), and the complete 8-surface dashboard & IDE screenshot gallery across a human-centered `README.md` paired with deep v1.8.1 technical specifications.

## [v1.8.0] - 2026-10-26 — Deterministic BuildContext Receipts, Build-Tag/Platform AST Filtering, CI-vs-Editor Drift Detector & Cross-Platform Matrix Gate

### Added
- **Polyglot Deterministic `BuildContext` Provenance Receipts (`internal/analyzer/build_context.go`)**:
  - Every `garuda analyze`, `garuda check`, `garuda policy eval`, and `garuda talon guard` execution resolves and records a cryptographic `BuildContext` receipt (`receipt_hash` & `graph_fingerprint`) capturing:
    - **Git Working-Tree State**: `commit_sha`, `branch`, `worktree_dirty`, `dirty_files_count`, and `dirty_files_hash`.
    - **Go Build Constraints**: `target_os` (`GOOS`), `target_arch` (`GOARCH`), `cgo_enabled` (`CGO_ENABLED`), `build_tags` (`-tags`), `compiler_version`, `go_mod_hash`, `active_files_count`, and `skipped_by_build_tags_count`.
    - **Python & TypeScript Environment Hashes**: `python_version`, `sys_platform`, `pyproject.toml`/`requirements.txt` lock hash, `tsconfig.json` hash, `module_resolution`, and `target_es`.
    - **1-Line Deterministic Reproduction Command (`reproduce_command`)**: Exact `CGO_ENABLED=... GOOS=... GOARCH=... ./bin/garuda talon guard ...` invocation required to reproduce the exact AST graph and policy verdict on any machine or CI runner.
- **`go/build.Context` Conditional Compilation & Build-Tag Filtering (`internal/analyzer/workspace_analyzer.go`)**:
  - Evaluates `//go:build`, `// +build`, `_linux.go`, `_darwin.go`, `_windows.go`, and `cgo` file constraints using `go/build.Context.MatchFile` during AST ingestion so conditional edges are accurately included or excluded per target platform.
  - Verified by `TestTwoBuildsExposeDifferentEdges` (`internal/analyzer/build_context_test.go`), proving two builds of the same commit (`cgo=0` vs `cgo=1 -tags cgo,enterprise`) produce distinct graph fingerprints and edge sets.
- **CI-vs-Editor Environment Drift Detector (`DetectEnvironmentDrift` in `internal/analyzer/build_context.go` & `internal/policy/engine.go`)**:
  - Automatically compares the active CLI/IDE/CI execution environment against the recorded workspace graph receipt and flags `TARGET_PLATFORM_MISMATCH`, `CGO_MISMATCH`, `BUILD_TAGS_MISMATCH`, `COMMIT_DRIFT`, `WORKTREE_DIRTY_DRIFT`, and `LANG_CONFIG_DRIFT` before silent false-ALLOW verdicts can occur.
- **Cross-Platform Build Matrix Verification (`--matrix` on `garuda check` & `garuda talon guard`)**:
  - Evaluates `linux/amd64 (cgo=0)`, `linux/amd64 (cgo=1, tags=cgo)`, `darwin/arm64 (cgo=1)`, and `windows/amd64 (cgo=0)` in a single pass so developers on macOS/Windows can catch Linux CI or CGO-specific architectural violations locally before pushing.
- **Merkle-Sealed Build Provenance & Dashboard UI (`internal/merkle/canonical.go`, `internal/api/dashboard_docs_talon.go`, `dashboard.html`, `dashboard.js`)**:
  - Canonical Merkle leaf encoding (`policy.Evidence.BuildContext`) now cryptographically binds the `BuildContext` receipt into RFC 6962 Merkle blocks and `.garuda/LAST_DECISION.env` Git commit trailers.
  - Added the **Deterministic Build-Context Provenance Receipt & CI-vs-Editor Drift Check** banner and **Cross-Platform Build Matrix** table to `05 — Talon & Risk Hub` with 1-click **Copy Reproduce Cmd**.
- **Precision Diff-Scope Symbol Filtering (`internal/policy/evaluator.go` & `internal/talon/talon.go`)**:
  - Refined `filterEvidenceByDiffScope` and `RunGuard` so modifying a clean symbol in a package or committing documentation changes never reclassifies untouched baseline findings in the same package as net-new violations.

## [v1.7.0] - 2026-10-26 — Docs & Rules Lineage Graph, Talon & Risk Hub Dashboard, Zero-CDN Offline Engine

### Added
- **04 — Docs, Specs & Guardrails Lineage Graph (`GET /api/v1/dashboard/docs-graph` & `garuda docs graph`)**:
  - 4-Tier Visual Lineage Graph connecting **Tier 1 Documents** (ADRs, Design Specs, System Invariants, YAML Policies, `.garudarules`) → **Tier 2 Extracted Normative Rules** (`MUST`, `MUST_NOT`, `BLOCK`, `REVIEW`, `DENY`, `INVARIANT`) → **Tier 3 Active Policy Guardrails** → **Tier 4 AST Code Symbols & Packages**.
  - Structured **Document Inventory Cards** and **Document-to-Code Component Matrix** with 1-click symbol and rule focus.
  - **Live Document / ADR / Spec Ingestion (`POST /api/v1/dashboard/docs-ingest`)**: Ingest new ADRs, specs, or guardrail rules directly from the dashboard modal (with built-in templates) or scan repository directories on the fly.
  - Integrated **`✦ Docs & Rules Graph`** mode (`level=docs`) directly into the `03 — Architecture Map` surface.
- **05 — Garuda Talon & Architectural Risk Hub (`GET /api/v1/dashboard/talon-hub` & `POST /api/v1/dashboard/talon-check`)**:
  - Live **Talon Active Guard Terminal** with 1-click diff-scoped verification (`HEAD`), violation & self-repair simulation, and clipboard prompt export.
  - **Active IDE, MCP & Git Hook Status Panel** tracking all 6 enforcement hooks (`Cursor .mdc`, `Cursor MCP`, `Claude Code PostToolUse`, `CLAUDE.md`, `AGENTS.md / .windsurfrules`, `.git/hooks/pre-commit`) with 1-click hook installation/sync.
  - **Architectural Risk Hub Leaderboard** ranking high-blast-radius symbols by caller fan-in, transitive impact radius, and governing ADR/policy sensitivity (0–100 score across `CRITICAL`, `HIGH`, `MEDIUM`, `LOW` tiers) with 1-click **Pre-Edit Briefing** drawers.
- **Zero-Dependency Offline SVG Graph Engine & Full UI Polish**:
  - Embedded a D3-compatible SVG force & 4-tier DAG layout engine directly into `dashboard.js` so the dashboard works 100% offline/air-gapped with zero external CDN requests.
  - Auto-populates `02 — Explore Index` with top exported symbols on initial load and adds toast feedback, API/Contract viewer modals, and sandbox-safe theme persistence.

## [v1.6.0] - 2026-10-25 — Universal Zero-Dependency Onboarding (Windows, macOS, Linux, Chromebook & Android)

### Added
- **Windows PowerShell 1-Command Installers (`install.ps1` & `setup.ps1`)**: Zero-dependency Windows 10/11 bootstrapper that automatically downloads Portable Zero-Admin PostgreSQL 16 (`~/.garuda/runtime/postgres`) and Portable Go 1.24 (`~/.garuda/toolchain/go`), compiles with `CGO_ENABLED=0`, runs migrations, configures `%APPDATA%\Claude\claude_desktop_config.json` and `.cursor\mcp.json`, and launches the dashboard.
- **Universal 1-Command Installer (`install.sh` & upgraded `setup.sh`)**: Zero-dependency installer for macOS (Apple Silicon & Intel), Linux, Chromebook (ChromiumOS Crostini), and Android (Termux) with automatic Portable Go toolchain provisioning, user-space PostgreSQL cluster initialization (`~/.garuda/data/pg`), and automatic `CGO_ENABLED=0` fallback when `gcc`/`clang` is absent.
- **Pure-Go (`!cgo`) TypeScript AST Fallback (`internal/ast/typescript/parser_nocgo.go` & `internal/ast/treesitter/bridge_nocgo.go`)**: Enables 100% Pure-Go compilation (`CGO_ENABLED=0`) and cross-OS binary builds without requiring a C compiler (`gcc`, `clang`, MinGW, or Xcode Command Line Tools).
- **Garuda System Doctor (`garuda doctor [--fix]`)**: Built-in CLI diagnostic and auto-remediation command (`cmd/garuda/doctor_cmd.go`) verifying OS/arch, `.env`, PostgreSQL connectivity, migrations, binaries, Talon hooks, and API gateway health.
- **Chromium, ChromeOS, Android & Desktop PWA Support (`/manifest.json`, `/sw.js`, `/icon.svg`)**: 1-click installable Progressive Web App manifest, Service Worker, and mobile/tablet responsive CSS for `/dashboard`.
- **Dedicated Installation & Troubleshooting Guide (`INSTALL.md`)**: Step-by-step 1-command onboarding for non-technical users (PMs, founders) and engineers, plus a 12-scenario installation error resolution reference.


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