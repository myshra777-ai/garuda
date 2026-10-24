# 🧠 Garuda Capabilities & Verification Matrix (v1.2.0)

> Grounded in the Garuda v1.2.0 multi-language semantic engine, graph-native policy evaluator, 3-state runtime verifier, and RFC 6962 Merkle ledger.

---

## 1. Plane 01 — Multi-Language AST & Call-Graph Extraction

| Capability | Status | Verification Tier | Supported Semantics | Enforced Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **Go Struct, Interface & Field Extraction** | 🟢 **Production (GA)** | `100% Benchmark (GAP-20)` | Struct tags (`json`, `db`, `validate`), line spans, pointer/slice flags, method sets | `ACGM Resolution Invariant` |
| **Go Method & Receiver Disambiguation** | 🟢 **Production (GA)** | `100% Benchmark (002-method-identity)` | Pointer and value receivers (`ReceiverType`), exported/unexported methods | `Canonical UUIDv5` |
| **Go Cross-Package & Intra-Package `CALLS`** | 🟢 **Production (GA)** | `go/types` + `*ast.Ident` & `*ast.SelectorExpr` | Both cross-package (`pkg.Fn()`) and same-package (`fn()`) direct function/method calls | `AST_EXACT Provenance` |
| **Go Interface Satisfaction (`IMPLEMENTS`)** | 🟢 **Production (GA)** | `go/types` + Cross-Package Structural Pass | Type-checked `types.Implements` plus cross-package structural method-set matching (`deriveWorkspaceInterfaceImplementations`) | `Full Recall Gate` |
| **Go Struct & Interface Embedding (`EMBEDS`)** | 🟢 **Production (GA)** | `100% Benchmark (006-embedding)` | Anonymous struct fields and embedded interfaces | `Composition Invariant` |
| **Python AST & Call Graph (`CALLS`, `INHERITS`, `IMPORTS`)** | 🟢 **Production (GA)** | `Multi-Repo Corpus (sqlmodel, fastapi)` | Modules, classes, methods, functions, `IMPORTS`, `INHERITS`, and body `CALLS` (`self.method()`, intra-module, and imported functions) | `Deterministic SHA-1 UUID` |
| **TypeScript Tree-Sitter Graph (`CALLS`, `IMPLEMENTS`, `INHERITS`, `IMPORTS`)** | 🟢 **Production (GA)** | `Tree-sitter TypeScript Grammar` | Classes, interfaces, functions, arrow functions, methods, relative/package `IMPORTS`, `INHERITS`, `IMPLEMENTS`, and `call_expression` `CALLS` | `Deterministic SHA-1 UUID` |
| **Cross-Repository Dependency Resolution** | 🟢 **Production (GA)** | `Postgres Workspace Symbol Cache` | Multi-repository module path and symbol bridge mapping within a workspace | `Workspace Isolation` |

---

## 2. Plane 02 — Epistemic Evidence & Runtime Drift Verification

| Capability | Status | Verification Tier | Supported Semantics | Enforced Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **3-State Epistemic Classification** | 🟢 **Production (GA)** | `Database Constraint + Verifier` | Every claim is classified as `SUPPORTED`, `UNVERIFIED`, or `CONTRADICTED` | `Honest Defaults (Never Upgrade Unverified)` |
| **OpenTelemetry Runtime Span Ingestion** | 🟢 **Production (GA)** | `POST /api/v1/telemetry/spans` | Correlates live service spans (`caller_symbol`, `target_endpoint`) against static workspace entities | `Tenant & Workspace Scoped` |
| **Bidirectional Runtime Contradiction Detection** | 🟢 **Production (GA)** | `10/10 Controlled Contradiction Suite` | Flags both explicit unapproved endpoints AND shadow runtime edges missing from static `claims` as `CONTRADICTED` | `Automatic Quarantine` |
| **Normative Documentation Claim Verification** | 🟢 **Production (GA)** | `scripts/doc_verify_test.py` | Extracts architectural claims from Markdown ADRs (`garuda docs ingest`) and verifies them against the static graph (`garuda docs verify`) | `Evidence-Backed Drift Detection` |

---

## 3. Plane 03 — Deterministic Graph-Native Policy Engine

| Predicate Type | Status | Query Mechanism | Purpose & Parameters |
| :--- | :--- | :--- | :--- |
| **`layer_violation`** | 🟢 **Production (GA)** | SQL Regex Join on `claims` + `entities` | Blocks direct `CALLS` or `IMPORTS` from `from_layer` package regex to `to_layer` package regex |
| **`transitive_claim_exists`** | 🟢 **Production (GA)** | Recursive CTE (`WITH RECURSIVE call_path`) | Detects multi-hop backdoor dependency chains (`from_name_pattern` -> ... -> `to_name_pattern` up to `max_depth`) |
| **`cycle_exists`** | 🟢 **Production (GA)** | Recursive CTE (`WITH RECURSIVE cycle_search`) | Detects circular package dependencies (`A -> B -> C -> A`) up to `max_depth` packages |
| **`blast_radius_exceeds`** | 🟢 **Production (GA)** | Inbound Caller Aggregation (`HAVING COUNT > N`) | Escalates changes to high-fan-in hubs when inbound callers exceed `max_callers` |
| **`claim_exists`** | 🟢 **Production (GA)** | Scoped `claims` Join (`scope.package`, `scope.pattern`) | Matches specific `claim_type`, `from_name_pattern`, `to_name_pattern`, and `min_confidence` |
| **`entity_exists`** | 🟢 **Production (GA)** | Scoped `entities` Filter (`scope.package`, `scope.pattern`) | Matches entities by `kind`, `name_pattern`, and `exported` status |
| **`contradiction_exists`** | 🟢 **Production (GA)** | `contradictions` + `claim_verifications` | Blocks or flags workspaces containing unresolved `CONTRADICTED` runtime or decision rows |
| **`verification_missing`** | 🟢 **Production (GA)** | Left-Join Anti-Match on `claim_verifications` | Warns when exported functions/methods lack `SUPPORTED` runtime verification |
| **`language_matches`** | 🟢 **Production (GA)** | Repository / Entity `language` Filter | Scopes policy rules to `go`, `python`, or `typescript` targets |

---

## 4. Plane 04 — Unified Change-Control & RFC 6962 Merkle Ledger

| Capability | Status | Verification Tier | Supported Semantics | Enforced Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **Unified Change-Control CLI (`garuda check`)** | 🟢 **Production (GA)** | `CLI + CI/CD Exit Code 2 Gate` | Combines Blast Radius + 3-State Evidence + Policy Evaluation + Merkle Block Seal into one Decision Card | `Deterministic Verdict` |
| **RFC 6962 Merkle Tree & Inclusion Proofs** | 🟢 **Production (GA)** | `Golden Vectors (internal/merkle)` | Domain-separated leaf hashing (`0x00` leaf, `0x01` internal node), odd-leaf duplication guard, tamper detection | `Cryptographic Tamper-Evidence` |
| **Cryptographic Decision & Policy Verification** | 🟢 **Production (GA)** | `VerifyDecisionCryptographic` & `VerifyEvaluation` | Recomputes canonical v1 byte encoding (`CanonicalDecision`, `CanonicalEvaluation`) and verifies `V1Proof` path to epoch root | `Zero Trust Verification` |
| **Dual-Tier Epoch Snapshots** | 🟢 **Production (GA)** | `merkle_snapshots` + `merkle_epoch_leaves` | Anchors both static analysis leaves and runtime verification leaves into periodic tenant epochs | `Bitemporal Auditability` |

---

## 5. Plane 05 — Agent MCP Surface, IDE & Zero-Config Runtime

| Capability | Status | Verification Tier | Supported Semantics | Enforced Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **MCP Stdio Server (`bin/garuda-mcp`)** | 🟢 **Production (GA)** | `23/23 checks in scripts/mcp_verify.py` | 21+ tools (`briefing`, `find_entity`, `neighbors`, `subclasses`, `implementers`, `blast_radius`, `handoff`, `resume`, `verify_policy_evaluation`) | `JSON-RPC 2.0 Spec Compliant` |
| **Transactional Multi-Agent Handoff & Resume** | 🟢 **Production (GA)** | `internal/store/handoff_store.go` | Atomic agent checkpoint locking, lineage tracking, and state resumption across agents | `Serializable Handoff Lineage` |
| **1-Command Zero-Config Onboarding (`./setup.sh` / `garuda init`)** | 🟢 **Production (GA)** | `End-to-End Automated Setup` | Auto-starts Docker Postgres, runs idempotent migrations, writes `.env` + Ed25519 JWT keys, provisions owner user, configures Cursor/Claude MCP | `Zero Manual Env Export` |
| **Multi-Tenant Web Console & Control Plane** | 🟢 **Production (GA)** | `cmd/garuda-api` & `garuda dev` | Session-authenticated `/dashboard`, interactive D3 force graph (`/graph`), and operator Control Plane (`/admin`) | `Strict Workspace Membership Isolation` |
