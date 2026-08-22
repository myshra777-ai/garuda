// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

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

	// 5. Create parent decision row (so foreign keys can be satisfied)
	title := fmt.Sprintf("Repository AST Schema Snapshot [%s]", result.Fingerprint[:8])
	statement := fmt.Sprintf("Extracted %d packages, %d structs, %d interfaces, %d functions",
		result.Stats.Packages, result.Stats.Structs, result.Stats.Interfaces, result.Stats.Functions)
	scopeJSON := []byte(`{"domain":"architecture","system":"ast-analyzer"}`)

	_, err = tx.Exec(ctx, `
		INSERT INTO decisions (
			tenant_id, id, title, statement, rationale, status,
			domain, system, scope_domain, scope_system, scope, owner, confidence, fingerprint,
			valid_from, approved_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $4, 'APPROVED',
			'architecture', 'ast-analyzer', 'architecture', 'ast-analyzer', $5, 'garuda-cli', 1.0, $6,
			NOW(), NOW(), NOW(), NOW()
		)
		ON CONFLICT (tenant_id, id) DO UPDATE SET
			title = EXCLUDED.title,
			statement = EXCLUDED.statement,
			rationale = EXCLUDED.rationale,
			fingerprint = EXCLUDED.fingerprint,
			scope = EXCLUDED.scope,
			updated_at = NOW()
	`, tenantID, decisionID, title, statement, scopeJSON, result.Fingerprint)
	if err != nil {
		return "", uuid.Nil, 0, fmt.Errorf("failed to upsert decision: %w", err)
	}

	// 6. Get next revision number and previous revision hash for this decision
	var revNumber int
	var prevHash []byte
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(revision_number), 0) + 1
		FROM decision_revisions
		WHERE tenant_id = $1 AND decision_id = $2
	`, tenantID, decisionID).Scan(&revNumber)
	if err != nil {
		return "", uuid.Nil, 0, fmt.Errorf("failed to get revision number: %w", err)
	}

	if revNumber > 1 {
		_ = tx.QueryRow(ctx, `
			SELECT decision_hash 
			FROM decision_revisions 
			WHERE tenant_id = $1 AND decision_id = $2 
			ORDER BY revision_number DESC LIMIT 1
		`, tenantID, decisionID).Scan(&prevHash)
	}
	if len(prevHash) == 0 {
		prevHash = make([]byte, 32)
	}

	// 7. Compute revision hash over the summary
	decisionHashBytes := sha256.Sum256(append([]byte(decisionID.String()), summaryJSON...))

	// 8. Insert immutable decision revision with tenant isolation & audit columns
	revisionID := uuid.New()
	emptyAssumptions := []byte("[]")
	_, err = tx.Exec(ctx, `
		INSERT INTO decision_revisions (
			id, tenant_id, decision_id, revision_number,
			assumptions, facts, canonical_json, payload_hash,
			decision_hash, previous_revision_hash,
			author, actor, valid_from, created_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10,
			'garuda-cli', 'garuda-cli', NOW(), NOW()
		)
	`, revisionID, tenantID, decisionID, revNumber,
		emptyAssumptions, summaryJSON, summaryJSON, astHashHex,
		decisionHashBytes[:], prevHash)
	if err != nil {
		return "", uuid.Nil, 0, fmt.Errorf("failed to insert decision revision: %w", err)
	}

	// 9. Store full AST payload in evidence_store with tenant isolation
	_, err = tx.Exec(ctx, `
		INSERT INTO evidence_store (tenant_id, block_hash, content, ref_count, created_at)
		VALUES ($1, $2, $3, 1, NOW())
		ON CONFLICT (tenant_id, block_hash) DO UPDATE
		SET ref_count = evidence_store.ref_count + 1
	`, tenantID, astHashBytes[:], string(fullASTJSON))
	if err != nil {
		return "", uuid.Nil, 0, fmt.Errorf("failed to insert isolated evidence: %w", err)
	}

	// 10. Update Merkle root
	// 10. Update Merkle root (using BYTEA binary representation)
	var currentRootBytes []byte
	err = tx.QueryRow(ctx, `
		SELECT root_hash FROM merkle_roots WHERE tenant_id = $1 FOR UPDATE
	`, tenantID).Scan(&currentRootBytes)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			currentRootBytes = make([]byte, 32)
		} else {
			return "", uuid.Nil, 0, fmt.Errorf("failed to fetch merkle root: %w", err)
		}
	}

	newRoot := sha256.Sum256(append(currentRootBytes, decisionHashBytes[:]...))

	_, err = tx.Exec(ctx, `
		INSERT INTO merkle_roots (tenant_id, root_hash, block_height, updated_at)
		VALUES ($1, $2, 1, NOW())
		ON CONFLICT (tenant_id) DO UPDATE
		SET root_hash = EXCLUDED.root_hash, block_height = merkle_roots.block_height + 1, updated_at = NOW()
`, tenantID, hex.EncodeToString(newRoot[:]))

	if err != nil {
		return "", uuid.Nil, 0, fmt.Errorf("failed to update merkle root: %w", err)
	}

	// 11. Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return "", uuid.Nil, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return decisionID.String(), revisionID, revNumber, nil
}
