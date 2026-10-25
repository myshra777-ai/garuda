# 🦅 Garuda Verification & Evidence Index

> **Current reference:** Use the root [EVIDENCE.md](../EVIDENCE.md) for current v1.5.0 measurements and scope. Reports below preserve individual benchmark/snapshot details; a historical result is not a current SLA, universal accuracy score, or production certification.

This directory contains empirical benchmark reports, multi-repository stress validation datasets, runtime contradiction logs, and cryptographic Merkle verification records.

---

## Directory Overview

* **[`reports/platform-readiness.md`](reports/platform-readiness.md):** Historical measured report for a 14-repository Go-only validation corpus and a local telemetry run; not a production certification.
* **[`benchmarks/gap20-grounding.md`](benchmarks/gap20-grounding.md):** GAP-20 controlled grounding comparison; interpret its precision/token results only within the tested cases.
* **[`benchmarks/runtime-correlation.md`](benchmarks/runtime-correlation.md):** Telemetry ingestion latency, asynchronous verification throughput, and contradiction quarantine rates.
* **[`validation/14-repository.md`](validation/14-repository.md):** Full entity, claim, and cross-repository bridge inventory for the 14-repo benchmark corpus.

---

## Core Verification Invariant

> **The Garuda Epistemic Principle:** *Unknown is strictly superior to fabricated certainty.* Evidence can be classified as `SUPPORTED`, `UNVERIFIED`, or `CONTRADICTED`. Whether a claim is surfaced, filtered or blocked depends on the specific command, tool, policy and integration; these labels do not imply universal graph completeness.
