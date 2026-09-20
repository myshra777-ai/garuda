# Garuda Specifications

**Version:** 0.1.x  
**Document status:** Current-state reference draft  
**Updated:** 2026-09-20  
**Scope:** Verified repository capabilities, validation boundaries, and explicitly marked future work.

> This document describes Garuda as implemented and verified in the current repository state. A capability is not considered current merely because it appears in a roadmap, design document, dashboard label, or generated artifact.

## Status legend

| Status | Meaning |
|---|---|
| Current | Implemented and verified by code, tests, or reproducible commands |
| Beta | Implemented, but constrained by scope, language, scale, or deployment maturity |
| Experimental | Available with confidence or coverage limitations |
| Planned | Designed or queued; not a current capability |
| Historical | Retained for context and not a current product contract |

## Contents

1. [Product boundary](#1-product-boundary)
2. [System architecture](#2-system-architecture)
3. [Semantic model](#3-semantic-model)
4. [Analyzer capabilities](#4-analyzer-capabilities)
5. [Workspace intelligence](#5-workspace-intelligence)
6. [Documentation and claims](#6-documentation-and-claims)
7. [Runtime verification](#7-runtime-verification)
8. [Policy engine](#8-policy-engine)
9. [Trust and Merkle layer](#9-trust-and-merkle-layer)
10. [MCP surface](#10-mcp-surface)
11. [Agent coordination](#11-agent-coordination)
12. [Workspace Console](#12-workspace-console)
13. [VS Code integration](#13-vs-code-integration)
14. [Tenant and workspace isolation](#14-tenant-and-workspace-isolation)
15. [CLI and HTTP interfaces](#15-cli-and-http-interfaces)
16. [CI/CD integration](#16-cicd-integration)
17. [Observability](#17-observability)
18. [Data and persistence](#18-data-and-persistence)
19. [Security and invariants](#19-security-and-invariants)
20. [Hygiene](#20-hygiene)
21. [Validation and evidence](#21-validation-and-evidence)
22. [Known boundaries](#22-known-boundaries)
23. [Roadmap](#23-roadmap)
24. [Versioning](#24-versioning)
25. [Glossary](#25-glossary)

---

## 1. Product boundary

Garuda is a PostgreSQL-backed semantic software-intelligence and governance system. It analyzes supported source repositories, stores scoped semantic state, correlates documentation and runtime evidence, evaluates policies, exposes query and coordination surfaces, and provides cryptographic verification for persisted governance records.

Garuda is not an autonomous coding agent, model provider, general-purpose agent framework, or compliance certification. The current repository does not provide a production execution harness that autonomously modifies code, reanalyzes a worktree, evaluates policy, and merges changes.

### Current product questions

Garuda is designed to answer questions such as:

- What entities and relationships exist in a scoped workspace?
- Which entities call, implement, inherit from, embed, or depend on another entity?
- What documentation claims apply to a workspace?
- What policy outcomes apply to the current state?
- What runtime evidence supports or contradicts a static claim?
- What graph-visible entities are affected by changing a symbol?
- What decisions and revisions exist, and can a persisted decision be independently verified?

### Artifact categories

| Artifact | Current meaning | Examples |
|---|---|---|
| Intent | Statements of what should be true | Policies, ADRs, documented requirements |
| Code | Analyzed implementation state | Packages, types, functions, methods, imports, calls |
| Runtime | Observed execution evidence | Spans, runtime observations, contradictions |
| Governance | Decisions and policy outcomes | Evaluations, decisions, revisions, Merkle records |

### Verification states

| State | Meaning |
|---|---|
| `SUPPORTED` | Available evidence supports the claim within the verified scope |
| `UNVERIFIED` | Evidence is insufficient; this is not proof of failure |
| `CONTRADICTED` | Available evidence conflicts with the claim |

---

## 2. System architecture

```text
Source repositories ── analyze ──> Semantic graph ── query ──> CLI / MCP / HTTP / Console
        │                              │
        └──── documentation claims ────┤
                                       │
Runtime observations ── correlate ────┤
                                       ↓
                                Policy evaluation
                                       ↓
                         Optional committed governance record
                                       ↓
                                Merkle verification
```

| Layer | Responsibility | Current status |
|---|---|---|
| Analyzer | Extract entities and relationships | Current for supported Go, Python, and TypeScript paths |
| Store | Persist tenant, workspace, repository, graph, policy, runtime, and governance state | Current |
| Knowledge layer | Ingest and verify documentation claims | Current within documented formats |
| Runtime layer | Correlate observations and record contradictions | Current with controlled and synthetic evidence boundaries |
| Policy layer | Evaluate declarative policy outcomes | Current |
| Trust layer | Anchor and verify committed records | Current |
| Integration layer | CLI, MCP, HTTP, dashboard, CI, and VS Code surfaces | Mixed; see individual sections |
| Execution harness | Autonomous sandboxed code-edit loop | Planned; not current |

Garuda’s persistent source of truth is PostgreSQL. The pure Hygiene analyzer is deliberately side-effect free and operates on typed in-memory semantic data.

---

## 3. Semantic model

### Entities

Entities represent semantic objects such as packages, files, structs, interfaces, functions, methods, fields, repositories, and workspaces. Entity records include identifiers and source metadata where available, including kind, name, package, file, language, line range, repository, tenant, and workspace.

### Relationships

Relationships are typed graph claims. Current relationship families include:

| Relationship | Meaning |
|---|---|
| `CALLS` | A function or method invokes another function or method |
| `IMPORTS` | A package or module imports another package or module |
| `IMPLEMENTS` | A concrete type satisfies an interface |
| `INHERITS` | A supported inheritance or derivation relationship |
| `EMBEDS` | A struct embeds another type |
| `REFERENCES` | A semantic reference between entities |
| Cross-repository bridge | A supported dependency crossing repository boundaries |

Relationships may carry confidence, resolution status, resolution method, source file, line range, evidence hash, tenant, workspace, and repository metadata.

### Claims and evidence

A documentation claim is a normalized statement extracted from an eligible source. A relationship is evidence about code structure. A runtime observation is evidence from execution. A policy evaluation combines applicable predicates and produces an outcome with evidence references.

Evidence authority is operation-specific. Compiler-backed resolution is stronger than name-only or structural heuristics, and heuristic results must remain identifiable as such.

### Identity and scope

Queries and persisted records use semantic identifiers and explicit tenant/workspace scope. The exact identifier format is an implementation contract of the analyzer and store layers; callers must not substitute display names for scoped identity when an ID is available.

---

## 4. Analyzer capabilities

### Go

Go analysis is compiler-backed within the supported scope. Current analyzer and benchmark coverage includes entities, signatures, receiver identity, type information, interfaces, embedding, generic forms, aliases, variadic signatures, closures, source locations, and call relationships.

Go call relationships can be resolved through `go/types` and carry resolution metadata. Unresolved or heuristic paths must not be presented as equivalent to authoritative compiler resolution.

### Python

Python support is structural. It can provide useful source and relationship information, but dynamic dispatch, imports generated at runtime, and other runtime behavior are outside the authority of purely structural analysis.

### TypeScript

TypeScript support is structural and uses the repository’s documented build requirements, including CGO and a C compiler where required. The capability is not equivalent to complete runtime or dynamic call-graph knowledge.

### Language status

| Language | Status |
|---|---|
| Go | Current; compiler-backed within supported scope |
| Python | Beta; structural |
| TypeScript | Beta; structural |
| Rust | Planned |
| Java | Planned |
| Elixir | Planned |

### Analysis persistence

The existing `garuda analyze` CLI can analyze repositories and, with the applicable save options, persist semantic results. The current Hygiene command does not trigger analysis or refresh the graph; it reads the already indexed workspace state.

---

## 5. Workspace intelligence

A workspace is a tenant-scoped logical grouping of repositories and semantic state. Workspace queries include repositories, packages, entities, relationships, documentation claims, policies, runtime evidence, decisions, and graph-visible impact.

Current workspace surfaces include:

- Workspace and repository summaries.
- Entity and relationship inspection.
- Search and graph exploration.
- Architectural hubs and centrality summaries.
- Cross-repository bridge summaries.
- Briefing and governance status.
- Impact and blast-radius queries.
- Documentation drift and contradiction views.

### Cross-repository capability

Cross-repository Go dependency resolution is implemented through the workspace model and has been exercised against multi-repository validation data. It remains subject to repository configuration, analyzer coverage, workspace-boundary validation, and scale limits. It is not a universal cross-language dependency guarantee.

### Blast radius versus decision impact

- `garuda.blast_radius` asks what graph-visible entities may be affected if a symbol changes.
- `garuda.get_impact` asks what decision-lineage entities may be affected if a governance decision changes.

These are different evidence sources and must not be treated as interchangeable.

---

## 6. Documentation and claims

Garuda can ingest supported Markdown, ADR, and plain-text documentation. The document parser extracts eligible normative statements according to the contract in [`docs/DOCUMENT_FORMATS.md`](docs/DOCUMENT_FORMATS.md).

The claim lifecycle is:

```text
discover → extract → normalize → correlate → verify
```

Claims can become `SUPPORTED`, `UNVERIFIED`, or `CONTRADICTED` depending on available evidence. Missing documentation does not automatically prove that implementation is invalid, and missing runtime evidence does not automatically prove that a path is absent.

Documentation-to-code and code-to-documentation drift are distinct queries. Runtime contradictions are a separate evidence class from static drift.

---

## 7. Runtime verification

The repository contains runtime observation and correlation paths for controlled and synthetic evidence. These paths can associate observations with semantic entities, record verification changes, and create contradiction records.

Runtime verification can support:

- Trace and span ingestion.
- Entity correlation.
- Runtime observation storage.
- Contradiction registration.
- Verification-state updates.
- Governance and dashboard reporting.

The current evidence does not justify claiming universal production coverage, continuous autonomous correlation, or complete dynamic behavior discovery. `SUPPORTED` means supported within the applicable workspace, source, and verification scope.

---

## 8. Policy engine

Policies express organizational intent as declarative rules over a tenant and workspace.

### Predicate families

Current policy predicates include:

- `entity_exists`.
- `claim_exists`.
- `contradiction_exists`.
- `verification_missing`.
- `language_matches`.

### Outcomes

| Outcome | Meaning |
|---|---|
| `ALLOW` | No applicable blocking condition was found |
| `WARN` | A non-blocking concern was found |
| `REVIEW` | Additional review is required |
| `BLOCK` | The evaluated policy requires the governed action to stop or be rejected |

### Preview and committed evaluation

CLI policy evaluation is preview by default. Add `--save` to persist and anchor the evaluation. `--reconcile` retires active policy rows whose YAML source is absent and requires `--save`.

The MCP `garuda.policy.evaluate` tool is preview-only by design and does not persist or anchor a decision.

### CLI examples

```bash
garuda policy validate ./policies
garuda policy evaluate ./policies
garuda policy evaluate ./policies --save
garuda policy evaluate ./policies --fail-on-block
```

Policy evaluation is not itself a merge engine. An `ALLOW` result does not authorize arbitrary automatic merging unless a separate integration explicitly defines and enforces that behavior.

---

## 9. Trust and Merkle layer

The Merkle layer anchors persisted governance records and supports independent verification of their inclusion and integrity.

It can establish that a recorded payload was included in a particular ledger state and that the proof recomputes under the applicable canonical encoding and verification rules.

It does not establish that:

- The semantic graph is complete.
- The analyzer is correct for every language.
- The policy is substantively correct.
- The source code is safe.
- Runtime observations cover every production path.

Useful commands include:

```bash
garuda policy verify <evaluation-id>
garuda verify
```

The canonical trust and invariant contracts are documented in the ADRs and [`docs/invariants.md`](docs/invariants.md).

---

## 10. MCP surface

### Transport and initialization

Garuda exposes MCP over line-delimited JSON-RPC on standard input and output. The current verified initialization response uses protocol version `2025-06-18`.

### Current tools

The current catalog contains 21 tools across session, graph, governance, and coordination groups.

| Group | Tools |
|---|---|
| Session context | `garuda.briefing` |
| Graph exploration | `garuda.entities`, `garuda.find_entity`, `garuda.inspect`, `garuda.neighbors`, `garuda.subclasses`, `garuda.implementers`, `garuda.blast_radius`, `garuda.query_claims`, `garuda.query` |
| Governance | `garuda.policy.list`, `garuda.policy.evaluate`, `garuda.verify_policy_evaluation`, `garuda.governance.status`, `garuda.check_drift`, `garuda.get_lineage`, `garuda.get_impact`, `garuda.detect_contradictions`, `garuda.propose_decision` |
| Coordination | `garuda.handoff`, `garuda.resume` |

The precise JSON schemas are defined by the server’s tool registry and should be discovered with `tools/list`; clients should not assume that a prose document is the schema authority.

### Mutation boundary

The current mutating tools are:

- `garuda.propose_decision`.
- `garuda.handoff`.
- `garuda.resume`.

`garuda.policy.evaluate` is a preview operation. Hygiene is not currently an MCP tool.

### Verification

The dependency-free reference verifier currently reports:

```text
38/38 checks passed
```

The verification suite covers initialization, tool discovery, protocol behavior, representative graph queries, negative paths, handoff/resume responses, unknown-tool behavior, and clean shutdown. It is not a complete proof of every tool’s semantic correctness.

Verified client categories include Cursor, Claude Desktop, Codex CLI, and the in-repository reference implementation.

---

## 11. Agent coordination

Garuda supports task and checkpoint coordination through transactional handoff and resume paths. It does not provide model inference, agent planning, sandbox execution, or autonomous code merging.

`garuda.handoff` transfers task ownership and creates a checkpoint under the applicable store transaction. `garuda.resume` restores and consumes a checkpoint; replaying a consumed checkpoint returns a structured failure rather than restoring it twice.

The handoff path uses Serializable transaction semantics and has regression coverage for repeated handoffs and happy-path MCP invocation.

The proposed autonomous execution harness is not part of the current Garuda product contract. It remains a future external orchestrator that must be designed around verified MCP contracts, sandbox boundaries, fresh-analysis guarantees, and explicit merge authorization.

---

## 12. Workspace Console

The tenant-facing dashboard is served at `/dashboard` under session-cookie authentication in the current web application.

Current dashboard families include:

- Workspace statistics and repository state.
- Search and graph exploration.
- Governance and policy state.
- Evidence and verification state.
- Decision history and lineage.
- MCP session activity where the corresponding data is available.

The dashboard is a presentation surface over scoped API and store queries. A dashboard label is not, by itself, evidence that an underlying metric or workflow is fully instrumented.

---

## 13. VS Code integration

The repository contains a VS Code extension under `vscode-extension/` for Garuda workspace exploration and architecture-oriented diagnostics.

The extension includes commands and views for workspace refresh, graph visualization, ledger state, and contradiction or architecture diagnostics according to the current extension manifest and compiled surface.

The extension is not an autonomous code-writing agent. It depends on the configured Garuda CLI, daemon, database, or MCP/API path appropriate to the installed version.

Current documentation must distinguish shipped commands from planned diagnostic expansion. Do not describe every future Hygiene finding as an existing IDE diagnostic.

---

## 14. Tenant and workspace isolation

Tenant isolation is enforced at the server and store boundaries, not by environment variables alone.

The required pattern is:

```text
resolve tenant
    ↓
resolve workspace within tenant
    ↓
apply tenant/workspace predicates to every scoped query
    ↓
return only authorized scoped state
```

Workspace names are not authorization boundaries. Same-named workspaces in different tenants must remain distinct. Cross-tenant reads must not return another tenant’s workspace, entities, relationships, policies, decisions, or runtime evidence.

The current codebase includes tenant and workspace isolation tests across API, store, benchmark, MCP, and workspace-resolution paths. New surfaces must reuse the established scope-resolution contract.

---

## 15. CLI and HTTP interfaces

### CLI families

The CLI contains commands for:

- Workspace and repository management.
- Repository analysis and semantic persistence.
- Entity, graph, and inspection queries.
- Summary, briefing, impact, and semantic diff workflows.
- Documentation ingestion and verification.
- Policy validation, preview, persistence, and proof verification.
- CI checks and contradiction gates.
- Agent coordination.
- Dashboard or development operation.
- Report-only static Hygiene.

Use command help as the authoritative flag reference:

```bash
garuda <command> --help
```

### HTTP and OpenAPI

The HTTP surface contains authentication, dashboard, workspace, graph, decision, policy, runtime, checkpoint, handoff, resume, and operational routes according to the current server and `openapi.yaml`.

The exact route contract belongs to the generated or maintained OpenAPI and handler definitions. Do not infer route availability from dashboard navigation alone.

---

## 16. CI/CD integration

Current CI-oriented capabilities include semantic analysis, policy gating, contradiction checks, and impact-oriented reporting.

| Capability | Current interface |
|---|---|
| Breaking-change gate | `garuda ci` with `--block-on-break` |
| Policy threshold gate | `garuda policy evaluate` with `--fail-on-block` |
| Contradiction gate | `garuda ci check` with `--fail-on-contradiction` |
| Impact reporting | CLI impact and graph reports; provider-specific PR annotations are not part of the verified current contract |

Example:

```bash
./bin/garuda ci check --fail-on-contradiction=true
./bin/garuda policy evaluate ./policies --fail-on-block
```

Do not use `--fail-on BLOCK` or `--format github` as current examples unless those flags are implemented and verified in a future release.

---

## 17. Observability

Current observability spans product evidence, runtime records, agent coordination, and server diagnostics. The exact availability depends on the deployed path and migrations applied.

Supported or implemented observability areas include:

- Request IDs and structured logs where configured.
- Runtime observations and correlation metadata.
- Contradiction records.
- MCP session and tool-call data where the corresponding tables and write path are enabled.
- Dashboard evidence and governance state.
- Error and query instrumentation where present in the deployment.

The system should not be documented as collecting arbitrary agent prompts or free-form business context. Whitelisted argument summaries and scoped operational records are the safer contract.

---

## 18. Data and persistence

PostgreSQL stores the persistent truth substrate. Major domains include:

- Tenants, users, sessions, and memberships.
- Workspaces and repositories.
- Entities and claims.
- Cross-repository edges.
- Documents and document claims.
- Runtime observations and contradictions.
- Policies, evaluations, decisions, revisions, and lineage.
- Agent tasks, checkpoints, and handoffs.
- Merkle roots, blocks, evidence, and proofs.
- Budgets and operational records.

Migration order and idempotency matter. Existing migration naming must be reviewed where numeric prefixes are duplicated.

Hygiene findings are currently ephemeral CLI results. There is no current persisted Hygiene finding table, baseline store, suppression store, or observation lifecycle.

---

## 19. Security and invariants

The canonical invariant contract is [`docs/invariants.md`](docs/invariants.md). This document records the practical boundaries relevant to current interfaces.

- Scope is enforced at server/store boundaries.
- Tenant and workspace IDs must remain attached to scoped records.
- Mutating handoff and resume paths must preserve transaction semantics.
- Merkle verification must fail rather than silently validate altered content.
- Heuristic relationships and findings must remain identifiable as heuristic or advisory.
- Preview policy evaluation must not persist or anchor.
- New integrations must not bypass the existing scope and authority model.
- Environment variables select context but do not replace authorization.
- A report-only command must not mutate the graph or source tree.

### No external dependency for core integrity

The Merkle and semantic core do not require an LLM provider. MCP clients and future harnesses are external integration layers, not part of the cryptographic integrity primitive.

### Autonomous harness boundary

No current release guarantee is made for sandbox escape prevention, automatic remediation, automatic merge, loop prevention, or LLM reasoning correctness. Those belong to a separate execution-harness design and validation arc.

---

## 20. Hygiene

Hygiene is the current report-only static-analysis capability described in [`docs/hygiene.md`](docs/hygiene.md).

### Commands

```bash
garuda hygiene .
garuda hygiene . --json -o hygiene.json
garuda ponytail .
```

`garuda ponytail` is a deprecated compatibility alias. Both commands use the same pure `internal/hygiene` analyzer and produce equivalent reports.

### Findings

| Finding | Current meaning | Not proof of |
|---|---|---|
| Static unreferenced candidate | Non-package, non-file entity has no incoming relationship in the indexed graph | Dead code or safe removal |
| Duplicate symbol-name candidate | Non-empty symbol name appears in multiple packages | Duplicated implementation |
| Standard-library alternative | Name contains `Contains` or `Sort` and matches a naming heuristic | Replaceable implementation |

### JSON compatibility

The CLI preserves the existing top-level fields:

```text
dead_code
duplications
stdlib_alternatives
summary
total_entities
total_relationships
```

The internal category names are more precise than the legacy JSON field names. `dead_code` is retained as a compatibility field but contains advisory static unreferenced candidates.

### Scope and behavior

Hygiene reads the resolved tenant/workspace graph and performs pure in-memory classification. It does not refresh analysis, persist findings, write source files, create decisions, anchor records, suppress findings, create baselines, expose an MCP tool, or merge code.

Malformed untyped graph rows with missing, non-string, or empty `from`, `to`, or `type` fields are skipped by the CLI adapter. Findings are deterministically sorted. The analyzer does not mutate its input slices or external state.

### Reference measurement

A verified reference run against `go-validation-10` measured:

```text
Entities: 22905
Relationships: 40956
Static unreferenced candidates: 11919
Duplicate symbol-name candidates: 1480
Standard-library alternatives: 59
```

These are counts for one indexed snapshot, not defect counts or quality scores.

### Future Hygiene work

Planned or undecided work includes:

- Analysis freshness metadata.
- Read-only MCP exposure.
- Persisted observations.
- Suppression and baseline semantics.
- VS Code finding lifecycle.
- Higher-confidence unused-import and over-scoped-file analysis.

None of these should be inferred from the current CLI.

---

## 21. Validation and evidence

### Repository verification

The current Hygiene arc has been verified with:

```bash
gofmt -d ...
go test ./... -count=1
go build ./...
go vet ./...
```

The exact command list and current measured outputs should be recorded in `EVIDENCE.md` when the evidence corpus is updated.

### MCP verification

```bash
WORKSPACE=go-validation-10 python3 scripts/mcp_verify.py
```

The current recorded reference result is:

```text
38/38 checks passed
```

### Evidence discipline

A capability claim should identify:

- The code or command that provides it.
- The test or reproducible run that verifies it.
- The tenant, workspace, corpus, or fixture scope.
- Whether the claim is current, beta, experimental, planned, or historical.
- What the evidence does not establish.

---

## 22. Known boundaries

- A graph is not a complete runtime model.
- Missing incoming references are not proof of dead code.
- Repeated names are not proof of duplicated code.
- A naming heuristic is not a refactoring proof.
- `UNVERIFIED` is not false, dead, or broken.
- A policy `ALLOW` is not automatically a merge authorization.
- A Merkle proof does not prove semantic completeness.
- An MCP protocol pass does not prove every semantic result is correct.
- Tenant isolation must be enforced by the server and store, not configuration strings alone.
- A dashboard panel does not prove that every underlying metric is fully instrumented.
- The current system is not an autonomous execution harness.
- The current system does not provide universal language coverage or complete dynamic call-graph resolution.

---

## 23. Roadmap

The following are planned or require a separate design and verification arc:

1. Hygiene evidence and contract expansion.
2. Optional read-only Hygiene MCP exposure.
3. Analysis freshness and workspace reanalysis contract.
4. External read-only MCP execution harness.
5. Patch artifacts and disposable worktrees.
6. Sandboxed single-agent execution loop.
7. Durable harness state and replay.
8. Human review and explicit merge authorization.
9. Multi-agent orchestration after the single-agent path is reliable.
10. Suppression, baselines, and persisted Hygiene observations.
11. Expanded language analyzers and larger multi-repository benchmarks.

The execution harness must not bypass Garuda’s tenant, workspace, evidence, policy, and transaction boundaries.

---

## 24. Versioning

Garuda has independent compatibility surfaces:

| Surface | Compatibility concern |
|---|---|
| Semantic model | Entity, relationship, claim, and identity representation |
| Analyzer | Language resolution and evidence behavior |
| Database | Ordered migrations and persisted state |
| Merkle layer | Canonical encoding, roots, blocks, proofs, and verification |
| MCP | Protocol version, tool names, input schemas, and response shapes |
| CLI | Commands, flags, JSON fields, output, and exit codes |
| HTTP/OpenAPI | Routes, schemas, authentication, and response behavior |
| VS Code | Commands, settings, views, diagnostics, and daemon compatibility |
| Hygiene | Finding kinds, advisory semantics, report fields, and alias behavior |

Breaking changes should state which surface changed, how compatibility is handled, and how the change was verified.

---

## 25. Glossary

| Term | Definition |
|---|---|
| Entity | A typed semantic object such as a package, function, type, method, or field |
| Relationship | A typed graph connection between entities |
| Claim | A normalized statement extracted from intent or documentation |
| Evidence | Source, runtime, semantic, or cryptographic material supporting a claim or decision |
| Supported | Evidence supports a claim within a defined scope |
| Unverified | Evidence is insufficient to establish a claim |
| Contradicted | Evidence conflicts with a claim |
| Workspace | Tenant-scoped grouping of repositories and semantic state |
| Tenant | Isolation and ownership boundary for users, workspaces, policies, and evidence |
| Blast radius | Graph-visible impact of changing a symbol |
| Decision impact | Lineage-visible impact of changing a governance decision |
| Policy evaluation | Applying policies to a scoped workspace state |
| Merkle root | Cryptographic summary of a ledger state |
| Inclusion proof | Proof that a record belongs to a committed Merkle state |
| Runtime observation | Recorded execution or telemetry evidence |
| Hygiene candidate | Advisory static finding requiring human or downstream analysis |
| Checkpoint | Persisted state used to transfer or resume agent work |
| Handoff | Transactional transfer of task ownership |
| MCP | Model Context Protocol interface for AI clients |
| Control Plane | Operator-facing platform surface distinct from the tenant Workspace Console |

---

## Evidence index

- [`README.md`](README.md) — product overview.
- [`PLAYBOOK.md`](PLAYBOOK.md) — installation and operational workflows.
- [`EVIDENCE.md`](EVIDENCE.md) — validation corpus and methodology.
- [`docs/CAPABILITIES.md`](docs/CAPABILITIES.md) — capability matrix.
- [`docs/hygiene.md`](docs/hygiene.md) — Hygiene command and report contract.
- [`docs/DOCUMENT_FORMATS.md`](docs/DOCUMENT_FORMATS.md) — documentation ingestion contract.
- [`docs/invariants.md`](docs/invariants.md) — invariant contract.
- [`docs/adr/`](docs/adr/) — architecture decision records.
- [`scripts/mcp_verify.py`](scripts/mcp_verify.py) — MCP reference verifier.
- [`openapi.yaml`](openapi.yaml) — HTTP contract.
- [`vscode-extension/`](vscode-extension/) — VS Code integration.

---

<p align="center">
  <strong>Garuda</strong><br>
  Analyze · Verify · Govern
</p>