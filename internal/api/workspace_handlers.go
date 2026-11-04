// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/store"
)

type workspaceListItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// TeamMemberDTO represents a user belonging to the current tenant.
type TeamMemberDTO struct {
	UserID      string `json:"user_id"`
	FullName    string `json:"full_name"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	AccountType string `json:"account_type"`
	JoinedAt    string `json:"joined_at"`
}

// HandleListWorkspaces returns every workspace the session user is a member of,
// scoped to their active tenant so cross-tenant workspaces never bleed across accounts.
func (s *Server) HandleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	applySecurityHeaders(w)
	ctx := r.Context()

	sess, ok := SessionFromContext(ctx)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "authentication required", nil)
		return
	}
	userID, err := uuid.Parse(sess.UserID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "invalid session user_id", nil)
		return
	}

	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "store unavailable", nil)
		return
	}

	list, err := pgStore.ListUserWorkspaces(ctx, userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "list workspaces failed", map[string]any{"detail": err.Error()})
		return
	}

	out := make([]workspaceListItem, 0, len(list))
	for _, uw := range list {
		out = append(out, workspaceListItem{
			ID:   uw.ID.String(),
			Name: uw.Name,
			Role: uw.Role,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// HandleDashboardMe returns the currently authenticated user's identity,
// tenant details, shareable invite code (for Tenant Owners), and RBAC surface permissions.
func (s *Server) HandleDashboardMe(w http.ResponseWriter, r *http.Request) {
	applySecurityHeaders(w)
	ctx := r.Context()

	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "store unavailable", nil)
		return
	}

	scope, err := workspaceScopeFromRequest(ctx, r, pgStore)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "session scope unavailable", nil)
		return
	}

	// Load all members in the current tenant
	var members []TeamMemberDTO
	rows, qErr := pgStore.Pool().Query(ctx, `
		SELECT u.id::text, COALESCE(u.full_name, ''), COALESCE(u.username, ''), u.email,
		       tm.role, COALESCE(u.account_type, 'tenant_owner'), tm.created_at
		  FROM tenant_members tm
		  JOIN users u ON u.id = tm.user_id
		 WHERE tm.tenant_id = $1
		 ORDER BY (tm.role = 'owner') DESC, tm.created_at ASC
	`, scope.TenantID)
	if qErr == nil {
		for rows.Next() {
			var m TeamMemberDTO
			var joined time.Time
			if rows.Scan(&m.UserID, &m.FullName, &m.Username, &m.Email, &m.Role, &m.AccountType, &joined) == nil {
				m.JoinedAt = joined.UTC().Format(time.RFC3339)
				members = append(members, m)
			}
		}
		rows.Close()
	}

	allowedSurfaces := []string{"about", "workspace", "search", "graph", "docsgraph", "talon", "review", "governance", "agents", "audit", "decisions"}
	if !scope.IsTenantOwner && scope.AccountType == "team_member" {
		// Team Members / Employees get a focused Developer Workbench without Tenant-Owner executive/governance surfaces
		allowedSurfaces = []string{"workspace", "search", "graph", "talon", "review", "agents"}
	}

	inviteCode := ""
	if scope.IsTenantOwner {
		inviteCode = scope.TenantInviteCode
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user_id":          scope.UserID.String(),
		"email":            scope.UserEmail,
		"full_name":        scope.UserFullName,
		"username":         scope.Username,
		"account_type":     scope.AccountType,
		"session_role":     scope.SessionRole,
		"workspace_role":   scope.WorkspaceRole,
		"is_tenant_owner":  scope.IsTenantOwner,
		"tenant_id":        scope.TenantID.String(),
		"tenant_name":      scope.TenantName,
		"invite_code":      inviteCode,
		"workspace_id":     scope.WorkspaceID.String(),
		"workspace":        scope.WorkspaceName,
		"allowed_surfaces": allowedSurfaces,
		"team_members":     members,
		"team_count":       len(members),
	})
}
