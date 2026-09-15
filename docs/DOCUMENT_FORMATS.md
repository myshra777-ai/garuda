# Garuda Document Formats

How to write documents that Garuda can extract verifiable claims from.

---

## Why format matters

Garuda reads documents and produces **claims**: structured
statements of the form `(subject, predicate, object, modality)`.

A claim is useful because it can be checked. `SQLModel MUST be a
class` is checkable against the semantic graph. `We should probably
think about using SQLModel` is not.

The parser reads one specific shape: **a bullet line that contains
a modal verb**. Everything else is ignored, on purpose. This is not
a limitation of ambition; it is the only way extraction can be
deterministic and honest.

---

## The bullet shape

The canonical form:

```
- <Subject> MUST <verb> <Object>
- <Subject> MUST NOT <verb> <Object>
- <Subject> SHOULD <verb> <Object>
```

Examples that produce claims:

```markdown
## Decision

- Payment service MUST use idempotency keys
- Database MUST NOT store raw credit card numbers
- HTTP server SHOULD log every request
```

Each line produces one claim:

| Subject | Predicate | Object | Modality |
|---|---|---|---|
| payment service | use | idempotency keys | MUST |
| database | store | raw credit card numbers | MUST_NOT |
| HTTP server | log | every request | SHOULD |

---

## The type-assertion shape

A subject that *is* something, rather than *does* something:

```
- <Subject> MUST be a <Type>
- <Subject> MUST be an <Type>
- <Subject> MUST be the <Type>
```

Examples:

```markdown
## Context

- SQLModel MUST be a class
- Engine MUST be an interface
```

Produces:

| Subject | Predicate | Object | Modality |
|---|---|---|---|
| SQLModel | is_a | class | MUST |
| Engine | is_a | interface | MUST |

The qualifier (`in sqlmodel.main`, `in the gin package`) is not part
of the claim. Put that information in the surrounding prose.

---

## The backtick shortcut

When both the subject and object are identifiers, wrap each in
backticks. The parser treats the first as subject, the second as
object, and produces a `CALLS` predicate.

```markdown
## Decision

- `RefundHandler` must call `PaymentStore`
- `RefundHandler` must not call `DirectBankAPI`
```

| Subject | Predicate | Object |
|---|---|---|
| RefundHandler | CALLS | PaymentStore |
| RefundHandler | CALLS | DirectBankAPI |

**Only `CALLS` claims are verified against the code graph.** A
`CALLS` claim flips to `SUPPORTED` if the graph contains a matching
edge and `CONTRADICTED` if a forbidden edge exists. Other predicates
remain `UNVERIFIED` by design.

---

## What is ignored

The parser reads bullet lines. It does not read:

- **Prose sentences.** "We should probably use SQLModel for the
  data layer" is not a claim. It has no clear subject, no clear
  object, and no explicit modality.
- **Tables.** `| I-01 | Canonical entity IDs are unique |` has no
  modal verb. It is a reference table, not a normative statement.
- **Code blocks.** Everything between ` ``` ` markers is skipped.
- **Headings without content.** A `## Decision` heading alone
  produces nothing.
- **Section headings outside the normative set.** A section titled
  "Introduction" or "Overview" is not scanned. Only sections whose
  heading contains one of: `decision, consequence, context,
  specification, invariant, requirement, rule, policy, constraint,
  approach, design, contract, behavior, behaviour`.

If a document produces fewer claims than you expect, check the
section headings and the line shapes. The parser is deterministic;
the same input always produces the same output.

---

## A worked example

A full ADR that produces clean claims:

```markdown
# ADR-0042: Payment idempotency

## Context

Payment handlers must be resilient to duplicate submissions. The
current implementation does not use idempotency keys, and this has
caused double-charges in production.

## Decision

- `RefundHandler` must call `IdempotencyStore`
- `RefundHandler` must not call `DirectBankAPI`
- Refund amount MUST NOT exceed captured amount

## Consequences

- Failed retries become safe.
- The idempotency store becomes a required dependency.
```

Ingest output:

```
✓ docs/adr/0042-payment-idempotency.md
   Extracted 3 claim(s)
   ├─ [MUST] RefundHandler CALLS IdempotencyStore (line 12)
   ├─ [MUST_NOT] RefundHandler CALLS DirectBankAPI (line 13)
   ├─ [MUST_NOT] Refund amount exceed captured amount (line 14)
```

---

## What you should know before ingesting

1. **Determinism.** The same document produces the same claims
   every time. This is not a heuristic parser that drifts.

2. **Completeness is not the goal.** A long document may produce
   three claims. That is correct. The parser extracts what it can
   verify, and ignores what it cannot.

3. **Verified answers can still be wrong.** A `SUPPORTED` claim
   means the static graph contains the referenced entity. It does
   not mean production behaves that way. A `CONTRADICTED` claim
   means the runtime observation conflicts with the static
   structure. Neither state is a substitute for reading the code.

4. **`UNVERIFIED` is not failure.** It means the claim exists but
   the verifier has not been able to check it. The most common
   reason is that the subject does not match any entity in the
   workspace. Fix the claim or analyze the repository it refers to.

5. **The graph is scoped to a workspace.** A claim about `SQLModel`
   only verifies if the workspace you ingest into contains the
   sqlmodel repository.