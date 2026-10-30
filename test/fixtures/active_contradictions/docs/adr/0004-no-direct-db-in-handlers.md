# ADR 0004: Strict Separation of Route Handlers and Database Storage

## Context
When AI coding agents or rapid patch developers modify API routes, they often bypass domain interfaces to execute direct SQL queries. This breaks multi-tenant transaction boundaries, bypasses caching, and leaks schema internals.

## Decision
1. All route handlers in `order` MUST call `store.OrderStore`.
2. The `OrderHandler` MUST NOT call `DirectDB.RawQuery` or import `database/sql` directly.
3. Metric calculations across teams MUST standardize unit denominators.

## Consequences
Any pull request or agent edit introducing direct database access from an HTTP route handler triggers an immediate `CONTRADICTED` status in the Garuda architectural code graph and fails the CI gate.

