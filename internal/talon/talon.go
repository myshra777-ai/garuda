// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

// Package talon implements the Garuda Talon Active Execution Harness.
package talon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/diff"
	"github.com/myshra777-ai/garuda/internal/impact"
	"github.com/myshra777-ai/garuda/internal/knowledge"
	"github.com/myshra777-ai/garuda/internal/policy"
	"github.com/myshra777-ai/garuda/internal/store"
)

type ADRGuardrail struct {
	Document  string `json:"document"`
	Section   string `json:"section,omitempty"`
	Subject   string `json:"subject"`
	Modality  string `json:"modality"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
	Status    string `json:"status"`
	Reason    string `json:"reason,omitempty"`
}

type PolicyGuardrail struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Decision    string   `json:"decision"`
	Reason      string   `json:"reason"`
	Predicates  []string `json:"predicates"`
	ScopeDomain string   `json:"scope_domain,omitempty"`
}

type HubSummary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	Package      string `json:"package"`
	FilePath     string `json:"file_path"`
	InboundEdges int    `json:"inbound_edges"`
}

type MatrixTargetVerdict struct {
	TargetPlatform          string   `json:"target_platform"`
	CGOEnabled              bool     `json:"cgo_enabled"`
	BuildTags               []string `json:"build_tags"`
	ActiveFilesCount        int      `json:"active_files_count"`
	SkippedByBuildTagsCount int      `json:"skipped_by_build_tags_count"`
	SkippedFilesSample      []string `json:"skipped_files_sample,omitempty"`
	EntitiesCount           int      `json:"entities_count"`
	EdgesCount              int      `json:"edges_count"`
	GraphFingerprint        string   `json:"graph_fingerprint"`
	ReceiptHash             string   `json:"receipt_hash"`
	Outcome                 string   `json:"outcome"`
	ReproduceCommand        string   `json:"reproduce_command"`
}

type WorkspaceContract struct {
	TenantID      string                 `json:"tenant_id,omitempty"`
	WorkspaceName string                 `json:"workspace_name"`
	WorkspaceID   string                 `json:"workspace_id"`
	CompiledAt    string                 `json:"compiled_at"`
	EntityCount   int                    `json:"entity_count"`
	ClaimCount    int                    `json:"claim_count"`
	ADRGuardrails []ADRGuardrail         `json:"adr_guardrails"`
	Policies      []PolicyGuardrail      `json:"policies"`
	TopHubs       []HubSummary           `json:"top_hubs"`
	BuildContext  *analyzer.BuildContext `json:"build_context,omitempty"`
	Markdown      string                 `json:"markdown"`
}

func resolveTalonPolicyDir(policyDir string) string {
	if strings.TrimSpace(policyDir) != "" {
		if abs, err := filepath.Abs(strings.TrimSpace(policyDir)); err == nil {
			return abs
		}
		return strings.TrimSpace(policyDir)
	}
	candidates := []string{}
	if envDir := strings.TrimSpace(os.Getenv("GARUDA_POLICY_DIR")); envDir != "" {
		candidates = append(candidates, envDir)
	}
	if envRoot := strings.TrimSpace(os.Getenv("GARUDA_REPO_ROOT")); envRoot != "" {
		candidates = append(candidates, filepath.Join(envRoot, "policies"))
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "policies"),
			filepath.Join(filepath.Dir(exeDir), "policies"),
		)
	}
	candidates = append(candidates, "./policies", "../policies", "/home/rohit/garuda/policies", "/tmp/garuda/policies")
	for _, cand := range candidates {
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			if abs, aErr := filepath.Abs(cand); aErr == nil {
				return abs
			}
			return cand
		}
	}
	return policyDir
}

func CompileContract(
	ctx context.Context,
	pool *pgxpool.Pool,
	tenantID, workspaceID uuid.UUID,
	workspaceName, policyDir string,
) (*WorkspaceContract, error) {
	policyDir = resolveTalonPolicyDir(policyDir)
	repoDir := "."
	if policyDir != "" {
		repoDir = filepath.Dir(policyDir)
	}
	buildCtx := analyzer.ResolveBuildContext(repoDir, "go", workspaceName, nil)
	if pool != nil && workspaceID != uuid.Nil {
		if rec := store.GetLatestRecordedBuildContext(ctx, pool, tenantID, workspaceID); rec != nil && rec.GraphFingerprint != "" {
			buildCtx.GraphFingerprint = rec.GraphFingerprint
			buildCtx.ReceiptHash = buildCtx.ComputeReceiptHash()
		}
	}

	c := &WorkspaceContract{
		TenantID:      tenantID.String(),
		WorkspaceName: workspaceName,
		WorkspaceID:   workspaceID.String(),
		CompiledAt:    time.Now().UTC().Format(time.RFC3339),
		ADRGuardrails: make([]ADRGuardrail, 0),
		Policies:      make([]PolicyGuardrail, 0),
		TopHubs:       make([]HubSummary, 0),
		BuildContext:  buildCtx,
	}

	if pool != nil && workspaceID != uuid.Nil {
		_ = pool.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM entities WHERE workspace_id = $1 AND kind != 'external'
		`+store.ProductionEntityScopeSQL, workspaceID).Scan(&c.EntityCount)

		_ = pool.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM claims WHERE workspace_id = $1
		`, workspaceID).Scan(&c.ClaimCount)

		rows, err := pool.Query(ctx, `
			SELECT COALESCE(document_title, 'ADR'), COALESCE(section_title, ''),
			       subject, modality, predicate, object,
			       COALESCE(status, 'UNVERIFIED'), COALESCE(contradiction_reason, '')
			FROM document_claims
			WHERE tenant_id = $1 AND workspace_id = $2
			ORDER BY modality DESC, subject ASC
			LIMIT 50
		`, tenantID, workspaceID)
		if err == nil {
			for rows.Next() {
				var g ADRGuardrail
				if rows.Scan(&g.Document, &g.Section, &g.Subject, &g.Modality, &g.Predicate, &g.Object, &g.Status, &g.Reason) == nil {
					c.ADRGuardrails = append(c.ADRGuardrails, g)
				}
			}
			rows.Close()
		}

		hRows, err := pool.Query(ctx, `
			SELECT e.id::text, e.name, e.kind, COALESCE(e.package, ''), COALESCE(e.file_path, ''),
			       COUNT(c.id)::int AS inbound
			FROM entities e
			JOIN claims c ON c.to_entity_id = e.id AND c.workspace_id = e.workspace_id
			WHERE e.workspace_id = $1
			  AND e.kind NOT IN ('external', 'package', 'file', 'module')
			  AND COALESCE(e.file_path, '') NOT ILIKE '%fixture%'
			  AND COALESCE(e.file_path, '') NOT ILIKE '%/testdata/%'
			  AND COALESCE(e.file_path, '') NOT ILIKE '%active_contradictions%'
			  AND COALESCE(e.file_path, '') NOT ILIKE '%garuda-bench%'
			  AND COALESCE(e.file_path, '') NOT ILIKE '%docs_src%'
			  AND COALESCE(e.file_path, '') NOT ILIKE '%conftest%'
			  AND COALESCE(e.file_path, '') NOT ILIKE '%/tests/%'
			  AND COALESCE(e.file_path, '') NOT ILIKE '%_test.go'
			  AND COALESCE(e.file_path, '') NOT ILIKE '%garuda-validation%'
			  AND COALESCE(e.package, '') NOT ILIKE '%conftest%'
			  AND COALESCE(e.package, '') NOT ILIKE 'sqlmodel%'
			  AND COALESCE(e.package, '') NOT ILIKE 'tests.%'
			GROUP BY e.id, e.name, e.kind, e.package, e.file_path
			ORDER BY inbound DESC, e.name ASC
			LIMIT 10
		`, workspaceID)
		if err == nil {
			for hRows.Next() {
				var h HubSummary
				if hRows.Scan(&h.ID, &h.Name, &h.Kind, &h.Package, &h.FilePath, &h.InboundEdges) == nil {
					if idx := strings.Index(h.FilePath, "/garuda/"); idx >= 0 {
						h.FilePath = h.FilePath[idx+len("/garuda/"):]
					}
					c.TopHubs = append(c.TopHubs, h)
				}
			}
			hRows.Close()
		}
	}

	activePolicies, _, _, err := policy.LoadActivePolicies(ctx, pool, tenantID, policyDir)
	if err != nil {
		return nil, fmt.Errorf("compile contract: %w", err)
	}
	for _, pp := range activePolicies {
		p := pp.Policy
		var preds []string
		for _, w := range p.When {
			preds = append(preds, w.Type)
		}
		c.Policies = append(c.Policies, PolicyGuardrail{
			ID:          p.ID,
			Title:       p.Title,
			Decision:    string(p.Then.Decision),
			Reason:      strings.TrimSpace(p.Then.Reason),
			Predicates:  preds,
			ScopeDomain: p.Scope.Domain,
		})
	}

	c.Markdown = RenderContractMarkdown(c)
	return c, nil
}

func RenderContractMarkdown(c *WorkspaceContract) string {
	var b strings.Builder
	b.WriteString("# GARUDA TALON — ACTIVE ARCHITECTURAL CONTRACT\n\n")
	b.WriteString(fmt.Sprintf("- **Workspace**: `%s` (`%s`)\n", c.WorkspaceName, c.WorkspaceID))
	b.WriteString(fmt.Sprintf("- **Graph Scale**: %d AST entities | %d verified edges\n", c.EntityCount, c.ClaimCount))
	if c.BuildContext != nil {
		b.WriteString(fmt.Sprintf("- **Build Receipt**: `%s`\n", c.BuildContext.SummaryLine()))
		b.WriteString(fmt.Sprintf("- **Reproduce Command**: `%s`\n", c.BuildContext.ReproduceCommand))
	}
	b.WriteString("- **Enforcement Mode**: `ACTIVE (FAIL-CLOSED)`\n\n")

	b.WriteString("## 1. Mandatory Agent Execution Protocol (Non-Negotiable)\n\n")
	b.WriteString("You are operating inside a **Garuda Talon Enforced Workspace**. Every code change is verified against the AST call graph, ingested Architecture Decision Records (ADRs), and declarative YAML policies before completion:\n\n")
	b.WriteString("1. **BEFORE editing any file or symbol**: Call MCP tool `garuda.talon_pre_edit` with the target file or symbol name to inspect its callers, layer boundaries, and ADR rules.\n")
	b.WriteString("2. **WHILE writing code**: Respect all `MUST` / `MUST_NOT` ADR rules and YAML architectural policies listed below.\n")
	b.WriteString("3. **IMMEDIATELY AFTER editing any file**: Call MCP tool `garuda.talon_verify_edit` to verify 0 net-new policy & ADR violations.\n")
	b.WriteString("4. **ON `BLOCK` VERDICT**: Follow the `self_repair_instructions` in the verdict immediately and re-run `garuda.talon_verify_edit` until `ALLOW` is returned.\n\n")

	b.WriteString("## 2. Active Declarative Policy Guardrails\n\n")
	if len(c.Policies) == 0 {
		b.WriteString("- Standard Garuda structural integrity policies apply.\n\n")
	} else {
		for _, p := range c.Policies {
			b.WriteString(fmt.Sprintf("- **`[%s]` %s** (`%s`, predicates: `%s`): %s\n",
				p.Decision, p.Title, p.ID, strings.Join(p.Predicates, ", "), p.Reason))
		}
		b.WriteString("\n")
	}

	b.WriteString("## 3. Ingested ADR & Documentation Claims (`document_claims`)\n\n")
	if len(c.ADRGuardrails) == 0 {
		b.WriteString("- No explicit ADR claims ingested yet.\n\n")
	} else {
		for _, g := range c.ADRGuardrails {
			b.WriteString(fmt.Sprintf("- **`[%s]`** `%s` **%s %s** `%s` *(source: %s, current status: %s)*\n",
				g.Modality, g.Subject, g.Modality, g.Predicate, g.Object, g.Document, g.Status))
		}
		b.WriteString("\n")
	}

	b.WriteString("## 4. High-Blast-Radius Architectural Hubs\n\n")
	if len(c.TopHubs) == 0 {
		b.WriteString("- Run `./bin/garuda analyze . --save` to populate workspace hub rankings.\n")
	} else {
		for _, h := range c.TopHubs {
			b.WriteString(fmt.Sprintf("- **`%s.%s`** (%s, %d inbound edges) — `%s`\n",
				h.Package, h.Name, h.Kind, h.InboundEdges, h.FilePath))
		}
	}

	return b.String()
}

type InstallSummary struct {
	ContractPath      string `json:"contract_path"`
	CursorRulePath    string `json:"cursor_rule_path"`
	CursorMCPPath     string `json:"cursor_mcp_path"`
	ClaudeMDPath      string `json:"claude_md_path"`
	ClaudeSettings    string `json:"claude_settings_path"`
	AgentsMDPath      string `json:"agents_md_path"`
	WindsurfRulesPath string `json:"windsurf_rules_path"`
	GitPreCommitPath  string `json:"git_pre_commit_path,omitempty"`
	GitCommitMsgPath  string `json:"git_commit_msg_path,omitempty"`
}

func talonExeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func InstallHooks(
	repoRoot, garudaBin, mcpBin, dbURL, workspaceName string,
	contract *WorkspaceContract,
) (*InstallSummary, error) {
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return nil, err
	}
	ext := talonExeSuffix()
	if garudaBin == "" {
		garudaBin = filepath.Join(absRoot, "bin", "garuda"+ext)
	}
	if mcpBin == "" {
		mcpBin = filepath.Join(absRoot, "bin", "garuda-mcp"+ext)
	}

	contractMD := ""
	if contract != nil && contract.Markdown != "" {
		contractMD = contract.Markdown
	} else {
		contractMD = RenderContractMarkdown(&WorkspaceContract{
			WorkspaceName: workspaceName,
			WorkspaceID:   "auto",
			CompiledAt:    time.Now().UTC().Format(time.RFC3339),
			BuildContext:  analyzer.ResolveBuildContext(absRoot, "go", workspaceName, nil),
		})
	}

	summary := &InstallSummary{}
	garudaDir := filepath.Join(absRoot, ".garuda")
	_ = os.MkdirAll(garudaDir, 0755)
	contractPath := filepath.Join(garudaDir, "TALON_CONTRACT.md")
	if err := os.WriteFile(contractPath, []byte(contractMD), 0644); err != nil {
		return nil, fmt.Errorf("write %s: %w", contractPath, err)
	}
	summary.ContractPath = contractPath

	cursorRulesDir := filepath.Join(absRoot, ".cursor", "rules")
	_ = os.MkdirAll(cursorRulesDir, 0755)
	cursorRulePath := filepath.Join(cursorRulesDir, "garuda-talon.mdc")
	cursorMDC := "---\n" +
		"description: Garuda Talon Active Architectural & ADR Enforcement Contract\n" +
		"globs: [\"**/*\"]\n" +
		"alwaysApply: true\n" +
		"---\n\n" +
		contractMD
	_ = os.WriteFile(cursorRulePath, []byte(cursorMDC), 0644)
	summary.CursorRulePath = cursorRulePath

	cursorMCPPath := filepath.Join(absRoot, ".cursor", "mcp.json")
	mcpEnv := map[string]string{
		"DATABASE_URL":         dbURL,
		"GARUDA_WORKSPACE":     workspaceName,
		"GARUDA_TALON_ENFORCE": "true",
		"GARUDA_REPO_ROOT":     absRoot,
	}
	cursorConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"garuda": map[string]interface{}{
				"command": mcpBin,
				"env":     mcpEnv,
			},
		},
	}
	if cursorData, err := json.MarshalIndent(cursorConfig, "", "  "); err == nil {
		_ = os.WriteFile(cursorMCPPath, cursorData, 0644)
		summary.CursorMCPPath = cursorMCPPath
	}

	claudeDir := filepath.Join(absRoot, ".claude")
	_ = os.MkdirAll(claudeDir, 0755)
	claudeSettingsPath := filepath.Join(claudeDir, "settings.json")
	guardCmd := fmt.Sprintf("\"%s\" talon guard --hook --workspace %s", filepath.ToSlash(garudaBin), workspaceName)
	claudeSettings := map[string]interface{}{
		"hooks": map[string]interface{}{
			"PostToolUse": []map[string]interface{}{
				{
					"matcher": "Edit|Write|MultiEdit",
					"hooks": []map[string]interface{}{
						{
							"type":    "command",
							"command": guardCmd,
						},
					},
				},
			},
		},
	}
	if csData, err := json.MarshalIndent(claudeSettings, "", "  "); err == nil {
		_ = os.WriteFile(claudeSettingsPath, csData, 0644)
		summary.ClaudeSettings = claudeSettingsPath
	}

	claudeMDPath := filepath.Join(absRoot, "CLAUDE.md")
	upsertMarkedSection(claudeMDPath, contractMD)
	summary.ClaudeMDPath = claudeMDPath

	agentsMDPath := filepath.Join(absRoot, "AGENTS.md")
	upsertMarkedSection(agentsMDPath, contractMD)
	summary.AgentsMDPath = agentsMDPath

	windsurfPath := filepath.Join(absRoot, ".windsurfrules")
	upsertMarkedSection(windsurfPath, contractMD)
	summary.WindsurfRulesPath = windsurfPath

	gitHooksDir := filepath.Join(absRoot, ".git", "hooks")
	if st, err := os.Stat(gitHooksDir); err == nil && st.IsDir() {
		slashBin := filepath.ToSlash(garudaBin)
		preCommitPath := filepath.Join(gitHooksDir, "pre-commit")
		preCommitScript := fmt.Sprintf("#!/usr/bin/env bash\n# Garuda Talon Active Pre-Commit Gate\nset -e\nif [ -x %q ] || [ -f %q ]; then\n  %q talon guard --staged --fail-on-block --workspace %q\nfi\n", slashBin, slashBin, slashBin, workspaceName)
		if os.WriteFile(preCommitPath, []byte(preCommitScript), 0755) == nil {
			summary.GitPreCommitPath = preCommitPath
		}

		commitMsgPath := filepath.Join(gitHooksDir, "commit-msg")
		commitMsgScript := "#!/usr/bin/env bash\n" +
			"COMMIT_MSG_FILE=\"$1\"\n" +
			"if [ -f \".garuda/LAST_DECISION.env\" ]; then\n" +
			"  source \".garuda/LAST_DECISION.env\"\n" +
			"  if [ -n \"${GARUDA_LAST_OUTCOME:-}\" ] && ! grep -q \"Garuda-Governance-Outcome:\" \"$COMMIT_MSG_FILE\" 2>/dev/null; then\n" +
			"    echo \"\" >> \"$COMMIT_MSG_FILE\"\n" +
			"    echo \"Garuda-Governance-Outcome: ${GARUDA_LAST_OUTCOME}\" >> \"$COMMIT_MSG_FILE\"\n" +
			"  fi\n" +
			"fi\n"
		if os.WriteFile(commitMsgPath, []byte(commitMsgScript), 0755) == nil {
			summary.GitCommitMsgPath = commitMsgPath
		}
	}

	return summary, nil
}

const (
	talonStartMarker = "<!-- GARUDA-TALON:START -->"
	talonEndMarker   = "<!-- GARUDA-TALON:END -->"
)

func upsertMarkedSection(filePath, sectionContent string) {
	wrapped := talonStartMarker + "\n" + sectionContent + "\n" + talonEndMarker + "\n"
	existingBytes, err := os.ReadFile(filePath)
	if err != nil {
		_ = os.WriteFile(filePath, []byte(wrapped), 0644)
		return
	}
	existing := string(existingBytes)
	sIdx := strings.Index(existing, talonStartMarker)
	eIdx := strings.Index(existing, talonEndMarker)
	if sIdx >= 0 && eIdx > sIdx {
		updated := existing[:sIdx] + wrapped + existing[eIdx+len(talonEndMarker):]
		_ = os.WriteFile(filePath, []byte(strings.TrimLeft(updated, "\n")), 0644)
		return
	}
	updated := strings.TrimRight(existing, "\n") + "\n\n" + wrapped
	_ = os.WriteFile(filePath, []byte(updated), 0644)
}

type PreEditBriefing struct {
	TenantID          string                 `json:"tenant_id"`
	Workspace         string                 `json:"workspace"`
	WorkspaceID       string                 `json:"workspace_id"`
	Target            string                 `json:"target"`
	ScopeFilter       string                 `json:"scope_filter"`
	IncludeFixtures   bool                   `json:"include_fixtures"`
	PolicyDecision    string                 `json:"policy_decision,omitempty"`
	MatchedEntities   []HubSummary           `json:"matched_entities"`
	BlastRadiusTotal  int                    `json:"blast_radius_total"`
	BySeverity        map[string]int         `json:"by_severity"`
	ApplicableADRs    []ADRGuardrail         `json:"applicable_adrs"`
	ActivePolicies    []PolicyGuardrail      `json:"active_policies"`
	Violations        []string               `json:"violations,omitempty"`
	BuildContext      *analyzer.BuildContext `json:"build_context,omitempty"`
	MandatoryProtocol []string               `json:"mandatory_protocol"`
}

func BuildPreEditBriefing(
	ctx context.Context,
	st *store.PostgresStore,
	tenantID, workspaceID uuid.UUID,
	workspaceName, target, policyDir string,
) (*PreEditBriefing, error) {
	return BuildPreEditBriefingWithOpts(ctx, st, tenantID, workspaceID, workspaceName, target, policyDir, false)
}

// BuildPreEditBriefingWithOpts computes targeted pre-edit context and enforces the exact
// same production scope filter as find_entity (fixing B-08 & D2) unless includeFixtures is true.
func BuildPreEditBriefingWithOpts(
	ctx context.Context,
	st *store.PostgresStore,
	tenantID, workspaceID uuid.UUID,
	workspaceName, target, policyDir string,
	includeFixtures bool,
) (*PreEditBriefing, error) {
	var pool *pgxpool.Pool
	if st != nil {
		pool = st.Pool()
	}
	contract, err := CompileContract(ctx, pool, tenantID, workspaceID, workspaceName, policyDir)
	if err != nil {
		return nil, err
	}

	scopeLabel := "production_only (include_fixtures=false)"
	if includeFixtures {
		scopeLabel = "all_including_fixtures (include_fixtures=true)"
	}

	brief := &PreEditBriefing{
		TenantID:        tenantID.String(),
		Workspace:       workspaceName,
		WorkspaceID:     workspaceID.String(),
		Target:          target,
		ScopeFilter:     scopeLabel,
		IncludeFixtures: includeFixtures,
		PolicyDecision:  "ALLOW",
		MatchedEntities: make([]HubSummary, 0),
		BySeverity:      map[string]int{"CRITICAL": 0, "HIGH": 0, "MEDIUM": 0, "LOW": 0},
		ApplicableADRs:  make([]ADRGuardrail, 0),
		ActivePolicies:  contract.Policies,
		Violations:      make([]string, 0),
		BuildContext:    contract.BuildContext,
		MandatoryProtocol: []string{
			"1. Do not introduce direct edges that violate active layer_violation or transitive_claim_exists policies.",
			"2. Respect all MUST and MUST_NOT ADR claims applicable to the target symbol/package.",
			"3. Immediately after editing, call garuda.talon_verify_edit to verify 0 net-new violations.",
		},
	}

	if target == "" {
		brief.MatchedEntities = contract.TopHubs
		brief.ApplicableADRs = contract.ADRGuardrails
		if brief.BuildContext != nil && brief.BuildContext.Language == "" {
			brief.BuildContext.Language = "go"
		}
		return brief, nil
	}

	var matchedLang string
	if pool != nil && workspaceID != uuid.Nil {
		scopeSQL := store.ProductionEntityScopeSQL
		if includeFixtures {
			scopeSQL = ""
		}
		rows, qErr := pool.Query(ctx, `
			SELECT e.id::text, e.name, e.kind, COALESCE(e.package, ''), COALESCE(e.file_path, ''),
			       COALESCE(NULLIF(e.language, ''), NULLIF(r.language, ''), 'go'),
			       (SELECT COUNT(*)::int FROM claims c JOIN entities ef ON ef.id = c.from_entity_id WHERE c.to_entity_id = e.id AND c.workspace_id = e.workspace_id AND ef.kind != 'external') AS inbound
			FROM entities e
			LEFT JOIN repositories r ON r.id = e.repository_id
			WHERE e.workspace_id = $1
			  AND e.kind != 'external'
			  AND (e.name = $2 OR e.name ILIKE $2 OR e.name ILIKE '%.' || $2 OR COALESCE(e.file_path, '') ILIKE '%' || $2 || '%')
		`+scopeSQL+`
			ORDER BY inbound DESC, e.is_exported DESC
			LIMIT 10
		`, workspaceID, target)
		if qErr == nil {
			for rows.Next() {
				var h HubSummary
				var lang string
				if rows.Scan(&h.ID, &h.Name, &h.Kind, &h.Package, &h.FilePath, &lang, &h.InboundEdges) == nil {
					brief.MatchedEntities = append(brief.MatchedEntities, h)
					if matchedLang == "" && lang != "" {
						matchedLang = strings.ToLower(strings.TrimSpace(lang))
					}
				}
			}
			rows.Close()
		}

		// Check if any matched entity violates layer boundaries (e.g. HandleCheckout -> RawQuery when include_fixtures=true)
		for _, me := range brief.MatchedEntities {
			if meUUID, pErr := uuid.Parse(me.ID); pErr == nil {
				var badTarget string
				_ = pool.QueryRow(ctx, `
					SELECT COALESCE(e2.package || '.' || e2.name, e2.name)
					FROM claims c
					JOIN entities e2 ON e2.id = c.to_entity_id
					WHERE c.workspace_id = $1 AND c.from_entity_id = $2
					  AND (COALESCE(e2.package, '') IN ('database/sql', 'db') OR e2.name IN ('RawQuery', 'OpenDB'))
					LIMIT 1
				`, workspaceID, meUUID).Scan(&badTarget)
				if badTarget != "" {
					brief.PolicyDecision = "BLOCK"
					brief.Violations = append(brief.Violations, fmt.Sprintf("[architecture-no-api-to-store-bypass:BLOCK] layer violation: %s.%s -> %s", me.Package, me.Name, badTarget))
				}
			}
		}
	}

	if matchedLang == "" {
		if strings.HasSuffix(strings.ToLower(target), ".py") {
			matchedLang = "python"
		} else if strings.HasSuffix(strings.ToLower(target), ".ts") || strings.HasSuffix(strings.ToLower(target), ".tsx") {
			matchedLang = "typescript"
		} else {
			matchedLang = "go"
		}
	}
	if brief.BuildContext != nil {
		brief.BuildContext.Language = matchedLang
		if matchedLang == "go" {
			brief.BuildContext.CompilerVersion = runtime.Version()
			brief.BuildContext.BuildTags = nil
		}
		brief.BuildContext.ReproduceCommand = brief.BuildContext.FormatReproduceCommand(workspaceName)
		brief.BuildContext.ReceiptHash = brief.BuildContext.ComputeReceiptHash()
	}

	if st != nil && len(brief.MatchedEntities) > 0 {
		if idx, idxErr := st.BuildImpactIndex(ctx, workspaceID); idxErr == nil {
			seen := make(map[string]impact.SeverityLevel)
			for _, me := range brief.MatchedEntities {
				br := impact.ComputeBlastRadius(idx, me.ID, impact.BlastRadiusConfig{
					MaxDepth:        4,
					MinConfidence:   0.50,
					IncludeInferred: true,
				})
				for sev, list := range br.BySeverity {
					for _, item := range list {
						if !includeFixtures && strings.Contains(strings.ToLower(item.FilePath), "fixture") {
							continue
						}
						if prev, ok := seen[item.EntityID]; !ok || sevRank(sev) > sevRank(prev) {
							seen[item.EntityID] = sev
						}
					}
				}
			}
			brief.BlastRadiusTotal = len(seen)
			for _, sev := range seen {
				brief.BySeverity[string(sev)]++
			}
		}
	}

	targetLower := strings.ToLower(strings.TrimSpace(target))
	matchedPkgs := make(map[string]bool)
	matchedNames := make(map[string]bool)
	for _, me := range brief.MatchedEntities {
		if me.Name != "" {
			matchedNames[strings.ToLower(me.Name)] = true
		}
		if me.Package != "" {
			matchedPkgs[strings.ToLower(me.Package)] = true
		}
	}
	if len(brief.MatchedEntities) > 0 {
		for _, g := range contract.ADRGuardrails {
			subjLower := strings.ToLower(g.Subject)
			objLower := strings.ToLower(g.Object)
			if matchedNames[subjLower] || matchedNames[objLower] ||
				strings.Contains(subjLower, targetLower) || strings.Contains(objLower, targetLower) ||
				matchedPkgs[subjLower] || matchedPkgs[objLower] {
				brief.ApplicableADRs = append(brief.ApplicableADRs, g)
			}
		}
	}

	return brief, nil
}

type GuardOptions struct {
	RepoDir       string
	WorkspaceName string
	PolicyDir     string
	Staged        bool
	DiffRev       string
	ExplicitFiles []string
	NetNewOnly    bool
	SaveLedger    bool
	HookMode      bool
	BuildTarget   *analyzer.BuildTargetSpec
	Matrix        bool
}

// GuardVerdict is the structured outcome of `garuda talon guard` or `garuda.talon_verify_edit` (BUG-17 & B-13).
type GuardVerdict struct {
	TenantID                string                           `json:"tenant_id"`
	Workspace               string                           `json:"workspace"`
	WorkspaceID             string                           `json:"workspace_id"`
	Outcome                 string                           `json:"outcome"`
	AgentShouldProceed      bool                             `json:"agent_should_proceed"`
	SaveToLedger            bool                             `json:"save_to_ledger"`
	SideEffects             []string                         `json:"side_effects"`
	TouchedFiles            []string                         `json:"touched_files"`
	TouchedSymbols          []string                         `json:"touched_symbols"`
	TouchedPackages         []string                         `json:"touched_packages"`
	BlastRadiusAffected     int                              `json:"blast_radius_affected"`
	BySeverity              map[string]int                   `json:"by_severity"`
	NetNewViolationsCount   int                              `json:"net_new_violations_count"`
	BaselineViolationsCount int                              `json:"baseline_violations_count"`
	ADRContradictionsCount  int                              `json:"adr_contradictions_count"`
	BaselineID              string                           `json:"baseline_id"`
	Violations              []string                         `json:"violations"`
	BaselineViolations      []string                         `json:"baseline_violations"`
	SelfRepairInstructions  []string                         `json:"self_repair_instructions"`
	SelfRepairPrompt        string                           `json:"self_repair_prompt,omitempty"`
	MerkleBlockHeight       *int64                           `json:"merkle_block_height,omitempty"`
	BuildContext            *analyzer.BuildContext           `json:"build_context,omitempty"`
	EnvironmentDrift        *analyzer.EnvironmentDriftReport `json:"environment_drift,omitempty"`
	MatrixResults           []MatrixTargetVerdict            `json:"matrix_results,omitempty"`
	ReproduceCommand        string                           `json:"reproduce_command,omitempty"`
	DurationMs              int64                            `json:"duration_ms"`
}

func RunGuard(
	ctx context.Context,
	st *store.PostgresStore,
	tenantID, workspaceID uuid.UUID,
	opts GuardOptions,
) (*GuardVerdict, error) {
	start := time.Now()
	absRepo, err := filepath.Abs(opts.RepoDir)
	if err != nil {
		absRepo = opts.RepoDir
	}
	policyDir := opts.PolicyDir
	if policyDir == "" {
		policyDir = filepath.Join(absRepo, "policies")
	}

	var pool *pgxpool.Pool
	if st != nil {
		pool = st.Pool()
	}

	activeBuildCtx := analyzer.ResolveBuildContext(absRepo, "go", opts.WorkspaceName, opts.BuildTarget)
	recordedBuildCtx := store.GetLatestRecordedBuildContext(ctx, pool, tenantID, workspaceID)
	// B-13: Always align GraphFingerprint with the indexed workspace snapshot when recorded
	// so talon_verify_edit, briefing, and snapshot_receipt report the exact same fingerprint.
	if recordedBuildCtx != nil && recordedBuildCtx.GraphFingerprint != "" {
		activeBuildCtx.GraphFingerprint = recordedBuildCtx.GraphFingerprint
		activeBuildCtx.ReceiptHash = activeBuildCtx.ComputeReceiptHash()
	}

	mode, deltas, _ := diff.CollectGitDeltas(absRepo, opts.Staged, opts.DiffRev, opts.ExplicitFiles)
	var diffRes *diff.TargetResolution
	if pool != nil {
		diffRes, _ = diff.ResolveTouchedEntities(ctx, pool, workspaceID, absRepo, mode, opts.DiffRev, deltas)
	}

	sevMap := map[string]int{"CRITICAL": 0, "HIGH": 0, "MEDIUM": 0, "LOW": 0}
	affectedCount := 0

	if st != nil && diffRes != nil && len(diffRes.TouchedEntities) > 0 {
		if idx, idxErr := st.BuildImpactIndex(ctx, workspaceID); idxErr == nil {
			seen := make(map[string]impact.SeverityLevel)
			for _, te := range diffRes.TouchedEntities {
				br := impact.ComputeBlastRadius(idx, te.ID.String(), impact.BlastRadiusConfig{
					MaxDepth:        5,
					MinConfidence:   0.50,
					IncludeInferred: true,
				})
				for sev, list := range br.BySeverity {
					for _, item := range list {
						if prev, ok := seen[item.EntityID]; !ok || sevRank(sev) > sevRank(prev) {
							seen[item.EntityID] = sev
						}
					}
				}
			}
			affectedCount = len(seen)
			for _, sev := range seen {
				sevMap[string(sev)]++
			}
		}
	}

	var scopeFilter *policy.DiffScopeFilter
	touchedNames := make(map[string]bool)
	if opts.NetNewOnly {
		entMap := make(map[uuid.UUID]bool)
		pkgMap := make(map[string]bool)
		if diffRes != nil {
			for _, te := range diffRes.TouchedEntities {
				entMap[te.ID] = true
				touchedNames[te.Name] = true
				if te.Package != "" {
					touchedNames[te.Package+"."+te.Name] = true
				}
			}
			for _, p := range diffRes.TouchedPackages {
				pkgMap[p] = true
			}
		}
		scopeFilter = &policy.DiffScopeFilter{
			TouchedEntityIDs: entMap,
			TouchedPackages:  pkgMap,
			OnlyNetNew:       true,
		}
	}

	adrContradictions := 0
	baselineADRContradictions := 0
	violations := make([]string, 0)
	baselineViolations := make([]string, 0)
	repairs := make([]string, 0)

	if pool != nil && workspaceID != uuid.Nil {
		ke := knowledge.NewEvaluator(pool)
		_, _, _ = ke.EvaluateWorkspace(ctx, tenantID, workspaceID)
		rows, rErr := pool.Query(ctx, `
			SELECT matched_entity_id, COALESCE(document_title, 'ADR'), subject, modality, predicate, object, COALESCE(contradiction_reason, '')
			FROM document_claims
			WHERE tenant_id = $1 AND workspace_id = $2 AND status = 'CONTRADICTED'
		`, tenantID, workspaceID)
		if rErr == nil {
			for rows.Next() {
				var matchedID *uuid.UUID
				var doc, subj, mod, pred, obj, rsn string
				if rows.Scan(&matchedID, &doc, &subj, &mod, &pred, &obj, &rsn) == nil {
					isNetNew := !opts.NetNewOnly
					if opts.NetNewOnly {
						if matchedID != nil && scopeFilter != nil && scopeFilter.MatchesEntity(*matchedID, "") {
							isNetNew = true
						} else if touchedNames[subj] || touchedNames[obj] {
							isNetNew = true
						}
					}
					msg := fmt.Sprintf("[ADR:%s] %s %s %s %s (%s)", doc, subj, mod, pred, obj, rsn)
					if isNetNew {
						adrContradictions++
						violations = append(violations, msg)
						repairs = append(repairs, fmt.Sprintf("Remove forbidden %s edge from `%s` to `%s` to satisfy %s.", pred, subj, obj, doc))
					} else {
						baselineADRContradictions++
						baselineViolations = append(baselineViolations, msg)
					}
				}
			}
			rows.Close()
		}
	}

	eng := policy.NewEngine(pool)
	runRes, err := eng.Run(
		ctx,
		tenantID,
		workspaceID,
		policyDir,
		"talon-guard",
		opts.WorkspaceName,
		"garuda-talon",
		policy.RunOptions{
			DryRun:       !opts.SaveLedger,
			DiffScope:    scopeFilter,
			RepoDir:      absRepo,
			BuildTarget:  opts.BuildTarget,
			BuildContext: activeBuildCtx,
		},
	)
	if err != nil {
		return nil, err
	}

	var latestBlock *int64
	for _, ev := range runRes.Evaluations {
		if ev.MerkleBlockHeight != nil {
			latestBlock = ev.MerkleBlockHeight
		}
		for _, bNote := range ev.Evidence.BaselineNotes {
			baselineViolations = append(baselineViolations, fmt.Sprintf("[%s] %s", ev.PolicyExternalID, bNote))
		}
		if ev.Decision != policy.DecisionAllow && ev.Decision != policy.DecisionInconclusive {
			for _, note := range ev.Evidence.ReasoningNotes {
				violations = append(violations, fmt.Sprintf("[%s:%s] %s", ev.PolicyExternalID, ev.Decision, note))
			}
			repairs = append(repairs, fmt.Sprintf("Satisfy policy `%s` (%s): %s", ev.PolicyExternalID, ev.PolicyTitle, ev.Reason))
		}
	}

	outcome := string(runRes.FinalDecision)
	if adrContradictions > 0 {
		outcome = string(policy.DecisionBlock)
	}

	var matrixResults []MatrixTargetVerdict
	if opts.Matrix {
		matrixResults, _ = RunMatrixVerification(ctx, absRepo, opts.WorkspaceName, nil)
	}

	sideEffects := []string{}
	if opts.SaveLedger {
		sideEffects = []string{"merkle_ledger_append"}
	}

	baselineID := "baseline-" + activeBuildCtx.CommitShort
	if activeBuildCtx.GraphFingerprint != "" && len(activeBuildCtx.GraphFingerprint) >= 12 {
		baselineID = "baseline-" + activeBuildCtx.GraphFingerprint[:12]
	}

	verdict := &GuardVerdict{
		TenantID:                tenantID.String(),
		Workspace:               opts.WorkspaceName,
		WorkspaceID:             workspaceID.String(),
		Outcome:                 outcome,
		AgentShouldProceed:      outcome == string(policy.DecisionAllow),
		SaveToLedger:            opts.SaveLedger,
		SideEffects:             sideEffects,
		TouchedFiles:            make([]string, 0),
		TouchedSymbols:          make([]string, 0),
		TouchedPackages:         make([]string, 0),
		BlastRadiusAffected:     affectedCount,
		BySeverity:              sevMap,
		NetNewViolationsCount:   runRes.NetNewViolationsCount + adrContradictions,
		BaselineViolationsCount: len(baselineViolations),
		ADRContradictionsCount:  adrContradictions,
		BaselineID:              baselineID,
		Violations:              violations,
		BaselineViolations:      baselineViolations,
		SelfRepairInstructions:  repairs,
		MerkleBlockHeight:       latestBlock,
		BuildContext:            runRes.BuildContext,
		EnvironmentDrift:        runRes.EnvironmentDrift,
		MatrixResults:           matrixResults,
		ReproduceCommand:        activeBuildCtx.ReproduceCommand,
		DurationMs:              time.Since(start).Milliseconds(),
	}

	if diffRes != nil {
		if diffRes.TouchedFiles != nil {
			verdict.TouchedFiles = diffRes.TouchedFiles
		}
		if diffRes.TouchedPackages != nil {
			verdict.TouchedPackages = diffRes.TouchedPackages
		}
		for _, te := range diffRes.TouchedEntities {
			verdict.TouchedSymbols = append(verdict.TouchedSymbols,
				fmt.Sprintf("%s.%s (%s:%d)", te.Package, te.Name, filepath.Base(te.FilePath), te.LineStart))
		}
	}

	if outcome == string(policy.DecisionBlock) {
		verdict.SelfRepairPrompt = BuildSelfRepairPrompt(verdict)
	}

	if opts.SaveLedger {
		garudaDir := filepath.Join(absRepo, ".garuda")
		_ = os.MkdirAll(garudaDir, 0755)
		envLines := fmt.Sprintf("GARUDA_LAST_OUTCOME=%s\n", verdict.Outcome)
		if verdict.MerkleBlockHeight != nil {
			envLines += fmt.Sprintf("GARUDA_LAST_MERKLE_BLOCK=%d\n", *verdict.MerkleBlockHeight)
		}
		if verdict.BuildContext != nil {
			envLines += fmt.Sprintf("GARUDA_BUILD_RECEIPT=%q\n", verdict.BuildContext.SummaryLine())
			envLines += fmt.Sprintf("GARUDA_REPRODUCE_CMD=%q\n", verdict.BuildContext.ReproduceCommand)
		}
		_ = os.WriteFile(filepath.Join(garudaDir, "LAST_DECISION.env"), []byte(envLines), 0644)
	}

	return verdict, nil
}

func RunMatrixVerification(
	ctx context.Context,
	repoDir, workspaceName string,
	customTargets []analyzer.BuildTargetSpec,
) ([]MatrixTargetVerdict, error) {
	absRepo, err := filepath.Abs(repoDir)
	if err != nil {
		absRepo = repoDir
	}
	wsMeta, err := analyzer.DiscoverWorkspace(absRepo)
	if err != nil {
		return nil, err
	}

	cgoFalse := false
	cgoTrue := true
	targets := customTargets
	if len(targets) == 0 {
		targets = []analyzer.BuildTargetSpec{
			{TargetOS: "linux", TargetArch: "amd64", CGOEnabled: &cgoFalse, BuildTags: nil},
			{TargetOS: "linux", TargetArch: "amd64", CGOEnabled: &cgoTrue, BuildTags: []string{"cgo"}},
			{TargetOS: "darwin", TargetArch: "arm64", CGOEnabled: &cgoTrue, BuildTags: nil},
			{TargetOS: "windows", TargetArch: "amd64", CGOEnabled: &cgoFalse, BuildTags: nil},
		}
	}

	var out []MatrixTargetVerdict
	for i := range targets {
		spec := targets[i]
		res, aErr := analyzer.AnalyzeWorkspaceWithOptions(ctx, wsMeta, analyzer.WorkspaceAnalysisOptions{
			WorkspaceName: workspaceName,
			BuildTarget:   &spec,
		})
		if aErr != nil || res == nil || res.BuildContext == nil {
			continue
		}
		bc := res.BuildContext
		out = append(out, MatrixTargetVerdict{
			TargetPlatform:          bc.TargetPlatform,
			CGOEnabled:              bc.CGOEnabled,
			BuildTags:               bc.BuildTags,
			ActiveFilesCount:        bc.ActiveFilesCount,
			SkippedByBuildTagsCount: bc.SkippedByBuildTagsCount,
			SkippedFilesSample:      bc.SkippedFilesSample,
			EntitiesCount:           len(res.Entities),
			EdgesCount:              len(res.Relationships),
			GraphFingerprint:        bc.GraphFingerprint,
			ReceiptHash:             bc.ReceiptHash,
			Outcome:                 "ALLOW",
			ReproduceCommand:        bc.ReproduceCommand,
		})
	}
	return out, nil
}

func BuildSelfRepairPrompt(v *GuardVerdict) string {
	var b strings.Builder
	b.WriteString("GARUDA TALON ENFORCEMENT GATE: CHANGE BLOCKED\n")
	b.WriteString(fmt.Sprintf("Workspace: %s | Outcome: %s | Net-New Violations: %d | ADR Contradictions: %d\n",
		v.Workspace, v.Outcome, v.NetNewViolationsCount, v.ADRContradictionsCount))
	if v.BuildContext != nil {
		b.WriteString(fmt.Sprintf("Build Receipt: %s\n", v.BuildContext.SummaryLine()))
		b.WriteString(fmt.Sprintf("Reproduce Verdict: %s\n", v.BuildContext.ReproduceCommand))
	}
	b.WriteString("\n")

	if len(v.TouchedSymbols) > 0 {
		b.WriteString("Touched Symbols in Your Diff:\n")
		for _, s := range v.TouchedSymbols {
			b.WriteString("  - " + s + "\n")
		}
		b.WriteString("\n")
	}

	if len(v.Violations) > 0 {
		b.WriteString("Detected Architectural & ADR Violations:\n")
		for _, viol := range v.Violations {
			b.WriteString("  ✗ " + viol + "\n")
		}
		b.WriteString("\n")
	}

	if len(v.SelfRepairInstructions) > 0 {
		b.WriteString("Mandatory Self-Repair Actions:\n")
		for i, inst := range v.SelfRepairInstructions {
			b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, inst))
		}
		b.WriteString("\n")
	}

	b.WriteString("Refactor the offending code now, then invoke `garuda.talon_verify_edit` to confirm `ALLOW`.\n")
	return b.String()
}

func RunHarnessLoop(
	ctx context.Context,
	st *store.PostgresStore,
	tenantID, workspaceID uuid.UUID,
	opts GuardOptions,
	agentArgs []string,
	maxRetries int,
	reIndexFn func(repoDir string),
) (*GuardVerdict, int, error) {
	if maxRetries <= 0 {
		maxRetries = 3
	}

	var lastVerdict *GuardVerdict
	repairPrompt := ""

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if len(agentArgs) > 0 {
			cmd := exec.CommandContext(ctx, agentArgs[0], agentArgs[1:]...)
			cmd.Dir = opts.RepoDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			env := os.Environ()
			env = append(env, "GARUDA_TALON_ENFORCE=true")
			env = append(env, fmt.Sprintf("GARUDA_TALON_ATTEMPT=%d", attempt))
			if repairPrompt != "" {
				env = append(env, "GARUDA_TALON_REPAIR_PROMPT="+repairPrompt)
				cmd.Stdin = bytes.NewBufferString(repairPrompt)
			}
			_ = cmd.Run()
		}

		if reIndexFn != nil {
			reIndexFn(opts.RepoDir)
		}

		verdict, err := RunGuard(ctx, st, tenantID, workspaceID, opts)
		if err != nil {
			return nil, attempt, err
		}
		lastVerdict = verdict

		if verdict.Outcome != string(policy.DecisionBlock) {
			return verdict, attempt, nil
		}
		repairPrompt = verdict.SelfRepairPrompt
	}

	return lastVerdict, maxRetries, nil
}

func sevRank(s impact.SeverityLevel) int {
	switch s {
	case impact.SeverityCritical:
		return 4
	case impact.SeverityHigh:
		return 3
	case impact.SeverityMedium:
		return 2
	case impact.SeverityLow:
		return 1
	default:
		return 0
	}
}
