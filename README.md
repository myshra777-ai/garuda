<p align="center">
  <img src="assets/garuda-logo.png" width="128" alt="Garuda">
</p>

<h1 align="center">Garuda</h1>

<p align="center">
  <strong>Keep AI agents aligned with what you're building.</strong>
</p>

<p align="center">
  <em>A verified state plane for the AI agents changing your software.</em>
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/releases"><img src="https://img.shields.io/badge/version-v0.1.x-blue.svg?style=flat-square" alt="Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache_2.0-green.svg?style=flat-square" alt="License"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.26-00ADD8.svg?style=flat-square" alt="Go Version"></a>
  <a href="scripts/mcp_verify.py"><img src="https://img.shields.io/badge/MCP-21%20tools%2C%2036%2F36%20verified-success.svg?style=flat-square" alt="MCP Verified"></a>
</p>

<p align="center">
  <a href="#getting-started">Get started</a> ·
  <a href="PLAYBOOK.md">Read the playbook</a> ·
  <a href="EVIDENCE.md">Review the evidence</a> ·
  <a href="https://github.com/myshra777-ai/garuda/issues">Report an issue</a>
</p>

---

## The problem everyone is quietly having

AI can now write, change, test, and fix software faster than any team can review it.

That speed is real. It's also creating a new kind of problem that nobody has a good name for yet.

A team writes down what it wants to build. A month later, the software has moved somewhere else. Not dramatically — not with a big rewrite — but a little at a time, change by change. Documentation says one thing, the code does another, and the running system does a third. Nobody notices until something breaks in production.

This happens with human developers too, but slowly. With AI agents, it happens fast enough that teams can lose track of their own software within weeks.

> **AI Development Drift is the growing gap between what an organization intended, what got built, and what the system actually does.**

It is not a bug in any single AI model. It is the natural result of asking fast-moving agents to reason about a system they can only partially see, across sessions they cannot remember, against intentions that keep changing.

---

## What Garuda does about it

Garuda is the layer your AI agents call before they change code. It answers three questions with compiler-backed evidence instead of a guess from the file that happens to be open:

- **What exists?** Every entity, relationship, and dependency in the workspace, resolved from compiler type information for Go and structural analysis for Python and TypeScript.
- **What are the rules?** Policies the organization has committed to, evaluated against the current state of the code.
- **What would break?** The transitive blast radius of a change, computed from the call and inheritance graph.

The same model is shared by every developer, every AI agent, and every session. Nothing is cached in one agent's context window. Nothing is re-derived from a fresh reading of the files. When a rule changes on Wednesday, the answer on Thursday reflects it — for every agent querying Garuda, in every client.

The four things Garuda connects:

```mermaid
flowchart TD
    INTENT["Intent: docs, policies, decisions"] --> GARUDA
    CODE["Code: what was built"] --> GARUDA
    RUNTIME["Runtime: what actually happens"] --> GARUDA
    GARUDA["Garuda: verified state plane"]
    GARUDA --> AGENTS["AI agents query via MCP"]
    GARUDA --> HUMANS["Developers query via CLI and dashboard"]
    GARUDA --> CI["CI queries via CLI"]
    AGENTS --> CHECK["Every change checked against the model"]
    HUMANS --> CHECK
    CI --> CHECK
    CHECK --> RECORD["Decisions anchored to a verifiable ledger"]
    RECORD --> GARUDA
```

When a developer or an AI agent proposes a change, Garuda checks it against the intent the organization has already committed to. When something drifts, you hear about it while it is still a small problem — not three months later during an incident review.

### What Garuda is not competing with

Garuda is not an agent framework. Microsoft Agent Framework, OpenAI's Agents SDK, and similar tools are excellent at building and orchestrating agents. Garuda sits underneath them.

Garuda is not a coding agent. Cursor, Claude Code, and Copilot are excellent at performing software changes. Garuda gives them a verified map of the system they are changing.

Garuda is not a code search tool. Sourcegraph is excellent at finding text and code across many repositories. Garuda answers a different question: not "where does this text appear" but "what is this, what does it depend on, and is it verified."

The category Garuda occupies is smaller and more specific than any of those. It is the verification and state layer that agent frameworks and coding agents call before and after they change your software. The value is not in being another agent — it is in being the thing every agent can rely on.

---

## A week in a team using Garuda

**Monday.** The team writes down a rule:

> *Payment services must never access the customer database directly.*

Garuda reads the rule and connects it to the parts of the code it applies to.

**Tuesday.** An AI agent generates a new payment feature. It introduces a direct database call.

Garuda catches it immediately.

```text
REVIEW — Payment services must not directly access the customer database

Change:     payments.HandleCharge
Rule:       payment-no-direct-db (v1)
Found in:   docs/payments.md
Recorded:   block #147
```

The agent is asked to rework the change. If the rule had been marked stricter, the change would not have been allowed through at all.

**Wednesday.** The architecture changes. The team decides read-only access is now allowed. Garuda updates what "correct" means.

**Thursday.** A second AI agent starts building an adjacent feature. It asks Garuda what rules apply. It gets the *current* answer, not Tuesday's version.

**Friday.** Production telemetry arrives. Garuda can now confirm the payment handler is actually doing what it claims. The requirement moves from "not yet verified" to "verified."

That's the whole loop. It runs continuously, not during quarterly architecture reviews.

---

## What you can do with Garuda

Garuda is easiest to understand through what it lets you do. Here are the workflows teams use it for.

### Onboard engineers into unfamiliar systems

Open Garuda and see your workspace from repositories down to packages, functions, and types. Inspect callers, dependencies, implementations, and cross-repository relationships without requiring a new engineer to reconstruct the architecture manually from scattered files.

### Ground AI agents in the real system

An AI coding agent can ask what calls a function, who implements an interface, or what would be affected by a change. Garuda returns structured answers from the shared semantic model instead of relying only on the file currently open in the agent context.

This reduces unsupported assumptions, invented callers, and unsafe "I think this is safe" changes. Experimental or heuristic relationships remain confidence-scored rather than being presented as authoritative.

### Keep documentation aligned with implementation

Garuda extracts supported normative claims from Markdown, ADR, and plain-text documentation and correlates them with the semantic graph. Claims can become **Supported**, remain **Unverified**, or become **Contradicted** when available runtime evidence conflicts with them.

The goal is not to declare that missing documentation proves missing behavior. The goal is to make the evidence boundary explicit in both directions: what documentation claims, and what the analyzed code actually contains.

### Catch policy violations before merge

Write rules such as *"Payment code must not import the database driver directly"* in policy configuration. Garuda evaluates them against the workspace and produces `ALLOW`, `WARN`, `REVIEW`, or `BLOCK` outcomes.

Policies can run in preview mode or be persisted and anchored when the organization needs an auditable governance decision.

### Understand change impact

Before changing an important symbol, query direct and transitive callers, implementers, subclasses, and related entities. `garuda impact` reports the blast radius present in the graph, including the important cases where the graph does not contain enough evidence to support a stronger claim.

### Connect repositories into one workspace model

A feature may span services, shared libraries, workers, and other repositories. Garuda can resolve cross-repository Go dependencies through the workspace model. Cross-repository resolution is currently beta and should be validated against the intended workspace and tenant boundaries.

### Preserve independently verifiable governance decisions

When a persisted policy evaluation records an allow, warning, review, or block decision, the result can be verified against the Merkle ledger. The proof establishes the integrity and inclusion of the recorded decision; it does not certify that the underlying semantic model is complete or correct.

---

## How the verification works

For claims about your software, Garuda tracks one of three verification states.

| State | What it means |
| :--- | :--- |
| **Supported** | Available evidence supports the claim within the verified scope. |
| **Unverified** | The claim has not received sufficient supporting evidence. This does not mean it is broken. |
| **Contradicted** | Available evidence conflicts with the claim. |

A phrase you'll see a lot in Garuda's design:

> **Absence of evidence is not evidence of absence.**

If Garuda has not seen a piece of code run, it says "unverified" — not "dead" or "broken." This keeps uncertainty visible instead of converting missing evidence into a false negative.

Go analysis is compiler-backed. Python and TypeScript analysis is structural. The confidence and authority of a result therefore depend on the language, relationship type, evidence source, and capability maturity shown in the current release documentation.

---

## Why the record is trustworthy

Every persisted governance decision is written into a tamper-evident Merkle ledger. Anyone with the evaluation ID and verification logic can independently verify that the recorded decision was included at a specific block and has not been altered since.

You can verify any recorded decision with a single command:

```bash
garuda policy verify <decision-id>
```

You do not need access to Garuda's servers, database, or operators to verify the proof. The invariants Garuda is built on are public. [`docs/invariants.md`](docs/invariants.md) lists the rules the truth substrate must never violate.

### What the ledger actually proves

The Merkle ledger proves that a decision was recorded at a specific block, that its content has not been altered since, and that the evaluation that produced it can be independently re-verified.

It does not prove that the semantic model the decision was made against is complete or correct. That is a different claim, and Garuda does not make it.

What Garuda does claim, precisely: **Garuda makes the evidence and decisions behind system governance independently verifiable.** A reader who trusts the Merkle algorithm and has the evaluation ID can confirm what was decided, when, and against which policy version — without trusting Garuda's servers, database, or operators. That is the guarantee. Everything else in this README describes what the model contains, not what the ledger certifies.

This is the piece none of the other tools have. A code graph is easy to build. Making the evidence and decisions behind it independently verifiable — without trusting the operators — is a different thing entirely.

---

## How Garuda differs from what you might already use

Garuda is complementary to most of what you already have. The category it occupies is the layer between your organization's intent and the AI agents changing your code.

| You might already use | What it does well | What Garuda adds |
| :--- | :--- | :--- |
| Agent frameworks (Microsoft Agent Framework, OpenAI Agents SDK) | Building and orchestrating agents | A verified model of the software the agents are operating on. Garuda is what those frameworks call, not a replacement for them. |
| Coding agents (Cursor, Copilot, Claude Code) | Performing software changes with useful context | A structured, evidence-backed map of the system being changed — delivered over MCP, shared across every agent and session |
| Code search (Sourcegraph, GitHub search) | Finding text fast across many repositories | A typed semantic model with stable identity, so a rename does not orphan every reference |
| Code knowledge graphs | Parsing many languages, showing relationships | Verification state — every relationship is `SUPPORTED`, `UNVERIFIED`, or `CONTRADICTED`, and every edge carries source evidence |
| Runtime observability (Datadog, Honeycomb) | Showing what happens in production | Correlating production spans with static entities, so you can ask *"is this declared path actually exercised?"* |
| Doc-code drift tools | Reporting where documentation and code disagree | Running the check in both directions and anchoring the result to a cryptographic record anyone can verify |
| Security and policy tools (Semgrep, others) | Detecting specific classes of violations | A general policy engine with authority attribution, evidence attachment, and Merkle-anchored decisions — not limited to security checks |

The integration path is MCP. Any agent that can call an MCP server can call Garuda.

---

## Built for AI agents, verified across three clients

Garuda speaks the Model Context Protocol — the same language Cursor, Claude Desktop, Codex, and other AI coding tools already use. An agent connected to Garuda can ask twenty-one different questions about your software and get structured, evidence-backed answers back.

The MCP server is verified against three independent clients, each with its own codebase and transport behavior:

| Client | What it proves |
| :--- | :--- |
| Cursor | The server works inside the IDE developers actually use |
| Claude Desktop | The server works with a client that validates responses strictly |
| Codex CLI | The server works with an independent client from a different vendor |

A fourth client is maintained in the repository: a few hundred lines of standard-library Python with no dependencies. It runs the verification suite against the server to check the protocol contract, not merely one vendor's compatibility.

```text
═══ 36/36 checks passed ═══
```

The script lives at [`scripts/mcp_verify.py`](scripts/mcp_verify.py). Run it against your own workspace to verify the server yourself.

### The twenty-one tools

| Group | Tools | What they answer |
| :--- | :--- | :--- |
| Session context | `garuda.briefing` | Workspace state, trust anchor, scale, hubs, policies, contradictions, and recent changes |
| Code graph exploration | `garuda.entities`, `garuda.find_entity`, `garuda.inspect`, `garuda.neighbors`, `garuda.subclasses`, `garuda.implementers`, `garuda.blast_radius`, `garuda.query_claims`, `garuda.query` | What exists, how entities connect, who implements or calls a subject, and what claims apply |
| Governance and decisions | `garuda.policy.list`, `garuda.policy.evaluate`, `garuda.verify_policy_evaluation`, `garuda.governance.status`, `garuda.check_drift`, `garuda.get_lineage`, `garuda.get_impact`, `garuda.detect_contradictions`, `garuda.propose_decision` | What rules apply, whether drift or contradictions exist, and how decisions are proposed and verified |
| Multi-agent coordination | `garuda.handoff`, `garuda.resume` | How agents transfer work through transactional checkpoints |

Read-only tools are everything except `garuda.propose_decision`, `garuda.handoff`, and `garuda.resume`. `garuda.policy.evaluate` is a preview by design and never persists. `garuda.handoff` and `garuda.resume` run in Serializable transactions; a second resume of the same checkpoint returns `{ "status": "not_found" }`.

---

## Who Garuda is for

### Teams already building with AI

If your team uses Cursor, Claude Code, GitHub Copilot, or similar tools and needs stronger guarantees about what those agents are changing, Garuda provides the shared verification layer beneath them.

### Developers joining unfamiliar codebases

If engineers spend weeks reconstructing architecture before making a meaningful commit, Garuda exposes the graph, evidence, and documented constraints they need to start with better context.

### Engineering leaders who need visibility

If you need to answer "where are we drifting?" and "which rules are being violated?" without a three-week audit, Garuda provides workspace-level evidence and governance queries.

### Teams in regulated or high-trust environments

If your software needs an independently verifiable record of governance decisions, Garuda provides technical evidence and cryptographic verification without presenting itself as a compliance certification.

### Platform and security teams

If architectural boundaries, tenant separation, policy enforcement, and audit evidence must be consistent across services and development workflows, Garuda provides a common semantic and governance substrate.

---

## What Garuda is not

To keep the story honest, here's what Garuda isn't:

- **Not an agent framework.** Garuda does not build or orchestrate agents; it provides the verified state and evidence layer they can query.
- **Not a coding agent.** Garuda does not autonomously implement software changes.
- **Not a replacement for human judgment.** Garuda shows evidence and decisions within its verified scope. Your team still decides what action to take.
- **Not a compliance certification.** Garuda produces technical evidence. Whether that evidence satisfies a regulator depends on your process and controls.
- **Not a magic fix for AI errors.** Garuda reduces unsupported assumptions but does not make any model incapable of being wrong.
- **Not a documentation authoring tool.** Write documents where you already maintain them; Garuda ingests and correlates supported normative content.
- **Not a decision-maker.** You define the rules. Garuda evaluates the current workspace against them.
- **Not a complete call graph for every language.** Go call edges are compiler-backed within the supported scope. Python and TypeScript analysis is structural, and dynamic call-graph tracing is experimental. `garuda impact` reports what the graph contains and preserves uncertainty where it does not contain enough evidence.

---

## Current capabilities

### Available today

- **Code understanding across Go, Python, and TypeScript.** Compiler-backed for Go; structural for Python and TypeScript.
- **Document ingestion.** Extract supported normative claims from Markdown, ADR, and plain-text documents. Prose, tables, and code blocks are ignored by design. See [`docs/DOCUMENT_FORMATS.md`](docs/DOCUMENT_FORMATS.md).
- **Verification states.** Track claims as `SUPPORTED`, `UNVERIFIED`, or `CONTRADICTED`.
- **Verifiable decision ledger.** Persist governance decisions and independently verify their Merkle inclusion proofs.
- **Policy engine.** Evaluate rules with `ALLOW`, `WARN`, `REVIEW`, or `BLOCK` outcomes.
- **Interactive workspace.** Explore repositories, packages, entities, relationships, evidence, and governance state.
- **Workspace-wide semantic queries.** Find entities, packages, files, relationships, and claims across the workspace.
- **Impact analysis.** Inspect direct and transitive graph-visible dependencies before changing a symbol.
- **Cross-repository intelligence.** Resolve supported multi-module Go dependencies through the PostgreSQL workspace cache; currently beta and subject to workspace-boundary validation.
- **MCP integration for AI agents.** Twenty-one tools, verified against Cursor, Claude Desktop, and Codex CLI, plus a reference implementation that checks the protocol contract. Run [`scripts/mcp_verify.py`](scripts/mcp_verify.py) yourself.
- **Transactional multi-agent coordination.** Transfer work through Serializable handoffs and single-consumption checkpoints.
- **CLI and HTTP APIs.** Script and automate supported workflows.
- **CI/CD integration.** Run analysis, policy evaluation, semantic diff, and impact checks in build pipelines.

Everything in this list is exercised end to end within its documented scope. The specific evidence and reproducible queries live in [`EVIDENCE.md`](EVIDENCE.md).

### In beta with design partners

- **Team accounts and access controls.** Implemented; early partner onboarding opens once the trust-floor audit completes.
- **Runtime verification at scale.** Connecting live production telemetry to a Garuda workspace.
- **Multi-repository benchmarks.** Formal quality gates for workspaces of 10, 25, and 100+ repositories.

### Launching soon

- **More languages.** Additional analyzers on the same semantic pipeline.
- **More agent tools.** Deeper integrations for agents that need impact and evidence, not just structure.
- **Deeper dashboard exploration.** Richer views of evidence, drift, and system health.

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

### Create a workspace

```bash
export GARUDA_WORKSPACE=my-workspace
./bin/garuda workspace create my-workspace
```

### Point Garuda at your code

```bash
./bin/garuda analyze /path/to/repo --save
```

Garuda detects Go, Python, and TypeScript projects from their module or project configuration files.

### Add your documentation

```bash
./bin/garuda docs ingest ./docs
```

### Evaluate your policies

```bash
./bin/garuda policy evaluate ./policies --workspace my-workspace
```

This is a preview by default — nothing is persisted and nothing is anchored. Add `--save` when you want a cryptographically recorded governance decision:

```bash
./bin/garuda policy evaluate ./policies --workspace my-workspace --save
```

### Connect your AI agent

Add Garuda to your MCP client's configuration:

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

Use an absolute binary path. Restart the MCP client after changing its configuration.

### Verify the MCP server yourself

```bash
WORKSPACE=my-workspace python3 scripts/mcp_verify.py
```

You should see `36/36 checks passed`. The script exercises every tool, including negative paths — an unknown ID must return a structured failure, not a transport error.

### Open the workspace

```bash
./bin/garuda dev
```

Then open `http://localhost:8080` in your browser.

---

## How it works, at a glance

```mermaid
flowchart LR
    DOCS["Your docs and rules"] --> GARUDA
    CODE["Your code"] --> GARUDA
    RUNTIME["Your running system"] --> GARUDA
    GARUDA["Garuda builds one shared understanding"]
    GARUDA --> VIEWS["Dashboard · CLI · AI agents · CI"]
    GARUDA --> RECORD["Verifiable decision log"]
```

Everything you see in Garuda — the graph, search results, policy checks, drift reports, and agent answers — comes from the same shared understanding of your software.

The graph is one way to look at it. The AI agent interface is another. Neither is the product on its own. The shared, evidence-aware state underneath them is the product.

---

## Where Garuda is going

Garuda's long-term direction is simple to say and hard to build:

> **Keep an organization's intent continuously connected to the software that implements it and the system that runs it.**

That means starting with what you have today — code, docs, and rules — and gradually connecting more of your organization's knowledge: runtime behavior, incident history, architectural decisions, and business rules. Each connection makes the understanding richer. None of them change what Garuda fundamentally is.

The vision is a world where software does not quietly lose the context behind why it was built. Where an AI agent three years from now can still tell you *why* a boundary exists, not just *that* it exists.

---

## Documentation

| Document | What's in it |
| :--- | :--- |
| [`PLAYBOOK.md`](PLAYBOOK.md) | Installation, operational procedures, and workflow examples |
| [`EVIDENCE.md`](EVIDENCE.md) | Validation results and methodology |
| [`docs/adr/`](docs/adr/) | Architecture decision records |
| [`docs/DOCUMENT_FORMATS.md`](docs/DOCUMENT_FORMATS.md) | Supported documents and claim-extraction format |
| [`docs/invariants.md`](docs/invariants.md) | Invariant contract for Garuda's truth substrate |
| [`SECURITY.md`](SECURITY.md) | Security model and issue-reporting process |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | Contribution guidelines |

---

## Contributing

Garuda is built in the open. Contributions are welcome — new language support, better parsers, deeper AI integrations, documentation improvements, and verification tooling. See [`CONTRIBUTING.md`](CONTRIBUTING.md) to get started.

---

## License

Apache License 2.0. See [LICENSE](LICENSE).

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