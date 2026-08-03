markdown
# Garuda

**The truth maintenance system for AI agents.**

Garuda gives AI agents shared memory, decision lineage, contradiction detection, and cost optimization—cutting cold starts from 8 minutes to 10ms and saving 40%+ on token costs.






# 🛡️ Garuda — Organizational Intelligence Runtime

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://go.dev/)
[![Architecture](https://img.shields.io/badge/Specification-GAS_v1.0-orange)](docs/GAS_Roadmap.md)
[![Status](https://img.shields.io/badge/Status-Phase_3_Active-green)](#-current-status--roadmap)

**Garuda** is an **Organizational Intelligence Runtime** designed to separate persistent organizational knowledge, reasoning, and governance from interchangeable foundation models. 

Instead of treating LLMs as sources of truth, Garuda operates as a persistent, append-only semantic substrate beneath models, applications, and multi-agent systems—ensuring every decision, evidence artifact, and policy action is **cryptographically verifiable, explainable, and non-repudiable**.

---

## 🌟 Key Architecture & Capabilities

* **🔒 Cryptographic Evidence Chain & Merkle Verification:** Every decision proposal generates a deterministic SHA-256 leaf hash linked to the tenant's Merkle root, offering $O(1)$ verification and tamper-evident audit trails.
* **⚡ Autonomous Contradiction Quarantine Engine:** Automatically detects conflicting decisions in the same `(scope_domain, scope_system)` pair in real-time, isolating conflicting proposals before they corrupt canonical knowledge.
* **📊 Token Budget Metering & Pre-Flight Ledger:** Prevents runaway agent cost loops by enforcing pre-flight token budget evaluation and writing consumption to an append-only ledger.
* **🔌 Model Context Protocol (MCP) Integration:** Exposes 5+ native MCP tools for dynamic agent discovery, reasoning traversal, and policy verification.
* **📜 Bitemporal & Replayable Truth:** Truth is append-only. Versions are never overwritten; state changes create immutable, hash-linked revisions for time-travel audits.

---

## 🏗️ Garuda Architecture Specification (GAS) Stack

Garuda enforces strict compliance with the **Garuda Architecture Specification (GAS)** layers:



┌───────────────────────────────────────────────────────────────────────────┐
│ GARUDA RUNTIME ARCHITECTURE │
├───────────────────────────────────────────────────────────────────────────┤
│ [Layer 7 / 8] Intent Governance & Continuous Runtime │
│ ├─ Cryptographic Merkle Attestation Engine (GAS Vol 009/011) │
│ ├─ Token Budget Metering & Pre-Flight Ledger (GAS Vol 001/007) │
│ └─ Model Context Protocol (MCP) Tool Suite (GAS Vol 011) │
├───────────────────────────────────────────────────────────────────────────┤
│ [Layer 0 / 1 / 6] Truth Foundation, Directed Graph & Cognition │
│ ├─ Autonomous Contradiction Quarantine Engine (GAS Vol 001/005) │
│ ├─ Append-Only Immutable Revisions & Decision Store (GAS Vol 001/005) │
│ └─ SHA-256 Hash-Chained Evidence Ledger (GAS Vol 001/009) │
└───────────────────────────────────────────────────────────────────────────┘



---

## 🚀 Quickstart Guide

### Prerequisites
* **Go** `1.22` or higher
* **PostgreSQL** `14` or higher
* **curl** and **jq** (for testing)

### 1. Clone & Build
```bash
git clone [https://github.com/myshra777-ai/garuda.git](https://github.com/myshra777-ai/garuda.git)
cd garuda
go build ./...


2. Configure Environment & Run Migrations
Set up your PostgreSQL database credentials:



Bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/garuda?sslmode=disable"
export JWT_SECRET="your-256-bit-secret"

# Apply all database schema migrations
go run cmd/migrate/main.go


3. Start the Garuda Gateway API



Bash
go run cmd/garuda-api/main.go


The secure API gateway will initialize on port :8080.
🧪 End-to-End Verification Walkthrough
Run the following test sequence in a terminal to verify Decision Proposal, Contradiction Quarantine, and Cryptographic Merkle Proof Generation:
Step 1: Obtain a Debug JWT Token



Bash
TOKEN=$(curl -s "http://localhost:8080/debug/token?actor=verifier&tenant_id=00000000-0000-0000-0000-000000000001" | jq -r '.token')


Step 2: Submit a Decision Proposal



Bash
DECISION_RESP=$(curl -s -X POST http://localhost:8080/api/v1/decisions/submit \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Enforce TLS 1.3 for external endpoints",
    "scope_domain": "security",
    "scope_system": "network"
  }')

echo "$DECISION_RESP" | jq .
DECISION_ID=$(echo "$DECISION_RESP" | jq -r '.id')


Step 3: Verify Cryptographic Merkle Inclusion Proof



Bash
curl -s -X GET "http://localhost:8080/api/v1/evidence/verify/$DECISION_ID" \
  -H "Authorization: Bearer $TOKEN" | jq .


Expected Attestation Output:



JSON
{
  "decision_id": "ef80caa2-3b19-4671-90c0-a63e5658cde5",
  "leaf_hash": "c6794103d1722bc5f692e280636ec4e0bccb0a3912b3da1ded859ee9025ac294",
  "parent_hash": "20b79ab70d6c1b43eaa927d0f8b1ede13353e43a450ae9ef9376eaba2231e0ab",
  "root_hash": "727c04d73838b82e2cc6c37c7ba33ad03b957741b383a6adad551f1003a73ebe",
  "block_height": 1,
  "tenant_id": "00000000-0000-0000-0000-000000000001",
  "is_verified": true,
  "created_at": "2026-08-03T23:39:58.972474+05:30"
}


📡 API Endpoints Reference
Method
Endpoint
Description
POST
/api/v1/decisions/submit
Propose a new decision; triggers quarantine checks and Merkle logging.
GET
/api/v1/evidence/verify/{id}
Retrieve cryptographic Merkle proof for a decision.
POST
/api/v1/attest/scope
Compute and attest a scope-wide Merkle root across decisions.
GET
/api/v1/decisions/{id}/lineage
Fetch parent and child decision lineage graphs.
GET
/debug/token
Development endpoint to generate tenant-scoped JWT tokens.

🗺️ Current Roadmap Status
Layer
Component
Status
Layer 0
Truth Foundation (Decisions, Revisions, Evidence, Merkle)
✅ 100% Operational
Layer 1
Directed Truth Graph (Nodes, Edges, Lineage)
✅ 85% Operational
Layer 2
Temporal Intelligence (Snapshots, Replay, Bitemporal)
🟡 65% Operational
Layer 6
Distributed Cognition (MCP Tools, Workspaces)
🟡 40% Operational
Layer 7
Intent Governance (Quarantine Engine, Policy Enforcement)
✅ 80% Operational
Layer 8
Organizational Intelligence Runtime (API Gateway, Telemetry)
🟡 60% Operational

🤝 Contributing
Contributions are welcome! Please ensure all code additions:
Conform to the Garuda Architecture Specification (GAS) laws.
Maintain append-only immutability principles.
Include relevant SQL migrations in migrations/.
Pass go build ./... and go test ./....
📄 License
Garuda is released under the Apache 2.0 License. See LICENSE for details.



<ElicitationsGroup message="How would you like to proceed with the repository setup?">
  <Elicitation label="Generate a LICENSE file (Apache 2.0)" query="Generate a standard Apache 2.0 LICENSE file for the GitHub repository."/>
  <Elicitation label="Generate a GitHub Actions CI workflow (.github/workflows/ci.yml)" query="Create a GitHub Actions CI workflow file to test and build Garuda on every push."/>
</ElicitationsGroup>


