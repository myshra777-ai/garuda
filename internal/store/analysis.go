// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/myshra777-ai/garuda/internal/analyzer"
)

// SaveAnalysisDecision persists an analysis result into the immutable ledger.
// It stores the full AST in evidence_store (content‑addressed) and a lightweight
// summary (with payload_hash) in decision_revisions, then updates the Merkle root.
// Returns: decisionID (string), revisionID (uuid.UUID), revisionNumber (int), error
func (s *PostgresStore) SaveAnalysisDecision(
	ctx context.Context,
	tenantIDStr string,
	result *analyzer.Result,
) (string, uuid.UUID, int, error) {
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return "", uuid.Nil, 0, fmt.Errorf("invalid tenant UUID string: %w", err)
	}

	if result == nil || result.Stats.Files == 0 {
		return "", uuid.Nil, 0, fmt.Errorf("cannot store empty analysis result into cryptographic ledger")
	}

	// 1. Marshal full AST and compute its SHA‑256 (content address)
	fullASTJSON, err := json.Marshal(result)
	if err != nil {
		return "", uuid.Nil, 0, fmt.Errorf("failed to marshal full AST payload: %w", err)
	}
	astHashBytes := sha256.Sum256(fullASTJSON)
	astHashHex := fmt.Sprintf("%x", astHashBytes[:])

	// 2. Build lightweight revision summary (includes payload_hash)
	summary := analyzer.RevisionSummary{
		Fingerprint: result.Fingerprint,
		Source:      result.Source,
		Stats:       result.Stats,
		PayloadHash: astHashHex,
	}
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return "", uuid.Nil, 0, fmt.Errorf("failed to marshal revision summary: %w", err)
	}

	// 3. Start atomic transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", uuid.Nil, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 4. Deterministic decision ID (v5 SHA‑1)
	decisionID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("analysis:"+tenantID.String()+":"+result.Fingerprint))

	// 5. Create parent decision row (so the foreign key can be satisfied)
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
		return "", uuid.Nil, 0, fmt.Errorf("failed to upsert decision: %w", err)
	}

	// 6. Get next revision number for this decision
	var revNumber int
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(revision_number), 0) + 1
		FROM decision_revisions WHERE decision_id = $1 AND tenant_id = $2
	`, decisionID, tenantID).Scan(&revNumber)
	if err != nil {
		return "", uuid.Nil, 0, fmt.Errorf("failed to get revision number: %w", err)
	}

	// 7. Get previous revision hash (maintains the chain)
	var prevHash []byte
	if revNumber > 1 {
		err = tx.QueryRow(ctx, `
			SELECT decision_hash FROM decision_revisions
			WHERE decision_id = $1 AND revision_number = $2 AND tenant_id = $3
		`, decisionID, revNumber-1, tenantID).Scan(&prevHash)
		if err != nil {
			return "", uuid.Nil, 0, fmt.Errorf("failed to fetch previous revision hash: %w", err)
		}
	} else {
		prevHash = make([]byte, 32) // zero hash for genesis
	}

	// 8. Compute revision hash over the summary (includes payload_hash)
	decisionHashBytes := sha256.Sum256(append([]byte(decisionID.String()), summaryJSON...))

	// 9. Insert immutable decision revision
	revisionID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO decision_revisions (
			id, decision_id, revision_number, canonical_json,
			decision_hash, previous_revision_hash, tenant_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`, revisionID, decisionID, revNumber, summaryJSON,
		decisionHashBytes[:], prevHash, tenantID)
	if err != nil {
		return "", uuid.Nil, 0, fmt.Errorf("failed to insert decision revision: %w", err)
	}

	// 10. Store full AST payload in evidence_store
	_, err = tx.Exec(ctx, `
		INSERT INTO evidence_store (tenant_id, block_hash, content, ref_count, created_at)
		VALUES ($1, $2, $3, 1, NOW())
		ON CONFLICT (tenant_id, block_hash) DO UPDATE
		SET ref_count = evidence_store.ref_count + 1
	`, tenantID, astHashBytes[:], fullASTJSON)
	if err != nil {
		return "", uuid.Nil, 0, fmt.Errorf("failed to insert isolated evidence: %w", err)
	}

	// 11. Update Merkle root (BYTEA – no hex conversion)
	var currentRoot []byte
	err = tx.QueryRow(ctx, `
		SELECT root_hash FROM merkle_roots WHERE tenant_id = $1 FOR UPDATE
	`, tenantID).Scan(&currentRoot)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			currentRoot = make([]byte, 32) // zero root for first entry
		} else {
			return "", uuid.Nil, 0, fmt.Errorf("failed to fetch merkle root: %w", err)
		}
	}

	newRoot := sha256.Sum256(append(currentRoot, decisionHashBytes[:]...))

	_, err = tx.Exec(ctx, `
		INSERT INTO merkle_roots (tenant_id, root_hash, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (tenant_id) DO UPDATE
		SET root_hash = EXCLUDED.root_hash, updated_at = NOW()
	`, tenantID, newRoot[:])
	if err != nil {
		return "", uuid.Nil, 0, fmt.Errorf("failed to update merkle root: %w", err)
	}

	// 12. Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return "", uuid.Nil, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return decisionID.String(), revisionID, revNumber, nil
}
