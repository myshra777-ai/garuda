

```markdown
# 🦅 Garuda

[![Release](https://img.shields.io/github/v/release/myshra777-ai/garuda?style=flat-square&color=blue)](https://github.com/myshra777-ai/garuda/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/myshra777-ai/garuda/garuda-ci.yml?branch=main&style=flat-square&label=CI)](https://github.com/myshra777-ai/garuda/actions)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg?style=flat-square)](LICENSE)
[![Benchmarks](https://img.shields.io/badge/Benchmarks-100%25-brightgreen?style=flat-square)](#-ast-semantic-benchmark-suite)

### Evidence-backed software intelligence for understanding, verifying, and governing software systems.

Garuda turns software repositories into a deterministic, inspectable semantic model of **entities, relationships, observations, claims, evidence, decisions, and historical state**.

It is designed around a simple principle:

> **Software intelligence should be backed by evidence, not reconstructed from guesses.**

Today, Garuda focuses on compiler-backed Go semantic analysis, deterministic identity, semantic relationships, impact analysis, evidence provenance, immutable state, cryptographic verification, multi-repository workspaces, and progressive architectural exploration.

The architecture is being extended with runtime telemetry so the same evidence model can eventually connect:

```text
Code
  ↓
Semantics
  ↓
Relationships
  ↓
Evidence
  ↓
Runtime Observations
  ↓
Claims
  ↓
Verification
  ↓
Architectural Decisions
  ↓
Governance
```

---

# 🎯 The Garuda Thesis

Most engineering tools answer questions from a single source of information.

Static analysis sees code.

Observability sees runtime behavior.

Documentation describes intended architecture.

AI agents generate conclusions from whatever context they receive.

Garuda is designed to connect these layers into a persistent semantic substrate.

```text
                 ┌───────────────────────┐
                 │       SOURCE CODE     │
                 └───────────┬───────────┘
                             │
                             ▼
                    AST + go/types
                             │
                             ▼
                ┌────────────────────────┐
                │  SEMANTIC GRAPH        │
                │                        │
                │  Entities              │
                │  Relationships         │
                │  Dependencies          │
                └───────────┬────────────┘
                            │
                            ▼
                      EVIDENCE LAYER
                            │
               ┌────────────┼────────────┐
               ▼            ▼            ▼
            Source        Commit       Content
             Line          SHA          Hash
               │            │            │
               └────────────┼────────────┘
                            ▼
                         CLAIMS
                            │
                 ┌──────────┼──────────┐
                 ▼          ▼          ▼
             SUPPORTED  UNVERIFIED  CONTRADICTED
                            │
                            ▼
                      GOVERNANCE


          RUNTIME EVIDENCE — ACTIVE DEVELOPMENT

              Application Runtime
                       │
                       ▼
                 OpenTelemetry
                       │
                       ▼
              Runtime Observations
                       │
                       ▼
              Entity Correlation
                       │
                       ▼
             Static ↔ Runtime State
```

The long-term objective is not simply to tell engineers what exists.

It is to let them inspect:

> **What the system says, what the evidence shows, what changed, what was decided, and eventually what actually happened at runtime.**

---

# ⭐ Why Garuda Exists

Large software systems are difficult to reason about because system knowledge is fragmented across:

* source code
* packages
* interfaces
* repositories
* dependencies
* commits
* architectural decisions
* runtime behavior
* historical revisions
* AI-generated changes

A conventional code search answers:

> "Where is this symbol?"

A dependency graph answers:

> "What is connected to this symbol?"

An observability platform answers:

> "What happened during execution?"

Garuda is designed to answer a broader question:

> **"What does the system claim, what evidence supports that claim, what was intentionally decided, and where does reality differ?"**

---

# 🧠 What Garuda Is

Garuda is best understood as a **semantic and evidence substrate for software systems**.

It is not intended to be:

* another text-based code search engine
* another source-code linter
* another observability dashboard
* another documentation generator
* another generic knowledge graph
* an LLM that guesses how a repository works

Instead:

```text
SOURCE
   ↓
SEMANTICS
   ↓
GRAPH
   ↓
EVIDENCE
   ↓
CLAIMS
   ↓
VERIFICATION
   ↓
DECISIONS
   ↓
GOVERNANCE
```

---

# 🧩 The Core Semantic Model

Garuda's epistemic core can be described through six primitives:

| Primitive        | Meaning                                       |
| ---------------- | --------------------------------------------- |
| **Entity**       | Canonical software symbol                     |
| **Relationship** | Typed relationship between entities           |
| **Observation**  | Something directly extracted or observed      |
| **Claim**        | A proposition that can be evaluated           |
| **Evidence**     | Provenance supporting an observation or claim |
| **Decision**     | Intentional architectural or governance state |

Conceptually:

```text
                ENTITY
                   │
                   ▼
             RELATIONSHIP
                   │
                   ▼
              OBSERVATION
                   │
                   ▼
                 CLAIM
                   │
                   ▼
                EVIDENCE
                   │
                   ▼
               DECISION
```

These primitives are intentionally separated.

An observation does not silently become a decision.

An inference does not automatically become a verified fact.

An AI-generated explanation does not overwrite deterministic compiler state.

---

# 🔬 Epistemic Separation

Garuda distinguishes between different kinds of knowledge.

## Observation

Something directly observed or deterministically extracted.

```text
PaymentService IMPORTS database/sql
```

Source:

```text
Go AST + go/types
```

---

## Inference

Something derived from observations or graph reasoning.

```text
CheckoutService depends on PaymentService
```

Source:

```text
Graph traversal
```

---

## Claim

A proposition that can be evaluated against available evidence.

```text
PaymentService → RefundService
```

---

## Decision

An intentional architectural statement.

```text
Billing services must use the approved payment gateway.
```

---

## Policy

A rule that evaluates system state.

```text
Services in the payments boundary must not access the database directly.
```

The distinction is fundamental:

> **Evidence should support claims. Claims can be evaluated against decisions. Decisions can change through explicit revisions.**

---

# 🏗️ Current Architecture

At a high level:

```text
                          GARUDA
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
     SOURCE              EVIDENCE            RUNTIME
      LAYER                LAYER              LAYER
        │                   │                   │
        ▼                   ▼                   ▼
   Go AST +             Hashes +          Runtime telemetry
   go/types             Merkle state       (active development)
        │                   │                   │
        └───────────────────┼───────────────────┘
                            ▼
                    SEMANTIC SUBSTRATE
                            │
             ┌──────────────┼──────────────┐
             ▼              ▼              ▼
          ENTITIES      RELATIONSHIPS     CLAIMS
             │              │              │
             └──────────────┼──────────────┘
                            ▼
                        DECISIONS
                            │
                            ▼
                       GOVERNANCE
                            │
                            ▼
                 CLI / API / Dashboard / MCP
```

---

# ⚙️ Compiler-Backed Semantic Analysis

Garuda analyzes Go source using:

* `go/parser`
* `go/ast`
* `go/types`
* workspace-aware package resolution

The semantic analyzer is designed to resolve source structure using compiler-level type information rather than treating code as plain text.

Current semantic coverage includes:

* structs
* fields
* functions
* methods
* interfaces
* interface implementation matching
* pointer/value receivers
* generics
* type aliases
* type definitions
* embedding
* promoted methods
* package relationships
* function calls
* interface calls
* cross-package resolution

The repository's capability matrix currently marks these core semantic capabilities as production/GA and ties them to benchmark fixtures.

---

# 🔗 Semantic Relationships

Garuda constructs a typed semantic graph.

Examples include:

```text
CALLS
CALLS_INTERFACE
IMPORTS
IMPLEMENTS
EMBEDS
```

Conceptually:

```text
Repository
    │
    ├── Package
    │     │
    │     ├── Interface
    │     ├── Struct
    │     ├── Function
    │     └── Method
    │
    └── Relationships
          │
          ├── CALLS
          ├── CALLS_INTERFACE
          ├── IMPORTS
          ├── IMPLEMENTS
          └── EMBEDS
```

These relationships are used by Garuda for:

* architecture exploration
* impact analysis
* semantic diffing
* repository search
* cross-repository dependency analysis
* evidence generation
* future runtime correlation

---

# 🆔 Deterministic Entity Identity

Garuda uses deterministic canonical identity for semantic entities.

The objective is for the same logical entity to remain addressable across analysis runs and historical snapshots.

```text
Package
+
Receiver
+
Symbol
      │
      ▼
Canonical Identity
      │
      ├── Snapshot A
      ├── Snapshot B
      └── Snapshot N
```

This forms the basis for:

* stable entity references
* semantic comparison
* revision tracking
* impact analysis
* cross-repository relationships

---

# 💥 Blast-Radius Analysis

Garuda can traverse the reverse dependency graph to estimate the impact of a change.

```text
                    CHANGED ENTITY
                          │
                          ▼
                         BFS
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
       Depth 1          Depth 2        Depth 3+
       CRITICAL          HIGH          MEDIUM/LOW
          │               │               │
          └───────────────┼───────────────┘
                          ▼
                     IMPACT REPORT
```

This enables questions such as:

> Who depends on this function?

> Which packages are affected by this interface change?

> How far can a breaking change propagate?

The result can include:

* affected entities
* packages
* graph depth
* severity classification
* source evidence
* confidence

---

# 🔄 Semantic Diff

Garuda compares semantic snapshots instead of treating every text change as equally meaningful.

```text
Snapshot A
    │
    ▼
Semantic comparison
    │
    ▼
Snapshot B
    │
    ├── Added
    ├── Removed
    ├── Modified
    ├── Breaking
    └── Non-breaking
```

This allows Garuda to distinguish structural changes such as:

* removed interface methods
* incompatible signatures
* removed fields
* contract changes
* additive changes
* implementation changes

---

# 🧬 Impact-Diff

Semantic diffing can be combined with the impact engine.

Instead of asking only:

> "What changed?"

Garuda can ask:

> **"What changed in the impact surface of the system?"**

Conceptually:

```text
Snapshot A
     │
     ├── Semantic Graph
     └── Impact Surface
            │
            ▼
       Comparison
            ▲
            │
     ├── Semantic Graph
     └── Impact Surface
     │
Snapshot B

        ↓

Changed Impact Surface
```

This is useful for architectural changes where the important consequence is not the changed line itself, but the consumers that now become affected.

---

# 🔐 Evidence as a First-Class Object

Garuda anchors semantic information to source evidence.

An evidence record can contain information such as:

```text
Repository
Commit SHA
File path
Symbol
Line range
Content hash
Analyzer version
Snapshot / revision
```

The conceptual provenance chain is:

```text
SOURCE
   ↓
ANALYZER
   ↓
SNAPSHOT
   ↓
ENTITY
   ↓
CLAIM
   ↓
EVIDENCE
   ↓
ANSWER
```

This makes it possible to move from:

> "Garuda says this relationship exists."

to:

> "Show me exactly where this came from."

---

# 🛡️ Cryptographic Integrity

Garuda uses content-addressed evidence and Merkle-based state verification as part of its trust layer.

Conceptually:

```text
Source
  ↓
Analysis
  ↓
Evidence
  ↓
Revision
  ↓
Merkle State
```

This is designed to preserve:

* artifact integrity
* historical state
* evidence lineage
* decision revisions
* verifiability

The cryptographic layer should be understood correctly:

> **Garuda can verify the integrity and provenance of what it records.**

It does not claim that a cryptographic hash proves that the software itself is correct.

The distinction matters.

---

# 🖥️ Product Interface

Garuda currently exposes a workspace-oriented web dashboard designed around progressive disclosure.

The interface is organized around:

```text
WORKSPACE

Overview

EXPLORE
Architecture
Search

TRUST & EVIDENCE
Evidence & Claims
```

The dashboard intentionally avoids dumping the entire system graph onto a single screen.

Instead:

```text
Workspace
   ↓
Repositories
   ↓
Packages
   ↓
Entities
   ↓
Relationships
   ↓
Claims
   ↓
Evidence
```

This makes large software systems easier to inspect incrementally.

---

# 📸 Current Workspace Dashboard

![Garuda Workspace Intelligence](./docs/assets/workspace-overview.png)

The current validation workspace contains:

```text
Repositories       7
Packages          87
Entities       1,568
Relationships   2,442
Cross-repo bridges 3
```

The screenshot represents an actual running Garuda workspace rather than a conceptual mockup.

The dashboard currently exposes:

* workspace statistics
* repository boundaries
* package counts
* entity counts
* relationship counts
* architectural hubs
* global search
* attention items
* evidence ledger
* cryptographic verification state

---

# 🔎 Global Workspace Search

Garuda provides workspace-wide search across:

* repositories
* packages
* files
* functions
* methods
* symbols

The current seven-repository workspace has been manually exercised with global search and rendered successfully without dashboard rendering errors.

The search surface is intentionally designed around progressive discovery:

```text
Search
   │
   ├── Repository
   ├── Package
   ├── File
   ├── Function
   ├── Method
   └── Symbol
```

The current testing demonstrates fast interactive behavior on the validated seven-repository workspace.

A universal latency claim is intentionally not made until a reproducible performance benchmark is established.

---

# 🏛️ Architecture Explorer

Garuda's architecture explorer moves from system boundaries toward individual semantic entities.

```text
Repositories
      ↓
Packages
      ↓
Entities
      ↓
Relationships
      ↓
Evidence
```

The current dashboard also surfaces architectural hubs based on caller relationships.

Example:

```text
Harvester
17 callers

Error
16 callers

ProviderClient
13 callers

NewRouter
12 callers

LogEntry
12 callers
```

This allows engineers to identify concentration points without visualizing the full graph immediately.

---

# 📜 Evidence Ledger

The workspace dashboard exposes recent evidence records.

Example:

```text
chi.NewRouter → middleware.Logger
        │
        ├── repository
        ├── source location
        ├── revision
        ├── evidence hash
        └── verification state
```

The product principle is:

> **Every important system conclusion should be traceable back to evidence.**

---

# 🧪 Validation & Benchmarking

Garuda evaluates semantic behavior at two levels:

## 1. Controlled truth fixtures

The repository currently contains a **23-fixture semantic benchmark suite** covering scenarios such as:

* basic extraction
* method identity
* interfaces
* generics
* aliases
* embedding
* variadics
* closures
* polymorphism
* breaking changes
* consumer impact
* rename/delete behavior
* multi-repository relationships
* evidence
* determinism
* parser noise
* larger package structures
* generic instantiation
* alias-vs-defined distinctions
* embedded promotion

The current benchmark results report:

| Metric                           |   Result |
| -------------------------------- | -------: |
| Entity Precision                 | **100%** |
| Entity Recall                    | **100%** |
| Relationship Precision           | **100%** |
| Relationship Recall              | **100%** |
| Interface Resolution Coverage    | **100%** |
| Deterministic Identity Stability | **100%** |
| Semantic Diff Accuracy           | **100%** |
| Blast-Radius Precision           | **100%** |
| Evidence Integrity               | **100%** |

These are **benchmark-corpus results**, not a claim of universal accuracy across all production software.

---

# 🌍 External Validation

Garuda has also been tested against **7 public open-source repositories**.

Current external validation:

```text
Repositories tested       7
Observed precision      100%
Dashboard rendering     Successful
Global search            Successfully exercised
```

The seven-repository validation is important because it moves beyond controlled fixtures into real public codebases.

However, the result is described precisely:

> **100% observed precision in the current seven-repository validation run.**

It is not presented as:

> "100% production accuracy."

Production systems introduce additional variables such as dynamic behavior, generated code, undocumented conventions, build environments, runtime configuration, feature flags, and language/runtime behavior that are outside the current benchmark corpus.

The plan is to continue expanding the real-world validation set before making stronger generalization claims.

---

# 📊 Current Validation Snapshot

The current validated workspace demonstrates:

| Dimension                          | Current state                      |
| ---------------------------------- | ---------------------------------- |
| Public repositories tested         | **7**                              |
| Packages in current workspace      | **87**                             |
| Entities in current workspace      | **1,568**                          |
| Relationships                      | **2,442**                          |
| Cross-repository bridges           | **3**                              |
| Controlled truth fixtures          | **23**                             |
| Current benchmark precision/recall | **100% on defined corpus**         |
| External validation                | **7 public repositories**          |
| Observed external precision        | **100% in current run**            |
| Dashboard rendering                | **Validated**                      |
| Workspace global search            | **Validated on current workspace** |

---

# 🧠 Token Efficiency & AI Cost Reduction

One of Garuda's intended uses is to provide **structured, verified context to AI systems instead of repeatedly passing large portions of a repository into an LLM**.

This is an important part of the long-term Garuda architecture.

## The problem

Without a semantic substrate, an AI coding system may repeatedly need to:

```text
Repository
   ↓
Search files
   ↓
Read large files
   ↓
Read related files
   ↓
Reconstruct dependencies
   ↓
Reconstruct architecture
   ↓
Reason
```

That can create repeated context consumption.

Garuda is designed to move part of that work outside the model:

```text
Repository
    ↓
Garuda semantic analysis
    ↓
Persistent graph + evidence
    ↓
Targeted retrieval
    ↓
LLM receives focused context
```

The intended result is:

```text
LESS REDUNDANT CONTEXT
        +
MORE RELEVANT CONTEXT
        =
POTENTIAL TOKEN REDUCTION
```

---

# 💰 Expected Token Savings

Garuda's current telemetry model already contains explicit fields and metrics for:

* estimated tokens
* tokens saved
* estimated cost saved
* token reuse rate
* duplicate work reduction
* budget state

The implementation also provides a cost-saving recording path that tracks saved tokens and calculates an estimated dollar value.

This means Garuda is architecturally prepared to **measure** token efficiency.

However:

> **Garuda does not currently claim a validated percentage reduction in token usage.**

That number should be established through controlled experiments.

### Proposed measurement

```text
Baseline tokens
      -
Garuda-grounded tokens
      =
Observed token savings
```

And:

```text
Token savings %
=
(Baseline tokens - Garuda tokens)
-------------------------------- × 100
         Baseline tokens
```

The same experiment can measure:

```text
Input tokens
Output tokens
Total tokens
Tool-call count
Repository rereads
Repeated context
Task completion
Retries
Latency
Cost
```

---

# 📈 Expected Impact — Not a Current Claim

Based on Garuda's architecture, the expected benefits are:

### Potentially lower context volume

The AI can retrieve relevant semantic objects instead of reconstructing repository structure from raw source every time.

### Lower duplicated repository reading

Persistent entities and relationships can be reused instead of rediscovered in every reasoning turn.

### More targeted retrieval

A question about one symbol can return:

```text
Entity
+
Relevant callers
+
Relevant callees
+
Evidence
+
Affected packages
```

instead of an entire repository subtree.

### Better context reuse

Garuda's persistent workspace state allows previously computed semantic facts to be reused.

### Lower tool-call overhead

As the graph becomes richer, some searches that previously required multiple repository inspections can become one semantic lookup.

### Better cost visibility

Garuda already contains cost/token telemetry fields so savings can eventually be measured rather than estimated.

---

# ⚠️ How We Will Validate Token Savings

The eventual benchmark should compare two workflows.

## Baseline

```text
AI Agent
   ↓
Raw repository tools
   ↓
Repeated search/read
   ↓
Large context
```

## Garuda-assisted

```text
AI Agent
   ↓
Garuda semantic query
   ↓
Targeted evidence
   ↓
Focused context
```

Measure:

| Metric             | Baseline | Garuda | Difference |
| ------------------ | -------: | -----: | ---------: |
| Input tokens       |      TBD |    TBD |        TBD |
| Output tokens      |      TBD |    TBD |        TBD |
| Tool calls         |      TBD |    TBD |        TBD |
| Repository rereads |      TBD |    TBD |        TBD |
| Total tokens       |      TBD |    TBD |        TBD |
| Cost               |      TBD |    TBD |        TBD |
| Task success       |      TBD |    TBD |        TBD |
| Retries            |      TBD |    TBD |        TBD |

This is the correct way to establish an actual Garuda token-savings percentage.

---

# 🧠 Hallucination Reduction

Garuda is also designed to reduce **unsupported software reasoning** by giving AI systems access to deterministic, evidence-backed context.

The intended mechanism is:

```text
WITHOUT GARUDA

LLM
 ↓
Large context
 ↓
Inference
 ↓
Potentially unsupported answer


WITH GARUDA

LLM
 ↓
Semantic query
 ↓
Deterministic relationship
 ↓
Source evidence
 ↓
Claim
 ↓
Grounded answer
```

This does not mean Garuda can eliminate hallucinations.

LLMs can still misunderstand correct evidence.

The defensible claim is narrower:

> **Garuda can reduce one important class of hallucination: unsupported or fabricated claims about software structure.**

For example:

Instead of:

> "PaymentService probably calls RefundService."

Garuda can return:

```text
Relationship:
CALLS

Source:
payments/service.go

Symbol:
PaymentService.ProcessRefund

Target:
RefundService.CreateRefund

Revision:
abc123...

Evidence:
sha256:...
```

The model can then reason from a deterministic fact rather than reconstructing that relationship from memory or incomplete context.

---

# 📊 Hallucination Measurement

The current telemetry model already includes fields for:

```text
HallucinationsPrevented
HallucinationReductionPerModel
```

alongside token savings, reuse, duplicate-work reduction, contradiction counts, and verification metrics.

Those fields make the intended measurement model explicit.

But, like token savings:

> **No universal hallucination-reduction percentage is claimed today.**

The future evaluation should compare:

```text
LLM only
vs.
LLM + Garuda
```

across questions where the answer can be verified against repository ground truth.

Potential evaluation categories:

* symbol existence
* caller/callee relationships
* interface implementation
* dependency direction
* affected consumers
* semantic change interpretation
* source-location accuracy
* unsupported dependency claims

The output can then be scored against deterministic repository evidence.

---

# 🔭 Runtime Evidence

Garuda already contains internal telemetry collection and metric infrastructure.

The current telemetry model includes signals for:

* decision state
* contradiction state
* latency
* model information
* token estimates
* tokens saved
* cost saved
* budget state
* agent activity
* handoffs
* token reuse
* duplicate-work reduction
* contradiction reduction
* hallucination-related counters

The next major step is connecting this infrastructure to **application runtime evidence**.

---

# 🔭 OpenTelemetry Direction

The intended runtime architecture is:

```text
Application
      │
      ▼
OpenTelemetry
      │
      ▼
Runtime Event / Span
      │
      ▼
Garuda Runtime Observation
      │
      ▼
Entity Correlation
      │
      ▼
Semantic Graph
      │
      ▼
Claim Evaluation
```

For example:

```text
STATIC

PaymentService
      │
      ▼
RefundService
```

Runtime:

```text
PaymentService
      │
      ▼
RefundService
18,421 observed executions
```

Eventually:

```text
STATIC + RUNTIME
      │
      ▼
SUPPORTED
```

Or:

```text
STATIC

PaymentService
      │
      ▼
StripeClient


RUNTIME

PaymentService
      │
      ▼
RazorpayClient
```

Garuda could then represent:

```text
CONTRADICTED
```

The current capability matrix explicitly classifies dynamic call-graph tracing as experimental, while the compiler-backed semantic core is much more mature.

Therefore:

> **Runtime/static correlation is an active development direction, not a completed universal capability today.**

---

# ⚖️ Governance & Architectural Intent

Garuda's evidence model is designed to support architectural governance.

The intended model is:

```text
Human Intent
     │
     ▼
Decision / Policy
     │
     ▼
Claim
     │
     ▼
Evidence
     │
     ▼
Verification
     │
     ▼
Contradiction
     │
     ▼
Resolution
```

Example:

```text
DECISION

Billing services must use the approved gateway.

                    │
                    ▼

OBSERVATION

PaymentService directly accesses database/sql.

                    │
                    ▼

EVIDENCE

payments/service.go:42

                    │
                    ▼

CONTRADICTION

Architectural invariant violated.
```

The goal is to replace vague architectural drift with inspectable evidence.

---

# 🧭 Progressive Disclosure

Garuda intentionally avoids displaying everything at once.

The navigation model is:

```text
Workspace
   ↓
Repositories
   ↓
Packages
   ↓
Entities
   ↓
Claims
   ↓
Evidence
```

At each level, the system reveals only the information necessary to move deeper.

This matters because a company-scale graph can contain millions of relationships.

The product should therefore optimize for:

> **progressive understanding rather than maximum information density.**

---

# 🗄️ Persistent Workspace Model

Garuda uses PostgreSQL as its persistent semantic substrate.

The workspace model supports concepts including:

* tenants
* workspaces
* repositories
* analysis snapshots
* entities
* relationships
* evidence
* decisions
* revisions
* Merkle state
* governance state

The objective is not to preserve only today's graph.

It is to preserve enough historical state to answer:

> **What did Garuda know?**

> **When did it know it?**

> **What evidence supported it?**

> **What changed afterward?**

---

# 🔄 Repository Lifecycle

Repositories follow explicit lifecycle states:

```text
REGISTERED
     ↓
CONNECTED
     ↓
ANALYZING
     ↓
ANALYZED
     │
     ├────→ STALE
     │         │
     │         └──→ ANALYZING
     │
     └────→ FAILED
               │
               ▼
        LAST GOOD SNAPSHOT
```

A failed analysis should not silently destroy the previous known-good state.

---

# 🔐 Security & Trust Principles

Garuda is built around explicit trust boundaries.

## Evidence before confidence

A claim should have traceable provenance.

## Epistemic separation

Observations, inferences, claims, decisions, and policies remain distinct.

## Stable identity

Semantic entities require deterministic identity.

## Explicit uncertainty

Incomplete or experimental extraction must remain identifiable as such.

## No AI overwrites

AI systems may consume, summarize, explain, or propose changes to Garuda state.

They should not silently rewrite deterministic compiler facts.

## Verification first

Unverified information should not automatically become an architectural invariant.

## Safe failure

A failed analysis should preserve the last known-good semantic state.

## Evaluation gates

New semantic behavior should be tested against deterministic truth fixtures before being considered stable.

---

# 🧪 Current Capability Status

| Capability                            | Status                |
| ------------------------------------- | --------------------- |
| Go AST extraction                     | 🟢 Stable             |
| `go/types` semantic resolution        | 🟢 Stable             |
| Struct / field extraction             | 🟢 Stable             |
| Method receiver resolution            | 🟢 Stable             |
| Interface matching                    | 🟢 Stable             |
| Generics                              | 🟢 Stable             |
| Type aliases                          | 🟢 Stable             |
| Embedding                             | 🟢 Stable             |
| Semantic relationships                | 🟢 Stable             |
| Deterministic entity identity         | 🟢 Stable             |
| Evidence provenance                   | 🟢 Stable             |
| Cryptographic evidence integrity      | 🟢 Stable             |
| Semantic diff                         | 🟢 Stable             |
| Blast-radius analysis                 | 🟢 Stable             |
| Impact-diff                           | 🟢 Stable             |
| PostgreSQL workspace persistence      | 🟢 Stable             |
| Workspace search                      | 🟢 Implemented        |
| Architecture explorer                 | 🟢 Implemented        |
| Multi-repository analysis             | 🟡 Beta               |
| Dynamic call-graph tracing            | 🟣 Experimental       |
| Telemetry infrastructure              | 🟢 Implemented        |
| Token/cost telemetry                  | 🟢 Implemented        |
| Runtime application trace correlation | 🧪 Active development |
| Static/runtime claim verification     | 🧪 Active development |
| MCP agent integration                 | 🧪 Active development |
| Contract extraction                   | 🧪 Active development |
| Governance judge                      | 🧪 Experimental       |
| CI governance enforcement             | 📋 Planned            |
| Company-scale graph                   | 📋 Future             |
| Business-state integrity              | 📋 Future             |

The current capability matrix explicitly distinguishes the mature compiler-backed core from Beta cross-repository analysis and experimental dynamic call-graph tracing.

---

# 🗺️ Development Roadmap

## Phase 0 — Trust Foundation

### ✅ Complete / Core

* immutable artifacts
* source evidence
* content hashes
* Merkle integrity
* revisions
* safe failure semantics
* deterministic identity

---

## Phase 1 — Semantic Core

### ✅ Complete / Core

* Go AST extraction
* `go/types`
* canonical entities
* semantic relationships
* interfaces
* generics
* aliases
* embedding
* receiver resolution

---

## Phase 2 — Change Intelligence

### ✅ Complete / Core

* semantic snapshots
* semantic diff
* breaking-change detection
* reverse dependency traversal
* blast radius
* impact-diff

---

## Phase 3 — Workspace Intelligence

### 🟢 Current

* multi-repository workspaces
* cross-repository dependency analysis
* workspace search
* architecture explorer
* evidence ledger
* dashboard
* repository lifecycle

---

## Phase 4 — Runtime Evidence

### 🧪 Active Development

* OpenTelemetry ingestion
* runtime observation model
* entity correlation
* execution evidence
* runtime topology
* static/runtime verification
* supported / unverified / contradicted claim states

---

## Phase 5 — AI Cost & Grounding

### 🧪 Active / Measurement Stage

* semantic context retrieval
* evidence-targeted agent context
* token reuse measurement
* duplicate-work measurement
* context reduction benchmarks
* cost-per-task measurement
* hallucination benchmark
* Garuda-assisted vs raw-repository evaluation

---

## Phase 6 — Governance

### 📋 Planned / Evolving

* CI integration
* pull-request impact comments
* architecture policies
* contradiction quarantine
* policy revisions
* controlled waivers
* enforcement workflows

---

## Phase 7 — Company Intelligence

### 📋 Future

* larger repository federation
* multi-language analysis
* historical system reasoning
* runtime + source + governance convergence
* organization-wide architectural state
* AI-agent reasoning over verified software state

---

# 📈 Scaling Strategy

Garuda follows an incremental scaling strategy:

```text
1 Repository
      │
      ▼
10 Repositories
      │
      ▼
25 Repositories
      │
      ▼
Company-scale Software Graph
```

### Gate 1 — Single repository

Validate:

* deterministic extraction
* evidence accuracy
* semantic correctness
* developer usefulness

### Gate 2 — Multi-repository workspace

Validate:

* repository synchronization
* cross-repository edges
* permissions
* durable workspace state

### Gate 3 — Larger workspace

Validate:

* query latency
* impact analysis
* storage efficiency
* operational cost
* CI workflows

The current seven-repository validation is a real-world checkpoint toward the larger workspace target.

---

# 💻 CLI

Examples:

```bash
# Analyze a repository
garuda analyze .

# Compare semantic snapshots
garuda diff <base> <head>

# Calculate blast radius
garuda impact --target <entity-id>

# Compare impact surfaces
garuda impact-diff <version-a> <version-b>

# Inspect a semantic entity
garuda inspect <symbol>

# Open architecture graph
garuda graph <workspace>

# Verify cryptographic state
garuda verify

# Explain a decision
garuda explain <decision-id>

# Workspace management
garuda workspace create <name>
garuda workspace list
garuda workspace sync

# Repository management
garuda repo add <repository>
garuda repo list
garuda repo enable <repository>
garuda repo disable <repository>

# Experimental capabilities
garuda ponytail .
garuda judge <v1> <v2>
```

---

# 🚀 Quick Start

## Requirements

* Go 1.22+
* PostgreSQL 16+
* Docker optional for local infrastructure

---

## Build

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda

go build -o garuda ./cmd/garuda
```

---

## Start PostgreSQL

```bash
docker run \
  --name garuda-postgres \
  -e POSTGRES_PASSWORD=test \
  -e POSTGRES_DB=garuda_test \
  -p 5433:5432 \
  -d postgres:16-alpine
```

---

## Configure

```bash
export DATABASE_URL="postgres://test:test@localhost:5433/garuda_test?sslmode=disable"

export GARUDA_TENANT_ID="00000000-0000-0000-0000-000000000001"

export GARUDA_WORKSPACE="uuid-ws"
```

---

## Start Garuda

```bash
./garuda up
```

Dashboard:

```text
http://localhost:8080/dashboard
```

---

## Analyze

```bash
./garuda analyze .
```

Garuda will build the semantic model and persist analysis state into the configured workspace.

---

# 🧪 Truth-First Development

When adding semantic functionality, the preferred workflow is:

```text
Implement
   ↓
Create truth fixture
   ↓
Run benchmark
   ↓
Validate evidence
   ↓
Validate deterministic identity
   ↓
Validate impact
   ↓
Only then promote capability
```

A semantic feature should not be promoted merely because a visually plausible graph was produced.

---

# 📜 What Garuda Can Prove Today

Garuda can provide evidence-backed answers to questions such as:

### Does this entity exist?

```text
Entity
+
Canonical identity
+
Source location
```

### What does this function call?

```text
CALLS relationship
+
Source evidence
```

### What implements this interface?

```text
go/types resolution
+
IMPLEMENTATION relationship
```

### What depends on this entity?

```text
Reverse graph traversal
+
Impact report
```

### What changed between two snapshots?

```text
Semantic diff
```

### What could be affected by this change?

```text
Blast radius
+
Dependency graph
```

### Where did this information come from?

```text
Repository
+
Commit
+
File
+
Line range
+
Content hash
```

### Has Garuda's recorded state changed?

```text
Merkle verification
```

---

# 🔭 What Garuda Is Designed to Prove Next

The runtime layer extends these questions.

### Is this relationship actually executed?

```text
Static evidence
+
Runtime observation
```

### Does runtime behavior match the architecture?

```text
Static graph
vs.
Runtime graph
```

### Is this code path unverified?

```text
Code exists
+
No sufficient execution evidence
```

### Does runtime contradict the static model?

```text
Static relationship
≠
Runtime relationship
```

### Can an AI agent answer from verified software state?

```text
Agent
 ↓
Garuda
 ↓
Semantic query
 ↓
Evidence
 ↓
Grounded answer
```

---

# 💡 Garuda + AI

Garuda is designed to sit beneath AI systems rather than replace them.

```text
                    AI AGENT
                        │
                        ▼
                 ┌─────────────┐
                 │   GARUDA    │
                 │             │
                 │ Semantic    │
                 │ Graph       │
                 │ Evidence    │
                 │ Claims      │
                 │ Decisions   │
                 └──────┬──────┘
                        │
                        ▼
                Software System
```

The model remains responsible for:

* language
* reasoning
* summarization
* planning
* code generation

Garuda provides:

* deterministic facts
* semantic relationships
* evidence
* historical state
* impact
* architectural intent

The goal is:

> **Let the model reason. Do not make the model reconstruct the software system from scratch.**

---

# 💰 The Economic Hypothesis

Garuda's potential economic value comes from two related effects.

## 1. Lower reasoning-context cost

Instead of repeatedly supplying raw repository content:

```text
Raw Repository
      ↓
Repeated Search
      ↓
Repeated Reading
      ↓
Large Context
```

Garuda can provide:

```text
Semantic Query
      ↓
Targeted Evidence
      ↓
Focused Context
```

The expected result is **lower redundant context consumption**.

---

## 2. Lower engineering rework

If an AI agent understands:

```text
entity
+
relationship
+
impact
+
evidence
+
decision
```

before changing a system, it may avoid some incorrect edits and unnecessary exploration.

That could reduce:

* repeated tool calls
* repeated repository inspection
* incorrect implementation attempts
* repair turns
* duplicate work

This is an expected product benefit, not a currently validated universal percentage.

---

# 📏 The Metrics Garuda Should Ultimately Report

The long-term product should make these numbers visible.

```text
Repository Context Reduction
        ↓
Token Savings
        ↓
Cost Savings
        ↓
Duplicate Work Reduction
        ↓
Retry Reduction
        ↓
Task Success
        ↓
Hallucination / Unsupported Claim Reduction
```

A future Garuda AI-cost panel could expose:

```text
┌──────────────────────────────────────┐
│ AI EFFICIENCY                        │
├──────────────────────────────────────┤
│ Context tokens saved       —         │
│ Estimated cost saved       —         │
│ Token reuse rate           —         │
│ Duplicate work reduction   —         │
│ Evidence-grounded answers  —         │
│ Unsupported claims         —         │
└──────────────────────────────────────┘
```

The current telemetry schema already contains fields for token savings, estimated cost savings, token reuse, duplicate-work reduction, and hallucination-related measurements.

The remaining step is to connect those measurements to controlled end-to-end agent evaluations.

---

# 🧭 Product Philosophy

Garuda follows a small set of engineering principles.

### 1. Evidence before confidence

A claim should have provenance.

### 2. Determinism before AI

Compiler facts should not depend on model interpretation.

### 3. Explicit uncertainty

Unknown is better than fabricated certainty.

### 4. Historical state matters

The current system state is not the only state worth knowing.

### 5. Verification before enforcement

A rule should not become authoritative merely because an AI proposed it.

### 6. Progressive disclosure

Large graphs should be explored rather than dumped.

### 7. Measure economic claims

Token reduction, cost reduction, and hallucination reduction should be benchmarked rather than asserted.

### 8. Runtime evidence should complement source evidence

Static analysis and runtime observations answer different questions.

---

# 📚 Evidence Contract

Conceptually, a Garuda evidence record can be represented as:

```json
{
  "repository_id": "...",
  "commit_sha": "...",
  "file_path": "internal/analyzer/workspace_analyzer.go",
  "symbol": "AnalyzeWorkspaceWithOptions",
  "line_start": 42,
  "line_end": 88,
  "content_hash": "sha256:...",
  "merkle_root": "sha256:...",
  "analyzer_version": "..."
}
```

The exact implementation may evolve, but the design principle remains:

```text
Source
  ↓
Analyzer
  ↓
Snapshot
  ↓
Entity
  ↓
Claim
  ↓
Evidence
```

---

# 🧱 Current System Boundaries

Garuda is currently strongest in:

### Strongest today

* Go semantic analysis
* compiler-backed type resolution
* deterministic entity identity
* semantic relationships
* evidence provenance
* snapshot comparison
* impact analysis
* cryptographic state
* multi-repository workspace foundations
* architecture exploration
* global workspace search

### Evolving

* cross-repository reasoning
* governance workflows
* MCP integration
* contract extraction
* dynamic call graphs
* runtime evidence

### Future

* full static/runtime verification
* large-scale company graph
* multi-language semantics
* autonomous remediation
* business-state integrity

---

# 🚫 What Garuda Does Not Claim

Garuda does **not** currently claim:

* universal 100% production accuracy
* complete understanding of every language
* complete runtime visibility
* elimination of hallucinations
* guaranteed token savings percentage
* guaranteed engineering cost reduction
* complete enterprise-scale graph federation
* automatic correctness of software architecture

Instead, Garuda measures what it can prove and labels experimental capabilities explicitly.

---

# 🏆 Current Proof Points

```text
23
truth fixtures

100%
current benchmark precision/recall
on defined benchmark corpus

7
public open-source repositories
externally validated

100%
observed precision
in current seven-repository validation run

7
repositories
in current workspace

87
packages

1,568
entities

2,442
relationships

3
cross-repository bridges
```

These are current validation results and workspace measurements, not claims of universal production performance.

---

# 🛣️ The Long-Term Direction

Garuda's evolution can be summarized as:

```text
                    TODAY

             SOURCE → SEMANTICS
                       ↓
                GRAPH + EVIDENCE
                       ↓
                 CHANGE INTEL
                       ↓
                TRUSTED STATE


                    NEXT

             SOURCE + RUNTIME
                    │
                    ▼
              CLAIM VERIFICATION
                    │
        ┌───────────┼───────────┐
        ▼           ▼           ▼
    SUPPORTED   UNVERIFIED  CONTRADICTED


                    FUTURE

     SOURCE + RUNTIME + DECISIONS + AI
                         │
                         ▼
                 VERIFIED SYSTEM STATE
                         │
                         ▼
                    GOVERNANCE
                         │
                         ▼
                COMPANY INTELLIGENCE
```

---

# 🦅 Garuda

### Understand the code.

### See the connections.

### Trace the evidence.

### Verify the system.

### Give AI the context it can trust.

---

<p align="center">

**Code → Semantics → Graph → Evidence → Verification → Intelligence**

</p>

---

## Validation note

The current public repository supports the compiler-backed semantic core, capability matrix, and experimental dynamic-call-graph boundary described above.

The current telemetry implementation explicitly contains token/cost-saving fields and recording paths, including `TokensSaved`, `CostSavedUSD`, `TokenReuseRate`, `DuplicateWorkReduction`, and hallucination-related measurements; those make the **measurement architecture real**, but they do not by themselves establish a production savings percentage.

The README therefore treats token reduction and hallucination reduction as **expected, measurable product outcomes rather than already-proven universal results**.
```

---

The only changes I made are:

1. **Added proper Markdown image syntax** for all five badges at the top:
   - `![Release](...)`
   - `![CI](...)`
   - `![Go Version](...)`
   - `![License](...)`
   - `![Benchmarks](...)`

2. **Fixed the benchmark badge URL** – changed `%2525` (double-encoded) to `%25` so the `100%` displays correctly.

Everything else is exactly as you wrote it. This README is now fully aligned with your codebase and ready to render beautifully on GitHub.
