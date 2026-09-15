<p align="center">
  <img src="assets/garuda-logo.png" width="128" alt="Garuda">
</p>

<h1 align="center">Garuda</h1>

<p align="center">
  <strong>Keep AI agents aligned with what you're building.</strong>
</p>

<p align="center">
  <em>A shared understanding of your software — for the humans and AI agents building it together.</em>
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/releases"><img src="https://img.shields.io/badge/version-v0.1.x-blue.svg?style=flat-square" alt="Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache_2.0-green.svg?style=flat-square" alt="License"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.26-00ADD8.svg?style=flat-square" alt="Go Version"></a>
  <a href="scripts/mcp_verify.py"><img src="https://img.shields.io/badge/MCP-16%20tools%2C%2018%2F18%20verified-success.svg?style=flat-square" alt="MCP Verified"></a>
</p>

---

## The problem everyone is quietly having

AI can now write, change, test, and fix software faster than any team can review it.

That speed is real. It's also creating a new kind of problem that nobody has a good name for yet.

A team writes down what they want to build. A month later, the software has moved somewhere else. Not dramatically — not with a big rewrite — but a little at a time, change by change. Documentation says one thing, the code does another, and the running system does a third. Nobody notices until something breaks in production.

This happens with human developers too, but slowly. With AI agents, it happens fast enough that teams can lose track of their own software within weeks.

> **AI Development Drift is the growing gap between what an organization intended, what got built, and what the system actually does.**

It is not a bug in any single AI model. It's the natural result of asking fast-moving agents to reason about a system they can only partially see, across sessions they can't remember, against intentions that keep changing.

---

## What Garuda does about it

Garuda gives everyone working on your software — people and AI agents — **one shared, verified understanding of what the software actually is.**

It connects the four things that normally live apart:

- What your organization **says** it wants (docs, plans, policies, requirements)
- What was **actually built** (your code)
- What the system **actually does** (runtime behavior)
- What **evidence** supports each of those claims

And it continuously checks whether they still agree.

```mermaid
flowchart TD
    INTENT["What you intend"] --> GARUDA["Garuda"]
    CODE["What you built"] --> GARUDA
    RUNTIME["What it actually does"] --> GARUDA
    GARUDA --> CHECK["Continuous check"]
    CHECK --> ALIGNED["Aligned"]
    CHECK --> DRIFT["Drifting"]
    CHECK --> UNKNOWN["Not enough evidence yet"]
    ALIGNED --> ACTION["Humans and AI act"]
    DRIFT --> ACTION
    UNKNOWN --> ACTION
    ACTION --> GARUDA
```

When a developer or an AI agent proposes a change, Garuda checks it against the intent your organization has already committed to. When something drifts, you hear about it while it's still a small problem — not three months later during an incident review.

---

## A week in a team using Garuda

**Monday.** The team writes down a rule:

> *Payment services must never access the customer database directly.*

Garuda reads the rule and connects it to the parts of the code it applies to.

**Tuesday.** An AI agent generates a new payment feature. It introduces a direct database call.

Garuda catches it immediately.

```
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

Garuda is easiest to understand through what it lets you do. Here are the things teams actually use it for.

### Onboard a new engineer in days instead of weeks

Open Garuda and see your whole workspace. Start at the level of repositories. Drill down into packages. Open a package and see its functions, its types, and how they connect. Click any function and see who calls it, what it depends on, and where it's used across every repository you own.

No more "read the code for two weeks before your first commit."

### Give AI agents the real answer instead of a guess

Your AI coding agent needs to know what calls a function, who implements an interface, or what breaks if you change something. Today it guesses from the file it has open. With Garuda, it asks directly and gets a real answer from a verified map of your code.

This means fewer hallucinated function names, fewer made-up callers, fewer "I think this is safe" moments that turn out not to be.

### Keep your documentation honest

Your documentation says a feature exists. Your code no longer does. Garuda can tell you that, and show you exactly where the mismatch is.

It works the other way too — if you have important code that nobody wrote down anywhere, Garuda will surface it as a knowledge gap.

### Catch policy violations before they merge

Write a rule like *"Payment code must not import the database driver directly"* in plain configuration, and Garuda will check every change against it. Violations get flagged, sent for review, or blocked depending on how strict you want to be.

The rule stops being something everyone is supposed to remember. It becomes something the system actually checks.

### Understand the impact of a change before you make it

Before you change an important function, ask Garuda who depends on it. See direct callers, indirect callers, and cross-repository connections in under a second.

This is the difference between *"I think this is safe"* and *"here are the eleven services that call this, and seven of them run in production."*

### Compare what your code says with what your system does

Source code describes what the system is designed to do. Runtime behavior shows what it actually does. Garuda connects the two. When they agree, you have real confidence. When they disagree, Garuda shows you the mismatch instead of letting you discover it during an incident.

### Work across all your repositories at once

Modern software doesn't live in one place. A single feature might touch a frontend, two services, a shared library, a database, and a background worker. Garuda understands them as one system, not as isolated codebases.

### Keep a verifiable record of every important decision

When Garuda makes a governance decision — allowing a change, blocking a change, flagging something for review — that decision is recorded in a way that anyone can independently verify later. No trust required in Garuda itself. If you need to show an auditor exactly what was decided and when, you can.

---

## How the verification works (in plain English)

For every claim about your software, Garuda keeps track of one of three states.

| State | What it means |
| :--- | :--- |
| **Supported** | We have evidence. The code exists, and the runtime agrees. |
| **Unverified** | The code exists, but we haven't seen it actually run yet. This does *not* mean it's broken — it just means we don't have proof. |
| **Contradicted** | Something is wrong. The system is behaving differently from what the code or the rules say it should. |

A phrase you'll see a lot in Garuda's design:

> **Absence of evidence is not evidence of absence.**

If Garuda hasn't seen a piece of code run, it says "unverified" — not "dead" or "broken." This is important because it means Garuda will never confidently tell you something is wrong when it simply doesn't know yet.

This is the same discipline good engineers already follow. Garuda just makes it systematic.

---

## Why the record is trustworthy

Every decision Garuda makes is written into a tamper-proof log. Anyone — including someone outside your organization — can independently verify that a decision was made at a specific moment and hasn't been changed since.

Think of it like a notarized ledger. Not because we expect people to lie, but because when decisions matter, "trust us" is a weaker guarantee than "verify us."

You can verify any recorded decision with a single command, without needing access to our systems, our database, or our keys.

This is the piece none of the other tools have. A code graph is easy to build. A code graph you can *prove* is a different thing entirely.

---

## How Garuda differs from what you might already use

If you are already using one of these categories of tool, here is what Garuda adds that the category does not.

| You might already use | What it does well | What Garuda adds |
| :--- | :--- | :--- |
| Code search (Sourcegraph, GitHub search) | Finds text fast across many repositories | A typed semantic model with stable identity, so a rename does not orphan every reference |
| Code knowledge graphs (Graphify and similar) | Parses many languages, shows relationships | Verification state — every relationship is `SUPPORTED`, `UNVERIFIED`, or `CONTRADICTED`, and every edge carries the source file and line that justifies it |
| Runtime observability (Datadog, Honeycomb) | Shows what is happening in production | Correlates production spans with static entities, so you can ask *"is this declared path actually exercised?"* |
| Doc-code drift tools | Reports where documentation and code disagree | Runs the check in both directions and anchors the result to a cryptographic record anyone can verify |
| AI coding agents (Cursor, Copilot, Claude Code) | Writes code fast with useful context | Gives those agents a structured, evidence-backed map of the system they are changing, over MCP |

These are complementary, not competing. Garuda is the layer that connects them, not a replacement for any one of them.

---

## Built for AI agents, verified like a spec

Garuda speaks the Model Context Protocol — the same language Cursor, Claude Desktop, and other AI coding tools already use. An agent connected to Garuda can ask sixteen different questions about your software and get structured answers back.

We didn't stop at "it works in one client." We wrote a second, tiny MCP client from scratch — a few hundred lines of Python, no dependencies — and ran the same test suite against both.

```
═══ 18/18 checks passed ═══
```

The script lives at [`scripts/mcp_verify.py`](scripts/mcp_verify.py). If you want to run it yourself against your own workspace, you can. That's the point.

The sixteen tools cover the questions agents actually ask:

- *What is around this entity?* — `garuda.neighbors`
- *Who inherits from this?* — `garuda.subclasses`
- *Who implements this interface?* — `garuda.implementers`
- *Where is this by name?* — `garuda.find_entity`
- *What would break if I changed this?* — `garuda.blast_radius`, `garuda.get_impact`
- *What are the rules here?* — `garuda.policy.list`, `garuda.policy.evaluate`
- *Is anything drifting?* — `garuda.check_drift`
- *What did we decide about this?* — `garuda.get_lineage`, `garuda.query`

Plus inspection, entity listing, governance status, and decision proposal.

---

## Who Garuda is for

### Teams already building with AI

If your team uses Cursor, Claude Code, GitHub Copilot, or similar tools and you're starting to worry about what those agents are quietly changing, Garuda is for you.

### Developers joining unfamiliar codebases

If you've ever spent two weeks reading code before making your first meaningful commit, Garuda is for you.

### Engineering leaders who need visibility

If you need to answer "where are we drifting?" and "which rules are being violated?" without a three-week audit, Garuda is for you.

### Teams in regulated or high-trust environments

If your software has to prove what it does, Garuda gives you a verifiable record without needing a compliance team to maintain it manually.

---

## What Garuda is not

To keep the story honest, here's what Garuda isn't:

- **Not a replacement for human judgment.** Garuda shows you what's true about your system. What to do about it is still a decision for your team.
- **Not a compliance certification.** Garuda produces technical evidence. Whether that evidence satisfies a regulator depends on your process, not on us.
- **Not a magic fix for AI errors.** Garuda dramatically reduces the opportunity for AI agents to work from wrong assumptions. It doesn't make any model incapable of being wrong.
- **Not a documentation authoring tool.** Write your docs where you already write them. Garuda reads them and connects them to your code.
- **Not a decision-maker.** You define the rules. Garuda checks whether your software still follows them.
- **Not a complete call graph for every language.** Go call edges are resolved through the compiler, but a single-language repository analyzed in isolation will have sparse call edges at the leaf — many functions have exactly one caller, and 90% have four or fewer. This is a property of the code, not a limitation of the analysis. `garuda impact` reports what the graph contains, and the graph is honest about what it does not contain. Structural edges — inheritance, implementation, imports — are dense and reliable; call edges are denser in workspaces with more repositories that consume each other.

---

## Current capabilities

### Available today

- **Code understanding across Go, Python, and TypeScript.** Compiler-grade for Go. Structural for Python and TypeScript.
- **Document ingestion (Beta).** Extract claims from Markdown and ADR documents that use the supported format, and match them to your code. Prose, tables, and code blocks are ignored by design. See [`docs/DOCUMENT_FORMATS.md`](docs/DOCUMENT_FORMATS.md).
- **Verification states.** Every claim is tracked as supported, unverified, or contradicted.
- **Tamper-proof decision log.** Every governance decision is anchored to a verifiable cryptographic record.
- **Policy engine.** Write rules in simple configuration. Garuda evaluates them against your system and produces clear outcomes: allow, warn, review, or block.
- **Interactive workspace.** Explore your entire system visually, from repositories down to individual functions.
- **Workspace-wide search.** Find any symbol, package, file, or repository across your organization.
- **Impact analysis.** See who depends on a change before you make it.
- **Cross-repository intelligence (Beta).** Understand software that spans multiple services. Same-language workspaces work today; cross-language resolution is heuristic.
- **MCP integration for AI agents.** Sixteen tools, verified against two independent clients — Cursor and a reference implementation we wrote ourselves to check the specification, not just one vendor's compatibility. Run the verification yourself: [`scripts/mcp_verify.py`](scripts/mcp_verify.py).
- **CLI and HTTP APIs.** Everything Garuda can do is scriptable and automatable.
- **CI/CD integration.** Run Garuda's checks as part of your existing build pipeline.

### Measured today

The figures below are from a workspace with nine repositories (Go, Python, and TypeScript) analyzed end to end. They are reproducible: run `garuda analyze` on the same commits and you will see the same counts.

| Metric | Count |
| :--- | :--- |
| Repositories analyzed | 9 |
| Entities extracted | 14,333 |
| Claims recorded | 37,511 |
| Languages covered | Go (Tier 5), Python (Tier 2), TypeScript (Tier 2) |
| Cross-repository edges | 1 |

Go is at **Tier 5** — compiler-resolved. Python and TypeScript are at **Tier 2** — imports resolved, calls heuristic. The distinction is honest: a claim about a Python function call is a weaker claim than a claim about a Go function call, and Garuda says so.

### In beta with design partners

- **Team accounts and access controls.** Running today in early partner environments. Enterprise single-sign-on on the way.
- **Runtime verification at scale.** Connecting live production telemetry to your Garuda workspace.
- **Multi-repository benchmarks.** Formal quality gates for workspaces of 10, 25, and 100+ repositories.

### Launching soon

- **More languages.** Rust, Java, and Elixir on the same pipeline. No changes to how you use Garuda — new languages just show up.
- **More agent tools.** Deeper integrations for AI agents that need to understand impact and coverage, not just structure.
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

Requires Go 1.26 or later. TypeScript support needs a C compiler.

### Or start the full stack with Docker

```bash
docker compose -f deploy/compose/docker-compose.prod.yml up -d
```

This starts Garuda, its database, and the background worker in one command.

### Point Garuda at your code

```bash
./bin/garuda analyze /path/to/repo --save --workspace my-workspace
```

Garuda figures out what language your project uses automatically.

### Add your documentation

```bash
./bin/garuda docs ingest ./docs --workspace my-workspace
```

### Evaluate your policies

```bash
./bin/garuda policy evaluate ./policies --workspace my-workspace
```

### Connect your AI agent

Add Garuda to your MCP client's configuration:

```json
{
  "mcpServers": {
    "garuda": {
      "command": "/absolute/path/to/bin/garuda-mcp",
      "env": {
        "DATABASE_URL": "postgres://user:pass@localhost:5432/garuda?sslmode=disable"
      }
    }
  }
}
```

### Verify the MCP server yourself

```bash
WORKSPACE=my-workspace python3 scripts/mcp_verify.py
```

You should see `18/18 checks passed`.

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

Everything you see in Garuda — the graph, the search results, the policy checks, the drift reports — comes from the same shared understanding of your software.

The graph is one way to look at it. The AI agent interface is another. Neither one is the product on its own. The shared understanding underneath them is the product.

---

## Where Garuda is going

Garuda's long-term direction is simple to say and hard to build:

> **Keep an organization's intent continuously connected to the software that implements it and the system that runs it.**

That means starting with what you have today — code, docs, rules — and gradually connecting more of your organization's knowledge: runtime behavior, incident history, architectural decisions, business rules. Each connection makes the understanding richer. None of them change what Garuda fundamentally is.

The vision is a world where software doesn't quietly lose the context behind why it was built. Where an AI agent three years from now can still tell you *why* a boundary exists, not just *that* it exists.

---

## Documentation

| Document | What's in it |
| :--- | :--- |
| [`PLAYBOOK.md`](PLAYBOOK.md) | Installation, operational procedures, workflow examples |
| [`EVIDENCE.md`](EVIDENCE.md) | Validation results and methodology |
| [`docs/adr/`](docs/adr/) | Architecture decision records |
| [`docs/DOCUMENT_FORMATS.md`](docs/DOCUMENT_FORMATS.md) | What documents Garuda reads, and how to format them |
| [`docs/invariants.md`](docs/invariants.md) | The twelve code-level invariants Garuda is built on |
| [`SECURITY.md`](SECURITY.md) | Security model and how to report issues |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | How to contribute |

---

## Contributing

Garuda is built in the open. Contributions are welcome — new language support, better parsers, deeper AI integrations, documentation improvements. See [`CONTRIBUTING.md`](CONTRIBUTING.md) to get started.

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