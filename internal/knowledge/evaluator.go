package knowledge

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Evaluator struct {
	pool *pgxpool.Pool
}

func NewEvaluator(pool *pgxpool.Pool) *Evaluator {
	return &Evaluator{pool: pool}
}

type HealthStats struct {
	TotalClaims      int
	Supported        int
	Unimplemented    int
	Contradicted     int
	TotalEntities    int
	UndocumentedCode int
}

type CodeDriftFinding struct {
	SymbolName string
	Kind       string
	Path       string
}

func (e *Evaluator) EvaluateWorkspace(ctx context.Context, tenantID uuid.UUID, workspace string) (*HealthStats, []CodeDriftFinding, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT id, subject, modality, predicate, object 
		FROM document_claims 
		WHERE tenant_id = $1 AND workspace = $2
	`, tenantID, workspace)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var claims []ClaimIR
	for rows.Next() {
		var c ClaimIR
		var mod, pred string
		if err := rows.Scan(&c.ID, &c.Subject, &mod, &pred, &c.Object); err != nil {
			continue
		}
		c.Modality, c.Predicate = Modality(mod), Predicate(pred)
		claims = append(claims, c)
	}
	rows.Close()

	stats := &HealthStats{TotalClaims: len(claims)}

	for _, c := range claims {
		status := StatusUnverified
		var reason string

		// Resolve Subject (handles strict match or package-prefixed match)
		var subjectID uuid.UUID
		err := e.pool.QueryRow(ctx, `
			SELECT id FROM entities 
			WHERE tenant_id = $1 AND (name = $2 OR name LIKE '%.' || $2)
			LIMIT 1
		`, tenantID, c.Subject).Scan(&subjectID)

		if err != nil {
			status = StatusUnverified
			reason = fmt.Sprintf("DOC->CODE DRIFT: Subject '%s' not found in codebase AST", c.Subject)
			stats.Unimplemented++
			e.updateClaim(ctx, c.ID, nil, status, reason)
			continue
		}

		// Resolve Object
		var objectID uuid.UUID
		err = e.pool.QueryRow(ctx, `
			SELECT id FROM entities 
			WHERE tenant_id = $1 AND (name = $2 OR name LIKE '%.' || $2)
			LIMIT 1
		`, tenantID, c.Object).Scan(&objectID)

		if err != nil && c.Predicate != PredicateIdempotent {
			status = StatusUnverified
			reason = fmt.Sprintf("DOC->CODE DRIFT: Target '%s' not found in codebase AST", c.Object)
			stats.Unimplemented++
			e.updateClaim(ctx, c.ID, &subjectID, status, reason)
			continue
		}

		// Evaluate Structural AST Relationships
		if c.Predicate == PredicateCalls {
			var exists bool
			_ = e.pool.QueryRow(ctx, `
				SELECT EXISTS(
					SELECT 1 FROM relationships 
					WHERE source_id = $1 AND target_id = $2 AND kind = 'calls'
				)
			`, subjectID, objectID).Scan(&exists)

			if c.Modality == ModalityMust {
				if exists {
					status = StatusSupported
					reason = "Verified: AST call relationship satisfied"
					stats.Supported++
				} else {
					status = StatusUnverified
					reason = "DOC->CODE DRIFT: Required call relationship not found in AST"
					stats.Unimplemented++
				}
			} else if c.Modality == ModalityMustNot {
				if exists {
					status = StatusContradicted
					reason = "ARCHITECTURE CONTRADICTION: Forbidden call relationship found in AST"
					stats.Contradicted++
				} else {
					status = StatusSupported
					reason = "Verified: Forbidden call is absent from AST"
					stats.Supported++
				}
			}
		} else {
			status = StatusUnverified
			reason = "Pending Runtime Telemetry / Invariant Validation"
			stats.Unimplemented++
		}

		e.updateClaim(ctx, c.ID, &subjectID, status, reason)
	}

	// Evaluate CODE -> DOC drift
	codeRows, err := e.pool.Query(ctx, `
		SELECT e.name, e.kind, e.file_path
		FROM entities e
		LEFT JOIN document_claims dc 
		  ON (e.name = dc.subject OR e.name LIKE '%.' || dc.subject) 
		  AND dc.tenant_id = e.tenant_id
		WHERE e.tenant_id = $1 
		  AND e.is_exported = true
		  AND e.kind IN ('function', 'method', 'struct')
		  AND dc.id IS NULL
		ORDER BY e.file_path, e.name ASC
		LIMIT 10
	`, tenantID)
	
	var undocumented []CodeDriftFinding
	if err == nil {
		defer codeRows.Close()
		for codeRows.Next() {
			var f CodeDriftFinding
			if err := codeRows.Scan(&f.SymbolName, &f.Kind, &f.Path); err == nil {
				undocumented = append(undocumented, f)
			}
		}
	}

	_ = e.pool.QueryRow(ctx, `SELECT COUNT(*) FROM entities WHERE tenant_id = $1 AND is_exported = true`, tenantID).Scan(&stats.TotalEntities)
	stats.UndocumentedCode = len(undocumented)

	return stats, undocumented, nil
}

func (e *Evaluator) updateClaim(ctx context.Context, id uuid.UUID, entityID *uuid.UUID, status VerificationStatus, reason string) {
	_, _ = e.pool.Exec(ctx, `
		UPDATE document_claims 
		SET status = $1, contradiction_reason = $2, matched_entity_id = $3, updated_at = NOW() 
		WHERE id = $4
	`, string(status), reason, entityID, id)
}
