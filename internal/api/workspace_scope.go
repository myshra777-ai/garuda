// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/myshra777-ai/garuda/internal/store"
)

// ErrNoWorkspaceAccess is returned by resolveWorkspaceForUser when
// the requested workspace is not in the session user's workspace_members
// set, or when the user has no workspaces at all.
var ErrNoWorkspaceAccess = errors.New("no workspace access for this user")

// WorkspaceScope is the resolved scope of one request: which user is
// making it, which workspace they are operating on, and which tenant
// that workspace belongs to, plus RBAC role and tenant invite metadata.
type WorkspaceScope struct {
	UserID           uuid.UUID
	UserEmail        string
	UserFullName     string
	Username         string
	AccountType      string // "tenant_owner" | "team_member" | "individual_dev"
	SessionRole      string // from users.role ("owner" | "admin" | "member" | "developer")
	WorkspaceID      uuid.UUID
	WorkspaceName    string
	TenantID         uuid.UUID
	TenantName       string
	TenantInviteCode string
	WorkspaceRole    string // "owner" | "admin" | "member", from workspace_members
	IsTenantOwner    bool
}

// resolveWorkspaceForUser returns the WorkspaceScope for the given
// user and workspace name, strictly prioritizing the user's primary_tenant_id
// so a newly registered user NEVER resolves to another tenant's workspace.
func resolveWorkspaceForUser(
	ctx context.Context,
	pgStore *store.PostgresStore,
	userID uuid.UUID,
	userEmail, sessionRole string,
	workspaceName string,
) (*WorkspaceScope, error) {
	var (
		workspaceID   uuid.UUID
		resolvedName  string
		tenantID      uuid.UUID
		workspaceRole string
		fullName      string
		username      string
		accountType   string
		primaryTenant *uuid.UUID
		tenantName    string
		inviteCode    string
	)

	// Load user extended metadata (with graceful fallback if columns don't exist in minimal unit tests)
	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COALESCE(full_name, ''), COALESCE(username, ''), COALESCE(account_type, 'tenant_owner'), primary_tenant_id
		  FROM users WHERE id = $1
	`, userID).Scan(&fullName, &username, &accountType, &primaryTenant)

	if workspaceName == "" {
		err := pgStore.Pool().QueryRow(ctx, `
			SELECT w.id, w.name, w.tenant_id, wm.role
			  FROM workspaces w
			  JOIN workspace_members wm ON wm.workspace_id = w.id
			 WHERE wm.user_id = $1
			 ORDER BY (CASE WHEN $2::uuid IS NOT NULL AND w.tenant_id = $2::uuid THEN 1 ELSE 0 END) DESC,
			          (CASE WHEN wm.role = 'owner' THEN 1 ELSE 0 END) DESC,
			          w.updated_at DESC, w.created_at DESC, w.id DESC
			 LIMIT 1
		`, userID, primaryTenant).Scan(&workspaceID, &resolvedName, &tenantID, &workspaceRole)
		if err != nil {
			return nil, ErrNoWorkspaceAccess
		}
	} else {
		err := pgStore.Pool().QueryRow(ctx, `
			SELECT w.id, w.name, w.tenant_id, wm.role
			  FROM workspaces w
			  JOIN workspace_members wm ON wm.workspace_id = w.id
			 WHERE wm.user_id = $1 AND w.name = $2
			 ORDER BY (CASE WHEN $3::uuid IS NOT NULL AND w.tenant_id = $3::uuid THEN 1 ELSE 0 END) DESC,
			          (CASE WHEN wm.role = 'owner' THEN 1 ELSE 0 END) DESC,
			          w.created_at DESC, w.id DESC
			 LIMIT 1
		`, userID, workspaceName, primaryTenant).Scan(&workspaceID, &resolvedName, &tenantID, &workspaceRole)
		if err != nil {
			return nil, ErrNoWorkspaceAccess
		}
	}

	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COALESCE(name, 'Organization Tenant'),
		       COALESCE(NULLIF(invite_code, ''), 'GRD-' || UPPER(SUBSTRING(REPLACE(id::text, '-', ''), 1, 6)))
		  FROM tenants WHERE id = $1
	`, tenantID).Scan(&tenantName, &inviteCode)

	if username == "" {
		if at := strings.IndexByte(userEmail, '@'); at > 0 {
			username = userEmail[:at]
		} else {
			username = userEmail
		}
	}
	if accountType == "" {
		if workspaceRole == "member" {
			accountType = "team_member"
		} else {
			accountType = "tenant_owner"
		}
	}

	isTenantOwner := (workspaceRole == "owner" || workspaceRole == "admin" || sessionRole == "owner" || sessionRole == "admin") && accountType != "team_member"

	return &WorkspaceScope{
		UserID:           userID,
		UserEmail:        userEmail,
		UserFullName:     fullName,
		Username:         username,
		AccountType:      accountType,
		SessionRole:      sessionRole,
		WorkspaceID:      workspaceID,
		WorkspaceName:    resolvedName,
		TenantID:         tenantID,
		TenantName:       tenantName,
		TenantInviteCode: inviteCode,
		WorkspaceRole:    workspaceRole,
		IsTenantOwner:    isTenantOwner,
	}, nil
}

func workspaceScopeFromRequest(
	ctx context.Context,
	r *http.Request,
	pgStore *store.PostgresStore,
) (*WorkspaceScope, error) {
	sess, ok := SessionFromContext(ctx)
	if !ok {
		return nil, ErrNoWorkspaceAccess
	}
	userID, err := uuid.Parse(sess.UserID)
	if err != nil {
		return nil, fmt.Errorf("session user_id is not a UUID: %w", err)
	}
	workspaceName := strings.TrimSpace(r.URL.Query().Get("workspace"))
	return resolveWorkspaceForUser(ctx, pgStore, userID, sess.UserEmail, sess.Role, workspaceName)
}

func writeWorkspaceNotFound(w http.ResponseWriter, r *http.Request, requested string) {
	if wantsHTML(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_ = parsedNotFoundTmpl.Execute(w, map[string]string{"Workspace": requested})
		return
	}
	writeJSONError(w, http.StatusNotFound, "workspace not found",
		map[string]any{"requested": requested})
}

// durableTenantStorageDir returns a persistent, owner-restricted (0700) storage
// directory scoped to (tenantID, workspaceID, category) under $GARUDA_DATA_DIR
// or ~/.garuda/tenant_storage, avoiding ephemeral /tmp storage for persistent
// source documents and approved rulebooks.
func durableTenantStorageDir(tenantID, workspaceID uuid.UUID, category string) string {
	base := strings.TrimSpace(os.Getenv("GARUDA_DATA_DIR"))
	if base == "" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			base = filepath.Join(home, ".garuda")
		} else {
			base = filepath.Join(os.TempDir(), "garuda-durable")
		}
	}
	cleanCat := filepath.Base(strings.TrimSpace(category))
	if cleanCat == "" || cleanCat == "." || cleanCat == ".." {
		cleanCat = "storage"
	}
	dir := filepath.Join(base, "tenant_storage", tenantID.String(), workspaceID.String(), cleanCat)
	_ = os.MkdirAll(dir, 0700)
	return dir
}
