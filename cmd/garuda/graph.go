package main

import (
	"github.com/myshra777-ai/garuda/internal/graph"
)

// generateGraphHTML creates an HTML file with an interactive D3.js visualization.
func generateGraphHTML(workspaceName string, nodes []graph.Node, edges []graph.Edge) (string, error) {
	return graph.Generate(workspaceName, nodes, edges)
}
