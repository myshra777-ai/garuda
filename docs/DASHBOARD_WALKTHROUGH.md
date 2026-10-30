# Garuda Dashboard Walkthrough

> **Scope:** Dashboard UI in the current `main` line (v1.9.6 code), as served from `/dashboard`. This guide explains the visible controls, panels, their purpose, and when to use them. The dashboard is a workspace view over Garuda's index, evidence, policy, and Talon services; it is not a replacement for the CLI or MCP interface.

## 1. Before you begin

Start Garuda and sign in:

```bash
garuda dev
# Open http://localhost:8080/dashboard
```

Choose the workspace you want to inspect from the sidebar selector. Dashboard reads are scoped to that workspace and the current authenticated session.

**Live data vs. sample data:** the main Overview, Symbol Search, Architecture Map, Docs & Rules, Talon, Change Review, Evidence, Agents, and Decision Lineage flows use live workspace responses by default. Empty or failed requests stay empty/unavailable instead of being filled with unrelated example records. To preview the explicit sample datasets, open `/dashboard?demo=1`; the demo banner is intentional and sample values are not evidence about your repository. The Data & Logic Audit API also includes built-in illustrative formula and financial-audit scenarios alongside workspace-enriched criticality context. Treat those scenario rows as examples until linked to imported source evidence.

## 2. Global navigation and controls

| Control | What it does | Why / when to use it |
|---|---|---|
| Workspace selector | Switches the active workspace using the session-protected workspace list. | Choose it before inspecting a different product, service family, or test environment. |
| Overview | Returns to the workspace summary. | Establish the current index and policy posture before drilling into a subsystem. |
| Explore → Symbol search | Opens the searchable symbol index. | Find a function, type, package, or file without navigating the graph first. |
| Explore → Architecture map | Opens the graph explorer. | Inspect dependency direction and start a blast-radius investigation. |
| Enforce → Docs & rules | Opens document-to-code lineage. | Compare ADR/spec requirements with indexed implementation. |
| Enforce → Talon & risk | Opens Talon status, hooks, risk hubs, and guard results. | Check agent enforcement or investigate high-impact symbols. |
| Enforce → Change review | Opens the change-review workbench. | Evaluate a proposed change against architecture and policy constraints. |
| Verify → Evidence & ledger | Opens evidence records and provenance views. | Inspect what Garuda observed and what has been anchored. |
| Verify → Agents & sessions | Opens MCP session activity. | Confirm coding clients are reaching Garuda. |
| Verify → Data & logic audit | Opens formula/criticality and financial discrepancy panels. | Investigate competing business logic or numeric outputs. |
| Verify → Decision lineage | Opens architectural decisions and supersession history. | Find the decision behind a constraint or trace how it changed. |
| Global search | Searches the current workspace; press `/` to focus it. | Quickly locate a symbol or package. |
| Command palette (`⌘K` / `Ctrl+K`) | Opens the command/search palette. Arrow keys navigate, Enter selects, Escape closes. | Jump to a destination or search without browsing the sidebar. |
| Add document | Opens the document-ingestion dialog. | Add an ADR, specification, invariant, or policy rule to the connected document/code view. |
| Refresh | Re-fetches workspace data. | Use after analysis, sync, ingestion, a Talon run, or policy evaluation. |
| Theme switch | Toggles graphite dark and warm light themes; the choice is saved locally. | Choose the most comfortable reading environment. |
| API Spec | Opens the API description modal. | Check an endpoint contract when integrating a client. |
| Talon Contract | Opens the compiled workspace contract. | Inspect guardrails supplied to agents. |
| Sign out | Ends the dashboard session. | Use on shared or untrusted machines. |
| Mobile navigation | Opens or closes the sidebar on narrow screens. | Reach other surfaces when the full sidebar is collapsed. |

The sidebar groups destinations by workflow rather than presenting ten equal-weight menu items: **Explore** for understanding structure, **Enforce** for rules and change control, and **Verify** for evidence and operational state.

## 3. Overview — the workspace briefing

The Overview is a summary, not a second copy of every other surface.

### System Integrity & Architectural Pulse

The headline communicates workspace context and the latest status returned by the API. The decision chip reports the latest available policy posture; the ledger chip reports an anchor only when one exists. “Unavailable” is a connection/data state, not an ALLOW verdict.

**Review changes** opens Change Review.

### Indexed Topology

Shows repository, entity, package, relationship, and cross-repository counts, plus language distribution. Select the card to open Architecture Map. Use it to check whether a repository sync has populated the workspace.

### Epistemic & Policy Gate

Shows supported, unverified, and contradicted claim counts, plus active policy totals. Interpret the distribution in the context of the latest index and verification run. Select the card to open Docs & Rules.

### RFC 6962 Provenance Ledger

Shows the latest known ledger anchor and safeguarded transaction count. Select the card to open Evidence & Ledger. An absent anchor is shown as absent; it is not treated as a verified genesis block.

### Progressive-disclosure panels

- **Needs Attention, Policy Evaluations & Documentation Drift** expands recent policy evaluations, attention items, and document/code drift. Select an evaluation to open its review context.
- **Highest-Centrality Architectural Hubs & Indexed Repositories** expands high-fan-in symbols and indexed repository summaries. Select a hub to focus Architecture Map; use repository actions to inspect package topology or search its symbols.

Keep these panels collapsed during routine use. Expand them when investigating a signal or comparing repositories.

## 4. Explore — symbols and architecture

### Symbol search

Enter a symbol, package, path, or entity kind. Use entity-kind filters to narrow results. Select a result to inspect its source identity and open its architectural context. An empty result means the query returned no match; it is not filled with sample symbols unless demo mode is enabled.

### Architecture map

The graph is a focused exploration tool, not a requirement to display every entity in a large workspace at once.

| Control / panel | What it does | When to use it |
|---|---|---|
| Entity / Package / Repository / Full | Changes the graph aggregation level. | Start broad at Repository or Package; move to Entity after choosing a target. |
| Docs | Switches to the document/code graph. | Inspect normative documents and their code relationships together. |
| Edge filters | Show all edges or narrow to calls, imports/dependencies, document governance, or contradictions. | Focus on one relationship type while tracing a path. |
| Focus/search | Loads a neighborhood around a symbol or package. | Investigate one-hop/two-hop impact instead of rendering a workspace-wide hairball. |
| Zoom / Fit | Adjusts graph scale and recenters the current map. | Use Fit after changing focus or window size. |
| Fullscreen | Expands the graph canvas. | Use for presentations or detailed dependency inspection. |
| Entity inspector | Shows selected-node identity and available relationship/evidence context. | Select a node when you need source identity and reasons behind its graph position. |

Treat the graph as navigation and inspection. Interpret visible edges alongside their evidence/provenance and epistemic state; graph presence alone does not prove runtime behavior.

## 5. Enforce — documents, Talon, and change review

### Docs & Rules

The document-to-code view connects ingested documents, extracted normative components, active rules, and matching code symbols.

- **Document-type filters** narrow the list/graph to ADRs, specs, invariants, policies, or supported document types.
- **Document cards / graph nodes** select a document or component and update the lineage inspector.
- **Component matrix** lists extracted requirements and their linked code symbols.
- **Add document** opens the ingestion dialog.

The dialog accepts a title, type, repository-relative path, and Markdown content. Use normative language and backtick code identifiers where appropriate. Review extracted components after ingestion; storing a document does not by itself mean every statement has been verified against implementation.

### Talon & Risk

This surface is the operational view of the Garuda Talon contract and risk profile of the indexed workspace.

- **Hook status** shows whether supported agent/IDE/Git integration files are present at the server's configured repository root. It does not prove every external client is connected at that instant.
- **Risk hubs** rank high-centrality or high-impact symbols. Select a hub for source, caller counts, risk tier, and governing-rule context.
- **Pre-edit briefing** gathers architecture context for a selected target before an agent changes it.
- **Run guard / Talon check** requests an evaluation against the current workspace and configured guardrails.
- **Build matrix** evaluates configured platform/build profiles where supported.
- **Simulate violation / drift** is a diagnostic demonstration; a simulation is not a production finding.
- **Install/sync hooks** updates supported local integration files when the backend action is available. Verify the returned result and resulting files; do not assume success from a click.

Use this surface to check agent readiness and investigate high-impact components. For actual enforcement, keep the CI/Talon command and MCP configuration authoritative.

### Change Review

The workbench focuses a proposed change:

1. Choose the target symbol or change context.
2. Review affected callers/dependencies and governing policies.
3. Inspect the decision, explanation, and evidence returned by the evaluator.
4. Resolve blocks or review-required findings before merging.

A green visual badge is not a replacement for the recorded evaluation and its cryptographic verification.

## 6. Verify — evidence, agents, audits, and decisions

### Evidence & Ledger

The evidence list records available analysis, policy, and runtime evidence. The ledger view provides record/anchor context and verification controls. Filter by result class, then open an entry for the underlying decision/evidence details.

A Merkle proof verifies record integrity and inclusion relative to an anchored root. It does **not** prove the semantic truth of a source claim; interpret it alongside the evidence type and epistemic state.

### Agents & Sessions

The active panel shows current MCP sessions; the recent panel shows recent sessions where available. Use it after configuring Garuda MCP/Talon in a coding client to confirm that the server is receiving sessions and tool calls. An empty list is a legitimate state; a failed request is not evidence that no agent connected.

### Data & Logic Audit

The formula/criticality area combines live workspace-enriched criticality tiers with built-in illustrative formula and financial-audit scenarios returned by the current dashboard API. Treat baseline financial discrepancy cards as examples, not automatically detected findings from your selected repository. Production findings should be backed by imported source evidence and an explicit analysis result.

- **Team context** switches the weighting/advisory context used by the view.
- **Formula/variant cards** expose the API's current example calculations, impact, and human decision state; confirm that a variant is grounded in your workspace before treating it as a production finding.
- **Approve / freeze** controls record a human disposition when the backend accepts the action.
- **Criticality tiers** summarize risk class and expected review level.
- **Financial discrepancy cards** compare built-in scenario values, formula/source lineage, and variance until those inputs are backed by imported workspace evidence.
- **Simulate Custom Unit-Scale Check** creates a local demonstration scenario; it is not an imported financial finding.
- **Rulebook editor** lets reviewers inspect/edit the proposed base rulebook, copy it, and submit approval. Verify the returned status/hash before treating it as approved.

Use this surface when reconciling competing business formulas or reviewing high-impact logic changes. It is not a substitute for source-system reconciliation or a formal financial audit.

### Decision Lineage

Lists architectural decision records and their revision/supersession chains. Select a decision to inspect metadata, status, revisions, and ledger details. **Ingest New ADR** opens the document-ingestion flow.

## 7. Repository onboarding: current capability

The dashboard ingests ADRs, design specifications, invariants, and policy documents. **It does not currently provide a repository upload, Git URL import, clone, or repository-sync control in the browser.** No inactive “Upload repo” button is shown.

Register and synchronize repositories with the CLI:

```bash
# Register a repository in an existing workspace
garuda repo add <workspace-name> https://github.com/org/garuda-assassin

# Synchronize enabled repositories and analyze them into the workspace
garuda workspace sync <workspace-name>

# Inspect repositories registered in that workspace
garuda repo list <workspace-name>
```

For a local checkout, use its `file://` URL only in an environment where the Garuda CLI and database can access that path. After sync completes, refresh the dashboard and verify repository, entity, package, and cross-repository counts.

### What a production dashboard importer should include

A browser importer is feasible, but it requires a real ingestion lifecycle rather than running `git clone` inside an HTTP handler:

1. **Source selection:** Git URL or supported archive upload, branch/tag/ref, and destination workspace/repository.
2. **Access control:** authenticated workspace membership checks; credentials through a secret-safe flow, never persisted in URLs or logs.
3. **Isolation:** clone into a dedicated per-job sandbox; reject path traversal and unsafe archive entries; enforce disk, file-count, size, and time limits.
4. **Job lifecycle:** enqueue a cancellable job and report queued/cloning/analyzing/persisting/complete/failed states.
5. **Deterministic analysis:** pin the resolved commit SHA, run the same analyzers and persistence path as the CLI, and store commit/fingerprint/receipt with the result.
6. **Recovery/idempotency:** avoid duplicate entries, make retries safe, clean temporary checkouts, and preserve prior indexed state if a new run fails.
7. **Honest progress:** show commit identity, analyzed counts, warnings, and actionable errors; do not mark a repository synced until analysis and persistence have completed.

This PR deliberately does not add a visual-only repository uploader or pretend the document-ingestion endpoint can ingest source repositories.

## 8. Suggested first-time workflow

1. Create/select a workspace and add one repository with the CLI.
2. Run workspace sync and confirm repository/entity/package counts in Overview.
3. Ingest the ADRs/specifications/policies relevant to that codebase.
4. Inspect Docs & Rules to check whether important requirements map to code symbols.
5. Use Explore and Architecture Map to locate a high-centrality symbol.
6. Run a Talon pre-edit briefing or guard check before changing it.
7. Review the proposed change and inspect evaluation evidence.
8. Confirm the resulting decision/anchor in Evidence & Ledger.
9. Configure a coding client with Garuda MCP/Talon and confirm its session appears in Agents & Sessions.

## 9. Motion, accessibility, and interaction notes

The dashboard uses short transitions for navigation, hover and disclosure changes; it avoids perpetual motion and respects `prefers-reduced-motion`. Keyboard focus is visible. On narrow screens, the sidebar becomes off-canvas navigation with a scrim. The light theme remains available for users who prefer a brighter work surface.
