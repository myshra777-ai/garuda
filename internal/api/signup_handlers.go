// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"os"
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

// isCanonicalBootstrapOwner reports whether email is the owner configured for this
// installation through GARUDA_OWNER_EMAIL. No address is trusted by default: a fixed
// address in the source would let anyone who registers it first take over the
// original tenant on every installation.
func isCanonicalBootstrapOwner(email string) bool {
	e := strings.ToLower(strings.TrimSpace(email))
	ownerEnv := strings.ToLower(strings.TrimSpace(os.Getenv("GARUDA_OWNER_EMAIL")))
	return ownerEnv != "" && e == ownerEnv
}

// HandleSignupGET renders the 3-mode registration page (Organization Tenant Owner,
// Team Member / Employee Join, and Individual Developer Workflow).
func (s *Server) HandleSignupGET(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(SessionCookieName); err == nil {
		if _, ok := s.sessions.Get(cookie.Value); ok && r.URL.Query().Get("switch") == "" {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
	}

	mode := strings.TrimSpace(r.URL.Query().Get("mode"))
	if mode == "" {
		mode = "tenant_owner"
	}
	invite := strings.TrimSpace(r.URL.Query().Get("invite"))
	if invite != "" && r.URL.Query().Get("mode") == "" {
		mode = "team_member"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Frame-Options", "DENY")
	_ = signupTmpl.Execute(w, map[string]any{
		"Error":          "",
		"Email":          "",
		"FullName":       "",
		"Username":       "",
		"OrgName":        "",
		"OrgDomain":      "",
		"WorkspaceName":  "default",
		"TenantJoinCode": invite,
		"AccountType":    mode,
	})
}

// HandleSignupPOST validates the form or JSON payload, provisions either:
//  1. An isolated Organization Tenant + Owner account ("tenant_owner")
//  2. A Team Member account bound ONLY to the target Tenant Invite Code / ID ("team_member")
//  3. An Individual Developer workflow account ("individual_dev")
//
// NEVER leaks canonical tenant access to other users!
func (s *Server) HandleSignupPOST(w http.ResponseWriter, r *http.Request) {
	var (
		email, password, confirm, fullName string
		opts                               store.SignupOptions
		hasConfirm                         bool
	)

	if strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		var body struct {
			Email          string  `json:"email"`
			Password       string  `json:"password"`
			Confirm        *string `json:"confirm"`
			FullName       string  `json:"full_name"`
			Username       string  `json:"username"`
			AccountType    string  `json:"account_type"`
			OrgName        string  `json:"org_name"`
			OrgDomain      string  `json:"org_domain"`
			WorkspaceName  string  `json:"workspace_name"`
			TenantJoinCode string  `json:"tenant_join_code"`
			JobTitle       string  `json:"job_title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			s.respondSignupError(w, r, "Invalid JSON submission", "", "", opts)
			return
		}
		email = strings.TrimSpace(body.Email)
		password = body.Password
		fullName = strings.TrimSpace(body.FullName)
		opts = store.SignupOptions{
			AccountType:    strings.TrimSpace(body.AccountType),
			Username:       strings.TrimSpace(body.Username),
			OrgName:        strings.TrimSpace(body.OrgName),
			OrgDomain:      strings.TrimSpace(body.OrgDomain),
			WorkspaceName:  strings.TrimSpace(body.WorkspaceName),
			TenantJoinCode: strings.TrimSpace(body.TenantJoinCode),
			JobTitle:       strings.TrimSpace(body.JobTitle),
		}
		if body.Confirm != nil {
			confirm = *body.Confirm
			hasConfirm = true
		} else {
			confirm = password
		}
	} else {
		if err := r.ParseForm(); err != nil {
			s.respondSignupError(w, r, "Invalid form submission", "", "", opts)
			return
		}
		email = strings.TrimSpace(r.FormValue("email"))
		password = r.FormValue("password")
		fullName = strings.TrimSpace(r.FormValue("full_name"))
		opts = store.SignupOptions{
			AccountType:    strings.TrimSpace(r.FormValue("account_type")),
			Username:       strings.TrimSpace(r.FormValue("username")),
			OrgName:        strings.TrimSpace(r.FormValue("org_name")),
			OrgDomain:      strings.TrimSpace(r.FormValue("org_domain")),
			WorkspaceName:  strings.TrimSpace(r.FormValue("workspace_name")),
			TenantJoinCode: strings.TrimSpace(r.FormValue("tenant_join_code")),
			JobTitle:       strings.TrimSpace(r.FormValue("job_title")),
		}
		if vals, ok := r.PostForm["confirm"]; ok && len(vals) > 0 && vals[0] != "" {
			confirm = vals[0]
			hasConfirm = true
		} else if vals, ok := r.Form["confirm"]; ok && len(vals) > 0 && vals[0] != "" {
			confirm = vals[0]
			hasConfirm = true
		} else {
			confirm = password
		}
	}

	if opts.AccountType == "" {
		opts.AccountType = "tenant_owner"
	}

	if email == "" || password == "" {
		s.respondSignupError(w, r, "Email and password are required", email, fullName, opts)
		return
	}
	if hasConfirm && password != confirm {
		s.respondSignupError(w, r, "Passwords do not match", email, fullName, opts)
		return
	}
	if len(password) < 8 {
		s.respondSignupError(w, r, "Password must be at least 8 characters", email, fullName, opts)
		return
	}
	if opts.AccountType == "team_member" && strings.TrimSpace(opts.TenantJoinCode) == "" {
		s.respondSignupError(w, r, "Please enter your Organization Tenant ID or Invite Code (e.g. GRD-GARUDA1)", email, fullName, opts)
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		slog.Error("signup: hash password", "error", err)
		s.respondSignupError(w, r, "Signup failed. Please try again.", email, fullName, opts)
		return
	}

	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		s.respondSignupError(w, r, "Signup unavailable", email, fullName, opts)
		return
	}

	user, ws, inviteCode, err := pgStore.SignupUserWithOptions(r.Context(), email, hash, fullName, opts)
	if err != nil {
		if errors.Is(err, store.ErrEmailExists) {
			slog.Info("signup: duplicate email", "email", email)
			// Never return the existing account's id, role or email to an unauthenticated caller,
			// and never change its access from a sign-up request.
			s.respondSignupError(w, r, "An account with this email already exists. Please sign in instead.", email, fullName, opts)
			return
		}
		if errors.Is(err, store.ErrTenantNotFound) {
			s.respondSignupError(w, r, "Tenant ID or Invite Code not found. Ask your Organization Owner for their Invite Code (e.g. GRD-GARUDA1) or Tenant UUID.", email, fullName, opts)
			return
		}
		slog.Error("signup: create user", "error", err, "email", email)
		s.respondSignupError(w, r, "Signup failed. Please try again.", email, fullName, opts)
		return
	}

	// STRICT MULTI-TENANT ISOLATION:
	// Only link the canonical CLI tenant (00000000-0000-0000-0000-000000000001)
	// if this signup is the owner configured through GARUDA_OWNER_EMAIL.
	// Never grant canonical tenant access to other tenants or external users!
	if isCanonicalBootstrapOwner(user.Email) && opts.AccountType == "tenant_owner" {
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
		"role", user.Role,
		"account_type", opts.AccountType,
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
			"account_type": opts.AccountType,
			"tenant_id":    ws.TenantID.String(),
			"invite_code":  inviteCode,
			"workspace_id": ws.ID.String(),
			"workspace":    ws.Name,
		})
		return
	}

	http.Redirect(w, r, "/dashboard?workspace="+ws.Name, http.StatusSeeOther)
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

	_, _ = pgStore.Pool().Exec(r.Context(), `
		UPDATE users SET primary_tenant_id = $2, role = 'owner', account_type = 'tenant_owner' WHERE id = $1
	`, userID, tenant.CanonicalID)
}

func (s *Server) respondSignupError(w http.ResponseWriter, r *http.Request, msg, email, fullName string, opts store.SignupOptions) {
	if isAPIClient(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": msg,
			"email": email,
		})
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusBadRequest)
	if opts.WorkspaceName == "" {
		opts.WorkspaceName = "default"
	}
	if opts.AccountType == "" {
		opts.AccountType = "tenant_owner"
	}
	_ = signupTmpl.Execute(w, map[string]any{
		"Error":          msg,
		"Email":          email,
		"FullName":       fullName,
		"Username":       opts.Username,
		"OrgName":        opts.OrgName,
		"OrgDomain":      opts.OrgDomain,
		"WorkspaceName":  opts.WorkspaceName,
		"TenantJoinCode": opts.TenantJoinCode,
		"AccountType":    opts.AccountType,
	})
}

const signupHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Onboarding &amp; Workspace Provisioning — Garuda</title>
<style>
:root {
    --font-serif: "Playfair Display", "Cormorant Garamond", "Lora", "Iowan Old Style", Georgia, serif;
    --font-sans: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    --font-mono: "JetBrains Mono", "SFMono-Regular", Consolas, monospace;
    --bg: #F7F6F2;
    --surface: #FFFFFF;
    --surface-subtle: #F0EFEA;
    --border: rgba(20, 21, 23, 0.10);
    --border-strong: rgba(20, 21, 23, 0.22);
    --ink: #141517;
    --text-2: #45484F;
    --muted: #686B73;
    --accent: #10B981;
    --accent-soft: rgba(16, 185, 129, 0.11);
    --accent-ink: #0F5338;
    --red: #991B1B;
}
* { box-sizing: border-box; }
body {
    margin: 0;
    min-height: 100vh;
    background: radial-gradient(circle at 80% 20%, rgba(16, 185, 129, 0.08) 0%, transparent 45%), var(--bg);
    color: var(--ink);
    font-family: var(--font-sans);
    font-size: 13.5px;
    display: grid;
    place-items: center;
    padding: 28px 16px;
}
.shell {
    width: min(560px, 96vw);
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 16px;
    padding: 34px 34px 28px;
    box-shadow: 0 20px 48px -12px rgba(20, 21, 23, 0.10);
}
.brand-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 20px;
}
.brand {
    display: flex;
    align-items: center;
    gap: 10px;
}
.brand-mark {
    width: 32px; height: 32px;
    border-radius: 8px;
    background: linear-gradient(135deg, #10B981 0%, #059669 100%);
    color: #FFFFFF;
    display: grid; place-items: center;
    font-weight: 800;
    font-size: 14px;
}
.brand-name { font-weight: 700; font-size: 14.5px; letter-spacing: -0.01em; }
.brand-sub { color: var(--muted); font-size: 11px; }
h1 {
    font-family: var(--font-serif);
    font-size: 28px;
    font-weight: 500;
    letter-spacing: -0.02em;
    margin: 0 0 6px;
}
.subtitle {
    color: var(--text-2);
    font-size: 13px;
    margin-bottom: 20px;
    line-height: 1.5;
}
.mode-tabs {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 6px;
    background: var(--surface-subtle);
    padding: 5px;
    border-radius: 11px;
    border: 1px solid var(--border);
    margin-bottom: 18px;
}
.mode-tab {
    border: none;
    background: transparent;
    padding: 9px 8px;
    border-radius: 8px;
    font-family: var(--font-sans);
    font-size: 12px;
    font-weight: 600;
    color: var(--text-2);
    cursor: pointer;
    text-align: center;
    transition: all 0.16s ease;
}
.mode-tab.active {
    background: var(--surface);
    color: var(--ink);
    box-shadow: 0 2px 8px rgba(20, 21, 23, 0.08);
}
.mode-banner {
    padding: 10px 13px;
    border-radius: 9px;
    background: var(--accent-soft);
    border: 1px solid rgba(16, 185, 129, 0.25);
    color: var(--accent-ink);
    font-size: 12px;
    line-height: 1.45;
    margin-bottom: 18px;
}
.grid-2 {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
}
label {
    display: block;
    font-size: 11.5px;
    font-weight: 600;
    color: var(--text-2);
    margin-bottom: 5px;
}
input[type=email], input[type=password], input[type=text] {
    width: 100%;
    height: 38px;
    border: 1px solid var(--border);
    background: var(--bg);
    border-radius: 8px;
    padding: 0 12px;
    color: var(--ink);
    outline: none;
    font: inherit;
    margin-bottom: 13px;
    transition: border-color 0.15s, background 0.15s;
}
input:focus {
    border-color: var(--ink);
    background: var(--surface);
}
.mono { font-family: var(--font-mono); font-size: 12.5px; }
button.submit-btn {
    width: 100%;
    height: 42px;
    border: 1px solid var(--ink);
    background: var(--ink);
    color: #FFFFFF;
    border-radius: 9px;
    font-weight: 600;
    font-size: 13.5px;
    cursor: pointer;
    margin-top: 4px;
    transition: transform 0.15s, opacity 0.15s;
}
button.submit-btn:hover { opacity: 0.92; transform: translateY(-1px); }
.error {
    background: #FDECEC;
    border: 1px solid #F3B8B8;
    color: var(--red);
    padding: 10px 12px;
    border-radius: 8px;
    font-size: 12px;
    margin-bottom: 16px;
}
.footer {
    margin-top: 18px;
    font-size: 12px;
    color: var(--muted);
    text-align: center;
}
.footer a { color: var(--ink); text-decoration: none; font-weight: 600; }
.footer a:hover { text-decoration: underline; }
@media (max-width: 520px) { .grid-2, .mode-tabs { grid-template-columns: 1fr; } }
</style>
</head>
<body>
<form class="shell" method="POST" action="/signup" autocomplete="off">
    <div class="brand-row">
        <div class="brand">
            <div class="brand-mark">G</div>
            <div>
                <div class="brand-name">GARUDA</div>
                <div class="brand-sub">Epistemic Software Intelligence &amp; Active Governance</div>
            </div>
        </div>
        <span class="mono" style="font-size:11px; color:var(--muted);">v1.9.9 Multi-Tenant RBAC</span>
    </div>

    <h1 id="form-heading">Provision your organization</h1>
    <div class="subtitle" id="form-subtitle">Choose how you want to onboard—create an isolated Organization Tenant, join your company's existing Tenant as a Team Member, or launch a personal Developer Workflow.</div>

    <input type="hidden" id="account_type" name="account_type" value="{{ .AccountType }}">

    <div class="mode-tabs" role="tablist">
        <button type="button" class="mode-tab" data-mode="tenant_owner" onclick="setSignupMode('tenant_owner')">🏢 Tenant Owner</button>
        <button type="button" class="mode-tab" data-mode="team_member" onclick="setSignupMode('team_member')">👥 Join Team</button>
        <button type="button" class="mode-tab" data-mode="individual_dev" onclick="setSignupMode('individual_dev')">⚡ Individual Dev</button>
    </div>

    <div class="mode-banner" id="mode-banner">
        <strong>Organization / Tenant Owner:</strong> Creates a dedicated, cryptographically isolated Tenant ID + Shareable Employee Invite Code, and unlocks the full 11-surface Executive &amp; Governance Dashboard.
    </div>

    {{ if .Error }}<div class="error">{{ .Error }}</div>{{ end }}

    <!-- Tenant Owner Fields -->
    <div id="section-tenant-owner">
        <div class="grid-2">
            <div>
                <label for="org_name">Organization / Company Name</label>
                <input type="text" id="org_name" name="org_name" value="{{ .OrgName }}" placeholder="e.g., Garuda Labs or Acme Corp">
            </div>
            <div>
                <label for="org_domain">Organization Domain / Slug</label>
                <input type="text" id="org_domain" name="org_domain" value="{{ .OrgDomain }}" placeholder="e.g., garuda.ai">
            </div>
        </div>
    </div>

    <!-- Team Member / Employee Join Fields -->
    <div id="section-team-member" style="display:none;">
        <label for="tenant_join_code">Organization Tenant ID or Invite Code (from your Tenant Owner)</label>
        <input type="text" id="tenant_join_code" name="tenant_join_code" class="mono" value="{{ .TenantJoinCode }}" placeholder="e.g., GRD-GARUDA1 or Tenant UUID">
    </div>

    <!-- Workspace Name (Used by Tenant Owner & Individual Dev) -->
    <div id="section-workspace-name">
        <label for="workspace_name" id="workspace-label">Initial Workspace Name</label>
        <input type="text" id="workspace_name" name="workspace_name" class="mono" value="{{ .WorkspaceName }}" placeholder="default">
    </div>

    <!-- Shared Identity Fields -->
    <div class="grid-2">
        <div>
            <label for="full_name">Full Name</label>
            <input type="text" id="full_name" name="full_name" value="{{ .FullName }}" placeholder="Rohit Mishra" required>
        </div>
        <div>
            <label for="username">Username (@handle)</label>
            <input type="text" id="username" name="username" class="mono" value="{{ .Username }}" placeholder="rohit_dev" required>
        </div>
    </div>

    <label for="email">Work or Developer Email</label>
    <input type="email" id="email" name="email" value="{{ .Email }}" placeholder="you@company.com" required>

    <div class="grid-2">
        <div>
            <label for="password">Password (min 8 chars)</label>
            <input type="password" id="password" name="password" autocomplete="new-password" required>
        </div>
        <div>
            <label for="confirm">Confirm Password</label>
            <input type="password" id="confirm" name="confirm" autocomplete="new-password" required>
        </div>
    </div>

    <button type="submit" class="submit-btn" id="submit-btn">Create Organization Tenant &amp; Enter Dashboard →</button>

    <div class="footer">
        Already have an account? <a href="/login">Sign in</a>
    </div>
</form>

<script>
function setSignupMode(mode) {
    document.getElementById("account_type").value = mode;
    document.querySelectorAll(".mode-tab").forEach(function(btn) {
        btn.classList.toggle("active", btn.getAttribute("data-mode") === mode);
    });
    var secOwner = document.getElementById("section-tenant-owner");
    var secMember = document.getElementById("section-team-member");
    var secWs = document.getElementById("section-workspace-name");
    var wsLabel = document.getElementById("workspace-label");
    var wsInput = document.getElementById("workspace_name");
    var heading = document.getElementById("form-heading");
    var banner = document.getElementById("mode-banner");
    var submitBtn = document.getElementById("submit-btn");

    if (mode === "team_member") {
        secOwner.style.display = "none";
        secMember.style.display = "block";
        secWs.style.display = "none";
        heading.textContent = "Join your team workspace";
        banner.innerHTML = "<strong>Team Member / Employee Flow:</strong> Minimal steps—enter your Organization Invite Code (e.g. <code>GRD-GARUDA1</code>), name, @username, email, and password. Grants scoped Developer Workbench access without exposing Tenant Owner executive controls.";
        submitBtn.textContent = "Join Team Workspace →";
    } else if (mode === "individual_dev") {
        secOwner.style.display = "none";
        secMember.style.display = "none";
        secWs.style.display = "block";
        wsLabel.textContent = "Personal Developer Workflow / Workspace Name";
        if (wsInput.value === "default") wsInput.value = "my-dev-workflow";
        heading.textContent = "Launch your developer workflow";
        banner.innerHTML = "<strong>Individual Developer Mode:</strong> Provisions an isolated personal sandbox workspace to index repositories, run Talon pre-edit/post-edit checks, and inspect call-graph blast radius.";
        submitBtn.textContent = "Create Developer Workflow →";
    } else {
        secOwner.style.display = "block";
        secMember.style.display = "none";
        secWs.style.display = "block";
        wsLabel.textContent = "Initial Organization Workspace Name";
        if (wsInput.value === "my-dev-workflow") wsInput.value = "default";
        heading.textContent = "Provision your organization";
        banner.innerHTML = "<strong>Organization / Tenant Owner:</strong> Creates a dedicated, cryptographically isolated Tenant ID + Shareable Employee Invite Code, and unlocks the full 11-surface Executive &amp; Governance Dashboard.";
        submitBtn.textContent = "Create Organization Tenant & Enter Dashboard →";
    }
}
setSignupMode(document.getElementById("account_type").value || "tenant_owner");
</script>
</body>
</html>`
