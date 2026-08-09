Markdown
# 🛡️ Garuda — Organizational Intelligence Runtime

**The truth maintenance system and persistent substrate for enterprise AI agents.**

Garuda separates persistent organizational knowledge, reasoning, and governance from interchangeable foundation models. It operates as a persistent, append-only semantic substrate beneath models, applications, and multi-agent systems—ensuring every decision, evidence artifact, and policy action is **cryptographically verifiable, explainable, and non-repudiable**.

---

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)](https://go.dev/)
[![Architecture](https://img.shields.io/badge/Specification-GAS_v1.0-orange)](docs/GAS_Roadmap.md)
[![Status](https://img.shields.io/badge/Status-Phase_3_Active-green)](#-current-status--roadmap)

---

## 🌟 Why Garuda?

### The Problem
* **Context Loss & Friction:** AI agents forget context between sessions, causing 8-minute cold starts and redundant execution cycles.
* **Token Inefficiency:** Uncoordinated multi-agent deployments duplicate work, leading to 40%–60% token budget waste.
* **Semantic Drift & Contradictions:** Unmonitored models output conflicting decisions, introducing operational risk and data corruption.
* **Compliance & Audit Blindspots:** Traditional model calls lack deterministic, tamper-proof execution trails required for enterprise governance.

### The Solution

| Feature | Function & Impact |
| :--- | :--- |
| **Cryptographic Evidence Chain** | Hashes every proposal and decision into a Merkle tree for immutable, tamper-proof audit trails. |
| **Pre-Flight Contradiction Shield** | Evaluates new proposals against active canonical policies in real-time, quarantining conflicts before state mutation. |
| **Token & Cost Metering** | Pre-flight budget enforcement and token savings heuristics prevent runaway model costs. |
| **Agent State Checkpoint & Handoff** | Enables cross-model and cross-agent execution handoffs with zero-downtime state resumption. |
| **Bitemporal & Replayable Truth** | Stores decisions as append-only revisions, allowing point-in-time state reconstruction and audit replay. |

---

## 🚀 System Capabilities & Features

<!-- AUTO-GENERATED:FEATURES:START -->
* **Decision Engine Shield:** Real-time pre-flight validation preventing policy-contradicting proposals.
* **Model-Attributed Telemetry:** Non-blocking async collector logging 32+ metrics into PostgreSQL/Supabase.
* **Agent Handoff & Resume:** Atomic state preservation across execution agents with checkpoint restoration.
<!-- AUTO-GENERATED:FEATURES:END -->

---

## 🏗️ Architecture & Component Design

<!-- AUTO-GENERATED:DIAGRAM:START -->
┌─────────────────────────────────────────────────────────────────────────────┐
│                          GARUDA RUNTIME ENGINE                              │
├─────────────────────────────────────────────────────────────────────────────┤
│  Layer 7/8: Governance, Telemetry & Ingestion                               │
│  ├─ Cryptographic Merkle Attestation Engine                                 │
│  ├─ Model-Attributed Telemetry Collector (Async Batch Ingestor)             │
│  └─ Token Budget Metering & Pre-Flight Ledger                               │
├─────────────────────────────────────────────────────────────────────────────┤
│  Layer 0/1/6: Truth Foundation, Directed Graph & Cognition                  │
│  ├─ Autonomous Contradiction Pre-Flight Shield                              │
│  ├─ Append-Only Immutable Revisions & Decision Store                        │
│  └─ SHA-256 Hash-Chained Evidence Ledger                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│  Storage Layer                                                              │
│  ├─ PostgreSQL / Supabase (Decisions, Evidence, Telemetry Events)           │
│  ├─ Redis (State Cache & SSE Broker Stream)                                 │
│  └─ Merkle Roots (Per-Tenant Cryptographic State)                           │
└─────────────────────────────────────────────────────────────────────────────┘

<!-- AUTO-GENERATED:DIAGRAM:END -->

---

## 📊 Live Observability & Telemetry Metrics

Garuda continuously captures model-attributed operational metrics to evaluate execution efficiency, policy compliance, and resource usage.

<!-- AUTO-GENERATED:METRICS:START -->
| Metric Category | Metrics Tracked |
| :--- | :--- |
| **Instance Context** | `instance_hash`, `session_id`, `garuda_version`, `agent_runtime` |
| **Decision Lifecycle** | `decision_status`, `scope_domain`, `scope_system`, `decision_confidence`, `contradictions_caught` |
| **Model Attribution** | `model_name`, `model_provider`, `model_route` |
| **Cost & Efficiency** | `tokens_estimated`, `tokens_saved`, `estimated_cost_usd`, `budget_remaining` |
| **Performance (ms)** | `cold_start_latency_p50/p95/p99`, `warm_start_latency_p50/p95/p99`, `api_latency_p50/p95/p99`, `handoff_latency` |
| **Efficacy Rates** | `handoff_success_rate`, `contradiction_reduction_rate`, `token_reuse_rate`, `hallucinations_prevented` |
<!-- AUTO-GENERATED:METRICS:END -->

---

## ⚡ Quickstart Guide

### Prerequisites
* **Go** `1.25` or higher
* **PostgreSQL** `14` or higher (or Supabase)
* **Redis** `7` or higher
* **Docker & Docker Compose** (Optional, for containerized deployments)

### 1. Installation

```bash
git clone [https://github.com/myshra777-ai/garuda.git](https://github.com/myshra777-ai/garuda.git)
cd garuda
go mod tidy
go build ./...
2. Local Environment Setup
Configure database connections and secrets:

Bash
export DATABASE_URL="postgres://garuda:garudapassword@localhost:5432/garuda?sslmode=disable"
export JWT_SECRET="your-256-bit-production-secret"
export GARUDA_TELEMETRY_ENABLED="true"

# Apply database migrations
go run cmd/migrate/main.go
3. Run via Docker Compose
To deploy the full production stack—including PostgreSQL, Redis, Garuda Gateway API, Background Worker, and Telemetry Collector—run:

Bash
docker compose -f deploy/compose/docker-compose.prod.yml up --build -d
🧪 Verification Walkthrough
Validate decision submission, policy validation, and Merkle inclusion proofs using curl:

Step 1: Issue Debug Auth Token
Bash
TOKEN=$(curl -s "http://localhost:8080/debug/token?actor=verifier&tenant_id=00000000-0000-0000-0000-000000000001" | jq -r '.token')
Step 2: Propose Policy Decision
Bash
DECISION_RESP=$(curl -s -X POST http://localhost:8080/api/v1/decisions/submit \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Model: claude-3-5-sonnet" \
  -H "X-Model-Provider: anthropic" \
  -d '{
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "title": "Enforce TLS 1.3 across external endpoints",
    "scope_domain": "security",
    "scope_system": "network"
  }')

echo "$DECISION_RESP" | jq .
DECISION_ID=$(echo "$DECISION_RESP" | jq -r '.id')
Step 3: Verify Cryptographic Merkle Inclusion
Bash
curl -s -X GET "http://localhost:8080/api/v1/evidence/verify/$DECISION_ID" \
  -H "Authorization: Bearer $TOKEN" | jq .
📡 Gateway API Reference
Method	Path	Function
POST	/api/v1/decisions/submit	Proposes a new decision; executes pre-flight contradiction check and records Merkle evidence.
GET	/api/v1/evidence/verify/{id}	Retrieves cryptographic Merkle inclusion proofs for target decision.
GET	/api/v1/decisions/{id}/lineage	Queries parent and child decision lineage relationships (DAG).
POST	/api/v1/agents/checkpoint	Persists execution agent state for thinking cycles.
GET	/api/v1/agents/checkpoint/{id}	Retrieves a saved execution state checkpoint.
POST	/api/v1/agents/handoff	Atomically transfers task execution across runtime agents.
POST	/api/v1/agents/warmup	Pre-heats runtime states and context buffers.
POST	/api/v1/budget/consume	Records token/execution unit consumption against tenant balance.
GET	/api/v1/budget	Fetches active budget allocation and remaining limits.
🗺️ Current Status & Roadmap
Layer	Component	Status
Layer 0	Truth Foundation (Decisions, Revisions, Evidence, Merkle Trees)	✅ 100% Operational
Layer 1	Directed Truth Graph (Nodes, Edges, Lineage DAGs)	✅ 100% Operational
Layer 2	Temporal Intelligence (Snapshots, Bitemporal Replay)	🟡 65% Operational
Layer 6	Distributed Cognition (MCP Tools, Workspaces)	🟡 40% Operational
Layer 7	Intent Governance (Contradiction Shield, Policy Enforcement)	✅ 90% Operational
Layer 8	Runtime Gateway & Telemetry (API Gateway, Telemetry Ingestor)	✅ 100% Operational
📄 License
Garuda is released under the Apache 2.0 License. See the LICENSE file for details.
