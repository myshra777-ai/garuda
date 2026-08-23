// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package runtime

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EntityCorrelator struct {
	pool *pgxpool.Pool
}

func NewEntityCorrelator(pool *pgxpool.Pool) *EntityCorrelator {
	return &EntityCorrelator{pool: pool}
}

func (c *EntityCorrelator) Correlate(ctx context.Context, tenantID, workspaceID uuid.UUID, obs *RuntimeObservation) CorrelationResult {
	funcName, _ := obs.Attributes["code.function"].(string)
	if funcName == "" {
		funcName = obs.Operation
	}

	namespace, _ := obs.Attributes["code.namespace"].(string)

	// Clean function name (strip pointer receiver or package prefixes)
	if idx := strings.LastIndex(funcName, "."); idx != -1 {
		funcName = funcName[idx+1:]
	}
	funcName = strings.TrimPrefix(funcName, "*")

	// 1. Exact Match on name and package suffix
	var entityID uuid.UUID
	err := c.pool.QueryRow(ctx, `
		SELECT id FROM entities 
		WHERE workspace_id = $1 
		  AND name = $2 
		  AND ($3 LIKE '%' || package OR package LIKE '%' || $3)
		ORDER BY (kind != 'external') DESC, is_exported DESC
		LIMIT 1
	`, workspaceID, funcName, namespace).Scan(&entityID)

	if err == nil && entityID != uuid.Nil {
		return CorrelationResult{EntityID: &entityID, Confidence: 1.0}
	}

	// 2. Fallback Match on function/entity name within workspace
	err = c.pool.QueryRow(ctx, `
		SELECT id FROM entities 
		WHERE workspace_id = $1 
		  AND name = $2
		ORDER BY (kind != 'external') DESC, is_exported DESC
		LIMIT 1
	`, workspaceID, funcName).Scan(&entityID)

	if err == nil && entityID != uuid.Nil {
		return CorrelationResult{EntityID: &entityID, Confidence: 0.85}
	}

	return CorrelationResult{EntityID: nil, Confidence: 0.0}
}
