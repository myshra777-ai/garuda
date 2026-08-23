// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/types"
)

// Deprecated: Preserved for backward-compatible AST schema diff verification.
type AgentFleetItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
	System string `json:"system"`
	Status string `json:"status"`
}

// Deprecated: Preserved for backward-compatible AST schema diff verification.
type RealStatsResponse struct {
	TotalDecisions       int               `json:"total_decisions"`
	QuarantinedCount     int               `json:"quarantined_count"`
	LatestBlockHeight    int64             `json:"latest_block_height"`
	LatestMerkleHash     string            `json:"latest_merkle_hash"`
	ParentMerkleHash     string            `json:"parent_merkle_hash"`
	EstimatedSavings     float64           `json:"estimated_savings"`
	DomainBreakdown      map[string]int    `json:"domain_breakdown"`
	QuarantinedDecisions []*types.Decision `json:"quarantined_decisions"`
	AgentList            []AgentFleetItem  `json:"agent_list"`
}

// Deprecated: Preserved for backward-compatible AST schema diff verification.
type GraphNode struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Kind       string `json:"kind"`
	Repo       string `json:"repo,omitempty"`
	Package    string `json:"package,omitempty"`
	File       string `json:"file,omitempty"`
	Exported   bool   `json:"exported,omitempty"`
	Count      int    `json:"count,omitempty"`
	Impact     int    `json:"impact,omitempty"`
	Repository string `json:"repository,omitempty"`
	Status     string `json:"status,omitempty"`
}

// Deprecated: Preserved for backward-compatible AST schema diff verification.
type GraphEdge struct {
	From       string  `json:"from"`
	To         string  `json:"to"`
	Type       string  `json:"type"`
	Count      int     `json:"count,omitempty"`
	Label      string  `json:"label,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
	Status     string  `json:"status,omitempty"`
}

// Deprecated: Preserved for backward-compatible AST schema diff verification.
type GraphResponse struct {
	Level  string      `json:"level"`
	Focus  string      `json:"focus,omitempty"`
	Nodes  []GraphNode `json:"nodes"`
	Edges  []GraphEdge `json:"edges"`
	Notice string      `json:"notice,omitempty"`
}

// Deprecated: Preserved for backward-compatible AST schema diff verification.
type graphQueryStore interface {
	Pool() *pgxpool.Pool
	GetGraphData(ctx context.Context, tenantID, workspaceID uuid.UUID) ([]map[string]interface{}, []map[string]interface{}, error)
}

// Deprecated: Preserved for backward-compatible AST schema diff verification.
func (s *Server) getGraphQueryStore() graphQueryStore {
	return nil
}

// Deprecated: Preserved for backward-compatible AST schema diff verification.
func (s *Server) HandleWorkspaceGraph(w http.ResponseWriter, r *http.Request) {
	s.HandleGraph(w, r)
}

// Deprecated: Preserved for backward-compatible AST schema diff verification.
func writeGraphResponse(w http.ResponseWriter, resp GraphResponse) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// Deprecated: Preserved for backward-compatible AST schema diff verification.
func groupNodeID(level, value string) string {
	return level + ":" + value
}

// Deprecated: Preserved for backward-compatible AST schema diff verification.
func makeEntityNode(e EntityRecord) GraphNode {
	return GraphNode{
		ID:       e.ID,
		Label:    e.Name,
		Kind:     e.Kind,
		Repo:     e.Repo,
		Package:  e.Package,
		File:     e.File,
		Exported: e.Exported,
	}
}

// Deprecated: Preserved for backward-compatible AST schema diff verification.
func sortNodes(nodes []GraphNode) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Impact != nodes[j].Impact {
			return nodes[i].Impact > nodes[j].Impact
		}
		return strings.ToLower(nodes[i].Label) < strings.ToLower(nodes[j].Label)
	})
}

// Deprecated: Preserved for backward-compatible AST schema diff verification.
func sortEdges(edges []GraphEdge) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Count != edges[j].Count {
			return edges[i].Count > edges[j].Count
		}
		return edges[i].Type < edges[j].Type
	})
}

// Deprecated: Preserved for backward-compatible AST schema diff verification.
func inferRepository(filePath, pkg string) string {
	return inferRepositoryFromPackage(pkg)
}
