# Garuda Walkthrough

**What the interface looks like, what each view shows, and how to open it.**

This document walks through the four surfaces you will use most often: the web dashboard, the interactive graph, the AI agent integration, and the CLI. It describes what each one shows and what to do with it.

Screenshots are deliberately sparse. The interface changes often enough that a screenshot is stale within weeks. Where a view is best seen rather than described, the doc tells you the exact command to open it and what to look for.

---

## The web dashboard

The dashboard is where you see everything about a workspace at once: repositories, entities, relationships, policies, documentation health, runtime state, and the Merkle ledger.

**Open it:**

```bash
garuda dev

Then visit http://localhost:8080/dashboard?workspace=<workspace-name> in your browser.

What you see
Sidebar. Workspace name, current repository count, entity count. Click it to switch workspaces.

Top bar. Search bar across every symbol, package, file, and repository in the workspace. Slash key focuses it from anywhere on the page.

Overview section. Five KPI cards across the top:

Repositories — how many codebases are in the workspace

Packages — architectural modules across all repositories

Entities — functions, structs, interfaces, methods

Relationships — the total count of typed edges between entities

Idempotency — how many transactions were deduplicated

Below that, four more cards:

Financial ROI (Saved) — estimated cost avoided by not re-exploring context

Active Policies — enforcement rules currently registered

Agent Peak (24h) — highest concurrent agent sessions in the last day

Drift Prevented — contract violations caught and quarantined

Languages panel. A horizontal bar showing codebase composition across all repositories. Hover any segment to see the exact percentage.

Policy Enforcement panel. Four cards — Blocked, Review, Warn, Allow — showing how many evaluations produced each outcome. Below the cards, a list of the ten most recent evaluations. Each row shows the decision badge, the policy title, the block height at which the decision was anchored, and the entities and claims involved. Click any row to open the evaluation drawer.

Documentation Claims and Knowledge Drift panels. Two cards side by side. The left shows documentation health: how many claims are supported, unverified, contradicted. The right shows drift in both directions — documentation that does not match code, and code that has no documentation.

Scanned repositories. A card per repository, with a language bar, commit, and last-analyzed timestamp. Click any repository to filter the package hierarchy to that codebase.

Trust strip. Three cards — Supported Claims, Unverified Claims, Contradictions. When runtime verification has not yet been attempted, the Unverified card says so explicitly rather than showing a zero.

Architecture explorer. Three large buttons — Repositories, Packages, Entities — that open the graph view at that level.

Architectural hubs. The five most-connected entities in the workspace, ranked by inbound edge count.

Needs attention. Quarantined runtime contradictions. Empty state says "Zero Active Violations" when there is nothing to show.

Recent evidence ledger. The three most recent runtime observations. Each row shows the trace, the operation, and the timestamp.

What to do with it
New to the workspace? Open the architecture explorer. Start at the repository level, drill down into packages, then into entities. This is the fastest way to understand an unfamiliar system.

Reviewing a change? Search for the changed symbol. Open its drawer. The upstream callers and downstream dependencies are listed with their evidence.

Checking drift? Look at the Knowledge Drift panel. It shows both directions of mismatch in a single view.

Auditing a decision? Open the Policy Enforcement panel, click an evaluation, click "Verify anchor." The drawer re-derives the Merkle inclusion proof.

The interactive graph
The graph is a force-directed visualization of the workspace. It opens from the dashboard or directly from the daemon.

Open it:

bash
garuda dev
Then visit http://localhost:8080/graph in your browser.

What it shows
Nodes are entities, repositories, or packages — depending on the level you are viewing. Edges are typed relationships between them.

Node size reflects the entity's centrality in the graph. A node with many inbound edges is larger. This makes architectural hubs visible at a glance: they dominate the layout.

Node color reflects the module or package the entity belongs to. Entities in the same package cluster together.

Edge style reflects the relationship type. Static edges are solid. Runtime contradictions appear as red dashed edges with a "VIOLATION" label.

How to navigate
Single-click any node to open its detail drawer.

Double-click a node to drill down. From the repository level, double-clicking a repository shows its packages. From the package level, double-clicking a package shows its entities.

Hover over any node to spotlight its connections. Nodes that are not connected dim. This is how you trace a specific entity's neighborhood.

Drag any node to reposition it. The layout is not fixed — you can pull a cluster apart to read it.

Zoom and pan with mouse wheel and drag. The controls in the top-right corner do the same, plus a Fit button that recenters everything.

The communities panel
On the right side of the graph view is a list of the modules and packages in the workspace, each with a checkbox and a colored dot matching the node color in the graph.

Uncheck a module to hide it. This is the fastest way to reduce a crowded graph to just the subsystem you care about. Click "Toggle All" to invert the selection.

What to do with it
Explore an unfamiliar system. Start at repository level. Drill into the repository that owns the subsystem you care about. Follow the edges.

Trace a specific symbol. Search for it by name, then hover to see what connects to it.

Find architectural hubs. Look for the largest nodes. Those are the entities with the most inbound edges — the parts of the system that everything depends on.

Spot runtime drift. Red dashed edges are contradicted relationships. They appear only when a runtime observation has conflicted with a static expectation.

The AI agent integration
Garuda speaks the Model Context Protocol. Any MCP-compatible client — Cursor, Claude Desktop, or a custom agent — can query the workspace directly.

Set it up:

Add this to your client's MCP configuration:

json
{
  "mcpServers": {
    "garuda": {
      "command": "/absolute/path/to/bin/garuda-mcp",
      "env": {
        "DATABASE_URL": "postgres://...",
        "GARUDA_WORKSPACE": "my-workspace"
      }
    }
  }
}
Restart the client. Ask it a question about your code.

What it feels like
Without Garuda, asking an agent "what calls HandleCharge" produces a best-effort answer based on whatever files the agent can find. The agent opens a file, greps for the name, reads what it finds, and answers from partial information. It will sometimes invent a caller that does not exist. It will miss callers that live in other packages.

With Garuda, the agent calls garuda.neighbors or garuda.find_entity and gets a compiler-resolved list. The answer names specific entities in specific packages. If there is a caller across a module boundary, it appears. If a caller does not exist, it does not appear.

The difference is visible immediately. The Garuda answer is specific. The guess is hedged.

The sixteen tools
The agent has access to sixteen tools grouped by what they answer:

Structure questions — garuda.neighbors, garuda.subclasses, garuda.implementers, garuda.find_entity, garuda.entities, garuda.inspect.

Governance questions — garuda.policy.list, garuda.policy.evaluate, garuda.governance.status.

Documentation questions — garuda.check_drift, garuda.query_claims.

Decision questions — garuda.query, garuda.get_lineage, garuda.get_impact, garuda.detect_contradictions, garuda.propose_decision.

The agent picks the right one based on the question. You do not need to know the tool names.

Verify the connection
bash
WORKSPACE=my-workspace python3 scripts/mcp_verify.py
Expected output:

text
═══ 18/18 checks passed ═══
If any check fails, the script names it. The most common failure is a wrong binary path in the MCP config.

The IDE extension
The Garuda VS Code / Cursor extension surfaces the same information inline, where you are already looking at code.

Install it:

Search for "Garuda" in the Extensions panel. Or build it from source:

bash
cd vscode-extension
npm install
npm run compile
Then copy the extension folder to ~/.vscode/extensions/.

Inline contradictions
When a runtime observation contradicts a static expectation for a symbol in an open file, the file line is marked with a red squiggle. The Problems panel shows the contradiction with its severity (ARCH_DRIFT_001 for runtime drift).

<p align="center"> <img src="../assets/screenshots/ide-contradictions.png" alt="Garuda inline contradiction detection" width="800" /> </p>
Blast-radius hover
Hover over any function, method, struct, or interface. The hover shows upstream callers, downstream dependencies, and a link to open the symbol in the graph visualizer.

<p align="center"> <img src="../assets/screenshots/ide-graph-view.png" alt="Garuda blast-radius hover context" width="800" /> </p>
Status bar
The status bar shows the current Merkle block height and the count of unresolved contradictions. Both update automatically as the daemon processes new observations.

The CLI
The CLI is where Garuda is most useful for scripting and CI, and where the deepest commands live. The full reference is in the Playbook. Three commands are worth knowing on day one.

garuda analyze
Point it at a directory, it extracts the semantic model and saves it to the workspace.

bash
garuda analyze /path/to/repo --save --workspace my-workspace
garuda impact
Given a symbol name, reports every caller — direct and transitive — across the workspace.

bash
garuda impact HandleCharge
garuda policy evaluate
Runs every policy in a directory against the current workspace. Every evaluation produces one of four outcomes and is anchored to the Merkle ledger.

bash
garuda policy evaluate ./policies --workspace my-workspace
Every evaluation gets an ID. Verify any one of them:

bash
garuda policy verify <evaluation-id>
That command re-derives the Merkle inclusion proof and confirms the evaluation was committed at a specific block height. Anyone with the source can run the same command and get the same answer.

Where to go next
README — what Garuda is, and why it exists

PLAYBOOK — installation, operations, and every command

EVIDENCE — validation results and methodology

DOCUMENT_FORMATS — what documents Garuda reads, and how to format them

invariants.md — the invariant contract Garuda's truth substrate is built on

text

---
