// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"time"
)

// BusinessMetrics is the data the Business tab renders. Every field
// has a formula in control-plane-metrics.md. Every read runs through
// the read-only Control Plane pool.
//
// Every metric has a paired Error field. When the query fails, the
// error is recorded here and the template renders "Not measured"
// instead of a fabricated zero. Silent failures are not acceptable on
// a governance dashboard.
type BusinessMetrics struct {
	RangeDays  int
	RangeLabel string

	Tenants      int
	TenantsError string

	NewTenants      int
	NewTenantsError string

	Workspaces      int
	WorkspacesError string

	UsersRecent   int
	UsersMeasured bool
	UsersError    string

	Sessions      int
	SessionsError string

	ClientBreakdown      []ClientCount
	ClientBreakdownError string

	// Error is a fatal failure — the read-only pool is missing, or
	// the first query failed so hard that no metric rendered. When
	// set, the template shows only this message.
	Error string
}

type ClientCount struct {
	Name  string
	Count int
}

// loadBusinessMetrics runs the six Business-tab queries against the
// read-only Control Plane pool. A failed query is recorded in the
// corresponding Error field; the tab renders with the other metrics
// intact.
//
// The interval parameter is cast to text before concatenation.
// pgx sends Go int parameters as int; Postgres's || operator is
// text||text. Writing ($1 || ' days') caused the query to fail and
// the previous version discarded the error, rendering 0 for every
// windowed metric. See control_plane_business.go history.
func (s *Server) loadBusinessMetrics(ctx context.Context, rangeParam string) BusinessMetrics {
	m := BusinessMetrics{RangeDays: 30, RangeLabel: "30 days"}

	if s.controlPool == nil {
		m.Error = "Control Plane database pool is not configured."
		return m
	}

	switch rangeParam {
	case "60":
		m.RangeDays = 60
		m.RangeLabel = "60 days"
	case "90":
		m.RangeDays = 90
		m.RangeLabel = "90 days"
	}

	qctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Tenants
	if err := s.controlPool.QueryRow(qctx,
		`SELECT COUNT(*) FROM tenants`,
	).Scan(&m.Tenants); err != nil {
		m.TenantsError = err.Error()
	}

	// New tenants
	if err := s.controlPool.QueryRow(qctx, `
		SELECT COUNT(*) FROM tenants
		 WHERE created_at >= NOW() - make_interval(days => $1)
	`, m.RangeDays).Scan(&m.NewTenants); err != nil {
		m.NewTenantsError = err.Error()
	}

	// Workspaces
	if err := s.controlPool.QueryRow(qctx,
		`SELECT COUNT(*) FROM workspaces`,
	).Scan(&m.Workspaces); err != nil {
		m.WorkspacesError = err.Error()
	}

	// Users — three-state. The first column counts rows with a
	// non-NULL last_seen_at. If it is zero, the metric renders "Not
	// yet measured" rather than 0. If the query errors, UsersError is
	// set and the template renders the same state.
	var hasAny int
	if err := s.controlPool.QueryRow(qctx, `
		SELECT
		  COUNT(*) FILTER (WHERE last_seen_at IS NOT NULL),
		  COUNT(DISTINCT user_id) FILTER (
		    WHERE last_seen_at >= NOW() - make_interval(days => $1)
		  )
		FROM tenant_members
	`, m.RangeDays).Scan(&hasAny, &m.UsersRecent); err != nil {
		m.UsersError = err.Error()
	} else if hasAny > 0 {
		m.UsersMeasured = true
	}

	// Sessions
	if err := s.controlPool.QueryRow(qctx, `
		SELECT COUNT(*) FROM mcp_sessions
		 WHERE started_at >= NOW() - make_interval(days => $1)
		   AND client_name != 'orphaned'
	`, m.RangeDays).Scan(&m.Sessions); err != nil {
		m.SessionsError = err.Error()
	}

	// Client breakdown — all-time, no window. The Sessions card is
	// windowed; this panel is not. The panel subtitle says "all time"
	// so the two numbers are not read as a contradiction.
	rows, err := s.controlPool.Query(qctx, `
		SELECT client_name, COUNT(*)
		  FROM mcp_sessions
		 WHERE client_name != 'orphaned'
		 GROUP BY client_name
		 ORDER BY COUNT(*) DESC
		 LIMIT 10
	`)
	if err != nil {
		m.ClientBreakdownError = err.Error()
	} else {
		defer rows.Close()
		for rows.Next() {
			var c ClientCount
			if rows.Scan(&c.Name, &c.Count) == nil {
				m.ClientBreakdown = append(m.ClientBreakdown, c)
			}
		}
	}

	return m
}
