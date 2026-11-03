---
name: change-impact
description: Calculates direct callers, transitive blast radius, criticality tier, and governing ADRs/policies before modifying any symbol or package.
version: "1.9.11"
license: "Apache-2.0"
read_only_default: true
---

# 02 — Garuda Change Impact Analysis

Calculates direct callers, transitive blast radius, criticality tier, and governing ADRs/policies before modifying any symbol or package.

## 1. Purpose & Triggers
Invoke this skill when the user or coding workflow asks:
- "What breaks if I change this function or interface?"
- "Assess the blast radius before editing"
- "Who calls this symbol across the workspace?"

## 2. Required Context
- `workspace`: Target Garuda workspace name (default: `"default"`).
- `repo_dir`: Repository root path (default: `"."`).
- `diff` / `target`: Git revision range (e.g. `"HEAD"`) or symbol/package name.

## 3. Ordered MCP Workflow
1. `garuda.find_entity` — Resolve the exact target symbol ID, package, file path, and signature.
2. `garuda.blast_radius` — Compute transitive callers and severity distribution (`CRITICAL`, `HIGH`, `MEDIUM`, `LOW`) up to depth 3–5.
3. `garuda.call_chain` — Trace concrete caller-to-callee execution paths.
4. `garuda.symbol_test_coverage` — Check whether unit or integration tests cover the target symbol and its callers.
5. `garuda.diff_impact` — Evaluate the combined impact of all currently modified files.

**CLI Fallback (when MCP stdio is not attached)**:
```bash
garuda talon brief <SymbolOrFile> --json
```

## 4. Decision & Epistemic Rules
- **Epistemic States**: Always distinguish `SUPPORTED` (backed by compiler/AST + runtime evidence where required), `UNVERIFIED` (statically extracted, awaiting runtime trace), and `CONTRADICTED` (violates ADR, policy, or static boundary).
- **Governance Verdicts**: `ALLOW` (proceed), `WARN` (proceed with note), `REVIEW` (requires human architectural review), `BLOCK` (halt and remediate net-new violation), `INCONCLUSIVE` (0 entities matched policy scope; never treat as vacuous `ALLOW`).

## 5. Output Contract
Format every report using this six-part structure:
`Finding -> Evidence (file:line & match_method) -> Affected Components (callers & tier) -> Policy / ADR -> Confidence / Status -> Recommended Action`

## 6. Safety & Limitations
- **Read-Only by Default**: Public installations run with `GARUDA_MCP_READ_ONLY=true`. Never invoke mutating tools (`garuda.approve_base_rulebook`, `garuda.decide_logic_variant`, `garuda.reconcile_data_audit`) without explicit human authorization (`GARUDA_MCP_APPROVER`).
- **Scope Qualification**: RFC 6962 Merkle proofs (`H_leaf = SHA256(0x00 || data)`) establish tamper-evident integrity and inclusion of recorded governance evaluations; they do not replace unit, integration, or dynamic security testing.
