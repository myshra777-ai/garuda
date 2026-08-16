package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/types"
)

type DashboardData struct {
	TenantID string
}

type AgentFleetItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
	System string `json:"system"`
	Status string `json:"status"`
}

type RealStatsResponse struct {
	TotalDecisions       int               `json:"total_decisions"`
	QuarantinedCount     int               `json:"quarantined_count"`
	LatestBlockHeight    int64             `json:"latest_block_height"`
	LatestMerkleHash     string            `json:"latest_merkle_hash"`
	ParentMerkleHash     string            `json:"parent_merkle_hash"`
	EstimatedSavings     float64           `json:"estimated_savings"`
	DomainBreakdown      map[string]int    `json:"domain_breakdown"`
	QuarantinedDecisions []*types.Decision `json:"quarantined_decisions"`
	AgentList            []AgentFleetItem  `json:"agent_list"`
}

// GraphResponse for D3 visualization
type GraphResponse struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type GraphNode struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Kind     string `json:"kind"`
	Package  string `json:"package"`
	File     string `json:"file"`
	Exported bool   `json:"exported"`
}

type GraphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

const prodDashboardHTML = `<!DOCTYPE html>
<html lang="en" class="dark">
<head>
    <script src="https://d3js.org/d3.v7.min.js"></script>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Garuda — AI Governance Platform</title>

    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/lucide@latest"></script>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">

    <script>
        tailwind.config = {
            darkMode: 'class',
            theme: {
                extend: {
                    fontFamily: {
                        sans: ['Inter', 'sans-serif'],
                        mono: ['JetBrains Mono', 'monospace'],
                    },
                    colors: {
                        brand: {
                            50: '#f4f0ff',
                            500: '#7c3aed',
                            600: '#6d28d9',
                            700: '#5b21b6',
                            900: '#2e1065',
                        }
                    }
                }
            }
        }
    </script>

    <style>
        body { background-color: #080c14; color: #f8fafc; font-family: 'Inter', sans-serif; }
        .summary-card { background: #0f172a; border: 1px solid #1e293b; box-shadow: 0 4px 20px -2px rgba(0, 0, 0, 0.3); }
        ::-webkit-scrollbar { width: 6px; height: 6px; }
        ::-webkit-scrollbar-track { background: #080c14; }
        ::-webkit-scrollbar-thumb { background: #334155; border-radius: 9999px; }
        #graph-container svg { display: block; width: 100%; height: 100%; }
        .node-label { font-size: 10px; font-weight: 500; fill: #f8fafc; }
        .link { stroke: #94a3b8; stroke-opacity: 0.5; stroke-width: 1.5; }
    </style>
</head>
<body class="min-h-screen flex flex-col selection:bg-brand-500 selection:text-white">

    <!-- TOP NAVIGATION BAR -->
    <header class="h-14 bg-slate-900 border-b border-slate-800 px-6 flex items-center justify-between sticky top-0 z-30">
        <div class="flex items-center space-x-4">
            <div class="flex items-center space-x-2.5 cursor-pointer" onclick="switchTab('summary')">
                <div class="w-8 h-8 rounded-lg bg-gradient-to-tr from-purple-600 via-indigo-600 to-brand-500 flex items-center justify-center text-white font-bold text-lg shadow-md">
                    🛡️
                </div>
                <span class="font-extrabold text-white tracking-tight text-base">GARUDA</span>
                <span class="text-[10px] font-bold px-2 py-0.5 rounded bg-purple-950/80 text-purple-300 border border-purple-800 font-mono">PROD v1.0</span>
            </div>
            <div class="h-4 w-px bg-slate-800 hidden sm:block"></div>
            <div class="hidden sm:flex items-center space-x-2 text-xs text-slate-400 font-mono">
                <span>Tenant: {{.TenantID}}</span>
            </div>
        </div>

        <div class="flex items-center space-x-3 text-xs">
            <div class="flex items-center space-x-1.5 px-2.5 py-1 bg-emerald-950/40 text-emerald-400 rounded-md border border-emerald-800/50">
                <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
                <span class="font-medium text-[11px]">Live PostgreSQL Sync</span>
            </div>
        </div>
    </header>

    <!-- TITLE BAR & SUB-NAVIGATION -->
    <div class="bg-slate-900 border-b border-slate-800 px-6 pt-5 pb-0">
        <div class="flex flex-col md:flex-row md:items-center justify-between mb-4 gap-3">
            <div>
                <h1 class="text-xl font-bold text-white tracking-tight">Real-Time Telemetry Command Center</h1>
                <p class="text-xs text-slate-400 mt-1">Live database execution metrics, Merkle proof tree heights, and domain isolation queues.</p>
            </div>

            <div class="flex items-center space-x-2 text-xs">
                <button onclick="openProvisionModal()" class="px-3 py-1.5 bg-brand-600 hover:bg-brand-700 text-white font-medium rounded-lg shadow-sm transition flex items-center space-x-1.5">
                    <i data-lucide="plus-circle" class="w-3.5 h-3.5"></i>
                    <span>Propose Decision</span>
                </button>
                <button onclick="fetchRealData()" class="px-3 py-1.5 bg-slate-800 border border-slate-700 text-slate-200 rounded-lg hover:bg-slate-700 transition flex items-center space-x-1.5">
                    <i data-lucide="refresh-cw" class="w-3.5 h-3.5"></i>
                    <span>Refresh DB Telemetry</span>
                </button>
            </div>
        </div>

        <!-- Sub-Navigation Tabs -->
        <div class="flex space-x-6 text-sm font-medium border-b border-transparent -mb-px overflow-x-auto">
            <button id="tab-summary" onclick="switchTab('summary')" class="tab-btn pb-3 border-b-2 border-brand-500 text-brand-400 font-semibold flex items-center space-x-2 whitespace-nowrap">
                <i data-lucide="layout-dashboard" class="w-4 h-4"></i>
                <span>Executive Summary</span>
            </button>
            <button id="tab-workforce" onclick="switchTab('workforce')" class="tab-btn pb-3 border-b-2 border-transparent text-slate-400 hover:text-slate-200 flex items-center space-x-2 whitespace-nowrap transition">
                <i data-lucide="network" class="w-4 h-4"></i>
                <span>AI Workforce Topology</span>
            </button>
            <button id="tab-governance" onclick="switchTab('governance')" class="tab-btn pb-3 border-b-2 border-transparent text-slate-400 hover:text-slate-200 flex items-center space-x-2 whitespace-nowrap transition">
                <i data-lucide="shield-alert" class="w-4 h-4"></i>
                <span>Quarantine Queue</span>
                <span id="quarantineBadge" class="bg-amber-500 text-slate-950 font-bold text-[10px] px-1.5 py-0.2 rounded-full ml-1">0</span>
            </button>
        </div>
    </div>

    <!-- MAIN CONTENT CONTAINER -->
    <main class="flex-1 p-6 space-y-6 max-w-[1600px] w-full mx-auto">

        <!-- TAB 1: EXECUTIVE SUMMARY -->
        <div id="view-summary" class="tab-content space-y-6">
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <div class="summary-card rounded-xl p-4 flex flex-col justify-between">
                    <span class="text-xs font-semibold text-slate-400 uppercase">Canonical Decisions</span>
                    <span id="real-total-decisions" class="text-3xl font-bold text-white mt-2 font-mono">0</span>
                    <span class="text-xs text-emerald-400 mt-1">✓ PostgreSQL Verified</span>
                </div>
                <div class="summary-card rounded-xl p-4 flex flex-col justify-between border-l-4 border-l-amber-500">
                    <span class="text-xs font-semibold text-slate-400 uppercase">Quarantined Conflicts</span>
                    <span id="real-quarantined-count" class="text-3xl font-bold text-amber-400 mt-2 font-mono">0</span>
                    <span class="text-xs text-amber-400 mt-1">Requires Operator Action</span>
                </div>
                <div class="summary-card rounded-xl p-4 flex flex-col justify-between">
                    <span class="text-xs font-semibold text-slate-400 uppercase">Merkle Block Height</span>
                    <span id="real-block-height" class="text-3xl font-bold text-purple-400 mt-2 font-mono">#0</span>
                    <span class="text-xs text-purple-300 mt-1">Worker Epoch Active</span>
                </div>
                <div class="summary-card rounded-xl p-4 flex flex-col justify-between">
                    <span class="text-xs font-semibold text-slate-400 uppercase">Token Cost Savings</span>
                    <span id="real-savings" class="text-3xl font-bold text-emerald-400 mt-2 font-mono">$0.00</span>
                    <span class="text-xs text-emerald-400/80 mt-1">Calculated via Deduplication</span>
                </div>
            </div>

            <!-- Merkle Root Chain Banner -->
            <div class="summary-card rounded-xl p-5 space-y-3">
                <div class="flex items-center justify-between">
                    <h2 class="text-sm font-bold text-white flex items-center">
                        <i data-lucide="binary" class="w-4 h-4 mr-2 text-brand-500"></i> Latest Cryptographic Merkle Root Snapshot
                    </h2>
                    <span id="real-parent-hash" class="text-xs font-mono text-purple-400 bg-purple-950/50 px-2 py-0.5 rounded border border-purple-800">
                        Parent: Genesis
                    </span>
                </div>
                <div id="real-merkle-hash" class="p-3 bg-slate-950 border border-purple-500/30 rounded-lg font-mono text-xs text-slate-200 break-all">
                    No Merkle root recorded yet
                </div>
            </div>

            <div class="summary-card rounded-xl p-5">
                <h2 class="text-sm font-bold text-white mb-4">Domain Operations Breakdown</h2>
                <div class="h-64 w-full">
                    <canvas id="domainOpsChart"></canvas>
                </div>
            </div>
        </div>

        <!-- TAB 2: WORKFORCE TOPOLOGY -->
        <div id="view-workforce" class="tab-content hidden space-y-6">
            <div class="summary-card rounded-xl p-5">
                <h2 class="text-sm font-bold text-white mb-2 flex items-center">
                    <i data-lucide="share-2" class="w-4 h-4 mr-2 text-brand-500"></i> Active Agent Mesh Network
                </h2>
                <p class="text-xs text-slate-400 mb-4">Live agent instances registered from PostgreSQL decision origins.</p>
                
                <div id="agentFleetContainer" class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6 font-mono text-xs">
                    <!-- Real Agents Rendered Dynamically -->
                </div>

                <!-- D3.js Graph Container -->
                <div id="graph-container" style="width:100%; height:400px; background:#0f172a; border-radius:12px; border:1px solid #1e293b; overflow:hidden;"></div>
            </div>
        </div>

        <!-- TAB 3: QUARANTINE LEDGER -->
        <div id="view-governance" class="tab-content hidden space-y-6">
            <div class="summary-card rounded-xl p-5 border-l-4 border-l-amber-500">
                <h2 class="text-base font-bold text-white mb-2">Quarantined Contradiction Queue</h2>
                <p class="text-xs text-slate-400 mb-4">Real isolated proposals fetched directly from PostgreSQL storage.</p>
                <div id="quarantineListContainer" class="space-y-3 font-mono text-xs">
                    <!-- Real items rendered dynamically -->
                </div>
            </div>
        </div>

    </main>

    <!-- PROPOSE DECISION MODAL -->
    <div id="provisionModal" class="fixed inset-0 bg-slate-950/80 backdrop-blur-sm z-50 hidden flex items-center justify-center p-4">
        <div class="bg-slate-900 border border-slate-800 rounded-xl w-full max-w-md p-5 space-y-4 shadow-2xl">
            <h3 class="text-sm font-bold text-white">Propose Governance Decision</h3>
            <div class="space-y-3 text-xs">
                <div>
                    <label class="block text-slate-400 mb-1">Title / Policy Statement</label>
                    <input type="text" id="newDecisionTitle" placeholder="e.g. Enforce OAuth2 for external endpoints" class="w-full bg-slate-950 border border-slate-800 rounded px-3 py-2 text-slate-100 focus:outline-none focus:border-brand-500">
                </div>
                <div>
                    <label class="block text-slate-400 mb-1">Scope Domain</label>
                    <input type="text" id="newDecisionDomain" value="security" class="w-full bg-slate-950 border border-slate-800 rounded px-3 py-2 text-slate-100 focus:outline-none focus:border-brand-500">
                </div>
                <div>
                    <label class="block text-slate-400 mb-1">Scope System</label>
                    <input type="text" id="newDecisionSystem" value="network" class="w-full bg-slate-950 border border-slate-800 rounded px-3 py-2 text-slate-100 focus:outline-none focus:border-brand-500">
                </div>
            </div>
            <div class="flex justify-end space-x-2 pt-2">
                <button onclick="closeProvisionModal()" class="px-3 py-1.5 bg-slate-800 text-slate-300 text-xs rounded">Cancel</button>
                <button onclick="submitRealDecision()" class="px-3 py-1.5 bg-brand-600 hover:bg-brand-700 text-white text-xs rounded font-medium">Submit to Engine</button>
            </div>
        </div>
    </div>

    <script>
        lucide.createIcons();

        let domainChart = null;

        function switchTab(tabId) {
            document.querySelectorAll('.tab-content').forEach(el => el.classList.add('hidden'));
            document.querySelectorAll('.tab-btn').forEach(el => {
                el.classList.remove('border-brand-500', 'text-brand-400', 'font-semibold');
                el.classList.add('border-transparent', 'text-slate-400');
            });

            document.getElementById('view-' + tabId).classList.remove('hidden');
            const activeBtn = document.getElementById('tab-' + tabId);
            if (activeBtn) {
                activeBtn.classList.remove('border-transparent', 'text-slate-400');
                activeBtn.classList.add('border-brand-500', 'text-brand-400', 'font-semibold');
            }

            if (tabId === 'workforce') {
                renderGraph();
            }
        }

        function openProvisionModal() { document.getElementById('provisionModal').classList.remove('hidden'); }
        function closeProvisionModal() { document.getElementById('provisionModal').classList.add('hidden'); }

        async function fetchRealData() {
            try {
                const res = await fetch('/api/v1/dashboard/stats');
                if (!res.ok) return;
                const data = await res.json();

                document.getElementById('real-total-decisions').innerText = data.total_decisions || 0;
                document.getElementById('real-quarantined-count').innerText = data.quarantined_count || 0;
                document.getElementById('quarantineBadge').innerText = data.quarantined_count || 0;
                document.getElementById('real-block-height').innerText = '#' + (data.latest_block_height || 0);
                document.getElementById('real-savings').innerText = '$' + (data.estimated_savings || 0).toFixed(2);

                if (data.latest_merkle_hash) {
                    document.getElementById('real-merkle-hash').innerText = data.latest_merkle_hash;
                }
                if (data.parent_merkle_hash) {
                    document.getElementById('real-parent-hash').innerText = 'Parent: ' + data.parent_merkle_hash;
                }

                renderQuarantineQueue(data.quarantined_decisions || []);
                renderDomainChart(data.domain_breakdown || {});
                renderAgentFleet(data.agent_list || []);
            } catch (err) {
                console.error('Error fetching dashboard stats:', err);
            }
        }

        function renderAgentFleet(agents) {
            const container = document.getElementById('agentFleetContainer');
            if (!container) return;
            container.innerHTML = '';

            if (agents.length === 0) {
                container.innerHTML = '<div class="col-span-3 text-slate-500 text-center py-4 font-sans">No active agents registered in database. Propose a decision to instantiate agent contexts.</div>';
                return;
            }

            agents.forEach(agent => {
                const card = document.createElement('div');
                card.className = 'p-3.5 bg-slate-950 border border-slate-800 rounded-lg space-y-1';
                card.innerHTML =
                    '<div class="flex items-center justify-between">' +
                        '<span class="font-bold text-purple-400">' + agent.name + '</span>' +
                        '<span class="px-1.5 py-0.5 bg-emerald-950/60 text-emerald-400 border border-emerald-800/50 rounded text-[9px]">ACTIVE</span>' +
                    '</div>' +
                    '<div class="text-[11px] text-slate-400">Domain: <span class="text-slate-200">' + agent.domain + '</span></div>' +
                    '<div class="text-[10px] text-slate-500">System: ' + agent.system + '</div>';
                container.appendChild(card);
            });
        }

        function renderQuarantineQueue(list) {
            const container = document.getElementById('quarantineListContainer');
            container.innerHTML = '';

            if (list.length === 0) {
                container.innerHTML = '<div class="p-4 text-center text-slate-500 font-sans">No quarantined conflicts in PostgreSQL ledger. All decisions are canonical.</div>';
                return;
            }

            list.forEach(item => {
                const domain = (item.scope && item.scope.domain) ? item.scope.domain : (item.scope_domain || 'general');
                const system = (item.scope && item.scope.system) ? item.scope.system : (item.scope_system || 'core');
                
                const card = document.createElement('div');
                card.className = 'p-3.5 bg-slate-950 border border-amber-500/30 rounded-lg flex items-center justify-between';
                card.innerHTML =
                    '<div>' +
                        '<div class="text-amber-400 font-bold">ID: ' + item.id + '</div>' +
                        '<p class="text-slate-200 text-xs font-sans mt-0.5">' + item.title + '</p>' +
                        '<span class="text-[10px] text-slate-500">Domain: ' + domain + ' / System: ' + system + '</span>' +
                    '</div>' +
                    '<span class="px-2 py-1 bg-amber-500/20 text-amber-300 border border-amber-500/40 rounded text-[10px] uppercase font-bold">Quarantined</span>';
                container.appendChild(card);
            });
        }

        function renderDomainChart(breakdown) {
            const ctx = document.getElementById('domainOpsChart').getContext('2d');
            const labels = Object.keys(breakdown);
            const values = Object.values(breakdown);

            if (domainChart) domainChart.destroy();

            domainChart = new Chart(ctx, {
                type: 'bar',
                data: {
                    labels: labels.length ? labels : ['No Active Domains'],
                    datasets: [{
                        label: 'Active Canonical Rules',
                        data: values.length ? values : [0],
                        backgroundColor: '#7c3aed',
                        borderRadius: 4
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: { legend: { display: false } },
                    scales: {
                        x: { grid: { display: false }, ticks: { color: '#94a3b8' } },
                        y: { grid: { color: '#1e293b' }, ticks: { color: '#94a3b8' }, beginAtZero: true }
                    }
                }
            });
        }

        // D3 Graph Rendering
        async function renderGraph() {
            const container = document.getElementById('graph-container');
            if (!container) return;
            const width = container.clientWidth || 800;
            const height = container.clientHeight || 400;

            try {
                const res = await fetch('/api/v1/graph?workspace=my-workspace');
                if (!res.ok) throw new Error('Failed to fetch graph data');
                const data = await res.json();

                // Clear previous
                d3.select(container).selectAll('*').remove();

                const svg = d3.select(container)
                    .append('svg')
                    .attr('width', width)
                    .attr('height', height)
                    .append('g');

                const nodes = (data.nodes || []).map(function(n) {
                    return {
                        id: n.id || n.ID,
                        label: n.label || n.Label,
                        kind: n.kind || n.Kind,
                        package: n.package || n.Package,
                        file: n.file || n.File,
                        exported: Boolean(n.exported ?? n.Exported)
                    };
                });
                const edges = (data.edges || []).map(function(e) {
                    return {
                        source: e.from || e.From,
                        target: e.to || e.To,
                        type: e.type || e.Type
                    };
                });

                const simulation = d3.forceSimulation(nodes)
                    .force('link', d3.forceLink(edges).id(function(d) { return d.id; }).distance(80).strength(0.3))
                    .force('charge', d3.forceManyBody().strength(-200))
                    .force('center', d3.forceCenter(width/2, height/2));

                const link = svg.append('g')
                    .selectAll('line')
                    .data(edges)
                    .enter().append('line')
                    .attr('class', 'link');

                const colorMap = { 'struct': '#3b82f6', 'interface': '#8b5cf6', 'func': '#f59e0b', 'type': '#10b981' };
                function getColor(k) { return colorMap[k] || '#ef4444'; }

                const node = svg.append('g')
                    .selectAll('g')
                    .data(nodes)
                    .enter().append('g')
                    .call(d3.drag()
                        .on('start', dragstarted)
                        .on('drag', dragged)
                        .on('end', dragended)
                    );

                node.append('circle')
                    .attr('r', 12)
                    .attr('fill', function(d) { return getColor(d.kind); })
                    .attr('stroke', '#fff')
                    .attr('stroke-width', 1.5);

                node.append('text')
                    .text(function(d) { return d.label; })
                    .attr('x', 16)
                    .attr('y', 4)
                    .attr('class', 'node-label');

                simulation.on('tick', function() {
                    link
                        .attr('x1', function(d) { return d.source.x; })
                        .attr('y1', function(d) { return d.source.y; })
                        .attr('x2', function(d) { return d.target.x; })
                        .attr('y2', function(d) { return d.target.y; });
                    node
                        .attr('transform', function(d) { return 'translate(' + d.x + ',' + d.y + ')'; });
                });

                function dragstarted(event, d) {
                    if (!event.active) simulation.alphaTarget(0.3).restart();
                    d.fx = d.x;
                    d.fy = d.y;
                }
                function dragged(event, d) {
                    d.fx = event.x;
                    d.fy = event.y;
                }
                function dragended(event, d) {
                    if (!event.active) simulation.alphaTarget(0);
                    d.fx = null;
                    d.fy = null;
                }

                // Resize handler
                const resizeHandler = function() {
                    const newWidth = container.clientWidth;
                    const newHeight = container.clientHeight;
                    svg.attr('width', newWidth).attr('height', newHeight);
                    simulation.force('center', d3.forceCenter(newWidth/2, newHeight/2));
                    simulation.alpha(0.3).restart();
                };
                window.removeEventListener('resize', resizeHandler);
                window.addEventListener('resize', resizeHandler);

            } catch (err) {
                console.error('Graph error:', err);
                container.innerHTML = '<div class="text-slate-400 text-sm p-4">Graph data unavailable. Ensure workspace has entities.</div>';
            }
        }

        async function submitRealDecision() {
            const rawInput = document.getElementById('newDecisionTitle').value.trim();
            let domain = document.getElementById('newDecisionDomain').value.trim();
            let system = document.getElementById('newDecisionSystem').value.trim();

            if (!rawInput) return;

            let title = rawInput;

            const titleMatch = rawInput.match(/"([^"]+)"/);
            if (titleMatch) {
                title = titleMatch[1];
            } else {
                title = rawInput.replace(/\/garuda\s+propose\s+|^propose\s+/, '')
                                .replace(/--scope-domain\s+[^\s]+/, '')
                                .replace(/--scope-system\s+[^\s]+/, '')
                                .trim();
            }

            const domainMatch = rawInput.match(/--scope-domain\s+([^\s]+)/);
            if (domainMatch) domain = domainMatch[1];

            const systemMatch = rawInput.match(/--scope-system\s+([^\s]+)/);
            if (systemMatch) system = systemMatch[1];

            try {
                const tokRes = await fetch('/debug/token?actor=dashboard-ui&tenant_id=00000000-0000-0000-0000-000000000001');
                const tokData = await tokRes.json();

                const res = await fetch('/api/v1/decisions/submit', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': 'Bearer ' + tokData.token
                    },
                    body: JSON.stringify({
                        title: title,
                        scope_domain: domain || "security",
                        scope_system: system || "network"
                    })
                });

                closeProvisionModal();
                fetchRealData();
            } catch (err) {
                alert('Error submitting decision: ' + err.message);
            }
        }

        fetchRealData();
        setInterval(fetchRealData, 5000);
        // Initial graph render after page load (if workforce tab is active)
        setTimeout(renderGraph, 1000);
    </script>
</body>
</html>`

var parsedProdDashboardTmpl = template.Must(template.New("dashboard").Parse(prodDashboardHTML))

func (s *Server) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	data := DashboardData{
		TenantID: tenantID.String(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = parsedProdDashboardTmpl.Execute(w, data)
}

func (s *Server) HandleDashboardStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	activeDecisions, err := s.store.GetDecisionsActiveAt(ctx, tenantID, time.Now().UTC(), types.Scope{}, nil)
	if err != nil {
		activeDecisions = []*types.Decision{}
	}

	canonicalCount := 0
	quarantinedList := []*types.Decision{}
	domainBreakdown := make(map[string]int)
	agentMap := make(map[string]AgentFleetItem)

	for _, d := range activeDecisions {
		if d == nil {
			continue
		}

		domain := d.ScopeDomain
		if domain == "" {
			domain = d.Scope.Domain
		}
		if domain == "" {
			domain = "general"
		}

		system := d.ScopeSystem
		if system == "" {
			system = d.Scope.System
		}
		if system == "" {
			system = "core"
		}

		owner := d.Owner
		if owner == "" {
			owner = "mcp-agent"
		}

		if d.Status == types.StatusQuarantined {
			quarantinedList = append(quarantinedList, d)
		} else {
			canonicalCount++
			domainBreakdown[domain]++
		}

		if _, exists := agentMap[owner]; !exists {
			agentMap[owner] = AgentFleetItem{
				ID:     uuid.NewSHA1(uuid.NameSpaceOID, []byte(owner)).String()[:8],
				Name:   owner,
				Domain: domain,
				System: system,
				Status: string(d.Status),
			}
		}
	}

	agentList := make([]AgentFleetItem, 0, len(agentMap))
	for _, agent := range agentMap {
		agentList = append(agentList, agent)
	}

	latestSnap, _ := s.store.GetLatestMerkleSnapshot(ctx, tenantID)
	latestHash := "No snapshots recorded yet"
	parentHash := "Genesis"
	var latestBlock int64 = 0

	if latestSnap != nil {
		latestHash = latestSnap.SnapshotHash
		latestBlock = latestSnap.BlockHeight
		if latestSnap.ParentSnapshotID != nil {
			parentHash = latestSnap.ParentSnapshotID.String()
		}
	}

	savings := float64(canonicalCount) * 0.20

	resp := RealStatsResponse{
		TotalDecisions:       canonicalCount,
		QuarantinedCount:     len(quarantinedList),
		LatestBlockHeight:    latestBlock,
		LatestMerkleHash:     latestHash,
		ParentMerkleHash:     parentHash,
		EstimatedSavings:     savings,
		DomainBreakdown:      domainBreakdown,
		QuarantinedDecisions: quarantinedList,
		AgentList:            agentList,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleLiveEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	rc := http.NewResponseController(w)
	flushFunc := func() error {
		return rc.Flush()
	}

	if err := rc.Flush(); err != nil {
		var flusher http.Flusher
		curr := w
		for curr != nil {
			if f, ok := curr.(http.Flusher); ok {
				flusher = f
				break
			}
			if unwrapper, ok := curr.(interface{ Unwrap() http.ResponseWriter }); ok {
				curr = unwrapper.Unwrap()
			} else {
				break
			}
		}
		if flusher != nil {
			flushFunc = func() error {
				flusher.Flush()
				return nil
			}
		}
	}

	_, _ = fmt.Fprintf(w, "event: connected\ndata: {\"status\": \"online\", \"timestamp\": \"%s\"}\n\n", time.Now().Format(time.RFC3339))
	_ = flushFunc()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case t := <-ticker.C:
			_, err := fmt.Fprintf(w, "event: ping\ndata: {\"time\": \"%s\"}\n\n", t.Format(time.RFC3339))
			if err != nil {
				return
			}
			_ = flushFunc()
		}
	}
}

// HandleGraph serves D3 graph data for the dashboard
func (s *Server) HandleGraph(w http.ResponseWriter, r *http.Request) {
	// Ensure we have a Postgres store (which has Pool() and graph query methods)
	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		http.Error(w, "Graph data not available", http.StatusServiceUnavailable)
		return
	}

	ctx := r.Context()
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	workspace := r.URL.Query().Get("workspace")
	if workspace == "" {
		workspace = "my-workspace"
	}

	var wsID uuid.UUID
	err := pgStore.Pool().QueryRow(ctx, `
		SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2
	`, tenantID, workspace).Scan(&wsID)
	if err != nil {
		http.Error(w, "Workspace not found", http.StatusNotFound)
		return
	}

	// Query entities
	rows, err := pgStore.Pool().Query(ctx, `
		SELECT id, name, kind, package, file_path, is_exported
		FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, wsID)
	if err != nil {
		http.Error(w, "Failed to query entities: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var nodes []GraphNode
	for rows.Next() {
		var id, name, kind, pkg, file string
		var exported bool
		if err := rows.Scan(&id, &name, &kind, &pkg, &file, &exported); err != nil {
			continue
		}
		nodes = append(nodes, GraphNode{
			ID:       id,
			Label:    name,
			Kind:     kind,
			Package:  pkg,
			File:     file,
			Exported: exported,
		})
	}

	// Query claims
	rows2, err := pgStore.Pool().Query(ctx, `
		SELECT from_entity_id, to_entity_id, claim_type
		FROM claims
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, wsID)
	if err != nil {
		http.Error(w, "Failed to query claims: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows2.Close()

	var edges []GraphEdge
	for rows2.Next() {
		var from, to, typ string
		if err := rows2.Scan(&from, &to, &typ); err != nil {
			continue
		}
		edges = append(edges, GraphEdge{From: from, To: to, Type: typ})
	}

	resp := GraphResponse{Nodes: nodes, Edges: edges}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
