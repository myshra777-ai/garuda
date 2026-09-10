# ADR 0001: Payment Refund Processing

## Context
When processing refunds across multiple gateway providers, network retries can lead to double execution.

## Decision
1. All invocations of `RefundHandler` must call `IdempotencyStore`.
2. The `RefundHandler` must be idempotent under repeated delivery.
3. The `RefundHandler` must not call `DirectBankAPI` directly.

## Consequences
Retries will safely exit early if the token has already been captured.
