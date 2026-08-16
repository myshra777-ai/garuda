package graph

// Node represents an entity in the graph.
type Node struct {
	ID       string `json:"id"`
	name     string `json:"name"`
	Label    string `json:"label"`
	Kind     string `json:"kind"`
	Package  string `json:"package"`
	File     string `json:"file"`
	Exported bool   `json:"exported"`
}

// Edge represents a relationship between two nodes.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}
