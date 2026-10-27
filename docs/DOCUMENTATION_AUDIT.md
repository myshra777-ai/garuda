# Documentation audit and source-of-truth map

This file explains how the Garuda documentation suite is organized and how public capability claims should be checked against the current implementation. It is a maintenance guide, not a separate feature specification.

## Product contract

Garuda connects a semantic model of supported code, graph-visible change impact, explicit architecture policies, available document/runtime evidence, agent workflow integrations, configured CI gates, and verifiable records of selected governance decisions.

Garuda Talon is the active agent-workflow layer. It compiles a workspace contract, provides pre-edit context, checks changed code through configured hooks/CLI, returns repair guidance, and can supervise a compatible CLI coding-agent command for a bounded number of attempts. Its hooks and instructions are not a universal sandbox or an unbypassable barrier.

## Capability-to-implementation map

| Capability | Implementation / verification source | Public documentation | Claim boundary |
|---|---|---|---|
| Source analysis and workspace graph | `cmd/garuda/analyze.go`, supported analyzers, `internal/impact` | [Capabilities](CAPABILITIES.md), [language coverage](MULTI_LANGUAGE_ROADMAP.md) | State exact language and relationship coverage. Go has compiler/type-aware paths for supported constructs; Python and TypeScript use structural AST parsing plus selected cross-file patterns. |
| Git-diff scope and change-control gate | `cmd/garuda/check_cmd.go`, `internal/diff`, `internal/policy` | [Quick start](QUICKSTART.md), [API/CLI](API.md), [Operations playbook](../PLAYBOOK.md) | `--fail-on-block` makes a `BLOCK` decision exit with code 2. CI can prevent merges only when its job is required by repository branch protection/rulesets. Confirm the workspace/index represents the candidate revision. |
| Talon pre-edit and post-edit workflow | `cmd/garuda/talon_cmd.go`, `internal/talon/talon.go`, generated integration files | [Talon guide](TALON.md), [Walkthrough](WALKTHROUGH.md) | Hooks vary by client. Repair is bounded; final outcome may remain `BLOCK`. Do not claim every agent action is intercepted or every change is repaired to `ALLOW`. |
| YAML architecture policies | `policies/*.yaml`, `internal/policy` | [API/CLI](API.md), [Capabilities](CAPABILITIES.md) | Nine supported predicate types are implemented. A successful evaluation checks only the configured policy bundle, indexed state, supported semantics and evaluated scope. |
| MCP tools and agent continuity | `cmd/garuda-mcp/main.go`, `cmd/garuda-mcp/session_tracking.go`, `scripts/mcp_verify.py` | [Talon guide](TALON.md), [Walkthrough](WALKTHROUGH.md) | Current public server exposes 23 tools plus a contract prompt/resources. The 30/30 verifier result covers reference protocol/integration checks, not the semantic correctness of every answer. |
| Documentation claims and runtime evidence | Documentation ingest/verify CLI, runtime span routes, verification state in the current server | [Capabilities](CAPABILITIES.md), [API/CLI](API.md), [Evidence](../EVIDENCE.md) | Runtime evidence means spans captured, ingested and mapped. Static/document/runtime states are `SUPPORTED`, `UNVERIFIED`, or `CONTRADICTED`; they are scoped evidence labels, not universal truth certificates. |
| Merkle governance records | `internal/merkle`, decision/evaluation persistence, `GET /api/v1/evidence/verify/{id}` | [Specifications](SPECS.md), [Evidence](../EVIDENCE.md) | Inclusion/integrity proofs apply to recorded data and the applicable root/proof. They do not prove graph completeness, semantic correctness, policy wisdom, or total compliance. |
| Dashboard and ADR Decision Lineage | `internal/api/templates/dashboard.html`, `internal/api/static/dashboard.js` | [Walkthrough](WALKTHROUGH.md) | Six core engineering surfaces are complemented by a separate Decision Lineage view. The review UI presents workspace context/saved evaluations; use the CLI/Talon guard for a fresh change-specific evaluation. |
| Owner Control Plane and operations metrics | `internal/api/control_plane*.go`, `cmd/garuda-mcp/session_tracking.go`, migrations 088/089/092 | [Control Plane design](specs/control-plane.md), [metrics](specs/control-plane-metrics.md), [MCP sessions](specs/mcp-sessions.md) | Separate surface under `/_/control`. Requires `GARUDA_CONTROL_TOKEN` and `CONTROL_DATABASE_URL` with its read-only role. MCP session tracking is best-effort; process-local request-error and query-latency metrics reset on restart. |
| Hygiene, semantic snapshot and plan utilities | `cmd/garuda/ponytail.go`, `cmd/garuda/ci.go`, `cmd/garuda/analyze.go`, `cmd/garuda/plan_cli.go`, `cmd/garuda/self_describe.go` | [API/CLI](API.md), [Capabilities](CAPABILITIES.md) | Hygiene findings are advisory candidates, not proof that code is dead. Baseline CI and before/after snapshot diff are distinct from the YAML policy gate. Plan requires the configured API service. |

## Current evidence and what it means

The current evidence reference is [EVIDENCE.md](../EVIDENCE.md):

- **24/24** strict compiler/AST benchmark cases passed.
- **30/30** reference MCP/Talon verifier checks passed.
- **10/10** injected contradictions were detected in the controlled runtime test.
- The **14-repository Go-only corpus** is separate from the **9-repository blended Go/Python/TypeScript workspace**. The latter snapshot records 14,333 entities and 40,956 relationships.

Those results apply to their named benchmark cases, verifier checks, controlled injection test or measured workspace snapshot. They do not establish universal recall, real-production drift-detection rates, or a complete semantic graph.

The GAP-20 token result belongs to a separate task-specific grounding benchmark. The platform readiness report contains local-environment observations such as latency/cadence. Treat both as scoped/historical measurements, not product-wide SLAs.

## Documentation hierarchy

- [README](../README.md): short user-facing problem, workflow and product contract.
- [Quick start](QUICKSTART.md): first setup, Talon context/guard and CLI/CI path.
- [Talon guide](TALON.md): generated contract, hooks, bounded supervisor and enforcement limits.
- [Capabilities](CAPABILITIES.md): implemented feature matrix and boundaries.
- [Positioning](POSITIONING.md): consistent product language and claim discipline.
- [API / CLI](API.md): command details and the current HTTP router inventory.
- [Walkthrough](WALKTHROUGH.md): what the dashboard and integrations actually show.
- [Specifications](SPECS.md): architecture, contracts, schema and limitations.
- [Evidence](../EVIDENCE.md): current validation summary and scope.
- [Operational playbook](../PLAYBOOK.md): deployment and CI runbook.
- [Source metadata](source/product.yaml), [capability metadata](source/capabilities.yaml), [roadmap metadata](source/roadmap.yaml): structured inputs for generated drafts.

## Historical and intentionally incomplete documents

- [Multi-language history](../EVIDENCE_MULTI_LANG.md) is a historical validation snapshot; use the current [multi-language roadmap](MULTI_LANGUAGE_ROADMAP.md) for the current analyzer contract.
- [v1.0 benchmark report](../BENCHMARKS_AND_VERIFICATION_v1.0.0.md) is historical; use the current [EVIDENCE.md](../EVIDENCE.md) for current-release verification evidence.
- [Workspace snapshot ADR](adr/0003-workspace-state-snapshots.md) remains proposed where the v1 content-addressed snapshot membership-proof path is not implemented.
- [Invariants](invariants.md) intentionally distinguishes fully enforced invariants from partial database-level protections. Do not remove those caveats simply to make the product sound stronger.

## Change-control rules for documentation

1. Public feature claims must identify the implementation path and the supporting test/evidence where one exists.
2. Distinguish implemented behavior, benchmark results, reference-workspace snapshots, roadmap items, and hypotheses about customer/business fit.
3. Never conflate a non-zero CLI exit code with repository branch protection, or a recorded Merkle proof with semantic correctness.
4. Do not describe the dashboard workbench or policy catalog as a live simulator unless a server-side implementation and current route exist.
5. Generated documentation and AI-written README drafts are artifacts for review; they must not silently overwrite canonical docs or push directly to `main`.
6. When a feature changes, update the code, test/evidence, relevant canonical docs and structured metadata together. Keep historical reports labelled as historical rather than rewriting their original results.
