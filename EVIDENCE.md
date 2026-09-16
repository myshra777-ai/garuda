```markdown
# 🦅 Garuda Evidence & Verification Summary

Garuda separates product claims from reproducible evidence. This document summarizes all empirical benchmarks, multi-repository validation datasets, and runtime telemetry tests.

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

Workspace: `go-validation-10` (`c71c76f5-5dad-4eb8-b35f-fc9b01e0fff7`)
Measured: **2026-09-15**

| Verification Dimension | Measured Metric | Verification State |
| :--- | :---: | :--- |
| **Tested Codebases** | **9 Repositories** | Go, Python, TypeScript |
| **AST Symbols Indexed** | **14,333 Entities** | Single unified semantic model |
| **Static Claims Indexed** | **40,956 Relationships** | Calls, imports, implements, embeds, references |
| **Cross-Repository Edges** | **2 Edges** | Drawn between two Go modules |
| **Languages in Workspace** | **3** | Go, Python, TypeScript coexisting |

Language coverage in this corpus, with the resolution tier each language operates at:

| Language | Resolution Tier | Method |
| :--- | :--- | :--- |
| Go | Tier 5 — compiler-resolved | Compiler type information |
| Python | Tier 2 — imports resolved, calls heuristic | Structural extraction |
| TypeScript | Tier 2 — imports resolved, calls heuristic | tree-sitter |

Tier 5 means a relationship is derived from compiler type information and is the strongest claim Garuda can make. Tier 2 means the relationship is derived from syntax and pattern matching — imports are reliable, calls are heuristic. Garuda reports the tier with each claim so the reader knows how strong the evidence is.

The two cross-repository edges reflect a real property of this workspace: the Python and TypeScript projects were analyzed but did not have declared inter-repository dependencies to any other module. The corpus tests that a mixed-language workspace produces a coherent semantic model; it does not test cross-language resolution, which is heuristic and is named as such under "What Is Not Measured" below.

### Reproducing Corpus B

```sql
SELECT
  (SELECT COUNT(*) FROM repositories      WHERE workspace_id = '<workspace-id>') AS repos,
  (SELECT COUNT(*) FROM entities          WHERE workspace_id = '<workspace-id>' AND kind != 'external') AS entities,
  (SELECT COUNT(*) FROM claims            WHERE workspace_id = '<workspace-id>') AS claims,
  (SELECT COUNT(*) FROM cross_repo_edges  WHERE workspace_id = '<workspace-id>') AS cross_edges;
```

Expected output on the reference corpus:

```
 repos | entities | claims | cross_edges
-------+----------+--------+-------------
     9 |    14333 |  40956 |           2
```

---

## 1. Controlled Runtime Verification Results

Runtime deviations were injected against the reference workspace on **2026-09-15**, targeting unapproved ports and unauthorized driver access. Each injection was a discrete observation posted to the telemetry ingestion endpoint. The verification engine correlated each observation against the static model.

```
CONTROLLED RUNTIME DRIFT DETECTION
┌────────────────────────────────────────────────────────────┐
│ Injected Observations:        10                           │
│ Successfully Ingested:        10                           │
│ Contradictions Produced:      10                           │
│ Missed / Unquarantined:        0                           │
│ Observed Detection Rate:    100%                           │
└────────────────────────────────────────────────────────────┘
```

The ten injected observations:

| # | Operation | Source |
| :-: | :--- | :--- |
| 1 | `IsOn` | test-harness |
| 2 | `SetIFDEQ` | test-harness |
| 3 | `CurryWith` | test-harness |
| 4 | `AddCallerSkip` | test-harness |
| 5 | `GetFeatureCount` | test-harness |
| 6 | `PFCount` | test-harness |
| 7 | `HScan` | test-harness |
| 8 | `DictObject` | test-harness |
| 9 | `RegisterServiceServerOption` | test-harness |
| 10 | `RecordTransition` | test-harness |

Each observation was correlated to a real entity in the static graph and checked against that entity's outgoing claims. All ten produced a `CONTRADICTED` verification.

### Verification state distribution

| Status | Count |
| :--- | ---: |
| `CONTRADICTED` | 10 |
| `SUPPORTED` | 188 |
| `UNVERIFIED` | 40,653 |

**Scope:** this measures the correlation path — whether a runtime observation that conflicts with the static model is detected and quarantined. It does not measure the rate at which real production systems produce such deviations. The observations are a controlled corpus, not live traffic.

Reproduce with:

```bash
psql "$DATABASE_URL" -c "
  SELECT status, COUNT(*)::int
    FROM claim_verifications
   WHERE workspace_id = '<workspace-id>'
   GROUP BY status;
"
```

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

The MCP server is validated against two independent clients: Cursor and a reference implementation written from scratch in the Garuda repository ([`scripts/mcp_verify.py`](../scripts/mcp_verify.py)).

| Verification Dimension | Result |
| :--- | :--- |
| **Tools exposed** | 16 |
| **Spec-compliance checks** | 18 / 18 |
| **Clients verified against** | Cursor + reference implementation |
| **Reference client dependencies** | Python standard library only |

The reference client exists because vendor compatibility is not the same as specification compliance. Cursor exposed one real bug — the server was responding to JSON-RPC notifications, which the specification forbids. A second implementation catches that class of bug before a user does.

Anyone can run the verification against their own workspace and reproduce the result:

```bash
WORKSPACE=<workspace> python3 scripts/mcp_verify.py
```

Expected output:

```
═══ 18/18 checks passed ═══
```

---

## 4. Empirical Benchmark Reports

- **[GAP-20 Grounding Benchmark — 0% structural hallucination, 87.2% token compression](benchmarks/gap20-grounding.md)**
- **[14-Repository Go Validation Report](reports/platform-readiness.md)**
- **[9-Repository Multi-Language Validation Report](reports/multi-language-validation.md)**
- **[Visual Interface & Screenshot Tour](../docs/WALKTHROUGH.md)**

---

## 5. What Is Not Measured

For every claim above, there is a limit. The limits are stated here so a reader does not infer more than the evidence supports.

- **Real-repository relationship precision is not measured.** The figures in both corpora describe what the analyzer produced. They do not measure what fraction of those relationships are correct against an independent ground truth. The Go analyzer is compiler-backed and therefore high-confidence by construction; the Python and TypeScript analyzers are structural and weaker.

- **The runtime detection rate is measured on injected observations.** Section 1 reports 10/10 on a controlled test. Production telemetry may include observation shapes not present in the test corpus.

- **Cross-language resolution is heuristic.** A Go workspace resolving into Python or TypeScript dependencies is not yet compiler-backed. Corpus B contains two cross-repository edges, both drawn between Go modules. Cross-language edges are not represented in this evidence.

- **Call edges at the leaf are sparse.** Median inbound call count per target is 1. The 90th percentile is 4. 73% of targets have exactly one caller. This is a property of the code under analysis, not a limitation of the analyzer. `garuda impact` reports what the graph contains; the graph is honest about what it does not.

- **Admission latency, cadence, and convergence times are environment-specific.** They describe the tested deployment, not a universal production SLA.

- **The Python and TypeScript analyzers have not been validated against a real cross-language workspace.** Their extraction runs correctly on their own corpora. Their interaction with the Go analyzer across module boundaries has not been measured.

---

## 6. Reproducing These Results

Every number in this document is produced by a command in the Garuda CLI. To reproduce:

```bash
# Analyze a repository and save the result to a workspace
garuda analyze /path/to/repo --save --workspace <workspace-name>

# Re-run the MCP verification against your own workspace
WORKSPACE=<workspace-name> python3 scripts/mcp_verify.py

# Inspect the Merkle ledger height and current roots
garuda status

# Query the verification state distribution
psql "$DATABASE_URL" -c "
  SELECT status, COUNT(*)::int
    FROM claim_verifications
   WHERE workspace_id = '<workspace-id>'
   GROUP BY status;
"
```

The corpus and commits for each report are documented in the linked files under `evidence/`.
```