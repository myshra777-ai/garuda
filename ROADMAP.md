# Garuda — Roadmap

**Analyze | Verify | Collaborate**

Building the evidence-backed intelligence layer for software.

---

## Current State

- **v0.2.0** — Go semantic analysis, runtime verification, cryptographic ledger
- **Validated on** 14 real-world Go repositories
- **Entities extracted:** 3,675 | **Relationships:** 5,679 | **Cross-repo bridges:** 55
- **Language support:** Go (production), Python/TypeScript/Rust (early testing)

---

## North-Star Product Loop

Intent → Semantic Model → Claims → Evidence → Implementation
→ Runtime → Verification → Agent/Developer Action → Re-verification

text

---

## Product Pillars

### Analyze
Extract typed semantic models from source code.
- Compiler-grade AST parsing (Go, Python, TypeScript, Rust)
- Deterministic entity identity (UUIDv5)
- Cross-repo relationship resolution

### Verify
Correlate intent, implementation, and runtime behavior.
- Document ingestion (Markdown, PDF, DOCX)
- Static ↔ runtime correlation via OpenTelemetry
- Tri-state model: SUPPORTED / UNVERIFIED / CONTRADICTED

### Collaborate
Expose verified context to humans and AI agents.
- MCP server for Claude, Cursor, GPT, Gemini
- CLI, IDE, dashboard, CI integration
- Shared workspace with cryptographic audit trail

---

## Knowledge Drift

Garuda detects inconsistencies between what is documented, implemented, and observed.

| Drift Class | Detection |
|-------------|-----------|
| DOC → CODE | Documented claim has no matching implementation |
| CODE → DOC | Implemented capability has no documentation |
| DOC → RUNTIME | Documented path not observed in production |
| CODE → RUNTIME | Declared path differs from observed execution |
| BUSINESS → CODE | Rule lacks enforcement mechanism |
| BUSINESS → RUNTIME | Observed state violates declared constraint |

---

## Execution Phases

### Phase 1: Core Hardening (Now → 3 months)
- Versioned benchmark corpus (20 fixtures)
- Security and isolation test suite
- Multi-language parser stability

### Phase 2: Documentation & Runtime (3 → 6 months)
- Document ingestion (Markdown, PDF, DOCX)
- Knowledge Drift v1
- Runtime coverage semantics

### Phase 3: Agent Verification Layer (6 → 12 months)
- MCP evidence queries
- Pre/post-change verification
- Design partner pilots

### Phase 4: Enterprise (12 → 24 months)
- Governance workflows
- Policy engine
- Business-state integrity

---

## Design Principles

1. **Evidence before confidence** — Every claim links to evidence.
2. **Unknown stays unknown** — UNVERIFIED is first-class, not a failure.
3. **Append-only history** — Corrections create new records, never rewrite.
4. **Small kernel, extensible perimeter** — Core stays stable; adapters evolve.
5. **One semantic layer, many interfaces** — CLI, MCP, IDE, dashboard query the same model.

---

## What Garuda Is Not

- ❌ Not a generic enterprise search product
- ❌ Not a documentation chatbot
- ❌ Not an HR/CRM/Slack ingestion platform
- ❌ Not a coding agent
- ❌ Not an observability platform

Garuda is the **verification substrate** that connects intent, implementation, and runtime.

---

## Follow Along

- **GitHub:** github.com/myshra777-ai/garuda
- **Evidence:** [EVIDENCE.md](./EVIDENCE.md)
- **Playbook:** [PLAYBOOK.md](./PLAYBOOK.md)