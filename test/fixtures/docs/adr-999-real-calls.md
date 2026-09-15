# ADR-999: Documented call edges against real entities

## Context

Synthetic document used to close the extraction → verification loop
end to end. Every bullet names a caller and a callee that both exist
as entities in the workspace where this document is ingested.

## Decision

- `usageTour` MUST call `NewClient`
- `usageTour` MUST call `Close`
- `readXInfoStreamConsumers` MUST call `DiscardNext`

## Consequences

- Two of the three claims flip to SUPPORTED against the redis
  repository in go-validation-10. The third, `usageTour` → `Close`,
  stays UNVERIFIED because multiple methods named Close exist in
  the workspace and the analyzer resolves the call to a different
  one. That is a known limitation, not a regression.
