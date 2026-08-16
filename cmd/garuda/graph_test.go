package main

import (
	"strings"
	"testing"
)

func TestGenerateGraphHTMLIncludesGraphData(t *testing.T) {
	nodes := []GraphNode{
		{ID: "n1", Label: "MyType", Group: "struct", File: "main.go", Package: "example.com/demo"},
	}
	edges := []GraphEdge{{From: "n1", To: "n2", Type: "calls"}}

	html := generateGraphHTML("demo", nodes, edges)
	if html == "" {
		t.Fatal("generateGraphHTML returned empty HTML")
	}
	if !strings.Contains(html, "MyType") {
		t.Fatal("graph HTML did not include node labels")
	}
	if !strings.Contains(html, "calls") {
		t.Fatal("graph HTML did not include relationship metadata")
	}
}

func TestConvertStoreGraphData(t *testing.T) {
	nodesData := []map[string]interface{}{
		{"id": "n1", "label": "MyType", "group": "struct", "file": "main.go", "package": "example.com/demo"},
	}
	edgesData := []map[string]interface{}{
		{"from": "n1", "to": "n2", "type": "calls"},
	}

	nodes, edges := convertStoreGraphData(nodesData, edgesData)
	if len(nodes) != 1 || nodes[0].Label != "MyType" {
		t.Fatalf("expected 1 converted node with label MyType, got %#v", nodes)
	}
	if len(edges) != 1 || edges[0].Type != "calls" {
		t.Fatalf("expected 1 converted edge with type calls, got %#v", edges)
	}

	html := generateGraphHTML("demo", nodes, edges)
	if !strings.Contains(html, "MyType") || !strings.Contains(html, "calls") {
		t.Fatal("converted graph data did not render into HTML")
	}
}
