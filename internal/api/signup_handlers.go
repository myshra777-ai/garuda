// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/tenant"
)

var signupTmpl = template.Must(template.New("signup").Parse(signupHTML))

func isAPIClient(r *http.Request) bool {
	accept := strings.ToLower(r.Header.Get("Accept"))
	ct := strings.ToLower(r.Header.Get("Content-Type"))
	ua := strings.ToLower(r.Header.Get("User-Agent"))
	return strings.Contains(accept, "application/json") ||
		strings.Contains(ct, "application/json") ||
		strings.HasPrefix(ua, "curl/") ||
		strings.HasPrefix(ua, "python") ||
		strings.HasPrefix(ua, "httpie/")
}

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

// HandleSignupPOST validates the form or JSON payload, creates the user
// and their personal tenant and default workspace (plus grants access to
// the canonical CLI tenant workspaces), issues a session cookie, and
// redirects browsers to /dashboard or returns JSON for CLI/API callers.
func (s *Server) HandleSignupPOST(w http.ResponseWriter, r *http.Request) {
	var email, password, confirm, fullName string
	hasConfirm := false

	if strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		var body struct {
			Email    string  `json:"email"`
			Password string  `json:"password"`
			Confirm  *string `json:"confirm"`
			FullName string  `json:"full_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			s.respondSignupError(w, r, "Invalid JSON submission", "", "")
			return
		}
		email = strings.TrimSpace(body.Email)
		password = body.Password
		fullName = strings.TrimSpace(body.FullName)
		if body.Confirm != nil {
			confirm = *body.Confirm
			hasConfirm = true
		} else {
			confirm = password
		}
	} else {
		if err := r.ParseForm(); err != nil {
			s.respondSignupError(w, r, "Invalid form submission", "", "")
			return
		}
		email = strings.TrimSpace(r.FormValue("email"))
		password = r.FormValue("password")
		fullName = strings.TrimSpace(r.FormValue("full_name"))
		if vals, ok := r.PostForm["confirm"]; ok && len(vals) > 0 {
			confirm = vals[0]
			hasConfirm = true
		} else if vals, ok := r.Form["confirm"]; ok && len(vals) > 0 {
			confirm = vals[0]
			hasConfirm = true
		} else {
			// CLI / curl callers often pass only -d email=... -d password=...
			confirm = password
		}
	}

	if email == "" || password == "" {
		s.respondSignupError(w, r, "Email and password are required", email, fullName)
		return
	}
	if hasConfirm && password != confirm {
		s.respondSignupError(w, r, "Passwords do not match", email, fullName)
		return
	}
	if len(password) < 8 {
		s.respondSignupError(w, r, "Password must be at least 8 characters", email, fullName)
		return
	}
	if fullName == "" {
		if at := strings.IndexByte(email, '@'); at > 0 {
			fullName = email[:at]
		} else {
			fullName = email
		}
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		slog.Error("signup: hash password", "error", err)
		s.respondSignupError(w, r, "Signup failed. Please try again.", email, fullName)
		return
	}

	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		s.respondSignupError(w, r, "Signup unavailable", email, fullName)
		return
	}

	user, ws, err := pgStore.SignupUser(r.Context(), email, hash, fullName)
	if err != nil {
		if errors.Is(err, store.ErrEmailExists) {
			slog.Info("signup: duplicate email", "email", email)
			// For idempotent CLI/API setup scripts, if the user already exists,
			// link them to canonical workspaces and return 200 OK in JSON mode.
			if isAPIClient(r) {
				if existingUser, uErr := pgStore.GetUserByEmail(r.Context(), email); uErr == nil && existingUser != nil {
					grantCanonicalWorkspaceAccess(r, pgStore, existingUser.ID)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"status":  "already_exists",
						"user_id": existingUser.ID,
						"email":   existingUser.Email,
						"role":    existingUser.Role,
					})
					return
				}
			}
			s.respondSignupError(w, r, "An account with this email already exists", email, fullName)
			return
		}
		slog.Error("signup: create user", "error", err, "email", email)
		s.respondSignupError(w, r, "Signup failed. Please try again.", email, fullName)
		return
	}

	// Ensure interactive users signed up via /signup can also view CLI-analyzed
	// workspaces under the canonical tenant on their local dashboard.
	if !strings.HasSuffix(email, "@local") {
		grantCanonicalWorkspaceAccess(r, pgStore, user.ID)
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

	if isAPIClient(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":       "created",
			"user_id":      user.ID,
			"email":        user.Email,
			"role":         user.Role,
			"tenant_id":    ws.TenantID.String(),
			"workspace_id": ws.ID.String(),
		})
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func grantCanonicalWorkspaceAccess(r *http.Request, pgStore *store.PostgresStore, userID string) {
	_, _ = pgStore.Pool().Exec(r.Context(), `
		INSERT INTO tenant_members (user_id, tenant_id, role, created_at)
		VALUES ($1, $2, 'owner', NOW())
		ON CONFLICT (user_id, tenant_id) DO NOTHING
	`, userID, tenant.CanonicalID)

	_, _ = pgStore.Pool().Exec(r.Context(), `
		INSERT INTO workspace_members (user_id, workspace_id, role, created_at)
		SELECT $1, w.id, 'owner', NOW()
		  FROM workspaces w
		 WHERE w.tenant_id = $2
		ON CONFLICT (user_id, workspace_id) DO NOTHING
	`, userID, tenant.CanonicalID)
}

func (s *Server) respondSignupError(w http.ResponseWriter, r *http.Request, msg, email, fullName string) {
	if isAPIClient(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": msg,
			"email": email,
		})
		return
	}
	s.renderSignupError(w, msg, email, fullName)
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
