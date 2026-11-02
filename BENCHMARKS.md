# Garuda — Architecture, Agent Continuity, Distribution & Three-Tier Benchmark Report

> Benchmark figures and the release-blocker matrix below were recorded on the `v1.9.9` line. `v1.9.10` changed installers, documentation, packaging and CI, not the analyzers (see [`CHANGELOG.md`](CHANGELOG.md)).

Garuda has evolved from a local developer analyzer into a **distributable engineering governance and multi-agent continuity platform**. All counts, versions, and verification gates in this report are synchronized with [`capabilities.json`](capabilities.json), [`internal/version/manifest.go`](internal/version/manifest.go), and [`EVIDENCE.md`](EVIDENCE.md) via `python3 scripts/sync_docs_from_manifest.py --check`.

---

## 1. The Four-Tier Document-to-Code Governance Graph (`garuda.docs_graph` & Surface `04`)

Garuda connects four distinct layers of a software organization into a single inspectable document-to-code governance graph:

| Tier | Layer | Artifacts & Entities Connected |
| :--- | :--- | :--- |
| **Tier 1** | **Organizational Intent** | ADRs (`docs/adr/*.md`), design specifications (`docs/specs/*.md`), invariants, declarative YAML policies (`policies/*.yaml`), and repository guardrails (`.garudarules`). |
| **Tier 2** | **Extracted Rules** | Normative requirements (`MUST`, `MUST_NOT`, `BLOCK`, `REVIEW`, `DENY`) with collision-free SHA-256 component identifiers. |
| **Tier 3** | **Executable Governance** | Active YAML policy predicates, compiled **Garuda Talon** contracts (`.garuda/TALON_CONTRACT.md`), and enforcement decisions (`ALLOW`, `WARN`, `REVIEW`, `BLOCK`). |
| **Tier 4** | **Actual Software (AST)** | Indexed Go (`go/types`), Python AST, and TypeScript AST functions, methods, structs, packages, and call/import edges. |

### Explicit Link-Provenance Qualification
Every document-to-code link returned by `/api/v1/dashboard/docs-graph` and MCP tool `garuda.docs_graph` includes explicit provenance metadata:
- `graph_source_mode`: `"live_indexed_ast"` when the workspace has indexed AST entities in PostgreSQL, or `"fallback_preview_unindexed"` / `"policy_and_claims_unindexed"` when operating on an unindexed workspace.
- `match_method`: Distinguishes `"ast_exact"` (exact symbol/package resolution) from `"package_or_keyword_match"` (heuristic keyword/package matching), `"document_claim_extracted"`, `"policy_enforced_entrypoint"`, and `"fallback_unindexed_preview"`.

---

## 2. Multi-Agent Coordination & Operational Continuity Layer

Garuda understands the work of agents, not just their code—preserving shared operational context across concurrent AI coding agents working on the same evolving codebase:

| Capability | Implementation Surface | What It Enables |
| :--- | :--- | :--- |
| **MCP Session Tracking** | `cmd/garuda-mcp/session_tracking.go` | Identifies connected MCP clients and agents, their active status, and session duration. |
| **Tool Execution Telemetry** | `cmd/garuda-mcp/telemetry.go` | Records tool names, execution duration, success/error counts, and whitelist-filtered arguments (`filterArgs`). |
| **Agent-Specific Briefings** | `garuda.briefing` / `garuda.preflight_bundle` | Tracks when an agent was last briefed and reports workspace changes since that watermark. |
| **Task Checkpoints** | `garuda.create_task` / `checkpoints` table | Persists structured context and scope metadata associated with a task UUID. |
| **Atomic Handoffs** | `garuda.handoff` (`internal/store/handoff_store.go`) | Locks and validates agents and task ownership inside a serializable PostgreSQL transaction (`ExecuteHandoffTransaction`), saves a checkpoint, transfers ownership, updates agent states, and fails closed (`status: "failed"`) on unknown task UUIDs. |
| **Resume** | `garuda.resume` | Allows a receiving agent to restore saved checkpoint data and marks the checkpoint consumed. |
| **Decision Lineage** | `garuda.propose_decision` & Surface `§` | Makes recorded architectural and formula decisions and their supersession relationships inspectable. |

---

## 3. Engineering Governance Distribution Platform

Garuda provides an end-to-end adoption and enforcement path:

`Developer machine → coding agent → GitHub repository → CI/CD gate → auditable governance record.`

| Delivery Component | Path | Behavior |
| :--- | :--- | :--- |
| **Composite GitHub Marketplace Action** | [`action.yml`](action.yml) | Builds zero-CGO binaries directly from `${GITHUB_ACTION_PATH}`, passes `--diff "${{ inputs.diff }}"` to `garuda talon guard`, captures exit codes without swallowing errors, and fails closed (`BLOCK`) if JSON output is unparsable or the gate exits non-zero. |
| **Draft GitHub Release** | [`.github/workflows/release.yml`](.github/workflows/release.yml) | On a `vX.Y.Z` tag: checks `version.txt` equals the tag, runs the documentation/manifest consistency and link checks and `go build/vet/test`, runs `scripts/build_release_matrix.sh`, attests build provenance and creates a **draft** release with `dist/releases/*` and `SHA256SUMS.txt`. |
| **GHCR Container Images** | [`.github/workflows/release-packages.yml`](.github/workflows/release-packages.yml), [`publish-mcp-image.yml`](.github/workflows/publish-mcp-image.yml) & [`Dockerfile`](Dockerfile) | Runs the consistency check and the MCP tool-contract tests, then pushes the multi-arch API image `ghcr.io/myshra777-ai/garuda:<version>` (all 4 static binaries: `garuda`, `garuda-api`, `garuda-mcp`, `garuda-talon`) and the MCP server image `ghcr.io/myshra777-ai/garuda-mcp:<version>`. |
| **Multi-OS Zero-CGO Release Matrix Builder** | [`scripts/build_release_matrix.sh`](scripts/build_release_matrix.sh) | Cross-compiles Windows (`amd64`, `arm64`), macOS (`arm64`, `amd64`), Linux (`amd64`, `arm64`, `.deb`), Android Termux, and FreeBSD packages directly into `dist/releases/` with `SHA256SUMS.txt`. |

---

## 4. Three-Tier Inspectable Evidence Standard

Garuda strictly separates deterministic compiler fixtures, synthetic fault-injection simulations, and live repository / MCP verification:

| Evidence Tier | Reported Score | Scope & Methodology |
| :--- | :--- | :--- |
| **Tier 1 — Deterministic Compiler & AST Fixtures** | **`24/24`** | Controlled Go (`go/types`, `go.work`, `vendor/`), Python AST, and TypeScript AST ground-truth fixtures (`make benchmark-strict` / `include_fixtures=true`). Excluded by default from live workspace queries via `ExcludeFixturesSQL`. |
| **Tier 2 — Synthetic Runtime Contradiction & Fault-Injection Simulations** | **`10/10`** | Controlled OTel span contradiction injections (`scripts/runtime_contradiction_test.py`) and opt-in reference simulations (`--demo-fixtures`, `simulate_violation`, `simulate_drift`). Never injected into live workspace queries unless explicitly requested. |
| **Tier 3A — Live Repository & MCP Protocol Verification** | **`63/63`** | End-to-end JSON-RPC 2.0 stdio protocol checks (`scripts/mcp_verify.py`) running against a live PostgreSQL instance, verifying active Talon contract metadata injection, fail-closed unknown-UUID handoffs, and RFC 6962 Merkle proofs. |
| **Tier 3B — 54-Tool Behavioral & Security Gate Regression Suite** | **`54/54` tools** (`108` subtests) | Positive and negative invocation cases for every registered MCP tool (`cmd/garuda-mcp/behavioral_tools_test.go`), verifying tool schema dispatch, argument whitelisting, `GARUDA_MCP_READ_ONLY=true` mutation blocking across all 13 mutating tools, `GARUDA_MCP_APPROVER` human co-sign enforcement, and tenant/workspace UUID isolation (runs store-free in unit CI and binds to live PostgreSQL when `DATABASE_URL` is set). |

---

## 5. Release-Blocker Resolution Matrix (`v1.9.9`)

| Audit Finding | Severity | Resolution Applied & Verified |
| :--- | :--- | :--- |
| **1. `action.yml` working directory, missing `--diff` flag, and fail-open `\|\| true` / `ALLOW` fallback** | **Critical** | Updated `action.yml` to compile from `${GITHUB_ACTION_PATH}`, pass `--diff "${{ inputs.diff }}"`, capture `GUARD_EXIT` without `\|\| true`, and default any unparsable or non-zero exit to `BLOCK` (`exit 2`). |
| **2. Canonical manifest (`capabilities.json`), `version.txt`, and version drift (`v1.9.1` / `v1.9.2` vs `v1.9.9`)** | **Critical** | Bundled `capabilities.json` and `version.txt` (`v1.9.9`) in the repository root and release archives; synchronized `Dockerfile`, `build/Dockerfile.release`, `scripts/build_release_matrix.sh`, and `setup.sh` to `v1.9.9`; added automated version & file checks to `scripts/sync_docs_from_manifest.py --check`. |
| **3. Release artifact path mismatch (`dist/` vs `dist/releases/*`) & single-binary `Dockerfile`** | **High** | Updated `scripts/build_release_matrix.sh` default `OUT_DIR` to `dist/releases` (matching `.github/workflows/release-packages.yml`) and updated `Dockerfile` + `build/Dockerfile.release` to compile and package all 4 static binaries (`garuda`, `garuda-api`, `garuda-mcp`, `garuda-talon`). |
| **4. Outdated `docs/POSITIONING.md` (16-tool claim & stale beta status)** | **Medium** | Rewrote `docs/POSITIONING.md` for `v1.9.9` (`54` MCP tools, `10` progressive-disclosure surfaces, 4-tier Docs & Rules Graph with provenance qualification, multi-agent continuity layer, and 3-tier evidence separation) and added automated drift validation in `scripts/sync_docs_from_manifest.py --check`. |
