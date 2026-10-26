<!-- GARUDA-TALON:START -->
# GARUDA TALON — ACTIVE ARCHITECTURAL CONTRACT

- **Workspace**: `default` (`48416063-0cb2-40ea-b89e-0c92ae6d4c13`)
- **Graph Scale**: 3686 AST entities | 10743 verified edges
- **Enforcement Mode**: `ACTIVE (FAIL-CLOSED)`

## 1. Mandatory Agent Execution Protocol (Non-Negotiable)

You are operating inside a **Garuda Talon Enforced Workspace**. Every code change is verified against the AST call graph, ingested Architecture Decision Records (ADRs), and declarative YAML policies before completion:

1. **BEFORE editing any file or symbol**: Call MCP tool `garuda.talon_pre_edit` (or `garuda.briefing` + `garuda.blast_radius`) with the target file or symbol name to inspect its callers, layer boundaries, and ADR rules.
2. **WHILE writing code**: Respect all `MUST` / `MUST_NOT` ADR rules and YAML architectural policies listed below. Never introduce forbidden cross-layer imports/calls or circular package dependencies.
3. **IMMEDIATELY AFTER editing any file**: Call MCP tool `garuda.talon_verify_edit` (or CLI `./bin/garuda talon guard --diff HEAD`) to incrementally re-index your edits and run diff-scoped policy & ADR verification.
4. **ON `BLOCK` VERDICT**: You MUST NOT stop, commit, or declare the task complete. Follow the `self_repair_instructions` in the verdict immediately and re-run `garuda.talon_verify_edit` until `ALLOW` is returned.

## 2. Active Declarative Policy Guardrails

- **`[BLOCK]` Payment code must not import database/sql directly** (`block-payment-direct-db`, predicates: `claim_exists`): Direct database/sql import from payment code violates hexagonal boundary.
- **`[WARN]` Exported functions should have runtime evidence** (`warn-missing-runtime`, predicates: `verification_missing`): Exported function has no runtime evidence. Consider adding a smoke test or re-running OTel ingestion.

## 3. Ingested ADR & Documentation Claims (`document_claims`)

- No explicit ADR claims ingested yet (run `./bin/garuda docs ingest <adr.md>` to register ADR rules).

## 4. High-Blast-Radius Architectural Hubs (Modify With Caution)

- **`sqlmodel.main.SQLModel`** (class, 171 inbound edges) — `/home/rohit/garuda-validation/py/sqlmodel/sqlmodel/main.py`
- **`github.com/myshra777-ai/garuda/internal/store.Pool`** (method, 51 inbound edges) — `/home/rohit/garuda/internal/store/postgres.go`
- **`github.com/spf13/cobra.Command`** (struct, 49 inbound edges) — `/home/rohit/garuda-validation/go/cobra/command.go`
- **`github.com/myshra777-ai/garuda/internal/api.Valid`** (method, 40 inbound edges) — `/home/rohit/garuda/internal/api/generated.go`
- **`github.com/spf13/cobra.Name`** (method, 36 inbound edges) — `/home/rohit/garuda-validation/go/cobra/command.go`
- **`github.com/myshra777-ai/garuda/internal/api.RespondWithError`** (method, 33 inbound edges) — `/home/rohit/garuda/internal/api/error.go`
- **`docs_src.tutorial.fastapi.update.tutorial002_py310.HeroBase`** (class, 30 inbound edges) — `/home/rohit/garuda-validation/py/sqlmodel/docs_src/tutorial/fastapi/update/tutorial002_py310.py`
- **`github.com/spf13/cobra.Flags`** (method, 29 inbound edges) — `/home/rohit/garuda-validation/go/cobra/command.go`
- **`github.com/spf13/cobra.HasParent`** (method, 26 inbound edges) — `/home/rohit/garuda-validation/go/cobra/command.go`
- **`github.com/myshra777-ai/garuda/test/benchmark/truth_fixtures/020-large-ish/store.PostgresStore`** (struct, 25 inbound edges) — `/home/rohit/garuda/test/benchmark/truth_fixtures/020-large-ish/store/postgres.go`

<!-- GARUDA-TALON:END -->



