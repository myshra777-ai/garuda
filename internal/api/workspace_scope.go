// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/myshra777-ai/garuda/internal/store"
)

// ErrNoWorkspaceAccess is returned by resolveWorkspaceForUser when
// the requested workspace is not in the session user's workspace_members
// set, or when the user has no workspaces at all. Callers translate
// this into HTTP 404. It is never a 403: a 403 would confirm that the
// workspace exists but the caller cannot see it, which is a
// cross-tenant existence leak.
var ErrNoWorkspaceAccess = errors.New("no workspace access for this user")

// WorkspaceScope is the resolved scope of one request: which user is
// making it, which workspace they are operating on, and which tenant
// that workspace belongs to.
//
// The tenant is not carried in the session. It is derived per request
// from the workspace, which is why a user can belong to more than one
// tenant without a session switch: the workspace name in the URL
// decides which tenant is checked.
type WorkspaceScope struct {
	UserID        uuid.UUID
	UserEmail     string
	SessionRole   string // from the session, not the workspace
	WorkspaceID   uuid.UUID
	WorkspaceName string // canonical name after resolution
	TenantID      uuid.UUID
	WorkspaceRole string // owner | admin | member, from workspace_members
}

// resolveWorkspaceForUser returns the WorkspaceScope for the given
// user and workspace name.
//
// If workspaceName is empty, the user's most recently updated
// workspace is selected. That is the "no explicit choice" path used
// by routes that do not carry a workspace in the URL.
//
// If workspaceName is non-empty, the workspace is matched by name
// against every workspace the user is a member of. A workspace with
// that name may exist in a different tenant, or in the same tenant
// under a team the user is not in; both produce ErrNoWorkspaceAccess.
//
// The query joins workspace_members, so membership is enforced at the
// database level. A workspace row that exists but has no
// workspace_members row for this user is invisible to this function.
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
	)

	if workspaceName == "" {
		err := pgStore.Pool().QueryRow(ctx, `
			SELECT w.id, w.name, w.tenant_id, wm.role
			  FROM workspaces w
			  JOIN workspace_members wm ON wm.workspace_id = w.id
			 WHERE wm.user_id = $1
			 ORDER BY w.updated_at DESC, w.created_at DESC, w.id DESC
			 LIMIT 1
		`, userID).Scan(&workspaceID, &resolvedName, &tenantID, &workspaceRole)
		if err != nil {
			return nil, ErrNoWorkspaceAccess
		}
	} else {
		err := pgStore.Pool().QueryRow(ctx, `
			SELECT w.id, w.name, w.tenant_id, wm.role
			  FROM workspaces w
			  JOIN workspace_members wm ON wm.workspace_id = w.id
			 WHERE wm.user_id = $1 AND w.name = $2
			 ORDER BY w.created_at ASC, w.id ASC
			 LIMIT 1
		`, userID, workspaceName).Scan(&workspaceID, &resolvedName, &tenantID, &workspaceRole)
		if err != nil {
			return nil, ErrNoWorkspaceAccess
		}
	}

	return &WorkspaceScope{
		UserID:        userID,
		UserEmail:     userEmail,
		SessionRole:   sessionRole,
		WorkspaceID:   workspaceID,
		WorkspaceName: resolvedName,
		TenantID:      tenantID,
		WorkspaceRole: workspaceRole,
	}, nil
}

// workspaceScopeFromRequest reads the session from ctx, reads the
// ?workspace= parameter from r, and resolves both into a scope.
//
// The session must already be in ctx — RequireSession puts it there.
// If the session is missing, this returns ErrNoWorkspaceAccess, which
// the caller treats as 404. That is conservative on purpose: a
// request that reached a session-protected handler without a session
// is a bug, and refusing to serve data is the right response.
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

// writeWorkspaceNotFound sends a 404 in the shape the caller expects:
// JSON for API paths, HTML for browser paths. The response body never
// distinguishes "workspace does not exist" from "workspace exists but
// you are not a member". Both produce the same message.
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
