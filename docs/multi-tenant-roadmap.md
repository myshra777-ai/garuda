# Multi-Tenant Roadmap

Status: design, awaiting implementation.
Scope: user identity, tenant membership, workspace membership, session model,
       admin telemetry, dual dashboards.
Depends on: Phase 4 (migration 076) — closed.
Supersedes: nothing. This is a new document.
Reference: internal/tenant, internal/auth, internal/api/auth_middleware.go,
           migrations/076_repositories_workspace_scoped_unique.sql.

---

## 1. Current state

What exists today, verified against the live database and the source tree.

### Real

- **Workspace namespacing.** Every semantic object carries `workspace_id`.
  Entities, claims, cross_repo_edges, document_claims, runtime_observations,
  workspace_modules. Queries filter by it.
- **A tenant column exists on workspaces.** `workspaces.tenant_id` is a UUID.
  Every query filters by it or by `workspace_id`.
- **Sessions.** JWT-based. Cookie holds a session ID; server holds a
  session map with user_id, email, role. Session TTL is enforced.
- **Password hashing.** bcrypt with constant-time comparison on the login
  path. The "wrong password" and "unknown email" paths are timed to match.
- **User registration exists at the service layer.** `AuthService.SignUp`
  creates a user, hashes the password, and issues a token. There is no
  HTTP route wired to it.
- **Bootstrap admin.** On first daemon start, `admin@local` is created
  with a random password, printed once to stderr.

### Not real

- **`dashboardTenantUUID` is a hardcoded constant** in
  `internal/api/dashboard_handlers.go`. Every dashboard request uses
  it, regardless of who signed in.
- **`authMiddleware` sets the tenant from the constant**, not from the
  session. See `internal/api/auth_middleware.go`.
- **No `tenants` table.** Tenant identity lives only as a UUID on
  `workspaces.tenant_id`. There is no name, no creation timestamp, no
  billing, no plan.
- **No membership tables.** There is no way to say "alice belongs to
  tenant X." There is no way to say "bob can see workspace Y."
- **No `/signup` route.** Registration exists as a function; the HTTP
  endpoint does not.
- **No admin dashboard.** The `/dashboard` route is the only dashboard.
- **No aggregate telemetry.** Counts of signups, active users, or
  workspaces must be computed by scanning the content tables.

### Consequence

The product is single-tenant in practice. Multiple users can sign in,
but they all see the same tenant. If two users create workspaces with
different names, both see both workspaces because the workspace
resolver is scoped by the constant, not by who is asking.

This is adequate for a founder-only deployment. It does not extend to
10 beta testers, and it does not extend to a customer with 20 teams.

---

## 2. Problem statement

The Company Brain category — per the YC Summer 2026 RFS — requires a
system that keeps a company's knowledge current and legible to AI
agents. Garuda's variant of that problem is narrower: keep a company's
*software structure* current and legible, with evidence, to both humans
and agents.

For that to be true for anyone other than the founder, four properties
must hold:

1. **A user sees only what they are a member of.** Not a filter on the
   query; a membership check that refuses by default.
2. **A tenant's data is not visible to another tenant's users.** Even
   by guessing a name.
3. **A team's workspace is not visible to another team in the same
   tenant.** Google's AI team and Google's Pixel team do not share.
4. **The operator can measure adoption without seeing content.**
   Counts of signups, active users, workspaces, agents, tokens saved.
   No prompts. No repository names. No entity names. No file paths.

Properties 1–3 are security. Property 4 is privacy. Both must be
structural — enforced by schema and middleware — not by policy or by
convention.

---

## 3. Design principles

These are the rules. Any implementation that violates one is wrong.

1. **Fail closed.** If the middleware cannot confirm membership, the
   request returns 404, not the requested data. Absence of evidence
   is not a grant of access.
2. **No tenant in the session.** The JWT carries `sub = user_id`. The
   tenant is derived per request from the workspace being accessed.
   This is what makes multi-tenant-per-user (a contractor working with
   two companies) work without a session switch.
3. **The admin dashboard reads only aggregates.** Structural, not
   policy. The admin dashboard's queries touch `telemetry_aggregates`
   and no other table with a `user_id`, `workspace_id`, or
   `repository_id` column.
4. **One middleware, every entry point.** HTTP, MCP, CLI. If a
   workspace-scoped operation bypasses the middleware, the isolation
   is fake.
5. **Team is a workspace.** Do not introduce a `teams` entity. If a
   team needs multiple workspaces, it creates multiple workspaces.
6. **Membership rows are facts, not claims.** A `workspace_members`
   row exists or it does not. There is no "pending" state in v1.
   Invitations that need acceptance are v1.1.
7. **No behavior change for the current deployment.** After every
   session below, the founder-only deployment works identically. The
   transition is invisible from the outside.
8. **Each session ends with a commit and a verification.** No session
   ships without a test that would fail if the session's change were
   reverted.

---

## 4. Data model

### 4.1 New table: `tenants`

```sql
CREATE TABLE tenants (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

No unique constraint on name. Two customers may legitimately want
the same tenant name; identity is the UUID.

4.2 New table: tenant_members
sql
CREATE TABLE tenant_members (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role        TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, tenant_id)
);
A user may belong to many tenants. A tenant has many users. The
composite primary key prevents duplicate membership.

Role semantics:

owner — created the tenant. Can delete it. Cannot be removed by
an admin.

admin — can invite, remove members, and manage tenant-wide
settings.

member — can access workspaces they are a member of.

4.3 New table: workspace_members
sql
CREATE TABLE workspace_members (
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    role            TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, workspace_id)
);
Role semantics at the workspace level mirror the tenant level. An
owner of a workspace can invite others to it. An admin can manage
members. A member can read and (in a later version) write.

No separate workspace_admins table. Roles live on the same row.

4.4 Modified table: users
sql
ALTER TABLE users ADD COLUMN last_login_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN is_disabled   BOOLEAN NOT NULL DEFAULT FALSE;
last_login_at is what makes active-user counts computable without
scanning session logs. is_disabled is what makes revocation possible
without deleting rows that audit records reference.

4.5 Modified table: workspaces
No schema change. tenant_id already exists. The current data has
every workspace under the canonical tenant UUID; the backfill in
Session A reassigns them.

4.6 New table: telemetry_aggregates
sql
CREATE TABLE telemetry_aggregates (
    bucket_date  DATE NOT NULL,
    tenant_id    UUID REFERENCES tenants(id) ON DELETE CASCADE,
    metric       TEXT NOT NULL,
    value        BIGINT NOT NULL DEFAULT 0,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_telemetry_aggregates_scoped
    ON telemetry_aggregates (bucket_date, tenant_id, metric)
    WHERE tenant_id IS NOT NULL;

CREATE UNIQUE INDEX idx_telemetry_aggregates_global
    ON telemetry_aggregates (bucket_date, metric)
    WHERE tenant_id IS NULL;
Two partial unique indexes because Postgres treats NULLs as distinct
in a unique index. A global bucket (tenant_id IS NULL) and a
tenant-scoped bucket (tenant_id IS NOT NULL) cannot collide.

The privacy boundary is this: the table has no column that can
identify a user, a repository, an entity, a file, or a prompt. The
admin dashboard cannot leak what the schema cannot express.

Metric keys used by the admin dashboard:

Metric	Meaning
signups	New users created that day
active_users	Distinct users who logged in that day
workspaces_created	New workspaces created that day
repositories_analyzed	Analyze runs completed that day
entities_extracted	Sum of entities persisted
agents_active	Distinct agent_runtime values seen that day
tokens_saved	Sum from telemetry_events
cost_saved_usd	Sum from telemetry_events
Adding a metric requires a schema review for the privacy boundary. If
a proposed metric cannot be computed from the existing content
without exposing identity, it is rejected.

5. Session model
5.1 JWT claims
text
sub = user_id  (UUID)
exp = expires_at
iat = issued_at
No tenant. No role. No email.

The reason: a user may belong to multiple tenants. Encoding one in the
JWT would require either a tenant-switch endpoint or accepting that
the session is scoped to one tenant and the user must log in again to
access another. Both are worse than deriving the tenant per request.

5.2 Session cookie
Unchanged from today. The cookie holds an opaque session ID; the
server holds a map[sessionID]{user_id, email, role, created_at}.
The JWT is derived from the session on each request for now; a later
change may cache it.

5.3 Middleware pipeline
Every request to a workspace-scoped endpoint runs:

text
1. Extract session cookie → sessionID
2. Look up session → user_id
3. Look up users row → refuse if missing or is_disabled
4. Parse ?workspace=<name> → look up workspaces row
5. Look up workspace_members (user_id, workspace_id) → refuse if missing
6. Set ctx: Session{UserID, WorkspaceID, TenantID, Role}
7. Handler runs, reads workspace_id from ctx, not from the query param
Step 5 returns 404, not 403. A 403 reveals that the workspace exists
but the user cannot access it. A 404 reveals nothing. That is the
correct privacy behavior.

5.4 Public routes
Routes that do not resolve a workspace do not run steps 4–5. These
are:

/login, /logout, /signup

/dashboard (the HTML shell; the JSON calls behind it go through
the full middleware)

/favicon.ico, /health, /metrics

The auth middleware's current bypass list is not a bypass list. Those
routes are simply not workspace-scoped, and the middleware's job is to
resolve a session for them if present, not to reject requests.

6. Roles and permissions
6.1 What each role can do
Action	Tenant owner	Tenant admin	Tenant member	Workspace owner	Workspace admin	Workspace member
View workspace	yes	yes	per-workspace	yes	yes	yes
Create workspace	yes	yes	yes (self-owned)	n/a	n/a	n/a
Invite to workspace	yes	yes	no	yes	yes	no
Remove from workspace	yes	yes	no	yes	yes	no
Delete workspace	yes	yes	no	yes	no	no
Invite to tenant	yes	yes	no	n/a	n/a	n/a
Delete tenant	yes	no	no	n/a	n/a	n/a
An invitation is two inserts: one into tenant_members (if the user
is not already in the tenant), one into workspace_members. Both go
through the same middleware check — the inviter must have the
corresponding role.

6.2 What this covers
The Google example:

Google is a tenant.

google-AI-team, google-pixel-team, google-product-team are three
workspaces in that tenant.

Alice is a member of tenant Google and a member of
google-AI-team only.

Bob is a member of tenant Google and a member of
google-pixel-team only.

Alice and Bob can each see that their tenant exists (they are
members of it), but neither can see the other's workspace.

Their AI agents inherit the same permissions because the agents
authenticate with the user's session. Alice's agent cannot query
google-pixel-team.

7. Signup and invitation flows
7.1 Signup
POST /signup with {email, password, full_name}.

Server:

Validate email format, password length (existing rules).

Refuse if the email already exists. Return generic error.

In one transaction:

Create users row.

Create tenants row named <email>'s tenant.

Create workspaces row named default.

Insert tenant_members (user, tenant, 'owner').

Insert workspace_members (user, workspace, 'owner').

Issue session cookie.

Redirect to /dashboard.

The new user is their own tenant with one workspace, member of both.

7.2 Invitation
POST /api/v1/workspaces/:id/members with {email, role}.

Server:

Middleware runs. Confirms the inviter is a workspace_members
row with role owner or admin.

Resolve email → user. If user does not exist, refuse. There is no
invitation-email flow in v1; the invitee must already be a Garuda
user.

In one transaction:

If user is not a member of the workspace's tenant, insert
tenant_members (user, tenant, 'member').

Insert workspace_members (user, workspace, role).

Both inserts are idempotent — ON CONFLICT DO NOTHING on the primary
key.

7.3 Removal
DELETE /api/v1/workspaces/:id/members/:user_id.

Server:

Middleware runs. Confirms the caller has owner or admin role
on the workspace.

Refuse to remove the last owner of a workspace.

Delete the workspace_members row.

Do not delete the tenant_members row automatically. The user
may still be a member of other workspaces in the same tenant.

8. Aggregate telemetry
8.1 Refresh job
A single function:

go
func RefreshAggregates(ctx context.Context, store *store.PostgresStore, date time.Time) error
Computes the day's aggregate rows and upserts them. Idempotent. Called:

On daemon start, for today and yesterday.

On a 15-minute ticker, for today.

The 15-minute freshness matters because a SaaS product that shows
yesterday's counts only is showing numbers nobody can act on.

8.2 What is aggregated, and where the boundary sits
Every aggregate value is computed from a source query that returns a
single number. No source query returns user, repository, entity, or
file identifiers. Concretely:

sql
-- signups
SELECT COUNT(*) FROM users
 WHERE created_at >= $1 AND created_at < $2
   AND ($3::uuid IS NULL OR tenant_id = $3);

-- active_users (24h or day bucket)
SELECT COUNT(DISTINCT user_id) FROM tenant_members
 WHERE user_id IN (
   SELECT id FROM users
    WHERE last_login_at >= $1 AND last_login_at < $2
 )
   AND ($3::uuid IS NULL OR tenant_id = $3);
The shape is always: one number out. No rows with columns.

8.3 The refresh job is the only writer
Only RefreshAggregates writes to telemetry_aggregates. No request
handler, no MCP tool, no CLI command writes there directly. This is
one place to review for the privacy boundary, not twenty.

9. The two dashboards
9.1 User dashboard: /dashboard
Unchanged in route, unchanged in HTML, changed only in data source.
Every query in HandleDashboardStats, HandleGraph,
HandleDashboardSearch, HandleDashboardPolicies, and the search
endpoint is scoped to the session's workspace_id from the
middleware. dashboardTenantUUID is deleted.

The user sees only their workspace. If they are a member of multiple
workspaces in the tenant, they can switch via a dropdown that lists
SELECT w.name FROM workspaces w JOIN workspace_members m ON m.workspace_id = w.id WHERE m.user_id = $1.

9.2 Admin dashboard: /admin
New route. Two guards, both required:

Session role is owner of a tenant whose ID is listed in
GARUDA_ADMIN_TENANT_ID env var, or the session's user is listed
in GARUDA_ADMIN_EMAILS.

Both env vars are read at process start. Empty means the admin
dashboard is disabled entirely. There is no way to reach it if it
is not configured.

Panels:

Panel	Source
Signups (24h, 7d, all-time)	telemetry_aggregates
Active users (24h, 7d)	telemetry_aggregates
Total users	telemetry_aggregates
Total tenants	SELECT COUNT(*) FROM tenants
Total workspaces	SELECT COUNT(*) FROM workspaces
Repositories analyzed (24h, 7d, all-time)	telemetry_aggregates
Entities extracted (24h, 7d, all-time)	telemetry_aggregates
Agents active (24h)	telemetry_aggregates
Tokens saved (all-time)	telemetry_aggregates
Cost saved USD (all-time)	telemetry_aggregates
The total tenants and total workspaces counts are the only
non-aggregate queries on the page. They are scalar counts of rows in
tables that do not carry content. If a stricter reading of the
privacy boundary wants those to also be pre-aggregated, that is a
one-hour change to add them to telemetry_aggregates.

Every panel whose underlying aggregate row does not exist renders
— with reason no aggregate row yet, per the MeasuredMetric
contract. Not 0.

9.3 What the admin dashboard never shows
User emails

Tenant names

Workspace names

Repository names

Entity names

File paths

Any prompt, any query string, any content

If a panel would show one of these, it does not go on the page.

10. Sessions
#	Session	Deliverable	Verification
0	Phase 4	Migration 076, committed.	repositories_tenant_workspace_name_uniq present. Probe insert of a duplicated name across two workspaces succeeds.
A	Schema	Migrations 077–081: tenants, tenant_members, workspace_members, users.last_login_at, users.is_disabled. Backfill: canonical tenant as one row, admin@local as owner, every existing workspace assigned to the canonical tenant.	SELECT COUNT(*) FROM tenant_members = 1. SELECT COUNT(*) FROM workspace_members = number of workspaces.
B	Signup	POST /signup endpoint. Creates user, tenant, workspace, two membership rows. No change to /login.	Sign up a second user, sign in, verify they have their own tenant and workspace. Founder login unchanged.
C	Middleware	Workspace resolution in middleware. resolveWorkspaceID → resolveWorkspaceForUser. All dashboard handlers read workspace from session. dashboardTenantUUID deleted.	Two users, two workspaces. User A asks for user B's workspace; response is 404.
D	Aggregates + Admin	telemetry_aggregates table. RefreshAggregates function + 15-min ticker. /admin route with the two guards.	Admin dashboard renders only aggregate panels. No content panel. Admin can see total users but not the user list.
E	Audit	Every SQL statement touching entities, claims, cross_repo_edges, document_claims, runtime_observations, repositories reviewed for workspace_id filter and membership precondition.	Grep-based audit script + manual review of every hit.
F	Onboard	Beta email to the 10 testers. Sign-ups open.	At least one external signup. Isolation probe from the previous section returns 404.
Session A ships three migrations in one commit. The migration sequence
must be applied in order. Rollback for each is a DROP TABLE or
DROP COLUMN — no data loss for anyone who has not signed up yet,
because the only row present is the founder's.

11. Rollback plan per session
Session	Rollback
A	DROP TABLE tenant_members, workspace_members, tenants CASCADE; ALTER TABLE users DROP COLUMN last_login_at, DROP COLUMN is_disabled;
B	Revert the endpoint commit. No schema change to roll back.
C	Revert the handler commit. The middleware change is one function; reverting restores the constant.
D	DROP TABLE telemetry_aggregates CASCADE; and revert the /admin route.
E	No code change. Revert any missed filter introduced by this session.
F	Revoke any tester accounts. The tenants they created are dropped with DELETE FROM tenants WHERE created_at > <date> CASCADE.
Every session is designed so a rollback does not require data
recovery. That is the reason the sequence is A → B → C → D rather than
shipping all four at once.

12. Isolation verification tests
These are the tests that must pass before Session F (onboarding).
They are the ones an external tester — or a founder with a public
profile — will run in the first hour.

Test 1 — Cross-tenant isolation
Two users, two tenants, two workspaces named tester-a-private and
tester-b-private.

bash
curl -s -b /tmp/tester-b-cookie.txt \
  "http://localhost:8080/api/v1/dashboard/stats?workspace=tester-a-private" \
  -o /dev/null -w "%{http_code}\n"
Expected: 404.

Test 2 — Cross-workspace isolation within a tenant
Tenant X. User Alice is a member of workspace ai-team. User Bob is a
member of workspace pixel-team in the same tenant. Bob asks for
ai-team.

bash
curl -s -b /tmp/bob-cookie.txt \
  "http://localhost:8080/api/v1/dashboard/stats?workspace=ai-team" \
  -o /dev/null -w "%{http_code}\n"
Expected: 404.

Test 3 — Membership is not name-based
Alice creates workspace my-workspace. Bob, in the same tenant,
creates a workspace with the same name. This must fail.

bash
curl -s -b /tmp/bob-cookie.txt -X POST \
  -d '{"name": "my-workspace"}' \
  "http://localhost:8080/api/v1/workspaces" \
  -o /dev/null -w "%{http_code}\n"
Expected: 409 or 400 with a message about duplicate name. The
UNIQUE (tenant_id, name) constraint on workspaces makes this
refuse. Which is correct — two workspaces in one tenant cannot share a
name. Two workspaces in different tenants can, and that is what
the tenant column enables.

Test 4 — Agent inherits user permissions
Alice's MCP server. Alice is a member of ai-team, not pixel-team.

bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"garuda.entities","arguments":{"workspace":"pixel-team"}}}' \
  | ./bin/garuda-mcp
Expected: JSON-RPC error, not a list of entities.

Test 5 — Admin dashboard has no content
Sign in as the admin user. Fetch /admin. Response HTML must not
contain any of: a user email, a tenant name, a workspace name, a
repository name, an entity name, a file path.

bash
curl -s -b /tmp/admin-cookie.txt "http://localhost:8080/admin" \
  | grep -Ei 'admin@local|go-validation-10|github\.com/|\.go:' || echo "clean"
Expected: clean.

Each test is written before the session that would make it pass. That
is, Test 1 is added as a test file in Session C, before C is
considered done. Tests 2–4 in Session C as well. Test 5 in Session D.

13. What this document does not cover
Billing and plans. No tenant_plan column, no Stripe integration,
no quota enforcement. When the first paying customer exists, a new
design document covers it.

SSO. Google/GitHub OAuth. Later.

Email verification on signup. For a beta with 10 invited
testers, manual verification is fine. For open signup, add a
verification step.

Password reset. Not in v1.

Audit log of membership changes. tenant_members and
workspace_members rows have created_at only. A later commit adds
a membership_events table if an auditor asks.

Per-workspace rate limits. Not in v1.

Cross-tenant collaboration. A user can belong to two tenants.
There is no shared workspace across tenants. That is intentional.

The teams entity. Every time it comes up, the answer is the
same: a team is a workspace. Two workspaces group with a naming
convention, not a foreign key. If a customer asks for team grouping
as a first-class concept, that is a v1.2 conversation with them.

14. After Session F
The immediate work after onboarding:

Watch the admin dashboard for the first week. Any panel that is
empty (—) with no aggregate row yet for more than a day is a
refresh job bug.

Ask two testers to attempt the isolation tests and report. Their
attempts are the external verification that internal tests cannot
provide.

If a tester asks for per-workspace invitations, that becomes the
next feature. If they ask for Slack/Jira/email ingestion, that is
a client-specific paid integration and out of scope.

The next roadmap document — after multi-tenant ships — is the one
that names what Garuda does with the identity it now has. Not
before.