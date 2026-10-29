# ADR-0004: Zero-Trust Agent Execution & Pre-Edit Gate

## Decision
- `RunGuard` must call `CompileContract` before evaluating agent diffs.
- `HandleDashboard` must not call `DirectBankAPI` directly.
- `AppendLeafAndSealTx` must be idempotent under repeated delivery.