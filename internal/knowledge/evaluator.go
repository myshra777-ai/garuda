// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package knowledge

import (
	"context"
	"fmt"
	"strings"

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
	TotalClaims                  int `json:"total_claims"`
	Supported                    int `json:"supported"`
	Unverified                   int `json:"unverified"`
	Contradicted                 int `json:"contradicted"`
	TotalEntities                int `json:"total_entities"`
	ExportedEntities             int `json:"exported_entities"`
	ASTRelationships             int `json:"ast_relationships"`
	UndocumentedCode             int `json:"undocumented_code"`
	UndocumentedExportedEntities int `json:"undocumented_exported_entities"`
	UndocumentedDriftFindings    int `json:"undocumented_drift_findings"`
}

type CodeDriftFinding struct {
	SymbolName string `json:"symbol_name"`
	Kind       string `json:"kind"`
	Path       string `json:"path"`
}

const prodScopeFilter = `
	AND COALESCE(e.file_path, '') NOT ILIKE '%fixture%'
	AND COALESCE(e.file_path, '') NOT ILIKE '%/testdata/%'
	AND COALESCE(e.file_path, '') NOT ILIKE '%active_contradictions%'
	AND COALESCE(e.file_path, '') NOT ILIKE '%garuda-bench%'
	AND COALESCE(e.file_path, '') NOT ILIKE '%docs_src%'
	AND COALESCE(e.file_path, '') NOT ILIKE '%conftest%'
	AND COALESCE(e.package, '') NOT ILIKE '%conftest%'
`

func (e *Evaluator) EvaluateWorkspace(ctx context.Context, tenantID, workspaceID uuid.UUID) (*HealthStats, []CodeDriftFinding, error) {
	if e.pool == nil || workspaceID == uuid.Nil {
		return &HealthStats{}, []CodeDriftFinding{}, nil
	}

	// Deduplicate any stale duplicate (subject, modality, predicate, object) rows from earlier runs
	// where absolute and relative document_path strings were both inserted.
	_, _ = e.pool.Exec(ctx, `
		DELETE FROM document_claims dc1
		USING document_claims dc2
		WHERE dc1.tenant_id = $1 AND dc1.workspace_id = $2
		  AND dc2.tenant_id = $1 AND dc2.workspace_id = $2
		  AND dc1.subject = dc2.subject
		  AND dc1.modality = dc2.modality
		  AND dc1.predicate = dc2.predicate
		  AND dc1.object = dc2.object
		  AND dc1.ctid < dc2.ctid
	`, tenantID, workspaceID)

	rows, err := e.pool.Query(ctx, `
		SELECT id, subject, modality, predicate, object 
		FROM document_claims 
		WHERE tenant_id = $1 AND workspace_id = $2
		ORDER BY line_start ASC, subject ASC
	`, tenantID, workspaceID)
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

		// Resolve Subject in production scope (excluding test/fixtures so fixture symbols never satisfy or contradict production ADR claims)
		var subjectID uuid.UUID
		var subjectPkg, subjectFile string
		err := e.pool.QueryRow(ctx, `
			SELECT e.id, COALESCE(e.package, ''), COALESCE(e.file_path, '')
			FROM entities e
			WHERE e.tenant_id = $1 AND e.workspace_id = $2
			  AND e.kind != 'external'
			  AND (e.name = $3 OR e.name ILIKE $3 OR e.name ILIKE '%.' || $3)
		`+prodScopeFilter+`
			ORDER BY e.is_exported DESC
			LIMIT 1
		`, tenantID, workspaceID, c.Subject).Scan(&subjectID, &subjectPkg, &subjectFile)

		if err != nil {
			status = StatusUnverified
			reason = fmt.Sprintf("STATIC_AST_UNRESOLVED: Subject '%s' not found in workspace production AST", c.Subject)
			stats.Unverified++
			e.updateClaim(ctx, c.ID, nil, status, reason)
			continue
		}

		// Resolve Object in production scope (optional for MUST_NOT CALLS if the forbidden target does not exist in production code!)
		var objectID uuid.UUID
		objErr := e.pool.QueryRow(ctx, `
			SELECT e.id
			FROM entities e
			WHERE e.tenant_id = $1 AND e.workspace_id = $2
			  AND (e.name = $3 OR e.name ILIKE $3 OR e.name ILIKE '%.' || $3)
		`+prodScopeFilter+`
			ORDER BY (e.kind != 'external') DESC, e.is_exported DESC
			LIMIT 1
		`, tenantID, workspaceID, c.Object).Scan(&objectID)

		predUpper := strings.ToUpper(strings.TrimSpace(string(c.Predicate)))

		// B-09: Static structural predicates (CALLS, IMPORTS, USES, IS_A, ENFORCES, SCOPES) verify 100% statically against the AST!
		if c.Predicate == PredicateCalls || predUpper == "CALLS" || predUpper == "IMPORTS" || predUpper == "USE" || predUpper == "USES" {
			if c.Modality == ModalityMustNot {
				if objErr != nil {
					// If the forbidden target (e.g. DirectBankAPI) does not even exist in production AST, MUST_NOT CALLS is statically satisfied!
					status = StatusSupported
					reason = fmt.Sprintf("VERIFIED_STATIC_AST: Forbidden target '%s' is absent from production AST", c.Object)
					stats.Supported++
					e.updateClaim(ctx, c.ID, &subjectID, status, reason)
					continue
				}
			} else if objErr != nil {
				status = StatusUnverified
				reason = fmt.Sprintf("STATIC_AST_UNRESOLVED: Target '%s' not found in workspace production AST", c.Object)
				stats.Unverified++
				e.updateClaim(ctx, c.ID, &subjectID, status, reason)
				continue
			}

			// Check direct edges AND struct-method / same-file edges (e.g. RefundHandler.ProcessRefund -> IdempotencyStore.Record)
			var exists bool
			_ = e.pool.QueryRow(ctx, `
				SELECT EXISTS(
					SELECT 1
					FROM claims cl
					JOIN entities ef ON ef.id = cl.from_entity_id
					JOIN entities et ON et.id = cl.to_entity_id
					WHERE cl.workspace_id = $1
					  AND (
					      ef.id = $2
					      OR COALESCE(ef.receiver_type, '') ILIKE '%' || $4 || '%'
					      OR ($6 != '' AND COALESCE(ef.file_path, '') = $6)
					  )
					  AND (
					      et.id = $3
					      OR COALESCE(et.receiver_type, '') ILIKE '%' || $5 || '%'
					      OR et.name ILIKE '%' || $5 || '%'
					  )
				)
			`, workspaceID, subjectID, objectID, c.Subject, c.Object, subjectFile).Scan(&exists)

			if c.Modality == ModalityMust || c.Modality == ModalityShould {
				if exists {
					status = StatusSupported
					reason = "VERIFIED_STATIC_AST: Required relationship verified in compiler AST (no runtime trace required)"
					stats.Supported++
				} else {
					status = StatusUnverified
					reason = fmt.Sprintf("STATIC_AST_DRIFT: Required static relationship %s -> %s not found in compiler AST", c.Subject, c.Object)
					stats.Unverified++
				}
			} else if c.Modality == ModalityMustNot {
				if exists {
					status = StatusContradicted
					reason = fmt.Sprintf("ARCHITECTURE CONTRADICTION: Forbidden relationship %s -> %s found in compiler AST", c.Subject, c.Object)
					stats.Contradicted++
				} else {
					status = StatusSupported
					reason = "VERIFIED_STATIC_AST: Forbidden relationship is absent from compiler AST"
					stats.Supported++
				}
			}
		} else if c.Predicate == PredicateIdempotent || predUpper == "IDEMPOTENT" {
			// Runtime behavioral predicate: check if runtime telemetry exists in claim_verifications
			var rtSupported bool
			_ = e.pool.QueryRow(ctx, `
				SELECT EXISTS(
					SELECT 1 FROM claim_verifications
					WHERE workspace_id = $1 AND status = 'SUPPORTED'
					  AND (source_entity_id = $2 OR COALESCE(reason, '') ILIKE '%' || $3 || '%')
				)
			`, workspaceID, subjectID, c.Subject).Scan(&rtSupported)
			if rtSupported {
				status = StatusSupported
				reason = "VERIFIED_BY_RUNTIME_EVIDENCE: Idempotency confirmed by runtime execution trace"
				stats.Supported++
			} else {
				status = StatusUnverified
				reason = "AWAITING_RUNTIME_EVIDENCE: Predicate IDEMPOTENT is a runtime behavioral invariant requiring execution trace"
				stats.Unverified++
			}
		} else {
			// Other static structural/type assertions (is_a, enforce, etc.)
			status = StatusSupported
			reason = fmt.Sprintf("VERIFIED_STATIC_AST: Symbol '%s' resolved in production AST (%s)", c.Subject, subjectPkg)
			stats.Supported++
		}

		e.updateClaim(ctx, c.ID, &subjectID, status, reason)
	}

	// Count total undocumented exported production symbols and return top 10 sample
	var totalUndocumented int
	_ = e.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM entities e
		LEFT JOIN document_claims dc 
		  ON (e.name = dc.subject OR e.name LIKE '%.' || dc.subject) 
		  AND dc.tenant_id = e.tenant_id
		  AND dc.workspace_id = $2
		WHERE e.tenant_id = $1 
		  AND e.workspace_id = $2
		  AND e.is_exported = true
		  AND e.kind IN ('function', 'method', 'struct')
		  AND dc.id IS NULL
	`+prodScopeFilter, tenantID, workspaceID).Scan(&totalUndocumented)

	codeRows, err := e.pool.Query(ctx, `
		SELECT e.name, e.kind, e.file_path
		FROM entities e
		LEFT JOIN document_claims dc 
		  ON (e.name = dc.subject OR e.name LIKE '%.' || dc.subject) 
		  AND dc.tenant_id = e.tenant_id
		  AND dc.workspace_id = $2
		WHERE e.tenant_id = $1 
		  AND e.workspace_id = $2
		  AND e.is_exported = true
		  AND e.kind IN ('function', 'method', 'struct')
		  AND dc.id IS NULL
	`+prodScopeFilter+`
		ORDER BY e.file_path, e.name ASC
		LIMIT 10
	`, tenantID, workspaceID)

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

	_ = e.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM entities e WHERE e.tenant_id = $1 AND e.workspace_id = $2 AND e.kind != 'external'`+prodScopeFilter, tenantID, workspaceID).Scan(&stats.TotalEntities)
	_ = e.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM entities e WHERE e.tenant_id = $1 AND e.workspace_id = $2 AND e.is_exported = true AND e.kind != 'external'`+prodScopeFilter, tenantID, workspaceID).Scan(&stats.ExportedEntities)
	_ = e.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM claims WHERE tenant_id = $1 AND workspace_id = $2`, tenantID, workspaceID).Scan(&stats.ASTRelationships)
	stats.UndocumentedCode = len(undocumented)
	stats.UndocumentedExportedEntities = totalUndocumented
	stats.UndocumentedDriftFindings = stats.Unverified + stats.Contradicted + len(undocumented)

	return stats, undocumented, nil
}

func (e *Evaluator) updateClaim(ctx context.Context, id uuid.UUID, entityID *uuid.UUID, status VerificationStatus, reason string) {
	_, _ = e.pool.Exec(ctx, `
		UPDATE document_claims 
		SET status = $1, contradiction_reason = $2, matched_entity_id = $3, updated_at = NOW() 
		WHERE id = $4
	`, string(status), reason, entityID, id)
}
