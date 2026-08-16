package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/auth"
)

// graphQueryStore is the Postgres-backed capability required by workspace graph APIs.
type graphQueryStore interface {
	Pool() *pgxpool.Pool
	GetGraphData(ctx context.Context, tenantID, workspaceID uuid.UUID) ([]map[string]interface{}, []map[string]interface{}, error)
}

func (s *Server) getGraphQueryStore() (graphQueryStore, bool) {
	gs, ok := s.store.(graphQueryStore)
	return gs, ok
}

// HandleWorkspaceGraph returns nodes and edges for a workspace in the dashboard-friendly schema.
func (s *Server) HandleWorkspaceGraph(w http.ResponseWriter, r *http.Request) {
	gstore, ok := s.getGraphQueryStore()
	if !ok || gstore == nil {
		s.RespondWithError(w, http.StatusServiceUnavailable, "workspace graph storage unavailable", "")
		return
	}

	tenantID, ok := auth.TenantIDFromContext(r.Context())
	if !ok {
		tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}

	workspaceName := r.URL.Query().Get("workspace")
	if workspaceName == "" {
		workspaceName = "my-workspace"
	}

	var wsID uuid.UUID
	if err := gstore.Pool().QueryRow(r.Context(), `
		SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2
	`, tenantID, workspaceName).Scan(&wsID); err != nil {
		s.RespondWithError(w, http.StatusNotFound, "workspace not found", "")
		return
	}

	rows, err := gstore.Pool().Query(r.Context(), `
		SELECT id, name, kind, package, file_path, is_exported
		FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, wsID)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, "failed to query entities", "")
		return
	}
	defer rows.Close()

	nodes := make([]GraphNode, 0)
	for rows.Next() {
		var id, name, kind, pkg, file string
		var exported bool
		if err := rows.Scan(&id, &name, &kind, &pkg, &file, &exported); err != nil {
			continue
		}
		nodes = append(nodes, GraphNode{
			ID:       id,
			Label:    name,
			Kind:     kind,
			Package:  pkg,
			File:     file,
			Exported: exported,
		})
	}

	rows2, err := gstore.Pool().Query(r.Context(), `
		SELECT from_entity_id, to_entity_id, claim_type
		FROM claims
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, wsID)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, "failed to query claims", "")
		return
	}
	defer rows2.Close()

	edges := make([]GraphEdge, 0)
	for rows2.Next() {
		var from, to, typ string
		if err := rows2.Scan(&from, &to, &typ); err != nil {
			continue
		}
		edges = append(edges, GraphEdge{From: from, To: to, Type: typ})
	}

	response := GraphResponse{Nodes: nodes, Edges: edges}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
