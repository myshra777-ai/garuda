// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/myshra777-ai/garuda/internal/store"
)

// adminEmailAllowed returns true iff the given email appears in the
// GARUDA_ADMIN_EMAILS environment variable.
//
// Empty GARUDA_ADMIN_EMAILS disables the admin dashboard entirely.
// There is no default. A deployment that has not opted in cannot
// reach /admin through any credential.
func adminEmailAllowed(email string) bool {
	list := os.Getenv("GARUDA_ADMIN_EMAILS")
	if list == "" {
		return false
	}
	email = strings.TrimSpace(strings.ToLower(email))
	for _, e := range strings.Split(list, ",") {
		if strings.TrimSpace(strings.ToLower(e)) == email {
			return true
		}
	}
	return false
}

// AdminDayRow is one day in the daily series shown on the admin
// dashboard.
type AdminDayRow struct {
	Date          string
	Signups       int64
	ActiveUsers   int64
	Workspaces    int64
	TelemetryRows int64
}

// AdminData is everything the admin template renders. Every field is
// a count or a scalar value. There is no field that could carry a
// user email, a tenant name, a workspace name, or any content.
type AdminData struct {
	// Current snapshot.
	UsersTotal      int64
	TenantsTotal    int64
	WorkspacesTotal int64

	// 24-hour and 7-day windows.
	SignupsToday    int64
	Signups7d       int64
	Signups30d      int64
	ActiveUsers24h  int64
	Workspaces7d    int64
	Tenants7d       int64
	TelemetryRows7d int64

	// All-time sums.
	TokensSavedAll    int64
	CostSavedUSDCents int64
	AgentsActiveToday int64

	// Daily series, most recent first.
	Daily []AdminDayRow

	// Diagnostic.
	LastRefreshed time.Time
	DeploymentEnv string
}

// HandleAdminDashboard renders the deploy-wide aggregate view.
//
// Access is gated by GARUDA_ADMIN_EMAILS. Empty env var means the
// route is disabled and returns 404. A signed-in user whose email is
// not in the allowlist gets the same 404, so the route's existence
// is not discoverable.
func (s *Server) HandleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	applySecurityHeaders(w)
	ctx := r.Context()

	sess, ok := SessionFromContext(ctx)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if !adminEmailAllowed(sess.UserEmail) {
		slog.Info("admin access denied", "email", sess.UserEmail, "remote", r.RemoteAddr)
		http.NotFound(w, r)
		return
	}

	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}

	data, err := buildAdminData(ctx, pgStore)
	if err != nil {
		slog.Error("admin dashboard build failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = adminTmpl.Execute(w, data)
}

// buildAdminData runs every query the admin dashboard needs. Every
// query reads telemetry_aggregates only. No query touches users,
// tenants, workspaces, entities, repositories, or any table that
// carries content. The privacy boundary is structural: if a future
// panel needs data this function cannot provide, the schema is the
// constraint, not the SQL.
func buildAdminData(ctx context.Context, pgStore *store.PostgresStore) (*AdminData, error) {
	d := &AdminData{}

	// Current snapshot — reads today's rows for the three snapshot
	// metrics. If the refresher has not run yet today, these are
	// zero; that is honest.
	readCurrent := func(metric string, into *int64) error {
		return pgStore.Pool().QueryRow(ctx, `
			SELECT value FROM telemetry_aggregates
			 WHERE bucket_date = CURRENT_DATE
			   AND tenant_id IS NULL
			   AND metric = $1
		`, metric).Scan(into)
	}
	if err := readCurrent(store.MetricUsersTotalCurrent, &d.UsersTotal); err != nil && !isNoRows(err) {
		return nil, fmt.Errorf("read users_total_current: %w", err)
	}
	if err := readCurrent(store.MetricTenantsTotalCurrent, &d.TenantsTotal); err != nil && !isNoRows(err) {
		return nil, fmt.Errorf("read tenants_total_current: %w", err)
	}
	if err := readCurrent(store.MetricWorkspacesTotalCurrent, &d.WorkspacesTotal); err != nil && !isNoRows(err) {
		return nil, fmt.Errorf("read workspaces_total_current: %w", err)
	}

	// Date-scoped sums. Every query is over the global scope
	// (tenant_id IS NULL) and a single metric.
	sumWindow := func(metric string, days int, into *int64) error {
		return pgStore.Pool().QueryRow(ctx, `
			SELECT COALESCE(SUM(value), 0)::bigint
			  FROM telemetry_aggregates
			 WHERE bucket_date >= CURRENT_DATE - $1::int
			   AND bucket_date <= CURRENT_DATE
			   AND tenant_id IS NULL
			   AND metric = $2
		`, days-1, metric).Scan(into)
	}
	if err := sumWindow(store.MetricSignups, 1, &d.SignupsToday); err != nil {
		return nil, fmt.Errorf("read signups today: %w", err)
	}
	if err := sumWindow(store.MetricSignups, 7, &d.Signups7d); err != nil {
		return nil, fmt.Errorf("read signups 7d: %w", err)
	}
	if err := sumWindow(store.MetricSignups, 30, &d.Signups30d); err != nil {
		return nil, fmt.Errorf("read signups 30d: %w", err)
	}
	if err := sumWindow(store.MetricWorkspacesCreated, 7, &d.Workspaces7d); err != nil {
		return nil, fmt.Errorf("read workspaces 7d: %w", err)
	}
	if err := sumWindow(store.MetricTenantsCreated, 7, &d.Tenants7d); err != nil {
		return nil, fmt.Errorf("read tenants 7d: %w", err)
	}
	if err := sumWindow(store.MetricTelemetryEventCount, 7, &d.TelemetryRows7d); err != nil {
		return nil, fmt.Errorf("read telemetry rows 7d: %w", err)
	}

	// "Today" active users. Different from the snapshot: reads
	// today's row for the day-scoped active_users_global metric.
	if err := readCurrent(store.MetricActiveUsersGlobal, &d.ActiveUsers24h); err != nil && !isNoRows(err) {
		return nil, fmt.Errorf("read active users: %w", err)
	}
	if err := readCurrent(store.MetricAgentsActive, &d.AgentsActiveToday); err != nil && !isNoRows(err) {
		return nil, fmt.Errorf("read agents active: %w", err)
	}

	// All-time sums. Sum across every bucket_date. These accumulate
	// and never decrease; if the numbers need to be monotonically
	// consistent, the read is the sum of every day's row.
	if err := sumWindow(store.MetricTokensSaved, 36500, &d.TokensSavedAll); err != nil {
		return nil, fmt.Errorf("read tokens saved all: %w", err)
	}
	if err := sumWindow(store.MetricCostSavedUSDCents, 36500, &d.CostSavedUSDCents); err != nil {
		return nil, fmt.Errorf("read cost saved all: %w", err)
	}

	// Daily series, last 14 days.
	rows, err := pgStore.Pool().Query(ctx, `
		SELECT bucket_date,
		       COALESCE(MAX(value) FILTER (WHERE metric = 'signups'), 0)::bigint,
		       COALESCE(MAX(value) FILTER (WHERE metric = 'active_users_global'), 0)::bigint,
		       COALESCE(MAX(value) FILTER (WHERE metric = 'workspaces_created'), 0)::bigint,
		       COALESCE(MAX(value) FILTER (WHERE metric = 'telemetry_event_count'), 0)::bigint
		  FROM telemetry_aggregates
		 WHERE bucket_date >= CURRENT_DATE - 14
		   AND bucket_date <= CURRENT_DATE
		   AND tenant_id IS NULL
		 GROUP BY bucket_date
		 ORDER BY bucket_date DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("read daily series: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var r AdminDayRow
		var t time.Time
		if err := rows.Scan(&t, &r.Signups, &r.ActiveUsers, &r.Workspaces, &r.TelemetryRows); err == nil {
			r.Date = t.Format("2006-01-02")
			d.Daily = append(d.Daily, r)
		}
	}

	d.LastRefreshed = time.Now().UTC()
	d.DeploymentEnv = os.Getenv("GARUDA_ENV")
	return d, nil
}

func isNoRows(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no rows in result set")
}

var adminTmpl = template.Must(template.New("admin").Parse(adminHTML))

const adminHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Garuda — Admin</title>
<style>
:root {
    --bg: #060810; --surface: #0e1424; --surface-2: #162035;
    --border: #1f2b45; --text: #f8fafc; --text-2: #94a3b8;
    --muted: #64748b; --brand: #38bdf8; --green: #34d399;
    --amber: #fbbf24; --red: #f43f5e;
}
* { box-sizing: border-box; }
body { margin: 0; background: var(--bg); color: var(--text);
       font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Inter, sans-serif; font-size: 13px; }
.topbar { height: 60px; padding: 0 32px; background: rgba(10, 14, 26, 0.9);
          border-bottom: 1px solid var(--border); display: flex; align-items: center; gap: 16px; }
.brand-mark { width: 32px; height: 32px; border-radius: 8px;
              background: linear-gradient(135deg, #0284c7, #38bdf8);
              display: grid; place-items: center; font-size: 16px; }
.brand-name { font-weight: 800; font-size: 14px; letter-spacing: 0.5px; }
.brand-sub { color: var(--muted); font-size: 10px; }
.badge { margin-left: auto; padding: 4px 10px; border-radius: 6px;
         font-size: 10px; font-weight: 800; text-transform: uppercase;
         background: rgba(56,189,248,0.12); color: var(--brand);
         border: 1px solid rgba(56,189,248,0.3); }
.content { max-width: 1200px; margin: 0 auto; padding: 32px; }
h1 { font-size: 22px; margin: 0 0 6px; }
.sub { color: var(--text-2); margin-bottom: 24px; }
.grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 14px; margin-bottom: 24px; }
.card { background: var(--surface); border: 1px solid var(--border); border-radius: 10px; padding: 16px; }
.card-label { color: var(--muted); font-size: 10px; font-weight: 750;
              text-transform: uppercase; letter-spacing: 0.06em; }
.card-value { font-size: 24px; font-weight: 800; margin-top: 6px; }
.card-sub { color: var(--muted); font-size: 11px; margin-top: 4px; }
table { width: 100%; border-collapse: collapse; margin-top: 8px; }
th, td { padding: 10px 12px; text-align: left; border-bottom: 1px solid var(--border); font-size: 12px; }
th { color: var(--muted); font-size: 10px; font-weight: 750;
     text-transform: uppercase; letter-spacing: 0.05em; }
td.num { text-align: right; font-variant-numeric: tabular-nums; }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.footer { margin-top: 32px; padding-top: 20px; border-top: 1px solid var(--border);
          color: var(--muted); font-size: 11px; }
</style>
</head>
<body>
<div class="topbar">
  <div class="brand-mark">🦅</div>
  <div>
    <div class="brand-name">GARUDA</div>
    <div class="brand-sub">Deployment administration</div>
  </div>
  <div class="badge">Aggregates only</div>
</div>
<div class="content">
  <h1>Deployment overview</h1>
  <div class="sub">Counts across the deployment. No user, workspace, or repository content is exposed here — the schema this page reads cannot express it.</div>

  <div class="grid">
    <div class="card">
      <div class="card-label">Users</div>
      <div class="card-value">{{ .UsersTotal }}</div>
      <div class="card-sub">registered</div>
    </div>
    <div class="card">
      <div class="card-label">Tenants</div>
      <div class="card-value">{{ .TenantsTotal }}</div>
      <div class="card-sub">active</div>
    </div>
    <div class="card">
      <div class="card-label">Workspaces</div>
      <div class="card-value">{{ .WorkspacesTotal }}</div>
      <div class="card-sub">created</div>
    </div>
    <div class="card">
      <div class="card-label">Active agents (24h)</div>
      <div class="card-value">{{ .AgentsActiveToday }}</div>
      <div class="card-sub">distinct agent runtimes</div>
    </div>
  </div>

  <div class="grid">
    <div class="card">
      <div class="card-label">Signups today</div>
      <div class="card-value">{{ .SignupsToday }}</div>
    </div>
    <div class="card">
      <div class="card-label">Signups 7d</div>
      <div class="card-value">{{ .Signups7d }}</div>
    </div>
    <div class="card">
      <div class="card-label">Workspaces 7d</div>
      <div class="card-value">{{ .Workspaces7d }}</div>
    </div>
    <div class="card">
      <div class="card-label">Telemetry rows 7d</div>
      <div class="card-value">{{ .TelemetryRows7d }}</div>
    </div>
  </div>

  <h2 style="font-size:15px; margin:24px 0 10px;">Last 14 days</h2>
  <table>
    <thead>
      <tr>
        <th>Date</th>
        <th class="num">Signups</th>
        <th class="num">Active users</th>
        <th class="num">Workspaces created</th>
        <th class="num">Telemetry events</th>
      </tr>
    </thead>
    <tbody>
      {{ range .Daily }}
      <tr>
        <td class="mono">{{ .Date }}</td>
        <td class="num">{{ .Signups }}</td>
        <td class="num">{{ .ActiveUsers }}</td>
        <td class="num">{{ .Workspaces }}</td>
        <td class="num">{{ .TelemetryRows }}</td>
      </tr>
      {{ else }}
      <tr><td colspan="5" style="color:var(--muted); text-align:center; padding:20px;">No aggregate rows yet. The refresher runs on daemon start and every 15 minutes.</td></tr>
      {{ end }}
    </tbody>
  </table>

  <div class="footer">
    Rendered {{ .LastRefreshed.Format "2006-01-02 15:04:05Z" }}.
    {{ if .DeploymentEnv }}Environment: {{ .DeploymentEnv }}.{{ end }}
    Source: telemetry_aggregates only.
  </div>
</div>
</body>
</html>`
