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

// ResolveWorkspaceTarget resolves workspace by name or falls back to the first available workspace.
// ResolveWorkspaceTarget resolves a workspace by name within the given
// tenant, or falls back to the most recently updated workspace for
// that tenant.
//
// The tenant_id filter is required on both queries. Without it the
// named lookup could match a workspace in a different tenant that
// happened to share a name, and the fallback returned whichever
// workspace was first in the table across the entire deployment.
//
// Same class of bug fixed in resolveWorkspaceID at the api layer
// (9ecf281) and in the tenant package.
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

// FindRepoByTarget searches for a repository by module path or URL match.
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

// GetWorkspaceCounts retrieves counts of repositories, internal entities, and claims.
func (s *PostgresStore) GetWorkspaceCounts(ctx context.Context, workspaceID uuid.UUID) (WorkspaceCounts, error) {
	var c WorkspaceCounts
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM repositories WHERE workspace_id = $1`, workspaceID).Scan(&c.Repositories)
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM entities WHERE workspace_id = $1 AND kind != 'external'`, workspaceID).Scan(&c.Entities)
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM claims WHERE workspace_id = $1`, workspaceID).Scan(&c.Claims)
	return c, nil
}

// GetTopArchitecturalHubs returns entities with the most inbound callers.
func (s *PostgresStore) GetTopArchitecturalHubs(ctx context.Context, workspaceID uuid.UUID, limit int) ([]ArchitecturalHub, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT e.name, e.kind, e.package, count(c.id) as callers
		FROM entities e
		JOIN claims c ON c.to_entity_id = e.id
		WHERE e.workspace_id = $1 AND e.kind != 'external'
		GROUP BY e.id, e.name, e.kind, e.package
		ORDER BY callers DESC
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

// GetCrossRepoBridges returns cross-repository calls in a workspace.
func (s *PostgresStore) GetCrossRepoBridges(ctx context.Context, workspaceID uuid.UUID, limit int) ([]CrossRepoBridge, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT r1.module_path, r2.module_path, count(*) 
		FROM claims c
		JOIN entities e1 ON c.from_entity_id = e1.id
		JOIN entities e2 ON c.to_entity_id = e2.id
		JOIN repositories r1 ON e1.repository_id = r1.id
		JOIN repositories r2 ON e2.repository_id = r2.id
		WHERE c.workspace_id = $1 AND r1.id != r2.id
		GROUP BY r1.module_path, r2.module_path
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

// GetRepoMetrics returns entity and distinct package counts for a repo.
func (s *PostgresStore) GetRepoMetrics(ctx context.Context, repoID uuid.UUID) (RepoSummaryMetrics, error) {
	var m RepoSummaryMetrics
	err := s.pool.QueryRow(ctx, `SELECT count(*), count(DISTINCT package) FROM entities WHERE repository_id = $1 AND kind != 'external'`, repoID).Scan(&m.EntityCount, &m.PackageCount)
	return m, err
}

// GetExportedSymbols returns top exported symbols for a repository.
func (s *PostgresStore) GetExportedSymbols(ctx context.Context, repoID uuid.UUID, limit int) ([]ExportedSymbol, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT name, kind, COALESCE(receiver_type, '')
		FROM entities 
		WHERE repository_id = $1 AND is_exported = true AND kind IN ('function', 'struct', 'interface')
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

// GetExternalDependencies returns external or cross-repo package references.
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

// GetSymbolSummaryDetail resolves an entity by ID or name and counts inbound/outbound claims.
func (s *PostgresStore) GetSymbolSummaryDetail(ctx context.Context, workspaceID uuid.UUID, target string) (*SymbolSummaryDetail, error) {
	var det SymbolSummaryDetail
	var err error

	if parsedID, parseErr := uuid.Parse(target); parseErr == nil {
		err = s.pool.QueryRow(ctx, `
			SELECT id, name, kind, package, receiver_type, file_path, signature, is_exported, line
			FROM entities WHERE workspace_id = $1 AND id = $2 LIMIT 1
		`, workspaceID, parsedID).Scan(&det.ID, &det.Name, &det.Kind, &det.Package, &det.ReceiverType, &det.FilePath, &det.Signature, &det.IsExported, &det.Line)
	} else {
		sym := target
		if dot := strings.LastIndex(target, "."); dot != -1 {
			sym = target[dot+1:]
		}
		err = s.pool.QueryRow(ctx, `
			SELECT id, name, kind, package, receiver_type, file_path, signature, is_exported, line
			FROM entities 
			WHERE workspace_id = $1 AND (name = $2 OR name ILIKE $2)
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

// BriefingDTO is the session-start summary for an MCP agent.
//
// Seven sections are always returned:
//
//	workspace   — which workspace the briefing describes
//	trust       — the tenant's Merkle state (root, block height, verification status)
//	scale       — repositories, entities, claims
//	hubs        — top entities by inbound edge count
//	policies    — active policy count and titles
//	attention   — open runtime contradictions and undocumented code
//	new_since   — diff against the caller's last briefing
//
// session_id and agent_id are returned for correlation. They are not
// keys; the watermark is keyed on agent_id only.
type BriefingDTO struct {
	Workspace   string               `json:"workspace"`
	WorkspaceID uuid.UUID            `json:"workspace_id"`
	SessionID   string               `json:"session_id"`
	AgentID     string               `json:"agent_id"`
	Trust       BriefingTrustDTO     `json:"trust"`
	Scale       BriefingScaleDTO     `json:"scale"`
	Hubs        []BriefingHubDTO     `json:"hubs"`
	Policies    BriefingPolicyDTO    `json:"policies"`
	Attention   BriefingAttentionDTO `json:"attention"`
	NewSince    BriefingDiffDTO      `json:"new_since"`
}

type BriefingTrustDTO struct {
	MerkleRoot  string `json:"merkle_root"`
	BlockHeight int64  `json:"block_height"`
	Status      string `json:"status"`
}

type BriefingScaleDTO struct {
	Repositories     int `json:"repositories"`
	Entities         int `json:"entities"`
	Claims           int `json:"claims"`
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
	OpenContradictions   int `json:"open_contradictions"`
	UndocumentedEntities int `json:"undocumented_entities"`
	DriftFindings        int `json:"drift_findings"`
}

type BriefingDiffDTO struct {
	ClaimsAdded        int  `json:"claims_added"`
	VerificationsSince int  `json:"verifications_since"`
	PoliciesChanged    int  `json:"policies_changed"`
	HasPriorBriefing   bool `json:"has_prior_briefing"`
}

// CountActivePolicies returns the number of active policies for a
// tenant and the titles of up to five of them.
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
		SELECT COALESCE(metadata->>'title', statement, 'unnamed policy') AS title
		  FROM policies
		 WHERE tenant_id = $1 AND status = 'active'
		 ORDER BY created_at DESC
		 LIMIT 5
	`, tenantID)
	if err != nil {
		return count, nil, fmt.Errorf("policy titles: %w", err)
	}
	defer rows.Close()

	var titles []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err == nil {
			titles = append(titles, t)
		}
	}
	return count, titles, rows.Err()
}

// CountOpenContradictions returns the number of runtime contradictions
// for a workspace. Reads claim_verifications, the same table the
// dashboard's "Needs Attention" panel reads.
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

// GetOrCreateWatermark returns the last briefing timestamp for an
// agent in a workspace. If no watermark exists, it writes one and
// returns a zero time.
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

	// No prior watermark. Write a zero-height row so the next call
	// has a starting point.
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

// UpdateWatermark records the moment of the current briefing.
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

// GetBriefing composes the seven-section briefing for an MCP agent.
//
// The watermark is read first. If a prior briefing exists, the diff
// is computed against it. Otherwise HasPriorBriefing is false and the
// diff counts are zero. After composition, the watermark is updated
// to the current moment and block height.
func (s *PostgresStore) GetBriefing(ctx context.Context, tenantID, workspaceID uuid.UUID, workspaceName, agentID, sessionID string) (*BriefingDTO, error) {
	lastBriefedAt, _, hasPrior, err := s.GetOrCreateWatermark(ctx, tenantID, workspaceID, agentID)
	if err != nil {
		return nil, err
	}

	counts, err := s.GetWorkspaceCounts(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

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

	briefing := &BriefingDTO{
		Workspace:   workspaceName,
		WorkspaceID: workspaceID,
		SessionID:   sessionID,
		AgentID:     agentID,
		Trust:       trust,
		Scale: BriefingScaleDTO{
			Repositories: counts.Repositories,
			Entities:     counts.Entities,
			Claims:       counts.Claims,
		},
		Hubs: hubs,
		Policies: BriefingPolicyDTO{
			ActiveCount: policyCount,
			Titles:      policyTitles,
		},
		Attention: BriefingAttentionDTO{
			OpenContradictions: contradictions,
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
