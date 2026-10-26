# Control Plane — design

**Status:** agreed 2026-09-17. Not yet implemented.
**Scope:** owner-facing platform management panel.
**Depends on:** `mcp-sessions.md` (Operations tab), `control-plane-metrics.md`
(Business and Tenants tabs).
**Blocks:** nothing.

---

## Non-goals

This spec does not build impersonation, feature flags, A/B experiments,
or the beta tester list. Those are queued. This spec builds the shell
they will eventually live in. It also does not build the tenant-facing
Workspace Console — that is `tenant-tabs.md`.

---

## Purpose

The Control Plane answers three questions no customer asks:

1. How is the business doing?
2. What is each tenant doing?
3. What is broken right now?

It is a distinct product from the tenant Workspace Console. Different URL,
different auth, different data scope, different layout. A user of one must
not be able to reach the other by guessing a URL, and the existence of the
Control Plane must not be advertised to a tenant.

---

## Naming and routing

| Layer | Term | URL |
| :--- | :--- | :--- |
| Tenant-facing | Workspace Console | `/dashboard` |
| Owner-facing | Control Plane | `/_/control` |

`/_/control` is deliberately non-obvious. It is not a word, it is not a
common path fragment, and it does not appear in any tenant-facing
documentation, log message, or error response.

**Unauthenticated requests return 404, not 401.** A 401 response confirms
the endpoint exists and requires a credential. A 404 response confirms
nothing. The distinction is small but real: an attacker scanning a
deployment sees `/_/control` and `/_/random` behave identically.

The application router must install the Control Plane handler **before**
any catch-all or SPA fallback handler. If the SPA fallback runs first,
`/_/control` returns the tenant dashboard HTML — the opposite of what is
intended. This ordering requirement is not a convention; it is a bug
waiting to happen if the router changes.

---

## Auth model

**Bearer token.** Two mechanisms:

- **Primary:** `GARUDA_CONTROL_TOKEN` environment variable on the server.
  Every request to `/_/control/*` must include
  `Authorization: Bearer <token>`.
- **Browser fallback:** `/_/control/login` accepts the token in a form and
  sets a 24-hour `HttpOnly` session cookie scoped to `/_/control`.

No user accounts, no passwords, no two-factor. If those are added later,
they slot in without changing the routing.

### Token lifecycle

**Generation.** The token is generated once per deployment, from a
cryptographically secure source (e.g., `openssl rand -base64 32`). It is
not derived from any other value.

**Rotation.** Rotating the token requires setting a new value in the
environment and restarting the API server. There is no dual-token window —
the moment the server sees the new value, the old token stops working.
Rotations are manual and documented in `PLAYBOOK.md`.

**Compromise.** If a token is suspected compromised, the response is:

1. Rotate the token on the server.
2. Review the `control_plane_access` audit log for unfamiliar source IPs.
3. Review `control_tenant_notes` and `impersonation_log` for unexpected
   entries.

The rotation invalidates every browser session cookie on the next request,
because cookie sessions are validated by re-checking the current token
value on each request. This is slower than a token-hash-based session but
simpler and correct.

### Rate limiting

Every failed auth attempt logs a `control-plane-auth-failure` event with
the source IP and the timestamp.

After 10 failures from the same IP in 15 minutes, all requests from that
IP are rejected for 30 minutes with a 404 response. The rejection is
in-memory; a server restart clears it.

**Why 404 on rate-limited requests, not 429.** A 429 confirms the endpoint
exists and that the rate limiter fired. A 404 confirms nothing. Same
reasoning as the unauthenticated case.

---

## Audit trail

Every request to `/_/control/*` writes a row to a `control_plane_access`
table:

```sql
CREATE TABLE control_plane_access (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  path         TEXT NOT NULL,
  method       TEXT NOT NULL,
  status       INT NOT NULL,
  source_ip    INET,
  user_agent   TEXT,
  auth_result  TEXT NOT NULL CHECK (auth_result IN ('ok', 'fail', 'rate_limited')),
  occurred_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX control_plane_access_recent_idx
  ON control_plane_access (occurred_at DESC);

CREATE INDEX control_plane_access_auth_idx
  ON control_plane_access (auth_result, occurred_at DESC)
  WHERE auth_result != 'ok';
```
Retention: 90 days. Pruned by the same nightly job that prunes
mcp_sessions.

This table is the answer to "who accessed the Control Plane and when."
Without it, a rotation or an impersonation leaves no trail beyond the
server's access log, which is not durable.

What gets logged, and what does not
Every request writes a row except read-only GETs to the aggregate
metric endpoints (/_/control/api/*). A single Business tab load makes
six metric fetches and writes zero rows. A login writes one row. A
failed auth writes one row. A mutation (write note, trigger backup)
writes one row.

Concretely, a row is written when:

The path is /_/control/login, /_/control/logout, or
/_/control/tenants/:id/notes.

The method is POST, PUT, PATCH, or DELETE.

The auth_result is fail or rate_limited.

The request is the first authenticated request from a new cookie
session (identified by the Set-Cookie header being present).

Reads are frequent and boring. Writes and auth events are the interesting
ones. Logging the former drowns the latter.

URL structure
text
/_/control                    default tab (business)
/_/control?tab=business       business tab
/_/control?tab=tenants        tenants tab
/_/control?tab=operations     operations tab
/_/control?tab=internal       internal tab
/_/control/tenants/:id        tenant detail
/_/control/login              token login form
/_/control/logout             clears the session cookie
Tabs are deep-linkable. Browser back and forward work. The active tab is
preserved across page reloads.

Tabs
Business
The investor-facing view. Growth, adoption, and usage metrics.

Panels:

Growth row — tenants, new tenants, workspaces, users, sessions

Adoption row — DAU / WAU / MAU, retention chart, activation

Usage row — entities, policies, decisions, tool calls, client breakdown

Revenue row — locked card, no billing yet

Every metric has a hover tooltip with its formula. See
control-plane-metrics.md §Business for definitions.

Tenants
Per-customer drill-down.

List view: a table of all tenants with the columns defined in
control-plane-metrics.md §Tenants tab. Sortable. Filterable by health
color.

Detail view (/_/control/tenants/:id):

Overview — name, tenant ID, created, plan (when plans exist)

Workspaces — list with names and entity counts

Recent activity — the tenant's last 20 sessions

Growth curve — the tenant's own entities-over-time series, using
repositories.last_analyzed_at as the time axis

Errors — recent tool-call errors scoped to this tenant

Actions — impersonate (queued, see below), add note, disable

Notes. The detail view includes a free-text notes field. Notes are
stored in control_tenant_notes:

```sql
CREATE TABLE control_tenant_notes (
  tenant_id   UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
  body        TEXT NOT NULL DEFAULT '',
  updated_by  TEXT NOT NULL,
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```
Notes are never joined by any tenant-scoped query. The endpoint that
renders notes is /_/control/* only. A tenant cannot see that notes
about them exist.

Why one row per tenant. The alternative — an append-only history of
note versions — is a feature, not a schema choice. The current design
says "there is a notes field." If history is wanted later, it becomes a
second table (control_tenant_note_versions) that references this one,
not a schema change here. Starting with one row means every query has
exactly one answer, and the render logic is trivial. Starting with many
rows means every read path must decide which one to display, and that
decision leaks into every future change.

Operations
Platform health. This is the tab the owner opens when something feels
wrong.

Panels:

Uptime and DB size

Query latency percentiles

Error rate and recent errors (blocked on errors_log)

Stalled MCP sessions

MCP tools with error rates above 5%

See control-plane-metrics.md §Operations for formulas.

Internal
Founder-only slice. Built last. Contains:

Roadmap view — a read-only render of docs/ROADMAP.md

Feature flags — per-tenant overrides (queued, see below)

Beta tester list — waitlist → invited → active → churned

Backup trigger — a button that writes a database dump to a configured
path

None of these are built in Arc C. They are described here so the tab
exists as a placeholder with a "Coming soon" body, which is cheaper than
adding the tab later.

Access control rules
Control Plane endpoints never return tenant-scoped workspace data. They
return aggregate counts and per-tenant metadata only.

A tenant's workspace contents — entity names, file paths, claim text —
never appear in the Control Plane.

control_tenant_notes is never joined by any tenant-scoped query.
Enforced at the endpoint level: the notes query lives in a package that
does not import the tenant handlers, so a future change cannot
accidentally reuse it from the tenant side.

No tenant ID is ever leaked in a URL the tenant can see. The Workspace
Console's HTML, CSS, and JavaScript contain no reference to
/_/control.

What "aggregate counts and per-tenant metadata" means precisely
Allowed:

COUNT(*) per tenant

Tenant name, tenant ID, creation timestamp

Session counts, tool-call counts

Policy counts, decision counts

Last activity timestamp

Not allowed:

Entity names

File paths

Claim text

Policy titles (the title may contain business context)

Error messages from tool calls (may contain file paths)

Anything from args_summary other than the presence/absence of a key

Data retention
Table	Retention	Pruned by
mcp_sessions	90 days	Nightly job
mcp_tool_calls	30 days	Nightly job
control_plane_access	90 days	Nightly job
control_tenant_notes	forever	—
impersonation_log	forever	—
The nightly job runs at 03:00 UTC. It is the same goroutine described in
mcp-sessions.md §Retention, extended to prune control_plane_access.

Queued items
Not built in Arc A, B, or C. Design is recorded here so it does not drift.

Impersonation
Open a tenant workspace as them, for debugging.

Requirements:

Available only from the Control Plane, not the Workspace Console.

Requires the owner token, not a tenant session.

The session cookie set by impersonation is tagged
impersonated_by=<owner_token_id>.

A persistent banner appears on every page of the tenant Workspace
Console during the impersonation: "You are viewing as <tenant-name>.
Every action is logged."

Every action during impersonation writes a row to impersonation_log:

```sql
CREATE TABLE impersonation_log (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL,
  owner_token_id  TEXT NOT NULL,
  action          TEXT NOT NULL,
  path            TEXT NOT NULL,
  occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```
Impersonation sessions expire after 30 minutes of inactivity.

Destructive actions are blocked during impersonation: delete workspace,
delete tenant, delete repository, remove policy.

Why this is queued, not built. Impersonation is a security-sensitive
feature. Building it before the audit log exists would mean a security
capability with no way to review its use. control_plane_access ships in
Arc C; impersonation_log is deferred to the arc that ships impersonation
itself.

Feature flags per tenant
Needed for staged rollouts. Rendered in the Internal tab. Requires a
feature_flags table with (tenant_id, flag_name, enabled) and a read
path in the tenant Workspace Console.

A/B experiment assignment
Per-tenant overrides for experiments. Requires the same table shape as
feature flags with an additional variant column.

Beta tester list
Waitlist → invited → active → churned status per person. Requires a
beta_testers table. Sensitive because it contains PII (email addresses).
Storage and access rules to be defined before the table is created.

What this is not
Not a customer-facing panel.

Not a per-workspace view. Everything here is cross-tenant.

Not a place to show tenant content. Aggregate counts and metadata only.

Not a source of financial claims until billing exists. A zeroed MRR
chart is worse than no chart.

Not a tenant administration tool. Tenant admins manage their own
workspace through the Workspace Console. The Control Plane is for the
platform owner.

Open questions
Should /_/control be served from the same process as /dashboard?
Current answer: yes, from the API server, on the same port, with
different routing. A separate binary would simplify process isolation
but complicate deployment. The routing separation is sufficient for
now.

Should the login form auto-logout after 24 hours of inactivity, or 24
hours from login? Current answer: 24 hours from login. An
inactivity-based session would require tracking per-request activity,
which the session cookie does not do. If inactivity-based expiry
becomes important, the cookie value becomes a token that is checked
against a server-side session table.

What happens on a fresh deployment with zero tenants? Current
answer: the Control Plane renders, all metrics show 0, the Tenants
table is empty with a message "No tenants yet." The Control Plane is
not hidden on a fresh deployment.

Fixture
Minimum state to verify the Definition of Done.

```sql
INSERT INTO tenants (id, name, created_at) VALUES
  ('11111111-1111-1111-1111-111111111111', 'fixture-control-tenant',
   NOW() - INTERVAL '15 days');

INSERT INTO workspaces (id, tenant_id, name) VALUES
  ('22222222-2222-2222-2222-222222222222',
   '11111111-1111-1111-1111-111111111111',
   'fixture-control-ws');

-- A control_tenant_notes row to verify idempotent writes
INSERT INTO control_tenant_notes (tenant_id, body, updated_by) VALUES
  ('11111111-1111-1111-1111-111111111111',
   'Initial fixture note.', 'owner');
```
The Control Plane reads through the garuda_control_ro role. The role
must be created in the same migration that creates control_plane_access:

```sql
CREATE ROLE garuda_control_ro WITH LOGIN PASSWORD '<generated>';
```
GRANT SELECT ON ALL TABLES IN SCHEMA public TO garuda_control_ro;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT ON TABLES TO garuda_control_ro;
The password is supplied via CONTROL_DATABASE_URL and is separate from
the tenant DATABASE_URL.

Definition of done
□ /_/control unauthenticated returns 404, not 401.
□ /_/control with a valid token renders the Business tab.
□ /_/control?tab=tenants with a valid token renders the Tenants tab
directly.
□ /_/control/login renders the token form; a valid token sets a
24-hour HttpOnly cookie scoped to /_/control.
□ An invalid token results in a logged control-plane-auth-failure
row in control_plane_access.
□ After 10 failed attempts from one IP in 15 minutes, subsequent
requests from that IP return 404 for 30 minutes.
□ Every request to /_/control/* writes exactly one row to
control_plane_access, except read-only GETs to
/_/control/api/*, which write zero.
□ /_/control/tenants/:id renders the six panels defined in §Tenants,
with empty states where data is absent.
□ Writing a note to a tenant twice results in a single row whose
body reflects the second write and whose updated_at is later
than the first.
□ Notes written on a tenant detail page persist and are not visible
from the tenant's Workspace Console.
□ The Internal tab renders with a "Coming soon" body.
□ The browser back button navigates between Control Plane tabs.
□ No string _control appears in the HTML, CSS, or JavaScript served
from /dashboard.
□ A query from the garuda_control_ro connection attempting an
INSERT fails with a permission error.
□ Restarting the API server with a new GARUDA_CONTROL_TOKEN
invalidates every existing cookie session on the next request.
Revision history
Date	Change
2026-09-17	Initial draft
2026-09-18	control_tenant_notes changed to one row per tenant, control_plane_access write rules narrowed to auth and mutation events, read-only role garuda_control_ro documented, indexes added
text
