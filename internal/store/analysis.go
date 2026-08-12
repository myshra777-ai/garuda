package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

// SaveAnalysisDecision persists an AST analysis result into the cryptographic ledger
// with content‑addressed evidence binding (non‑repudiation).
func (s *PostgresStore) SaveAnalysisDecision(
	ctx context.Context,
	tenantID uuid.UUID,
	result *analyzer.AnalysisResult,
	provenance *analyzer.Provenance,
) (uuid.UUID, int, error) {
	if result == nil || result.Stats.Files == 0 {
		return uuid.Nil, 0, fmt.Errorf("cannot store empty analysis result")
	}

	// 1. Marshal full heavy AST payload & compute content hash
	fullASTJSON, err := json.Marshal(result)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("marshal full AST: %w", err)
	}
	astHashBytes := sha256.Sum256(fullASTJSON)
	astHashHex := hex.EncodeToString(astHashBytes[:])

	// 2. Build lightweight revision summary (includes payload_hash)
	summary := analyzer.RevisionSummary{
		Fingerprint: result.Fingerprint,
		Source:      result.Source,
		Stats:       result.Stats,
		PayloadHash: astHashHex,
		Provenance:  provenance,
	}
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("marshal summary: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 3. Deterministic Decision ID (v5 SHA‑1)
	namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	decisionID := uuid.NewSHA1(namespace, []byte("analysis:"+tenantID.String()+":"+result.Fingerprint))

	title := fmt.Sprintf("AST Snapshot [%s]", result.Fingerprint[:8])
	statement := fmt.Sprintf("%d packages, %d structs, %d interfaces, %d functions",
		result.Stats.Packages, result.Stats.Structs, result.Stats.Interfaces, result.Stats.Functions)

	// 4. Upsert parent decision (including provenance columns if provided)
	var workspaceID, repoID interface{}
	workspaceID = nil
	repoID = nil
	commitSHA := ""
	if provenance != nil {
		workspaceID = provenance.WorkspaceID
		repoID = provenance.RepoID
		commitSHA = provenance.CommitSHA
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO decisions (
			id, tenant_id, title, statement, status, domain, system, owner,
			fingerprint, workspace_id, repository_id, commit_sha,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, 'APPROVED', 'architecture', 'ast-analyzer', 'garuda-cli', $5, $6, $7, $8, NOW(), NOW())
		ON CONFLICT (tenant_id, id) DO UPDATE
		SET title = EXCLUDED.title,
		    statement = EXCLUDED.statement,
		    fingerprint = EXCLUDED.fingerprint,
		    workspace_id = EXCLUDED.workspace_id,
		    repository_id = EXCLUDED.repository_id,
		    commit_sha = EXCLUDED.commit_sha,
		    updated_at = NOW()
	`, decisionID, tenantID, title, statement, result.Fingerprint,
		workspaceID, repoID, commitSHA)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("upsert decision: %w", err)
	}

	// 5. Next revision number
	var revNumber int
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(revision_number), 0) + 1
		FROM decision_revisions WHERE decision_id = $1 AND tenant_id = $2
	`, decisionID, tenantID).Scan(&revNumber)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("get revision: %w", err)
	}

	// 6. Previous hash
	var prevHash []byte
	if revNumber > 1 {
		err = tx.QueryRow(ctx, `
			SELECT decision_hash FROM decision_revisions
			WHERE decision_id = $1 AND revision_number = $2 AND tenant_id = $3
		`, decisionID, revNumber-1, tenantID).Scan(&prevHash)
		if err != nil {
			return uuid.Nil, 0, fmt.Errorf("fetch prev hash: %w", err)
		}
	} else {
		prevHash = make([]byte, 32)
	}

	// 7. Compute revision hash over summary (includes payload_hash)
	decisionHashBytes := sha256.Sum256(append([]byte(decisionID.String()), summaryJSON...))

	// 8. Insert immutable revision (lightweight summary)
	_, err = tx.Exec(ctx, `
		INSERT INTO decision_revisions (
			decision_id, revision_number, canonical_json, decision_hash,
			previous_revision_hash, tenant_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`, decisionID, revNumber, summaryJSON, decisionHashBytes[:], prevHash, tenantID)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("insert revision: %w", err)
	}

	// 9. Insert full AST into evidence_store (tenant‑isolated, content‑addressed)
	_, err = tx.Exec(ctx, `
		INSERT INTO evidence_store (tenant_id, block_hash, content, ref_count, created_at)
		VALUES ($1, $2, $3, 1, NOW())
		ON CONFLICT (tenant_id, block_hash) DO UPDATE
		SET ref_count = evidence_store.ref_count + 1
	`, tenantID, astHashBytes[:], fullASTJSON)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("insert evidence: %w", err)
	}

	// 10. Update Merkle Root (hex‑encoded)
	// Ensure a row exists first (idempotent) – using a placeholder hex string
	_, err = tx.Exec(ctx, `
		INSERT INTO merkle_roots (tenant_id, root_hash, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (tenant_id) DO NOTHING
	`, tenantID, "0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("ensure merkle row: %w", err)
	}

	var currentRootHex string
	err = tx.QueryRow(ctx, `
		SELECT root_hash FROM merkle_roots WHERE tenant_id = $1 FOR UPDATE
	`, tenantID).Scan(&currentRootHex)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("fetch merkle root with lock: %w", err)
	}

	currentRoot, err := hex.DecodeString(currentRootHex)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("decode current root: %w", err)
	}
	if len(currentRoot) == 0 {
		currentRoot = make([]byte, 32)
	}

	combined := append(currentRoot, decisionHashBytes[:]...)
	newRootBytes := sha256.Sum256(combined)
	newRootHex := hex.EncodeToString(newRootBytes[:])

	_, err = tx.Exec(ctx, `
		UPDATE merkle_roots SET root_hash = $1, updated_at = NOW()
		WHERE tenant_id = $2
	`, newRootHex, tenantID)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("update merkle root: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, 0, fmt.Errorf("commit tx: %w", err)
	}

	return decisionID, revNumber, nil
}
