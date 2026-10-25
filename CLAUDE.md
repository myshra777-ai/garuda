<!-- GARUDA-TALON:START -->
# GARUDA TALON — ACTIVE ARCHITECTURAL CONTRACT

- **Workspace**: `default` (`48416063-0cb2-40ea-b89e-0c92ae6d4c13`)
- **Graph Scale**: 1849 AST entities | 3349 verified edges
- **Workflow Mode**: `TALON CONFIGURED` (hooks + policy checks)

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

## 4. High-Blast-Radius Architectural Hubs (Modify With Caution)

- **`sqlmodel.main.SQLModel`** (class, 171 inbound edges) — `/home/rohit/garuda-validation/py/sqlmodel/sqlmodel/main.py`
- **`github.com/spf13/cobra.Command`** (struct, 49 inbound edges) — `/home/rohit/garuda-validation/go/cobra/command.go`
- **`github.com/spf13/cobra.Name`** (method, 36 inbound edges) — `/home/rohit/garuda-validation/go/cobra/command.go`
- **`docs_src.tutorial.fastapi.update.tutorial002_py310.HeroBase`** (class, 30 inbound edges) — `/home/rohit/garuda-validation/py/sqlmodel/docs_src/tutorial/fastapi/update/tutorial002_py310.py`
- **`github.com/spf13/cobra.Flags`** (method, 29 inbound edges) — `/home/rohit/garuda-validation/go/cobra/command.go`
- **`github.com/spf13/cobra.HasParent`** (method, 26 inbound edges) — `/home/rohit/garuda-validation/go/cobra/command.go`
- **`github.com/spf13/cobra.CommandPath`** (method, 22 inbound edges) — `/home/rohit/garuda-validation/go/cobra/command.go`
- **`github.com/spf13/cobra.WriteStringAndCheck`** (function, 21 inbound edges) — `/home/rohit/garuda-validation/go/cobra/cobra.go`
- **`github.com/spf13/cobra.IsAvailableCommand`** (method, 18 inbound edges) — `/home/rohit/garuda-validation/go/cobra/command.go`
- **`github.com/spf13/cobra.Commands`** (method, 15 inbound edges) — `/home/rohit/garuda-validation/go/cobra/command.go`

<!-- GARUDA-TALON:END -->

