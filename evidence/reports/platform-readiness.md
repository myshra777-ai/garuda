# Garuda Platform Readiness & Multi-Repository Validation Report

> **Scope and age note:** This report is a measured snapshot of a 14-repository Go-only validation corpus and a specific local telemetry run. It is not a certification of production readiness or an SLA. For the current v1.5.0 evidence index, corpus separation, and known limits, see [EVIDENCE.md](../../EVIDENCE.md).

---

## 1. Go-only multi-repository validation corpus (14 repositories)

| Metric Dimension | Measured Corpus Value | Ledger Verification Status |
| :--- | :---: | :--- |
| **Tested Codebases** | **14 Repositories** | Isolated namespaces (Gin, Chi, Prometheus, Zap, Garuda, etc.) |
| **Architectural Packages** | **143 Modules** | Fully indexed across module boundaries |
| **Indexed AST Entities** | **4,888 Entities** | UUIDv5 canonical symbol addressing |
| **Indexed Static Claims** | **5,894 Claims** | Type declarations, method receivers, interface implementations |
| **Cross-Repo Bridges** | **55 Bridges** | Inter-module API invocation edges |
| **Controlled runtime test** | **10/10 injected contradictions** | Observed in this controlled test |
| **Ledger Epoch / Height** | **Block #898+** | Dual Merkle Roots continuously verified |

---

## 2. Telemetry Pipeline Performance
Application (OTel SDK)
│
▼ (gRPC / HTTP Spans)
OpenTelemetry Collector (garudaexporter)
│
▼ (POST /api/v1/telemetry/spans)
Garuda Ingestion Engine ──▶ [ HTTP 202 Accepted · ~1.8ms p95 ]
│
▼ (Asynchronous Verification Worker · 10s Epochs)
Dual-Root Merkle Ledger Commit
│
▼
IDE Squiggles (ARCH_DRIFT_001) & Visualizer (~1–2 min propagation)


* **Admission Latency:** `~1.8 ms` p95 HTTP 202 acknowledgment in the measured local run.
* **Verification Cadence:** Background epoch work configured at `10s` in the measured local run.
* **Dashboard / IDE Diagnostic Propagation:** `~1–2 min` observed propagation in that environment; not a guaranteed SLA.
* **Controlled contradiction test:** `10 / 10` manually injected runtime observations detected and quarantined. This does not measure recall on arbitrary production traffic.


These figures are historical environment-specific measurements. Workspace snapshot v0 rows should not be confused with the v1 decision-log Merkle inclusion-proof contract. Use EVIDENCE.md for current measured scope and limitations.
