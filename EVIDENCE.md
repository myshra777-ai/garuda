# Garuda Evidence & Verification Summary

Garuda separates product claims from reproducible evidence. This document records empirical benchmarks, multi-repository validation datasets, runtime verification tests, MCP compatibility checks, workspace demonstrations, and known measurement boundaries.

Every number below is a measured result, a snapshot value, or an explicitly labelled benchmark observation. A benchmark result is not presented as a universal guarantee. Where a claim covers only a subset of Garuda's capabilities, that subset is named.

> **Current-state note:** Historical evidence remains useful when it explains how a capability was validated. Current product claims should use the latest reproducible result and should not silently reuse superseded tool counts, client counts, or check totals.

---

## Evidence at a Glance

| Area | Current evidence |
|---|---|
| Reference MCP verifier | `38/38` checks passed |
| MCP tools | 21 tools exposed |
| MCP clients tested | Cursor, Claude Desktop, Codex CLI, and dependency-free reference verifier |
| Go-only validation | 14 repositories |
| Blended validation | 9 repositories across Go, Python, and TypeScript |
| Blended workspace snapshot | 14,333 entities and 40,956 relationships |
| Controlled runtime contradiction test | 10/10 injected contradictions detected |
| Workspace Console | Workspace, Agents, Governance, Decisions, and Graph surfaces exercised |
| VS Code integration | Ledger, contradictions, diagnostics, graph, refresh, reanalysis, and Go hover workflows exercised |

---

## Evidence Model

Garuda records evidence in several layers:

```text
Source code and compiler/type information
                  ↓
Semantic entities and typed relationships
                  ↓
Documentation claims and policy predicates
                  ↓
Runtime observations and contradictions
                  ↓
Governance decisions and Merkle proofs
```

The evidence state is explicit:

| State | Meaning |
|---|---|
| `SUPPORTED` | Available evidence supports the claim within the verified scope |
| `UNVERIFIED` | Evidence is insufficient; this does not mean the claim is false |
| `CONTRADICTED` | Available evidence conflicts with the claim |

---

## Corpus A — Go-Only Validation

**Scope:** 14 repositories
**Purpose:** Measure analyzer behavior when every language in the corpus is compiler-backed Go. This is the strongest evidence set for supported Go relationship classes because the corpus contains no Python or TypeScript heuristic component.

| Verification dimension | Measured metric | Verification state |
|---|---:|---|
| Tested codebases | **14 repositories** | Go-only, multi-module workspaces |
| AST symbols indexed | **4,888 entities** | Deterministic identity within the measured corpus |
| Static claims indexed | **5,894 relationships** | Calls, imports, implementations, embeddings, and references |
| Cross-module bridges | **55 bridges** | Inter-repository edges observed in the corpus |
| Controlled contradictions | **10 / 10 detected** | 100% observed quarantine rate in the controlled test |
| Merkle ledger | **Block #898+** | Cryptographically verified in the reported run |

### Corpus progression

```text
Repositories:   5 repos ──▶ 7 repos ──▶ 10 repos ──▶ 14 repos
Entities:       1,420    ──▶ 2,181    ──▶ 3,333     ──▶ 4,888
Relationships:  1,890    ──▶ 2,442    ──▶ 4,510     ──▶ 5,894
Cross-repo:     3        ──▶ 12       ──▶ 28        ──▶ 55
```

These values describe analyzer output on the named validation sources. They are not universal product limits and do not independently prove real-repository relationship precision.

### Interpretation

- Go relationships in this corpus were derived through compiler/type information within the supported analyzer scope.
- Cross-repository bridge counts represent resolved edges observed in the tested workspaces.
- The corpus supports claims about extraction and persistence behavior; it does not prove that every possible dynamic runtime relationship is discoverable.

---

## Corpus B — Blended Multi-Language Validation

**Workspace:** `go-validation-10`
**Workspace ID:** `c71c76f5-5dad-4eb8-b35f-fc9b01e0fff7`
**Measured:** 2026-09-15
**Scope:** 9 repositories across Go, Python, and TypeScript.

| Verification dimension | Measured metric | Verification state |
|---|---:|---|
| Tested codebases | **9 repositories** | Go, Python, and TypeScript |
| AST symbols indexed | **14,333 entities** | Single unified semantic model |
| Static claims indexed | **40,956 relationships** | Calls, imports, implementations, embeddings, and references |
| Cross-repository edges | **2 edges** | Both observed between Go modules |
| Languages in workspace | **3** | Coexisting analyzers and persisted graph state |

### Resolution tiers

| Language | Resolution tier | Method |
|---|---|---|
| Go | Tier 5 — compiler-resolved | Compiler/type information |
| Python | Tier 2 — structural | Imports are stronger; dynamic calls are heuristic |
| TypeScript | Tier 2 — structural | Structural/tree-sitter analysis |

Tier 5 is the strongest supported static resolution tier in this evidence set. Tier 2 results are useful structural evidence but should not be interpreted as equivalent to compiler-resolved relationships.

The two cross-repository edges reflect the measured workspace: the Python and TypeScript repositories were analyzed, but no declared cross-repository dependency to another module was present for those projects. This corpus validates a unified multi-language workspace model; it does not validate compiler-backed cross-language resolution.

### Reproduce Corpus B counts

```sql
SELECT
  (SELECT COUNT(*) FROM repositories     WHERE workspace_id = '<workspace-id>') AS repos,
  (SELECT COUNT(*) FROM entities         WHERE workspace_id = '<workspace-id>' AND kind != 'external') AS entities,
  (SELECT COUNT(*) FROM claims           WHERE workspace_id = '<workspace-id>') AS claims,
  (SELECT COUNT(*) FROM cross_repo_edges WHERE workspace_id = '<workspace-id>') AS cross_edges;
```

Expected reference output:

```text
 repos | entities | claims | cross_edges
-------+----------+--------+------------
     9 |    14333 |  40956 |           2
```

---

## Controlled Runtime Verification

Runtime deviations were injected against the reference workspace on 2026-09-15. The test targeted unapproved ports and unauthorized driver-access patterns. Each injection was posted as a discrete telemetry observation and correlated against the static model.

```text
CONTROLLED RUNTIME DRIFT DETECTION
┌────────────────────────────────────────────────────────────┐
│ Injected observations:        10                           │
│ Successfully ingested:       10                           │
│ Contradictions produced:     10                           │
│ Missed / unquarantined:       0                           │
│ Observed detection rate:   100%                           │
└────────────────────────────────────────────────────────────┘
```

| # | Operation | Source |
|---:|---|---|
| 1 | `WithTransportCredentials` | test harness |
| 2 | `Channel` | test harness |
| 3 | `WithCodec` | test harness |
| 4 | `ApplyServerOptions` | test harness |
| 5 | `SetExtraHeader` | test harness |
| 6 | `test_default_bool` | test harness |
| 7 | `XRevRangeN` | test harness |
| 8 | `SetPickedCluster` | test harness |
| 9 | `StaticMethod` | test harness |
| 10 | `ApplyDefaultsWithPoolSize` | test harness |

Each observation was correlated to a real entity in the static graph and checked against outgoing claims. All ten produced a `CONTRADICTED` verification in the controlled test.

### Verification-state distribution

| Status | Count |
|---|---:|
| `CONTRADICTED` | 10 |
| `SUPPORTED` | 188 |
| `UNVERIFIED` | 40,653 |

**Scope:** This measures whether injected observations that conflict with the static model are detected and quarantined. It does not measure the rate at which real production systems produce deviations.

### Reproduce the contradiction test

```bash
DATABASE_URL=... python3 scripts/runtime_contradiction_test.py
```

Expected output:

```text
═══ 10/10 contradicted ═══
```

The script cleans its test residue on entry so repeated runs do not accumulate the same fixtures. The count is taken from the verification table consumed by the attention and governance surfaces.

---

## Documentation Extraction and Verification

`scripts/doc_verify_test.py` exercises the extraction-to-verification loop. It writes a synthetic ADR containing:

- A real `CALLS` relationship.
- A real entity with an invented target.
- Fully invented names.

It then runs documentation ingestion and verification and counts the resulting states.

```bash
DATABASE_URL=... python3 scripts/doc_verify_test.py
```

Expected output:

```text
═══ Status distribution ═══
  SUPPORTED     1
  UNVERIFIED    2
  CONTRADICTED  0

═══ PASS ═══
```

This test demonstrates that a claim naming a real supported edge can become `SUPPORTED`, while claims with missing targets remain `UNVERIFIED` with reasons rather than being treated as contradictions.

---

## Telemetry Pipeline Characteristics

| Metric dimension | Observed characteristic | Verification protocol |
|---|---|---|
| Ingestion endpoint | `POST /api/v1/telemetry/spans` | OpenTelemetry/OTLP span mapping |
| Admission latency | Approximately `1.8 ms` p95 | HTTP `202 Accepted` acknowledgment |
| Verification engine | Asynchronous worker | Static-versus-runtime correlation |
| Merkle commit cadence | Approximately `10 seconds` per epoch | Static/runtime root recomputation |
| UI/IDE convergence | Approximately `1–2 minutes` | Propagation of `ARCH_DRIFT_001` markers |

These values describe the tested environment: local Linux x86_64 deployment, PostgreSQL 16 on NVMe storage, single-node worker, and no external network hop between ingestion and database. They are observations, not production SLAs.

---

## MCP Server Verification

The current MCP server exposes 21 tools and has been tested against four clients:

| Client | Verification scope |
|---|---|
| Cursor | MCP integration in the developer IDE |
| Claude Desktop | MCP integration in a desktop client with strict response handling |
| Codex CLI | MCP integration in an independent CLI client |
| Reference verifier | Dependency-free protocol, tool, and negative-path verification |

### Current reference result

```text
═══ 38/38 checks passed ═══
```

Run the current verifier against your workspace:

```bash
WORKSPACE=<workspace> python3 scripts/mcp_verify.py
```

The verifier checks:

- Initialization and request identifiers.
- MCP protocol version `2025-06-18`.
- Tool discovery and 21-tool availability.
- Entity lookup and graph relationships.
- Subclass and implementer semantics.
- Policy-proof failure behavior for unknown evaluations.
- Blast-radius output and severity fields.
- Handoff and resume negative paths.
- Structured unknown-tool errors.
- Notification response suppression.
- Clean server shutdown.

The dependency-free verifier checks protocol behavior and server contracts. Client-specific UI rendering, schema caching, and transport behavior are validated separately through Cursor, Claude Desktop, and Codex CLI usage.

---

## Workspace Console Evidence

The tenant-facing Workspace Console has been exercised across the following surfaces:

| Surface | Observed capability |
|---|---|
| Workspace | Repository, package, entity, relationship, language, policy, evidence, and health summaries |
| Agents | MCP sessions, tool-call activity, active/recent session views, and coordination state |
| Governance | Policy outcomes, documentation claims, drift, contradictions, verification states, and cryptographic trust |
| Decisions | Decision history, evidence, ancestors, descendants, revisions, and anchor status |
| Graph | Interactive topology, communities, repositories, packages, relationships, and exploration |
| Evidence and Trust | Ledger status, block height, root, parent/genesis, observation time, and verification controls |

A reference workspace view has displayed values such as 9 repositories, 1,567 packages, 14,333 entities, 40,956 relationships, 2 cross-repository bridges, 2 active policies, and 10 contradiction or attention records. These are workspace snapshot values, not product limits.

### Governance evidence example

A policy review record can display:

```text
Decision:     REVIEW
Predicate:    verification_missing
Entities:     10
Claims:       0
Contradictions: 0
Block height: #66
Proof stored: Yes
```

This demonstrates a review decision based on missing runtime evidence rather than an unsupported claim that the code is broken.

### Cryptographic trust evidence

The Evidence and Trust view can display:

- Verified ledger state.
- Block height.
- Current root.
- Parent root or genesis marker.
- Observation time.
- Stored proof state.
- Anchor verification controls.

---

## VS Code Integration Evidence

Garuda's VS Code extension is located under `vscode-extension/` and has been exercised for editor-integrated architecture and evidence workflows.

### Tested extension surfaces

- Cryptographic Ledger view.
- Quarantined Contradictions view.
- Problems-panel diagnostics.
- Policy and runtime contradiction markers.
- Architecture graph visualizer.
- Workspace AST reanalysis.
- Ledger and verification-state refresh.
- Go symbol hover with blast-radius and dependency information.

The extension manifest identifies the Garuda Architecture Shield, its commands, views, settings, Go activation paths, daemon URL, executable path, database URL, and hover-blast-radius option.

### IDE boundary

The extension is tested for editor-integrated diagnostics and exploration. Its update behavior depends on configured state, daemon/database connectivity, and refresh or reanalysis events. The evidence does not claim unrestricted continuous analysis of every file mutation without an analysis event.

---

## Multi-Agent and Session Evidence

Garuda's current MCP and coordination surface includes:

- 21 MCP tools.
- Transactional handoff.
- Checkpoint creation.
- Resume and single-consumption behavior.
- Session and tool-call activity visibility.
- Client and agent metadata where available.
- Tool duration and status tracking.
- Whitelisted argument summaries.

Handoff and resume verification includes failure paths for unknown task, agent, and checkpoint identifiers. The current reference verifier confirms structured failure responses rather than transport-level failures.

---

## What Is Not Measured

The following boundaries remain explicit:

- Real-repository relationship precision is not measured against an independent complete ground-truth corpus.
- Go compiler-backed results are high-confidence within supported relationship classes but do not prove universal runtime coverage.
- Python and TypeScript structural analysis is weaker for dynamic behavior than Go compiler-backed analysis.
- Runtime detection is measured on controlled injected observations, not on representative live production traffic.
- Cross-language resolution is not established as compiler-backed by Corpus B.
- Call edges at graph leaves can be sparse because of actual code topology and analysis scope.
- Admission latency, Merkle cadence, and UI/IDE convergence are environment-specific observations, not universal SLAs.
- Larger production-scale runtime verification remains a beta validation area.
- Dead-code and unused-import detection exposed through MCP are next-arc capabilities.
- The Merkle ledger proves integrity and inclusion of recorded evidence and decisions; it does not prove semantic-model completeness or correctness.
- Dashboard screenshots demonstrate exercised product surfaces but do not replace automated regression tests.

---

## Reproducing Results

### Analyze and save a repository

```bash
garuda analyze /path/to/repo --save --workspace <workspace-name>
```

### Reproduce controlled contradiction detection

```bash
DATABASE_URL=... python3 scripts/runtime_contradiction_test.py
```

Expected:

```text
═══ 10/10 contradicted ═══
```

### Reproduce documentation verification

```bash
DATABASE_URL=... python3 scripts/doc_verify_test.py
```

Expected: one `SUPPORTED`, two `UNVERIFIED`, zero `CONTRADICTED`, and `PASS`.

### Re-run MCP verification

```bash
WORKSPACE=<workspace-name> python3 scripts/mcp_verify.py
```

Expected:

```text
═══ 38/38 checks passed ═══
```

### Inspect ledger status

```bash
garuda status
```

### Verify ledger integrity

```bash
garuda verify
```

### Query verification-state distribution

```bash
psql "$DATABASE_URL" -c "
  SELECT status, COUNT(*)::int
    FROM claim_verifications
   WHERE workspace_id = '<workspace-id>'
   GROUP BY status;
"
```

---

## Evidence Sources

- [`README.md`](README.md) — Product claims and public positioning.
- [`PLAYBOOK.md`](PLAYBOOK.md) — Operational workflows and command reference.
- [`docs/SPECS.md`](docs/SPECS.md) — Detailed product and system specification.
- [`docs/CAPABILITIES.md`](docs/CAPABILITIES.md) — Capability maturity and analyzer matrix.
- [`docs/WALKTHROUGH.md`](docs/WALKTHROUGH.md) — Visual interface and screenshot tour.
- [`docs/adr/`](docs/adr/) — Architecture decision records.
- [`docs/invariants.md`](docs/invariants.md) — Invariant contract.
- [`scripts/mcp_verify.py`](scripts/mcp_verify.py) — MCP reference verifier.
- [`scripts/runtime_contradiction_test.py`](scripts/runtime_contradiction_test.py) — Controlled runtime contradiction test.
- [`scripts/doc_verify_test.py`](scripts/doc_verify_test.py) — Documentation extraction and verification test.
- [`vscode-extension/`](vscode-extension/) — VS Code integration source and manifest.

Historical evidence documents may contain earlier corpus sizes or MCP check totals. Such entries should be read as historical milestones; the current MCP reference result is `38/38` and the current public product surface is documented in `README.md`, `PLAYBOOK.md`, and `docs/SPECS.md`.
