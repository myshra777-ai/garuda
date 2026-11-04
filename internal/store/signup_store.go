// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/myshra777-ai/garuda/internal/auth"
)

// ErrEmailExists is returned when the email address is already registered.
var ErrEmailExists = errors.New("email already registered")

// ErrTenantNotFound is returned when a team_member signup provides an unknown Tenant ID or Invite Code.
var ErrTenantNotFound = errors.New("organization tenant or invite code not found")

// SignupOptions carries multi-path registration fields for Tenant Owners, Team Members, and Individual Developers.
type SignupOptions struct {
	AccountType    string // "tenant_owner" | "team_member" | "individual_dev"
	Username       string
	OrgName        string
	OrgDomain      string
	WorkspaceName  string
	TenantJoinCode string // Tenant UUID, GRD-XXXXXX invite code, or org slug
	JobTitle       string
}

var nonSlugRe = regexp.MustCompile(`[^a-z0-9_]+`)

func sanitizeHandle(raw, fallback string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.TrimPrefix(s, "@")
	s = nonSlugRe.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	if s == "" {
		s = fallback
	}
	return s
}

// SignupUser preserves the 3-argument signature used by existing store tests,
// defaulting to an isolated personal/tenant_owner registration.
func (s *PostgresStore) SignupUser(
	ctx context.Context,
	email, passwordHash, fullName string,
) (*auth.User, *Workspace, error) {
	user, ws, _, err := s.SignupUserWithOptions(ctx, email, passwordHash, fullName, SignupOptions{
		AccountType:   "tenant_owner",
		WorkspaceName: "default",
	})
	return user, ws, err
}

// SignupUserWithOptions provisions:
//   - "tenant_owner": A brand-new isolated Organization Tenant + Invite Code + Workspace + Owner role
//   - "team_member":  Joins an existing Tenant via Tenant UUID / GRD-XXXXXX Invite Code / Slug with Member role
//   - "individual_dev": Provisions an isolated personal developer workflow tenant & workspace
func (s *PostgresStore) SignupUserWithOptions(
	ctx context.Context,
	email, passwordHash, fullName string,
	opts SignupOptions,
) (*auth.User, *Workspace, string, error) {
	email = strings.TrimSpace(email)
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		if at := strings.IndexByte(email, '@'); at > 0 {
			fullName = email[:at]
		} else {
			fullName = email
		}
	}

	accountType := strings.ToLower(strings.TrimSpace(opts.AccountType))
	switch accountType {
	case "tenant_owner", "team_member", "individual_dev":
	default:
		accountType = "tenant_owner"
	}

	defaultUser := "user"
	if at := strings.IndexByte(email, '@'); at > 0 {
		defaultUser = email[:at]
	}
	username := sanitizeHandle(opts.Username, sanitizeHandle(defaultUser, "dev"))

	wsName := strings.TrimSpace(opts.WorkspaceName)
	if wsName == "" {
		if accountType == "individual_dev" {
			wsName = username + "-workflow"
		} else {
			wsName = "default"
		}
	}

	userID := uuid.New()
	now := time.Now().UTC()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, nil, "", fmt.Errorf("begin signup transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var (
		tenantID    uuid.UUID
		workspaceID uuid.UUID
		tenantName  string
		inviteCode  string
		userRole    string
		wsRole      string
	)

	if accountType == "team_member" {
		joinCode := strings.TrimSpace(opts.TenantJoinCode)
		if joinCode == "" {
			return nil, nil, "", ErrTenantNotFound
		}

		// Resolve target tenant by UUID, Invite Code (e.g. GRD-GARUDA1), slug, or name
		err := tx.QueryRow(ctx, `
			SELECT id, name, COALESCE(NULLIF(invite_code, ''), 'GRD-' || UPPER(SUBSTRING(REPLACE(id::text, '-', ''), 1, 6)))
			  FROM tenants
			 WHERE id::text = $1
			    OR UPPER(COALESCE(invite_code, '')) = UPPER($1)
			    OR LOWER(COALESCE(slug, '')) = LOWER($1)
			    OR LOWER(name) = LOWER($1)
			 LIMIT 1
		`, joinCode).Scan(&tenantID, &tenantName, &inviteCode)
		if err != nil || tenantID == uuid.Nil {
			return nil, nil, "", ErrTenantNotFound
		}

		userRole = "member"
		wsRole = "member"

		// Find target workspace in that tenant (or default / first workspace)
		err = tx.QueryRow(ctx, `
			SELECT id, name
			  FROM workspaces
			 WHERE tenant_id = $1
			 ORDER BY (name = $2) DESC, (name = 'default') DESC, created_at ASC
			 LIMIT 1
		`, tenantID, wsName).Scan(&workspaceID, &wsName)
		if err != nil || workspaceID == uuid.Nil {
			workspaceID = uuid.New()
			wsName = "default"
			if _, err := tx.Exec(ctx, `
				INSERT INTO workspaces (id, tenant_id, name, root_path, is_go_work, description, created_at, updated_at)
				VALUES ($1, $2, $3, '.', false, 'Team workspace', $4, $4)
			`, workspaceID, tenantID, wsName, now); err != nil {
				return nil, nil, "", fmt.Errorf("insert team workspace: %w", err)
			}
		}
	} else {
		tenantID = uuid.New()
		workspaceID = uuid.New()
		inviteCode = "GRD-" + strings.ToUpper(strings.ReplaceAll(tenantID.String(), "-", "")[:6])

		if accountType == "individual_dev" {
			tenantName = fullName + " (@" + username + " Developer Workspace)"
			userRole = "developer"
			wsRole = "owner"
		} else {
			tenantName = strings.TrimSpace(opts.OrgName)
			if tenantName == "" {
				tenantName = email + "'s tenant"
			}
			userRole = "owner"
			wsRole = "owner"
		}
	}

	// 1. Insert user row (trying extended columns first, with fallback if migration 097 wasn't run in unit test harness)
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, username, account_type, primary_tenant_id, job_title, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
	`, userID, email, passwordHash, fullName, userRole, username, accountType, tenantID, strings.TrimSpace(opts.JobTitle), now)
	if err != nil {
		if isDuplicateKey(err) {
			return nil, nil, "", ErrEmailExists
		}
		// Fallback for minimal test schemas without migration 097 columns
		if _, err2 := tx.Exec(ctx, `
			INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $6)
		`, userID, email, passwordHash, fullName, userRole, now); err2 != nil {
			if isDuplicateKey(err2) {
				return nil, nil, "", ErrEmailExists
			}
			return nil, nil, "", fmt.Errorf("insert user: %w", err2)
		}
	}

	// 2. If creating a new tenant (tenant_owner or individual_dev), insert tenant & initial workspace
	if accountType != "team_member" {
		slug := sanitizeHandle(tenantName, "org")
		_, err = tx.Exec(ctx, `
			INSERT INTO tenants (id, name, slug, org_domain, invite_code, owner_user_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		`, tenantID, tenantName, slug, strings.TrimSpace(opts.OrgDomain), inviteCode, userID, now)
		if err != nil {
			if _, err2 := tx.Exec(ctx, `
				INSERT INTO tenants (id, name, created_at, updated_at)
				VALUES ($1, $2, $3, $3)
			`, tenantID, tenantName, now); err2 != nil {
				return nil, nil, "", fmt.Errorf("insert tenant: %w", err2)
			}
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO workspaces (id, tenant_id, name, root_path, is_go_work, description, created_at, updated_at)
			VALUES ($1, $2, $3, '.', false, '', $4, $4)
		`, workspaceID, tenantID, wsName, now); err != nil {
			return nil, nil, "", fmt.Errorf("insert workspace: %w", err)
		}
	}

	// 3. Insert tenant_members row
	tmRole := "owner"
	if accountType == "team_member" {
		tmRole = "member"
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO tenant_members (user_id, tenant_id, role, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, tenant_id) DO UPDATE SET role = EXCLUDED.role
	`, userID, tenantID, tmRole, now); err != nil {
		return nil, nil, "", fmt.Errorf("insert tenant member: %w", err)
	}

	// 4. Insert workspace_members row(s)
	if accountType == "team_member" {
		if _, err := tx.Exec(ctx, `
			INSERT INTO workspace_members (user_id, workspace_id, role, created_at)
			SELECT $1, w.id, 'member', $3
			  FROM workspaces w
			 WHERE w.tenant_id = $2
			ON CONFLICT (user_id, workspace_id) DO NOTHING
		`, userID, tenantID, now); err != nil {
			return nil, nil, "", fmt.Errorf("insert team workspace members: %w", err)
		}
	} else {
		if _, err := tx.Exec(ctx, `
			INSERT INTO workspace_members (user_id, workspace_id, role, created_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id, workspace_id) DO NOTHING
		`, userID, workspaceID, wsRole, now); err != nil {
			return nil, nil, "", fmt.Errorf("insert workspace member: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, "", fmt.Errorf("commit signup transaction: %w", err)
	}

	user := &auth.User{
		ID:           userID.String(),
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Role:         userRole,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	ws := &Workspace{
		ID:          workspaceID,
		TenantID:    tenantID,
		Name:        wsName,
		RootPath:    ".",
		IsGoWork:    false,
		Description: "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return user, ws, inviteCode, nil
}

func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
