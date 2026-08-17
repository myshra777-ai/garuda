package main

import (
	"strings"
	"testing"

	"github.com/myshra777-ai/garuda/internal/graph"
)

func TestGenerateGraphHTMLIncludesGraphData(t *testing.T) {
	nodes := []graph.Node{
		{ID: "n1", Label: "MyType", Kind: "struct", File: "main.go", Package: "example.com/demo"},
	}
	edges := []graph.Edge{{From: "n1", To: "n2", Type: "calls"}}

	html, err := generateGraphHTML("demo", nodes, edges)
	if err != nil {
		t.Fatalf("generateGraphHTML failed: %v", err)
	}
	if html == "" {
		t.Fatal("generateGraphHTML returned empty HTML")
	}
	if !strings.Contains(html, "MyType") {
		t.Fatal("graph HTML did not include node labels")
	}
	if !strings.Contains(html, "calls") {
		t.Fatal("graph HTML did not include relationship metadata")
	}
	if !strings.Contains(html, "d3.v7.min.js") {
		t.Fatal("graph HTML did not include D3.js")
	}
}

func TestGenerateGraphHTMLEmbedsJSONPayload(t *testing.T) {
	nodes := []graph.Node{
		{ID: "n1", Label: "MyType", Kind: "struct", File: "main.go", Package: "example.com/demo", Exported: true},
	}
	edges := []graph.Edge{{From: "n1", To: "n2", Type: "calls"}}

	html, err := generateGraphHTML("demo", nodes, edges)
	if err != nil {
		t.Fatalf("generateGraphHTML failed: %v", err)
	}
	if !strings.Contains(html, `"label":"MyType"`) {
		t.Fatal("graph HTML did not embed node JSON")
	}
	if !strings.Contains(html, `"type":"calls"`) {
		t.Fatal("graph HTML did not embed edge JSON")
	}
}
