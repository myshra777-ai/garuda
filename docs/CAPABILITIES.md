# 🧠 Garuda Capabilities & AST Verification Matrix

> Auto-generated on `2026-11-02 04:10:56 UTC`. Grounded in AST snapshot and benchmark gates.

## Snapshot Extraction Metrics

| Metric | Count |
| :--- | :--- |
| **Parsed Files** | `325` |
| **Packages** | `106` |
| **Discovered Structs** | `552` |
| **Discovered Interfaces** | `26` |
| **Functions & Methods** | `1525` |
| **Total Struct Fields** | `0` |

## Feature Verification & Status

### Multi-Language AST, Incremental Sync & Call-Graph Extraction

*Go compiler-backed AST/type/go.work resolution and Python/TypeScript cross-file alias and instance call-graph extraction (with Pure-Go CGO_ENABLED=0 fallback)*

| Capability | Status | Verification Tier | Supported Semantics | Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **Go Struct, Interface & Field Extraction** | 🟢 **Production (GA)** | `100% Benchmark (24/24 Strict Suite)` | Tags (json, db, validate), line spans, pointer/slice flags, method sets | `ACGM Resolution Invariant` |
| **Go Method & Receiver Disambiguation** | 🟢 **Production (GA)** | `100% Benchmark (002-method-identity)` | Pointer and value receivers (ReceiverType), exported/unexported | `Canonical UUIDv5` |
| **Go Intra-Package, Cross-Package & go.work Call Graph (CALLS)** | 🟢 **Production (GA)** | `go/types (*ast.Ident, *ast.SelectorExpr, go.work)` | Direct intra-package calls, cross-package calls, and multi-module go.work resolution at 1.0 confidence | `AST_EXACT Provenance` |
| **Go Interface Implementation Matching (IMPLEMENTS)** | 🟢 **Production (GA)** | `100% Benchmark (003, 009, 010, 024)` | go/types satisfaction plus cross-module structural method-set matching at 1.0 confidence | `Full Recall Gate` |
| **Incremental Working-Tree Re-Index (garuda analyze --incremental)** | 🟢 **Production (GA)** | `Sub-Second Dirty File Sync` | Shared in-memory package caching and targeted file re-indexing for uncommitted edits | `Deterministic Graph Parity` |
| **Python AST, Alias & Instance Call Graph Extraction** | 🟢 **Production (GA)** | `Multi-Repo Corpus + Cross-File Alias/Instance Tests` | Modules, classes, methods, functions, import aliases (import X as Y), and instance method calls (0.95 confidence) | `Deterministic SHA-1 UUID` |
| **TypeScript Tree-Sitter & Pure-Go (!cgo) Call Graph** | 🟢 **Production (GA)** | `Tree-sitter (cgo) + Pure-Go Lexical/Scope (!cgo) + Cross-File Alias/Instance Tests` | Classes, interfaces, functions, named/namespace import aliases, new Class() instances, and method calls (0.95 confidence, zero C compiler needed) | `Deterministic SHA-1 UUID` |

### Graph-Native Policy, Diff Scoping & Change-Control Engine

*Declarative YAML policy evaluation, Git-diff AST targeting, and RFC 6962 Merkle governance receipts*

| Capability | Status | Verification Tier | Supported Semantics | Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **Diff-Scoped Change-Control CLI (garuda check --diff)** | 🟢 **Production (GA)** | `CLI & CI/CD Exit Code 2 Gate` | Git diff hunk-to-AST mapping + union blast radius + 3-state evidence + net-new vs baseline policy evaluation + Merkle block seal | `Deterministic Verdict` |
| **Transitive Call & Layer Boundary Predicates** | 🟢 **Production (GA)** | `Recursive SQL CTEs` | transitive_claim_exists, layer_violation, cycle_exists, blast_radius_exceeds | `Zero Hidden Coupling` |
| **3-State Edge-Level Runtime Drift & Contradiction Engine** | 🟢 **Production (GA)** | `10/10 Controlled Contradiction Suite` | Edge-level caller+callee corroboration (SUPPORTED, UNVERIFIED, CONTRADICTED) | `Automatic Quarantine` |
| **RFC 6962 Cryptographic Merkle Verification** | 🟢 **Production (GA)** | `Golden Vectors (internal/merkle)` | VerifyDecisionCryptographic, VerifyEvaluation, dual-tier epoch snapshots | `Cryptographic Tamper-Evidence` |

### Garuda Talon Active Agent Harness, Universal Runtime & 54-Tool MCP Surface

*Active pre-edit and post-edit enforcement across Cursor, Claude Code, Codex, Windsurf, and Git hooks with 1-command cross-OS provisioning*

| Capability | Status | Verification Tier | Supported Semantics | Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **Universal Zero-Dependency Installer & Doctor (install.ps1, install.sh, garuda doctor)** | 🟢 **Production (GA)** | `Windows PowerShell, macOS, Linux, Chromebook & Android` | Auto-provisions Portable Go, Portable PostgreSQL, CGO_ENABLED=0 Pure-Go binaries, PWA manifest/service-worker, and self-healing diagnostics (garuda doctor --fix) | `Zero Pre-Installed Dependencies` |
| **Active Contract & Hook Installer (garuda talon init)** | 🟢 **Production (GA)** | `internal/talon/talon_test.go` | Compiles .garuda/TALON_CONTRACT.md, .cursor/rules/garuda-talon.mdc, .claude/settings.json PostToolUse hook, CLAUDE.md, AGENTS.md, .windsurfrules, and Git hooks | `Fail-Closed Agent Binding` |
| **Active MCP Server Enforcement (bin/garuda-mcp)** | 🟢 **Production (GA)** | `63/63 Checks in scripts/mcp_verify.py` | 54 MCP tools, prompts/list (garuda.architectural_contract), resources/list (garuda://workspace/contract), and automatic first-call guardrail injection | `JSON-RPC 2.0 Spec Compliant` |
| **Deterministic Self-Repair Loop (garuda talon guard & run)** | 🟢 **Production (GA)** | `internal/talon + End-to-End Harness Verification` | Incremental AST re-index on dirty files, net-new policy & ADR verification, structured Self-Repair Prompt generation, and supervisor retry loop | `Zero Unverified Merges` |
