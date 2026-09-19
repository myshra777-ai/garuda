// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"time"
)

// TenantRow is one row in the Tenants tab table. Every column has a
// formula in control-plane-metrics.md §Tenants tab.
//
// Health is derived, not stored: green if the last MCP session
// activity is within 7 days, amber if within 30, red otherwise or if
// the tenant's error rate in the last 24 hours exceeds 5%.
type TenantRow struct {
	ID   string
	Name string

	Workspaces int
	Users      int
	Entities   int
	Sessions7d int
	Policies   int

	LastActivity     string // human-formatted for display, "—" if none
	LastActivityUnix int64  // 0 if none, used only for sorting

	Health       string // "green" | "amber" | "red"
	HealthReason string // tooltip text
}

// TenantsMetrics is the payload the Tenants tab renders.
//
// A failed query sets Error and leaves Rows empty. The template
// renders the error banner and nothing else. This is the same
// discipline as BusinessMetrics: a broken query must never render
// as a zero.
type TenantsMetrics struct {
	Rows  []TenantRow
	Total int
	Error string
}

// loadTenantsMetrics runs one query against the read-only Control
// Plane pool and derives health for each tenant. Reads only. Never
// writes to the tenant database.
func (s *Server) loadTenantsMetrics(ctx context.Context) TenantsMetrics {
	m := TenantsMetrics{Rows: []TenantRow{}}

	if s.controlPool == nil {
		m.Error = "Control Plane database pool is not configured."
		return m
	}

	qctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.controlPool.Query(qctx, `
		SELECT
		  t.id,
		  t.name,
		  (SELECT COUNT(*) FROM workspaces w
		     WHERE w.tenant_id = t.id),
		  (SELECT COUNT(*) FROM tenant_members tm
		     WHERE tm.tenant_id = t.id),
		  (SELECT COUNT(*) FROM entities e
		     WHERE e.tenant_id = t.id),
		  (SELECT COUNT(*) FROM mcp_sessions s
		     WHERE s.tenant_id = t.id
		       AND s.started_at >= NOW() - INTERVAL '7 days'
		       AND s.client_name != 'orphaned'),
		  (SELECT COUNT(*) FROM policies p
		     WHERE p.tenant_id = t.id AND p.status = 'active'),
		  (SELECT MAX(s.last_activity_at) FROM mcp_sessions s
		     WHERE s.tenant_id = t.id
		       AND s.client_name != 'orphaned'),
		  (SELECT
		     CASE
		       WHEN COUNT(*) = 0 THEN NULL
		       ELSE COUNT(*) FILTER (WHERE c.status = 'error')::float / COUNT(*)
		     END
		     FROM mcp_tool_calls c
		     JOIN mcp_sessions s ON s.id = c.session_id
		    WHERE s.tenant_id = t.id
		      AND s.client_name != 'orphaned'
		      AND c.called_at >= NOW() - INTERVAL '24 hours')
		  FROM tenants t
		 ORDER BY t.name
	`)
	if err != nil {
		m.Error = err.Error()
		return m
	}
	defer rows.Close()

	now := time.Now().UTC()
	for rows.Next() {
		var r TenantRow
		var lastActivity *time.Time
		var errorRate *float64

		if err := rows.Scan(
			&r.ID, &r.Name,
			&r.Workspaces, &r.Users, &r.Entities,
			&r.Sessions7d, &r.Policies,
			&lastActivity, &errorRate,
		); err != nil {
			continue
		}

		if lastActivity != nil {
			r.LastActivity = lastActivity.Format("2006-01-02 15:04")
			r.LastActivityUnix = lastActivity.Unix()
		} else {
			r.LastActivity = "—"
		}

		r.Health, r.HealthReason = deriveTenantHealth(now, lastActivity, errorRate)
		m.Rows = append(m.Rows, r)
	}

	m.Total = len(m.Rows)
	sortTenantsByHealth(m.Rows)
	return m
}

// deriveTenantHealth returns the color and a one-line reason.
//
// Precedence: an elevated error rate overrides the time-based rule,
// because a tenant with recent crashes matters more than one that is
// simply idle.
func deriveTenantHealth(now time.Time, lastActivity *time.Time, errorRate *float64) (string, string) {
	if errorRate != nil && *errorRate > 0.05 {
		return "red", "Error rate over 5% in the last 24 hours"
	}
	if lastActivity == nil {
		return "red", "No MCP sessions on record"
	}
	age := now.Sub(*lastActivity)
	switch {
	case age <= 7*24*time.Hour:
		return "green", "Active within the last 7 days"
	case age <= 30*24*time.Hour:
		return "amber", "Last activity 7–30 days ago"
	default:
		return "red", "No activity in over 30 days"
	}
}

// sortTenantsByHealth orders rows: red first, then amber, then green.
// Within the same color, most-recent activity first. Ties broken by
// name. The Tenants tab opens with the tenants that need attention at
// the top.
func sortTenantsByHealth(rows []TenantRow) {
	rank := func(s string) int {
		switch s {
		case "red":
			return 0
		case "amber":
			return 1
		case "green":
			return 2
		}
		return 3
	}
	// Small n; simple insertion sort avoids importing sort for one use.
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0; j-- {
			a, b := rows[j-1], rows[j]
			swap := false
			if rank(a.Health) > rank(b.Health) {
				swap = true
			} else if rank(a.Health) == rank(b.Health) {
				if a.LastActivityUnix < b.LastActivityUnix {
					swap = true
				} else if a.LastActivityUnix == b.LastActivityUnix && a.Name > b.Name {
					swap = true
				}
			}
			if !swap {
				break
			}
			rows[j-1], rows[j] = rows[j], rows[j-1]
		}
	}
}
