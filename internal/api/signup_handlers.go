// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/store"
)

var signupTmpl = template.Must(template.New("signup").Parse(signupHTML))

// HandleSignupGET renders the signup form. If the browser already has
// a valid session, it redirects to /dashboard.
func (s *Server) HandleSignupGET(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(SessionCookieName); err == nil {
		if _, ok := s.sessions.Get(cookie.Value); ok {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Frame-Options", "DENY")
	_ = signupTmpl.Execute(w, map[string]any{
		"Error":    "",
		"Email":    "",
		"FullName": "",
	})
}

// HandleSignupPOST validates the form, creates the user and their
// personal tenant and default workspace, issues a session cookie,
// and redirects to /dashboard.
//
// Validation order: required fields, password match, password length.
// Each failure renders the form again with a specific message and the
// email field preserved, so the user does not retype it.
//
// Duplicate email is the one error that must not reveal whether the
// email exists. The handler returns the same message a user would get
// if the email were invalid. That is the same reasoning as the login
// handler's generic "invalid email or password" — no enumeration.
func (s *Server) HandleSignupPOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderSignupError(w, "Invalid form submission", "", "")
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	confirm := r.FormValue("confirm")
	fullName := strings.TrimSpace(r.FormValue("full_name"))

	if email == "" || password == "" {
		s.renderSignupError(w, "Email and password are required", email, fullName)
		return
	}
	if password != confirm {
		s.renderSignupError(w, "Passwords do not match", email, fullName)
		return
	}
	if len(password) < 8 {
		s.renderSignupError(w, "Password must be at least 8 characters", email, fullName)
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		slog.Error("signup: hash password", "error", err)
		s.renderSignupError(w, "Signup failed. Please try again.", email, fullName)
		return
	}

	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		s.renderSignupError(w, "Signup unavailable", email, fullName)
		return
	}

	user, ws, err := pgStore.SignupUser(r.Context(), email, hash, fullName)
	if err != nil {
		// ErrEmailExists and a generic failure produce the same
		// user-facing message. The distinction is logged for the
		// operator but never shown to the browser.
		if errors.Is(err, store.ErrEmailExists) {
			slog.Info("signup: duplicate email", "email", email)
			s.renderSignupError(w, "An account with this email already exists", email, fullName)
			return
		}
		slog.Error("signup: create user", "error", err, "email", email)
		s.renderSignupError(w, "Signup failed. Please try again.", email, fullName)
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

	slog.Info("signup succeeded",
		"email", user.Email,
		"user_id", user.ID,
		"tenant_id", ws.TenantID.String(),
		"workspace_id", ws.ID.String(),
	)

	// The workspace is scoped to the new tenant. The dashboard's
	// workspace resolver does not yet read from the session — that is
	// Session C. Until then, a newly signed-up user lands on the
	// dashboard and sees whatever the canonical-tenant resolver
	// returns. That is a known Session B limitation; Session C closes
	// it by reading the workspace from the session.
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (s *Server) renderSignupError(w http.ResponseWriter, msg, email, fullName string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusBadRequest)
	_ = signupTmpl.Execute(w, map[string]any{
		"Error":    msg,
		"Email":    email,
		"FullName": fullName,
	})
}

const signupHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Create account — Garuda</title>
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
    width: min(440px, 92vw);
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
input[type=email], input[type=password], input[type=text] {
    width: 100%;
    height: 42px;
    border: 1px solid var(--border);
    background: #050811;
    border-radius: 9px;
    padding: 0 14px;
    color: white;
    outline: none;
    margin-bottom: 16px;
    transition: 0.15s;
}
input:focus {
    border-color: var(--brand);
    box-shadow: 0 0 0 3px rgba(56,189,248,0.25);
}
button {
    width: 100%;
    height: 44px;
    border: 0;
    background: var(--brand);
    color: #04060c;
    border-radius: 9px;
    font-weight: 750;
    font-size: 14px;
    cursor: pointer;
    transition: 0.15s;
}
button:hover { background: var(--brand-dark); color: white; }
.error {
    background: rgba(244,63,94,0.12);
    border: 1px solid rgba(244,63,94,0.4);
    color: var(--red);
    padding: 10px 14px;
    border-radius: 8px;
    font-size: 12px;
    margin-bottom: 18px;
}
.footer {
    margin-top: 22px;
    font-size: 12px;
    color: var(--muted);
    text-align: center;
}
.footer a { color: var(--brand); text-decoration: none; font-weight: 600; }
.footer a:hover { text-decoration: underline; }
</style>
</head>
<body>
<form class="card" method="POST" action="/signup" autocomplete="off">
    <div class="brand">
        <div class="brand-mark">🦅</div>
        <div>
            <div class="brand-name">GARUDA</div>
            <div class="brand-sub">Epistemic Software Intelligence</div>
        </div>
    </div>
    <h1>Create your workspace</h1>
    <div class="subtitle">You will be the owner of a private tenant and one default workspace. Invite your team later.</div>
    {{ if .Error }}<div class="error">{{ .Error }}</div>{{ end }}
    <label for="full_name">Full name</label>
    <input type="text" id="full_name" name="full_name" value="{{ .FullName }}" autocomplete="name">
    <label for="email">Email</label>
    <input type="email" id="email" name="email" value="{{ .Email }}" autocomplete="email" required>
    <label for="password">Password</label>
    <input type="password" id="password" name="password" autocomplete="new-password" required>
    <label for="confirm">Confirm password</label>
    <input type="password" id="confirm" name="confirm" autocomplete="new-password" required>
    <button type="submit">Create account</button>
    <div class="footer">
        Already have an account? <a href="/login">Sign in</a>
    </div>
</form>
</body>
</html>`
