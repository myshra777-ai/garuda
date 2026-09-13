// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/store"
)

// workspaceListItem is one entry in the workspace picker response.
// It carries only what the picker needs: an ID for future use, a
// name to render, and the caller's role. No tenant information, no
// repository count, no content.
type workspaceListItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// HandleListWorkspaces returns every workspace the session user is a
// member of. Used by the workspace picker on the dashboard sidebar.
//
// The list is scoped by workspace_members.user_id. A user in two
// workspaces sees two entries. A user in one sees one. A user in
// none sees an empty array, not an error.
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
