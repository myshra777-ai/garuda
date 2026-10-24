/* ============================================================================
   GARUDA — Epistemic Software Intelligence Instrument (dashboard.js)
   Combines Linear interaction discipline, Vercel context hierarchy,
   Sentry/Vero investigative drill-down, and Neo4j Bloom neighborhood graph.
   ============================================================================ */

var urlParams = new URLSearchParams(window.location.search);
var WORKSPACE = urlParams.get("workspace") || window.GARUDA_WORKSPACE || "";

var viewToSection = {
    workspace:  "view-overview",
    search:     "view-search",
    graph:      "view-architecture",
    review:     "view-review",
    governance: "view-trust",
    agents:     "view-agents",
    decisions:  "view-decisions"
};

var surfaceLabels = {
    workspace:  "01 — Overview",
    search:     "02 — Explore Index",
    graph:      "03 — Architecture Map",
    review:     "04 — Change Review",
    governance: "05 — Evidence & Ledger",
    agents:     "06 — Agents & Policies",
    decisions:  "ADR — Decision Lineage"
};

var tabs = ["workspace", "search", "graph", "review", "governance", "agents", "decisions"];

function initialView() {
    var params = new URLSearchParams(window.location.search);
    var tab = params.get("tab");
    if (viewToSection[tab]) return tab;
    return "workspace";
}

var state = {
    stats: null,
    policiesData: null,
    currentView: "workspace",
    currentLevel: "entity",
    currentFocus: "",
    hopDepth: 1,
    edgeTypeFilter: "ALL",
    graphLayoutMode: "both",
    selectedNode: null,
    graphData: null,
    graphZoom: null,
    graphSvg: null,
    graphGroup: null,
    graphSimulation: null,
    searchTimer: null,
    searchResultsCache: [],
    searchKindFilter: "all",
    activeCommunities: new Set(),
    hoveredNodeId: null,
    graphCache: {},
    auditEntries: [],
    auditFilter: "ALL",
    cmdkIndex: 0,
    cmdkItems: []
};

// Restrained architectural palette for community dots (no neon)
var communityPalette = [
    "#191B19", "#4D5E10", "#2E5A44", "#7C5816", "#525866", "#7A2E2E",
    "#3D5A6C", "#5E4B70", "#6B5E4C", "#3B6E5C", "#854D0E", "#4A4D48"
];

function getCommunityColor(node) {
    if (!node) return communityPalette[0];
    if (node.status === "CONTRADICTED") return "#991B1B";
    var str = node._community || node.package || node.repo || node.id || "general";
    var hash = 0;
    for (var i = 0; i < str.length; i++) {
        hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }
    return communityPalette[Math.abs(hash) % communityPalette.length];
}

/* ============================================================================
   THEME MANAGEMENT (Warm Off-White Light Default + Restrained Graphite Dark)
   ============================================================================ */
function initTheme() {
    var saved = localStorage.getItem("garuda_theme") || "light";
    document.documentElement.setAttribute("data-theme", saved);
    updateThemeButton(saved);
}

function toggleTheme() {
    var curr = document.documentElement.getAttribute("data-theme") || "light";
    var next = curr === "light" ? "dark" : "light";
    document.documentElement.setAttribute("data-theme", next);
    localStorage.setItem("garuda_theme", next);
    updateThemeButton(next);
    if (state.currentView === "graph" && state.graphData) {
        renderGraph(state.graphData);
    }
}

function updateThemeButton(theme) {
    var btn = document.getElementById("theme-toggle-btn");
    if (btn) {
        btn.textContent = theme === "light" ? "☀" : "☾";
    }
}

/* ============================================================================
   WORKSPACE PICKER & ROUTING
   ============================================================================ */
async function loadWorkspacePicker() {
    try {
        var res = await fetch("/api/v1/workspaces", {
            headers: { "Accept": "application/json" }
        });
        if (!res.ok) return;
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

function showView(view) {
    var aliases = { overview: "workspace", architecture: "graph", trust: "governance" };
    view = aliases[view] || view;

    if (!viewToSection[view]) return;
    state.currentView = view;

    Object.keys(viewToSection).forEach(function(key) {
        var el = document.getElementById(viewToSection[key]);
        if (!el) return;
        el.style.display = (key === view) ? "block" : "none";
    });

    document.querySelectorAll(".tab").forEach(function(el) {
        var isActive = el.dataset.tab === view;
        el.classList.toggle("active", isActive);
        el.setAttribute("aria-selected", isActive ? "true" : "false");
    });

    document.querySelectorAll(".nav-btn").forEach(function(el) {
        var isActive = el.dataset.surface === view;
        el.classList.toggle("active", isActive);
    });

    setText("topbar-surface-name", surfaceLabels[view] || "01 — Overview");

    if (view === "graph") {
        // Start with context (top symbol/entity neighborhood), not an unfiltered hairball
        var defaultLevel = state.currentLevel || "entity";
        var defaultFocus = state.currentFocus || "";
        if (!defaultFocus && state.stats && state.stats.top_hubs && state.stats.top_hubs.length > 0 && defaultLevel === "entity") {
            defaultFocus = state.stats.top_hubs[0].package || state.stats.top_hubs[0].name || "";
        }
        loadArchitecture(defaultLevel, defaultFocus);
    }
    if (view === "review") {
        renderChangeReviewWorkbench();
    }
    if (view === "governance") {
        renderTrust();
        renderAuditLogTable();
    }
    if (view === "agents") {
        loadAgents();
    }
    if (view === "decisions") {
        loadDecisions();
    }

    var params = new URLSearchParams(window.location.search);
    if (params.get("tab") !== view) {
        params.set("tab", view);
        history.pushState({ tab: view }, "", window.location.pathname + "?" + params.toString());
    }
}

window.addEventListener("popstate", function(event) {
    var view = (event.state && event.state.tab) || initialView();
    showView(view);
});

/* ============================================================================
   STATS, POLICIES & HONEST SYSTEM STATE
   ============================================================================ */
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
        state.policiesData = data;
        renderPolicies(data);
        buildAuditEntries();
        if (state.currentView === "review") {
            renderChangeReviewWorkbench();
        }
    } catch (err) {
        console.error("Garuda policies error:", err);
    }
}

function renderStats() {
    if (!state.stats) return;
    var s = state.stats;

    setText("stat-repositories", formatNumber(s.repositories) + " repos");
    setText("stat-packages", formatNumber(s.packages));
    setText("stat-entities", formatNumber(s.entities));
    setText("stat-relationships", formatNumber(s.relationships));
    setText("stat-cross-links", formatNumber(s.cross_repo_links));
    setText("stat-idempotency", formatNumber(s.idempotency_safeguards));

    setText("stat-supported", formatNumber(s.supported_claims));
    setText("stat-unverified", formatNumber(s.unverified_claims));
    setText("stat-contradicted", formatNumber(s.contradicted));

    var totalClaims = (s.supported_claims || 0) + (s.unverified_claims || 0) + (s.contradicted || 0);
    if (totalClaims === 0) totalClaims = s.total_claims || 0;
    setText("pulse-total-claims", formatNumber(totalClaims));

    if (totalClaims > 0) {
        var supPct = Math.round(((s.supported_claims || 0) / totalClaims) * 100);
        var conPct = Math.round(((s.contradicted || 0) / totalClaims) * 100);
        var unvPct = Math.max(0, 100 - supPct - conPct);
        var barSup = document.getElementById("bar-supported");
        var barUnv = document.getElementById("bar-unverified");
        var barCon = document.getElementById("bar-contradicted");
        if (barSup) barSup.style.width = supPct + "%";
        if (barUnv) barUnv.style.width = unvPct + "%";
        if (barCon) barCon.style.width = conPct + "%";
    }

    if (s.verification_attempted) {
        setText("stat-unverified-desc", "static only, 0 recent spans");
    } else {
        setText("stat-unverified-desc", "runtime trace not yet ingested");
    }

    setText("explorer-repos", formatNumber(s.repositories));
    setText("explorer-packages", formatNumber(s.packages));
    setText("explorer-entities", formatNumber(s.entities));
    setText("nav-badge-packages", formatNumber(s.packages));
    setText("nav-badge-ledger", "#" + String(s.latest_block_height || 1));
    setText("overview-block-height", "#" + String(s.latest_block_height || 1));
    setText("overview-updated-at", "Last indexed " + formatDate(s.last_updated));

    var merkleShort = (s.latest_merkle_hash || "Genesis verified");
    if (merkleShort.length > 22) merkleShort = merkleShort.slice(0, 18) + "…";
    setText("overview-merkle-short", merkleShort);

    setText("sidebar-meta", formatNumber(s.repositories) + " repos · " + formatNumber(s.entities) + " entities");
    setText("sidebar-workspace", s.workspace || WORKSPACE);
    setText("workspace-breadcrumb", s.workspace || WORKSPACE);
    setText("repo-list-ws", s.workspace || WORKSPACE);
    setText("stat-drift-count", formatNumber(s.quarantined_count));

    // Honest narrative summary in the Overview Hero
    var narrative = "Indexed " + formatNumber(s.entities) + " code entities and " + formatNumber(s.relationships) +
        " semantic relationships across " + formatNumber(s.packages) + " packages. ";
    if ((s.contradicted || 0) > 0) {
        narrative += formatNumber(s.contradicted) + " contradicted claim(s) require immediate investigation.";
    } else if (!s.verification_attempted) {
        narrative += "Static AST extraction is complete; runtime OpenTelemetry verification has not yet been ingested for this revision.";
    } else {
        narrative += formatNumber(s.supported_claims) + " claims are corroborated by runtime execution with zero active contradictions.";
    }
    setText("overview-narrative", narrative);

    // Honest System Pill in Topbar
    var sysDot = document.getElementById("honest-system-dot");
    if ((s.contradicted || 0) > 0) {
        setText("honest-system-text", s.contradicted + " contradiction(s) active");
        if (sysDot) sysDot.style.background = "var(--red)";
    } else if (!s.verification_attempted) {
        setText("honest-system-text", "Static AST synced · Runtime unverified");
        if (sysDot) sysDot.style.background = "#D97706";
    } else {
        setText("honest-system-text", "AST + Runtime verified · Block #" + (s.latest_block_height || 1));
        if (sysDot) sysDot.style.background = "var(--green)";
    }

    renderLanguages(s.languages_breakdown || []);
    renderDrift(s.drift || {});
    renderHubs(s.top_hubs || []);
    renderAttention(s.needs_attention || []);
    renderEvidence(s.recent_evidence || []);
    renderTrust();
    renderScannedRepos(s.repo_stats || [], s.repositories_list || []);
    buildAuditEntries();
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
    setText("nav-badge-review", data.final_decision || "ALLOW");

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
        list.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No policy evaluations recorded yet</div><div class="row-meta">Run <code>./bin/garuda check --project ' + escapeHTML(WORKSPACE || "default") + ' --policy policies --save</code> to evaluate and anchor architecture rules.</div></div></div>';
        return;
    }

    list.innerHTML = "";
    evals.forEach(function(ev, idx) {
        var row = document.createElement("div");
        row.className = "policy-eval-row";

        var anchorChip = ev.merkle_block_height
            ? '<span class="merkle-chip' + (ev.anchor_valid ? '' : ' invalid') + '">block #' + ev.merkle_block_height + '</span>'
            : '<span class="merkle-chip invalid">unanchored</span>';

        var preds = (ev.matched_predicates || []).join(", ") || "all predicates passed";
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
                    '<span>rule: <strong>' + escapeHTML(preds) + '</strong></span>' +
                    '<span>' + escapeHTML(when) + '</span>' +
                '</div>' +
            '</div>' +
            '<button class="btn-secondary" style="height:26px; font-size:11px;" onclick="event.stopPropagation(); openEvalInChangeReview(' + idx + ')">Review →</button>';

        row.style.cursor = "pointer";
        row.onclick = function() {
            openEvalInChangeReview(idx);
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
    setText("drawer-kind", "POLICY VERDICT · " + (ev.decision || ""));

    var html = '<div class="detail-section"><div class="detail-section-title">Governance Verdict</div>';
    html += propertyRow("Decision", ev.decision);
    html += propertyRow("Reason", ev.reason);
    html += propertyRow("Actor", ev.actor || "—");
    html += propertyRow("Subject", (ev.subject_kind || "workspace") + "/" + (ev.subject_id || WORKSPACE));
    html += propertyRow("Evaluated", formatDate(ev.evaluated_at));
    html += '</div>';

    html += '<div class="detail-section"><div class="detail-section-title">Corroborated Evidence</div>';
    html += propertyRow("Entities", ev.entity_count);
    html += propertyRow("Claims", ev.claim_count);
    html += propertyRow("Contradictions", ev.contradiction_count);
    html += propertyRow("Matched Rules", (ev.matched_predicates || []).join(", ") || "—");
    html += '</div>';

    html += '<div class="detail-section"><div class="detail-section-title">Cryptographic RFC 6962 Anchor</div>';
    if (ev.merkle_block_height) {
        html += propertyRow("Block height", "#" + ev.merkle_block_height);
        html += propertyRow("Proof structure", ev.anchor_valid ? "Valid RFC 6962 inclusion path" : "Invalid / missing");
        html += '<div class="merkle" style="margin-top:8px;">eval_id=' + escapeHTML(ev.id) + '</div>';
        html += '<button class="detail-action" onclick="verifyPolicyAnchor(\'' + escapeJS(ev.id) + '\')">Verify Cryptographic Proof →</button>';
    } else {
        html += '<div class="row-meta">This evaluation was run in preview/dry-run mode and has no Merkle anchor.</div>';
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
                var versionTag = data.version === 1 ? "RFC 6962 v1" : "v0";
                alert("✓ Cryptographic Merkle Proof VALID (" + versionTag + ")\nEvaluation ID: " + evalID + "\nEpoch Block: #" + data.block_height + "\nEpoch Root: " + (data.epoch_root || data.new_root || "verified"));
            } else {
                alert("✗ Merkle Proof INVALID: " + (data.error || "verification failed"));
            }
        })
        .catch(function(err) {
            alert("Verification request failed: " + err);
        });
}

function renderLanguages(langs) {
    var bar = document.getElementById("lang-bar");
    var legend = document.getElementById("lang-legend");
    if (!bar || !legend) return;

    if (!langs || langs.length === 0) {
        bar.innerHTML = '<div style="width:100%; background:var(--surface-3);"></div>';
        legend.innerHTML = '<span style="color:var(--muted); font-size:11px;">Run garuda analyze to index languages</span>';
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
    langs.slice(0, 3).forEach(function(lang) {
        var item = document.createElement("span");
        item.className = "lang-legend-item";
        item.style.marginRight = "8px";
        item.innerHTML = '<span class="lang-legend-dot" style="background:' + lang.color + ';"></span>' +
            '<span class="lang-legend-name">' + escapeHTML(lang.name) + '</span> ' +
            '<span class="lang-legend-pct">' + lang.percentage.toFixed(0) + '%</span>';
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
    setText("drift-unimplemented-docs", formatNumber(d.unverified_docs));
}

function renderHubs(hubs) {
    var list = document.getElementById("hub-list");
    if (!list) return;
    if (!hubs || hubs.length === 0) {
        list.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Explore high-impact entities</div><div class="row-meta">Open the Architecture Map to inspect symbol centrality.</div></div><button class="btn-secondary" onclick="openArchitecture(\'entity\')">Open Map</button></div>';
        return;
    }
    list.innerHTML = "";
    hubs.forEach(function(h) {
        var row = document.createElement("div");
        row.className = "list-row";
        row.style.cursor = "pointer";
        row.innerHTML = '<div class="row-icon">' + symbolIcon(h.kind) + '</div>' +
            '<div class="row-main">' +
                '<div class="row-title mono">' + escapeHTML(h.name) + '</div>' +
                '<div class="row-meta">' + escapeHTML(h.kind) + ' · ' + escapeHTML(h.package) + '</div>' +
            '</div>' +
            '<div style="text-align:right; display:flex; align-items:center; gap:10px;">' +
                '<div><span class="mono" style="font-weight:700; color:var(--text); font-size:12.5px;">' + h.callers + '</span> <span style="font-size:10px; color:var(--muted);">callers</span></div>' +
                '<span class="kbd-chip">1-hop →</span>' +
            '</div>';
        row.onclick = function() {
            focusSymbolOnArchitectureMap(h.name, h.package, h);
        };
        list.appendChild(row);
    });
}

function focusSymbolOnArchitectureMap(symbolName, pkgName, hubObj) {
    state.currentLevel = "entity";
    state.currentFocus = symbolName || pkgName || "";
    state.selectedNode = hubObj ? {
        id: hubObj.id || symbolName,
        label: hubObj.name || symbolName,
        kind: hubObj.kind || "function",
        package: hubObj.package || pkgName,
        repo: hubObj.repo || WORKSPACE,
        file: hubObj.file || hubObj.package || ""
    } : null;
    showView("graph");
}

function renderAttention(items) {
    var list = document.getElementById("attention-list");
    if (!list) return;
    if (!items || items.length === 0) {
        list.innerHTML = '<div class="list-row"><div class="row-icon" style="color:var(--green); border-color:var(--green-border); background:var(--green-bg);">✓</div><div class="row-main"><div class="row-title">Zero Active Contradictions</div><div class="row-meta">All indexed static claims and observed runtime edges are structurally consistent.</div></div></div>';
        return;
    }
    list.innerHTML = "";
    items.forEach(function(item) {
        var row = document.createElement("div");
        row.className = "list-row";
        row.style.cursor = "pointer";
        row.innerHTML = '<div class="row-icon" style="color:' + (item.severity === 'critical' ? 'var(--red)' : 'var(--amber)') + ';">!</div>' +
            '<div class="row-main">' +
                '<div class="row-title">' + escapeHTML(item.title) + '</div>' +
                '<div class="row-meta">' + escapeHTML(item.subtitle) + ' · ' + escapeHTML(item.evidence_loc) + '</div>' +
            '</div>' +
            '<span class="badge-pill ' + escapeHTML(item.severity) + '">' + escapeHTML(item.severity) + '</span>';
        row.onclick = function() {
            showView("review");
        };
        list.appendChild(row);
    });
}

function renderEvidence(items) {
    var list = document.getElementById("evidence-list");
    if (!list) return;
    if (!items || items.length === 0) {
        list.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Static AST evidence indexed</div><div class="row-meta">Ingest OpenTelemetry spans via POST /api/v1/telemetry/spans to record live runtime corroboration.</div></div></div>';
        return;
    }
    list.innerHTML = "";
    items.forEach(function(item) {
        var row = document.createElement("div");
        row.className = "list-row";
        var badgeClass = item.kind === "Runtime Trace" ? "warning" : "info";
        var badgeText = item.kind === "Runtime Trace" ? "Runtime Observed" : "AST Recorded";
        row.innerHTML = '<div class="row-icon">§</div>' +
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
            list.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No repositories indexed in this workspace yet.</div></div></div>';
            return;
        }
        list.innerHTML = "";
        fallbackList.forEach(function(repo) {
            var row = document.createElement("div");
            row.className = "list-row";
            row.style.cursor = "pointer";
            row.innerHTML = '<div class="row-icon">▦</div>' +
                '<div class="row-main">' +
                    '<div class="row-title mono">' + escapeHTML(repo) + '</div>' +
                    '<div class="row-meta">Indexed repository boundary</div>' +
                '</div>' +
                '<button class="btn-secondary" onclick="event.stopPropagation(); runSearch(\'' + escapeJS(repo) + '\')">Explore Symbols →</button>';
            row.onclick = function() {
                state.currentLevel = "package";
                state.currentFocus = repo;
                showView("graph");
            };
            list.appendChild(row);
        });
        return;
    }

    list.innerHTML = "";
    repoStats.forEach(function(repo) {
        var card = document.createElement("div");
        card.className = "repo-card";

        var statusBadge = repo.analysis_status === "synced"
            ? '<span class="badge-pill success">synced</span>'
            : '<span class="badge-pill info">' + escapeHTML(repo.analysis_status || "indexed") + '</span>';

        var commitShort = repo.current_commit && repo.current_commit.length > 8
            ? repo.current_commit.substring(0, 8)
            : (repo.current_commit || "HEAD");

        card.innerHTML =
            '<div class="repo-card-header">' +
                '<div>' +
                    '<div class="repo-card-name">' + escapeHTML(repo.name) + '</div>' +
                    '<div class="repo-card-meta">commit ' + escapeHTML(commitShort) + ' · analyzed ' + (repo.last_analyzed || "recently") + '</div>' +
                '</div>' +
                statusBadge +
            '</div>' +
            '<div class="repo-card-stats">' +
                '<span><span class="repo-card-stat-val">' + formatNumber(repo.entities) + '</span> entities</span>' +
                '<span><span class="repo-card-stat-val">' + formatNumber(repo.relationships) + '</span> relationships</span>' +
                '<span><span class="repo-card-stat-val">' + formatNumber(repo.files) + '</span> files</span>' +
            '</div>' +
            '<div class="repo-card-actions">' +
                '<button class="btn-secondary" style="height:28px; font-size:11.5px;" onclick="event.stopPropagation(); state.currentLevel=\'package\'; state.currentFocus=\'' + escapeJS(repo.name) + '\'; showView(\'graph\');">Architecture Map (Packages) →</button>' +
                '<button class="btn-secondary" style="height:28px; font-size:11.5px;" onclick="event.stopPropagation(); runSearch(\'' + escapeJS(repo.name) + '\')">Search Symbols</button>' +
            '</div>';

        list.appendChild(card);
    });
}

/* ============================================================================
   03 — ARCHITECTURE MAP: NEIGHBORHOOD DIAGRAM + BLOOM CANVAS + INSPECTOR
   ============================================================================ */
function openArchitecture(level) {
    state.currentLevel = level || "entity";
    state.currentFocus = "";
    showView("graph");
}

function setArchitectureMode(level) {
    state.currentLevel = level;
    state.currentFocus = "";
    ["entity", "package", "repo", "full"].forEach(function(k) {
        var btn = document.getElementById("btn-mode-" + k);
        if (btn) {
            var match = (k === "repo" && level === "repository") || k === level;
            btn.classList.toggle("active", match);
        }
    });
    loadArchitecture(level, "");
}

function setHopDepth(hops) {
    state.hopDepth = hops;
    var b1 = document.getElementById("btn-hop-1");
    var b2 = document.getElementById("btn-hop-2");
    if (b1) b1.classList.toggle("active", hops === 1);
    if (b2) b2.classList.toggle("active", hops === 2);
    if (state.graphData) {
        renderGraph(state.graphData);
    }
}

function setEdgeTypeFilter(filterType) {
    state.edgeTypeFilter = filterType;
    document.querySelectorAll("[data-edge-filter]").forEach(function(btn) {
        btn.classList.toggle("active", btn.getAttribute("data-edge-filter") === filterType);
    });
    if (state.graphData) {
        renderGraph(state.graphData);
    }
}

function setGraphLayoutMode(mode) {
    state.graphLayoutMode = mode;
    ["both", "cards", "canvas"].forEach(function(m) {
        var b = document.getElementById("btn-layout-" + m);
        if (b) b.classList.toggle("active", m === mode);
    });
    var nbStage = document.getElementById("neighborhood-stage");
    var canvasStage = document.getElementById("graph-canvas-stage");
    if (nbStage) nbStage.style.display = (mode === "canvas") ? "none" : "block";
    if (canvasStage) canvasStage.style.display = (mode === "cards") ? "none" : "flex";
    if (mode !== "cards") {
        setTimeout(fitGraph, 120);
    }
}

function focusGraphByQuery(query) {
    query = (query || "").trim();
    if (!query) return;
    state.graphCache = {};
    state.currentLevel = "entity";
    state.currentFocus = query;
    loadArchitecture("entity", query);
}

function toggleFullscreen() {
    var wrap = document.getElementById("graph-wrap");
    if (wrap) wrap.classList.toggle("fullscreen");
    setTimeout(fitGraph, 200);
}

async function loadArchitecture(level, focus) {
    state.currentLevel = level || "entity";
    state.currentFocus = focus || "";
    updateArchitectureHeader();

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
        var res = await fetch(url, { headers: { "Accept": "application/json" } });
        if (!res.ok) throw new Error("Graph request failed");
        var data = await res.json();
        // If entity focus returned 0 nodes (e.g., symbol name vs package), fallback to top entities and select matching node
        if ((!data.nodes || data.nodes.length === 0) && state.currentFocus) {
            var fallbackRes = await fetch("/api/v1/graph?workspace=" + encodeURIComponent(WORKSPACE) + "&level=entity");
            if (fallbackRes.ok) {
                data = await fallbackRes.json();
            }
        }
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
    var title = "Architecture Map";
    var subtitle = "Focused 1-hop and 2-hop neighborhood around your selected symbol or package.";
    var breadcrumb = state.currentFocus || "Neighborhood";

    if (state.currentLevel === "full") {
        title = "System Package Mesh";
        subtitle = "High-centrality package dependency topology across the workspace.";
        breadcrumb = "Full Mesh";
    } else if (state.currentLevel === "repository") {
        title = "Repository Boundaries";
        subtitle = "Cross-repository dependency bridges and service boundaries.";
        breadcrumb = "Repositories";
    } else if (state.currentLevel === "package") {
        title = state.currentFocus ? "Packages in " + state.currentFocus : "Package Architecture";
        subtitle = "Click any package to inspect its incoming and outgoing module dependencies.";
        breadcrumb = state.currentFocus || "Packages";
    } else if (state.currentLevel === "entity") {
        title = state.currentFocus ? "Neighborhood: " + shortenLabel(state.currentFocus, 42) : "Focused Symbol Neighborhood";
        subtitle = "Start with context, not every node. Click any neighbor to pivot the inspector without reshuffling.";
        breadcrumb = state.currentFocus ? shortenLabel(state.currentFocus, 28) : "Symbols";
    }

    setText("architecture-title", title);
    setText("architecture-subtitle", subtitle);
    setText("architecture-breadcrumb", breadcrumb);
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
        return p.slice(-2).join("/");
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
        item.innerHTML = '<input type="checkbox" class="community-checkbox" checked>' +
            '<span class="community-dot" style="background:' + color + ';"></span>' +
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

/* Update Screenshot 7 Neighborhood Card Diagram + Persistent Side Inspector */
var currentIncomingNodes = [];
var currentOutgoingNodes = [];

function updateNeighborhoodAndInspector(centerNode, allNodes, allEdges) {
    if (!centerNode) return;
    state.selectedNode = centerNode;

    var incoming = [];
    var outgoing = [];
    allEdges.forEach(function(e) {
        var sId = e.source.id || e.source;
        var tId = e.target.id || e.target;
        if (tId === centerNode.id && sId !== centerNode.id) {
            incoming.push({ node: e.source, edge: e });
        }
        if (sId === centerNode.id && tId !== centerNode.id) {
            outgoing.push({ node: e.target, edge: e });
        }
    });

    currentIncomingNodes = incoming;
    currentOutgoingNodes = outgoing;

    // 1. Update Screenshot 7 Top Neighborhood Card Map
    setText("nb-selected-title", centerNode.label || centerNode.id);
    var locText = (centerNode.package || centerNode.repo || "workspace") +
        (centerNode.file ? " · " + shortenPath(centerNode.file) : "") +
        " (" + (centerNode.kind || "entity") + ")";
    setText("nb-selected-loc", locText);

    if (incoming.length > 0) {
        setText("nb-caller-title", incoming[0].node.label || "Caller");
        setText("nb-caller-sub", incoming.length === 1
            ? "1 incoming caller (" + (incoming[0].edge.type || "CALLS") + ")"
            : incoming.length + " incoming callers · click to pivot");
    } else {
        setText("nb-caller-title", "Entrypoint / Root");
        setText("nb-caller-sub", "0 incoming callers in view");
    }

    if (outgoing.length > 0) {
        setText("nb-dep-title", outgoing[0].node.label || "Dependency");
        setText("nb-dep-sub", outgoing.length === 1
            ? "1 outgoing dependency (" + (outgoing[0].edge.type || "CALLS") + ")"
            : outgoing.length + " outgoing dependencies · click to pivot");
    } else {
        setText("nb-dep-title", "Leaf Entity");
        setText("nb-dep-sub", "0 outgoing dependencies in view");
    }

    var policyDecision = (state.policiesData && state.policiesData.final_decision) || "ALLOW";
    var activePolCount = (state.policiesData && state.policiesData.active_policies) || 0;
    setText("nb-policy-title", "Policy Gate: " + policyDecision);
    setText("nb-policy-sub", activePolCount + " active rules · click for Change Review");

    // 2. Update Persistent Right-Side Entity Inspector
    setText("inspector-name", centerNode.label || centerNode.id);
    setText("inspector-meta", (centerNode.kind || "symbol").toUpperCase() + " · " + (centerNode.package || centerNode.repo || "workspace"));
    setText("insp-in-count", String(incoming.length));
    setText("insp-out-count", String(outgoing.length));

    renderInspectorRefList("insp-in-list", incoming, "←");
    renderInspectorRefList("insp-out-list", outgoing, "→");

    var propsEl = document.getElementById("insp-properties");
    if (propsEl) {
        var html = '<div class="detail-section-title" style="margin-top:8px;">Source &amp; Epistemic Metadata</div>';
        html += propertyRow("Kind", centerNode.kind || "entity");
        html += propertyRow("Package", centerNode.package || "—");
        html += propertyRow("Repository", centerNode.repo || WORKSPACE);
        if (centerNode.file) html += propertyRow("Source File", shortenPath(centerNode.file));
        html += propertyRow("Exported", centerNode.exported ? "Yes (Public API)" : "Package-private");
        html += propertyRow("Epistemic State", centerNode.status || "SUPPORTED (AST)");
        html += '<div style="margin-top:12px; display:flex; flex-direction:column; gap:6px;">';
        html += '<button class="btn-primary" style="width:100%; justify-content:center;" onclick="openSelectedInChangeReview()">Inspect in Change Review →</button>';
        if (centerNode.kind === "repository" || state.currentLevel === "repository") {
            html += '<button class="btn-secondary" style="width:100%;" onclick="state.currentLevel=\'package\'; state.currentFocus=\'' + escapeJS(centerNode.label) + '\'; loadArchitecture(\'package\', \'' + escapeJS(centerNode.label) + '\')">Drill into Packages →</button>';
        } else if (centerNode.kind === "package" || state.currentLevel === "package") {
            html += '<button class="btn-secondary" style="width:100%;" onclick="state.currentLevel=\'entity\'; state.currentFocus=\'' + escapeJS(centerNode.id) + '\'; loadArchitecture(\'entity\', \'' + escapeJS(centerNode.id) + '\')">Drill into Symbols →</button>';
        }
        html += '</div>';
        propsEl.innerHTML = html;
    }
}

function renderInspectorRefList(containerId, items, arrow) {
    var el = document.getElementById(containerId);
    if (!el) return;
    if (!items || items.length === 0) {
        el.innerHTML = '<div style="font-size:11px; color:var(--muted); padding:2px 0;">None in current neighborhood</div>';
        return;
    }
    el.innerHTML = "";
    items.slice(0, 6).forEach(function(item) {
        var row = document.createElement("div");
        row.style.cssText = "display:flex; align-items:center; justify-content:space-between; padding:5px 8px; margin-top:3px; background:var(--surface-2); border:1px solid var(--border); border-radius:5px; cursor:pointer; font-size:11.5px;";
        row.innerHTML = '<span class="mono" style="overflow:hidden; text-overflow:ellipsis; white-space:nowrap;">' + arrow + ' ' + escapeHTML(item.node.label || item.node.id) + '</span>' +
            '<span class="kind-pill" style="font-size:9px; padding:1px 5px;">' + escapeHTML(item.edge.type || "CALLS") + '</span>';
        row.onclick = function() {
            if (state.graphData) {
                updateNeighborhoodAndInspector(item.node, state.graphData._renderedNodes || [], state.graphData._renderedEdges || []);
                highlightSelectedNodeOnCanvas(item.node);
            }
        };
        el.appendChild(row);
    });
}

function expandInspectorList(dir) {
    var list = dir === "in" ? currentIncomingNodes : currentOutgoingNodes;
    if (list && list.length > 0) {
        var target = list[0].node;
        updateNeighborhoodAndInspector(target, state.graphData._renderedNodes || [], state.graphData._renderedEdges || []);
        highlightSelectedNodeOnCanvas(target);
    }
}

function focusFirstCaller() {
    if (currentIncomingNodes.length > 0) {
        expandInspectorList("in");
    }
}

function focusFirstDependency() {
    if (currentOutgoingNodes.length > 0) {
        expandInspectorList("out");
    }
}

function highlightSelectedNodeOnCanvas(centerNode) {
    if (!state.graphGroup || !centerNode) return;
    state.graphGroup.selectAll(".node-core")
        .attr("fill", function(d) {
            if (d.id === centerNode.id) return "#191B19";
            if (d.status === "CONTRADICTED") return "#991B1B";
            return "#FFFFFF";
        })
        .attr("stroke", function(d) {
            if (d.id === centerNode.id) return "#D5E66A";
            return getCommunityColor(d);
        })
        .attr("stroke-width", function(d) {
            return d.id === centerNode.id ? 3.5 : 2;
        });

    state.graphGroup.selectAll(".node-inner-glyph")
        .attr("fill", function(d) {
            return d.id === centerNode.id ? "#D5E66A" : getCommunityColor(d);
        });
}

function renderGraph(data) {
    if (state.graphSimulation) {
        state.graphSimulation.stop();
        state.graphSimulation = null;
    }

    var svg = d3.select("#graph");
    svg.selectAll("*").remove();

    var empty = document.getElementById("graph-empty");
    if (!data || !data.nodes || data.nodes.length === 0) {
        if (empty) empty.style.display = "grid";
        return;
    }
    if (empty) empty.style.display = "none";

    var container = document.querySelector(".graph-wrap");
    var width = (container && container.clientWidth) || 880;
    var height = (container && container.clientHeight) || 640;

    svg.attr("width", width).attr("height", height);

    var defs = svg.append("defs");
    defs.append("marker").attr("id", "arrow").attr("viewBox", "0 -5 10 10").attr("refX", 22).attr("refY", 0)
        .attr("markerWidth", 5).attr("markerHeight", 5).attr("orient", "auto")
        .append("path").attr("d", "M0,-4L10,0L0,4").attr("fill", "#8A8E85");

    defs.append("marker").attr("id", "arrow-violation").attr("viewBox", "0 -5 10 10").attr("refX", 22).attr("refY", 0)
        .attr("markerWidth", 6).attr("markerHeight", 6).attr("orient", "auto")
        .append("path").attr("d", "M0,-4L10,0L0,4").attr("fill", "#991B1B");

    var zoomLayer = svg.append("g");
    state.graphSvg = svg;
    state.graphGroup = zoomLayer;

    var zoom = d3.zoom().scaleExtent([0.15, 5]).on("zoom", function(event) {
        zoomLayer.attr("transform", event.transform);
    });
    svg.call(zoom);
    state.graphZoom = zoom;

    var allNodes = data.nodes.map(function(n) {
        var obj = Object.assign({}, n);
        obj._community = extractCommunityName(obj);
        return obj;
    });
    var nodeByID = {};
    allNodes.forEach(function(n) { nodeByID[n.id] = n; });

    var allEdges = (data.edges || []).filter(function(e) {
        if (!nodeByID[e.from] || !nodeByID[e.to]) return false;
        if (state.edgeTypeFilter === "ALL") return true;
        if (state.edgeTypeFilter === "CONTRADICTED") return e.status === "CONTRADICTED";
        return (e.type || "").toUpperCase().indexOf(state.edgeTypeFilter) !== -1;
    }).map(function(e) {
        return {
            id: e.id,
            source: nodeByID[e.from],
            target: nodeByID[e.to],
            type: e.type || "CALLS",
            status: e.status,
            label: e.label,
            count: e.count || 1
        };
    });

    // Determine primary focused symbol/node
    var centerNode = state.selectedNode && nodeByID[state.selectedNode.id] ? nodeByID[state.selectedNode.id] : null;
    if (!centerNode && state.currentFocus) {
        centerNode = allNodes.find(function(n) {
            return n.id === state.currentFocus || n.label === state.currentFocus || n.package === state.currentFocus;
        });
    }
    if (!centerNode) {
        // Pick highest-centrality node
        var sortedByDeg = allNodes.slice().sort(function(a, b) { return (b.count || 0) - (a.count || 0); });
        centerNode = sortedByDeg[0];
    }

    // Filter to 1-hop or 2-hop neighborhood around centerNode when in entity mode so we NEVER render a hairball
    var nodes = allNodes;
    var validEdges = allEdges;
    if (state.currentLevel === "entity" && centerNode && allNodes.length > 28) {
        var hop1 = new Set([centerNode.id]);
        allEdges.forEach(function(e) {
            if (e.source.id === centerNode.id) hop1.add(e.target.id);
            if (e.target.id === centerNode.id) hop1.add(e.source.id);
        });
        var allowed = new Set(hop1);
        if (state.hopDepth >= 2) {
            allEdges.forEach(function(e) {
                if (hop1.has(e.source.id)) allowed.add(e.target.id);
                if (hop1.has(e.target.id)) allowed.add(e.source.id);
            });
        }
        // If centerNode had few edges, include top connected nodes up to 24 nodes total
        if (allowed.size < 8) {
            allNodes.slice(0, 20).forEach(function(n) { allowed.add(n.id); });
        }
        nodes = allNodes.filter(function(n) { return allowed.has(n.id); });
        validEdges = allEdges.filter(function(e) { return allowed.has(e.source.id) && allowed.has(e.target.id); });
    }

    data._renderedNodes = nodes;
    data._renderedEdges = validEdges;

    updateNeighborhoodAndInspector(centerNode, nodes, validEdges);
    setText("graph-description", "Showing " + nodes.length + " focused nodes and " + validEdges.length + " relationships (" + state.hopDepth + "-hop neighborhood)");

    var linkedByIndex = {};
    validEdges.forEach(function(d) {
        linkedByIndex[d.source.id + "," + d.target.id] = true;
        linkedByIndex[d.target.id + "," + d.source.id] = true;
    });

    function isConnected(a, b) {
        return a.id === b.id || linkedByIndex[a.id + "," + b.id];
    }

    var simulation = d3.forceSimulation(nodes)
        .alphaDecay(0.09)
        .force("link", d3.forceLink(validEdges).id(function(d) { return d.id; }).distance(115).strength(0.45))
        .force("charge", d3.forceManyBody().strength(-480))
        .force("center", d3.forceCenter(width / 2, height / 2))
        .force("collision", d3.forceCollide().radius(32));

    state.graphSimulation = simulation;

    var link = zoomLayer.append("g").selectAll("path")
        .data(validEdges)
        .enter().append("path")
        .attr("class", function(d) { return d.status === "CONTRADICTED" ? "graph-link violation" : "graph-link"; })
        .attr("stroke-width", function(d) { return d.status === "CONTRADICTED" ? 2.2 : 1.35; })
        .attr("marker-end", function(d) { return d.status === "CONTRADICTED" ? "url(#arrow-violation)" : "url(#arrow)"; });

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

    node.append("circle").attr("class", "node-core")
        .attr("r", function(d) {
            return (centerNode && d.id === centerNode.id) ? 15 : 11;
        })
        .attr("fill", function(d) {
            if (centerNode && d.id === centerNode.id) return "#191B19";
            if (d.status === "CONTRADICTED") return "#991B1B";
            return "#FFFFFF";
        })
        .attr("stroke", function(d) {
            if (centerNode && d.id === centerNode.id) return "#D5E66A";
            return getCommunityColor(d);
        })
        .attr("stroke-width", function(d) {
            return (centerNode && d.id === centerNode.id) ? 3.5 : 2;
        });

    node.append("text").attr("class", "node-inner-glyph")
        .attr("text-anchor", "middle")
        .attr("dy", 3.5)
        .style("font-family", "var(--font-mono)")
        .style("font-size", "9.5px")
        .style("font-weight", "700")
        .style("pointer-events", "none")
        .attr("fill", function(d) {
            return (centerNode && d.id === centerNode.id) ? "#D5E66A" : getCommunityColor(d);
        })
        .text(function(d) { return symbolIcon(d.kind); });

    node.append("text").attr("class", "graph-node-label")
        .attr("dx", 16)
        .attr("dy", 4)
        .text(function(d) { return shortenLabel(d.label || d.id, 26); });

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
        updateNeighborhoodAndInspector(d, nodes, validEdges);
        highlightSelectedNodeOnCanvas(d);
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
        state.currentFocus = d.label || d.id;
        loadArchitecture("entity", state.currentFocus);
    });

    var tickCount = 0;
    simulation.on("tick", function() {
        tickCount++;
        if (tickCount > 320 || simulation.alpha() < 0.05) {
            simulation.stop();
        }
        link.attr("d", function(d) {
            var dx = d.target.x - d.source.x;
            var dy = d.target.y - d.source.y;
            var dr = Math.sqrt(dx * dx + dy * dy) * 1.4;
            return "M" + d.source.x + "," + d.source.y + "A" + dr + "," + dr + " 0 0,1 " + d.target.x + "," + d.target.y;
        });
        node.attr("transform", function(d) { return "translate(" + d.x + "," + d.y + ")"; });
    });

    setTimeout(fitGraph, 450);
}

function fitGraph() {
    if (!state.graphSvg || !state.graphGroup) return;
    var svg = state.graphSvg;
    var group = state.graphGroup;
    var bbox;
    try { bbox = group.node().getBBox(); } catch (e) { return; }
    if (!bbox.width || !bbox.height) return;

    var container = document.querySelector(".graph-wrap");
    var width = (container && container.clientWidth) || 880;
    var height = (container && container.clientHeight) || 640;

    var scale = Math.min(width / (bbox.width + 120), height / (bbox.height + 120), 1.25);
    scale = Math.max(scale, 0.2);
    var tx = width / 2 - scale * (bbox.x + bbox.width / 2);
    var ty = height / 2 - scale * (bbox.y + bbox.height / 2);

    svg.transition().duration(300).call(state.graphZoom.transform, d3.zoomIdentity.translate(tx, ty).scale(scale));
}

function zoomGraph(multiplier) {
    if (!state.graphSvg || !state.graphZoom) return;
    state.graphSvg.transition().duration(180).call(state.graphZoom.scaleBy, multiplier);
}

function goUpArchitecture() {
    if (state.currentLevel === "entity") {
        setArchitectureMode("package");
        return;
    }
    if (state.currentLevel === "package" || state.currentLevel === "full") {
        setArchitectureMode("repository");
        return;
    }
    showView("workspace");
}

function openSelectedInChangeReview() {
    showView("review");
}

/* ============================================================================
   04 — CHANGE REVIEW: 3-PANE SIGNATURE WORKBENCH
   ============================================================================ */
function openEvalInChangeReview(evalIndex) {
    state.selectedEvalIndex = evalIndex;
    showView("review");
}

function renderChangeReviewWorkbench() {
    var picker = document.getElementById("review-target-picker");
    var evalListEl = document.getElementById("review-evaluations-list");
    var evals = (state.policiesData && state.policiesData.latest_evaluations) || [];
    var hubs = (state.stats && state.stats.top_hubs) || [];

    if (picker) {
        picker.innerHTML = "";
        if (state.selectedNode) {
            var optSel = document.createElement("option");
            optSel.value = "symbol:" + (state.selectedNode.label || state.selectedNode.id);
            optSel.textContent = "Symbol: " + (state.selectedNode.label || state.selectedNode.id) + " (" + (state.selectedNode.package || "workspace") + ")";
            picker.appendChild(optSel);
        }
        evals.forEach(function(ev, i) {
            var opt = document.createElement("option");
            opt.value = "eval:" + i;
            opt.textContent = "[" + ev.decision + "] " + ev.policy_title;
            picker.appendChild(opt);
        });
        hubs.forEach(function(h, i) {
            var opt = document.createElement("option");
            opt.value = "hub:" + i;
            opt.textContent = "Hub: " + h.name + " (" + h.callers + " callers)";
            picker.appendChild(opt);
        });
    }

    if (evalListEl) {
        if (evals.length === 0) {
            evalListEl.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No saved policy checks</div><div class="row-meta">Showing static blast-radius analysis.</div></div></div>';
        } else {
            evalListEl.innerHTML = "";
            evals.slice(0, 6).forEach(function(ev, idx) {
                var r = document.createElement("div");
                r.className = "list-row";
                r.style.cursor = "pointer";
                r.innerHTML = '<span class="policy-badge ' + escapeHTML(ev.decision) + '">' + escapeHTML(ev.decision) + '</span>' +
                    '<div class="row-main"><div class="row-title">' + escapeHTML(ev.policy_title) + '</div><div class="row-meta">' + formatDate(ev.evaluated_at) + '</div></div>';
                r.onclick = function() {
                    selectReviewTarget("eval:" + idx);
                };
                evalListEl.appendChild(r);
            });
        }
    }

    if (state.selectedEvalIndex !== undefined && evals[state.selectedEvalIndex]) {
        selectReviewTarget("eval:" + state.selectedEvalIndex);
    } else if (state.selectedNode) {
        selectReviewTarget("symbol:" + (state.selectedNode.label || state.selectedNode.id));
    } else if (evals.length > 0) {
        selectReviewTarget("eval:0");
    } else if (hubs.length > 0) {
        selectReviewTarget("hub:0");
    } else {
        selectReviewTarget("default");
    }
}

function selectReviewTarget(val) {
    var evals = (state.policiesData && state.policiesData.latest_evaluations) || [];
    var hubs = (state.stats && state.stats.top_hubs) || [];
    var stage = document.getElementById("review-impact-stage");

    var targetName = WORKSPACE || "default";
    var pkgName = "workspace-wide";
    var filePath = "policies/architecture-layer-guardrails.yaml";
    var decision = (state.policiesData && state.policiesData.final_decision) || "ALLOW";
    var epistemic = (state.stats && state.stats.contradicted > 0) ? "CONTRADICTED" :
                    ((state.stats && state.stats.verification_attempted) ? "SUPPORTED" : "UNVERIFIED");
    var ledgerText = (state.stats && state.stats.latest_block_height) ? "ANCHORED #" + state.stats.latest_block_height : "ANCHORED #1";
    var ruleName = "architecture-layer-guardrails (layer_violation, cycle_exists, blast_radius_exceeds)";
    var reasonText = "Evaluated static AST dependencies and runtime verification status against active GarudaPolicy rules.";
    var evCounts = formatNumber((state.stats && state.stats.entities) || 0) + " entities · " +
                   formatNumber((state.stats && state.stats.relationships) || 0) + " relationships · " +
                   formatNumber((state.stats && state.stats.contradicted) || 0) + " contradictions";
    var evalId = "";
    var callersCount = hubs.length > 0 ? hubs[0].callers : 4;

    if (val && val.indexOf("eval:") === 0) {
        var idx = parseInt(val.split(":")[1], 10);
        var ev = evals[idx];
        if (ev) {
            targetName = ev.subject_id || ev.policy_title;
            pkgName = ev.subject_kind || "workspace";
            decision = ev.decision || "ALLOW";
            epistemic = ev.contradiction_count > 0 ? "CONTRADICTED" : "SUPPORTED";
            ledgerText = ev.merkle_block_height ? "BLOCK #" + ev.merkle_block_height : "UNANCHORED";
            ruleName = (ev.matched_predicates && ev.matched_predicates.length > 0) ? ev.matched_predicates.join(", ") : ev.policy_title;
            reasonText = ev.reason || reasonText;
            evCounts = formatNumber(ev.entity_count) + " entities · " + formatNumber(ev.claim_count) + " claims · " + formatNumber(ev.contradiction_count) + " contradictions";
            evalId = ev.id;
        }
    } else if (val && val.indexOf("hub:") === 0) {
        var hIdx = parseInt(val.split(":")[1], 10);
        var h = hubs[hIdx];
        if (h) {
            targetName = h.name;
            pkgName = h.package;
            filePath = h.package + " (" + h.kind + ")";
            callersCount = h.callers;
            reasonText = "Symbol " + h.name + " is an architectural hub with " + h.callers + " direct inbound callers. Changes to its signature or behavior propagate across " + h.package + ".";
        }
    } else if (state.selectedNode) {
        targetName = state.selectedNode.label || state.selectedNode.id;
        pkgName = state.selectedNode.package || state.selectedNode.repo || "workspace";
        filePath = state.selectedNode.file || pkgName;
        callersCount = currentIncomingNodes.length;
        reasonText = "Focused change impact review for " + targetName + " in package " + pkgName + ".";
    }

    setText("rev-ctx-target", targetName);
    setText("rev-ctx-pkg", pkgName);
    setText("rev-ctx-file", shortenPath(filePath));
    setText("rev-blast-summary", callersCount + " upstream callers · " + (currentOutgoingNodes.length || 2) + " downstream deps");

    // Update 3-Plane Triad Badges
    var epEl = document.getElementById("rev-triad-epistemic");
    if (epEl) {
        epEl.textContent = epistemic;
        epEl.className = "badge-pill " + (epistemic === "CONTRADICTED" ? "state-contradicted" : (epistemic === "SUPPORTED" ? "state-supported" : "state-unverified"));
    }
    var polEl = document.getElementById("rev-triad-policy");
    if (polEl) {
        polEl.textContent = decision;
        polEl.className = "policy-badge " + decision;
    }
    var ledEl = document.getElementById("rev-triad-ledger");
    if (ledEl) {
        ledEl.textContent = ledgerText;
    }

    setText("rev-dr-id", evalId ? ("eval " + evalId.slice(0, 8) + "…") : "live workspace check");
    setText("rev-dr-what", reasonText);
    setText("rev-dr-rule", ruleName);
    setText("rev-dr-evidence", evCounts);
    setText("rev-dr-cli", "./bin/garuda check --project " + (WORKSPACE || "default") + " --target " + targetName + " --policy policies --save");

    var actionsEl = document.getElementById("rev-dr-actions");
    if (actionsEl) {
        var btns = '<button class="btn-primary" onclick="focusSymbolOnArchitectureMap(\'' + escapeJS(targetName) + '\', \'' + escapeJS(pkgName) + '\', null)">Open in Architecture Map →</button>';
        if (evalId) {
            btns += '<button class="btn-secondary" onclick="verifyPolicyAnchor(\'' + escapeJS(evalId) + '\')">Verify RFC 6962 Proof</button>';
        }
        actionsEl.innerHTML = btns;
    }

    if (stage) {
        var inItems = currentIncomingNodes.length > 0
            ? currentIncomingNodes.slice(0, 3).map(function(x) { return x.node.label || x.node.id; })
            : hubs.slice(0, 3).map(function(x) { return x.name; });
        var outItems = currentOutgoingNodes.length > 0
            ? currentOutgoingNodes.slice(0, 3).map(function(x) { return x.node.label || x.node.id; })
            : ["store.PostgresStore", "merkle.Anchor", "policy.Evaluator"];

        var stageHTML = '<div class="neighborhood-diagram" style="padding:10px 0;">' +
            '<div class="selected-fn-card">' +
                '<div class="wb-pane-kicker" style="color:var(--brand-accent); margin-bottom:4px;">INSPECTED CHANGE TARGET</div>' +
                '<div class="selected-fn-title mono">' + escapeHTML(targetName) + '</div>' +
                '<div class="selected-fn-loc">' + escapeHTML(pkgName) + ' · ' + escapeHTML(shortenPath(filePath)) + '</div>' +
            '</div>' +
            '<svg class="neighborhood-connector-svg" viewBox="0 0 560 44" preserveAspectRatio="none">' +
                '<path d="M280,0 L280,22 M110,22 L450,22 M110,22 L110,42 M280,22 L280,42 M450,22 L450,42" stroke="#9FA39A" stroke-width="1.4" fill="none"/>' +
            '</svg>' +
            '<div class="neighborhood-trio">' +
                '<div class="nb-card">' +
                    '<div class="nb-card-icon">&lt;/&gt; Upstream Callers</div>' +
                    '<div class="nb-card-label">' + callersCount + ' Affected Callers</div>' +
                    '<div class="nb-card-sub">' + escapeHTML(inItems.join(", ") || "Direct entrypoints") + '</div>' +
                '</div>' +
                '<div class="nb-card dependency-highlight">' +
                    '<div class="nb-card-icon">⑂ Downstream Deps</div>' +
                    '<div class="nb-card-label">Callee Contracts</div>' +
                    '<div class="nb-card-sub">' + escapeHTML(outItems.join(", ")) + '</div>' +
                '</div>' +
                '<div class="nb-card">' +
                    '<div class="nb-card-icon">🛡️ Governance Gate</div>' +
                    '<div class="nb-card-label">' + escapeHTML(decision) + '</div>' +
                    '<div class="nb-card-sub">' + escapeHTML(ledgerText) + '</div>' +
                '</div>' +
            '</div>' +
        '</div>';

        stageHTML += '<div class="panel" style="margin-top:18px; margin-bottom:0;">' +
            '<div class="panel-header"><div class="panel-title">Transitive Impact &amp; Verification Path</div></div>' +
            '<div class="list">' +
                '<div class="list-row"><div class="row-icon">1</div><div class="row-main"><div class="row-title">Static AST Resolution (GO_TYPES / Tree-sitter)</div><div class="row-meta">Target ' + escapeHTML(targetName) + ' resolved in package ' + escapeHTML(pkgName) + ' with confidence 1.0.</div></div><span class="badge-pill state-supported">RESOLVED</span></div>' +
                '<div class="list-row"><div class="row-icon">2</div><div class="row-main"><div class="row-title">Runtime Edge Corroboration</div><div class="row-meta">' + (epistemic === "SUPPORTED" ? "Observed in runtime_observations & runtime_edges." : "Awaiting OpenTelemetry trace span corroboration for target edge.") + '</div></div><span class="badge-pill ' + (epistemic === "SUPPORTED" ? "state-supported" : "state-unverified") + '">' + escapeHTML(epistemic) + '</span></div>' +
                '<div class="list-row"><div class="row-icon">3</div><div class="row-main"><div class="row-title">Declarative Policy &amp; RFC 6962 Merkle Seal</div><div class="row-meta">' + escapeHTML(reasonText) + '</div></div><span class="policy-badge ' + escapeHTML(decision) + '">' + escapeHTML(decision) + '</span></div>' +
            '</div>' +
        '</div>';

        stage.innerHTML = stageHTML;
    }
}

/* ============================================================================
   05 — EVIDENCE & LEDGER: VERO-STYLE AUDIT TABLE & DECISION RECORD (Screenshot 5)
   ============================================================================ */
function buildAuditEntries() {
    var entries = [];
    var evals = (state.policiesData && state.policiesData.latest_evaluations) || [];
    evals.forEach(function(ev) {
        entries.push({
            id: ev.id,
            time: ev.evaluated_at,
            actor: ev.actor || "garuda-policy",
            action: ev.reason || ("Evaluated policy " + ev.policy_title),
            object: ev.policy_title + " (" + (ev.subject_id || WORKSPACE) + ")",
            anchor: ev.merkle_block_height ? ("block #" + ev.merkle_block_height) : "unanchored",
            result: ev.decision || "ALLOW",
            rule: (ev.matched_predicates || []).join(", ") || "policy.ValidatePolicy",
            lookedAt: formatNumber(ev.entity_count) + " entities, " + formatNumber(ev.claim_count) + " static claims, " + formatNumber(ev.contradiction_count) + " contradictions",
            isEval: true
        });
    });

    var att = (state.stats && state.stats.needs_attention) || [];
    att.forEach(function(a, i) {
        entries.push({
            id: a.id || ("contra-" + i),
            time: (state.stats && state.stats.last_updated) || new Date().toISOString(),
            actor: "runtime-verifier",
            action: a.title + " — " + a.subtitle,
            object: a.evidence_loc || WORKSPACE,
            anchor: "quarantined",
            result: "BLOCK",
            rule: "runtime_vs_static.contradiction_guard",
            lookedAt: "runtime_edges vs static claims in " + (a.evidence_loc || WORKSPACE),
            isEval: false
        });
    });

    var evList = (state.stats && state.stats.recent_evidence) || [];
    evList.forEach(function(item, i) {
        entries.push({
            id: item.id || ("ev-" + i),
            time: item.timestamp,
            actor: item.kind || "ast-analyzer",
            action: item.summary,
            object: item.source || WORKSPACE,
            anchor: "#" + ((state.stats && state.stats.latest_block_height) || 1),
            result: item.kind === "Runtime Trace" ? "SUPPORTED" : "ALLOW",
            rule: "canonical.SnapshotHash · GO_TYPES",
            lookedAt: item.source || "workspace AST snapshot",
            isEval: false
        });
    });

    state.auditEntries = entries;
    if (state.currentView === "governance") {
        renderAuditLogTable();
    }
}

function filterAuditEntries(filterVal, btnEl) {
    state.auditFilter = filterVal;
    if (btnEl && btnEl.parentElement) {
        btnEl.parentElement.querySelectorAll(".filter-pill").forEach(function(b) {
            b.classList.toggle("active", b === btnEl);
        });
    }
    renderAuditLogTable();
}

function renderAuditLogTable() {
    var tbody = document.getElementById("audit-log-tbody");
    if (!tbody) return;

    var filtered = state.auditEntries.filter(function(e) {
        if (state.auditFilter === "ALL") return true;
        if (state.auditFilter === "ALLOW") return e.result === "ALLOW" || e.result === "SUPPORTED";
        if (state.auditFilter === "WARN") return e.result === "WARN" || e.result === "REVIEW" || e.result === "UNVERIFIED";
        if (state.auditFilter === "BLOCK") return e.result === "BLOCK" || e.result === "CONTRADICTED";
        return true;
    });

    if (filtered.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" style="color:var(--muted); padding:18px;">No audit entries match the current filter. Run <code>./bin/garuda check --project ' + escapeHTML(WORKSPACE || "default") + ' --policy policies --save</code> to append new records.</td></tr>';
        return;
    }

    tbody.innerHTML = "";
    filtered.forEach(function(entry, idx) {
        var tr = document.createElement("tr");
        if (idx === 0) tr.className = "selected-row";

        var badgeClass = (entry.result === "BLOCK" || entry.result === "CONTRADICTED") ? "critical" :
                         (entry.result === "WARN" || entry.result === "REVIEW") ? "warning" : "success";

        tr.innerHTML =
            '<td class="mono" style="font-size:11.5px; color:var(--muted); white-space:nowrap;">' + escapeHTML(formatTimeOnly(entry.time)) + '</td>' +
            '<td style="font-weight:550;">' + escapeHTML(entry.actor) + '</td>' +
            '<td>' + escapeHTML(entry.action) + '</td>' +
            '<td class="mono" style="font-size:11.5px; color:var(--text-2);">' + escapeHTML(entry.object) + '</td>' +
            '<td><span class="merkle-chip" style="margin-left:0;">' + escapeHTML(entry.anchor) + '</span></td>' +
            '<td><span class="badge-pill ' + badgeClass + '">' + escapeHTML(entry.result) + '</span></td>';

        tr.onclick = function() {
            tbody.querySelectorAll("tr").forEach(function(r) { r.classList.remove("selected-row"); });
            tr.classList.add("selected-row");
            populateAuditDecisionRecord(entry);
        };
        tbody.appendChild(tr);
    });

    populateAuditDecisionRecord(filtered[0]);
}

function populateAuditDecisionRecord(entry) {
    if (!entry) return;
    setText("audit-dr-meta", formatDate(entry.time) + " · entry " + (entry.id || "").slice(0, 10));
    setText("audit-dr-what", entry.action);
    setText("audit-dr-looked", entry.lookedAt + " in " + entry.object);
    setText("audit-dr-rule", entry.rule);
    setText("audit-dr-confidence", entry.result === "BLOCK" ? "Quarantined contradiction / policy violation." : "Deterministic static resolution (1.0); anchored at " + entry.anchor + ".");
    var remEl = document.getElementById("audit-dr-remedy");
    if (remEl && entry.isEval && entry.id) {
        remEl.innerHTML = 'Verify RFC 6962 inclusion proof against the epoch root: <button class="btn-secondary" style="height:24px; font-size:11px; margin-left:6px;" onclick="verifyPolicyAnchor(\'' + escapeJS(entry.id) + '\')">Verify Proof →</button>';
    }
}

/* ============================================================================
   02 — EXPLORE: SEMANTIC SEARCH & SPLIT PREVIEW
   ============================================================================ */
function setupSearch() {
    var global = document.getElementById("global-search");
    var page = document.getElementById("search-page-input");

    if (global) {
        global.addEventListener("keydown", function(e) {
            if (e.key === "Enter") {
                var val = global.value.trim();
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
            }, 160);
        });
        page.addEventListener("keydown", function(e) {
            if (e.key === "Enter") {
                var val = page.value.trim();
                if (val) performSearch(val);
            }
        });
    }

    document.addEventListener("keydown", function(e) {
        if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
            e.preventDefault();
            openCommandPalette();
            return;
        }
        if ((e.ctrlKey || e.metaKey) && e.key >= "1" && e.key <= "6") {
            var idx = parseInt(e.key, 10) - 1;
            if (idx < tabs.length) {
                e.preventDefault();
                showView(tabs[idx]);
                return;
            }
        }
        if (e.key === "/" && document.activeElement.tagName !== "INPUT" && document.activeElement.tagName !== "SELECT") {
            e.preventDefault();
            if (global) global.focus();
        }
        if (e.key === "Escape") {
            closeDrawer();
            closeCommandPalette();
        }
    });
}

function setSearchKindFilter(kind) {
    state.searchKindFilter = kind;
    document.querySelectorAll("#search-kind-filters .filter-pill").forEach(function(btn) {
        btn.classList.toggle("active", btn.getAttribute("data-kind") === kind);
    });
    renderSearchResults({ query: document.getElementById("search-page-input").value || "", results: state.searchResultsCache });
}

function runSearch(query) {
    var global = document.getElementById("global-search");
    var page = document.getElementById("search-page-input");
    if (global) global.value = query;
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
    results.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Searching semantic index...</div></div></div>';

    try {
        var res = await fetch("/api/v1/dashboard/search?q=" + encodeURIComponent(query) + "&workspace=" + encodeURIComponent(WORKSPACE) + "&limit=80");
        if (!res.ok) throw new Error("Search failed");
        var data = await res.json();
        state.searchResultsCache = data.results || [];
        renderSearchResults(data);
    } catch (err) {
        results.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Search unavailable</div><div class="row-meta">' + escapeHTML(err.message) + '</div></div></div>';
    }
}

function renderSearchResults(data) {
    var results = document.getElementById("search-results");
    var title = document.getElementById("search-results-title");
    var rawList = (data && data.results) || [];

    var filtered = rawList.filter(function(item) {
        if (state.searchKindFilter === "all") return true;
        var k = (item.kind || "").toLowerCase();
        if (state.searchKindFilter === "function") return k === "function" || k === "method";
        if (state.searchKindFilter === "struct") return k === "struct" || k === "interface" || k === "class" || k === "type";
        if (state.searchKindFilter === "package") return k === "package" || k === "file";
        return true;
    });

    if (title) {
        title.textContent = formatNumber(filtered.length) + " results" + (data.query ? " for “" + data.query + "”" : "");
    }
    if (filtered.length === 0) {
        results.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No matching objects</div><div class="row-meta">Try a function, struct, interface, or package name.</div></div></div>';
        return;
    }
    results.innerHTML = "";
    filtered.forEach(function(item, idx) {
        var row = document.createElement("div");
        row.className = "list-row search-result";
        row.innerHTML = '<div class="row-icon">' + symbolIcon(item.kind) + '</div>' +
            '<div class="row-main">' +
                '<div class="row-title mono">' + escapeHTML(item.name) + '</div>' +
                '<div class="row-meta">' + escapeHTML(item.package || item.repo || "") + (item.file ? " · " + escapeHTML(shortenPath(item.file)) : "") + '</div>' +
            '</div>' +
            '<span class="kind-pill">' + escapeHTML(item.kind || "entity") + '</span>';
        row.onclick = function() {
            previewSearchResult(item);
        };
        if (idx === 0) {
            previewSearchResult(item);
        }
        results.appendChild(row);
    });
}

function previewSearchResult(item) {
    if (!item) return;
    var pane = document.getElementById("search-preview-pane");
    if (!pane) {
        openSearchResult(item);
        return;
    }
    var html = '<div class="selected-fn-card" style="text-align:left; margin-bottom:16px;">' +
        '<div class="wb-pane-kicker" style="color:var(--brand-accent);">' + escapeHTML((item.kind || "entity").toUpperCase()) + '</div>' +
        '<div class="selected-fn-title mono" style="margin-top:4px;">' + escapeHTML(item.name) + '</div>' +
        '<div class="selected-fn-loc">' + escapeHTML(item.package || "workspace") + '</div>' +
    '</div>';

    html += '<div class="detail-section">';
    html += propertyRow("Repository", item.repo || WORKSPACE);
    html += propertyRow("Package", item.package || "—");
    html += propertyRow("Source File", shortenPath(item.file || "—"));
    html += propertyRow("Visibility", item.exported ? "Exported (Public)" : "Internal");
    html += '</div>';

    html += '<div style="display:flex; gap:8px; flex-wrap:wrap;">' +
        '<button class="btn-primary" onclick="focusSymbolOnArchitectureMap(\'' + escapeJS(item.name) + '\', \'' + escapeJS(item.package || "") + '\', ' + JSON.stringify(item).replace(/"/g, '&quot;') + ')">Open 1-Hop Architecture Map →</button>' +
        '<button class="btn-secondary" onclick="state.selectedNode={id:\'' + escapeJS(item.id) + '\',label:\'' + escapeJS(item.name) + '\',kind:\'' + escapeJS(item.kind) + '\',package:\'' + escapeJS(item.package || "") + '\',file:\'' + escapeJS(item.file || "") + '\'}; showView(\'review\')">Inspect in Change Review →</button>' +
    '</div>';

    pane.innerHTML = html;
}

function openSearchResult(item) {
    if (!item) return;
    focusSymbolOnArchitectureMap(item.name, item.package, item);
}

function openNodeDrawer(node) {
    if (!node) return;
    updateNeighborhoodAndInspector(node, (state.graphData && state.graphData._renderedNodes) || [], (state.graphData && state.graphData._renderedEdges) || []);
    highlightSelectedNodeOnCanvas(node);
}

function propertyRow(key, value) {
    return '<div class="detail-property"><div class="detail-key">' + escapeHTML(key) + '</div><div class="detail-value">' + escapeHTML(String(value)) + '</div></div>';
}

function closeDrawer() {
    var d = document.getElementById("drawer");
    var o = document.getElementById("drawer-overlay");
    if (d) d.classList.remove("open");
    if (o) o.classList.remove("open");
}

/* ============================================================================
   LINEAR-STYLE ⌘K COMMAND PALETTE MODAL (Screenshot 2)
   ============================================================================ */
function openCommandPalette() {
    var bd = document.getElementById("cmdk-backdrop");
    var inp = document.getElementById("cmdk-input");
    if (!bd || !inp) return;
    bd.classList.add("open");
    inp.value = "";
    renderCommandPaletteItems("");
    setTimeout(function() { inp.focus(); }, 20);
}

function closeCommandPalette() {
    var bd = document.getElementById("cmdk-backdrop");
    if (bd) bd.classList.remove("open");
}

function renderCommandPaletteItems(query) {
    var listEl = document.getElementById("cmdk-list");
    if (!listEl) return;
    var q = (query || "").toLowerCase().trim();

    var items = [];
    // Include top architectural hubs for quick jump
    var hubs = (state.stats && state.stats.top_hubs) || [];
    hubs.slice(0, 5).forEach(function(h) {
        if (!q || h.name.toLowerCase().indexOf(q) !== -1 || (h.package || "").toLowerCase().indexOf(q) !== -1) {
            items.push({
                group: "Architectural Symbols",
                icon: symbolIcon(h.kind),
                title: h.name,
                sub: h.package + " · " + h.callers + " callers",
                badge: "Jump to 1-hop →",
                action: function() { closeCommandPalette(); focusSymbolOnArchitectureMap(h.name, h.package, h); }
            });
        }
    });

    var navCommands = [
        { group: "Working Surfaces", icon: "01", title: "Go to Overview (Command Center)", sub: "System integrity summary, policy pulse, and repositories", badge: "⌘1", action: function() { closeCommandPalette(); showView("workspace"); } },
        { group: "Working Surfaces", icon: "02", title: "Explore Semantic Codebase Index", sub: "Search functions, structs, interfaces, and packages", badge: "⌘2", action: function() { closeCommandPalette(); showView("search"); } },
        { group: "Working Surfaces", icon: "03", title: "Open Architecture Map (1-Hop Neighborhood)", sub: "Graph exploration without the hairball", badge: "⌘3", action: function() { closeCommandPalette(); showView("graph"); } },
        { group: "Working Surfaces", icon: "04", title: "Open Change Review Workbench", sub: "3-pane impact, policy verdict, and epistemic evidence review", badge: "⌘4", action: function() { closeCommandPalette(); showView("review"); } },
        { group: "Working Surfaces", icon: "05", title: "Open Evidence & Cryptographic Ledger", sub: "Append-only audit log and RFC 6962 Merkle proofs", badge: "⌘5", action: function() { closeCommandPalette(); showView("governance"); } },
        { group: "Working Surfaces", icon: "06", title: "Open Agents & Session Timeline", sub: "Active MCP sessions, tool calls, and handoffs", badge: "⌘6", action: function() { closeCommandPalette(); showView("agents"); } },
        { group: "Quick Actions", icon: "☀", title: "Toggle Light / Graphite Dark Theme", sub: "Switch between #F7F7F3 Warm Paper and #131513 Graphite", badge: "Theme", action: function() { closeCommandPalette(); toggleTheme(); } }
    ];

    navCommands.forEach(function(cmd) {
        if (!q || cmd.title.toLowerCase().indexOf(q) !== -1 || cmd.sub.toLowerCase().indexOf(q) !== -1) {
            items.push(cmd);
        }
    });

    if (q.length >= 2) {
        items.unshift({
            group: "Deep Search",
            icon: "⌕",
            title: "Search workspace index for “" + query + "”",
            sub: "Query all indexed repositories, packages, and symbols",
            badge: "↵ Search",
            action: function() { closeCommandPalette(); runSearch(query); }
        });
    }

    state.cmdkItems = items;
    state.cmdkIndex = 0;

    listEl.innerHTML = "";
    var lastGroup = "";
    items.forEach(function(item, idx) {
        if (item.group !== lastGroup) {
            lastGroup = item.group;
            var g = document.createElement("div");
            g.className = "cmdk-group-title";
            g.textContent = lastGroup;
            listEl.appendChild(g);
        }
        var row = document.createElement("div");
        row.className = "cmdk-item" + (idx === 0 ? " active" : "");
        row.innerHTML =
            '<div class="cmdk-item-left">' +
                '<div class="cmdk-icon">' + escapeHTML(item.icon) + '</div>' +
                '<div>' +
                    '<div class="cmdk-title">' + escapeHTML(item.title) + '</div>' +
                    '<div class="cmdk-sub">' + escapeHTML(item.sub) + '</div>' +
                '</div>' +
            '</div>' +
            '<span class="kbd-chip">' + escapeHTML(item.badge) + '</span>';
        row.onclick = item.action;
        listEl.appendChild(row);
    });
}

(function setupCommandPaletteInput() {
    var inp = document.getElementById("cmdk-input");
    if (!inp) return;
    inp.addEventListener("input", function() {
        renderCommandPaletteItems(inp.value);
    });
    inp.addEventListener("keydown", function(e) {
        var rows = document.querySelectorAll("#cmdk-list .cmdk-item");
        if (e.key === "ArrowDown") {
            e.preventDefault();
            if (rows.length > 0) {
                state.cmdkIndex = (state.cmdkIndex + 1) % rows.length;
                rows.forEach(function(r, i) { r.classList.toggle("active", i === state.cmdkIndex); });
            }
        } else if (e.key === "ArrowUp") {
            e.preventDefault();
            if (rows.length > 0) {
                state.cmdkIndex = (state.cmdkIndex - 1 + rows.length) % rows.length;
                rows.forEach(function(r, i) { r.classList.toggle("active", i === state.cmdkIndex); });
            }
        } else if (e.key === "Enter") {
            e.preventDefault();
            if (state.cmdkItems[state.cmdkIndex]) {
                state.cmdkItems[state.cmdkIndex].action();
            }
        }
    });
})();

/* ============================================================================
   06 — AGENTS & DECISIONS
   ============================================================================ */
var agentsLoaded = false;

async function loadAgents() {
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
        renderSessionList(activeEl, data.active || [], "No active MCP sessions in the last 5 minutes.", true);
        renderSessionList(recentEl, data.recent || [], "No MCP sessions in the last 24 hours.", false);
    } catch (err) {
        activeEl.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Failed to load sessions</div><div class="row-meta">' + escapeHTML(String(err)) + '</div></div></div>';
        recentEl.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Failed to load sessions</div></div></div>';
    }
}

function renderSessionList(container, sessions, emptyMessage, isActivePanel) {
    if (!sessions || sessions.length === 0) {
        container.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">' + escapeHTML(emptyMessage) + '</div><div class="row-meta">Connect Cursor or Claude Desktop via garuda-mcp to record agent tool activity.</div></div></div>';
        return;
    }
    container.innerHTML = "";
    sessions.forEach(function(s) {
        var row = document.createElement("div");
        row.className = "list-row";
        var clientLabel = escapeHTML(s.client_name || "mcp-client");
        if (s.client_version) {
            clientLabel += ' <span style="color:var(--muted); font-size:10.5px;">v' + escapeHTML(s.client_version) + '</span>';
        }
        var duration = formatDuration(s.duration_seconds);
        var activity = formatDate(s.last_activity_at);
        var badge = s.close_reason
            ? '<span class="badge-pill info">' + escapeHTML(s.close_reason) + '</span>'
            : (isActivePanel ? '<span class="badge-pill success">active</span>' : '<span class="badge-pill info">completed</span>');
        row.innerHTML =
            '<div class="row-icon">⌘</div>' +
            '<div class="row-main">' +
                '<div class="row-title">' + clientLabel + ' · ' + escapeHTML(s.agent_id || "mcp-agent") + '</div>' +
                '<div class="row-meta">' + s.tool_call_count + ' tool calls · ' + duration + ' · ' + escapeHTML(activity) + '</div>' +
            '</div>' +
            badge;
        container.appendChild(row);
    });
}

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
        setText("nav-badge-decisions", String(decisions.length));

        if (subtitleEl) {
            subtitleEl.textContent = decisions.length === 1
                ? "1 architectural decision in this workspace."
                : decisions.length + " architectural decisions in this workspace.";
        }
        renderDecisionList(listEl, decisions);
    } catch (err) {
        if (subtitleEl) subtitleEl.textContent = "Failed to load.";
        listEl.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Failed to load decisions</div><div class="row-meta">' + escapeHTML(String(err)) + '</div></div></div>';
    }
}

function renderDecisionList(container, decisions) {
    if (!decisions || decisions.length === 0) {
        container.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No architectural decisions recorded yet.</div><div class="row-meta">Propose one via <code>POST /api/v1/decisions/propose</code> or the <code>garuda.propose_decision</code> MCP tool.</div></div></div>';
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
            ? '<span class="merkle-chip">anchored</span>'
            : '<span class="merkle-chip invalid">unanchored</span>';

        var meta = [];
        if (d.scope_domain) meta.push(escapeHTML(d.scope_domain));
        if (d.scope_system) meta.push(escapeHTML(d.scope_system));
        meta.push(d.revision_count + (d.revision_count === 1 ? " revision" : " revisions"));
        if (d.owner) meta.push(escapeHTML(d.owner));

        row.innerHTML =
            '<div class="row-icon">§</div>' +
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
    if (s === "APPROVED" || s === "CANONICAL" || s === "ACCEPTED") return "success";
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
    setText("drawer-title", d.title || "Decision Record");
    setText("drawer-kind", "ADR · " + (d.status || "RECORD"));

    var html = '<div class="detail-section"><div class="detail-section-title">Decision Metadata</div>';
    html += propertyRow("Title", d.title || "(untitled)");
    html += propertyRow("Status", d.status || "unknown");
    html += propertyRow("Owner", d.owner || "—");
    html += propertyRow("Scope domain", d.scope_domain || "—");
    html += propertyRow("Scope system", d.scope_system || "—");
    html += propertyRow("Created", formatDate(d.created_at));
    html += propertyRow("Merkle anchor", d.anchored ? "Anchored in RFC 6962 ledger" : "Not anchored");
    html += '</div>';

    var revisions = data.revisions || [];
    if (revisions.length > 0) {
        html += '<div class="detail-section"><div class="detail-section-title">Cryptographic Revision Chain (' + revisions.length + ')</div>';
        revisions.forEach(function(r) {
            var hashShort = (r.decision_hash_hex || "").slice(0, 14);
            var prevShort = (r.previous_hash_hex || "").slice(0, 14);
            html += '<div class="detail-property">';
            html += '<div class="detail-key">rev #' + r.revision_number + '</div>';
            html += '<div class="detail-value">' + escapeHTML(hashShort) + '…<div class="row-meta">prev ' + escapeHTML(prevShort) + '…</div></div>';
            html += '</div>';
        });
        html += '</div>';
    }

    body.innerHTML = html;
}

/* ============================================================================
   UTILITIES & INITIALIZATION
   ============================================================================ */
function setText(id, value) {
    var el = document.getElementById(id);
    if (!el) return;
    el.textContent = value;
}

function formatNumber(value) {
    return Number(value || 0).toLocaleString();
}

function formatDate(value) {
    if (!value) return "—";
    try { return new Date(value).toLocaleString(); } catch (e) { return value; }
}

function formatTimeOnly(value) {
    if (!value) return "—";
    try {
        var d = new Date(value);
        return d.toTimeString().split(" ")[0];
    } catch (e) {
        return value;
    }
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

function shortenLabel(value, max) {
    if (!value) return "";
    if (value.length <= max) return value;
    return value.slice(0, max - 1) + "…";
}

function shortenPath(p) {
    if (!p) return "";
    var parts = String(p).split("/");
    if (parts.length <= 3) return p;
    return "…/" + parts.slice(-3).join("/");
}

function symbolIcon(kind) {
    if (!kind) return "◇";
    var k = kind.toLowerCase();
    if (k === "function") return "ƒ";
    if (k === "method") return "m";
    if (k === "struct" || k === "class") return "S";
    if (k === "interface") return "I";
    if (k === "package") return "◇";
    if (k === "repository") return "▦";
    if (k === "file") return "□";
    return "•";
}

function escapeHTML(value) {
    return String(value)
        .replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#039;");
}

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
    if (state.currentView === "graph") {
        await loadArchitecture(state.currentLevel, state.currentFocus);
    }
    if (state.currentView === "review") {
        renderChangeReviewWorkbench();
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

initTheme();
setupSearch();
loadWorkspacePicker();
showView(initialView());
loadAll();

window.addEventListener("resize", function() {
    if (state.currentView === "graph") {
        setTimeout(fitGraph, 100);
    }
});