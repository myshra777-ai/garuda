// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package knowledge

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GraphNode struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Type   string `json:"type"` // document, claim, entity
	Status string `json:"status,omitempty"`
}

type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Kind   string `json:"kind"` // asserts, matches, calls, violates
}

type UnifiedGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type GraphService struct {
	pool *pgxpool.Pool
}

func NewGraphService(pool *pgxpool.Pool) *GraphService {
	return &GraphService{pool: pool}
}

func (g *GraphService) BuildUnifiedGraph(ctx context.Context, tenantID, workspaceID uuid.UUID) (*UnifiedGraph, error) {
	graph := &UnifiedGraph{}
	nodeMap := make(map[string]bool)

	addNode := func(id, label, nType, status string) {
		if !nodeMap[id] {
			graph.Nodes = append(graph.Nodes, GraphNode{ID: id, Label: label, Type: nType, Status: status})
			nodeMap[id] = true
		}
	}

	// 1. Fetch Entities (AST)
	entityRows, err := g.pool.Query(ctx, `
		SELECT id, name, kind FROM entities WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID)
	if err == nil {
		defer entityRows.Close()
		for entityRows.Next() {
			var id uuid.UUID
			var name, kind string
			if entityRows.Scan(&id, &name, &kind) == nil {
				addNode(id.String(), name, "entity", "active")
			}
		}
	}

	// 2. Fetch Document Claims & Map Edges
	claimRows, err := g.pool.Query(ctx, `
		SELECT id, subject, predicate, object, modality, status, matched_entity_id 
		FROM document_claims WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID)
	if err == nil {
		defer claimRows.Close()
		for claimRows.Next() {
			var claimID, subj, pred, obj, mod, status string
			var matchedEntityID *uuid.UUID
			if claimRows.Scan(&claimID, &subj, &pred, &obj, &mod, &status, &matchedEntityID) == nil {
				claimLabel := fmt.Sprintf("[%s] %s %s %s", mod, subj, pred, obj)
				addNode(claimID, claimLabel, "claim", status)

				// Edge from Claim -> Subject Entity if matched
				if matchedEntityID != nil {
					graph.Edges = append(graph.Edges, GraphEdge{
						Source: claimID,
						Target: matchedEntityID.String(),
						Kind:   "asserts",
					})
				}
			}
		}
	}

	// 3. Fetch AST Relationships (Calls, etc.)
	relRows, err := g.pool.Query(ctx, `
		SELECT source_id, target_id, kind FROM relationships WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID)
	if err == nil {
		defer relRows.Close()
		for relRows.Next() {
			var src, tgt uuid.UUID
			var kind string
			if relRows.Scan(&src, &tgt, &kind) == nil {
				graph.Edges = append(graph.Edges, GraphEdge{
					Source: src.String(),
					Target: tgt.String(),
					Kind:   kind,
				})
			}
		}
	}

	return graph, nil
}
