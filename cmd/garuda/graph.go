package main

type GraphNode struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Group   string `json:"group"` // struct, interface, etc.
	File    string `json:"file"`
	Package string `json:"package"`
}

type GraphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

// generateGraphHTML creates an HTML file with D3.js visualization
