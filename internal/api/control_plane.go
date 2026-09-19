// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"crypto/subtle"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/myshra777-ai/garuda/internal/store"
)

const (
	// ControlPlaneCookieName is the name of the owner session cookie.
	// Distinct from the tenant session cookie so the two never
	// collide, even if a browser holds both.
	ControlPlaneCookieName = "garuda_control_session"

	// ControlPlaneTokenEnv is the environment variable holding the
	// owner bearer token. When empty, the Control Plane is not
	// served at all — every /_/control request returns 404.
	ControlPlaneTokenEnv = "GARUDA_CONTROL_TOKEN"

	// ControlPlaneCookieTTL is the browser session lifetime. Matches
	// the spec: 24 hours from login. Rotating the server token
	// invalidates the cookie because the cookie stores the token
	// itself and is re-validated on every request.
	ControlPlaneCookieTTL = 24 * time.Hour

	// controlPlaneAuthWindow is the failure window. After N failures
	// in this window, the IP is blocked.
	controlPlaneAuthWindow = 15 * time.Minute

	// controlPlaneAuthMaxFailures is the threshold. Ten failures in
	// the window block the IP.
	controlPlaneAuthMaxFailures = 10

	// controlPlaneAuthBlockDuration is how long the block lasts.
	controlPlaneAuthBlockDuration = 30 * time.Minute
)

// controlPlaneToken returns the configured owner token, or empty if
// unset. A deployment without the env var does not serve the Control
// Plane.
func controlPlaneToken() string {
	return strings.TrimSpace(os.Getenv(ControlPlaneTokenEnv))
}

// controlPlaneFailures tracks failed auth attempts per source IP. It
// is in-memory; a server restart clears it. That matches the spec:
// rate limit state is not durable.
type controlPlaneFailures struct {
	mu           sync.Mutex
	attempts     map[string][]time.Time
	blockedUntil map[string]time.Time
}

var cpFailures = &controlPlaneFailures{
	attempts:     make(map[string][]time.Time),
	blockedUntil: make(map[string]time.Time),
}

func (c *controlPlaneFailures) record(ip string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-controlPlaneAuthWindow)
	kept := c.attempts[ip][:0]
	for _, t := range c.attempts[ip] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	kept = append(kept, now)
	c.attempts[ip] = kept
	if len(kept) >= controlPlaneAuthMaxFailures {
		c.blockedUntil[ip] = now.Add(controlPlaneAuthBlockDuration)
	}
}

func (c *controlPlaneFailures) isBlocked(ip string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	until, ok := c.blockedUntil[ip]
	if !ok {
		return false
	}
	if time.Now().After(until) {
		delete(c.blockedUntil, ip)
		delete(c.attempts, ip)
		return false
	}
	return true
}

// remoteIP returns the client IP without the port. r.RemoteAddr is
// "host:port" by default, which Postgres rejects as INET.
func remoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// recordControlAccess writes one row to control_plane_access. Best
// effort: a failure logs at WARN and the request proceeds.
func (s *Server) recordControlAccess(r *http.Request, status int, authResult string) {
	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if _, err := pgStore.Pool().Exec(ctx, `
		INSERT INTO control_plane_access
		  (path, method, status, source_ip, user_agent, auth_result)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, r.URL.Path, r.Method, status, remoteIP(r), r.UserAgent(), authResult); err != nil {
		slog.Warn("control-plane access log write failed", "error", err)
	}
}

// controlPlaneAuth gates every /_/control route except login and
// logout. Precedence: cookie first, then Authorization: Bearer.
//
// Unauthenticated requests return 404, not 401. A 401 confirms the
// endpoint exists; a 404 confirms nothing. A scanner sees /_/control
// and /_/random behave identically.
func (s *Server) controlPlaneAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		configured := controlPlaneToken()
		if configured == "" {
			http.NotFound(w, r)
			return
		}

		ip := remoteIP(r)
		if cpFailures.isBlocked(ip) {
			s.recordControlAccess(r, http.StatusNotFound, "rate_limited")
			http.NotFound(w, r)
			return
		}

		if c, err := r.Cookie(ControlPlaneCookieName); err == nil && c.Value != "" {
			if subtle.ConstantTimeCompare([]byte(c.Value), []byte(configured)) == 1 {
				next.ServeHTTP(w, r)
				return
			}
		}

		if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			presented := strings.TrimPrefix(auth, "Bearer ")
			if subtle.ConstantTimeCompare([]byte(presented), []byte(configured)) == 1 {
				next.ServeHTTP(w, r)
				return
			}
		}

		cpFailures.record(ip)
		s.recordControlAccess(r, http.StatusNotFound, "fail")
		slog.Warn("control-plane auth failure", "remote", r.RemoteAddr, "path", r.URL.Path)
		http.NotFound(w, r)
	})
}

// HandleControlPlane renders the shell for the selected tab. Read-only;
// no access-log write per the spec.
func (s *Server) HandleControlPlane(w http.ResponseWriter, r *http.Request) {
	tab := r.URL.Query().Get("tab")
	switch tab {
	case "business", "tenants", "operations", "internal":
	default:
		tab = "business"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = controlPlaneTmpl.Execute(w, map[string]any{"Tab": tab})
}

// HandleControlPlaneLoginGET renders the token form. If the browser
// already carries a valid control cookie, redirect to the shell.
func (s *Server) HandleControlPlaneLoginGET(w http.ResponseWriter, r *http.Request) {
	if controlPlaneToken() == "" {
		http.NotFound(w, r)
		return
	}
	if c, err := r.Cookie(ControlPlaneCookieName); err == nil {
		if subtle.ConstantTimeCompare([]byte(c.Value), []byte(controlPlaneToken())) == 1 {
			http.Redirect(w, r, "/_/control", http.StatusSeeOther)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = controlPlaneLoginTmpl.Execute(w, nil)
}

// HandleControlPlaneLoginPOST validates the submitted token and sets
// the session cookie on success. One access-log row per attempt.
func (s *Server) HandleControlPlaneLoginPOST(w http.ResponseWriter, r *http.Request) {
	if controlPlaneToken() == "" {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		s.renderControlLoginError(w, "Invalid form submission")
		return
	}
	presented := strings.TrimSpace(r.FormValue("token"))
	configured := controlPlaneToken()

	if presented == "" || subtle.ConstantTimeCompare([]byte(presented), []byte(configured)) != 1 {
		cpFailures.record(remoteIP(r))
		s.recordControlAccess(r, http.StatusUnauthorized, "fail")
		slog.Warn("control-plane login failure", "remote", r.RemoteAddr)
		s.renderControlLoginError(w, "Invalid token")
		return
	}

	s.recordControlAccess(r, http.StatusOK, "ok")
	http.SetCookie(w, &http.Cookie{
		Name:     ControlPlaneCookieName,
		Value:    configured,
		Path:     "/_/control",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(ControlPlaneCookieTTL / time.Second),
	})
	http.Redirect(w, r, "/_/control", http.StatusSeeOther)
}

// HandleControlPlaneLogout clears the control cookie. Logged as an
// event.
func (s *Server) HandleControlPlaneLogout(w http.ResponseWriter, r *http.Request) {
	s.recordControlAccess(r, http.StatusSeeOther, "ok")
	http.SetCookie(w, &http.Cookie{
		Name:     ControlPlaneCookieName,
		Value:    "",
		Path:     "/_/control",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/_/control/login", http.StatusSeeOther)
}

func (s *Server) renderControlLoginError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_ = controlPlaneLoginTmpl.Execute(w, map[string]any{"Error": msg})
}

// ─── Templates ────────────────────────────────────────────────────────

const controlPlaneShellHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Garuda — Control Plane</title>
<style>
:root { --bg:#060810; --surface:#0e1424; --border:#1f2b45; --text:#f8fafc; --text-2:#94a3b8; --muted:#64748b; --brand:#38bdf8; }
* { box-sizing: border-box; }
body { margin:0; background:var(--bg); color:var(--text); font-family:-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; font-size:13px; }
.app { display:flex; flex-direction:column; min-height:100vh; }
.header { padding:16px 28px; border-bottom:1px solid var(--border); display:flex; align-items:center; gap:20px; }
.brand { font-weight:800; letter-spacing:0.5px; font-size:15px; }
.brand-sub { color:var(--muted); font-size:11px; margin-top:2px; }
.tabs { display:flex; gap:4px; padding:0 28px; border-bottom:1px solid var(--border); background:rgba(6,8,16,0.85); }
.tab { padding:14px 16px; color:var(--text-2); text-decoration:none; font-weight:600; font-size:13px; border-bottom:2px solid transparent; }
.tab:hover { color:var(--text); }
.tab.active { color:var(--brand); border-bottom-color:var(--brand); }
.content { flex:1; padding:48px 28px; max-width:1200px; margin:0 auto; width:100%; }
.placeholder { padding:60px 20px; text-align:center; color:var(--muted); }
.placeholder h1 { color:var(--text); font-size:22px; margin:0 0 8px; font-weight:700; }
.logout { margin-left:auto; }
.logout button { background:transparent; border:1px solid var(--border); color:var(--text-2); padding:6px 12px; border-radius:6px; cursor:pointer; font-size:12px; }
.logout button:hover { color:var(--text); border-color:var(--text-2); }
</style>
</head>
<body>
<div class="app">
  <header class="header">
    <div>
      <div class="brand">GARUDA CONTROL PLANE</div>
      <div class="brand-sub">Platform operations</div>
    </div>
    <form class="logout" method="POST" action="/_/control/logout">
      <button type="submit">Sign out</button>
    </form>
  </header>
  <nav class="tabs">
    <a href="/_/control?tab=business"   class="tab {{if eq .Tab "business"}}active{{end}}">Business</a>
    <a href="/_/control?tab=tenants"    class="tab {{if eq .Tab "tenants"}}active{{end}}">Tenants</a>
    <a href="/_/control?tab=operations" class="tab {{if eq .Tab "operations"}}active{{end}}">Operations</a>
    <a href="/_/control?tab=internal"   class="tab {{if eq .Tab "internal"}}active{{end}}">Internal</a>
  </nav>
  <main class="content">
    {{if eq .Tab "business"}}
      <div class="placeholder">
        <h1>Business</h1>
        <p>Growth, adoption, and usage metrics.</p>
        <p style="margin-top:16px;font-size:12px;">Lands in the next commit.</p>
      </div>
    {{else if eq .Tab "tenants"}}
      <div class="placeholder">
        <h1>Tenants</h1>
        <p>Per-customer drilldown and health.</p>
        <p style="margin-top:16px;font-size:12px;">Lands after Business.</p>
      </div>
    {{else if eq .Tab "operations"}}
      <div class="placeholder">
        <h1>Operations</h1>
        <p>Platform health, error rates, session activity.</p>
        <p style="margin-top:16px;font-size:12px;">Lands last.</p>
      </div>
    {{else if eq .Tab "internal"}}
      <div class="placeholder">
        <h1>Internal</h1>
        <p>Roadmap, feature flags, beta tester list.</p>
        <p style="margin-top:16px;font-size:12px;">Queued.</p>
      </div>
    {{end}}
  </main>
</div>
</body>
</html>`

const controlPlaneLoginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Garuda — Control Plane</title>
<style>
body { background:#060810; color:#f8fafc; font-family:-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin:0; min-height:100vh; display:grid; place-items:center; }
.card { width:380px; padding:32px; background:#0e1424; border:1px solid #1f2b45; border-radius:12px; }
h1 { font-size:18px; margin:0 0 8px; font-weight:700; }
.sub { color:#94a3b8; font-size:13px; margin-bottom:24px; }
label { display:block; font-size:11px; color:#94a3b8; text-transform:uppercase; letter-spacing:0.05em; margin-bottom:6px; font-weight:600; }
input { width:100%; padding:12px; background:#050811; border:1px solid #1f2b45; border-radius:8px; color:#f8fafc; font-size:14px; font-family:ui-monospace, monospace; outline:none; }
input:focus { border-color:#38bdf8; }
button { width:100%; padding:12px; margin-top:20px; background:#0284c7; color:white; border:0; border-radius:8px; font-size:14px; font-weight:600; cursor:pointer; }
button:hover { background:#0369a1; }
.error { color:#f43f5e; font-size:13px; margin-bottom:16px; }
</style>
</head>
<body>
<form class="card" method="POST" action="/_/control/login">
  <h1>Control Plane</h1>
  <div class="sub">Owner access only.</div>
  {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
  <label for="token">Owner token</label>
  <input id="token" name="token" type="password" required autofocus autocomplete="off">
  <button type="submit">Sign in</button>
</form>
</body>
</html>`

var controlPlaneTmpl = template.Must(template.New("control_plane").Parse(controlPlaneShellHTML))
var controlPlaneLoginTmpl = template.Must(template.New("control_plane_login").Parse(controlPlaneLoginHTML))
