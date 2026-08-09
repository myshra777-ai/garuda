# 🛡️ Garuda — Organizational Intelligence Runtime

**The truth maintenance system and persistent substrate for enterprise AI agents.**

Garuda separates persistent organizational knowledge, reasoning, and governance from interchangeable foundation models. It operates as a persistent, append-only semantic substrate beneath models, applications, and multi-agent systems—ensuring every decision, evidence artifact, and policy action is **cryptographically verifiable, explainable, and non-repudiable**.

---

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)](https://go.dev/)
[![Architecture](https://img.shields.io/badge/Specification-GAS_v1.0-orange)](docs/GAS_Roadmap.md)
[![Status](https://img.shields.io/badge/Status-Phase_3_Active-green)](#-current-status--roadmap)

---

## 🧠 The Story Behind Garuda

> *"I have 100 AI agents running on ChatGPT, Claude, and Gemini. They don't talk to each other. They re-read the same code, re-burn the same tokens, and sometimes contradict each other. That's chaos. That's wasted money. That's the problem Garuda solves."*

This was the problem statement that started Garuda. The boom of AI agentic workflows has been extraordinary. Every engineering team is delegating tasks to autonomous agents. But at scale, multi-agent systems break down fast:

- **Agent A** spends 8 minutes re-reading the codebase to understand context that **Agent B** already figured out an hour ago.
- **Agent C** quietly overrides a critical security policy that **Agent D** just implemented.
- Different agents using ChatGPT, Claude, and Gemini operate in silos, wasting **40–60% of your token budget** re-ingesting state and making conflicting decisions.

**Garuda was built to solve this.**

It is not just another MCP tool. It is the **central brain and governance layer** for your entire fleet of AI agents. It doesn't matter if half your team is using Claude, some are using GPT-4o, or others are on Gemini. You install Garuda with a single command, and it immediately organizes your team of AI agents so they work as a single, coordinated unit.

---

## 🌟 Why Garuda?

### The Problem
- **Context Loss & Friction:** AI agents forget context between sessions, causing 8-minute cold starts and redundant execution cycles.
- **Token Inefficiency:** Uncoordinated multi-agent deployments duplicate work, leading to 40%–60% token budget waste.
- **Semantic Drift & Contradictions:** Unmonitored models output conflicting decisions, introducing operational risk and data corruption.
- **Compliance & Audit Blindspots:** Traditional model calls lack deterministic, tamper-proof execution trails required for enterprise governance.

### The Solution

Garuda gives every agent a **permanent, auditable, shareable memory**:

| Feature | Function & Impact |
| :--- | :--- |
| **Cryptographic Evidence Chain** | Every decision is SHA‑256 hashed and linked into a per‑tenant Merkle tree – tamper‑proof, auditable, legally verifiable. |
| **Autonomous Contradiction Quarantine** | Real‑time conflict detection within `(scope_domain, scope_system)` – zero‑trust governance, conflicting proposals quarantined before they corrupt truth. |
| **Token & Cost Metering** | Pre-flight budget enforcement and token savings heuristics prevent runaway model costs. |
| **Agent State Checkpoint & Handoff** | Enables cross-model and cross-agent execution handoffs with zero-downtime state resumption – Claude ↔ GPT ↔ Gemini seamless. |
| **Bitemporal & Replayable Truth** | Stores decisions as append-only revisions, allowing point-in-time state reconstruction and audit replay. |
| **Native MCP Integration** | `/garuda` slash commands inside Cursor, Claude Desktop, or any MCP‑compatible client – zero‑learning‑curve for agents and humans. |

---

## Features
<!-- FEATURES_START -->
### 🚀 Latest Capabilities (Updated 2026-08-09)

- **Decision Registry & Immutable Truth Foundation** – versioned, append-only ledger with full metadata (who, when, why, evidence).
- **Directed Lineage Graph (DAG)** – full traceability of dependencies, supersession, and impact analysis.
- **Autonomous Contradiction Quarantine Engine** – real‑time detection and isolation of conflicting decisions.
- **Multi‑Agent Handoff & Checkpointing** – atomic, crash‑safe transfer of tasks with CAS‑deduplicated context.
- **Cryptographic Merkle Audit Ledger** – SHA‑256 hashed event chaining; every API response includes `X-Garuda-Merkle-Root`.
- **Token Budget Metering & Circuit Breakers** – real‑time tracking with pre‑flight checks and automatic fallback.
- **Bitemporal Validity Queries** – point‑in‑time truth reconstruction: "What was the state on 2026‑06‑01?"
- **Automatic Checkpointing & Idle Watchdog** – token exhaustion and 60‑min inactivity auto‑checkpoint.
- **Dynamic Model Router & Pre‑Flight Classifier** – intelligent task routing based on intent, token depth, budget, and SLA.
- **Secret Redaction & PII Protection** – real‑time regex‑based detection and masking of API keys, certificates, personal data.
- **Native MCP Integration** – slash commands inside Cursor, Claude Desktop, Gemini CLI, etc.
- **Mission Control Dashboard** – embedded dark‑mode web UI with live SSE metrics, lineage, contradictions, and budget.
- **Single‑Turn Agent Bootstrapper** – `/system/bootstrap` returns endpoints, Merkle state, budget, and MCP tools in one call.
- **Self‑Healing Error Remediation** – error responses include machine‑actionable remediation hints.
- **OpenAPI 3.0.3 Specification & Swagger UI** – interactive API docs at `/docs`.

---
_This section is auto-generated by AI from the latest commits._
<!-- FEATURES_END -->

---

## 🏗️ Architecture & Component Design

<!-- AUTO-GENERATED:DIAGRAM:START -->
```
┌──────────────────────────────────────────────────────────────────────────────────────────────┐
│                                 GARUDA RUNTIME ENGINE                                        │
├──────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────────────────────┐   │
│  │  Layer 7/8 – Intent Governance & Continuous Runtime                                  │   │
│  │  ├─ Cryptographic Merkle Attestation Engine                                         │   │
│  │  ├─ Model‑Attributed Telemetry Collector (Async Batch Ingestor)                     │   │
│  │  ├─ Token Budget Metering & Pre‑Flight Ledger                                       │   │
│  │  └─ MCP Bridge & Slash‑Command Router                                               │   │
│  └──────────────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────────────────────┐   │
│  │  Layer 0/1/6 – Truth Foundation, Directed Graph & Cognition                          │   │
│  │  ├─ Autonomous Contradiction Pre‑Flight Shield                                       │   │
│  │  ├─ Append‑Only Immutable Revisions & Decision Store                                  │   │
│  │  ├─ SHA‑256 Hash‑Chained Evidence Ledger                                             │   │
│  │  └─ Agent State Checkpoint & Handoff Engine                                          │   │
│  └──────────────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────────────────────┐   │
│  │  Storage Layer                                                                        │   │
│  │  ├─ PostgreSQL / Supabase (Decisions, Evidence, Telemetry Events)                    │   │
│  │  ├─ Redis (State Cache & SSE Broker Stream)                                          │   │
│  │  └─ Merkle Roots (Per‑Tenant Cryptographic State)                                    │   │
│  └──────────────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────────────────────┐   │
│  │  Background Daemons                                                                    │   │
│  │  ├─ Merkle Snapshot Worker (Epochic SHA‑256 parent‑chained snapshots)                │   │
│  │  └─ Telemetry Collector (Async, batched ingestion)                                   │   │
│  └──────────────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                              │
└──────────────────────────────────────────────────────────────────────────────────────────────┘
```
<!-- AUTO-GENERATED:DIAGRAM:END -->

---

## 📊 Live Observability & Telemetry Metrics

Garuda continuously captures model‑attributed operational metrics to evaluate execution efficiency, policy compliance, and resource usage.

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
* **PostgreSQL** `14` or higher (or Supabase / Render Postgres)
* **Redis** `7` or higher (optional, for caching)
* **Docker & Docker Compose** (Optional, for containerized deployments)

### 1. One‑Command Install (Recommended)

```bash
# Install Garuda globally (like Ollama, npm, or go install)
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | sh

# Start the full runtime
garuda up
```

That's it. You now have:
- ✅ PostgreSQL + Redis (managed via Docker)
- ✅ API Gateway on `:8080`
- ✅ Merkle snapshot worker
- ✅ Mission Control dashboard at `http://localhost:8080/dashboard`
- ✅ MCP bridge for slash commands

### 2. Manual Build (Without Docker)

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
go mod tidy
go build ./...
export DATABASE_URL="postgres://garuda:garudapassword@localhost:5432/garuda?sslmode=disable"
export JWT_SECRET="your-256-bit-production-secret"
export GARUDA_TELEMETRY_ENABLED="true"
go run cmd/migrate/main.go
go run cmd/garuda-api/main.go
```

---

## 🧪 Verification Walkthrough

Validate decision submission, policy validation, and Merkle inclusion proofs using `curl`:

### Step 1: Issue Debug Auth Token

```bash
TOKEN=$(curl -s "http://localhost:8080/debug/token?actor=verifier&tenant_id=00000000-0000-0000-0000-000000000001" | jq -r '.token')
```

### Step 2: Propose Policy Decision

```bash
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
```

### Step 3: Verify Cryptographic Merkle Inclusion

```bash
curl -s -X GET "http://localhost:8080/api/v1/evidence/verify/$DECISION_ID" \
  -H "Authorization: Bearer $TOKEN" | jq .
```

### Step 4: Try a Contradictory Proposal (Quarantine Test)

```bash
curl -s -X POST http://localhost:8080/api/v1/decisions/submit \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "title": "Disable TLS 1.3 for legacy endpoints",
    "scope_domain": "security",
    "scope_system": "network"
  }' | jq .
# Expected output: {"status":"quarantined","reason":"contradiction with existing decision"}
```

---

## 📡 Gateway API Reference

| Method | Path | Function |
| :--- | :--- | :--- |
| `POST` | `/api/v1/decisions/submit` | Propose a new decision; executes pre‑flight contradiction check and records Merkle evidence. |
| `GET` | `/api/v1/evidence/verify/{id}` | Retrieve cryptographic Merkle inclusion proofs for target decision. |
| `GET` | `/api/v1/decisions/{id}/lineage` | Query parent and child decision lineage relationships (DAG). |
| `POST` | `/api/v1/agents/checkpoint` | Persist execution agent state for thinking cycles. |
| `GET` | `/api/v1/agents/checkpoint/{id}` | Retrieve a saved execution state checkpoint. |
| `POST` | `/api/v1/agents/handoff` | Atomically transfer task execution across runtime agents. |
| `POST` | `/api/v1/agents/warmup` | Pre‑heat runtime states and context buffers. |
| `POST` | `/api/v1/budget/consume` | Record token/execution unit consumption against tenant balance. |
| `GET` | `/api/v1/budget` | Fetch active budget allocation and remaining limits. |
| `GET` | `/api/v1/decisions/active?at=...` | Point‑in‑time truth reconstruction. |
| `GET` | `/api/v1/audit/verify` | Verify an audit event's Merkle inclusion. |
| `POST` | `/mcp/bridge` | Execute quote‑aware slash commands (`/garuda propose`, etc.). |
| `GET` | `/dashboard` | Mission Control dark‑mode web dashboard. |
| `GET` | `/docs` | Swagger UI interactive API documentation. |

---

## 🧠 MCP Integration – Slash Commands

Garuda speaks MCP (Model Context Protocol). Inside **Cursor, Claude Desktop, or any MCP‑compatible client**, you can use slash commands:

```text
/garuda propose "Use Redis for caching" --scope-domain infrastructure --scope-system cache
/garuda verify <decision_id>
/garuda lineage <decision_id>
/garuda status
/garuda handoff <task_id> <source_agent> <target_agent>
```

No new APIs to learn – your agents just talk to Garuda.

---

## 🧩 CLI Reference

| Command | Description |
| :--- | :--- |
| `garuda init` | Set up your local environment. |
| `garuda up` | Start all services (API, Worker, Dashboard). |
| `garuda down` | Stop everything. |
| `garuda status` | Check health, budget, and Merkle root. |
| `garuda propose "<title>" [--scope-domain domain] [--scope-system system]` | Add a new decision to the brain. |
| `garuda verify <id>` | Get cryptographic proof of a decision. |
| `garuda lineage <id>` | See the full family tree of a decision. |
| `garuda handoff <task_id> <source> <target>` | Handoff task between agents. |
| `garuda resume <agent_id>` | Resume agent from checkpoint. |
| `garuda dashboard` | Open Mission Control. |
| `garuda --version` | Show version. |

---

## 🗺️ Current Status & Roadmap

| Layer | Component | Status |
| :--- | :--- | :--- |
| Layer 0 | Truth Foundation (Decisions, Revisions, Evidence, Merkle Trees) | ✅ 100% Operational |
| Layer 1 | Directed Truth Graph (Nodes, Edges, Lineage DAGs) | ✅ 100% Operational |
| Layer 2 | Temporal Intelligence (Snapshots, Bitemporal Replay) | 🟡 65% Operational |
| Layer 6 | Distributed Cognition (MCP Tools, Workspaces) | 🟡 40% Operational |
| Layer 7 | Intent Governance (Contradiction Shield, Policy Enforcement) | ✅ 90% Operational |
| Layer 8 | Runtime Gateway & Telemetry (API Gateway, Telemetry Ingestor) | ✅ 100% Operational |

**Overall Platform: ~70% of Full GAS Roadmap**

---

## 📊 Live Dashboard

Mission Control (`/dashboard`) shows:
- **Active decisions** – what the brain knows.
- **Quarantined conflicts** – catch contradictions before they cause damage.
- **Token budget & ROI** – real‑time cost savings.
- **Merkle snapshot chain** – cryptographic proof of everything.
- **SSE event stream** – live telemetry feed.

---

## 🔐 Security & Trust

- JWT authentication with tenant isolation.
- Ed25519 signing for non‑repudiation.
- Cryptographic Merkle hashing prevents tampering.
- Append‑only ledger – no destructive updates.
- Built‑in secret redaction for API keys and PII.
- GDPR/DPDP‑compliant telemetry with explicit consent and opt‑out.

---

## 🌍 Why Garuda Over Competitors?

| Competitor | Their Focus | Garuda's Edge |
| :--- | :--- | :--- |
| **Hyper** | Temporal knowledge graphs | Contradiction detection + multi‑agent handoff |
| **Trace** | Workflow orchestration | Decision lineage + cryptographic proof |
| **Graphify** | Code graph extraction | Business decision linking + MCP governance tools |
| **Glen** | Metric consistency | Tribal knowledge harvesting + handoff |
| **Coasty** | RPA + SOP | Thinking Mode + contradiction resolution |

---

## 🤝 Contributing

Contributions are welcome! Please ensure all code additions:

- Follow the **Garuda Architecture Specification (GAS)** laws.
- Maintain **append‑only immutability** principles.
- Include relevant SQL migrations in `migrations/`.
- Pass `go build ./...` and `go test ./...`.
- Update the README (auto‑generated by AI, or manually).

---

## 📄 License

Garuda is released under the **Apache 2.0 License**. See the [LICENSE](LICENSE) file for details.

---

**Star ⭐ us on GitHub** and help build the memory layer for enterprise AI.

**🔗 https://github.com/myshra777-ai/garuda**
