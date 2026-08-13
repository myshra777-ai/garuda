package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/myshra777-ai/garuda/internal/analyzer"
)

func (s *PostgresStore) SaveAnalysisDecision(ctx context.Context, tenantIDStr string, result *analyzer.Result) (string, int, error) {
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid tenant UUID string: %w", err)
	}

	if result == nil || result.Stats.Files == 0 {
		return "", 0, fmt.Errorf("cannot store empty analysis result into cryptographic ledger")
	}

	fullASTJSON, err := json.Marshal(result)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal full AST payload: %w", err)
	}
	astHashBytes := sha256.Sum256(fullASTJSON)
	astHashHex := hex.EncodeToString(astHashBytes[:])

	summary := analyzer.CanonicalRevisionSummary{
		Fingerprint: result.Fingerprint,
		Source:      result.Source,
		Stats:       result.Stats,
		PayloadHash: astHashHex,
	}
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal revision summary: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	decisionID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("analysis:"+tenantID.String()+":"+result.Fingerprint))

	title := fmt.Sprintf("Repository AST Schema Snapshot [%s]", result.Fingerprint[:8])
	statement := fmt.Sprintf("Extracted %d packages, %d structs, %d interfaces, %d functions",
		result.Stats.Packages, result.Stats.Structs, result.Stats.Interfaces, result.Stats.Functions)

	_, err = tx.Exec(ctx, `
        INSERT INTO decisions (id, tenant_id, title, statement, status, domain, system, owner, fingerprint, created_at, updated_at)
        VALUES ($1, $2, $3, $4, 'APPROVED', 'architecture', 'ast-analyzer', 'garuda-cli', $5, NOW(), NOW())
        ON CONFLICT (tenant_id, id) DO UPDATE 
        SET title = EXCLUDED.title, statement = EXCLUDED.statement, fingerprint = EXCLUDED.fingerprint, updated_at = NOW()
    `, decisionID, tenantID, title, statement, result.Fingerprint)
	if err != nil {
		return "", 0, fmt.Errorf("failed to upsert decision: %w", err)
	}

	var revNumber int
	err = tx.QueryRow(ctx, `
        SELECT COALESCE(MAX(revision_number), 0) + 1
        FROM decision_revisions WHERE decision_id = $1 AND tenant_id = $2
    `, decisionID, tenantID).Scan(&revNumber)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get revision number: %w", err)
	}

	var prevHash []byte
	if revNumber > 1 {
		err = tx.QueryRow(ctx, `
            SELECT decision_hash FROM decision_revisions 
            WHERE decision_id = $1 AND revision_number = $2 AND tenant_id = $3
        `, decisionID, revNumber-1, tenantID).Scan(&prevHash)
		if err != nil {
			return "", 0, fmt.Errorf("failed to fetch previous revision hash: %w", err)
		}
	} else {
		prevHash = make([]byte, 32)
	}

	decisionHashBytes := sha256.Sum256(append([]byte(decisionID.String()), summaryJSON...))

	_, err = tx.Exec(ctx, `
        INSERT INTO decision_revisions (decision_id, revision_number, canonical_json, decision_hash, previous_revision_hash, tenant_id, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW())
    `, decisionID, revNumber, summaryJSON, decisionHashBytes[:], prevHash, tenantID)
	if err != nil {
		return "", 0, fmt.Errorf("failed to insert decision revision: %w", err)
	}
	// inside SaveAnalysisDecision, after decisionHashBytes is computed

	var currentRootHex string
	err = tx.QueryRow(ctx, `
    SELECT root_hash FROM merkle_roots WHERE tenant_id = $1 FOR UPDATE
`, tenantID).Scan(&currentRootHex)

	var currentRoot []byte
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			currentRoot = make([]byte, 32) // zero root for first entry
		} else {
			return "", 0, fmt.Errorf("failed to fetch merkle root: %w", err)
		}
	} else {
		currentRoot, err = hex.DecodeString(currentRootHex)
		if err != nil {
			return "", 0, fmt.Errorf("failed to decode merkle root hex: %w", err)
		}
	}

	newRoot := sha256.Sum256(append(currentRoot, decisionHashBytes[:]...))
	newRootHex := hex.EncodeToString(newRoot[:])

	_, err = tx.Exec(ctx, `
    INSERT INTO merkle_roots (tenant_id, root_hash, updated_at)
    VALUES ($1, $2, NOW())
    ON CONFLICT (tenant_id) DO UPDATE
    SET root_hash = EXCLUDED.root_hash, updated_at = NOW()
`, tenantID, newRootHex)
	_, err = tx.Exec(ctx, `
        INSERT INTO evidence_store (tenant_id, block_hash, content, ref_count, created_at)
        VALUES ($1, $2, $3, 1, NOW())
        ON CONFLICT (tenant_id, block_hash) DO UPDATE 
        SET ref_count = evidence_store.ref_count + 1
    `, tenantID, astHashBytes[:], fullASTJSON)
	if err != nil {
		return "", 0, fmt.Errorf("failed to insert isolated evidence: %w", err)
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO merkle_roots (tenant_id, root_hash, updated_at)
        VALUES ($1, $2, NOW())
        ON CONFLICT (tenant_id) DO UPDATE SET root_hash = EXCLUDED.root_hash, updated_at = NOW()
    `, tenantID, newRootHex)
	if err != nil {
		return "", 0, fmt.Errorf("failed to update merkle root: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return decisionID.String(), revNumber, nil
}
