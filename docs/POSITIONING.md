# Garuda positioning

**Purpose:** Keep product language clear, useful to developers and consistent with the implementation. External claims should link to code or current measured evidence where possible.

## One-sentence description

**Garuda is architecture-aware change control for AI-written software, with an active agent harness called Talon.**

It builds a semantic model of supported repositories, provides agent-aware context, checks selected edits against architectural constraints, and makes the result available through a developer workflow and configured CI gate.

## The problem in a sentence

An AI-generated diff can compile and pass tests while still breaking an architectural boundary, adding a dependency cycle, contradicting an ADR, or changing a high-impact symbol used across the codebase. The visible diff does not always show the full consequence.

## The solution

Garuda connects five things that are often separated:

1. **Code structure:** supported entities and relationships extracted from source.
2. **Impact:** graph-visible callers, dependencies and affected symbols.
3. **Intent:** YAML policies and supported claims extracted from Markdown/ADRs.
4. **Observed behavior:** OpenTelemetry spans that can be correlated with indexed entities.
5. **Governance:** outcomes that can fail a configured gate and, when persisted, carry a verifiable Merkle inclusion record.

### Two cooperating layers

- **Garuda's semantic and governance engine** indexes supported code structures, evaluates policies, correlates available evidence, and records selected decisions.
- **Garuda Talon** brings those checks into supported agent workflows: compile the active architecture contract, provide pre-edit context, check edits incrementally, generate repair instructions, and optionally supervise a CLI coding agent for a bounded number of attempts.

A graph alone describes relationships. Talon and the change-control CLI aim to make those relationships actionable during development and review.

## Product category

Primary category: **architecture-aware change control for AI coding agents**.

Supporting descriptions:
- Active agent guardrails and execution harness
- Software architecture intelligence
- Diff-scoped policy evaluation
- Evidence-backed software governance

Garuda does not supply the code-generating model. Talon wraps supported coding-agent commands and installs configured integration files; it is not a universal sandbox or a guarantee that every agent process can be intercepted.

## Who it is for

The initial user is an engineering team using coding agents on a codebase with architecture boundaries that matter. The first user may be a staff/principal engineer, platform engineer, developer productivity lead, or engineer who owns a sensitive subsystem.

The buyer may be a CTO/VP Engineering or a platform organization responsible for keeping architecture rules consistent as code generation scales. This is a target customer hypothesis, not evidence of customer validation.

The core user question is:

> Before this agent change moves forward, what does it affect, which constraints apply, and why did the check pass or block?

## Message hierarchy

**Headline:** Let agents write code. Keep architecture in the loop.

**Supporting line:** Garuda gives coding agents a model of supported software relationships, checks edits against explicit architecture policies, and can fail a configured CI gate when the result is `BLOCK`.

**Proof points:**
- Diff-to-symbol mapping and graph-visible blast radius.
- Nine supported YAML policy predicate types.
- Talon pre-edit briefing, post-edit guard and bounded CLI-agent supervisor loop.
- 23 MCP tools plus contract prompt/resources and active session contract injection.
- Static/document/runtime evidence states and Merkle verification for persisted records.

## What Garuda is not

- Not a guarantee that every dependency or dynamic runtime behavior is discovered.
- Not a replacement for unit/integration tests, code review, threat modeling or engineering judgment.
- Not a code-generating model.
- Not a general-purpose observability platform.
- Not proof that a policy is correct just because it evaluates successfully.
- Not a compliance certification or external timestamp authority.
- Not a guarantee that an agent instruction or MCP tool cannot be ignored or bypassed.

## Claim discipline

Prefer:
- “Maps supported relationships and estimates graph-visible change impact.”
- “Checks the active Git diff against configured architecture policies.”
- “Separates net-new findings from baseline findings where diff-scoped evaluation applies.”
- “Can fail the CI job when a blocking outcome is returned.”
- “Talon adds pre-edit context and post-edit checks at configured integration points.”
- “Correlates the runtime spans that have been captured, ingested and mapped.”
- “Verifies integrity and inclusion of a persisted record against a supplied Merkle root and proof.”

Avoid unless independently demonstrated for the exact scope:
- “100% accurate”, “complete graph”, “zero hallucinations”, or “every dependency found”.
- “Forces every agent to comply” or “blocks every possible bad merge”.
- “Proves the code is correct” or “proves total compliance”.
- “Zero production runtime misses”.
- “Production-ready” as a blanket claim across every subsystem.
- Token savings or performance numbers without the benchmark, commit, command, environment and measurement scope.

## How to report validation

Use [EVIDENCE.md](../EVIDENCE.md) as the current reference. Name the test/corpus, commit, command and date; separate the Go-only corpus from the blended multi-language workspace; distinguish static confidence values from measured precision/recall; and state what the test did not measure.

## Documentation hierarchy

- [README](../README.md): user-facing product overview.
- [Quick start](QUICKSTART.md): first setup and check.
- [Talon guide](TALON.md): active agent workflow and operational boundaries.
- [Capabilities](CAPABILITIES.md): implementation and limits.
- [Specifications](SPECS.md): architecture and interfaces.
- [API reference](API.md): commands, config, MCP and HTTP.
- [Evidence](../EVIDENCE.md): reproducible measurements and scope.
- [Security policy](../SECURITY.md): disclosure and security boundaries.
