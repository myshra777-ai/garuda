### Repository URL resolution in `garuda analyze`

`garuda analyze <path>` determines the repository URL in this order:

1. `--repo <url>` flag, if supplied. This wins unconditionally.
2. `git -C <path> remote get-url origin`, if the path is inside a
   git working tree with an `origin` remote.
3. `file://<absolute-path>`, if neither of the above resolves.

The flag is the correct input for CI, demos, and any analysis of a
subdirectory that belongs to a larger repository. Without it, every
subdirectory of the same working tree resolves to the same URL —
which collapses what the caller intended to be N repositories into
one, and defeats cross-repository analysis.

Example — analyze two fixture modules as separate repositories so
that cross-repo edges can be detected:

    ./bin/garuda analyze ./fixtures/module_auth \
      --save --workspace demo --repo file:///tmp/fixture-auth

    ./bin/garuda analyze ./fixtures/module_gateway \
      --save --workspace demo --repo file:///tmp/fixture-gateway

Without the two `--repo` flags, both modules resolve to the same
`repositories` row and no cross-repo edge can be produced.

A related consequence: `git remote get-url origin` returns different
strings for SSH and HTTPS clones of the same repository. Two
developers on different clone configurations will create two rows
in `repositories` for the same logical repository. The
`UNIQUE (workspace_id, url)` constraint prevents within-workspace
duplication of a given URL, but does not normalize SSH vs HTTPS.
Prefer `--repo` in any context where reproducibility matters.


# 🧠 Garuda Capabilities & AST Verification Matrix

> Auto-generated on `2026-09-07 02:41:51 UTC`. Grounded in AST snapshot and benchmark gates.

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
