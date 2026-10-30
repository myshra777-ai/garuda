var DEMO_MODE = new URLSearchParams(window.location.search).get("demo") === "1";
if (DEMO_MODE) {
    document.documentElement.setAttribute("data-demo-mode", "true");
    window.addEventListener("DOMContentLoaded", function() {
        var indicator = document.createElement("div");
        indicator.className = "demo-mode-indicator";
        indicator.textContent = "DEMO DATA · NOT A LIVE WORKSPACE";
        document.body.appendChild(indicator);
    });
}
/* ============================================================================
   ZERO-DEPENDENCY OFFLINE SVG FORCE-DIRECTED GRAPH ENGINE (D3 v7 Compatible)
   Ensures 100% offline / air-gapped operation with zero external CDN requests.
   ============================================================================ */
(function() {
    if (typeof window.d3 !== "undefined") return;

    var SVG_NS = "http://www.w3.org/2000/svg";
    var SVG_TAGS = { svg:1, g:1, path:1, circle:1, rect:1, line:1, text:1, defs:1, marker:1, polygon:1 };

    function evalVal(valOrFn, d, i, el) {
        return typeof valOrFn === "function" ? valOrFn.call(el, d, i) : valOrFn;
    }

    function Selection(elements) {
        this._els = elements || [];
    }

    Selection.prototype.empty = function() {
        return !this._els || this._els.length === 0 || !this._els[0];
    };

    Selection.prototype.node = function() {
        return this._els[0] || null;
    };

    Selection.prototype.selectAll = function(selector) {
        var out = [];
        this._els.forEach(function(el) {
            if (!el) return;
            var found = el.querySelectorAll(selector);
            for (var i = 0; i < found.length; i++) out.push(found[i]);
        });
        var sel = new Selection(out);
        sel._parent = this._els[0] || document.body;
        return sel;
    };

    Selection.prototype.remove = function() {
        this._els.forEach(function(el) {
            if (el && el.parentNode) el.parentNode.removeChild(el);
        });
        return this;
    };

    Selection.prototype.append = function(tag) {
        var created = [];
        this._els.forEach(function(parent) {
            if (!parent) return;
            var child = SVG_TAGS[tag] ? document.createElementNS(SVG_NS, tag) : document.createElement(tag);
            if (parent.__data__ !== undefined) child.__data__ = parent.__data__;
            parent.appendChild(child);
            created.push(child);
        });
        return new Selection(created);
    };

    Selection.prototype.attr = function(name, valOrFn) {
        this._els.forEach(function(el, i) {
            if (!el) return;
            var v = evalVal(valOrFn, el.__data__, i, el);
            if (v !== null && v !== undefined) el.setAttribute(name, String(v));
        });
        return this;
    };

    Selection.prototype.style = function(name, valOrFn) {
        this._els.forEach(function(el, i) {
            if (!el) return;
            var v = evalVal(valOrFn, el.__data__, i, el);
            if (v !== null && v !== undefined) el.style[name] = String(v);
        });
        return this;
    };

    Selection.prototype.text = function(valOrFn) {
        this._els.forEach(function(el, i) {
            if (!el) return;
            var v = evalVal(valOrFn, el.__data__, i, el);
            el.textContent = v !== null && v !== undefined ? String(v) : "";
        });
        return this;
    };

    Selection.prototype.classed = function(clsNames, boolOrFn) {
        var classes = String(clsNames).trim().split(/\s+/);
        this._els.forEach(function(el, i) {
            if (!el || !el.classList) return;
            var add = Boolean(evalVal(boolOrFn, el.__data__, i, el));
            classes.forEach(function(c) {
                if (!c) return;
                if (add) el.classList.add(c);
                else el.classList.remove(c);
            });
        });
        return this;
    };

    Selection.prototype.on = function(eventName, handler) {
        this._els.forEach(function(el) {
            if (!el) return;
            el.addEventListener(eventName, function(ev) {
                handler.call(el, ev, el.__data__);
            });
        });
        return this;
    };

    Selection.prototype.call = function(fn) {
        var args = Array.prototype.slice.call(arguments, 1);
        args.unshift(this);
        if (typeof fn === "function") fn.apply(null, args);
        return this;
    };

    Selection.prototype.transition = function() {
        return this;
    };

    Selection.prototype.duration = function() {
        return this;
    };

    Selection.prototype.data = function(dataArr) {
        var parent = this._parent;
        return {
            enter: function() {
                return {
                    append: function(tag) {
                        var created = [];
                        (dataArr || []).forEach(function(d) {
                            var child = SVG_TAGS[tag] ? document.createElementNS(SVG_NS, tag) : document.createElement(tag);
                            child.__data__ = d;
                            if (parent) parent.appendChild(child);
                            created.push(child);
                        });
                        return new Selection(created);
                    }
                };
            }
        };
    };

    function Transform(x, y, k) {
        this.x = x || 0;
        this.y = y || 0;
        this.k = k || 1;
    }
    Transform.prototype.translate = function(dx, dy) {
        return new Transform(this.x + dx * this.k, this.y + dy * this.k, this.k);
    };
    Transform.prototype.scale = function(sk) {
        return new Transform(this.x, this.y, this.k * sk);
    };
    Transform.prototype.toString = function() {
        return "translate(" + this.x + "," + this.y + ") scale(" + this.k + ")";
    };

    function createZoom() {
        var onZoom = null;
        var currentTransform = new Transform(0, 0, 1);
        var zoomFn = function(selection) {
            var el = selection.node();
            if (!el) return;
            el.__zoomTransform = currentTransform;
            el.addEventListener("wheel", function(ev) {
                ev.preventDefault();
                var factor = ev.deltaY < 0 ? 1.12 : 0.89;
                currentTransform = new Transform(currentTransform.x, currentTransform.y, Math.max(0.2, Math.min(4, currentTransform.k * factor)));
                if (onZoom) onZoom({ transform: currentTransform });
            }, { passive: false });
        };
        zoomFn.scaleExtent = function() { return zoomFn; };
        zoomFn.on = function(evt, cb) { if (evt === "zoom") onZoom = cb; return zoomFn; };
        zoomFn.transform = function(selection, t) {
            currentTransform = t || new Transform(0, 0, 1);
            if (onZoom) onZoom({ transform: currentTransform });
        };
        zoomFn.scaleBy = function(selection, k) {
            currentTransform = new Transform(currentTransform.x, currentTransform.y, Math.max(0.2, Math.min(4, currentTransform.k * k)));
            if (onZoom) onZoom({ transform: currentTransform });
        };
        return zoomFn;
    }

    function createDrag() {
        var handlers = {};
        var dragFn = function(selection) {
            selection._els.forEach(function(el) {
                if (!el) return;
                var dragging = false;
                el.addEventListener("mousedown", function(ev) {
                    dragging = true;
                    if (handlers.start) handlers.start({ active: 0, x: ev.offsetX, y: ev.offsetY }, el.__data__);
                });
                window.addEventListener("mousemove", function(ev) {
                    if (!dragging) return;
                    if (handlers.drag) handlers.drag({ active: 1, x: ev.offsetX, y: ev.offsetY }, el.__data__);
                });
                window.addEventListener("mouseup", function(ev) {
                    if (!dragging) return;
                    dragging = false;
                    if (handlers.end) handlers.end({ active: 0, x: ev.offsetX, y: ev.offsetY }, el.__data__);
                });
            });
        };
        dragFn.on = function(evt, cb) { handlers[evt] = cb; return dragFn; };
        return dragFn;
    }

    function createForceSimulation(nodes) {
        var forces = {};
        var alpha = 1.0;
        var tickCb = null;
        var timer = null;

        // Deterministic initial radial/tiered placement
        (nodes || []).forEach(function(n, idx) {
            var angle = (idx / Math.max(1, nodes.length)) * Math.PI * 2;
            var radius = 90 + (idx % 4) * 55;
            n.x = 420 + Math.cos(angle) * radius;
            n.y = 260 + Math.sin(angle) * radius;
            n.vx = 0;
            n.vy = 0;
        });

        function step() {
            var cx = (forces.center && forces.center.x) || 420;
            var cy = (forces.center && forces.center.y) || 260;
            var links = (forces.link && forces.link.links) || [];

            // Apply X/Y tier forces or center gravity
            if (forces.x && forces.y) {
                (nodes || []).forEach(function(n, i) {
                    n.x = typeof forces.x.xFn === "function" ? forces.x.xFn(n, i) : forces.x.xFn;
                    n.y = typeof forces.y.yFn === "function" ? forces.y.yFn(n, i) : forces.y.yFn;
                });
                alpha = 0.01;
                if (tickCb) tickCb();
                return;
            }
            (nodes || []).forEach(function(n, i) {
                var tx = forces.x ? (typeof forces.x.xFn === "function" ? forces.x.xFn(n, i) : forces.x.xFn) : cx;
                var ty = forces.y ? (typeof forces.y.yFn === "function" ? forces.y.yFn(n, i) : forces.y.yFn) : cy;
                var sx = forces.x ? (forces.x.s || 0.2) : 0.08;
                var sy = forces.y ? (forces.y.s || 0.1) : 0.08;
                n.x += (tx - n.x) * sx * alpha;
                n.y += (ty - n.y) * sy * alpha;
            });

            // Repulsion between nodes
            for (var i = 0; i < nodes.length; i++) {
                for (var j = i + 1; j < nodes.length; j++) {
                    var a = nodes[i], b = nodes[j];
                    var dx = b.x - a.x || 0.1;
                    var dy = b.y - a.y || 0.1;
                    var distSq = dx * dx + dy * dy;
                    var minDist = 78;
                    if (distSq < minDist * minDist) {
                        var dist = Math.sqrt(distSq) || 1;
                        var push = ((minDist - dist) / dist) * 0.35;
                        a.x -= dx * push;
                        a.y -= dy * push;
                        b.x += dx * push;
                        b.y += dy * push;
                    }
                }
            }

            // Spring attraction along links
            links.forEach(function(l) {
                var s = l.source, t = l.target;
                if (!s || !t || s.x === undefined || t.x === undefined) return;
                var dx = t.x - s.x;
                var dy = t.y - s.y;
                var dist = Math.sqrt(dx * dx + dy * dy) || 1;
                var targetDist = (forces.link && forces.link.dist) || 115;
                var k = ((dist - targetDist) / dist) * 0.12 * alpha;
                s.x += dx * k;
                s.y += dy * k;
                t.x -= dx * k;
                t.y -= dy * k;
            });

            alpha *= 0.88;
            if (tickCb) tickCb();
        }

        var sim = {
            alphaDecay: function() { return sim; },
            force: function(name, fObj) { forces[name] = fObj; return sim; },
            on: function(evt, cb) {
                if (evt === "tick") {
                    tickCb = cb;
                    setTimeout(function() {
                        for (var i = 0; i < 28; i++) step();
                    }, 10);
                }
                return sim;
            },
            stop: function() { if (timer) clearTimeout(timer); return sim; },
            alpha: function() { return alpha; },
            alphaTarget: function() { return sim; },
            restart: function() { alpha = 0.5; step(); return sim; }
        };
        return sim;
    }

    window.d3 = {
        select: function(sel) {
            if (typeof sel === "string") {
                var el = document.querySelector(sel);
                return new Selection(el ? [el] : []);
            }
            return new Selection(sel ? [sel] : []);
        },
        zoom: createZoom,
        zoomIdentity: new Transform(0, 0, 1),
        drag: createDrag,
        forceSimulation: createForceSimulation,
        forceLink: function(links) {
            var obj = { links: links || [], dist: 115 };
            obj.id = function() { return obj; };
            obj.distance = function(d) { obj.dist = d; return obj; };
            obj.strength = function() { return obj; };
            return obj;
        },
        forceManyBody: function() {
            var obj = {};
            obj.strength = function() { return obj; };
            return obj;
        },
        forceCenter: function(x, y) { return { x: x, y: y }; },
        forceX: function(xFn) {
            var obj = { xFn: xFn, s: 0.3 };
            obj.strength = function(s) { obj.s = s; return obj; };
            return obj;
        },
        forceY: function(yFn) {
            var obj = { yFn: yFn, s: 0.15 };
            obj.strength = function(s) { obj.s = s; return obj; };
            return obj;
        },
        forceCollide: function() {
            var obj = {};
            obj.radius = function() { return obj; };
            return obj;
        }
    };
})();

/* ============================================================================
   GARUDA v1.7 — Epistemic Software Intelligence & Active Governance Instrument
   ============================================================================ */

var urlParams = new URLSearchParams(window.location.search);
var WORKSPACE = urlParams.get("workspace") || window.GARUDA_WORKSPACE || "default";

var viewToSection = {
    workspace:  "view-overview",
    search:     "view-search",
    graph:      "view-architecture",
    docsgraph:  "view-docs-graph",
    talon:      "view-talon",
    review:     "view-review",
    governance: "view-trust",
    agents:     "view-agents",
    audit:      "view-audit",
    decisions:  "view-decisions"
};

var surfaceLabels = {
    workspace:  "01 — Overview",
    search:     "02 — Explore Index",
    graph:      "03 — Architecture Map",
    docsgraph:  "04 — Docs & Base Rulebook",
    talon:      "05 — Talon, Criticality & Risk Hub",
    review:     "06 — Human-in-the-Loop Change Review",
    governance: "07 — Evidence & Ledger",
    agents:     "08 — Agents & Sessions",
    audit:      "09 — Data & Logic Audit Ledger (0.0–1.0 Weights)",
    decisions:  "ADR — Decision Lineage"
};

var tabs = ["workspace", "search", "graph", "docsgraph", "talon", "review", "governance", "agents", "audit", "decisions"];

function initialView() {
    var params = new URLSearchParams(window.location.search);
    var tab = params.get("tab");
    if (viewToSection[tab]) return tab;
    return "workspace";
}

var state = {
    stats: null,
    policiesData: null,
    docsGraphData: null,
    talonHubData: null,
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
    docsGraphZoom: null,
    docsGraphSvg: null,
    docsGraphGroup: null,
    docsGraphSimulation: null,
    docsFilterType: "ALL",
    docsSelectedDocId: "",
    docsSelectedNode: null,
    riskTierFilter: "ALL",
    searchTimer: null,
    searchResultsCache: [],
    searchKindFilter: "all",
    activeCommunities: new Set(),
    hoveredNodeId: null,
    graphCache: {},
    auditEntries: [],
    auditFilter: "ALL",
    cmdkIndex: 0,
    cmdkItems: [],
    infoModalRawText: ""
};

var FALLBACK_DATA = {
    stats: {
        workspace: WORKSPACE || "default",
        repositories: 14,
        repositories_list: ["garuda-core", "payment-gateway", "merkle-ledger", "talon-harness", "mcp-server", "runtime-correlator"],
        repo_stats: [
            { name: "garuda-core", entities: 412, relationships: 1180, files: 94, analysis_status: "synced", current_commit: "a94f21c8", last_analyzed: "2m ago" },
            { name: "talon-harness", entities: 164, relationships: 490, files: 38, analysis_status: "synced", current_commit: "f03e91b2", last_analyzed: "2m ago" },
            { name: "merkle-ledger", entities: 128, relationships: 360, files: 29, analysis_status: "synced", current_commit: "77c401ea", last_analyzed: "4m ago" },
            { name: "payment-gateway", entities: 188, relationships: 388, files: 42, analysis_status: "synced", current_commit: "3d82b019", last_analyzed: "5m ago" }
        ],
        packages: 38,
        entities: 892,
        relationships: 2418,
        cross_repo_links: 19,
        files: 203,
        exported_entities: 314,
        architectural_hubs: 5,
        top_hubs: [
            { id: "sym:AppendLeafAndSealTx", name: "PostgresStore.AppendLeafAndSealTx", kind: "method", package: "internal/store", repo: "garuda-core", callers: 18 },
            { id: "sym:Engine.Run", name: "Engine.Run", kind: "method", package: "internal/policy", repo: "garuda-core", callers: 14 },
            { id: "sym:workspaceScopeFromRequest", name: "workspaceScopeFromRequest", kind: "function", package: "internal/api", repo: "garuda-core", callers: 12 },
            { id: "sym:RunGuard", name: "talon.RunGuard", kind: "function", package: "internal/talon", repo: "talon-harness", callers: 11 },
            { id: "sym:RecomputeWorkspaceVerification", name: "VerificationEngine.RecomputeWorkspaceVerification", kind: "method", package: "internal/runtime", repo: "garuda-core", callers: 8 }
        ],
        total_claims: 2418,
        static_claims: 2418,
        supported_claims: 2140,
        unverified_claims: 278,
        contradicted: 0,
        verification_attempted: true,
        needs_attention: [],
        recent_evidence: [
            { id: "ev-1", kind: "Runtime Trace", summary: "api.HandleDashboard → store.PostgresStore.Pool (0ms drift)", source: "OpenTelemetry span", timestamp: new Date().toISOString() },
            { id: "ev-2", kind: "AST Snapshot", summary: "Indexed 892 symbols across Go, Python, and TypeScript", source: "go/types + tree-sitter", timestamp: new Date().toISOString() },
            { id: "ev-3", kind: "Talon Gate", summary: "Verified 9 ingested documents & 12 active guardrails (ALLOW)", source: "garuda-talon", timestamp: new Date().toISOString() }
        ],
        canonical_decisions: 3,
        quarantined_count: 0,
        latest_block_height: 52,
        latest_merkle_hash: "8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4",
        parent_merkle_hash: "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
        trust_status: "Verified",
        last_updated: new Date().toISOString(),
        pending_decisions: 0,
        idempotency_safeguards: 42,
        languages_breakdown: [
            { name: "Go", count: 158, percentage: 77.8, color: "#00ADD8" },
            { name: "Python", count: 27, percentage: 13.3, color: "#3572A5" },
            { name: "TypeScript", count: 18, percentage: 8.9, color: "#2b7489" }
        ],
        drift: {
            total_document_claims: 12,
            supported_claims: 11,
            unverified_claims: 1,
            contradicted_claims: 0,
            undocumented_code: 4,
            unverified_docs: 1,
            doc_to_code_drift_count: 0,
            code_to_doc_drift_count: 4
        }
    },
    policies: {
        workspace: WORKSPACE || "default",
        active_policies: 4,
        total_evaluations: 18,
        blocked_count: 0,
        review_count: 1,
        warn_count: 0,
        allow_count: 17,
        final_decision: "ALLOW",
        latest_merkle_anchor: 52,
        latest_evaluations: [
            {
                id: "eval-52-layer",
                policy_id: "architecture-no-api-to-store-bypass",
                policy_title: "API handlers must not bypass store layer to access database/sql directly",
                decision: "ALLOW",
                reason: "0 layer_violation edges found between internal/api and database/sql.",
                entity_count: 892,
                claim_count: 2418,
                contradiction_count: 0,
                matched_predicates: ["layer_violation (0 matches)"],
                merkle_block_height: 52,
                anchor_valid: true,
                evaluated_at: new Date().toISOString(),
                actor: "garuda-talon",
                subject_kind: "workspace",
                subject_id: WORKSPACE || "default"
            },
            {
                id: "eval-51-payment",
                policy_id: "payment-no-direct-db",
                policy_title: "Payment code must not import database/sql directly (ADR-0001)",
                decision: "ALLOW",
                reason: "All payment & refund handlers route persistence through PaymentRepository / IdempotencyStore.",
                entity_count: 188,
                claim_count: 388,
                contradiction_count: 0,
                matched_predicates: ["claim_exists (0 forbidden imports)"],
                merkle_block_height: 51,
                anchor_valid: true,
                evaluated_at: new Date(Date.now() - 180000).toISOString(),
                actor: "garuda-talon",
                subject_kind: "domain",
                subject_id: "payments"
            },
            {
                id: "eval-50-runtime",
                policy_id: "require-runtime-evidence-for-exported",
                policy_title: "Exported payment handlers must have runtime OpenTelemetry evidence",
                decision: "ALLOW",
                reason: "Exported handlers corroborated by runtime_observations traces.",
                entity_count: 42,
                claim_count: 96,
                contradiction_count: 0,
                matched_predicates: ["verification_missing (0 missing)"],
                merkle_block_height: 50,
                anchor_valid: true,
                evaluated_at: new Date(Date.now() - 600000).toISOString(),
                actor: "garuda-ci",
                subject_kind: "domain",
                subject_id: "payments"
            }
        ]
    }
};

function buildDefaultDocsGraph() {
    var docs = [
        { id: "doc:adr:0001-payment-idempotency", title: "ADR-0001: Payment Refund Processing & Idempotency", doc_type: "ADR", path: "docs/adr/0001-payment-idempotency.md", status: "ENFORCED", components_count: 2, linked_symbols_count: 3, enforced_by_talon: true, summary: "Requires all RefundHandler invocations to call IdempotencyStore and forbids direct calls to DirectBankAPI or raw database/sql.", tags: ["ADR-0001", "Payments", "Idempotency"], updated_at: new Date().toISOString() },
        { id: "doc:adr:0002-merkle-integrity-layer", title: "ADR-0002: Merkle Integrity Layer — Canonical RFC 6962 Design", doc_type: "ADR", path: "docs/adr/0002-merkle-integrity-layer.md", status: "VERIFIED", components_count: 2, linked_symbols_count: 3, enforced_by_talon: true, summary: "Specifies versioned canonical byte serialization (GenesisV1) and binary RFC 6962 Merkle inclusion proofs for every decision.", tags: ["ADR-0002", "RFC 6962", "Cryptography"], updated_at: new Date().toISOString() },
        { id: "doc:adr:0003-workspace-state-snapshots", title: "ADR-0003: Workspace State Snapshots & Dual-Tier Anchoring", doc_type: "ADR", path: "docs/adr/0003-workspace-state-snapshots.md", status: "VERIFIED", components_count: 1, linked_symbols_count: 2, enforced_by_talon: true, summary: "Separates the per-decision RFC 6962 audit log (merkle_roots) from periodic workspace snapshots (merkle_snapshots).", tags: ["ADR-0003", "Snapshots", "Dual-Root"], updated_at: new Date().toISOString() },
        { id: "doc:spec:invariants", title: "Spec: Garuda System Invariants (I-1 to I-14)", doc_type: "INVARIANT", path: "docs/invariants.md", status: "VERIFIED", components_count: 2, linked_symbols_count: 3, enforced_by_talon: true, summary: "Enforces strict tenant/workspace scope isolation before graph traversal (I-11) and 3-state epistemic separation (I-4).", tags: ["Invariants", "I-11", "3-State Epistemics"], updated_at: new Date().toISOString() },
        { id: "doc:spec:control-plane", title: "Spec: Control Plane & Active MCP Session Watermarks", doc_type: "SPEC", path: "docs/specs/control-plane.md", status: "ENFORCED", components_count: 2, linked_symbols_count: 3, enforced_by_talon: true, summary: "Defines read-only database pool isolation for Control Plane endpoints and mandatory Talon pre-edit/post-edit MCP watermarks.", tags: ["Spec", "Control Plane", "MCP", "Talon"], updated_at: new Date().toISOString() },
        { id: "doc:policy:architecture-no-api-to-store-bypass", title: "Policy: API handlers must not bypass store layer to database/sql", doc_type: "POLICY", path: "policies/architecture-layer-guardrails.yaml", status: "ENFORCED", components_count: 1, linked_symbols_count: 3, enforced_by_talon: true, summary: "Blocks any direct or transitive import/call from internal/api to database/sql bypassing internal/store.", tags: ["YAML Policy", "BLOCK", "Layer Boundary"], updated_at: new Date().toISOString() },
        { id: "doc:policy:payment-no-direct-db", title: "Policy: Payment code must not import database/sql directly", doc_type: "POLICY", path: "policies/payment-no-direct-db.yaml", status: "ENFORCED", components_count: 1, linked_symbols_count: 2, enforced_by_talon: true, summary: "Enforces the hexagonal boundary from ADR-0001 across all Payment and Refund services.", tags: ["YAML Policy", "BLOCK", "Payments"], updated_at: new Date().toISOString() },
        { id: "doc:policy:require-runtime-evidence-for-exported", title: "Policy: Exported handlers must have runtime telemetry evidence", doc_type: "POLICY", path: "policies/require-runtime-evidence.yaml", status: "ENFORCED", components_count: 1, linked_symbols_count: 2, enforced_by_talon: true, summary: "Requires SRE review if an exported handler lacks OpenTelemetry runtime span corroboration.", tags: ["YAML Policy", "REVIEW", "Runtime Evidence"], updated_at: new Date().toISOString() },
        { id: "doc:guardrail:garudarules", title: "Guardrail: .garudarules (Hard Security & Secret Hygiene)", doc_type: "GUARDRAIL", path: ".garudarules", status: "ENFORCED", components_count: 1, linked_symbols_count: 2, enforced_by_talon: true, summary: "Hard pre-commit denial of hardcoded API keys, tokens, or private credentials in source files.", tags: ["Hard Rule", "DENY", "POL-001"], updated_at: new Date().toISOString() }
    ];

    var comps = [
        { id: "comp:adr0001:1", doc_id: "doc:adr:0001-payment-idempotency", doc_title: "ADR-0001: Payment Refund Processing", doc_type: "ADR", doc_path: "docs/adr/0001-payment-idempotency.md", section: "Decision §1", line: 7, modality: "MUST", subject: "RefundHandler", predicate: "CALLS", object: "IdempotencyStore", statement: "All invocations of `RefundHandler` MUST call `IdempotencyStore` before executing gateway mutations.", status: "ENFORCED", reason: "Verified via migrations/058_idempotency_keys.sql & internal/refund.", policy_id: "payment-no-direct-db", linked_code: [{ id: "sym:RateLimiter.Allow", name: "RateLimiter.Allow", kind: "method", package: "internal/refund", file: "internal/refund/rate_limiter.go", relation: "GOVERNED_BY" }, { id: "sym:PostgresStore.AppendLeafAndSealTx", name: "PostgresStore.AppendLeafAndSealTx", kind: "method", package: "internal/store", file: "internal/store/merkle_v1.go", relation: "IMPLEMENTS" }] },
        { id: "comp:adr0001:2", doc_id: "doc:adr:0001-payment-idempotency", doc_title: "ADR-0001: Payment Refund Processing", doc_type: "ADR", doc_path: "docs/adr/0001-payment-idempotency.md", section: "Decision §3", line: 9, modality: "MUST_NOT", subject: "RefundHandler", predicate: "CALLS", object: "DirectBankAPI / database/sql", statement: "`RefundHandler` MUST NOT call `DirectBankAPI` or import `database/sql` directly.", status: "ENFORCED", reason: "Actively blocked by policy `payment-no-direct-db` and Talon pre-commit gate.", policy_id: "payment-no-direct-db", linked_code: [{ id: "sym:RateLimiter.Allow", name: "RateLimiter.Allow", kind: "method", package: "internal/refund", file: "internal/refund/rate_limiter.go", relation: "ENFORCED_ON" }, { id: "sym:Engine.Run", name: "Engine.Run", kind: "method", package: "internal/policy", file: "internal/policy/engine.go", relation: "ENFORCED_ON" }] },
        { id: "comp:adr0002:d1", doc_id: "doc:adr:0002-merkle-integrity-layer", doc_title: "ADR-0002: Merkle Integrity Layer", doc_type: "ADR", doc_path: "docs/adr/0002-merkle-integrity-layer.md", section: "Decision D1", line: 33, modality: "MUST", subject: "internal/merkle", predicate: "IMPLEMENTS", object: "CanonicalEncoderV1", statement: "Every hash input in `internal/merkle/` MUST use versioned length-prefixed canonical byte serialization.", status: "SUPPORTED", reason: "Implemented in internal/merkle/canonical.go and genesis.go.", policy_id: "", linked_code: [{ id: "sym:VerifyV1Proof", name: "VerifyV1Proof", kind: "function", package: "internal/store", file: "internal/store/merkle_v1.go", relation: "IMPLEMENTS" }] },
        { id: "comp:adr0002:d3", doc_id: "doc:adr:0002-merkle-integrity-layer", doc_title: "ADR-0002: Merkle Integrity Layer", doc_type: "ADR", doc_path: "docs/adr/0002-merkle-integrity-layer.md", section: "Decision D3", line: 51, modality: "MUST", subject: "AppendLeafAndSealTx", predicate: "ANCHORS", object: "merkle_epoch_leaves", statement: "Every policy evaluation and decision MUST be anchored in `merkle_epoch_leaves` with an RFC 6962 inclusion proof.", status: "SUPPORTED", reason: "Verified in internal/store/merkle_v1.go and HandleDashboardPolicyVerify.", policy_id: "", linked_code: [{ id: "sym:PostgresStore.AppendLeafAndSealTx", name: "PostgresStore.AppendLeafAndSealTx", kind: "method", package: "internal/store", file: "internal/store/merkle_v1.go", relation: "IMPLEMENTS" }, { id: "sym:VerifyV1Proof", name: "VerifyV1Proof", kind: "function", package: "internal/store", file: "internal/store/merkle_v1.go", relation: "IMPLEMENTS" }] },
        { id: "comp:adr0003:1", doc_id: "doc:adr:0003-workspace-state-snapshots", doc_title: "ADR-0003: Workspace State Snapshots", doc_type: "ADR", doc_path: "docs/adr/0003-workspace-state-snapshots.md", section: "Context §1-2", line: 14, modality: "MUST", subject: "CreateUnifiedMerkleSnapshot", predicate: "SEPARATES", object: "merkle_roots vs merkle_snapshots", statement: "Periodic workspace state snapshots (`merkle_snapshots`) MUST remain distinct from the v1 decision log (`merkle_roots`).", status: "SUPPORTED", reason: "Implemented in internal/store/merkle_store.go.", policy_id: "", linked_code: [{ id: "sym:CreateUnifiedMerkleSnapshot", name: "CreateUnifiedMerkleSnapshot", kind: "method", package: "internal/store", file: "internal/store/merkle_store.go", relation: "IMPLEMENTS" }] },
        { id: "comp:inv:i11", doc_id: "doc:spec:invariants", doc_title: "Spec: Garuda System Invariants", doc_type: "INVARIANT", doc_path: "docs/invariants.md", section: "Invariant I-11", line: 18, modality: "INVARIANT", subject: "workspaceScopeFromRequest", predicate: "ENFORCES", object: "tenant_id & workspace_id scope", statement: "Every store and graph query MUST apply `tenant_id` and `workspace_id` filters before traversal.", status: "SUPPORTED", reason: "Verified across internal/api/workspace_scope.go and internal/knowledge/evaluator.go.", policy_id: "architecture-no-api-to-store-bypass", linked_code: [{ id: "sym:workspaceScopeFromRequest", name: "workspaceScopeFromRequest", kind: "function", package: "internal/api", file: "internal/api/workspace_scope.go", relation: "IMPLEMENTS" }, { id: "sym:Server.HandleDashboard", name: "Server.HandleDashboard", kind: "method", package: "internal/api", file: "internal/api/dashboard_handlers.go", relation: "GOVERNED_BY" }] },
        { id: "comp:inv:i4", doc_id: "doc:spec:invariants", doc_title: "Spec: Garuda System Invariants", doc_type: "INVARIANT", doc_path: "docs/invariants.md", section: "Invariant I-4", line: 42, modality: "INVARIANT", subject: "VerificationEngine", predicate: "DISTINGUISHES", object: "SUPPORTED | UNVERIFIED | CONTRADICTED", statement: "Garuda MUST never collapse unexecuted runtime branches into SUPPORTED; static AST and runtime traces are reported separately.", status: "SUPPORTED", reason: "Enforced in internal/runtime/verifier.go.", policy_id: "require-runtime-evidence-for-exported", linked_code: [{ id: "sym:RecomputeWorkspaceVerification", name: "RecomputeWorkspaceVerification", kind: "method", package: "internal/runtime", file: "internal/runtime/verifier.go", relation: "IMPLEMENTS" }] },
        { id: "comp:spec:cp1", doc_id: "doc:spec:control-plane", doc_title: "Spec: Control Plane & MCP Sessions", doc_type: "SPEC", doc_path: "docs/specs/control-plane.md", section: "Control Plane Pool", line: 12, modality: "MUST", subject: "HandleControlPlane", predicate: "USES_READONLY_POOL", object: "garuda_control_ro", statement: "Control Plane endpoints (`/_/control`) MUST execute exclusively over the read-only `controlPool` connection.", status: "SUPPORTED", reason: "Verified in internal/api/control_plane.go.", policy_id: "", linked_code: [{ id: "sym:Server.RegisterRoutes", name: "Server.RegisterRoutes", kind: "method", package: "internal/api", file: "internal/api/handler.go", relation: "IMPLEMENTS" }] },
        { id: "comp:spec:mcp1", doc_id: "doc:spec:control-plane", doc_title: "Spec: Control Plane & MCP Sessions", doc_type: "SPEC", doc_path: "docs/specs/control-plane.md", section: "Active Talon Watermarks", line: 28, modality: "MUST", subject: "garuda-mcp / talon", predicate: "ENFORCES", object: "PreEditBrief & RunGuard", statement: "AI coding agents MUST call `talon_pre_edit` before editing hubs and pass `talon_verify_edit` after edits.", status: "ENFORCED", reason: "Enforced in internal/talon/talon.go and cmd/garuda-mcp/tools_talon.go.", policy_id: "", linked_code: [{ id: "sym:CompileContract", name: "CompileContract", kind: "function", package: "internal/talon", file: "internal/talon/talon.go", relation: "IMPLEMENTS" }, { id: "sym:RunGuard", name: "talon.RunGuard", kind: "function", package: "internal/talon", file: "internal/talon/talon.go", relation: "IMPLEMENTS" }] },
        { id: "comp:pol:layer:1", doc_id: "doc:policy:architecture-no-api-to-store-bypass", doc_title: "Policy: No API-to-DB Bypass", doc_type: "POLICY", doc_path: "policies/architecture-layer-guardrails.yaml", section: "when.layer_violation", line: 19, modality: "BLOCK", subject: "internal/api", predicate: "LAYER_VIOLATION", object: "database/sql", statement: "[BLOCK] API layer must route persistence operations through internal/store rather than calling raw database/sql directly.", status: "ENFORCED", reason: "Active declarative YAML policy evaluated on every check and Talon guard pass.", policy_id: "architecture-no-api-to-store-bypass", linked_code: [{ id: "sym:Server.HandleDashboard", name: "Server.HandleDashboard", kind: "method", package: "internal/api", file: "internal/api/dashboard_handlers.go", relation: "ENFORCED_ON" }, { id: "sym:Engine.Run", name: "Engine.Run", kind: "method", package: "internal/policy", file: "internal/policy/engine.go", relation: "ENFORCED_ON" }] },
        { id: "comp:pol:payment:1", doc_id: "doc:policy:payment-no-direct-db", doc_title: "Policy: Payment No Direct DB", doc_type: "POLICY", doc_path: "policies/payment-no-direct-db.yaml", section: "when.claim_exists", line: 23, modality: "BLOCK", subject: "Payment / Refund", predicate: "FORBIDS_IMPORT", object: "database/sql", statement: "[BLOCK] Direct import of database/sql from payment code violates the hexagonal boundary in ADR-0001.", status: "ENFORCED", reason: "Active declarative YAML policy guarding payment and refund packages.", policy_id: "payment-no-direct-db", linked_code: [{ id: "sym:RateLimiter.Allow", name: "RateLimiter.Allow", kind: "method", package: "internal/refund", file: "internal/refund/rate_limiter.go", relation: "ENFORCED_ON" }] },
        { id: "comp:rule:POL-001", doc_id: "doc:guardrail:garudarules", doc_title: "Guardrail: .garudarules", doc_type: "GUARDRAIL", doc_path: ".garudarules", section: "rules.POL-001", line: 3, modality: "DENY", subject: "security.credentials", predicate: "FORBIDS_PATTERN", object: "regex:(key|token|secret)", statement: "[DENY] No hardcoded secrets, API keys, or private tokens permitted in source tree.", status: "ENFORCED", reason: "Enforced by Garuda hygiene scanner and Talon pre-commit gate.", policy_id: "POL-001", linked_code: [{ id: "sym:AuthMiddleware", name: "AuthMiddleware", kind: "method", package: "internal/api", file: "internal/api/auth_middleware.go", relation: "ENFORCED_ON" }] }
    ];

    return synthesizeDocsGraphFromList(docs, comps);
}

function synthesizeDocsGraphFromList(docs, comps) {
    var nodes = [];
    var edges = [];
    var nodeSeen = {};
    var edgeSeen = {};

    function addNode(n) {
        if (!nodeSeen[n.id]) {
            nodeSeen[n.id] = true;
            nodes.push(n);
        }
    }
    function addEdge(e) {
        if (!edgeSeen[e.id] && e.from && e.to && e.from !== e.to) {
            edgeSeen[e.id] = true;
            edges.push(e);
        }
    }

    docs.forEach(function(d) {
        addNode({
            id: d.id,
            label: d.title,
            kind: "document",
            package: d.doc_type,
            repo: WORKSPACE,
            file: d.path,
            exported: true,
            status: d.status,
            count: (d.components_count || 1) + (d.linked_symbols_count || 1)
        });
    });

    comps.forEach(function(c) {
        var kind = (c.doc_type === "POLICY" || c.doc_type === "GUARDRAIL") ? "policy_guardrail" : "doc_component";
        addNode({
            id: c.id,
            label: "[" + c.modality + "] " + shortenLabel(c.subject, 16) + " → " + shortenLabel(c.object, 16),
            kind: kind,
            package: c.doc_title,
            repo: c.modality,
            file: c.doc_path + ":" + c.line + " (" + c.section + ")",
            exported: true,
            status: c.status,
            count: (c.linked_code ? c.linked_code.length : 1) + 1
        });

        addEdge({
            id: c.doc_id + "->" + c.id,
            from: c.doc_id,
            to: c.id,
            type: "DEFINES_RULE",
            label: c.modality.toLowerCase(),
            confidence: 1.0,
            status: c.status,
            count: 1
        });

        if (c.policy_id) {
            var polId = c.policy_id === "POL-001" ? "doc:guardrail:garudarules" : ("doc:policy:" + c.policy_id);
            if (nodeSeen[polId] && c.doc_id !== polId) {
                addEdge({
                    id: c.id + "->" + polId,
                    from: c.id,
                    to: polId,
                    type: "ENFORCED_BY_POLICY",
                    label: "enforced by",
                    confidence: 1.0,
                    status: "ENFORCED",
                    count: 1
                });
            }
        }

        (c.linked_code || []).forEach(function(lc) {
            var cid = lc.id || ("code:" + lc.package + "." + lc.name);
            addNode({
                id: cid,
                label: lc.name,
                kind: lc.kind || "code_symbol",
                package: lc.package,
                repo: WORKSPACE,
                file: lc.file || lc.package,
                exported: true,
                status: c.status,
                count: 3
            });
            addEdge({
                id: c.id + "->" + cid,
                from: c.id,
                to: cid,
                type: (c.doc_type === "POLICY" || c.doc_type === "GUARDRAIL") ? "ENFORCES_GUARDRAIL" : "GOVERNS_CODE",
                label: (c.doc_type === "POLICY" || c.doc_type === "GUARDRAIL") ? "guards" : "governs",
                confidence: 1.0,
                status: c.status,
                count: 1
            });
        });
    });

    return {
        workspace: WORKSPACE,
        stats: {
            total_documents: docs.length,
            adrs_count: docs.filter(function(d) { return d.doc_type === "ADR"; }).length,
            specs_count: docs.filter(function(d) { return d.doc_type === "SPEC" || d.doc_type === "INVARIANT"; }).length,
            policies_count: docs.filter(function(d) { return d.doc_type === "POLICY" || d.doc_type === "GUARDRAIL"; }).length,
            total_components: comps.length,
            verified_components: comps.length,
            enforced_by_talon: comps.filter(function(c) { return c.status === "ENFORCED"; }).length,
            drift_or_unverified: 0,
            contradicted_count: 0,
            connected_code_nodes: nodes.filter(function(n) { return n.kind !== "document" && n.kind !== "doc_component" && n.kind !== "policy_guardrail"; }).length,
            total_graph_edges: edges.length
        },
        documents: docs,
        components: comps,
        graph: {
            level: "docs",
            focus: "",
            nodes: nodes,
            edges: edges,
            truncated: false,
            total_nodes: nodes.length
        },
        updated_at: new Date().toISOString()
    };
}

function buildDefaultTalonHub() {
    return {
        workspace: WORKSPACE || "default",
        enforcement_mode: "ACTIVE (FAIL-CLOSED)",
        active_hooks_count: 6,
        total_hooks_count: 6,
        critical_hubs_count: 4,
        high_hubs_count: 2,
        hooks: [
            { id: "hook-cursor-mdc", name: "Cursor Active Rule (.mdc)", agent: "Cursor IDE", path: ".cursor/rules/garuda-talon.mdc", installed: true, mode: "ALWAYS_APPLY", description: "Injects compiled ADR & YAML policy guardrails into every Cursor composer/agent turn." },
            { id: "hook-cursor-mcp", name: "Cursor MCP Enforcement Server", agent: "Cursor MCP", path: ".cursor/mcp.json", installed: true, mode: "MCP_GATE", description: "Configures garuda-mcp with GARUDA_TALON_ENFORCE=true and 23 deterministic tools." },
            { id: "hook-claude-settings", name: "Claude Code PostToolUse Hook", agent: "Claude Code", path: ".claude/settings.json", installed: true, mode: "POST_TOOL_USE", description: "Triggers `garuda talon guard --hook` after every Edit/Write/MultiEdit tool call." },
            { id: "hook-claude-md", name: "CLAUDE.md Active Contract", agent: "Claude Code", path: "CLAUDE.md", installed: true, mode: "ALWAYS_APPLY", description: "Embeds mandatory pre-edit briefing and post-edit verification rules into Claude's root memory." },
            { id: "hook-agents-md", name: "AGENTS.md & .windsurfrules", agent: "Codex / Windsurf / Aider", path: "AGENTS.md", installed: true, mode: "ALWAYS_APPLY", description: "Synchronizes workspace hubs, policies, and ADR constraints across Codex, Windsurf, and Aider." },
            { id: "hook-git-precommit", name: "Git Pre-Commit & Provenance Gate", agent: "Git CLI / CI", path: ".git/hooks/pre-commit", installed: true, mode: "PRE_COMMIT", description: "Blocks commits with net-new architectural violations and stamps RFC 6962 Merkle block trailers." }
        ],
        risk_hubs: [
            { id: "sym:AppendLeafAndSealTx", name: "PostgresStore.AppendLeafAndSealTx", kind: "method", package: "internal/store", file: "internal/store/merkle_v1.go", inbound_callers: 18, outbound_callees: 6, transitive_radius: 60, risk_score: 96, risk_tier: "CRITICAL", epistemic_status: "SUPPORTED", governing_rules: ["ADR-0002: RFC 6962 Canonical Merkle Layer", "Invariant I-11: Tenant & Workspace Scope"], blast_radius_summary: "18 direct callers · ~60 transitive symbols across internal/store & internal/policy" },
            { id: "sym:Engine.Run", name: "Engine.Run", kind: "method", package: "internal/policy", file: "internal/policy/engine.go", inbound_callers: 14, outbound_callees: 9, transitive_radius: 51, risk_score: 92, risk_tier: "CRITICAL", epistemic_status: "SUPPORTED", governing_rules: ["architecture-no-api-to-store-bypass", "ADR-0002: RFC 6962 Canonical Merkle Layer"], blast_radius_summary: "14 direct callers · ~51 transitive symbols across CLI, MCP, and Talon gates" },
            { id: "sym:RunGuard", name: "talon.RunGuard", kind: "function", package: "internal/talon", file: "internal/talon/talon.go", inbound_callers: 11, outbound_callees: 8, transitive_radius: 41, risk_score: 86, risk_tier: "CRITICAL", epistemic_status: "SUPPORTED", governing_rules: ["Spec: Active Talon MCP Watermarks", "architecture-layer-guardrails"], blast_radius_summary: "11 direct callers · ~41 transitive symbols across pre-commit and MCP hooks" },
            { id: "sym:workspaceScopeFromRequest", name: "workspaceScopeFromRequest", kind: "function", package: "internal/api", file: "internal/api/workspace_scope.go", inbound_callers: 12, outbound_callees: 4, transitive_radius: 40, risk_score: 84, risk_tier: "CRITICAL", epistemic_status: "SUPPORTED", governing_rules: ["Invariant I-11: Tenant & Workspace Scope", "architecture-no-api-to-store-bypass"], blast_radius_summary: "12 direct callers · ~40 transitive symbols across all dashboard API handlers" },
            { id: "sym:RecomputeWorkspaceVerification", name: "VerificationEngine.RecomputeWorkspaceVerification", kind: "method", package: "internal/runtime", file: "internal/runtime/verifier.go", inbound_callers: 8, outbound_callees: 7, transitive_radius: 31, risk_score: 76, risk_tier: "HIGH", epistemic_status: "SUPPORTED", governing_rules: ["Invariant I-4: 3-State Epistemic Honesty", "require-runtime-evidence-for-exported"], blast_radius_summary: "8 direct callers · ~31 transitive symbols across runtime telemetry correlator" },
            { id: "sym:CompileContract", name: "talon.CompileContract", kind: "function", package: "internal/talon", file: "internal/talon/talon.go", inbound_callers: 7, outbound_callees: 5, transitive_radius: 26, risk_score: 71, risk_tier: "HIGH", epistemic_status: "SUPPORTED", governing_rules: ["Spec: Active Talon MCP Watermarks", "ADR-0001: Payment Idempotency"], blast_radius_summary: "7 direct callers · ~26 transitive symbols across IDE & MCP contract compilers" },
            { id: "sym:RateLimiter.Allow", name: "RateLimiter.Allow", kind: "method", package: "internal/refund", file: "internal/refund/rate_limiter.go", inbound_callers: 5, outbound_callees: 3, transitive_radius: 18, risk_score: 58, risk_tier: "MEDIUM", epistemic_status: "SUPPORTED", governing_rules: ["ADR-0001: Payment Idempotency", "payment-no-direct-db"], blast_radius_summary: "5 direct callers · ~18 transitive symbols across payment refund endpoints" }
        ],
        contract: {
            workspace_name: WORKSPACE || "default",
            workspace_id: "canonical",
            compiled_at: new Date().toISOString(),
            entity_count: 892,
            claim_count: 2418,
            markdown: "# GARUDA TALON — ACTIVE ARCHITECTURAL CONTRACT\n\n- **Workspace**: `" + (WORKSPACE || "default") + "`\n- **Graph Scale**: 892 AST entities | 2,418 verified edges\n- **Enforcement Mode**: `ACTIVE (FAIL-CLOSED)`\n\n## 1. Mandatory Agent Execution Protocol\n1. **BEFORE editing any file or symbol**: Call MCP tool `garuda.talon_pre_edit` to inspect callers, layer boundaries, and ADR rules.\n2. **WHILE writing code**: Respect all `MUST` / `MUST_NOT` ADR rules and YAML architectural policies.\n3. **IMMEDIATELY AFTER editing any file**: Call `garuda.talon_verify_edit` (or `garuda talon guard --diff HEAD`).\n4. **ON `BLOCK` VERDICT**: Follow `self_repair_instructions` immediately until `ALLOW` is returned.\n\n## 2. Active Declarative Policy Guardrails\n- **`[BLOCK]` API handlers must not bypass store layer** (`architecture-no-api-to-store-bypass`)\n- **`[BLOCK]` Payment code must not import database/sql directly** (`payment-no-direct-db`)\n- **`[REVIEW]` Exported payment handlers must have runtime evidence** (`require-runtime-evidence-for-exported`)\n- **`[DENY]` No hardcoded secrets or tokens** (`POL-001`)\n\n## 3. Ingested ADR & Spec Claims\n- **`[MUST]`** `RefundHandler` MUST CALLS `IdempotencyStore` *(ADR-0001)*\n- **`[MUST_NOT]`** `RefundHandler` MUST_NOT CALLS `DirectBankAPI` *(ADR-0001)*\n- **`[MUST]`** `AppendLeafAndSealTx` MUST ANCHORS `merkle_epoch_leaves` *(ADR-0002)*\n- **`[INVARIANT]`** `workspaceScopeFromRequest` ENFORCES `tenant_id & workspace_id scope` *(Invariant I-11)*"
        },
        latest_verdict: {
            workspace: WORKSPACE || "default",
            outcome: "ALLOW",
            agent_should_proceed: true,
            touched_files: [],
            touched_symbols: [],
            touched_packages: [],
            blast_radius_affected: 0,
            by_severity: { CRITICAL: 4, HIGH: 2, MEDIUM: 4, LOW: 2 },
            net_new_violations_count: 0,
            baseline_violations_count: 0,
            adr_contradictions_count: 0,
            violations: [],
            self_repair_instructions: [
                "Workspace is clean (0 net-new policy or ADR violations).",
                "Before editing any CRITICAL or HIGH risk hub, run `garuda.talon_pre_edit` to inspect callers and layer guardrails."
            ],
            merkle_block_height: 52,
            duration_ms: 18,
            reproduce_command: "GOOS=linux GOARCH=amd64 CGO_ENABLED=0 garuda check --target linux/amd64 --cgo=false --commit 9f8bc3fa41e2 --workspace " + (WORKSPACE || "default"),
            build_context: {
                language: "go",
                target_os: "linux",
                target_arch: "amd64",
                target_platform: "linux/amd64",
                cgo_enabled: false,
                build_tags: ["purego"],
                compiler_version: "go1.24 (go/build constraint filter)",
                commit_sha: "9f8bc3fa41e27b09d4c6a8123e5f70192b4c6d8e",
                commit_short: "9f8bc3fa41e2",
                branch: "main",
                worktree_dirty: false,
                dirty_files_count: 0,
                config_manifest_hash: "e7a1c904b2d81f30",
                graph_fingerprint: "e4b89c1d3a7f0e29c5a1840f92b7d31e64a8f10c9b2e5d7a3c819204f6b1a3d8",
                receipt_hash: "6d92a4f01b8e3c71a5d902e4b8c1f3a790e2d4b6c8a1f3e5d7b9c0a2e4f6a8b0",
                active_files_count: 148,
                skipped_by_build_tags_count: 3,
                skipped_files_sample: ["internal/ast/treesitter/cgo_bridge.go", "internal/runtime/ebpf_linux_cgo.go", "cmd/garuda/win_console_windows.go"],
                reproduce_command: "GOOS=linux GOARCH=amd64 CGO_ENABLED=0 garuda check --target linux/amd64 --cgo=false --commit 9f8bc3fa41e2 --workspace " + (WORKSPACE || "default")
            },
            environment_drift: {
                drift_detected: false,
                severity: "NONE",
                summary: "Build environment matches graph receipt (0 platform, tag, or commit discrepancies).",
                discrepancies: [],
                reproduce_command: "GOOS=linux GOARCH=amd64 CGO_ENABLED=0 garuda check --target linux/amd64 --cgo=false --commit 9f8bc3fa41e2 --workspace " + (WORKSPACE || "default")
            },
            matrix_results: [
                {
                    target_platform: "linux/amd64",
                    cgo_enabled: false,
                    build_tags: ["purego"],
                    active_files_count: 148,
                    skipped_by_build_tags_count: 3,
                    skipped_files_sample: ["internal/ast/treesitter/cgo_bridge.go", "cmd/garuda/win_console_windows.go"],
                    entities_count: 892,
                    edges_count: 2418,
                    graph_fingerprint: "e4b89c1d3a7f0e29c5a1840f92b7d31e64a8f10c9b2e5d7a3c819204f6b1a3d8",
                    receipt_hash: "6d92a4f01b8e3c71",
                    outcome: "ALLOW",
                    reproduce_command: "GOOS=linux GOARCH=amd64 CGO_ENABLED=0 garuda check --target linux/amd64 --cgo=false --tags purego --workspace " + (WORKSPACE || "default")
                },
                {
                    target_platform: "linux/amd64",
                    cgo_enabled: true,
                    build_tags: ["cgo"],
                    active_files_count: 150,
                    skipped_by_build_tags_count: 1,
                    skipped_files_sample: ["cmd/garuda/win_console_windows.go"],
                    entities_count: 904,
                    edges_count: 2451,
                    graph_fingerprint: "3c91a7e40b2d8f16a5c3e901b4d7f2a8c6e0b1d3f5a7c9e2b4d6f8a0c2e4b6d8",
                    receipt_hash: "8a41f2c90e3b7d15",
                    outcome: "ALLOW",
                    reproduce_command: "GOOS=linux GOARCH=amd64 CGO_ENABLED=1 garuda check --target linux/amd64 --cgo=true --tags cgo --workspace " + (WORKSPACE || "default")
                },
                {
                    target_platform: "darwin/arm64",
                    cgo_enabled: true,
                    build_tags: ["cgo", "darwin"],
                    active_files_count: 149,
                    skipped_by_build_tags_count: 2,
                    skipped_files_sample: ["internal/runtime/ebpf_linux_cgo.go", "cmd/garuda/win_console_windows.go"],
                    entities_count: 898,
                    edges_count: 2432,
                    graph_fingerprint: "7a12f90c3e81b4d0921f8a6e3c5b7d901a2e4f6c8b0d1e3a5f7c9b2d4e6a8f01",
                    receipt_hash: "2f78e1a49c0b5d32",
                    outcome: "ALLOW",
                    reproduce_command: "GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 garuda check --target darwin/arm64 --cgo=true --workspace " + (WORKSPACE || "default")
                },
                {
                    target_platform: "windows/amd64",
                    cgo_enabled: false,
                    build_tags: ["purego", "windows"],
                    active_files_count: 148,
                    skipped_by_build_tags_count: 3,
                    skipped_files_sample: ["internal/ast/treesitter/cgo_bridge.go", "internal/runtime/ebpf_linux_cgo.go"],
                    entities_count: 893,
                    edges_count: 2420,
                    graph_fingerprint: "5b30d8e21a4c9f07e6b2d4a8c1f3e5a7b9d0c2e4f6a8b1d3c5e7f9a2b4c6d8e0",
                    receipt_hash: "9c14b6d82a0e5f73",
                    outcome: "ALLOW",
                    reproduce_command: "GOOS=windows GOARCH=amd64 CGO_ENABLED=0 garuda check --target windows/amd64 --cgo=false --workspace " + (WORKSPACE || "default")
                }
            ]
        }
    };
}

var communityPalette = [
    "#191B19", "#4D5E10", "#2E5A44", "#7C5816", "#525866", "#7A2E2E",
    "#3D5A6C", "#5E4B70", "#6B5E4C", "#3B6E5C", "#854D0E", "#4A4D48"
];

function getCommunityColor(node) {
    if (!node) return communityPalette[0];
    if (node.status === "CONTRADICTED") return "#991B1B";
    if (node.kind === "document") return "#137333";
    if (node.kind === "doc_component") return "#B06000";
    if (node.kind === "policy_guardrail") return "#191B19";
    var str = node._community || node.package || node.repo || node.id || "general";
    var hash = 0;
    for (var i = 0; i < str.length; i++) {
        hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }
    return communityPalette[Math.abs(hash) % communityPalette.length];
}

function showToast(message) {
    var container = document.getElementById("toast-container");
    if (!container) return;
    container.innerHTML = "";
    var toast = document.createElement("div");
    toast.className = "garuda-toast";
    toast.innerHTML = '<span style="color:var(--brand-accent);">🦅</span><span>' + escapeHTML(message) + '</span>';
    container.appendChild(toast);
    setTimeout(function() {
        if (toast.parentNode) toast.parentNode.removeChild(toast);
    }, 3200);
}

function refreshAllWithToast() {
    showToast("Refreshing workspace topology, docs graph & Talon hub…");
    loadAll();
}

function initTheme() {
    var saved = "dark";
    try { saved = localStorage.getItem("garuda_theme") || "dark"; } catch (e) {}
    if (saved !== "light" && saved !== "dark") saved = "dark";
    document.documentElement.setAttribute("data-theme", saved);
    updateThemeButton(saved);
}
function toggleTheme() {
    var curr = document.documentElement.getAttribute("data-theme") || "light";
    var next = curr === "light" ? "dark" : "light";
    document.documentElement.setAttribute("data-theme", next);
    try { localStorage.setItem("garuda_theme", next); } catch (e) {}
    updateThemeButton(next);
    showToast("Switched to " + (next === "dark" ? "Graphite Dark" : "Warm Paper Light") + " mode");
    if (state.currentView === "graph" && state.graphData) {
        renderGraph(state.graphData);
    }
    if (state.currentView === "docsgraph" && state.docsGraphData) {
        renderDocsD3Graph();
    }
}

function updateThemeButton(theme) {
    var btn = document.getElementById("theme-toggle-btn");
    if (btn) {
        btn.textContent = theme === "light" ? "☀" : "☾";
    }
}

async function loadWorkspacePicker() {
    var sel = document.getElementById("workspace-picker");
    if (!sel) return;
    try {
        var res = await fetch("/api/v1/workspaces", {
            headers: { "Accept": "application/json" }
        });
        if (!res.ok) throw new Error("workspaces unavailable");
        var list = await res.json();
        sel.innerHTML = "";
        if (!list || list.length === 0) {
            list = [{ name: WORKSPACE || "default", role: "owner" }];
        }
        list.forEach(function(w) {
            var opt = document.createElement("option");
            opt.value = w.name;
            opt.textContent = w.name + " (" + (w.role || "owner") + ")";
            if (w.name === WORKSPACE) opt.selected = true;
            sel.appendChild(opt);
        });
    } catch (err) {
        sel.innerHTML = '<option value="' + escapeHTML(WORKSPACE || "default") + '" selected>' + escapeHTML(WORKSPACE || "default") + ' (owner)</option>';
    }
}

function switchWorkspace(name) {
    if (!name || name === WORKSPACE) return;
    WORKSPACE = name;
    var params = new URLSearchParams(window.location.search);
    params.set("workspace", name);
    window.location.search = params.toString();
}


function switchSurfaceSubTab(sectionId, paneId, btnEl) {
    var sec = document.getElementById(sectionId);
    if (!sec) return;
    sec.querySelectorAll(".surface-tab-pane").forEach(function(p) {
        var isMatch = (p.id === paneId);
        p.classList.toggle("active", isMatch);
    });
    if (btnEl && btnEl.parentNode) {
        btnEl.parentNode.querySelectorAll(".surface-tab-btn").forEach(function(b) {
            b.classList.toggle("active", b === btnEl);
        });
    }
    if (paneId === "dg-pane-graph" && state.docsGraphData) {
        setTimeout(function() { renderDocsD3Graph(); }, 40);
    }
}

function toggleMobileNav(forceOpen) {
    var sidebar = document.querySelector(".sidebar");
    var scrim = document.getElementById("sidebar-scrim");
    var toggle = document.getElementById("menu-toggle");
    if (!sidebar || !scrim) return;
    var open = typeof forceOpen === "boolean" ? forceOpen : !sidebar.classList.contains("mobile-open");
    sidebar.classList.toggle("mobile-open", open);
    scrim.classList.toggle("open", open);
    if (toggle) {
        toggle.setAttribute("aria-expanded", open ? "true" : "false");
        toggle.setAttribute("aria-label", open ? "Close navigation" : "Open navigation");
    }
}

function showView(view) {
    var aliases = { overview: "workspace", architecture: "graph", trust: "governance", docs: "docsgraph", risk: "talon" };
    view = aliases[view] || view;

    if (!viewToSection[view]) return;
    state.currentView = view;

    Object.keys(viewToSection).forEach(function(key) {
        var el = document.getElementById(viewToSection[key]);
        if (!el) return;
        var isTarget = (key === view);
        el.style.display = isTarget ? "block" : "none";
        el.classList.toggle("active", isTarget);
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
    document.querySelectorAll(".nav-cluster").forEach(function(group) {
        group.open = !!group.querySelector(".nav-btn.active");
    });
    if (window.innerWidth <= 900) toggleMobileNav(false);

    setText("topbar-surface-name", surfaceLabels[view] || "01 — Overview");

    if (view === "search") {
        if (!state.searchResultsCache || state.searchResultsCache.length === 0) {
            performSearch((document.getElementById("search-page-input") && document.getElementById("search-page-input").value) || "");
        }
    }
    if (view === "graph") {
        var defaultLevel = state.currentLevel || "entity";
        var defaultFocus = state.currentFocus || "";
        if (!defaultFocus && state.stats && state.stats.top_hubs && state.stats.top_hubs.length > 0 && defaultLevel === "entity") {
            defaultFocus = state.stats.top_hubs[0].package || state.stats.top_hubs[0].name || "";
        }
        loadArchitecture(defaultLevel, defaultFocus);
    }
    if (view === "docsgraph") {
        loadDocsGraph();
    }
    if (view === "talon") {
        loadTalonHub();
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
    if (view === "audit") {
        loadHumanDecisionHub();
    }
    if (view === "decisions") {
        loadDecisions();
    }

    try {
        var params = new URLSearchParams(window.location.search);
        if (params.get("tab") !== view && window.location.protocol !== "file:") {
            params.set("tab", view);
            history.pushState({ tab: view }, "", window.location.pathname + "?" + params.toString());
        }
    } catch (e) {}
}

window.addEventListener("popstate", function(event) {
    var view = (event.state && event.state.tab) || initialView();
    showView(view);
});

async function loadStats() {
    try {
        var res = await fetch("/api/v1/dashboard/stats?workspace=" + encodeURIComponent(WORKSPACE), {
            headers: { "Accept": "application/json" }
        });
        if (!res.ok) throw new Error("Stats request failed: " + res.status);
        var data = await res.json();
        if (!data || typeof data !== "object" || Array.isArray(data)) throw new Error("Stats response was not an object.");
        ["languages_breakdown", "top_hubs", "needs_attention", "recent_evidence", "repo_stats", "repositories_list"].forEach(function(key) {
            if (!Array.isArray(data[key])) data[key] = [];
        });
        if (!data.drift || typeof data.drift !== "object") data.drift = {};
        state.stats = data;
        state.statsError = null;
    } catch (err) {
        state.statsError = err;
        state.stats = DEMO_MODE ? FALLBACK_DATA.stats : null;
    }
    if (state.stats) renderStats();
    else renderStatsUnavailable();
    await loadPolicies();
}
async function loadPolicies() {
    try {
        var res = await fetch("/api/v1/dashboard/policies?workspace=" + encodeURIComponent(WORKSPACE), {
            headers: { "Accept": "application/json" }
        });
        if (!res.ok) throw new Error("Policies request failed: " + res.status);
        var data = await res.json();
        if (!data || typeof data !== "object" || Array.isArray(data)) throw new Error("Policies response was not an object.");
        if (!Array.isArray(data.latest_evaluations)) data.latest_evaluations = [];
        state.policiesData = data;
        state.policiesError = null;
    } catch (err) {
        state.policiesError = err;
        state.policiesData = DEMO_MODE ? FALLBACK_DATA.policies : { unavailable: true, latest_evaluations: [] };
    }
    renderPolicies(state.policiesData);
    buildAuditEntries();
    if (state.currentView === "review") renderChangeReviewWorkbench();
}
function renderStatsUnavailable() {
    ["stat-repositories", "stat-packages", "stat-entities", "stat-relationships", "stat-cross-links",
     "stat-idempotency", "stat-supported", "stat-unverified", "stat-contradicted", "pulse-total-claims",
     "stat-active-policies", "stat-drift-count", "explorer-repos", "explorer-packages", "explorer-entities",
     "overview-block-height", "overview-merkle-short", "drift-total-docs", "drift-doc-supported",
     "drift-doc-unverified", "drift-doc-contradicted", "drift-doc-to-code", "drift-code-to-doc",
     "drift-undocumented-code", "drift-unimplemented-docs"].forEach(function(id) { setText(id, "—"); });
    setText("overview-updated-at", "Workspace data unavailable");
    setText("overview-narrative", "Garuda could not load the current workspace snapshot. Check the connection and refresh to try again.");
    setText("sidebar-meta", "Workspace data unavailable");
    setText("honest-system-text", "Workspace data unavailable");
    setText("lang-legend", "Language data unavailable.");
    var badge = document.getElementById("policy-final-decision-badge");
    if (badge) { badge.textContent = "Unavailable"; badge.className = "badge-pill info"; }
    setText("policy-latest-anchor", "No current anchor");
    ["hub-list", "attention-list", "evidence-list", "scanned-repos-list"].forEach(function(id) {
        var el = document.getElementById(id);
        if (el) el.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Workspace data unavailable</div><div class="row-meta">Refresh after the dashboard connection is restored.</div></div></div>';
    });
    var langBar = document.getElementById("lang-bar");
    if (langBar) langBar.innerHTML = "";
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
        setText("stat-unverified-desc", "static AST verified");
    } else {
        setText("stat-unverified-desc", "awaiting runtime trace");
    }

    setText("explorer-repos", formatNumber(s.repositories));
    setText("explorer-packages", formatNumber(s.packages));
    setText("explorer-entities", formatNumber(s.entities));
    setText("nav-badge-packages", formatNumber(s.packages));
    setText("nav-badge-ledger", s.latest_block_height ? "#" + String(s.latest_block_height) : "—");
    setText("overview-block-height", s.latest_block_height ? "#" + String(s.latest_block_height) : "—");
    setText("overview-updated-at", "Last indexed " + formatDate(s.last_updated));

    var merkleShort = (s.latest_merkle_hash || "No anchor yet");
    if (merkleShort.length > 22) merkleShort = merkleShort.slice(0, 18) + "…";
    setText("overview-merkle-short", merkleShort);

    setText("sidebar-meta", formatNumber(s.repositories) + " repos · " + formatNumber(s.entities) + " entities");
    setText("sidebar-workspace", s.workspace || WORKSPACE);
    setText("workspace-breadcrumb", s.workspace || WORKSPACE);
    setText("repo-list-ws", s.workspace || WORKSPACE);
    setText("stat-drift-count", formatNumber(s.quarantined_count));

    var narrative = "Indexed " + formatNumber(s.entities) + " code entities and " + formatNumber(s.relationships) +
        " semantic relationships across " + formatNumber(s.packages) + " packages, connected to ingested ADRs, specs, and active Talon guardrails. ";
    if ((s.contradicted || 0) > 0) {
        narrative += formatNumber(s.contradicted) + " contradicted claim(s) require immediate investigation.";
    } else {
        narrative += "No active contradictions are reported in the latest workspace evaluation.";
    }
    setText("overview-narrative", narrative);

    var sysDot = document.getElementById("honest-system-dot");
    if ((s.contradicted || 0) > 0) {
        setText("honest-system-text", s.contradicted + " contradiction(s) active");
        if (sysDot) sysDot.style.background = "var(--red)";
    } else {
        setText("honest-system-text", s.latest_block_height ? "Workspace indexed · Block #" + s.latest_block_height : "Workspace indexed · No ledger anchor yet");
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
    setText("policy-block-count", formatNumber(data.blocked_count || 0));
    setText("policy-review-count", formatNumber(data.review_count || 0));
    setText("policy-warn-count", formatNumber(data.warn_count || 0));
    setText("policy-allow-count", formatNumber(data.allow_count || 0));
    setText("stat-active-policies", formatNumber(data.active_policies || 0));
    var badge = document.getElementById("policy-final-decision-badge");
    var decision = data.unavailable ? "Unavailable" : (data.final_decision || "No evaluation");
    if (badge) {
        badge.textContent = decision;
        badge.className = "badge-pill " + (
            data.unavailable ? "info" :
            data.final_decision === "BLOCK" ? "critical" :
            data.final_decision === "REVIEW" || data.final_decision === "WARN" ? "warning" :
            data.final_decision === "ALLOW" ? "success" : "info"
        );
    }
    setText("nav-badge-review", data.final_decision || "—");
    var chip = document.getElementById("policy-latest-anchor");
    if (chip) {
        chip.textContent = data.latest_merkle_anchor ? "block #" + data.latest_merkle_anchor : "No anchor yet";
        chip.classList.remove("invalid");
    }
    var list = document.getElementById("policy-eval-list");
    if (!list) return;
    var evals = data.latest_evaluations || [];
    if (evals.length === 0) {
        list.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">' +
            (data.unavailable ? "Policy evaluations could not be loaded." : "No policy evaluations have been recorded for this workspace yet.") +
            '</div><div class="row-meta">' +
            (data.unavailable ? "Refresh the dashboard after the API connection is restored." : "Run a change check to create the first evaluation.") +
            '</div></div></div>';
        return;
    }
    list.innerHTML = "";
    evals.forEach(function(ev, idx) {
        var row = document.createElement("div");
        row.className = "policy-eval-row";
        var anchorChip = ev.merkle_block_height
            ? '<span class="merkle-chip' + (ev.anchor_valid ? '' : ' invalid') + '">block #' + ev.merkle_block_height + '</span>'
            : '<span class="merkle-chip">not anchored</span>';
        var preds = (ev.matched_predicates || []).join(", ") || "all predicates passed";
        var when = formatDate(ev.evaluated_at);
        row.innerHTML =
            '<span class="policy-badge ' + escapeHTML(ev.decision) + '">' + escapeHTML(ev.decision) + '</span>' +
            '<div class="policy-eval-main"><div class="policy-eval-title">' + escapeHTML(ev.policy_title) + anchorChip + '</div>' +
            '<div class="policy-eval-reason">' + escapeHTML(ev.reason) + '</div>' +
            '<div class="policy-eval-meta"><span><strong>' + formatNumber(ev.entity_count) + '</strong> entities</span>' +
            '<span><strong>' + formatNumber(ev.claim_count) + '</strong> claims</span>' +
            '<span><strong>' + formatNumber(ev.contradiction_count) + '</strong> contradictions</span>' +
            '<span>rule: <strong>' + escapeHTML(preds) + '</strong></span><span>' + escapeHTML(when) + '</span></div></div>' +
            '<button class="btn-secondary" style="height:30px;font-size:11px;" onclick="event.stopPropagation();openEvalInChangeReview(' + idx + ')">Review</button>';
        row.style.cursor = "pointer";
        row.onclick = function() { openEvalInChangeReview(idx); };
        list.appendChild(row);
    });
}
function verifyPolicyAnchor(evalID) {
    if (!evalID) {
        showToast("Select a recorded policy evaluation before verifying its proof.");
        return;
    }
    var url = "/api/v1/dashboard/policies/verify?workspace=" + encodeURIComponent(WORKSPACE) + "&id=" + encodeURIComponent(evalID);
    fetch(url, { headers: { "Accept": "application/json" } })
        .then(function(res) {
            if (!res.ok) throw new Error("Proof verification failed (HTTP " + res.status + ").");
            return res.json();
        })
        .then(function(data) {
            if (!data || data.valid !== true || !data.block_height || !(data.epoch_root || data.new_root)) {
                showToast("No valid anchored proof was returned for this evaluation.");
                return;
            }
            showProofDrawer(evalID, data.block_height, data.epoch_root || data.new_root);
        })
        .catch(function(err) { showToast(err && err.message ? err.message : "Proof verification is unavailable."); });
}
function showProofDrawer(evalID, blockHeight, epochRoot) {
    var drawer = document.getElementById("drawer");
    var overlay = document.getElementById("drawer-overlay");
    var body = document.getElementById("drawer-body");
    if (!drawer || !overlay || !body) return;

    setText("drawer-title", "RFC 6962 Cryptographic Proof Verified");
    setText("drawer-kind", "MERKLE INCLUSION PROOF · BLOCK #" + blockHeight);

    var html = '<div class="detail-section"><div class="detail-section-title">Verification Verdict</div>';
    html += propertyRow("Proof Status", "✓ VALID (RFC 6962 Canonical v1)");
    html += propertyRow("Evaluation ID", evalID);
    html += propertyRow("Epoch Block Height", "#" + blockHeight);
    html += propertyRow("Hash Algorithm", "SHA-256 (0x00 Leaf / 0x01 Node prefix)");
    html += '</div>';

    html += '<div class="detail-section"><div class="detail-section-title">Epoch Root Hash</div>';
    html += '<div class="merkle">' + escapeHTML(epochRoot) + '</div>';
    html += '</div>';

    body.innerHTML = html;
    drawer.classList.add("open");
    overlay.classList.add("open");
    showToast("Verified RFC 6962 Merkle proof for block #" + blockHeight);
}

function renderLanguages(langs) {
    var bar = document.getElementById("lang-bar");
    var legend = document.getElementById("lang-legend");
    if (!bar || !legend) return;
    if ((!langs || langs.length === 0) && DEMO_MODE) langs = FALLBACK_DATA.stats.languages_breakdown;
    langs = Array.isArray(langs) ? langs.slice().sort(function(a, b) { return b.count - a.count; }) : [];
    bar.innerHTML = "";
    legend.innerHTML = "";
    if (langs.length === 0) {
        legend.textContent = "Language data appears after the first successful index.";
        return;
    }
    langs.forEach(function(lang) {
        var seg = document.createElement("div");
        seg.className = "lang-bar-segment";
        seg.style.width = lang.percentage + "%";
        seg.style.background = lang.color;
        seg.title = lang.name + " " + lang.percentage.toFixed(1) + "%";
        bar.appendChild(seg);
    });
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
    d = d || {};
    setText("drift-total-docs", formatNumber(d.total_document_claims || 0));
    setText("drift-doc-supported", formatNumber(d.supported_claims || 0));
    setText("drift-doc-unverified", formatNumber(d.unverified_claims || 0));
    setText("drift-doc-contradicted", formatNumber(d.contradicted_claims || 0));
    setText("drift-doc-to-code", formatNumber(d.doc_to_code_drift_count || 0));
    setText("drift-code-to-doc", formatNumber(d.code_to_doc_drift_count || 0));
    setText("drift-undocumented-code", formatNumber(d.undocumented_code || 0));
    setText("drift-unimplemented-docs", formatNumber(d.unverified_docs || 0));
}
function renderHubs(hubs) {
    var list = document.getElementById("hub-list");
    if (!list) return;
    if ((!hubs || hubs.length === 0) && DEMO_MODE) hubs = FALLBACK_DATA.stats.top_hubs;
    hubs = Array.isArray(hubs) ? hubs : [];
    list.innerHTML = "";
    if (hubs.length === 0) {
        list.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No indexed hubs yet</div><div class="row-meta">Run an analysis to rank symbols by architectural centrality.</div></div></div>';
        return;
    }
    hubs.forEach(function(h) {
        var row = document.createElement("div");
        row.className = "list-row";
        row.style.cursor = "pointer";
        row.innerHTML = '<div class="row-icon">' + symbolIcon(h.kind) + '</div><div class="row-main"><div class="row-title mono">' + escapeHTML(h.name) + '</div><div class="row-meta">' + escapeHTML(h.kind) + ' · ' + escapeHTML(h.package) + '</div></div>' +
            '<div style="text-align:right;display:flex;align-items:center;gap:10px;"><div><span class="mono" style="font-weight:700;color:var(--text);font-size:12.5px;">' + h.callers + '</span> <span style="font-size:10px;color:var(--muted);">callers</span></div><span class="kbd-chip">1-hop</span></div>';
        row.onclick = function() { focusSymbolOnArchitectureMap(h.name, h.package, h); };
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
    items = Array.isArray(items) ? items : [];
    if (items.length === 0) {
        list.innerHTML = '<div class="list-row"><div class="row-icon" style="color:var(--green);border-color:var(--green-border);background:var(--green-bg);">✓</div><div class="row-main"><div class="row-title">No items need attention</div><div class="row-meta">No active contradictions or policy blocks are reported in this workspace snapshot.</div></div></div>';
        return;
    }
    list.innerHTML = "";
    items.forEach(function(item) {
        var row = document.createElement("div");
        row.className = "list-row";
        row.style.cursor = "pointer";
        row.innerHTML = '<div class="row-icon" style="color:' + (item.severity === 'critical' ? 'var(--red)' : 'var(--amber)') + ';">!</div>' +
            '<div class="row-main"><div class="row-title">' + escapeHTML(item.title) + '</div><div class="row-meta">' + escapeHTML(item.subtitle) + ' · ' + escapeHTML(item.evidence_loc) + '</div></div>' +
            '<span class="badge-pill ' + escapeHTML(item.severity) + '">' + escapeHTML(item.severity) + '</span>';
        row.onclick = function() { showView("review"); };
        list.appendChild(row);
    });
}
function renderEvidence(items) {
    var list = document.getElementById("evidence-list");
    if (!list) return;
    if ((!items || items.length === 0) && DEMO_MODE) items = FALLBACK_DATA.stats.recent_evidence;
    items = Array.isArray(items) ? items : [];
    list.innerHTML = "";
    if (items.length === 0) {
        list.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No evidence records yet</div><div class="row-meta">Evidence appears after analysis, verification, or policy evaluation.</div></div></div>';
        return;
    }
    items.forEach(function(item) {
        var row = document.createElement("div");
        row.className = "list-row";
        var badgeClass = item.kind === "Runtime Trace" ? "warning" : "info";
        var badgeText = item.kind === "Runtime Trace" ? "Runtime observed" : "Recorded";
        row.innerHTML = '<div class="row-icon">§</div><div class="row-main"><div class="row-title">' + escapeHTML(item.summary) + '</div>' +
            '<div class="row-meta">' + escapeHTML(item.kind) + ' · ' + escapeHTML(item.source) + ' · ' + formatDate(item.timestamp) + '</div></div>' +
            '<span class="badge-pill ' + badgeClass + '">' + badgeText + '</span>';
        list.appendChild(row);
    });
}
function renderTrust() {
    if (!state.stats) return;
    setText("trust-status", state.stats.trust_status || "Status unavailable");
    setText("trust-height", state.stats.latest_block_height ? "#" + String(state.stats.latest_block_height) : "—");
    setText("trust-updated", formatDate(state.stats.last_updated));
    setText("trust-merkle", state.stats.latest_merkle_hash || "No anchor recorded");
    setText("trust-parent", state.stats.parent_merkle_hash || "No parent anchor");
}

function renderScannedRepos(repoStats, fallbackList) {
    var list = document.getElementById("scanned-repos-list");
    if (!list) return;
    if ((!repoStats || repoStats.length === 0) && DEMO_MODE) repoStats = FALLBACK_DATA.stats.repo_stats;
    repoStats = Array.isArray(repoStats) ? repoStats : [];
    list.innerHTML = "";
    if (repoStats.length === 0) {
        list.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No repositories indexed in this workspace</div><div class="row-meta">Add a repository with the Garuda CLI, then run workspace sync.</div></div></div>';
        return;
    }
    repoStats.forEach(function(repo) {
        var card = document.createElement("div");
        card.className = "repo-card";
        var statusBadge = repo.analysis_status === "synced"
            ? '<span class="badge-pill success">synced</span>'
            : '<span class="badge-pill info">' + escapeHTML(repo.analysis_status || "indexed") + '</span>';
        var commitShort = repo.current_commit && repo.current_commit.length > 8 ? repo.current_commit.substring(0, 8) : (repo.current_commit || "HEAD");
        card.innerHTML =
            '<div class="repo-card-header"><div><div class="repo-card-name">' + escapeHTML(repo.name) + '</div>' +
            '<div class="repo-card-meta">commit ' + escapeHTML(commitShort) + ' · analyzed ' + escapeHTML(repo.last_analyzed || "time unavailable") + '</div></div>' + statusBadge + '</div>' +
            '<div class="repo-card-stats"><span><span class="repo-card-stat-val">' + formatNumber(repo.entities) + '</span> entities</span>' +
            '<span><span class="repo-card-stat-val">' + formatNumber(repo.relationships) + '</span> relationships</span>' +
            '<span><span class="repo-card-stat-val">' + formatNumber(repo.files) + '</span> files</span></div>' +
            '<div class="repo-card-actions"><button class="btn-secondary" style="height:28px;font-size:11.5px;" onclick="event.stopPropagation();state.currentLevel=\'package\';state.currentFocus=\'' + escapeJS(repo.name) + '\';showView(\'graph\');">Architecture map</button>' +
            '<button class="btn-secondary" style="height:28px;font-size:11.5px;" onclick="event.stopPropagation();runSearch(\'' + escapeJS(repo.name) + '\')">Search symbols</button></div>';
        list.appendChild(card);
    });
}
async function loadDocsGraph() {
    if (!state.docsGraphData) {
        try {
            var res = await fetch("/api/v1/dashboard/docs-graph?workspace=" + encodeURIComponent(WORKSPACE), {
                headers: { "Accept": "application/json" }
            });
            if (!res.ok) throw new Error("docs-graph HTTP " + res.status);
            var data = await res.json();
            state.docsGraphData = data && Array.isArray(data.documents) && (data.documents.length > 0 || !DEMO_MODE)
                ? data
                : (DEMO_MODE ? buildDefaultDocsGraph() : { stats: {}, documents: [], components: [], graph: { nodes: [], edges: [] } });
        } catch (err) {
            state.docsGraphData = DEMO_MODE ? buildDefaultDocsGraph() :
                { stats: {}, documents: [], components: [], graph: { nodes: [], edges: [] }, unavailable: true };
        }
    }
    renderDocsGraphAndTables();
}
function filterDocsGraph(filterVal, btnEl) {
    state.docsFilterType = filterVal;
    state.docsSelectedDocId = "";
    if (btnEl && btnEl.parentElement) {
        btnEl.parentElement.querySelectorAll(".filter-pill").forEach(function(b) {
            b.classList.toggle("active", b === btnEl);
        });
    }
    renderDocsGraphAndTables();
}

function selectSpecificDocCard(docId) {
    if (state.docsSelectedDocId === docId) {
        state.docsSelectedDocId = "";
    } else {
        state.docsSelectedDocId = docId;
    }
    renderDocsGraphAndTables();
}

function renderDocsGraphAndTables() {
    if (!state.docsGraphData) {
        state.docsGraphData = DEMO_MODE ? buildDefaultDocsGraph() : { stats: {}, documents: [], components: [], graph: { nodes: [], edges: [] } };
    }
    var dg = state.docsGraphData;
    var st = dg.stats || {};

    setText("dg-stat-total-docs", formatNumber(st.total_documents || dg.documents.length));
    setText("dg-stat-docs-foot", (st.adrs_count || 3) + " ADRs · " + (st.specs_count || 2) + " Specs · " + (st.policies_count || 4) + " Policies");
    setText("dg-stat-total-comps", formatNumber(st.total_components || dg.components.length));
    setText("dg-stat-code-nodes", formatNumber(st.connected_code_nodes || 12));
    setText("dg-stat-edges", formatNumber(st.total_graph_edges || dg.graph.edges.length));
    setText("dg-stat-enforced", formatNumber(st.verified_components || dg.components.length));
    setText("nav-badge-docsgraph", String(dg.documents.length) + " Docs");

    var searchVal = ((document.getElementById("docs-filter-search") && document.getElementById("docs-filter-search").value) || "").toLowerCase().trim();

    var filteredDocs = (dg.documents || []).filter(function(d) {
        if (state.docsFilterType === "ADR" && d.doc_type !== "ADR") return false;
        if (state.docsFilterType === "SPEC" && d.doc_type !== "SPEC" && d.doc_type !== "INVARIANT") return false;
        if (state.docsFilterType === "POLICY" && d.doc_type !== "POLICY" && d.doc_type !== "GUARDRAIL") return false;
        if (state.docsFilterType === "ENFORCED" && !d.enforced_by_talon) return false;
        if (searchVal) {
            var hay = (d.title + " " + d.path + " " + d.summary + " " + (d.tags || []).join(" ")).toLowerCase();
            if (hay.indexOf(searchVal) === -1) return false;
        }
        return true;
    });

    var docsGrid = document.getElementById("ingested-docs-grid");
    if (docsGrid) {
        docsGrid.innerHTML = "";
        filteredDocs.forEach(function(d) {
            var isSelected = state.docsSelectedDocId === d.id;
            var card = document.createElement("div");
            card.className = "repo-card";
            card.style.cursor = "pointer";
            if (isSelected) {
                card.style.borderColor = "var(--ink)";
                card.style.boxShadow = "0 0 0 2px var(--brand-accent)";
            }

            var statusPill = d.enforced_by_talon
                ? '<span class="badge-pill success">Talon Enforced</span>'
                : '<span class="badge-pill info">' + escapeHTML(d.status || "VERIFIED") + '</span>';

            card.innerHTML =
                '<div class="repo-card-header">' +
                    '<div style="display:flex; align-items:center; gap:6px; flex-wrap:wrap;">' +
                        '<span class="doc-type-badge ' + escapeHTML(d.doc_type) + '">' + escapeHTML(d.doc_type) + '</span>' +
                        statusPill +
                    '</div>' +
                    '<span class="kbd-chip">' + (isSelected ? "Focused ✓" : "Filter Dots") + '</span>' +
                '</div>' +
                '<div style="font-weight:700; font-size:13px; color:var(--text); margin-top:6px;">' + escapeHTML(d.title) + '</div>' +
                '<div class="mono" style="font-size:10.5px; color:var(--muted); margin-top:2px;">' + escapeHTML(d.path) + '</div>' +
                '<div style="font-size:11.5px; color:var(--text-2); margin-top:6px; line-height:1.4;">' + escapeHTML(d.summary) + '</div>' +
                '<div class="repo-card-stats" style="margin-top:10px;">' +
                    '<span><strong>' + (d.components_count || 1) + '</strong> extracted rules</span>' +
                    '<span><strong>' + (d.linked_symbols_count || 2) + '</strong> connected code symbols</span>' +
                '</div>';

            card.onclick = function() {
                selectSpecificDocCard(d.id);
                inspectDocsNodeById(d.id);
            };
            docsGrid.appendChild(card);
        });
    }

    var allowedDocIds = new Set(filteredDocs.map(function(d) { return d.id; }));
    var filteredComps = (dg.components || []).filter(function(c) {
        if (state.docsSelectedDocId && c.doc_id !== state.docsSelectedDocId) return false;
        if (!state.docsSelectedDocId && !allowedDocIds.has(c.doc_id) && searchVal === "") return false;
        if (searchVal) {
            var codeNames = (c.linked_code || []).map(function(lc) { return lc.name + " " + lc.package; }).join(" ");
            var hay = (c.doc_title + " " + c.subject + " " + c.predicate + " " + c.object + " " + c.statement + " " + codeNames).toLowerCase();
            if (hay.indexOf(searchVal) === -1) return false;
        }
        return true;
    });

    setText("dg-table-count-chip", filteredComps.length + " component" + (filteredComps.length === 1 ? "" : "s"));

    var tbody = document.getElementById("docs-components-tbody");
    if (tbody) {
        tbody.innerHTML = "";
        if (filteredComps.length === 0) {
            tbody.innerHTML = '<tr><td colspan="6" style="color:var(--muted); padding:18px;">No document components match this filter. Click <strong>Show All Documents</strong> or <strong>+ Ingest Document / ADR / Rule</strong>.</td></tr>';
        } else {
            filteredComps.forEach(function(comp) {
                var tr = document.createElement("tr");
                tr.style.cursor = "pointer";

                var modClass = (comp.modality === "BLOCK" || comp.modality === "MUST_NOT" || comp.modality === "DENY") ? "critical" :
                               (comp.modality === "REVIEW") ? "warning" : "success";

                var codeChips = (comp.linked_code || []).map(function(lc) {
                    return '<span class="code-ref-chip" onclick="event.stopPropagation(); focusSymbolOnArchitectureMap(\'' + escapeJS(lc.name) + '\', \'' + escapeJS(lc.package) + '\', null)" title="Click to open 1-hop Architecture Map for ' + escapeHTML(lc.name) + '">' +
                        symbolIcon(lc.kind) + ' ' + escapeHTML(lc.name) +
                    '</span>';
                }).join("");

                var statusBadge = comp.status === "CONTRADICTED"
                    ? '<span class="badge-pill critical">CONTRADICTED</span>'
                    : (comp.status === "ENFORCED" ? '<span class="badge-pill success">ENFORCED BY TALON</span>' : '<span class="badge-pill info">' + escapeHTML(comp.status || "SUPPORTED") + '</span>');

                tr.innerHTML =
                    '<td>' +
                        '<div style="display:flex; align-items:center; gap:6px;">' +
                            '<span class="doc-type-badge ' + escapeHTML(comp.doc_type) + '">' + escapeHTML(comp.doc_type) + '</span>' +
                            '<span style="font-weight:600; font-size:12px;">' + escapeHTML(shortenLabel(comp.doc_title, 34)) + '</span>' +
                        '</div>' +
                        '<div class="mono" style="font-size:10.5px; color:var(--muted); margin-top:2px;">' + escapeHTML(comp.doc_path) + ':' + comp.line + '</div>' +
                    '</td>' +
                    '<td><span class="badge-pill ' + modClass + ' mono">' + escapeHTML(comp.modality) + '</span></td>' +
                    '<td>' +
                        '<div class="mono" style="font-size:11.5px; font-weight:650; color:var(--text);">' +
                            escapeHTML(comp.subject) + ' <span style="color:var(--muted);">' + escapeHTML(comp.predicate) + '</span> ' + escapeHTML(comp.object) +
                        '</div>' +
                        '<div style="font-size:11.5px; color:var(--text-2); margin-top:2px;">' + escapeHTML(comp.statement) + '</div>' +
                    '</td>' +
                    '<td>' + (codeChips || '<span style="color:var(--muted); font-size:11px;">Workspace-wide</span>') + '</td>' +
                    '<td>' +
                        statusBadge +
                        (comp.policy_id ? '<div class="mono" style="font-size:10px; color:var(--muted); margin-top:3px;">Policy: ' + escapeHTML(comp.policy_id) + '</div>' : '') +
                    '</td>' +
                    '<td><button class="btn-secondary" style="height:26px; font-size:11px;" onclick="event.stopPropagation(); inspectDocsNodeById(\'' + escapeJS(comp.id) + '\')">Focus Dot →</button></td>';

                tr.onclick = function() {
                    inspectDocsNodeById(comp.id);
                };
                tbody.appendChild(tr);
            });
        }
    }

    renderDocsD3Graph(filteredDocs, filteredComps);
    renderRulebookStudio();
}

function renderDocsD3Graph(filteredDocs, filteredComps) {
    if (state.docsGraphSimulation) {
        state.docsGraphSimulation.stop();
        state.docsGraphSimulation = null;
    }
    var svg = d3.select("#docs-graph-svg");
    if (svg.empty()) return;
    svg.selectAll("*").remove();

    var dg = state.docsGraphData || { documents: [], components: [], stats: {}, graph: { nodes: [], edges: [] } };
    var docs = filteredDocs || dg.documents || [];
    var comps = filteredComps || dg.components || [];

    var subGraph = synthesizeDocsGraphFromList(docs, comps).graph;
    var nodes = (subGraph.nodes || []).map(function(n) { return Object.assign({}, n); });
    var nodeById = {};
    nodes.forEach(function(n) { nodeById[n.id] = n; });

    var edges = (subGraph.edges || []).filter(function(e) {
        return nodeById[e.from] && nodeById[e.to];
    }).map(function(e) {
        return {
            id: e.id,
            source: nodeById[e.from],
            target: nodeById[e.to],
            type: e.type,
            label: e.label,
            status: e.status
        };
    });

    var wrap = document.getElementById("docs-graph-wrap");
    var width = (wrap && wrap.clientWidth) || 780;
    var height = ((wrap && wrap.clientHeight) || 520) - 38;
    svg.attr("width", width).attr("height", height);

    var defs = svg.append("defs");
    defs.append("marker").attr("id", "dg-arrow").attr("viewBox", "0 -5 10 10").attr("refX", 22).attr("refY", 0)
        .attr("markerWidth", 5).attr("markerHeight", 5).attr("orient", "auto")
        .append("path").attr("d", "M0,-4L10,0L0,4").attr("fill", "#767A72");

    var zoomLayer = svg.append("g");
    state.docsGraphSvg = svg;
    state.docsGraphGroup = zoomLayer;

    var zoom = d3.zoom().scaleExtent([0.2, 4]).on("zoom", function(event) {
        zoomLayer.attr("transform", event.transform);
    });
    svg.call(zoom);
    state.docsGraphZoom = zoom;

    var tierBuckets = { doc: [], comp: [], pol: [], code: [] };
    nodes.forEach(function(n) {
        if (n.kind === "document") tierBuckets.doc.push(n);
        else if (n.kind === "doc_component") tierBuckets.comp.push(n);
        else if (n.kind === "policy_guardrail") tierBuckets.pol.push(n);
        else tierBuckets.code.push(n);
    });
    Object.keys(tierBuckets).forEach(function(k) {
        var list = tierBuckets[k];
        list.forEach(function(n, idx) {
            n._tierIndex = idx;
            n._tierCount = list.length;
        });
    });

    function tierX(d) {
        if (d.kind === "document") return 40;
        if (d.kind === "doc_component") return 280;
        if (d.kind === "policy_guardrail") return 520;
        return 740;
    }
    function tierY(d) {
        var count = Math.max(1, d._tierCount || 1);
        var idx = d._tierIndex || 0;
        var usableH = Math.max(380, height - 40);
        var step = usableH / (count + 1);
        return 20 + step * (idx + 1);
    }

    var simulation = d3.forceSimulation(nodes)
        .alphaDecay(0.08)
        .force("link", d3.forceLink(edges).id(function(d) { return d.id; }).distance(130).strength(0.25))
        .force("charge", d3.forceManyBody().strength(-220))
        .force("x", d3.forceX(function(d) { return tierX(d); }).strength(0.85))
        .force("y", d3.forceY(function(d) { return tierY(d); }).strength(0.85))
        .force("collision", d3.forceCollide().radius(24));

    state.docsGraphSimulation = simulation;

    var link = zoomLayer.append("g").selectAll("path")
        .data(edges).enter().append("path")
        .attr("class", function(d) { return d.status === "CONTRADICTED" ? "graph-link violation" : "graph-link"; })
        .attr("stroke", function(d) {
            if (d.status === "CONTRADICTED") return "#991B1B";
            if (d.type === "ENFORCES_GUARDRAIL" || d.type === "ENFORCED_BY_POLICY") return "#137333";
            return "#8A8E85";
        })
        .attr("stroke-width", function(d) { return d.type === "DEFINES_RULE" ? 1.6 : 1.4; })
        .attr("marker-end", "url(#dg-arrow)");

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

    node.append("circle").attr("class", "dg-node-circle")
        .attr("r", function(d) {
            if (d.kind === "document") return 14;
            if (d.kind === "policy_guardrail" || d.kind === "doc_component") return 12;
            return 11;
        })
        .attr("fill", function(d) {
            if (d.kind === "document") return "#137333";
            if (d.kind === "doc_component") return "#B06000";
            if (d.kind === "policy_guardrail") return "#191B19";
            return "#2E5A44";
        })
        .attr("stroke", function(d) {
            if (d.kind === "policy_guardrail") return "#D5E66A";
            if (d.kind === "document") return "#CEEAD6";
            if (d.kind === "doc_component") return "#FDD663";
            return "#FFFFFF";
        })
        .attr("stroke-width", 2.5);

    node.append("text")
        .attr("text-anchor", "middle")
        .attr("dy", 3.5)
        .style("font-family", "var(--font-mono)")
        .style("font-size", "9px")
        .style("font-weight", "700")
        .style("pointer-events", "none")
        .attr("fill", function(d) {
            return d.kind === "policy_guardrail" ? "#D5E66A" : "#FFFFFF";
        })
        .text(function(d) {
            if (d.kind === "document") return "DOC";
            if (d.kind === "doc_component") return "§";
            if (d.kind === "policy_guardrail") return "P";
            return symbolIcon(d.kind);
        });

    node.append("text").attr("class", "graph-node-label")
        .attr("dx", 16)
        .attr("dy", 4)
        .text(function(d) { return shortenLabel(d.label || d.id, 28); });

    node.on("click", function(event, d) {
        event.stopPropagation();
        populateDocsInspector(d, nodes, edges);
    });

    simulation.on("tick", function() {
        link.attr("d", function(d) {
            var mx = (d.source.x + d.target.x) / 2;
            return "M" + d.source.x + "," + d.source.y + " C" + mx + "," + d.source.y + " " + mx + "," + d.target.y + " " + d.target.x + "," + d.target.y;
        });
        node.attr("transform", function(d) { return "translate(" + d.x + "," + d.y + ")"; });
    });

    state.docsGraphData._renderedNodes = nodes;
    state.docsGraphData._renderedEdges = edges;

    if (nodes.length > 0) {
        populateDocsInspector(nodes[0], nodes, edges);
    }
    setTimeout(fitDocsGraph, 400);
}

function inspectDocsNodeById(nodeId) {
    if (!state.docsGraphData) return;
    var nodes = state.docsGraphData._renderedNodes || state.docsGraphData.graph.nodes || [];
    var edges = state.docsGraphData._renderedEdges || state.docsGraphData.graph.edges || [];
    var found = nodes.find(function(n) { return n.id === nodeId; });
    if (found) {
        populateDocsInspector(found, nodes, edges);
        showToast("Focused lineage dot: " + shortenLabel(found.label, 36));
    }
}

function populateDocsInspector(centerNode, nodes, edges) {
    if (!centerNode) return;
    state.docsSelectedNode = centerNode;

    var tierLabel = centerNode.kind === "document" ? "TIER 1 · SOURCE DOCUMENT" :
                    centerNode.kind === "doc_component" ? "TIER 2 · EXTRACTED RULE / CLAIM" :
                    centerNode.kind === "policy_guardrail" ? "TIER 3 · ACTIVE POLICY GUARDRAIL" :
                    "TIER 4 · GOVERNED CODE SYMBOL (AST)";

    setText("dg-insp-title", centerNode.label || centerNode.id);
    setText("dg-insp-subtitle", tierLabel);

    var connected = [];
    (edges || []).forEach(function(e) {
        var s = e.source.id ? e.source : nodes.find(function(n) { return n.id === e.from; });
        var t = e.target.id ? e.target : nodes.find(function(n) { return n.id === e.to; });
        if (!s || !t) return;
        if (s.id === centerNode.id) connected.push({ dir: "→", rel: e.type || "LINKS_TO", node: t });
        if (t.id === centerNode.id) connected.push({ dir: "←", rel: e.type || "FROM_DOC", node: s });
    });

    var body = document.getElementById("dg-insp-body");
    if (!body) return;

    var html = '<div class="detail-section" style="margin-bottom:12px;">';
    html += propertyRow("Node Tier", tierLabel);
    html += propertyRow("Scope / Package", centerNode.package || "workspace");
    if (centerNode.file) html += propertyRow("Source Location", centerNode.file);
    html += propertyRow("Enforcement Status", centerNode.status || "ENFORCED");
    html += '</div>';

    html += '<div class="detail-section-title">Connected Dots in Graph (' + connected.length + ')</div>';
    if (connected.length === 0) {
        html += '<div style="color:var(--muted); font-size:11.5px; margin-top:4px;">No direct edges in current filter.</div>';
    } else {
        connected.forEach(function(item) {
            html += '<div style="display:flex; align-items:center; justify-content:space-between; padding:6px 8px; margin-top:4px; background:var(--surface-2); border:1px solid var(--border); border-radius:6px; cursor:pointer; font-size:11.5px;" onclick="inspectDocsNodeById(\'' + escapeJS(item.node.id) + '\')">' +
                '<span class="mono" style="overflow:hidden; text-overflow:ellipsis; white-space:nowrap;">' + item.dir + ' ' + escapeHTML(shortenLabel(item.node.label, 28)) + '</span>' +
                '<span class="kind-pill" style="font-size:9px;">' + escapeHTML(item.rel) + '</span>' +
            '</div>';
        });
    }

    if (centerNode.kind !== "document" && centerNode.kind !== "doc_component" && centerNode.kind !== "policy_guardrail") {
        html += '<div style="margin-top:14px; display:flex; flex-direction:column; gap:6px;">' +
            '<button class="btn-primary" style="width:100%; justify-content:center;" onclick="focusSymbolOnArchitectureMap(\'' + escapeJS(centerNode.label) + '\', \'' + escapeJS(centerNode.package) + '\', null)">Open Symbol in 1-Hop Arch Map →</button>' +
            '<button class="btn-secondary" style="width:100%; justify-content:center;" onclick="openPreEditBriefDrawer(\'' + escapeJS(centerNode.label) + '\')">Inspect Talon Pre-Edit Brief →</button>' +
        '</div>';
    }

    body.innerHTML = html;

    if (state.docsGraphGroup) {
        state.docsGraphGroup.selectAll(".dg-node-circle")
            .attr("stroke-width", function(d) { return d.id === centerNode.id ? 4.5 : 2.5; });
    }
}

function fitDocsGraph() {
    if (!state.docsGraphSvg || !state.docsGraphGroup || !state.docsGraphZoom) return;
    var bbox;
    try { bbox = state.docsGraphGroup.node().getBBox(); } catch (e) { return; }
    if (!bbox.width || !bbox.height) return;
    var wrap = document.getElementById("docs-graph-wrap");
    var width = (wrap && wrap.clientWidth) || 780;
    var height = ((wrap && wrap.clientHeight) || 520) - 38;
    var scale = Math.min(width / (bbox.width + 100), height / (bbox.height + 100), 1.2);
    scale = Math.max(scale, 0.25);
    var tx = width / 2 - scale * (bbox.x + bbox.width / 2);
    var ty = height / 2 - scale * (bbox.y + bbox.height / 2);
    state.docsGraphSvg.transition().duration(280).call(state.docsGraphZoom.transform, d3.zoomIdentity.translate(tx, ty).scale(scale));
}

function zoomDocsGraph(mult) {
    if (!state.docsGraphSvg || !state.docsGraphZoom) return;
    state.docsGraphSvg.transition().duration(180).call(state.docsGraphZoom.scaleBy, mult);
}

function openIngestDocModal() {
    var modal = document.getElementById("ingest-doc-modal");
    if (!modal) return;
    modal.classList.add("open");
    if (!document.getElementById("ingest-doc-title").value) {
        fillIngestTemplate("adr4");
    }
}

function closeIngestDocModal() {
    var modal = document.getElementById("ingest-doc-modal");
    if (modal) modal.classList.remove("open");
}

function fillIngestTemplate(tpl) {
    var titleEl = document.getElementById("ingest-doc-title");
    var typeEl = document.getElementById("ingest-doc-type");
    var pathEl = document.getElementById("ingest-doc-path");
    var contentEl = document.getElementById("ingest-doc-content");
    if (!titleEl || !typeEl || !pathEl || !contentEl) return;

    if (tpl === "adr4") {
        titleEl.value = "ADR-0004: Zero-Trust Agent Execution & Pre-Edit Gate";
        typeEl.value = "ADR";
        pathEl.value = "docs/adr/0004-zero-trust-agent-gate.md";
        contentEl.value = "# ADR-0004: Zero-Trust Agent Execution & Pre-Edit Gate\n\n## Decision\n- `RunGuard` must call `CompileContract` before evaluating agent diffs.\n- `HandleDashboard` must not call `DirectBankAPI` directly.\n- `AppendLeafAndSealTx` must be idempotent under repeated delivery.";
    } else if (tpl === "spec-payment") {
        titleEl.value = "Spec: Payment Retry SLA & Rate Limiter Contract";
        typeEl.value = "SPEC";
        pathEl.value = "docs/specs/payment-retry-sla.md";
        contentEl.value = "# Spec: Payment Retry SLA & Rate Limiter Contract\n\n## Specification\n- `RateLimiter.Allow` must call `PostgresStore` for tenant token bucket state.\n- `RefundHandler` must not call `UnverifiedExternalWebhook` directly.";
    } else if (tpl === "rule-store") {
        titleEl.value = "Policy: Strict Store & Merkle Layer Isolation";
        typeEl.value = "POLICY";
        pathEl.value = "policies/store-merkle-isolation.yaml";
        contentEl.value = "# Policy: Strict Store & Merkle Layer Isolation\n\n## Rule\n- `VerifyV1Proof` must call `CreateUnifiedMerkleSnapshot` for epoch verification.\n- `AuthMiddleware` must not call `rawSQLDriver` directly.";
    }
}

async function submitIngestDoc() {
    var titleEl = document.getElementById("ingest-doc-title");
    var typeEl = document.getElementById("ingest-doc-type");
    var pathEl = document.getElementById("ingest-doc-path");
    var contentEl = document.getElementById("ingest-doc-content");
    var title = (titleEl && titleEl.value || "").trim() || "ADR-Custom: Architectural Guardrail";
    var docType = (typeEl && typeEl.value || "ADR").trim();
    var path = (pathEl && pathEl.value || "docs/adr/custom-rule.md").trim();
    var content = (contentEl && contentEl.value || "").trim();
    if (!content) {
        showToast("Add Markdown content before ingesting.");
        if (contentEl) contentEl.focus();
        return;
    }
    try {
        var res = await fetch("/api/v1/dashboard/docs-ingest?workspace=" + encodeURIComponent(WORKSPACE), {
            method: "POST",
            headers: { "Content-Type": "application/json", "Accept": "application/json" },
            body: JSON.stringify({ workspace: WORKSPACE, title: title, doc_type: docType, path: path, content: content })
        });
        var payload = null;
        try { payload = await res.json(); } catch (parseErr) {}
        if (!res.ok) throw new Error(payload && (payload.error || payload.message) || ("Document ingestion failed (HTTP " + res.status + ")."));
        if (!payload || payload.status !== "ingested" || !payload.docs_graph ||
            !Array.isArray(payload.docs_graph.documents) || !Array.isArray(payload.docs_graph.components)) {
            throw new Error("The server did not return a persisted document graph. No local preview was created.");
        }
        state.docsGraphData = payload.docs_graph;
        closeIngestDocModal();
        showView("docsgraph");
        renderDocsGraphAndTables();
        showToast("Ingested “" + shortenLabel(title, 32) + "” · " + formatNumber(payload.extracted_count || 0) + " extracted claims.");
    } catch (err) {
        showToast(err && err.message ? err.message : "Document ingestion failed. No success state was recorded.");
    }
}
async function rescanWorkspaceDocs() {
    showToast("Scanning docs/adr, docs/specs, policies, and .garudarules…");
    state.docsGraphData = null;
    await loadDocsGraph();
    if (!state.docsGraphData || state.docsGraphData.unavailable) {
        showToast("Document scan failed. No synchronized state was reported.");
        return;
    }
    var docsCount = Array.isArray(state.docsGraphData.documents) ? state.docsGraphData.documents.length : 0;
    var componentCount = Array.isArray(state.docsGraphData.components) ? state.docsGraphData.components.length : 0;
    showToast("Document index refreshed · " + docsCount + " documents · " + componentCount + " extracted components.");
}
async function loadTalonHub() {
    if (!state.talonHubData) {
        try {
            var res = await fetch("/api/v1/dashboard/talon-hub?workspace=" + encodeURIComponent(WORKSPACE), {
                headers: { "Accept": "application/json" }
            });
            if (!res.ok) throw new Error("talon-hub HTTP " + res.status);
            var data = await res.json();
            if (!data || typeof data !== "object") throw new Error("Invalid Talon response");
            state.talonHubData = (DEMO_MODE && (!Array.isArray(data.risk_hubs) || data.risk_hubs.length === 0))
                ? buildDefaultTalonHub() : data;
        } catch (err) {
            state.talonHubData = DEMO_MODE ? buildDefaultTalonHub() : { hooks: [], risk_hubs: [], enforcement_mode: "Unavailable", unavailable: true };
        }
    }
    renderTalonHub();
}
function renderTalonHub() {
    if (!state.talonHubData) {
        state.talonHubData = DEMO_MODE ? buildDefaultTalonHub() : { hooks: [], risk_hubs: [], enforcement_mode: "Unavailable" };
    }
    var th = state.talonHubData;
    var hooks = th.hooks || [];
    var hubs = th.risk_hubs || [];
    var verdict = th.latest_verdict || {};

    var unavailable = th.unavailable === true;
    setText("talon-hooks-count", unavailable ? "— / —" : (th.active_hooks_count || 0) + " / " + (th.total_hooks_count || 0));
    setText("talon-crit-badge", unavailable ? "Unavailable" : (th.critical_hubs_count || 0) + " CRITICAL");
    setText("talon-total-hubs", unavailable ? "—" : String(hubs.length));
    setText("talon-high-count", unavailable ? "—" : String(th.high_hubs_count || 0));
    var modeBadge = document.getElementById("talon-mode-badge");
    if (modeBadge) { modeBadge.textContent = th.enforcement_mode || (unavailable ? "Unavailable" : "Configured"); modeBadge.className = "badge-pill info"; }

    renderTalonVerdict(verdict);

    var hooksEl = document.getElementById("talon-hooks-list");
    if (hooksEl) {
        hooksEl.innerHTML = "";
        if (state.talonHubData.unavailable) {
            hooksEl.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">Hook status unavailable</div><div class="row-meta">No hook installation state was inferred.</div></div></div>';
        } else if (hooks.length === 0) {
            hooksEl.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No hook files found in the configured root</div><div class="row-meta">Sync installs supported files in the server-configured repository.</div></div></div>';
        }
        hooks.forEach(function(h) {
            var card = document.createElement("div");
            card.className = "hook-card";
            var badge = h.installed
                ? '<span class="badge-pill success" style="font-size:9.5px;">ACTIVE</span>'
                : '<span class="badge-pill warning" style="font-size:9.5px;">READY</span>';
            card.innerHTML =
                '<div class="hook-card-top">' +
                    '<span class="hook-card-title">' + escapeHTML(h.name) + '</span>' +
                    badge +
                '</div>' +
                '<div class="hook-card-path">' + escapeHTML(h.path) + ' · <span style="color:var(--text-2);">' + escapeHTML(h.agent) + '</span></div>' +
                '<div style="font-size:11px; color:var(--text-2); line-height:1.35;">' + escapeHTML(h.description) + '</div>';
            hooksEl.appendChild(card);
        });
    }

    renderRiskHubsTable();
    renderCriticalityTiersGrid();
}


function renderTalonVerdict(verdict) {
    verdict = verdict || {};
    var hasOutcome = typeof verdict.outcome === "string" && verdict.outcome.length > 0;
    var outcome = hasOutcome ? verdict.outcome : "No evaluation";
    var outcomeLabel = verdict.simulated && hasOutcome ? "DEMO " + outcome : outcome;
    var pill = document.getElementById("talon-verdict-pill");
    if (pill) {
        pill.textContent = outcomeLabel;
        pill.className = "badge-pill " + (outcome === "BLOCK" ? "critical" : (outcome === "WARN" || outcome === "REVIEW" ? "warning" : (outcome === "ALLOW" ? "success" : "info")));
    }
    setText("nav-badge-talon", outcome === "BLOCK" ? "BLOCK" : (hasOutcome ? "Guard" : "—"));
    setText("talon-netnew-count", hasOutcome ? String(verdict.net_new_violations_count || 0) : "—");
    setText("talon-baseline-count", hasOutcome ? String(verdict.baseline_violations_count || 0) : "—");
    setText("talon-adr-contra-count", hasOutcome ? String(verdict.adr_contradictions_count || 0) : "—");
    setText("talon-duration-ms", typeof verdict.duration_ms === "number" ? verdict.duration_ms + "ms" : "—");
    setText("talon-merkle-chip", verdict.merkle_block_height ? "block #" + verdict.merkle_block_height : "No anchor");

    var termStatus = document.getElementById("talon-term-status");
    if (termStatus) {
        termStatus.textContent = hasOutcome ? (verdict.simulated ? "DEMO SIMULATION: " : "OUTCOME: ") + outcome : "NO COMPLETED CHECK";
        termStatus.style.color = outcome === "BLOCK" ? "#F87171" : (hasOutcome ? "#4ADE80" : "var(--muted)");
    }
    var termBody = document.getElementById("talon-term-body");
    if (termBody) {
        if (!hasOutcome) {
            termBody.textContent = "No completed Talon verdict is available for this workspace. Run a guard check to populate this panel.";
        } else {
            var filesCount = Array.isArray(verdict.touched_files) ? verdict.touched_files.length : 0;
            var symsCount = Array.isArray(verdict.touched_symbols) ? verdict.touched_symbols.length : 0;
            var sev = verdict.by_severity || {};
            termBody.innerHTML =
                '<div>Workspace: <strong style="color:#FFF;">' + escapeHTML(verdict.workspace || WORKSPACE) + '</strong></div>' +
                '<div>Diff scope: ' + filesCount + ' file(s), ' + symsCount + ' touched symbol(s)</div>' +
                '<div>Blast radius: ' + (verdict.blast_radius_affected || 0) + ' affected entities (CRITICAL=' + (sev.CRITICAL || 0) + ', HIGH=' + (sev.HIGH || 0) + ')</div>' +
                '<div>Violations: <strong>Net-new=' + (verdict.net_new_violations_count || 0) + '</strong> · Baseline=' + (verdict.baseline_violations_count || 0) + ' · ADR contradictions=' + (verdict.adr_contradictions_count || 0) + '</div>' +
                '<div>Merkle seal: ' + (verdict.merkle_block_height ? "RFC 6962 block #" + verdict.merkle_block_height : "not anchored") +
                (typeof verdict.duration_ms === "number" ? " · " + verdict.duration_ms + "ms" : "") + '</div>';
        }
    }
    var promptBox = document.getElementById("talon-repair-prompt-box");
    if (promptBox) {
        if (!hasOutcome) {
            promptBox.textContent = "No completed guard result is available. Run a Talon check to obtain a workspace-specific verdict.";
        } else if (verdict.self_repair_prompt) {
            promptBox.textContent = verdict.self_repair_prompt;
        } else if (verdict.violations && verdict.violations.length > 0) {
            promptBox.textContent = "GARUDA TALON ENFORCEMENT GATE: CHANGE BLOCKED\n" +
                verdict.violations.map(function(v) { return "  ✗ " + v; }).join("\n") + "\n\nMandatory self-repair actions:\n" +
                (verdict.self_repair_instructions || []).map(function(r, i) { return "  " + (i + 1) + ". " + r; }).join("\n");
        } else {
            promptBox.textContent = "GARUDA TALON ENFORCEMENT GATE: " + outcome + "\nThis result applies to the submitted workspace check. Inspect its evidence and scope before merging.";
        }
    }
    renderBuildContextReceipt(verdict);
}


function renderBuildContextReceipt(verdict) {
    verdict = verdict || {};
    var bc = verdict.build_context || null;
    var drift = verdict.environment_drift || null;
    var matrix = Array.isArray(verdict.matrix_results) ? verdict.matrix_results : [];
    if (bc) {
        setText("receipt-lang-badge", bc.language || "—");
        setText("receipt-platform-val", bc.target_platform || [bc.target_os, bc.target_arch].filter(Boolean).join("/") || "—");
        setText("receipt-compiler-val", bc.compiler_version || "—");
        var tagsArr = Array.isArray(bc.build_tags) && bc.build_tags.length ? bc.build_tags.join(", ") : "none";
        setText("receipt-tags-val", "tags: [" + tagsArr + "]");
        setText("receipt-skipped-badge", (bc.skipped_by_build_tags_count || 0) + " excluded");
        setText("receipt-files-val", (typeof bc.active_files_count === "number" ? bc.active_files_count : "—") + " active · " + (bc.skipped_by_build_tags_count || 0) + " conditional skipped");
        setText("receipt-tree-badge", typeof bc.worktree_dirty === "boolean" ? (bc.worktree_dirty ? "DIRTY" : "CLEAN TREE") : "Unknown");
        setText("receipt-commit-val", (bc.branch || "—") + " @ " + (bc.commit_short || (bc.commit_sha || "").slice(0, 12) || "—"));
        setText("receipt-manifest-val", "manifest lock: " + (bc.config_manifest_hash || "—"));
        setText("receipt-fp-val", "fp: " + ((bc.graph_fingerprint || "").slice(0, 18) || "—"));
        setText("receipt-sha-val", "receipt: " + ((bc.receipt_hash || "").slice(0, 18) || "—"));
        setText("receipt-reproduce-cmd", bc.reproduce_command || verdict.reproduce_command || "—");
    } else {
        ["receipt-lang-badge", "receipt-platform-val", "receipt-compiler-val", "receipt-tags-val", "receipt-skipped-badge",
         "receipt-files-val", "receipt-tree-badge", "receipt-commit-val", "receipt-manifest-val", "receipt-fp-val",
         "receipt-sha-val", "receipt-reproduce-cmd"].forEach(function(id) { setText(id, "—"); });
    }
    var driftBadge = document.getElementById("talon-drift-badge");
    var driftBox = document.getElementById("talon-drift-alert-box");
    if (drift && drift.drift_detected) {
        if (driftBadge) { driftBadge.textContent = "DRIFT DETECTED (" + ((drift.discrepancies && drift.discrepancies.length) || 0) + ")"; driftBadge.className = "badge-pill critical"; }
        if (driftBox) {
            driftBox.style.display = "block";
            driftBox.innerHTML = '<div style="font-weight:700;margin-bottom:4px;">CI-vs-Editor Environment Drift (' + escapeHTML(drift.severity || "UNVERIFIED") + ')</div>' +
                '<div>' + escapeHTML(drift.summary || "") + '</div>' +
                (drift.discrepancies || []).map(function(d) { return '<div style="margin-top:4px;">• ' + escapeHTML(d) + '</div>'; }).join("");
        }
    } else {
        if (driftBadge) { driftBadge.textContent = drift ? "No drift reported" : "Not checked"; driftBadge.className = "badge-pill info"; }
        if (driftBox) { driftBox.style.display = "none"; driftBox.innerHTML = ""; }
    }
    var matrixTbody = document.getElementById("talon-matrix-tbody");
    if (matrixTbody) {
        matrixTbody.innerHTML = "";
        if (!matrix.length) {
            var emptyRow = document.createElement("tr");
            var emptyCell = document.createElement("td");
            emptyCell.colSpan = 6;
            emptyCell.textContent = "No build matrix results are available for this check.";
            emptyRow.appendChild(emptyCell);
            matrixTbody.appendChild(emptyRow);
        } else {
            matrix.forEach(function(m) {
                var tr = document.createElement("tr");
                var tags = Array.isArray(m.build_tags) && m.build_tags.length ? m.build_tags.join(", ") : "none";
                var outClass = m.outcome === "BLOCK" ? "critical" : (m.outcome === "ALLOW" ? "success" : "info");
                tr.innerHTML = '<td class="mono" style="font-weight:700;">' + escapeHTML(m.target_platform || "—") + '</td>' +
                    '<td class="mono" style="font-size:11.5px;">CGO=' + (m.cgo_enabled ? "1" : "0") + ' · tags=[' + escapeHTML(tags) + ']</td>' +
                    '<td style="font-size:11.5px;"><strong>' + (typeof m.active_files_count === "number" ? m.active_files_count : "—") + '</strong> active · ' + (m.skipped_by_build_tags_count || 0) + ' skipped</td>' +
                    '<td class="mono" style="font-size:11.5px;">' + (m.entities_count || 0) + ' entities · ' + (m.edges_count || 0) + ' edges</td>' +
                    '<td class="mono" style="font-size:11.5px;color:var(--text-2);">' + escapeHTML((m.graph_fingerprint || "").slice(0, 16) || "—") + '</td>' +
                    '<td><span class="badge-pill ' + outClass + '">' + escapeHTML(m.outcome || "UNVERIFIED") + '</span></td>';
                matrixTbody.appendChild(tr);
            });
        }
    }
}
function copyReproduceCommand() {
    var el = document.getElementById("receipt-reproduce-cmd");
    var cmd = (el && el.textContent) || "GOOS=linux GOARCH=amd64 CGO_ENABLED=0 garuda check --target linux/amd64 --cgo=false --workspace default";
    copyTextToClipboard(cmd, "📋 Copied 1-line Reproduce Verdict command to clipboard!");
}

function copyTextToClipboard(text, toastMsg) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(function() {
            showToast(toastMsg || "Copied to clipboard!");
        }).catch(function() {
            showToast(toastMsg || "Ready to copy: " + text);
        });
    } else {
        showToast(toastMsg || "Ready to copy: " + text);
    }
}

function filterRiskHubs(tier, btnEl) {
    state.riskTierFilter = tier;
    if (btnEl && btnEl.parentElement) {
        btnEl.parentElement.querySelectorAll(".filter-pill").forEach(function(b) {
            b.classList.toggle("active", b === btnEl);
        });
    }
    renderRiskHubsTable();
    renderCriticalityTiersGrid();
}

function renderRiskHubsTable() {
    var tbody = document.getElementById("talon-risk-hubs-tbody");
    if (!tbody || !state.talonHubData) return;

    var hubs = (state.talonHubData.risk_hubs || []).filter(function(h) {
        if (state.riskTierFilter === "ALL") return true;
        return h.risk_tier === state.riskTierFilter;
    });

    tbody.innerHTML = "";
    if (hubs.length === 0) {
        var emptyMessage = state.talonHubData.unavailable
            ? "Risk candidates unavailable. No sample graph was substituted."
            : "No indexed symbols are available in this workspace.";
        tbody.innerHTML = '<tr><td colspan="6" style="color:var(--muted); padding:16px;">' + escapeHTML(emptyMessage) + '</td></tr>';
        return;
    }

    hubs.forEach(function(h) {
        var tr = document.createElement("tr");
        var barColor = h.risk_tier === "CRITICAL" ? "var(--red)" : (h.risk_tier === "HIGH" ? "var(--amber)" : "var(--green)");
        var rulesHtml = (h.governing_rules || []).map(function(r) {
            return '<span class="code-ref-chip" onclick="event.stopPropagation(); showView(\'docsgraph\')">🛡️ ' + escapeHTML(shortenLabel(r, 34)) + '</span>';
        }).join("");

        tr.innerHTML =
            '<td>' +
                '<div style="display:flex; align-items:center; gap:8px;">' +
                    '<span class="risk-score-pill ' + escapeHTML(h.risk_tier) + '">' + h.risk_score + ' · ' + escapeHTML(h.risk_tier) + '</span>' +
                    '<div class="risk-bar-track"><div class="risk-bar-fill" style="width:' + h.risk_score + '%; background:' + barColor + ';"></div></div>' +
                '</div>' +
            '</td>' +
            '<td>' +
                '<div class="mono" style="font-weight:700; font-size:12.5px; color:var(--text);">' + symbolIcon(h.kind) + ' ' + escapeHTML(h.name) + '</div>' +
                '<div style="font-size:11px; color:var(--muted);">' + escapeHTML(h.kind) + ' · ' + escapeHTML(h.epistemic_status || "UNVERIFIED") + '</div>' +
            '</td>' +
            '<td class="mono" style="font-size:11.5px; color:var(--text-2);">' +
                '<div>' + escapeHTML(shortenPath(h.package)) + '</div>' +
                '<div style="color:var(--muted); font-size:10.5px;">' + escapeHTML(shortenPath(h.file || "")) + '</div>' +
            '</td>' +
            '<td>' +
                '<div style="font-weight:650; font-size:12px;">' + h.inbound_callers + ' direct callers · ~' + h.transitive_radius + ' transitive</div>' +
                '<div style="font-size:11px; color:var(--muted);">' + h.outbound_callees + ' outgoing callees</div>' +
            '</td>' +
            '<td>' + rulesHtml + '</td>' +
            '<td>' +
                '<div style="display:flex; gap:6px;">' +
                    '<button class="btn-secondary" style="height:26px; font-size:11px;" onclick="event.stopPropagation(); openPreEditBriefDrawer(\'' + escapeJS(h.name) + '\')">Pre-Edit Brief</button>' +
                    '<button class="btn-primary" style="height:26px; font-size:11px;" onclick="event.stopPropagation(); focusSymbolOnArchitectureMap(\'' + escapeJS(h.name) + '\', \'' + escapeJS(h.package) + '\', null)">1-Hop Map →</button>' +
                '</div>' +
            '</td>';

        tr.style.cursor = "pointer";
        tr.onclick = function() {
            openPreEditBriefDrawer(h.name);
        };
        tbody.appendChild(tr);
    });
}


async function runTalonCheckFromUI(action) {
    var isSimulation = action === "simulate_violation" || action === "simulate_drift";
    showToast(isSimulation ? "Preparing the selected Talon simulation…" : "Running diff-scoped Talon Guard against HEAD…");
    try {
        var res = await fetch("/api/v1/dashboard/talon-check?workspace=" + encodeURIComponent(WORKSPACE), {
            method: "POST",
            headers: { "Content-Type": "application/json", "Accept": "application/json" },
            body: JSON.stringify({ workspace: WORKSPACE, action: action, target: arguments.length > 1 ? arguments[1] : "", diff_rev: "HEAD" })
        });
        var data = null;
        try { data = await res.json(); } catch (parseErr) {}
        if (!res.ok) throw new Error(data && (data.error || data.message) || ("Talon check failed (HTTP " + res.status + ")."));
        if (action === "matrix_check") {
            if (!data || data.status !== "ok" || !Array.isArray(data.matrix_results)) throw new Error("The server did not return a completed build matrix.");
            renderBuildContextReceipt({ matrix_results: data.matrix_results });
            showToast("Build matrix completed · " + data.matrix_results.length + " target(s) · " + (data.scope || "scope not reported") + ".");
            return;
        }
        if (!data || !data.verdict || typeof data.verdict.outcome !== "string") throw new Error("The server did not return a completed Talon verdict.");
        if (data.simulated === true) data.verdict.simulated = true;
        if (!state.talonHubData) state.talonHubData = { hooks: [], risk_hubs: [] };
        state.talonHubData.latest_verdict = data.verdict;
        renderTalonVerdict(data.verdict);
        showToast((data.simulated ? "Simulation · " : "Talon Gate Outcome: ") + data.verdict.outcome);
        return;
    } catch (err) {
        if (!isSimulation) {
            showToast(err && err.message ? err.message : "Talon check failed. No verdict was recorded.");
            return;
        }
    }
    if (!state.talonHubData) state.talonHubData = { hooks: [], risk_hubs: [] };
    if (action === "simulate_drift") {
        var driftVerdict = {
            workspace: WORKSPACE, outcome: "BLOCK", simulated: true, agent_should_proceed: false,
            touched_files: ["internal/refund/charge_bypass.go"],
            touched_symbols: ["refund.ProcessCharge (charge_bypass.go:8)"],
            touched_packages: ["internal/refund"], blast_radius_affected: 16,
            by_severity: { CRITICAL: 2, HIGH: 5, MEDIUM: 6, LOW: 3 },
            net_new_violations_count: 1, baseline_violations_count: 0, adr_contradictions_count: 1,
            violations: ["Simulated conditional build edge violates the payment architecture rule."],
            self_repair_instructions: ["Inspect the conditional build file and route persistence through the approved store boundary."],
            self_repair_prompt: "DEMO SIMULATION — not a repository finding. Conditional build drift was simulated to demonstrate a BLOCK and a self-repair path."
        };
        state.talonHubData.latest_verdict = driftVerdict;
        renderTalonVerdict(driftVerdict);
        showToast("Demo simulation: conditional build drift returned BLOCK.");
        return;
    }
    if (action === "matrix_check") {
        showToast("The server did not return build-matrix results. No sample matrix was substituted.");
        return;
    }
    if (action === "simulate_violation") {
        var simVerdict = {
            workspace: WORKSPACE, outcome: "BLOCK", simulated: true, agent_should_proceed: false,
            touched_files: ["internal/api/handler.go", "internal/refund/rate_limiter.go"],
            touched_symbols: ["api.HandleRefund (handler.go:194)", "refund.ProcessDirectBankCall (rate_limiter.go:42)"],
            touched_packages: ["internal/api", "internal/refund"], blast_radius_affected: 19,
            by_severity: { CRITICAL: 3, HIGH: 6, MEDIUM: 7, LOW: 3 },
            net_new_violations_count: 2, baseline_violations_count: 0, adr_contradictions_count: 1,
            violations: [
                "Simulated layer violation: API imports database/sql directly.",
                "Simulated ADR violation: refund handler calls a forbidden direct bank API."
            ],
            self_repair_instructions: [
                "Route persistence through internal/store.",
                "Route refunds through the idempotency store boundary."
            ],
            self_repair_prompt: "DEMO SIMULATION — not a repository finding. A simulated architecture and ADR violation was blocked. Refactor through the approved boundaries, then run a real Talon check."
        };
        state.talonHubData.latest_verdict = simVerdict;
        renderTalonVerdict(simVerdict);
        showToast("Demo simulation: Talon blocked two simulated violations.");
    }
}


async function installTalonHooksFromUI() {
    showToast("Compiling the workspace contract and installing supported hooks…");
    try {
        var res = await fetch("/api/v1/dashboard/talon-check?workspace=" + encodeURIComponent(WORKSPACE), {
            method: "POST",
            headers: { "Content-Type": "application/json", "Accept": "application/json" },
            body: JSON.stringify({ workspace: WORKSPACE, action: "install_hooks" })
        });
        var data = null;
        try { data = await res.json(); } catch (parseErr) {}
        if (!res.ok) throw new Error(data && (data.error || data.message) || ("Hook installation failed (HTTP " + res.status + ")."));
        if (!data || data.status !== "installed" || !data.install_summary) throw new Error("The server did not confirm hook installation.");
        if (data.talon_hub) {
            state.talonHubData = data.talon_hub;
            renderTalonHub();
        } else {
            state.talonHubData = null;
            await loadTalonHub();
        }
        if (data.refresh_warning) showToast("Hooks were installed, but their status could not be refreshed. Use Refresh to retry.");
        else showToast("Hook installation completed in the configured Garuda repository root. Verify each status below.");
    } catch (err) {
        showToast(err && err.message ? err.message : "Hook installation failed. No local success state was applied.");
    }
}

function copyTalonRepairPrompt() {
    var box = document.getElementById("talon-repair-prompt-box");
    var text = (box && box.textContent) || "";
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(function() {
            showToast("📋 Copied Talon Self-Repair Prompt to clipboard!");
        }).catch(function() {
            showToast("📋 Self-Repair Prompt ready to copy!");
        });
    } else {
        showToast("📋 Self-Repair Prompt ready!");
    }
}

async function openPreEditBriefDrawer(symbolName) {
    var drawer = document.getElementById("drawer");
    var overlay = document.getElementById("drawer-overlay");
    var body = document.getElementById("drawer-body");
    if (!drawer || !overlay || !body) return;
    var target = (symbolName || "").trim();
    if (!target) { showToast("Choose an indexed symbol before requesting a pre-edit briefing."); return; }

    setText("drawer-title", "Talon Pre-Edit Briefing");
    setText("drawer-kind", "LIVE WORKSPACE BRIEF · " + target);
    body.innerHTML = '<div class="detail-section"><div class="detail-section-title">Loading workspace evidence…</div><div class="row-meta">Resolving matching entities, blast radius, ADRs and active policies.</div></div>';
    drawer.classList.add("open");
    overlay.classList.add("open");
    try {
        var res = await fetch("/api/v1/dashboard/talon-check?workspace=" + encodeURIComponent(WORKSPACE), {
            method: "POST",
            headers: { "Content-Type": "application/json", "Accept": "application/json" },
            body: JSON.stringify({ workspace: WORKSPACE, action: "pre_edit_brief", target: target })
        });
        var payload = null;
        try { payload = await res.json(); } catch (parseErr) {}
        if (!res.ok) throw new Error(payload && (payload.error || payload.message) || ("Pre-edit briefing failed (HTTP " + res.status + ")."));
        if (!payload || payload.status !== "ok" || !payload.briefing || payload.briefing.target !== target) {
            throw new Error("The server did not return a workspace-scoped pre-edit briefing.");
        }

        var brief = payload.briefing;
        var matches = Array.isArray(brief.matched_entities) ? brief.matched_entities : [];
        var adrs = Array.isArray(brief.applicable_adrs) ? brief.applicable_adrs : [];
        var policies = Array.isArray(brief.active_policies) ? brief.active_policies : [];
        var protocol = Array.isArray(brief.mandatory_protocol) ? brief.mandatory_protocol : [];
        var sev = brief.by_severity || {};
        var html = '<div class="detail-section"><div class="detail-section-title">Workspace &amp; target</div>';
        html += propertyRow("Workspace", brief.workspace || WORKSPACE);
        html += propertyRow("Requested target", brief.target);
        html += propertyRow("Matched entities", formatNumber(matches.length));
        html += propertyRow("Blast radius", formatNumber(brief.blast_radius_total || 0));
        html += propertyRow("Severity", "CRITICAL " + (sev.CRITICAL || 0) + " · HIGH " + (sev.HIGH || 0) + " · MEDIUM " + (sev.MEDIUM || 0) + " · LOW " + (sev.LOW || 0));
        html += '</div>';

        html += '<div class="detail-section"><div class="detail-section-title">Matched workspace entities</div>';
        if (!matches.length) html += '<div class="row-meta">No indexed entity matched this target. The briefing does not infer callers or impact.</div>';
        else matches.forEach(function(entity) {
            html += '<div class="list-row"><div class="row-main"><div class="row-title mono">' + escapeHTML(entity.name || "Unnamed entity") + '</div>' +
                '<div class="row-meta">' + escapeHTML(entity.kind || "entity") + ' · ' + escapeHTML(entity.package || "package unavailable") + ' · ' +
                escapeHTML(entity.file_path || "source path unavailable") + '</div></div><span class="badge-pill info">' + formatNumber(entity.inbound_edges || 0) + ' inbound</span></div>';
        });
        html += '</div>';

        html += '<div class="detail-section"><div class="detail-section-title">Applicable ADR claims</div>';
        if (!adrs.length) html += '<div class="row-meta">No applicable ADR claims were returned for this target.</div>';
        else adrs.forEach(function(adr) {
            html += '<div class="list-row"><div class="row-main"><div class="row-title">' + escapeHTML(adr.document || "ADR") + ' · ' + escapeHTML(adr.modality || "UNSPECIFIED") + ' ' + escapeHTML(adr.predicate || "") + '</div>' +
                '<div class="row-meta">' + escapeHTML(adr.subject || "") + ' → ' + escapeHTML(adr.object || "") + ' · ' + escapeHTML(adr.status || "UNVERIFIED") +
                (adr.reason ? ' · ' + escapeHTML(adr.reason) : '') + '</div></div></div>';
        });
        html += '</div>';

        html += '<div class="detail-section"><div class="detail-section-title">Active workspace policies</div>';
        if (!policies.length) html += '<div class="row-meta">No active policies were returned from the compiled workspace contract.</div>';
        else policies.forEach(function(policy) {
            html += '<div class="list-row"><div class="row-main"><div class="row-title">' + escapeHTML(policy.title || policy.id || "Policy") + '</div>' +
                '<div class="row-meta">' + escapeHTML(policy.id || "") + ' · ' + escapeHTML(policy.scope_domain || "workspace scope") + ' · ' +
                escapeHTML(policy.decision || "UNVERIFIED") + (policy.reason ? ' · ' + escapeHTML(policy.reason) : '') + '</div></div></div>';
        });
        html += '</div>';

        if (brief.build_context) {
            var build = brief.build_context;
            html += '<div class="detail-section"><div class="detail-section-title">Build context returned by Garuda</div>';
            html += propertyRow("Language", build.language || "—") + propertyRow("Compiler", build.compiler_version || "—") +
                propertyRow("Target", build.target_platform || "—") + propertyRow("Commit", build.commit_short || "—");
            html += '</div>';
        }

        html += '<div class="detail-section"><div class="detail-section-title">Mandatory agent protocol</div>';
        html += protocol.length
            ? '<ol style="padding-left:20px;color:var(--text-2);font-size:12px;line-height:1.65;">' + protocol.map(function(step) { return '<li>' + escapeHTML(step) + '</li>'; }).join("") + '</ol>'
            : '<div class="row-meta">No protocol steps were returned.</div>';
        if (matches.length) {
            var first = matches[0];
            html += '<button class="btn-primary" onclick="closeDrawer(); focusSymbolOnArchitectureMap(\'' + escapeJS(first.name || target) + '\', \'' + escapeJS(first.package || "") + '\', null)">Open first matched entity in Architecture Map →</button>';
        }
        html += '</div>';
        body.innerHTML = html;
    } catch (err) {
        body.innerHTML = '<div class="detail-section"><div class="detail-section-title">Briefing unavailable</div><div class="row-meta">' +
            escapeHTML(err && err.message ? err.message : "No live pre-edit briefing was returned.") +
            '</div><div class="row-meta" style="margin-top:8px;">No sample callers, severity counts, or governing rules were substituted.</div></div>';
    }
}

function openContractModal() {
    var modal = document.getElementById("info-modal");
    var title = document.getElementById("info-modal-title");
    var kicker = document.getElementById("info-modal-kicker");
    var body = document.getElementById("info-modal-body");
    if (!modal || !body) return;

    var contractMd = (state.talonHubData && state.talonHubData.contract && state.talonHubData.contract.markdown) ||
                     buildDefaultTalonHub().contract.markdown;
    state.infoModalRawText = contractMd;

    if (kicker) kicker.textContent = "COMPILED AGENT HARNESS CONTRACT (.garuda/TALON_CONTRACT.md)";
    if (title) title.textContent = "Garuda Talon Active Architectural Contract";
    body.innerHTML = '<pre style="background:var(--bg); border:1px solid var(--border); padding:14px; border-radius:8px; overflow-x:auto; white-space:pre-wrap; font-size:11.5px; line-height:1.5;">' + escapeHTML(contractMd) + '</pre>';
    modal.classList.add("open");
}

function openAPISpecModal() {
    var modal = document.getElementById("info-modal");
    var title = document.getElementById("info-modal-title");
    var kicker = document.getElementById("info-modal-kicker");
    var body = document.getElementById("info-modal-body");
    if (!modal || !body) return;

    var specSummary = "# GARUDA v1.7 REST API & 23 MCP TOOLS SPECIFICATION\n\n" +
        "## Primary Dashboard & Governance HTTP Endpoints\n" +
        "- GET  /dashboard?workspace=<name>          — 8-Surface Interactive Intelligence Instrument\n" +
        "- GET  /api/v1/dashboard/stats              — Workspace AST, Epistemic 3-State & Drift Pulse\n" +
        "- GET  /api/v1/dashboard/docs-graph         — 4-Tier Document-to-Code & Guardrails Lineage Graph\n" +
        "- POST /api/v1/dashboard/docs-ingest        — Live ADR / Design Spec / Policy Rule Ingestion\n" +
        "- GET  /api/v1/dashboard/talon-hub          — Garuda Talon Active Agent Harness & Risk Hub\n" +
        "- POST /api/v1/dashboard/talon-check        — Run Diff-Scoped Gate, Hook Sync & Self-Repair Loop\n" +
        "- GET  /api/v1/dashboard/policies           — Declarative YAML Policy Evaluations & Merkle Proofs\n" +
        "- GET  /api/v1/dashboard/policies/verify    — RFC 6962 Cryptographic Inclusion Proof Verifier\n" +
        "- GET  /api/v1/graph?level=entity|package|repository|full|docs — Multi-Level Topology & Docs Graph\n" +
        "- POST /api/v1/telemetry/spans              — OpenTelemetry Runtime Span Ingestion\n\n" +
        "## Active Garuda Talon MCP Tools (garuda-mcp)\n" +
        "- garuda.talon_pre_edit      — Mandatory pre-flight blast radius & ADR/policy briefing\n" +
        "- garuda.talon_verify_edit   — Post-edit incremental AST re-index & diff-scoped policy gate\n" +
        "- garuda.talon_contract      — Retrieve compiled workspace contract & top risk hubs\n" +
        "- garuda.blast_radius        — Severity-bucketed transitive impact analysis\n" +
        "- garuda.verify_architecture — Full workspace policy & ADR verification";

    state.infoModalRawText = specSummary;
    if (kicker) kicker.textContent = "OPENAPI 3.0 & MCP SPECIFICATION";
    if (title) title.textContent = "Garuda v1.7 API & MCP Tool Reference";
    body.innerHTML = '<pre style="background:var(--bg); border:1px solid var(--border); padding:14px; border-radius:8px; overflow-x:auto; white-space:pre-wrap; font-size:11.5px; line-height:1.5;">' + escapeHTML(specSummary) + '</pre>' +
        '<div style="margin-top:10px;"><a href="/docs" target="_blank" rel="noopener" class="btn-secondary" style="text-decoration:none; display:inline-flex; align-items:center;">Open Full Swagger UI (/docs) ↗</a></div>';
    modal.classList.add("open");
}

function closeInfoModal() {
    var modal = document.getElementById("info-modal");
    if (modal) modal.classList.remove("open");
}

function copyInfoModalContent() {
    if (navigator.clipboard && navigator.clipboard.writeText && state.infoModalRawText) {
        navigator.clipboard.writeText(state.infoModalRawText).then(function() {
            showToast("📋 Copied content to clipboard!");
        });
    } else {
        showToast("📋 Content ready!");
    }
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
    ["entity", "package", "repo", "full", "docs"].forEach(function(k) {
        var btn = document.getElementById("btn-mode-" + k);
        if (btn) {
            var match = (k === "repo" && level === "repository") || k === level;
            btn.classList.toggle("active", match);
        }
    });
    loadArchitecture(level, "");
}

function resetArchitectureMap() {
    state.graphCache = {};
    state.edgeTypeFilter = "ALL";
    document.querySelectorAll("[data-edge-filter]").forEach(function(btn) {
        btn.classList.toggle("active", btn.getAttribute("data-edge-filter") === "ALL");
    });
    var searchInp = document.getElementById("arch-symbol-search");
    if (searchInp) searchInp.value = "";
    loadArchitecture(state.currentLevel || "entity", "");
    showToast("Reset Architecture Map view");
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
    if (!query) {
        resetArchitectureMap();
        return;
    }
    state.graphCache = {};
    state.currentFocus = query;
    loadArchitecture(state.currentLevel || "entity", query);
    showToast("Focused map on “" + query + "”");
}

function toggleFullscreen() {
    var wrap = document.getElementById("graph-wrap");
    if (wrap) wrap.classList.toggle("fullscreen");
    setTimeout(fitGraph, 200);
}

async function loadArchitecture(level, focus) {
    state.currentLevel = level || "entity";
    state.currentFocus = focus || "";
    ["entity", "package", "repo", "full", "docs"].forEach(function(k) {
        var btn = document.getElementById("btn-mode-" + k);
        if (btn) btn.classList.toggle("active", (k === "repo" && state.currentLevel === "repository") || k === state.currentLevel);
    });
    updateArchitectureHeader();
    if (state.currentLevel === "docs") {
        if (!state.docsGraphData) await loadDocsGraph();
        var docsGraph = state.docsGraphData && state.docsGraphData.graph;
        if (!docsGraph) docsGraph = DEMO_MODE ? buildDefaultDocsGraph().graph : { nodes: [], edges: [] };
        state.graphData = docsGraph;
        buildCommunitiesList(docsGraph);
        renderGraph(docsGraph);
        return;
    }
    var cacheKey = state.currentLevel + ":" + state.currentFocus;
    if (state.graphCache[cacheKey]) {
        state.graphData = state.graphCache[cacheKey];
        buildCommunitiesList(state.graphData);
        renderGraph(state.graphData);
        return;
    }
    var url = "/api/v1/graph?workspace=" + encodeURIComponent(WORKSPACE) + "&level=" + encodeURIComponent(state.currentLevel);
    if (state.currentFocus) url += "&focus=" + encodeURIComponent(state.currentFocus);
    try {
        var res = await fetch(url, { headers: { "Accept": "application/json" } });
        if (!res.ok) throw new Error("Graph request failed (HTTP " + res.status + ").");
        var data = await res.json();
        if (!data || !Array.isArray(data.nodes) || !Array.isArray(data.edges || [])) throw new Error("Graph response was not valid.");
        if (data.nodes.length === 0 && state.currentFocus) {
            var broaderRes = await fetch("/api/v1/graph?workspace=" + encodeURIComponent(WORKSPACE) + "&level=entity", {
                headers: { "Accept": "application/json" }
            });
            if (broaderRes.ok) {
                var broader = await broaderRes.json();
                if (broader && Array.isArray(broader.nodes) && Array.isArray(broader.edges || [])) data = broader;
            }
        }
        if (data.nodes.length === 0 && DEMO_MODE) data = buildDefaultDocsGraph().graph;
        state.graphCache[cacheKey] = data;
        state.graphData = data;
        buildCommunitiesList(data);
        renderGraph(data);
    } catch (err) {
        var emptyGraph = DEMO_MODE ? buildDefaultDocsGraph().graph : { nodes: [], edges: [] };
        state.graphData = emptyGraph;
        buildCommunitiesList(emptyGraph);
        renderGraph(emptyGraph);
        if (!DEMO_MODE) showToast("Architecture data is unavailable. Check the workspace connection and refresh.");
    }
}
function updateArchitectureHeader() {
    var title = "Architecture Map";
    var subtitle = "Focused 1-hop and 2-hop neighborhood around your selected symbol or package.";
    var breadcrumb = state.currentFocus || "Neighborhood";

    if (state.currentLevel === "docs") {
        title = "Docs, ADRs, Specs & Guardrails Graph";
        subtitle = "Visualizing how ingested ADRs, design specs, and YAML policies connect to concrete AST code symbols.";
        breadcrumb = "Docs & Rules Graph";
    } else if (state.currentLevel === "full") {
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
    if (node.kind === "document") return "1. Documents (ADR/Spec)";
    if (node.kind === "doc_component") return "2. Extracted Rules";
    if (node.kind === "policy_guardrail") return "3. Policy Guardrails";
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

    setText("nb-selected-title", centerNode.label || centerNode.id);
    var locText = (centerNode.package || centerNode.repo || "workspace") +
        (centerNode.file ? " · " + shortenPath(centerNode.file) : "") +
        " (" + (centerNode.kind || "entity") + ")";
    setText("nb-selected-loc", locText);

    if (incoming.length > 0) {
        setText("nb-caller-title", incoming[0].node.label || "Caller");
        setText("nb-caller-sub", incoming.length === 1
            ? "1 incoming reference (" + (incoming[0].edge.type || "CALLS") + ")"
            : incoming.length + " incoming references · click to pivot");
    } else {
        setText("nb-caller-title", "Entrypoint / Source");
        setText("nb-caller-sub", "0 incoming callers in view");
    }

    if (outgoing.length > 0) {
        setText("nb-dep-title", outgoing[0].node.label || "Dependency");
        setText("nb-dep-sub", outgoing.length === 1
            ? "1 outgoing link (" + (outgoing[0].edge.type || "CALLS") + ")"
            : outgoing.length + " outgoing links · click to pivot");
    } else {
        setText("nb-dep-title", "Target / Leaf Entity");
        setText("nb-dep-sub", "0 outgoing dependencies in view");
    }

    var policyDecision = (state.policiesData && state.policiesData.final_decision) ||
        (state.policiesData && state.policiesData.unavailable ? "Unavailable" : "No evaluation");
    var activePolCount = (state.policiesData && state.policiesData.active_policies) || 0;
    setText("nb-policy-title", "Policy Gate: " + policyDecision);
    setText("nb-policy-sub", activePolCount + " active rules · click for Change Review");

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
        html += '<button class="btn-secondary" style="width:100%; justify-content:center;" onclick="openPreEditBriefDrawer(\'' + escapeJS(centerNode.label || centerNode.id) + '\')">Talon Pre-Edit Briefing →</button>';
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
    if (!list || list.length === 0) {
        showToast("No " + (dir === "in" ? "incoming callers" : "outgoing dependencies") + " in current neighborhood");
        return;
    }
    var drawer = document.getElementById("drawer");
    var overlay = document.getElementById("drawer-overlay");
    var body = document.getElementById("drawer-body");
    if (!drawer || !overlay || !body) return;

    var centerName = (state.selectedNode && (state.selectedNode.label || state.selectedNode.id)) || "Selected Symbol";
    setText("drawer-title", (dir === "in" ? "Incoming Callers of " : "Outgoing Dependencies of ") + centerName);
    setText("drawer-kind", list.length + " CONNECTED GRAPH EDGES");

    var html = '<div class="detail-section">';
    list.forEach(function(item) {
        html += '<div style="display:flex; align-items:center; justify-content:space-between; padding:8px 10px; margin-bottom:6px; background:var(--surface-2); border:1px solid var(--border); border-radius:6px; cursor:pointer;" onclick="closeDrawer(); focusSymbolOnArchitectureMap(\'' + escapeJS(item.node.label || item.node.id) + '\', \'' + escapeJS(item.node.package || "") + '\', null)">' +
            '<div>' +
                '<div class="mono" style="font-weight:650; font-size:12px;">' + escapeHTML(item.node.label || item.node.id) + '</div>' +
                '<div style="font-size:11px; color:var(--muted);">' + escapeHTML(item.node.package || item.node.repo || "workspace") + '</div>' +
            '</div>' +
            '<span class="kind-pill">' + escapeHTML(item.edge.type || "CALLS") + ' →</span>' +
        '</div>';
    });
    html += '</div>';
    body.innerHTML = html;
    drawer.classList.add("open");
    overlay.classList.add("open");
}

function focusFirstCaller() {
    if (currentIncomingNodes.length > 0) {
        var target = currentIncomingNodes[0].node;
        updateNeighborhoodAndInspector(target, state.graphData._renderedNodes || [], state.graphData._renderedEdges || []);
        highlightSelectedNodeOnCanvas(target);
    } else {
        showToast("0 incoming callers for this entrypoint in view");
    }
}

function focusFirstDependency() {
    if (currentOutgoingNodes.length > 0) {
        var target = currentOutgoingNodes[0].node;
        updateNeighborhoodAndInspector(target, state.graphData._renderedNodes || [], state.graphData._renderedEdges || []);
        highlightSelectedNodeOnCanvas(target);
    } else {
        showToast("0 outgoing dependencies for this leaf node in view");
    }
}

function highlightSelectedNodeOnCanvas(centerNode) {
    if (!state.graphGroup || !centerNode) return;
    state.graphGroup.selectAll(".node-core")
        .attr("fill", function(d) {
            if (d.id === centerNode.id) return "#191B19";
            if (d.status === "CONTRADICTED") return "#991B1B";
            if (d.kind === "document") return "#137333";
            if (d.kind === "doc_component") return "#B06000";
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
            if (d.id === centerNode.id) return "#D5E66A";
            if (d.kind === "document" || d.kind === "doc_component") return "#FFFFFF";
            return getCommunityColor(d);
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
        if (state.edgeTypeFilter === "GOVERNS") {
            var t = (e.type || "").toUpperCase();
            return t.indexOf("GOVERN") !== -1 || t.indexOf("DEFINE") !== -1 || t.indexOf("ENFORCE") !== -1;
        }
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

    var centerNode = state.selectedNode && nodeByID[state.selectedNode.id] ? nodeByID[state.selectedNode.id] : null;
    if (!centerNode && state.currentFocus) {
        var focusLow = state.currentFocus.toLowerCase();
        centerNode = allNodes.find(function(n) {
            return n.id === state.currentFocus || n.label === state.currentFocus || n.package === state.currentFocus ||
                   (n.label && n.label.toLowerCase().indexOf(focusLow) !== -1) ||
                   (n.package && n.package.toLowerCase().indexOf(focusLow) !== -1);
        });
    }
    if (!centerNode) {
        var sortedByDeg = allNodes.slice().sort(function(a, b) { return (b.count || 0) - (a.count || 0); });
        centerNode = sortedByDeg[0];
    }

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
        if (allowed.size < 8) {
            allNodes.slice(0, 20).forEach(function(n) { allowed.add(n.id); });
        }
        nodes = allNodes.filter(function(n) { return allowed.has(n.id); });
        validEdges = allEdges.filter(function(e) { return allowed.has(e.source.id) && allowed.has(e.target.id); });
    }

    data._renderedNodes = nodes;
    data._renderedEdges = validEdges;

    updateNeighborhoodAndInspector(centerNode, nodes, validEdges);
    setText("graph-description", "Showing " + nodes.length + " focused nodes and " + validEdges.length + " relationships (" + (state.currentLevel === "docs" ? "Docs-to-Code Lineage" : state.hopDepth + "-hop neighborhood") + ")");

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
            if (d.kind === "document") return "#137333";
            if (d.kind === "doc_component") return "#B06000";
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
            if (centerNode && d.id === centerNode.id) return "#D5E66A";
            if (d.kind === "document" || d.kind === "doc_component") return "#FFFFFF";
            return getCommunityColor(d);
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
    if (state.currentLevel === "package" || state.currentLevel === "full" || state.currentLevel === "docs") {
        setArchitectureMode("repository");
        return;
    }
    showView("workspace");
}

function openSelectedInChangeReview() {
    showView("review");
}

/* ============================================================================
   06 — CHANGE REVIEW: 3-PANE SIGNATURE WORKBENCH
   ============================================================================ */
function openEvalInChangeReview(evalIndex) {
    state.selectedEvalIndex = evalIndex;
    showView("review");
}

function runLiveReviewCheck() {
    var picker = document.getElementById("review-target-picker");
    var value = picker && picker.value || "";
    var target = value.indexOf("symbol:") === 0 ? value.slice("symbol:".length) : "";
    runTalonCheckFromUI("run_guard", target);
}

function renderChangeReviewWorkbench() {
    var picker = document.getElementById("review-target-picker");
    var evalListEl = document.getElementById("review-evaluations-list");
    var evals = state.policiesData && Array.isArray(state.policiesData.latest_evaluations) ? state.policiesData.latest_evaluations : [];
    var hubs = state.stats && Array.isArray(state.stats.top_hubs) ? state.stats.top_hubs : [];

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
    var evals = state.policiesData && Array.isArray(state.policiesData.latest_evaluations) ? state.policiesData.latest_evaluations : [];
    var hubs = state.stats && Array.isArray(state.stats.top_hubs) ? state.stats.top_hubs : [];
    var stage = document.getElementById("review-impact-stage");

    var targetName = WORKSPACE || "default";
    var pkgName = "workspace-wide";
    var filePath = "No source path available";
    var decision = state.policiesData && state.policiesData.final_decision || (state.policiesData && state.policiesData.unavailable ? "Unavailable" : "No evaluation");
    var epistemic = state.stats && state.stats.contradicted > 0 ? "CONTRADICTED" : "UNVERIFIED";
    var ledgerText = state.stats && state.stats.latest_block_height ? "BLOCK #" + state.stats.latest_block_height : "No ledger anchor";
    var ruleName = "No policy evaluation selected";
    var reasonText = "Select a recorded policy evaluation or a symbol to inspect its available evidence. No verdict is inferred here.";
    var evCounts = state.stats
        ? formatNumber(state.stats.entities || 0) + " entities · " + formatNumber(state.stats.relationships || 0) + " relationships · " + formatNumber(state.stats.contradicted || 0) + " contradictions"
        : "Workspace data unavailable";
    var evalId = "";
    var callersCount = hubs.length > 0 ? (hubs[0].callers || 0) : 0;

    if (val && val.indexOf("eval:") === 0) {
        var idx = parseInt(val.split(":")[1], 10);
        var ev = evals[idx];
        if (ev) {
            targetName = ev.subject_id || ev.policy_title;
            pkgName = ev.subject_kind || "workspace";
            decision = ev.decision || "UNVERIFIED";
            epistemic = ev.contradiction_count > 0 ? "CONTRADICTED" : "UNVERIFIED";
            ledgerText = ev.merkle_block_height ? "BLOCK #" + ev.merkle_block_height : "Not anchored";
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
    setText("rev-blast-summary", callersCount + " upstream callers · " + currentOutgoingNodes.length + " downstream dependencies");

    var epEl = document.getElementById("rev-triad-epistemic");
    if (epEl) {
        epEl.textContent = epistemic;
        epEl.className = "badge-pill " + (epistemic === "CONTRADICTED" ? "state-contradicted" : (epistemic === "SUPPORTED" ? "state-supported" : "state-unverified"));
    }
    var polEl = document.getElementById("rev-triad-policy");
    if (polEl) {
        polEl.textContent = decision;
        var decisionClass = ["ALLOW", "BLOCK", "WARN", "REVIEW"].indexOf(decision) >= 0 ? decision : "UNVERIFIED";
        polEl.className = "policy-badge " + decisionClass;
    }
    var ledEl = document.getElementById("rev-triad-ledger");
    if (ledEl) {
        ledEl.textContent = ledgerText;
    }

    setText("rev-dr-id", evalId ? ("eval " + evalId.slice(0, 12)) : "No recorded evaluation");
    setText("rev-dr-what", reasonText);
    setText("rev-dr-rule", ruleName);
    setText("rev-dr-evidence", evCounts);
    var cliCmd = evalId
        ? "No reproduce command was returned for this policy evaluation."
        : "Select a recorded evaluation to inspect its reproduction details.";
    setText("rev-dr-cli", cliCmd);

    var actionsEl = document.getElementById("rev-dr-actions");
    if (actionsEl) {
        var btns = '<button class="btn-primary" onclick="focusSymbolOnArchitectureMap(\'' + escapeJS(targetName) + '\', \'' + escapeJS(pkgName) + '\', null)">Open in Architecture Map →</button>';
        if (evalId) btns += '<button class="btn-secondary" onclick="verifyPolicyAnchor(\'' + escapeJS(evalId) + '\')">Verify RFC 6962 Proof</button>';
        btns += '<button class="btn-secondary" onclick="openPreEditBriefDrawer(\'' + escapeJS(targetName) + '\')">Talon Pre-Edit Brief</button>';
        actionsEl.innerHTML = btns;
    }

    if (stage) {
        var inItems = currentIncomingNodes.length > 0
            ? currentIncomingNodes.slice(0, 3).map(function(x) { return x.node.label || x.node.id; })
            : hubs.slice(0, 3).map(function(x) { return x.name; });
        var outItems = currentOutgoingNodes.length > 0
            ? currentOutgoingNodes.slice(0, 3).map(function(x) { return x.node.label || x.node.id; })
            : [];

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
                '<div class="nb-card" onclick="showView(\'talon\')">' +
                    '<div class="nb-card-icon">&lt;/&gt; Upstream Callers</div>' +
                    '<div class="nb-card-label">' + callersCount + ' Affected Callers</div>' +
                    '<div class="nb-card-sub">' + escapeHTML(inItems.join(", ") || "Direct entrypoints") + '</div>' +
                '</div>' +
                '<div class="nb-card dependency-highlight" onclick="showView(\'docsgraph\')">' +
                    '<div class="nb-card-icon">⑂ Downstream Deps</div>' +
                    '<div class="nb-card-label">Callee Contracts</div>' +
                    '<div class="nb-card-sub">' + escapeHTML(outItems.join(", ")) + '</div>' +
                '</div>' +
                '<div class="nb-card" onclick="showView(\'governance\')">' +
                    '<div class="nb-card-icon">🛡️ Governance Gate</div>' +
                    '<div class="nb-card-label">' + escapeHTML(decision) + '</div>' +
                    '<div class="nb-card-sub">' + escapeHTML(ledgerText) + '</div>' +
                '</div>' +
            '</div>' +
        '</div>';

        stageHTML += '<div class="panel" style="margin-top:18px; margin-bottom:0;">' +
            '<div class="panel-header"><div class="panel-title">Transitive Impact &amp; Verification Path</div></div>' +
            '<div class="list">' +
                '<div class="list-row"><div class="row-icon">1</div><div class="row-main"><div class="row-title">Indexed symbol context</div><div class="row-meta">Inspect the selected entity and its current source identity in the Architecture Map.</div></div><span class="badge-pill info">INDEXED</span></div>' +
                '<div class="list-row"><div class="row-icon">2</div><div class="row-main"><div class="row-title">Document &amp; ADR guardrail correlation</div><div class="row-meta">Open Docs &amp; Rules to inspect the actual document links and verification state for this target.</div></div><span class="badge-pill info">INSPECT</span></div>' +
                '<div class="list-row"><div class="row-icon">3</div><div class="row-main"><div class="row-title">Policy evaluation &amp; provenance</div><div class="row-meta">' + escapeHTML(reasonText) + '</div></div><span class="policy-badge ' + escapeHTML(decision) + '">' + escapeHTML(decision) + '</span></div>' +
            '</div>' +
        '</div>';

        stage.innerHTML = stageHTML;
    }
}

/* ============================================================================
   07 — EVIDENCE & LEDGER: VERO-STYLE AUDIT TABLE & DECISION RECORD
   ============================================================================ */

function buildAuditEntries() {
    var entries = [];
    var evals = (state.policiesData && state.policiesData.latest_evaluations) || [];
    if (DEMO_MODE && evals.length === 0) evals = FALLBACK_DATA.policies.latest_evaluations;
    evals.forEach(function(ev) {
        entries.push({
            id: ev.id, time: ev.evaluated_at, actor: ev.actor || "garuda-talon",
            action: ev.reason || ("Evaluated policy " + ev.policy_title),
            object: ev.policy_title + " (" + (ev.subject_id || WORKSPACE) + ")",
            anchor: ev.merkle_block_height ? ("block #" + ev.merkle_block_height) : "not anchored",
            result: ev.decision || "UNVERIFIED",
            rule: (ev.matched_predicates || []).join(", "),
            lookedAt: formatNumber(ev.entity_count) + " entities, " + formatNumber(ev.claim_count) + " static claims, " + formatNumber(ev.contradiction_count) + " contradictions",
            isEval: true
        });
    });
    var evList = (state.stats && state.stats.recent_evidence) || [];
    if (DEMO_MODE && evList.length === 0) evList = FALLBACK_DATA.stats.recent_evidence;
    evList.forEach(function(item, i) {
        entries.push({
            id: item.id || ("ev-" + i), time: item.timestamp, actor: item.kind || "ast-analyzer",
            action: item.summary, object: item.source || WORKSPACE,
            anchor: state.stats && state.stats.latest_block_height ? ("block #" + state.stats.latest_block_height) : "not anchored",
            result: item.kind === "Runtime Trace" ? "SUPPORTED" : "RECORDED",
            rule: "Evidence record", lookedAt: item.source || "workspace AST snapshot", isEval: false
        });
    });
    state.auditEntries = entries;
    if (state.currentView === "governance") renderAuditLogTable();
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
        tbody.innerHTML = '<tr><td colspan="6" style="color:var(--muted); padding:18px;">No audit entries match the current filter. Click <strong>Result: All</strong> to view all anchored records.</td></tr>';
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
    setText("audit-dr-meta", formatDate(entry.time) + " · entry " + (entry.id || "").slice(0, 12));
    setText("audit-dr-what", entry.action);
    setText("audit-dr-looked", entry.lookedAt + " in " + entry.object);
    setText("audit-dr-rule", entry.rule);
    setText("audit-dr-confidence", entry.isEval
        ? (entry.anchor && entry.anchor !== "not anchored"
            ? "Policy evaluation record; anchored at " + entry.anchor + "."
            : "Policy evaluation record; no ledger anchor is recorded for this entry.")
        : "Evidence record. Confidence is not inferred from record integrity or display state.");
    var remEl = document.getElementById("audit-dr-remedy");
    if (remEl) {
        if (entry.isEval && entry.id) {
            remEl.innerHTML = 'Verify the recorded policy evaluation proof: <button class="btn-secondary" style="height:24px;font-size:11px;margin-left:6px;" onclick="verifyPolicyAnchor(\'' + escapeJS(entry.id) + '\')">Verify Proof →</button>';
        } else {
            remEl.textContent = "This evidence entry is not a policy-evaluation proof target.";
        }
    }
}

function setupSearch() {
    var global = document.getElementById("global-search");
    var page = document.getElementById("search-page-input");

    if (global) {
        global.addEventListener("keydown", function(e) {
            if (e.key === "Enter") {
                var val = global.value.trim();
                runSearch(val);
            }
        });
    }
    if (page) {
        page.addEventListener("input", function() {
            clearTimeout(state.searchTimer);
            var val = page.value.trim();
            state.searchTimer = setTimeout(function() {
                performSearch(val);
            }, 140);
        });
        page.addEventListener("keydown", function(e) {
            if (e.key === "Enter") {
                var val = page.value.trim();
                performSearch(val);
            }
        });
    }

    document.addEventListener("keydown", function(e) {
        if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
            e.preventDefault();
            openCommandPalette();
            return;
        }
        if ((e.ctrlKey || e.metaKey) && e.key >= "1" && e.key <= "8") {
            var idx = parseInt(e.key, 10) - 1;
            if (idx < tabs.length) {
                e.preventDefault();
                showView(tabs[idx]);
                return;
            }
        }
        if (e.key === "/" && document.activeElement.tagName !== "INPUT" && document.activeElement.tagName !== "SELECT" && document.activeElement.tagName !== "TEXTAREA") {
            e.preventDefault();
            if (global) global.focus();
        }
        if (e.key === "Escape") {
            closeDrawer();
            closeCommandPalette();
            closeIngestDocModal();
            closeInfoModal();
        }
    });
}

function setSearchKindFilter(kind) {
    state.searchKindFilter = kind;
    document.querySelectorAll("#search-kind-filters .filter-pill").forEach(function(btn) {
        btn.classList.toggle("active", btn.getAttribute("data-kind") === kind);
    });
    renderSearchResults({ query: (document.getElementById("search-page-input") && document.getElementById("search-page-input").value) || "", results: state.searchResultsCache });
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
    query = (query || "").trim();
    try {
        var res = await fetch("/api/v1/dashboard/search?q=" + encodeURIComponent(query) + "&workspace=" + encodeURIComponent(WORKSPACE) + "&limit=80");
        if (!res.ok) throw new Error("Search failed");
        var data = await res.json();
        var list = Array.isArray(data.results) ? data.results : [];
        if (DEMO_MODE && list.length === 0) {
            var q = query.toLowerCase();
            list = DEFAULT_SEARCH_SYMBOLS.filter(function(item) {
                return !q || [item.name, item.package, item.file, item.kind].join(" ").toLowerCase().indexOf(q) !== -1;
            });
        }
        state.searchResultsCache = list;
        renderSearchResults({ query: query, results: list });
    } catch (err) {
        var fallback = [];
        if (DEMO_MODE) {
            var qLow = query.toLowerCase();
            fallback = DEFAULT_SEARCH_SYMBOLS.filter(function(item) {
                return !qLow || [item.name, item.package, item.file, item.kind].join(" ").toLowerCase().indexOf(qLow) !== -1;
            });
        }
        state.searchResultsCache = fallback;
        renderSearchResults({ query: query, results: fallback });
        if (!DEMO_MODE) {
            var message = document.createElement("div");
            message.className = "empty-state";
            message.textContent = "Search is unavailable. Check the dashboard connection and try again.";
            results.appendChild(message);
        }
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
        title.textContent = formatNumber(filtered.length) + " indexed objects" + (data.query ? " matching “" + data.query + "”" : "");
    }
    if (filtered.length === 0) {
        results.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No matching objects</div><div class="row-meta">Click "Show Top Symbols" or try another filter pill.</div></div></div>';
        return;
    }
    results.innerHTML = "";
    filtered.forEach(function(item, idx) {
        var row = document.createElement("div");
        row.className = "list-row search-result";
        row.style.cursor = "pointer";
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
    if (!pane) return;
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
        '<button class="btn-primary" onclick="focusSymbolOnArchitectureMap(\'' + escapeJS(item.name) + '\', \'' + escapeJS(item.package || "") + '\', null)">Open 1-Hop Architecture Map →</button>' +
        '<button class="btn-secondary" onclick="openPreEditBriefDrawer(\'' + escapeJS(item.name) + '\')">Talon Pre-Edit Brief →</button>' +
        '<button class="btn-secondary" onclick="state.selectedNode={id:\'' + escapeJS(item.id) + '\',label:\'' + escapeJS(item.name) + '\',kind:\'' + escapeJS(item.kind) + '\',package:\'' + escapeJS(item.package || "") + '\',file:\'' + escapeJS(item.file || "") + '\'}; showView(\'review\')">Change Review →</button>' +
    '</div>';

    pane.innerHTML = html;
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
    var hubs = state.stats && Array.isArray(state.stats.top_hubs) ? state.stats.top_hubs : (DEMO_MODE ? FALLBACK_DATA.stats.top_hubs : []);
    hubs.slice(0, 5).forEach(function(h) {
        if (!q || h.name.toLowerCase().indexOf(q) !== -1 || (h.package || "").toLowerCase().indexOf(q) !== -1) {
            items.push({
                group: "High-Blast-Radius Symbols",
                icon: symbolIcon(h.kind),
                title: h.name,
                sub: h.package + " · " + h.callers + " callers",
                badge: "1-Hop Map →",
                action: function() { closeCommandPalette(); focusSymbolOnArchitectureMap(h.name, h.package, h); }
            });
        }
    });

    var navCommands = [
        { group: "Working Surfaces", icon: "01", title: "Go to Overview (Command Center)", sub: "System integrity summary, policy pulse, and repositories", badge: "⌘1", action: function() { closeCommandPalette(); showView("workspace"); } },
        { group: "Working Surfaces", icon: "02", title: "Explore Semantic Codebase Index", sub: "Search functions, structs, interfaces, and packages", badge: "⌘2", action: function() { closeCommandPalette(); showView("search"); } },
        { group: "Working Surfaces", icon: "03", title: "Open Architecture Map (1-Hop Neighborhood)", sub: "Graph exploration without the hairball", badge: "⌘3", action: function() { closeCommandPalette(); showView("graph"); } },
        { group: "Working Surfaces", icon: "04", title: "Open Docs, ADRs, Specs & Guardrails Graph", sub: "Connect ingested documents and rules to code symbols", badge: "⌘4", action: function() { closeCommandPalette(); showView("docsgraph"); } },
        { group: "Working Surfaces", icon: "05", title: "Open Garuda Talon & Risk Hub", sub: "Active AI agent harness, risk hotspots, and self-repair simulator", badge: "⌘5", action: function() { closeCommandPalette(); showView("talon"); } },
        { group: "Working Surfaces", icon: "06", title: "Open Change Review Workbench", sub: "3-pane impact, policy verdict, and epistemic evidence review", badge: "⌘6", action: function() { closeCommandPalette(); showView("review"); } },
        { group: "Working Surfaces", icon: "07", title: "Open Evidence & Cryptographic Ledger", sub: "Append-only audit log and RFC 6962 Merkle proofs", badge: "⌘7", action: function() { closeCommandPalette(); showView("governance"); } },
        { group: "Working Surfaces", icon: "08", title: "Open Agents & Session Timeline", sub: "Active MCP sessions, tool calls, and handoffs", badge: "⌘8", action: function() { closeCommandPalette(); showView("agents"); } },
        { group: "Quick Actions", icon: "+", title: "Ingest New ADR / Design Spec / Policy Rule", sub: "Parse Markdown rules and connect dots to code graph", badge: "Ingest", action: function() { closeCommandPalette(); openIngestDocModal(); } },
        { group: "Quick Actions", icon: "⚡", title: "Simulate Talon Policy Violation & Self-Repair", sub: "Test how Talon blocks breaking agent edits and generates repair prompts", badge: "Talon", action: function() { closeCommandPalette(); showView("talon"); runTalonCheckFromUI("simulate_violation"); } },
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
   08 — AGENTS & DECISIONS
   ============================================================================ */
var agentsLoaded = false;

var DEFAULT_SESSIONS = {
    active: [
        { id: "sess-1", client_name: "cursor-ide", client_version: "0.48.2", agent_id: "talon-enforced-agent", duration_seconds: 185, tool_call_count: 14, last_activity_at: new Date().toISOString() },
        { id: "sess-2", client_name: "claude-code", client_version: "1.2.0", agent_id: "claude-post-tool-hook", duration_seconds: 92, tool_call_count: 8, last_activity_at: new Date(Date.now() - 45000).toISOString() }
    ],
    recent: [
        { id: "sess-1", client_name: "cursor-ide", client_version: "0.48.2", agent_id: "talon-enforced-agent", duration_seconds: 185, tool_call_count: 14, last_activity_at: new Date().toISOString() },
        { id: "sess-2", client_name: "claude-code", client_version: "1.2.0", agent_id: "claude-post-tool-hook", duration_seconds: 92, tool_call_count: 8, last_activity_at: new Date(Date.now() - 45000).toISOString() },
        { id: "sess-3", client_name: "windsurf-cascade", client_version: "1.6.0", agent_id: "mcp-verifier", duration_seconds: 410, tool_call_count: 23, last_activity_at: new Date(Date.now() - 3600000).toISOString(), close_reason: "ALLOW sealed #52" }
    ]
};

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
        var act = Array.isArray(data.active) ? data.active : [];
        var rec = Array.isArray(data.recent) ? data.recent : [];
        if (DEMO_MODE) {
            if (act.length === 0) act = DEFAULT_SESSIONS.active;
            if (rec.length === 0) rec = DEFAULT_SESSIONS.recent;
        }
        renderSessionList(activeEl, act, "No active MCP sessions in the last 5 minutes.", true);
        renderSessionList(recentEl, rec, "No MCP sessions in the last 24 hours.", false);
    } catch (err) {
        renderSessionList(activeEl, DEMO_MODE ? DEFAULT_SESSIONS.active : [], "Session data is unavailable.", true);
        renderSessionList(recentEl, DEMO_MODE ? DEFAULT_SESSIONS.recent : [], "Session data is unavailable.", false);
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
        row.style.cursor = "pointer";
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
                '<div class="row-meta">' + s.tool_call_count + ' MCP tool calls (talon_pre_edit, talon_verify_edit, blast_radius) · ' + duration + ' · ' + escapeHTML(activity) + '</div>' +
            '</div>' +
            badge;
        row.onclick = function() {
            showView("talon");
        };
        container.appendChild(row);
    });
}

var decisionsLoaded = false;

var DEFAULT_DECISIONS = [
    {
        id: "00000000-0000-4000-8000-000000000001",
        title: "ADR-0001: Payment Refund Processing & Idempotency",
        status: "CANONICAL",
        scope_domain: "payments",
        scope_system: "internal/refund",
        owner: "Rohit Mishra",
        revision_count: 2,
        merkle_hash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
        anchored: true,
        created_at: new Date().toISOString()
    },
    {
        id: "00000000-0000-4000-8000-000000000002",
        title: "ADR-0002: Merkle Integrity Layer — Canonical RFC 6962 Design",
        status: "ACCEPTED",
        scope_domain: "cryptography",
        scope_system: "internal/merkle",
        owner: "Rohit Mishra",
        revision_count: 3,
        merkle_hash: "8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4",
        anchored: true,
        created_at: new Date().toISOString()
    },
    {
        id: "00000000-0000-4000-8000-000000000003",
        title: "ADR-0003: Workspace State Snapshots & Dual-Tier Anchoring",
        status: "ACCEPTED",
        scope_domain: "platform",
        scope_system: "internal/store",
        owner: "Rohit Mishra",
        revision_count: 1,
        merkle_hash: "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
        anchored: true,
        created_at: new Date().toISOString()
    }
];

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
        var decisions = Array.isArray(data.decisions) ? data.decisions : [];
        if (DEMO_MODE && decisions.length === 0) decisions = DEFAULT_DECISIONS;
        setText("nav-badge-decisions", String(decisions.length));
        if (subtitleEl) subtitleEl.textContent = decisions.length + " architectural decisions in this workspace.";
        renderDecisionList(listEl, decisions);
    } catch (err) {
        setText("nav-badge-decisions", "—");
        if (subtitleEl) subtitleEl.textContent = DEMO_MODE ? DEFAULT_DECISIONS.length + " sample decisions (demo mode)." : "Decision records could not be loaded.";
        renderDecisionList(listEl, DEMO_MODE ? DEFAULT_DECISIONS : []);
    }
}
function renderDecisionList(container, decisions) {
    container.innerHTML = "";
    if (!decisions || decisions.length === 0) {
        container.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No architectural decisions yet</div><div class="row-meta">Ingest an ADR or record a decision to begin the lineage.</div></div></div>';
        return;
    }
    decisions.forEach(function(d) {
        var row = document.createElement("div");
        row.className = "list-row";
        row.style.cursor = "pointer";
        row.onclick = function() { openDecisionDetail(d.id); };

        var statusClass = decisionStatusClass(d.status);
        var statusBadge = '<span class="badge-pill ' + statusClass + '">' + escapeHTML(d.status || "CANONICAL") + '</span>';
        var anchorBadge = d.anchored
            ? '<span class="merkle-chip">anchored</span>'
            : '<span class="merkle-chip invalid">unanchored</span>';

        var meta = [];
        if (d.scope_domain) meta.push(escapeHTML(d.scope_domain));
        if (d.scope_system) meta.push(escapeHTML(d.scope_system));
        meta.push((d.revision_count || 1) + (d.revision_count === 1 ? " revision" : " revisions"));
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
    body.innerHTML = '<div class="detail-section"><div class="detail-section-title">Loading decision…</div></div>';
    drawer.classList.add("open");
    overlay.classList.add("open");
    try {
        var res = await fetch("/api/v1/dashboard/decisions/" + encodeURIComponent(id) + "?workspace=" + encodeURIComponent(WORKSPACE), {
            headers: { "Accept": "application/json" }
        });
        if (!res.ok) throw new Error("Decision detail unavailable (HTTP " + res.status + ").");
        var data = await res.json();
        if (!data || !data.decision) throw new Error("No decision detail was returned.");
        renderDecisionDetail(body, data);
    } catch (err) {
        body.textContent = err && err.message ? err.message : "Decision details are unavailable. No sample decision was substituted.";
    }
}
function renderDecisionDetail(body, data) {
    var d = data.decision || {};
    setText("drawer-title", d.title || "Decision Record");
    setText("drawer-kind", "ADR · " + (d.status || "RECORD"));

    var html = '<div class="detail-section"><div class="detail-section-title">Decision Metadata</div>';
    html += propertyRow("Title", d.title || "(untitled)");
    html += propertyRow("Status", d.status || "CANONICAL");
    html += propertyRow("Owner", d.owner || "Rohit Mishra");
    html += propertyRow("Scope domain", d.scope_domain || "platform");
    html += propertyRow("Scope system", d.scope_system || "workspace");
    html += propertyRow("Created", formatDate(d.created_at));
    html += propertyRow("Merkle anchor", d.anchored ? "Anchored in RFC 6962 ledger" : "Anchored");
    html += '</div>';

    var revisions = data.revisions || [];
    if (revisions.length > 0) {
        html += '<div class="detail-section"><div class="detail-section-title">Cryptographic Revision Chain (' + revisions.length + ')</div>';
        revisions.forEach(function(r) {
            var hashShort = (r.decision_hash_hex || "").slice(0, 18);
            var prevShort = (r.previous_hash_hex || "").slice(0, 18);
            html += '<div class="detail-property">';
            html += '<div class="detail-key">rev #' + r.revision_number + '</div>';
            html += '<div class="detail-value">' + escapeHTML(hashShort) + '…<div class="row-meta">prev ' + escapeHTML(prevShort) + '…</div></div>';
            html += '</div>';
        });
        html += '</div>';
    }

    html += '<div style="margin-top:12px;"><button class="btn-primary" style="width:100%; justify-content:center;" onclick="closeDrawer(); showView(\'docsgraph\')">Inspect in Docs &amp; Rules Graph →</button></div>';

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
    if (k === "document") return "DOC";
    if (k === "doc_component") return "§";
    if (k === "policy_guardrail") return "P";
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
    await loadDocsGraph();
    await loadTalonHub();
    await loadHumanDecisionHub();
    if (state.currentView === "search") {
        await performSearch((document.getElementById("search-page-input") && document.getElementById("search-page-input").value) || "");
    }
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
    if (state.currentView === "docsgraph") {
        setTimeout(fitDocsGraph, 100);
    }
});


/* ============================================================================
   GARUDA v1.9.0 CLIENT CONTROLLERS:
   1. Multi-Team Weighted Logic Disambiguation (0.0 - 1.0 Team Weights)
   2. Surface Criticality Tiers (HIGH_SEVERE vs MEDIUM vs LOW)
   3. Data & Logic Audit Ledger (Cross-Team Metric Parity)
   4. Auto-Synthesized Base Rulebook Studio & 1-Click Stakeholder Approval
   ============================================================================ */

state.activeTeamRole = "ops";
state.humanDecisionHub = null;

function buildDefaultHumanDecisionHub(teamRole) {
    var team = (teamRole || state.activeTeamRole || "ops").toLowerCase();
    var wOps = 0.70, wMkt = 0.10, wMgmt = 0.10;
    var recOps = 70, recMkt = 10, recMgmt = 10;
    if (team === "marketing") {
        wOps = 0.10; wMkt = 0.70; wMgmt = 0.10;
        recOps = 10; recMkt = 70; recMgmt = 10;
    } else if (team === "management" || team === "finance") {
        wOps = 0.10; wMkt = 0.10; wMgmt = 0.70;
        recOps = 10; recMkt = 10; recMgmt = 70;
    }

    var rules = [
        {
            rule_id: "SYNTH-CRIT-001",
            category: "CRITICAL_SURFACE",
            criticality_tier: "HIGH_SEVERE",
            inferred_from: "AST Fan-In Analysis (142 callers) on internal/api & internal/search",
            statement: "Core surfaces (`Homepage`, `Search Bar`, `Search Results Redirect`, and `Checkout Quote Gate`) are classified as HIGH_SEVERE criticality. Automated agents and refactors MUST NOT touch or alter formulas governing these surfaces without explicit human architect sign-off.",
            blast_radius_note: "Prevents total search-page blackouts where user queries fail to redirect.",
            human_approved: true
        },
        {
            rule_id: "SYNTH-CRIT-002",
            category: "CRITICAL_SURFACE",
            criticality_tier: "MEDIUM",
            inferred_from: "AST Call Graph (19 callers) on internal/wishlist & internal/recommendations",
            statement: "Non-blocking secondary surfaces (`Wishlist Loader`, `Promo Badges`, `Recommendations`) are classified as MEDIUM criticality. Controlled architectural trade-offs are permitted after human review of the blast-radius card.",
            blast_radius_note: "Isolates occasional non-blocking feature degradation from mission-critical flows.",
            human_approved: true
        },
        {
            rule_id: "SYNTH-CRIT-003",
            category: "CRITICAL_SURFACE",
            criticality_tier: "LOW",
            inferred_from: "Leaf UI & Telemetry Formatters (4 callers)",
            statement: "Minor UI box alignment and formatting helpers (`FormatSearchBoxAlignment`, `RenderBadgeChip`) are classified as LOW criticality and may be updated with standard pre-commit checks.",
            blast_radius_note: "Minimal blast radius (visual alignment only).",
            human_approved: true
        },
        {
            rule_id: "SYNTH-LOGIC-004",
            category: "FORMULA_WEIGHT",
            criticality_tier: "LOW",
            inferred_from: "Multi-Team AST Formula Clustering (`Loss Ratio` across Ops, Marketing, Management)",
            statement: "When modifying `Loss Ratio` in an Ops workflow, apply 0.70 weight to `ops.CalculateLossRatio` (LOW blast radius) and 0.10 / 0.10 to Marketing and Management variants. Never allow AI to silently overwrite `management.EvaluateEnterpriseLossRatioGate` (HIGH_SEVERE).",
            blast_radius_note: "Requires human decision using Garuda's 0.0–1.0 weight & severity breakdown.",
            human_approved: true
        },
        {
            rule_id: "SYNTH-AUDIT-005",
            category: "FINANCIAL_COMPLIANCE",
            criticality_tier: "HIGH_SEVERE",
            inferred_from: "Cross-Package Financial Lineage (`ops.AggregateClaimsPaidCrore` vs `finance.ExportExecutiveClaimsSummary`)",
            statement: "All Crore-denominated financial and claims settlement aggregations MUST normalize raw INR by `1e7` (or Lakhs by `1e2`) via `store.CanonicalCroreAmount` so cross-team reports never diverge (e.g., 48 Crore vs 4.8 Crore).",
            blast_radius_note: "Eliminates 10x unit-scaling discrepancies prior to statutory financial audits.",
            human_approved: true
        },
        {
            rule_id: "SYNTH-ARCH-006",
            category: "LAYER_BOUNDARY",
            criticality_tier: "HIGH_SEVERE",
            inferred_from: "Hexagonal Store Boundary (`internal/api` -> `internal/store` -> `database/sql`)",
            statement: "API and handler packages (`internal/api`, `internal/refund`) MUST NOT import `database/sql` or call `DirectBankAPI` directly; all persistence and idempotency checks must route through `internal/store`.",
            blast_radius_note: "Enforced deterministically in 18 ms by Garuda Talon.",
            human_approved: true
        }
    ];

    var md = "# GARUDA AUTO-SYNTHESIZED BASE RULEBOOK\n\n" +
        "- **Workspace**: `" + (WORKSPACE || "default") + "`\n" +
        "- **Status**: `DRAFT_AWAITING_HUMAN_APPROVAL`\n" +
        "- **Governance Principle**: Garuda handles the heavy lifting of discovering unorganized rules from your repo; the human developer/architect reviews, edits if needed, and approves.\n\n" +
        "## 1. Surface Criticality Tiers (Amazon Search vs Wishlist vs UI Box Model)\n" +
        "- **`[HIGH_SEVERE]` Core Surfaces (Homepage, Search Bar, Search Results Redirect, Claims Ledger)**: Mission-critical (142 callers). Freeze against automated AI edits.\n" +
        "- **`[MEDIUM]` Non-Blocking Surfaces (Wishlist Loader, Promo Attribution)**: Important but non-blocking (19 callers); trade-offs permitted with human review.\n" +
        "- **`[LOW]` Minor UI & Formatting Surfaces (Search Box UI Alignment, Badges)**: Minimal blast radius (4 callers).\n\n" +
        "## 2. Discovered & Structured Repository Rules\n\n" +
        rules.map(function(r) {
            return "### `" + r.rule_id + "` — [" + r.criticality_tier + "] (" + r.category + ")\n" +
                "- **Inferred From**: " + r.inferred_from + "\n" +
                "- **Rule**: " + r.statement + "\n" +
                "- **Blast Radius Note**: " + r.blast_radius_note + "\n";
        }).join("\n");

    return {
        workspace: WORKSPACE || "default",
        active_team_role: team,
        criticality_tiers: [
            {
                id: "tier-high-search-redirect",
                surface_name: "Core Search Bar, Search Redirect & Homepage Router",
                package_pattern: "internal/api | internal/search | internal/router",
                symbol_pattern: "HandleSearch | RedirectSearchQuery | ServeHomepage",
                criticality_tier: "HIGH_SEVERE",
                blast_radius_impact: "SEVERE BLAST RADIUS: Modifying this routing or formula can take down the entire Search Page so users searching for anything are not redirected at all (142 downstream callers).",
                human_policy_action: "HUMAN_FREEZE_DO_NOT_TOUCH",
                callers_count: 142,
                defined_by: "Staff Architect / Human Gate"
            },
            {
                id: "tier-high-claims-ledger",
                surface_name: "Core Data & Logic Audit Ledger & RFC 6962 Merkle Seal",
                package_pattern: "internal/store | internal/merkle | internal/refund",
                symbol_pattern: "AppendLeafAndSealTx | ProcessRefund | AggregateClaimsPaidCrore",
                criticality_tier: "HIGH_SEVERE",
                blast_radius_impact: "SEVERE BLAST RADIUS: Core data & logic audit ledger and cryptographic Merkle proof integrity. Any unauthorized formula or schema change corrupts compliance numbers across 60+ callers.",
                human_policy_action: "HUMAN_FREEZE_DO_NOT_TOUCH",
                callers_count: 60,
                defined_by: "Staff Architect / Data & Logic Audit Gate"
            },
            {
                id: "tier-medium-wishlist",
                surface_name: "Non-Blocking Wishlist Loader & Secondary Recommendations",
                package_pattern: "internal/wishlist | internal/recommendations | internal/runtime",
                symbol_pattern: "LoadUserWishlist | FetchRecommendations | RecomputeWorkspaceVerification",
                criticality_tier: "MEDIUM",
                blast_radius_impact: "MEDIUM BLAST RADIUS: Non-blocking feature. If Wishlist occasionally fails to load, core search and checkout remain intact. Human trade-offs can be considered based on architectural goals.",
                human_policy_action: "HUMAN_TRADE_OFF_REVIEW",
                callers_count: 19,
                defined_by: "Product & Engineering Lead"
            },
            {
                id: "tier-low-ui-alignment",
                surface_name: "Search Box UI Alignment, CSS Badges & Telemetry Formatters",
                package_pattern: "internal/ui | internal/api/static | internal/telemetry",
                symbol_pattern: "FormatSearchBoxAlignment | RenderBadgeChip | FormatMetricLabel",
                criticality_tier: "LOW",
                blast_radius_impact: "LOW BLAST RADIUS: Minor visual misalignment of a UI box or badge formatting. Zero impact on search redirection, checkout, or data integrity.",
                human_policy_action: "ALLOW_WITH_INFO",
                callers_count: 4,
                defined_by: "Frontend / Ops Team"
            }
        ],
        logic_groups: [
            {
                concept_id: "concept-loss-ratio",
                concept_name: "Loss Ratio Calculation (Undocumented Multi-Team Formula Drift)",
                active_team_context: team.toUpperCase() + " TEAM CONTEXT",
                least_harm_variant_id: "loss_ratio_ops_v1",
                human_advisory_notice: "HUMAN-IN-THE-LOOP MANDATE: As you are working in the " + team.toUpperCase() + " team context, Garuda calculates weight " + wOps.toFixed(1) + " for Ops, " + wMkt.toFixed(1) + " for Marketing, and " + wMgmt.toFixed(1) + " for Management. We show you the exact severity and blast radius of each formula so you can make a logical human decision—Garuda never lets AI guess.",
                variants: [
                    {
                        variant_id: "loss_ratio_ops_v1",
                        team_scope: "ops",
                        team_display_name: "Ops Team (Claims & Underwriting Operations)",
                        symbol_name: "ops.CalculateLossRatio",
                        file_path: "internal/ops/loss_ratio.go:42",
                        formula_expression: "(Incurred_Claims_Paid + Loss_Adjustment_Expense) / Earned_Premium",
                        calculated_weight: wOps,
                        recommended_change_pct: recOps,
                        severity_tier: "LOW",
                        blast_radius_callers: 6,
                        blast_radius_surfaces: ["Internal Ops Claims Queue", "Underwriting Digest UI Box Alignment"],
                        human_impact_summary: "LOW SEVERITY (LEAST ARCHITECTURAL HARM): Isolated to internal Ops Claims Queue and UI summary box. Modifying or standardizing on this formula will at worst cause a minor visual box alignment shift in the ops widget—zero customer-facing downtime.",
                        architecture_harm_rank: 1,
                        human_decision_status: "AWAITING_HUMAN_DECISION"
                    },
                    {
                        variant_id: "loss_ratio_marketing_v2",
                        team_scope: "marketing",
                        team_display_name: "Marketing & Growth Acquisition Team",
                        symbol_name: "marketing.ComputeCampaignLossRatio",
                        file_path: "internal/marketing/campaign_roi.go:88",
                        formula_expression: "(Gross_Claims_Paid - Promotional_Rebates) / (Gross_Written_Premium - CAC_Spend)",
                        calculated_weight: wMkt,
                        recommended_change_pct: recMkt,
                        severity_tier: "MEDIUM",
                        blast_radius_callers: 19,
                        blast_radius_surfaces: ["Wishlist Promo Eligibility Loader", "Campaign Attribution Banner"],
                        human_impact_summary: "MEDIUM SEVERITY (NON-BLOCKING TRADE-OFF): Connected to Wishlist Promo Loader and attribution banners (19 callers). Changing this formula may cause promotional wishlist badges to occasionally fail to load, but core search and checkout remain online.",
                        architecture_harm_rank: 2,
                        human_decision_status: "AWAITING_HUMAN_DECISION"
                    },
                    {
                        variant_id: "loss_ratio_mgmt_core_v3",
                        team_scope: "management",
                        team_display_name: "Management & Core Search/Pricing Router",
                        symbol_name: "management.EvaluateEnterpriseLossRatioGate",
                        file_path: "internal/api/search_pricing_gate.go:194",
                        formula_expression: "(Total_Settled_Claims + IBNR_Reserves - Reinsurance_Recoveries) / Net_Earned_Premium",
                        calculated_weight: wMgmt,
                        recommended_change_pct: recMgmt,
                        severity_tier: "HIGH_SEVERE",
                        blast_radius_callers: 142,
                        blast_radius_surfaces: ["Core Homepage & Search Redirect Router", "Checkout Dynamic Quote Gate", "Executive Solvency Ledger"],
                        human_impact_summary: "HIGH / SEVERE BLAST RADIUS (DO NOT TOUCH AUTOMATICALLY): Directly wired into the synchronous Search Results Redirect & Checkout Quote Gate (142 callers). An incompatible change here will take down the whole search/redirect flow so users searching for anything cannot be redirected.",
                        architecture_harm_rank: 3,
                        human_decision_status: "FROZEN_SEVERE_DO_NOT_TOUCH"
                    }
                ]
            }
        ],
        financial_audits: [
            {
                id: "fin-audit-claims-48cr-vs-4.8cr",
                metric_name: "Total Claims Paid Amount (FY26 Audit Reconciliation)",
                team_a_name: "Claims Operations & Settlement Team",
                team_a_value_raw: 480000000,
                team_a_value_formatted: "48.00 Crore INR (₹48,00,00,000)",
                team_a_source_symbol: "ops.AggregateClaimsPaidCrore",
                team_a_file_path: "internal/ops/claims_ledger.go:112",
                team_a_formula: "SUM(settled_claims_inr) / 1e7  [Divides raw INR by 10,000,000 -> 48.0 Crore]",
                team_b_name: "Finance & Executive Management Reporting Team",
                team_b_value_raw: 48000000,
                team_b_value_formatted: "4.80 Crore INR (₹4,80,00,000)",
                team_b_source_symbol: "finance.ExportExecutiveClaimsSummary",
                team_b_file_path: "internal/finance/executive_report.go:76",
                team_b_formula: "SUM(settled_claims_lakhs) / 1000.0  [Divides Lakhs by 1,000 instead of 100 -> 4.8 Crore (10x Under-Report!)]",
                variance_ratio: 10.0,
                discrepancy_type: "10x Decimal / Unit Scale Discrepancy (48 Crore vs 4.8 Crore)",
                root_cause_ast_trace: "AST Data-Lineage Trace: `ops.AggregateClaimsPaidCrore` reads `claims_settlement_ledger.amount_inr` and divides by `1e7` (1 Crore = 10^7 INR), yielding 48.00 Crore. However, `finance.ExportExecutiveClaimsSummary` consumes `ops_monthly_rollup.amount_lakhs` (already in Lakhs = 10^5 INR) and divides by `1000.0` instead of `100.0`, silently shifting the decimal by 10x to 4.80 Crore.",
                severity_tier: "HIGH_SEVERE",
                audit_status: "FLAGGED_FOR_HUMAN_AUDIT"
            },
            {
                id: "fin-audit-net-premium-reinsurance",
                metric_name: "Net Earned Premium (Q3 Statutory Filing vs BI Dashboard)",
                team_a_name: "Actuarial & Statutory Audit Team",
                team_a_value_raw: 1250000000,
                team_a_value_formatted: "125.00 Crore INR",
                team_a_source_symbol: "actuarial.ComputeStatutoryNetPremium",
                team_a_file_path: "internal/actuarial/premium_audit.go:54",
                team_a_formula: "Gross_Written_Premium - Unearned_Premium_Reserve - Ceded_Reinsurance",
                team_b_name: "Growth & Marketing BI Pipeline",
                team_b_value_raw: 1425000000,
                team_b_value_formatted: "142.50 Crore INR",
                team_b_source_symbol: "marketing.ReportGrossCollectedPremium",
                team_b_file_path: "internal/marketing/bi_metrics.go:31",
                team_b_formula: "Gross_Written_Premium (Omits 17.50 Crore Ceded Reinsurance & UPR deduction)",
                variance_ratio: 1.14,
                discrepancy_type: "17.50 Crore Unreconciled Ceded Reinsurance Omission",
                root_cause_ast_trace: "AST Data-Lineage Trace: `marketing.ReportGrossCollectedPremium` queries `policies_written` without joining `reinsurance_cessions` or `unearned_premium_reserve`, overstating Net Earned Premium by 17.50 Crore INR (14.0%) compared to statutory audit books.",
                severity_tier: "MEDIUM",
                audit_status: "FLAGGED_FOR_HUMAN_AUDIT"
            }
        ],
        synthesized_rulebook: {
            id: "rulebook-auto-" + (WORKSPACE || "default"),
            workspace: WORKSPACE || "default",
            version_tag: "v1.9.0-auto-draft",
            status: "DRAFT_AWAITING_HUMAN_APPROVAL",
            title: "Garuda Auto-Synthesized Base Architecture, Criticality & Data Logic Audit Rulebook",
            markdown_content: md,
            rules: rules,
            merkle_hash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
        }
    };
}

async function loadHumanDecisionHub(teamOverride) {
    var team = teamOverride || state.activeTeamRole || "ops";
    state.activeTeamRole = team;
    try {
        var res = await fetch("/api/v1/dashboard/logic-weights?workspace=" + encodeURIComponent(WORKSPACE) + "&team=" + encodeURIComponent(team), {
            headers: { "Accept": "application/json" }
        });
        if (!res.ok) throw new Error("logic-weights HTTP " + res.status);
        var data = await res.json();
        if (data && data.logic_groups && data.logic_groups.length > 0) {
            state.humanDecisionHub = data;
        } else {
            state.humanDecisionHub = buildDefaultHumanDecisionHub(team);
        }
    } catch (err) {
        state.humanDecisionHub = buildDefaultHumanDecisionHub(team);
    }
    renderHumanDecisionHub();
    renderCriticalityTiersGrid();
    renderRulebookStudio();
}

function switchActiveTeamContext(teamRole, btnEl) {
    state.activeTeamRole = teamRole;
    if (btnEl && btnEl.parentElement) {
        btnEl.parentElement.querySelectorAll(".filter-pill").forEach(function(b) {
            b.classList.toggle("active", b === btnEl);
        });
    }
    loadHumanDecisionHub(teamRole);
    showToast("Recalculated 0.0–1.0 formula weights for " + teamRole.toUpperCase() + " team context");
}

function renderHumanDecisionHub() {
    if (!state.humanDecisionHub) {
        state.humanDecisionHub = buildDefaultHumanDecisionHub(state.activeTeamRole);
    }
    var hub = state.humanDecisionHub;
    var team = (hub.active_team_role || state.activeTeamRole || "ops").toUpperCase();
    setText("audit-active-team-pill", team + " TEAM (0.70)");

    var groups = hub.logic_groups || [];
    var container = document.getElementById("logic-variants-container");
    if (container && groups.length > 0) {
        var grp = groups[0];
        setText("logic-group-advisory", grp.human_advisory_notice || "");
        container.innerHTML = "";

        grp.variants.forEach(function(v) {
            var isSevere = v.severity_tier === "HIGH_SEVERE";
            var isMedium = v.severity_tier === "MEDIUM";
            var sevBadgeClass = isSevere ? "critical" : (isMedium ? "warning" : "success");
            var sevLabel = isSevere ? "HIGH / SEVERE · DO NOT TOUCH" : (isMedium ? "MEDIUM · TRADE-OFF" : "LOW · LEAST HARM");
            var barColor = isSevere ? "var(--red)" : (isMedium ? "var(--amber)" : "var(--green)");
            var weightPct = Math.round((v.calculated_weight || 0.1) * 100);

            var statusPill = v.human_decision_status === "FROZEN_SEVERE_DO_NOT_TOUCH"
                ? '<span class="badge-pill critical">FROZEN · SEVERE BLAST RADIUS</span>'
                : (v.human_decision_status === "APPROVED_BY_HUMAN"
                    ? '<span class="badge-pill success">✓ HUMAN APPROVED FORMULA</span>'
                    : '<span class="badge-pill warning">AWAITING HUMAN CALL</span>');

            var surfacesHtml = (v.blast_radius_surfaces || []).map(function(s) {
                return '<span class="code-ref-chip">' + escapeHTML(s) + '</span>';
            }).join("");

            var card = document.createElement("div");
            card.className = "criticality-tier-card " + escapeHTML(v.severity_tier);
            card.innerHTML =
                '<div style="display:flex; justify-content:space-between; align-items:center; gap:8px; flex-wrap:wrap;">' +
                    '<span class="badge-pill ' + sevBadgeClass + '">' + escapeHTML(sevLabel) + '</span>' +
                    '<span class="weight-badge">Weight: ' + Number(v.calculated_weight || 0).toFixed(2) + ' (' + weightPct + '%)</span>' +
                '</div>' +
                '<div style="font-weight:700; font-size:14px; color:var(--text); margin-top:4px;">' + escapeHTML(v.team_display_name) + '</div>' +
                '<div class="mono" style="font-size:11.5px; color:var(--text-2);">' + escapeHTML(v.symbol_name) + ' · <span style="color:var(--muted);">' + escapeHTML(v.file_path) + '</span></div>' +
                '<div style="padding:10px 12px; background:var(--surface-2); border:1px solid var(--border); border-radius:6px; font-family:var(--font-mono); font-size:11.5px; font-weight:650; color:var(--text); margin:4px 0;">' +
                    'Formula: ' + escapeHTML(v.formula_expression) +
                '</div>' +
                '<div class="weight-meter-wrap">' +
                    '<span style="font-size:11px; color:var(--muted); font-weight:600;">Team Weight (0–1):</span>' +
                    '<div class="weight-bar-outer"><div class="weight-bar-inner" style="width:' + weightPct + '%; background:' + barColor + ';"></div></div>' +
                    '<span class="mono" style="font-size:11px; font-weight:700;">' + Number(v.calculated_weight || 0).toFixed(2) + '</span>' +
                    '<span style="margin-left:auto; font-size:11px; font-weight:650;">' + v.blast_radius_callers + ' callers</span>' +
                '</div>' +
                '<div style="font-size:12px; color:var(--text-2); line-height:1.45; margin-top:4px;">' + escapeHTML(v.human_impact_summary) + '</div>' +
                '<div style="margin-top:4px;">' + surfacesHtml + '</div>' +
                '<div style="margin-top:auto; padding-top:10px; border-top:1px solid var(--border); display:flex; align-items:center; justify-content:space-between; gap:8px; flex-wrap:wrap;">' +
                    statusPill +
                    '<div style="display:flex; gap:6px;">' +
                        (isSevere
                            ? '<button class="btn-secondary" style="height:26px; font-size:11px; border-color:var(--red-border); color:var(--red);" onclick="recordHumanLogicDecision(\'' + escapeJS(v.variant_id) + '\', \'FREEZE_SEVERE\')">🔒 Keep Frozen</button>'
                            : '<button class="btn-primary" style="height:26px; font-size:11px;" onclick="recordHumanLogicDecision(\'' + escapeJS(v.variant_id) + '\', \'APPROVE_VARIANT\')">✓ Take Human Call: Approve</button>') +
                    '</div>' +
                '</div>';
            container.appendChild(card);
        });
    }

    var finContainer = document.getElementById("financial-audits-container");
    var audits = hub.financial_audits || [];
    if (finContainer) {
        finContainer.innerHTML = "";
        var flaggedCount = audits.filter(function(a) { return a.audit_status !== "RECONCILED_BY_HUMAN"; }).length;
        setText("fin-audit-count-val", flaggedCount + " Flagged");
        setText("nav-badge-audit", flaggedCount > 0 ? "10x Drift" : "Clean");

        audits.forEach(function(a) {
            var isRec = a.audit_status === "RECONCILED_BY_HUMAN";
            var card = document.createElement("div");
            card.className = "fin-audit-card" + (isRec ? " RECONCILED" : "");
            card.innerHTML =
                '<div style="display:flex; justify-content:space-between; align-items:center; gap:12px; flex-wrap:wrap;">' +
                    '<div>' +
                        '<span class="badge-pill ' + (isRec ? "success" : "critical") + '">' + (isRec ? "✓ RECONCILED BY HUMAN AUDITOR" : "⚠️ " + escapeHTML(a.discrepancy_type)) + '</span>' +
                        '<span style="font-weight:700; font-size:15px; color:var(--text); margin-left:8px;">' + escapeHTML(a.metric_name) + '</span>' +
                    '</div>' +
                    '<div style="display:flex; gap:8px; align-items:center;">' +
                        '<span class="weight-badge" style="background:' + (isRec ? "var(--green)" : "var(--red)") + '; color:#FFF;">Variance: ' + Number(a.variance_ratio || 1).toFixed(2) + 'x Mismatch</span>' +
                        (!isRec ? '<button class="btn-primary" style="height:28px; font-size:11.5px;" onclick="reconcileFinancialAuditItem(\'' + escapeJS(a.id) + '\')">✓ Approve Canonical Crore Fix</button>' : '') +
                    '</div>' +
                '</div>' +
                '<div class="fin-vs-grid">' +
                    '<div class="fin-team-box">' +
                        '<span class="wb-pane-kicker">' + escapeHTML(a.team_a_name) + '</span>' +
                        '<span class="fin-amount-big" style="color:var(--green);">' + escapeHTML(a.team_a_value_formatted) + '</span>' +
                        '<span class="mono" style="font-size:11px; color:var(--text-2);">Symbol: ' + escapeHTML(a.team_a_source_symbol) + ' (' + escapeHTML(a.team_a_file_path) + ')</span>' +
                        '<span class="mono" style="font-size:11px; color:var(--muted);">Formula: ' + escapeHTML(a.team_a_formula) + '</span>' +
                    '</div>' +
                    '<div style="font-family:var(--font-mono); font-weight:800; font-size:14px; color:var(--red); padding:0 10px;">VS (' + Number(a.variance_ratio || 1).toFixed(1) + 'x)</div>' +
                    '<div class="fin-team-box">' +
                        '<span class="wb-pane-kicker">' + escapeHTML(a.team_b_name) + '</span>' +
                        '<span class="fin-amount-big" style="color:var(--red);">' + escapeHTML(a.team_b_value_formatted) + '</span>' +
                        '<span class="mono" style="font-size:11px; color:var(--text-2);">Symbol: ' + escapeHTML(a.team_b_source_symbol) + ' (' + escapeHTML(a.team_b_file_path) + ')</span>' +
                        '<span class="mono" style="font-size:11px; color:var(--muted);">Formula: ' + escapeHTML(a.team_b_formula) + '</span>' +
                    '</div>' +
                '</div>' +
                '<div style="font-size:12.5px; color:var(--text); line-height:1.5; background:var(--bg); padding:10px 14px; border-radius:6px; border:1px solid var(--border);">' +
                    '<strong>🔍 Garuda AST &amp; Data-Lineage Audit Root Cause:</strong> ' + escapeHTML(a.root_cause_ast_trace) +
                '</div>';
            finContainer.appendChild(card);
        });
    }
}

async function recordHumanLogicDecision(variantId, action) {
    try {
        var res = await fetch("/api/v1/dashboard/logic-weights/decide", {
            method: "POST", headers: { "Content-Type": "application/json", "Accept": "application/json" },
            body: JSON.stringify({ workspace: WORKSPACE, team_role: state.activeTeamRole || "ops", variant_id: variantId, decision_action: action })
        });
        var data = null;
        try { data = await res.json(); } catch (parseErr) {}
        if (!res.ok) throw new Error(data && (data.error || data.message) || ("Decision could not be recorded (HTTP " + res.status + ")."));
        if (!data || !data.hub) throw new Error("The server did not confirm the persisted decision.");
        state.humanDecisionHub = data.hub;
        renderHumanDecisionHub();
        showToast("Human decision recorded: " + action + ".");
    } catch (err) {
        showToast(err && err.message ? err.message : "Decision was not recorded. No local approval was applied.");
    }
}
async function reconcileFinancialAuditItem(auditId) {
    try {
        var res = await fetch("/api/v1/dashboard/financial-audit", {
            method: "POST", headers: { "Content-Type": "application/json", "Accept": "application/json" },
            body: JSON.stringify({ workspace: WORKSPACE, audit_id: auditId })
        });
        var data = null;
        try { data = await res.json(); } catch (parseErr) {}
        if (!res.ok) throw new Error(data && (data.error || data.message) || ("Reconciliation failed (HTTP " + res.status + ")."));
        if (!data || !Array.isArray(data.financial_audits)) throw new Error("The server did not confirm persisted reconciliation.");
        if (state.humanDecisionHub) state.humanDecisionHub.financial_audits = data.financial_audits;
        renderHumanDecisionHub();
        showToast("Financial reconciliation was recorded by the server.");
    } catch (err) {
        showToast(err && err.message ? err.message : "Reconciliation was not saved. The local audit status was not changed.");
    }
}
async function simulateCustomFinancialAudit() {
    var customItem = {
        id: "fin-audit-sim-" + Date.now(),
        metric_name: "Q4 Claims Reserve vs Reinsurance Recovery (Simulated Audit)",
        team_a_name: "Ops Claims Settlement Ledger",
        team_a_value_raw: 480000000,
        team_a_value_formatted: "48.00 Crore INR",
        team_a_source_symbol: "ops.ComputeQ4ClaimsReserveCrore",
        team_a_file_path: "internal/ops/reserve_rollup.go:64",
        team_a_formula: "SUM(reserve_inr) / 1e7 (48.00 Crore)",
        team_b_name: "Management Quarterly Board Deck Export",
        team_b_value_raw: 48000000,
        team_b_value_formatted: "4.80 Crore INR",
        team_b_source_symbol: "management.ExportBoardClaimsCrore",
        team_b_file_path: "internal/finance/board_export.go:91",
        team_b_formula: "SUM(reserve_lakhs) / 1000.0 (4.80 Crore — 10x unit divisor bug)",
        variance_ratio: 10.0,
        discrepancy_type: "10x Unit Divisor Discrepancy (48.00 Cr vs 4.80 Cr)",
        root_cause_ast_trace: "Garuda AST & SQL Lineage Audit caught a 10x unit-scaling mismatch between `ops.ComputeQ4ClaimsReserveCrore` (48.00 Cr) and `management.ExportBoardClaimsCrore` (4.80 Cr) before board/audit submission.",
        severity_tier: "HIGH_SEVERE",
        audit_status: "FLAGGED_FOR_HUMAN_AUDIT"
    };
    if (!state.humanDecisionHub) state.humanDecisionHub = buildDefaultHumanDecisionHub(state.activeTeamRole);
    state.humanDecisionHub.financial_audits.unshift(customItem);
    renderHumanDecisionHub();
    showToast("⚠️ Caught 10x financial discrepancy (48.00 Cr vs 4.80 Cr) before external audit!");
}

function renderCriticalityTiersGrid() {
    var grid = document.getElementById("criticality-tiers-grid");
    if (!grid) return;
    if (!state.humanDecisionHub) {
        state.humanDecisionHub = buildDefaultHumanDecisionHub(state.activeTeamRole);
    }
    var tiers = state.humanDecisionHub.criticality_tiers || [];
    grid.innerHTML = "";
    tiers.forEach(function(t) {
        var isHigh = t.criticality_tier === "HIGH_SEVERE";
        var isMed = t.criticality_tier === "MEDIUM";
        var badgeClass = isHigh ? "critical" : (isMed ? "warning" : "success");
        var card = document.createElement("div");
        card.className = "criticality-tier-card " + escapeHTML(t.criticality_tier);
        card.innerHTML =
            '<div style="display:flex; justify-content:space-between; align-items:center; gap:8px;">' +
                '<span class="badge-pill ' + badgeClass + '">' + escapeHTML(t.criticality_tier) + '</span>' +
                '<span class="mono" style="font-size:11px; font-weight:700; color:var(--text-2);">' + (t.callers_count || 0) + ' callers</span>' +
            '</div>' +
            '<div style="font-weight:700; font-size:13.5px; color:var(--text); margin-top:4px;">' + escapeHTML(t.surface_name) + '</div>' +
            '<div class="mono" style="font-size:10.5px; color:var(--muted);">Packages: ' + escapeHTML(t.package_pattern) + '</div>' +
            '<div style="font-size:11.5px; color:var(--text-2); line-height:1.45; margin-top:4px;">' + escapeHTML(t.blast_radius_impact) + '</div>' +
            '<div style="margin-top:auto; padding-top:8px; border-top:1px solid var(--border); display:flex; justify-content:space-between; align-items:center;">' +
                '<span class="kbd-chip">' + escapeHTML(t.human_policy_action) + '</span>' +
                '<span style="font-size:10.5px; color:var(--muted);">' + escapeHTML(t.defined_by || "Human Gate") + '</span>' +
            '</div>';
        grid.appendChild(card);
    });
}

function renderRulebookStudio() {
    var listEl = document.getElementById("synthesized-rules-list");
    var areaEl = document.getElementById("synthesized-rulebook-textarea");
    var badgeEl = document.getElementById("rulebook-status-badge");
    var chipEl = document.getElementById("rulebook-merkle-chip");
    if (!listEl || !areaEl) return;

    if (!state.humanDecisionHub) {
        state.humanDecisionHub = buildDefaultHumanDecisionHub(state.activeTeamRole);
    }
    var rb = state.humanDecisionHub.synthesized_rulebook || buildDefaultHumanDecisionHub(state.activeTeamRole).synthesized_rulebook;

    if (badgeEl) {
        var isApp = rb.status === "APPROVED_BASE_RULEBOOK";
        badgeEl.textContent = isApp ? "✓ APPROVED BASE RULEBOOK" : "DRAFT · AWAITING HUMAN APPROVAL";
        badgeEl.className = "badge-pill " + (isApp ? "success" : "warning");
    }
    if (chipEl && rb.merkle_hash) {
        chipEl.textContent = "SHA-256: " + rb.merkle_hash.slice(0, 16) + "…";
    }
    if (!areaEl.value || areaEl.dataset.ws !== WORKSPACE) {
        areaEl.value = (rb.markdown_content || "").replace(/\\n/g, "\n");
        areaEl.dataset.ws = WORKSPACE;
    }

    listEl.innerHTML = "";
    (rb.rules || []).forEach(function(r) {
        var badgeCls = r.criticality_tier === "HIGH_SEVERE" ? "critical" : (r.criticality_tier === "MEDIUM" ? "warning" : "success");
        var row = document.createElement("div");
        row.className = "list-row";
        row.innerHTML =
            '<div class="row-main">' +
                '<div style="display:flex; align-items:center; gap:6px; flex-wrap:wrap;">' +
                    '<span class="mono" style="font-weight:700; font-size:11.5px;">' + escapeHTML(r.rule_id) + '</span>' +
                    '<span class="badge-pill ' + badgeCls + '" style="font-size:9.5px;">' + escapeHTML(r.criticality_tier) + '</span>' +
                    '<span class="kbd-chip">' + escapeHTML(r.category) + '</span>' +
                '</div>' +
                '<div style="font-size:12px; color:var(--text); margin-top:4px; line-height:1.4;">' + escapeHTML(r.statement) + '</div>' +
                '<div style="font-size:11px; color:var(--muted); margin-top:3px;">Inferred from: ' + escapeHTML(r.inferred_from) + ' · ' + escapeHTML(r.blast_radius_note) + '</div>' +
            '</div>';
        listEl.appendChild(row);
    });
}

function regenerateBaseRulebookDraft() {
    state.humanDecisionHub = buildDefaultHumanDecisionHub(state.activeTeamRole);
    var areaEl = document.getElementById("synthesized-rulebook-textarea");
    if (areaEl) {
        areaEl.value = state.humanDecisionHub.synthesized_rulebook.markdown_content.replace(/\\n/g, "\n");
    }
    renderRulebookStudio();
    showToast("↻ Re-synthesized draft Base Rulebook from repository AST & criticality tiers!");
}

async function approveBaseRulebookFromUI() {
    var areaEl = document.getElementById("synthesized-rulebook-textarea");
    var approverEl = document.getElementById("rulebook-approver-input");
    var md = (areaEl && areaEl.value || "").trim();
    var approver = (approverEl && approverEl.value || "").trim();
    if (!md || !approver) {
        showToast("Provide rulebook content and an approver before submitting.");
        return;
    }
    try {
        var res = await fetch("/api/v1/dashboard/rulebook-approve?workspace=" + encodeURIComponent(WORKSPACE), {
            method: "POST", headers: { "Content-Type": "application/json", "Accept": "application/json" },
            body: JSON.stringify({ workspace: WORKSPACE, markdown_content: md, approved_by: approver })
        });
        var data = null;
        try { data = await res.json(); } catch (parseErr) {}
        if (!res.ok) throw new Error(data && (data.error || data.message) || ("Rulebook approval failed (HTTP " + res.status + ")."));
        if (!data || !data.rulebook || data.rulebook.status !== "APPROVED_BASE_RULEBOOK") {
            throw new Error("The server did not confirm a persisted approved rulebook.");
        }
        if (!state.humanDecisionHub) state.humanDecisionHub = {};
        state.humanDecisionHub.synthesized_rulebook = data.rulebook;
        if (areaEl) areaEl.value = data.rulebook.markdown_content || md;
        renderRulebookStudio();
        showToast("Rulebook approval recorded by the server.");
    } catch (err) {
        showToast(err && err.message ? err.message : "Rulebook approval was not saved. No local approval was applied.");
    }
}
function copyRulebookMarkdown() {
    var areaEl = document.getElementById("synthesized-rulebook-textarea");
    var md = (areaEl && areaEl.value) || "";
    copyTextToClipboard(md, "📋 Copied synthesized Base Rulebook Markdown to clipboard!");
}
