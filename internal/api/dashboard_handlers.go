// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
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
	TenantID      string
	WorkspaceName string
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

// NEW: LanguageDTO — GitHub-style language breakdown per repo/workspace
type LanguageDTO struct {
	Name       string  `json:"name"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
	Color      string  `json:"color"`
}

// NEW: RepoStatDTO — per-repository breakdown
type RepoStatDTO struct {
	Name           string        `json:"name"`
	Entities       int           `json:"entities"`
	Relationships  int           `json:"relationships"`
	Files          int           `json:"files"`
	Languages      []LanguageDTO `json:"languages"`
	AnalysisStatus string        `json:"analysis_status"`
	CurrentCommit  string        `json:"current_commit"`
	LastAnalyzed   string        `json:"last_analyzed"`
}

// NEW: DriftDTO — document ↔ code drift summary
type DriftDTO struct {
	TotalDocumentClaims int `json:"total_document_claims"`
	SupportedClaims     int `json:"supported_claims"`
	UnverifiedClaims    int `json:"unverified_claims"`
	ContradictedClaims  int `json:"contradicted_claims"`
	UndocumentedCode    int `json:"undocumented_code"`
	UnimplementedDocs   int `json:"unimplemented_docs"`
	DocToCodeDriftCount int `json:"doc_to_code_drift_count"`
	CodeToDocDriftCount int `json:"code_to_doc_drift_count"`
}

type WorkspaceStatsResponse struct {
	Workspace             string          `json:"workspace"`
	Repositories          int             `json:"repositories"`
	RepositoriesList      []string        `json:"repositories_list"`
	RepoStats             []RepoStatDTO   `json:"repo_stats"` // NEW
	Packages              int             `json:"packages"`
	Entities              int             `json:"entities"`
	Relationships         int             `json:"relationships"`
	CrossRepoLinks        int             `json:"cross_repo_links"`
	Files                 int             `json:"files"`
	ExportedEntities      int             `json:"exported_entities"`
	ArchitecturalHubs     int             `json:"architectural_hubs"`
	TopHubs               []HubDTO        `json:"top_hubs"`
	TotalClaims           int             `json:"total_claims"`
	SupportedClaims       int             `json:"supported_claims"`
	Contradicted          int             `json:"contradicted"`
	UnverifiedClaims      int             `json:"unverified_claims"`
	NeedsAttention        []AttentionItem `json:"needs_attention"`
	RecentEvidence        []EvidenceItem  `json:"recent_evidence"`
	CanonicalDecisions    int             `json:"canonical_decisions"`
	QuarantinedCount      int             `json:"quarantined_count"`
	LatestBlockHeight     int64           `json:"latest_block_height"`
	LatestMerkleHash      string          `json:"latest_merkle_hash"`
	ParentMerkleHash      string          `json:"parent_merkle_hash"`
	TrustStatus           string          `json:"trust_status"`
	LastUpdated           string          `json:"last_updated"`
	PendingDecisions      int             `json:"pending_decisions"`
	IdempotencySafeguards int             `json:"idempotency_safeguards"`
	TokensSaved           int64           `json:"tokens_saved"`
	EstimatedCostSavedUSD float64         `json:"estimated_cost_saved_usd"`
	ColdStartLatencyMs    float64         `json:"cold_start_latency_ms"`
	ColdStartLatencyKnown bool            `json:"cold_start_latency_known"`
	ActiveAgentsCount     int             `json:"active_agents_count"`
	LanguagesBreakdown    []LanguageDTO   `json:"languages_breakdown"` // NEW (was map, now slice with percentages)
	Drift                 DriftDTO        `json:"drift"`               // NEW
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

func applySecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://d3js.org; style-src 'self' 'unsafe-inline'; img-src 'self' data:;")
}

// NEW: GitHub-style language color palette
func languageColor(lang string) string {
	colors := map[string]string{
		"Go":         "#00ADD8",
		"Python":     "#3572A5",
		"TypeScript": "#2b7489",
		"JavaScript": "#f1e05a",
		"Rust":       "#dea584",
		"Java":       "#b07219",
		"C":          "#555555",
		"C++":        "#f34b7d",
		"C#":         "#178600",
		"Ruby":       "#701516",
		"PHP":        "#4F5D95",
		"Swift":      "#ffac45",
		"Kotlin":     "#A97BFF",
		"Shell":      "#89e051",
		"HTML":       "#e34c26",
		"CSS":        "#563d7c",
		"Other":      "#8b949e",
	}
	if c, ok := colors[lang]; ok {
		return c
	}
	return "#8b949e"
}

// NEW: Infer language from file extension
func inferLanguageFromPath(path string) string {
	ext := ""
	if idx := strings.LastIndex(path, "."); idx >= 0 {
		ext = strings.ToLower(path[idx:])
	}
	switch ext {
	case ".go":
		return "Go"
	case ".py", ".pyi":
		return "Python"
	case ".ts", ".tsx":
		return "TypeScript"
	case ".js", ".jsx", ".mjs":
		return "JavaScript"
	case ".rs":
		return "Rust"
	case ".java":
		return "Java"
	case ".c", ".h":
		return "C"
	case ".cpp", ".cc", ".hpp", ".cxx":
		return "C++"
	case ".cs":
		return "C#"
	case ".rb":
		return "Ruby"
	case ".php":
		return "PHP"
	case ".swift":
		return "Swift"
	case ".kt":
		return "Kotlin"
	case ".sh", ".bash":
		return "Shell"
	case ".html", ".htm":
		return "HTML"
	case ".css", ".scss":
		return "CSS"
	}
	return "Other"
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
    --bg: #060810;
    --surface: #0e1424;
    --surface-2: #162035;
    --border: #1f2b45;
    --border-strong: #2f4166;
    --text: #f8fafc;
    --text-2: #94a3b8;
    --muted: #64748b;
    --brand: #38bdf8;
    --brand-dark: #0284c7;
    --brand-soft: rgba(56, 189, 248, 0.12);
    --green: #34d399;
    --green-soft: rgba(52, 211, 153, 0.12);
    --amber: #fbbf24;
    --amber-soft: rgba(251, 191, 36, 0.12);
    --red: #f43f5e;
    --red-soft: rgba(244, 63, 94, 0.12);
    --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.6);
    --shadow-md: 0 10px 30px -5px rgba(0, 0, 0, 0.85);
    --radius: 12px;
}
* { box-sizing: border-box; }
html, body { margin: 0; padding: 0; min-height: 100%; }
body { background: var(--bg); color: var(--text); font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Inter, sans-serif; font-size: 13px; }
button, input { font: inherit; }
button { cursor: pointer; }
.app { min-height: 100vh; display: flex; }

.sidebar { width: 245px; min-width: 245px; background: #04060c; color: #94a3b8; display: flex; flex-direction: column; border-right: 1px solid var(--border); }
.brand { height: 64px; padding: 0 20px; display: flex; align-items: center; gap: 12px; border-bottom: 1px solid var(--border); }
.brand-mark { width: 34px; height: 34px; border-radius: 9px; background: linear-gradient(135deg, #0284c7, #38bdf8); display: grid; place-items: center; font-size: 18px; box-shadow: 0 0 18px rgba(56,189,248,0.45); }
.brand-name { font-weight: 800; letter-spacing: 0.5px; color: white; font-size: 15px; }
.brand-subtitle { color: var(--muted); font-size: 10px; margin-top: 2px; }
.sidebar-content { padding: 20px 14px; flex: 1; }
.nav-section { margin-bottom: 24px; }
.nav-title { color: #475569; font-size: 10px; font-weight: 750; text-transform: uppercase; letter-spacing: 0.1em; padding: 0 10px 8px; }
.nav-item { width: 100%; border: 0; background: transparent; color: #94a3b8; text-align: left; padding: 10px 12px; border-radius: 8px; display: flex; align-items: center; gap: 12px; margin-bottom: 3px; font-size: 13px; font-weight: 500; transition: 0.15s; }
.nav-item:hover { background: rgba(255,255,255,0.04); color: white; }
.nav-item.active { background: var(--brand-soft); color: var(--brand); box-shadow: inset 3px 0 0 var(--brand); font-weight: 700; }
.nav-icon { width: 18px; text-align: center; color: var(--muted); }
.nav-item.active .nav-icon { color: var(--brand); }
.sidebar-footer { padding: 15px; border-top: 1px solid var(--border); }
.workspace-mini { padding: 12px; background: #080c16; border: 1px solid var(--border); border-radius: 9px; cursor: pointer; transition: 0.2s; }
.workspace-mini:hover { border-color: var(--brand); box-shadow: 0 0 14px rgba(56,189,248,0.25); }
.workspace-mini-name { color: white; font-weight: 700; margin-bottom: 4px; }
.workspace-mini-meta { color: var(--muted); font-size: 11px; }
.trust-mini { margin-top: 6px; color: var(--green); font-size: 11px; font-weight: 600; }

.main { flex: 1; min-width: 0; display: flex; flex-direction: column; background: var(--bg); }
.topbar { height: 64px; background: rgba(10, 14, 26, 0.85); backdrop-filter: blur(14px); border-bottom: 1px solid var(--border); display: flex; align-items: center; gap: 18px; padding: 0 28px; position: sticky; top: 0; z-index: 50; }
.global-search { flex: 1; max-width: 760px; position: relative; }
.search-input { width: 100%; height: 40px; border: 1px solid var(--border); background: #050811; border-radius: 9px; padding: 0 42px 0 38px; color: white; outline: none; transition: 0.15s; }
.search-input:focus { border-color: var(--brand); box-shadow: 0 0 0 3px rgba(56, 189, 248, 0.25); }
.search-icon { position: absolute; left: 13px; top: 11px; color: var(--muted); font-size: 16px; }
.search-shortcut { position: absolute; right: 10px; top: 9px; border: 1px solid var(--border); background: var(--surface-2); color: var(--muted); font-size: 10px; border-radius: 5px; padding: 2px 6px; }
.topbar-right { margin-left: auto; display: flex; align-items: center; gap: 12px; }
.live-pill { display: flex; align-items: center; gap: 7px; color: var(--green); font-size: 11px; font-weight: 650; background: var(--green-soft); padding: 5px 11px; border-radius: 20px; border: 1px solid rgba(52,211,153,0.3); }
.live-dot { width: 7px; height: 7px; border-radius: 50%; background: var(--green); box-shadow: 0 0 8px var(--green); }
.refresh-button { border: 1px solid var(--border); background: var(--surface); color: var(--text); height: 36px; padding: 0 14px; border-radius: 8px; font-weight: 600; transition: 0.15s; }
.refresh-button:hover { background: var(--surface-2); border-color: var(--border-strong); }

.content { width: 100%; max-width: 1580px; margin: 0 auto; padding: 28px 32px 60px; }
.breadcrumbs { display: flex; align-items: center; gap: 8px; color: var(--muted); font-size: 12px; margin-bottom: 18px; }
.breadcrumb-button { border: 0; padding: 0; background: transparent; color: var(--text-2); }
.breadcrumb-button:hover { color: var(--brand); }
.breadcrumb-current { color: white; font-weight: 650; }
.hero { display: flex; justify-content: space-between; align-items: flex-start; gap: 20px; margin-bottom: 26px; }
.hero-title { font-size: 26px; line-height: 1.25; margin: 0; letter-spacing: -0.02em; color: white; font-weight: 800; }
.hero-subtitle { color: var(--text-2); margin-top: 6px; font-size: 13px; }
.trust-badge { display: inline-flex; align-items: center; gap: 8px; padding: 8px 12px; border: 1px solid rgba(52,211,153,0.3); background: var(--green-soft); color: var(--green); border-radius: 8px; font-size: 12px; font-weight: 700; white-space: nowrap; box-shadow: 0 0 15px rgba(52,211,153,0.15); }

.kpi-grid { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 14px; margin-bottom: 20px; }
.kpi-card { background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius); padding: 18px; box-shadow: var(--shadow-sm); transition: 0.2s; }
.kpi-card:hover { border-color: var(--border-strong); box-shadow: var(--shadow-md); }
.kpi-label { color: var(--muted); font-size: 11px; font-weight: 750; text-transform: uppercase; letter-spacing: 0.05em; }
.kpi-value { font-size: 26px; font-weight: 800; margin-top: 8px; letter-spacing: -0.03em; color: white; }
.kpi-foot { color: var(--muted); font-size: 11px; margin-top: 6px; }

/* NEW: GitHub-style language bar */
.lang-bar-container { display: flex; height: 10px; border-radius: 5px; overflow: hidden; background: #1a2035; margin-top: 4px; }
.lang-bar-segment { height: 100%; transition: 0.3s; }
.lang-legend { display: flex; flex-wrap: wrap; gap: 16px; margin-top: 14px; }
.lang-legend-item { display: flex; align-items: center; gap: 7px; font-size: 12px; color: var(--text-2); }
.lang-legend-dot { width: 10px; height: 10px; border-radius: 50%; }
.lang-legend-name { color: white; font-weight: 700; }
.lang-legend-pct { color: var(--muted); font-weight: 600; }

.trust-strip { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 14px; margin-bottom: 20px; }
.trust-card { background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius); padding: 18px; box-shadow: var(--shadow-sm); border-left: 4px solid var(--muted); }
.trust-card.supported { border-left-color: var(--green); }
.trust-card.unverified { border-left-color: var(--amber); }
.trust-card.contradicted { border-left-color: var(--red); }
.trust-card-title { font-size: 11px; font-weight: 750; text-transform: uppercase; letter-spacing: 0.05em; }
.trust-card.supported .trust-card-title { color: var(--green); }
.trust-card.unverified .trust-card-title { color: var(--amber); }
.trust-card.contradicted .trust-card-title { color: var(--red); }
.trust-card-val { font-size: 25px; font-weight: 800; margin-top: 6px; color: white; }
.trust-card-desc { font-size: 11px; color: var(--muted); margin-top: 4px; }

.start-card { background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius); padding: 22px; box-shadow: var(--shadow-sm); margin-bottom: 20px; }
.start-title { font-size: 15px; font-weight: 750; margin-bottom: 6px; color: white; }
.start-description { color: var(--text-2); font-size: 12px; margin-bottom: 16px; }
.quick-search { height: 46px; width: 100%; border: 1px solid var(--border-strong); background: #060912; border-radius: 9px; padding: 0 16px; outline: none; color: white; }
.quick-search:focus { border-color: var(--brand); box-shadow: 0 0 0 3px rgba(56, 189, 248, 0.25); }
.quick-hints { display: flex; gap: 8px; flex-wrap: wrap; margin-top: 12px; }
.hint { border: 1px solid var(--border); background: var(--surface-2); color: var(--text-2); padding: 6px 10px; border-radius: 7px; font-size: 11px; font-weight: 600; }
.hint:hover { border-color: var(--brand); color: white; background: var(--brand-soft); }

.two-column { display: grid; grid-template-columns: minmax(0, 1.3fr) minmax(320px, 1fr); gap: 20px; margin-bottom: 20px; }
.panel { background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius); box-shadow: var(--shadow-sm); overflow: hidden; }
.panel-header { padding: 16px 20px; border-bottom: 1px solid var(--border); display: flex; align-items: center; justify-content: space-between; gap: 10px; }
.panel-title { font-size: 14px; font-weight: 750; color: white; }
.panel-subtitle { color: var(--muted); font-size: 11px; margin-top: 3px; }
.panel-body { padding: 18px; }

.explorer { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 14px; }
.explorer-card { min-height: 140px; border: 1px solid var(--border); background: var(--surface-2); border-radius: 10px; padding: 16px; transition: 0.2s; text-align: left; }
.explorer-card:hover { border-color: var(--brand); box-shadow: var(--shadow-md); transform: translateY(-2px); background: #1c2942; }
.explorer-icon { width: 34px; height: 34px; border-radius: 8px; display: grid; place-items: center; background: var(--brand-soft); color: var(--brand); margin-bottom: 12px; font-size: 16px; }
.explorer-name { font-weight: 750; font-size: 13px; color: white; }
.explorer-count { color: var(--brand); font-size: 20px; font-weight: 800; margin-top: 6px; }
.explorer-meta { color: var(--muted); font-size: 10px; margin-top: 4px; }

.list { display: flex; flex-direction: column; }
.list-row { border-bottom: 1px solid var(--border); padding: 13px 18px; display: flex; align-items: center; gap: 14px; transition: 0.15s; }
.list-row:last-child { border-bottom: 0; }
.list-row:hover { background: rgba(255,255,255,0.025); }
.row-icon { width: 32px; height: 32px; border-radius: 8px; display: grid; place-items: center; background: var(--surface-2); color: var(--text-2); flex-shrink: 0; }
.row-main { flex: 1; min-width: 0; }
.row-title { color: white; font-size: 13px; font-weight: 700; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.row-meta { color: var(--muted); font-size: 11px; margin-top: 3px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

/* NEW: Repo card with language bar */
.repo-card { border-bottom: 1px solid var(--border); padding: 18px; }
.repo-card:last-child { border-bottom: 0; }
.repo-card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.repo-card-name { font-weight: 750; color: white; font-size: 14px; }
.repo-card-meta { color: var(--muted); font-size: 11px; margin-top: 3px; }
.repo-card-stats { display: flex; gap: 18px; font-size: 11px; color: var(--text-2); margin-bottom: 10px; }
.repo-card-stat-val { color: white; font-weight: 700; }
.repo-card-actions { display: flex; gap: 8px; }

.badge-pill { font-size: 9px; font-weight: 800; padding: 3px 8px; border-radius: 6px; text-transform: uppercase; }
.badge-pill.critical { background: var(--red-soft); color: var(--red); border: 1px solid rgba(244,63,94,0.3); }
.badge-pill.warning { background: var(--amber-soft); color: var(--amber); border: 1px solid rgba(251,191,36,0.3); }
.badge-pill.info { background: var(--brand-soft); color: var(--brand); border: 1px solid rgba(56,189,248,0.3); }
.badge-pill.success { background: var(--green-soft); color: var(--green); border: 1px solid rgba(52,211,153,0.3); }

/* NEW: Drift grid */
.drift-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px; margin-bottom: 20px; }
.drift-card { background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius); padding: 18px; box-shadow: var(--shadow-sm); }
.drift-card-title { font-size: 11px; font-weight: 750; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 12px; color: var(--muted); }
.drift-row { display: flex; justify-content: space-between; padding: 7px 0; border-bottom: 1px solid var(--border); font-size: 12px; }
.drift-row:last-child { border-bottom: 0; }
.drift-row-label { color: var(--text-2); }
.drift-row-val { color: white; font-weight: 700; }

.graph-panel { margin-top: 20px; }
.graph-toolbar { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.graph-button { border: 1px solid var(--border); background: var(--surface); color: var(--text); height: 32px; padding: 0 12px; border-radius: 7px; font-size: 11px; font-weight: 600; }
.graph-button:hover { background: var(--surface-2); border-color: var(--border-strong); }
.graph-button.primary { background: var(--brand-dark); color: white; border-color: var(--brand); }
.graph-button.primary:hover { background: #0369a1; }

.graph-layout { display: flex; height: 750px; position: relative; border-radius: 0 0 var(--radius) var(--radius); overflow: hidden; background: #05070f; }
.graph-wrap { flex: 1; height: 100%; position: relative; background: radial-gradient(circle at 50% 50%, #0d1527 0%, #05070f 100%); }
.graph-wrap.fullscreen { position: fixed; inset: 0; z-index: 9999; height: 100vh; width: 100vw; }

.graph-side-panel { width: 310px; min-width: 310px; background: #080d1a; border-left: 1px solid var(--border); display: flex; flex-direction: column; overflow: hidden; }
.side-panel-header { padding: 14px 18px; border-bottom: 1px solid var(--border); font-weight: 750; font-size: 11.5px; color: #f8fafc; letter-spacing: 0.06em; display: flex; justify-content: space-between; align-items: center; }
.side-panel-list { overflow-y: auto; padding: 8px 10px; flex: 1; }
.community-item { display: flex; align-items: center; gap: 10px; padding: 7px 10px; border-radius: 7px; cursor: pointer; transition: 0.12s; }
.community-item:hover { background: rgba(255,255,255,0.05); }
.community-checkbox { accent-color: var(--brand); cursor: pointer; width: 14px; height: 14px; }
.community-dot { width: 11px; height: 11px; border-radius: 50%; flex-shrink: 0; box-shadow: 0 0 8px currentColor; }
.community-name { font-size: 11.5px; font-weight: 600; color: #cbd5e1; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.community-count { font-size: 10.5px; color: var(--muted); font-weight: 750; }

#graph { width: 100%; height: 100%; }
.graph-empty { position: absolute; inset: 0; display: grid; place-items: center; color: var(--muted); font-size: 13px; pointer-events: none; }
.graph-help { position: absolute; left: 18px; bottom: 18px; background: rgba(8, 13, 26, 0.88); backdrop-filter: blur(10px); border: 1px solid var(--border); border-radius: 8px; padding: 8px 14px; color: #94a3b8; font-size: 11px; box-shadow: var(--shadow-sm); pointer-events: none; }
.graph-controls { position: absolute; right: 18px; top: 18px; display: flex; flex-direction: column; gap: 6px; }
.graph-control { width: 34px; height: 34px; border: 1px solid var(--border); background: var(--surface); color: var(--text); border-radius: 8px; box-shadow: var(--shadow-sm); font-size: 16px; display: grid; place-items: center; }
.graph-control:hover { background: var(--surface-2); border-color: var(--brand); color: var(--brand); }

.graph-node-label { font-size: 11px; fill: #ffffff; pointer-events: none; font-weight: 700; text-shadow: 0 1px 4px rgba(0,0,0,0.95), 0 0 10px rgba(0,0,0,0.85); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; }
.graph-link { fill: none; stroke: #38bdf8; stroke-opacity: 0.28; transition: stroke-opacity 0.2s, stroke-width 0.2s; }
.graph-link.highlighted { stroke-opacity: 0.95 !important; stroke-width: 2.4px !important; }
.graph-link.dimmed { stroke-opacity: 0.04 !important; }
.graph-link.violation { stroke: #f43f5e !important; stroke-opacity: 0.95 !important; stroke-dasharray: 5, 4; animation: dash-pulse 1.4s linear infinite; }
.graph-link-label { fill: #f43f5e; font-size: 9.5px; font-weight: 800; pointer-events: none; text-shadow: 0 1px 3px rgba(0,0,0,0.9); }
.graph-node-group { transition: opacity 0.2s ease-out; }
.graph-node-group.dimmed { opacity: 0.08 !important; }
.graph-node-group.highlighted { opacity: 1 !important; }

@keyframes dash-pulse { from { stroke-dashoffset: 18; } to { stroke-dashoffset: 0; } }

.drawer-overlay { position: fixed; inset: 0; background: rgba(0, 0, 0, 0.7); backdrop-filter: blur(5px); z-index: 10000; display: none; }
.drawer-overlay.open { display: block; }
.drawer { position: fixed; right: 0; top: 0; height: 100vh; width: min(480px, 92vw); background: var(--surface); border-left: 1px solid var(--border); box-shadow: -15px 0 45px rgba(0,0,0,0.85); z-index: 10001; transform: translateX(100%); transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1); display: flex; flex-direction: column; color: white; }
.drawer.open { transform: translateX(0); }
.drawer-header { padding: 20px; border-bottom: 1px solid var(--border); display: flex; justify-content: space-between; gap: 12px; background: #060913; }
.drawer-title { font-size: 17px; font-weight: 800; color: white; }
.drawer-kind { color: var(--brand); font-size: 10px; font-weight: 800; text-transform: uppercase; margin-top: 4px; letter-spacing: 0.08em; }
.drawer-close { border: 0; background: var(--surface-2); color: white; width: 32px; height: 32px; border-radius: 8px; font-size: 16px; }
.drawer-body { overflow-y: auto; padding: 20px; }
.detail-section { margin-bottom: 24px; }
.detail-section-title { font-size: 10px; text-transform: uppercase; letter-spacing: 0.1em; color: var(--muted); font-weight: 800; margin-bottom: 10px; }
.detail-property { display: grid; grid-template-columns: 110px 1fr; gap: 12px; padding: 9px 0; border-bottom: 1px solid var(--border); }
.detail-key { color: var(--muted); font-size: 11px; }
.detail-value { color: var(--text); font-size: 12px; overflow-wrap: anywhere; font-weight: 500; }
.detail-action { width: 100%; border: 1px solid rgba(56, 189, 248, 0.4); background: var(--brand-soft); color: var(--brand); border-radius: 8px; padding: 10px 12px; font-size: 12px; font-weight: 750; margin-top: 8px; transition: 0.15s; }
.detail-action:hover { background: var(--brand-dark); color: white; }

.search-view { display: none; }
.search-view.active { display: block; }
.search-result { cursor: pointer; }
.search-result:hover { background: rgba(56, 189, 248, 0.08); }
.kind-pill { font-size: 9px; font-weight: 800; padding: 4px 8px; border-radius: 6px; background: var(--surface-2); color: #94a3b8; text-transform: uppercase; }
.merkle { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 10px; color: #94a3b8; word-break: break-all; background: #050811; padding: 10px; border-radius: 8px; margin-top: 8px; border: 1px solid var(--border); }
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
                    <span>Company Graph</span>
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
            <div class="workspace-mini" onclick="promptWorkspaceSwitch()">
                <div class="workspace-mini-name" id="sidebar-workspace">{{ .WorkspaceName }}</div>
                <div class="workspace-mini-meta" id="sidebar-meta">Loading workspace...</div>
                <div class="trust-mini">✓ Transactionally verified</div>
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
                    <span class="breadcrumb-current" id="workspace-breadcrumb">{{ .WorkspaceName }}</span>
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
                    <div class="kpi-card" style="border-left: 4px solid var(--brand);">
                        <div class="kpi-label" style="color:var(--brand);">Idempotency</div>
                        <div class="kpi-value" id="stat-idempotency">—</div>
                        <div class="kpi-foot">Safeguarded transactions</div>
                    </div>
                </div>

                <div class="kpi-grid" style="grid-template-columns: repeat(4, minmax(0, 1fr)); margin-bottom: 20px;">
                    <div class="kpi-card" style="border-left: 4px solid var(--green);">
                        <div class="kpi-label" style="color:var(--green);">Financial ROI (Saved)</div>
                        <div class="kpi-value" id="stat-cost-saved">$—</div>
                        <div class="kpi-foot"><span id="stat-tokens-saved">—</span> context tokens conserved</div>
                    </div>
                    <div class="kpi-card" style="border-left: 4px solid var(--brand);">
                        <div class="kpi-label" style="color:var(--brand);">Cold Start Latency</div>
                        <div class="kpi-value" id="stat-cold-start">—</div>
                        <div class="kpi-foot">Recorded cold-start telemetry</div>
                    </div>
                    <div class="kpi-card" style="border-left: 4px solid var(--amber);">
                        <div class="kpi-label" style="color:var(--amber);">Active Agent Swarms</div>
                        <div class="kpi-value" id="stat-active-agents">—</div>
                        <div class="kpi-foot">Concurrent AI/CLI sessions</div>
                    </div>
                    <div class="kpi-card" style="border-left: 4px solid var(--red);">
                        <div class="kpi-label" style="color:var(--red);">Drift Prevented</div>
                        <div class="kpi-value" id="stat-drift-count">0</div>
                        <div class="kpi-foot">Quarantined contract violations</div>
                    </div>
                </div>

                <!-- NEW: Languages breakdown (GitHub-style) -->
                <div class="panel" style="margin-bottom: 20px;">
                    <div class="panel-header">
                        <div>
                            <div class="panel-title">Languages</div>
                            <div class="panel-subtitle">Codebase composition across all scanned repositories</div>
                        </div>
                    </div>
                    <div class="panel-body">
                        <div class="lang-bar-container" id="lang-bar"></div>
                        <div class="lang-legend" id="lang-legend">
                            <span style="color:var(--muted); font-size:11px;">Loading languages...</span>
                        </div>
                    </div>
                </div>

                <!-- NEW: Doc × Code Drift overview -->
                <div class="drift-grid">
                    <div class="drift-card">
                        <div class="drift-card-title">📄 Documentation Claims</div>
                        <div class="drift-row">
                            <span class="drift-row-label">Total documented capabilities</span>
                            <span class="drift-row-val" id="drift-total-docs">—</span>
                        </div>
                        <div class="drift-row">
                            <span class="drift-row-label">Supported by code</span>
                            <span class="drift-row-val" id="drift-doc-supported" style="color:var(--green);">—</span>
                        </div>
                        <div class="drift-row">
                            <span class="drift-row-label">Unverified</span>
                            <span class="drift-row-val" id="drift-doc-unverified" style="color:var(--amber);">—</span>
                        </div>
                        <div class="drift-row">
                            <span class="drift-row-label">Contradicted</span>
                            <span class="drift-row-val" id="drift-doc-contradicted" style="color:var(--red);">—</span>
                        </div>
                    </div>
                    <div class="drift-card">
                        <div class="drift-card-title">🔍 Knowledge Drift</div>
                        <div class="drift-row">
                            <span class="drift-row-label">Doc → Code drift (documented, not implemented)</span>
                            <span class="drift-row-val" id="drift-doc-to-code" style="color:var(--amber);">—</span>
                        </div>
                        <div class="drift-row">
                            <span class="drift-row-label">Code → Doc drift (implemented, not documented)</span>
                            <span class="drift-row-val" id="drift-code-to-doc" style="color:var(--amber);">—</span>
                        </div>
                        <div class="drift-row">
                            <span class="drift-row-label">Undocumented code entities</span>
                            <span class="drift-row-val" id="drift-undocumented-code" style="color:var(--amber);">—</span>
                        </div>
                        <div class="drift-row">
                            <span class="drift-row-label">Unimplemented doc claims</span>
                            <span class="drift-row-val" id="drift-unimplemented-docs" style="color:var(--amber);">—</span>
                        </div>
                    </div>
                </div>

                <div class="panel" style="margin-bottom: 20px;">
                    <div class="panel-header">
                        <div>
                            <div class="panel-title">Scanned repositories</div>
                            <div class="panel-subtitle">Active codebases bound to workspace '<span id="repo-list-ws">{{ .WorkspaceName }}</span>'. Click any repository to filter its package hierarchy.</div>
                        </div>
                    </div>
                    <div class="list" id="scanned-repos-list">
                        <div class="list-row"><div class="row-main"><div class="row-title">Loading repositories...</div></div></div>
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
                    <span class="breadcrumb-current" id="architecture-breadcrumb">Company Graph</span>
                </div>
                <div class="hero">
                    <div>
                        <h1 class="hero-title" id="architecture-title">Company Graph</h1>
                        <div class="hero-subtitle" id="architecture-subtitle">Pristine Graphify-inspired force-directed topology map.</div>
                    </div>
                    <div class="graph-toolbar">
                        <button class="graph-button" id="btn-mode-repo" onclick="setArchitectureMode('repository')">Repositories</button>
                        <button class="graph-button" id="btn-mode-package" onclick="setArchitectureMode('package')">All Packages</button>
                        <button class="graph-button" id="btn-mode-full" onclick="setArchitectureMode('full')">Full Knowledge Graph</button>
                        <button class="graph-button" onclick="goUpArchitecture()">← Up Level</button>
                        <button class="graph-button" onclick="loadArchitecture(state.currentLevel, state.currentFocus)">Refresh</button>
                    </div>
                </div>

                <div class="panel graph-panel">
                    <div class="panel-header">
                        <div>
                            <div class="panel-title">Interactive topology explorer</div>
                            <div class="panel-subtitle" id="graph-description">Workspace-level repository mesh.</div>
                        </div>
                        <div class="graph-toolbar">
                            <button class="graph-button primary" onclick="fitGraph()">Fit Canvas</button>
                        </div>
                    </div>
                    <div class="graph-layout">
                        <div class="graph-wrap" id="graph-wrap">
                            <svg id="graph"></svg>
                            <div id="graph-empty" class="graph-empty" style="display:none;">No architecture data available for this view.</div>
                            <div class="graph-controls">
                                <button class="graph-control" onclick="toggleFullscreen()" title="Toggle Fullscreen">⛶</button>
                                <button class="graph-control" onclick="zoomGraph(1.25)" title="Zoom in">+</button>
                                <button class="graph-control" onclick="zoomGraph(0.8)" title="Zoom out">−</button>
                                <button class="graph-control" onclick="fitGraph()" title="Fit graph">⌂</button>
                            </div>
                            <div class="graph-help">Click to inspect · Double-click to expand · Hover to spotlight connections</div>
                        </div>
                        <aside class="graph-side-panel">
                            <div class="side-panel-header">
                                <span>COMMUNITIES</span>
                                <span style="font-size:10px; color:var(--brand); cursor:pointer;" onclick="toggleAllCommunities()">Toggle All</span>
                            </div>
                            <div class="side-panel-list" id="communities-list">
                                <div style="color:var(--muted); font-size:11px; padding:8px;">Analyzing topology...</div>
                            </div>
                        </aside>
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
var urlParams = new URLSearchParams(window.location.search);
var WORKSPACE = urlParams.get("workspace") || "{{ .WorkspaceName }}";

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
    hoveredNodeId: null
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

function promptWorkspaceSwitch() {
    var ws = prompt("Enter workspace name to switch context:", WORKSPACE);
    if (ws && ws.trim() !== "" && ws !== WORKSPACE) {
        window.location.search = "?workspace=" + encodeURIComponent(ws.trim());
    }
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
    setText("stat-idempotency", formatNumber(s.idempotency_safeguards));

    setText("stat-supported", formatNumber(s.supported_claims));
    setText("stat-unverified", formatNumber(s.unverified_claims));
    setText("stat-contradicted", formatNumber(s.contradicted));

    setText("explorer-repos", formatNumber(s.repositories));
    setText("explorer-packages", formatNumber(s.packages));
    setText("explorer-entities", formatNumber(s.entities));

    setText("sidebar-meta", formatNumber(s.repositories) + " repositories · " + formatNumber(s.entities) + " entities");
    setText("sidebar-workspace", s.workspace || WORKSPACE);
    setText("workspace-breadcrumb", s.workspace || WORKSPACE);
    setText("repo-list-ws", s.workspace || WORKSPACE);

    setText("stat-cost-saved", "$" + Number(s.estimated_cost_saved_usd || 0).toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2}));
    setText("stat-tokens-saved", formatNumber(s.tokens_saved));
    setText("stat-cold-start", s.cold_start_latency_known ? Number(s.cold_start_latency_ms).toFixed(1) + " ms" : "—");
    setText("stat-active-agents", formatNumber(s.active_agents_count || 0));
    setText("stat-drift-count", formatNumber(s.quarantined_count));

    // NEW: render languages
    renderLanguages(s.languages_breakdown || []);

    // NEW: render drift
    renderDrift(s.drift || {});

    renderHubs(s.top_hubs || []);
    renderAttention(s.needs_attention || []);
    renderEvidence(s.recent_evidence || []);
    renderTrust();
    renderScannedRepos(s.repo_stats || [], s.repositories_list || []);
}

// NEW: GitHub-style language bar
function renderLanguages(langs) {
    var bar = document.getElementById("lang-bar");
    var legend = document.getElementById("lang-legend");
    if (!bar || !legend) return;

    if (!langs || langs.length === 0) {
        bar.innerHTML = '<div style="width:100%; background:#1a2035;"></div>';
        legend.innerHTML = '<span style="color:var(--muted); font-size:11px;">No language data yet — run garuda analyze on at least one repository.</span>';
        return;
    }

    // Sort by count descending
    langs.sort(function(a, b) { return b.count - a.count; });

    // Build bar
    bar.innerHTML = "";
    langs.forEach(function(lang) {
        var seg = document.createElement("div");
        seg.className = "lang-bar-segment";
        seg.style.width = lang.percentage + "%";
        seg.style.background = lang.color;
        seg.title = lang.name + " " + lang.percentage.toFixed(1) + "%";
        bar.appendChild(seg);
    });

    // Build legend
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

// NEW: Drift overview
function renderDrift(d) {
    setText("drift-total-docs", formatNumber(d.total_document_claims));
    setText("drift-doc-supported", formatNumber(d.supported_claims));
    setText("drift-doc-unverified", formatNumber(d.unverified_claims));
    setText("drift-doc-contradicted", formatNumber(d.contradicted_claims));
    setText("drift-doc-to-code", formatNumber(d.doc_to_code_drift_count));
    setText("drift-code-to-doc", formatNumber(d.code_to_doc_drift_count));
    setText("drift-undocumented-code", formatNumber(d.undocumented_code));
    setText("drift-unimplemented-docs", formatNumber(d.unimplemented_docs));
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

// NEW: Enhanced repo list with per-repo language bar
function renderScannedRepos(repoStats, fallbackList) {
    var list = document.getElementById("scanned-repos-list");
    if (!list) return;

    // Use rich stats if available, otherwise fall back to plain names
    if (!repoStats || repoStats.length === 0) {
        if (!fallbackList || fallbackList.length === 0) {
            list.innerHTML = '<div class="list-row"><div class="row-main"><div class="row-title">No repositories registered in this workspace.</div></div></div>';
            return;
        }
        // Fallback: simple list
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

    // Rich repo cards with language bars
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
    var subtitle = "Pristine Graphify-inspired force-directed topology map.";
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

    var linkDist = state.currentLevel === "full" ? 65 : (state.currentLevel === "repository" ? 140 : 90);
    var chargeForce = state.currentLevel === "full" ? -260 : (state.currentLevel === "repository" ? -750 : -420);

    var simulation = d3.forceSimulation(nodes)
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

    simulation.on("tick", function() {
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
	applySecurityHeaders(w)
	ctx := r.Context()
	wsName := strings.TrimSpace(r.URL.Query().Get("workspace"))

	pgStore, ok := s.store.(*store.PostgresStore)
	if wsName == "" && ok && pgStore != nil {
		_ = pgStore.Pool().QueryRow(ctx, `SELECT name FROM workspaces ORDER BY updated_at DESC LIMIT 1`).Scan(&wsName)
	}
	if wsName == "" {
		wsName = "default"
	}

	data := DashboardData{
		TenantID:      dashboardTenantID,
		WorkspaceName: wsName,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = parsedProdDashboardTmpl.Execute(w, data)
}

func inferRepositoryFromPackage(pkg string) string {
	if pkg == "" {
		return "unknown"
	}

	pkg = strings.TrimPrefix(pkg, "file://")
	if idx := strings.Index(pkg, "/go/pkg/mod/"); idx != -1 {
		pkg = pkg[idx+len("/go/pkg/mod/"):]
		if atIdx := strings.Index(pkg, "@"); atIdx != -1 {
			pkg = pkg[:atIdx]
		}
	}

	parts := strings.Split(strings.Trim(pkg, "/"), "/")
	firstSegment := parts[0]

	if !strings.Contains(firstSegment, ".") && firstSegment != "garuda" && firstSegment != "myshra777-ai" {
		return "stdlib"
	}

	if strings.Contains(pkg, "myshra777-ai/garuda") || strings.HasPrefix(pkg, "github.com/myshra777-ai/garuda") {
		return "garuda"
	}

	if len(parts) >= 3 && (parts[0] == "github.com" || parts[0] == "golang.org") {
		return parts[0] + "/" + parts[1] + "/" + parts[2]
	}
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}
	return pkg
}

// / NEW: Normalize language names to canonical display form
func normalizeLanguageName(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "go", "golang":
		return "Go"
	case "python", "py":
		return "Python"
	case "typescript", "ts":
		return "TypeScript"
	case "javascript", "js":
		return "JavaScript"
	case "rust", "rs":
		return "Rust"
	case "java":
		return "Java"
	case "c":
		return "C"
	case "c++", "cpp":
		return "C++"
	case "c#", "csharp":
		return "C#"
	case "ruby", "rb":
		return "Ruby"
	case "php":
		return "PHP"
	case "swift":
		return "Swift"
	case "kotlin", "kt":
		return "Kotlin"
	case "shell", "bash", "sh":
		return "Shell"
	case "html":
		return "HTML"
	case "css", "scss":
		return "CSS"
	}
	return lang
}

// computeRepoStats — matches entities by repository_id, not name-LIKE.
// Fixes garuda-self and grpc-go showing 0 entities.
func computeRepoStats(ctx context.Context, pgStore *store.PostgresStore, workspaceID uuid.UUID) []RepoStatDTO {
	rows, err := pgStore.Pool().Query(ctx, `
		SELECT 
			id,
			COALESCE(name, '') AS name,
			COALESCE(analysis_status, 'pending') AS status,
			COALESCE(current_commit, '') AS commit,
			last_analyzed_at
		FROM repositories 
		WHERE workspace_id = $1
		ORDER BY name
	`, workspaceID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var stats []RepoStatDTO
	for rows.Next() {
		var repoID uuid.UUID
		var name, status, commit string
		var lastAnalyzed *time.Time
		if err := rows.Scan(&repoID, &name, &status, &commit, &lastAnalyzed); err != nil {
			continue
		}

		stat := RepoStatDTO{
			Name:           name,
			AnalysisStatus: status,
			CurrentCommit:  commit,
		}
		if lastAnalyzed != nil {
			stat.LastAnalyzed = lastAnalyzed.Format(time.RFC3339)
		}

		// FIXED: match by repository_id — not package LIKE
		_ = pgStore.Pool().QueryRow(ctx, `
			SELECT COUNT(*)::int, COUNT(DISTINCT file_path)::int
			FROM entities 
			WHERE workspace_id = $1 AND repository_id = $2 AND kind != 'external'
		`, workspaceID, repoID).Scan(&stat.Entities, &stat.Files)

		_ = pgStore.Pool().QueryRow(ctx, `
			SELECT COUNT(*)::int 
			FROM claims c
			JOIN entities e ON e.id = c.from_entity_id
			WHERE c.workspace_id = $1 AND e.repository_id = $2
		`, workspaceID, repoID).Scan(&stat.Relationships)

		// Per-repo language breakdown — FIXED to use repository_id
		langRows, err := pgStore.Pool().Query(ctx, `
			SELECT 
				COALESCE(NULLIF(language, ''), '') AS lang,
				file_path,
				COUNT(*)::int AS cnt
			FROM entities 
			WHERE workspace_id = $1 AND repository_id = $2 AND kind != 'external'
			GROUP BY language, file_path
		`, workspaceID, repoID)
		if err == nil {
			repoLangs := make(map[string]int)
			repoTotal := 0
			for langRows.Next() {
				var lang, filePath string
				var cnt int
				if err := langRows.Scan(&lang, &filePath, &cnt); err != nil {
					continue
				}
				if lang == "" {
					lang = inferLanguageFromPath(filePath)
				}
				lang = normalizeLanguageName(lang)
				repoLangs[lang] += cnt
				repoTotal += cnt
			}
			langRows.Close()

			if repoTotal > 0 {
				var ls []LanguageDTO
				for l, c := range repoLangs {
					ls = append(ls, LanguageDTO{
						Name:       l,
						Count:      c,
						Percentage: float64(c) / float64(repoTotal) * 100.0,
						Color:      languageColor(l),
					})
				}
				sort.Slice(ls, func(i, j int) bool { return ls[i].Count > ls[j].Count })
				stat.Languages = ls
			}
		}

		stats = append(stats, stat)
	}
	return stats
}

// computeLanguageBreakdown — workspace-wide aggregation using shared normalizer
func computeLanguageBreakdown(ctx context.Context, pgStore *store.PostgresStore, workspaceID uuid.UUID) []LanguageDTO {
	rows, err := pgStore.Pool().Query(ctx, `
		SELECT 
			COALESCE(NULLIF(language, ''), '') AS lang,
			file_path,
			COUNT(*)::int AS cnt
		FROM entities 
		WHERE workspace_id = $1 AND kind != 'external'
		GROUP BY language, file_path
	`, workspaceID)
	if err != nil {
		return []LanguageDTO{}
	}
	defer rows.Close()

	langTotals := make(map[string]int)
	total := 0

	for rows.Next() {
		var lang, filePath string
		var cnt int
		if err := rows.Scan(&lang, &filePath, &cnt); err != nil {
			continue
		}
		if lang == "" {
			lang = inferLanguageFromPath(filePath)
		}
		lang = normalizeLanguageName(lang)
		langTotals[lang] += cnt
		total += cnt
	}

	if total == 0 {
		return []LanguageDTO{}
	}

	var result []LanguageDTO
	for lang, cnt := range langTotals {
		pct := float64(cnt) / float64(total) * 100.0
		result = append(result, LanguageDTO{
			Name:       lang,
			Count:      cnt,
			Percentage: pct,
			Color:      languageColor(lang),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	// Group languages < 0.5% into "Other"
	var major []LanguageDTO
	var otherCount int
	var otherPct float64
	for _, l := range result {
		if l.Percentage < 0.5 {
			otherCount += l.Count
			otherPct += l.Percentage
		} else {
			major = append(major, l)
		}
	}
	if otherCount > 0 {
		major = append(major, LanguageDTO{
			Name:       "Other",
			Count:      otherCount,
			Percentage: otherPct,
			Color:      languageColor("Other"),
		})
	}

	return major
}

// NEW: Compute doc-code drift summary
func computeDrift(ctx context.Context, pgStore *store.PostgresStore, workspaceID uuid.UUID, workspaceName string) DriftDTO {
	var d DriftDTO

	// Document claims by status
	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM document_claims WHERE workspace = $1
	`, workspaceName).Scan(&d.TotalDocumentClaims)

	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM document_claims WHERE workspace = $1 AND status = 'SUPPORTED'
	`, workspaceName).Scan(&d.SupportedClaims)

	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM document_claims WHERE workspace = $1 AND status = 'UNVERIFIED'
	`, workspaceName).Scan(&d.UnverifiedClaims)

	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM document_claims WHERE workspace = $1 AND status = 'CONTRADICTED'
	`, workspaceName).Scan(&d.ContradictedClaims)

	// Doc → Code drift = claims that have no matching code entity
	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM document_claims 
		WHERE workspace = $1 AND matched_entity_id IS NULL
	`, workspaceName).Scan(&d.DocToCodeDriftCount)

	// Code → Doc drift = code entities with no matching doc claim
	// Count entities (functions/methods/structs) that aren't referenced by any doc claim
	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM entities e
		WHERE e.workspace_id = $1 
		  AND e.kind IN ('function', 'method', 'struct', 'interface')
		  AND NOT EXISTS (
			SELECT 1 FROM document_claims dc 
			WHERE dc.workspace = $2 
			  AND dc.subject ILIKE '%' || e.name || '%'
		  )
	`, workspaceID, workspaceName).Scan(&d.CodeToDocDriftCount)

	// Undocumented code entities (exported only, filtered)
	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM entities e
		WHERE e.workspace_id = $1 
		  AND e.kind IN ('function', 'method', 'struct', 'interface')
		  AND e.is_exported = TRUE
		  AND NOT EXISTS (
			SELECT 1 FROM document_claims dc 
			WHERE dc.workspace = $2 
			  AND dc.subject ILIKE '%' || e.name || '%'
		  )
	`, workspaceID, workspaceName).Scan(&d.UndocumentedCode)

	// Unimplemented docs = doc claims with no matched entity
	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM document_claims 
		WHERE workspace = $1 AND matched_entity_id IS NULL
	`, workspaceName).Scan(&d.UnimplementedDocs)

	return d
}

func (s *Server) HandleDashboardStats(w http.ResponseWriter, r *http.Request) {
	applySecurityHeaders(w)
	ctx := r.Context()
	tenantID := getDashboardTenant()

	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		http.Error(w, "dashboard store unavailable", http.StatusServiceUnavailable)
		return
	}

	workspaceName := strings.TrimSpace(r.URL.Query().Get("workspace"))
	var workspaceID uuid.UUID

	if workspaceName != "" {
		err := pgStore.Pool().QueryRow(ctx, `SELECT id, name FROM workspaces WHERE name = $1 LIMIT 1`, workspaceName).Scan(&workspaceID, &workspaceName)
		if err != nil {
			_ = pgStore.Pool().QueryRow(ctx, `SELECT id, name FROM workspaces ORDER BY updated_at DESC LIMIT 1`).Scan(&workspaceID, &workspaceName)
		}
	} else {
		err := pgStore.Pool().QueryRow(ctx, `SELECT id, name FROM workspaces ORDER BY updated_at DESC LIMIT 1`).Scan(&workspaceID, &workspaceName)
		if err != nil {
			workspaceName = "default"
		}
	}

	repoList := make([]string, 0)
	repoRows, err := pgStore.Pool().Query(ctx, `SELECT name FROM repositories WHERE workspace_id = $1 ORDER BY name`, workspaceID)
	if err == nil {
		defer repoRows.Close()
		for repoRows.Next() {
			var name string
			if err := repoRows.Scan(&name); err == nil && name != "" {
				repoList = append(repoList, name)
			}
		}
	}
	repositories := len(repoList)

	// NEW: per-repo stats with languages
	repoStats := computeRepoStats(ctx, pgStore, workspaceID)

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

	// NEW: workspace-wide language breakdown
	languagesBreakdown := computeLanguageBreakdown(ctx, pgStore, workspaceID)

	var relationships int
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COUNT(*)::int FROM claims WHERE workspace_id = $1`, workspaceID).Scan(&relationships)

	var activeContradictions int
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COALESCE(COUNT(*)::int, 0) FROM claim_verifications WHERE workspace_id = $1 AND status = 'CONTRADICTED'`, workspaceID).Scan(&activeContradictions)

	var supportedClaims int
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COALESCE(COUNT(*)::int, 0) FROM claim_verifications WHERE workspace_id = $1 AND status = 'SUPPORTED'`, workspaceID).Scan(&supportedClaims)

	var unverifiedClaims int
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COALESCE(COUNT(*)::int, 0) FROM claim_verifications WHERE workspace_id = $1 AND status = 'UNVERIFIED'`, workspaceID).Scan(&unverifiedClaims)

	var idempotencySafeguards int
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COALESCE(COUNT(*)::int, 0) FROM idempotency_keys WHERE tenant_id = $1`, tenantID).Scan(&idempotencySafeguards)

	var pendingDecisions int
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COALESCE(COUNT(*)::int, 0) FROM decisions WHERE status = 'PENDING'`).Scan(&pendingDecisions)

	var canonicalDecisions int
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COALESCE(COUNT(*)::int, 0) FROM decisions WHERE tenant_id = $1 AND status = 'CANONICAL'`, tenantID).Scan(&canonicalDecisions)

	if supportedClaims == 0 && unverifiedClaims == 0 && activeContradictions == 0 {
		unverifiedClaims = relationships
	}
	totalClaims := relationships
	if totalClaims == 0 {
		totalClaims = supportedClaims + unverifiedClaims + activeContradictions
	}

	var tokensSaved int64
	var costSavedUSD float64
	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COALESCE(SUM(tokens_saved), 0), COALESCE(SUM(cost_saved_usd), 0)
		FROM telemetry_events
	`).Scan(&tokensSaved, &costSavedUSD)

	var activeAgents int
	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COALESCE(MAX(active_agents), 0)::int
		FROM telemetry_events
		WHERE created_at > NOW() - INTERVAL '24 hours'
	`).Scan(&activeAgents)

	var coldStartLatency float64
	var coldStartSamples int
	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COALESCE(AVG(cold_start_latency_ms), 0), COUNT(cold_start_latency_ms)
		FROM telemetry_events
		WHERE created_at > NOW() - INTERVAL '7 days'
	`).Scan(&coldStartLatency, &coldStartSamples)

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
		SELECT trace_id, source_service || ' → (' || operation || ')', 'Runtime Trace', observed_at
		FROM runtime_observations
		WHERE workspace_id = $1
		ORDER BY observed_at DESC
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
	latestHash := "Genesis"
	parentHash := "Genesis"
	trustStatus := "Genesis"
	var latestBlock int64 = 1
	if latestSnap != nil {
		trustStatus = "Verified"
		latestHash = latestSnap.SnapshotHash
		latestBlock = latestSnap.BlockHeight
		if latestSnap.ParentSnapshotID != nil {
			parentHash = latestSnap.ParentSnapshotID.String()
		}
	}

	// NEW: compute drift
	drift := computeDrift(ctx, pgStore, workspaceID, workspaceName)

	resp := WorkspaceStatsResponse{
		Workspace:             workspaceName,
		Repositories:          repositories,
		RepositoriesList:      repoList,
		RepoStats:             repoStats,
		Packages:              packages,
		Entities:              entities,
		Relationships:         relationships,
		CrossRepoLinks:        crossRepoLinks,
		Files:                 files,
		ExportedEntities:      exportedEntities,
		ArchitecturalHubs:     len(topHubs),
		TopHubs:               topHubs,
		TotalClaims:           totalClaims,
		SupportedClaims:       supportedClaims,
		Contradicted:          activeContradictions,
		UnverifiedClaims:      unverifiedClaims,
		NeedsAttention:        needsAttention,
		RecentEvidence:        recentEvidence,
		CanonicalDecisions:    canonicalDecisions,
		QuarantinedCount:      activeContradictions,
		LatestBlockHeight:     latestBlock,
		LatestMerkleHash:      latestHash,
		ParentMerkleHash:      parentHash,
		TrustStatus:           trustStatus,
		LastUpdated:           time.Now().UTC().Format(time.RFC3339),
		PendingDecisions:      pendingDecisions,
		IdempotencySafeguards: idempotencySafeguards,
		TokensSaved:           tokensSaved,
		EstimatedCostSavedUSD: costSavedUSD,
		ColdStartLatencyMs:    coldStartLatency,
		ColdStartLatencyKnown: coldStartSamples > 0,
		ActiveAgentsCount:     activeAgents,
		LanguagesBreakdown:    languagesBreakdown,
		Drift:                 drift,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleDashboardSearch(w http.ResponseWriter, r *http.Request) {
	applySecurityHeaders(w)
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

	workspaceName := strings.TrimSpace(r.URL.Query().Get("workspace"))
	var workspaceID uuid.UUID

	if workspaceName != "" {
		err := pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE name = $1 LIMIT 1`, workspaceName).Scan(&workspaceID)
		if err != nil {
			_ = pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces ORDER BY updated_at DESC LIMIT 1`).Scan(&workspaceID)
		}
	} else {
		_ = pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces ORDER BY updated_at DESC LIMIT 1`).Scan(&workspaceID)
	}

	limit := normalizeLimit(r.URL.Query().Get("limit"), 50, 100)

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
		workspaceID, searchPattern, limit,
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
	applySecurityHeaders(w)
	ctx := r.Context()
	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}

	workspaceName := strings.TrimSpace(r.URL.Query().Get("workspace"))
	var workspaceID uuid.UUID

	if workspaceName != "" {
		err := pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE name = $1 LIMIT 1`, workspaceName).Scan(&workspaceID)
		if err != nil {
			_ = pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces ORDER BY updated_at DESC LIMIT 1`).Scan(&workspaceID)
		}
	} else {
		_ = pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces ORDER BY updated_at DESC LIMIT 1`).Scan(&workspaceID)
	}

	level := r.URL.Query().Get("level")
	if level == "" {
		level = "repository"
	}
	focus := strings.TrimSpace(r.URL.Query().Get("focus"))

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

	if level == "full" {
		for _, e := range entityMap {
			if e.Repo == "stdlib" || e.Package == "" {
				continue
			}
			if _, exists := nodesMap[e.Package]; !exists {
				nodesMap[e.Package] = GraphNodeDTO{
					ID: e.Package, Label: e.Package, Kind: "package",
					Repo: e.Repo, Package: e.Package, Status: "SUPPORTED", Exported: true, Count: 1,
				}
			} else {
				node := nodesMap[e.Package]
				node.Count++
				nodesMap[e.Package] = node
			}
		}

		for _, c := range claims {
			src := entityMap[c.from]
			tgt := entityMap[c.to]
			if src.Package == "" || tgt.Package == "" || src.Repo == "stdlib" || tgt.Repo == "stdlib" {
				continue
			}
			if src.Package != tgt.Package {
				edgeKey := src.Package + "->" + tgt.Package
				edge := edgesMap[edgeKey]
				if edge.Source == "" {
					edge = GraphEdgeDTO{ID: edgeKey, Source: src.Package, Target: tgt.Package, Type: "PACKAGE_DEPENDENCY", Status: "SUPPORTED", Count: 0, Confidence: 1.0}
				}
				edge.Count++
				edgesMap[edgeKey] = edge
			}
		}
	} else if level == "repository" {
		repoCounts := make(map[string]int)
		for _, e := range entityMap {
			if e.Repo == "stdlib" || e.Repo == "" {
				continue
			}
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

		if _, exists := nodesMap["garuda"]; !exists {
			nodesMap["garuda"] = GraphNodeDTO{ID: "garuda", Label: "garuda", Kind: "repository", Repo: "garuda", Status: "SUPPORTED", Exported: true, Count: 100}
		}

		for _, c := range claims {
			src := entityMap[c.from]
			tgt := entityMap[c.to]
			if src.Repo == "" || tgt.Repo == "" || src.Repo == "stdlib" || tgt.Repo == "stdlib" {
				continue
			}
			if src.Repo != tgt.Repo {
				edgeKey := src.Repo + "->" + tgt.Repo
				edge := edgesMap[edgeKey]
				if edge.Source == "" {
					edge = GraphEdgeDTO{ID: edgeKey, Source: src.Repo, Target: tgt.Repo, Type: "CROSS_REPO_BRIDGE", Status: "SUPPORTED", Label: "depends on", Count: 0, Confidence: 1.0}
				}
				edge.Count++
				edgesMap[edgeKey] = edge
			}
		}

		for repo := range nodesMap {
			if repo != "garuda" && !strings.HasPrefix(repo, "ext-") {
				edgeKey := "garuda->" + repo
				if _, exists := edgesMap[edgeKey]; !exists {
					edgesMap[edgeKey] = GraphEdgeDTO{ID: edgeKey, Source: "garuda", Target: repo, Type: "MODULE_DEPENDENCY", Status: "SUPPORTED", Label: "imports", Count: 1, Confidence: 1.0}
				}
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
			if e.Repo == "stdlib" || e.Package == "" {
				continue
			}
			if focus != "" && e.Repo != focus {
				continue
			}
			pkgCounts[e.Package]++
			if _, exists := nodesMap[e.Package]; !exists {
				nodesMap[e.Package] = GraphNodeDTO{ID: e.Package, Label: e.Package, Kind: "package", Repo: e.Repo, Package: e.Package, Status: "SUPPORTED", Exported: true}
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
			if src.Package == "" || tgt.Package == "" || src.Repo == "stdlib" || tgt.Repo == "stdlib" {
				continue
			}
			if focus != "" && (src.Repo != focus || tgt.Repo != focus) {
				continue
			}
			if src.Package != tgt.Package {
				edgeKey := src.Package + "->" + tgt.Package
				edge := edgesMap[edgeKey]
				if edge.Source == "" {
					edge = GraphEdgeDTO{ID: edgeKey, Source: src.Package, Target: tgt.Package, Type: "STATIC_DEPENDENCY", Status: "SUPPORTED", Count: 0, Confidence: 1.0}
				}
				edge.Count++
				edgesMap[edgeKey] = edge
			}
		}

		for _, cv := range contras {
			src := entityMap[cv.src]
			if src.Package != "" && (focus == "" || src.Repo == focus) {
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
			if focus == "" || e.Package == focus {
				nodesMap[e.ID] = GraphNodeDTO{ID: e.ID, Label: e.Name, Kind: e.Kind, Repo: e.Repo, Package: e.Package, Exported: e.Exported, Status: "SUPPORTED"}
			}
		}
		for _, c := range claims {
			src := entityMap[c.from]
			tgt := entityMap[c.to]
			if focus == "" || src.Package == focus || tgt.Package == focus {
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
	applySecurityHeaders(w)
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
