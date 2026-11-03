---
name: violation-repair
description: Guides an agent through remediating a `BLOCK` or `REVIEW` governance finding without introducing secondary boundary violations.
version: "1.9.11"
license: "Apache-2.0"
read_only_default: true
---

# 05 — Garuda Assisted Violation Repair

Guides an agent through remediating a `BLOCK` or `REVIEW` governance finding without introducing secondary boundary violations.

## 1. Purpose & Triggers
Invoke this skill when the user or coding workflow asks:
- "Fix this Talon violation without breaking service boundaries"
- "Remediate the architectural BLOCK reported by Garuda"
- "Propose a minimal repair plan for this policy failure"

## 2. Required Context
- `workspace`: Target Garuda workspace name (default: `"default"`).
- `repo_dir`: Repository root path (default: `"."`).
- `diff` / `target`: Git revision range (e.g. `"HEAD"`) or symbol/package name.

## 3. Ordered MCP Workflow
1. `garuda.talon_verify_edit` — Retrieve the exact `violations` list and `self_repair_instructions`.
2. `garuda.explain_claim` — Inspect the exact source line, resolution method (`GO_TYPES`, `AST`, `DOC_RULE`), and governing policy.
3. `garuda.talon_pre_edit` — Check the proposed replacement interface or store helper before editing.
4. Propose a minimal diff (requesting human confirmation before touching `HIGH_SEVERE` surfaces), apply the fix, and re-run `garuda.talon_verify_edit` to confirm `ALLOW`.

**CLI Fallback (when MCP stdio is not attached)**:
```bash
garuda check . --diff HEAD --policy policies
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
