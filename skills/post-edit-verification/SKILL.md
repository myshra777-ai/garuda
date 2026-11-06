---
name: post-edit-verification
description: "Use after finishing a code change, and before committing or opening a pull request, to check that the edit introduced no new architectural or policy violations. Garuda re-indexes only the modified files, evaluates the git diff against configured YAML policies and ADR claims, and separates net-new violations from findings that already existed on the baseline, so reviewers see only what this change caused. It confirms the build context behind the result (platform, CGO, build tags, commit) and can verify that conditional build tags are not hiding violations on other operating systems. Returns an allow, review or block outcome with file-and-line evidence. Read-only by default."
version: "1.9.11"
license: "Apache-2.0"
read_only_default: true
---

# 04 — Garuda Post-Edit Verification

Evaluates the completed git diff against configured YAML policies and ADR claims, separating net-new violations from pre-existing baseline findings.

## 1. Purpose & Triggers
Invoke this skill when the user or coding workflow asks:
- "I've finished the change. Verify it before I commit."
- "Run Garuda diff-scoped verification on my working tree"
- "Did my edit introduce any new architectural or policy violations?"

## 2. Required Context
- `workspace`: Target Garuda workspace name (default: `"default"`).
- `repo_dir`: Repository root path (default: `"."`).
- `diff` / `target`: Git revision range (e.g. `"HEAD"`) or symbol/package name.

## 3. Ordered MCP Workflow
1. `garuda.talon_verify_edit` (with `save_to_ledger: false` in read-only mode) — Incrementally analyze modified files and evaluate net-new violations vs baseline violations.
2. `garuda.verify_receipt` — Confirm the deterministic `BuildContext` (`GOOS/GOARCH`, `CGO_ENABLED`, build tags, commit SHA, and `receipt_hash`).
3. `garuda.build_matrix_check` — Optionally verify that conditional build tags do not hide violations across `linux/amd64`, `darwin/arm64`, and `windows/amd64`.

**CLI Fallback (when MCP stdio is not attached)**:
```bash
garuda talon guard . --diff HEAD --fail-on-block
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
