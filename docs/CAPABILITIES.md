# Garuda capabilities

Garuda connects source-code structure, architecture intent, change scope, runtime observations and governance records. The highest-value workflow is not simply viewing a graph: it is using the graph and explicit rules to guide an agent edit, check the resulting diff, and fail a configured CI job when a blocking condition is found.

This is a current capability guide—not a certification of production readiness. Feature behavior depends on the current binary, database migrations, workspace state, supported syntax, configured rules and integration path.

## At a glance

| Capability | What is implemented | Boundary to keep in mind |
|---|---|---|
| Go semantic analysis | Compiler/type-aware extraction for supported packages, calls, interfaces, embeddings and multi-module workspaces | Resolution depends on what the Go toolchain can load; dynamic behavior and unsupported build configurations may be absent |
| Python analysis | AST-based modules, classes, functions, imports and selected cross-file alias/instance-call patterns | Not a full Python type checker; dynamic dispatch and unresolved references remain uncertain |
| TypeScript analysis | Tree-sitter extraction, imports, classes and selected cross-file alias/instance-call patterns | Not a replacement for the TypeScript compiler or language server |
| Workspace graph | Persisted entities and typed relationships with supported cross-repository mappings | Completeness depends on indexed repositories, parser support and resolution success |
| Blast-radius analysis | Caller/dependency neighborhoods, interface relationships and confidence-weighted multi-hop impact | Impact is only as complete as the indexed relationships and chosen scope |
| Incremental indexing | Re-index dirty files or selected files to refresh working-tree state before a check | Validate the repository root, files and workspace match the candidate change |
| Diff-scoped change control | Maps Git diff hunks to touched AST entities and evaluates changed scope | Unusual/generated diffs or unresolved source spans can yield incomplete target mapping; inspect touched files/symbols |
| Baseline separation | Distinguishes net-new violations in the active diff from pre-existing findings where diff-scoped filtering applies | Existing violations remain visible; this is not a blanket waiver for legacy debt |
| YAML policy engine | Nine predicates for layers, transitive paths, cycles, blast radius, claims, entities, contradictions, missing verification and language scope | Only implemented predicates, supported graph semantics and configured scope are evaluated |
| Documentation claims | Extracts supported claim shapes from Markdown/ADRs and checks them against the indexed model | Deterministic extraction is deliberately narrower than natural-language understanding; a document can still be wrong |
| Runtime evidence | Ingests supported OpenTelemetry spans and correlates observations with static entities | Runtime evidence reflects telemetry actually captured, ingested and mapped |
| Garuda Talon | Compiles an agent-facing contract; provides pre-edit context, post-edit guard, MCP calls and a bounded CLI-agent supervisor/self-repair loop | Hooks are client-specific and can be misconfigured or bypassed; validate failure semantics and the exact enforcement boundary |
| MCP surface | 23 tools plus an architectural-contract prompt and workspace contract/guardrail resources | Tool schemas are defined by the live server; a protocol check does not establish semantic accuracy of every tool |
| Multi-agent continuity | Handoff, resume and checkpoint paths for structured task/state transfer | Not a general-purpose multi-agent planner or model runtime |
| Governance and Merkle verification | Records selected evaluations/decisions and verifies integrity/inclusion against a supplied Merkle root/proof | Does not prove source-analysis completeness, policy wisdom, external timestamping or software correctness |
| Dashboard and IDE | Six dashboard workspaces and a VS Code extension for evidence, graph, diagnostics and selected symbol workflows | UI presence is not proof that each panel is populated or that every path is continuously refreshed |
| Owner Control Plane and MCP operations | Separate `/_/control` surface with Business, Tenants, Operations and Internal views; session/tool-call records, stalled sessions, tool error rates, rolling in-process request error rate, query-latency percentiles and recent errors | Requires `GARUDA_CONTROL_TOKEN` and `CONTROL_DATABASE_URL` with the read-only role; session tracking is best-effort and in-process metrics reset on restart |
| Hygiene report | Advisory static unreferenced candidates, duplicate symbol-name candidates and naming-based standard-library suggestions | No finding is proof of dead code or safe deletion; the report does not mutate code or persist a finding lifecycle |
| Snapshot/plan utilities | Semantic snapshot comparison, baseline CI assessment, and structured plan/self-description endpoints or reports | These utilities have their own input and server prerequisites; use the API reference for exact flags |

## Garuda's differentiator: connect the loop

### 1. Before the edit
Talon can compile active ADR claims, YAML policies and high-impact symbols into a workspace contract. A pre-edit briefing gives the agent or developer relevant rules and graph-visible impact before a change is made.

### 2. After the edit
Talon can re-index dirty files, target the Git diff, check net-new violations, verify relevant ADR claims, and return a structured verdict plus repair-oriented instructions. Its supervisor command can rerun a CLI coding agent with that feedback for a configured number of attempts.

### 3. Before the merge
`garuda check` or a policy evaluation can be wired into CI. With `--fail-on-block`, a `BLOCK` verdict exits with code `2`; the CI provider can fail the job and a required branch-protection check can stop the merge. Garuda does not configure repository branch protection on its own.

### 4. After the decision
When a decision is persisted, the Merkle implementation can supply an inclusion proof. That receipt lets reviewers verify the integrity and inclusion of recorded decision data; it does not make the original analysis complete by itself.

## Evidence terminology

| Term | Meaning |
|---|---|
| `SUPPORTED` | Available evidence supports a claim within the checked scope |
| `UNVERIFIED` | Evidence is insufficient; not proof of a defect |
| `CONTRADICTED` | Available evidence conflicts with a claim |
| `ALLOW` | The configured evaluator found no blocking condition in the evaluated scope |
| `WARN` / `REVIEW` | A non-blocking warning or review outcome |
| `BLOCK` | An active policy/evidence condition produced a blocking outcome |

A passing result applies to the scope actually evaluated. Confirm index freshness, touched-symbol mapping, policy files, telemetry availability and integration failure behavior before using it as a mandatory merge gate.

## Current validation record

The current [evidence report](../EVIDENCE.md) records:

- 24/24 strict compiler/AST benchmark cases passed.
- 30/30 reference MCP/Talon verifier checks passed.
- 10/10 injected contradictions detected in the controlled runtime test.
- A 14-repository Go-only validation corpus and a separate 9-repository mixed-language workspace.

The 14-repository Go corpus and 9-repository mixed-language workspace are separate datasets. The 14,333-entity / 40,956-relationship figure belongs to the documented mixed-language workspace snapshot, not the 14-repository Go-only corpus. These counts are snapshot facts, not universal product limits or complete-graph accuracy scores.

## Related documentation

- [Quick start](QUICKSTART.md)
- [Talon guide](TALON.md)
- [CLI, MCP and HTTP API](API.md)
- [System specifications](SPECS.md)
- [End-to-end walkthrough](WALKTHROUGH.md)
- [Multi-language coverage](MULTI_LANGUAGE_ROADMAP.md)
- [Evidence and verification](../EVIDENCE.md)
