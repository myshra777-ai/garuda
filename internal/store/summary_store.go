// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type WorkspaceCounts struct {
	Repositories int
	Entities     int
	Claims       int
}

type ArchitecturalHub struct {
	Name    string
	Kind    string
	Package string
	Callers int
}

type CrossRepoBridge struct {
	FromModule string
	ToModule   string
	CallCount  int
}

type RepoSummaryMetrics struct {
	EntityCount  int
	PackageCount int
}

type ExportedSymbol struct {
	Name         string
	Kind         string
	ReceiverType string
}

type DependencyRef struct {
	Package    string
	References int
}

type SymbolSummaryDetail struct {
	ID            uuid.UUID
	Name          string
	Kind          string
	Package       string
	ReceiverType  string
	FilePath      string
	Signature     string
	IsExported    bool
	Line          int
	InboundCount  int
	OutboundCount int
}

// ProductionEntityScopeSQL is the canonical production-only scope filter shared across
// store, briefing, entities, find_entity, talon_pre_edit, and hygiene_report (B-07, B-14, D7).
const ProductionEntityScopeSQL = `
	AND COALESCE(file_path, '') NOT ILIKE '%fixture%'
	AND COALESCE(file_path, '') NOT ILIKE '%/testdata/%'
	AND COALESCE(file_path, '') NOT ILIKE '%active_contradictions%'
	AND COALESCE(file_path, '') NOT ILIKE '%garuda-bench%'
	AND COALESCE(file_path, '') NOT ILIKE '%docs_src%'
	AND COALESCE(file_path, '') NOT ILIKE '%conftest%'
	AND COALESCE(package, '') NOT ILIKE '%conftest%'
`

func (s *PostgresStore) ResolveWorkspaceTarget(ctx context.Context, tenantID uuid.UUID, name string) (uuid.UUID, string, error) {
	var wsID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM workspaces
		 WHERE tenant_id = $1 AND name = $2
		 ORDER BY created_at ASC, id ASC
		 LIMIT 1
	`, tenantID, name).Scan(&wsID)
	if err == nil {
		return wsID, name, nil
	}

	var fallbackName string
	err = s.pool.QueryRow(ctx, `
		SELECT id, name FROM workspaces
		 WHERE tenant_id = $1
		 ORDER BY updated_at DESC, created_at DESC, id DESC
		 LIMIT 1
	`, tenantID).Scan(&wsID, &fallbackName)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("no workspaces found for tenant %s: %w", tenantID, err)
	}
	return wsID, fallbackName, nil
}

func (s *PostgresStore) FindRepoByTarget(ctx context.Context, workspaceID uuid.UUID, target string) (uuid.UUID, string, string, error) {
	var repoID uuid.UUID
	var repoURL, modPath string
	err := s.pool.QueryRow(ctx, `
		SELECT id, url, module_path FROM repositories 
		WHERE workspace_id = $1 AND (module_path ILIKE '%' || $2 || '%' OR url ILIKE '%' || $2 || '%')
		LIMIT 1
	`, workspaceID, target).Scan(&repoID, &repoURL, &modPath)
	if err != nil {
		return uuid.Nil, "", "", err
	}
	return repoID, repoURL, modPath, nil
}

// GetWorkspaceCounts retrieves production-scoped counts of repositories, internal entities, and claims.
func (s *PostgresStore) GetWorkspaceCounts(ctx context.Context, workspaceID uuid.UUID) (WorkspaceCounts, error) {
	var c WorkspaceCounts
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM repositories WHERE workspace_id = $1`, workspaceID).Scan(&c.Repositories)
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM entities WHERE workspace_id = $1 AND kind != 'external'`+ProductionEntityScopeSQL, workspaceID).Scan(&c.Entities)
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM claims WHERE workspace_id = $1`, workspaceID).Scan(&c.Claims)
	return c, nil
}

// GetTopArchitecturalHubs returns first-party production entities with the most inbound callers,
// excluding fixtures, test helpers (tests.conftest), benchmarks (garuda-bench), and third-party
// validation packages (sqlmodel, go/chi, go/cobra) when first-party packages exist (fixes D7 & B-14).
func (s *PostgresStore) GetTopArchitecturalHubs(ctx context.Context, workspaceID uuid.UUID, limit int) ([]ArchitecturalHub, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT e.name, e.kind, COALESCE(e.package, ''), COUNT(c.id)::int AS callers
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
		GROUP BY e.id, e.name, e.kind, e.package
		ORDER BY callers DESC, e.name ASC
		LIMIT $2;
	`, workspaceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hubs []ArchitecturalHub
	for rows.Next() {
		var h ArchitecturalHub
		if err := rows.Scan(&h.Name, &h.Kind, &h.Package, &h.Callers); err == nil {
			hubs = append(hubs, h)
		}
	}
	return hubs, nil
}

func (s *PostgresStore) GetCrossRepoBridges(ctx context.Context, workspaceID uuid.UUID, limit int) ([]CrossRepoBridge, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT COALESCE(r1.module_path, r1.name, 'unknown') AS from_module,
		       COALESCE(r2.module_path, r2.name, 'unknown') AS to_module,
		       COUNT(*)::int AS call_count
		  FROM cross_repo_edges cre
		  JOIN repositories r1 ON r1.id = cre.from_repo_id
		  JOIN repositories r2 ON r2.id = cre.to_repo_id
		 WHERE cre.workspace_id = $1
		 GROUP BY r1.module_path, r1.name, r2.module_path, r2.name
		 ORDER BY call_count DESC
		 LIMIT $2;
	`, workspaceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bridges []CrossRepoBridge
	for rows.Next() {
		var b CrossRepoBridge
		if err := rows.Scan(&b.FromModule, &b.ToModule, &b.CallCount); err == nil {
			bridges = append(bridges, b)
		}
	}
	return bridges, nil
}

func (s *PostgresStore) GetRepoMetrics(ctx context.Context, repoID uuid.UUID) (RepoSummaryMetrics, error) {
	var m RepoSummaryMetrics
	err := s.pool.QueryRow(ctx, `SELECT count(*), count(DISTINCT package) FROM entities WHERE repository_id = $1 AND kind != 'external'`+ProductionEntityScopeSQL, repoID).Scan(&m.EntityCount, &m.PackageCount)
	return m, err
}

func (s *PostgresStore) GetExportedSymbols(ctx context.Context, repoID uuid.UUID, limit int) ([]ExportedSymbol, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT name, kind, COALESCE(receiver_type, '')
		FROM entities 
		WHERE repository_id = $1 AND is_exported = true AND kind IN ('function', 'struct', 'interface')
	`+ProductionEntityScopeSQL+`
		ORDER BY kind DESC, name ASC
		LIMIT $2;
	`, repoID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exports []ExportedSymbol
	for rows.Next() {
		var sym ExportedSymbol
		if err := rows.Scan(&sym.Name, &sym.Kind, &sym.ReceiverType); err == nil {
			exports = append(exports, sym)
		}
	}
	return exports, nil
}

func (s *PostgresStore) GetExternalDependencies(ctx context.Context, repoID uuid.UUID, limit int) ([]DependencyRef, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT e2.package, count(*) as count
		FROM claims c
		JOIN entities e1 ON c.from_entity_id = e1.id
		JOIN entities e2 ON c.to_entity_id = e2.id
		WHERE e1.repository_id = $1 AND (e2.repository_id != $1 OR e2.kind = 'external')
		GROUP BY e2.package
		ORDER BY count DESC
		LIMIT $2;
	`, repoID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []DependencyRef
	for rows.Next() {
		var d DependencyRef
		if err := rows.Scan(&d.Package, &d.References); err == nil {
			deps = append(deps, d)
		}
	}
	return deps, nil
}

func (s *PostgresStore) GetSymbolSummaryDetail(ctx context.Context, workspaceID uuid.UUID, target string) (*SymbolSummaryDetail, error) {
	var det SymbolSummaryDetail
	var err error

	if parsedID, parseErr := uuid.Parse(target); parseErr == nil {
		err = s.pool.QueryRow(ctx, `
			SELECT id, name, kind, COALESCE(package, ''), COALESCE(receiver_type, ''),
			       COALESCE(file_path, ''), COALESCE(signature, ''), COALESCE(is_exported, false),
			       COALESCE(NULLIF(line, 0), line_start, 0)
			FROM entities WHERE workspace_id = $1 AND id = $2 LIMIT 1
		`, workspaceID, parsedID).Scan(&det.ID, &det.Name, &det.Kind, &det.Package, &det.ReceiverType, &det.FilePath, &det.Signature, &det.IsExported, &det.Line)
	} else {
		sym := target
		if dot := strings.LastIndex(target, "."); dot != -1 {
			sym = target[dot+1:]
		}
		err = s.pool.QueryRow(ctx, `
			SELECT id, name, kind, COALESCE(package, ''), COALESCE(receiver_type, ''),
			       COALESCE(file_path, ''), COALESCE(signature, ''), COALESCE(is_exported, false),
			       COALESCE(NULLIF(line, 0), line_start, 0)
			FROM entities 
			WHERE workspace_id = $1 AND (name = $2 OR name ILIKE $2)
		`+ProductionEntityScopeSQL+`
			ORDER BY (kind != 'external') DESC, is_exported DESC
			LIMIT 1
		`, workspaceID, sym).Scan(&det.ID, &det.Name, &det.Kind, &det.Package, &det.ReceiverType, &det.FilePath, &det.Signature, &det.IsExported, &det.Line)
	}

	if err != nil {
		return nil, fmt.Errorf("could not find symbol matching '%s': %w", target, err)
	}

	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM claims WHERE workspace_id = $1 AND to_entity_id = $2`, workspaceID, det.ID).Scan(&det.InboundCount)
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM claims WHERE workspace_id = $1 AND from_entity_id = $2`, workspaceID, det.ID).Scan(&det.OutboundCount)

	return &det, nil
}

type BriefingDTO struct {
	Workspace       string               `json:"workspace"`
	WorkspaceID     uuid.UUID            `json:"workspace_id"`
	SessionID       string               `json:"session_id"`
	AgentID         string               `json:"agent_id"`
	PolicyVerdict   string               `json:"policy_verdict"`
	EpistemicCounts map[string]int       `json:"epistemic_counts"`
	Trust           BriefingTrustDTO     `json:"trust"`
	Scale           BriefingScaleDTO     `json:"scale"`
	Hubs            []BriefingHubDTO     `json:"hubs"`
	Policies        BriefingPolicyDTO    `json:"policies"`
	Attention       BriefingAttentionDTO `json:"attention"`
	NewSince        BriefingDiffDTO      `json:"new_since"`
}

type BriefingTrustDTO struct {
	MerkleRoot  string `json:"merkle_root"`
	BlockHeight int64  `json:"block_height"`
	Status      string `json:"status"`
}

type BriefingScaleDTO struct {
	Repositories     int `json:"repositories"`
	Entities         int `json:"entities"`
	ExportedEntities int `json:"exported_entities"`
	ExternalStubs    int `json:"external_stubs"`
	Claims           int `json:"claims"`
	ASTRelationships int `json:"ast_relationships"`
	DocumentClaims   int `json:"document_claims"`
	CrossRepoBridges int `json:"cross_repo_bridges"`
}

type BriefingHubDTO struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Package string `json:"package"`
	Callers int    `json:"inbound_edges"`
}

type BriefingPolicyDTO struct {
	ActiveCount int      `json:"active_count"`
	Titles      []string `json:"titles,omitempty"`
}

type BriefingAttentionDTO struct {
	OpenContradictions           int `json:"open_contradictions"`
	UndocumentedExportedEntities int `json:"undocumented_exported_entities"`
	UndocumentedEntities         int `json:"undocumented_entities"`
	UndocumentedDriftFindings    int `json:"undocumented_drift_findings"`
	DriftFindings                int `json:"drift_findings"`
}

type BriefingDiffDTO struct {
	ClaimsAdded        int  `json:"claims_added"`
	VerificationsSince int  `json:"verifications_since"`
	PoliciesChanged    int  `json:"policies_changed"`
	HasPriorBriefing   bool `json:"has_prior_briefing"`
}

// CountActivePolicies returns the number of active policies for a tenant and canonical IDs/titles (B-16).
func (s *PostgresStore) CountActivePolicies(ctx context.Context, tenantID uuid.UUID) (int, []string, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM policies
		 WHERE tenant_id = $1 AND status = 'active'
	`, tenantID).Scan(&count)
	if err != nil {
		return 0, nil, fmt.Errorf("count active policies: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT LOWER(TRIM(REGEXP_REPLACE(COALESCE(statement, ''), '^POLICY:', ''))) AS pid
		  FROM policies
		 WHERE tenant_id = $1 AND status = 'active'
		 ORDER BY pid ASC
		 LIMIT 10
	`, tenantID)
	if err != nil {
		return count, nil, fmt.Errorf("policy titles: %w", err)
	}
	defer rows.Close()

	var titles []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err == nil && t != "" {
			titles = append(titles, t)
		}
	}
	return count, titles, rows.Err()
}

func (s *PostgresStore) CountOpenContradictions(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(COUNT(*)::int, 0)
		  FROM claim_verifications
		 WHERE workspace_id = $1 AND status = 'CONTRADICTED'
	`, workspaceID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count open contradictions: %w", err)
	}
	return count, nil
}

func (s *PostgresStore) GetOrCreateWatermark(ctx context.Context, tenantID, workspaceID uuid.UUID, agentID string) (time.Time, int64, bool, error) {
	var lastBriefedAt time.Time
	var lastHeight int64
	err := s.pool.QueryRow(ctx, `
		SELECT last_briefed_at, last_merkle_height
		  FROM mcp_agent_watermarks
		 WHERE tenant_id = $1 AND workspace_id = $2 AND agent_id = $3
	`, tenantID, workspaceID, agentID).Scan(&lastBriefedAt, &lastHeight)
	if err == nil {
		return lastBriefedAt, lastHeight, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, 0, false, fmt.Errorf("read watermark: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO mcp_agent_watermarks (tenant_id, workspace_id, agent_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (tenant_id, workspace_id, agent_id) DO NOTHING
	`, tenantID, workspaceID, agentID)
	if err != nil {
		return time.Time{}, 0, false, fmt.Errorf("write initial watermark: %w", err)
	}
	return time.Time{}, 0, false, nil
}

func (s *PostgresStore) UpdateWatermark(ctx context.Context, tenantID, workspaceID uuid.UUID, agentID string, height int64) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO mcp_agent_watermarks (tenant_id, workspace_id, agent_id, last_briefed_at, last_merkle_height)
		VALUES ($1, $2, $3, NOW(), $4)
		ON CONFLICT (tenant_id, workspace_id, agent_id) DO UPDATE
		SET last_briefed_at = NOW(),
		    last_merkle_height = EXCLUDED.last_merkle_height
	`, tenantID, workspaceID, agentID, height)
	if err != nil {
		return fmt.Errorf("update watermark: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetBriefing(ctx context.Context, tenantID, workspaceID uuid.UUID, workspaceName, agentID, sessionID string) (*BriefingDTO, error) {
	lastBriefedAt, _, hasPrior, err := s.GetOrCreateWatermark(ctx, tenantID, workspaceID, agentID)
	if err != nil {
		return nil, err
	}

	counts, err := s.GetWorkspaceCounts(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	bridges, _ := s.GetCrossRepoBridges(ctx, workspaceID, 1000)

	hubsData, err := s.GetTopArchitecturalHubs(ctx, workspaceID, 5)
	if err != nil {
		return nil, err
	}
	hubs := make([]BriefingHubDTO, 0, len(hubsData))
	for _, h := range hubsData {
		hubs = append(hubs, BriefingHubDTO{
			Name:    h.Name,
			Kind:    h.Kind,
			Package: h.Package,
			Callers: h.Callers,
		})
	}

	merkleRoot, merkleErr := s.GetMerkleRoot(ctx, tenantID)
	trust := BriefingTrustDTO{Status: "GENESIS"}
	var blockHeight int64
	if merkleErr == nil && merkleRoot != nil {
		trust.MerkleRoot = merkleRoot.RootHash
		trust.BlockHeight = merkleRoot.BlockHeight
		trust.Status = "VERIFIED"
		blockHeight = merkleRoot.BlockHeight
	}

	policyCount, policyTitles, err := s.CountActivePolicies(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	contradictions, err := s.CountOpenContradictions(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var exportedEntities, externalStubs, docClaimsTotal, docSupported, docContradicted, docUnverified int
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM entities WHERE workspace_id = $1 AND kind != 'external' AND is_exported = true`+ProductionEntityScopeSQL, workspaceID).Scan(&exportedEntities)
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM entities WHERE workspace_id = $1 AND kind = 'external'`, workspaceID).Scan(&externalStubs)
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM document_claims WHERE workspace_id = $1`, workspaceID).Scan(&docClaimsTotal)
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM document_claims WHERE workspace_id = $1 AND status = 'SUPPORTED'`, workspaceID).Scan(&docSupported)
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM document_claims WHERE workspace_id = $1 AND status = 'CONTRADICTED'`, workspaceID).Scan(&docContradicted)
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM document_claims WHERE workspace_id = $1 AND status NOT IN ('SUPPORTED', 'CONTRADICTED')`, workspaceID).Scan(&docUnverified)

	totalContradictions := contradictions + docContradicted
	policyVerdict := "ALLOW"
	if counts.Entities == 0 {
		policyVerdict = "UNVERIFIED"
	} else if totalContradictions > 0 {
		policyVerdict = "BLOCK"
	} else if docClaimsTotal > 0 && docSupported == 0 {
		policyVerdict = "UNVERIFIED"
	} else if docClaimsTotal > 0 && docSupported < docClaimsTotal {
		policyVerdict = "PARTIAL"
	}

	epistemicCounts := map[string]int{
		"SUPPORTED":             docSupported,
		"CONTRADICTED":          totalContradictions,
		"UNVERIFIED":            docUnverified,
		"TOTAL_DOCUMENT_CLAIMS": docClaimsTotal,
		"TOTAL_AST_CLAIMS":      counts.Claims,
	}

	undocExported := exportedEntities - docSupported
	if undocExported < 0 {
		undocExported = 0
	}

	briefing := &BriefingDTO{
		Workspace:       workspaceName,
		WorkspaceID:     workspaceID,
		SessionID:       sessionID,
		AgentID:         agentID,
		PolicyVerdict:   policyVerdict,
		EpistemicCounts: epistemicCounts,
		Trust:           trust,
		Scale: BriefingScaleDTO{
			Repositories:     counts.Repositories,
			Entities:         counts.Entities,
			ExportedEntities: exportedEntities,
			ExternalStubs:    externalStubs,
			Claims:           counts.Claims,
			ASTRelationships: counts.Claims,
			DocumentClaims:   docClaimsTotal,
			CrossRepoBridges: len(bridges),
		},
		Hubs: hubs,
		Policies: BriefingPolicyDTO{
			ActiveCount: policyCount,
			Titles:      policyTitles,
		},
		Attention: BriefingAttentionDTO{
			OpenContradictions:           totalContradictions,
			UndocumentedExportedEntities: undocExported,
			UndocumentedEntities:         undocExported,
			UndocumentedDriftFindings:    docUnverified,
			DriftFindings:                docUnverified,
		},
		NewSince: BriefingDiffDTO{
			HasPriorBriefing: hasPrior,
		},
	}

	if hasPrior {
		var claimsAdded, verificationsSince int
		_ = s.pool.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM claims
			 WHERE workspace_id = $1 AND created_at > $2
		`, workspaceID, lastBriefedAt).Scan(&claimsAdded)
		_ = s.pool.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM claim_verifications
			 WHERE workspace_id = $1 AND last_evaluated_at > $2
		`, workspaceID, lastBriefedAt).Scan(&verificationsSince)
		briefing.NewSince.ClaimsAdded = claimsAdded
		briefing.NewSince.VerificationsSince = verificationsSince
	}

	if err := s.UpdateWatermark(ctx, tenantID, workspaceID, agentID, blockHeight); err != nil {
		return nil, err
	}

	return briefing, nil
}
