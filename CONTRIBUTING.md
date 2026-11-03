# Contributing to Garuda

Garuda is developed in the open. Contributions are welcome: language parsers, developer tooling, documentation, tests and bug fixes.

This document is deliberately short. If something is missing, open an issue.

## Before you start

Open an issue describing what you intend to work on. This is not a gate; it is a chance to check that the work is in scope and that no one else is already on it. For small fixes (typos, one-line improvements), skip the issue and open the pull request directly.

For anything that changes a public interface (a CLI command, an MCP tool, a JSON response shape), start with an issue. Those changes affect downstream consumers and are worth discussing before the code is written.

## Licence and file headers

All contributions fall under Apache 2.0. Every source file starts with:

```go
// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
```

Use `#` instead of `//` in Python, shell and YAML files. Some older files carry an extra "Law Enforcement" comment line below the header; new files do not need it.

## Development setup

You need Go 1.26 or later. `make build` and the release binaries use `CGO_ENABLED=0`, so no C compiler is needed; TypeScript is then parsed by a pure-Go structural parser. To use the tree-sitter TypeScript parser, build with `CGO_ENABLED=1` and a C compiler on your `PATH` (see `internal/ast/treesitter`).

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
make build            # bin/garuda, bin/garuda-api, bin/garuda-mcp, bin/garuda-talon (CGO disabled)
```

Most commands, and everything that stores decisions, need PostgreSQL 13 or newer with the `pgcrypto` extension. The quickest route is the installer, which sets up a private PostgreSQL for you (see [INSTALL.md](INSTALL.md)):

```bash
./setup.sh            # Windows: .\setup.ps1
```

To use your own database instead:

```bash
docker run -d --name garuda-dev-pg \
  -e POSTGRES_USER=garuda -e POSTGRES_PASSWORD=garuda -e POSTGRES_DB=garuda \
  -p 127.0.0.1:5432:5432 postgres:16

export DATABASE_URL="postgres://garuda:garuda@127.0.0.1:5432/garuda?sslmode=disable"
```

The API server and the CLI apply any pending files in `migrations/` when they start, so run them from the repository root. That password is for a throwaway local container only.

## Running the tests

```bash
make test             # go test -v -race ./internal/... ./cmd/...
make benchmark-strict # strict ground-truth analyzer benchmark (also a CI gate)
```

The MCP server has a protocol check that runs against the compiled binary and prints `═══ N/N checks passed ═══`:

```bash
make build
python3 scripts/mcp_verify.py
```

## Before you open a pull request

Run these in order. If one fails, fix it before requesting review.

```bash
gofmt -l .            # prints nothing when formatting is clean
go vet ./...
go build ./...
make test
```

If your change touches documentation, manifests, version numbers, installers or workflows, also run the repository checks (CI runs the same ones):

```bash
python3 scripts/sync_docs_from_manifest.py --check   # versions and counts match capabilities.json
python3 scripts/check_doc_links.py                   # every relative Markdown link resolves
python3 scripts/check_marketplace_manifests.py       # plugin and registry files are tracked and consistent
python3 scripts/check_data_files.py                  # JSON and YAML files parse
```

`capabilities.json` is the single source of truth for the version and the tool, migration and skill counts. Change it first, then run `python3 scripts/sync_docs_from_manifest.py` to update the files that mirror it. To prepare a release, see [docs/LISTING_GUIDE.md](docs/LISTING_GUIDE.md#cutting-a-release).

## Sign your commits (Developer Certificate of Origin)

Every commit must carry a `Signed-off-by` line. It states that you wrote the change, or have the right to submit it, under the project's Apache-2.0 license (the [Developer Certificate of Origin](https://developercertificate.org/)).

```bash
git commit -s -m "fix: describe the change"      # adds: Signed-off-by: Your Name <you@example.com>
git commit --amend -s --no-edit                  # add it to the last commit
```

Use your real name and an email address you control. A pull request with unsigned commits will be asked to fix them before review.

## What makes a good pull request

- **One change per PR.** A PR that fixes a bug and refactors an unrelated file is two PRs. Small PRs get reviewed fast; large ones sit.
- **A commit message that says why, not what.** The diff shows what changed. The message should say why. Reference the issue if there is one.
- **Tests for behavior changes.** If you fix a bug, add a test that fails before the fix and passes after. If you add a feature, add a test that exercises the new code path.
- **No fabricated claims in code or docs.** If a metric is measured on a specific corpus, name the corpus. If a behavior is not verified, say so. The project's core discipline is that claims are backed by evidence. Apply it to your own contribution.
- **No new dependency without justification.** Garuda has a small dependency surface. If you need a library, explain why in the PR description.

## Review process

Every PR gets reviewed by a maintainer. Most reviews happen within a week. A maintainer may approve and merge, request changes, ask for the change to be split into smaller PRs, or decline the change if it is out of scope. If your PR is declined, the reasoning will be in the review comments. You are welcome to open an issue to discuss further.

## Documentation

Documentation changes are first-class contributions. If a doc is wrong, outdated or missing, fixing it is a real contribution. These are the maintained entry points:

- [README.md](README.md): what Garuda is and why it exists
- [INSTALL.md](INSTALL.md): installation on Windows, macOS, Linux, ChromeOS and Android
- [PLAYBOOK.md](PLAYBOOK.md): operations and workflows
- [EVIDENCE.md](EVIDENCE.md): validation results and methodology
- [docs/POSITIONING.md](docs/POSITIONING.md): positioning reference
- [docs/WALKTHROUGH.md](docs/WALKTHROUGH.md): visual tour of the interface
- [docs/DOCUMENT_FORMATS.md](docs/DOCUMENT_FORMATS.md): document ingestion contract
- [docs/invariants.md](docs/invariants.md): the invariant contract

Architecture decision records in [docs/adr/](docs/adr/) are immutable. To change a decision, write a new ADR that supersedes the old one. Do not edit an existing ADR.

## Scope

In scope and welcome:

- New language analyzers on the existing pipeline
- Improvements to the Go analyzer
- MCP tool additions
- Policy predicate additions
- Documentation improvements
- Test coverage
- Bug fixes

Out of scope and will be declined:

- Changes that add a dependency for a feature the standard library already provides
- Changes that loosen the "evidence before confidence" discipline
- Changes that add a claim to a public doc without a corresponding reproducible test
- Rewrites of working code without a stated problem being solved
- Style-only changes to files you are not otherwise touching

## Code of conduct

Be direct, be honest, be respectful. Disagree with the idea, not the person. Attack claims with evidence, not with tone. If a conversation turns into an argument about intent rather than substance, take a break and come back.

The project does not have a formal code of conduct document. If you feel one is needed, open an issue.

## Getting help

- Questions and bugs: [GitHub Issues](https://github.com/myshra777-ai/garuda/issues)
- Security: see [SECURITY.md](SECURITY.md)
