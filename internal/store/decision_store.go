package store

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/merkle"
	"github.com/myshra777-ai/garuda/internal/types"
)

// SaveDecision inserts or updates a decision record and registers its Merkle leaf.
func (s *PostgresStore) SaveDecision(ctx context.Context, d *types.Decision) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if d.Status == "" {
		d.Status = types.StatusDraft
	}
	if d.Owner == "" {
		d.Owner = "agent-system"
	}

	// Fallback sync between nested Scope struct and flat fields
	if d.ScopeDomain == "" && d.Scope.Domain != "" {
		d.ScopeDomain = d.Scope.Domain
	}
	if d.ScopeSystem == "" && d.Scope.System != "" {
		d.ScopeSystem = d.Scope.System
	}

	// 1. Fetch parent Merkle root for tenant (creates genesis if missing)
	root, err := s.GetMerkleRoot(ctx, d.TenantID)
	if err != nil {
		return fmt.Errorf("failed to fetch parent Merkle root: %w", err)
	}
	parentHash := root.RootHash

	// 2. Compute decision leaf hash
	evidenceIDs := make([]string, len(d.EvidenceIDs))
	for i, h := range d.EvidenceIDs {
		evidenceIDs[i] = hex.EncodeToString(h[:])
	}

	decisionHash := merkle.HashDecision(
		d.ID,
		d.Title,
		string(d.Status),
		d.ScopeDomain,
		d.ScopeSystem,
		d.Owner,
		evidenceIDs,
	)

	d.MerkleHash = decisionHash
	d.ParentMerkleHash = parentHash

	// 3. Upsert decision into PostgreSQL matching the (tenant_id, id) constraint	insertQuery := `
	// 3. Upsert decision into PostgreSQL matching the (tenant_id, id) constraint
	insertQuery := `
		INSERT INTO decisions (
			id, tenant_id, title, status, scope_domain, scope_system, owner, 
			merkle_hash, parent_merkle_hash, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		ON CONFLICT (tenant_id, id) DO UPDATE SET
			title = EXCLUDED.title,
			status = EXCLUDED.status,
			scope_domain = EXCLUDED.scope_domain,
			scope_system = EXCLUDED.scope_system,
			merkle_hash = EXCLUDED.merkle_hash,
			parent_merkle_hash = EXCLUDED.parent_merkle_hash,
			updated_at = NOW();
	`

	_, err = s.pool.Exec(ctx, insertQuery,
		d.ID, d.TenantID, d.Title, d.Status, d.ScopeDomain, d.ScopeSystem, d.Owner,
		d.MerkleHash, d.ParentMerkleHash,
	)
	if err != nil {
		return fmt.Errorf("failed to insert decision row: %w", err)
	}

	// 4. Append decision hash to tenant Merkle chain
	_, err = s.AppendMerkleChain(ctx, d.TenantID, decisionHash)
	if err != nil {
		return fmt.Errorf("failed to append to Merkle chain: %w", err)
	}

	return nil
}

// GetDecision retrieves a decision by tenant and ID.
func (s *PostgresStore) GetDecision(ctx context.Context, tenantID, decisionID uuid.UUID) (*types.Decision, error) {
	query := `
		SELECT id, tenant_id, title, status, scope_domain, scope_system, owner,
		       COALESCE(merkle_hash, ''), COALESCE(parent_merkle_hash, ''), created_at, updated_at
		FROM decisions
		WHERE tenant_id = $1 AND id = $2;
	`

	var d types.Decision
	err := s.pool.QueryRow(ctx, query, tenantID, decisionID).Scan(
		&d.ID, &d.TenantID, &d.Title, &d.Status, &d.ScopeDomain, &d.ScopeSystem, &d.Owner,
		&d.MerkleHash, &d.ParentMerkleHash, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get decision: %w", err)
	}

	d.Scope = types.Scope{Domain: d.ScopeDomain, System: d.ScopeSystem}
	return &d, nil
}

// GetDecisionRevisions returns historical snapshots for a decision.
func (s *PostgresStore) GetDecisionRevisions(ctx context.Context, tenantID, decisionID uuid.UUID) ([]types.DecisionRevision, error) {
	query := `
		SELECT id, revision_number, snapshot_json, created_at
		FROM decision_revisions
		WHERE tenant_id = $1 AND decision_id = $2
		ORDER BY revision_number ASC;
	`
	rows, err := s.pool.Query(ctx, query, tenantID, decisionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query revisions: %w", err)
	}
	defer rows.Close()

	var revisions []types.DecisionRevision
	for rows.Next() {
		var rev types.DecisionRevision
		var snapshotJSON []byte
		if err := rows.Scan(&rev.ID, &rev.RevisionNumber, &snapshotJSON, &rev.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan revision: %w", err)
		}
		rev.TenantID = tenantID
		rev.DecisionID = decisionID
		rev.SnapshotJSON = snapshotJSON
		revisions = append(revisions, rev)
	}
	return revisions, nil
}

// GetDecisionsByScope fetches decisions with optional domain/system filters.
func (s *PostgresStore) GetDecisionsByScope(ctx context.Context, tenantID uuid.UUID, domain, system string) ([]*types.Decision, error) {
	query := `
		SELECT id, tenant_id, title, status, scope_domain, scope_system, owner,
		       COALESCE(merkle_hash, ''), COALESCE(parent_merkle_hash, ''), created_at, updated_at
		FROM decisions
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIdx := 2

	if domain != "" {
		query += fmt.Sprintf(" AND scope_domain = $%d", argIdx)
		args = append(args, domain)
		argIdx++
	}
	if system != "" {
		query += fmt.Sprintf(" AND scope_system = $%d", argIdx)
		args = append(args, system)
		argIdx++
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query decisions by scope: %w", err)
	}
	defer rows.Close()

	var decisions []*types.Decision
	for rows.Next() {
		var d types.Decision
		if err := rows.Scan(
			&d.ID, &d.TenantID, &d.Title, &d.Status, &d.ScopeDomain, &d.ScopeSystem, &d.Owner,
			&d.MerkleHash, &d.ParentMerkleHash, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan decision row: %w", err)
		}
		d.Scope = types.Scope{Domain: d.ScopeDomain, System: d.ScopeSystem}
		decisions = append(decisions, &d)
	}
	return decisions, nil
}

// ListDecisionsByParent fetches decisions that have a specific parent.
func (s *PostgresStore) ListDecisionsByParent(ctx context.Context, tenantID, parentID uuid.UUID) ([]*types.Decision, error) {
	query := `
		SELECT id, tenant_id, title, status, scope_domain, scope_system, owner,
		       COALESCE(merkle_hash, ''), COALESCE(parent_merkle_hash, ''), created_at, updated_at
		FROM decisions
		WHERE tenant_id = $1 AND parent_id = $2
		ORDER BY created_at DESC;
	`
	rows, err := s.pool.Query(ctx, query, tenantID, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list decisions by parent: %w", err)
	}
	defer rows.Close()

	var results []*types.Decision
	for rows.Next() {
		var d types.Decision
		if err := rows.Scan(
			&d.ID, &d.TenantID, &d.Title, &d.Status, &d.ScopeDomain, &d.ScopeSystem, &d.Owner,
			&d.MerkleHash, &d.ParentMerkleHash, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan decision row: %w", err)
		}
		d.Scope = types.Scope{Domain: d.ScopeDomain, System: d.ScopeSystem}
		results = append(results, &d)
	}
	return results, nil
}

// parseStatus converts a database string to DecisionStatus.
func parseStatus(s string) types.DecisionStatus {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "draft":
		return types.StatusDraft
	case "review":
		return types.StatusReview
	case "approved":
		return types.StatusApproved
	case "canonical":
		return types.StatusCanonical
	case "superseded":
		return types.StatusSuperseded
	case "archived":
		return types.StatusArchived
	case "deprecated":
		return types.StatusDeprecated
	default:
		return types.StatusDraft
	}
}
