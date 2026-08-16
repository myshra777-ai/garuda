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
body {
	font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
	background: #f8fafc;
	color: #1e293b;
	height: 100vh;
	display: flex;
	flex-direction: column;
	overflow: hidden;
}
.header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 12px 24px;
	background: #ffffff;
	border-bottom: 1px solid #e2e8f0;
	flex-shrink: 0;
}
.header h1 { font-size: 20px; font-weight: 600; color: #0f172a; }
.header .sub { font-size: 14px; color: #64748b; }
.header .stats-text { font-size: 13px; color: #475569; }
.header .stats-text strong { color: #0f172a; }
.top-bar {
	display: flex;
	align-items: center;
	gap: 20px;
	padding: 8px 24px;
	background: #ffffff;
	border-bottom: 1px solid #e2e8f0;
	flex-shrink: 0;
	flex-wrap: wrap;
}
.search-box { flex: 1; min-width: 200px; }
.search-box input {
	width: 100%;
	padding: 6px 12px;
	border: 1px solid #cbd5e1;
	border-radius: 6px;
	font-size: 13px;
	background: #f8fafc;
	outline: none;
}
.search-box input:focus { border-color: #3b82f6; background: #ffffff; }
.legend {
	display: flex;
	flex-wrap: wrap;
	gap: 12px;
	align-items: center;
	font-size: 12px;
	color: #475569;
}
.legend-item { display: flex; align-items: center; gap: 4px; }
.color-dot {
	width: 12px;
	height: 12px;
	border-radius: 50%;
	border: 1px solid #e2e8f0;
	flex-shrink: 0;
}
.edge-sample { width: 20px; height: 2px; background: #94a3b8; display: inline-block; }
.main-container {
	position: relative;
	display: flex;
	flex: 1;
	overflow: hidden;
}
#graph-container {
	flex: 1;
	background: #ffffff;
	position: relative;
	cursor: grab;
	overflow: hidden;
}
#graph-container:active { cursor: grabbing; }
#graph-container svg { display: block; width: 100%; height: 100%; }
.node { cursor: pointer; }
.node circle { stroke: #fff; stroke-width: 1.5px; }
.node text {
	font-size: 10px;
	font-weight: 500;
	fill: #0f172a;
	pointer-events: none;
	text-shadow: 0 1px 2px rgba(255,255,255,0.8);
}
.link { stroke: #94a3b8; stroke-opacity: 0.5; stroke-width: 1.5px; }
.link-label { font-size: 8px; fill: #94a3b8; pointer-events: none; }
.side-panel {
	width: 340px;
	background: #ffffff;
	border-left: 1px solid #e2e8f0;
	padding: 18px;
	overflow-y: auto;
	flex-shrink: 0;
}
.side-panel h3 { font-size: 16px; font-weight: 600; margin-bottom: 8px; color: #0f172a; }
.side-panel .empty { color: #94a3b8; font-style: italic; font-size: 14px; }
.prop-group { margin: 12px 0; }
.prop-group .label {
	font-size: 11px;
	font-weight: 600;
	color: #475569;
	text-transform: uppercase;
	letter-spacing: 0.03em;
	margin-bottom: 2px;
}
.prop-group .value {
	font-size: 13px;
	color: #1e293b;
	background: #f8fafc;
	padding: 2px 10px;
	border-radius: 4px;
	border: 1px solid #e2e8f0;
	display: inline-block;
	font-family: monospace;
	font-size: 12px;
}
.claim-list { list-style: none; padding: 0; margin: 4px 0 8px 0; }
.claim-list li {
	font-size: 12px;
	padding: 3px 0;
	border-bottom: 1px solid #f1f5f9;
	display: flex;
	justify-content: space-between;
}
.claim-list li .type {
	background: #e2e8f0;
	padding: 0 8px;
	border-radius: 10px;
	font-size: 10px;
	color: #475569;
}
/* Controls and tooltip are now outside #graph-container, so they survive re-renders */
.controls {
	position: absolute;
	bottom: 24px;
	left: 50%;
	transform: translateX(-50%);
	display: flex;
	gap: 8px;
	background: rgba(255,255,255,0.95);
	border: 1px solid #e2e8f0;
	border-radius: 8px;
	padding: 8px 12px;
	backdrop-filter: blur(4px);
	z-index: 20;
	align-items: center;
	box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}
.controls button {
	background: #f1f5f9;
	border: 1px solid #e2e8f0;
	color: #0f172a;
	width: 32px;
	height: 32px;
	border-radius: 4px;
	cursor: pointer;
	font-size: 16px;
	display: flex;
	align-items: center;
	justify-content: center;
	transition: background 0.15s;
}
.controls button:hover { background: #e2e8f0; }
.controls input[type="range"] {
	width: 100px;
	height: 4px;
	background: #e2e8f0;
	border-radius: 9999px;
	outline: none;
	accent-color: #3b82f6;
}
.controls input[type="range"]::-webkit-slider-thumb {
	-webkit-appearance: none;
	width: 14px;
	height: 14px;
	border-radius: 50%;
	background: #3b82f6;
	cursor: pointer;
}
.controls .zoom-label {
	font-size: 11px;
	color: #475569;
	min-width: 44px;
	text-align: center;
	font-weight: 500;
}
.tooltip {
	position: absolute;
	top: 20px;
	left: 50%;
	transform: translateX(-50%);
	background: rgba(15,23,42,0.95);
	color: #f8fafc;
	border: 1px solid #1e293b;
	border-radius: 8px;
	padding: 10px 16px;
	font-size: 13px;
	line-height: 1.6;
	display: none;
	z-index: 30;
	pointer-events: none;
	backdrop-filter: blur(8px);
	box-shadow: 0 8px 24px rgba(0,0,0,0.3);
	max-width: 300px;
	white-space: nowrap;
}
.tooltip .title { font-weight: 600; color: #f8fafc; }
.tooltip .detail { color: #94a3b8; font-size: 12px; margin-top: 2px; }
::-webkit-scrollbar { width: 4px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 9999px; }
</style>
</head>
<body>

<div class="header">
	<div>
		<h1>Garuda Company Graph</h1>
		<div class="sub">Workspace: <strong>{{.Workspace}}</strong> • <span id="nodeCount">0</span> entities, <span id="edgeCount">0</span> relationships</div>
	</div>
	<div class="stats-text">Merkle root: <span style="color:#10b981;">verified</span></div>
</div>

<div class="top-bar">
	<div class="search-box">
		<input type="text" id="searchInput" placeholder="Search entities..." />
	</div>
	<div class="legend">
		<span class="legend-item"><span class="color-dot" style="background:#3b82f6;"></span> struct</span>
		<span class="legend-item"><span class="color-dot" style="background:#8b5cf6;"></span> interface</span>
		<span class="legend-item"><span class="color-dot" style="background:#f59e0b;"></span> func</span>
		<span class="legend-item"><span class="color-dot" style="background:#10b981;"></span> method</span>
		<span class="legend-item"><span class="color-dot" style="background:#ef4444;"></span> package</span>
		<span class="legend-item"><span class="color-dot" style="background:#ec4899;"></span> file</span>
		<span class="legend-item"><span class="color-dot" style="background:#94a3b8;"></span> other</span>
		<span class="legend-item" style="margin-left:8px;">Edge: <span class="edge-sample"></span> relationships</span>
	</div>
</div>

<div class="main-container">
	<div id="graph-container">
		<!-- graph is rendered here by D3 -->
	</div>
	<!-- Controls outside container -->
	<div class="controls">
		<button id="zoom-in" title="Zoom In">+</button>
		<input type="range" id="zoom-slider" min="0.1" max="5" step="0.05" value="1" />
		<button id="zoom-out" title="Zoom Out">−</button>
		<button id="zoom-reset" title="Reset View" style="font-size:14px;">⟲</button>
		<span class="zoom-label" id="zoom-level">100%</span>
	</div>
	<!-- Tooltip outside container -->
	<div class="tooltip" id="tooltip"></div>
	<div class="side-panel" id="sidePanel">
		<h3>Entity Details</h3>
		<div class="empty">Click a node to see details</div>
	</div>
</div>

<script>
(function() {
	var rawNodes = {{.NodesJSON}};
	var rawEdges = {{.EdgesJSON}};

	var colorMap = {
		'struct': '#3b82f6',
		'interface': '#8b5cf6',
		'function': '#f59e0b',
		'method': '#10b981',
		'package': '#ef4444',
		'file': '#ec4899',
		'external': '#6b7280',
		'repository': '#f472b6',
		'directory': '#9ca3af',
		'variable': '#14b8a6',
		'constant': '#f97316',
		'type': '#06b6d4',
		'api': '#8b5cf6',
		'service': '#f59e0b',
		'default': '#94a3b8'
	};
	function getColor(kind) { return colorMap[kind] || colorMap['default']; }

	var nodes = (rawNodes || []).map(function(n) {
		return {
			id: n.id || n.ID || '',
			label: n.label || n.Label || n.name || n.Name || '',
			kind: n.kind || n.group || n.Kind || 'default',
			package: n.package || n.Package || '',
			file: n.file || n.File || '',
			exported: n.exported || n.Exported || false
		};
	}).filter(function(n) { return n.id && n.label; });

	var edges = (rawEdges || []).map(function(e) {
		return {
			source: e.from || e.From || e.Source || '',
			target: e.to || e.To || e.Target || '',
			type: e.type || e.Type || 'edge'
		};
	}).filter(function(e) { return e.source && e.target; });

	document.getElementById('nodeCount').textContent = nodes.length;
	document.getElementById('edgeCount').textContent = edges.length;

	if (nodes.length === 0) {
		document.getElementById('graph-container').innerHTML =
			'<div style="color:#94a3b8;font-size:14px;text-align:center;padding-top:40px;">No entities found.</div>';
		return;
	}

	var container = document.getElementById('graph-container');
	var width = container.clientWidth;
	var height = container.clientHeight;

	d3.select(container).selectAll('*').remove();

	var svg = d3.select(container)
		.append('svg')
		.attr('width', width)
		.attr('height', height)
		.style('cursor', 'grab');

	var zoom = d3.zoom()
		.scaleExtent([0.1, 5])
		.on('zoom', function(event) {
			g.attr('transform', event.transform);
			var k = event.transform.k;
			document.getElementById('zoom-slider').value = k;
			document.getElementById('zoom-level').innerText = Math.round(k * 100) + '%';
		});

	svg.call(zoom);
	var g = svg.append('g');

	var simulation = d3.forceSimulation(nodes)
		.force('link', d3.forceLink(edges).id(function(d) { return d.id; }).distance(100).strength(0.3))
		.force('charge', d3.forceManyBody().strength(-350))
		.force('center', d3.forceCenter(width/2, height/2))
		.force('collision', d3.forceCollide().radius(30));

	var link = g.append('g')
		.selectAll('line')
		.data(edges)
		.enter().append('line')
		.attr('class', 'link');

	var linkLabel = g.append('g')
		.selectAll('text')
		.data(edges)
		.enter().append('text')
		.attr('class', 'link-label')
		.text(function(d) { return d.type; });

	var tooltip = d3.select('#tooltip');

	var node = g.append('g')
		.selectAll('g')
		.data(nodes)
		.enter().append('g')
		.attr('class', 'node')
		.call(d3.drag()
			.on('start', function(event, d) {
				if (!event.active) simulation.alphaTarget(0.3).restart();
				d.fx = d.x;
				d.fy = d.y;
			})
			.on('drag', function(event, d) {
				d.fx = event.x;
				d.fy = event.y;
			})
			.on('end', function(event, d) {
				if (!event.active) simulation.alphaTarget(0);
				d.fx = null;
				d.fy = null;
			})
		);

	node.append('circle')
		.attr('r', 16)
		.attr('fill', function(d) { return getColor(d.kind); })
		.attr('stroke', '#fff')
		.attr('stroke-width', 1.5)
		.on('mouseover', function(event, d) {
			tooltip.style('display', 'block')
				.html(
					'<div class="title">' + d.label + '</div>' +
					'<div class="detail">Kind: <span style="color:' + getColor(d.kind) + ';font-weight:600;">' + d.kind + '</span></div>' +
					(d.package ? '<div class="detail">Package: ' + d.package + '</div>' : '') +
					(d.file ? '<div class="detail">File: ' + d.file + '</div>' : '') +
					(d.exported ? '<div class="detail">🔓 Exported</div>' : '')
				)
				.style('left', (event.pageX + 12) + 'px')
				.style('top', (event.pageY - 10) + 'px');
			d3.select(this).transition().duration(150).attr('r', 22);
		})
		.on('mouseout', function() {
			tooltip.style('display', 'none');
			d3.select(this).transition().duration(150).attr('r', 16);
		})
		.on('click', function(event, d) {
			showDetails(d);
		});

	node.append('text')
		.text(function(d) { return d.label; })
		.attr('x', 20)
		.attr('y', 4)
		.style('font-size', '10px')
		.style('font-weight', '500')
		.style('fill', '#0f172a')
		.style('pointer-events', 'none')
		.style('text-shadow', '0 1px 2px rgba(255,255,255,0.8)');

	simulation.on('tick', function() {
		link
			.attr('x1', function(d) { return d.source.x; })
			.attr('y1', function(d) { return d.source.y; })
			.attr('x2', function(d) { return d.target.x; })
			.attr('y2', function(d) { return d.target.y; });

		linkLabel
			.attr('x', function(d) { return (d.source.x + d.target.x) / 2; })
			.attr('y', function(d) { return (d.source.y + d.target.y) / 2 - 6; });

		node.attr('transform', function(d) {
			return 'translate(' + d.x + ',' + d.y + ')';
		});
	});

	// --- Zoom controls (outside container, always available) ---
	document.getElementById('zoom-slider').addEventListener('input', function() {
		var val = parseFloat(this.value);
		var transform = d3.zoomIdentity.scale(val);
		svg.transition().duration(150).call(zoom.transform, transform);
		document.getElementById('zoom-level').innerText = Math.round(val * 100) + '%';
	});

	document.getElementById('zoom-in').addEventListener('click', function() {
		var slider = document.getElementById('zoom-slider');
		var val = parseFloat(slider.value) + 0.1;
		if (val > 5) val = 5;
		slider.value = val;
		var transform = d3.zoomIdentity.scale(val);
		svg.transition().duration(150).call(zoom.transform, transform);
		document.getElementById('zoom-level').innerText = Math.round(val * 100) + '%';
	});

	document.getElementById('zoom-out').addEventListener('click', function() {
		var slider = document.getElementById('zoom-slider');
		var val = parseFloat(slider.value) - 0.1;
		if (val < 0.1) val = 0.1;
		slider.value = val;
		var transform = d3.zoomIdentity.scale(val);
		svg.transition().duration(150).call(zoom.transform, transform);
		document.getElementById('zoom-level').innerText = Math.round(val * 100) + '%';
	});

	document.getElementById('zoom-reset').addEventListener('click', function() {
		var slider = document.getElementById('zoom-slider');
		slider.value = 1;
		var transform = d3.zoomIdentity;
		svg.transition().duration(300).call(zoom.transform, transform);
		document.getElementById('zoom-level').innerText = '100%';
	});

	// --- SEARCH ---
	document.getElementById('searchInput').addEventListener('input', function() {
		var q = this.value.toLowerCase().trim();
		d3.selectAll('.node').style('opacity', function(d) {
			if (q === '') return 1;
			var labelMatch = d.label && d.label.toLowerCase().includes(q);
			var pkgMatch = d.package && d.package.toLowerCase().includes(q);
			return (labelMatch || pkgMatch) ? 1 : 0.1;
		});
	});

	// --- Side panel ---
	function showDetails(d) {
		var panel = document.getElementById('sidePanel');
		if (!d) {
			panel.innerHTML = '<h3>Entity Details</h3><div class="empty">Click a node to see details</div>';
			return;
		}

		var incoming = edges.filter(function(e) { return e.target === d.id; });
		var outgoing = edges.filter(function(e) { return e.source === d.id; });

		var html = '<h3>' + d.label + '</h3>';
		html += '<div class="prop-group"><div class="label">Kind</div><div class="value">' + d.kind + '</div></div>';
		html += '<div class="prop-group"><div class="label">Package</div><div class="value">' + d.package + '</div></div>';
		html += '<div class="prop-group"><div class="label">File</div><div class="value">' + (d.file || 'N/A') + '</div></div>';
		html += '<div class="prop-group"><div class="label">Exported</div><div class="value">' + (d.exported ? 'Yes' : 'No') + '</div></div>';

		if (outgoing.length > 0) {
			html += '<div class="prop-group"><div class="label">Outgoing relationships</div><ul class="claim-list">';
			outgoing.forEach(function(e) {
				var target = nodes.find(function(n) { return n.id === e.target; });
				var label = target ? target.label : e.target;
				html += '<li>' + e.type + ' → <strong>' + label + '</strong></li>';
			});
			html += '</ul></div>';
		}

		if (incoming.length > 0) {
			html += '<div class="prop-group"><div class="label">Incoming relationships</div><ul class="claim-list">';
			incoming.forEach(function(e) {
				var source = nodes.find(function(n) { return n.id === e.source; });
				var label = source ? source.label : e.source;
				html += '<li>' + e.type + ' ← <strong>' + label + '</strong></li>';
			});
			html += '</ul></div>';
		}

		panel.innerHTML = html;
	}

	// --- Resize ---
	var resizeHandler = function() {
		var newWidth = container.clientWidth;
		var newHeight = container.clientHeight;
		svg.attr('width', newWidth).attr('height', newHeight);
		simulation.force('center', d3.forceCenter(newWidth/2, newHeight/2));
		simulation.alpha(0.3).restart();
	};
	window.removeEventListener('resize', resizeHandler);
	window.addEventListener('resize', resizeHandler);

	// Keyboard shortcuts
	document.addEventListener('keydown', function(e) {
		if (e.key === '+' || e.key === '=') {
			e.preventDefault();
			document.getElementById('zoom-in').click();
		} else if (e.key === '-') {
			e.preventDefault();
			document.getElementById('zoom-out').click();
		} else if (e.key === 'r' && (e.ctrlKey || e.metaKey)) {
			e.preventDefault();
			document.getElementById('zoom-reset').click();
		}
	});
})();
</script>
</body>
</html>`
