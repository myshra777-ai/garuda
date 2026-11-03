---
name: pre-edit-governance
description: Acts as the mandatory pre-change checkpoint before an AI coding agent modifies code, returning a proceed, revise, or escalate decision.
version: "1.9.11"
license: "Apache-2.0"
read_only_default: true
---

# 03 — Garuda Pre-Edit Governance Gate

Acts as the mandatory pre-change checkpoint before an AI coding agent modifies code, returning a proceed, revise, or escalate decision.

## 1. Purpose & Triggers
Invoke this skill when the user or coding workflow asks:
- "Before modifying this code, check Garuda's policies"
- "Run pre-edit governance for my planned change"
- "Can an AI agent safely edit this package?"

## 2. Required Context
- `workspace`: Target Garuda workspace name (default: `"default"`).
- `repo_dir`: Repository root path (default: `"."`).
- `diff` / `target`: Git revision range (e.g. `"HEAD"`) or symbol/package name.

## 3. Ordered MCP Workflow
1. `garuda.preflight_bundle` (or `garuda.talon_pre_edit`) — Fetch caller fan-in, package criticality tier, active YAML policies, and applicable `MUST` / `MUST_NOT` ADR claims in one call.
2. `garuda.disambiguate_logic` — If the target symbol implements a domain metric or calculation, check for competing cross-team variants (`0.0–1.0` weights) and `FROZEN_SEVERE_DO_NOT_TOUCH` locks.
3. Evaluate whether the target surface is `HIGH_SEVERE` or governed by a `BLOCK` policy before writing code.

**CLI Fallback (when MCP stdio is not attached)**:
```bash
garuda talon brief <SymbolOrFile> && garuda audit criticality
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
