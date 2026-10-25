<!-- GARUDA-TALON:START -->
# GARUDA TALON — ACTIVE ARCHITECTURAL CONTRACT

- **Workspace**: `default` (generated workspace ID; refresh with `garuda talon init` before relying on it)
- **Graph Scale**: snapshot values from the last contract generation; recompile before relying on counts or impact
- **Workflow Mode**: `TALON CONFIGURED` (hooks + policy checks)

> **Snapshot note:** The policy list and high-impact symbols below were generated from a prior validation workspace. They are reference context, not a live graph of this checkout. Re-run `./bin/garuda talon init . --workspace <workspace> --policies ./policies` to compile the current workspace contract.

## 1. Mandatory Agent Execution Protocol (Non-Negotiable)

You are working in a Talon-configured repository. Follow this pre-edit and post-edit protocol for supported tool operations. Talon provides context and configured checks; hooks vary by client and are not a universal interception barrier or sandbox. Run the guard against the changed diff and rely on a required CI check for the merge boundary when configured:

1. **BEFORE editing any file or symbol**: Call MCP tool `garuda.talon_pre_edit` (or `garuda.briefing` + `garuda.blast_radius`) with the target file or symbol name to inspect its callers, layer boundaries, and ADR rules.
2. **WHILE writing code**: Respect all `MUST` / `MUST_NOT` ADR rules and YAML architectural policies listed below. Never introduce forbidden cross-layer imports/calls or circular package dependencies.
3. **IMMEDIATELY AFTER editing any file**: Call MCP tool `garuda.talon_verify_edit` (or CLI `./bin/garuda talon guard --diff HEAD`) to incrementally re-index your edits and run diff-scoped policy & ADR verification.
4. **ON `BLOCK` VERDICT**: You MUST NOT stop, commit, or declare the task complete. Follow `self_repair_instructions` and retry only within the configured supervisor limit. If the result remains `BLOCK`, stop and report the unresolved findings; Talon does not guarantee that every change can be repaired to `ALLOW`.

## 2. Active Declarative Policy Guardrails

- **`[BLOCK]` Payment code must not import database/sql directly** (`block-payment-direct-db`, predicates: `claim_exists`): Direct database/sql import from payment code violates hexagonal boundary.
- **`[WARN]` Exported functions should have runtime evidence** (`warn-missing-runtime`, predicates: `verification_missing`): Exported function has no runtime evidence. Consider adding a smoke test or re-running OTel ingestion.

## 3. Ingested ADR & Documentation Claims (`document_claims`)

- No explicit ADR claims ingested yet (run `./bin/garuda docs ingest <adr.md>` to register ADR rules).

## 4. Repository-specific high-impact symbols

- The checked-in contract does not contain current repository-specific hub rankings. Run `./bin/garuda analyze . --save` and recompile the contract to populate them for the selected workspace.

<!-- GARUDA-TALON:END -->

