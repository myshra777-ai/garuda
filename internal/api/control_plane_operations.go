// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"fmt"
	"time"
)

// StalledSession is one row in the Operations stalled-sessions panel.
type StalledSession struct {
	ID             string
	ClientName     string
	LastActivityAt string
	IdleFor        string
}

// ToolErrorRow is one row in the MCP tool errors panel.
type ToolErrorRow struct {
	ToolName string
	Total    int
	Errors   int
	RatePct  float64
}

// OperationsMetrics is the payload the Operations tab renders.
//
// The four working panels each have a paired Error field. When the
// query fails, the template renders "Not measured" with the error in
// the tooltip.
//
// The three locked panels (latency, error rate, recent errors) are
// declared here as booleans so the template can render the "not yet
// instrumented" state explicitly. They are not errors — they are
// honest signals that the underlying infrastructure does not exist.
type OperationsMetrics struct {
	// Working panels
	UptimeSeconds int64
	UptimeHuman   string

	DBSizeBytes int64
	DBSizeHuman string
	DBSizeError string

	StalledSessions      []StalledSession
	StalledSessionsError string

	ToolErrors      []ToolErrorRow
	ToolErrorsError string

	// Locked panels
	LatencyInstrumented   bool
	ErrorRateInstrumented bool
	ErrorLogInstrumented  bool

	// Fatal pool error — the read-only pool is missing.
	Error string
}

// loadOperationsMetrics runs the four working Operations panels.
//
// Latency, error rate, and the recent-errors log are not instrumented
// yet. The corresponding Locked flags are false so the template
// renders them as locked cards rather than fabricated zeros.
func (s *Server) loadOperationsMetrics(ctx context.Context) OperationsMetrics {
	m := OperationsMetrics{
		LatencyInstrumented:   false,
		ErrorRateInstrumented: false,
		ErrorLogInstrumented:  false,
	}

	if s.controlPool == nil {
		m.Error = "Control Plane database pool is not configured."
		return m
	}

	// Uptime — in-process, no query.
	if !s.processStart.IsZero() {
		d := time.Since(s.processStart)
		m.UptimeSeconds = int64(d.Seconds())
		m.UptimeHuman = humanDuration(d)
	} else {
		m.UptimeHuman = "unknown"
	}

	qctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// DB size
	var dbSize int64
	if err := s.controlPool.QueryRow(qctx,
		`SELECT pg_database_size(current_database())`,
	).Scan(&dbSize); err != nil {
		m.DBSizeError = err.Error()
	} else {
		m.DBSizeBytes = dbSize
		m.DBSizeHuman = humanBytes(dbSize)
	}

	// Stalled sessions — closed_at IS NULL and last_activity_at older
	// than 30 minutes. Orphaned sessions are excluded.
	rows, err := s.controlPool.Query(qctx, `
		SELECT id, client_name, last_activity_at,
		       EXTRACT(EPOCH FROM (NOW() - last_activity_at))::int
		  FROM mcp_sessions
		 WHERE closed_at IS NULL
		   AND last_activity_at < NOW() - INTERVAL '30 minutes'
		   AND client_name != 'orphaned'
		 ORDER BY last_activity_at ASC
		 LIMIT 50
	`)
	if err != nil {
		m.StalledSessionsError = err.Error()
	} else {
		defer rows.Close()
		for rows.Next() {
			var row StalledSession
			var lastActivity time.Time
			var idleSec int
			if err := rows.Scan(&row.ID, &row.ClientName, &lastActivity, &idleSec); err != nil {
				continue
			}
			row.LastActivityAt = lastActivity.Format("2006-01-02 15:04")
			row.IdleFor = humanDuration(time.Duration(idleSec) * time.Second)
			m.StalledSessions = append(m.StalledSessions, row)
		}
	}

	// MCP tool errors — tools with error rate over 5% in the last hour.
	rows2, err := s.controlPool.Query(qctx, `
		SELECT tool_name,
		       COUNT(*)::int AS total,
		       COUNT(*) FILTER (WHERE status = 'error')::int AS errors,
		       (COUNT(*) FILTER (WHERE status = 'error')::float / NULLIF(COUNT(*), 0) * 100)::float AS rate_pct
		  FROM mcp_tool_calls
		 WHERE called_at >= NOW() - INTERVAL '1 hour'
		 GROUP BY tool_name
		HAVING COUNT(*) FILTER (WHERE status = 'error')::float / NULLIF(COUNT(*), 0) > 0.05
		 ORDER BY rate_pct DESC
		 LIMIT 50
	`)
	if err != nil {
		m.ToolErrorsError = err.Error()
	} else {
		defer rows2.Close()
		for rows2.Next() {
			var r ToolErrorRow
			if err := rows2.Scan(&r.ToolName, &r.Total, &r.Errors, &r.RatePct); err != nil {
				continue
			}
			m.ToolErrors = append(m.ToolErrors, r)
		}
	}

	return m
}

// humanDuration renders a duration in a compact human form:
// "3s", "1m 20s", "2h 15m", "4d 6h".
func humanDuration(d time.Duration) string {
	if d < time.Second {
		return "<1s"
	}
	sec := int64(d.Seconds())
	if sec < 60 {
		return fmt.Sprintf("%ds", sec)
	}
	min := sec / 60
	sec = sec % 60
	if min < 60 {
		if sec == 0 {
			return fmt.Sprintf("%dm", min)
		}
		return fmt.Sprintf("%dm %ds", min, sec)
	}
	hr := min / 60
	min = min % 60
	if hr < 24 {
		if min == 0 {
			return fmt.Sprintf("%dh", hr)
		}
		return fmt.Sprintf("%dh %dm", hr, min)
	}
	day := hr / 24
	hr = hr % 24
	if hr == 0 {
		return fmt.Sprintf("%dd", day)
	}
	return fmt.Sprintf("%dd %dh", day, hr)
}

// humanBytes renders a byte count in KiB / MiB / GiB.
func humanBytes(b int64) string {
	const (
		KiB = 1024
		MiB = 1024 * KiB
		GiB = 1024 * MiB
	)
	switch {
	case b >= GiB:
		return fmt.Sprintf("%.2f GiB", float64(b)/float64(GiB))
	case b >= MiB:
		return fmt.Sprintf("%.2f MiB", float64(b)/float64(MiB))
	case b >= KiB:
		return fmt.Sprintf("%.2f KiB", float64(b)/float64(KiB))
	}
	return fmt.Sprintf("%d B", b)
}
