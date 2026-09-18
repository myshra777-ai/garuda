// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/tenant"
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

// PolicyEvaluationDTO — one enforcement decision for the dashboard.
type PolicyEvaluationDTO struct {
	ID                 string   `json:"id"`
	PolicyID           string   `json:"policy_id"`
	PolicyTitle        string   `json:"policy_title"`
	Decision           string   `json:"decision"`
	Reason             string   `json:"reason"`
	EntityCount        int      `json:"entity_count"`
	ClaimCount         int      `json:"claim_count"`
	ContradictionCount int      `json:"contradiction_count"`
	MatchedPredicates  []string `json:"matched_predicates"`
	MerkleBlockHeight  *int64   `json:"merkle_block_height,omitempty"`
	AnchorValid        bool     `json:"anchor_valid"`
	EvaluatedAt        string   `json:"evaluated_at"`
	Actor              string   `json:"actor"`
	SubjectKind        string   `json:"subject_kind"`
	SubjectID          string   `json:"subject_id"`
}

// PolicyEnforcementResponse — aggregate for the dashboard.
type PolicyEnforcementResponse struct {
	Workspace          string                `json:"workspace"`
	ActivePolicies     int                   `json:"active_policies"`
	TotalEvaluations   int                   `json:"total_evaluations"`
	BlockedCount       int                   `json:"blocked_count"`
	ReviewCount        int                   `json:"review_count"`
	WarnCount          int                   `json:"warn_count"`
	AllowCount         int                   `json:"allow_count"`
	LatestEvaluations  []PolicyEvaluationDTO `json:"latest_evaluations"`
	FinalDecision      string                `json:"final_decision"`
	LatestMerkleAnchor *int64                `json:"latest_merkle_anchor,omitempty"`
}

// LanguageDTO — language breakdown per repo/workspace.
type LanguageDTO struct {
	Name       string  `json:"name"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
	Color      string  `json:"color"`
}

// RepoStatDTO — per-repository breakdown.
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

// DriftDTO — document ↔ code drift summary.
type DriftDTO struct {
	TotalDocumentClaims int `json:"total_document_claims"`
	SupportedClaims     int `json:"supported_claims"`
	UnverifiedClaims    int `json:"unverified_claims"`
	ContradictedClaims  int `json:"contradicted_claims"`
	UndocumentedCode    int `json:"undocumented_code"`
	UnverifiedDocs      int `json:"unverified_docs"`
	DocToCodeDriftCount int `json:"doc_to_code_drift_count"`
	CodeToDocDriftCount int `json:"code_to_doc_drift_count"`
}

// MeasuredMetric distinguishes three states that must never be
// collapsed into one number:
//
//	HasData = true          → the value is a real measurement
//	HasData = false         → the measurement was attempted but the
//	                          source has no rows for this scope/time
//	IsMeasured = false      → the measurement was not attempted, or
//	                          the query failed. Value is meaningless.
//
// The previous implementation returned 0 for all three cases,
// which read as "we measured and the answer is zero." This is the
// same epistemic failure mode that was fixed for the governance
// verdict in the MCP layer and for claim classifications at the DB
// layer.
type MeasuredMetric struct {
	// Value is the aggregated number. Valid only when IsMeasured &&
	// HasData. Clients must not display Value in any other case.
	Value float64 `json:"value"`

	// IsMeasured is true iff the backing query executed successfully.
	// False means we do not know the value — not that the value is
	// zero.
	IsMeasured bool `json:"is_measured"`

	// HasData is true iff the backing table had at least one row in
	// the aggregation window with a non-NULL value for this column.
	HasData bool `json:"has_data"`

	// RowCount is the number of rows the aggregate considered. Useful
	// for "we measured 0 of 0 rows" style messaging.
	RowCount int64 `json:"row_count"`

	// Reason is a human-readable explanation, populated when
	// IsMeasured is false or HasData is false.
	Reason string `json:"reason,omitempty"`
}

type WorkspaceStatsResponse struct {
	Workspace             string          `json:"workspace"`
	Repositories          int             `json:"repositories"`
	RepositoriesList      []string        `json:"repositories_list"`
	RepoStats             []RepoStatDTO   `json:"repo_stats"`
	Packages              int             `json:"packages"`
	Entities              int             `json:"entities"`
	Relationships         int             `json:"relationships"`
	CrossRepoLinks        int             `json:"cross_repo_links"`
	Files                 int             `json:"files"`
	ExportedEntities      int             `json:"exported_entities"`
	ArchitecturalHubs     int             `json:"architectural_hubs"`
	TopHubs               []HubDTO        `json:"top_hubs"`
	TotalClaims           int             `json:"total_claims"`
	StaticClaims          int             `json:"static_claims"`
	SupportedClaims       int             `json:"supported_claims"`
	Contradicted          int             `json:"contradicted"`
	UnverifiedClaims      int             `json:"unverified_claims"`
	VerificationAttempted bool            `json:"verification_attempted"`
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
	// Class B runtime metrics. Each carries explicit measurement state.
	// See MeasuredMetric doc comment above.
	TokensSavedMetric      MeasuredMetric `json:"tokens_saved_metric"`
	CostSavedUSDMetric     MeasuredMetric `json:"cost_saved_usd_metric"`
	ColdStartLatencyMetric MeasuredMetric `json:"cold_start_latency_metric"`
	ActiveAgentsMetric     MeasuredMetric `json:"active_agents_metric"`
	LanguagesBreakdown     []LanguageDTO  `json:"languages_breakdown"`
	Drift                  DriftDTO       `json:"drift"`
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
	Level      string         `json:"level"`
	Focus      string         `json:"focus"`
	Nodes      []GraphNodeDTO `json:"nodes"`
	Edges      []GraphEdgeDTO `json:"edges"`
	Truncated  bool           `json:"truncated,omitempty"`
	TotalNodes int            `json:"total_nodes,omitempty"`
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

// -----------------------------------------------------------------------------
// Package-level constants and helpers
// -----------------------------------------------------------------------------

const dashboardTenantID = tenant.CanonicalIDStr

// Parse the tenant UUID once at package load. Any malformed constant
// panics at process start, not on every request.

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

// writeJSONError writes a small JSON error body with the given status code.
func writeJSONError(w http.ResponseWriter, status int, msg string, extra map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body := map[string]any{"error": msg}
	for k, v := range extra {
		body[k] = v
	}
	_ = json.NewEncoder(w).Encode(body)
}

// proofIsInternallyValid checks that a stored Merkle proof is well-formed
// and structurally self-consistent. It does not re-derive the full
// inclusion proof against the current epoch root — that lives in the CLI
// `garuda policy verify` path. This is the dashboard's conservative
// answer to "should we show a green anchor chip?".
//
// Previously this was `merkle_proof IS NOT NULL`, which showed green for
// any non-null blob. That was a lie of omission.
func proofIsInternallyValid(proofJSON []byte) bool {
	if len(proofJSON) == 0 {
		return false
	}
	if v1, err := store.UnmarshalV1Proof(proofJSON); err == nil {
		if v1.LeafHash == "" || v1.EpochRoot == "" {
			return false
		}
		return true
	}
	var v0 struct {
		NewRoot string `json:"new_root"`
	}
	if err := json.Unmarshal(proofJSON, &v0); err != nil {
		return false
	}
	return v0.NewRoot != ""
}

// languageColor — standard language color palette.
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

// inferLanguageFromPath — infer language from file extension.
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

// normalizeLanguageName — canonical display form.
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

// -----------------------------------------------------------------------------
// Embedded Dashboard HTML
// -----------------------------------------------------------------------------

//go:embed templates/dashboard.html
var prodDashboardHTML string

//go:embed templates/not_found.html
var notFoundHTML string

var parsedProdDashboardTmpl = template.Must(
	template.New("dashboard").Parse(prodDashboardHTML),
)

var parsedNotFoundTmpl = template.Must(
	template.New("notfound").Parse(notFoundHTML),
)

// -----------------------------------------------------------------------------
// Dashboard HTTP Handlers
// -----------------------------------------------------------------------------

func (s *Server) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	applySecurityHeaders(w)
	ctx := r.Context()
	wsName := strings.TrimSpace(r.URL.Query().Get("workspace"))

	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}

	scope, err := workspaceScopeFromRequest(ctx, r, pgStore)
	if err != nil {
		writeWorkspaceNotFound(w, r, wsName)
		return
	}

	data := DashboardData{
		TenantID:      scope.TenantID.String(),
		WorkspaceName: scope.WorkspaceName,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = parsedProdDashboardTmpl.Execute(w, data)
}

// computeRepoStats — batched queries: 3 total instead of 3×N.
//
// Previous version issued three queries per repository
// (entity count, relationship count, language breakdown). On a
// workspace with 50 repositories that is 150 round-trips per
// dashboard load. This version issues one query for each dimension
// and assembles the per-repo DTOs in Go.
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
	statIndexByRepo := make(map[uuid.UUID]int)
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
		stats = append(stats, stat)
		statIndexByRepo[repoID] = len(stats) - 1
	}
	if len(stats) == 0 {
		return stats
	}

	// Batch 1: entity + file counts per repo
	if er, err := pgStore.Pool().Query(ctx, `
		SELECT repository_id, COUNT(*)::int, COUNT(DISTINCT file_path)::int
		FROM entities
		WHERE workspace_id = $1 AND kind != 'external' AND repository_id IS NOT NULL
		GROUP BY repository_id
	`, workspaceID); err == nil {
		for er.Next() {
			var rid uuid.UUID
			var ec, fc int
			if er.Scan(&rid, &ec, &fc) == nil {
				if idx, ok := statIndexByRepo[rid]; ok {
					stats[idx].Entities = ec
					stats[idx].Files = fc
				}
			}
		}
		er.Close()
	}

	// Batch 2: relationship counts per repo (edges whose source is in the repo)
	if rr, err := pgStore.Pool().Query(ctx, `
		SELECT e.repository_id, COUNT(*)::int
		FROM claims c
		JOIN entities e ON e.id = c.from_entity_id
		WHERE c.workspace_id = $1 AND e.repository_id IS NOT NULL
		GROUP BY e.repository_id
	`, workspaceID); err == nil {
		for rr.Next() {
			var rid uuid.UUID
			var cnt int
			if rr.Scan(&rid, &cnt) == nil {
				if idx, ok := statIndexByRepo[rid]; ok {
					stats[idx].Relationships = cnt
				}
			}
		}
		rr.Close()
	}

	// Batch 3: language breakdown per repo
	langTotals := make(map[uuid.UUID]map[string]int)
	if langRows, err := pgStore.Pool().Query(ctx, `
		SELECT repository_id, COALESCE(NULLIF(language, ''), '') AS lang, file_path
		FROM entities
		WHERE workspace_id = $1 AND kind != 'external' AND repository_id IS NOT NULL
	`, workspaceID); err == nil {
		for langRows.Next() {
			var rid uuid.UUID
			var lang, filePath string
			if langRows.Scan(&rid, &lang, &filePath) == nil {
				if lang == "" {
					lang = inferLanguageFromPath(filePath)
				}
				lang = normalizeLanguageName(lang)
				if langTotals[rid] == nil {
					langTotals[rid] = make(map[string]int)
				}
				langTotals[rid][lang]++
			}
		}
		langRows.Close()
	}
	for rid, langs := range langTotals {
		total := 0
		for _, c := range langs {
			total += c
		}
		if total == 0 {
			continue
		}
		var ls []LanguageDTO
		for l, c := range langs {
			ls = append(ls, LanguageDTO{
				Name:       l,
				Count:      c,
				Percentage: float64(c) / float64(total) * 100.0,
				Color:      languageColor(l),
			})
		}
		sort.Slice(ls, func(i, j int) bool { return ls[i].Count > ls[j].Count })
		if idx, ok := statIndexByRepo[rid]; ok {
			stats[idx].Languages = ls
		}
	}

	return stats
}

// computeLanguageBreakdown — workspace-wide aggregation.
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

// computeDrift — document ↔ code drift summary.
//
// NOTE: this function reads two tables that are keyed differently.
// document_claims is keyed by workspace name (string); entities is
// keyed by workspace_id (uuid). If a workspace is ever renamed, the
// document_claims queries will silently return zero. Fixing this
// requires adding workspace_id to document_claims — a schema change
// tracked separately. Until then the workspace name must not change.
// computeDrift — document ↔ code drift summary.
//
// All queries scope by workspace_id. The previous implementation scoped
// by the workspace name (string), which silently returned zero after a
// rename because the string in document_claims was left stale while the
// workspaces row kept its stable UUID.
func computeDrift(ctx context.Context, pgStore *store.PostgresStore, workspaceID uuid.UUID) DriftDTO {
	var d DriftDTO

	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM document_claims WHERE workspace_id = $1
	`, workspaceID).Scan(&d.TotalDocumentClaims)

	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM document_claims WHERE workspace_id = $1 AND status = 'SUPPORTED'
	`, workspaceID).Scan(&d.SupportedClaims)

	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM document_claims WHERE workspace_id = $1 AND status = 'UNVERIFIED'
	`, workspaceID).Scan(&d.UnverifiedClaims)

	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM document_claims WHERE workspace_id = $1 AND status = 'CONTRADICTED'
	`, workspaceID).Scan(&d.ContradictedClaims)

	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM document_claims
		WHERE workspace_id = $1 AND matched_entity_id IS NULL
	`, workspaceID).Scan(&d.DocToCodeDriftCount)

	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM entities e
		WHERE e.workspace_id = $1
		  AND e.kind IN ('function', 'method', 'struct', 'interface')
		  AND NOT EXISTS (
			SELECT 1 FROM document_claims dc
			WHERE dc.workspace_id = $1
			  AND dc.subject ILIKE '%' || e.name || '%'
		  )
	`, workspaceID).Scan(&d.CodeToDocDriftCount)

	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM entities e
		WHERE e.workspace_id = $1
		  AND e.kind IN ('function', 'method', 'struct', 'interface')
		  AND e.is_exported = TRUE
		  AND NOT EXISTS (
			SELECT 1 FROM document_claims dc
			WHERE dc.workspace_id = $1
			  AND dc.subject ILIKE '%' || e.name || '%'
		  )
	`, workspaceID).Scan(&d.UndocumentedCode)

	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM document_claims
		WHERE workspace_id = $1
		  AND matched_entity_id IS NOT NULL
		  AND status = 'UNVERIFIED'
	`, workspaceID).Scan(&d.UnverifiedDocs)

	return d
}

func (s *Server) HandleDashboardStats(w http.ResponseWriter, r *http.Request) {
	applySecurityHeaders(w)
	ctx := r.Context()

	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "dashboard store unavailable", nil)
		return
	}

	scope, err := workspaceScopeFromRequest(ctx, r, pgStore)
	if err != nil {
		writeWorkspaceNotFound(w, r, strings.TrimSpace(r.URL.Query().Get("workspace")))
		return
	}
	workspaceID := scope.WorkspaceID
	workspaceName := scope.WorkspaceName
	tenantID := scope.TenantID

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

	repoStats := computeRepoStats(ctx, pgStore, workspaceID)

	// Cross-repo bridges: query the resolved cross_repo_edges table
	// directly. Before this fix the number was derived by comparing
	// heuristic that was written for Go and produced nonsense on
	// Python and TypeScript. Every distinct dotted TS package pair
	// counted as a bridge.
	//
	// The real count comes from cross_repo_edges, populated by the
	// analyzer when a claim's source and target entities resolve to
	// different repositories within the same workspace.
	// Cross-repo bridges: distinct (from_repo, to_repo) pairs with at
	// least one edge between them. resolved = true is intentionally
	// NOT a filter: an unresolved edge is still a bridge — the
	// analyzer found an import from repo A into a package of repo B
	// and failed to map it to a specific entity in B. The dependency
	// exists. Filtering it out undercounts.
	//
	// The string concatenation is used because Postgres does not
	// support COUNT(DISTINCT (a, b)) directly; the ::text cast of a
	// UUID is unambiguous and short.
	var crossRepoLinks int
	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(DISTINCT (from_repo_id::text || '->' || to_repo_id::text))
		FROM cross_repo_edges
		WHERE workspace_id = $1
	`, workspaceID).Scan(&crossRepoLinks)

	var packages, entities, files, exportedEntities int
	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int, COUNT(DISTINCT package)::int, COUNT(DISTINCT file_path)::int, COUNT(*) FILTER (WHERE is_exported = TRUE)::int
		FROM entities WHERE workspace_id = $1 AND kind != 'external'
	`, workspaceID).Scan(&entities, &packages, &files, &exportedEntities)

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

	// FIX: tenant-scoped. Previously counted every tenant's pending
	// decisions, same class of leak as the Active Policies panel.
	var pendingDecisions int
	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COALESCE(COUNT(*)::int, 0)
		FROM decisions
		WHERE tenant_id = $1 AND status = 'PENDING'
	`, tenantID).Scan(&pendingDecisions)

	var canonicalDecisions int
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COALESCE(COUNT(*)::int, 0) FROM decisions WHERE tenant_id = $1 AND status = 'CANONICAL'`, tenantID).Scan(&canonicalDecisions)

	// Three orthogonal counts, kept separate on purpose:
	//
	//   - supportedClaims / unverifiedClaims / activeContradictions are
	//     runtime verification results. Zero when no evaluator has run.
	//
	//   - staticClaims is the count of semantic edges in the graph.
	//     Every one exists. None have been runtime-verified unless the
	//     counters above are non-zero.
	//
	// The frontend uses VerificationAttempted to distinguish
	// "attempted and found nothing" from "never attempted".
	var staticClaims int
	_ = pgStore.Pool().QueryRow(ctx,
		`SELECT COUNT(*)::int FROM claims WHERE workspace_id = $1`,
		workspaceID).Scan(&staticClaims)

	runtimeTotal := supportedClaims + unverifiedClaims + activeContradictions
	totalClaims := runtimeTotal
	if runtimeTotal == 0 && staticClaims > 0 {
		totalClaims = staticClaims
	}

	// Class B runtime metrics. These are deployment-wide telemetry
	// figures computed from telemetry_events, which is keyed by
	// instance_hash and session_id and has no tenant or workspace
	// column. A per-workspace value cannot be derived from it, and
	// showing the deployment-wide number on a workspace dashboard
	// would be read as this workspace's contribution. The four
	// metrics are therefore reported as not-measured on this path,
	// with the reason pointing to the operator view at /admin, which
	// is where the deployment-wide totals belong.
	//
	// These queries are workspace-agnostic on purpose. telemetry_events
	// is keyed by instance_hash and session_id, not by tenant or
	// workspace. Until the aggregates table (Session D) is introduced,
	// the values reflect every telemetry event on this deployment.
	// When a deployment hosts more than one tenant, this read path
	// must move to telemetry_aggregates, which is scoped per tenant.
	// Deployment-wide metric, not per-workspace. telemetry_events is
	// keyed by instance_hash and session_id only; there is no tenant
	// or workspace column. Showing the deployment-wide number here
	// would be read as this workspace's contribution, which it is
	// not. The operator view at /admin is where this figure belongs.
	var tokensSavedMetric = MeasuredMetric{
		IsMeasured: false,
		HasData:    false,
		Reason:     "deployment-wide metric; see /admin",
	}

	var costSavedMetric = MeasuredMetric{
		IsMeasured: false,
		HasData:    false,
		Reason:     "deployment-wide metric; see /admin",
	}

	var coldStartMetric = MeasuredMetric{
		IsMeasured: false,
		HasData:    false,
		Reason:     "deployment-wide metric; see /admin",
	}

	var activeAgentsMetric = MeasuredMetric{
		IsMeasured: false,
		HasData:    false,
		Reason:     "deployment-wide metric; see /admin",
	}

	hubRows, err := pgStore.Pool().Query(ctx, `
		SELECT e.id, e.name, e.kind, e.package,
		       COALESCE(r.name, 'unknown') AS repo, count(c.id) as callers
		FROM entities e
		JOIN claims c ON c.to_entity_id = e.id
		LEFT JOIN repositories r ON r.id = e.repository_id
		WHERE e.workspace_id = $1 AND e.kind != 'external'
		GROUP BY e.id, e.name, e.kind, e.package, r.name
		ORDER BY callers DESC
		LIMIT 5
	`, workspaceID)
	var topHubs []HubDTO
	if err == nil {
		defer hubRows.Close()
		for hubRows.Next() {
			var h HubDTO
			if err := hubRows.Scan(&h.ID, &h.Name, &h.Kind, &h.Package, &h.Repo, &h.Callers); err == nil {
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

	// FIX: read the decision log's current root from merkle_roots, not
	// the worker's periodic snapshot counter in merkle_snapshots. The
	// two advance at different rates. The dashboard "block height" is
	// the decision log's height.
	latestHash := "Genesis"
	parentHash := "Genesis"
	trustStatus := "Genesis"
	var latestBlock int64 = 0
	var rootHash string
	var rootHeight int64
	rootErr := pgStore.Pool().QueryRow(ctx, `
		SELECT root_hash, block_height
		FROM merkle_roots
		WHERE tenant_id = $1
	`, tenantID).Scan(&rootHash, &rootHeight)
	if rootErr == nil {
		trustStatus = "Verified"
		latestHash = rootHash
		latestBlock = rootHeight
	}

	drift := computeDrift(ctx, pgStore, workspaceID)

	resp := WorkspaceStatsResponse{
		Workspace:              workspaceName,
		Repositories:           repositories,
		RepositoriesList:       repoList,
		RepoStats:              repoStats,
		Packages:               packages,
		Entities:               entities,
		Relationships:          relationships,
		CrossRepoLinks:         crossRepoLinks,
		Files:                  files,
		ExportedEntities:       exportedEntities,
		ArchitecturalHubs:      len(topHubs),
		TopHubs:                topHubs,
		TotalClaims:            totalClaims,
		SupportedClaims:        supportedClaims,
		Contradicted:           activeContradictions,
		StaticClaims:           staticClaims,
		VerificationAttempted:  runtimeTotal > 0,
		UnverifiedClaims:       unverifiedClaims,
		NeedsAttention:         needsAttention,
		RecentEvidence:         recentEvidence,
		CanonicalDecisions:     canonicalDecisions,
		QuarantinedCount:       activeContradictions,
		LatestBlockHeight:      latestBlock,
		LatestMerkleHash:       latestHash,
		ParentMerkleHash:       parentHash,
		TrustStatus:            trustStatus,
		LastUpdated:            time.Now().UTC().Format(time.RFC3339),
		PendingDecisions:       pendingDecisions,
		IdempotencySafeguards:  idempotencySafeguards,
		TokensSavedMetric:      tokensSavedMetric,
		CostSavedUSDMetric:     costSavedMetric,
		ColdStartLatencyMetric: coldStartMetric,
		ActiveAgentsMetric:     activeAgentsMetric,
		LanguagesBreakdown:     languagesBreakdown,
		Drift:                  drift,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleDashboardSearch(w http.ResponseWriter, r *http.Request) {
	applySecurityHeaders(w)
	ctx := r.Context()
	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "search unavailable", nil)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SearchResponse{Query: "", Results: []SearchResult{}})
		return
	}

	scope, err := workspaceScopeFromRequest(ctx, r, pgStore)
	if err != nil {
		writeWorkspaceNotFound(w, r, strings.TrimSpace(r.URL.Query().Get("workspace")))
		return
	}
	workspaceID := scope.WorkspaceID

	limit := normalizeLimit(r.URL.Query().Get("limit"), 50, 100)

	searchPattern := "%" + query + "%"
	rows, err := pgStore.Pool().Query(
		ctx,
		`
		SELECT e.id, e.name, e.kind, e.package, e.file_path, e.is_exported,
		       COALESCE(r.name, 'unknown')
		FROM entities e
		LEFT JOIN repositories r ON r.id = e.repository_id
		WHERE e.workspace_id = $1
		  AND (e.name ILIKE $2 OR e.package ILIKE $2 OR e.file_path ILIKE $2 OR e.kind ILIKE $2)
		ORDER BY (e.kind != 'external') DESC, e.is_exported DESC, e.name
		LIMIT $3
		`,
		workspaceID, searchPattern, limit,
	)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "search query failed", map[string]any{"detail": err.Error()})
		return
	}
	defer rows.Close()

	results := make([]SearchResult, 0)
	for rows.Next() {
		var res SearchResult
		if err := rows.Scan(&res.ID, &res.Name, &res.Kind, &res.Package, &res.File, &res.Exported, &res.Repo); err == nil {
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
		writeJSONError(w, http.StatusServiceUnavailable, "store unavailable", nil)
		return
	}

	scope, err := workspaceScopeFromRequest(ctx, r, pgStore)
	if err != nil {
		writeWorkspaceNotFound(w, r, strings.TrimSpace(r.URL.Query().Get("workspace")))
		return
	}
	workspaceID := scope.WorkspaceID
	workspaceName := scope.WorkspaceName

	level := r.URL.Query().Get("level")
	if level == "" {
		level = "repository"
	}
	focus := strings.TrimSpace(r.URL.Query().Get("focus"))

	entityMap := make(map[string]EntityRecord)
	// External dependency stubs (kind='external') ARE loaded here.
	// They are the endpoints of every cross-repo and library
	// dependency edge. Filtering them out of entityMap caused every
	// claim whose source or target resolved to a stub to be dropped
	// silently — the graph rendered nodes without edges. The
	// node-building blocks below decide what to render; the entity
	// level filters zero-degree entities when no focus is set.
	const maxGraphEntities = 50000
	eRows, err := pgStore.Pool().Query(ctx, `
        SELECT e.id::text, e.name, e.kind, e.package, e.file_path, e.is_exported,
               COALESCE(r.name, 'unknown')
        FROM entities e
        LEFT JOIN repositories r ON r.id = e.repository_id
        WHERE e.workspace_id = $1
        LIMIT $2
    `, workspaceID, maxGraphEntities)
	if err == nil {
		defer eRows.Close()
		for eRows.Next() {
			var e EntityRecord
			if err := eRows.Scan(&e.ID, &e.Name, &e.Kind, &e.Package, &e.File, &e.Exported, &e.Repo); err == nil {
				entityMap[e.ID] = e
			}
		}
		if len(entityMap) >= maxGraphEntities {
			slog.Warn("graph entity cap reached", "workspace", workspaceName, "cap", maxGraphEntities)
		}
	}

	type rawClaim struct{ from, to string }
	var claims []rawClaim
	const maxGraphClaims = 200000
	cRows, err := pgStore.Pool().Query(ctx, `
        SELECT from_entity_id::text, to_entity_id::text
        FROM claims
        WHERE workspace_id = $1
        LIMIT $2
    `, workspaceID, maxGraphClaims)
	if err == nil {
		defer cRows.Close()
		for cRows.Next() {
			var c rawClaim
			if err := cRows.Scan(&c.from, &c.to); err == nil {
				claims = append(claims, c)
			}
		}
		if len(claims) >= maxGraphClaims {
			slog.Warn("graph claim cap reached", "workspace", workspaceName, "cap", maxGraphClaims)
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

	// Cross-repo edges detected by package/import analysis.
	// Grouped in SQL to aggregate link density per pair.
	type rawCrossRepo struct {
		fromRepo, toRepo string
		count            int
	}
	var crossRepoPairs []rawCrossRepo
	crRows, err := pgStore.Pool().Query(ctx, `
        SELECT r1.name AS from_repo, r2.name AS to_repo, COUNT(*)::int AS weight
        FROM cross_repo_edges cre
        JOIN repositories r1 ON r1.id = cre.from_repo_id
        JOIN repositories r2 ON r2.id = cre.to_repo_id
        WHERE cre.workspace_id = $1
        GROUP BY r1.name, r2.name
    `, workspaceID)
	if err == nil {
		defer crRows.Close()
		for crRows.Next() {
			var p rawCrossRepo
			if err := crRows.Scan(&p.fromRepo, &p.toRepo, &p.count); err == nil {
				crossRepoPairs = append(crossRepoPairs, p)
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

		// Merge cross_repo_edges and ensure endpoints exist in nodesMap
		for _, p := range crossRepoPairs {
			if p.fromRepo == "" || p.toRepo == "" || p.fromRepo == p.toRepo || p.fromRepo == "stdlib" || p.toRepo == "stdlib" {
				continue
			}

			if _, exists := nodesMap[p.fromRepo]; !exists {
				nodesMap[p.fromRepo] = GraphNodeDTO{
					ID:       p.fromRepo,
					Label:    p.fromRepo,
					Kind:     "repository",
					Repo:     p.fromRepo,
					Status:   "SUPPORTED",
					Exported: true,
					Count:    1,
				}
			}
			if _, exists := nodesMap[p.toRepo]; !exists {
				nodesMap[p.toRepo] = GraphNodeDTO{
					ID:       p.toRepo,
					Label:    p.toRepo,
					Kind:     "repository",
					Repo:     p.toRepo,
					Status:   "SUPPORTED",
					Exported: true,
					Count:    1,
				}
			}

			edgeKey := p.fromRepo + "->" + p.toRepo
			edge := edgesMap[edgeKey]
			if edge.Source == "" {
				edge = GraphEdgeDTO{
					ID:         edgeKey,
					Source:     p.fromRepo,
					Target:     p.toRepo,
					Type:       "CROSS_REPO_BRIDGE",
					Status:     "SUPPORTED",
					Label:      "depends on",
					Count:      0,
					Confidence: 1.0,
				}
			}
			edge.Count += p.count
			edgesMap[edgeKey] = edge
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
		// Compute degree — how many claims each entity participates in
		// as source or target. The no-focus view ("Top architectural
		// symbols") uses this to select the top N by connectivity.
		// Without it, the cap falls through to alphabetical order and
		// keeps 200 unconnected stubs (../../vendor/...), producing a
		// scattered cloud with no visible edges.
		degree := make(map[string]int)
		for _, c := range claims {
			degree[c.from]++
			degree[c.to]++
		}

		for _, e := range entityMap {
			if focus != "" && e.Package != focus {
				continue
			}
			// No focus = "top architectural symbols." An entity with
			// zero claims is not architectural; it is a standalone
			// type, an external stub, or an unused symbol. Filter it
			// out so the cap keeps real neighborhood hubs.
			if focus == "" && degree[e.ID] == 0 {
				continue
			}
			nodesMap[e.ID] = GraphNodeDTO{
				ID:       e.ID,
				Label:    e.Name,
				Kind:     e.Kind,
				Repo:     e.Repo,
				Package:  e.Package,
				Exported: e.Exported,
				Status:   "SUPPORTED",
				Count:    degree[e.ID],
			}
		}
		for _, c := range claims {
			src := entityMap[c.from]
			tgt := entityMap[c.to]
			if focus == "" || src.Package == focus || tgt.Package == focus {
				if _, exists := nodesMap[src.ID]; !exists {
					nodesMap[src.ID] = GraphNodeDTO{
						ID: src.ID, Label: src.Name, Kind: src.Kind,
						Package: src.Package, Repo: src.Repo,
						Status: "SUPPORTED", Count: degree[src.ID],
					}
				}
				if _, exists := nodesMap[tgt.ID]; !exists {
					nodesMap[tgt.ID] = GraphNodeDTO{
						ID: tgt.ID, Label: tgt.Name, Kind: tgt.Kind,
						Package: tgt.Package, Repo: tgt.Repo,
						Status: "SUPPORTED", Count: degree[tgt.ID],
					}
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

	// Cap the number of nodes returned per level. At 1,567 packages the
	// browser cannot render a useful graph — the layout becomes a
	// hairball and the simulation never settles. Sorting by Count and
	// keeping the top N preserves the highest-centrality nodes, which
	// is what the user is actually looking for. The response is marked
	// truncated so the UI can say "showing 200 of 1,567."
	const maxNodesPerGraph = 200

	var allNodes []GraphNodeDTO
	for _, n := range nodesMap {
		allNodes = append(allNodes, n)
	}
	totalNodes := len(allNodes)

	truncated := false
	if len(allNodes) > maxNodesPerGraph {
		sort.Slice(allNodes, func(i, j int) bool {
			if allNodes[i].Count != allNodes[j].Count {
				return allNodes[i].Count > allNodes[j].Count
			}
			return allNodes[i].Label < allNodes[j].Label
		})
		allNodes = allNodes[:maxNodesPerGraph]
		truncated = true
	}

	// Drop edges whose source or target was cut. Otherwise the graph
	// renders edges that point at nodes the client never received, and
	// the layout produces visible fragments.
	kept := make(map[string]bool, len(allNodes))
	for _, n := range allNodes {
		kept[n.ID] = true
	}
	var edges []GraphEdgeDTO
	for _, e := range edgesMap {
		if kept[e.Source] && kept[e.Target] {
			edges = append(edges, e)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(GraphResponseDTO{
		Level:      level,
		Focus:      focus,
		Nodes:      allNodes,
		Edges:      edges,
		Truncated:  truncated,
		TotalNodes: totalNodes,
	})
}

// HandleDashboardPolicies — returns policy enforcement state for the workspace.
func (s *Server) HandleDashboardPolicies(w http.ResponseWriter, r *http.Request) {
	applySecurityHeaders(w)
	ctx := r.Context()

	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "store unavailable", nil)
		return
	}

	scope, err := workspaceScopeFromRequest(ctx, r, pgStore)
	if err != nil {
		writeWorkspaceNotFound(w, r, strings.TrimSpace(r.URL.Query().Get("workspace")))
		return
	}
	tenantID := scope.TenantID
	workspaceID := scope.WorkspaceID
	workspaceName := scope.WorkspaceName

	resp := PolicyEnforcementResponse{
		Workspace:         workspaceName,
		LatestEvaluations: []PolicyEvaluationDTO{},
		FinalDecision:     "ALLOW",
	}

	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int FROM policies WHERE tenant_id = $1 AND status = 'active'
	`, tenantID).Scan(&resp.ActivePolicies)

	_ = pgStore.Pool().QueryRow(ctx, `
		SELECT
			COUNT(*)::int,
			COUNT(*) FILTER (WHERE decision = 'BLOCK')::int,
			COUNT(*) FILTER (WHERE decision = 'REVIEW')::int,
			COUNT(*) FILTER (WHERE decision = 'WARN')::int,
			COUNT(*) FILTER (WHERE decision = 'ALLOW')::int,
			MAX(merkle_block_height)
		FROM policy_evaluations
		WHERE workspace_id = $1
	`, workspaceID).Scan(
		&resp.TotalEvaluations,
		&resp.BlockedCount,
		&resp.ReviewCount,
		&resp.WarnCount,
		&resp.AllowCount,
		&resp.LatestMerkleAnchor,
	)

	switch {
	case resp.BlockedCount > 0:
		resp.FinalDecision = "BLOCK"
	case resp.ReviewCount > 0:
		resp.FinalDecision = "REVIEW"
	case resp.WarnCount > 0:
		resp.FinalDecision = "WARN"
	default:
		resp.FinalDecision = "ALLOW"
	}

	// Fetch proof bytes alongside the row so we can evaluate AnchorValid
	// honestly (not just "proof column is non-null").
	rows, err := pgStore.Pool().Query(ctx, `
		SELECT
			pe.id,
			pe.policy_id,
			COALESCE(p.metadata->>'title', p.statement, 'unnamed policy'),
			pe.decision,
			pe.reason,
			COALESCE(jsonb_array_length(pe.evidence->'entity_ids'), 0),
			COALESCE(jsonb_array_length(pe.evidence->'claim_ids'), 0),
			COALESCE(jsonb_array_length(pe.evidence->'contradiction_ids'), 0),
			pe.evidence->'matched_predicates',
			pe.merkle_block_height,
			pe.merkle_proof,
			pe.evaluated_at,
			pe.actor,
			pe.subject_kind,
			pe.subject_id
		FROM policy_evaluations pe
		LEFT JOIN policies p ON p.id = pe.policy_id
		WHERE pe.workspace_id = $1
		ORDER BY pe.evaluated_at DESC
		LIMIT 10
	`, workspaceID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var d PolicyEvaluationDTO
			var matchedJSON []byte
			var proofBytes []byte
			var evalTime time.Time
			if err := rows.Scan(
				&d.ID, &d.PolicyID, &d.PolicyTitle, &d.Decision, &d.Reason,
				&d.EntityCount, &d.ClaimCount, &d.ContradictionCount,
				&matchedJSON, &d.MerkleBlockHeight, &proofBytes,
				&evalTime, &d.Actor, &d.SubjectKind, &d.SubjectID,
			); err != nil {
				continue
			}
			if len(matchedJSON) > 0 {
				_ = json.Unmarshal(matchedJSON, &d.MatchedPredicates)
			}
			d.AnchorValid = proofIsInternallyValid(proofBytes)
			d.EvaluatedAt = evalTime.Format(time.RFC3339)
			resp.LatestEvaluations = append(resp.LatestEvaluations, d)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// HandleDashboardPolicyVerify — verifies the Merkle anchor of one evaluation.
//
// Rewritten to handle both v1 (RFC 6962 canonical) and v0 (linear chain)
// proof formats. The previous implementation tried to unmarshal every
// proof as v0 and returned "malformed proof" for every well-formed v1
// proof, which is what every current evaluation produces. The verify
// button was effectively broken for all live data.
//
// The endpoint is also workspace-scoped: a caller must specify which
// workspace the evaluation belongs to, and the query filters by it.
func (s *Server) HandleDashboardPolicyVerify(w http.ResponseWriter, r *http.Request) {
	applySecurityHeaders(w)
	ctx := r.Context()
	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "store unavailable", nil)
		return
	}

	workspaceName := strings.TrimSpace(r.URL.Query().Get("workspace"))
	if workspaceName == "" {
		writeJSONError(w, http.StatusBadRequest, "workspace parameter required", nil)
		return
	}
	scope, err := workspaceScopeFromRequest(ctx, r, pgStore)
	if err != nil {
		writeWorkspaceNotFound(w, r, workspaceName)
		return
	}
	workspaceID := scope.WorkspaceID

	evalIDStr := r.URL.Query().Get("id")
	evalID, err := uuid.Parse(evalIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}

	var decision, reason string
	var blockHeight *int64
	var proof []byte
	err = pgStore.Pool().QueryRow(ctx, `
		SELECT decision, reason, merkle_block_height, merkle_proof
		FROM policy_evaluations
		WHERE id = $1 AND workspace_id = $2
	`, evalID, workspaceID).Scan(&decision, &reason, &blockHeight, &proof)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "evaluation not found", nil)
		return
	}

	// Try v1 first — the format every new evaluation produces.
	if v1, err := store.UnmarshalV1Proof(proof); err == nil {
		leafBytes, decodeErr := hex.DecodeString(v1.LeafHash)
		if decodeErr != nil {
			writeJSONError(w, http.StatusOK, "invalid leaf hash", map[string]any{"valid": false})
			return
		}
		ok, verifyErr := store.VerifyV1Proof(leafBytes, v1)
		if verifyErr != nil {
			writeJSONError(w, http.StatusOK, verifyErr.Error(), map[string]any{"valid": false})
			return
		}
		valid := ok && blockHeight != nil && v1.EpochHeight == *blockHeight
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"valid":        valid,
			"version":      1,
			"block_height": v1.EpochHeight,
			"decision":     decision,
			"epoch_root":   v1.EpochRoot,
			"parent_root":  v1.ParentRoot,
		})
		return
	}

	// Fall back to v0 verification for historical rows.
	var proofData struct {
		DecisionHash string `json:"decision_hash"`
		PrevRoot     string `json:"prev_root"`
		NewRoot      string `json:"new_root"`
		BlockHeight  int64  `json:"block_height"`
	}
	if err := json.Unmarshal(proof, &proofData); err != nil {
		writeJSONError(w, http.StatusOK, "malformed proof", map[string]any{"valid": false})
		return
	}

	valid := blockHeight != nil && proofData.BlockHeight == *blockHeight && proofData.NewRoot != ""
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"valid":        valid,
		"version":      0,
		"block_height": proofData.BlockHeight,
		"decision":     decision,
		"prev_root":    proofData.PrevRoot,
		"new_root":     proofData.NewRoot,
	})
}

const maxSSEConnections = 1000

func (s *Server) HandleLiveEvents(w http.ResponseWriter, r *http.Request) {
	applySecurityHeaders(w)

	if s.sseCount != nil {
		select {
		case s.sseCount <- struct{}{}:
			defer func() { <-s.sseCount }()
		default:
			w.Header().Set("Retry-After", "30")
			http.Error(w, `{"error":"too many concurrent streams"}`, http.StatusServiceUnavailable)
			return
		}
	}

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
