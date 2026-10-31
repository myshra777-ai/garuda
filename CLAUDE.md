<!-- GARUDA-TALON:START -->
# GARUDA TALON — ACTIVE ARCHITECTURAL CONTRACT

- **Workspace**: `default` (`48416063-0cb2-40ea-b89e-0c92ae6d4c13`)
- **Graph Scale**: 3030 AST entities | 12371 verified edges
- **Build Receipt**: `lang=go platform=linux/amd64 cgo=off tags=[none] commit=uncommitted graph_fp=7c2be4de86ee`
- **Reproduce Command**: `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 garuda check --target linux/amd64 --cgo=false --workspace default`
- **Enforcement Mode**: `ACTIVE (FAIL-CLOSED)`

## 1. Mandatory Agent Execution Protocol (Non-Negotiable)

You are operating inside a **Garuda Talon Enforced Workspace**. Every code change is verified against the AST call graph, ingested Architecture Decision Records (ADRs), and declarative YAML policies before completion:

1. **BEFORE editing any file or symbol**: Call MCP tool `garuda.talon_pre_edit` with the target file or symbol name to inspect its callers, layer boundaries, and ADR rules.
2. **WHILE writing code**: Respect all `MUST` / `MUST_NOT` ADR rules and YAML architectural policies listed below.
3. **IMMEDIATELY AFTER editing any file**: Call MCP tool `garuda.talon_verify_edit` to verify 0 net-new policy & ADR violations.
4. **ON `BLOCK` VERDICT**: Follow the `self_repair_instructions` in the verdict immediately and re-run `garuda.talon_verify_edit` until `ALLOW` is returned.

## 2. Active Declarative Policy Guardrails

- **`[WARN]` API handlers must not bypass store layer to access database/sql directly or transitively** (`architecture-no-api-to-store-bypass`, predicates: `verification_missing`): API handlers must not bypass store layer to access database/sql directly or transitively
- **`[WARN]` Exported payment handlers must have runtime evidence** (`require-runtime-evidence-for-exported`, predicates: `verification_missing`): Exported payment handlers must have runtime evidence
- **`[WARN]` Exported functions should have runtime evidence** (`warn-missing-runtime`, predicates: `verification_missing`): Exported function has no runtime evidence. Consider adding a smoke test or re-running OTel ingestion.

## 3. Ingested ADR & Documentation Claims (`document_claims`)

- **`[MUST_NOT]`** `HandleDashboard` **MUST_NOT CALLS** `DirectBankAPI` *(source: ADR-0004: Zero-Trust Agent Execution & Pre-Edit Gate, current status: SUPPORTED)*
- **`[MUST_NOT]`** `RefundHandler` **MUST_NOT CALLS** `DirectBankAPI` *(source: ADR 0001: Payment Refund Processing, current status: UNVERIFIED)*
- **`[MUST]`** `AppendLeafAndSealTx` **MUST IDEMPOTENT** `IdempotencyKey` *(source: ADR-0004: Zero-Trust Agent Execution & Pre-Edit Gate, current status: UNVERIFIED)*
- **`[MUST]`** `RefundHandler` **MUST IDEMPOTENT** `IdempotencyKey` *(source: ADR 0001: Payment Refund Processing, current status: UNVERIFIED)*
- **`[MUST]`** `RefundHandler` **MUST CALLS** `IdempotencyStore` *(source: ADR 0001: Payment Refund Processing, current status: UNVERIFIED)*
- **`[MUST]`** `RunGuard` **MUST CALLS** `CompileContract` *(source: ADR-0004: Zero-Trust Agent Execution & Pre-Edit Gate, current status: SUPPORTED)*

## 4. High-Blast-Radius Architectural Hubs

- **`github.com/myshra777-ai/garuda/internal/store.Pool`** (method, 100 inbound edges) — `internal/store/postgres.go`
- **`github.com/myshra777-ai/garuda/cmd/garuda-mcp.safePool`** (method, 59 inbound edges) — `cmd/garuda-mcp/tools_audit.go`
- **`github.com/myshra777-ai/garuda/cmd/garuda-mcp.resolveTenantAndWorkspace`** (method, 54 inbound edges) — `cmd/garuda-mcp/tools_governance.go`
- **`github.com/myshra777-ai/garuda/cmd/garuda-mcp.resolveMCPWorkspaceName`** (method, 45 inbound edges) — `cmd/garuda-mcp/tools_audit.go`
- **`github.com/myshra777-ai/garuda/internal/api.Valid`** (method, 40 inbound edges) — `internal/api/generated.go`
- **`github.com/myshra777-ai/garuda/internal/types.Evidence`** (struct, 34 inbound edges) — `internal/types/types.go`
- **`github.com/myshra777-ai/garuda/internal/api.RespondWithError`** (method, 33 inbound edges) — `internal/api/error.go`
- **`github.com/myshra777-ai/garuda/cmd/garuda-mcp.sanitizeMCPFilePath`** (function, 24 inbound edges) — `cmd/garuda-mcp/tools_semantic.go`
- **`github.com/myshra777-ai/garuda/internal/api.applyEditors`** (method, 22 inbound edges) — `internal/api/generated.go`
- **`github.com/myshra777-ai/garuda/internal/store.Close`** (method, 22 inbound edges) — `internal/store/postgres.go`

<!-- GARUDA-TALON:END -->























































