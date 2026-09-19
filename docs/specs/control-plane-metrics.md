# Control Plane — metric definitions

**Status:** agreed 2026-09-17.
**Scope:** every metric shown on the Control Plane Business tab and the
Tenants tab table.

Every metric has a formula, a window, a refresh cadence, and an empty
state. Two operators looking at this dashboard must see the same number.
The rule that makes this possible is: **a metric without a formula is a
claim, and claims are not allowed here.**

---

## Non-goals

This spec does not define revenue tracking, churn prediction, or cohort
analysis beyond the retention curve. Metrics are historical counts with
formulas. They are not forecasts. Any metric that would require a model
to produce is out of scope.

---

## How to add a new metric

Every metric added to Business, Tenants, or Operations must declare all
four of:

1. A plain-English definition.
2. An exact SQL formula.
3. A refresh cadence.
4. An empty state (numeric zero, "not yet measured," or "not
   instrumented" — see §Empty state variants).

A metric missing any of the four is rejected in review, not fixed in
follow-up. The four-part rule is what makes two operators see the same
number.

---

## Data access model

The Control Plane reads from the same PostgreSQL database as the tenant
Workspace Console, using a **read-only role** (`garuda_control_ro`). The
role has `SELECT` on every table but no `INSERT`, `UPDATE`, or `DELETE`.

Every Control Plane query runs through this role. The application
connection for `/_/control/*` is separate from the tenant application
connection and is established from a different pool.

**Why a separate role rather than a scoped query.** A scoped query is a
convention enforced by the developer. A separate role is enforced by the
database. If a future query to the Control Plane accidentally joins a
tenant-scoped table in a way that would return tenant content, the role
does not help — but if a future code path attempts to write from the
Control Plane, the role stops it. The two protections are complementary;
the role is the cheap one.

**Why no replica.** A replica would require operational work (replication
setup, lag monitoring, failover) that does not pay for itself at current
volume. When the primary becomes the bottleneck, the Control Plane's
queries are the first to move.

---

## Refresh cadence

| Class | Cadence | Examples |
| :--- | :--- | :--- |
| Cheap aggregates | On page load | Tenants, workspaces, entities |
| Rolling window counts | 60-second cache | Sessions (7d), tool calls |
| Expensive scans | 5-minute cache | Client breakdown, retention curve |
| Live process state | On page load, no cache | Uptime, DB size |

The cache is in-memory in the API server. Keys are metric name plus
parameter tuple. The cache is cleared on server restart; no persistence
is needed because a cold read is cheap.

---

## Time ranges

Every metric that accepts a range accepts one of: 30, 60, or 90 days. The
default is 30.

A range of "30 days" means a rolling 30×24-hour window ending at the
moment of the query. It does **not** mean calendar days. This matters for
metrics like "new tenants" — a query at 14:00 on Tuesday covering "the
last 30 days" includes from 14:00 thirty days prior, not from midnight.

All timestamps in formulas are UTC. The dashboard displays UTC. There is
no user-local time because there is no user identity — the Control Plane
is accessed by the owner, from one machine, and UTC is what the logs use.

---

## Business tab — Growth

**Tenants**
- Formula: `SELECT COUNT(*) FROM tenants`
- Empty state: `0`
- Cadence: on page load

**New tenants**
- Formula: `SELECT COUNT(*) FROM tenants WHERE created_at >= NOW() - INTERVAL '<range> days'`
- Empty state: `0`

**Workspaces**
- Formula: `SELECT COUNT(*) FROM workspaces`
- Empty state: `0`

**Users**
- Formula: `SELECT COUNT(DISTINCT user_id) FROM tenant_members WHERE last_seen_at >= NOW() - INTERVAL '<range> days'`
- Empty state:
  - `0` if `last_seen_at` is populated for at least one row.
  - **"Not yet measured"** if every `last_seen_at` in `tenant_members` is NULL. The presence of the column is not the condition; the presence of at least one non-NULL value is.
- **Dependency note:** requires `tenant_members.last_seen_at` to be
  updated on every authenticated tenant request.

**Sessions**
- Formula: `SELECT COUNT(*) FROM mcp_sessions WHERE started_at >= NOW() - INTERVAL '<range> days' AND client_name != 'orphaned'`
- Empty state: `0`
- Cadence: 60-second cache

---

## Business tab — Adoption

**DAU / WAU / MAU**
- DAU: `SELECT COUNT(DISTINCT user_id) FROM tenant_members WHERE last_seen_at >= NOW() - INTERVAL '1 day'`
- WAU: same, 7 days
- MAU: same, 30 days
- Empty state:
  - `0` if `last_seen_at` is populated for at least one row.
  - **"Not yet measured"** if every `last_seen_at` in `tenant_members` is NULL. Same condition as Users above.

**Retention (cohort)**
- Formula: cohort tenants by signup week (`DATE_TRUNC('week', tenants.created_at)`). For each cohort, compute the percentage of tenants with at least one `mcp_sessions` row in week N after signup.
- Empty state: **the chart is hidden** until at least two complete cohorts exist. A retention curve with one point is a line, not a curve, and would be misread.
- Cadence: 5-minute cache

**Activation**
- Formula: `SELECT COUNT(*) FROM tenants t WHERE EXISTS (SELECT 1 FROM mcp_sessions s WHERE s.tenant_id = t.id AND s.started_at < t.created_at + INTERVAL '24 hours') / COUNT(*) FROM tenants`
- Empty state: **the metric is hidden** until at least 10 tenants exist. A percentage over fewer than 10 samples is not stable enough to be displayed.

**Paying**
- Status: **absent** until billing is instrumented.
- Rendered as a locked card: *"Billing not yet instrumented."*
- No zeroed chart, no `$0 MRR` placeholder. A number that reads as real but is fabricated is worse than an honest gap.

---

## Business tab — Usage

**Entities analyzed**
- Formula: `SELECT COUNT(*) FROM entities`
- Empty state: `0`

**Policies evaluated**
- Formula: `SELECT COUNT(*) FROM policy_evaluations`
- Empty state: `0`

**Decisions recorded**
- Formula: `SELECT COUNT(*) FROM decisions`
- Empty state: `0`

**MCP tool calls**
- Formula: `SELECT COUNT(*) FROM mcp_tool_calls`
- Empty state: `0`
- Cadence: 60-second cache

**Client breakdown**
- Formula: `SELECT client_name, COUNT(*) FROM mcp_sessions WHERE client_name != 'orphaned' GROUP BY client_name ORDER BY COUNT(*) DESC`
- Empty state: *"No sessions yet"* as the chart caption
- Cadence: 5-minute cache

---

## Business tab — Revenue

Not shown. Locked card: *"Billing not yet instrumented."*

---

## Tenants tab — table

| Column | Formula | Sort |
| :--- | :--- | :--- |
| Tenant | `tenants.name` | alphabetical |
| Workspaces | `SELECT COUNT(*) FROM workspaces WHERE tenant_id = t.id` | numeric |
| Users | `SELECT COUNT(*) FROM tenant_members WHERE tenant_id = t.id` | numeric |
| Entities | `SELECT COUNT(*) FROM entities WHERE tenant_id = t.id` | numeric |
| Sessions (7d) | `SELECT COUNT(*) FROM mcp_sessions WHERE tenant_id = t.id AND started_at >= NOW() - INTERVAL '7 days' AND client_name != 'orphaned'` | numeric |
| Policies | `SELECT COUNT(*) FROM policies WHERE tenant_id = t.id AND status = 'active'` | numeric |
| Last activity | `SELECT MAX(last_activity_at) FROM mcp_sessions WHERE tenant_id = t.id AND client_name != 'orphaned'` | descending |
| Health | derived — see below | by color, red first |

**Health derivation:**

- **green** — last activity within 7 days
- **amber** — last activity between 7 and 30 days
- **red** — last activity more than 30 days ago, **or** error rate over 5% in the last 24 hours

The error-rate clause requires a per-tenant error query:

```sql
SELECT
  COUNT(*) FILTER (WHERE c.status = 'error')::float / NULLIF(COUNT(*), 0)
FROM mcp_tool_calls c
JOIN mcp_sessions s ON s.id = c.session_id
WHERE s.tenant_id = $1
  AND s.client_name != 'orphaned'
  AND c.called_at >= NOW() - INTERVAL '24 hours'
If the result is NULL — no calls in the window — the health indicator
falls back to the time-based rule alone.

Operations tab — metrics
Uptime

Formula: NOW() - process_start_time

Empty state: 0s

Cadence: on page load, no cache

DB size

Formula: SELECT pg_database_size(current_database())

Empty state: 0 bytes (cannot occur in practice)

Average query latency

Formula: P50 / P95 / P99 of the API server's own query durations over the last 5 minutes, from an in-process ring buffer

Empty state: "Not yet measured"

Cadence: on page load

Error rate

Formula: errors / total requests over the last hour, from an in-process counter

Empty state: 0%

Recent errors

Formula: SELECT * FROM errors_log WHERE occurred_at >= NOW() - INTERVAL '24 hours' ORDER BY occurred_at DESC LIMIT 50

Dependency note: the errors_log table does not exist. This metric is blocked on a separate migration. Until that ships, the panel shows: "Error log not yet instrumented."

Empty state: "No errors in the last 24 hours"

Stalled sessions

Formula: SELECT * FROM mcp_sessions WHERE closed_at IS NULL AND last_activity_at < NOW() - INTERVAL '30 minutes' AND client_name != 'orphaned'

Empty state: "All sessions healthy"

MCP tool errors

Formula:

```sql
SELECT tool_name, COUNT(*) AS total,
       COUNT(*) FILTER (WHERE status = 'error') AS errors
  FROM mcp_tool_calls
 WHERE called_at >= NOW() - INTERVAL '1 hour'
 GROUP BY tool_name
HAVING COUNT(*) FILTER (WHERE status = 'error')::float
     / NULLIF(COUNT(*), 0) > 0.05
Empty state: "No tool errors"

Cadence: 60-second cache

Hover tooltips
Every metric displays its formula on hover. Example for "New tenants":

New tenants
Count of tenants created in the selected range.
SELECT COUNT(*) FROM tenants WHERE created_at >= NOW() - INTERVAL '30 days'
Updated every 5 minutes.

The tooltip contains three things: the plain-English definition, the
exact SQL, and the refresh cadence. The first is for readers, the second
for auditors, the third for anyone wondering why the number did not
change when they hit refresh.

Empty state variants

Kind	Rendering	Example
Numeric zero	The digit 0	Tenants on a fresh deployment
Not yet measured	Italic text in the metric slot	Latency before any query runs
Not instrumented	Greyed card with a lock icon	Billing, error log
A 0 is a real number. A "not yet measured" is an honest absence. A
"not instrumented" is a dependency that has not been built. Rendering
all three the same way is how dashboards lie.

What is not measured
Any metric requiring a data source the schema does not contain. The
dashboard shows "Not yet measured" rather than a fabricated zero.

Financial projections.

Any metric derived from free-form user input. The args_summary
whitelist ensures free-form text never reaches the observability tables,
and no Control Plane query reads it as if it had.

Per-human attribution beyond user_id. The Control Plane reports
tenants, sessions, and counts. It does not report people.

Fixture
Minimum state to verify the Definition of Done.

```sql
INSERT INTO tenants (id, name, created_at) VALUES
  ('11111111-1111-1111-1111-111111111111', 'fixture-tenant',
   NOW() - INTERVAL '20 days');

INSERT INTO workspaces (id, tenant_id, name) VALUES
  ('22222222-2222-2222-2222-222222222222',
   '11111111-1111-1111-1111-111111111111',
   'fixture-ws');

INSERT INTO tenant_members (tenant_id, user_id, last_seen_at) VALUES
  ('11111111-1111-1111-1111-111111111111',
   '33333333-3333-3333-3333-333333333333',
   NOW() - INTERVAL '2 hours');

INSERT INTO mcp_sessions (
  id, tenant_id, workspace_id, session_token,
  client_name, agent_id, started_at, last_activity_at
) VALUES
  ('44444444-4444-4444-4444-444444444444',
   '11111111-1111-1111-1111-111111111111',
   '22222222-2222-2222-2222-222222222222',
   'fixture-metrics-session-1',
   'cursor', 'fixture-agent-1',
   NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
  ('55555555-5555-5555-5555-555555555555',
   '11111111-1111-1111-1111-111111111111',
   '22222222-2222-2222-2222-222222222222',
   'fixture-metrics-session-2',
   'claude', 'fixture-agent-2',
   NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day');
```
Definition of done
□ Every metric listed above has a formula in a Go constant or a named
query function; no SQL string is inlined into a handler.
□ Every metric that returns a numeric value renders 0 correctly, not
NaN, not blank.
□ The "New tenants" tooltip shows both the plain-English definition and
the exact SQL, and the SQL matches the query that produced the number.
□ The "Paying" card renders with the lock icon and the text
"Billing not yet instrumented" on a fresh deployment.
□ The Retention chart is hidden, not zeroed, when fewer than two
complete cohorts exist.
□ The Activation metric is hidden, not zeroed, when fewer than 10
tenants exist.
□ Users, DAU, WAU, and MAU render "Not yet measured" when every
tenant_members.last_seen_at is NULL, and render 0 when at least
one row has a non-NULL last_seen_at but the window is empty.
□ The Tenants tab's Health column renders green / amber / red correctly
for the three time windows.
□ The Health column's red-on-error-rate path fires for a tenant whose
tool errors exceed 5% in the last 24 hours, with at least one tool
call in the window.
□ All queries run through the read-only role. Attempting an INSERT
from the Control Plane's connection fails with a permission error.
Revision history
Date	Change
2026-09-17	Initial draft
2026-09-18	Read-only role documented, three-state empty handling for Users/DAU/WAU/MAU, "How to add a new metric" section, client_name != 'orphaned' guards on all session queries
text

---
