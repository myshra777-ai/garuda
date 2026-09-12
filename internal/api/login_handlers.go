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
    --bg: #060810;
    --surface: #0e1424;
    --border: #1f2b45;
    --text: #f8fafc;
    --text-2: #94a3b8;
    --muted: #64748b;
    --brand: #38bdf8;
    --brand-dark: #0284c7;
    --red: #f43f5e;
}
* { box-sizing: border-box; }
body {
    margin: 0;
    min-height: 100vh;
    background: var(--bg);
    color: var(--text);
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Inter, sans-serif;
    font-size: 14px;
    display: grid;
    place-items: center;
}
.card {
    width: min(420px, 92vw);
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 14px;
    padding: 36px 32px;
    box-shadow: 0 20px 60px -10px rgba(0,0,0,0.8);
}
.brand {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 28px;
}
.brand-mark {
    width: 40px; height: 40px;
    border-radius: 10px;
    background: linear-gradient(135deg, #0284c7, #38bdf8);
    display: grid; place-items: center;
    font-size: 20px;
    box-shadow: 0 0 22px rgba(56,189,248,0.5);
}
.brand-name { font-weight: 800; font-size: 18px; letter-spacing: 0.5px; }
.brand-sub { color: var(--muted); font-size: 11px; margin-top: 2px; }
h1 { font-size: 20px; font-weight: 700; margin: 0 0 6px; }
.subtitle { color: var(--text-2); font-size: 13px; margin-bottom: 24px; }
label { display: block; font-size: 12px; font-weight: 650; color: var(--text-2); margin-bottom: 6px; }
input[type=email], input[type=password] {
    width: 100%;
    height: 42px;
    border: 1px solid var(--border);
    background: #050811;
    border-radius: 9px;
    padding: 0 14px;
    color: var(--text);
    outline: none;
    font: inherit;
    margin-bottom: 16px;
    transition: 0.15s;
}
input:focus {
    border-color: var(--brand);
    box-shadow: 0 0 0 3px rgba(56, 189, 248, 0.22);
}
button {
    width: 100%;
    height: 44px;
    border: 0;
    border-radius: 9px;
    background: var(--brand-dark);
    color: white;
    font-size: 14px;
    font-weight: 700;
    cursor: pointer;
    transition: 0.15s;
}
button:hover { background: #0369a1; }
.error {
    background: rgba(244, 63, 94, 0.12);
    border: 1px solid rgba(244, 63, 94, 0.4);
    color: var(--red);
    border-radius: 8px;
    padding: 10px 14px;
    font-size: 12px;
    margin-bottom: 18px;
}
.footnote {
    color: var(--muted);
    font-size: 11px;
    text-align: center;
    margin-top: 22px;
}
</style>
</head>
<body>
<form class="card" method="POST" action="/login">
    <div class="brand">
        <div class="brand-mark">🦅</div>
        <div>
            <div class="brand-name">GARUDA</div>
            <div class="brand-sub">Epistemic Software Intelligence</div>
        </div>
    </div>

    <h1>Sign in</h1>
    <div class="subtitle">Authenticate to access the workspace.</div>

    {{ if .Error }}<div class="error">{{ .Error }}</div>{{ end }}

    <input type="hidden" name="next" value="{{ .Next }}">

    <label for="email">Email</label>
    <input id="email" type="email" name="email" required autofocus autocomplete="username">

    <label for="password">Password</label>
    <input id="password" type="password" name="password" required autocomplete="current-password">

    <button type="submit">Sign in</button>

    <div class="footnote">Garuda · Analyze | Verify | Govern</div>
</form>
</body>
</html>`
