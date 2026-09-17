# Garuda Analyzer Capabilities & Self-Snapshot

> **Document scope:** This document has two distinct purposes: it defines the analyzer capabilities and maturity levels, and it records a dated self-snapshot of the Garuda repository used for verification. The repository metrics below are not universal product limits or benchmark claims.
>
> **Generated:** 2026-09-14 03:10:49 UTC  
> **Grounding:** AST snapshot and benchmark gates  
> **Analyzer scope:** Go AST parsing, type checking, and relationship resolution

---

## Verification Status Legend

| Status | Meaning |
|---|---|
| 🟢 **Production (GA)** | Validated for production workflows and covered by the stated benchmark gate |
| 🟡 **Beta (AST Verified)** | Functionally verified against the AST or workspace model but still subject to release or deployment constraints |
| 🟣 **Experimental (Heuristic)** | Available for exploration; results are confidence-scored and must not be treated as authoritative without additional verification |

---

## Garuda Self-Snapshot

The following counts describe the Garuda repository represented by the dated AST snapshot. They are extraction metrics for that snapshot, not general claims about every repository analyzed by Garuda.

| Metric | Count | Interpretation |
|---|---:|---|
| **Parsed files** | `289` | Source files included in the AST snapshot |
| **Packages** | `98` | Packages discovered during extraction |
| **Discovered structs** | `474` | Struct declarations identified by the analyzer |
| **Discovered interfaces** | `27` | Interface declarations identified by the analyzer |
| **Functions and methods** | `1,245` | Functions and methods identified by the analyzer |

**Struct-field metric:** The supplied snapshot reported `Total struct fields: 0`, but the field count is not considered measured or reliable for this snapshot. It is intentionally excluded from the metrics table rather than presented as a valid zero.

---

## AST Capabilities Matrix

*Go AST parsing, type checking, and relationship resolution.*

| Capability | Status | Verification tier | Supported semantics | Governing invariant | Operational interpretation |
|---|---|---|---|---|---|
| **Struct and field extraction** | 🟢 **Production (GA)** | `V5` | Struct declarations, field metadata, `json`/`db`/`validate` tags, source line spans, pointer and slice flags | `ACGM Resolution Invariant` | Suitable for production semantic indexing and downstream governance queries |
| **Method receiver disambiguation** | 🟢 **Production (GA)** | `002-method-identity` | Pointer receivers, value receivers, exported methods, unexported methods | `Canonical UUIDv5` | Distinguishes methods with receiver-aware identity resolution |
| **Interface implementation matching** | 🟢 **Production (GA)** | `003`, `009`, `010` | Dynamic method-set satisfaction and multi-interface polymorphism | `Full Recall Gate` | Resolves implementation relationships across supported method sets |
| **Generics and type parameters** | 🟢 **Production (GA)** | `004-generics` | Type-parameter filtering and generic-container extraction | `Zero False Reference Claims` | Prevents generic syntax from producing unsupported or spurious reference claims |
| **Type aliases versus definitions** | 🟢 **Production (GA)** | `005-alias` | Distinguishes `type X = Y` aliases from `type X Y` definitions | `Kind Disambiguation` | Preserves semantic identity and declaration kind |
| **Struct embedding** | 🟢 **Production (GA)** | `006-embedding` | Anonymous fields and `EMBEDS` relationship emission | `Composition Invariant` | Represents composition and promoted members as explicit semantic relationships |
| **Variadic and closure handling** | 🟢 **Production (GA)** | `007`, `008` | Ellipsis signatures and nested function-literal traversal | `Scope Resolution` | Maintains correct lexical and callable scope through nested constructs |
| **Cross-repository dependency resolution** | 🟡 **Beta (AST Verified)** | `workspace-cache` | Multi-module Go import mapping across repositories and workspaces | `Tenant Isolation` | Use for workspace-scale dependency analysis; validate repository and workspace boundaries in deployment |
| **Dynamic call-graph tracing** | 🟣 **Experimental (Heuristic)** | `heuristic` | SSA call-site resolution | `Epistemic Confidence < 0.8` | Treat results as confidence-scored hypotheses rather than authoritative call relationships |

### Benchmark identifier convention

The **Verification tier** column uses benchmark identifiers only. Human-readable descriptions belong in the capability semantics and operational-interpretation columns; this keeps the matrix traceable and consistent.

---

## Verification Policy

The status in this matrix describes capability maturity, not merely implementation presence:

- **Production (GA)** capabilities may support authoritative semantic and governance workflows within their documented scope.
- **Beta** capabilities are usable but should be monitored for workspace, repository, cache, and tenant-boundary behavior.
- **Experimental** capabilities must remain explicitly confidence-aware and should not silently override compiler-backed or persisted evidence.

For governance decisions, authoritative relationships should be derived from validated AST/type information and persisted evidence. Heuristic relationships should be labeled separately and excluded from strict compliance judgments unless an explicit policy permits them.

### Evidence authority order

```text
Compiler-backed AST/type evidence
              ↓
Persisted semantic relationships with source evidence
              ↓
Validated documentation claims
              ↓
Confidence-scored heuristic relationships
```

Lower-tier evidence must not silently replace higher-tier evidence. When evidence conflicts or required verification is unavailable, the consuming workflow should preserve the uncertainty and apply its configured fail-closed behavior.

---

## Evidence and Release Gates

| Evidence class | Required interpretation |
|---|---|
| AST snapshot metrics | Descriptive extraction coverage for the generated self-snapshot |
| Benchmark gate | Capability-specific correctness evidence for the named benchmark identifier |
| Workspace cache verification | Evidence for cross-repository resolution behavior in the configured PostgreSQL workspace |
| Heuristic evaluation | Confidence-scored output requiring independent validation |

A capability should only be promoted to a higher maturity tier after its benchmark, persistence, isolation, and failure-path behavior have been validated for the target release.

---

## Snapshot Integrity Checklist

Before publishing a new matrix, verify:

- Parsed-file, package, struct, interface, and function counts match the AST snapshot.
- Struct-field totals are either reconciled against emitted field entities or explicitly marked as unmeasured.
- Benchmark identifiers and pass rates match the current release evidence.
- Cross-repository results are checked against the intended workspace and tenant scope.
- Heuristic outputs include confidence values and are not represented as authoritative relationships.
- Generated timestamp, release commit, analyzer version, and database snapshot identity are recorded.
- Self-snapshot metrics are clearly separated from product capability claims.

---

## Aether Governance Alignment

**Aether is the internal governance framework used here to describe authority, tenant isolation, integrity, state, and evidence-handling requirements.** This section is intended for internal readers; external versions may omit it or replace it with a product-neutral governance statement.

The matrix distinguishes persisted, compiler-backed evidence from heuristic inference. Storage presence alone does not establish visibility or authority; semantic exposure should remain dependent on the applicable registry, verification, approval, and revocation conditions. Cross-tenant resolution must remain structurally unavailable, and any integrity or audit failure must fail closed.