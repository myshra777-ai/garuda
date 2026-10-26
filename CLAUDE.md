<!-- GARUDA-TALON:START -->
# GARUDA TALON — ACTIVE ARCHITECTURAL CONTRACT

- **Workspace**: `evidence-corpus-14` (`4ccd1077-ca6f-4182-9b8f-79e1b301bf26`)
- **Graph Scale**: 0 AST entities | 0 verified edges
- **Enforcement Mode**: `ACTIVE (FAIL-CLOSED)`

## 1. Mandatory Agent Execution Protocol (Non-Negotiable)

You are operating inside a **Garuda Talon Enforced Workspace**. Every code change is verified against the AST call graph, ingested Architecture Decision Records (ADRs), and declarative YAML policies before completion:

1. **BEFORE editing any file or symbol**: Call MCP tool `garuda.talon_pre_edit` (or `garuda.briefing` + `garuda.blast_radius`) with the target file or symbol name to inspect its callers, layer boundaries, and ADR rules.
2. **WHILE writing code**: Respect all `MUST` / `MUST_NOT` ADR rules and YAML architectural policies listed below. Never introduce forbidden cross-layer imports/calls or circular package dependencies.
3. **IMMEDIATELY AFTER editing any file**: Call MCP tool `garuda.talon_verify_edit` (or CLI `./bin/garuda talon guard --diff HEAD`) to incrementally re-index your edits and run diff-scoped policy & ADR verification.
4. **ON `BLOCK` VERDICT**: You MUST NOT stop, commit, or declare the task complete. Follow the `self_repair_instructions` in the verdict immediately and re-run `garuda.talon_verify_edit` until `ALLOW` is returned.

## 2. Active Declarative Policy Guardrails

- **`[BLOCK]` API handlers must not bypass store layer to access database/sql directly or transitively** (`architecture-no-api-to-store-bypass`, predicates: `layer_violation`): API layer must route persistence operations through internal/store interfaces
rather than calling raw database/sql packages directly.
- **`[BLOCK]` Payment code must not import database/sql directly** (`payment-no-direct-db`, predicates: `claim_exists`): Direct import of database/sql from payment code violates the hexagonal
boundary established in ADR-017. Route through PaymentRepository.
- **`[REVIEW]` Exported payment handlers must have runtime evidence** (`require-runtime-evidence-for-exported`, predicates: `verification_missing`): Exported handler has no runtime evidence. Requires SRE review
before merge. Add a smoke test or confirm the handler is exercised.

## 3. Ingested ADR & Documentation Claims (`document_claims`)

- **`[MUST_NOT]`** `HandleDashboard` **MUST_NOT CALLS** `DirectBankAPI` *(source: ADR-0004: Zero-Trust Agent Execution & Pre-Edit Gate, current status: UNVERIFIED)*
- **`[MUST_NOT]`** `RefundHandler` **MUST_NOT CALLS** `DirectBankAPI` *(source: ADR 0001: Payment Refund Processing, current status: UNVERIFIED)*
- **`[MUST]`** `AppendLeafAndSealTx` **MUST IDEMPOTENT** `IdempotencyKey` *(source: ADR-0004: Zero-Trust Agent Execution & Pre-Edit Gate, current status: UNVERIFIED)*
- **`[MUST]`** `RefundHandler` **MUST CALLS** `IdempotencyStore` *(source: ADR 0001: Payment Refund Processing, current status: UNVERIFIED)*
- **`[MUST]`** `RefundHandler` **MUST IDEMPOTENT** `IdempotencyKey` *(source: ADR 0001: Payment Refund Processing, current status: UNVERIFIED)*
- **`[MUST]`** `RunGuard` **MUST CALLS** `CompileContract` *(source: ADR-0004: Zero-Trust Agent Execution & Pre-Edit Gate, current status: UNVERIFIED)*

## 4. High-Blast-Radius Architectural Hubs (Modify With Caution)

- Run `./bin/garuda analyze . --save` to populate workspace hub rankings.

<!-- GARUDA-TALON:END -->










