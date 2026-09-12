# Garuda Positioning

**Internal reference.** This document is the canonical statement of
what Garuda is, what it is not, and how it should be described. All
public-facing material — README, pitch decks, website copy, investor
emails — derives from this document. If a claim is not in this
document, it is not a claim Garuda makes.

## The One-Sentence Pitch

Garuda is the verification layer for modern software: every change is
checked against what the organization intended, what the code
implements, and what production actually does — and every check is
independently verifiable.

## The Problem

Software organizations now run faster than they can verify. AI coding
agents produce changes at a rate that human review cannot match. The
knowledge needed to judge whether a change is correct — what the
architecture intended, what the code actually does, what production
observes — is scattered across source files, design documents, tickets,
runtime systems, and engineers' memory.

The result is that "correct" and "consistent with what we decided" have
drifted apart. A change can pass its tests and still violate an
architectural decision made three years ago. Nobody catches it, because
nobody has the full picture in front of them at the moment of review.

Existing tools address parts of this problem. Search engines find
symbols. Documentation platforms organize knowledge. Observability
platforms record runtime behavior. None of them connect these layers
into a single model that can be verified.

## The Solution

Garuda builds a persistent semantic model of a software system. Every
element in that model — an entity, a relationship, a decision, a policy
outcome — is backed by evidence pointing to the exact source, commit,
and line that justifies it. Nothing is asserted without a pointer to
why.

Then Garuda verifies. A YAML policy declares organizational intent. An
evaluator checks it against the semantic model. The result is a
deterministic decision — allow, warn, review, or block — anchored to a
cryptographic ledger alongside the evidence it cited. An auditor with
no access to Garuda's infrastructure can independently confirm the
decision happened at a specific point in the ledger's history.

## What Garuda Is

- **A verification layer.** It answers the question "should this change
  be allowed?" with evidence, not with a score.
- **A semantic model.** It connects code structure, documentation,
  runtime observations, and organizational decisions into one
  inspectable graph.
- **A cryptographic audit trail.** It makes the reasoning behind every
  decision permanent and independently checkable.
- **A shared workspace.** It serves developers, architects, and AI
  agents from the same model, so they stop reconstructing the same
  reality independently.

## What Garuda Is Not

- **Not a code search engine.** Search is a commodity. Garuda is
  interested in what search cannot answer: whether a relationship is
  true, and whether it should be permitted.
- **Not a knowledge graph product.** The graph is the substrate. The
  verification layer on top of it is the product. Competing on graph
  visualization features is a category Garuda deliberately does not
  enter.
- **Not an observability platform.** Runtime telemetry is one evidence
  source among several. Garuda is interested in whether runtime
  behavior confirms or contradicts what the architecture intended.
- **Not a documentation generator.** Documents are one source of
  organizational intent. Garuda checks whether the code matches them,
  not the reverse.
- **Not an AI coding agent.** Garuda does not write code. It tells
  agents and reviewers what is known, what is uncertain, and what
  changed.

## The Moat

Three capabilities, held together, that no current category of
software delivers:

**1. Evidence-backed claims.** Every relationship in the model carries
the file, line, commit, and analyzer version that produced it. An
engineer can click any edge and land on the exact source line that
justifies it. Nothing in the model is asserted without proof.

**2. Merkle-anchored governance.** Every policy decision is anchored
to the same cryptographic ledger as the evidence it cited. Any party
can independently re-derive the inclusion proof and verify that a
specific decision was committed at a specific block height — without
trusting Garuda, without access to Garuda's database, without reading
Garuda's source.

**3. Deterministic verification.** Wherever possible, extraction and
evaluation are deterministic. The same source, analyzed twice, produces
byte-identical output. Where determinism is not possible, the model
labels the difference explicitly. Model-assisted reasoning is
permitted; model-assisted fabrication is not.

## The Refusal To Fabricate

Garuda treats `UNVERIFIED` as a first-class state, not a failure mode.
Absence of evidence is never presented as certainty. Every metric on
every dashboard either has real data or explicitly says "not
measured" with a reason.

This is a product principle and a technical discipline. Every number
the system produces can be traced to a query, a golden vector, or a
frozen test fixture. A claim that cannot be traced is not a claim
Garuda makes.

## The Competitive Landscape

Garuda operates at the intersection of four categories that have
historically been separate:

| Category | What it delivers | What Garuda adds on top |
|---|---|---|
| Code search and code graphs | Find symbols and structure | Verify that a relationship is true, and whether it is permitted |
| Documentation and knowledge platforms | Connect organizational knowledge | Check whether code matches the decisions those documents describe |
| Observability and runtime telemetry | Record what systems do | Correlate runtime behavior with architectural intent |
| Engineering governance and policy | Declare standards | Enforce standards deterministically and anchor the outcome cryptographically |

Garuda does not compete with any of these directly. It consumes
evidence from each of them and produces a verified state that none of
them produce alone.

## The Buyer

Garuda is bought by engineering leaders at organizations where three
conditions hold:

1. **Software changes fast enough that human review cannot keep up.**
   Typically this means AI coding agents are already in use.
2. **The cost of architectural drift is material.** Regulated
   industries, payment systems, and any organization where a wrong
   change has compliance or financial consequences.
3. **Evidence matters.** Auditors, compliance officers, or technical
   due-diligence processes require proof of how decisions were made.

The buyer is not looking for a prettier graph. They are looking for a
defensible answer to "why is this software in this state, and can you
prove it?"

## The One Metric That Matters

Time from a question to a source-backed answer.

Not "tokens saved." Not "nodes in the graph." Not "languages
supported." The measure that determines whether Garuda is useful is how
quickly a developer, an architect, or an AI agent can reach a
conclusion they are willing to defend, with evidence attached.

## What Garuda Publishes

- Measured coverage numbers, published with methodology.
- Per-language resolution rates, updated when the analyzers change.
- Benchmark results from versioned, reproducible corpora.
- The exact scope of every claim: what was checked, what was not, and
  why.

Garuda does not publish percent-precision figures without a named
corpus. It does not publish "zero hallucinations" or "100% accuracy."
It does not claim capabilities that its roadmap lists as planned.