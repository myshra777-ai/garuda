package store

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/merkle"
	"github.com/myshra777-ai/garuda/internal/types"
)

// SaveDecision inserts or updates a decision record writing flat scope fields, JSONB scope, and registers its Merkle leaf.
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
	if d.ValidFrom.IsZero() {
		d.ValidFrom = time.Now().UTC()
	}

	// Fallback sync between nested Scope struct and flat fields
	if d.ScopeDomain == "" && d.Scope.Domain != "" {
		d.ScopeDomain = d.Scope.Domain
	}
	if d.ScopeSystem == "" && d.Scope.System != "" {
		d.ScopeSystem = d.Scope.System
	}
	if d.Scope.Domain == "" && d.ScopeDomain != "" {
		d.Scope.Domain = d.ScopeDomain
	}
	if d.Scope.System == "" && d.ScopeSystem != "" {
		d.Scope.System = d.ScopeSystem
	}

	scopeJSON, err := json.Marshal(d.Scope)
	if err != nil {
		return fmt.Errorf("failed to marshal scope: %w", err)
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

	// 3. Upsert decision into PostgreSQL writing scope_domain, scope_system AND scope (JSONB)
	insertQuery := `
		INSERT INTO decisions (
			tenant_id, id, title, status, scope_domain, scope_system, scope, owner, confidence,
			merkle_hash, parent_merkle_hash, created_at, updated_at, approved_at, valid_from, valid_to
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW(), $12, $13, $14)
		ON CONFLICT (tenant_id, id) DO UPDATE SET
			title = EXCLUDED.title,
			status = EXCLUDED.status,
			scope_domain = EXCLUDED.scope_domain,
			scope_system = EXCLUDED.scope_system,
			scope = EXCLUDED.scope,
			owner = EXCLUDED.owner,
			confidence = EXCLUDED.confidence,
			merkle_hash = EXCLUDED.merkle_hash,
			parent_merkle_hash = EXCLUDED.parent_merkle_hash,
			updated_at = NOW(),
			approved_at = EXCLUDED.approved_at,
			valid_from = EXCLUDED.valid_from,
			valid_to = EXCLUDED.valid_to;
	`

	_, err = s.pool.Exec(ctx, insertQuery,
		d.TenantID, d.ID, d.Title, d.Status.String(),
		d.ScopeDomain, d.ScopeSystem, scopeJSON, d.Owner, d.Confidence,
		d.MerkleHash, d.ParentMerkleHash, d.ApprovedAt,
		d.ValidFrom, d.ValidTo,
	)
	if err != nil {
		return fmt.Errorf("failed to save decision: %w", err)
	}

	// 4. Append decision hash to tenant Merkle chain
	_, err = s.AppendMerkleChain(ctx, d.TenantID, decisionHash)
	if err != nil {
		return fmt.Errorf("failed to append to Merkle chain: %w", err)
	}

	return nil
}

// GetDecision retrieves a decision by tenant and ID including Merkle hashes and temporal fields.
func (s *PostgresStore) GetDecision(ctx context.Context, tenantID, decisionID uuid.UUID) (*types.Decision, error) {
	query := `
		SELECT id, tenant_id, title, status, scope_domain, scope_system, owner,
		       COALESCE(merkle_hash, ''), COALESCE(parent_merkle_hash, ''), 
		       approved_at, valid_from, valid_to, created_at, updated_at
		FROM decisions
		WHERE tenant_id = $1 AND id = $2;
	`

	var d types.Decision
	err := s.pool.QueryRow(ctx, query, tenantID, decisionID).Scan(
		&d.ID, &d.TenantID, &d.Title, &d.Status, &d.ScopeDomain, &d.ScopeSystem, &d.Owner,
		&d.MerkleHash, &d.ParentMerkleHash, &d.ApprovedAt, &d.ValidFrom, &d.ValidTo,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get decision: %w", err)
	}

	d.Scope = types.Scope{Domain: d.ScopeDomain, System: d.ScopeSystem}
	return &d, nil
}

// GetDecisionRevisions returns historical snapshots for a decision.
// GetDecisionRevisions is DEPRECATED. Use the new revision store.
func (s *PostgresStore) GetDecisionRevisions(ctx context.Context, tenantID, decisionID uuid.UUID) ([]types.DecisionRevision, error) {
	return []types.DecisionRevision{}, nil
}

// GetDecisionsByScope fetches decisions with optional domain/system filters including temporal fields.
func (s *PostgresStore) GetDecisionsByScope(ctx context.Context, tenantID uuid.UUID, domain, system string) ([]*types.Decision, error) {
	query := `
		SELECT id, tenant_id, title, status, scope_domain, scope_system, owner,
		       COALESCE(merkle_hash, ''), COALESCE(parent_merkle_hash, ''), 
		       approved_at, valid_from, valid_to, created_at, updated_at
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
			&d.MerkleHash, &d.ParentMerkleHash, &d.ApprovedAt, &d.ValidFrom, &d.ValidTo,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan decision row: %w", err)
		}
		d.Scope = types.Scope{Domain: d.ScopeDomain, System: d.ScopeSystem}
		decisions = append(decisions, &d)
	}
	return decisions, nil
}

// ListDecisionsByParent fetches decisions that have a specific parent including temporal fields.
func (s *PostgresStore) ListDecisionsByParent(ctx context.Context, tenantID, parentID uuid.UUID) ([]*types.Decision, error) {
	query := `
		SELECT id, tenant_id, title, status, scope_domain, scope_system, owner,
		       COALESCE(merkle_hash, ''), COALESCE(parent_merkle_hash, ''), 
		       approved_at, valid_from, valid_to, created_at, updated_at
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
			&d.MerkleHash, &d.ParentMerkleHash, &d.ApprovedAt, &d.ValidFrom, &d.ValidTo,
			&d.CreatedAt, &d.UpdatedAt,
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
	case "active":
		return types.StatusActive
	case "quarantined":
		return types.StatusQuarantined
	default:
		return types.StatusDraft
	}
}
