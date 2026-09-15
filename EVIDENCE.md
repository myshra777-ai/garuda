```markdown
# 🦅 Garuda Evidence & Verification Summary

Garuda separates product claims from reproducible evidence. This document summarizes all empirical benchmarks, multi-repository validation datasets, and runtime telemetry stress tests.

Every number below is measured. Where a number is a benchmark result rather than an absolute guarantee, it is labelled as such. Where a claim covers a subset of what the tool can do, the subset is named.

Two independent corpora are documented here. They are not the same test at different points in time — they are two distinct evaluations, each with its own purpose and its own scope. Both are still valid. Both are reported separately because they measure different things.

---

## Corpus A — Go-only validation (14 repositories)

Purpose: measure the analyzer's behaviour when every language in the corpus is compiler-backed. This is the strongest evidence Garuda can produce for the Go analyzer, because there is no heuristic component anywhere in the pipeline.

| Verification Dimension | Measured Metric | Verification State |
| :--- | :---: | :--- |
| **Tested Codebases** | **14 Repositories** | Go-only, multi-module workspaces |
| **AST Symbols Indexed** | **4,888 Entities** | Deterministic identity, stable across analysis runs |
| **Static Claims Indexed** | **5,894 Relationships** | Calls, imports, implements, embeds, references |
| **Cross-Module Bridges** | **55 Bridges** | Inter-repository edges |
| **Controlled Contradictions** | **10 / 10 Detected** | 100% quarantine rate |
| **Merkle Ledger Height** | **Block #898+** | Cryptographically verified |

Corpus progression across the Go-only milestones:

```
Repositories:     5 Repos  ──▶ 7 Repos  ──▶ 10 Repos ──▶ 14 Repos
Entities:         1,420    ──▶ 2,181    ──▶ 3,333    ──▶ 4,888 Entities
Relationships:    1,890    ──▶ 2,442    ──▶ 4,510    ──▶ 5,894 Claims
Cross-Repo:       3        ──▶ 12       ──▶ 28       ──▶ 55 Bridges
```

Every relationship in this corpus was derived from compiler type information. There is no heuristic component. The entity counts, claim counts, and bridge counts are the analyzer's output on those commits, reproducible by running `garuda analyze` against the same sources.

---

## Corpus B — Blended multi-language validation (9 repositories)

Purpose: measure the analyzer's behaviour when the workspace contains Go, Python, and TypeScript side by side. This corpus validates the claim that Garuda produces one unified semantic model across languages, and it measures how the pipeline behaves when different languages coexist in the same workspace.

| Verification Dimension | Measured Metric | Verification State |
| :--- | :---: | :--- |
| **Tested Codebases** | **9 Repositories** | Go, Python, TypeScript |
| **AST Symbols Indexed** | **14,333 Entities** | Single unified semantic model |
| **Static Claims Indexed** | **37,511 Relationships** | Calls, imports, implements, embeds, references |
| **Cross-Repository Edges** | **1 Edge** | Drawn between two Go modules |
| **Languages in Workspace** | **3** | Go, Python, TypeScript coexisting |

Language coverage in this corpus, with the resolution tier each language operates at:

| Language | Resolution Tier | Method |
| :--- | :--- | :--- |
| Go | Tier 5 — compiler-resolved | Compiler type information |
| Python | Tier 2 — imports resolved, calls heuristic | Structural extraction |
| TypeScript | Tier 2 — imports resolved, calls heuristic | tree-sitter |

Tier 5 means a relationship is derived from compiler type information and is the strongest claim Garuda can make. Tier 2 means the relationship is derived from syntax and pattern matching — imports are reliable, calls are heuristic. Garuda reports the tier with each claim so the reader knows how strong the evidence is.

The single cross-repository edge reflects a real property of this workspace: the Python and TypeScript projects were analyzed but did not have declared inter-repository dependencies to any other module. The corpus tests that a mixed-language workspace produces a coherent semantic model; it does not test cross-language resolution, which is heuristic and is named as such under "What Is Not Measured" below.

---

## 1. Controlled Runtime Verification Results

```
CONTROLLED RUNTIME DRIFT DETECTION
┌────────────────────────────────────────────────────────┐
│ Injected Observations:        10                       │
│ Successfully Ingested:        10                       │
│ Contradictions Detected:      10                       │
│ Missed / Unquarantined:        0                       │
│ Observed Detection Rate:    100%                       │
└────────────────────────────────────────────────────────┘
```

Method: runtime deviations were injected against a controlled workspace, targeting unapproved ports and unauthorized driver access. Each injection was a discrete observation posted to the telemetry ingestion endpoint. The verification engine correlated each observation against the static model and produced a contradiction in every case.

Scope: this measures the correlation path — whether a runtime observation that conflicts with the static model is detected. It does not measure the rate at which real production systems produce such deviations, nor does it claim to detect every class of runtime drift that a real system might exhibit.

---

## 2. Telemetry Pipeline Characteristics

| Metric Dimension | Observed Characteristic | Verification Protocol |
| :--- | :--- | :--- |
| **Ingestion Endpoint** | `POST /api/v1/telemetry/spans` | OpenTelemetry OTLP span mapping |
| **Admission Latency** | `~1.8 ms` (p95) | HTTP 202 Accepted acknowledgment |
| **Verification Engine** | Asynchronous worker | Evaluates static vs runtime parity |
| **Merkle Commit Cadence** | `10 seconds` per epoch | Recomputes static and runtime roots |
| **UI / IDE Convergence** | `~1–2 minutes` | Propagates `ARCH_DRIFT_001` markers |

These figures describe the tested environment: a local Linux x86_64 deployment, PostgreSQL 16 on NVMe storage, single-node worker, no external network hop between the ingestion endpoint and the database. Production deployments with different topology, network latency, and worker concurrency will see different numbers.

---

## 3. MCP Server Verification

The MCP server is validated against two independent clients: Cursor and a reference implementation written from scratch in the Garuda repository ([`scripts/mcp_verify.py`](scripts/mcp_verify.py)).

| Verification Dimension | Result |
| :--- | :--- |
| **Tools exposed** | 16 |
| **Spec-compliance checks** | 18 / 18 |
| **Clients verified against** | Cursor + reference implementation |
| **Reference client dependencies** | Python standard library only |

The reference client exists because vendor compatibility is not the same as specification compliance. Cursor exposed one real bug — the server was responding to JSON-RPC notifications, which the specification forbids. A second implementation catches that class of bug before a user does.

The verification script is included in the repository. Anyone can run it against their own workspace and reproduce the result:

```bash
WORKSPACE=<workspace> python3 scripts/mcp_verify.py
```

---

## 4. Empirical Benchmark Reports

- **[GAP-20 Grounding Benchmark — 0% structural hallucination, 87.2% token compression](evidence/benchmarks/gap20-grounding.md)**
- **[14-Repository Go Validation Report](evidence/reports/platform-readiness.md)**
- **[9-Repository Multi-Language Validation Report](evidence/reports/multi-language-validation.md)**
- **[Visual Interface & Screenshot Tour](docs/WALKTHROUGH.md)**

---

## 5. What Is Not Measured

For every claim above, there is a limit. The limits are stated here so a reader does not infer more than the evidence supports.

- **Real-repository relationship precision is not measured.** The figures in both corpora describe what the analyzer produced. They do not measure what fraction of those relationships are correct against an independent ground truth. The Go analyzer is compiler-backed and therefore high-confidence by construction; the Python and TypeScript analyzers are structural and weaker.
- **The runtime detection rate is measured on injected observations.** Section 1 reports 10/10 on a controlled test. Production telemetry may include observation shapes not present in the test corpus.
- **Cross-language resolution is heuristic.** A Go workspace resolving into Python or TypeScript dependencies is not yet compiler-backed. Corpus B contains one cross-repository edge, drawn between two Go modules. Cross-language edges are not represented in this evidence.
- **Call edges at the leaf are sparse.** Median inbound call count per target is 1. The 90th percentile is 4. 73% of targets have exactly one caller. This is a property of the code under analysis, not a limitation of the analyzer. `garuda impact` reports what the graph contains; the graph is honest about what it does not.
- **Admission latency, cadence, and convergence times are environment-specific.** They describe the tested deployment, not a universal production SLA.

---

## Reproducing These Results

Every number in this document is produced by a command in the Garuda CLI. To reproduce:

```bash
# Analyze a repository and save the result to a workspace
garuda analyze /path/to/repo --save --workspace <name>

# Re-run the MCP verification against your own workspace
WORKSPACE=<name> python3 scripts/mcp_verify.py

# Inspect the Merkle ledger height and current roots
garuda status
```

The corpus and commits for each report are documented in the linked files under `evidence/`.
```