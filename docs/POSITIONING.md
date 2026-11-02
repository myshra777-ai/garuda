# Garuda Positioning

**Internal & canonical reference.** This document is the canonical statement of what Garuda is, what it is not, and how its capabilities and evidence should be described. All public-facing material — `README.md`, `docs/CAPABILITIES.md`, `EVIDENCE.md`, release notes, and pitch materials — derives from this document and [`capabilities.json`](../capabilities.json). If a claim is not backed by [`capabilities.json`](../capabilities.json) and [`EVIDENCE.md`](../EVIDENCE.md), it is not a claim Garuda makes.

---

## The one-sentence pitch

Garuda is the distributable engineering governance and operational continuity platform that keeps AI coding agents, software architecture, and business logic aligned as software evolves.

## The shorter pitch

Garuda gives everyone working on a software system — developers, reviewers, CI/CD pipelines, and concurrent AI agents — one shared, verifiable model of what the software is, which rules govern it, and how work moves between agents.

## The problem

AI coding agents can now write, refactor, test, and ship code faster than any engineering organization can manually review it. That speed creates a structural failure mode we call **AI Development Drift**: the compounding gap between what an organization intended, what rules were written down, what code got built, and what the running system or cross-team pipelines actually do.

Four failure modes appear together as multi-agent autonomy increases:

1. **Agents operate without cross-session and cross-agent continuity.** When work moves from one coding agent or session to another, the receiving agent reconstructs context from scratch—missing prior checkpoints, active task boundaries, and changes made since the last briefing.
2. **Agents ship syntactically valid code that breaks architectural boundaries.** A diff can compile cleanly and pass unit tests while making an API handler call `database/sql` directly or modifying a high-blast-radius hub with 140+ callers.
3. **Teams and agents silently diverge on business formulas and unit scales.** Operations, Marketing, and Management often maintain competing undocumented formulas for the same metric (such as *Loss Ratio*) or mismatched unit scalings (`48.00 Crore` vs. `4.80 Crore` due to Lakh-to-Crore `/ 100` vs. `/ 1000` division).
4. **Intent, rules, enforcement, and code live in disconnected silos.** ADRs sit in Markdown files, policies sit in YAML, agent instructions sit in `.cursor` or `CLAUDE.md`, and the actual call graph sits in source code—with no single graph showing how organizational intent governs concrete symbols.

## The four product pillars

### 1. The Four-Tier Document-to-Code Governance Graph (`garuda.docs_graph` & Surface `04`)

Garuda connects four layers of a software organization into a single inspectable document-to-code governance graph:

| Governance Tier | Layer Name | What It Represents |
| :--- | :--- | :--- |
| **Tier 1** | **Organizational Intent** | Ingested ADRs (`docs/adr/*.md`), design specifications, system invariants, YAML policies (`policies/*.yaml`), and repository guardrails (`.garudarules`). |
| **Tier 2** | **Extracted Normative Rules** | Structured normative claims (`MUST`, `MUST_NOT`, `BLOCK`, `REVIEW`, `DENY`) extracted from documents with collision-free SHA-256 component IDs. |
| **Tier 3** | **Executable Governance** | Active declarative YAML policies, compiled **Garuda Talon** contracts (`.garuda/TALON_CONTRACT.md`), and pre-edit / post-edit gate verdicts. |
| **Tier 4** | **Actual Software (AST)** | Indexed Go, Python, and TypeScript functions, methods, structs, packages, and call/import relationships. |

**Honest qualification on link provenance:** Every document-to-code link in `garuda.docs_graph` and `/api/v1/dashboard/docs-graph` carries an explicit `match_method` (`"ast_exact"`, `"package_or_keyword_match"`, or `"fallback_unindexed_preview"`) and `graph_source_mode` (`"live_indexed_ast"` vs. `"fallback_preview_unindexed"`). We describe this as a **document-to-code governance graph** that combines indexed document claims, active policies, exact AST symbol matches, and keyword/curated rules—never claiming that heuristic or preview links are compiler-proven AST call edges.

### 2. Multi-Agent Coordination & Operational Continuity (`internal/store/handoff_store.go` & `cmd/garuda-mcp/`)

Garuda understands the work of agents, not just their code. It provides shared operational context for teams running multiple AI coding agents across the same evolving repository:

| Capability | Implementation Surface | What It Enables |
| :--- | :--- | :--- |
| **MCP Session Tracking** | `cmd/garuda-mcp/session_tracking.go` | Identifies connected MCP clients and agents, active session state, and session duration. |
| **Tool Execution Telemetry** | `cmd/garuda-mcp/telemetry.go` | Records tool names, execution latency, success/error rates, and whitelist-filtered arguments (`filterArgs`). |
| **Agent-Specific Briefings** | `garuda.briefing` / `garuda.preflight_bundle` | Tracks when an agent was last briefed and reports net-new workspace deltas since that watermark. |
| **Task Checkpoints** | `garuda.create_task` / `checkpoints` table | Persists structured execution context and scope metadata associated with a task UUID. |
| **Serializable Atomic Handoffs** | `garuda.handoff` (`ExecuteHandoffTransaction`) | Locks and validates source/target agents and task ownership inside a serializable PostgreSQL transaction, saves a checkpoint, transfers ownership, and fails closed on unknown task UUIDs. |
| **Checkpoint Resume** | `garuda.resume` | Restores saved checkpoint state for the receiving agent and marks the checkpoint consumed. |
| **Decision Lineage** | `garuda.propose_decision` & Surface `§` | Records MADR-compatible architectural and formula decisions with supersession chains and human co-sign status. |

### 3. Distributable Engineering Governance Platform (`action.yml`, GHCR & Multi-OS Release Matrix)

Garuda is designed to operate across the full software delivery lifecycle:

`Developer machine → AI coding agent (MCP / Talon hooks) → GitHub repository → CI/CD gate (GitHub Action) → Auditable RFC 6962 governance record.`

- **Composite GitHub Marketplace Action (`action.yml`)**: Builds from `${GITHUB_ACTION_PATH}`, evaluates `--diff "${{ inputs.diff }}"`, posts the deterministic `receipt_hash` and `reproduce_command` to `$GITHUB_STEP_SUMMARY`, and fails closed (`BLOCK`) on non-zero exit codes or unparsable output.
- **Multi-Arch GHCR Container (`Dockerfile` & `.github/workflows/release-packages.yml`)**: Packages all 4 static zero-CGO binaries (`garuda`, `garuda-api`, `garuda-mcp`, `garuda-talon`), `capabilities.json`, `version.txt`, `migrations/`, and `policies/` into `ghcr.io/myshra777-ai/garuda:latest` (and a tag per release).
- **Cross-Platform Zero-CGO Release Matrix (`scripts/build_release_matrix.sh`)**: Outputs release bundles into `dist/` for Windows (`amd64`, `arm64`), macOS (`arm64`, `amd64`), Linux (`amd64`, `arm64`, `.deb`), Android Termux, and FreeBSD.

### 4. Three-Tier Inspectable Evidence Standard (`EVIDENCE.md`)

Garuda strictly separates compiler fixtures, synthetic fault-injection simulations, and live repository / MCP verification so benchmark or demo data never pollutes production workspaces:

| Evidence Tier | Suite & Activation Flag | Default in Production | What It Proves & Honest Scope |
| :--- | :--- | :--- | :--- |
| **Tier 1 — Compiler & AST Fixtures** | `24/24` (`make benchmark-strict` / `include_fixtures=true`) | **Excluded by default** (`ExcludeFixturesSQL`) | Deterministic extraction precision/recall on controlled Go (`go/types`, `go.work`, `vendor/`), Python AST, and TypeScript AST fixtures, plus checkout-path-invariant cache fingerprints (`TestDoubleRunDeterminism`). |
| **Tier 2 — Synthetic Contradiction & Fault Simulations** | `10/10` (`scripts/runtime_contradiction_test.py`, `--demo-fixtures`) | **Opt-in only** (`INSUFFICIENT_DATA` when `include_fixtures=false`) | Quarantine of injected unapproved OTel service edges (`10/10`), Talon self-repair prompt synthesis under simulated `BLOCK`/drift, and reference case studies (`Loss Ratio` `0.0–1.0` weights & `48.00 Cr vs 4.80 Cr`). |
| **Tier 3A — Live PostgreSQL & MCP Protocol Verification** | `63/63` checks (`scripts/mcp_verify.py`) | **Active by default** | End-to-end JSON-RPC 2.0 stdio protocol execution against a live PostgreSQL instance, verifying active Talon contract metadata injection, fail-closed unknown-UUID handoffs, and RFC 6962 proofs. |
| **Tier 3B — 54-Tool Behavioral & Security Gate Regression Suite** | `54/54` tools / `108` subtests (`cmd/garuda-mcp/behavioral_tools_test.go`) | **Active in CI (`go test`)** | Positive and negative contract, schema, tenant-isolation, `GARUDA_MCP_READ_ONLY`, and `GARUDA_MCP_APPROVER` regression tests across all 54 registered MCP tools (runs store-free in unit CI and binds to PostgreSQL when `DATABASE_URL` is set). |

---

## Where Garuda is today

**Shipped in the repository:**
- **4 Zero-CGO Binaries**: `garuda` (CLI), `garuda-mcp` (54-tool MCP server), `garuda-api` (HTTP gateway & PWA server), and `garuda-talon` (active agent harness).
- **54 Registered MCP Tools**: 41 read-only tools and 13 mutating tools protected by strict argument whitelisting (`tools_whitelist.go`), fail-closed `GARUDA_MCP_READ_ONLY=true` mode, and `GARUDA_MCP_APPROVER` human co-sign attestation.
- **10-Surface Progressive-Disclosure Dashboard (`/dashboard`)**: Redesigned to display at most 4 primary cards on first view with collapsible `<details>` drawers, tabbed workbenches, and `cubic-bezier(0.16, 1, 0.3, 1)` motion in Light and Dark modes.
- **Multi-Language AST & Graph Engine**: Compiler-backed Go analyzer (`go/types`) and structural Python and TypeScript AST analyzers backed by 91 idempotent PostgreSQL migrations and an RFC 6962 Merkle ledger.
- **Human-in-the-Loop Data & Logic Audit Ledger (`095_data_logic_audit_ledger_and_hygiene.sql`)**: `0.0–1.0` team formula disambiguation, 3-tier surface criticality (`HIGH_SEVERE`, `MEDIUM`, `LOW`), and unit-scaling discrepancy detection (`data_logic_audit_ledger`).
- **Distribution & Packaging Pipeline**: Fail-closed `action.yml`, multi-binary `Dockerfile`, `scripts/build_release_matrix.sh` (`dist/*`), and `scripts/sync_docs_from_manifest.py` backed by `capabilities.json` and `version.txt`.

**Honest qualifications & ongoing priorities:**
- **Go vs. Python/TypeScript depth**: Go analysis uses full `go/types` compiler type resolution; Python and TypeScript analyzers perform structural AST extraction (imports, calls, class hierarchies, and unit-scaling divisors) rather than full cross-package type inference.
- **Docs & Rules Graph link provenance**: Combines exact AST matches (`ast_exact`) with package/keyword matching (`package_or_keyword_match`) and explicit preview fallbacks (`fallback_unindexed_preview`) when a workspace has not yet been indexed.
- **Behavioral test scope**: `behavioral_tools_test.go` (`54/54` tools) validates tool contracts, input validation, tenant isolation, and security gates across all 54 handlers, complemented by `scripts/mcp_verify.py` (`63/63` checks) against live PostgreSQL.
- **Commercial stage**: Open-source Apache-2.0 release; enterprise SSO, hosted multi-tenant cloud billing, and remote Merkle notary services remain future roadmap items.

## The one line to remember

> Garuda keeps AI coding agents, software architecture, and organizational intent aligned as your system evolves.