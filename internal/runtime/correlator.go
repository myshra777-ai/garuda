// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

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
	funcName, namespace := spanFunctionName(obs)
	if funcName == "" {
		return CorrelationResult{EntityID: nil, Confidence: 0.0}
	}

	var entityID uuid.UUID

	// 1. Exact match on name and package suffix. Only possible when the span
	// names a namespace: with an empty namespace the package condition below
	// matches every package, which reported same-named functions in unrelated
	// packages at full confidence.
	if namespace != "" {
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
	}

	// 2. Fallback Match on function/entity name within workspace
	err := c.pool.QueryRow(ctx, `
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

// spanFunctionName extracts the bare function name and namespace a span
// refers to. The name comes from the "code.function" attribute, falling back
// to the span operation; package, receiver and pointer prefixes are removed,
// so "payments.(*Handler).Charge" becomes "Charge".
func spanFunctionName(obs *RuntimeObservation) (name, namespace string) {
	name, _ = obs.Attributes["code.function"].(string)
	if name == "" {
		name = obs.Operation
	}
	namespace, _ = obs.Attributes["code.namespace"].(string)
	namespace = strings.TrimSpace(namespace)

	name = strings.TrimSpace(name)
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	name = strings.Trim(name, "(*)")
	return name, namespace
}
