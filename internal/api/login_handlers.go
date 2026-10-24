// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/tenant"
)

// bootstrapAdminEmail is the fixed email of the automatically created
// administrator. The password is random and printed once to stderr on
// first daemon startup.
const bootstrapAdminEmail = "admin@local"

var loginTmpl = template.Must(template.New("login").Parse(loginHTML))

// HandleLoginGET renders the login form. If the browser already has a
// valid session, it redirects to /dashboard.
func (s *Server) HandleLoginGET(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(SessionCookieName); err == nil {
		if _, ok := s.sessions.Get(cookie.Value); ok {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Frame-Options", "DENY")
	_ = loginTmpl.Execute(w, map[string]any{
		"Error": "",
		"Next":  r.URL.Query().Get("next"),
	})
}

// HandleLoginPOST verifies credentials and issues a session cookie.
func (s *Server) HandleLoginPOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderLoginError(w, "Invalid form submission", "")
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	next := r.FormValue("next")

	if email == "" || password == "" {
		s.renderLoginError(w, "Email and password are required", next)
		return
	}

	user, err := s.authService.SignIn(r.Context(), email, password)
	if err != nil {
		// Do not log the attempted password. Do not reveal whether
		// the email exists.
		slog.Warn("login failed", "email", email, "remote", r.RemoteAddr)
		s.renderLoginError(w, "Invalid email or password", next)
		return
	}

	sess := s.sessions.Create(user.ID, user.Email, user.Role)

	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sess.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(SessionTTL / time.Second),
	})

	slog.Info("login succeeded", "email", user.Email, "role", user.Role)

	// last_login_at drives the active_users aggregate. Written here
	// rather than in SignIn so a failed signin does not count as
	// activity. Best-effort: a failed write does not fail the login.
	if pgStore, ok := s.store.(*store.PostgresStore); ok && pgStore != nil {
		if _, err := pgStore.Pool().Exec(r.Context(),
			`UPDATE users SET last_login_at = NOW() WHERE id = $1`, user.ID); err != nil {
			slog.Warn("login: last_login_at update failed", "user_id", user.ID, "error", err)
		}
	}

	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		next = "/dashboard"
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}

// HandleLogout deletes the session and clears the cookie.
func (s *Server) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(SessionCookieName); err == nil {
		s.sessions.Delete(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Server) renderLoginError(w http.ResponseWriter, msg, next string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusUnauthorized)
	_ = loginTmpl.Execute(w, map[string]any{
		"Error": msg,
		"Next":  next,
	})
}

func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

// BootstrapAdminIfEmpty creates the default admin user if no users
// exist. Called once at daemon startup.
//
// The password is random, printed once to stderr, and cannot be
// recovered from the database — only the bcrypt hash is stored. If
// the operator loses the password, they must delete the row and
// restart the daemon.
func (s *Server) BootstrapAdminIfEmpty(ctx context.Context) error {
	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok {
		return fmt.Errorf("bootstrap requires a PostgresStore, got %T", s.store)
	}

	_, err := pgStore.GetUserByEmail(ctx, bootstrapAdminEmail)
	if err == nil {
		return nil // already exists
	}
	if !errors.Is(err, auth.ErrUserNotFound) {
		return fmt.Errorf("bootstrap: query user: %w", err)
	}

	password, err := generateBootstrapPassword()
	if err != nil {
		return err
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("bootstrap: hash password: %w", err)
	}

	now := time.Now().UTC()
	user := &auth.User{
		ID:           uuid.New().String(),
		Email:        bootstrapAdminEmail,
		PasswordHash: hash,
		FullName:     "Administrator",
		Role:         "admin",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := pgStore.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("bootstrap: create user: %w", err)
	}

	// The bootstrap admin is not created through /signup, so it does
	// not get the two membership rows that the signup flow creates.
	// Without these rows, the Session C middleware refuses every
	// workspace-scoped request for this user with a 404 — including
	// the workspaces the founder was already using before Session A.
	//
	// Owner of the canonical tenant, and owner of every workspace in
	// it. ON CONFLICT DO NOTHING so a re-run against an existing
	// database is a no-op.
	if _, err := pgStore.Pool().Exec(ctx, `
		INSERT INTO tenant_members (user_id, tenant_id, role, created_at)
		VALUES ($1, $2, 'owner', NOW())
		ON CONFLICT (user_id, tenant_id) DO NOTHING
	`, user.ID, tenant.CanonicalID); err != nil {
		return fmt.Errorf("bootstrap: insert tenant_members: %w", err)
	}

	if _, err := pgStore.Pool().Exec(ctx, `
		INSERT INTO workspace_members (user_id, workspace_id, role, created_at)
		SELECT $1, w.id, 'owner', NOW()
		  FROM workspaces w
		 WHERE w.tenant_id = $2
		ON CONFLICT (user_id, workspace_id) DO NOTHING
	`, user.ID, tenant.CanonicalID); err != nil {
		return fmt.Errorf("bootstrap: insert workspace_members: %w", err)
	}

	// Print to stderr, not to structured logs, so that operators see
	// it in the terminal and can copy it before it scrolls.
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "  GARUDA BOOTSTRAP ADMIN CREATED")
	fmt.Fprintln(os.Stderr, "  Email:    "+bootstrapAdminEmail)
	fmt.Fprintln(os.Stderr, "  Password: "+password)
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  Save this password. It is not stored in plaintext")
	fmt.Fprintln(os.Stderr, "  and cannot be recovered. To regenerate, delete the")
	fmt.Fprintln(os.Stderr, "  admin@local row and restart the daemon.")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "")

	return nil
}

func generateBootstrapPassword() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("bootstrap: crypto/rand: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

const loginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Sign in — Garuda</title>
<style>
:root {
    --bg: #F7F7F3;
    --surface: #FFFFFF;
    --border: #E7E7E1;
    --ink: #191B19;
    --text: #191B19;
    --text-2: #4A4D48;
    --muted: #767A72;
    --accent: #D5E66A;
    --red: #991B1B;
}
* { box-sizing: border-box; }
body {
    margin: 0;
    min-height: 100vh;
    background: var(--bg);
    color: var(--text);
    font-family: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    font-size: 13.5px;
    display: grid;
    place-items: center;
}
.card {
    width: min(410px, 92vw);
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 34px 32px;
    box-shadow: 0 12px 32px -4px rgba(25, 27, 25, 0.08);
}
.brand {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 24px;
}
.brand-mark {
    width: 32px; height: 32px;
    border-radius: 7px;
    background: var(--ink);
    color: var(--accent);
    display: grid; place-items: center;
    font-weight: 800;
    font-size: 14px;
}
.brand-name { font-weight: 750; font-size: 15px; letter-spacing: -0.01em; color: var(--text); }
.brand-sub { color: var(--muted); font-size: 11px; margin-top: 1px; }
h1 { font-size: 20px; font-weight: 650; letter-spacing: -0.02em; margin: 0 0 6px; }
.subtitle { color: var(--text-2); font-size: 13px; margin-bottom: 22px; }
label { display: block; font-size: 11.5px; font-weight: 600; color: var(--text-2); margin-bottom: 6px; }
input[type=email], input[type=password] {
    width: 100%;
    height: 38px;
    border: 1px solid var(--border);
    background: var(--bg);
    border-radius: 7px;
    padding: 0 12px;
    color: var(--text);
    outline: none;
    font: inherit;
    margin-bottom: 15px;
    transition: border-color 0.15s, background 0.15s;
}
input:focus {
    border-color: var(--ink);
    background: var(--surface);
}
button {
    width: 100%;
    height: 40px;
    border: 1px solid var(--ink);
    border-radius: 7px;
    background: var(--ink);
    color: #FFFFFF;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: opacity 0.15s;
}
button:hover { opacity: 0.9; }
.error {
    background: #FDECEC;
    border: 1px solid #F3B8B8;
    color: var(--red);
    border-radius: 7px;
    padding: 10px 12px;
    font-size: 12px;
    margin-bottom: 16px;
}
.footnote {
    color: var(--muted);
    font-size: 11.5px;
    text-align: center;
    margin-top: 20px;
}
.footnote a { color: var(--text); font-weight: 600; text-decoration: none; }
.footnote a:hover { text-decoration: underline; }
</style>
</head>
<body>
<form class="card" method="POST" action="/login">
    <div class="brand">
        <div class="brand-mark">G</div>
        <div>
            <div class="brand-name">GARUDA</div>
            <div class="brand-sub">Software Intelligence Instrument</div>
        </div>
    </div>

    <h1>Sign in</h1>
    <div class="subtitle">Authenticate to inspect architecture, policies, and evidence.</div>

    {{ if .Error }}<div class="error">{{ .Error }}</div>{{ end }}

    <input type="hidden" name="next" value="{{ .Next }}">

    <label for="email">Email</label>
    <input id="email" type="email" name="email" required autofocus autocomplete="username">

    <label for="password">Password</label>
    <input id="password" type="password" name="password" required autocomplete="current-password">

    <button type="submit">Sign in →</button>

    <div class="footnote">New workspace? <a href="/signup">Create account</a></div>
</form>
</body>
</html>`
