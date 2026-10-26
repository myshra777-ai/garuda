# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

Garuda processes AST structures, dependency graphs, and code claims. If you discover a security vulnerability or ingestion isolation issue:

1. **Do not create a public issue.**
2. Send report details and reproduction steps to `security@garuda-engine.dev` or directly to Rohit Mishra.
3. You will receive an acknowledgment within 48 hours with an estimated remediation timeline.

# Security Policy

## Reporting a vulnerability

Report security issues privately. **Do not open a public GitHub issue.**

**Email:** security@garuda.dev

Include:

- A description of the issue and its impact
- Steps to reproduce, or a proof of concept
- The version or commit hash where you observed it
- Your name and contact, if you want credit
- Whether you intend to publish, and on what timeline

You will receive an acknowledgement within 72 hours. If the issue is valid, we will:

1. Confirm the affected versions.
2. Develop and test a fix.
3. Release the fix and credit you (unless you prefer otherwise).
4. Publish a short advisory describing the issue and the fix.

If the issue is not valid, we will explain why and close the report.

## Supported versions

Security fixes are backported to the current minor release and the previous one. Anything older receives fixes only if the issue is critical and the fix is small.

| Version | Supported |
| :--- | :--- |
| v0.1.x | Yes |
| Older | Critical issues only |

## Scope

In scope:

- The Garuda API server (`cmd/garuda-api/`)
- The Garuda CLI and daemon (`cmd/garuda/`, `internal/api/`)
- The MCP server (`cmd/garuda-mcp/`)
- The Merkle trust layer (`internal/merkle/`)
- The policy engine (`internal/policy/`)
- The PostgreSQL schema and migrations (`migrations/`)

Out of scope:

- Findings that require an attacker to already have database credentials
- Findings that require an attacker to already have filesystem access to the host
- Denial of service from unbounded resource consumption without authentication
- Missing rate limits on endpoints that already require authentication
- Findings in third-party dependencies where the maintainer has already published an advisory

## Security model

The following properties are part of the design. Findings that show any of them is violated are in scope and treated as high severity.

**Authentication.** Mutation endpoints are guarded by Ed25519 JWTs. The signing key is loaded from `JWT_PRIVATE_KEY_HEX`. In production, the process refuses to start if the key is not set.

**Integrity.** Every governance decision is anchored to an RFC 6962 Merkle tree with a canonical, language-independent byte encoding. The inclusion proof is a pure function of the leaf hash, the proof path, and the epoch root. It does not depend on Garuda's database, servers, or keys.

**Idempotency.** Decision commits pass through `idempotency_keys`. A duplicate submission is detected and rejected before it is written.

**Isolation.** Every tenant-scoped query filters by `tenant_id`. A workspace belonging to one tenant is not visible to another.

**Non-root runtime.** The API binary runs in a scratch container as a non-root user (`65532:65532`).

## What is not a vulnerability

The following are known limitations, documented so they are not reported as findings:

- The Merkle ledger is anchored locally. Independent verification requires the source of `internal/merkle/`, not an external timestamp authority or a public ledger. If you want external anchoring, the source is available for you to integrate.
- Call edges at the leaf are sparse (median 1, 90th percentile 4). This is a property of the code under analysis, not a defect in the analyzer.
- The Python and TypeScript analyzers are structural and have not been validated against a real cross-language workspace. Claims from these analyzers are labelled at the tier they were derived at.
- Runtime verification correlates static structure with OpenTelemetry spans. The correlation path is exercised against synthetic spans. Production correlation is in beta.

## Credit

We credit reporters in the release notes and the advisory, unless they prefer to remain anonymous.

## History

No advisories have been published as of 2026-09-15.