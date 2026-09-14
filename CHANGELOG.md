## [Unreleased]

### Added

* **Multi-Tenancy & Workspace Governance**
  * Added database schemas for tenants and memberships, workspace resolution, and session-scoped tenant handling.
  * Added `/signup` endpoint to initialize user, tenant, and workspace resources.
  * Added dedicated admin dashboard at `/admin` and workspace picker sidebar component on the user dashboard.
  * Added `telemetry_aggregates` aggregation table and automated refresh job.

* **Model Context Protocol (MCP)**
  * Implemented stdio long-lived MCP server process (`Phase 2.1`).
  * Added read-only governance tools (`Phase 2.2a`), dry-run policy evaluation, and semantic analysis tools (`Phase 2.2b/2.2c`).
  * Added MCP invocation telemetry tracking tool executions per invocation.

* **Merkle Verification Ledger (ADR-0002)**
  * Added RFC 6962 Merkle tree implementation with canonical byte encoding and golden test vectors.
  * Added canonical encoders for genesis, decision, and policy evaluation artifacts.
  * Added tenant-scoped genesis generation, epoch leaves schema, and `EpochRoot` computation.
  * Added v1 write primitives, decision/policy write paths, and evidence attachments to graph edges.

* **Language Analysis & Semantic Features**
  * Added import resolution and class taxonomy infrastructure for multi-language analysis.
  * Enhanced graph views with entity degree calculations, 200-node caps, caching, and cross-repo merging.
  * Added support for type alias extraction, defined types, and `MeasuredMetric` contract for runtime metrics.
  * Added browser session authentication.

---

### Changed

* **Workspace & Tenant Isolation**
  * Enforced strictly scoped `workspace_id` constraints across CLI, knowledge readers, entity lookups, and database tables (`workspace_id NOT NULL` migration 082).
  * Removed `'default'` workspace fallback across all CLI commands and HTTP request handlers.
  * Scoped repository unique constraints to individual workspace contexts.
  * Replaced package path inference (`inferRepositoryFromPackage`) with explicit foreign key joins to `repositories`.

* **Dashboard & Metrics**
  * Segregated deployment-wide administration metrics from workspace-level user dashboards.

* **Tooling & Dependencies**
  * Upgraded project toolchain to Go 1.26 and promoted `tree-sitter` to a direct dependency.
  * Updated README and documentation regarding repository URL resolution and current capabilities.

---

### Fixed

* **Security & Multi-Tenant Data Leaks**
  * Resolved cross-tenant and cross-workspace data leakage vectors in `resolveWorkspaceID` and knowledge readers.
  * Fixed concurrent race condition when calling `CreateWorkspace` with duplicate names.
  * Cleaned up residue leaks in tenant isolation test fixtures.
  * Hardened dashboard authentication and session security.

* **AST Parsing & Semantic Graph Extraction**
  * Prevented self-referential `IMPLEMENTS` edges on interface definitions.
  * Restored `DEFINES`, `EMBEDS`, and `REFERENCES` extraction routines and corrected alias references.

* **Database & Migrations**
  * Fixed goose migration runner to execute strictly the `Up` migration section.
  * Corrected SQL statement in `SaveClaims` INSERT queries.
  * Fixed `cross_repo_edges` persistent writes and cross-repository bridge counts on the dashboard.

---

### Removed

* Removed dead internal polyglot AST parser (`internal/ast/polyglot.go`) and deprecated budget code.
* Removed suspended deployment manifests (`render.yaml` and `fly.toml`).
* Removed obsolete test suites (`resolve_workspace_id_test.go`).

---

### Semantic Changes

* **Epistemic Class Terminology**: Renamed `Unimplemented` status/class to `Unverified`.