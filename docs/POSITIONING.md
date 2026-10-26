# Garuda Positioning

**Internal reference.** This document is the canonical statement of what Garuda is, what it is not, and how it should be described. All public-facing material — README, pitch decks, website copy, investor emails — derives from this document. If a claim is not in this document, it is not a claim Garuda makes.

---

## The one-sentence pitch

Garuda keeps AI agents aligned with what you're building.

## The shorter pitch

Garuda gives everyone working on your software — people and AI agents — one shared, verified understanding of what the software actually is.

## The problem

AI can now write, change, test, and fix software faster than any team can review it. That speed is real. It is also creating a failure mode that no one has named yet.

A team writes down what they want to build. A month later, the software has moved somewhere else. Not dramatically. Not with a rewrite. A little at a time, change by change. Documentation says one thing, the code does another, and the running system does a third. Nobody notices until something breaks in production.

We call this **AI Development Drift**. It is the growing gap between what an organization intended, what got built, and what the system actually does.

It is not a bug in any single AI model. It is the structural consequence of asking fast-moving agents to reason about a system they can only partially see, across sessions they cannot remember, against intentions that keep changing.

The three failure modes that appear together as autonomy increases:

**Agents operate on incomplete and changing context.** Even with a perfect initial context, the system changes while the agent works. Another agent modifies something. A policy changes. Documentation moves. A dependency shifts. The problem is not memory. It is continuous alignment with a moving source of truth.

**Agents ship changes that break invariants.** A change can be syntactically valid, type-correct, tested, and reasonable in review, and still violate a constraint that defines the system. Compilation and tests are the wrong layer of correctness for this failure mode.

**The organization loses the ability to re-align.** When documentation is deep and code has moved on, humans and AI both struggle to reconstruct what happened. Re-alignment takes weeks. Sometimes the accumulated changes redefine the vision entirely, and nobody notices until a customer does.

## The solution

Garuda builds a persistent semantic model of a software system. Every element in that model — an entity, a relationship, a decision, a policy outcome — is backed by evidence pointing to the exact source file and line that justifies it. Nothing is asserted without a pointer to why.

Then Garuda verifies. A declarative policy states organizational intent. An evaluator checks it against the model. The result is a deterministic decision — allow, warn, review, or block — anchored to a cryptographic ledger alongside the evidence it cited. An auditor with no access to Garuda's infrastructure can independently confirm the decision happened at a specific point in the ledger's history.

The same verified model is served to developers through a CLI and a dashboard, and to AI agents through a Model Context Protocol server. Neither side reconstructs the codebase from scratch. Both see the same verified state.

## What Garuda is

- **An alignment layer.** It keeps what an organization intended, what got built, and what actually runs connected as the system evolves.
- **A verified semantic model.** It connects code structure, documentation, runtime observations, and organizational decisions into one inspectable graph, where every edge carries the source file and line that justifies it.
- **A cryptographic audit trail.** Every governance decision is permanent and independently checkable without trusting Garuda.
- **A shared workspace.** It serves developers and AI agents from the same model, so neither has to reconstruct the same reality independently.

## What Garuda is not

- **Not a code search engine.** Search returns results. Garuda returns verified facts.
- **Not a knowledge graph product.** The graph is the substrate. The verification layer on top of it is the product.
- **Not an observability platform.** Runtime telemetry is one evidence source. Garuda is interested in whether runtime behavior confirms or contradicts what the architecture intended.
- **Not a documentation generator.** Documents are one source of intent. Garuda checks whether the code matches them, not the reverse.
- **Not an AI coding agent.** Garuda does not write code. It tells agents and reviewers what is known, what is uncertain, and what changed.
- **Not a compliance certification.** Garuda produces technical evidence. Whether that evidence satisfies a regulator depends on the deployment and process, not on us.
- **Not a replacement for human judgment.** Garuda shows what is true about the system. What to do about it is still a decision for the team.

## The moat

Three capabilities, held together, that no current category of software delivers:

**1. Evidence-backed claims.** Every relationship in the model carries the file, line, commit, and analyzer version that produced it. An engineer can click any edge and land on the exact source line that justifies it. Nothing in the model is asserted without proof.

**2. Merkle-anchored governance.** Every policy decision is anchored to the same cryptographic ledger as the evidence it cited. Any party can independently re-derive the inclusion proof and verify that a specific decision was committed at a specific block height — without trusting Garuda, without access to Garuda's database, without reading Garuda's source.

**3. Deterministic verification.** Wherever possible, extraction and evaluation are deterministic. The same source, analyzed twice, produces byte-identical output. Where determinism is not possible, the model labels the difference explicitly. Model-assisted reasoning is permitted. Model-assisted fabrication is not.

## The refusal to fabricate

Garuda treats `UNVERIFIED` as a first-class state, not a failure mode. Absence of evidence is never presented as certainty. Every metric on every dashboard either has real data or explicitly says "not measured" with a reason.

This is a product principle and a technical discipline. Every number the system produces can be traced to a query, a golden vector, or a frozen test fixture. A claim that cannot be traced is not a claim Garuda makes.

The same discipline applies to the marketing. The README describes capabilities that ship. The roadmap describes capabilities that do not yet ship. The two are never mixed. When a claim is Beta, it says Beta. When a number is measured on a specific corpus, the corpus is named.

## The competitive landscape

Garuda operates at the intersection of five categories that have historically been separate:

| Category | What it delivers | What Garuda adds on top |
| :--- | :--- | :--- |
| Code search (Sourcegraph, GitHub search) | Find text fast across many repositories | A typed semantic model with stable identity — a rename does not orphan every reference |
| Code knowledge graphs (Graphify and similar) | Parse many languages, show relationships | Verification state on every relationship, with source file and line as evidence |
| Runtime observability (Datadog, Honeycomb) | Show what is happening in production | Correlate production spans with static entities to answer "is this declared path actually exercised?" |
| Documentation and drift tools | Connect organizational knowledge or flag drift | Two-way check with a cryptographic record anyone can verify |
| AI coding agents (Cursor, Copilot, Claude Code) | Write code fast with useful context | Give those agents a structured, evidence-backed map of the system they are changing |

These categories are complementary, not competing. Garuda is the layer that connects them, not a replacement for any one of them.

## The buyer

The initial buyer is not "software teams." It is narrower and more urgent.

**Primary buyer:** AI-native engineering teams building toward increasingly autonomous software development. Teams using coding agents heavily, running multiple agents concurrently, shipping rapidly, and concerned about architectural consistency as more code is generated than reviewed.

**Economic buyer:** VP Engineering, CTO, Head of Engineering, Head of AI Engineering.

**Daily users:** Developers, platform engineers, and the AI agents themselves.

**Expansion path:** Once Garuda becomes trusted inside the engineering workflow, the same intelligence layer extends to organization-wide software understanding, governance workflows, and runtime integrity. Business-critical workflow integrity is a longer-term direction, not a current product.

## The one metric that matters

Time from a question to a source-backed answer.

Not tokens saved. Not nodes in the graph. Not languages supported. The measure that determines whether Garuda is useful is how quickly a developer, an architect, or an AI agent can reach a conclusion they are willing to defend, with evidence attached.

## What Garuda publishes

- Measured coverage numbers, published with methodology.
- Per-language resolution rates, updated when the analyzers change.
- Benchmark results from versioned, reproducible corpora.
- The exact scope of every claim: what was checked, what was not, and why.

Garuda does not publish percent-precision figures without a named corpus. It does not publish "zero hallucinations" or "100% accuracy." It does not claim capabilities that its roadmap lists as planned.

## Where Garuda is today

Stated plainly, because this is the internal reference and the current state matters more here than anywhere else.

**Stable today:**
- Go analyzer (compiler-backed), Python and TypeScript analyzers (structural)
- Merkle v1 trust layer with RFC 6962 compliance and independent verification
- Policy engine with five predicates and four outcomes
- Semantic graph explorer with repository, package, entity, and neighborhood views
- MCP server with sixteen tools, verified against Cursor and a reference client
- CLI and HTTP APIs for every capability
- CI integration via `garuda ci` and `garuda judge`
- Documentation ingestion for Markdown, ADR, and plain text

**In beta:**
- Documentation-to-code alignment at scale
- Cross-repository intelligence beyond same-language workspaces
- Runtime verification against live production telemetry

**Launching soon:**
- Additional language analyzers (the pipeline is language-agnostic; new languages plug in without schema changes)
- Expanded MCP surface
- Deeper dashboard exploration

**Not yet started:**
- Billing and subscription infrastructure
- Enterprise identity (SSO, workspace membership, per-workspace access control)
- Hosted Merkle trust layer (current trust layer runs locally)

**Honest gap:** no paying customer yet. Design partner conversations are in progress. The first paid pilot is the goal for the next quarter. Charging before that would be guessing.

## What the roadmap is not

The roadmap is not a marketing document. It is not a promise. It is a record of intent, subject to change based on what design partners ask for. If a feature is on the roadmap and not in the README, it has not shipped. If a feature is on the roadmap and someone asks about it, the honest answer is "planned, not yet shipped." Never "coming soon."

## The one line to remember

> Garuda keeps AI agents aligned with what you're building.

Everything else in this document is evidence for that line, or an honest limit on it.