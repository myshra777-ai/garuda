package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/knowledge"
)

type ClaimStore struct {
	pool *pgxpool.Pool
}

func NewClaimStore(pool *pgxpool.Pool) *ClaimStore {
	return &ClaimStore{pool: pool}
}

// SaveClaims clears existing claims for the target document path and persists the fresh batch atomically.
func (s *ClaimStore) SaveClaims(ctx context.Context, claims []knowledge.ClaimIR) error {
	if len(claims) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Idempotency check: Group by document path and clear old extractions for this file
	docPaths := make(map[string]bool)
	var tenantID uuid.UUID
	var workspace string

	for _, c := range claims {
		docPaths[c.DocumentPath] = true
		tenantID = c.TenantID
		workspace = c.Workspace
	}

	for path := range docPaths {
		_, err := tx.Exec(ctx, `
			DELETE FROM document_claims 
			WHERE tenant_id = $1 AND workspace = $2 AND document_path = $3
		`, tenantID, workspace, path)
		if err != nil {
			return fmt.Errorf("failed to clear old claims for %s: %w", path, err)
		}
	}

	stmt := `
		INSERT INTO document_claims (
			id, workspace, tenant_id, document_path, document_title, document_type,
			section_title, line_start, line_end, raw_statement, subject,
			modality, predicate, object, provenance_class, confidence, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`

	for _, c := range claims {
		_, err := tx.Exec(ctx, stmt,
			c.ID,
			c.Workspace,
			c.TenantID,
			c.DocumentPath,
			c.DocumentTitle,
			c.DocumentType,
			c.SectionTitle,
			c.LineStart,
			c.LineEnd,
			c.RawStatement,
			c.Subject,
			string(c.Modality),
			string(c.Predicate),
			c.Object,
			string(c.Provenance),
			c.Confidence,
			string(c.Status),
		)
		if err != nil {
			return fmt.Errorf("failed to insert claim %s: %w", c.ID, err)
		}
	}

	return tx.Commit(ctx)
}

func (s *ClaimStore) GetClaimsByWorkspace(ctx context.Context, tenantID uuid.UUID, workspace string) ([]knowledge.ClaimIR, error) {
	query := `
		SELECT 
			id, workspace, tenant_id, document_path, document_title, document_type,
			section_title, line_start, line_end, raw_statement, subject,
			modality, predicate, object, provenance_class, confidence, status,
			matched_entity_id, contradiction_reason, created_at
		FROM document_claims
		WHERE tenant_id = $1 AND workspace = $2
		ORDER BY document_path, line_start ASC
	`

	rows, err := s.pool.Query(ctx, query, tenantID, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to query document claims: %w", err)
	}
	defer rows.Close()

	var results []knowledge.ClaimIR
	for rows.Next() {
		var c knowledge.ClaimIR
		var mod, pred, prov, stat string
		err := rows.Scan(
			&c.ID, &c.Workspace, &c.TenantID, &c.DocumentPath, &c.DocumentTitle, &c.DocumentType,
			&c.SectionTitle, &c.LineStart, &c.LineEnd, &c.RawStatement, &c.Subject,
			&mod, &pred, &c.Object, &prov, &c.Confidence, &stat,
			&c.MatchedEntityID, &c.ContradictionMsg, &c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan claim row: %w", err)
		}
		c.Modality = knowledge.Modality(mod)
		c.Predicate = knowledge.Predicate(pred)
		c.Provenance = knowledge.ProvenanceClass(prov)
		c.Status = knowledge.VerificationStatus(stat)
		results = append(results, c)
	}

	return results, rows.Err()
}
