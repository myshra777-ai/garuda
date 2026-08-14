package graph

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Garuda Graph - {{.Workspace}}</title>
<script src="https://d3js.org/d3.v7.min.js"></script>
<style>
* { margin:0; padding:0; box-sizing:border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; background: #ffffff; color: #1e293b; padding: 20px; display: flex; flex-direction: column; height: 100vh; }
.header { display: flex; justify-content: space-between; align-items: center; padding-bottom: 12px; border-bottom: 1px solid #e2e8f0; }
.header h1 { font-weight: 600; font-size: 20px; color: #0f172a; }
.header .sub { font-size: 14px; color: #64748b; }
.controls { display: flex; gap: 20px; align-items: center; margin: 12px 0; flex-wrap: wrap; }
.controls input { padding: 8px 12px; border: 1px solid #cbd5e1; border-radius: 6px; font-size: 14px; width: 260px; background: #f8fafc; }
.controls input:focus { outline: none; border-color: #3b82f6; background: #ffffff; }
.legend { display: flex; gap: 20px; align-items: center; font-size: 13px; flex-wrap: wrap; }
.legend-item { display: flex; align-items: center; gap: 6px; }
.color-dot { width: 12px; height: 12px; border-radius: 50%; border: 1px solid #e2e8f0; }
.main-container { display: flex; flex: 1; overflow: hidden; gap: 20px; margin-top: 8px; }
#graph-container { flex: 1; background: #ffffff; border-radius: 12px; border: 1px solid #e2e8f0; position: relative; overflow: hidden; }
#graph-container svg { width: 100%; height: 100%; }
.node circle { stroke: #fff; stroke-width: 1.5px; }
.node text { font-size: 10px; font-weight: 500; fill: #0f172a; pointer-events: none; text-shadow: 0 1px 2px rgba(255,255,255,0.8); }
.edge line { stroke: #94a3b8; stroke-opacity: 0.5; stroke-width: 1.5px; }
.side-panel { width: 340px; background: #ffffff; border-radius: 12px; border: 1px solid #e2e8f0; padding: 18px; overflow-y: auto; flex-shrink: 0; }
.side-panel h3 { font-size: 16px; font-weight: 600; margin-bottom: 8px; color: #0f172a; }
.side-panel .empty { color: #94a3b8; font-style: italic; font-size: 14px; }
.prop-group { margin: 12px 0; }
.prop-group .label { font-size: 12px; font-weight: 600; color: #475569; text-transform: uppercase; letter-spacing: 0.03em; margin-bottom: 4px; }
.prop-group .value { font-size: 13px; color: #1e293b; background: #f8fafc; padding: 4px 10px; border-radius: 4px; border: 1px solid #e2e8f0; display: inline-block; font-family: monospace; font-size: 12px; }
.prop-group .value-block { font-size: 13px; color: #1e293b; background: #f8fafc; padding: 6px 10px; border-radius: 4px; border: 1px solid #e2e8f0; font-family: monospace; font-size: 12px; white-space: pre-wrap; word-break: break-all; margin-top: 2px; }
.claim-list { list-style: none; padding: 0; margin: 4px 0 8px 0; }
.claim-list li { font-size: 12px; padding: 3px 0; border-bottom: 1px solid #f1f5f9; display: flex; justify-content: space-between; }
.claim-list li .type { background: #e2e8f0; padding: 0 8px; border-radius: 10px; font-size: 10px; color: #475569; }
.tooltip { position: absolute; background: white; border: 1px solid #cbd5e1; border-radius: 6px; padding: 8px 12px; font-size: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); pointer-events: none; opacity: 0; transition: opacity 0.15s; }
.tooltip.visible { opacity: 1; }
.stats-bar { display: flex; gap: 24px; font-size: 13px; color: #475569; padding: 6px 0; }
.stats-bar strong { color: #0f172a; }
</style>
</head>
<body>
<div class="header">
<div><h1>Garuda Company Graph</h1><div class="sub">Workspace: {{.Workspace}} • <span id="nodeCount">0</span> entities, <span id="edgeCount">0</span> relationships</div></div>
<div class="stats-bar"><span>Merkle root: <span id="merkleStatus">verified</span></span></div>
</div>
<div class="controls">
<input type="text" id="searchInput" placeholder="Search entities..." />
<div class="legend">
<span class="legend-item"><span class="color-dot" style="background:#3b82f6;"></span> struct</span>
<span class="legend-item"><span class="color-dot" style="background:#8b5cf6;"></span> interface</span>
<span class="legend-item"><span class="color-dot" style="background:#f59e0b;"></span> func</span>
<span class="legend-item"><span class="color-dot" style="background:#10b981;"></span> type</span>
<span class="legend-item"><span class="color-dot" style="background:#ef4444;"></span> other</span>
<span class="legend-item" style="margin-left:8px;">Edge: <span style="background:#94a3b8; width:20px; height:2px; display:inline-block;"></span> relationships</span>
</div>
</div>
<div class="main-container">
<div id="graph-container"></div>
<div class="side-panel" id="sidePanel">
<h3>Entity Details</h3>
<div class="empty">Click a node to see details</div>
</div>
</div>
<script>
var data = {{.DataJSON}};
var colorMap = { 'struct':'#3b82f6', 'interface':'#8b5cf6', 'func':'#f59e0b', 'type':'#10b981' };
function getColor(k) { return colorMap[k] || '#ef4444'; }
var nodes = data.Nodes.map(function(n) { return { id: n.ID, label: n.Label, kind: n.Kind, package: n.Package, file: n.File, exported: n.Exported, group: n.Kind }; });
var edges = data.Edges.map(function(e) { return { source: e.From, target: e.To, type: e.Type }; });

var container = document.getElementById('graph-container');
var width = container.clientWidth, height = container.clientHeight;
var svg = d3.select('#graph-container').append('svg').attr('width', width).attr('height', height).append('g');

var simulation = d3.forceSimulation(nodes)
    .force('link', d3.forceLink(edges).id(function(d) { return d.id; }).distance(80).strength(0.3))
    .force('charge', d3.forceManyBody().strength(-200))
    .force('center', d3.forceCenter(width/2, height/2))
    .force('x', d3.forceX(width/2).strength(0.02))
    .force('y', d3.forceY(height/2).strength(0.02));

var link = svg.append('g').selectAll('line').data(edges).enter().append('line')
    .attr('stroke', '#94a3b8').attr('stroke-opacity', 0.5).attr('stroke-width', 1.5);

var node = svg.append('g').selectAll('g').data(nodes).enter().append('g').attr('class', 'node')
    .call(d3.drag().on('start', function(event,d){ if(!event.active) simulation.alphaTarget(0.3).restart(); d.fx=d.x; d.fy=d.y; })
    .on('drag', function(event,d){ d.fx=event.x; d.fy=event.y; })
    .on('end', function(event,d){ if(!event.active) simulation.alphaTarget(0); d.fx=null; d.fy=null; }));

node.append('circle').attr('r', 12).attr('fill', function(d) { return getColor(d.kind); }).attr('stroke', '#ffffff').attr('stroke-width', 1.5);
node.append('text').text(function(d) { return d.label; }).attr('x', 16).attr('y', 4).attr('font-size', 10).attr('fill', '#0f172a').attr('font-weight', 500).attr('pointer-events', 'none');

node.on('click', function(event, d) { showDetails(d); });

var tooltip = d3.select('body').append('div').attr('class', 'tooltip');
node.on('mouseover', function(event, d) {
    tooltip.html('<strong>' + d.label + '</strong><br/>' + d.kind + ' • ' + d.package).classed('visible', true)
        .style('left', (event.pageX + 12) + 'px').style('top', (event.pageY - 10) + 'px');
}).on('mousemove', function(event, d) {
    tooltip.style('left', (event.pageX + 12) + 'px').style('top', (event.pageY - 10) + 'px');
}).on('mouseout', function() { tooltip.classed('visible', false); });

simulation.on('tick', function() {
    link.attr('x1', function(d) { return d.source.x; }).attr('y1', function(d) { return d.source.y; })
        .attr('x2', function(d) { return d.target.x; }).attr('y2', function(d) { return d.target.y; });
    node.attr('transform', function(d) { return 'translate(' + d.x + ',' + d.y + ')'; });
});

window.addEventListener('resize', function() {
    var w = container.clientWidth, h = container.clientHeight;
    svg.attr('width', w).attr('height', h);
    simulation.force('center', d3.forceCenter(w/2, h/2));
    simulation.force('x', d3.forceX(w/2).strength(0.02));
    simulation.force('y', d3.forceY(h/2).strength(0.02));
    simulation.alpha(0.3).restart();
});

function showDetails(d) {
    var panel = document.getElementById('sidePanel');
    var claimsIn = edges.filter(function(e) { return e.target === d.id; });
    var claimsOut = edges.filter(function(e) { return e.source === d.id; });
    var html = '<h3>' + d.label + '</h3>';
    html += '<div class="prop-group"><div class="label">Kind</div><div class="value">' + d.kind + '</div></div>';
    html += '<div class="prop-group"><div class="label">Package</div><div class="value">' + d.package + '</div></div>';
    html += '<div class="prop-group"><div class="label">File</div><div class="value">' + (d.file || 'N/A') + '</div></div>';
    html += '<div class="prop-group"><div class="label">Exported</div><div class="value">' + (d.exported ? 'Yes' : 'No') + '</div></div>';
    if (claimsOut.length) {
        html += '<div class="prop-group"><div class="label">Outgoing claims</div><ul class="claim-list">';
        claimsOut.forEach(function(e) {
            var target = nodes.find(function(n) { return n.id === e.target; });
            html += '<li>' + e.type + ' → <strong>' + (target ? target.label : e.target) + '</strong></li>';
        });
        html += '</ul></div>';
    }
    if (claimsIn.length) {
        html += '<div class="prop-group"><div class="label">Incoming claims</div><ul class="claim-list">';
        claimsIn.forEach(function(e) {
            var source = nodes.find(function(n) { return n.id === e.source; });
            html += '<li>' + e.type + ' ← <strong>' + (source ? source.label : e.source) + '</strong></li>';
        });
        html += '</ul></div>';
    }
    panel.innerHTML = html;
}

document.getElementById('searchInput').addEventListener('input', function(e) {
    var q = e.target.value.toLowerCase();
    node.style('opacity', function(d) { return q === '' || d.label.toLowerCase().includes(q) ? 1 : 0.1; });
});

document.getElementById('nodeCount').textContent = nodes.length;
document.getElementById('edgeCount').textContent = edges.length;
</script>
</body>
</html>`
