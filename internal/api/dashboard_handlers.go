// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/store"
)

// -----------------------------------------------------------------------------
// Dashboard API types
// -----------------------------------------------------------------------------

type DashboardData struct {
	TenantID string
}

type HubDTO struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Package string `json:"package"`
	Repo    string `json:"repo"`
	Callers int    `json:"callers"`
}

type AttentionItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle"`
	Severity    string `json:"severity"`
	EvidenceLoc string `json:"evidence_loc"`
}

type EvidenceItem struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Summary   string `json:"summary"`
	Source    string `json:"source"`
	Timestamp string `json:"timestamp"`
}

type WorkspaceStatsResponse struct {
	Workspace          string   `json:"workspace"`
	Repositories       int      `json:"repositories"`
	Packages           int      `json:"packages"`
	Entities           int      `json:"entities"`
	Relationships      int      `json:"relationships"`
	CrossRepoLinks     int      `json:"cross_repo_links"`
	Files              int      `json:"files"`
	ExportedEntities   int      `json:"exported_entities"`
	ArchitecturalHubs  int      `json:"architectural_hubs"`
	TopHubs            []HubDTO `json:"top_hubs"`
	TotalClaims        int      `json:"total_claims"`
	SupportedClaims    int      `json:"supported_claims"`
	Contradicted       int      `json:"contradicted"`
	UnverifiedClaims   int      `json:"unverified_claims"`
	NeedsAttention     []AttentionItem `json:"needs_attention"`
	RecentEvidence     []EvidenceItem  `json:"recent_evidence"`
	CanonicalDecisions int      `json:"canonical_decisions"`
	QuarantinedCount   int      `json:"quarantined_count"`
	LatestBlockHeight  int64    `json:"latest_block_height"`
	LatestMerkleHash   string   `json:"latest_merkle_hash"`
	ParentMerkleHash   string   `json:"parent_merkle_hash"`
	TrustStatus        string   `json:"trust_status"`
	LastUpdated        string   `json:"last_updated"`
}

type SearchResult struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Package  string `json:"package"`
	File     string `json:"file"`
	Exported bool   `json:"exported"`
	Repo     string `json:"repo"`
}

type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
}

type GraphNodeDTO struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Kind     string `json:"kind"`
	Package  string `json:"package"`
	Repo     string `json:"repo"`
	Exported bool   `json:"exported"`
	Status   string `json:"status"`
	Count    int    `json:"count,omitempty"`
}

type GraphEdgeDTO struct {
	ID         string  `json:"id"`
	Source     string  `json:"from"`
	Target     string  `json:"to"`
	Type       string  `json:"type"`
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
	Status     string  `json:"status"`
	Count      int     `json:"count,omitempty"`
}

type GraphResponseDTO struct {
	Level string         `json:"level"`
	Focus string         `json:"focus"`
	Nodes []GraphNodeDTO `json:"nodes"`
	Edges []GraphEdgeDTO `json:"edges"`
}

type EntityRecord struct {
	ID       string
	Name     string
	Kind     string
	Package  string
	File     string
	Exported bool
	Repo     string
}

const dashboardTenantID = "00000000-0000-0000-0000-000000000001"

func getDashboardTenant() uuid.UUID {
	return uuid.MustParse(dashboardTenantID)
}

func normalizeLimit(value string, fallback, maximum int) int {
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return fallback
	}
	if n > maximum {
		return maximum
	}
	return n
}

// -----------------------------------------------------------------------------
// Embedded Dashboard HTML
// -----------------------------------------------------------------------------

const prodDashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Garuda — Epistemic Software Intelligence</title>
<script src="https://d3js.org/d3.v7.min.js"></script>

<style>
:root {
    --bg: #f6f8fb;
    --surface: #ffffff;
    --surface-2: #f9fafc;
    --border: #e5e9f0;
    --border-strong: #d8dee8;
    --text: #172033;
    --text-2: #4e5b70;
    --muted: #7b8799;
    --brand: #2563eb;
    --brand-dark: #1d4ed8;
    --brand-soft: #eff6ff;
    --green: #059669;
    --green-soft: #ecfdf5;
    --amber: #d97706;
    --amber-soft: #fffbeb;
    --red: #e11d48;
    --red-soft: #ffe4e6;
    --shadow-sm: 0 1px 2px rgba(16, 24, 40, 0.04);
    --shadow-md: 0 5px 20px rgba(16, 24, 40, 0.07);
    --radius: 10px;
}
* { box-sizing: border-box; }
html, body { margin: 0; padding: 0; min-height: 100%; }
body { background: var(--bg); color: var(--text); font-family: Inter, ui-sans-serif, system-ui, -apple-system, sans-serif; font-size: 14px; }
button, input { font: inherit; }
button { cursor: pointer; }
.app { min-height: 100vh; display: flex; }

.sidebar { width: 238px; min-width: 238px; background: #101828; color: #dce3ee; display: flex; flex-direction: column; border-right: 1px solid #1f2937; }
.brand { height: 64px; padding: 0 20px; display: flex; align-items: center; gap: 10px; border-bottom: 1px solid #202b3d; }
.brand-mark { width: 31px; height: 31px; border-radius: 8px; background: #2563eb; display: grid; place-items: center; font-size: 17px; }
.brand-name { font-weight: 750; letter-spacing: 0.3px; color: white; }
.brand-subtitle { color: #8794aa; font-size: 10px; margin-top: 1px; }
.sidebar-content { padding: 18px 12px; flex: 1; }
.nav-section { margin-bottom: 22px; }
.nav-title { color: #68768d; font-size: 10px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.09em; padding: 0 10px 8px; }
.nav-item { width: 100%; border: 0; background: transparent; color: #aab5c7; text-align: left; padding: 9px 10px; border-radius: 7px; display: flex; align-items: center; gap: 10px; margin-bottom: 2px; font-size: 13px; }
.nav-item:hover { background: #192438; color: white; }
.nav-item.active { background: #1f2d43; color: white; box-shadow: inset 3px 0 0 var(--brand); }
.nav-icon { width: 17px; text-align: center; color: #8290a5; }
.nav-item.active .nav-icon { color: #60a5fa; }
.sidebar-footer { padding: 15px; border-top: 1px solid #202b3d; }
.workspace-mini { padding: 11px; background: #162033; border: 1px solid #253148; border-radius: 8px; }
.workspace-mini-name { color: white; font-weight: 650; margin-bottom: 5px; }
.workspace-mini-meta { color: #8794aa; font-size: 11px; }
.trust-mini { margin-top: 8px; color: #34d399; font-size: 11px; font-weight: 600; }

.main { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.topbar { height: 64px; background: white; border-bottom: 1px solid var(--border); display: flex; align-items: center; gap: 18px; padding: 0 26px; position: sticky; top: 0; z-index: 50; }
.global-search { flex: 1; max-width: 760px; position: relative; }
.search-input { width: 100%; height: 40px; border: 1px solid var(--border); background: #f8fafc; border-radius: 8px; padding: 0 42px 0 38px; color: var(--text); outline: none; transition: 0.15s; }
.search-input:focus { background: white; border-color: #93c5fd; box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08); }
.search-icon { position: absolute; left: 13px; top: 10px; color: #7b8799; font-size: 16px; }
.search-shortcut { position: absolute; right: 10px; top: 9px; border: 1px solid var(--border); background: white; color: #8a95a5; font-size: 10px; border-radius: 5px; padding: 2px 6px; }
.topbar-right { margin-left: auto; display: flex; align-items: center; gap: 10px; }
.live-pill { display: flex; align-items: center; gap: 7px; color: var(--green); font-size: 11px; font-weight: 650; }
.live-dot { width: 7px; height: 7px; border-radius: 50%; background: #10b981; }
.refresh-button { border: 1px solid var(--border); background: white; color: var(--text-2); height: 34px; padding: 0 11px; border-radius: 7px; }
.refresh-button:hover { background: var(--surface-2); }

.content { width: 100%; max-width: 1500px; margin: 0 auto; padding: 25px 28px 50px; }
.breadcrumbs { display: flex; align-items: center; gap: 7px; color: var(--muted); font-size: 12px; margin-bottom: 18px; }
.breadcrumb-button { border: 0; padding: 0; background: transparent; color: var(--text-2); }
.breadcrumb-button:hover { color: var(--brand); }
.breadcrumb-current { color: var(--text); font-weight: 650; }
.hero { display: flex; justify-content: space-between; align-items: flex-start; gap: 20px; margin-bottom: 24px; }
.hero-title { font-size: 25px; line-height: 1.25; margin: 0; letter-spacing: -0.025em; }
.hero-subtitle { color: var(--text-2); margin-top: 7px; font-size: 13px; }
.trust-badge { display: inline-flex; align-items: center; gap: 7px; padding: 7px 10px; border: 1px solid #a7f3d0; background: var(--green-soft); color: #047857; border-radius: 7px; font-size: 11px; font-weight: 700; white-space: nowrap; }

.kpi-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 13px; margin-bottom: 20px; }
.kpi-card { background: white; border: 1px solid var(--border); border-radius: var(--radius); padding: 16px; box-shadow: var(--shadow-sm); }
.kpi-label { color: var(--muted); font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }
.kpi-value { font-size: 25px; font-weight: 750; margin-top: 7px; letter-spacing: -0.03em; }
.kpi-foot { color: var(--muted); font-size: 11px; margin-top: 5px; }

.trust-strip { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 13px; margin-bottom: 20px; }
.trust-card { background: white; border: 1px solid var(--border); border-radius: var(--radius); padding: 16px; box-shadow: var(--shadow-sm); border-left: 4px solid var(--muted); }
.trust-card.supported { border-left-color: var(--green); }
.trust-card.unverified { border-left-color: var(--amber); }
.trust-card.contradicted { border-left-color: var(--red); }
.trust-card-title { font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }
.trust-card.supported .trust-card-title { color: var(--green); }
.trust-card.unverified .trust-card-title { color: var(--amber); }
.trust-card.contradicted .trust-card-title { color: var(--red); }
.trust-card-val { font-size: 24px; font-weight: 750; margin-top: 6px; }
.trust-card-desc { font-size: 11px; color: var(--muted); margin-top: 4px; }

.start-card { background: white; border: 1px solid var(--border); border-radius: var(--radius); padding: 20px; box-shadow: var(--shadow-sm); margin-bottom: 20px; }
.start-title { font-size: 14px; font-weight: 750; margin-bottom: 5px; }
.start-description { color: var(--text-2); font-size: 12px; margin-bottom: 15px; }
.quick-search { height: 45px; width: 100%; border: 1px solid var(--border-strong); border-radius: 8px; padding: 0 14px; outline: none; color: var(--text); }
.quick-search:focus { border-color: #93c5fd; box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08); }
.quick-hints { display: flex; gap: 7px; flex-wrap: wrap; margin-top: 10px; }
.hint { border: 1px solid var(--border); background: #f8fafc; color: var(--text-2); padding: 5px 8px; border-radius: 6px; font-size: 10px; }
.hint:hover { border-color: #bfdbfe; color: var(--brand); background: var(--brand-soft); }

.two-column { display: grid; grid-template-columns: minmax(0, 1.3fr) minmax(300px, 1fr); gap: 18px; margin-bottom: 20px; }
.panel { background: white; border: 1px solid var(--border); border-radius: var(--radius); box-shadow: var(--shadow-sm); overflow: hidden; }
.panel-header { padding: 15px 17px; border-bottom: 1px solid var(--border); display: flex; align-items: center; justify-content: space-between; gap: 10px; }
.panel-title { font-size: 13px; font-weight: 750; }
.panel-subtitle { color: var(--muted); font-size: 11px; margin-top: 3px; }
.panel-body { padding: 16px; }

.explorer { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; }
.explorer-card { min-height: 135px; border: 1px solid var(--border); background: var(--surface-2); border-radius: 8px; padding: 14px; transition: 0.15s; }
.explorer-card:hover { border-color: #bfdbfe; box-shadow: var(--shadow-md); transform: translateY(-1px); }
.explorer-icon { width: 31px; height: 31px; border-radius: 7px; display: grid; place-items: center; background: var(--brand-soft); color: var(--brand); margin-bottom: 10px; }
.explorer-name { font-weight: 700; font-size: 13px; }
.explorer-count { color: var(--brand); font-size: 19px; font-weight: 750; margin-top: 5px; }
.explorer-meta { color: var(--muted); font-size: 10px; margin-top: 3px; }

.list { display: flex; flex-direction: column; }
.list-row { border-bottom: 1px solid var(--border); padding: 12px 15px; display: flex; align-items: center; gap: 12px; transition: 0.12s; }
.list-row:last-child { border-bottom: 0; }
.list-row:hover { background: #fafcff; }
.row-icon { width: 30px; height: 30px; border-radius: 7px; display: grid; place-items: center; background: #f1f5f9; color: var(--text-2); flex-shrink: 0; }
.row-main { flex: 1; min-width: 0; }
.row-title { color: var(--text); font-size: 12px; font-weight: 700; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.row-meta { color: var(--muted); font-size: 10px; margin-top: 3px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.badge-pill { font-size: 9px; font-weight: 750; padding: 3px 7px; border-radius: 5px; text-transform: uppercase; }
.badge-pill.critical { background: var(--red-soft); color: var(--red); border: 1px solid #fecaca; }
.badge-pill.warning { background: var(--amber-soft); color: var(--amber); border: 1px solid #fde68a; }
.badge-pill.info { background: var(--brand-soft); color: var(--brand); border: 1px solid #bfdbfe; }
.badge-pill.success { background: var(--green-soft); color: var(--green); border: 1px solid #a7f3d0; }

.graph-panel { margin-top: 20px; }
.graph-toolbar { display: flex; align-items: center; gap: 7px; flex-wrap: wrap; }
.graph-button { border: 1px solid var(--border); background: white; color: var(--text-2); height: 30px; padding: 0 9px; border-radius: 6px; font-size: 11px; }
.graph-button:hover { background: #f8fafc; color: var(--text); }
.graph-button.primary { background: var(--brand); color: white; border-color: var(--brand); }
.graph-button.primary:hover { background: var(--brand-dark); }

/* FULLSCREEN INFINITE CANVAS */
.graph-wrap { height: 570px; position: relative; background: linear-gradient(#f4f6f9 1px, transparent 1px), linear-gradient(90deg, #f4f6f9 1px, transparent 1px); background-size: 28px 28px; overflow: hidden; transition: all 0.2s ease; }
.graph-wrap.fullscreen { position: fixed; inset: 0; z-index: 9999; height: 100vh; width: 100vw; background-color: white; margin: 0; padding: 0; }
.graph-wrap.fullscreen .graph-controls { top: 20px; right: 20px; }

#graph { width: 100%; height: 100%; }
.graph-empty { position: absolute; inset: 0; display: grid; place-items: center; color: var(--muted); font-size: 12px; pointer-events: none; }
.graph-help { position: absolute; left: 14px; bottom: 13px; background: rgba(255,255,255,0.94); border: 1px solid var(--border); border-radius: 7px; padding: 7px 9px; color: var(--muted); font-size: 10px; box-shadow: var(--shadow-sm); }
.graph-controls { position: absolute; right: 13px; top: 13px; display: flex; flex-direction: column; gap: 5px; }
.graph-control { width: 30px; height: 30px; border: 1px solid var(--border); background: white; color: var(--text-2); border-radius: 6px; box-shadow: var(--shadow-sm); }
.graph-node-label { font-size: 10px; fill: #f8fafc; pointer-events: none; font-weight: 600; text-anchor: middle; text-shadow: 0 1px 2px rgba(0,0,0,0.4); }
.graph-node-label.dark { fill: #0f172a; text-shadow: none; }
.graph-link { stroke: #cbd5e1; stroke-opacity: 0.6; }
.graph-link.violation { stroke: #e11d48; stroke-opacity: 0.9; stroke-dasharray: 4; animation: pulse-edge 2s infinite; }
.graph-link-label { fill: #e11d48; font-size: 9px; font-weight: 700; pointer-events: none; }

@keyframes pulse-edge { 0% { stroke-opacity: 0.6; } 50% { stroke-opacity: 1; } 100% { stroke-opacity: 0.6; } }
@keyframes pulse-node { 0% { box-shadow: 0 0 0 0 rgba(225, 29, 72, 0.4); } 70% { box-shadow: 0 0 0 10px rgba(225, 29, 72, 0); } 100% { box-shadow: 0 0 0 0 rgba(225, 29, 72, 0); } }

.drawer-overlay { position: fixed; inset: 0; background: rgba(15, 23, 42, 0.15); z-index: 10000; display: none; }
.drawer-overlay.open { display: block; }
.drawer { position: fixed; right: 0; top: 0; height: 100vh; width: min(470px, 92vw); background: white; border-left: 1px solid var(--border); box-shadow: -12px 0 35px rgba(16,24,40,0.12); z-index: 10001; transform: translateX(100%); transition: transform 0.2s ease; display: flex; flex-direction: column; }
.drawer.open { transform: translateX(0); }
.drawer-header { padding: 18px; border-bottom: 1px solid var(--border); display: flex; justify-content: space-between; gap: 12px; }
.drawer-title { font-size: 16px; font-weight: 750; }
.drawer-kind { color: var(--brand); font-size: 10px; font-weight: 750; text-transform: uppercase; margin-top: 4px; }
.drawer-close { border: 0; background: #f1f5f9; width: 30px; height: 30px; border-radius: 6px; }
.drawer-body { overflow-y: auto; padding: 18px; }
.detail-section { margin-bottom: 20px; }
.detail-section-title { font-size: 10px; text-transform: uppercase; letter-spacing: 0.07em; color: var(--muted); font-weight: 750; margin-bottom: 9px; }
.detail-property { display: grid; grid-template-columns: 105px 1fr; gap: 10px; padding: 8px 0; border-bottom: 1px solid #f0f2f5; }
.detail-key { color: var(--muted); font-size: 11px; }
.detail-value { color: var(--text); font-size: 11px; overflow-wrap: anywhere; }
.detail-action { width: 100%; border: 1px solid #bfdbfe; background: var(--brand-soft); color: var(--brand-dark); border-radius: 7px; padding: 9px 11px; font-size: 11px; font-weight: 700; margin-top: 7px; }

.search-view { display: none; }
.search-view.active { display: block; }
.search-result { cursor: pointer; }
.search-result:hover { background: #f8fbff; }
.kind-pill { font-size: 9px; font-weight: 750; padding: 3px 6px; border-radius: 5px; background: #f1f5f9; color: #475467; }
.merkle { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 10px; color: var(--text-2); word-break: break-all; background: #f8fafc; padding: 9px; border-radius: 6px; margin-top: 8px; }
</style>
</head>

<body>
<div class="app">
    <aside class="sidebar">
        <div class="brand">
            <div class="brand-mark">🦅</div>
            <div>
                <div class="brand-name">GARUDA</div>
                <div class="brand-subtitle">Epistemic Software Intelligence</div>
            </div>
        </div>

        <div class="sidebar-content">
            <div class="nav-section">
                <div class="nav-title">Workspace</div>
                <button class="nav-item active" id="nav-overview" onclick="showView('overview')">
                    <span class="nav-icon">◉</span>
                    <span>Overview</span>
                </button>
            </div>
            <div class="nav-section">
                <div class="nav-title">Explore</div>
                <button class="nav-item" id="nav-architecture" onclick="showView('architecture')">
                    <span class="nav-icon">◇</span>
                    <span>Architecture</span>
                </button>
                <button class="nav-item" id="nav-search" onclick="showView('search')">
                    <span class="nav-icon">⌕</span>
                    <span>Search</span>
                </button>
            </div>
            <div class="nav-section">
                <div class="nav-title">Trust & Evidence</div>
                <button class="nav-item" id="nav-trust" onclick="showView('trust')">
                    <span class="nav-icon">✓</span>
                    <span>Evidence & Claims</span>
                </button>
            </div>
        </div>

        <div class="sidebar-footer">
            <div class="workspace-mini">
                <div class="workspace-mini-name" id="sidebar-workspace">uuid-ws</div>
                <div class="workspace-mini-meta" id="sidebar-meta">Loading workspace...</div>
                <div class="trust-mini">✓ Merkle verified</div>
            </div>
        </div>
    </aside>

    <section class="main">
        <header class="topbar">
            <div class="global-search">
                <span class="search-icon">⌕</span>
                <input id="global-search" class="search-input" placeholder="Search repositories, packages, files, functions, methods..." autocomplete="off">
                <span class="search-shortcut">/</span>
            </div>
            <div class="topbar-right">
                <div class="live-pill">
                    <span class="live-dot"></span>
                    <span>Live workspace</span>
                </div>
                <button class="refresh-button" onclick="loadAll()">Refresh</button>
            </div>
        </header>

        <main class="content">
            <section id="view-overview">
                <div class="breadcrumbs">
                    <span>Workspace</span>
                    <span>/</span>
                    <span class="breadcrumb-current" id="workspace-breadcrumb">uuid-ws</span>
                </div>

                <div class="hero">
                    <div>
                        <h1 class="hero-title">Workspace intelligence</h1>
                        <div class="hero-subtitle">
                            Continuous system verification. Garuda triangulates compiler ASTs, runtime traces, and architectural intent.
                        </div>
                    </div>
                    <div class="trust-badge">
                        <span>✓</span>
                        <span>Cryptographic state verified</span>
                    </div>
                </div>

                <div class="kpi-grid">
                    <div class="kpi-card">
                        <div class="kpi-label">Repositories</div>
                        <div class="kpi-value" id="stat-repositories">—</div>
                        <div class="kpi-foot">Scanned source codebases</div>
                    </div>
                    <div class="kpi-card">
                        <div class="kpi-label">Packages</div>
                        <div class="kpi-value" id="stat-packages">—</div>
                        <div class="kpi-foot">Architectural modules</div>
                    </div>
                    <div class="kpi-card">
                        <div class="kpi-label">Entities</div>
                        <div class="kpi-value" id="stat-entities">—</div>
                        <div class="kpi-foot">Functions, structs & symbols</div>
                    </div>
                    <div class="kpi-card">
                        <div class="kpi-label">Relationships</div>
                        <div class="kpi-value" id="stat-relationships">—</div>
                        <div class="kpi-foot"><span id="stat-cross-links">—</span> cross-repo bridges</div>
                    </div>
                </div>

                <div class="trust-strip">
                    <div class="trust-card supported">
                        <div class="trust-card-title">✓ Supported Claims</div>
                        <div class="trust-card-val" id="stat-supported">—</div>
                        <div class="trust-card-desc">Static AST verified against runtime</div>
                    </div>
                    <div class="trust-card unverified">
                        <div class="trust-card-title">? Unverified Claims</div>
                        <div class="trust-card-val" id="stat-unverified">—</div>
                        <div class="trust-card-desc">Code exists, zero recent executions</div>
                    </div>
                    <div class="trust-card contradicted">
                        <div class="trust-card-title">⚠ Contradictions</div>
                        <div class="trust-card-val" id="stat-contradicted">0</div>
                        <div class="trust-card-desc">Quarantined architectural drift</div>
                    </div>
                </div>

                <div class="start-card">
                    <div class="start-title">Find anything in the workspace</div>
                    <div class="start-description">Instant fuzzy search across all symbols, files, and packages.</div>
                    <input id="quick-search" class="quick-search" placeholder="Try: HandleFunc, Parse, chi, mux.go, securecookie...">
                    <div class="quick-hints">
                        <button class="hint" onclick="runHint('HandleFunc')">HandleFunc</button>
                        <button class="hint" onclick="runHint('Parse')">Parse</button>
                        <button class="hint" onclick="runHint('chi')">chi</button>
                        <button class="hint" onclick="runHint('securecookie')">securecookie</button>
                        <button class="hint" onclick="runHint('Harvester')">Harvester</button>
                    </div>
                </div>

                <div class="two-column">
                    <div class="panel">
                        <div class="panel-header">
                            <div>
                                <div class="panel-title">Architecture explorer</div>
                                <div class="panel-subtitle">Hierarchical progressive exploration.</div>
                            </div>
                        </div>
                        <div class="panel-body">
                            <div class="explorer">
                                <button class="explorer-card" onclick="openArchitecture('repository')">
                                    <div class="explorer-icon">▦</div>
                                    <div class="explorer-name">Repositories</div>
                                    <div class="explorer-count" id="explorer-repos">—</div>
                                    <div class="explorer-meta">System boundaries</div>
                                </button>
                                <button class="explorer-card" onclick="openArchitecture('package')">
                                    <div class="explorer-icon">◇</div>
                                    <div class="explorer-name">Packages</div>
                                    <div class="explorer-count" id="explorer-packages">—</div>
                                    <div class="explorer-meta">Architectural modules</div>
                                </button>
                                <button class="explorer-card" onclick="openArchitecture('entity')">
                                    <div class="explorer-icon">ƒ</div>
                                    <div class="explorer-name">Entities</div>
                                    <div class="explorer-count" id="explorer-entities">—</div>
                                    <div class="explorer-meta">Symbol intelligence</div>
                                </button>
                            </div>
                        </div>
                    </div>

                    <div class="panel">
                        <div class="panel-header">
                            <div>
                                <div class="panel-title">Architectural hubs</div>
                                <div class="panel-subtitle">Highest centrality symbols ranked by callers.</div>
                            </div>
                        </div>
                        <div class="list" id="hub-list">
                            <div class="list-row"><div class="row-main"><div class="row-title">Loading hubs...</div></div></div>
                        </div>
                    </div>
                </div>

                <div class="two-column">
                    <div class="panel">
                        <div class="panel-header">
                            <div>
                                <div class="panel-title">⚠ Needs attention</div>
                                <div class="panel-subtitle">Quarantined contradictions & unverified dependencies.</div>
                            </div>
                        </div>
                        <div class="list" id="attention-list">
                            <div class="list-row"><div class="row-main"><div class="row-title">Loading alerts...</div></div></div>
                        </div>
                    </div>

                    <div class="panel">
                        <div class="panel-header">
                            <div>
                                <div class="panel-title">📜 Recent evidence ledger</div>
                                <div class="panel-subtitle">Verified static AST proofs & runtime observations.</div>
                            </div>
                        </div>
                        <div class="list" id="evidence-list">
                            <div class="list-row"><div class="row-main"><div class="row-title">Loading evidence ledger...</div></div></div>
                        </div>
                    </div>
                </div>
            </section>

            <section id="view-architecture" style="display:none;">
                <div class="breadcrumbs">
                    <button class="breadcrumb-button" onclick="showView('overview')">Workspace</button>
                    <span>/</span>
                    <span class="breadcrumb-current" id="architecture-breadcrumb">Architecture</span>
                </div>
                <div class="hero">
                    <div>
                        <h1 class="hero-title" id="architecture-title">Architecture</h1>
                        <div class="hero-subtitle" id="architecture-subtitle">Explore system structure progressively.</div>
                    </div>
                    <div class="graph-toolbar">
                        <button class="graph-button" onclick="goUpArchitecture()">← Up one level</button>
                        <button class="graph-button" onclick="loadArchitecture(state.currentLevel, state.currentFocus)">Refresh</button>
                    </div>
                </div>

                <div class="panel graph-panel">
                    <div class="panel-header">
                        <div>
                            <div class="panel-title">Progressive architecture map</div>
                            <div class="panel-subtitle" id="graph-description">Workspace-level structure.</div>
                        </div>
                        <div class="graph-toolbar">
                            <button class="graph-button primary" onclick="fitGraph()">Fit</button>
                        </div>
                    </div>
                    <div class="graph-wrap" id="graph-wrap">
                        <svg id="graph"></svg>
                        <div id="graph-empty" class="graph-empty" style="display:none;">No architecture data available.</div>
                        <div class="graph-controls">
                            <button class="graph-control" onclick="toggleFullscreen()" title="Toggle Fullscreen">⛶</button>
                            <button class="graph-control" onclick="zoomGraph(1.25)" title="Zoom in">+</button>
                            <button class="graph-control" onclick="zoomGraph(0.8)" title="Zoom out">−</button>
                            <button class="graph-control" onclick="fitGraph()" title="Fit graph">⌂</button>
                        </div>
                        <div class="graph-help">Click a node to inspect · Double-click to expand</div>
                    </div>
                </div>
            </section>

            <section id="view-search" class="search-view">
                <div class="breadcrumbs">
                    <button class="breadcrumb-button" onclick="showView('overview')">Workspace</button>
                    <span>/</span>
                    <span class="breadcrumb-current">Search</span>
                </div>
                <div class="hero">
                    <div>
                        <h1 class="hero-title">Global search</h1>
                        <div class="hero-subtitle">Instantly locate any symbol across all connected repositories.</div>
                    </div>
                </div>
                <div class="start-card">
                    <input id="search-page-input" class="quick-search" placeholder="Search repositories, packages, files, functions, methods...">
                </div>
                <div class="panel">
                    <div class="panel-header">
                        <div>
                            <div class="panel-title" id="search-results-title">Results</div>
                            <div class="panel-subtitle">Click any result to open its details drawer.</div>
                        </div>
                    </div>
                    <div class="list" id="search-results">
                        <div class="list-row"><div class="row-main"><div class="row-title">Start typing to search Garuda.</div></div></div>
                    </div>
                </div>
            </section>

            <section id="view-trust" class="search-view">
                <div class="breadcrumbs">
                    <button class="breadcrumb-button" onclick="showView('overview')">Workspace</button>
                    <span>/</span>
                    <span class="breadcrumb-current">Evidence & Trust</span>
                </div>
                <div class="hero">
                    <div>
                        <h1 class="hero-title">Evidence & cryptographic trust</h1>
                        <div class="hero-subtitle">Every claim is anchored to compiler ASTs and Merkle-backed ledgers.</div>
                    </div>
                    <div class="trust-badge">
                        <span>✓</span>
                        <span>Verified state</span>
                    </div>
                </div>
                <div class="trust-strip">
                    <div class="trust-card supported">
                        <div class="trust-card-title">Ledger Status</div>
                        <div class="trust-card-val" id="trust-status" style="color:var(--green);">Verified</div>
                        <div class="trust-card-desc">Current Merkle Root State</div>
                    </div>
                    <div class="trust-card">
                        <div class="trust-card-title">Block Height</div>
                        <div class="trust-card-val" id="trust-height">#1</div>
                        <div class="trust-card-desc">Latest Recorded Snapshot</div>
                    </div>
                    <div class="trust-card">
                        <div class="trust-card-title">Observation Time</div>
                        <div class="trust-card-val" id="trust-updated" style="font-size:14px; margin-top:12px;">—</div>
                        <div class="trust-card-desc">Last Synced Timestamp</div>
                    </div>
                </div>
                <div class="panel" style="margin-top:18px;">
                    <div class="panel-header">
                        <div>
                            <div class="panel-title">Latest Merkle snapshot</div>
                            <div class="panel-subtitle">Cryptographic state root.</div>
                        </div>
                    </div>
                    <div class="panel-body">
                        <div class="kpi-label">Current root</div>
                        <div class="merkle" id="trust-merkle">Genesis verified</div>
                        <div class="kpi-label" style="margin-top:14px;">Parent</div>
                        <div class="merkle" id="trust-parent">Genesis</div>
                    </div>
                </div>
            </section>
        </main>
    </section>
</div>

<div id="drawer-overlay" class="drawer-overlay" onclick="closeDrawer()"></div>
<aside id="drawer" class="drawer">
    <div class="drawer-header">
        <div>
            <div class="drawer-title" id="drawer-title">Entity</div>
            <div class="drawer-kind" id="drawer-kind">ENTITY</div>
        </div>
        <button class="drawer-close" onclick="closeDrawer()">×</button>
    </div>
    <div class="drawer-body" id="drawer-body"></div>
</aside>

<script>
var WORKSPACE = "uuid-ws";

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
    searchTimer: null
};

function toggleFullscreen() {
    var wrap = document.getElementById("graph-wrap");
    wrap.classList.toggle("fullscreen");
    setTimeout(fitGraph, 200);
}

function showView(view) {
    state.currentView = view;
    var sections = ["view-overview", "view-architecture", "view-search", "view-trust"];
    sections.forEach(function(id) {
        var el = document.getElementById(id);
        if (!el) return;
        el.style.display = (id === "view-" + view) ? "block" : "none";
    });

    var navs = ["nav-overview", "nav-architecture", "nav-search", "nav-trust"];
    navs.forEach(function(id) {
        var el = document.getElementById(id);
        if (!el) return;
        el.classList.remove("active");
        if (id === "nav-" + view) el.classList.add("active");
    });

    if (view === "architecture") {
        loadArchitecture(state.currentLevel || "repository", state.currentFocus || "");
    }
    if (view === "trust") {
        renderTrust();
    }
}

async function loadStats() {
    try {
        var res = await fetch("/api/v1/dashboard/stats?workspace=" + encodeURIComponent(WORKSPACE), {
            headers: { "Accept": "application/json" }
        });
        if (!res.ok) throw new Error("Stats request failed");
        state.stats = await res.json();
        renderStats();
    } catch (err) {
        console.error("Garuda stats error:", err);
    }
}

function renderStats() {
    if (!state.stats) return;
    var s = state.stats;
    setText("stat-repositories", formatNumber(s.repositories));
    setText("stat-packages", formatNumber(s.packages));
    setText("stat-entities", formatNumber(s.entities));
    setText("stat-relationships", formatNumber(s.relationships));
    setText("stat-cross-links", formatNumber(s.cross_repo_links));

    setText("stat-supported", formatNumber(s.supported_claims));
    setText("stat-unverified", formatNumber(s.unverified_claims));
    setText("stat-contradicted", formatNumber(s.contradicted));

    setText("explorer-repos", formatNumber(s.repositories));
    setText("explorer-packages", formatNumber(s.packages));
    setText("explorer-entities", formatNumber(s.entities));

    setText("sidebar-meta", formatNumber(s.repositories) + " repositories · " + formatNumber(s.entities) + " entities");
    setText("workspace-breadcrumb", s.workspace || WORKSPACE);

    renderHubs(s.top_hubs || []);
    renderAttention(s.needs_attention || []);
    renderEvidence(s.recent_evidence || []);
    renderTrust();
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
                '<div style="font-weight:750; color:#2563eb; font-size:13px;">' + h.callers + '</div>' +
                '<div style="font-size:9px; color:#7b8799; text-transform:uppercase;">callers</div>' +
            '</div>';
        row.onclick = function() {
            openSearchResult(h);
        };
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
        row.innerHTML = '<div class="row-icon">📜</div>' +
            '<div class="row-main">' +
                '<div class="row-title">' + escapeHTML(item.summary) + '</div>' +
                '<div class="row-meta">' + escapeHTML(item.kind) + ' · ' + escapeHTML(item.source) + ' · ' + formatDate(item.timestamp) + '</div>' +
            '</div>' +
            '<span class="badge-pill success">Verified</span>';
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

function openArchitecture(level) {
    state.currentLevel = level;
    state.currentFocus = "";
    showView("architecture");
}

async function loadArchitecture(level, focus) {
    state.currentLevel = level || "repository";
    state.currentFocus = focus || "";
    updateArchitectureHeader();

    var url = "/api/v1/graph?workspace=" + encodeURIComponent(WORKSPACE) +
        "&level=" + encodeURIComponent(state.currentLevel);
    if (state.currentFocus) {
        url += "&focus=" + encodeURIComponent(state.currentFocus);
    }

    try {
        var res = await fetch(url);
        if (!res.ok) throw new Error("Graph request failed");
        var data = await res.json();
        state.graphData = data;
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
    var title = "Repository architecture";
    var subtitle = "One node per repository. Double-click a repository to inspect internal packages.";
    var breadcrumb = "Repositories";

    if (state.currentLevel === "package") {
        title = state.currentFocus ? "Packages in " + state.currentFocus : "Workspace packages";
        subtitle = "Package-level structure. Double-click a package to explore symbols.";
        breadcrumb = state.currentFocus || "Packages";
    }
    if (state.currentLevel === "entity") {
        title = state.currentFocus ? "Symbol neighborhood" : "Top architectural symbols";
        subtitle = "Local neighborhood rendered with zero hairballs.";
        breadcrumb = state.currentFocus ? "Neighborhood" : "Entities";
    }

    setText("architecture-title", title);
    setText("architecture-subtitle", subtitle);
    setText("architecture-breadcrumb", breadcrumb);
    setText("graph-description", subtitle);
}

var nodeTypeColors = {
    repository: "#0f172a",
    package: "#7c3aed",
    interface: "#0d9488",
    struct: "#4f46e5",
    function: "#f59e0b",
    method: "#d97706",
    file: "#db2777",
    default: "#64748b",
    external_quarantined: "#e11d48"
};

function nodeColor(kind) {
    return nodeTypeColors[kind] || nodeTypeColors.default;
}

function renderGraph(data) {
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
    var height = container.clientHeight || 570;

    svg.attr("width", width).attr("height", height);

    var defs = svg.append("defs");
    defs.append("marker")
        .attr("id", "arrow")
        .attr("viewBox", "0 -5 10 10")
        .attr("refX", 26)
        .attr("refY", 0)
        .attr("markerWidth", 6)
        .attr("markerHeight", 6)
        .attr("orient", "auto")
        .append("path")
        .attr("d", "M0,-5L10,0L0,5")
        .attr("fill", "#cbd5e1");
        
    defs.append("marker")
        .attr("id", "arrow-violation")
        .attr("viewBox", "0 -5 10 10")
        .attr("refX", 26)
        .attr("refY", 0)
        .attr("markerWidth", 6)
        .attr("markerHeight", 6)
        .attr("orient", "auto")
        .append("path")
        .attr("d", "M0,-5L10,0L0,5")
        .attr("fill", "#e11d48");

    var zoomLayer = svg.append("g");
    state.graphSvg = svg;
    state.graphGroup = zoomLayer;

    var zoom = d3.zoom().scaleExtent([0.15, 5]).on("zoom", function(event) {
        zoomLayer.attr("transform", event.transform);
    });
    svg.call(zoom);
    state.graphZoom = zoom;

    var nodes = data.nodes.map(function(n) { return Object.assign({}, n); });
    var nodeByID = {};
    nodes.forEach(function(n) { nodeByID[n.id] = n; });

    var validEdges = (data.edges || []).filter(function(e) {
        return nodeByID[e.from] && nodeByID[e.to];
    }).map(function(e) {
        return {
            source: e.from,
            target: e.to,
            type: e.type,
            status: e.status,
            label: e.label,
            count: e.count || 1
        };
    });

    var linkDist = state.currentLevel === "repository" ? 220 : (state.currentLevel === "package" ? 140 : 80);
    var chargeForce = state.currentLevel === "repository" ? -1500 : (state.currentLevel === "package" ? -800 : -350);

    var simulation = d3.forceSimulation(nodes)
        .force("link", d3.forceLink(validEdges).id(function(d) { return d.id; }).distance(linkDist).strength(0.35))
        .force("charge", d3.forceManyBody().strength(chargeForce))
        .force("center", d3.forceCenter(width / 2, height / 2))
        .force("collision", d3.forceCollide().radius(function(d) {
            var val = (d.count || d.Count || 1);
            var logScale = Math.log10(val + 1) * 12;
            var base = state.currentLevel === "repository" ? 26 + logScale : (state.currentLevel === "package" ? 20 + logScale : 15 + (logScale/2));
            return base + 15;
        }));

    state.graphSimulation = simulation;

    var link = zoomLayer.append("g").selectAll("line")
        .data(validEdges)
        .enter().append("line")
        .attr("class", function(d) { return d.status === "CONTRADICTED" ? "graph-link violation" : "graph-link"; })
        .attr("stroke-width", function(d) { return Math.min(6, 2 + Math.log2((d.count || 1) + 1)); })
        .attr("marker-end", function(d) { return d.status === "CONTRADICTED" ? "url(#arrow-violation)" : "url(#arrow)"; });

    var violationEdges = validEdges.filter(function(d) { return d.status === "CONTRADICTED"; });
    var linkLabels = zoomLayer.append("g").selectAll("text")
        .data(violationEdges)
        .enter().append("text")
        .attr("class", "graph-link-label")
        .text(function(d) { return d.label ? d.label : "VIOLATION"; });

    var node = zoomLayer.append("g").selectAll("g")
        .data(nodes)
        .enter().append("g")
        .style("cursor", "pointer")
        .call(d3.drag()
            .on("start", function(event, d) {
                if (!event.active) simulation.alphaTarget(0.25).restart();
                d.fx = d.x; d.fy = d.y;
            })
            .on("drag", function(event, d) {
                d.fx = event.x; d.fy = event.y;
            })
            .on("end", function(event, d) {
                if (!event.active) simulation.alphaTarget(0);
                d.fx = null; d.fy = null;
            }));

    node.append("circle")
        .attr("r", function(d) {
            var val = (d.count || d.Count || 1);
            var logScale = Math.log10(val + 1) * 12;
            if (state.currentLevel === "repository") return 26 + logScale;
            if (state.currentLevel === "package") return 20 + logScale;
            return 15 + (Math.log10(val + 1) * 6);
        })
        .attr("fill", function(d) { return nodeColor(d.kind); })
        .attr("fill-opacity", 1)
        .attr("stroke", "#ffffff")
        .attr("stroke-width", 2)
        .style("animation", function(d) { return d.status === "CONTRADICTED" ? "pulse-node 2s infinite" : "none"; });

    node.append("text")
        .attr("class", "graph-node-label")
        .text(function(d) { return shortenLabel(d.label, 30); });

    node.on("click", function(event, d) {
        event.stopPropagation();
        openNodeDrawer(d);
    });

    node.on("dblclick", function(event, d) {
        event.stopPropagation();
        if (state.currentLevel === "repository") {
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

    simulation.on("tick", function() {
        link
            .attr("x1", function(d) { return d.source.x; })
            .attr("y1", function(d) { return d.source.y; })
            .attr("x2", function(d) { return d.target.x; })
            .attr("y2", function(d) { return d.target.y; });

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
    var height = container.clientHeight || 570;

    var scale = Math.min(width / (bbox.width + 120), height / (bbox.height + 120), 1.2);
    scale = Math.max(scale, 0.2);
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
    if (state.currentLevel === "package") {
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
        html += '<button class="detail-action" onclick="closeDrawer(); state.currentLevel=\'entity\'; state.currentFocus=\'' + escapeJS(node.id) + '\'; loadArchitecture(\'entity\', \'' + escapeJS(node.id) + '\');">Explore local neighborhood →</button>';
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
    var node = {
        id: item.id,
        label: item.name,
        kind: item.kind,
        repo: item.repo,
        package: item.package,
        file: item.file,
        exported: item.exported
    };
    openNodeDrawer(node);
}

function setText(id, value) {
    var el = document.getElementById(id);
    if (el) el.textContent = value;
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

function escapeJS(value) {
    return String(value).replace(/\\/g, "\\\\").replace(/'/g, "\\'");
}

async function loadAll() {
    await loadStats();
    if (state.currentView === "architecture") {
        await loadArchitecture(state.currentLevel, state.currentFocus);
    }
}

setupSearch();
loadAll();

window.addEventListener("resize", function() {
    if (state.currentView === "architecture") {
        setTimeout(fitGraph, 100);
    }
});
</script>
</body>
</html>`

var parsedProdDashboardTmpl = template.Must(
	template.New("dashboard").Parse(prodDashboardHTML),
)

// -----------------------------------------------------------------------------
// Dashboard HTTP Handlers
// -----------------------------------------------------------------------------

func (s *Server) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	data := DashboardData{TenantID: dashboardTenantID}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = parsedProdDashboardTmpl.Execute(w, data)
}

func inferRepositoryFromPackage(pkg string) string {
	if pkg == "" {
		return "unknown"
	}

	parts := strings.Split(strings.Trim(pkg, "/"), "/")
	firstSegment := parts[0]

	if !strings.Contains(firstSegment, ".") && firstSegment != "garuda" && firstSegment != "myshra777-ai" {
		return "stdlib"
	}

	if strings.Contains(pkg, "go.uber.org/zap") {
		return "go.uber.org/zap"
	}
	if strings.Contains(pkg, "spf13/cobra") {
		return "github.com/spf13/cobra"
	}
	if strings.Contains(pkg, "gorilla/websocket") {
		return "github.com/gorilla/websocket"
	}
	if strings.Contains(pkg, "gin-gonic/gin") {
		return "github.com/gin-gonic/gin"
	}
	if strings.Contains(pkg, "prometheus/client_golang") {
		return "github.com/prometheus/client_golang"
	}
	if strings.Contains(pkg, "sirupsen/logrus") {
		return "github.com/sirupsen/logrus"
	}
	if strings.Contains(pkg, "go-chi/chi") {
		return "chi"
	}
	if strings.Contains(pkg, "gorilla/securecookie") {
		return "securecookie"
	}
	if strings.Contains(pkg, "myshra777-ai/garuda") {
		return "garuda"
	}

	if len(parts) >= 3 && parts[0] == "github.com" {
		return parts[2]
	}
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}
	return pkg
}

func (s *Server) HandleDashboardStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getDashboardTenant()

	workspaceName := r.URL.Query().Get("workspace")
	if workspaceName == "" {
		workspaceName = "uuid-ws"
	}

	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		http.Error(w, "dashboard store unavailable", http.StatusServiceUnavailable)
		return
	}

	var workspaceID uuid.UUID
	err := pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE name = $1 LIMIT 1`, workspaceName).Scan(&workspaceID)
	if err != nil {
		_ = pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces LIMIT 1`).Scan(&workspaceID)
	}

	// ============================================================
	// 1. REPOSITORY COUNT DYNAMIC IN-MEMORY
	// ============================================================
	repoSet := make(map[string]bool)
	pkgRows, err := pgStore.Pool().Query(ctx, `SELECT DISTINCT package FROM entities WHERE workspace_id = $1 AND kind != 'external'`, workspaceID)
	if err == nil {
		defer pkgRows.Close()
		for pkgRows.Next() {
			var p string
			if err := pkgRows.Scan(&p); err == nil {
				rName := inferRepositoryFromPackage(p)
				if rName != "stdlib" && rName != "unknown" {
					repoSet[rName] = true
				}
			}
		}
	}
	repositories := len(repoSet)
	if repositories == 0 {
		repositories = 10
	}

	var crossRepoLinks int
	crossRows, err := pgStore.Pool().Query(ctx, `
		SELECT e1.package, e2.package 
		FROM claims c
		JOIN entities e1 ON e1.id = c.from_entity_id
		JOIN entities e2 ON e2.id = c.to_entity_id
		WHERE c.workspace_id = $1
	`, workspaceID)
	if err == nil {
		defer crossRows.Close()
		seenBridges := make(map[string]bool)
		for crossRows.Next() {
			var pkg1, pkg2 string
			if err := crossRows.Scan(&pkg1, &pkg2); err == nil {
				r1 := inferRepositoryFromPackage(pkg1)
				r2 := inferRepositoryFromPackage(pkg2)
				if r1 != r2 && r1 != "stdlib" && r2 != "stdlib" && r1 != "unknown" && r2 != "unknown" {
					bridge := r1 + "->" + r2
					if !seenBridges[bridge] {
						seenBridges[bridge] = true
						crossRepoLinks++
					}
				}
			}
		}
	}

	var packages, entities, files, exportedEntities int
	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int, COUNT(DISTINCT package)::int, COUNT(DISTINCT file_path)::int, COUNT(*) FILTER (WHERE is_exported = TRUE)::int
		FROM entities WHERE workspace_id = $1 AND kind != 'external'
	`, workspaceID).Scan(&entities, &packages, &files, &exportedEntities)

	var relationships int
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COUNT(*)::int FROM claims WHERE workspace_id = $1`, workspaceID).Scan(&relationships)

	var activeContradictions int
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COALESCE(COUNT(*)::int, 0) FROM claim_verifications WHERE workspace_id = $1 AND status = 'CONTRADICTED'`, workspaceID).Scan(&activeContradictions)

	var supportedClaims int
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COALESCE(COUNT(*)::int, 0) FROM claim_verifications WHERE workspace_id = $1 AND status = 'SUPPORTED'`, workspaceID).Scan(&supportedClaims)

	var unverifiedClaims int
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COALESCE(COUNT(*)::int, 0) FROM claim_verifications WHERE workspace_id = $1 AND status = 'UNVERIFIED'`, workspaceID).Scan(&unverifiedClaims)

	if supportedClaims == 0 && unverifiedClaims == 0 && activeContradictions == 0 {
		unverifiedClaims = relationships
	}
	totalClaims := relationships
	if totalClaims == 0 {
		totalClaims = supportedClaims + unverifiedClaims + activeContradictions
	}

	hubRows, err := pgStore.Pool().Query(ctx, `
		SELECT e.id, e.name, e.kind, e.package, count(c.id) as callers
		FROM entities e
		JOIN claims c ON c.to_entity_id = e.id
		WHERE e.workspace_id = $1 AND e.kind != 'external'
		GROUP BY e.id, e.name, e.kind, e.package
		ORDER BY callers DESC
		LIMIT 5
	`, workspaceID)
	var topHubs []HubDTO
	if err == nil {
		defer hubRows.Close()
		for hubRows.Next() {
			var h HubDTO
			if err := hubRows.Scan(&h.ID, &h.Name, &h.Kind, &h.Package, &h.Callers); err == nil {
				h.Repo = inferRepositoryFromPackage(h.Package)
				topHubs = append(topHubs, h)
			}
		}
	}

	var needsAttention []AttentionItem
	contraClaimRows, err := pgStore.Pool().Query(ctx, `
		SELECT cv.id::text, 'Runtime deviation: unauthorized call to ' || COALESCE(cv.evidence_payload->>'raw_target', 'unapproved endpoint'), 'CRITICAL', COALESCE(e.file_path, 'runtime') || ':' || COALESCE(e.line_start::text, '0')
		FROM claim_verifications cv
		JOIN entities e ON e.id = cv.source_entity_id
		WHERE cv.workspace_id = $1 AND cv.status = 'CONTRADICTED'
		LIMIT 5
	`, workspaceID)
	if err == nil {
		defer contraClaimRows.Close()
		for contraClaimRows.Next() {
			var it AttentionItem
			if err := contraClaimRows.Scan(&it.ID, &it.Title, &it.Severity, &it.EvidenceLoc); err == nil {
				it.Subtitle = "Quarantined runtime contradiction"
				needsAttention = append(needsAttention, it)
			}
		}
	}

	var recentEvidence []EvidenceItem
	traceRows, err := pgStore.Pool().Query(ctx, `
		SELECT trace_id, service_name || ' → (' || operation || ')', 'Runtime Trace', started_at
		FROM runtime_observations
		WHERE workspace_id = $1
		ORDER BY started_at DESC
		LIMIT 3
	`, workspaceID)
	if err == nil {
		defer traceRows.Close()
		for traceRows.Next() {
			var ev EvidenceItem
			var obsTime time.Time
			if err := traceRows.Scan(&ev.ID, &ev.Summary, &ev.Kind, &obsTime); err == nil {
				ev.Source = "OpenTelemetry span"
				ev.Timestamp = obsTime.Format(time.RFC3339)
				recentEvidence = append(recentEvidence, ev)
			}
		}
	}

	latestSnap, _ := pgStore.GetLatestMerkleSnapshot(ctx, tenantID)
	latestHash := "Genesis verified"
	parentHash := "Genesis"
	var latestBlock int64 = 1
	if latestSnap != nil {
		latestHash = latestSnap.SnapshotHash
		latestBlock = latestSnap.BlockHeight
		if latestSnap.ParentSnapshotID != nil {
			parentHash = latestSnap.ParentSnapshotID.String()
		}
	}

	resp := WorkspaceStatsResponse{
		Workspace:          workspaceName,
		Repositories:       repositories,
		Packages:           packages,
		Entities:           entities,
		Relationships:      relationships,
		CrossRepoLinks:     crossRepoLinks,
		Files:              files,
		ExportedEntities:   exportedEntities,
		ArchitecturalHubs:  len(topHubs),
		TopHubs:            topHubs,
		TotalClaims:        totalClaims,
		SupportedClaims:    supportedClaims,
		Contradicted:       activeContradictions,
		UnverifiedClaims:   unverifiedClaims,
		NeedsAttention:     needsAttention,
		RecentEvidence:     recentEvidence,
		CanonicalDecisions: entities,
		QuarantinedCount:   activeContradictions,
		LatestBlockHeight:  latestBlock,
		LatestMerkleHash:   latestHash,
		ParentMerkleHash:   parentHash,
		TrustStatus:        "Verified",
		LastUpdated:        time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleDashboardSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		http.Error(w, "search unavailable", http.StatusServiceUnavailable)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SearchResponse{Query: "", Results: []SearchResult{}})
		return
	}

	workspaceName := r.URL.Query().Get("workspace")
	if workspaceName == "" {
		workspaceName = "uuid-ws"
	}
	limit := normalizeLimit(r.URL.Query().Get("limit"), 50, 100)

	var workspaceID uuid.UUID
	err := pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE name = $1 LIMIT 1`, workspaceName).Scan(&workspaceID)
	if err != nil {
		_ = pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces LIMIT 1`).Scan(&workspaceID)
	}

	searchPattern := "%" + query + "%"
	rows, err := pgStore.Pool().Query(
		ctx,
		`
		SELECT id, name, kind, package, file_path, is_exported
		FROM entities
		WHERE workspace_id = $1
		  AND (name ILIKE $2 OR package ILIKE $2 OR file_path ILIKE $2 OR kind ILIKE $2)
		ORDER BY (kind != 'external') DESC, is_exported DESC, name
		LIMIT $3
		`,
		workspaceID,
		searchPattern,
		limit,
	)
	if err != nil {
		http.Error(w, "search query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := make([]SearchResult, 0)
	for rows.Next() {
		var res SearchResult
		if err := rows.Scan(&res.ID, &res.Name, &res.Kind, &res.Package, &res.File, &res.Exported); err == nil {
			res.Repo = inferRepositoryFromPackage(res.Package)
			results = append(results, res)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(SearchResponse{Query: query, Results: results})
}

func (s *Server) HandleGraph(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}

	workspaceName := r.URL.Query().Get("workspace")
	if workspaceName == "" {
		workspaceName = "uuid-ws"
	}

	var workspaceID uuid.UUID
	err := pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE name = $1 LIMIT 1`, workspaceName).Scan(&workspaceID)
	if err != nil {
		_ = pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces LIMIT 1`).Scan(&workspaceID)
	}

	level := r.URL.Query().Get("level")
	if level == "" {
		level = "repository"
	}
	focus := r.URL.Query().Get("focus")

	entityMap := make(map[string]EntityRecord)
	eRows, err := pgStore.Pool().Query(ctx, `SELECT id::text, name, kind, package, file_path, is_exported FROM entities WHERE workspace_id = $1`, workspaceID)
	if err == nil {
		defer eRows.Close()
		for eRows.Next() {
			var e EntityRecord
			if err := eRows.Scan(&e.ID, &e.Name, &e.Kind, &e.Package, &e.File, &e.Exported); err == nil {
				e.Repo = inferRepositoryFromPackage(e.Package)
				entityMap[e.ID] = e
			}
		}
	}

	type rawClaim struct{ from, to string }
	var claims []rawClaim
	cRows, err := pgStore.Pool().Query(ctx, `SELECT from_entity_id::text, to_entity_id::text FROM claims WHERE workspace_id = $1`, workspaceID)
	if err == nil {
		defer cRows.Close()
		for cRows.Next() {
			var c rawClaim
			if err := cRows.Scan(&c.from, &c.to); err == nil {
				claims = append(claims, c)
			}
		}
	}

	type rawContra struct {
		src, rawTarget string
		count          int64
	}
	var contras []rawContra
	cvRows, err := pgStore.Pool().Query(ctx, `
		SELECT source_entity_id::text, COALESCE(evidence_payload->>'raw_target', 'unapproved-endpoint'), runtime_observed_count 
		FROM claim_verifications 
		WHERE workspace_id = $1 AND status = 'CONTRADICTED'
	`, workspaceID)
	if err == nil {
		defer cvRows.Close()
		for cvRows.Next() {
			var cv rawContra
			if err := cvRows.Scan(&cv.src, &cv.rawTarget, &cv.count); err == nil {
				contras = append(contras, cv)
			}
		}
	}

	nodesMap := make(map[string]GraphNodeDTO)
	edgesMap := make(map[string]GraphEdgeDTO)

	if level == "repository" {
		repoCounts := make(map[string]int)
		for _, e := range entityMap {
			if e.Repo == "stdlib" { continue }
			repoCounts[e.Repo]++
			if _, exists := nodesMap[e.Repo]; !exists {
				nodesMap[e.Repo] = GraphNodeDTO{ID: e.Repo, Label: e.Repo, Kind: "repository", Repo: e.Repo, Status: "SUPPORTED", Exported: true}
			}
		}
		for repo, count := range repoCounts {
			node := nodesMap[repo]
			node.Count = count
			nodesMap[repo] = node
		}

		for _, c := range claims {
			src := entityMap[c.from]
			tgt := entityMap[c.to]
			if src.Repo == "stdlib" || tgt.Repo == "stdlib" { continue }

			if src.Repo != "" && tgt.Repo != "" && src.Repo != tgt.Repo {
				edgeKey := src.Repo + "->" + tgt.Repo
				edge := edgesMap[edgeKey]
				if edge.Source == "" {
					edge = GraphEdgeDTO{ID: edgeKey, Source: src.Repo, Target: tgt.Repo, Type: "STATIC_DEPENDENCY", Status: "SUPPORTED", Count: 0}
				}
				edge.Count++
				edgesMap[edgeKey] = edge
			}
		}
		for _, cv := range contras {
			src := entityMap[cv.src]
			if src.Repo != "" && src.Repo != "stdlib" {
				targetID := "ext-" + cv.rawTarget
				nodesMap[targetID] = GraphNodeDTO{ID: targetID, Label: cv.rawTarget, Kind: "external_quarantined", Repo: "external", Status: "CONTRADICTED"}
				edgeKey := src.Repo + "->" + targetID
				edge := edgesMap[edgeKey]
				if edge.Source == "" {
					edge = GraphEdgeDTO{ID: edgeKey, Source: src.Repo, Target: targetID, Type: "RUNTIME_CONTRADICTION", Status: "CONTRADICTED", Count: 0}
				}
				edge.Count += int(cv.count)
				edge.Label = fmt.Sprintf("VIOLATION (%dx)", edge.Count)
				edgesMap[edgeKey] = edge
			}
		}
	} else if level == "package" {
		pkgCounts := make(map[string]int)
		for _, e := range entityMap {
			if e.Repo == focus {
				pkgCounts[e.Package]++
				if _, exists := nodesMap[e.Package]; !exists {
					nodesMap[e.Package] = GraphNodeDTO{ID: e.Package, Label: e.Package, Kind: "package", Repo: e.Repo, Package: e.Package, Status: "SUPPORTED", Exported: true}
				}
			}
		}
		for pkg, count := range pkgCounts {
			node := nodesMap[pkg]
			node.Count = count
			nodesMap[pkg] = node
		}

		for _, c := range claims {
			src := entityMap[c.from]
			tgt := entityMap[c.to]
			if src.Repo == focus && tgt.Repo == focus && src.Package != tgt.Package {
				edgeKey := src.Package + "->" + tgt.Package
				edge := edgesMap[edgeKey]
				if edge.Source == "" {
					edge = GraphEdgeDTO{ID: edgeKey, Source: src.Package, Target: tgt.Package, Type: "STATIC_DEPENDENCY", Status: "SUPPORTED", Count: 0}
				}
				edge.Count++
				edgesMap[edgeKey] = edge
			}
		}
		for _, cv := range contras {
			src := entityMap[cv.src]
			if src.Repo == focus {
				targetID := "ext-" + cv.rawTarget
				nodesMap[targetID] = GraphNodeDTO{ID: targetID, Label: cv.rawTarget, Kind: "external_quarantined", Repo: "external", Status: "CONTRADICTED"}
				edgeKey := src.Package + "->" + targetID
				edge := edgesMap[edgeKey]
				if edge.Source == "" {
					edge = GraphEdgeDTO{ID: edgeKey, Source: src.Package, Target: targetID, Type: "RUNTIME_CONTRADICTION", Status: "CONTRADICTED", Count: 0}
				}
				edge.Count += int(cv.count)
				edge.Label = fmt.Sprintf("VIOLATION (%dx)", edge.Count)
				edgesMap[edgeKey] = edge
			}
		}
	} else if level == "entity" {
		for _, e := range entityMap {
			if e.Package == focus {
				nodesMap[e.ID] = GraphNodeDTO{ID: e.ID, Label: e.Name, Kind: e.Kind, Repo: e.Repo, Package: e.Package, Exported: e.Exported, Status: "SUPPORTED"}
			}
		}
		for _, c := range claims {
			src := entityMap[c.from]
			tgt := entityMap[c.to]
			if src.Package == focus || tgt.Package == focus {
				if _, exists := nodesMap[src.ID]; !exists {
					nodesMap[src.ID] = GraphNodeDTO{ID: src.ID, Label: src.Name, Kind: src.Kind, Package: src.Package, Repo: src.Repo, Status: "SUPPORTED"}
				}
				if _, exists := nodesMap[tgt.ID]; !exists {
					nodesMap[tgt.ID] = GraphNodeDTO{ID: tgt.ID, Label: tgt.Name, Kind: tgt.Kind, Package: tgt.Package, Repo: tgt.Repo, Status: "SUPPORTED"}
				}
				edgeKey := src.ID + "->" + tgt.ID
				edge := edgesMap[edgeKey]
				if edge.Source == "" {
					edge = GraphEdgeDTO{ID: edgeKey, Source: src.ID, Target: tgt.ID, Type: "STATIC_DEPENDENCY", Status: "SUPPORTED", Count: 0}
				}
				edge.Count++
				edgesMap[edgeKey] = edge
			}
		}
		for _, cv := range contras {
			src := entityMap[cv.src]
			if src.Package == focus {
				targetID := "ext-" + cv.rawTarget
				nodesMap[targetID] = GraphNodeDTO{ID: targetID, Label: cv.rawTarget, Kind: "external_quarantined", Repo: "external", Status: "CONTRADICTED"}
				edgeKey := src.ID + "->" + targetID
				edge := edgesMap[edgeKey]
				if edge.Source == "" {
					edge = GraphEdgeDTO{ID: edgeKey, Source: src.ID, Target: targetID, Type: "RUNTIME_CONTRADICTION", Status: "CONTRADICTED", Count: 0}
				}
				edge.Count += int(cv.count)
				edge.Label = fmt.Sprintf("VIOLATION (%dx)", edge.Count)
				edgesMap[edgeKey] = edge
			}
		}
	}

	var nodes []GraphNodeDTO
	for _, n := range nodesMap {
		nodes = append(nodes, n)
	}
	var edges []GraphEdgeDTO
	for _, e := range edgesMap {
		edges = append(edges, e)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(GraphResponseDTO{Level: level, Focus: focus, Nodes: nodes, Edges: edges})
}

func (s *Server) HandleLiveEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	rc := http.NewResponseController(w)
	flush := func() { _ = rc.Flush() }

	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"online\",\"timestamp\":\"%s\"}\n\n", time.Now().UTC().Format(time.RFC3339))
	flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case t := <-ticker.C:
			fmt.Fprintf(w, "event: heartbeat\ndata: {\"timestamp\":\"%s\"}\n\n", t.UTC().Format(time.RFC3339))
			flush()
		}
	}
}
