<p align="center">
  <img src="assets/garuda-logo.png" width="112" alt="Garuda">
</p>

<h1 align="center">Garuda</h1>

<p align="center">
  <strong>The software intelligence and verification layer for AI-native engineering.</strong>
</p>

<p align="center">
  Give AI agents and engineering teams a shared, evidence-backed understanding of what a software system is, what it is intended to do, what changed, and what the available evidence supports.
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/releases/tag/v0.2.0"><img src="https://img.shields.io/badge/release-v0.2.0-blue.svg?style=flat-square" alt="Release v0.2.0"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache_2.0-green.svg?style=flat-square" alt="Apache 2.0 License"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.26-00ADD8.svg?style=flat-square" alt="Go 1.26+"></a>
  <a href="scripts/mcp_verify.py"><img src="https://img.shields.io/badge/MCP-21%20tools%2C%2038%2F38%20reference%20checks-success.svg?style=flat-square" alt="MCP 21 tools, 38 of 38 checks passed"></a>
  <a href="SECURITY.md"><img src="https://img.shields.io/badge/security-documented-informational.svg?style=flat-square" alt="Security documentation"></a>
</p>

<p align="center">
  <a href="#getting-started">Get started</a> ·
  <a href="#why-now">Why now</a> ·
  <a href="#what-garuda-does">What Garuda does</a> ·
  <a href="docs/SPECS.md">Specifications</a> ·
  <a href="EVIDENCE.md">Evidence</a> ·
  <a href="PLAYBOOK.md">Playbook</a>
</p>

---

## The thesis

AI is making software execution abundant.

The scarce resource is becoming **trustworthy system context**: knowing what a system is, what it is supposed to do, what changed, what depends on it, and whether available evidence supports the resulting state.

Garuda is the software intelligence and verification layer between AI agents and the software systems they change.

```text
AI agents
    │
    ▼
GARUDA — software intelligence, evidence, and governance
    │
    ▼
Software systems
```

Garuda connects organizational intent, source code, runtime observations, policies, decisions, and verifiable history into one persistent workspace model.

> **Understand the system. Verify the change. Keep AI aligned.**

---

## Why now

Software development is moving from AI-assisted coding toward increasingly agentic execution. Agents can now plan multi-step work, modify repositories, run tools, review changes, coordinate with other agents, and operate for longer periods with less direct supervision.

The important consequence is not simply that more code gets written. It is that **software can change faster than teams can reconstruct and verify system context**.

This creates a growing gap between:

```text
What was intended  →  What was implemented  →  What is observed
```

Garuda calls this **AI Development Drift**.

The problem becomes more important as organizations adopt:

- Parallel coding agents.
- Background and long-running engineering tasks.
- Multi-repository architectures.
- Policy-driven development workflows.
- Runtime-aware governance.
- AI-generated changes that require human review at system scale.

DORA's AI-capabilities research similarly frames AI adoption as a systems and organizational capability problem, not merely a model or editor problem. Garuda is designed for the missing layer between agent execution and system understanding.

---

## The business problem

| Business cost | What creates it | What Garuda provides |
|---|---|---|
| **Rework** | A locally reasonable change violates a dependency, architecture rule, or documented requirement | Semantic relationships, policies, and impact analysis |
| **Review bottlenecks** | Reviewers reconstruct architecture and intent before trusting agent-generated changes | Persistent system state with source-backed evidence |
| **Context reconstruction** | Agents and engineers rediscover the same facts across sessions and tools | Shared workspace state through MCP, CLI, HTTP, dashboard, CI, and IDE interfaces |
| **Operational risk** | Documentation, code, and runtime behavior diverge | Explicit `SUPPORTED`, `UNVERIFIED`, and `CONTRADICTED` states |
| **Governance overhead** | Teams cannot easily prove which rule was applied to which system state | Deterministic policy evaluation and verifiable decision records |
| **Multi-agent coordination cost** | Agents duplicate work or transfer incomplete context | Transactional handoff, checkpoint, and resume primitives |

These are mechanisms of value, not guaranteed dollar savings. Actual outcomes depend on codebase complexity, workflow design, agent autonomy, and adoption depth.

---

## What Garuda is

Garuda is a **persistent software state, evidence, and governance layer for AI-native engineering teams**.

It maintains a workspace model that connects four artifact categories:

| Artifact | What it represents | Examples |
|---|---|---|
| **Intent** | What an organization says should be true | Requirements, ADRs, specifications, policies |
| **Code** | What has been implemented | Repositories, packages, types, functions, methods, relationships |
| **Runtime** | What the system has been observed doing | Traces, spans, operations, runtime observations |
| **Evidence** | What supports a claim or decision | Source locations, verification records, contradictions, Merkle proofs |

Garuda does not attempt to become the model, the coding agent, the observability platform, or the entire agent runtime. It provides the shared state and verification layer those systems can use.

### The questions Garuda answers

- What exists in this workspace?
- What is this symbol and what does it depend on?
- Who calls or implements it?
- What documentation claims apply here?
- What policies govern this change?
- What evidence supports the claim?
- What is still unverified?
- What contradicts the expected state?
- What could be affected if this symbol changes?
- What decision introduced this constraint?
- Can the recorded decision be independently verified?

---

## What Garuda does

### 1. Build a semantic model of software

Garuda analyzes repositories and represents software as typed entities and relationships rather than only text matches.

The model can include:

- Repositories and packages.
- Structs, interfaces, functions, methods, and fields.
- Imports, calls, references, implementations, embedding, and inheritance where supported.
- Cross-repository bridges.
- Source locations and evidence metadata.
- Stable semantic identity.
- Analyzer resolution tier and confidence boundaries.

Go analysis is compiler-backed within the supported scope. Python and TypeScript analysis are structural and carry lower authority for dynamic behavior.

### 2. Connect documentation to implementation

Garuda ingests supported engineering documents and extracts normative claims. Claims are correlated with semantic and runtime evidence rather than accepted as unquestioned truth.

| State | Meaning |
|---|---|
| `SUPPORTED` | Available evidence supports the claim within the active verification scope |
| `UNVERIFIED` | There is not enough evidence to support or contradict the claim |
| `CONTRADICTED` | Available evidence conflicts with the claim |

An unverified claim is not automatically false. It means the evidence boundary remains open.

### 3. Enforce engineering policies

Policies define scope, authority, predicates, outcomes, and reasons. The current decision outcomes are:

- `ALLOW`
- `WARN`
- `REVIEW`
- `BLOCK`

Policies can run as non-persisting previews or as committed evaluations anchored to the Merkle ledger.

Example policy intent:

```text
Payment services must not import database/sql directly.
```

Garuda converts the rule into an executable governance check against the current workspace state.

### 4. Show impact before changes become surprises

Garuda provides two distinct impact paths:

```text
garuda.blast_radius  →  symbol and graph impact
garuda.get_impact     →  decision and lineage impact
```

The result is a structured view of graph-visible consequences, not a guarantee that every dynamic runtime effect has been discovered.

### 5. Correlate static expectations with runtime evidence

Garuda can associate runtime observations with semantic entities and identify whether observed behavior supports or contradicts the expected system state.

```text
Intent / static model
          ↓
Expected behavior
          ↓
Runtime observation
          ↓
SUPPORTED / UNVERIFIED / CONTRADICTED
          ↓
Policy or human review
```

The runtime path is implemented and validated with controlled and synthetic evidence. Production-scale continuous verification remains a validation area rather than a claim that every production execution is continuously observed.

### 6. Give agents persistent system context

Garuda exposes workspace state through MCP, CLI, HTTP, dashboard, CI, and VS Code interfaces.

An AI client can ask:

```text
What implements this interface?
What calls this function?
What policies apply to this change?
Which claims are contradicted?
What could be affected if I change this type?
What decision introduced this constraint?
```

The answer comes from the shared workspace model rather than only the file currently open in the agent context.

### 7. Coordinate work between agents

`garuda.handoff` transfers task ownership while creating a checkpoint. `garuda.resume` restores and consumes that checkpoint transactionally.

This is coordination state, not a replacement for a model harness. Garuda does not itself run a general planner/executor/observer loop, launch arbitrary model fleets, or provide a general-purpose code-execution sandbox.

### 8. Meter and classify agent work

The repository also contains supporting controls for agent infrastructure:

- Tenant-level token and execution metering.
- Budget consumption checks.
- Pre-flight task classification.
- Payload redaction for common secret-like patterns.
- Model-selection recommendations based on task domain and budget state.

This layer is a pre-flight classifier and routing policy, not a complete model-serving fabric or provider gateway.

---

## Where Garuda fits

```mermaid
flowchart LR
    AGENTS["AI coding agents\nCursor · Claude · Codex · MCP clients"] --> GARUDA["GARUDA\nSoftware intelligence + verification"]
    GARUDA --> CODE["Code\nrepositories · packages · symbols"]
    GARUDA --> INTENT["Intent\ndocs · ADRs · policies"]
    GARUDA --> RUNTIME["Runtime\nobservations · traces"]
    GARUDA --> GOVERNANCE["Governance\npolicies · decisions"]
    GARUDA --> EVIDENCE["Evidence\nprovenance · Merkle proofs"]
    GARUDA --> INTERFACES["Interfaces\nCLI · API · CI · IDE · Console"]
```

The agent remains responsible for acting. Garuda helps the agent understand the system it is acting on and preserves evidence about important governance state.

---

## What Garuda is not competing with

Garuda is complementary to most of the tools teams already use.

| Category | What it does well | What Garuda adds |
|---|---|---|
| Agent frameworks | Build and orchestrate agents | A verified software state model for the agents to query |
| Coding agents | Make software changes | Evidence-backed context about the system being changed |
| Code search | Find text quickly | Typed entities, relationships, impact, claims, and verification states |
| Knowledge graphs | Show objects and connections | Evidence tiers, runtime correlation, policy, lineage, and decision integrity |
| Observability platforms | Show production behavior | Correlation between runtime observations and software structure |
| Documentation drift tools | Find disagreement between docs and code | Bidirectional claim verification with explicit uncertainty |
| Security and policy tools | Detect specific rule violations | General governance with authority, evidence, and Merkle-backed decisions |

Garuda's category is the layer between organizational intent and the agents changing software.

---

## User workflows

### Give an agent system context

An agent queries the workspace rather than reconstructing architecture from one local file or one short-lived context window.

### Review a critical change

Inspect callers, implementers, dependencies, claims, policies, and graph-visible impact before accepting a change.

### Connect requirements to implementation

Ingest normative documentation and determine whether implementation or runtime evidence supports, contradicts, or fails to verify the requirement.

### Turn architecture rules into governance

Express important engineering boundaries as policies and evaluate them in preview or committed mode.

### Preserve important decisions

Anchor committed evaluations and decision records so their recorded state can be independently checked later.

### Transfer work between agents

Use durable checkpoints and transactional handoffs instead of depending on one session retaining all context.

### Investigate runtime drift

Correlate observed behavior with static expectations and route contradictions or missing evidence to policy review.

---

## Workspace Console and VS Code

### Workspace Console

The tenant-facing Workspace Console provides:

| Surface | Capability |
|---|---|
| **Workspace** | Repository, package, entity, relationship, language, policy, evidence, and health summaries |
| **Agents** | MCP sessions, tool-call activity, active and recent sessions, and coordination state |
| **Governance** | Policy outcomes, documentation claims, drift, contradictions, verification states, and cryptographic trust |
| **Decisions** | Decision history, evidence, ancestors, descendants, revisions, and anchor status |
| **Graph** | Interactive topology, communities, repositories, packages, relationships, and exploration |

### VS Code extension

Garuda includes a VS Code extension under [`vscode-extension/`](vscode-extension/).

Current extension workflows include:

- Cryptographic Ledger view.
- Quarantined Contradictions view.
- Problems-panel diagnostics.
- Policy and runtime contradiction markers.
- Architecture graph visualizer.
- Workspace AST reanalysis.
- Ledger and verification-state refresh.
- Go symbol hover with blast-radius and dependency information.

The extension is tested for editor-integrated diagnostics and exploration. Its update behavior depends on configured state, daemon/database connectivity, and refresh or reanalysis events. “Real-time” should not be read as a guarantee that every file mutation is continuously analyzed without an analysis event.

---

## Evidence and trust

Garuda follows a simple rule:

> **Evidence before confidence.**

The trust layer uses a Merkle structure for committed governance records.

A verifier can check whether a specific record is included under a known root and whether its recorded content has changed since inclusion.

That does not prove:

- The semantic model is complete.
- A policy is substantively correct.
- Runtime telemetry covers every production path.
- The underlying software is safe.
- Every language has compiler-grade resolution.

It proves the narrower property that matters for auditability:

> **What Garuda recorded as a governance decision can be independently verified as a recorded fact.**

Verify a committed evaluation with:

```bash
garuda policy verify <evaluation-id>
```

---

## Current proof points

| Area | Current evidence |
|---|---|
| **MCP surface** | 21 tools in the MCP server registry |
| **MCP verification** | `scripts/mcp_verify.py` reports 38/38 reference checks |
| **Client validation** | Cursor, Claude Desktop, Codex CLI, and dependency-free reference verifier |
| **Semantic analysis** | Go compiler-backed within supported scope; Python and TypeScript structural analyzers |
| **Documentation ingestion** | Supported Markdown, PDF, DOCX, TXT, RST, and normalized ADR-style content |
| **Governance** | Declarative policy evaluation with `ALLOW`, `WARN`, `REVIEW`, and `BLOCK` outcomes |
| **Trust layer** | Merkle commitments, inclusion proofs, and decision-chain verification |
| **Multi-agent state** | Transactional handoff, checkpoint, resume, and duplicate-consumption protection |
| **Multi-tenancy** | Users, tenants, memberships, workspace resolution, and scoped state |
| **Agent metering** | Token and execution budgets with consumption state |
| **Pre-flight controls** | Task classification and secret-like payload redaction |
| **User surfaces** | Workspace Console, Agents, Governance, Decisions, Graph, and VS Code integration |

These are implementation and validation facts within the documented scope, not claims that every capability has been tested at every production scale.

See [`EVIDENCE.md`](EVIDENCE.md), [`BENCHMARKS_AND_VERIFICATION_v1.0.0.md`](BENCHMARKS_AND_VERIFICATION_v1.0.0.md), and [`docs/CAPABILITIES.md`](docs/CAPABILITIES.md) for methodology and boundaries.

---

## Current maturity

**Current status: Early / testing.**

The latest tagged public release is `v0.2.0`. The `main` branch may contain additional unreleased work beyond that tag, including current MCP, workspace, coordination, dashboard, and operational capabilities.

### Stable within the current development line

- Semantic software modeling within documented analyzer scopes.
- Workspace-scoped state and graph persistence.
- Policy evaluation and governance outcomes.
- Merkle-backed governance records.
- MCP integration and reference verification.
- Transactional handoff and resume.
- Workspace Console and IDE integrations.

### Beta or validation-sensitive

- Python and TypeScript semantic depth.
- Runtime verification against production-scale telemetry.
- Large multi-repository validation beyond named corpora.
- Deployment hardening across every operational access path.

### Explicit non-claims for the current release

- General-purpose agent execution sandboxing.
- Enterprise secrets or credential brokering.
- Durable autonomous scheduling.
- Complete model-provider orchestration.
- Cross-agent shared memory as a finished capability.

Capability maturity is tracked in [`docs/CAPABILITIES.md`](docs/CAPABILITIES.md).

---

## Current limitations and non-goals

Garuda will not claim:

- Zero hallucinations.
- 100% semantic accuracy across every language.
- Complete knowledge of every code path.
- Guaranteed prevention of every unsafe change.
- Universal runtime coverage.
- A cryptographic ledger as proof that the underlying software is correct.
- A model router as a replacement for a complete inference gateway.
- A checkpoint system as a complete agent runtime.

The product is designed to become more trustworthy by making its limits visible.

---

## Getting started

### Build from source

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda

go build -o bin/garuda ./cmd/garuda
go build -o bin/garuda-mcp ./cmd/garuda-mcp
```

Requires Go 1.26 or later. TypeScript support requires `CGO_ENABLED=1` and a C compiler.

### Start PostgreSQL

```bash
docker run -d \
  --name garuda-pg \
  -e POSTGRES_USER=garuda \
  -e POSTGRES_PASSWORD=garuda \
  -e POSTGRES_DB=garuda \
  -p 5432:5432 \
  postgres:16
```

```bash
export DATABASE_URL="postgres://garuda:garuda@localhost:5432/garuda?sslmode=disable"
```

Apply migrations:

```bash
for f in migrations/[0-9]*.sql; do
  psql "$DATABASE_URL" -f "$f"
done
```

If duplicate migration prefixes exist in a working tree, verify ordering before applying them.

### Create and analyze a workspace

```bash
export GARUDA_WORKSPACE=my-workspace
./bin/garuda workspace create my-workspace
./bin/garuda analyze /path/to/repo --save
```

Analyze each repository against the same workspace to build a multi-repository model.

### Ingest and verify documentation

```bash
./bin/garuda docs ingest ./docs
./bin/garuda docs verify
```

### Validate and evaluate policies

```bash
./bin/garuda policy validate ./policies
./bin/garuda policy evaluate ./policies --workspace my-workspace
```

Preview mode does not persist or anchor the result. Add `--save` for a recorded governance decision:

```bash
./bin/garuda policy evaluate ./policies --workspace my-workspace --save
```

### Start the workspace service

```bash
./bin/garuda dev
```

Open `http://localhost:8080`.

### Verify MCP

```bash
WORKSPACE=my-workspace python3 scripts/mcp_verify.py
```

Expected:

```text
═══ 38/38 checks passed ═══
```

---

## Interfaces

| Interface | Purpose |
|---|---|
| **MCP** | Agent-native access to workspace state, governance, evidence, impact, and coordination |
| **CLI** | Analysis, policies, documentation, verification, impact, CI, and operations |
| **HTTP / OpenAPI** | Machine-readable programmatic integrations |
| **Workspace Console** | Human-facing workspace, agents, governance, decisions, and graph views |
| **VS Code extension** | In-editor architecture, ledger, contradiction, diagnostics, and exploration workflows |
| **CI integration** | Policy gates, semantic analysis, and automated engineering checks |

---

## Security and isolation

Garuda includes application-level controls for multi-tenant and workspace-scoped access, including:

- Authenticated sessions and protected API paths.
- Tenant and workspace membership.
- Workspace-scoped request resolution.
- SQL scoping and isolation checks.
- Idempotent governance writes.
- Request identifiers.
- Non-root container support for the API image.
- Protected operator access.
- Token and execution budgeting.
- Audit and Merkle verification paths.

These controls are part of the current product architecture. They should not be interpreted as a claim that Garuda provides the complete OS-level isolation envelope of a hardened multi-user agent execution platform.

See [`SECURITY.md`](SECURITY.md) for the security model and reporting process.

---

## The moat we are building

Garuda's differentiation is the combination of:

### Persistent system state

A durable model of software structure, intent, evidence, runtime observations, and decisions.

### Evidence-aware semantics

Relationships tied to their source and resolution tier, with uncertainty preserved rather than hidden.

### Governance as an executable layer

Policies applied to the same state agents and humans are reading.

### Cryptographically verifiable history

Important decisions committed to records that can be independently checked.

### Agent-native integration

MCP makes the same state directly queryable where AI decisions are made.

The long-term defensibility comes from the state model, evidence model, governance model, and integration surface becoming deeply embedded in engineering workflows.

---

## Commercial direction

Garuda starts with a focused wedge: **AI-native software engineering**.

```text
AI-enabled engineering team
        ↓
Shared software intelligence
        ↓
Engineering governance and verification
        ↓
Organization-wide software state
        ↓
Business-critical system integrity
```

The first product solves a concrete engineering problem: agents and people need reliable system context and a way to verify important changes.

Longer-term expansion may include business-critical domains such as payments, billing, bookings, refunds, and other workflows where correctness depends on explicit invariants and state transitions.

These are future expansion areas, not claims about the current release.

---

## What we measure

Garuda deliberately avoids presenting proxy metrics as business outcomes.

The product metrics that matter include:

- Time to reach a source-backed answer.
- Time spent reconstructing unfamiliar system context.
- Review time for agent-generated changes.
- Policy and architecture violations detected before merge.
- Rework caused by system drift.
- Time to recover context when work moves between agents.
- Percentage of important claims with explicit verification state.
- Governance decisions that can be independently reproduced.
- Controlled agent-task success with and without Garuda.

Token usage, graph size, tool count, and latency remain useful engineering measurements, but they are secondary to whether Garuda reduces real engineering friction.

---

## Roadmap direction

The roadmap extends the same system rather than creating unrelated products.

### Near term

- Deeper runtime verification.
- Stronger cross-repository and cross-language analysis.
- Autonomous verification workflows.
- Broader dashboard and IDE integration.
- Stale-session lifecycle handling.
- Deeper policy and contract analysis.

### Agent intelligence layer

- Durable shared memory across agent contexts.
- Context compression for long-running work.
- Richer handoff semantics.
- More complete agent attribution and cost accounting.

### System-integrity layer

- Stronger policy coverage for business-critical invariants.
- Richer runtime-to-intent verification.
- Expanded evidence exchange between engineering systems.
- Integrations for larger organizational software estates.

Planned capabilities are not current shipped capabilities.

---

## Documentation and verification

| Document | Purpose |
|---|---|
| [`docs/SPECS.md`](docs/SPECS.md) | Detailed public system and product specification |
| [`docs/CAPABILITIES.md`](docs/CAPABILITIES.md) | Capability maturity and analyzer verification status |
| [`EVIDENCE.md`](EVIDENCE.md) | Evidence methodology and current validation record |
| [`BENCHMARKS_AND_VERIFICATION_v1.0.0.md`](BENCHMARKS_AND_VERIFICATION_v1.0.0.md) | Detailed benchmark and verification report |
| [`PLAYBOOK.md`](PLAYBOOK.md) | Installation, workflows, and operational reference |
| [`docs/DOCUMENT_FORMATS.md`](docs/DOCUMENT_FORMATS.md) | Document normalization and claim-extraction contract |
| [`docs/invariants.md`](docs/invariants.md) | Trust-substrate invariant contract |
| [`docs/adr/`](docs/adr/) | Architecture decisions |
| [`openapi.yaml`](openapi.yaml) | Machine-readable HTTP API contract |
| [`scripts/mcp_verify.py`](scripts/mcp_verify.py) | Dependency-free MCP reference verifier |
| [`SECURITY.md`](SECURITY.md) | Security model and vulnerability reporting |

---

## Contributing

Garuda is built in the open. Contributions are welcome across analyzers, semantic resolution, policy evaluation, runtime evidence, MCP integrations, dashboards, IDE tooling, verification, security, documentation, and testing.

See [`CONTRIBUTING.md`](CONTRIBUTING.md) to get started.

---

## License

Apache License 2.0. See [`LICENSE`](LICENSE).

---

<p align="center">
  <strong>Garuda</strong><br>
  Understand the system. Verify the change. Keep AI aligned.
</p>