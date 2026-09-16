# Contributing to Garuda

Thank you for contributing to Garuda.

## Principles & Requirements

1. **License & Headers**: All contributions fall under Apache 2.0. Every source file must contain the standard header:
   ```go
   // Copyright 2026 Rohit Mishra
   // SPDX-License-Identifier: Apache-2.0
   //
   // Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

cat << 'EOF' > SECURITY.md
# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

# Contributing to Garuda

Garuda is developed in the open. Contributions are welcome — language parsers, developer tooling, documentation, tests, and bug fixes.

This document covers what you need to know to contribute. It is deliberately short. If something is missing, open an issue.

## Before you start

Open an issue describing what you intend to work on. This is not a gate — it is a chance to check that the work is in scope and that no one else is already on it. For small fixes (typos, one-line improvements), skip the issue and open the pull request directly.

For anything that changes a public interface — a CLI command, an MCP tool, a JSON response shape — start with an issue. Those changes affect downstream consumers and are worth discussing before the code is written.

## Development setup

Requires Go 1.26 or later. TypeScript parsing requires CGO and a C compiler on your `PATH`.

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
go build -o bin/garuda ./cmd/garuda
go build -o bin/garuda-mcp ./cmd/garuda-mcp

For database work, you need PostgreSQL 14 or later:

bash
docker run -d \
  --name garuda-dev-pg \
  -e POSTGRES_USER=garuda \
  -e POSTGRES_PASSWORD=garuda \
  -e POSTGRES_DB=garuda \
  -p 5432:5432 \
  postgres:16

export DATABASE_URL="postgres://garuda:garuda@localhost:5432/garuda?sslmode=disable"

for f in migrations/[0-9]*.sql; do
  psql "$DATABASE_URL" -f "$f"
done
Running the tests
bash
go test ./...
The MCP server has a spec-compliance test that runs against the compiled binary:

bash
go build -o bin/garuda-mcp ./cmd/garuda-mcp
python3 scripts/mcp_verify.py
Expected output: ═══ 18/18 checks passed ═══.

Before you open a pull request
Three checks, in order. If any of them fails, fix it before requesting review.

bash
gofmt -w .
go vet ./...
go build ./...
Then run the test suite. If your change touches the analyzer, run the benchmark suite too:

bash
./bin/garuda bench
What makes a good pull request
One change per PR. A PR that fixes a bug and refactors an unrelated file is two PRs. Small PRs get reviewed fast; large ones sit.

A commit message that says why, not what. The diff shows what changed. The commit message should say why it changed. Reference the issue if there is one.

Tests for behavior changes. If you fix a bug, add a test that fails before the fix and passes after. If you add a feature, add a test that exercises the new code path.

No fabricated claims in code or docs. If a metric is measured on a specific corpus, name the corpus. If a behavior is not verified, say so. The project's core discipline is that claims are backed by evidence. Apply it to your own contribution.

No new dependency without justification. Garuda has a small dependency surface. Adding a library should be a deliberate choice, not a convenience. If you need one, explain why in the PR description.

Review process
Every PR gets reviewed by a maintainer. Most reviews happen within a week. Small PRs are faster.

A maintainer may:

Approve and merge.

Request changes.

Ask for the change to be split into smaller PRs.

Decline the change if it is out of scope.

If your PR is declined, the reasoning will be in the review comments. You are welcome to open an issue to discuss further.

Documentation
Documentation changes are first-class contributions. If a doc is wrong, outdated, or missing, fixing it is a real contribution.

The following docs are maintained and reviewed:

README.md — what Garuda is and why it exists

PLAYBOOK.md — installation, operations, workflows

EVIDENCE.md — validation results and methodology

docs/POSITIONING.md — internal positioning reference

docs/WALKTHROUGH.md — visual tour of the interface

docs/DOCUMENT_FORMATS.md — document ingestion contract

docs/invariants.md — the invariant contract

ADRs in docs/adr/ are immutable. To change a decision, write a new ADR that supersedes the old one. Do not edit an existing ADR.

Scope
The following are in scope and welcome:

New language analyzers on the existing pipeline

Improvements to the Go analyzer

MCP tool additions

Policy predicate additions

Documentation improvements

Test coverage

Bug fixes

The following are out of scope and will be declined:

Changes that add a dependency for a feature already provided by the standard library

Changes that loosen the "evidence before confidence" discipline

Changes that add a claim to a public doc without a corresponding reproducible test

Rewrites of working code without a stated problem being solved

Style-only changes to files you are not otherwise touching

Code of conduct
Be direct, be honest, be respectful. Disagree with the idea, not the person. Attack claims with evidence, not with tone. If a conversation turns into an argument about intent rather than substance, take a break and come back.

The project does not have a formal code of conduct document. If you feel one is needed, open an issue.

Getting help
Questions: GitHub Discussions

Bugs: GitHub Issues

Security: see SECURITY.md

