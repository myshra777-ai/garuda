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
	"github.com/myshra777-ai/garuda/internal/canonical"
	"github.com/myshra777-ai/garuda/internal/merkle"
	"github.com/myshra777-ai/garuda/internal/types"
)

// SubmitDecision atomically creates an immutable revision with:
// - Decision identity creation and row-level lock
// - Hash chain / Merkle root update
// - Append-only revision record
// - Audit event generation
// - Reference-counted evidence store
// - Idempotency tracking
// Everything is executed in a single, atomic, serializable transaction.
func (s *PostgresStore) SubmitDecision(
	ctx context.Context,
	req *types.SubmitDecisionRequest,
	actor string,
	requestID string,
) (*types.SubmitDecisionResult, error) {
	// 1. Check idempotency outside transaction to fast-path duplicated incoming requests
	if req.IdempotencyKey != "" {
		existing, err := s.getDecisionByidempotencyKey(ctx, req.TenantID, req.IdempotencyKey)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	// 2. Begin serializable transaction
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start serializable transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 3. Upsert decision identity and acquire row lock (serialize concurrent updates to same decision)
	decisionID := req.DecisionID
	if decisionID == uuid.Nil {
		decisionID = uuid.New()
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO decisions (id, tenant_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (tenant_id, id) DO NOTHING
	`, decisionID, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create decision identity: %w", err)
	}

	var lockedID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT id FROM decisions WHERE tenant_id = $1 AND id = $2 FOR UPDATE
	`, req.TenantID, decisionID).Scan(&lockedID)
	if err != nil {
		return nil, fmt.Errorf("failed to lock decision identity: %w", err)
	}

	// 4. Calculate next revision number
	var revisionNumber int
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(revision_number), 0) + 1
		FROM decision_revisions
		WHERE tenant_id = $1 AND decision_id = $2
	`, req.TenantID, decisionID).Scan(&revisionNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to compute revision number: %w", err)
	}

	// 5. Build canonical content and hashes
	content := canonical.DecisionContent{
		Title:      req.Title,
		Statement:  req.Statement,
		Scope:      req.Scope,
		Owner:      req.Owner,
		Confidence: req.Confidence,
		ParentID:   req.ParentID,
	}

	contentHash, err := content.Hash()
	if err != nil {
		return nil, fmt.Errorf("failed to hash decision content: %w", err)
	}
	contentHashBytes := contentHash[:]

	canonicalJSON, err := content.CanonicalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to canonicalize content: %w", err)
	}

	// 6. Lock/Get Merkle root with robust UPSERT lock pattern
	var prevRootHex string
	err = tx.QueryRow(ctx, `
		SELECT root_hash FROM merkle_roots WHERE tenant_id = $1 FOR UPDATE
	`, req.TenantID).Scan(&prevRootHex)

	if errors.Is(err, pgx.ErrNoRows) {
		// Hex-string genesis initialization
		prevRootHex = merkle.HashDecision(uuid.Nil, "GENESIS_ROOT", "active", "system", "core", "garuda", nil)
	} else if err != nil {
		return nil, fmt.Errorf("failed to fetch and lock merkle root: %w", err)
	}

	prevRootBytes, err := hex.DecodeString(prevRootHex)
	if err != nil {
		return nil, fmt.Errorf("decode previous merkle root: %w", err)
	}

	// 7. Compute updated Merkle root hash chain: SHA256(prevRoot || contentHash)
	combined := append(prevRootBytes, contentHashBytes...)
	newRoot := sha256.Sum256(combined)
	newRootHex := hex.EncodeToString(newRoot[:])

	// 8. Insert decision revision (append-only)
	revisionID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO decision_revisions (
			id, tenant_id, decision_id, revision_number,
			canonical_json, decision_hash, previous_revision_hash,
			actor, request_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
	`, revisionID, req.TenantID, decisionID, revisionNumber,
		canonicalJSON, contentHashBytes, prevRootBytes, actor, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert decision revision: %w", err)
	}

	// 9. Update Merkle root INSIDE transaction (BEGIN tx context)
	_, err = tx.Exec(ctx, `
        UPDATE merkle_roots
        SET root_hash = $1, height = height + 1, updated_at = NOW()
        WHERE tenant_id = $2
    `, newRootHex, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to update merkle root in transaction: %w", err)
	}

	// 10. Write audit log entry
	auditPayload := map[string]interface{}{
		"decision_id": decisionID,
		"revision_id": revisionID,
		"revision":    revisionNumber,
		"hash":        hex.EncodeToString(contentHashBytes),
		"actor":       actor,
		"request_id":  requestID,
		"idempotency": req.IdempotencyKey,
	}
	auditJSON, err := json.Marshal(auditPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal audit payload: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO audit_events (tenant_id, event_type, actor, payload, request_id, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, req.TenantID, "decision_submitted", actor, auditJSON, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to write audit event: %w", err)
	}

	// 11. Reference-counted evidence store insertion
	if len(req.Evidence) > 0 {
		for _, ev := range req.Evidence {
			hashBytes := ev.Hash[:]
			_, err = tx.Exec(ctx, `
				INSERT INTO evidence_store (block_hash, content, ref_count, created_at)
				VALUES ($1, $2, 1, NOW())
				ON CONFLICT (block_hash) DO UPDATE
				SET ref_count = evidence_store.ref_count + 1
			`, hashBytes, ev.Content)
			if err != nil {
				return nil, fmt.Errorf("failed to insert evidence: %w", err)
			}
		}
	}

	// 12. Save idempotency entry safely
	if req.IdempotencyKey != "" {
		_, err = tx.Exec(ctx, `
			INSERT INTO idempotency_keys (tenant_id, idempotency_key, decision_id, revision_id, created_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (tenant_id, idempotency_key) DO NOTHING
		`, req.TenantID, req.IdempotencyKey, decisionID, revisionID)
		if err != nil {
			return nil, fmt.Errorf("failed to record idempotency key: %w", err)
		}
	}

	// 13. Commit all changes atomically
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	return &types.SubmitDecisionResult{
		DecisionID:     decisionID,
		RevisionID:     revisionID,
		RevisionNumber: revisionNumber,
		ContentHash:    contentHashBytes,
		MerkleRoot:     newRoot[:],
	}, nil
}

// getDecisionByidempotencyKey retrieves a decision from a previous submission with the same key.
func (s *PostgresStore) getDecisionByidempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (*types.SubmitDecisionResult, error) {
	var decisionID, revisionID uuid.UUID
	var revNum int
	var hash, root []byte

	err := s.pool.QueryRow(ctx, `
		SELECT ik.decision_id, ik.revision_id, r.revision_number, r.decision_hash, mr.root_hash
		FROM idempotency_keys ik
		JOIN decision_revisions r ON ik.revision_id = r.id
		JOIN merkle_roots mr ON ik.tenant_id = mr.tenant_id
		WHERE ik.tenant_id = $1 AND ik.idempotency_key = $2
	`, tenantID, key).Scan(&decisionID, &revisionID, &revNum, &hash, &root)

	if err != nil {
		return nil, err
	}

	return &types.SubmitDecisionResult{
		DecisionID:     decisionID,
		RevisionID:     revisionID,
		RevisionNumber: revNum,
		ContentHash:    hash,
		MerkleRoot:     root,
	}, nil
}
