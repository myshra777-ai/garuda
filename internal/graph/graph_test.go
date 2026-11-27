// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package graph

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

func TestGenerate_EmbedsNodesAndEdges(t *testing.T) {
	nodes := []Node{
		{ID: "n1", Name: "Charge", Label: "Charge", Kind: "function", Package: "payments", File: "charge.go", Exported: true, Impact: 3},
		{ID: "n2", Name: "repo", Kind: "struct", Package: "payments"},
	}
	edges := []Edge{{From: "n1", To: "n2", Type: "CALLS", CrossRepo: true}}

	html, err := Generate("payments", nodes, edges)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.HasPrefix(html, "<!DOCTYPE html>") {
		t.Fatal("output should be a full HTML document")
	}

	gotNodes := extractJSONArray(t, html, "rawNodes")
	var decodedNodes []Node
	if err := json.Unmarshal([]byte(gotNodes), &decodedNodes); err != nil {
		t.Fatalf("embedded nodes are not valid JSON: %v\n%s", err, gotNodes)
	}
	if len(decodedNodes) != 2 || decodedNodes[0] != nodes[0] || decodedNodes[1] != nodes[1] {
		t.Fatalf("nodes did not round-trip: %+v", decodedNodes)
	}

	gotEdges := extractJSONArray(t, html, "rawEdges")
	var decodedEdges []Edge
	if err := json.Unmarshal([]byte(gotEdges), &decodedEdges); err != nil {
		t.Fatalf("embedded edges are not valid JSON: %v", err)
	}
	if len(decodedEdges) != 1 || decodedEdges[0] != edges[0] {
		t.Fatalf("edges did not round-trip: %+v", decodedEdges)
	}
}

func TestGenerate_EmptyGraph(t *testing.T) {
	html, err := Generate("empty", nil, nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(html, "var rawNodes = null;") && !strings.Contains(html, "var rawNodes = [];") {
		t.Fatalf("expected an empty node list to be embedded, got: %s", snippet(html, "rawNodes"))
	}
}

func TestGenerate_EscapesHostileNames(t *testing.T) {
	// Entity names come from analysed code, so they must never be able to
	// close the <script> block or inject markup.
	hostile := `</script><script>alert(1)</script>`
	nodes := []Node{{ID: "x", Name: hostile, Label: hostile}}
	html, err := Generate(`<img src=x onerror=alert(1)>`, nodes, nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if strings.Contains(html, "<script>alert(1)") {
		t.Fatal("a node name broke out of the script block")
	}
	if strings.Contains(html, "<img src=x onerror=alert(1)>") {
		t.Fatal("the workspace name was not escaped")
	}
	// The data must still be recoverable.
	var decoded []Node
	if err := json.Unmarshal([]byte(extractJSONArray(t, html, "rawNodes")), &decoded); err != nil {
		t.Fatalf("escaped nodes are not valid JSON: %v", err)
	}
	if decoded[0].Name != hostile {
		t.Fatalf("name changed after escaping: %q", decoded[0].Name)
	}
}

func extractJSONArray(t *testing.T, html, varName string) string {
	t.Helper()
	re := regexp.MustCompile(`var ` + varName + ` = (.*);`)
	m := re.FindStringSubmatch(html)
	if m == nil {
		t.Fatalf("variable %s not found in output", varName)
	}
	return m[1]
}

func snippet(html, near string) string {
	i := strings.Index(html, near)
	if i < 0 {
		return "(not found)"
	}
	end := i + 80
	if end > len(html) {
		end = len(html)
	}
	return html[i:end]
}
