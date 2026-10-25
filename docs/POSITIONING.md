# Garuda positioning

**Purpose:** Internal reference for product language. Public claims should be consistent with the implementation, tests and current evidence. When in doubt, narrow the claim and link to the relevant validation.

## One-sentence description

**Garuda is architecture-aware change control for AI-written code.**

It builds a semantic model of supported software repositories, evaluates changes against explicit architectural constraints, and exposes evidence-backed outcomes through developer tools, AI-agent integrations and CI workflows.

## The problem

Coding agents can make large changes quickly, but conventional tests and linters do not enforce every architectural constraint. A change may compile and pass tests while violating a dependency rule, creating a package cycle, contradicting an architecture decision or unexpectedly affecting a widely used symbol.

As teams use more agents and parallel workflows, it becomes harder to keep code, documented intent and observed runtime behavior aligned.

## The solution

Garuda connects four kinds of information:

1. **Code structure** — supported entities and relationships extracted from source.
2. **Intent** — architectural claims extracted from supported documentation and policies defined by the team.
3. **Runtime observations** — telemetry that can be correlated with the static model.
4. **Governance decisions** — evaluation outcomes and, when persisted, records that can be cryptographically verified.

Garuda can scope checks to a target or Git diff, estimate impact from the indexed graph, evaluate configured policies, and return a decision for developer review or a configured gate. Garuda Talon brings this workflow into supported AI-agent integrations and can provide repair feedback.

## Product category

Primary category: **architecture-aware change control for AI coding agents**.

Supporting descriptions:
- Software architecture intelligence
- Policy-based change checks
- Agent-integrated architectural guardrails
- Evidence-backed software governance

Use “agent harness” when describing Talon's execution and integration layer. Do not use “autonomous coding agent” to describe Garuda as a whole: Garuda does not provide its own code-generating model.

## What Garuda is

- A semantic model of supported source-code structures and relationships.
- A way to compare selected architectural intent against an indexed code model.
- A policy evaluator for implemented predicates and configured scopes.
- An integration layer for CLI, MCP, dashboard and selected agent/Git workflows.
- A way to preserve and verify the integrity of selected governance records.

## What Garuda is not

- A guarantee that every relationship or runtime behavior is discovered.
- A replacement for tests, code review, threat modeling or engineering judgment.
- A coding model that independently writes software.
- A general-purpose observability platform.
- A proof that a policy is correct merely because the evaluator passed it.
- A compliance certification or external timestamp authority.
- A guarantee that every AI agent can be forced to obey instructions or call MCP tools.

## Why the combination matters

The product value comes from connecting analysis, intent, change scope, policy evaluation and recorded evidence in one workflow. A graph alone helps people explore relationships; Garuda aims to use those relationships to evaluate changes against explicit constraints.

The cryptographic layer is a separate property: it can help verify that a persisted record is consistent with a supplied Merkle root and proof. It does not make the analysis semantically complete or make the underlying decision correct.

## Primary customer

Initial target: engineering teams that rely heavily on AI coding agents and need better visibility into architectural impact and policy violations as code changes.

Likely stakeholders:
- CTO, VP Engineering or Head of Engineering
- Platform and developer productivity teams
- Staff/principal engineers responsible for architecture
- Teams building agent-heavy development workflows

The first user is likely an engineer who needs to answer: “What could this change affect, which rules apply, and why did the check pass or block?”

## Recommended message

**Headline:** Architecture-aware change control for AI-written code.

**Supporting line:** Garuda maps software dependencies, checks changes against documented architecture and explicit policies, and gives developers and coding agents evidence-backed feedback.

**Short version:** Let agents write code. Keep architectural constraints visible and checkable.

## Claim discipline

Prefer:
- “Checks supported architectural constraints against the indexed model.”
- “Estimates change impact from indexed relationships.”
- “Evaluates configured policies within the selected scope.”
- “Correlates ingested runtime traces with static entities.”
- “Verifies persisted records against a supplied Merkle proof and root.”
- “Talon integrates checks into configured agent workflows.”

Avoid unless independently demonstrated for the exact scope:
- “100% accurate” or “complete graph”
- “Zero hallucinations” or “zero blind edits”
- “Forces every agent to comply”
- “Blocks every architectural violation”
- “Proves the code is correct”
- “Independently proves when a decision happened” without an external trusted timestamp/anchor
- “Production-ready” as a blanket claim across every subsystem

## How to describe benchmark results

Always name the benchmark suite or corpus, commit, command and date. Separate Go compiler-backed results from structural Python/TypeScript results. Report false positives and false negatives when measured. A passing test suite demonstrates the tested cases; it does not establish universal accuracy.

## Documentation hierarchy

- [README](../README.md): concise user-facing overview and entry points.
- [Capabilities](CAPABILITIES.md): implemented surfaces and limits.
- [Specifications](SPECS.md): architecture and interface contract.
- [API reference](API.md): commands, config, MCP and HTTP.
- [Evidence](../EVIDENCE.md): measured results and methodology.
- [Security policy](../SECURITY.md): responsible disclosure and security boundaries.
