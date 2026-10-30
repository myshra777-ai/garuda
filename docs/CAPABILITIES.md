# 🧠 Garuda Capabilities & Verification Matrix (`v1.9.1` Universal)

> Grounded in the Garuda `v1.9.1` multi-language compiler/AST engine (`go/types`, Python AST, and dual Tree-sitter / Pure-Go `CGO_ENABLED=0` TypeScript AST), Garuda Talon active agent execution harness, diff-scoped graph policy evaluator, 3-state runtime verifier, RFC 6962 Merkle ledger, and zero-dependency cross-OS installer matrix.

---

## 1. Plane 01 — Multi-Language AST, Incremental Sync & Call-Graph Extraction

| Capability | Status | Verification Tier | Supported Semantics | Enforced Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **Go Struct, Interface & Field Extraction** | 🟢 **Production (GA)** | `100% Benchmark (24/24 Strict Suite)` | Struct tags (`json`, `db`, `validate`), line spans, pointer/slice flags, method sets | `ACGM Resolution Invariant` |
| **Go Method & Receiver Disambiguation** | 🟢 **Production (GA)** | `100% Benchmark (002-method-identity)` | Pointer and value receivers (`ReceiverType`), exported/unexported methods | `Canonical UUIDv5` |
| **Go Cross-Package, Intra-Package & `go.work` `CALLS`** | 🟢 **Production (GA)** | `go/types` + `go.work` + `vendor/` (`1.0` confidence) | Direct intra-package calls, cross-package calls, and multi-module `go.work` workspace calls | `AST_EXACT Provenance` |
| **Go Interface Satisfaction (`IMPLEMENTS`)** | 🟢 **Production (GA)** | `100% Benchmark (003, 009, 010, 024)` | Type-checked `types.Implements` + cross-module `go.work` structural satisfaction at `1.0` confidence | `Full Recall Gate` |
| **Go Struct & Interface Embedding (`EMBEDS`)** | 🟢 **Production (GA)** | `100% Benchmark (006-embedding)` | Anonymous struct fields and embedded interfaces | `Composition Invariant` |
| **Sub-Second Incremental Working-Tree Sync** | 🟢 **Production (GA)** | `garuda analyze --incremental --files` | Shared in-memory package cache + targeted file re-indexing for dirty working-tree edits | `Deterministic Graph Parity` |
| **Python AST, Alias & Instance Call Graph** | 🟢 **Production (GA)** | `Multi-Repo Corpus + Alias/Receiver Tests` | Modules, classes, methods, functions, `import X as Y`, `from X import Y as Z`, and `self.attr.method()` / instance calls (`0.95` confidence) | `Deterministic SHA-1 UUID` |
| **TypeScript Tree-Sitter & Pure-Go (`!cgo`) AST Graph** | 🟢 **Production (GA)** | `Tree-sitter (cgo) + Pure-Go Lexical/Scope (!cgo)` | Interfaces, classes, functions, named/namespace import aliases (`import { X as Y }`, `import * as NS`), `new Class()` instance calls (`0.95` confidence, zero C compiler needed) | `Deterministic SHA-1 UUID` |
| **Cross-Repository Dependency Resolution** | 🟢 **Production (GA)** | `Postgres Workspace Symbol Cache` | Multi-repository module path and symbol bridge mapping within a workspace | `Workspace Isolation` |

---

## 2. Plane 02 — Epistemic Evidence & Edge-Level Runtime Drift Verification

| Capability | Status | Verification Tier | Supported Semantics | Enforced Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **3-State Epistemic Classification** | 🟢 **Production (GA)** | `Database Constraint + Verifier` | Every claim is classified as `SUPPORTED`, `UNVERIFIED`, or `CONTRADICTED` | `Honest Defaults (Never Upgrade Unverified)` |
| **Edge-Level `SUPPORTED` Promotion** | 🟢 **Production (GA)** | `internal/runtime/verifier.go` | Static claim (`from -> to`) is promoted to `SUPPORTED` only when runtime spans confirm **both** caller and callee in correlated traces | `Strict Edge Corroboration` |
| **OpenTelemetry Runtime Span Ingestion** | 🟢 **Production (GA)** | `POST /api/v1/telemetry/spans` | Correlates live service spans (`caller_symbol`, `target_endpoint`) against static workspace entities | `Tenant & Workspace Scoped` |
| **Bidirectional Runtime Contradiction Detection** | 🟢 **Production (GA)** | `10/10 Controlled Contradiction Suite` | Flags both explicit unapproved endpoints AND shadow runtime edges missing from static `claims` as `CONTRADICTED` | `Automatic Quarantine` |
| **Normative Documentation & ADR Claim Verification** | 🟢 **Production (GA)** | `garuda docs ingest` & `garuda docs verify` | Extracts `MUST` / `MUST_NOT` architectural claims from Markdown ADRs and verifies them against static & runtime edges | `Evidence-Backed Drift Detection` |

---

## 3. Plane 03 — Deterministic Graph-Native Policy Engine & Diff Scoping

| Predicate / Capability | Status | Query Mechanism | Purpose & Parameters |
| :--- | :--- | :--- | :--- |
| **`layer_violation`** | 🟢 **Production (GA)** | SQL Regex Join on `claims` + `entities` | Blocks direct `CALLS` or `IMPORTS` from `from_layer` package regex to `to_layer` package regex |
| **`transitive_claim_exists`** | 🟢 **Production (GA)** | Recursive CTE (`WITH RECURSIVE call_path`) | Detects multi-hop backdoor dependency chains (`from_name_pattern` -> ... -> `to_name_pattern` up to `max_depth`) |
| **`cycle_exists`** | 🟢 **Production (GA)** | Recursive CTE (`WITH RECURSIVE cycle_search`) | Detects circular package dependencies (`A -> B -> C -> A`) up to `max_depth` packages |
| **`blast_radius_exceeds`** | 🟢 **Production (GA)** | Inbound Caller Aggregation (`HAVING COUNT > N`) | Escalates changes to high-fan-in hubs when inbound callers exceed `max_callers` |
| **`claim_exists`** | 🟢 **Production (GA)** | Scoped `claims` Join (`scope.package`, `scope.pattern`) | Matches specific `claim_type`, `from_name_pattern`, `to_name_pattern`, and `min_confidence` |
| **`entity_exists`** | 🟢 **Production (GA)** | Scoped `entities` Filter (`scope.package`, `scope.pattern`) | Matches entities by `kind`, `name_pattern`, and `exported` status |
| **`contradiction_exists`** | 🟢 **Production (GA)** | `contradictions` + `claim_verifications` | Blocks or flags workspaces containing unresolved `CONTRADICTED` runtime or decision rows |
| **`verification_missing`** | 🟢 **Production (GA)** | Left-Join Anti-Match on `claim_verifications` | Warns when exported functions/methods lack `SUPPORTED` runtime verification |
| **`language_matches`** | 🟢 **Production (GA)** | Repository / Entity `language` Filter | Scopes policy rules to `go`, `python`, or `typescript` targets |
| **Net-New vs. Baseline Diff Scoping (`DiffScopeFilter`)** | 🟢 **Production (GA)** | `internal/diff` + `Evaluator.EvaluateWithScope` | Maps `git diff --unified=0` hunks to AST line spans (`[line_start, line_end]`) and isolates `net_new_violations` from `baseline_violations` |

---

## 4. Plane 04 — Unified Change-Control & RFC 6962 Merkle Ledger

| Capability | Status | Verification Tier | Supported Semantics | Enforced Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **Diff-Scoped Change-Control CLI (`garuda check`)** | 🟢 **Production (GA)** | `--diff`, `--staged`, `--files`, `--target` | Combines Git-Diff AST Targeting + Union Blast Radius + 3-State Evidence + Net-New Policy Evaluation + Merkle Block Seal | `Deterministic Verdict` |
| **RFC 6962 Merkle Tree & Inclusion Proofs** | 🟢 **Production (GA)** | `Golden Vectors (internal/merkle)` | Domain-separated leaf hashing (`0x00` leaf, `0x01` internal node), odd-leaf duplication guard, tamper detection | `Cryptographic Tamper-Evidence` |
| **Cryptographic Decision & Policy Verification** | 🟢 **Production (GA)** | `VerifyDecisionCryptographic` & `VerifyEvaluation` | Recomputes canonical v1 byte encoding (`CanonicalDecision`, `CanonicalEvaluation`) and verifies `V1Proof` path to epoch root | `Zero Trust Verification` |
| **Dual-Tier Epoch Snapshots** | 🟢 **Production (GA)** | `merkle_snapshots` + `merkle_epoch_leaves` | Anchors both static analysis leaves and runtime verification leaves into periodic tenant epochs | `Bitemporal Auditability` |

---

## 5. Plane 05 — Garuda Talon Active Agent Harness & 35-Tool MCP Surface

| Capability | Status | Verification Tier | Supported Semantics | Enforced Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **Active Contract Compilation (`garuda talon init`)** | 🟢 **Production (GA)** | `internal/talon/talon.go` | Synthesizes live ADRs (`document_claims`), YAML policies, and top hubs into `.garuda/TALON_CONTRACT.md`, `.cursor/rules/garuda-talon.mdc`, `CLAUDE.md`, `AGENTS.md`, and `.windsurfrules` | `Always-On Agent Context` |
| **Active MCP Session Enforcement (`GARUDA_TALON_ENFORCE=true`)** | 🟢 **Production (GA)** | `43/43 checks in scripts/mcp_verify.py` | 35 MCP tools + `prompts/list` (`garuda.architectural_contract`) + `resources/list` (`garuda://workspace/contract`) + automatic first-call guardrail injection | `Fail-Closed Agent Orientation` |
| **Pre-Edit & Post-Edit Guard (`garuda.talon_pre_edit` / `verify_edit`)** | 🟢 **Production (GA)** | `internal/talon` + `cmd/garuda-mcp` | Pre-flight blast radius & ADR check before edits; incremental AST sync + diff-scoped policy check + deterministic Self-Repair Prompt after edits | `Zero Blind Agent Edits` |
| **Claude Code & Git Hook Enforcement** | 🟢 **Production (GA)** | `.claude/settings.json` + `.git/hooks/` | `PostToolUse` hook (`garuda talon guard --hook` exiting `2` on `BLOCK`) and Git `pre-commit` / `commit-msg` Merkle provenance trailers | `Pre-Commit & Edit-Time Gate` |
| **Autonomous Supervisor Loop (`garuda talon run`)** | 🟢 **Production (GA)** | `talon.RunHarnessLoop` | Wraps any CLI coding agent, intercepts dirty edits, evaluates `talon guard`, and loops the agent with self-repair prompts until `ALLOW` | `Automated Self-Repair Convergence` |
| **10-Surface Software Intelligence Dashboard (`/dashboard` PWA)** | 🟢 **Production (GA)** | `cmd/garuda-api` & `garuda dev` | `Overview`, `Symbol Search`, `Architecture Map`, `Docs & Rules`, `Talon & Risk`, `Change Review`, `Evidence & Ledger`, `Agents & Sessions`, `Data & Logic Audit`, and `Decision Lineage` + `⌘K` palette + theme switch + responsive navigation + `/manifest.json` & `/sw.js` PWA | `Ed25519 JWT & Session Isolation` |

---

## 6. Plane 06 — Universal Zero-Dependency Runtime & Cross-OS Diagnostics (`v1.9.1`)

| Capability | Status | Verification Tier | Supported Semantics | Enforced Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **Windows PowerShell 5.1 & 7 Zero-Dependency Installer (`install.ps1` / `setup.ps1`)** | 🟢 **Production (GA)** | `Windows 10/11 PS 5.1 & PS 7 + Git Bash` | Pure 7-bit ASCII scripts (`CP1252`-safe), native `curl.exe`/`tar.exe` acceleration, stderr-safe `Invoke-QuietCmd`, Portable PostgreSQL 16.4 (`$HOME\.garuda\runtime\postgres`), and Portable Go `1.26.0` (`$HOME\.garuda\toolchain\go`) without Admin or GCC | `1-Command Windows Onboarding` |
| **macOS / Linux / Chromebook / Android Fast Installer (`install.sh` / `setup.sh`)** | 🟢 **Production (GA)** | `Darwin (ARM64/Intel) / Linux / Crostini / Termux` | Shallow clone (`--depth 1`), 3-second pre-built Portable PostgreSQL `16.4.0` (`~18 MB`, avoiding slow Homebrew source builds), single-download Portable Go `1.26.0` matching `go.mod`, and 1-pass `CGO_ENABLED=0` build with `bin/.build_stamp` caching | `1-Command Universal Onboarding` |
| **Explicit IPv4 `127.0.0.1:5433` & Socket Isolation (`env.go` & `init_cmd.go`)** | 🟢 **Production (GA)** | `macOS IPv6 [::1] Bypass + ~/.garuda/data/run` | Normalizes `@localhost:` to `@127.0.0.1:`, initializes clusters with `LC_ALL=C --no-locale -E UTF8`, and isolates Unix sockets in `~/.garuda/data/run` (`chmod 700`) | `Zero Connection-Refused Boot` |
| **Canonical Schema Hardening (`migrations/057_schema_hardening.sql`)** | 🟢 **Production (GA)** | `Idempotent PostgreSQL Migration` | Aligns `workspaces`, `repositories`, `entities` (`canonical_name`, `file_path`), `relationships`, and `cross_repo_edges` across fresh and upgraded databases | `100% Schema Parity` |
| **Self-Diagnosis & Auto-Repair CLI (`garuda doctor`)** | 🟢 **Production (GA)** | `cmd/garuda/doctor_cmd.go` (`--fix`, `--json`) | Auto-starts stopped user-space PostgreSQL clusters (`~/.garuda/data/pg`) or Docker, applies migrations, rebuilds missing binaries, and repairs `.env` & Talon hooks | `Self-Healing Local Runtime` |


---

## 7. Plane 07 — Deterministic `BuildContext` Provenance, Conditional Build-Tag AST & Docs-to-Code Lineage (`v1.7`–`v1.9.1`)

| Capability | Status | Verification Tier | Supported Semantics | Enforced Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **Polyglot `BuildContext` Provenance Receipts** | 🟢 **Production (GA)** | `internal/analyzer/build_context_test.go` | Captures Git SHA/dirty hash, `GOOS`, `GOARCH`, `CGO_ENABLED`, `-tags`, `go.mod` hash, Python lock hash, `tsconfig.json` hash, and 1-line `reproduce_command` | `Canonical Merkle Leaf Binding` |
| **`go/build.Context` Conditional Build-Tag & OS Filtering** | 🟢 **Production (GA)** | `TestTwoBuildsExposeDifferentEdges` | Evaluates `//go:build`, `// +build`, `_linux.go`, `_darwin.go`, `_windows.go`, and `cgo` constraints via `go/build.Context.MatchFile` | `Platform-Accurate Edge Graph` |
| **CI-vs-Editor Environment Drift Detector** | 🟢 **Production (GA)** | `TestDetectEnvironmentDrift` | Detects `TARGET_PLATFORM_MISMATCH`, `CGO_MISMATCH`, `BUILD_TAGS_MISMATCH`, `COMMIT_DRIFT`, `WORKTREE_DIRTY_DRIFT`, and `LANG_CONFIG_DRIFT` | `Zero Silent False-ALLOW` |
| **4-Profile Cross-Platform Build Matrix Gate (`--matrix`)** | 🟢 **Production (GA)** | `TestEvaluateBuildMatrix` | Evaluates `linux/amd64 (cgo=0)`, `linux/amd64 (cgo=1, tags=cgo)`, `darwin/arm64 (cgo=1)`, and `windows/amd64 (cgo=0)` in one pass | `Pre-Push Cross-OS Parity` |
| **Relative-Path Package Cache Determinism (`v1.8.1`)** | 🟢 **Production (GA)** | `TestDoubleRunDeterminism (24/24)` | Hashes repository-relative paths (`filepath.Rel`) + file sizes + SHA-256 digests so cache fingerprints are checkout-path invariant | `100% Deterministic Fingerprints` |
| **4-Tier Docs, Specs & Guardrails Lineage Graph** | 🟢 **Production (GA)** | `GET /api/v1/dashboard/docs-graph` & `garuda docs graph` | Connects Documents (ADRs/Specs/Policies) → Extracted Rules (`MUST`/`MUST_NOT`) → Active Guardrails → Governed AST Symbols | `Traceable Normative Lineage` |
| **Architectural Risk Hub & Blast-Radius Leaderboard** | 🟢 **Production (GA)** | `GET /api/v1/dashboard/talon-hub` | Ranks symbols `0–100` (`CRITICAL`/`HIGH`/`MEDIUM`/`LOW`) by caller fan-in, transitive blast radius, and ADR sensitivity | `Pre-Edit Risk Visibility` |


---

## 8. Plane 08 — Human-in-the-Loop Formula Disambiguation, Surface Criticality Tiers, Financial Data Audit & Auto-Rulebook (`v1.9.1`)

| Capability | Status | Verification Tier | Supported Semantics | Enforced Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **Multi-Team Formula Disambiguation (`0.0–1.0` Weights)** | 🟢 **Production (GA)** | `garuda audit logic` & `garuda.disambiguate_logic` | Computes `0.00–1.00` team weights (`0.70` active team vs `0.10`/`0.10`), ranks variants by Least Architectural Harm, and persists human decisions in `logic_variants_registry` | `Human-in-the-Loop Mandate (Zero Silent AI Overwrites)` |
| **3-Tier Surface Criticality Classification** | 🟢 **Production (GA)** | `garuda audit criticality` & `garuda.criticality_tiers` | Classifies surfaces into `HIGH_SEVERE` (`HUMAN_FREEZE_DO_NOT_TOUCH`), `MEDIUM` (`HUMAN_TRADE_OFF_REVIEW`), and `LOW` (`ALLOW_WITH_INFO`) | `Core Search & Settlement Freeze` |
| **Continuous Financial & Business Data Lineage Audit** | 🟢 **Production (GA)** | `garuda audit financial` & `garuda.audit_financial_discrepancy` | Traces AST & SQL aggregation formulas across teams to catch 10x unit-scaling discrepancies (`48.00 Cr` vs `4.80 Cr`) and omitted cessions pre-audit | `Pre-Audit Crore Normalization (SYNTH-AUDIT-005)` |
| **Auto-Synthesized Base Rulebook Studio** | 🟢 **Production (GA)** | `garuda rulebook generate/approve` & `garuda.approve_base_rulebook` | Synthesizes unorganized repo rules into `.garuda/DRAFT_BASE_RULEBOOK.md` and seals human-approved baseline into `.garuda/APPROVED_BASE_RULEBOOK.md` | `RFC 6962 SHA-256 Rulebook Seal` |
| **35-Tool Extended MCP Server (`bin/garuda-mcp`)** | 🟢 **Production (GA)** | `43/43 checks in scripts/mcp_verify.py` + `TestToolArgWhitelistCoversAllTools` | Exposes all 35 tools over stdio JSON-RPC 2.0 with strict argument whitelisting and first-call active contract injection | `100% CLI-to-MCP Parity` |
