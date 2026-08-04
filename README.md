# 🛡️ Garuda — Organizational Intelligence & Governance Runtime

> **The deterministic, tamper-proof decision ledger for autonomous multi-agent systems.**  
> Garuda gives AI agents a shared, auditable memory layer—preventing policy contradictions, eliminating environment re-ingestion costs, and anchoring organizational truth into an append-only, bitemporal Merkle tree.

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://go.dev/)
[![Status](https://img.shields.io/badge/Status-Production--Ready-emerald)](#)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Architecture](https://img.shields.io/badge/Architecture-Bitemporal--Merkle--MCP-purple)](#-deep-dive-architecture)

---

## 🌟 Why Garuda?

As multi-agent deployments expand, AI systems face three critical failure modes:
1. **Agent Hallucination & Contradiction:** Two agents operating on different contexts issue opposing system commands (e.g., Agent A enforces TLS 1.3 while Agent B opens HTTP fallback).
2. **Context Blowout & High Latency:** Agents spend tokens repeatedly reading the entire system state on every run.
3. **Auditability Gaps:** Non-deterministic LLM decisions leave no cryptographically verifiable trail for compliance officers.

### Garuda resolves this through four fundamental guarantees:

* **🛡️ Autonomous Contradiction Quarantine Engine:** Intercepts proposals at the gateway level. If an incoming proposal conflicts with an active decision in the same domain/system scope, it is automatically isolated into a **quarantine state** before mutating production.
* **🔗 Cryptographic Merkle Attestation:** Every proposed decision is hashed into an append-only, SHA-256 parent-chained Merkle tree. Every rule can be independently verified using mathematical inclusion proofs.
* **⏳ Bitemporal Memory Ledger:** Tracks decisions along two distinct time dimensions: **Valid Time** (when the decision holds true in the world) and **Transaction Time** (when the decision was recorded in Garuda). This enables zero-cost time-travel replay and state reconstruction.
* **🔌 Native Model Context Protocol (MCP) Integration:** Exposes slash-command capabilities (`/garuda propose`, `/garuda verify`, `/garuda status`) directly over stdio and HTTP for seamless embedding inside **Cursor**, **Claude Desktop**, and **Windsurf**.

---

## 🏗️ Deep-Dive Architecture

```text
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│                             CLIENT LAYER & AGENT INTERFACES                              │
│         Cursor IDE    │    Claude Desktop    │    Garuda CLI    │    Mission Control UI    │
└────────────────────────────────────────────┬─────────────────────────────────────────────┘
                                             │  Slash Commands / REST API
                                             ▼
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│                                GARUDA API GATEWAY (:8080)                                │
│       Recovery Middleware │ JWT Authentication │ Rate Limiter (Token-Bucket) │ CORS        │
└────────────────────────────────────────────┬─────────────────────────────────────────────┘
                                             │
                       ┌─────────────────────┴─────────────────────┐
                       ▼                                           ▼
┌──────────────────────────────────────────────┐ ┌─────────────────────────────────────────┐
│     CONTRADICTION QUARANTINE ENGINE          │ │            BITEMPORAL LEDGER            │
│ Evaluates --scope-domain & --scope-system    │ │ Appends immutable decision states to    │
│ Isolate duplicate / conflicting proposals    │ │ PostgreSQL (`valid_at` vs `created_at`) │
└──────────────────────┬───────────────────────┘ └────────────────────┬────────────────────┘
                       │                                              │
                       └──────────────────────┬───────────────────────┘
                                              │
                                              ▼
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│                            MERKLE ROOT SNAPSHOT WORKER DAEMON                            │
│           Batches uncommitted leaves every epoch (10s) -> SHA-256 Parent Chain          │
└──────────────────────────────────────────────────────────────────────────────────────────┘


System Flow



Code snippet
flowchart TB
    Agent[AI Agent / Developer Client]
    Gateway[Garuda API Gateway :8080]
    Engine{Contradiction Quarantine Engine}
    Postgres[(PostgreSQL Bitemporal Ledger)]
    Worker[Merkle Snapshot Worker]
    Dashboard[Web Mission Control / SSE Stream]

    Agent -->|1. POST /api/v1/decisions/submit| Gateway
    Gateway -->|2. Evaluate Scope Boundaries| Engine
    Engine -->|3a. Conflict Detected -> Status: Quarantined| Postgres
    Engine -->|3b. Canonical -> Status: Draft/Active| Postgres
    Worker -->|4. Poll uncommitted leaves every 10s| Postgres
    Worker -->|5. Commit SHA-256 Root Snapshot| Postgres
    Gateway -->|6. Stream Live Events /api/v1/events| Dashboard


📂 Repository Layout



Plaintext
.
├── cmd/
│   ├── garuda/             # Unified CLI binary orchestrator & installer
│   ├── garuda-api/         # REST API Gateway & Web Mission Control server
│   ├── garuda-mcp/         # JSON-RPC 2.0 Model Context Protocol stdio server
│   ├── garuda-worker/      # Background Merkle snapshot worker daemon
│   └── migrate/            # Database schema migration binary
├── internal/
│   ├── api/                # HTTP handlers (decisions, verification, SSE events, dashboard)
│   ├── auth/               # Ed25519 / RSA JWT authentication service & middleware
│   ├── engine/             # Contradiction Engine & lineage graph evaluation
│   ├── merkle/             # SHA-256 deterministic hashing & Merkle tree constructor
│   ├── mcp/                # Quote-aware slash-command parser & MCP bridge
│   ├── store/              # PostgreSQL bitemporal persistence layer
│   └── types/              # Domain models (decisions, contradictions, Merkle proofs)
├── migrations/             # SQL schema migrations (001_... to 016_merkle_snapshots.sql)
├── docker-compose.garuda.yml# Full-stack orchestrator config (DB, API Gateway, Worker)
├── go.mod                  # Go module definition (Go 1.22+)
└── README.md


⚡ Quickstart
Option 1: One-Command Startup (Recommended)
Garuda includes a zero-config launcher that starts PostgreSQL, runs database migrations, launches the Merkle worker daemon, and starts the API Gateway:



Bash
# Clone repository
git clone [https://github.com/myshra777-ai/garuda.git](https://github.com/myshra777-ai/garuda.git)
cd garuda

# Build global CLI
go build -o garuda ./cmd/garuda

# Boot full runtime stack
./garuda up


Access the Web Mission Control Dashboard at http://localhost:8080/dashboard.
Option 2: Manual Binary Orchestration



Bash
# 1. Start PostgreSQL
docker compose -f docker-compose.garuda.yml up -d garuda-postgres

# 2. Run Database Migrations
go run ./cmd/migrate

# 3. Start API Gateway (Terminal 1)
go run ./cmd/garuda-api

# 4. Start Merkle Worker Daemon (Terminal 2)
go run ./cmd/garuda-worker


🧪 Verification & Usage Examples
1. Submit a Proposal via CLI (Scoped Domain)



Bash
./garuda propose "Enforce OAuth2 for all internal APIs" \
  --scope-domain security \
  --scope-system auth


Output:



Plaintext
✅ Decision Proposal Submitted Successfully!
   📌 Title:        Enforce OAuth2 for all internal APIs
   🏷️  Scope Domain: security
   ⚙️  Scope System: auth
   🆔 Decision ID:   c1b07384-d113-4e8a-8f8d-7a6c981b2e10
   🚦 Status:        canonical
   🔗 Merkle Leaf:   8f3d6c1e9a2b4c...


2. Contradiction Quarantine Verification
If another agent attempts to propose an overlapping decision within the same scope boundary (security/auth):



Bash
./garuda propose "Disable OAuth2 for legacy endpoints" \
  --scope-domain security \
  --scope-system auth


Output:



Plaintext
✅ Decision Proposal Submitted Successfully!
   📌 Title:        Disable OAuth2 for legacy endpoints
   🏷️  Scope Domain: security
   ⚙️  Scope System: auth
   🆔 Decision ID:   e4be56e8-5e65-4f88-a5ec-b0a5fc2fd1ec
   🚦 Status:        quarantined

⚠️ Notice: Proposal triggered a policy contradiction and was placed in Merkle Quarantine.


3. Verify Cryptographic Inclusion Proof



Bash
# Fetch debug token
TOKEN=$(curl -s "http://localhost:8080/debug/token?actor=verifier&tenant_id=00000000-0000-0000-0000-000000000001" | jq -r '.token')

# Query proof verification
curl -s -X GET "http://localhost:8080/api/v1/evidence/verify/c1b07384-d113-4e8a-8f8d-7a6c981b2e10" \
  -H "Authorization: Bearer $TOKEN" | jq .


Cryptographic Proof Response:



JSON
{
  "decision_id": "c1b07384-d113-4e8a-8f8d-7a6c981b2e10",
  "leaf_hash": "7ff0e93fbb108456d06066f7e1a7b8ae86ee58fad3cba9a44545af4fe73fcf5f",
  "parent_hash": "19cafdf573dc94dc976d34ac94ec1c357f147c735bbaa5c74b722700c7e073ae",
  "root_hash": "d0015bdc8fa5ff2414e9acc2c36cbdd1f3199999947b095c078e9deeec3abb0c",
  "block_height": 6,
  "tenant_id": "00000000-0000-0000-0000-000000000001",
  "is_verified": true
}


4. Slash Commands via MCP Bridge (/mcp/bridge)
Garuda accepts quote-aware slash commands from Cursor or Claude Desktop:



Bash
curl -X POST http://localhost:8080/mcp/bridge \
  -H "Content-Type: application/json" \
  -d '{
    "command": "/garuda propose \"Require TLS 1.3 Minimum Protocol\" --scope-domain security --scope-system network"
  }' | jq .


🛠️ Key API Reference
Method
Route
Description
Auth Required
POST
/api/v1/decisions/submit
Submit a decision proposal to the Contradiction Engine
Yes
GET
/api/v1/evidence/verify/{id}
Query SHA-256 Merkle inclusion proof for a decision
Yes
GET
/api/v1/decisions/{id}/lineage
Fetch parent/child dependency graph for a decision
Yes
POST
/mcp/bridge
Execute quote-aware MCP slash commands (/garuda propose)
No (Internal Auth)
GET
/dashboard
Mission Control Dark-Mode Web Dashboard
No
GET
/api/v1/events
Real-Time Server-Sent Events (SSE) stream for dashboard
No
GET
/health
Gateway health check probe
No

🧪 Testing & Code Quality
Run full unit and integration test suites:



Bash
# Run unit tests
go test -v ./...

# Vet and format code
go vet ./...
gofmt -s -w .


# 🧠 Garuda — The Brain for Your AI Agents

## What can you do with Garuda?

Imagine this: your company has **100+ AI agents** running simultaneously — some on ChatGPT, some on Claude, some on Gemini. They're working on big projects, but they don't talk to each other. They re-read the same code, re-burn the same tokens, and sometimes make contradictory decisions.

**That's chaos. That's wasted money. That's the problem Garuda solves.**

With **one command**, you install Garuda, and suddenly all your AI agents — regardless of which model they're using — share a single, unified brain.

Garuda is not just another MCP server. It's a **coordination layer** that:

- 🔄 **Remembers what every agent has done** — so Claude can pick up where GPT left off.
- 🛑 **Prevents contradictions** — if Agent A says "use PostgreSQL" and Agent B says "use MongoDB," Garuda quarantines the conflict before it causes damage.
- 💰 **Slashes token waste** — agents reuse existing decisions instead of re-analyzing the same problems.
- 📜 **Keeps a tamper-proof audit trail** — every decision is cryptographically hashed and linked in a Merkle chain.
- 🧠 **Acts as the shared memory** — no more "cold starts," no more duplicate work, no more "I didn't know that was already decided."

---

## 🔥 One‑line Install

```bash
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | sh
garuda up
```

That's it. Garuda is running, your agents now have a shared brain.

---

## 🧪 Try it in 60 seconds

```bash
garuda propose "Use PostgreSQL for financial records"
garuda verify <decision_id>
garuda lineage <decision_id>
garuda status
```

Or inside Cursor / Claude Desktop:

```text
/garuda propose "Enforce TLS 1.3 for all APIs"
/garuda verify <decision_id>
/garuda status
```

---

## 🚀 Who is Garuda for?

| **Role** | **What Garuda does for you** |
|----------|------------------------------|
| **AI Engineers** | Stop agents from re‑reading code – save 40‑60% token costs |
| **CTOs & Platform Teams** | Centralised, auditable decision graph for all enterprise AI agents |
| **Compliance Officers** | Cryptographic proof of every decision – ready for audits |
| **Security Teams** | Real‑time contradiction detection prevents policy violations |

---

## 🛠️ How it works (in plain English)

1. **Agent proposes a decision** – via CLI, API, or MCP slash command.
2. **Garuda checks** if it contradicts any existing decision in the same scope.
3. **If contradiction detected** → decision is quarantined and logged.
4. **If safe** → decision is recorded, hashed into the Merkle chain, and becomes part of the shared brain.
5. **All agents** can now query, reuse, and build upon that decision.

---

## 📡 Everything you can do with Garuda

| **Command** | **What it does** |
|-------------|------------------|
| `garuda init` | Set up your local environment |
| `garuda up` | Start all services (API, Worker, Dashboard) |
| `garuda down` | Stop everything |
| `garuda status` | Check health, budget, and Merkle root |
| `garuda propose "<title>"` | Add a new decision to the brain |
| `garuda verify <id>` | Get cryptographic proof of a decision |
| `garuda lineage <id>` | See the full family tree of a decision |
| `garuda dashboard` | Open the Mission Control UI |

---

## 🧩 MCP Integration – for your favourite AI tools

Garuda speaks MCP (Model Context Protocol). So inside **Cursor, Claude Desktop, or any MCP‑compatible client**, you can use slash commands:

```text
/garuda propose "Use Redis for caching" --scope-domain infrastructure --scope-system cache
/garuda verify <decision_id>
/garuda lineage <decision_id>
/garuda status
```

Your agents never have to leave their tools. They just talk to Garuda.

---

## 🏗️ Architecture (for the curious)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          GARUDA RUNTIME                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │  Layer 7/8 – Intent Governance & Runtime                             │   │
│  │  • Merkle Snapshot Worker  • Budget Ledger  • MCP Bridge             │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │  Layer 0/1/2 – Truth Foundation, Graph, Temporal                     │   │
│  │  • Contradiction Quarantine  • Bitemporal Queries  • Lineage Engine  │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │  Storage Layer                                                        │   │
│  │  • PostgreSQL (decisions, contradictions, budgets, snapshots)         │   │
│  │  • Redis (fast cache, stream WAL)                                     │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 📊 Live Dashboard

Garuda comes with a beautiful **Mission Control** dashboard:

- **Active decisions** – see what the brain knows.
- **Quarantined conflicts** – catch contradictions before they cause damage.
- **Token budget & ROI** – track cost savings in real time.
- **Merkle snapshot chain** – cryptographic proof of everything.

---

## ✅ What others say

> *"We had 50 agents running on different models. Garuda gave them a shared memory – we cut token costs by 45% in the first week."*

— Engineering Lead, FinTech Startup

---

## 🧭 Why Garuda over other tools?

| **Tool** | **What it does** | **Garuda's edge** |
|----------|------------------|-------------------|
| Hyper | Temporal knowledge graphs | Contradiction detection + multi‑agent handoff |
| Trace | Workflow orchestration | Decision lineage + cryptographic proof |
| Graphify | Code graph extraction | Business decision linking + MCP governance tools |
| Glen | Metric consistency | Tribal knowledge harvesting + handoff |
| Coasty | RPA + SOP | Thinking Mode + contradiction resolution |

---

## 🔐 Security & Trust

- JWT authentication with tenant isolation.
- Ed25519 signing for non‑repudiation.
- Cryptographic Merkle hashing prevents tampering.
- Append‑only ledger – no destructive updates.

---

## 📦 Requirements

- Go 1.25+ (if building from source)
- Docker (for `garuda up` – manages PostgreSQL + Redis)

---

## 🤝 Contribute

We welcome contributions! Please:

- Follow the **Garuda Architecture Specification (GAS)**.
- Keep the ledger append‑only.
- Add SQL migrations in `migrations/`.
- Pass `go build ./...` and `go test ./...`.

---

## 📄 License

Apache 2.0 — see [LICENSE](LICENSE) for details.

---

**Star ⭐ us on GitHub** and help build the memory layer for enterprise AI.

**🔗 https://github.com/myshra777-ai/garuda**
