# 🧠 Garuda Capabilities & AST Verification Matrix

> Auto-generated on `2026-08-31 03:21:31 UTC`. Grounded in AST snapshot and benchmark gates.

## Snapshot Extraction Metrics

| Metric | Count |
| :--- | :--- |
| **Parsed Files** | `215` |
| **Packages** | `78` |
| **Discovered Structs** | `372` |
| **Discovered Interfaces** | `26` |
| **Functions & Methods** | `920` |
| **Total Struct Fields** | `0` |

## Feature Verification & Status

### AST Extraction & Semantic Analysis

*Go AST parsing, type checking, and relation resolution*

| Capability | Status | Verification Tier | Supported Semantics | Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **Struct & Field Extraction** | 🟢 **Production (GA)** | `100% Benchmark (V5)` | Tags (json, db, validate), line spans, pointer/slice flags | `ACGM Resolution Invariant` |
| **Method Receiver Disambiguation** | 🟢 **Production (GA)** | `100% Benchmark (002-method-identity)` | Pointer and value receivers, exported/unexported | `Canonical UUIDv5` |
| **Interface Implementation Matching** | 🟢 **Production (GA)** | `100% Benchmark (003, 009, 010)` | Dynamic method set satisfaction, multi-interface polymorphism | `Full Recall Gate` |
| **Generics & Type Parameters** | 🟢 **Production (GA)** | `100% Benchmark (004-generics)` | Type parameter filtering, generic container extraction | `Zero False Reference Claims` |
| **Type Aliases vs Definitions** | 🟢 **Production (GA)** | `100% Benchmark (005-alias)` | type X = Y (alias) vs type X Y (type) | `Kind Disambiguation` |
| **Struct Embedding** | 🟢 **Production (GA)** | `100% Benchmark (006-embedding)` | Anonymous field EMBEDS relation emission | `Composition Invariant` |
| **Variadic & Closure Handling** | 🟢 **Production (GA)** | `100% Benchmark (007, 008)` | Ellipsis signatures, nested literal traversals | `Scope Resolution` |
| **Cross-Repo Dependency Resolution** | 🟡 **Beta (AST Verified)** | `Postgres Workspace Cache` | Multi-module Go import mapping across workspaces | `Tenant Isolation` |
| **Dynamic Call Graph Tracing** | 🟣 **Experimental (Heuristic)** | `Heuristic` | SSA call site resolution | `Epistemic Confidence < 0.8` |
