---
name: architecture-review
description: Reviews actual repository architecture against documented ADRs, design specs, and declarative YAML policies with file-and-line AST evidence.
version: "1.9.11"
license: "Apache-2.0"
read_only_default: true
---

# 01 — Garuda Architecture Review

Reviews actual repository architecture against documented ADRs, design specs, and declarative YAML policies with file-and-line AST evidence.

## 1. Purpose & Triggers
Invoke this skill when the user or coding workflow asks:
- "Review this codebase's architecture"
- "Are we following our service or layer boundaries?"
- "Audit package coupling and ADR compliance"

## 2. Required Context
- `workspace`: Target Garuda workspace name (default: `"default"`).
- `repo_dir`: Repository root path (default: `"."`).
- `diff` / `target`: Git revision range (e.g. `"HEAD"`) or symbol/package name.

## 3. Ordered MCP Workflow
1. `garuda.summary` — Retrieve workspace entity, relationship, package, and epistemic verification counts.
2. `garuda.docs_graph` — Inspect the 4-tier Document -> Rule -> Policy -> Code lineage graph and check `match_method` (`ast_exact` vs `package_or_keyword_match`).
3. `garuda.criticality_tiers` — Rank packages across `HIGH_SEVERE`, `MEDIUM`, and `LOW` blast-radius tiers.
4. `garuda.contradictions` — Surface any active static-vs-ADR or runtime-vs-static contradictions.
5. `garuda.policy_evaluate` — Run all 9 graph-native YAML policy predicates in dry-run mode.

**CLI Fallback (when MCP stdio is not attached)**:
```bash
garuda summary && garuda check . --policy policies --json
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
