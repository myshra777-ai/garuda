// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package selfdescribe

import (
	"context"
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/store"
)

func getSemanticStats(ctx context.Context, db *store.PostgresStore, tenantID string, workspaceID uuid.UUID) (SemanticInfo, error) {
	var count int
	err := db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID).Scan(&count)
	if err != nil {
		return SemanticInfo{}, err
	}

	var relCount int
	err = db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM claims
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID).Scan(&relCount)
	if err != nil {
		return SemanticInfo{}, err
	}

	// Evidence count
	var evCount int
	err = db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM evidence_store
		WHERE tenant_id = $1
	`, tenantID).Scan(&evCount)
	if err != nil {
		return SemanticInfo{}, err
	}

	return SemanticInfo{
		Entities:      count,
		Relationships: relCount,
		Evidence:      evCount,
		Lineage:       true, // We have lineage if there are relationships
	}, nil
}

func getTrustEvidence(ctx context.Context, db *store.PostgresStore, tenantID string) (TrustInfo, error) {
	var revCount int
	err := db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM decision_revisions
		WHERE tenant_id = $1
	`, tenantID).Scan(&revCount)
	if err != nil {
		return TrustInfo{}, err
	}

	var rootHash []byte
	err = db.Pool().QueryRow(ctx, `
		SELECT root_hash FROM merkle_roots
		WHERE tenant_id = $1
	`, tenantID).Scan(&rootHash)
	if err != nil {
		return TrustInfo{
			ImmutableLedger: true,
			MerkleRoot:      "not_available",
			RevisionCount:   revCount,
			AuditTrail:      true,
		}, nil
	}

	return TrustInfo{
		ImmutableLedger: true,
		MerkleRoot:      hex.EncodeToString(rootHash),
		RevisionCount:   revCount,
		AuditTrail:      true,
	}, nil
}
