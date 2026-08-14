package graph

import (
	"bytes"
	"encoding/json"
	"html/template"
)

// Generate returns an HTML page containing the interactive graph.
func Generate(workspaceName string, nodes []Node, edges []Edge) (string, error) {
	data := struct {
		Workspace string
		DataJSON  string
	}{
		Workspace: workspaceName,
	}

	// Marshal nodes and edges into JSON for the JavaScript variable
	payload := struct {
		Nodes []Node `json:"Nodes"`
		Edges []Edge `json:"Edges"`
	}{
		Nodes: nodes,
		Edges: edges,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	data.DataJSON = string(jsonData)

	// Parse and execute the template
	tmpl, err := template.New("graph").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
