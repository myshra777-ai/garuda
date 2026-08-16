package main

import (
	"encoding/json"
	"fmt"
)

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

// generateGraphHTML creates an HTML file with a lightweight D3.js visualization.
func generateGraphHTML(workspaceName string, nodes []GraphNode, edges []GraphEdge) string {
	dataNodes := "[]"
	dataEdges := "[]"
	if len(nodes) > 0 {
		dataNodes = marshalGraphData(nodes)
	}
	if len(edges) > 0 {
		dataEdges = marshalGraphData(edges)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8" />
  <title>Garuda Graph - %s</title>
  <style>
    body { margin: 0; font-family: Arial, sans-serif; background: #0f172a; color: #e2e8f0; }
    #graph { width: 100vw; height: 100vh; }
    .node { fill: #38bdf8; stroke: #dbeafe; stroke-width: 1.5px; }
    .node:hover { stroke: #facc15; }
    .link { stroke: rgba(148, 163, 184, 0.8); stroke-width: 1.5; }
    .label { fill: #e2e8f0; font-size: 12px; pointer-events: none; }
    .controls { position: absolute; top: 12px; left: 12px; background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(148, 163, 184, 0.7); border-radius: 8px; padding: 8px 10px; }
    .controls input { width: 220px; padding: 6px 8px; border-radius: 6px; border: 1px solid #475569; background: #0f172a; color: #e2e8f0; }
  </style>
</head>
<body>
  <div class="controls">
    <input id="filter" type="text" placeholder="Filter entities..." />
  </div>
  <svg id="graph"></svg>

  <script src="https://cdn.jsdelivr.net/npm/d3@7"></script>
  <script>
    const nodes = %s;
    const edges = %s;

    const svg = document.getElementById('graph');
    const width = window.innerWidth;
    const height = window.innerHeight;

    const simulation = d3.forceSimulation(nodes)
      .force('link', d3.forceLink(edges).id(function(d) { return d.id; }).distance(120))
      .force('charge', d3.forceManyBody().strength(-300))
      .force('center', d3.forceCenter(width / 2, height / 2));

    const link = d3.select(svg)
      .attr('width', width)
      .attr('height', height)
      .append('g')
      .selectAll('line')
      .data(edges)
      .enter().append('line')
      .attr('class', 'link');

    const node = d3.select(svg).selectAll('.node')
      .data(nodes)
      .enter().append('g')
      .attr('class', 'node');

    node.append('circle')
      .attr('r', 10)
      .attr('fill', '#38bdf8');

    node.append('text')
      .attr('class', 'label')
      .attr('dx', 12)
      .attr('dy', 4)
      .text(function(d) { return d.label || d.id; });

    simulation.on('tick', function() {
      link
        .attr('x1', function(d) { return d.source.x; })
        .attr('y1', function(d) { return d.source.y; })
        .attr('x2', function(d) { return d.target.x; })
        .attr('y2', function(d) { return d.target.y; });

      node.attr('transform', function(d) {
        return 'translate(' + d.x + ',' + d.y + ')';
      });
    });

    const filterInput = document.getElementById('filter');
    filterInput.addEventListener('input', function () {
      const term = this.value.trim().toLowerCase();
      if (!term) {
        node.style('opacity', 1);
        link.style('opacity', 1);
        return;
      }

      node.style('opacity', function(d) {
        return String(d.label || d.id).toLowerCase().includes(term) ? 1 : 0.15;
      });
      link.style('opacity', function(d) {
        const sourceMatch = String((d.source && (d.source.label || d.source.id)) || '').toLowerCase().includes(term);
        const targetMatch = String((d.target && (d.target.label || d.target.id)) || '').toLowerCase().includes(term);
        return sourceMatch || targetMatch ? 1 : 0.15;
      });
    });
  </script>
</body>
</html>`, workspaceName, dataNodes, dataEdges)
}

func marshalGraphData[T any](items []T) string {
	data, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(data)
}
