# Garuda Specifications

**Version:** 0.1.x  
**Document status:** Public reference specification  
**Scope:** Semantic analysis, evidence verification, governance, MCP integration, workspace intelligence, runtime correlation, dashboards, and IDE integration  
**Updated:** 2026-09-19

> This document describes Garuda as a verified state and governance plane for AI-assisted software development. It documents shipped behavior, validation boundaries, and planned capabilities. It is not a claim that Garuda is a general-purpose agent runtime, model-orchestration framework, or compliance certification.

---

## Table of Contents

1. [Product Boundary](#1-product-boundary)
2. [Core Concepts](#2-core-concepts)
3. [System Architecture](#3-system-architecture)
4. [Semantic Model](#4-semantic-model)
5. [Analyzer Capabilities](#5-analyzer-capabilities)
6. [Workspace Intelligence](#6-workspace-intelligence)
7. [Documentation and Claims](#7-documentation-and-claims)
8. [Runtime Verification](#8-runtime-verification)
9. [Policy Engine](#9-policy-engine)
10. [Trust and Merkle Layer](#10-trust-and-merkle-layer)
11. [MCP Surface](#11-mcp-surface)
12. [Multi-Agent Coordination](#12-multi-agent-coordination)
13. [Workspace Console](#13-workspace-console)
14. [VS Code Integration](#14-vs-code-integration)
15. [Tenant and Workspace Isolation](#15-tenant-and-workspace-isolation)
16. [Control Plane](#16-control-plane)
17. [CLI and HTTP Interfaces](#17-cli-and-http-interfaces)
18. [CI/CD Integration](#18-cicd-integration)
19. [Observability](#19-observability)
20. [Data and Persistence Model](#20-data-and-persistence-model)
21. [Security and Invariants](#21-security-and-invariants)
22. [Validation and Evidence](#22-validation-and-evidence)
23. [Operational Workflows](#23-operational-workflows)
24. [Known Boundaries](#24-known-boundaries)
25. [Roadmap](#25-roadmap)
26. [Versioning](#26-versioning)
27. [Glossary](#27-glossary)

---

## 1. Product Boundary

Garuda is a semantic software-intelligence and governance platform for teams using AI-assisted development workflows. It builds a shared model of code, relationships, documentation claims, runtime observations, policies, and governance decisions.

AI agents and developers query that model through MCP, CLI, HTTP, dashboards, CI, and the VS Code extension. Garuda is designed to answer questions such as:

- What exists in the workspace?
- What depends on this symbol?
- Who implements this interface?
- What documentation claims apply to this code?
- What policy rules apply to the current state?
- What runtime evidence supports or contradicts the static model?
- What would be affected by changing a symbol or decision?
- What was decided, under which policy version, and can that decision be independently verified?

### Product category

Garuda is a **verified state plane for AI-assisted software development**.

It is not an agent framework, coding agent, general-purpose code search engine, runtime observability replacement, or compliance certification. It can integrate with those categories and provide the semantic, governance, and evidence layer beneath them.

### Primary users

| User | Primary value |
|---|---|
| AI-enabled engineering teams | Ground agents in a shared semantic model rather than only local file context |
| Platform and architecture teams | Inspect dependencies, boundaries, policy violations, drift, and change impact |
| Security and governance teams | Record and verify policy decisions with evidence and cryptographic proofs |
| Engineering leaders | View workspace health, governance state, runtime evidence, and decision history |
| Developers in unfamiliar codebases | Navigate entities, callers, implementers, relationships, and claims faster |

---

## 2. Core Concepts

### Four artifact categories

Garuda connects four classes of artifacts:

| Artifact | Meaning | Examples |
|---|---|---|
| Intent | What an organization says should be true | Policies, requirements, ADRs, specifications |
| Code | What has been implemented | Packages, structs, interfaces, functions, methods, imports, calls |
| Runtime | What the system has been observed doing | Traces, spans, runtime observations, contradictions |
| Evidence | What supports a claim or decision | Source locations, relationship evidence, verification records, Merkle proofs |

### Three verification states

| State | Meaning |
|---|---|
| `SUPPORTED` | Available evidence supports the claim within the verified scope |
| `UNVERIFIED` | The claim lacks sufficient evidence; this does not mean it is false or broken |
| `CONTRADICTED` | Available evidence conflicts with the claim |

### Evidence authority

Garuda preserves an explicit evidence hierarchy:

```text
Compiler-backed AST and type evidence
              ↓
Persisted semantic relationships with source evidence
              ↓
Validated documentation claims
              ↓
Runtime observations and correlation results
              ↓
Confidence-scored heuristic relationships
```

The exact authority order depends on the operation. Heuristic relationships must remain labeled and must not silently override stronger evidence.

### Maturity levels

| Status | Meaning |
|---|---|
| Production (GA) | Shipped, tested, and suitable for authoritative workflows within the documented scope |
| Beta | Functional and validated, but subject to scale, language, deployment, or release-boundary constraints |
| Experimental | Available for exploration; confidence-scored or heuristic behavior requires independent validation |
| Upcoming | Designed or queued; not a current shipped capability |

---

## 3. System Architecture

```mermaid
flowchart TD
    DOCS["Documentation, ADRs, and policies"] --> INGEST["Ingestion and claim extraction"]
    CODE["Source repositories"] --> ANALYZE["AST and semantic analyzers"]
    RUNTIME["Telemetry and runtime observations"] --> VERIFY["Runtime correlation and verification"]
    ANALYZE --> GRAPH["Tenant/workspace semantic graph"]
    INGEST --> GRAPH
    VERIFY --> GRAPH
    GRAPH --> POLICY["Policy engine"]
    GRAPH --> MCP["MCP server"]
    GRAPH --> API["CLI and HTTP APIs"]
    GRAPH --> CONSOLE["Workspace Console"]
    GRAPH --> IDE["VS Code extension"]
    POLICY --> LEDGER["Merkle trust layer"]
    LEDGER --> PROOF["Independent verification"]
    MCP --> AGENTS["AI clients and agents"]
```

### Architectural layers

| Layer | Responsibility |
|---|---|
| Analyzer layer | Parse source, resolve types where supported, and emit entities and relationships |
| Knowledge layer | Store claims, entities, relationships, evidence, and verification states |
| Runtime layer | Ingest observations, correlate observations with semantic entities, and record contradictions |
| Policy layer | Evaluate declarative rules and produce governance outcomes |
| Trust layer | Anchor committed evaluations and decisions in Merkle structures |
| Integration layer | Expose MCP, CLI, HTTP, dashboard, CI, and IDE interfaces |
| Persistence layer | Store tenant, workspace, repository, graph, policy, runtime, session, and ledger state in PostgreSQL |

---

## 4. Semantic Model

### Entities

Entities represent identifiable software symbols and architectural objects, including:

- Packages.
- Structs.
- Interfaces.
- Functions.
- Methods.
- Fields.
- Repositories.
- Workspaces.
- External or unresolved references where the analyzer preserves them as graph objects.

Entities carry stable identifiers and source metadata. The analyzer preserves declaration kind, package context, file path, line span, language, repository, tenant, and workspace scope where available.

### Relationships

Relationships are typed rather than represented as unstructured links.

| Relationship | Meaning |
|---|---|
| `CALLS` | One function or method invokes another |
| `IMPORTS` | A package or module imports another package or module |
| `IMPLEMENTS` | A concrete type satisfies an interface |
| `INHERITS` | A type inherits from or structurally derives from another supported type |
| `EMBEDS` | A struct embeds another type |
| `REFERENCES` | A symbol refers to another semantic entity |
| Cross-repository bridge | A supported dependency crosses repository boundaries |

Each relationship should retain source evidence such as file and line location, derivation method, and resolution tier.

### Stable identity

Garuda uses canonical identities to avoid treating every textual rename as an unrelated object. The exact identity representation depends on the semantic model and language adapter; the public contract is that queries and persisted relationships operate on stable semantic identities rather than raw text matches alone.

### Resolution tiers

| Tier | Typical source | Authority |
|---|---|---|
| Tier 5 | Go compiler/type information | Highest supported static authority |
| Tier 2 | Structural Python and TypeScript analysis | Useful structural evidence; lower authority for dynamic behavior |
| Heuristic | SSA or dynamic call approximation | Confidence-scored and non-authoritative unless explicitly permitted |

---

## 5. Analyzer Capabilities

### Go analyzer

Go analysis is compiler-backed within the supported scope. Current validated capabilities include:

- Struct and field extraction.
- JSON, database, and validation tags.
- Source line spans.
- Pointer and slice flags.
- Pointer and value receiver disambiguation.
- Interface implementation matching.
- Generic type-parameter handling.
- Type alias versus type-definition distinction.
- Struct embedding and `EMBEDS` relationships.
- Variadic signatures.
- Nested closure and function-literal traversal.
- Cross-repository dependency resolution through the workspace model.

The supplied benchmark matrix reports production status for the named benchmark gates, including method identity, generics, aliases, embedding, variadic handling, and closure handling.

### Python analyzer

Python analysis is structural. Import relationships are more reliable than dynamic call resolution. Python results should be interpreted according to relationship type and corpus coverage rather than a single universal accuracy number.

### TypeScript analyzer

TypeScript analysis is structural and requires CGO and a C compiler in the documented build path. It supports structural extraction and relationship discovery within the supported analyzer scope.

### Current and queued languages

| Language | Status |
|---|---|
| Go | Production (GA), compiler-backed within supported scope |
| Python | Beta, structural |
| TypeScript | Beta, structural |
| Rust | Upcoming |
| Java | Upcoming |
| Elixir | Upcoming |

### Experimental call-graph behavior

Dynamic or heuristic call-graph results must carry confidence information. They are not equivalent to compiler-resolved relationships and should not be silently promoted into strict governance evidence.

---

## 6. Workspace Intelligence

A workspace is a logical, tenant-scoped collection of repositories and their semantic state. Garuda presents repositories as parts of one system rather than isolated codebases.

### Workspace views

The Workspace Console can expose:

- Repository, package, entity, and relationship totals.
- Language composition.
- Architectural hubs ranked by graph centrality.
- Recent evidence and runtime observations.
- Policy enforcement state.
- Contradictions and unverified claims.
- Documentation drift.
- Decision history.
- Interactive topology and community exploration.

### Cross-repository intelligence

Garuda supports cross-repository relationships for validated multi-module workspaces. The current capability has been exercised across multi-repository Go corpora and the `go-validation-10` workspace.

The current operational contract is:

- Supported repository relationships are persisted in the workspace model.
- Cross-repository bridges can be surfaced in briefing and dashboard views.
- The quality of a bridge depends on language, import resolution, repository configuration, and analyzer maturity.
- Cross-language resolution and larger-scale quality gates remain separate validation concerns.

### Example workspace snapshot

A validated workspace view has displayed values such as:

- 9 repositories.
- 1,567 packages.
- 14,333 entities.
- 40,956 relationships or graph claims.
- 2 cross-repository bridges.

These values describe a particular snapshot, not product limits.

---

## 7. Documentation and Claims

### Supported document formats

Garuda can ingest Markdown, ADR, and plain-text documents according to the documented format contract.

A line is eligible for normative extraction when:

1. It contains a supported modal term such as `MUST`, `SHOULD`, `MUST NOT`, or `SHALL`.
2. It occurs under a heading suggesting normative content, such as `Decision`, `Requirement`, `Constraint`, `Policy`, or `Specification`.

Free-form prose, tables, and code blocks are intentionally excluded unless they satisfy the parser contract.

### Claim lifecycle

```text
Document discovery
      ↓
Normative line extraction
      ↓
Claim normalization
      ↓
Semantic graph correlation
      ↓
Runtime correlation where available
      ↓
SUPPORTED / UNVERIFIED / CONTRADICTED
```

### Drift directions

Garuda can inspect both directions of drift:

- Documentation claims without matching implementation evidence.
- Code entities and relationships without corresponding documentation claims.
- Static claims whose runtime observations are missing.
- Static or documented expectations contradicted by runtime observations.

The system must not interpret undocumented code as automatically invalid. It reports the evidence gap and leaves the governance decision to policy or human review.

---

## 8. Runtime Verification

Garuda accepts runtime observations and correlates them with semantic entities. The runtime path can support:

- Trace and span ingestion.
- Entity association.
- Runtime evidence records.
- Contradiction creation.
- Verification-state changes.
- Policy predicates such as `verification_missing`.
- Dashboard presentation of recent evidence and contradictions.

### Example runtime workflow

```text
OpenTelemetry or runtime source
          ↓
Runtime observation ingestion
          ↓
Entity and operation correlation
          ↓
Evidence or contradiction record
          ↓
Policy evaluation
          ↓
ALLOW / WARN / REVIEW / BLOCK
```

### Current boundary

The runtime verification path is implemented and exercised with controlled and synthetic evidence. Autonomous, large-scale, continuously scheduled correlation, dead-code detection, and unused-import detection are part of the next capability arc.

The correct interpretation of `SUPPORTED` is therefore scope-dependent: it means the current evidence supports the claim under the active verification path. It does not mean every claim in every workspace has been observed in production.

### Contradictions

A contradiction represents an evidence conflict, such as a runtime operation that violates a statically or normatively established expectation. Contradictions can appear in workspace dashboards, attention panels, policy evidence, and IDE diagnostics.

---

## 9. Policy Engine

Policies are declarative rules evaluated against a tenant and workspace.

### Policy structure

A policy may define:

- Identifier.
- Version.
- Title.
- Priority.
- Language or domain scope.
- Authority.
- Predicates.
- Decision outcome.
- Human-readable reason.

### Predicates

| Predicate | Meaning |
|---|---|
| `entity_exists` | Matching entities exist |
| `claim_exists` | Matching relationships or claims exist |
| `contradiction_exists` | A contradiction exists in the applicable scope |
| `verification_missing` | Required verification evidence is absent |
| `language_matches` | The scope contains a target language |

### Outcomes

| Outcome | Meaning |
|---|---|
| `ALLOW` | No blocking condition was found |
| `WARN` | A non-blocking condition is reported |
| `REVIEW` | Human or specialist review is required |
| `BLOCK` | The policy violation should prevent the governed action or merge |

### Evaluation modes

| Mode | Behavior |
|---|---|
| Preview | Evaluates without persisting or anchoring the result |
| Committed | Persists the evaluation and anchors it to the Merkle ledger; requires `--save` |
| Reconcile | Opt-in state reconciliation that retires active policies whose YAML files are absent; requires `--save` |

### Authority attribution

Policies can carry an authority such as a team or organizational owner. This preserves who is responsible for the rule and provides context for review and audit workflows.

---

## 10. Trust and Merkle Layer

### Purpose

The trust layer provides an independently verifiable record of persisted decisions and evidence commitments.

### What it proves

The ledger can prove that:

- A decision or evaluation was recorded in a specific tenant ledger.
- The record was included at a specific block height.
- The recorded content has not changed since inclusion, assuming the verifier trusts the published root and verification algorithm.
- The stored inclusion proof can be independently re-derived and checked.

### What it does not prove

The ledger does not prove that:

- The semantic model is complete.
- The analyzer is correct for every language or relationship type.
- The underlying code is safe.
- The runtime corpus represents every production path.
- The policy itself is substantively correct.

### Verification commands

```bash
garuda policy verify <evaluation-id>
garuda verify
```

The first verifies an individual policy evaluation or anchor. The second verifies the broader ledger and decision-chain integrity exposed by the CLI.

### Merkle terminology

Garuda uses the terms **tamper-evident**, **cryptographically anchored**, and **independently verifiable**. These terms describe the proof guarantee without claiming that every surrounding operational system is impossible to compromise.

### Dashboard presentation

The Evidence and Trust view can display:

- Ledger status.
- Block height.
- Current root.
- Parent root or genesis marker.
- Observation time.
- Verification state.
- Policy evaluation evidence.
- Anchor verification controls.

---

## 11. MCP Surface

### Transport

Garuda's MCP server uses line-delimited JSON-RPC 2.0 over standard input and output.

The current initialization contract reports MCP protocol version `2025-06-18`.

### Tool catalog

| Group | Tool | Purpose | Access |
|---|---|---|---|
| Session | `garuda.briefing` | Workspace state, trust anchor, scale, hubs, policies, contradictions, and recent changes | Read-only |
| Graph | `garuda.entities` | List semantic entities | Read-only |
| Graph | `garuda.find_entity` | Find entities by name, kind, package, or path | Read-only |
| Graph | `garuda.inspect` | Inspect one entity and relationships | Read-only |
| Graph | `garuda.neighbors` | List one-hop inbound and outbound connections | Read-only |
| Graph | `garuda.subclasses` | Find inheritance or embedding relationships | Read-only |
| Graph | `garuda.implementers` | Find interface implementers | Read-only |
| Graph | `garuda.blast_radius` | Compute graph-visible impact of changing a symbol | Read-only |
| Graph | `garuda.query_claims` | Query documentation claims related to a subject | Read-only |
| Graph | `garuda.query` | Query the knowledge graph using natural language | Read-only |
| Governance | `garuda.policy.list` | List tenant policies | Read-only |
| Governance | `garuda.policy.evaluate` | Preview policy outcomes | Read-only; preview by design |
| Governance | `garuda.verify_policy_evaluation` | Verify a persisted evaluation proof | Read-only |
| Governance | `garuda.governance.status` | Aggregate governance health | Read-only |
| Governance | `garuda.check_drift` | Report documentation-to-code drift | Read-only |
| Governance | `garuda.get_lineage` | Retrieve decision lineage | Read-only |
| Governance | `garuda.get_impact` | Compute impact of a decision change | Read-only |
| Governance | `garuda.detect_contradictions` | List unresolved contradictions | Read-only |
| Governance | `garuda.propose_decision` | Create a budget-checked draft | Mutating |
| Coordination | `garuda.handoff` | Transfer work and create a checkpoint | Mutating |
| Coordination | `garuda.resume` | Restore and consume a checkpoint | Mutating |

### Mutation semantics

The mutating tools are:

- `garuda.propose_decision`.
- `garuda.handoff`.
- `garuda.resume`.

`garuda.policy.evaluate` is a preview by default and does not persist. CLI committed evaluation requires `--save`.

### Tool distinction

- `garuda.blast_radius` answers what may be affected if a symbol changes, using the semantic graph.
- `garuda.get_impact` answers what may be affected if a decision changes, using decision lineage.

### Client validation

The MCP server has been tested against:

| Client | Scope |
|---|---|
| Cursor | IDE MCP integration |
| Claude Desktop | Desktop MCP integration |
| Codex CLI | Independent CLI MCP integration |
| Reference verifier | Dependency-free protocol, tool, and negative-path checks |

Current reference verification:

```text
═══ 38/38 checks passed ═══
```

The reference verifier checks initialization, protocol version, tool discovery, entity lookup, graph relationships, policy-proof failures, blast-radius output, handoff and resume failure paths, unknown-tool errors, and clean shutdown.

---

## 12. Multi-Agent Coordination

Garuda coordinates agent work at the state and task-transfer layer. It does not itself provide model inference or general agent-process orchestration.

### Handoff

`garuda.handoff` transfers task ownership between a source agent and a target agent. The transaction can:

1. Validate task and agent preconditions.
2. Lock the relevant task and checkpoint state.
3. Create a checkpoint.
4. Record the handoff.
5. Transition agent and task state.
6. Commit atomically.

### Resume

`garuda.resume` restores an active checkpoint and consumes it transactionally. A second attempt to consume the same checkpoint returns a structured `not_found` result rather than restoring it twice.

### Isolation level

The handoff and resume paths use Serializable transaction semantics and row locking where required by the store path. Concurrency behavior is part of the release verification contract.

### Sessions and tool calls

The current MCP product surface includes session and tool-call observability:

- MCP process session identity.
- Client and client-version metadata where available.
- Agent and workspace association.
- Tool-call duration.
- Tool-call status.
- Whitelisted argument summaries.
- Graceful and inferred session closure state.

Free-form query and business-context fields should not be recorded in argument summaries.

---

## 13. Workspace Console

The Workspace Console is the tenant-facing user dashboard.

### Workspace

The Workspace view presents:

- Repository count.
- Package count.
- Entity count.
- Relationship count.
- Language composition.
- Active policy count.
- Drift or quarantined-contract indicators.
- Architectural hubs.
- Recent evidence.
- Workspace health and trust status.

### Agents

The Agents view presents MCP activity, including active and recent sessions, tool-call activity, and agent coordination state.

### Governance

The Governance view presents:

- Policy decision summaries.
- `ALLOW`, `WARN`, `REVIEW`, and `BLOCK` counts.
- Documentation claim totals.
- Supported, unverified, and contradicted claims.
- Knowledge drift.
- Contradictions and attention items.
- Merkle and cryptographic trust state.

### Decisions

The Decisions view presents:

- Decision history.
- Decision status.
- Policy and scope metadata.
- Ancestors and descendants.
- Revision chains.
- Merkle or anchor status.
- Detail drawers for evidence and lineage.

### Graph

The Graph view presents:

- Interactive topology.
- Entity communities.
- Repository and package relationships.
- Architectural hubs.
- Level-based navigation.
- Local and full-graph exploration.
- Relationship highlighting and inspection.

---

## 14. VS Code Integration

Garuda provides a VS Code extension under `vscode-extension/`.

### Extension identity

| Field | Value |
|---|---|
| Display name | Garuda Epistemic Architecture Shield |
| Version | `0.1.0` |
| Primary scope | Go workspace architecture and verification assistance |
| Activation | Go language, workspace containing `go.mod`, or startup completion |
| Main entry point | `out/extension.js` |

### Commands

| Command | Purpose |
|---|---|
| `garuda.refreshState` | Refresh ledger and verification state |
| `garuda.openVisualizer` | Open graph visualizer |
| `garuda.reanalyzeWorkspace` | Re-index workspace AST |

### Extension views

- Cryptographic Ledger.
- Quarantined Contradictions.

### Configuration

| Setting | Purpose |
|---|---|
| `garuda.executablePath` | Path to the Garuda CLI binary |
| `garuda.databaseUrl` | PostgreSQL state-store connection string |
| `garuda.daemonUrl` | Unified daemon HTTP API URL |
| `garuda.enableHoverBlastRadius` | Enable blast-radius and dependency information on Go symbol hover |

### Editor diagnostics

The extension can surface analyzer-backed architecture and runtime issues through the VS Code Problems view. The supplied validation evidence shows diagnostics for quarantined runtime contradictions and violations such as `ARCH_DRIFT_001`.

The extension also supports architecture-oriented hover and graph workflows, including blast-radius and dependency information where enabled.

### Verification boundary

The extension should be described as **tested and shipped for editor-integrated diagnostics and exploration**. “Real-time” should be interpreted according to the configured refresh, reanalysis, daemon, and diagnostic-update paths. It should not be read as a guarantee that every repository mutation is continuously analyzed without a refresh or analysis event.

---

## 15. Tenant and Workspace Isolation

Garuda models tenants, users, memberships, workspaces, repositories, entities, claims, policies, runtime evidence, decisions, and ledger state with explicit scope fields.

### Tenant scope

Tenant-scoped data is resolved through authenticated or configured tenant context. Queries against tenant-owned resources must include the applicable tenant boundary.

### Workspace scope

Workspace-scoped data is resolved through the active workspace and tenant. Workspace relationships are used to prevent one workspace's semantic state from being incorrectly presented as another workspace's state.

### Shared workspaces

A workspace can be shared by authorized developers and agents. The workspace is the collaboration boundary for repositories, graph state, policies, evidence, and decisions.

### Access-control principles

- Resolve tenant before reading tenant-owned state.
- Resolve workspace within tenant scope.
- Preserve workspace identifiers in semantic and governance records.
- Reject or return structured failure for unauthorized scope resolution.
- Keep cross-tenant reads structurally unavailable through query scoping and access checks.
- Treat tenant and workspace isolation tests as release gates.

### Scope caveat

Isolation guarantees apply to the implemented authentication, API, CLI, MCP, dashboard, and store paths covered by the current release tests. New access paths must inherit the same scope-resolution contract rather than introducing default or implicit cross-scope behavior.

---

## 16. Control Plane

The Control Plane is an owner-facing operational surface and is distinct from the tenant-facing Workspace Console. It is not the primary public product surface documented for ordinary workspace users.

Where enabled, it can provide operational and tenant-level visibility with protected access and read-only database access for reporting paths.

### Separation of concerns

| Surface | Audience | Scope |
|---|---|---|
| Workspace Console | Tenant users, developers, agents | Current tenant and workspace |
| Control Plane | Owner/operator | Platform-level operational administration |

Internal roadmap, feature-flag, and beta-operations views are not treated as tenant-facing product capabilities.

---

## 17. CLI and HTTP Interfaces

### CLI

The CLI supports workflows for:

- Analysis.
- Workspace management.
- Repository management.
- Entity and graph inspection.
- Impact analysis.
- Semantic diff.
- Policy validation and evaluation.
- Evaluation display and verification.
- Documentation ingestion and verification.
- Status and ledger verification.
- CI operation.
- Benchmarks and governance judgment.
- Development daemon and dashboard operation.

Use command-specific help for the authoritative flags:

```bash
garuda <command> --help
```

### HTTP and OpenAPI

Garuda exposes HTTP handlers for supported workspace, authentication, dashboard, policy, runtime, handoff, budget, ledger, and operational workflows. `openapi.yaml` provides the machine-readable API contract for the documented HTTP surface.

### Automation boundary

Supported CLI, HTTP, MCP, and CI workflows are scriptable. A dashboard-only presentation is not automatically equivalent to an independently callable API; new dashboard actions should be backed by an explicit endpoint or MCP tool before being described as automation capabilities.

---

## 18. CI/CD Integration

Garuda can be integrated into build and review pipelines.

### Policy gate

```bash
./bin/garuda policy evaluate ./policies --fail-on BLOCK
```

`--fail-on` supports `BLOCK`, `REVIEW`, `WARN`, and `ALLOW` thresholds.

### Impact annotation

```bash
./bin/garuda impact <changed-symbol> --format github
```

### Semantic diff

```bash
./bin/garuda analyze . --save -o /tmp/before.json
./bin/garuda analyze . --save -o /tmp/after.json
./bin/garuda diff /tmp/before.json /tmp/after.json
```

### Recommended pipeline

1. Start or provision PostgreSQL.
2. Apply migrations.
3. Resolve the tenant and workspace.
4. Analyze and persist the repository state.
5. Validate policies.
6. Run preview or committed policy evaluation.
7. Fail on the configured decision threshold.
8. Verify documentation or runtime evidence where applicable.
9. Publish semantic diff, impact, and verification artifacts.

---

## 19. Observability

Garuda's observability surface includes product and platform evidence relevant to workspace operation.

### MCP activity

MCP sessions and tool calls can record:

- Client identity.
- Version metadata.
- Agent identity.
- Session lifecycle.
- Tool name.
- Duration.
- Success or error status.
- Whitelisted argument summary.

### Runtime evidence

Runtime observation records can include trace identifiers, span identifiers, parent spans, service names, operations, entity association, timing, and contradiction state.

### Dashboard indicators

The Workspace Console can present:

- Supported claims.
- Unverified claims.
- Contradictions.
- Runtime evidence.
- Policy attention items.
- Architectural hubs.
- Trust state.

### Privacy boundary

Free-form business context should not be included in whitelisted session argument summaries. Runtime and error data should be scoped according to tenant and operator access rules.

---

## 20. Data and Persistence Model

PostgreSQL is the system of record for Garuda's persistent semantic and governance state.

### Major data domains

| Domain | Representative data |
|---|---|
| Identity | Users, sessions, tenants, memberships |
| Workspace | Workspaces, repositories, workspace membership |
| Semantic graph | Entities, packages, claims, relationships, cross-repository bridges |
| Documentation | Documents, document claims, verification states |
| Runtime | Observations, spans, contradictions, correlation records |
| Governance | Policies, evaluations, violations, decisions, revisions, lineage |
| Agent coordination | Tasks, agent state, checkpoints, handoffs |
| Trust | Merkle roots, blocks, proofs, evidence records |
| Observability | MCP sessions, tool calls, operational records |
| Resource control | Tenant budgets and execution accounting |

### Migration discipline

Migrations define schema evolution. They should be applied in order, tested for idempotence, and verified against the active release. Duplicate numeric prefixes must be reviewed before automated migration execution.

### Scope discipline

Tables containing tenant- or workspace-owned data must preserve the relevant scope key and enforce it in reads and writes. Store methods are expected to keep scope resolution explicit rather than relying on caller assumptions.

---

## 21. Security and Invariants

Garuda's security posture is based on explicit invariants rather than on UI labels alone.

### Core invariants

- Tenant boundaries must be preserved on tenant-owned reads and writes.
- Workspace boundaries must be preserved on workspace-owned reads and writes.
- Storage presence does not automatically imply semantic visibility.
- Evidence and decisions must retain source and scope context.
- Committed governance records must be auditable and independently verifiable.
- Policy evaluation and anchoring must not silently downgrade failures into success.
- Heuristic evidence must remain distinguishable from compiler-backed evidence.
- Mutating agent-coordination operations must be transactional and idempotent where required.
- Revocation or invalidation state must dominate dependent exposure paths where applicable.

The canonical invariant contract is maintained in [`docs/invariants.md`](docs/invariants.md). This specification summarizes the model; the invariant document remains authoritative for exact rules and notation.

### Fail-closed behavior

When required integrity, scope, audit, or verification work fails, the governed operation should preserve the failure state rather than expose an unverified success. The exact behavior is operation-specific and must be tested at the store, API, CLI, MCP, and dashboard boundaries.

---

## 22. Validation and Evidence

### Reference verification

```bash
WORKSPACE=go-validation-10 python3 scripts/mcp_verify.py
```

Current reference result:

```text
═══ 38/38 checks passed ═══
```

### Workspace validation

Validated workspace evidence includes a multi-repository Go corpus with dashboard-visible scale, graph relationships, policy evaluations, runtime observations, contradictions, and Merkle state.

### MCP clients

MCP integration has been tested against Cursor, Claude Desktop, Codex CLI, and the dependency-free reference verifier.

### Dashboard evidence

The user-facing Workspace Console has been exercised across:

- Workspace intelligence.
- Graph exploration.
- Governance and policy enforcement.
- Evidence and cryptographic trust.
- Decision history.
- Agent/session activity.

### IDE evidence

The VS Code extension has been exercised with:

- Cryptographic ledger view.
- Contradiction view.
- Problems-panel diagnostics.
- Architecture and graph commands.
- Go symbol hover support.
- Workspace reanalysis and state refresh commands.

### Evidence discipline

Every capability status should be accompanied by:

- Release or commit identity.
- Test, benchmark, or reproducible command.
- Corpus or workspace scope.
- Known limitations.
- Whether the result is authoritative, beta, or heuristic.

---

## 23. Operational Workflows

### First local workspace

```bash
export DATABASE_URL="postgres://garuda:garuda@localhost:5432/garuda?sslmode=disable"
export GARUDA_WORKSPACE=my-workspace
./bin/garuda workspace create my-workspace
./bin/garuda analyze /path/to/repo --save
./bin/garuda summary
```

### Documentation verification

```bash
./bin/garuda docs ingest ./docs
./bin/garuda docs verify
```

### Policy preview and commit

```bash
./bin/garuda policy validate ./policies
./bin/garuda policy evaluate ./policies
./bin/garuda policy evaluate ./policies --save
```

### Policy proof verification

```bash
./bin/garuda policy show <evaluation-id>
./bin/garuda policy verify <evaluation-id>
```

### MCP verification

```bash
go build -o bin/garuda-mcp ./cmd/garuda-mcp
WORKSPACE=my-workspace python3 scripts/mcp_verify.py
```

### Impact review

```bash
./bin/garuda impact <entity-name>
```

### VS Code workflow

1. Build the extension with `npm run compile` from `vscode-extension/`.
2. Configure the Garuda executable and daemon/database settings.
3. Open a Go workspace containing `go.mod`.
4. Run `Garuda: Re-index Workspace AST`.
5. Inspect the Garuda Architecture Shield views.
6. Review Problems-panel diagnostics and hover impact information.
7. Refresh ledger and verification state after policy or runtime changes.

---

## 24. Known Boundaries

Garuda deliberately preserves uncertainty and capability boundaries.

- Go compiler-backed analysis is not equivalent to universal runtime knowledge.
- Python and TypeScript structural analysis has lower authority for dynamic behavior.
- Heuristic call-graph results are confidence-scored.
- A graph-visible blast radius is not a proof that every runtime dependency was discovered.
- `UNVERIFIED` is not the same as false, dead, or broken.
- A Merkle proof verifies recorded decision integrity, not semantic-model completeness.
- Runtime verification depends on the quality and coverage of the observation source.
- Cross-repository intelligence depends on repository configuration and resolvable dependency paths.
- Dashboard state is scoped to the selected tenant and workspace.
- Public product documentation excludes private operator-only Control Plane details unless explicitly marked.
- Upcoming autonomous verification, dead-code detection, unused-import detection, and shared agent memory are not current GA capabilities.

---

## 25. Roadmap

### Next capability arc

- Autonomous runtime correlation.
- Dead-code and unused-import detection exposed through MCP.
- Ponytail-oriented unused-symbol and over-scoped-file analysis.
- Rust, Java, and Elixir analyzers.
- Cross-agent shared memory.
- Memory compression for long-running agent workflows.
- Stale MCP-session cleanup.
- Deeper lint coverage for store-versus-DTO and contract mismatches.
- Expanded operational observability.

### Promotion rule

A roadmap capability should not be described as shipped until its implementation, persistence, isolation, failure-path behavior, and reproducible validation are complete for the target release.

---

## 26. Versioning

Garuda has multiple compatibility surfaces:

| Surface | Compatibility concern |
|---|---|
| Semantic model | Entity, relationship, claim, and identity representation |
| Analyzer version | Language resolution and evidence behavior |
| Database schema | Ordered migrations and persisted state |
| Merkle version | Canonical encoding, root, proof, and verification behavior |
| MCP protocol | JSON-RPC transport, protocol version, tools, schemas, and response shapes |
| CLI | Commands, flags, output formats, and exit codes |
| HTTP/OpenAPI | Routes, request schemas, response schemas, and auth behavior |
| VS Code extension | Commands, settings, views, diagnostics, and daemon compatibility |

Breaking changes should be documented with the affected surface, migration path, verification impact, and release identifier.

---

## 27. Glossary

| Term | Definition |
|---|---|
| Entity | A stable semantic object such as a package, type, function, method, or field |
| Relationship | A typed connection between semantic entities |
| Claim | A normalized statement extracted from documentation or another intent source |
| Evidence | Source, runtime, or cryptographic material supporting a claim or decision |
| Supported | Evidence supports the claim within the verified scope |
| Unverified | Evidence is insufficient to establish the claim |
| Contradicted | Evidence conflicts with the claim |
| Workspace | A logical group of repositories and semantic state |
| Tenant | An isolation and ownership boundary for users, workspaces, policies, evidence, and decisions |
| Blast radius | Graph-visible impact of changing a symbol |
| Decision impact | Lineage-visible impact of changing a governance decision |
| Policy evaluation | A run of policies against the current workspace state |
| Merkle root | Cryptographic summary of a ledger state |
| Block | A ledger position containing anchored records or state |
| Inclusion proof | Evidence that a record belongs to a Merkle root or block |
| Runtime observation | A recorded observation from an execution or telemetry source |
| Contradiction | A detected conflict between available evidence and an expected claim or relationship |
| Tier 5 | Compiler-backed semantic resolution tier used for supported Go analysis |
| Tier 2 | Structural analysis tier used for supported Python and TypeScript paths |
| MCP | Model Context Protocol interface used by AI clients to query and operate on Garuda |
| Checkpoint | Persisted state used to transfer or resume agent work |
| Handoff | Transactional transfer of task ownership between agents |
| Control Plane | Operator-facing platform surface distinct from the tenant Workspace Console |

---

## Evidence and Documentation Index

- [`README.md`](README.md) — Product overview and positioning.
- [`PLAYBOOK.md`](PLAYBOOK.md) — Installation, setup, workflows, and command reference.
- [`EVIDENCE.md`](EVIDENCE.md) — Validation evidence and methodology.
- [`docs/CAPABILITIES.md`](docs/CAPABILITIES.md) — Analyzer and product capability matrix.
- [`docs/DOCUMENT_FORMATS.md`](docs/DOCUMENT_FORMATS.md) — Documentation ingestion contract.
- [`docs/invariants.md`](docs/invariants.md) — Canonical invariant contract.
- [`docs/adr/`](docs/adr/) — Architecture decision records.
- [`vscode-extension/`](vscode-extension/) — VS Code integration.
- [`scripts/mcp_verify.py`](scripts/mcp_verify.py) — Dependency-free MCP reference verifier.
- [`openapi.yaml`](openapi.yaml) — HTTP API contract.

---

<p align="center">
  <strong>Garuda</strong><br>
  Analyze · Verify · Govern
</p>