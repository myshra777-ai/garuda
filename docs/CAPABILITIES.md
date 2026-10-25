# Garuda capabilities

This document describes the main implemented capability areas and the boundaries users should understand. It is not a certification of production readiness. Feature availability depends on the current checkout, configuration, supported input and integration path.

## At a glance

| Area | What is implemented | Important boundary |
|---|---|---|
| Go analysis | Compiler/type-aware extraction for supported packages, calls, interfaces and embedding | Resolution is limited to code the Go toolchain can load and analyze; dynamic behavior is not fully represented |
| Python analysis | AST-based extraction for modules, classes, functions and selected alias/instance-call patterns | Structural analysis is not equivalent to complete Python runtime resolution |
| TypeScript analysis | Tree-sitter-based extraction for supported syntax, imports, classes and selected call patterns | Coverage depends on parser support and implemented resolution rules |
| Workspace graph | Persisted entities and typed relationships, including supported cross-repository mappings | Graph completeness depends on indexing scope and resolution success |
| Impact analysis | Neighborhood, caller/dependency and transitive blast-radius queries | Results are based on relationships present in the graph |
| Git diff targeting | Maps supported diff hunks to source entities and evaluates change-scoped policy outcomes | Line mapping and entity resolution can be incomplete; inspect results for generated or unusual diffs |
| Architecture policies | YAML-configured predicates for layers, cycles, claims, entities, contradictions and related conditions | Only implemented predicates and configured scopes are evaluated |
| Documentation claims | Ingests supported Markdown/ADR content and checks extracted claims against the model | Claim extraction does not establish that a document is correct or complete |
| Runtime evidence | Ingests supported OpenTelemetry spans and correlates observations with static entities | Traces show observed execution, not every possible production path |
| Garuda Talon | Contract generation, pre-edit briefing, post-edit guard, MCP integration and a CLI-agent supervisor loop | Enforcement depends on installed hooks, client behavior, configured policies and fail-closed behavior |
| MCP | Agent-facing query, impact, policy, evidence and handoff tools, plus prompts/resources | Tool availability and behavior depend on the built server and connected client |
| Dashboard | Workspace briefing, architecture map, change review, policies, evidence and governance views | UI presence alone does not establish that every metric is populated or operationally complete |
| Merkle verification | Hash-based integrity and inclusion-proof verification for persisted records | Proves consistency with the supplied root/proof, not semantic truth or external timestamping |
| Hygiene analysis | Structural candidates for dead ends, orphaned symbols and other cleanup | Findings are candidates for review, not proof that code is safe to delete |

## Change-control workflow

A typical workflow connects these capabilities:

1. Index the repository and inspect unresolved or high-impact relationships.
2. Define architecture constraints in the supported policy format and document relevant decisions.
3. Use `garuda check` or Talon to scope evaluation to a target or Git diff.
4. Review the verdict and the difference between new findings and baseline findings, when available.
5. Persist the decision when needed, then verify the record through the supported Merkle verification path.
6. Run the same checks in CI only after the team has tested expected behavior, failure modes and bypass paths.

## Garuda Talon

Talon includes code paths for:

- Compiling workspace ADR claims, YAML policies and selected high-impact symbols into an agent-facing contract.
- Producing pre-edit context and impact information.
- Re-indexing changed files and evaluating post-edit changes.
- Returning repair-oriented feedback for blocked changes.
- Registering MCP tools and generating client-specific configuration/hooks.
- Running a supervisor loop around a CLI coding agent.

These are distinct from a guarantee that all agent edits are intercepted. Generated rules can be ignored by an agent; hooks can be misconfigured or bypassed; and a wrapper's behavior depends on the command and environment. Teams should verify the exact enforcement boundary in their chosen editor, agent, Git workflow and CI environment.

## Evidence and terminology

| Term | Meaning |
|---|---|
| `SUPPORTED` | Available evidence supports the claim within the checked scope |
| `UNVERIFIED` | Evidence is insufficient; not proof of a defect |
| `CONTRADICTED` | Available evidence conflicts with the claim |
| `ALLOW` | The configured evaluator found no blocking condition in its evaluated scope |
| `WARN` / `REVIEW` | The configured evaluator returned a non-blocking warning or review outcome |
| `BLOCK` | The configured evaluator found a blocking condition under the active rules |

A clean result is only as complete as the input model, supported semantics, active policies and evaluation scope. It is not a general proof of correctness.

## Validation evidence

See [EVIDENCE.md](../EVIDENCE.md) for reported benchmarks, corpus sizes, MCP verification and runtime contradiction tests. Treat each result as scoped to the named test or corpus. Before using a number in external materials, reproduce it from the documented command and preserve the environment, commit and output.

## Related documentation

- [Quick start](QUICKSTART.md)
- [CLI, MCP and HTTP API](API.md)
- [System specifications](SPECS.md)
- [End-to-end walkthrough](WALKTHROUGH.md)
- [Evidence and verification](../EVIDENCE.md)
