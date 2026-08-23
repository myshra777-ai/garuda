# 🦅 Garuda Evidence & Verification Summary

Garuda separates product claims from reproducible evidence. This document summarizes all empirical benchmarks, multi-repository validation datasets, and runtime telemetry stress tests.

---

## 1. System Scale & Validation Scope
Corpus Progression Across Validation Milestones:

Repositories:     5 Repos ──▶ 7 Repos ──▶ 10 Repos ──▶ 14 Repos (Current)
Entities:         1,420   ──▶ 2,181   ──▶ 3,333    ──▶ 4,888 Entities
Relationships:    1,890   ──▶ 2,442   ──▶ 4,510    ──▶ 5,894 Claims
Bridges:          3       ──▶ 12      ──▶ 28       ──▶ 55 Cross-Repo Bridges


| Verification Dimension | Measured Corpus Metric | Verification State |
| :--- | :---: | :--- |
| **Tested Codebases** | **14 Repositories** | Multi-module Go workspaces |
| **AST Symbols Indexed** | **4,888 Entities** | UUIDv5 deterministic canonical identity |
| **Static Claims Indexed** | **5,894 Relationships** | Methods, interfaces, struct embeddings |
| **Cross-Module Bridges** | **55 Bridges** | Inter-repository call graph edges |
| **Controlled Contradictions** | **10 / 10 Detected** | 100% quarantine rate |
| **Dual Merkle Ledger Height** | **Block #898+** | Cryptographically verified |

---

## 2. Controlled Runtime Verification Results

CONTROLLED RUNTIME DRIFT DETECTION:
┌────────────────────────────────────────────────────────┐
│ Injected Observations:        10                       │
│ Successfully Ingested:        10                       │
│ Contradictions Detected:      10                       │
│ Missed / Unquarantined:        0                       │
│ Observed Detection Rate:    100%                       │
└────────────────────────────────────────────────────────┘
*Note: Evaluated using manually injected runtime deviations targeting unapproved ports/drivers.


---

## 3. Telemetry Pipeline Characteristics

| Metric Dimension | Observed Characteristic | Verification Protocol |
| :--- | :--- | :--- |
| **Ingestion Endpoint** | `POST /api/v1/telemetry/spans` | OpenTelemetry OTLP span mapping |
| **Admission Latency** | `~1.8 ms` (p95) | HTTP 202 Accepted acknowledgment |
| **Verification Engine** | Asynchronous worker | Evaluates static vs runtime parity |
| **Merkle Commit Cadence** | `10 seconds` per epoch | Recomputes $R_{static}$ and $R_{runtime}$ |
| **UI / IDE Convergence** | `~1–2 minutes` | Propagates `ARCH_DRIFT_001` squiggles |

---

## 4. Empirical Benchmark Reports

* **[GAP-20 Grounding Benchmark (0% Hallucination, 87.2% Compression)](evidence/benchmarks/gap20-grounding.md)**
* **[14-Repository Multi-Module Validation Report](evidence/reports/platform-readiness.md)**
* **[Visual Interface & Screenshot Tour](docs/WALKTHROUGH.md)**
