# Garuda Analyzer Capabilities & Product Snapshot

> **Document scope:** This document has three distinct purposes: it defines
> the analyzer capabilities and maturity levels, records a dated
> self-snapshot of the Garuda repository used for verification, and
> tracks product-level capabilities shipped and upcoming. The
> repository metrics below are not universal product limits or
> benchmark claims.
>
> **Generated:** 2026-09-19
> **Grounding:** AST snapshot, benchmark gates, runtime verification
> **Analyzer scope:** Go AST parsing, type checking, and relationship resolution
> **Product scope:** MCP surface, dashboard, control plane, trust layer, policy engine

---

## Verification Status Legend

| Status | Meaning |
|---|---|
| 🟢 **Production (GA)** | Validated for production workflows and covered by the stated benchmark gate |
| 🟡 **Beta** | Functionally verified against the AST, workspace model, or runtime path but still subject to release or deployment constraints |
| 🟣 **Experimental (Heuristic)** | Available for exploration; results are confidence-scored and must not be treated as authoritative without additional verification |
| 🔵 **Upcoming** | Designed and scheduled for the next arc. Not yet shipped. Listed so the document does not trail the roadmap. |

---

## Garuda Self-Snapshot

The following counts describe the Garuda repository represented by the dated AST snapshot. They are extraction metrics for that snapshot, not general claims about every repository analyzed by Garuda.

| Metric | Count | Interpretation |
|---|---:|---|
| **Parsed files** | `289` | Source files included in the AST snapshot |
| **Packages** | `98` | Packages discovered during extraction |
| **Discovered structs** | `474` | Struct declarations identified by the analyzer |
| **Discovered interfaces** | `27` | Interface declarations identified by the analyzer |
| **Functions and methods** | `1,245` | Functions and methods identified by the analyzer |

**Struct-field metric:** The supplied snapshot reported `Total struct fields: 0`, but the field count is not considered measured or reliable for this snapshot. It is intentionally excluded from the metrics table rather than presented as a valid zero.

---

## AST Capabilities Matrix

*Go AST parsing, type checking, and relationship resolution.*

| Capability | Status | Verification tier | Supported semantics | Governing invariant | Operational interpretation |
|---|---|---|---|---|---|
| **Struct and field extraction** | 🟢 **Production (GA)** | `V5` | Struct declarations, field metadata, `json`/`db`/`validate` tags, source line spans, pointer and slice flags | `ACGM Resolution Invariant` | Suitable for production semantic indexing and downstream governance queries |
| **Method receiver disambiguation** | 🟢 **Production (GA)** | `002-method-identity` | Pointer receivers, value receivers, exported methods, unexported methods | `Canonical UUIDv5` | Distinguishes methods with receiver-aware identity resolution |
| **Interface implementation matching** | 🟢 **Production (GA)** | `003`, `009`, `010` | Dynamic method-set satisfaction and multi-interface polymorphism | `Full Recall Gate` | Resolves implementation relationships across supported method sets |
| **Generics and type parameters** | 🟢 **Production (GA)** | `004-generics` | Type-parameter filtering and generic-container extraction | `Zero False Reference Claims` | Prevents generic syntax from producing unsupported or spurious reference claims |
| **Type aliases versus definitions** | 🟢 **Production (GA)** | `005-alias` | Distinguishes `type X = Y` aliases from `type X Y` definitions | `Kind Disambiguation` | Preserves semantic identity and declaration kind |
| **Struct embedding** | 🟢 **Production (GA)** | `006-embedding` | Anonymous fields and `EMBEDS` relationship emission | `Composition Invariant` | Represents composition and promoted members as explicit semantic relationships |
| **Variadic and closure handling** | 🟢 **Production (GA)** | `007`, `008` | Ellipsis signatures and nested function-literal traversal | `Scope Resolution` | Maintains correct lexical and callable scope through nested constructs |
| **Cross-repository dependency resolution** | 🟡 **Beta (AST Verified)** | `workspace-cache` | Multi-module Go import mapping across repositories and workspaces | `Tenant Isolation` | Use for workspace-scale dependency analysis; validate repository and workspace boundaries in deployment |
| **Dynamic call-graph tracing** | 🟣 **Experimental (Heuristic)** | `heuristic` | SSA call-site resolution | `Epistemic Confidence < 0.8` | Treat results as confidence-scored hypotheses rather than authoritative call relationships |

### Benchmark identifier convention

The **Verification tier** column uses benchmark identifiers only. Human-readable descriptions belong in the capability semantics and operational-interpretation columns; this keeps the matrix traceable and consistent.

---

## Product Capabilities Matrix

*Product-level capabilities grouped by surface. Analyzer-specific capabilities are in the matrix above.*

### Semantic Model

| Capability | Status | Verified by | Notes |
|---|---|---|---|
| **Go extraction** | 🟢 **Production (GA)** | `go-validation-10` reference workspace | Compiler-backed. Tier 5. Precision reported at 100% for supported relationship classes. |
| **Python extraction** | 🟡 **Beta** | `sqlmodel` reference repository | Structural AST. Tier 2. Imports reliable; calls heuristic. |
| **TypeScript extraction** | 🟡 **Beta** | `trpc` reference repository | Structural AST. Tier 2. Requires CGO and a C compiler. |
| **Rust extraction** | 🔵 **Upcoming** | — | Next language on the same pipeline. No schema change required. |
| **Java extraction** | 🔵 **Upcoming** | — | Queued after Rust. |
| **Elixir extraction** | 🔵 **Upcoming** | — | Queued. |

### MCP Surface

| Capability | Status | Verified by | Notes |
|---|---|---|---|
| **21 MCP tools** | 🟢 **Production (GA)** | `scripts/mcp_verify.py` 38/38 | Verified against four independent clients: Cursor, Claude Desktop, Codex CLI, and a reference implementation with no dependencies. |
| **Session tracking** | 🟢 **Production (GA)** | `mcp_sessions` table | One row per MCP process lifetime. Graceful shutdown writes `closed_at`. |
| **Tool call tracking** | 🟢 **Production (GA)** | `mcp_tool_calls` table | Per-tool duration, status, and whitelisted args. Free-form text is never recorded. |
| **Stale session cleanup** | 🔵 **Upcoming** | — | Spec calls for a job that marks crashed sessions `close_reason='stale'`. Schema and index exist; the job does not. |

### Trust Layer

| Capability | Status | Verified by | Notes |
|---|---|---|---|
| **RFC 6962 Merkle tree** | 🟢 **Production (GA)** | `garuda policy verify <id>` on live anchors | Domain-separated leaf and internal hashes. Golden vectors frozen in testdata. |
| **Inclusion proofs** | 🟢 **Production (GA)** | `VerifyInclusion` and `VerifyV1Proof` | Pure functions. No store, no DB, no network. A third party with the proof and root can verify without trusting Garuda. |
| **Policy evaluation anchoring** | 🟢 **Production (GA)** | `policy_evaluations` table + verify path | Every committed evaluation is anchored to a specific block. `--save` required; preview mode never anchors. |
| **Decision chain verification** | 🟢 **Production (GA)** | `garuda verify` | Per-decision revision chains walked independently. 72 revisions across 17 decisions verified. |
| **Anchor verification for decisions** | 🟡 **Beta** | `internal/policy/anchor.go` | Currently filters by `id` only. Tracked as B23. |

### Policy Engine

| Capability | Status | Verified by | Notes |
|---|---|---|---|
| **Five predicates** | 🟢 **Production (GA)** | `entity_exists`, `claim_exists`, `contradiction_exists`, `verification_missing`, `language_matches` | Declarative YAML. Deterministic evaluation. |
| **Four decision outcomes** | 🟢 **Production (GA)** | `ALLOW`, `WARN`, `REVIEW`, `BLOCK` | Highest severity across evaluations determines the final decision. |
| **Preview vs committed mode** | 🟢 **Production (GA)** | `policy evaluate` without `--save` | Preview mode makes no writes, leaves no Merkle trace. Committed mode requires `--save`. |
| **Directory reconcile** | 🟢 **Production (GA)** | `policy evaluate --reconcile --save` | Marks active policies whose YAML file is absent as superseded. Opt-in by design. |
| **Unique constraint on (tenant_id, statement)** | 🟢 **Production (GA)** | Migration 087 | Concurrent insert cannot duplicate. |

### Multi-Agent Coordination

| Capability | Status | Verified by | Notes |
|---|---|---|---|
| **Checkpoint / resume** | 🟢 **Production (GA)** | `TestHandoffAndResume` | Serializable isolation. A second resume of the same checkpoint returns `not_found`. |
| **Two-agent handoff** | 🟢 **Production (GA)** | Go test, commit `c2e3164` | Seven assertions covering state transitions on both agents, checkpoint lifecycle, and resume rejection. |
| **Cross-agent shared memory** | 🔵 **Upcoming** | — | Named as an RFS capability gap. No shared pool today; watermarks are per-agent. |
| **Memory compression** | 🔵 **Upcoming** | — | Named as an RFS capability gap. Raw data exists; no summarization layer. |

### User Dashboard

*Tenant-facing. `/dashboard`. Session cookie auth.*

| Tab | Status | Notes |
|---|---|---|
| **Workspace** | 🟢 **Production (GA)** | KPIs, language breakdown, scanned repositories, architectural hubs, recent evidence. |
| **Agents** | 🟢 **Production (GA)** | MCP surface catalog, active sessions, recent sessions, session detail. |
| **Governance** | 🟢 **Production (GA)** | Policy enforcement, documentation claims, knowledge drift, contradictions, Merkle trust. |
| **Decisions** | 🟢 **Production (GA)** | Decision list, detail drawer with ancestors/children, revision chain with genesis marker. |
| **Graph** | 🟢 **Production (GA)** | Interactive topology, communities sidebar, level switcher. |

### Control Plane

*Owner-facing. `/_/control`. Bearer token auth. Reads from `garuda_control_ro`, a read-only database role.*

| Tab | Status | Notes |
|---|---|---|
| **Auth + audit log** | 🟢 **Production (GA)** | 404-not-401 on unauthenticated. Every auth event and mutation writes one row to `control_plane_access`. Reads do not. |
| **Read-only role** | 🟢 **Production (GA)** | `garuda_control_ro` has `SELECT` only. Migration 089. INSERT through the role fails with a permission error. |
| **Business** | 🟢 **Production (GA)** | Growth, adoption, and usage metrics. Every metric has a formula on hover. Three-state empty handling. |
| **Tenants** | 🟢 **Production (GA)** | Cross-tenant list with derived health. Red first. Sort by color, then last activity. |
| **Operations** | 🔵 **Upcoming** | Next arc. Four panels have data sources today; three require new infrastructure. |
| **Internal** | 🔵 **Upcoming** | Queued after Operations. Roadmap view, feature flags, beta tester list. |

### Observability (next arc)

These three infrastructure pieces unblock the Operations tab's locked panels.

| Capability | Status | Notes |
|---|---|---|
| **`errors_log` table** | 🔵 **Upcoming** | Migration + logger wiring. Blocked on Operations tab. |
| **Query latency ring buffer** | 🔵 **Upcoming** | In-process P50/P95/P99 over a rolling window. |
| **Request error counter** | 🔵 **Upcoming** | In-process counter for the Operations error rate panel. |

### Quality & Lint

| Capability | Status | Verified by | Notes |
|---|---|---|---|
| **SQL scoping lint rule** | 🟢 **Production (GA)** | `tools/lint/` (5 tests) | Flags workspace-scoped queries missing a `workspace_id` filter. |
| **DTO completeness lint rule** | 🔵 **Upcoming** | — | Broaden the lint rule to catch store-vs-DTO field drift, which the current SQL check does not. |

---

## Verification Policy

The status in these matrices describes capability maturity, not merely implementation presence:

- **Production (GA)** capabilities may support authoritative semantic and governance workflows within their documented scope.
- **Beta** capabilities are usable but should be monitored for workspace, repository, cache, and tenant-boundary behavior.
- **Experimental** capabilities must remain explicitly confidence-aware and should not silently override compiler-backed or persisted evidence.
- **Upcoming** capabilities are scheduled. They must not be presented as shipped until the arc that delivers them closes with a verification.

For governance decisions, authoritative relationships should be derived from validated AST/type information and persisted evidence. Heuristic relationships should be labeled separately and excluded from strict compliance judgments unless an explicit policy permits them.

### Evidence authority order

```text
Compiler-backed AST/type evidence
              ↓
Persisted semantic relationships with source evidence
              ↓
Validated documentation claims
              ↓
Confidence-scored heuristic relationships

Lower-tier evidence must not silently replace higher-tier evidence. When evidence conflicts or required verification is unavailable, the consuming workflow should preserve the uncertainty and apply its configured fail-closed behavior.

Evidence and Release Gates
Evidence class	Required interpretation
AST snapshot metrics	Descriptive extraction coverage for the generated self-snapshot
Benchmark gate	Capability-specific correctness evidence for the named benchmark identifier
Workspace cache verification	Evidence for cross-repository resolution behavior in the configured PostgreSQL workspace
Heuristic evaluation	Confidence-scored output requiring independent validation
Runtime verification	Direct observation of the capability in the running system. Required before promoting a status from Upcoming or Beta to Production.
A capability should only be promoted to a higher maturity tier after its benchmark, persistence, isolation, and failure-path behavior have been validated for the target release.

Snapshot Integrity Checklist
Before publishing a new matrix, verify:

Parsed-file, package, struct, interface, and function counts match the AST snapshot.

Struct-field totals are either reconciled against emitted field entities or explicitly marked as unmeasured.

Benchmark identifiers and pass rates match the current release evidence.

Cross-repository results are checked against the intended workspace and tenant scope.

Heuristic outputs include confidence values and are not represented as authoritative relationships.

Generated timestamp, release commit, analyzer version, and database snapshot identity are recorded.

Self-snapshot metrics are clearly separated from product capability claims.

Every Upcoming entry has a named arc or a queued item in docs/TRACKING.md.

No Upcoming entry is described in the present tense anywhere in the document.

Aether Governance Alignment
Aether is the internal governance framework used here to describe authority, tenant isolation, integrity, state, and evidence-handling requirements. This section is intended for internal readers; external versions may omit it or replace it with a product-neutral governance statement.

The matrices distinguish persisted, compiler-backed evidence from heuristic inference. Storage presence alone does not establish visibility or authority; semantic exposure should remain dependent on the applicable registry, verification, approval, and revocation conditions. Cross-tenant resolution must remain structurally unavailable, and any integrity or audit failure must fail closed.