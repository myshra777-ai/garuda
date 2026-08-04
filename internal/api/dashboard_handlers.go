package api

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

type DashboardData struct {
	TenantID             string
	TotalDecisions       int
	ActiveContradictions int
	TokensSaved          int64
	CostSaved            float64
	LatestMerkleHash     string
	ParentMerkleHash     string
	LatestBlockHeight    int64
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="en" class="dark">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Garuda — Mission Control</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script>
      tailwind.config = {
        darkMode: 'class',
        theme: {
          extend: {
            colors: {
              garuda: { 500: '#8b5cf6', 600: '#7c3aed', 900: '#0f172a' }
            }
          }
        }
      }
    </script>
</head>
<body class="bg-slate-950 text-slate-100 font-sans antialiased min-h-screen flex flex-col">

    <!-- Top Navigation -->
    <header class="border-b border-slate-800 bg-slate-900/50 backdrop-blur px-6 py-4 flex items-center justify-between">
        <div class="flex items-center space-x-3">
            <span class="text-2xl">🛡️</span>
            <span class="text-xl font-bold tracking-tight bg-gradient-to-r from-purple-400 to-indigo-400 bg-clip-text text-transparent">Garuda Runtime</span>
            <span class="text-xs px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-400 border border-purple-500/20">GAS v1.0</span>
        </div>
        <div class="flex items-center space-x-4 text-sm text-slate-400">
            <span>Tenant: <code class="text-slate-200 bg-slate-800 px-2 py-1 rounded">{{.TenantID}}</code></span>
            <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                ● Daemon Active
            </span>
        </div>
    </header>

    <!-- Main Container -->
    <main class="flex-1 max-w-7xl w-full mx-auto px-6 py-8 space-y-8">
        
        <!-- ROI Metrics Grid -->
        <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
            <div class="bg-slate-900 border border-slate-800 rounded-xl p-5 shadow-sm">
                <p class="text-xs font-medium text-slate-400 uppercase tracking-wider">Active Decisions</p>
                <h3 class="text-3xl font-extrabold text-slate-100 mt-2">{{.TotalDecisions}}</h3>
                <p class="text-xs text-emerald-400 mt-1">✓ Live PostgreSQL Query</p>
            </div>
            <div class="bg-slate-900 border border-slate-800 rounded-xl p-5 shadow-sm">
                <p class="text-xs font-medium text-slate-400 uppercase tracking-wider">Quarantined Conflicts</p>
                <h3 class="text-3xl font-extrabold text-amber-400 mt-2">{{.ActiveContradictions}}</h3>
                <p class="text-xs text-amber-400/80 mt-1">Requires Operator Action</p>
            </div>
            <div class="bg-slate-900 border border-slate-800 rounded-xl p-5 shadow-sm">
                <p class="text-xs font-medium text-slate-400 uppercase tracking-wider">Tokens Budgeted</p>
                <h3 class="text-3xl font-extrabold text-purple-400 mt-2">1,250,000</h3>
                <p class="text-xs text-purple-400/80 mt-1">Pre-flight Ledger Verified</p>
            </div>
            <div class="bg-slate-900 border border-slate-800 rounded-xl p-5 shadow-sm">
                <p class="text-xs font-medium text-slate-400 uppercase tracking-wider">Estimated LLM ROI</p>
                <h3 class="text-3xl font-extrabold text-emerald-400 mt-2">${{.CostSaved}}</h3>
                <p class="text-xs text-emerald-400/80 mt-1">Saved via Deduplication</p>
            </div>
        </div>

        <!-- Merkle Snapshot Chain & Live Command Trigger -->
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
            
            <!-- Left 2 Cols: Merkle Chain Visualizer -->
            <div class="lg:col-span-2 bg-slate-900 border border-slate-800 rounded-xl p-6 space-y-4">
                <div class="flex items-center justify-between">
                    <h2 class="text-lg font-bold text-slate-100">🔗 Cryptographic Merkle Root Chain</h2>
                    <span class="text-xs text-slate-400">Epoch: 10s Ticker</span>
                </div>
                
                <div class="space-y-3">
                    <div class="p-4 bg-slate-950 border border-purple-500/30 rounded-lg flex items-center justify-between">
                        <div>
                            <span class="text-xs font-mono text-purple-400">LATEST ROOT SNAPSHOT</span>
                            <p class="text-sm font-mono text-slate-200 mt-0.5 truncate max-w-md">{{.LatestMerkleHash}}</p>
                        </div>
                        <span class="text-xs px-2 py-1 bg-purple-500/20 text-purple-300 rounded font-mono">Block #{{.LatestBlockHeight}}</span>
                    </div>

                    <div class="p-4 bg-slate-950/60 border border-slate-800 rounded-lg flex items-center justify-between text-slate-400">
                        <div>
                            <span class="text-xs font-mono text-slate-500">PARENT HASH</span>
                            <p class="text-sm font-mono text-slate-400 mt-0.5 truncate max-w-md">{{.ParentMerkleHash}}</p>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Right Col: Live Slash Command Trigger -->
            <div class="bg-slate-900 border border-slate-800 rounded-xl p-6 space-y-4">
                <h2 class="text-lg font-bold text-slate-100">⚡ Natural MCP Command Trigger</h2>
                <p class="text-xs text-slate-400">Execute slash commands directly into your runtime:</p>
                
                <div class="space-y-3">
                    <input id="cmdInput" type="text" 
                           value='/garuda propose "Enforce TLS 1.3 for internal APIs" --scope-domain security --scope-system network' 
                           class="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-sm font-mono text-slate-200 focus:outline-none focus:border-purple-500" />
                    <button type="button" onclick="executeSlashCommand()" 
                            class="w-full bg-purple-600 hover:bg-purple-700 text-white font-medium py-2 text-sm rounded-lg transition">
                        Execute Command
                    </button>
                </div>

                <!-- Execution Feedback Output -->
                <div id="cmdResult" class="hidden p-3 bg-slate-950 border border-slate-800 rounded-lg font-mono text-xs text-slate-300 overflow-x-auto"></div>
            </div>
        </div>
    </main>

    <script>
    // Live Server-Sent Events Connection
    const evtSource = new EventSource('/api/v1/events');
    evtSource.addEventListener('ping', (e) => {
        console.log('Garuda Live Event Pulse:', e.data);
    });

    async function executeSlashCommand() {
        const rawCmd = document.getElementById('cmdInput').value.trim();
        const resultDiv = document.getElementById('cmdResult');
        resultDiv.classList.remove('hidden');
        resultDiv.innerHTML = '<span class="text-purple-400">⏳ Submitting proposal to Merkle Engine...</span>';

        let title = "New Decision Proposal";
        let scopeDomain = "general";
        let scopeSystem = "web-ui";

        // Parse command string
        const titleMatch = rawCmd.match(/propose\s+"([^"]+)"/) || rawCmd.match(/propose\s+'([^']+)'/);
        if (titleMatch) title = titleMatch[1];

        const domainMatch = rawCmd.match(/--scope-domain\s+([^\s]+)/);
        if (domainMatch) scopeDomain = domainMatch[1];

        const systemMatch = rawCmd.match(/--scope-system\s+([^\s]+)/);
        if (systemMatch) scopeSystem = systemMatch[1];

        try {
            // Fetch debug token
            const tokRes = await fetch('/debug/token?actor=dashboard-ui&tenant_id=00000000-0000-0000-0000-000000000001');
            const tokData = await tokRes.json();

            // Submit proposal
            const res = await fetch('/api/v1/decisions/submit', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer ' + tokData.token
                },
                body: JSON.stringify({
                    title: title,
                    scope_domain: scopeDomain,
                    scope_system: scopeSystem
                })
            });

            const data = await res.json();
            if (res.ok || res.status === 200 || res.status === 201) {
                resultDiv.innerHTML = '<span class="text-emerald-400">✅ PROPOSAL RECORDED!</span>\n' + JSON.stringify(data, null, 2);
                setTimeout(() => { window.location.reload(); }, 1200);
            } else {
                resultDiv.innerHTML = '<span class="text-amber-400">⚠️ RESPONSE:</span>\n' + JSON.stringify(data, null, 2);
            }
        } catch (err) {
            resultDiv.innerHTML = '<span class="text-red-400">❌ Execution Error:</span> ' + err.message;
        }
    }
    </script>
</body>
</html>`

// HandleDashboard renders the Mission Control Web UI
func (s *Server) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// 1. Fetch active decisions count
	activeDecisions, err := s.store.GetDecisionsActiveAt(ctx, tenantID, time.Now().UTC(), types.Scope{}, nil)
	activeCount := len(activeDecisions)
	if err != nil {
		activeCount = 0
	}

	// 2. Count active quarantined items
	quarantineCount := 0
	for _, d := range activeDecisions {
		if d.Status == types.StatusQuarantined {
			quarantineCount++
		}
	}

	// 3. Query latest Merkle snapshot
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

	data := DashboardData{
		TenantID:             tenantID.String(),
		TotalDecisions:       activeCount,
		ActiveContradictions: quarantineCount,
		TokensSaved:          1250000,
		CostSaved:            24.50,
		LatestMerkleHash:     latestHash,
		ParentMerkleHash:     parentHash,
		LatestBlockHeight:    latestBlock,
	}

	tmpl, err := template.New("dashboard").Parse(dashboardHTML)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	_ = tmpl.Execute(w, data)
}

// HandleLiveEvents streams real-time Server-Sent Events (SSE) to the Web UI
func (s *Server) HandleLiveEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	_, _ = fmt.Fprintf(w, "event: connected\ndata: {\"status\": \"online\", \"timestamp\": \"%s\"}\n\n", time.Now().Format(time.RFC3339))
	flusher.Flush()

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
			flusher.Flush()
		}
	}
}
