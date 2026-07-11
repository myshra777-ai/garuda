package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/techtaytor/garuda/internal/types"
)

func (s *PostgresStore) SaveDecision(ctx context.Context, d *types.Decision) error {
	// Convert EvidenceIDs to byte slices
	evidenceIDs := make([][]byte, len(d.EvidenceIDs))
	for i, h := range d.EvidenceIDs {
		evidenceIDs[i] = h[:]
	}

	// Serialize TemporalMetadata to JSONB
	temporalJSON, err := json.Marshal(d.TemporalMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal temporal metadata: %w", err)
	}

	query := `
		INSERT INTO decisions (
			tenant_id, id, title, status, fingerprint,
			evidence_ids, temporal_metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (tenant_id, id) DO UPDATE SET
			status = EXCLUDED.status,
			title = EXCLUDED.title,
			fingerprint = EXCLUDED.fingerprint,
			evidence_ids = EXCLUDED.evidence_ids,
			temporal_metadata = EXCLUDED.temporal_metadata,
			updated_at = NOW();
	`

	_, err = s.pool.Exec(ctx, query,
		d.TenantID, d.ID, d.Title, string(d.Status), d.Fingerprint,
		evidenceIDs, temporalJSON, d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save decision: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetDecision(ctx context.Context, tenantID, decisionID uuid.UUID) (*types.Decision, error) {
	var d types.Decision
	var statusStr string
	var evidenceIDs [][]byte
	var temporalJSON []byte

	query := `
		SELECT id, title, status, fingerprint, evidence_ids, temporal_metadata, created_at, updated_at
		FROM decisions
		WHERE tenant_id = $1 AND id = $2;
	`

	err := s.pool.QueryRow(ctx, query, tenantID, decisionID).Scan(
		&d.ID, &d.Title, &statusStr, &d.Fingerprint,
		&evidenceIDs, &temporalJSON, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get decision: %w", err)
	}

	// Convert status string to DecisionStatus
	d.Status = types.DecisionStatus(statusStr)

	// Convert evidenceIDs from [][]byte to []types.EvidenceHash
	d.EvidenceIDs = make([]types.EvidenceHash, len(evidenceIDs))
	for i, b := range evidenceIDs {
		if len(b) != 32 {
			return nil, fmt.Errorf("evidence hash has unexpected length: %d", len(b))
		}
		copy(d.EvidenceIDs[i][:], b)
	}

	// Unmarshal temporal metadata
	if err := json.Unmarshal(temporalJSON, &d.TemporalMetadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal temporal metadata: %w", err)
	}

	d.TenantID = tenantID

	return &d, nil
}

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
