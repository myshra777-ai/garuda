<p align="center">
  <img src="assets/garuda-logo.png" width="128" alt="Garuda">
</p>

<h1 align="center">Garuda</h1>

<p align="center">
  <strong>Keep AI agents aligned with the software they change.</strong>
</p>

<p align="center">
  <em>A verified state plane for AI-assisted software development.</em>
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/releases"><img src="https://img.shields.io/badge/version-v0.1.x-blue.svg?style=flat-square" alt="Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache_2.0-green.svg?style=flat-square" alt="License"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.26-00ADD8.svg?style=flat-square" alt="Go Version"></a>
  <a href="scripts/mcp_verify.py"><img src="https://img.shields.io/badge/MCP-21%20tools%2C%2038%2F38%20verified-success.svg?style=flat-square" alt="MCP Verified"></a>
</p>

<p align="center">
  <a href="#getting-started">Get started</a> ·
  <a href="docs/SPECS.md">Read the specifications</a> ·
  <a href="PLAYBOOK.md">Read the playbook</a> ·
  <a href="EVIDENCE.md">Review the evidence</a> ·
  <a href="https://github.com/myshra777-ai/garuda/issues">Report an issue</a>
</p>

---

## The problem: AI development drift

AI can now write, change, test, and repair software faster than teams can review it.

That speed creates a new operational problem. An organization records an intention, an AI agent changes the code, and the running system evolves again. Over time, documentation, implementation, and runtime behavior can diverge without a single dramatic failure announcing the change.

> **AI Development Drift is the growing gap between what an organization intended, what was built, and what the system actually does.**

Garuda makes that gap visible, queryable, and governable.

---

## What Garuda does

Garuda is the layer AI agents call before and after they change software. It provides a shared, evidence-aware model of the system and answers three practical questions:

- **What exists?** Entities, relationships, dependencies, and evidence across the workspace.
- **What rules apply?** Policies and architectural constraints evaluated against the current system state.
- **What could change?** Callers, implementations, dependencies, claims, and graph-visible blast radius.

Go analysis is compiler-backed within the supported scope. Python and TypeScript analysis is structural. Runtime observations and documentation claims are represented with explicit verification states rather than being silently converted into certainty.

The same persisted model can be queried by developers, AI agents, CLI workflows, dashboards, CI pipelines, and the VS Code extension. The goal is not to replace those clients; it is to give them a consistent state and evidence layer.

```mermaid
flowchart TD
    INTENT["Intent: docs, policies, decisions"] --> GARUDA["Garuda: verified state plane"]
    CODE["Code: what was built"] --> GARUDA
    RUNTIME["Runtime: observed behavior"] --> GARUDA
    GARUDA --> AGENTS["AI agents via MCP"]
    GARUDA --> HUMANS["Developers via Console, CLI, and VS Code"]
    GARUDA --> CI["CI and automation"]
    AGENTS --> GOVERN["Query, evaluate, verify"]
    HUMANS --> GOVERN
    CI --> GOVERN
    GOVERN --> RECORD["Evidence and decisions anchored to a Merkle ledger"]
```

### What Garuda is not competing with

Garuda is not an agent framework. Microsoft Agent Framework, OpenAI's Agents SDK, and similar tools are excellent at building and orchestrating agents. Garuda sits underneath them.

Garuda is not a coding agent. Cursor, Claude Code, and Copilot are excellent at performing software changes. Garuda gives them a verified map of the system they are changing.

Garuda is not a code search tool. Sourcegraph is excellent at finding text and code across many repositories. Garuda answers a different question: not "where does this text appear" but "what is this, what does it depend on, and what evidence supports it?"

The category Garuda occupies is narrower and more specific: the verification and state layer that agent frameworks and coding agents can query before and after they change software.

---

## A representative workflow

**Define intent.** A team records a policy such as:

> Payment services must not import `database/sql` directly.

**Analyze implementation.** Garuda extracts entities and relationships from the workspace.

**Evaluate the policy.** The policy engine returns `ALLOW`, `WARN`, `REVIEW`, or `BLOCK` based on the current graph and claims.

**Check runtime evidence.** Exported handlers without runtime evidence can be routed to review through a `verification_missing` predicate.

**Preserve the decision.** A committed evaluation can be anchored to the tenant's Merkle ledger and independently verified later.

**Continue the work.** An AI agent can query the same workspace through MCP, inspect affected symbols, understand the blast radius, and hand work to another agent through a transactional checkpoint.

---

## What teams use Garuda for

### Ground AI agents in the real system

An AI coding agent can ask what calls a function, who implements an interface, what claims apply, or what could be affected by a change. Garuda returns structured answers from the shared semantic model instead of relying only on the file currently open in the agent context.

### Explore large workspaces

Navigate from repositories to packages, entities, relationships, architectural hubs, evidence, policies, and decisions. Interactive graph exploration provides a system-level view across repositories rather than treating each codebase as an isolated island.

### Keep documentation aligned with implementation

Garuda ingests supported normative claims from Markdown, ADR, and plain-text documents and correlates them with analyzed code and available runtime evidence. Claims are tracked as `SUPPORTED`, `UNVERIFIED`, or `CONTRADICTED`.

An unverified claim is not automatically false. It means the current evidence is insufficient.

### Evaluate engineering and governance policies

Policies are declarative YAML rules with authority, scope, predicates, and outcomes. Evaluations can run as previews or be committed and anchored when an auditable decision is required.

### Understand change impact

Use graph relationships to inspect callers, implementers, subclasses, neighbors, and transitive impact. `garuda.blast_radius` answers what may be affected by changing a symbol; `garuda.get_impact` answers what may be affected by changing a decision in the lineage graph.

### Correlate static structure with runtime behavior

Runtime observations can be associated with semantic entities and used to surface unsupported or contradictory behavior. Current development extends this toward more autonomous verification, dead-code detection, and unused-import analysis exposed through MCP.

### Coordinate agent work

MCP exposes transactional handoff and resume operations. Agents can transfer task ownership, create checkpoints, and reject duplicate checkpoint consumption through Serializable transaction paths.

### Preserve governance evidence

Persisted policy evaluations and decision records can be anchored to Merkle structures. Independent verification confirms what was recorded, at which block, and against which policy version. The ledger does not claim that every semantic model is complete; it makes the recorded evidence and decision history independently verifiable.

---

## Two dashboards and an IDE

### Workspace Console

The tenant-facing Workspace Console provides:

| Surface | Capability |
|---|---|
| **Workspace** | Repository, package, entity, and relationship counts; language composition; architectural hubs; recent evidence; workspace health |
| **Agents** | MCP surface, active and recent sessions, tool-call activity, and agent coordination state |
| **Governance** | Policy enforcement, documentation claims, drift, contradictions, verification states, and Merkle trust |
| **Decisions** | Decision history, policy metadata, evidence, ancestry, descendants, revision chains, and anchor status |
| **Graph** | Interactive topology, communities, repositories, packages, relationships, and impact-oriented exploration |

The console displays policy violations, contradictions, unverified claims, graph relationships, runtime evidence, and impact-related governance information within the selected workspace.

### VS Code extension

Garuda includes a VS Code extension under [`vscode-extension/`](vscode-extension/).

The extension provides:

- Cryptographic Ledger view.
- Quarantined Contradictions view.
- Garuda Problems-panel diagnostics.
- Architecture graph visualizer command.
- Workspace AST reanalysis command.
- Ledger and verification-state refresh command.
- Go symbol hover information with blast-radius and dependency counts.

The extension is tested for editor-integrated diagnostics and exploration. Its diagnostic behavior depends on configured Garuda state, daemon/database connectivity, and the configured refresh or reanalysis path. “Real-time” should therefore be understood as editor-integrated and refresh-aware, not as a guarantee that every file mutation is continuously analyzed without an analysis event.

### Operator surface

Garuda may also be deployed with an owner-facing operational surface distinct from the tenant-facing Workspace Console. This internal surface is not required for ordinary workspace use and is not the primary public product boundary.

---

## Verification states

| State | Meaning |
|---|---|
| **SUPPORTED** | Available evidence supports the claim within the verified scope |
| **UNVERIFIED** | The claim has not received sufficient supporting evidence; this does not mean it is false or broken |
| **CONTRADICTED** | Available evidence conflicts with the claim |

> **Absence of evidence is not evidence of absence.**

Evidence authority depends on source and maturity tier. Compiler-backed AST/type evidence and persisted source-attributed relationships have a different authority level from heuristic inference. Heuristic relationships remain labeled and cannot silently override stronger evidence.

---

## Why the record is trustworthy

Persisted governance decisions are recorded in a tamper-evident Merkle ledger. Independent verification can confirm that a decision was included at a specific block, that its recorded content has not been altered since inclusion, and that its proof can be re-derived.

```bash
garuda policy verify <decision-id>
```

### What the ledger actually proves

The Merkle ledger proves that a decision was recorded at a specific block, that its content has not been altered since, and that the evaluation that produced it can be independently re-verified.

It does not prove that the semantic model the decision was made against is complete or correct. That is a different claim, and Garuda does not make it.

What Garuda does claim, precisely: **Garuda makes the evidence and decisions behind system governance independently verifiable.** A reader with the evaluation ID and verification logic can confirm what was decided, when, and against which policy version without trusting Garuda's servers, database, or operators.

The canonical invariant contract is documented in [`docs/invariants.md`](docs/invariants.md). The complete system specification is available in [`docs/SPECS.md`](docs/SPECS.md).

---

## How Garuda differs from adjacent tools

| You might already use | What it does well | What Garuda adds |
|---|---|---|
| Agent frameworks | Building and orchestrating agents | A verified model of the software those agents operate on |
| Coding agents | Performing software changes with useful context | A structured, evidence-backed map of the system being changed |
| Code search | Finding text quickly across repositories | Typed semantic identity, relationships, evidence, and verification state |
| Code knowledge graphs | Showing entities and relationships | Policy, evidence status, runtime correlation, lineage, and decision integrity |
| Runtime observability | Showing production behavior | Correlation between runtime observations and static semantic entities |
| Doc-code drift tools | Reporting disagreement between docs and code | Bidirectional claim verification with explicit uncertainty and evidence |
| Security and policy tools | Detecting specific violation classes | A general governance engine with authority, evidence, and Merkle-anchored decisions |

Garuda is complementary to these systems. Its integration path is MCP, alongside CLI, HTTP, dashboard, CI, and IDE interfaces.

---

## Built for AI agents, verified across four clients

Garuda speaks the Model Context Protocol over line-delimited JSON-RPC 2.0 over stdio. The current MCP surface exposes 21 tools for workspace briefing, graph exploration, policy and governance queries, decision lineage, impact analysis, contradictions, proposals, and multi-agent coordination.

The MCP surface has been tested against four clients:

| Client | Verification scope |
|---|---|
| **Cursor** | MCP integration in the developer IDE |
| **Claude Desktop** | MCP integration in a desktop client with strict response handling |
| **Codex CLI** | MCP integration in an independent CLI client |
| **Reference verifier** | Dependency-free protocol, tool, and negative-path checks through [`scripts/mcp_verify.py`](scripts/mcp_verify.py) |

Current reference-verifier result:

```text
═══ 38/38 checks passed ═══
```

The verifier checks initialization, protocol version, tool discovery, entity and relationship queries, policy-proof failure handling, blast-radius output, handoff and resume negative paths, unknown-tool errors, and clean server shutdown.

### MCP tool groups

| Group | Tools | Purpose |
|---|---|---|
| **Session context** | `garuda.briefing` | Workspace state, trust anchor, scale, hubs, policies, contradictions, and recent changes |
| **Code graph exploration** | `garuda.entities`, `garuda.find_entity`, `garuda.inspect`, `garuda.neighbors`, `garuda.subclasses`, `garuda.implementers`, `garuda.blast_radius`, `garuda.query_claims`, `garuda.query` | Semantic discovery, relationships, implementations, claims, and graph-visible impact |
| **Governance and decisions** | `garuda.policy.list`, `garuda.policy.evaluate`, `garuda.verify_policy_evaluation`, `garuda.governance.status`, `garuda.check_drift`, `garuda.get_lineage`, `garuda.get_impact`, `garuda.detect_contradictions`, `garuda.propose_decision` | Policy state, drift, contradictions, lineage, impact, proposals, and proof verification |
| **Multi-agent coordination** | `garuda.handoff`, `garuda.resume` | Transactional task transfer and checkpoint restoration |

`garuda.policy.evaluate` runs a preview by default; add `--save` to persist and anchor the result. The `--reconcile` flag retires active policy rows whose YAML file is no longer present and requires `--save`.

`garuda.blast_radius` answers “what breaks if I change this symbol” from the graph. `garuda.get_impact` answers “what breaks if I change this decision” from the lineage DAG.

Read-only tools are everything except `garuda.propose_decision`, `garuda.handoff`, and `garuda.resume`. `garuda.handoff` and `garuda.resume` use Serializable transaction paths; a second resume of the same checkpoint returns `{ "status": "not_found" }`.

---

## Current capabilities

### Available today

- **Go, Python, and TypeScript analysis.** Go is compiler-backed within the supported scope; Python and TypeScript are structural analyzers.
- **Semantic graph.** Entities, packages, functions, methods, interfaces, fields, imports, calls, embeddings, implementations, claims, and source evidence.
- **Cross-repository intelligence.** Workspace-level analysis of supported multi-module dependencies, including validated cross-repository bridges.
- **Documentation ingestion.** Normative claim extraction from supported Markdown, ADR, and plain-text formats.
- **Verification states.** `SUPPORTED`, `UNVERIFIED`, and `CONTRADICTED` claim and evidence states.
- **Policy engine.** Declarative predicates and `ALLOW`, `WARN`, `REVIEW`, and `BLOCK` outcomes.
- **Merkle trust layer.** Decision anchoring, inclusion proofs, revision-chain verification, and independent proof validation.
- **MCP integration.** 21 tools tested against Cursor, Claude Desktop, Codex CLI, and the dependency-free reference verifier; current reference result: 38/38 checks passed.
- **MCP session tracking.** MCP process sessions and tool calls are recorded with status, duration, and whitelisted argument summaries; free-form text is not recorded.
- **Agent activity dashboard.** Active and recent MCP sessions, tool-call counts, and per-session activity.
- **Multi-agent coordination.** Serializable task handoffs, checkpoints, resume operations, and duplicate-resume rejection.
- **Workspace Console.** Workspace, Agents, Governance, Decisions, and Graph surfaces.
- **VS Code integration.** Ledger, contradiction, diagnostic, graph, reanalysis, refresh, and Go hover workflows.
- **Decision history.** Decision records, ancestry, descendants, revision chains, and genesis markers.
- **Runtime evidence ingestion.** Runtime observations and contradiction workflows for correlating execution evidence with semantic entities.
- **CLI, HTTP, MCP, dashboard, IDE, and CI interfaces.** Supported workflows are scriptable or queryable through machine-readable interfaces.
- **SQL-scoping lint.** Detection of workspace-scoped queries missing required workspace filters.

### In beta

- **Multi-tenant onboarding and workspace membership.** Team accounts, tenant membership, and shared workspace operation are implemented and under continued release validation.
- **Runtime verification at scale.** Correlation of live production telemetry with the semantic graph; current validation includes synthetic and controlled runtime evidence.
- **Large multi-repository benchmarks.** Formal quality gates for larger repository sets, including 10-, 25-, and 100-plus-repository targets.
- **Cross-language relationship depth.** Python and TypeScript structural analysis is available, but relationship authority differs from compiler-backed Go analysis.

### Upcoming

- **Rust, Java, and Elixir analyzers.** Additional language support on the semantic pipeline.
- **Autonomous verification workflows.** Automatic runtime correlation, dead-code detection, and unused-import detection exposed through MCP.
- **Cross-agent shared memory.** Workspace-keyed memory that allows one agent to consume durable knowledge produced by another.
- **Memory compression.** Summarization of older agent context for long-running workflows.
- **Stale MCP-session cleanup.** Automatic closure of sessions that terminate without a graceful shutdown.
- **Deeper lint coverage.** Store-versus-DTO field drift and additional contract-mismatch checks.

---

## Who Garuda is for

### AI-enabled engineering teams

Teams using Cursor, Claude Desktop, Codex, GitHub Copilot, or other MCP-capable clients can use Garuda as a shared state and governance layer beneath their development workflows.

### Platform and architecture teams

Teams responsible for boundaries, dependency visibility, policy enforcement, runtime evidence, and cross-repository architecture can use Garuda to replace fragmented manual analysis with a shared semantic model.

### Engineering leaders

Leaders who need evidence about drift, policy outcomes, architectural impact, and governance decisions can use the dashboards, CLI, APIs, and decision ledger.

### High-trust engineering environments

Teams that need independently verifiable technical evidence can use Garuda's Merkle-backed decision records. Garuda provides technical evidence; it is not itself a regulatory certification.

---

## What Garuda is not

- **Not an agent framework.** It does not build or orchestrate model agents.
- **Not a coding agent.** It does not autonomously implement software changes.
- **Not a general code-search replacement.** It answers semantic and governance questions rather than only locating text.
- **Not a compliance certification.** It provides technical evidence and verification mechanisms.
- **Not a replacement for human judgment.** Teams decide what actions to take.
- **Not a universal complete call graph.** Coverage depends on language, relationship type, repository scope, and analyzer maturity.
- **Not a guarantee that the semantic model is complete.** The ledger verifies recorded evidence and decisions, not universal correctness.

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

The URL must match the PostgreSQL container you started. If another container is running on a different port, Garuda uses whichever database is specified by the exported `DATABASE_URL`.

Apply migrations:

```bash
for f in migrations/[0-9]*.sql; do
  psql "$DATABASE_URL" -f "$f"
done
```

If the migration directory contains duplicate numeric prefixes, verify the intended ordering before applying them.

### Create a workspace

```bash
export GARUDA_WORKSPACE=my-workspace
./bin/garuda workspace create my-workspace
```

### Analyze a repository

```bash
./bin/garuda analyze /path/to/repo --save
```

Run the command once per repository when building a multi-repository workspace.

### Ingest documentation

```bash
./bin/garuda docs ingest ./docs
./bin/garuda docs verify
```

### Validate and evaluate policies

```bash
./bin/garuda policy validate ./policies
./bin/garuda policy evaluate ./policies --workspace my-workspace
```

Evaluation is a preview by default. Nothing is persisted or anchored. Add `--save` to create a recorded governance decision:

```bash
./bin/garuda policy evaluate ./policies --workspace my-workspace --save
```

### Connect an AI client through MCP

```json
{
  "mcpServers": {
    "garuda": {
      "command": "/absolute/path/to/bin/garuda-mcp",
      "env": {
        "DATABASE_URL": "postgres://garuda:garuda@localhost:5432/garuda?sslmode=disable",
        "GARUDA_WORKSPACE": "my-workspace"
      }
    }
  }
}
```

Use an absolute path to `garuda-mcp`. Restart the MCP client after changing its configuration.

### Verify the MCP server

```bash
WORKSPACE=my-workspace python3 scripts/mcp_verify.py
```

Expected reference-verifier result:

```text
═══ 38/38 checks passed ═══
```

The verifier exercises tool discovery, positive and negative paths, structured errors, and clean shutdown behavior.

### Build the VS Code extension

```bash
cd vscode-extension
npm install
npm run compile
```

Open the repository in VS Code, configure `garuda.executablePath`, `garuda.databaseUrl`, and `garuda.daemonUrl`, then use the Garuda Architecture Shield commands and views.

### Start the workspace service

```bash
./bin/garuda dev
```

Then open `http://localhost:8080` in a browser.

---

## Architecture at a glance

```mermaid
flowchart LR
    DOCS["Documentation and policies"] --> GRAPH["Semantic graph"]
    CODE["Source repositories"] --> GRAPH
    RUNTIME["Runtime observations"] --> GRAPH
    GRAPH --> POLICY["Policy and governance engine"]
    GRAPH --> MCP["MCP server"]
    GRAPH --> CONSOLE["Workspace Console"]
    GRAPH --> IDE["VS Code extension"]
    POLICY --> LEDGER["Merkle trust layer"]
    MCP --> AGENTS["AI clients and agents"]
    CONSOLE --> DEVELOPERS["Developers and operators"]
    IDE --> DEVELOPERS
    LEDGER --> VERIFY["Independent verification"]
```

Garuda's surfaces share one persisted model. The graph, policy engine, dashboards, CLI, MCP tools, runtime evidence, IDE diagnostics, and ledger are different interfaces over connected state—not isolated product features.

---

## Documentation

| Document | Purpose |
|---|---|
| [`docs/SPECS.md`](docs/SPECS.md) | Detailed public product and system specification |
| [`PLAYBOOK.md`](PLAYBOOK.md) | Installation, setup, workflows, and command reference |
| [`EVIDENCE.md`](EVIDENCE.md) | Validation results and evidence methodology |
| [`docs/CAPABILITIES.md`](docs/CAPABILITIES.md) | Analyzer and product capability maturity matrix |
| [`docs/specs/`](docs/specs/) | Subsystem specifications and operational contracts |
| [`docs/adr/`](docs/adr/) | Architecture decision records |
| [`docs/DOCUMENT_FORMATS.md`](docs/DOCUMENT_FORMATS.md) | Supported document formats and claim extraction |
| [`docs/invariants.md`](docs/invariants.md) | Invariant contract for the trust substrate |
| [`vscode-extension/`](vscode-extension/) | VS Code integration |
| [`scripts/mcp_verify.py`](scripts/mcp_verify.py) | Dependency-free MCP reference verifier |
| [`openapi.yaml`](openapi.yaml) | HTTP API contract |
| [`SECURITY.md`](SECURITY.md) | Security model and vulnerability reporting |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | Contribution guidelines |

---

## Roadmap direction

Garuda's direction is to keep organizational intent continuously connected to the software that implements it, the runtime behavior that exercises it, and the AI agents that change it.

The next product arc focuses on autonomous verification, deeper language support, workspace-keyed agent memory, operational observability, dead-code and unused-import detection, stronger IDE automation, and synchronized evidence across every client and documentation surface.

---

## Contributing

Garuda is built in the open. Contributions are welcome across language analyzers, semantic resolution, policy evaluation, runtime evidence, MCP integrations, dashboards, IDE tooling, verification, documentation, and testing.

See [`CONTRIBUTING.md`](CONTRIBUTING.md) to get started.

---

## License

Apache License 2.0. See [`LICENSE`](LICENSE).

---

<p align="center">
  <strong>Garuda</strong><br>
  Analyze · Verify · Govern
</p>

<p align="center">
  <sub>
    Understand what you intended.<br>
    See what was built.<br>
    Verify what is actually happening.<br>
    Keep humans and AI aligned.
  </sub>
</p>