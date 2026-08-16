package graph

import (
	"encoding/json"
	"html/template"
	"strings"
)

// TemplateData holds data for the HTML template.
type TemplateData struct {
	Workspace string
	NodesJSON template.JS
	EdgesJSON template.JS
}

// Generate produces the full HTML for the graph.
func Generate(workspaceName string, nodes []Node, edges []Edge) (string, error) {
	nodesJSON, err := json.Marshal(nodes)
	if err != nil {
		return "", err
	}
	edgesJSON, err := json.Marshal(edges)
	if err != nil {
		return "", err
	}

	// Use template.JS to safely embed JSON in JavaScript.
	data := TemplateData{
		Workspace: workspaceName,
		NodesJSON: template.JS(nodesJSON),
		EdgesJSON: template.JS(edgesJSON),
	}

	tmpl, err := template.New("graph").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", err
	}
	return sb.String(), nil
}
