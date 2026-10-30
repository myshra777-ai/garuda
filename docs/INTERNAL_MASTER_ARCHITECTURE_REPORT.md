# Garuda Comprehensive System Architecture, Epistemic Governance Engine, and Engineering Reference Manual

## 1. Executive System Identity and Core Architectural Mandate

### 1.1 The Autonomous Code Generation Crisis Requires a Deterministic Verification Substrate

Software engineering has undergone a structural phase transition driven by autonomous and semi-autonomous coding assistants, including [Cursor](https://www.cursor.com/), [Anthropic Claude Code](https://docs.anthropic.com/en/docs/agents-and-tools/claude-code/overview), [OpenAI Codex](https://openai.com/index/openai-codex/), and [Windsurf](https://codeium.com/windsurf). While code generation throughput has accelerated by an order of magnitude, verification infrastructure has remained anchored to legacy paradigms designed for human typing speeds. Traditional static linters validate localized syntax and formatting rules, whereas unit test suites evaluate isolated function inputs and outputs—frequently using mock objects synthesized by the very same language model that authored the implementation. Neither mechanism enforces macro-architectural invariants, cross-service dependency boundaries, multi-team business formula consistency, or statutory financial data lineage.

[Garuda](https://github.com/myshra777-ai/garuda) (`github.com/myshra777-ai/garuda`, release version `v1.9.2`) addresses this structural deficit by operating as a local-first, PostgreSQL-backed software intelligence instrument, change-control gate, active AI agent execution harness (`Garuda Talon`), human-in-the-loop business formula disambiguator, and cryptographic audit ledger. Rather than competing with generative models to produce source code, Garuda functions as the deterministic referee that governs what code is permitted to enter a production repository. Every architectural claim, dependency edge, runtime trace, policy evaluation, and human stakeholder approval is anchored in a relational semantic graph and sealed inside an [RFC 6962 Certificate Transparency](https://datatracker.ietf.org/doc/html/rfc6962) Merkle tree.

### 1.2 Four Failure Modes Eliminated by the Garuda Control Plane

Enterprise engineering organizations deploying autonomous coding agents encounter four recurring classes of catastrophic failure that pass conventional continuous integration pipelines undetected. First, agents introduce silent architectural layer bypasses: to satisfy a local prompt quickly, an agent modifies an HTTP transport handler to invoke [`database/sql`](https://pkg.go.dev/database/sql) or an external payment client directly, bypassing the domain idempotency store mandated by internal Architecture Decision Records (`ADR-0001`). Unit tests pass because the local database state updates as expected, yet production idempotency guarantees are silently destroyed.

Second, growing codebases accumulate undocumented competing business formulas across organizational silos. In insurance, fintech, and e-commerce platforms, Operations, Marketing, and Executive Management teams routinely implement distinct mathematical definitions for shared business concepts such as `Loss Ratio` or `Net Revenue`. When an autonomous agent is tasked with refactoring a shared pricing or search routing module, it lacks the organizational context to distinguish between a localized campaign metric and a core solvency gate. Overwriting a core pricing formula with a promotional variant can break downstream callers across the entire customer-facing application.

Third, cross-team data pipelines suffer from silent unit-scaling and decimal drift that remain invisible until statutory financial audits. A canonical failure pattern occurs when a Claims Operations pipeline aggregates settled claims in raw Indian Rupees (`SUM(settled_claims_inr) / 1e7`) and accurately reports `48.00 Crore INR`, while a Finance executive reporting export reads an intermediate monthly rollup table denominated in Lakhs (`settled_claims_lakhs`) and mistakenly divides by `1000.0` rather than `100.0`, reporting `4.80 Crore INR`. Because both SQL queries and Go functions are syntactically valid and pass unit tests, this tenfold (`10.0x`, or `43.20 Crore INR`) discrepancy survives until external auditors flag the contradiction.

Fourth, most commercial software repositories lack formal, machine-readable documentation of their architectural rules. Expecting engineering leads to author hundreds of lines of declarative YAML policies or Markdown ADRs from scratch creates an insurmountable onboarding barrier. Garuda eliminates all four failure modes through compiler-grade static analysis, pre-commit agent interception, dynamic `0.00` to `1.00` team formula weighting, AST-to-SQL financial lineage tracing, and automated base rulebook synthesis.

| Enterprise Failure Class | Legacy Linter and Unit Test Behavior | Garuda Deterministic Governance Mechanism | Enforced System Invariant |
| :--- | :--- | :--- | :--- |
| **Silent Architectural Layer Bypass** | Passes syntax checks and mocked unit tests despite HTTP handlers calling `database/sql` directly. | Evaluates `layer_violation` and `transitive_claim_exists` predicates over the resolved AST call graph in `<20 ms`. | `ACGM Resolution Invariant` and `Fail-Closed Policy Gate` |
| **Competing Undocumented Business Formulas** | Allows AI agents to overwrite core pricing gates with localized marketing or ops formulas. | Computes `0.00–1.00` team relevance weights, ranks variants by Least Architectural Harm, and requires human sign-off. | `Human-in-the-Loop Mandate (Zero Silent AI Overwrites)` |
| **10x Financial Unit-Scaling Drift (48 Cr vs. 4.8 Cr)** | Ignores semantic divisor mismatches (`/ 1000.0` vs. `/ 100.0` on Lakhs-to-Crore conversions). | Traces AST and SQL aggregation lineage across team boundaries and quarantines `SCALE_FACTOR_10X_MISMATCH` pre-audit. | `Pre-Audit Crore Normalization (SYNTH-AUDIT-005)` |
| **Undocumented Legacy Repository Structure** | Provides zero protection until engineers manually write custom lint rules or ADRs. | Reverse-engineers call graphs, fan-in hubs, and layer boundaries into `.garuda/DRAFT_BASE_RULEBOOK.md` for 1-click approval. | `RFC 6962 SHA-256 Rulebook Seal` |
| **Cross-Platform Build-Tag Blindness** | Analyzes only the host OS, missing violations hidden inside `_windows.go` or `//go:build legacy_bypass` files. | Executes a 4-profile cross-platform build matrix (`--matrix`) and binds cryptographic `BuildContext` provenance receipts. | `Pre-Push Cross-OS Parity` and `Zero Silent False-ALLOW` |

---

## 2. Organizational Stakeholders, Operational Personas, and Value Delivery

### 2.1 Multi-Persona Operational Alignment Across Technical and Executive Workflows

Garuda is architected to serve seven distinct operational personas across the software delivery, architectural governance, and financial compliance lifecycle. Rather than confining software intelligence to command-line compiler output, the platform translates deep Abstract Syntax Tree (AST) topology and relational graph queries into role-specific interfaces across its command-line binary (`bin/garuda`), its 35-tool [Model Context Protocol](https://modelcontextprotocol.io/specification/2025-06-18) server (`bin/garuda-mcp`), and its 9-surface Progressive Web Application (`/dashboard`).

For **Software Engineers and Full-Stack Developers**, the primary operational friction in large repositories is getting penalized for pre-existing technical debt or inadvertently breaking an unfamiliar upstream service during a routine change. Garuda resolves this through diff-scoped violation partitioning (`DiffScopeFilter`) and pre-edit impact briefings (`garuda talon brief`). When an engineer modifies a file, Garuda maps the exact `git diff --unified=0` line hunks to touched AST symbols and separates `net_new_violations` from `baseline_violations`. Consequently, a developer working in a legacy repository with 25 baseline architectural violations can merge clean code immediately as long as `net_new_violations == 0`.

For **Staff and Principal Architects**, Garuda replaces manual wiki maintenance with executable governance. Through the Auto-Synthesized Base Rulebook Studio (`04 — Docs & Base Rulebook`), architects can ingest existing Markdown ADRs (`garuda docs ingest`) or instruct Garuda to synthesize an initial baseline rulebook from an unorganized repository (`garuda rulebook generate`). The system constructs a 4-tier lineage graph connecting Tier 1 Normative Documents, Tier 2 Extracted Claims (`MUST` and `MUST_NOT`), Tier 3 Active YAML Policy Guardrails, and Tier 4 Governed Code Symbols.

For **Finance Controllers, Chief Financial Officers (CFOs), Actuaries, and Statutory Auditors**, Garuda provides continuous pre-audit assurance over code-generated financial figures. Using Surface `09 — Financial & Logic Audit` and the `garuda audit financial` CLI, financial stakeholders inspect side-by-side lineage comparisons between operational ledgers and executive reporting exports, reconcile unit-scaling anomalies (`48.00 Crore INR` versus `4.80 Crore INR`), govern competing actuarial or revenue formulas via `0.00–1.00` weights, and verify tamper-evident RFC 6962 cryptographic inclusion proofs (`garuda verify`).

For **Non-Technical Founders, Product Managers, and Engineering Leads**, Garuda translates technical call graphs into business risk tiers using the Amazon Search versus Wishlist criticality model. High-fan-in customer pathways—such as the Homepage, Search Bar, Search Results Redirect Router (142 callers), and Claims Settlement Ledger—are classified as `HIGH_SEVERE` (`HUMAN_FREEZE_DO_NOT_TOUCH`), ensuring that automated refactors never take down revenue-critical flows. Secondary features such as the Wishlist Loader (19 callers) are classified as `MEDIUM` (`HUMAN_TRADE_OFF_REVIEW`), while cosmetic UI helpers (4 callers) are marked `LOW` (`ALLOW_WITH_INFO`).

For **DevOps, Platform Engineers, and Site Reliability Engineers (SREs)**, Garuda eliminates environment drift between developer laptops and continuous integration runners (`DetectEnvironmentDrift`) and correlates live [OpenTelemetry](https://opentelemetry.io/docs/specs/otel/) production spans (`POST /api/v1/telemetry/spans`) against the static call graph to quarantine unapproved shadow dependencies (`CONTRADICTED`). Finally, for **Autonomous AI Coding Agents** (`Cursor`, `Claude Code`, `Codex`, `Windsurf`) and **Technical Due-Diligence Evaluators**, Garuda provides a deterministic 35-tool JSON-RPC 2.0 MCP surface and a reproducible benchmark suite (`24/24` compiler benchmarks, `43/43` MCP protocol checks, and `10/10` runtime contradiction quarantines).

| Stakeholder Persona | Core Operational Objective | Primary Garuda Interface Surface | Key CLI Commands and MCP Primitives | Enforced Governance Outcome |
| :--- | :--- | :--- | :--- | :--- |
| **Software Engineer / Full-Stack Developer** | Prevent blind regressions and isolate personal diff changes from legacy technical debt. | `06 — Change Review Workbench` and `03 — Architecture Map` | `garuda talon brief`, `garuda talon guard --diff HEAD`, `garuda.talon_pre_edit` | Merge unblocked when `net_new_violations == 0`; instant self-repair guidance on `BLOCK`. |
| **Staff / Principal Architect** | Codify architectural boundaries, eliminate manual ADR authoring, and prevent circular dependencies. | `04 — Docs & Base Rulebook` and `08 — Policy Studio` | `garuda rulebook generate/approve`, `garuda docs ingest`, `garuda policy evaluate` | 4-tier traceability from Markdown ADRs and YAML predicates to exact AST symbols. |
| **Finance Controller, CFO & Statutory Auditor** | Detect cross-team financial discrepancies (48 Cr vs. 4.8 Cr) and govern competing business formulas. | `09 — Financial & Logic Audit` and `07 — Evidence & Ledger` | `garuda audit financial`, `garuda audit logic`, `garuda verify` | Pre-audit quarantine of 10x divisor bugs and RFC 6962 cryptographic audit sign-off. |
| **Founder, Product Manager & Eng Lead** | Protect mission-critical customer flows (Search, Checkout) while maintaining high feature velocity. | `01 — Overview Briefing` and `05 — Talon & Risk Hub` | `./setup.sh` / `.\setup.ps1`, `garuda audit criticality`, `garuda summary` | `HIGH_SEVERE` surfaces frozen against AI edits; `MEDIUM` trade-offs reviewed by humans. |
| **DevOps, Platform Engineer & SRE** | Enforce cross-OS build parity (`linux`, `darwin`, `windows`, `cgo`) and quarantine shadow runtime calls. | `05 — Talon & Risk Hub` (`--matrix`) and `/admin` Control Plane | `garuda check --matrix`, `garuda doctor --fix`, `POST /api/v1/telemetry/spans` | Zero CI-vs-Editor environment drift and automatic quarantine of unapproved runtime edges. |
| **AI Coding Agent (Cursor, Claude Code, Codex)** | Receive active architectural contracts automatically and self-repair violations in `<20 ms`. | `bin/garuda-mcp` (35 JSON-RPC Tools) and `.garuda/TALON_CONTRACT.md` | `garuda.talon_pre_edit`, `garuda.talon_verify_edit`, `garuda.disambiguate_logic` | Fail-closed orientation on first tool call and automated supervisor convergence loop. |
| **Investor & Technical Evaluator** | Independently verify compiler precision, determinism, MCP compliance, and cryptographic proofs. | `EVIDENCE.md` and Automated Verification Harness | `make benchmark-strict`, `python3 scripts/mcp_verify.py`, `scripts/runtime_contradiction_test.py` | 100% precision/recall across 24 compiler cases and 43/43 verified MCP protocol checks. |

---

## 3. Eight-Plane System Architecture and Epistemic State Machine

### 3.1 Structural Organization of the Eight Execution Planes

The internal architecture of Garuda is partitioned into eight cohesive execution planes that transform raw source repositories, normative Markdown documents, and live runtime telemetry into cryptographically verifiable governance decisions. At the foundation lies a multi-tenant [PostgreSQL 16](https://www.postgresql.org/docs/16/) persistence engine that stores canonical semantic entities, typed relationships, epistemic verification states, policy evaluations, and Merkle epoch leaves. Every plane communicates through strongly typed Go interfaces (`internal/analyzer`, `internal/store`, `internal/knowledge`, `internal/runtime`, `internal/policy`, `internal/merkle`, `internal/talon`, and `internal/api`), ensuring that the core truth substrate remains 100% deterministic and completely independent of external Large Language Model (LLM) APIs (`Invariant #10`).

1. **Plane 01 (Multi-Language AST, Incremental Sync, and Call-Graph Extraction)**: Parses Go (`go/ast`, `go/types`, `golang.org/x/mod`), Python (`internal/ast/python`), and TypeScript (`github.com/smacker/go-tree-sitter` and pure-Go `!cgo` lexical/scope parser `internal/ast/typescript`) into a unified semantic graph of `entities`, `claims` (relationships), and `cross_repo_edges`.
2. **Plane 02 (Epistemic Evidence and Edge-Level Runtime Drift Verification)**: Classifies every architectural claim into a strict 3-state epistemic model (`SUPPORTED`, `UNVERIFIED`, `CONTRADICTED`), ingests OpenTelemetry execution spans, and corroborates or quarantines static edges against live production traces.
3. **Plane 03 (Deterministic Graph-Native Policy Engine and Diff Scoping)**: Evaluates nine declarative YAML predicates (`layer_violation`, `transitive_claim_exists`, `cycle_exists`, `blast_radius_exceeds`, `claim_exists`, `entity_exists`, `contradiction_exists`, `verification_missing`, `language_matches`) using recursive SQL Common Table Expressions (CTEs) and partitions findings by `git diff` AST line spans.
4. **Plane 04 (Unified Change-Control and RFC 6962 Merkle Ledger)**: Orchestrates the 5-step change-control pipeline (`garuda check`) and anchors every decision and policy evaluation into a domain-separated SHA-256 Merkle tree with independently verifiable inclusion proofs (`V1Proof`).
5. **Plane 05 (Garuda Talon Active Agent Harness and 35-Tool MCP Surface)**: Compiles live workspace rules into IDE agent contracts (`.garuda/TALON_CONTRACT.md`, `.cursor/rules/garuda-talon.mdc`, `CLAUDE.md`, `AGENTS.md`, `.windsurfrules`), enforces pre-commit and `PostToolUse` hooks, and exposes 35 tools over JSON-RPC 2.0.
6. **Plane 06 (Universal Zero-Dependency Runtime and Cross-OS Diagnostics)**: Provisions portable user-space PostgreSQL `16.4.0` (`127.0.0.1:5433`) and Go `1.26.0` toolchains across Windows PowerShell 5.1/7, macOS (ARM64/Intel), Linux, Chromebook (Crostini), and Android (Termux) with self-healing diagnostics (`garuda doctor --fix`).
7. **Plane 07 (Deterministic BuildContext Provenance and 4-Tier Docs-to-Code Lineage)**: Binds OS, architecture, `CGO_ENABLED`, build tags, and dependency lockfile hashes into cryptographic `BuildContext` receipts and maps normative documents across four lineage tiers.
8. **Plane 08 (Human-in-the-Loop Formula Disambiguation, Surface Criticality Tiers, Financial Audit, and Base Rulebook Studio)**: Governs multi-team formula weights (`0.00–1.00`), enforces 3-tier surface criticality freezes, reconciles cross-team financial discrepancies (`48.00 Cr` versus `4.80 Cr`), and synthesizes baseline rulebooks for undocumented repositories.

### 3.2 The Three-State Epistemic Model and the Ten Immutable Laws

Unlike conventional static analyzers that treat every heuristic match as binary truth, Garuda implements a rigorous epistemic state machine governed by the **ACGM Resolution Invariant** and the **10 Immutable Laws** (`docs/invariants.md`). Every relationship claim in the workspace exists in one of three mutually exclusive epistemic states:

- **`SUPPORTED`**: Authoritative evidence confirms that the relationship holds within the verified scope. For static Go analysis, `go/types` compiler resolution establishes `1.0` static confidence; for runtime verification (`internal/runtime/verifier.go`), a static claim (`from_entity_id -> to_entity_id`) is promoted to `SUPPORTED` only when ingested runtime spans confirm **both** the caller and the exact target callee within correlated trace execution (`Statement 3` edge-level corroboration).
- **`UNVERIFIED`**: Available evidence is insufficient to confirm or refute the claim at runtime. Under Garuda's Honest Defaults (`migrations/064_honest_defaults.sql`), `UNVERIFIED` claims are never silently upgraded to `SUPPORTED` merely because the caller function executed without observing the target edge.
- **`CONTRADICTED`**: Observed evidence directly conflicts with an architectural invariant, a normative ADR claim (`MUST_NOT`), or the static dependency model. Contradictions occur in two directions: when a static code edge violates a documented rule or YAML policy predicate, or when live runtime telemetry reveals a shadow call between entities that possesses no corresponding static `claims` record. All `CONTRADICTED` records are automatically quarantined in the `contradictions` and `claim_verifications` tables.

| Law # | Immutable Law Title | Technical Mechanism and Enforcing Code Path | Current Enforcement Status |
| :--- | :--- | :--- | :--- |
| **Law 1** | **Revisions Are Append-Only** | `internal/store/revision_store.go` performs `INSERT` operations exclusively on `decision_revisions`; obsolete decisions are marked `superseded` rather than mutated or deleted (`migrations/005` and `046`). | `Partial` (Application-enforced; DB trigger queued) |
| **Law 2** | **Content Hash Excludes Metadata** | `content_hash` is computed as the SHA-256 digest of the canonical JSON encoding of `DecisionContent`, excluding `actor`, `timestamp`, and `revision_id` so identical content yields identical hashes. | `Enforced` (`internal/store/revision_store.go`) |
| **Law 3** | **Actor Derived From Authentication** | `internal/auth/jwt.go` extracts `actor` and `tenant_id` directly from cryptographically verified [Ed25519](https://datatracker.ietf.org/doc/html/rfc8032) JWT claims (`AuthMiddleware`), ignoring unverified client body fields. | `Enforced` (`internal/api/auth_middleware.go`) |
| **Law 4** | **Tenant Isolation Is Strict** | Every SQL query in `internal/store/` and `internal/api/` filters by `tenant_id`; cross-tenant lookups return empty sets rather than leaking existence errors, backed by RLS on `harvested_decisions`. | `Partial` (100% SQL query filtering; full table RLS queued) |
| **Law 5** | **Every Mutation Is Atomic** | `SaveDecision` (`internal/store/postgres.go`), revision creation, and policy evaluations execute inside a single ACID PostgreSQL transaction that commits the row, Merkle leaf, and audit event atomically. | `Enforced` (`internal/store/postgres.go`) |
| **Law 6** | **Hash Structure Is Independently Verifiable** | Uses v1 RFC 6962 Merkle trees (`0x00` leaf prefix, `0x01` internal node prefix, odd-node promotion closing `CVE-2012-2459`) with canonical byte encoders (`internal/merkle/decision.go`). | `Enforced` (`migrations/067` through `070`) |
| **Law 7** | **Safe Client Error Surfaces** | Internal database and stack errors are collapsed at API boundaries; client error payloads include a correlation `request_id` (`internal/api/handoff_handlers.go`). | `Partial` (Enforced on handoff/API; dashboard unification queued) |
| **Law 8** | **Idempotent Mutating Operations** | Mutating requests check `idempotency_keys` (`migrations/058_idempotency_keys.sql`) prior to execution, returning cached results on retry without duplicating state. | `Enforced` (`migrations/058_idempotency_keys.sql`) |
| **Law 9** | **End-to-End Request Observability** | Operations propagate `request_id` across structured logs, audit events, and OpenTelemetry spans (`internal/telemetry/`). | `Partial` (Core API and handoff routes enforced) |
| **Law 10** | **Zero External Dependencies for Core Integrity** | `internal/merkle/`, `internal/analyzer/`, and `internal/policy/` rely strictly on deterministic algorithms and the Go standard library—never external LLM calls. | `Enforced` (`internal/merkle/` golden vectors) |

---

## 4. Deep-Dive Methods and Algorithms: How Garuda Analyzes, Verifies, and Governs Code

### 4.1 Polyglot Compiler-Backed AST Extraction and Canonical Identity (`UUIDv5`)

The foundational analytical method in Garuda is multi-pass static semantic extraction (`internal/analyzer/workspace_analyzer.go`). For Go workspaces, Garuda discovers module boundaries via `go.mod` and multi-module `go.work` files (`DiscoverWorkspace`), constructs a custom `MultiModuleImporter` (`internal/analyzer/importer.go`) that resolves packages across workspace modules and local `vendor/` directories before falling back to `go/importer.Default()`, and filters source files through `go/build.Context.MatchFile` (`internal/analyzer/build_context.go`). This ensures that conditional build constraints (`//go:build`, `// +build`), platform filename suffixes (`_linux.go`, `_darwin.go`, `_windows.go`), and `CGO_ENABLED` gates are honored exactly as the Go compiler sees them.

Analysis executes in three deterministic passes over sorted package paths:

- **Pass 1 (Type Checking and Entity Extraction)**: Each package is parsed via `go/parser.ParseFile` and type-checked via `types.Config.Check` (`go/types`), populating `types.Info` (`Types`, `Defs`, `Uses`, `Implicits`, `Selections`). Garuda extracts `KindPackage`, `KindStruct`, `KindInterface`, `KindFunction`, `KindMethod` (recording pointer and value receiver identity in `ReceiverType`), and `KindField` entities, computing a cryptographic `SignatureHash` (`SHA-256(kind|name|receiver|signature)`) to detect contract-breaking changes.
- **Pass 2 (Semantic Relationship Extraction)**: Once all workspace entities are registered, `extractWorkspaceRelationships` inspects AST call expressions (`*ast.CallExpr`), selector expressions (`*ast.SelectorExpr`), and identifier uses (`info.Uses[fun]`) to emit `CALLS`, `IMPORTS`, and `REFERENCES` relationships at `1.0` compiler-verified confidence (`AST_EXACT` provenance).
- **Pass 3 (Cross-Package Structural Interface and Embedding Satisfaction)**: `deriveWorkspaceInterfaceImplementations` and `deriveWorkspaceStructEmbeddings` evaluate `types.Implements` and structural method-set compatibility across all modules in the workspace, emitting `IMPLEMENTS` and `EMBEDS` edges at `1.0` confidence (verified by Benchmark 024 `TestBenchmark_024_GoWork_CrossModule_ResolutionAndBlastRadius`).

For **Python** (`internal/ast/python`) and **TypeScript** (`internal/ast/typescript`), Garuda executes a two-pass structural and scope-aware resolution pipeline. Pass 1 extracts modules, classes, interfaces, functions, arrow functions, methods, and import alias bindings (`import X as Y`, `from X import Y as Z` in Python; `import { X as Y }`, `import * as NS` in TypeScript). Pass 2 resolves intra-module calls, cross-file aliased calls, constructor instantiations (`new Class()`), and instance attribute method invocations (`self.repo.fetch()`, `this.service.execute()`) at `0.95` static confidence. Notably, TypeScript analysis supports dual compilation modes: C-backed Tree-sitter (`github.com/smacker/go-tree-sitter`) when `CGO_ENABLED=1`, and a 100% Pure-Go lexical and scope parser (`internal/ast/typescript/parser_nocgo.go` and `internal/ast/treesitter/bridge_nocgo.go`) when compiled with `CGO_ENABLED=0`, enabling zero-dependency execution on machines without a C compiler.

Every extracted entity is assigned a deterministic **Canonical Identity** (`internal/canonical/identity.go`) using [RFC 4122 UUIDv5](https://datatracker.ietf.org/doc/html/rfc4122) (`SHA-1` namespace hashing over `tenant_id | workspace_id | repo_module_path | package_path | kind | receiver_type | entity_name`). Because identity is content- and scope-addressed rather than auto-incremented, re-analyzing a repository across different machines or operating systems produces the exact same entity UUIDs.

### 4.2 Content-Addressed Incremental Package Caching and Deterministic Graph Fingerprinting

To achieve sub-second latency during developer inner-loop checks (`garuda talon guard` and `garuda analyze --incremental`), Garuda implements content-addressed package caching (`internal/analyzer/cache_interface.go` and `internal/analyzer/hasher.go`). For each package directory, `ComputePackageTreeHash` reads all non-test `.go` files in alphabetical order and computes a SHA-256 digest over `fileName : fileContent`. This raw tree hash is then mixed with the active build constraints:

$$\text{TreeHash}_{\text{pkg}} = \text{SHA256}\left(\text{RawTreeHash}_{\text{pkg}} \;\|\; \text{"|"} \;\|\; \text{GOOS} \;\|\; \text{"/"} \;\|\; \text{GOARCH} \;\|\; \text{"|cgo="} \;\|\; \text{CGO} \;\|\; \text{"|tags="} \;\|\; \text{BuildTags}\right)$$

During workspace analysis, Garuda queries `PackageCache.GetPackage(ctx, tenantID, pkgPath, treeHash)`. On a cache hit, cached `Entities` and `Relationships` are appended directly without re-parsing AST files (`result.Stats.Files == 0` on warm runs), while Pass 3 cross-package interface derivation still executes over the complete entity set. Furthermore, in `v1.9.2`, the workspace-wide `GraphFingerprint` is derived strictly from the build context header and the sorted sequence of `(pkgPath, treeHash)` pairs:

$$\text{GraphFingerprint} = \text{SHA256}\left(\text{BuildHeader} \;\|\; \bigoplus_{p \in \text{SortedPkgs}} \left(p \;\|\; \text{":"} \;\|\; \text{hex}(\text{TreeHash}_p) \;\|\; \text{":"}\right)\right)$$

Excluding transient runtime parsing counters (`result.Stats.Files`) from `repoHasher` guarantees mathematical fingerprint identity between cold full-parse runs and warm cache-skipping runs (`TestIncrementalWorkspaceAnalysis_CacheSkipping`).

![Multi-Repository Semantic Extraction Progression](corpus_progression_chart.png)

### 4.3 Graph-Native Policy Predicates, Recursive CTEs, and Diff-Scoped Violation Partitioning

Garuda's policy engine (`internal/policy/parser.go`, `evaluator.go`, `engine.go`) evaluates declarative YAML bundles (`policies/*.yaml`) directly against the PostgreSQL semantic graph. Rather than pulling millions of edges into application memory, the evaluator compiles predicates into parameterized SQL queries and recursive Common Table Expressions (`WITH RECURSIVE`):

- **Transitive Dependency Chain Detection (`transitive_claim_exists`)**: Uses a depth-bounded recursive CTE (`call_path`) with cycle prevention (`NOT (c.to_entity_id = ANY(cp.path))`) to discover hidden multi-hop backdoors—such as an HTTP handler calling a logging helper that calls a cache utility that opens a raw SQL connection (`depth <= max_depth`).
- **Circular Dependency Detection (`cycle_exists`)**: Traverses package-level import and call edges up to `max_depth` hops to identify architectural cycles (`PkgA -> PkgB -> PkgC -> PkgA`).
- **Layer Isolation (`layer_violation`)** and **Hub Escalation (`blast_radius_exceeds`)**: Joins `claims` against source and target `entities` filtered by POSIX regular expressions (`~`) and aggregates distinct inbound callers (`HAVING COUNT(DISTINCT c.from_entity_id) > $max_callers`).

When invoked in diff-scoped mode (`garuda check --diff HEAD` or `garuda talon guard --staged`), `internal/diff/diff.go` executes `git diff --unified=0`, parses modified line ranges `[startLine, endLine]`, and resolves every AST entity whose source span `[line_start, line_end]` overlaps the diff (`ResolveTouchedEntities`). `Evaluator.EvaluateWithScope` then applies `filterEvidenceByDiffScope`: only policy violations whose source or target entity belongs to `TouchedEntityIDs` are classified as **Net-New Violations** (`net_new_violations_count`), whereas pre-existing violations in untouched functions—even within the same file or package—are isolated as **Baseline Debt** (`baseline_violations_count`). This ensures developers and AI agents are gated strictly on the architectural impact of their own edits.

### 4.4 Human-in-the-Loop Formula Weighting (`0.00–1.00`), Criticality Tiers, and Pre-Audit Financial Lineage

In `v1.9.0` through `v1.9.2`, Garuda introduced deterministic governance over ambiguous business logic and financial reporting (`internal/api/dashboard_audit_rulebook.go`, `cmd/garuda/audit_cmd.go`, `migrations/093_human_in_the_loop_audit_rulebook.sql`).

When multiple teams implement divergent formulas for a shared concept (such as `Loss Ratio`), Garuda clusters the candidate functions in `logic_variants_registry` and computes two orthogonal metrics:

1. **Architectural Harm Rank and Surface Criticality Tier**: Based on inbound transitive caller fan-in and downstream surface bindings (`surface_criticality_tiers`), each variant is classified into `LOW` (`<= 10` callers, leaf UI helpers, Harm Rank #1), `MEDIUM` (`11–50` callers, non-blocking features like Wishlist Promo Loader, Harm Rank #2), or `HIGH_SEVERE` (`> 50` callers or bound to core Homepage, Search Redirect, or Settlement Ledger surfaces, Harm Rank #3 — `HUMAN_FREEZE_DO_NOT_TOUCH`).
2. **Dynamic Team Relevance Weight (`0.00` to `1.00`)**: Given an active developer or agent working in team scope $T_{\text{active}} \in \{\text{ops}, \text{marketing}, \text{management}\}$, Garuda assigns a contextual relevance distribution across the $K$ discovered variants:

$$W(v_i \mid T_{\text{active}}) = \begin{cases} 0.70 & \text{if } \text{team}(v_i) = T_{\text{active}} \\ \frac{0.20}{K - 1} & \text{if } \text{team}(v_i) \neq T_{\text{active}} \end{cases}$$

with the remaining `0.10` probability mass reserved for unmapped legacy call sites. Crucially, even when a `HIGH_SEVERE` variant receives a `0.70` contextual weight (for example, when inspecting from the `--team management` perspective), Garuda keeps its status locked as `FROZEN_SEVERE` (`HUMAN_FREEZE_DO_NOT_TOUCH`) until a human architect or financial controller explicitly records an approval via `garuda audit logic --decide` or `garuda.decide_logic_variant`.

For **Financial Discrepancy Auditing** (`financial_audit_reconciliations`), Garuda compares metric-producing symbols across teams ($V_A$ in Team A versus $V_B$ in Team B), computes the variance ratio:

$$\text{VarianceRatio} = \frac{\max(|V_A|, |V_B|)}{\min(|V_A|, |V_B|)}$$

and inspects the AST binary expression divisors and embedded SQL aggregation strings. When $\text{VarianceRatio} \approx 10.0$ (such as `48.00 Crore INR` in `ops.AggregateClaimsPaidCrore` versus `4.80 Crore INR` in `finance.ExportExecutiveClaimsSummary`), Garuda automatically flags `SCALE_FACTOR_10X_MISMATCH`, identifies the `/ 1000.0` versus `/ 100.0` Lakhs-to-Crore divisor bug in the AST trace, and prescribes routing both pipelines through `store.CanonicalCroreAmount` (`SYNTH-AUDIT-005`).

![Surface Criticality Blast Radius and Financial Audit Detection](logic_weights_criticality_chart.png)

### 4.5 RFC 6962 Cryptographic Merkle Ledger and Canonical Byte Encoding

To provide tamper-evident auditability for compliance and statutory reviews, every persisted governance decision (`decisions`) and policy evaluation (`policy_evaluations`) is anchored into an append-only RFC 6962 Merkle tree (`internal/merkle/tree.go`, `decision.go`, `evaluation.go`, `epoch.go`). To prevent second-preimage attacks and CVE-2012-2459 (odd-leaf duplication vulnerabilities), Garuda enforces strict domain separation and deterministic canonical serialization (`verification_version = 1`):

$$\text{LeafHash} = \text{SHA256}\left(\texttt{0x00} \;\|\; \text{CanonicalBytes}(D)\right)$$

$$\text{InternalNodeHash}(L, R) = \text{SHA256}\left(\texttt{0x01} \;\|\; L \;\|\; R\right)$$

When a tree level contains an odd number of nodes, the trailing node is promoted directly to the next level rather than duplicated. Each record stores its sibling audit path (`V1Proof`) and epoch root hash (`merkle_roots` and `merkle_epoch_leaves`). Any auditor can run `./bin/garuda verify` or `GET /api/v1/evidence/verify/{id}` to recompute the canonical byte digest from the raw database columns and verify mathematical inclusion up to the tenant's Merkle root without trusting external servers or keys (`Law #6` and `Law #10`).

---

## 5. Complete Technology Stack, Binary Topology, and Storage Schema

### 5.1 Core Language, Compiler Libraries, Cryptography, and Runtime Stack

Garuda is implemented in **Go `1.26.0`** (with backward-compatible toolchain fallback to Go `1.24+`), chosen for its native compiler introspection packages (`go/ast`, `go/parser`, `go/token`, `go/types`, `go/build`), high-concurrency goroutine scheduling, static single-binary compilation (`CGO_ENABLED=0`), and sub-millisecond startup latency for CLI and git-hook invocations. The complete technology stack is structured to guarantee zero external cloud dependencies and 100% air-gapped execution capability.

| Architectural Layer | Technology / Library Package | Version | Specific Engineering Role in Garuda |
| :--- | :--- | :--- | :--- |
| **Core Language & Compiler AST** | Go Standard Library (`go/ast`, `go/types`, `go/build`, `golang.org/x/mod`) | `Go 1.26.0` / `v0.40.0` | Exact type-checked Go AST extraction, `go.work` multi-module resolution, and conditional build-tag filtering. |
| **Polyglot AST Parsing** | `github.com/smacker/go-tree-sitter` + Pure-Go `!cgo` Lexical Parser | `v0.0.0-20240827` | Dual-mode TypeScript/JavaScript AST parsing (CGO Tree-sitter or zero-dependency Pure-Go `!cgo`) and Python AST extraction. |
| **Primary Relational & Graph Store** | PostgreSQL (`github.com/jackc/pgx/v5`, `github.com/lib/pq`) | `PostgreSQL 16.4` / `pgx v5.10.0` | ACID persistence for 93 schema migrations, recursive CTE graph traversals, JSONB evidence storage, and symbol caching. |
| **Optional High-Speed Cache** | Redis (`github.com/redis/go-redis/v9`) + In-Memory `MemoryPackageCache` | `v9.20.0` | Optional distributed API response caching and thread-safe content-addressed package AST caching. |
| **Cryptographic Primitives** | `crypto/ed25519`, `crypto/sha256`, `github.com/golang-jwt/jwt/v5`, `github.com/google/uuid` | `Go stdlib` / `v5.3.1` / `v1.6.0` | Ed25519 asymmetric JWT signing/verification, RFC 6962 SHA-256 Merkle trees, and RFC 4122 UUIDv5 canonical entity IDs. |
| **HTTP Gateway & Routing** | `github.com/gorilla/mux`, `github.com/labstack/echo/v4`, `golang.org/x/time/rate` | `v1.8.1` / `v4.15.4` / `v0.16.0` | REST API routing, token-bucket rate limiting, session cookie middleware, and embedded 9-surface PWA serving. |
| **CLI & Interactive Shell** | `github.com/spf13/cobra`, `github.com/AlecAivazis/survey/v2` | `v1.10.2` / `v2.3.7` | Command-line interface orchestration (`garuda`, `garuda-talon`) and interactive terminal initialization wizards. |
| **Configuration & Policy Formats** | `gopkg.in/yaml.v3`, `github.com/goccy/go-yaml`, `github.com/BurntSushi/toml` | `v3.0.1` / `v1.19.2` / `v1.6.0` | Fail-closed YAML policy bundle parsing, OpenAPI 3.0 specification loading, and `pyproject.toml` dependency hashing. |
| **Git & Diff Introspection** | `github.com/go-git/go-git/v5` + Native `git diff --unified=0` Parser | `v5.19.2` | Commit history mining (`internal/ingest/git_miner.go`), dirty working-tree hashing, and diff-to-AST line span mapping. |
| **Observability & Runtime Telemetry** | `go.opentelemetry.io/otel/sdk` + OTLP Span Correlator | `v1.43.0` | Live runtime span ingestion (`POST /api/v1/telemetry/spans`), caller/callee symbol correlation, and contradiction quarantine. |
| **Frontend Dashboard Instrument** | Vanilla ES2022 JS (`dashboard.js`), CSS3 (`dashboard.css`), Custom Offline SVG Force/DAG Engine | `v1.9.1 Embedded` | Zero-CDN, 100% offline 9-surface Software Intelligence Instrument and installable Chromium/Mobile PWA. |
| **Portable All-OS Runtime Provisioner** | Pre-built Portable PostgreSQL `16.4.0` + Portable Go `1.26.0` (`setup.sh` / `setup.ps1`) | `v1.9.1 Universal` | Zero-admin user-space provisioning (`~/.garuda/runtime/postgres` and `~/.garuda/toolchain/go`) across 5 OS families. |

### 5.2 Executable Binary Topology (`cmd/`) and Internal Package Architecture (`internal/`)

The repository compiles seven specialized entrypoints under `cmd/` (with the four primary operational binaries `garuda`, `garuda-api`, `garuda-mcp`, and `garuda-talon` built in a single pass into `./bin/`):

1. **`cmd/garuda` (`./bin/garuda`)**: The unified command-line interface exposing 28 top-level and nested commands for repository initialization (`init`), health diagnostics (`doctor`), multi-language AST analysis (`analyze`), 5-plane change control (`check`), Talon agent harness control (`talon`), declarative policy management (`policy`), documentation lineage (`docs`), business and financial auditing (`audit`), rulebook synthesis (`rulebook`), static hygiene (`hygiene`), and daemon management (`dev`, `status`, `verify`).
2. **`cmd/garuda-api` (`./bin/garuda-api`)**: The HTTP REST gateway, Ed25519 authentication server, OpenTelemetry span ingestion endpoint, and embedded web server hosting the 9-surface `/dashboard` PWA and `/admin` multi-tenant operator control plane.
3. **`cmd/garuda-mcp` (`./bin/garuda-mcp`)**: The JSON-RPC 2.0 Model Context Protocol server (`protocolVersion: "2025-06-18"`) over `stdio`, exposing 35 deterministic tools, MCP Prompts (`garuda.architectural_contract`), MCP Resources (`garuda://workspace/contract`), strict argument whitelisting (`toolArgWhitelist`), and first-call active contract injection.
4. **`cmd/garuda-talon` (`./bin/garuda-talon`)**: Dedicated standalone binary for high-speed execution of the Garuda Talon pre-edit, post-edit, git hook, and autonomous supervisor loop workflows.
5. **`cmd/garuda-worker`**: Asynchronous background daemon executing periodic runtime verification (`RecomputeWorkspaceVerification`) and 10-second Merkle epoch root anchoring.
6. **`cmd/telemetry-collector`**: Standalone OpenTelemetry collector bridge for high-throughput span normalization.
7. **`cmd/garuda-slack` and `cmd/migrate`**: Slack governance notification bot (`internal/harvester/slack.go`) and standalone idempotent SQL schema migration runner.

### 5.3 PostgreSQL Relational and Graph Schema Architecture (93 Idempotent Migrations)

Garuda's persistence layer (`internal/store/` and `migrations/000_create_extensions.sql` through `093_human_in_the_loop_audit_rulebook.sql`) consists of 93 idempotent SQL migrations organized across six functional schema domains:

| Schema Domain | Core PostgreSQL Tables | Primary Keys, Indexes, and Architectural Purpose |
| :--- | :--- | :--- |
| **1. Multi-Tenant & Workspace Isolation** | `tenants`, `users`, `tenant_members`, `workspaces`, `workspace_members`, `repositories` | Scopes all records by `tenant_id` and `workspace_id` (`uq_repositories_workspace_name`). Enforces workspace-level access control and tracks disabled/last-login states (`migrations/077`–`082`). |
| **2. Polyglot AST & Semantic Graph** | `entities`, `claims`, `relationships`, `cross_repo_edges`, `workspace_symbol_cache` | Stores canonical UUIDv5 entities (`kind`, `canonical_name`, `file_path`, `line_start`, `line_end`, `signature`, `ast_hash`) and typed edges (`CALLS`, `IMPORTS`, `IMPLEMENTS`, `EMBEDS`, `INHERITS`) with confidence scores and resolution methods (`migrations/032`–`074`). |
| **3. Epistemic Verification & Runtime Telemetry** | `document_claims`, `claim_verifications`, `runtime_observations`, `runtime_edges`, `contradictions`, `telemetry_aggregates` | Tracks extracted ADR `MUST`/`MUST_NOT` rules, 3-state verification (`SUPPORTED`, `UNVERIFIED`, `CONTRADICTED`), live OTel spans, and quarantined shadow runtime calls (`migrations/050`–`084`). |
| **4. Policy Engine & Governance Decisions** | `policies`, `policy_rules`, `policy_evaluations`, `decisions`, `decision_revisions`, `idempotency_keys` | Persists declarative YAML rules, evaluation outcomes (`ALLOW`, `WARN`, `REVIEW`, `BLOCK`), bitemporal validity windows (`valid_from`, `valid_to`), and append-only revisions (`migrations/001`–`070`). |
| **5. RFC 6962 Cryptographic Merkle Ledger** | `merkle_roots`, `merkle_snapshots`, `merkle_epoch_leaves`, `evidence` | Stores tenant Merkle root hashes, block heights, `verification_version = 1` domain-separated leaves, and JSONB `V1Proof` inclusion paths (`migrations/012`–`070`). |
| **6. Agent Coordination & v1.9.1 Human Audit** | `agent_checkpoints`, `agent_handoffs`, `mcp_sessions`, `mcp_agent_watermarks`, `surface_criticality_tiers`, `logic_variants_registry`, `financial_audit_reconciliations`, `synthesized_rulebooks` | Manages transactional multi-agent handoffs, MCP session telemetry, `HIGH_SEVERE`/`MEDIUM`/`LOW` criticality tiers, `0.00–1.00` team formula weights, 48 Cr vs. 4.8 Cr financial audits, and approved base rulebooks (`migrations/009`, `017`, `086`, `088`, `093`). |

---

## 6. Chronological Engineering Evolution: How Garuda Was Built (`v1.0` to `v1.9.2`)

### 6.1 Phase I (`v1.0` – `v1.3.3`): Epistemic Foundation, Multi-Language AST, and Cryptographic Hardening

Garuda began with a foundational architectural premise: static code analysis and architectural documentation fail in enterprise environments because they lack a shared, verifiable truth substrate. In `v1.0` and `v1.1.0`, the core PostgreSQL semantic graph, the Go `go/types` analyzer, the initial 20-tool MCP server, and the dual-root Merkle ledger were constructed. However, early validation revealed three critical engineering gaps that required deep architectural iteration across `v1.2.0` and `v1.3.0`:

First, developers needed a single command that unified blast-radius analysis, epistemic evidence, policy evaluation, and Merkle sealing. In `v1.2.0`, `garuda check` (`cmd/garuda/check_cmd.go`) was built alongside the four graph-native recursive SQL policy predicates (`layer_violation`, `transitive_claim_exists`, `cycle_exists`, `blast_radius_exceeds`) and multi-language call-graph parity for Python and TypeScript. Simultaneously, all 90+ SQL migrations were refactored with idempotent `DO $$ BEGIN ... EXCEPTION WHEN duplicate_object` guards (`internal/store/migrate.go`) so database upgrades never failed on pre-existing types or constraints.

Second, in `v1.3.0`, following a rigorous security and due-diligence audit (the Astra Due-Diligence Remediation), six production-hardening upgrades were engineered:
- **Cryptographic Ed25519 JWT Enforcement**: `Server.AuthMiddleware` (`internal/api/auth_middleware.go`) was wired directly to `jwtConfig.ValidateToken(token)`, rejecting forged or expired tokens with `401 Unauthorized` and binding `tenant_id` and `actor` strictly from verified cryptographic claims.
- **Strict Edge-Level `SUPPORTED` Promotion**: `RecomputeWorkspaceVerification` (`internal/runtime/verifier.go`) was hardened so a static claim (`from -> to`) is promoted to `SUPPORTED` only when runtime spans confirm **both** the caller and callee in correlated traces.
- **Fail-Closed Policy Engine**: `ParseDirectory` and `Engine.Run` (`internal/policy/`) were updated to abort immediately on malformed YAML syntax, unknown predicates, or Merkle anchoring failures rather than defaulting to `ALLOW`.
- **RFC 6962 v1 Canonical Merkle Proofs**: Linear v0 hash chains were superseded by domain-separated RFC 6962 Merkle trees (`migrations/067` through `070`) with full inclusion proof verification (`VerifyDecisionCryptographic`).

### 6.2 Phase II (`v1.4.0` – `v1.6.1`): Software Intelligence Instrument, Garuda Talon, and Zero-Dependency Onboarding

In `v1.4.0`, the web interface (`/dashboard`) was transformed from a basic D3 force graph into a multi-surface **Software Intelligence Instrument** featuring an Executive Briefing, a 1-hop/2-hop Neighborhood Architecture Map with a synchronized 4-tab Inspector Drawer, a pre-merge Change Review Workbench, a Policy Studio, and a global `⌘K` / `Ctrl+K` command palette. Benchmark 024 (`go.work` cross-module interface satisfaction) was also hardened to achieve `24/24` (`100%`) precision and recall across the compiler benchmark suite.

In `v1.5.0`, recognizing that passive post-commit CI checks occur too late in an AI coding workflow, **Garuda Talon** (`internal/talon/talon.go`, `cmd/garuda/talon_cmd.go`) was architected. Talon introduced active contract compilation (`garuda talon init`), writing live workspace ADRs and YAML guardrails directly into `.cursor/rules/garuda-talon.mdc`, `CLAUDE.md`, `.claude/settings.json` (`PostToolUse` hook), `AGENTS.md`, `.windsurfrules`, and `.git/hooks/pre-commit`. It also introduced git-diff hunk-to-AST mapping (`internal/diff/diff.go`), partitioning `net_new_violations` from `baseline_violations`, and an autonomous supervisor loop (`garuda talon run`) that feeds deterministic Self-Repair Prompts back to CLI coding agents until `ALLOW` is reached.

In `v1.6.0` and `v1.6.1`, feedback from early evaluators highlighted onboarding friction when users lacked Docker, Go, PostgreSQL, or a C compiler (`gcc`/`clang`/Xcode). To solve this permanently, a Pure-Go (`!cgo`) TypeScript AST fallback (`internal/ast/typescript/parser_nocgo.go`) and universal 1-command zero-dependency installers (`install.sh`/`setup.sh` and `install.ps1`/`setup.ps1`) were built, alongside `garuda doctor --fix` (`cmd/garuda/doctor_cmd.go`) and full Progressive Web App (`/manifest.json`, `/sw.js`) support across macOS, Windows, Linux, Chromebook (Crostini), and Android (Termux).

### 6.3 Phase III (`v1.7.0` – `v1.9.2`): BuildContext Provenance, Cross-OS Polish, Financial Audits, and Cache Determinism

In `v1.7.0` and `v1.8.0`, Garuda added the 4-tier **Docs, Specs & Guardrails Lineage Graph** (`04 — Docs & Base Rulebook`), the **Architectural Risk Hub Leaderboard** (`05 — Talon & Risk Hub`), conditional `go/build.Context` build-tag filtering, cryptographic `BuildContext` provenance receipts, the CI-vs-Editor Environment Drift Detector, and the 4-profile Cross-Platform Build Matrix (`--matrix`).

Between `v1.8.1` and `v1.8.4`, intense real-world installation testing across Windows PowerShell 5.1 (`CP1252` encoding) and macOS (Apple Silicon and Intel) drove crucial systems engineering fixes:
- Stripping all multi-byte UTF-8 emojis from `install.ps1` and `setup.ps1` so Windows PowerShell 5.1 never throws `ParserError: UnexpectedToken`, and wrapping native binaries in `Invoke-QuietCmd` so informational `stderr` output from `initdb.exe` or `git` never triggers a terminating `NativeCommandError`.
- Provisioning pre-built Portable PostgreSQL `16.4.0` binaries (`~18 MB`) in 3 seconds in `~/.garuda/runtime/postgres` (avoiding 40-minute Homebrew source builds), normalizing `@localhost:` to explicit IPv4 `@127.0.0.1:5433` (`internal/config/env.go`) to prevent macOS IPv6 `[::1]` connection-refused errors, and adding `pg_isready` WAL recovery loops.

In `v1.9.0` and `v1.9.1`, Garuda expanded from pure engineering governance to enterprise business and financial assurance: adding **Multi-Team Formula Disambiguation (`0.00–1.00` weights)**, **3-Tier Surface Criticality Classification**, **Continuous Financial Discrepancy Auditing (`48.00 Cr` vs. `4.80 Cr`)**, the **Auto-Synthesized Base Rulebook Studio**, and expanding `bin/garuda-mcp` to **35 production tools** (`43/43` verification checks passing). Finally, in `v1.9.2`, `AnalyzeWorkspaceWithOptions` (`internal/analyzer/workspace_analyzer.go`) was refined to remove transient runtime parsed-file counters (`active_files=%d:skipped=%d`) from `repoHasher`, ensuring 100% deterministic graph fingerprint equality between cold full-parse runs and warm incremental cache hits (`TestIncrementalWorkspaceAnalysis_CacheSkipping`).

| Release Version | Milestone Date | Key Architectural Capabilities Introduced | Verification & Benchmark Gate |
| :--- | :--- | :--- | :--- |
| **`v1.1.0`** | `2026-09-20` | Initial Go AST analyzer, PostgreSQL semantic graph, 20-tool MCP server, and dual-root Merkle ledger. | Corpus A (14 Go repos) & Corpus B (9 polyglot repos) |
| **`v1.2.0`** | `2026-10-24` | Unified `garuda check` 5-plane CLI, recursive CTE policy predicates, Python/TS call graphs, idempotent migrations. | `23/23` MCP checks & RFC 6962 decision verification |
| **`v1.3.0` – `v1.3.3`** | `2026-10-24` | Ed25519 JWT auth enforcement, edge-level `SUPPORTED` runtime promotion, fail-closed YAML policy parser, K8s hardening. | `10/10` controlled runtime contradiction suite |
| **`v1.4.0`** | `2026-10-24` | 6-surface Software Intelligence Instrument (`/dashboard`), 1-hop/2-hop neighborhood map, Benchmark 024 `go.work` fix. | `24/24` (`100%`) strict compiler benchmark suite |
| **`v1.5.0` – `v1.5.1`** | `2026-10-25` | **Garuda Talon** active agent harness, IDE/Git hook compilation, diff-scoped `net_new_violations`, self-repair loop. | Sub-second incremental AST sync & Talon unit suite |
| **`v1.6.0` – `v1.6.1`** | `2026-10-25` | Pure-Go (`!cgo`) TypeScript parser, Windows/macOS/Linux/Chromebook/Android 1-command installers, `garuda doctor --fix`, PWA. | `CGO_ENABLED=0` cross-OS compilation matrix |
| **`v1.7.0` – `v1.8.0`** | `2026-10-26` | 4-Tier Docs-to-Code Lineage Graph, Risk Hub Leaderboard, `BuildContext` provenance receipts, `--matrix` 4-OS gate. | `TestTwoBuildsExposeDifferentEdges` & Drift Detector |
| **`v1.8.1` – `v1.8.4`** | `2026-10-27` | Pure-ASCII Windows PS 5.1 scripts, 3-second Portable PostgreSQL `16.4`, IPv4 `127.0.0.1:5433` normalization, schema `057`. | `TestDoubleRunDeterminism` (24/24) & all-OS smoke |
| **`v1.9.0` – `v1.9.1`** | `2026-10-29` | `0.0–1.0` formula weights, 3-tier surface criticality, 48 Cr vs. 4.8 Cr financial audit, Base Rulebook Studio, 35 MCP tools. | `43/43` MCP verifier checks & migration `093` |
| **`v1.9.2`** | `2026-10-29` | Deterministic workspace graph fingerprinting across cold and warm package-cache runs (`workspace_analyzer.go`). | `go test -v -race ./internal/...` 100% green |

---

## 7. Subsystem Reference: Garuda Talon, 35-Tool MCP Server, and 9-Surface Dashboard PWA

### 7.1 Garuda Talon Active Agent Harness and Six-Point Hook Matrix

`Garuda Talon` (`internal/talon/talon.go`) bridges static repository governance and live AI agent execution. When `garuda talon init` runs (automatically invoked during `./setup.sh`, `.\setup.ps1`, `garuda init`, or `garuda mcp install`), it queries the workspace's ingested ADR claims (`document_claims`), active YAML policy files (`./policies/*.yaml`), surface criticality tiers (`surface_criticality_tiers`), and top high-fan-in architectural hubs, compiling them into a single canonical Markdown contract (`.garuda/TALON_CONTRACT.md`) and propagating enforcement across six integration points:

1. **Cursor Always-On Rule (`.cursor/rules/garuda-talon.mdc`)**: Written with frontmatter `alwaysApply: true`, instructing Cursor's agent composer to call `garuda.talon_pre_edit` before modifying code and `garuda.talon_verify_edit` immediately after editing.
2. **Cursor MCP Configuration (`.cursor/mcp.json`)**: Configures `bin/garuda-mcp` with `GARUDA_TALON_ENFORCE=true` and explicit `DATABASE_URL` and `GARUDA_WORKSPACE` bindings.
3. **Claude Code Post-Tool-Use Hook (`.claude/settings.json` and `CLAUDE.md`)**: Configures a `PostToolUse` hook matching file-mutation tools (`Edit`, `Write`, `MultiEdit`) that executes `./bin/garuda talon guard --hook`. If a net-new violation is introduced, the hook exits with code `2` and injects the deterministic Self-Repair Prompt directly into Claude Code's context window.
4. **Universal Agent Instructions (`AGENTS.md` and `.windsurfrules`)**: Provides workspace-specific architectural contracts and forbidden dependency patterns for OpenAI Codex CLI, Windsurf, and GitHub Copilot Coding Agent.
5. **Git Pre-Commit Gate (`.git/hooks/pre-commit`)**: Executes `garuda talon guard --staged --fail-on-block`, preventing any human or agent commit containing net-new policy or ADR violations from entering local git history.
6. **Git Commit-Msg Provenance Trailer (`.git/hooks/commit-msg`)**: Appends cryptographic provenance trailers (`Garuda-Verdict: ALLOW`, `Garuda-Merkle-Root: <sha256>`, `Garuda-Build-Receipt: <hash>`) read from `.garuda/LAST_DECISION.env` onto verified commits.

### 7.2 Complete 35-Tool Model Context Protocol (`bin/garuda-mcp`) Inventory

The MCP server (`cmd/garuda-mcp/main.go`, `tools_talon.go`, `tools_graph.go`, `tools_governance.go`, `tools_semantic.go`, `tools_audit.go`, `tools_whitelist.go`) implements JSON-RPC 2.0 over `stdio`. To prevent injection or hallucinated parameters, every incoming `tools/call` request is validated against `toolArgWhitelist` (`TestToolArgWhitelistCoversAllTools`). Furthermore, when `GARUDA_TALON_ENFORCE=true` is active, the session tracker (`session_tracking.go`) intercepts the very first tool call in any agent session and attaches `_garuda_active_contract` containing the workspace's active `MUST`/`MUST_NOT` rules and YAML guardrails.

| # | MCP Tool Name | Category | Whitelisted Input Parameters | Deterministic Output & Persistence Behavior |
| :--- | :--- | :--- | :--- | :--- |
| **1** | `garuda.talon_pre_edit` | Active Talon Harness | `workspace`, `target`, `tenant_id` | Returns callers, callees, risk score (0–100), criticality tier, and governing ADR/policy rules before editing. |
| **2** | `garuda.talon_verify_edit` | Active Talon Harness | `workspace`, `repo_dir`, `staged`, `diff`, `files`, `tenant_id` | Runs incremental AST sync, evaluates diff-scoped net-new violations, seals Merkle leaf, and emits Self-Repair Prompt. |
| **3** | `garuda.briefing` | Orientation & Discovery | `workspace`, `tenant_id` | Executive workspace summary: repository count, entity/claim totals, Merkle block height, and contradiction count. |
| **4** | `garuda.entities` | Orientation & Discovery | `workspace`, `package`, `kind`, `limit` | Paginated directory of indexed AST entities filtered by package path or semantic kind. |
| **5** | `garuda.find_entity` | Orientation & Discovery | `workspace`, `name_pattern`, `kind`, `package`, `limit` | Case-insensitive and suffix-aware (`%.Symbol`) entity search ranked by inbound caller fan-in. |
| **6** | `garuda.inspect` | Orientation & Discovery | `workspace`, `entity_id` | Deep entity profile: canonical UUIDv5, file path, line span, signature hash, and direct relationships. |
| **7** | `garuda.query` | Orientation & Discovery | *(none)* | Read-only structural workspace overview across repositories and languages. |
| **8** | `garuda.neighbors` | Topology & Blast Radius | `workspace`, `entity_id`, `limit` | 1-hop inbound callers and outbound dependencies with edge types and confidence scores. |
| **9** | `garuda.blast_radius` | Topology & Blast Radius | `workspace`, `entity_id`, `depth`, `min_confidence` | Transitive upstream impact tree up to `depth` hops with `CRITICAL`/`HIGH`/`MEDIUM`/`LOW` severity breakdown. |
| **10** | `garuda.implementers` | Topology & Blast Radius | `workspace`, `entity_id`, `name`, `limit` | Resolves all concrete structs or classes satisfying an interface (`IMPLEMENTS`) with `resolveTopClaimTarget` fallback. |
| **11** | `garuda.subclasses` | Topology & Blast Radius | `workspace`, `entity_id`, `name`, `language`, `limit` | Resolves derived classes and struct embeddings (`INHERITS` and `EMBEDS`) across Go, Python, and TypeScript. |
| **12** | `garuda.get_impact` | Topology & Blast Radius | `decision_id`, `tenant_id` | Retrieves graph entities and downstream services impacted by a recorded architectural decision. |
| **13** | `garuda.get_lineage` | Topology & Blast Radius | `decision_id`, `tenant_id` | Traverses bitemporal ancestor and descendant revision chains (`decision_revisions`). |
| **14** | `garuda.policy.list` | Governance & Policy | `tenant_id` | Lists registered declarative architecture policies and their severity tiers. |
| **15** | `garuda.policy.evaluate` | Governance & Policy | `workspace` | Evaluates active YAML policies against the workspace graph and returns per-rule outcomes. |
| **16** | `garuda.verify_policy_evaluation` | Governance & Policy | `evaluation_id`, `tenant_id` | Recomputes canonical evaluation bytes and verifies the RFC 6962 `V1Proof` inclusion path against `merkle_roots`. |
| **17** | `garuda.governance.status` | Governance & Policy | `workspace` | Returns current Merkle root hash, block height, verification version, and active quarantine counts. |
| **18** | `garuda.propose_decision` | Governance & Policy | `scope_domain`, `scope_system`, `tenant_id` | Proposes a scoped architectural decision, checks for contradictions, and anchors an atomic Merkle revision. |
| **19** | `garuda.detect_contradictions` | Epistemic Truth & Drift | `tenant_id` | Lists unresolved static, ADR, and runtime `CONTRADICTED` records requiring remediation. |
| **20** | `garuda.check_drift` | Epistemic Truth & Drift | `workspace` | Evaluates documentation-to-code drift and CI-vs-Editor `BuildContext` environment drift. |
| **21** | `garuda.query_claims` | Epistemic Truth & Drift | `workspace`, `subject` | Queries extracted normative ADR claims (`document_claims`) and static graph relationships for a subject. |
| **22** | `garuda.handoff` | Multi-Agent Continuity | `task_id`, `source_agent_id`, `target_agent_id` | Atomically transfers task ownership between agents inside a PostgreSQL transaction and creates a checkpoint. |
| **23** | `garuda.resume` | Multi-Agent Continuity | `agent_id`, `checkpoint_id` | Consumes a single-use agent handoff checkpoint and restores architectural state and watermarks. |
| **24** | `garuda.disambiguate_logic` | Business Logic Audit | `workspace`, `team`, `concept`, `tenant_id` | Returns competing team formulas (e.g., `Loss Ratio`) with `0.00–1.00` weights, caller counts, and harm ranks. |
| **25** | `garuda.decide_logic_variant` | Business Logic Audit | `workspace`, `team`, `variant_id`, `decision_action`, `custom_weight`, `tenant_id` | Records a human stakeholder's approval (`APPROVED_BY_HUMAN`) or freeze (`FROZEN_SEVERE`) in `logic_variants_registry`. |
| **26** | `garuda.audit_financial_discrepancy` | Financial Data Audit | `workspace`, `metric`, `team_a_name`, `team_a_crore`, `team_b_name`, `team_b_crore`, `tenant_id` | Audits cross-team Crore/INR metrics (e.g., `48.0 Cr` vs. `4.8 Cr`) and traces AST/SQL divisor bugs. |
| **27** | `garuda.reconcile_financial_audit` | Financial Data Audit | `workspace`, `audit_id`, `metric`, `approved_by`, `tenant_id` | Marks a flagged financial discrepancy as `RECONCILED_BY_HUMAN` with auditor provenance in PostgreSQL. |
| **28** | `garuda.synthesize_rulebook` | Base Rulebook Studio | `workspace`, `tenant_id` | Reverse-engineers repository topology into `.garuda/DRAFT_BASE_RULEBOOK.md` and `synthesized_rulebooks`. |
| **29** | `garuda.approve_base_rulebook` | Base Rulebook Studio | `workspace`, `approved_by`, `tenant_id` | Seals `.garuda/APPROVED_BASE_RULEBOOK.md` with an RFC 6962 SHA-256 digest upon human approval. |
| **30** | `garuda.criticality_tiers` | Surface Criticality | `workspace`, `tier_filter`, `surface_name`, `package_pattern`, `criticality_tier`, `tenant_id` | Lists or registers `HIGH_SEVERE`, `MEDIUM`, and `LOW` architectural surfaces in `surface_criticality_tiers`. |
| **31** | `garuda.docs_graph` | Docs-to-Code Lineage | `workspace`, `doc_type`, `tenant_id` | Returns the 4-tier Document -> Extracted Rule -> Policy Guardrail -> Code Symbol lineage graph. |
| **32** | `garuda.docs_ingest` | Docs-to-Code Lineage | `workspace`, `path`, `title`, `doc_type`, `tenant_id` | Ingests Markdown ADRs or specifications into `document_claims` and links matching AST symbols. |
| **33** | `garuda.doctor_diagnose` | Diagnostics & Matrix | `workspace`, `fix`, `tenant_id` | Runs the 6-point System Doctor diagnostic and optional auto-repair over MCP. |
| **34** | `garuda.build_matrix_check` | Diagnostics & Matrix | `workspace`, `repo_dir`, `diff`, `tenant_id` | Evaluates the 4-profile cross-platform build matrix (`linux`, `darwin`, `windows`, `cgo`) and returns `BuildContext`. |
| **35** | `garuda.hygiene_report` | Static Code Hygiene | `workspace`, `limit`, `tenant_id` | Runs the side-effect-free hygiene analyzer to report unreferenced exported symbols and duplicate names. |

### 7.3 The Nine-Surface Software Intelligence Instrument (`/dashboard` PWA)

Served by `bin/garuda-api` (`internal/api/dashboard_handlers.go`, `dashboard_docs_talon.go`, `dashboard_audit_rulebook.go`) and rendered via zero-CDN offline assets (`dashboard.html`, `dashboard.js`, `dashboard.css`), the web console organizes software intelligence into nine specialized operational surfaces:

1. **Surface `01 — Overview Briefing`**: Displays workspace vital signs, indexed entity/relationship counts, 3-state epistemic distribution (`SUPPORTED` / `UNVERIFIED` / `CONTRADICTED`), current RFC 6962 Merkle block height, and direct navigation shortcuts to the Base Rulebook Studio, Risk Hub, and Financial Audit workbench.
2. **Surface `02 — Explore Index`**: Provides a high-speed searchable directory of all packages, structs, interfaces, functions, and methods ranked by inbound caller centrality, integrated with the global `⌘K` / `Ctrl+K` command palette.
3. **Surface `03 — Architecture Map`**: Renders a focused 1-hop or 2-hop SVG neighborhood graph (`[Upstream Callers] -> [Focal Symbol] -> [Downstream Dependencies & Policy Gate]`) paired with a synchronized 4-tab Inspector Drawer (`Overview`, `Callers & Blast Radius`, `Governing Rules`, `Source & Evidence`).
4. **Surface `04 — Docs & Base Rulebook`**: Visualizes the 4-tier Document-to-Code Lineage DAG, displays Document Inventory Cards, provides live ADR ingestion modals, and hosts the **Auto-Synthesized Base Rulebook Studio** with inline Markdown editing and 1-click human approval.
5. **Surface `05 — Talon, Criticality & Risk Hub`**: Features the live Talon Active Guard Terminal, 6-point IDE/Git hook status toggles, the `0–100` Architectural Risk Hub Leaderboard, the 3-Tier Surface Criticality table (`HIGH_SEVERE` / `MEDIUM` / `LOW`), the `BuildContext` provenance receipt banner, and the 4-profile Cross-Platform Build Matrix table.
6. **Surface `06 — Change Review Workbench`**: Enables interactive pre-merge simulation (`POST /api/v1/dashboard/workbench/simulate`), comparing `net_new_violations` against `baseline_violations` before a pull request is opened.
7. **Surface `07 — Evidence & Cryptographic Ledger`**: Exposes quarantined runtime contradictions (`CONTRADICTED`), OpenTelemetry span correlation status, and interactive 1-click RFC 6962 Merkle inclusion proof verification.
8. **Surface `08 — Policy Studio`**: Displays active YAML policy predicates, match counts, severity badges (`BLOCK`, `REVIEW`, `WARN`), and remediation instructions.
9. **Surface `09 — Financial & Logic Audit`**: Houses the **Multi-Team Formula Disambiguation Workbench** (`0.00–1.00` team weights across Ops, Marketing, and Management with `✓ Take Human Call: Approve` and `🔒 Keep Frozen` controls) and the **Cross-Team Financial Discrepancy Detector** (`48.00 Crore INR` versus `4.80 Crore INR` AST/SQL lineage trace and interactive custom Crore audit simulator).

---

## 8. Empirical Validation Benchmarks, Reproducibility Protocols, and Known Boundaries

### 8.1 Ground-Truth Compiler Suite (`24/24`), Multi-Repo Corpora, and Runtime Quarantine (`10/10`)

Garuda maintains a strict separation between marketing assertions and reproducible engineering evidence (`EVIDENCE.md`). Every capability is backed by automated verification suites located in `test/benchmark/`, `scripts/mcp_verify.py`, `scripts/runtime_contradiction_test.py`, and `scripts/doc_verify_test.py`:

- **Ground-Truth Compiler & Determinism Suite (`make benchmark-strict`)**: Evaluates 24 synthetic and multi-module Go test fixtures (`test/benchmark/truth_fixtures/`) covering method receiver identity (`002-method-identity`), interface satisfaction (`003`, `009`, `010`), struct/interface embedding (`006-embedding`), generics, variadic functions, closures, and cross-module `go.work` workspaces (`024-go-work`). Garuda achieves **`24/24` (`100%` precision and `100%` recall, `0.0%` false-positive rate)** and passes `TestDoubleRunDeterminism` across all 24 cases.
- **Corpus A (14-Repository Compiler-Backed Go Validation)**: Indexed 14 production Go repositories into a unified workspace, extracting **4,888 canonical AST entities**, **5,894 typed relationships**, and **55 cross-repository module bridges**, anchored past **Merkle Block #898**.
- **Corpus B (9-Repository Polyglot Validation — Workspace `go-validation-10`)**: Indexed 9 repositories spanning Go (Tier 5 compiler-resolved at `1.0` confidence), Python (Tier 2 structural + cross-file alias/instance resolution at `0.95` confidence), and TypeScript (Tier 2 Tree-sitter + cross-file alias/instance resolution at `0.95` confidence), persisting **14,333 entities** and **40,956 relationships** in PostgreSQL.
- **Controlled Runtime Contradiction Quarantine (`scripts/runtime_contradiction_test.py`)**: Injected 10 unauthorized runtime driver and gRPC transport observations (`WithTransportCredentials`, `Channel`, `WithCodec`, `ApplyServerOptions`, `SetExtraHeader`, `test_default_bool`, `XRevRangeN`, `SetPickedCluster`, `StaticMethod`, `ApplyDefaultsWithPoolSize`) over authenticated `POST /api/v1/telemetry/spans` (`1.8 ms` p95 admission latency). All **10 out of 10 (`100%`)** were correlated to static entities and quarantined as `CONTRADICTED`.

### 8.2 Explicit Validation Boundaries and Future Engineering Roadmap

In accordance with Garuda's epistemic honesty contract (`docs/SPECS.md` §22), four technical boundaries are explicitly documented for internal engineers and external evaluators:

1. **Static Resolution Authority by Language**: Go relationships carry Tier 5 compiler-backed authority (`go/types` at `1.0` confidence). Python and TypeScript relationships carry Tier 2 structural and scope-aware authority (`0.95` confidence); highly dynamic constructs such as Python `eval()`/`importlib` or untyped JavaScript `Reflect` calls cannot be proven statically and rely on Plane 02 OpenTelemetry runtime span correlation.
2. **Cryptographic Integrity versus Semantic Completeness**: An RFC 6962 Merkle inclusion proof mathematically guarantees that a governance decision, `BuildContext` receipt, or policy evaluation has not been tampered with since commitment (`Law #6`); it does not independently prove that a human-authored policy rule covers every possible business edge case.
3. **Database-Level Row-Level Security (RLS) Completion**: While 100% of application SQL queries in `internal/store/` enforce `WHERE tenant_id = $1`, extending PostgreSQL native Row-Level Security policies from `harvested_decisions` to all remaining tables and adding database-level `BEFORE UPDATE OR DELETE` immutability triggers on `decision_revisions` remain scheduled hardening milestones.
4. **Next-Arc Language Expansion**: Compiler-backed analyzers for **Rust** (`rust-analyzer` / `syn` bridge), **Java** (bytecode and AST extractor), and **Elixir** are queued on the multi-language roadmap (`docs/MULTI_LANGUAGE_ROADMAP.md`).

---

## 9. References and Authoritative Source Repositories

- Garuda Primary Source Repository and Release Suite (`v1.9.1` / `v1.9.2`): [GitHub Repository: myshra777-ai/garuda](https://github.com/myshra777-ai/garuda)
- IETF Certificate Transparency and Merkle Hash Tree Standard: [RFC 6962 Certificate Transparency](https://datatracker.ietf.org/doc/html/rfc6962)
- Edwards-Curve Digital Signature Algorithm (Ed25519) Standard: [RFC 8032 EdDSA / Ed25519](https://datatracker.ietf.org/doc/html/rfc8032)
- Universally Unique Identifier (UUIDv5 SHA-1 Name-Based) Standard: [RFC 4122 UUID Specification](https://datatracker.ietf.org/doc/html/rfc4122)
- Model Context Protocol (MCP) JSON-RPC 2.0 Specification: [Model Context Protocol 2025-06-18 Specification](https://modelcontextprotocol.io/specification/2025-06-18)
- OpenTelemetry Distributed Tracing and Span Specification: [OpenTelemetry Specification](https://opentelemetry.io/docs/specs/otel/)
- Go Standard Library Type Checker and Compiler AST Packages: [Go Standard Library go/types](https://pkg.go.dev/go/types)
- PostgreSQL 16 Relational Database and Recursive CTE Documentation: [PostgreSQL 16 Official Documentation](https://www.postgresql.org/docs/16/)
