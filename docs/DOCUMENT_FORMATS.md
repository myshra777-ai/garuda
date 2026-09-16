```markdown
# Garuda Document Formats

How to write documents that Garuda can extract verifiable claims from.

---

## Why format matters

Garuda reads documents and produces **claims**: structured statements of the form `(subject, predicate, object, modality)`.

A claim is useful because it can be checked. `SQLModel MUST be a class` is checkable against the semantic graph. `We should probably think about using SQLModel` is not.

The parser reads one specific shape: **a bullet line that contains a modal verb**. Everything else is ignored, on purpose. This is not a limitation of ambition; it is the only way extraction can be deterministic and honest.

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
| Payment service | use | idempotency keys | MUST |
| Database | store | raw credit card numbers | MUST_NOT |
| HTTP server | log | every request | SHOULD |

Subject and object preserve their original casing. The parser does not lowercase them. The table above shows the values as the parser would extract them from the examples.

### Recognized modality words

The parser recognizes more than the three primary modal verbs. Each group below produces the same modality.

| Modality | Recognized words |
| :--- | :--- |
| MUST | `must`, `shall`, `requires`, `require`, `enforces`, `enforce`, `has to`, `have to` |
| MUST NOT | `must not`, `shall not`, `cannot`, `may not`, `never` |
| SHOULD | `should`, `recommended`, `ought to` |

Use the primary form (`MUST`, `MUST NOT`, `SHOULD`) in new documents. The synonyms are supported so existing prose does not have to be rewritten before ingestion.

### Bullet characters

Four prefixes are recognized. Any of them works.

```
-     (hyphen)
*     (asterisk)
+     (plus)
•     (bullet)
```

A line that does not begin with one of these prefixes is treated as prose and skipped, even if it contains a modal verb.

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

The qualifier (`in sqlmodel.main`, `in the gin package`) is not part of the claim. Put that information in the surrounding prose.

---

## The backtick shortcut

When both the subject and object are identifiers, wrap each in backticks. The parser treats the first as subject, the second as object, and produces a `CALLS` predicate.

```markdown
## Decision

- `RefundHandler` must call `PaymentStore`
- `RefundHandler` must not call `DirectBankAPI`
```

| Subject | Predicate | Object |
|---|---|---|
| RefundHandler | CALLS | PaymentStore |
| RefundHandler | CALLS | DirectBankAPI |

**Only `CALLS` claims are verified against the code graph.** A `CALLS` claim flips to `SUPPORTED` if the graph contains a matching edge and `CONTRADICTED` if a forbidden edge exists. Other predicates remain `UNVERIFIED` by design.

### The idempotency shortcut

When a line contains a single backticked identifier and the word `idempotent` or `idempotency`, the parser produces a fixed claim:

```markdown
- `RefundHandler` must be idempotent
```

| Subject | Predicate | Object |
|---|---|---|
| RefundHandler | idempotent | IdempotencyKey |

The object is always `IdempotencyKey` — it is a marker, not a reference. This is a shorthand for a common assertion that would otherwise be ambiguous to parse.

---

## What is ignored

The parser reads bullet lines. It does not read:

- **Prose sentences.** "We should probably use SQLModel for the data layer" is not a claim. It has no clear subject, no clear object, and no explicit modality.
- **Tables.** `| I-01 | Canonical entity IDs are unique |` has no modal verb. It is a reference table, not a normative statement.
- **Code blocks.** Everything between ` ``` ` markers is skipped.
- **Headings without content.** A `## Decision` heading alone produces nothing.
- **Section headings outside the normative set.** A section titled "Introduction" or "Overview" is not scanned. Only sections whose heading *contains* one of the keywords below are read.

### Normative section keywords

A section is scanned if its heading contains any of these substrings, case-insensitively:

```
decision       consequence    context        specification
invariant      requirement    rule           policy
constraint     approach       design         contract
behavior       behaviour
```

The match is a substring, not an equality. A section titled `## Decisions and Tradeoffs` matches because it contains `decision`. A section titled `## Notes` does not match.

### Subject and object validation

Not every line that matches the bullet shape produces a claim. A few conditions reject the parsed result.

**Subject rejected if:**
- It is empty.
- It contains a comma, semicolon, or colon. A comma in a subject means the parser picked up a sentence fragment, not a noun phrase.
- It is longer than six words. Long subjects are almost always prose the parser should not have touched.
- It begins with one of: `this`, `that`, `these`, `those`, `it`, `they`, `we`, `there`. These pronouns never refer to a specific entity.

**Object rejected if:**
- It is empty or shorter than two characters.
- It is a stopword: `a`, `an`, `the`, `it`, `this`, `that`, `and`, `or`, `but`, `so`, `if`, `as`.

A rejected subject or object means the whole line produces no claim. This is intentional. The parser would rather produce zero claims than a claim that cannot be verified.

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

Note what the third claim extracted. The subject `Refund amount`, the predicate `exceed`, and the object `captured amount`. The parser did not know what `exceed` means as a predicate — it recorded the verb as-is. That claim will remain `UNVERIFIED` after verification, because only `CALLS` predicates have a verification path. Recording it is still useful: the claim is searchable, and a reader can find every document that mentions `captured amount`.

---

## Troubleshooting: why did my document produce no claims?

Five reasons, in order of likelihood.

**1. The document has no bullet lines.** Prose paragraphs are not scanned. If the assertions you want to check are written as sentences, convert them to bullets.

**2. The bullet lines do not contain a modal verb.** `RefundHandler calls IdempotencyStore` produces nothing. `RefundHandler MUST call IdempotencyStore` produces a claim.

**3. The bullets are in a section that is not normative.** Bullets in `## Introduction` or `## Notes` are skipped. Move them to a section whose heading contains one of the normative keywords.

**4. The subject was rejected.** A subject with a comma, more than six words, or a leading pronoun is discarded. Break the sentence into a cleaner noun phrase.

**5. The repository the claim refers to is not in the workspace.** A claim about `SQLModel` will not verify unless the workspace contains the repository that defines `SQLModel`. Ingest the repository first.

If none of those apply and the claim still does not appear, run `garuda docs ingest <file>` on a single file and check the output. The parser prints the exact count and the reason each line was or was not parsed.

---

## What you should know before ingesting

1. **Determinism.** The same document produces the same claims every time. This is not a heuristic parser that drifts.

2. **Completeness is not the goal.** A long document may produce three claims. That is correct. The parser extracts what it can verify, and ignores what it cannot.

3. **Verified answers can still be wrong.** A `SUPPORTED` claim means the static graph contains the referenced entity. It does not mean production behaves that way. A `CONTRADICTED` claim means the runtime observation conflicts with the static structure. Neither state is a substitute for reading the code.

4. **`UNVERIFIED` is not failure.** It means the claim exists but the verifier has not been able to check it. The most common reason is that the subject does not match any entity in the workspace. Fix the claim or analyze the repository it refers to.

5. **The graph is scoped to a workspace.** A claim about `SQLModel` only verifies if the workspace you ingest into contains the sqlmodel repository.
```

