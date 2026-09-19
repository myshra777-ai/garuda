var urlParams = new URLSearchParams(window.location.search);
var WORKSPACE = urlParams.get("workspace") || window.GARUDA_WORKSPACE || "";

// viewToSection maps a tab name to the section ID that renders it.
// Search is a virtual view — not in the tab strip — reachable from
// the top bar and the search hint buttons.
var viewToSection = {
    workspace:  "view-overview",
    agents:     "view-agents",
    governance: "view-trust",
    decisions:  "view-decisions",
    graph:      "view-architecture",
    search:     "view-search"
};

var tabs = ["workspace", "agents", "governance", "decisions", "graph"];

// initialView reads ?tab= from the URL. Unknown or missing values
// fall back to workspace. Search is never a URL-routed tab.
function initialView() {
    var params = new URLSearchParams(window.location.search);
    var tab = params.get("tab");
    if (viewToSection[tab] && tab !== "search") return tab;
    return "workspace";
}

var state = {
    stats: null,
    currentView: "overview",
    currentLevel: "repository",
    currentFocus: "",
    graphData: null,
    graphZoom: null,
    graphSvg: null,
    graphGroup: null,
    graphSimulation: null,
    searchTimer: null,
    activeCommunities: new Set(),
    hoveredNodeId: null,
    graphCache: {},
};

var communityPalette = [
    "#38bdf8", "#f59e0b", "#ef4444", "#10b981", "#a855f7", "#06b6d4",
    "#ec4899", "#84cc16", "#818cf8", "#f97316", "#14b8a6", "#eab308"
];

function getCommunityColor(node) {
    if (!node) return communityPalette[0];
    if (node.status === "CONTRADICTED") return "#f43f5e";
    var str = node._community || node.package || node.repo || node.id || "general";
    var hash = 0;
    for (var i = 0; i < str.length; i++) {
        hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }
    return communityPalette[Math.abs(hash) % communityPalette.length];
}

// loadWorkspacePicker fetches the list of workspaces the current
// session user is a member of and populates the sidebar select.
// The list is scoped to membership by the server; the client does
// not filter.
async function loadWorkspacePicker() {
    try {
        var res = await fetch("/api/v1/workspaces", {
            headers: { "Accept": "application/json" }
        });
        if (!res.ok) {
            console.error("workspace picker: HTTP " + res.status);
            return;
        }
        var list = await res.json();
        var sel = document.getElementById("workspace-picker");
        if (!sel) return;
        sel.innerHTML = "";
        if (!list || list.length === 0) {
            var opt = document.createElement("option");
            opt.value = "";
            opt.textContent = "No workspaces";
            opt.disabled = true;
            opt.selected = true;
            sel.appendChild(opt);
            return;
        }
        list.forEach(function(w) {
            var opt = document.createElement("option");
            opt.value = w.name;
            opt.textContent = w.name + " (" + w.role + ")";
            if (w.name === WORKSPACE) opt.selected = true;
            sel.appendChild(opt);
        });
    } catch (err) {
        console.error("workspace picker load failed:", err);
    }
}

function switchWorkspace(name) {
    if (!name || name === WORKSPACE) return;
    var params = new URLSearchParams(window.location.search);
    params.set("workspace", name);
    window.location.search = params.toString();
}

function toggleFullscreen() {
    var wrap = document.getElementById("graph-wrap");
    wrap.classList.toggle("fullscreen");
    setTimeout(fitGraph, 200);
}

function setArchitectureMode(level) {
    state.currentLevel = level;
    state.currentFocus = "";
    loadArchitecture(level, "");
}

function showView(view) {
    // Compatibility: legacy call sites use the old view names
    // (overview, architecture, trust). New ones use tab names.
    var aliases = { overview: "workspace", architecture: "graph", trust: "governance" };
    view = aliases[view] || view;

    if (!viewToSection[view]) return;
    state.currentView = view;

    // Section visibility: exactly one section shown at a time.
    Object.keys(viewToSection).forEach(function(key) {
        var el = document.getElementById(viewToSection[key]);
        if (!el) return;
        el.style.display = (key === view) ? "block" : "none";
    });

    // Tab active state. Search has no tab, so all tabs go inactive.
    document.querySelectorAll(".tab").forEach(function(el) {
        var isActive = el.dataset.tab === view;
        el.classList.toggle("active", isActive);
        el.setAttribute("aria-selected", isActive ? "true" : "false");
    });

    // Lazy-load only when this view becomes visible.
    if (view === "graph") {
        loadArchitecture(state.currentLevel || "repository", state.currentFocus || "");
    }
    if (view === "governance") {
        renderTrust();
    }
    if (view === "agents") {
        loadAgents();
    }
    if (view === "decisions") {
        loadDecisions();
    }

    // URL. Search is a transient overlay; it does not change ?tab=.
    if (view !== "search") {
        var params = new URLSearchParams(window.location.search);
        if (params.get("tab") !== view) {
            params.set("tab", view);
            history.pushState({ tab: view }, "", window.location.pathname + "?" + params.toString());
        }
    }
}

window.addEventListener("popstate", function(event) {
    var view = (event.state && event.state.tab) || initialView();
    showView(view);
});

document.addEventListener("keydown", function(e) {
    if ((e.ctrlKey || e.metaKey) && e.key >= "1" && e.key <= "5") {
        var idx = parseInt(e.key, 10) - 1;
        if (idx < tabs.length) {
            e.preventDefault();
            showView(tabs[idx]);
        }
    }
});

async function loadStats() {
    try {
        var res = await fetch("/api/v1/dashboard/stats?workspace=" + encodeURIComponent(WORKSPACE), {
            headers: { "Accept": "application/json" }
        });
        if (!res.ok) throw new Error("Stats request failed: " + res.status);
        state.stats = await res.json();
        renderStats();
    } catch (err) {
        console.error("Garuda stats error:", err);
    }
    await loadPolicies();
}

async function loadPolicies() {
    try {
        var res = await fetch("/api/v1/dashboard/policies?workspace=" + encodeURIComponent(WORKSPACE), {
            headers: { "Accept": "application/json" }
        });
        if (!res.ok) throw new Error("Policies request failed: " + res.status);
        var data = await res.json();
        renderPolicies(data);
    } catch (err) {
        console.error("Garuda policies error:", err);
    }
}

function renderPolicies(data) {
    if (!data) return;

    setText("policy-block-count",  formatNumber(data.blocked_count || 0));
    setText("policy-review-count", formatNumber(data.review_count || 0));
    setText("policy-warn-count",   formatNumber(data.warn_count || 0));
    setText("policy-allow-count",  formatNumber(data.allow_count || 0));
    setText("stat-active-policies", formatNumber(data.active_policies || 0));

    var badge = document.getElementById("policy-final-decision-badge");
    if (badge) {
        badge.textContent = data.final_decision || "ALLOW";
        badge.className = "badge-pill " + (
            data.final_decision === "BLOCK"  ? "critical" :
            data.final_decision === "REVIEW" ? "warning"  :
            data.final_decision === "WARN"   ? "warning"  : "success"
        );
    }

    var chip = document.getElementById("policy-latest-anchor");
    if (chip) {
        if (data.latest_merkle_anchor) {
            chip.textContent = "block #" + data.latest_merkle_anchor;
            chip.classList.remove("invalid");
        } else {
            chip.textContent = "unanchored";
            chip.classList.add("invalid");
        }
    }

    var list = document.getElementById("policy-eval-list");
    if (!list) return;

    var evals = data.latest_evaluations || [];
    if (evals.length === 0) {
        list.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No evaluations yet</div><div class="row-meta">Run <code>garuda policy evaluate &lt;dir&gt;</code> to record decisions.</div></div></div>';
        return;
    }

    list.innerHTML = "";
    evals.forEach(function(ev) {
        var row = document.createElement("div");
        row.className = "policy-eval-row";

        var anchorChip = "";
        if (ev.merkle_block_height) {
            anchorChip = '<span class="merkle-chip' + (ev.anchor_valid ? '' : ' invalid') + '">block #' + ev.merkle_block_height + '</span>';
        } else {
            anchorChip = '<span class="merkle-chip invalid">unanchored</span>';
        }

        var preds = (ev.matched_predicates || []).join(", ") || "—";
        var when = formatDate(ev.evaluated_at);

        row.innerHTML =
            '<span class="policy-badge ' + escapeHTML(ev.decision) + '">' + escapeHTML(ev.decision) + '</span>' +
            '<div class="policy-eval-main">' +
                '<div class="policy-eval-title">' + escapeHTML(ev.policy_title) + anchorChip + '</div>' +
                '<div class="policy-eval-reason">' + escapeHTML(ev.reason) + '</div>' +
                '<div class="policy-eval-meta">' +
                    '<span><strong>' + formatNumber(ev.entity_count) + '</strong> entities</span>' +
                    '<span><strong>' + formatNumber(ev.claim_count) + '</strong> claims</span>' +
                    '<span><strong>' + formatNumber(ev.contradiction_count) + '</strong> contradictions</span>' +
                    '<span>predicates: <strong>' + escapeHTML(preds) + '</strong></span>' +
                    '<span>' + escapeHTML(when) + '</span>' +
                '</div>' +
            '</div>';

        row.style.cursor = "pointer";
        row.onclick = function() {
            openPolicyDrawer(ev);
        };

        list.appendChild(row);
    });
}

function openPolicyDrawer(ev) {
    if (!ev) return;
    var drawer = document.getElementById("drawer");
    var overlay = document.getElementById("drawer-overlay");
    var body = document.getElementById("drawer-body");

    setText("drawer-title", ev.policy_title || "Policy Evaluation");
    setText("drawer-kind", "POLICY " + (ev.decision || ""));

    var html = '<div class="detail-section"><div class="detail-section-title">Decision</div>';
    html += propertyRow("Decision", ev.decision);
    html += propertyRow("Reason", ev.reason);
    html += propertyRow("Actor", ev.actor || "—");
    html += propertyRow("Subject", (ev.subject_kind || "manual") + "/" + (ev.subject_id || ""));
    html += propertyRow("Evaluated", formatDate(ev.evaluated_at));
    html += '</div>';

    html += '<div class="detail-section"><div class="detail-section-title">Evidence</div>';
    html += propertyRow("Entities", ev.entity_count);
    html += propertyRow("Claims", ev.claim_count);
    html += propertyRow("Contradictions", ev.contradiction_count);
    html += propertyRow("Predicates", (ev.matched_predicates || []).join(", ") || "—");
    html += '</div>';

    html += '<div class="detail-section"><div class="detail-section-title">Cryptographic Anchor</div>';
    if (ev.merkle_block_height) {
        html += propertyRow("Block height", "#" + ev.merkle_block_height);
        html += propertyRow("Proof stored", ev.anchor_valid ? "Yes" : "No");
        html += '<div class="merkle" style="margin-top:8px;">eval_id=' + escapeHTML(ev.id) + '</div>';
        html += '<button class="detail-action" onclick="verifyPolicyAnchor(\'' + escapeJS(ev.id) + '\')">Verify anchor →</button>';
    } else {
        html += '<div class="row-meta">This evaluation has no Merkle anchor.</div>';
    }
    html += '</div>';

    body.innerHTML = html;
    drawer.classList.add("open");
    overlay.classList.add("open");
}

function verifyPolicyAnchor(evalID) {
    var url = "/api/v1/dashboard/policies/verify?workspace=" +
        encodeURIComponent(WORKSPACE) + "&id=" + encodeURIComponent(evalID);
    fetch(url)
        .then(function(res) { return res.json(); })
        .then(function(data) {
            if (data.valid) {
                var versionTag = data.version === 1 ? "v1" : "v0";
                alert("Merkle anchor VALID (" + versionTag + ") for evaluation " + evalID + " at block #" + data.block_height);
            } else {
                alert("Merkle anchor INVALID: " + (data.error || "unknown"));
            }
        })
        .catch(function(err) {
            alert("Verification request failed: " + err);
        });
}

// renderMeasuredMetric renders a Class B metric with three honest
// states:
//
//   has_data=true               → the value
//   has_data=false, measured    → "No data"
//   is_measured=false           → "Not measured"
//
// The previous code rendered 0 in all three cases, which read as
// "we measured and the answer is zero."
function renderMeasuredMetric(elementId, metric, valueFormatter) {
    var el = document.getElementById(elementId);
    if (!el) return;
    if (!metric) {
        el.textContent = "—";
        el.title = "Not measured";
        el.style.opacity = "0.5";
        return;
    }
    if (!metric.is_measured) {
        el.textContent = "—";
        el.title = metric.reason || "Not measured";
        el.style.opacity = "0.5";
        return;
    }
    if (!metric.has_data) {
        el.textContent = "—";
        el.title = metric.reason || "No data in this window";
        el.style.opacity = "0.6";
        return;
    }
    el.textContent = valueFormatter(metric);
    el.title = "Measured across " + metric.row_count + " telemetry event(s)";
    el.style.opacity = "1";
}

function renderStats() {
    if (!state.stats) return;
    var s = state.stats;
    setText("stat-repositories", formatNumber(s.repositories));
    setText("stat-packages", formatNumber(s.packages));
    setText("stat-entities", formatNumber(s.entities));
    setText("stat-relationships", formatNumber(s.relationships));
    setText("stat-cross-links", formatNumber(s.cross_repo_links));
    setText("stat-idempotency", formatNumber(s.idempotency_safeguards));

    setText("stat-supported", formatNumber(s.supported_claims));
    setText("stat-unverified", formatNumber(s.unverified_claims));
    setText("stat-contradicted", formatNumber(s.contradicted));

    // Honest card description: distinguish "attempted and found none"
    // from "never attempted".
    if (s.verification_attempted) {
        setText("stat-unverified-desc", "Code exists, zero recent executions");
    } else {
        setText("stat-unverified-desc", "Runtime verification not yet attempted");
    }

    setText("explorer-repos", formatNumber(s.repositories));
    setText("explorer-packages", formatNumber(s.packages));
    setText("explorer-entities", formatNumber(s.entities));

    setText("sidebar-meta", formatNumber(s.repositories) + " repositories · " + formatNumber(s.entities) + " entities");
    setText("sidebar-workspace", s.workspace || WORKSPACE);
    setText("workspace-breadcrumb", s.workspace || WORKSPACE);
    setText("repo-list-ws", s.workspace || WORKSPACE);

    // The deployment-wide metrics (tokens saved, cost saved, agent
    // peak) were removed from this dashboard. They belong on the
    // operator view at /admin, where the schema can support them
    // without leaking one tenant's numbers to another.
    setText("stat-drift-count", formatNumber(s.quarantined_count));

    renderLanguages(s.languages_breakdown || []);
    renderDrift(s.drift || {});
    renderHubs(s.top_hubs || []);
    renderAttention(s.needs_attention || []);
    renderEvidence(s.recent_evidence || []);
    renderTrust();
    renderScannedRepos(s.repo_stats || [], s.repositories_list || []);
}

function renderLanguages(langs) {
    var bar = document.getElementById("lang-bar");
    var legend = document.getElementById("lang-legend");
    if (!bar || !legend) return;

    if (!langs || langs.length === 0) {
        bar.innerHTML = '<div style="width:100%; background:#1a2035;"></div>';
        legend.innerHTML = '<span style="color:var(--muted); font-size:11px;">No language data yet — run garuda analyze on at least one repository.</span>';
        return;
    }

    langs.sort(function(a, b) { return b.count - a.count; });

    bar.innerHTML = "";
    langs.forEach(function(lang) {
        var seg = document.createElement("div");
        seg.className = "lang-bar-segment";
        seg.style.width = lang.percentage + "%";
        seg.style.background = lang.color;
        seg.title = lang.name + " " + lang.percentage.toFixed(1) + "%";
        bar.appendChild(seg);
    });

    legend.innerHTML = "";
    langs.forEach(function(lang) {
        var item = document.createElement("div");
        item.className = "lang-legend-item";
        item.innerHTML = '<span class="lang-legend-dot" style="background:' + lang.color + ';"></span>' +
            '<span class="lang-legend-name">' + escapeHTML(lang.name) + '</span>' +
            '<span class="lang-legend-pct">' + lang.percentage.toFixed(1) + '%</span>';
        legend.appendChild(item);
    });
}

function renderDrift(d) {
    setText("drift-total-docs", formatNumber(d.total_document_claims));
    setText("drift-doc-supported", formatNumber(d.supported_claims));
    setText("drift-doc-unverified", formatNumber(d.unverified_claims));
    setText("drift-doc-contradicted", formatNumber(d.contradicted_claims));
    setText("drift-doc-to-code", formatNumber(d.doc_to_code_drift_count));
    setText("drift-code-to-doc", formatNumber(d.code_to_doc_drift_count));
    setText("drift-undocumented-code", formatNumber(d.undocumented_code));
    // Field name matches DriftDTO JSON tag (unverified_docs), not the DOM id.
    setText("drift-unimplemented-docs", formatNumber(d.unverified_docs));
}

function renderHubs(hubs) {
    var list = document.getElementById("hub-list");
    if (!list) return;
    if (!hubs || hubs.length === 0) {
        list.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Explore high-impact entities</div><div class="row-meta">Open the architecture map to inspect relationship centrality.</div></div><button class="graph-button" onclick="openArchitecture(\'entity\')">Open</button></div>';
        return;
    }
    list.innerHTML = "";
    hubs.forEach(function(h) {
        var row = document.createElement("div");
        row.className = "list-row";
        row.style.cursor = "pointer";
        row.innerHTML = '<div class="row-icon">' + symbolIcon(h.kind) + '</div>' +
            '<div class="row-main">' +
                '<div class="row-title">' + escapeHTML(h.name) + '</div>' +
                '<div class="row-meta">' + escapeHTML(h.kind) + ' · ' + escapeHTML(h.repo) + ' (' + escapeHTML(h.package) + ')</div>' +
            '</div>' +
            '<div style="text-align:right;">' +
                '<div style="font-weight:750; color:var(--brand); font-size:13px;">' + h.callers + '</div>' +
                '<div style="font-size:9px; color:var(--muted); text-transform:uppercase;">callers</div>' +
            '</div>';
        row.onclick = function() { openSearchResult(h); };
        list.appendChild(row);
    });
}

function renderAttention(items) {
    var list = document.getElementById("attention-list");
    if (!list) return;
    if (!items || items.length === 0) {
        list.innerHTML = '<div class="list-row"><div class="row-icon" style="color:var(--green);">✓</div><div class="row-main"><div class="row-title">Zero Active Violations</div><div class="row-meta">All static and runtime claims are structurally consistent.</div></div></div>';
        return;
    }
    list.innerHTML = "";
    items.forEach(function(item) {
        var row = document.createElement("div");
        row.className = "list-row";
        row.innerHTML = '<div class="row-icon" style="color:' + (item.severity === 'critical' ? 'var(--red)' : 'var(--amber)') + ';">●</div>' +
            '<div class="row-main">' +
                '<div class="row-title">' + escapeHTML(item.title) + '</div>' +
                '<div class="row-meta">' + escapeHTML(item.subtitle) + ' (' + escapeHTML(item.evidence_loc) + ')</div>' +
            '</div>' +
            '<span class="badge-pill ' + escapeHTML(item.severity) + '">' + escapeHTML(item.severity) + '</span>';
        list.appendChild(row);
    });
}

function renderEvidence(items) {
    var list = document.getElementById("evidence-list");
    if (!list) return;
    if (!items || items.length === 0) {
        list.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Evidence ledger active</div><div class="row-meta">Awaiting additional runtime span ingestions.</div></div></div>';
        return;
    }
    list.innerHTML = "";
    items.forEach(function(item) {
        var row = document.createElement("div");
        row.className = "list-row";
        // Previously every row was badged "Verified" regardless of source.
        // Runtime observations are records, not verified claims. Badge them
        // by kind so a trace does not look like a Merkle anchor.
        var badgeClass = "info";
        var badgeText = "Recorded";
        if (item.kind === "Runtime Trace") { badgeText = "Observed"; }
        row.innerHTML = '<div class="row-icon">📜</div>' +
            '<div class="row-main">' +
                '<div class="row-title">' + escapeHTML(item.summary) + '</div>' +
                '<div class="row-meta">' + escapeHTML(item.kind) + ' · ' + escapeHTML(item.source) + ' · ' + formatDate(item.timestamp) + '</div>' +
            '</div>' +
            '<span class="badge-pill ' + badgeClass + '">' + badgeText + '</span>';
        list.appendChild(row);
    });
}

function renderTrust() {
    if (!state.stats) return;
    setText("trust-status", state.stats.trust_status || "Verified");
    setText("trust-height", "#" + String(state.stats.latest_block_height || 1));
    setText("trust-updated", formatDate(state.stats.last_updated));
    setText("trust-merkle", state.stats.latest_merkle_hash || "Genesis verified");
    setText("trust-parent", state.stats.parent_merkle_hash || "Genesis");
}

function renderScannedRepos(repoStats, fallbackList) {
    var list = document.getElementById("scanned-repos-list");
    if (!list) return;

    if (!repoStats || repoStats.length === 0) {
        if (!fallbackList || fallbackList.length === 0) {
            list.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No repositories registered in this workspace.</div></div></div>';
            return;
        }
        list.innerHTML = "";
        fallbackList.forEach(function(repo) {
            var row = document.createElement("div");
            row.className = "list-row";
            row.style.cursor = "pointer";
            row.innerHTML = '<div class="row-icon">📦</div>' +
                '<div class="row-main">' +
                    '<div class="row-title">' + escapeHTML(repo) + '</div>' +
                    '<div class="row-meta">Source code boundary</div>' +
                '</div>' +
                '<button class="graph-button" onclick="event.stopPropagation(); runSearch(\'' + escapeJS(repo) + '\')">Explore Symbols →</button>';
            row.onclick = function() {
                state.currentLevel = "package";
                state.currentFocus = repo;
                showView("architecture");
            };
            list.appendChild(row);
        });
        return;
    }

    list.innerHTML = "";
    repoStats.forEach(function(repo) {
        var card = document.createElement("div");
        card.className = "repo-card";

        var langBarHTML = "";
        if (repo.languages && repo.languages.length > 0) {
            langBarHTML = '<div class="lang-bar-container" style="margin-top:8px;">';
            repo.languages.forEach(function(l) {
                langBarHTML += '<div class="lang-bar-segment" style="width:' + l.percentage + '%; background:' + l.color + ';" title="' + escapeHTML(l.name) + ' ' + l.percentage.toFixed(1) + '%"></div>';
            });
            langBarHTML += '</div><div style="display:flex; flex-wrap:wrap; gap:12px; margin-top:8px;">';
            repo.languages.forEach(function(l) {
                langBarHTML += '<div class="lang-legend-item"><span class="lang-legend-dot" style="background:' + l.color + ';"></span>' +
                    '<span style="font-size:11px;">' + escapeHTML(l.name) + ' <span class="lang-legend-pct">' + l.percentage.toFixed(1) + '%</span></span></div>';
            });
            langBarHTML += '</div>';
        }

        var statusBadge = repo.analysis_status === "synced"
            ? '<span class="badge-pill success">synced</span>'
            : '<span class="badge-pill info">' + escapeHTML(repo.analysis_status || "pending") + '</span>';

        var commitShort = repo.current_commit && repo.current_commit.length > 8
            ? repo.current_commit.substring(0, 8)
            : (repo.current_commit || "—");

        card.innerHTML =
            '<div class="repo-card-header">' +
                '<div>' +
                    '<div class="repo-card-name">📦 ' + escapeHTML(repo.name) + '</div>' +
                    '<div class="repo-card-meta">Commit ' + escapeHTML(commitShort) + ' · ' + (repo.last_analyzed || "never analyzed") + '</div>' +
                '</div>' +
                statusBadge +
            '</div>' +
            '<div class="repo-card-stats">' +
                '<span><span class="repo-card-stat-val">' + formatNumber(repo.entities) + '</span> entities</span>' +
                '<span><span class="repo-card-stat-val">' + formatNumber(repo.relationships) + '</span> relationships</span>' +
                '<span><span class="repo-card-stat-val">' + formatNumber(repo.files) + '</span> files</span>' +
            '</div>' +
            langBarHTML +
            '<div class="repo-card-actions" style="margin-top:12px;">' +
                '<button class="graph-button" onclick="event.stopPropagation(); runSearch(\'' + escapeJS(repo.name) + '\')">Explore Symbols →</button>' +
                '<button class="graph-button" onclick="event.stopPropagation(); state.currentLevel=\'package\'; state.currentFocus=\'' + escapeJS(repo.name) + '\'; showView(\'architecture\');">View Packages</button>' +
            '</div>';

        list.appendChild(card);
    });
}

function openArchitecture(level) {
    state.currentLevel = level;
    state.currentFocus = "";
    showView("architecture");
}

async function loadArchitecture(level, focus) {
    state.currentLevel = level || "repository";
    state.currentFocus = focus || "";
    updateArchitectureHeader();

    // Cache the graph JSON by (level, focus). Navigating away and back
    // reuses the response instead of paying the fetch cost again. The
    // Refresh button clears state.graphCache before calling this, so
    // explicit refresh still hits the server.
    var cacheKey = state.currentLevel + ":" + state.currentFocus;
    if (state.graphCache[cacheKey]) {
        state.graphData = state.graphCache[cacheKey];
        buildCommunitiesList(state.graphData);
        renderGraph(state.graphData);
        return;
    }

    var url = "/api/v1/graph?workspace=" + encodeURIComponent(WORKSPACE) +
        "&level=" + encodeURIComponent(state.currentLevel);
    if (state.currentFocus) {
        url += "&focus=" + encodeURIComponent(state.currentFocus);
    }

    try {
        var res = await fetch(url);
        if (!res.ok) throw new Error("Graph request failed");
        var data = await res.json();
        state.graphCache[cacheKey] = data;
        state.graphData = data;
        buildCommunitiesList(data);
        renderGraph(data);
    } catch (err) {
        console.error("Architecture error:", err);
        var empty = document.getElementById("graph-empty");
        if (empty) {
            empty.style.display = "grid";
            empty.textContent = "Architecture data unavailable.";
        }
    }
}

function updateArchitectureHeader() {
    var title = "Company Graph";
    var subtitle = "Force-directed semantic topology map.";
    var breadcrumb = "Graph";

    if (state.currentLevel === "full") {
        title = "Entire Knowledge Graph";
        subtitle = "Multi-repository package and symbol dependency mesh.";
        breadcrumb = "Knowledge Graph";
    } else if (state.currentLevel === "package") {
        title = state.currentFocus ? "Packages in " + state.currentFocus : "Workspace packages";
        subtitle = "Package-level structure. Double-click a package to explore symbols.";
        breadcrumb = state.currentFocus || "Packages";
    } else if (state.currentLevel === "entity") {
        title = state.currentFocus ? "Symbol neighborhood" : "Top architectural symbols";
        subtitle = "Local neighborhood rendered with zero hairballs.";
        breadcrumb = state.currentFocus ? "Neighborhood" : "Entities";
    }

    setText("architecture-title", title);
    setText("architecture-subtitle", subtitle);
    setText("architecture-breadcrumb", breadcrumb);
    setText("graph-description", subtitle);
}

function extractCommunityName(node) {
    var comm = node.package || node.repo || "general";
    if (comm.indexOf("myshra777-ai/garuda/") !== -1) {
        var sub = comm.split("myshra777-ai/garuda/")[1];
        var parts = sub.split("/");
        return parts[0] + (parts.length > 1 ? "/" + parts[1] : "");
    }
    if (comm.indexOf("/") !== -1) {
        var p = comm.split("/");
        return p[p.length - 1];
    }
    return comm;
}

function buildCommunitiesList(data) {
    var container = document.getElementById("communities-list");
    if (!container) return;
    if (!data || !data.nodes || data.nodes.length === 0) {
        container.innerHTML = '<div style="color:var(--muted); font-size:11px; padding:8px;">No modules detected.</div>';
        return;
    }

    var counts = {};
    state.activeCommunities = new Set();
    data.nodes.forEach(function(n) {
        var comm = extractCommunityName(n);
        n._community = comm;
        counts[comm] = (counts[comm] || 0) + 1;
        state.activeCommunities.add(comm);
    });

    var sorted = Object.keys(counts).sort(function(a, b) { return counts[b] - counts[a]; });
    container.innerHTML = "";

    sorted.forEach(function(comm) {
        var dummyNode = { _community: comm };
        var color = getCommunityColor(dummyNode);
        var item = document.createElement("div");
        item.className = "community-item";
        item.innerHTML = '<input type="checkbox" class="community-checkbox" checked id="chk-' + escapeHTML(comm) + '">' +
            '<span class="community-dot" style="background:' + color + '; color:' + color + ';"></span>' +
            '<span class="community-name" title="' + escapeHTML(comm) + '">' + escapeHTML(comm) + '</span>' +
            '<span class="community-count">' + counts[comm] + '</span>';

        var chk = item.querySelector("input");
        chk.onchange = function() {
            if (chk.checked) state.activeCommunities.add(comm);
            else state.activeCommunities.delete(comm);
            applyCommunityFilter();
        };

        item.onclick = function(e) {
            if (e.target !== chk) {
                chk.checked = !chk.checked;
                chk.onchange();
            }
        };

        container.appendChild(item);
    });
}

function toggleAllCommunities() {
    var checkboxes = document.querySelectorAll(".community-checkbox");
    var allChecked = true;
    checkboxes.forEach(function(c) { if (!c.checked) allChecked = false; });
    var target = !allChecked;
    checkboxes.forEach(function(c) {
        c.checked = target;
        c.onchange();
    });
}

function applyCommunityFilter() {
    if (!state.graphGroup) return;
    state.graphGroup.selectAll(".graph-node-group")
        .style("display", function(d) {
            return state.activeCommunities.has(d._community) ? "inline" : "none";
        });

    state.graphGroup.selectAll(".graph-link")
        .style("display", function(d) {
            var sComm = d.source._community;
            var tComm = d.target._community;
            return (state.activeCommunities.has(sComm) && state.activeCommunities.has(tComm)) ? "inline" : "none";
        });
}

function renderGraph(data) {
    // Stop any running simulation before creating a new one. The
    // previous render's simulation keeps ticking on D3's timer,
    // computing positions for node objects the SVG no longer holds.
    // Each navigation from and back to the graph leaves one more
    // simulation alive; memory and CPU grow with each round trip.
    if (state.graphSimulation) {
        state.graphSimulation.stop();
        state.graphSimulation = null;
    }

    var svg = d3.select("#graph");
    svg.selectAll("*").remove();

    var empty = document.getElementById("graph-empty");
    if (!data || !data.nodes || data.nodes.length === 0) {
        empty.style.display = "grid";
        return;
    }
    empty.style.display = "none";

    var container = document.querySelector(".graph-wrap");
    var width = container.clientWidth || 900;
    var height = container.clientHeight || 750;

    svg.attr("width", width).attr("height", height);

    var defs = svg.append("defs");

    var filter = defs.append("filter")
        .attr("id", "neon-glow")
        .attr("x", "-50%").attr("y", "-50%")
        .attr("width", "200%").attr("height", "200%");
    filter.append("feGaussianBlur").attr("stdDeviation", "4").attr("result", "coloredBlur");
    var feMerge = filter.append("feMerge");
    feMerge.append("feMergeNode").attr("in", "coloredBlur");
    feMerge.append("feMergeNode").attr("in", "SourceGraphic");

    defs.append("marker").attr("id", "arrow").attr("viewBox", "0 -5 10 10").attr("refX", 24).attr("refY", 0)
        .attr("markerWidth", 5.5).attr("markerHeight", 5.5).attr("orient", "auto")
        .append("path").attr("d", "M0,-5L10,0L0,5").attr("fill", "#38bdf8");

    defs.append("marker").attr("id", "arrow-violation").attr("viewBox", "0 -5 10 10").attr("refX", 24).attr("refY", 0)
        .attr("markerWidth", 6.5).attr("markerHeight", 6.5).attr("orient", "auto")
        .append("path").attr("d", "M0,-5L10,0L0,5").attr("fill", "#f43f5e");

    var zoomLayer = svg.append("g");
    state.graphSvg = svg;
    state.graphGroup = zoomLayer;

    var zoom = d3.zoom().scaleExtent([0.1, 8]).on("zoom", function(event) {
        zoomLayer.attr("transform", event.transform);
    });
    svg.call(zoom);
    state.graphZoom = zoom;

    var nodes = data.nodes.map(function(n) {
        var obj = Object.assign({}, n);
        obj._community = extractCommunityName(obj);
        return obj;
    });
    var nodeByID = {};
    nodes.forEach(function(n) { nodeByID[n.id] = n; });

    var validEdges = (data.edges || []).filter(function(e) {
        return nodeByID[e.from] && nodeByID[e.to];
    }).map(function(e) {
        return {
            id: e.id,
            source: nodeByID[e.from],
            target: nodeByID[e.to],
            type: e.type,
            status: e.status,
            label: e.label,
            count: e.count || 1
        };
    });

    var linkedByIndex = {};
    validEdges.forEach(function(d) {
        linkedByIndex[d.source.id + "," + d.target.id] = true;
        linkedByIndex[d.target.id + "," + d.source.id] = true;
    });

    function isConnected(a, b) {
        return a.id === b.id || linkedByIndex[a.id + "," + b.id];
    }

    var linkDist = state.currentLevel === "full" ? 90 : (state.currentLevel === "repository" ? 140 : 90);
    var chargeForce = state.currentLevel === "full" ? -260 : (state.currentLevel === "repository" ? -750 : -420);

    var simulation = d3.forceSimulation(nodes)
        .alphaDecay(0.08)
        .force("link", d3.forceLink(validEdges).id(function(d) { return d.id; }).distance(linkDist).strength(0.35))
        .force("charge", d3.forceManyBody().strength(chargeForce))
        .force("center", d3.forceCenter(width / 2, height / 2))
        .force("collision", d3.forceCollide().radius(function(d) {
            var val = (d.count || d.Count || 1);
            var logScale = Math.log10(val + 1) * 5.5;
            return 9 + logScale + 8;
        }));

    state.graphSimulation = simulation;

    var link = zoomLayer.append("g").selectAll("path")
        .data(validEdges)
        .enter().append("path")
        .attr("class", function(d) { return d.status === "CONTRADICTED" ? "graph-link violation" : "graph-link"; })
        .attr("stroke", function(d) {
            if (d.status === "CONTRADICTED") return "#f43f5e";
            return getCommunityColor(d.source);
        })
        .attr("stroke-width", function(d) { return Math.min(3.5, 1.2 + Math.log2((d.count || 1) + 1)); })
        .attr("marker-end", function(d) { return d.status === "CONTRADICTED" ? "url(#arrow-violation)" : "url(#arrow)"; });

    var violationEdges = validEdges.filter(function(d) { return d.status === "CONTRADICTED"; });
    var linkLabels = zoomLayer.append("g").selectAll("text")
        .data(violationEdges).enter().append("text")
        .attr("class", "graph-link-label")
        .text(function(d) { return d.label ? d.label : "VIOLATION"; });

    var node = zoomLayer.append("g").selectAll("g")
        .data(nodes).enter().append("g")
        .attr("class", "graph-node-group")
        .style("cursor", "pointer")
        .call(d3.drag()
            .on("start", function(event, d) {
                if (!event.active) simulation.alphaTarget(0.25).restart();
                d.fx = d.x; d.fy = d.y;
            })
            .on("drag", function(event, d) { d.fx = event.x; d.fy = event.y; })
            .on("end", function(event, d) {
                if (!event.active) simulation.alphaTarget(0);
                d.fx = null; d.fy = null;
            }));

    node.append("circle").attr("class", "node-halo")
        .attr("r", function(d) {
            var val = (d.count || d.Count || 1);
            return Math.max(8, 9 + Math.log10(val + 1) * 5) + 6;
        })
        .attr("fill", function(d) { return getCommunityColor(d); })
        .attr("opacity", 0.3)
        .style("filter", "url(#neon-glow)");

    node.append("circle").attr("class", "node-core")
        .attr("r", function(d) {
            var val = (d.count || d.Count || 1);
            return Math.max(7, 8 + Math.log10(val + 1) * 5);
        })
        .attr("fill", function(d) { return getCommunityColor(d); })
        .attr("stroke", "#ffffff")
        .attr("stroke-width", 1.5)
        .style("filter", "drop-shadow(0 0 6px rgba(0,0,0,0.85))");

    node.append("text").attr("class", "graph-node-label")
        .attr("dx", function(d) {
            var val = (d.count || d.Count || 1);
            return Math.max(7, 8 + Math.log10(val + 1) * 5) + 6;
        })
        .attr("dy", 3.5)
        .text(function(d) { return shortenLabel(d.label || d.id, 24); });

    node.on("mouseover", function(event, d) {
        state.hoveredNodeId = d.id;
        node.classed("dimmed", function(o) { return !isConnected(d, o); });
        node.classed("highlighted", function(o) { return isConnected(d, o); });
        link.classed("highlighted", function(o) { return o.source.id === d.id || o.target.id === d.id; });
        link.classed("dimmed", function(o) { return o.source.id !== d.id && o.target.id !== d.id; });
    });

    node.on("mouseout", function() {
        state.hoveredNodeId = null;
        node.classed("dimmed", false);
        node.classed("highlighted", false);
        link.classed("highlighted", false);
        link.classed("dimmed", false);
    });

    node.on("click", function(event, d) {
        event.stopPropagation();
        openNodeDrawer(d);
    });

    node.on("dblclick", function(event, d) {
        event.stopPropagation();
        if (state.currentLevel === "repository" || state.currentLevel === "full") {
            state.currentLevel = "package";
            state.currentFocus = d.label;
            loadArchitecture(state.currentLevel, state.currentFocus);
            return;
        }
        if (state.currentLevel === "package") {
            state.currentLevel = "entity";
            state.currentFocus = d.id;
            loadArchitecture(state.currentLevel, state.currentFocus);
            return;
        }
        if (state.currentLevel === "entity") {
            openNodeDrawer(d);
        }
    });

    var tickCount = 0;
    simulation.on("tick", function() {
        // Stop the simulation once the layout has settled. alphaDecay
        // 0.08 reaches alpha < 0.05 in roughly 40 ticks at this node
        // count. The 400-tick ceiling is a safety net for pathological
        // graphs. Without either condition, D3's default alphaMin lets
        // the simulation run for thousands of ticks, producing minutes
        // of visible drift.
        tickCount++;
        if (tickCount > 400 || simulation.alpha() < 0.05) {
            simulation.stop();
        }

        link.attr("d", function(d) {
            var dx = d.target.x - d.source.x;
            var dy = d.target.y - d.source.y;
            var dr = Math.sqrt(dx * dx + dy * dy) * 1.25;
            return "M" + d.source.x + "," + d.source.y + "A" + dr + "," + dr + " 0 0,1 " + d.target.x + "," + d.target.y;
        });

        linkLabels
            .attr("x", function(d) { return (d.source.x + d.target.x) / 2; })
            .attr("y", function(d) { return (d.source.y + d.target.y) / 2 - 5; });

        node.attr("transform", function(d) { return "translate(" + d.x + "," + d.y + ")"; });
    });

    setTimeout(function() { fitGraph(); }, 600);
}

function fitGraph() {
    if (!state.graphSvg || !state.graphGroup) return;
    var svg = state.graphSvg;
    var group = state.graphGroup;
    var bbox;
    try { bbox = group.node().getBBox(); } catch (e) { return; }
    if (!bbox.width || !bbox.height) return;

    var container = document.querySelector(".graph-wrap");
    var width = container.clientWidth || 900;
    var height = container.clientHeight || 750;

    var scale = Math.min(width / (bbox.width + 140), height / (bbox.height + 140), 1.2);
    scale = Math.max(scale, 0.15);
    var tx = width / 2 - scale * (bbox.x + bbox.width / 2);
    var ty = height / 2 - scale * (bbox.y + bbox.height / 2);

    svg.transition().duration(350).call(state.graphZoom.transform, d3.zoomIdentity.translate(tx, ty).scale(scale));
}

function zoomGraph(multiplier) {
    if (!state.graphSvg || !state.graphZoom) return;
    state.graphSvg.transition().duration(200).call(state.graphZoom.scaleBy, multiplier);
}

function goUpArchitecture() {
    if (state.currentLevel === "entity") {
        state.currentLevel = "package";
        state.currentFocus = "";
        loadArchitecture(state.currentLevel, "");
        return;
    }
    if (state.currentLevel === "package" || state.currentLevel === "full") {
        state.currentLevel = "repository";
        state.currentFocus = "";
        loadArchitecture(state.currentLevel, "");
        return;
    }
    showView("overview");
}

function openNodeDrawer(node) {
    if (!node) return;
    var drawer = document.getElementById("drawer");
    var overlay = document.getElementById("drawer-overlay");
    var body = document.getElementById("drawer-body");

    setText("drawer-title", node.label || node.name || "Unknown");
    setText("drawer-kind", (node.kind || "entity").toUpperCase());

    var html = '<div class="detail-section"><div class="detail-section-title">Identity</div>';
    html += propertyRow("Type", node.kind || "—");
    if (node.repo) html += propertyRow("Repository", node.repo);
    if (node.package) html += propertyRow("Package", node.package);
    if (node.file) html += propertyRow("File", node.file);
    if (node.exported !== undefined) html += propertyRow("Exported", node.exported ? "Yes" : "No");
    if (node.Count || node.count) html += propertyRow("Contained/Centrality", formatNumber(node.Count || node.count));
    html += '</div>';

    html += '<div class="detail-section"><div class="detail-section-title">Actions</div>';
    if (node.kind === "repository" || state.currentLevel === "repository") {
        html += '<button class="detail-action" onclick="closeDrawer(); state.currentLevel=\'package\'; state.currentFocus=\'' + escapeJS(node.label) + '\'; loadArchitecture(\'package\', \'' + escapeJS(node.label) + '\');">Explore packages →</button>';
    }
    if (node.kind === "package" || state.currentLevel === "package") {
        html += '<button class="detail-action" onclick="closeDrawer(); state.currentLevel=\'entity\'; state.currentFocus=\'' + escapeJS(node.id) + '\'; loadArchitecture(\'entity\', \'' + escapeJS(node.id) + '\');">Explore symbols →</button>';
    }
       if (state.currentLevel === "entity" || (node.kind !== "repository" && node.kind !== "package")) {
        // The entity level filters by package, not by entity ID. Passing
        // node.id here sent a UUID to a query that filters on
        // e.Package = focus, which matched zero rows and returned an
        // empty graph. node.package is the field the entity level
        // expects.
        //
        // The cache is cleared before navigating so a stale entry for
        // the same (level, focus) key does not return the previous
        // render's nodes.
        var focusPkg = node.package || node.Package || "";
        html += '<button class="detail-action" onclick="closeDrawer(); state.graphCache = {}; state.currentLevel=\'entity\'; state.currentFocus=\'' + escapeJS(focusPkg) + '\'; showView(\'architecture\');">Explore local neighborhood →</button>';
    }
    html += '</div>';

    body.innerHTML = html;
    drawer.classList.add("open");
    overlay.classList.add("open");
}

function propertyRow(key, value) {
    return '<div class="detail-property"><div class="detail-key">' + escapeHTML(key) + '</div><div class="detail-value">' + escapeHTML(String(value)) + '</div></div>';
}

function closeDrawer() {
    document.getElementById("drawer").classList.remove("open");
    document.getElementById("drawer-overlay").classList.remove("open");
}

function setupSearch() {
    var global = document.getElementById("global-search");
    var quick = document.getElementById("quick-search");
    var page = document.getElementById("search-page-input");

    if (global) {
        global.addEventListener("keydown", function(e) {
            if (e.key === "Enter") {
                var val = global.value.trim();
                if (val) runSearch(val);
            }
        });
    }
    if (quick) {
        quick.addEventListener("keydown", function(e) {
            if (e.key === "Enter") {
                var val = quick.value.trim();
                if (val) runSearch(val);
            }
        });
    }
    if (page) {
        page.addEventListener("input", function() {
            clearTimeout(state.searchTimer);
            var val = page.value.trim();
            state.searchTimer = setTimeout(function() {
                if (val.length >= 2) performSearch(val);
            }, 180);
        });
    }
    document.addEventListener("keydown", function(e) {
        if (e.key === "/" && document.activeElement.tagName !== "INPUT") {
            e.preventDefault();
            if (global) global.focus();
        }
        if (e.key === "Escape") closeDrawer();
    });
}

function runSearch(query) {
    var global = document.getElementById("global-search");
    var quick = document.getElementById("quick-search");
    var page = document.getElementById("search-page-input");
    if (global) global.value = query;
    if (quick) quick.value = query;
    if (page) page.value = query;
    showView("search");
    performSearch(query);
}

function runHint(value) {
    runSearch(value);
}

async function performSearch(query) {
    var results = document.getElementById("search-results");
    if (!results) return;
    results.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Searching...</div></div></div>';

    try {
        var res = await fetch("/api/v1/dashboard/search?q=" + encodeURIComponent(query) + "&workspace=" + encodeURIComponent(WORKSPACE) + "&limit=80");
        if (!res.ok) throw new Error("Search failed");
        var data = await res.json();
        renderSearchResults(data);
    } catch (err) {
        results.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Search unavailable</div><div class="row-meta">' + escapeHTML(err.message) + '</div></div></div>';
    }
}

function renderSearchResults(data) {
    var results = document.getElementById("search-results");
    var title = document.getElementById("search-results-title");
    if (title) {
        title.textContent = formatNumber((data.results || []).length) + " results for “" + data.query + "”";
    }
    if (!data.results || data.results.length === 0) {
        results.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No matching objects</div><div class="row-meta">Try a symbol, package, repository or filename.</div></div></div>';
        return;
    }
    results.innerHTML = "";
    data.results.forEach(function(item) {
        var row = document.createElement("div");
        row.className = "list-row search-result";
        row.innerHTML = '<div class="row-icon">' + symbolIcon(item.kind) + '</div>' +
            '<div class="row-main">' +
                '<div class="row-title">' + escapeHTML(item.name) + '</div>' +
                '<div class="row-meta">' + escapeHTML(item.repo || "") + " · " + escapeHTML(item.package || "") + " · " + escapeHTML(item.file || "") + '</div>' +
            '</div>' +
            '<span class="kind-pill">' + escapeHTML(item.kind || "entity") + '</span>';
        row.onclick = function() {
            openSearchResult(item);
        };
        results.appendChild(row);
    });
}

function openSearchResult(item) {
    if (!item) return;
    var drawer = document.getElementById("drawer");
    var overlay = document.getElementById("drawer-overlay");
    var body = document.getElementById("drawer-body");

    setText("drawer-title", item.name || "Symbol");
    setText("drawer-kind", (item.kind || "entity").toUpperCase());

    var html = '<div class="detail-section"><div class="detail-section-title">Identity</div>';
    html += propertyRow("Kind", item.kind || "—");
    html += propertyRow("Repository", item.repo || "—");
    html += propertyRow("Package", item.package || "—");
    html += propertyRow("File", item.file || "—");
    html += propertyRow("Exported", item.exported ? "Yes" : "No");
    html += '</div>';

    html += '<div class="detail-section"><div class="detail-section-title">Actions</div>';
    html += '<button class="detail-action" onclick="closeDrawer(); state.graphCache = {}; state.currentLevel=\'entity\'; state.currentFocus=\'' + escapeJS(item.package || item.id) + '\'; showView(\'architecture\');">Explore local neighborhood →</button>';
    html += '</div>';

    body.innerHTML = html;
    drawer.classList.add("open");
    overlay.classList.add("open");
}

function setText(id, value) {
    var el = document.getElementById(id);
    if (!el) {
        console.warn("setText: missing element #" + id);
        return;
    }
    el.textContent = value;
}

function formatNumber(value) {
    return Number(value || 0).toLocaleString();
}

function formatDate(value) {
    if (!value) return "—";
    try { return new Date(value).toLocaleString(); } catch (e) { return value; }
}

function shortenLabel(value, max) {
    if (!value) return "";
    if (value.length <= max) return value;
    return value.slice(0, max - 1) + "…";
}

function symbolIcon(kind) {
    if (!kind) return "◇";
    if (kind === "function") return "ƒ";
    if (kind === "method") return "m";
    if (kind === "struct") return "S";
    if (kind === "interface") return "I";
    if (kind === "package") return "◇";
    if (kind === "file") return "□";
    return "•";
}

function escapeHTML(value) {
    return String(value)
        .replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#039;");
}

// escapeJS — broader than the previous version. Escapes backslash,
// both quote styles, newlines, carriage returns, and the Unicode line
// and paragraph separators that terminate JS source. Values embedded
// in inline handlers (onclick="f('...')") cannot break out.
function escapeJS(value) {
    return String(value)
        .replace(/\\/g, "\\\\")
        .replace(/'/g, "\\'")
        .replace(/"/g, '\\"')
        .replace(/\n/g, "\\n")
        .replace(/\r/g, "\\r")
        .replace(/\u2028/g, "\\u2028")
        .replace(/\u2029/g, "\\u2029");
}

async function loadAll() {
    await loadStats();
    if (state.currentView === "architecture") {
        await loadArchitecture(state.currentLevel, state.currentFocus);
    }
    if (state.currentView === "agents") {
        agentsLoaded = false;
        await loadAgents();
    }
    if (state.currentView === "decisions") {
        decisionsLoaded = false;
        await loadDecisions();
    }
}

setupSearch();
loadWorkspacePicker();
showView(initialView());
loadAll();

window.addEventListener("resize", function() {
    if (state.currentView === "architecture") {
        setTimeout(fitGraph, 100);
    }
});

// ─────────────────────────────────────────────────────────────────────────────
// Agents tab — MCP session activity
// ─────────────────────────────────────────────────────────────────────────────

var agentsLoaded = false;

async function loadAgents() {
    // Idempotent: only fetch on the first activation. Manual refresh
    // through the top-bar Refresh button calls loadAll which can also
    // re-trigger this.
    if (agentsLoaded) return;
    agentsLoaded = true;

    var activeEl = document.getElementById("agents-active-list");
    var recentEl = document.getElementById("agents-recent-list");
    if (!activeEl || !recentEl) return;

    try {
        var res = await fetch("/api/v1/dashboard/sessions?workspace=" + encodeURIComponent(WORKSPACE), {
            headers: { "Accept": "application/json" }
        });
        if (!res.ok) throw new Error("HTTP " + res.status);
        var data = await res.json();
        renderSessionList(activeEl, data.active || [], "No active sessions.", true);
        renderSessionList(recentEl, data.recent || [], "No sessions in the last 24 hours.", false);
    } catch (err) {
        activeEl.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Failed to load sessions</div><div class="row-meta">' + escapeHTML(String(err)) + '</div></div></div>';
        recentEl.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Failed to load sessions</div></div></div>';
    }
}

function renderSessionList(container, sessions, emptyMessage, isActivePanel) {
    if (!sessions || sessions.length === 0) {
        container.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">' + escapeHTML(emptyMessage) + '</div></div></div>';
        return;
    }
    container.innerHTML = "";
    sessions.forEach(function(s) {
        var row = document.createElement("div");
        row.className = "list-row";
        var clientLabel = escapeHTML(s.client_name || "unknown");
        if (s.client_version) {
            clientLabel += " <span style=\"color:var(--muted); font-size:10px;\">" + escapeHTML(s.client_version) + "</span>";
        }
        var duration = formatDuration(s.duration_seconds);
        var activity = formatDate(s.last_activity_at);
        var badge = "";
        if (s.close_reason) {
            badge = '<span class="badge-pill info">' + escapeHTML(s.close_reason) + '</span>';
        } else if (isActivePanel) {
            badge = '<span class="badge-pill success">active</span>';
        }
        row.innerHTML =
            '<div class="row-main">' +
                '<div class="row-title">' + clientLabel + ' · ' + escapeHTML(s.agent_id || "mcp-agent") + '</div>' +
                '<div class="row-meta">' + s.tool_call_count + ' tool calls · ' + duration + ' · last activity ' + escapeHTML(activity) + '</div>' +
            '</div>' +
            badge;
        container.appendChild(row);
    });
}

function formatDuration(seconds) {
    if (!seconds || seconds < 1) return "<1s";
    if (seconds < 60) return seconds + "s";
    var m = Math.floor(seconds / 60);
    var s = seconds % 60;
    if (m < 60) return m + "m" + (s > 0 ? " " + s + "s" : "");
    var h = Math.floor(m / 60);
    return h + "h " + (m % 60) + "m";
}


// ─────────────────────────────────────────────────────────────────────────────
// Decisions tab — decision log, lineage, revision chain
// ─────────────────────────────────────────────────────────────────────────────

var decisionsLoaded = false;

async function loadDecisions() {
    if (decisionsLoaded) return;
    decisionsLoaded = true;

    var listEl = document.getElementById("decisions-list");
    var subtitleEl = document.getElementById("decisions-subtitle");
    if (!listEl) return;

    try {
        var res = await fetch("/api/v1/dashboard/decisions?workspace=" + encodeURIComponent(WORKSPACE), {
            headers: { "Accept": "application/json" }
        });
        if (!res.ok) throw new Error("HTTP " + res.status);
        var data = await res.json();
        var decisions = data.decisions || [];

        if (subtitleEl) {
            subtitleEl.textContent = decisions.length === 1
                ? "1 decision in this workspace."
                : decisions.length + " decisions in this workspace.";
        }
        renderDecisionList(listEl, decisions);
    } catch (err) {
        if (subtitleEl) subtitleEl.textContent = "Failed to load.";
        listEl.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Failed to load decisions</div><div class="row-meta">' + escapeHTML(String(err)) + '</div></div></div>';
    }
}

function renderDecisionList(container, decisions) {
    if (!decisions || decisions.length === 0) {
        container.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No decisions recorded.</div><div class="row-meta">Propose one with <code>garuda propose_decision</code> or via the MCP tool.</div></div></div>';
        return;
    }
    container.innerHTML = "";
    decisions.forEach(function(d) {
        var row = document.createElement("div");
        row.className = "list-row";
        row.style.cursor = "pointer";
        row.onclick = function() { openDecisionDetail(d.id); };

        var statusClass = decisionStatusClass(d.status);
        var statusBadge = '<span class="badge-pill ' + statusClass + '">' + escapeHTML(d.status || "unknown") + '</span>';
        var anchorBadge = d.anchored
            ? '<span class="badge-pill success" title="Merkle-anchored">anchored</span>'
            : '<span class="badge-pill info" title="Not anchored">unanchored</span>';

        var meta = [];
        if (d.scope_domain) meta.push(escapeHTML(d.scope_domain));
        if (d.scope_system) meta.push(escapeHTML(d.scope_system));
        meta.push(d.revision_count + (d.revision_count === 1 ? " revision" : " revisions"));
        if (d.owner) meta.push(escapeHTML(d.owner));

        row.innerHTML =
            '<div class="row-main">' +
                '<div class="row-title">' + escapeHTML(d.title || "(untitled)") + '</div>' +
                '<div class="row-meta">' + meta.join(' · ') + '</div>' +
            '</div>' +
            statusBadge + anchorBadge;
        container.appendChild(row);
    });
}

function decisionStatusClass(status) {
    var s = (status || "").toUpperCase();
    if (s === "APPROVED" || s === "CANONICAL") return "success";
    if (s === "DRAFT") return "info";
    if (s === "PENDING" || s === "REVIEW") return "warning";
    if (s === "QUARANTINED" || s === "BLOCKED") return "critical";
    return "info";
}

async function openDecisionDetail(id) {
    var drawer = document.getElementById("drawer");
    var overlay = document.getElementById("drawer-overlay");
    var body = document.getElementById("drawer-body");
    if (!drawer || !overlay || !body) return;

    body.innerHTML = '<div class="detail-section"><div class="detail-section-title">Loading decision...</div></div>';
    drawer.classList.add("open");
    overlay.classList.add("open");

    try {
        var res = await fetch("/api/v1/dashboard/decisions/" + encodeURIComponent(id) + "?workspace=" + encodeURIComponent(WORKSPACE), {
            headers: { "Accept": "application/json" }
        });
        if (!res.ok) throw new Error("HTTP " + res.status);
        var data = await res.json();
        renderDecisionDetail(body, data);
    } catch (err) {
        body.innerHTML = '<div class="detail-section"><div class="detail-section-title">Failed to load decision</div><div class="detail-value">' + escapeHTML(String(err)) + '</div></div>';
    }
}

function renderDecisionDetail(body, data) {
    var d = data.decision || {};
    var html = '';

    html += '<div class="detail-section"><div class="detail-section-title">Decision</div>';
    html += propertyRow("Title", d.title || "(untitled)");
    html += propertyRow("Status", d.status || "unknown");
    html += propertyRow("Owner", d.owner || "—");
    html += propertyRow("Scope domain", d.scope_domain || "—");
    html += propertyRow("Scope system", d.scope_system || "—");
    html += propertyRow("Created", d.created_at || "—");
    html += propertyRow("Updated", d.updated_at || "—");
    html += propertyRow("Merkle anchor", d.anchored ? "anchored" : "not anchored");
    html += '</div>';

    var ancestors = data.ancestors || [];
    if (ancestors.length > 1) {
        html += '<div class="detail-section"><div class="detail-section-title">Ancestors (' + (ancestors.length - 1) + ')</div>';
        for (var i = 1; i < ancestors.length; i++) {
            html += propertyRow("→ " + (i), ancestors[i].title || ancestors[i].id);
        }
        html += '</div>';
    }

    var children = data.children || [];
    if (children.length > 0) {
        html += '<div class="detail-section"><div class="detail-section-title">Superseded by (' + children.length + ')</div>';
        children.forEach(function(c) {
            html += propertyRow("→", c.title || c.id);
        });
        html += '</div>';
    }

    var revisions = data.revisions || [];
    if (revisions.length > 0) {
        html += '<div class="detail-section"><div class="detail-section-title">Revision chain (' + revisions.length + ')</div>';
        revisions.forEach(function(r) {
            var hashShort = (r.decision_hash_hex || "").slice(0, 12);
            var prevShort = (r.previous_hash_hex || "").slice(0, 12);
            var isGenesis = (r.previous_hash_hex || "").replace(/0/g, "") === "";
            html += '<div class="detail-property">';
            html += '<div class="detail-key">rev ' + r.revision_number + '</div>';
            html += '<div class="detail-value">';
            html += escapeHTML(hashShort) + '…';
            if (isGenesis) {
                html += ' <span class="badge-pill info">genesis</span>';
            }
            html += '<div class="row-meta" style="margin-top:4px;">prev ' + escapeHTML(prevShort) + '… · ' + escapeHTML(r.created_at || "") + '</div>';
            html += '</div></div>';
        });
        html += '</div>';
    } else {
        html += '<div class="detail-section"><div class="detail-section-title">Revision chain</div>';
        html += '<div class="row-meta">No revisions recorded for this decision.</div>';
        html += '</div>';
    }

    body.innerHTML = html;
}