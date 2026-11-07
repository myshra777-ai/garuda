---
name: evidence-verification
description: "Use when someone asks you to prove that an architectural claim or requirement actually holds, verify a governance receipt, or export an audit trail. Garuda checks natural-language or ADR statements against the indexed code and returns each claim as supported, unverified or contradicted, with evidence. It validates deterministic receipt hashes and graph fingerprints so a result can be reproduced, verifies RFC 6962 SHA-256 Merkle inclusion proofs, and exports the full decision, policy evaluation and proof bundle for auditors or compliance review. Use it whenever a conclusion must be checkable by someone else rather than taken on trust. Read-only by default."
version: "1.9.11"
license: "Apache-2.0"
read_only_default: true
---

# 06 — Garuda Evidence & Merkle Ledger Verification

Substantiates architectural claims (`SUPPORTED`, `UNVERIFIED`, `CONTRADICTED`) and verifies RFC 6962 SHA-256 Merkle inclusion proofs and governance receipts.

## 1. Purpose & Triggers
Invoke this skill when the user or coding workflow asks:
- "Prove whether this architectural requirement is actually satisfied"
- "Verify the RFC 6962 Merkle receipt and epistemic state for this workspace"
- "Export the governance audit trail"

## 2. Required Context
- `workspace`: Target Garuda workspace name (default: `"default"`).
- `repo_dir`: Repository root path (default: `"."`).
- `diff` / `target`: Git revision range (e.g. `"HEAD"`) or symbol/package name.

## 3. Ordered MCP Workflow
1. `garuda.query_claims` — Retrieve claims filtered by `status` (`SUPPORTED`, `UNVERIFIED`, `CONTRADICTED`).
2. `garuda.verify_claim` — Verify whether a natural-language or ADR statement holds in the indexed AST.
3. `garuda.verify_receipt` — Validate the deterministic `receipt_hash` and `graph_fingerprint`.
4. `garuda.audit_export` — Export the complete governance decision, policy evaluation, and RFC 6962 Merkle proof bundle.

**CLI Fallback (when MCP stdio is not attached)**:
```bash
garuda doctor && garuda check . --json
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
