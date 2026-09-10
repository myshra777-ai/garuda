// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package knowledge

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TelemetryEvent struct {
	TenantID       uuid.UUID
	Workspace      string
	TraceID        string
	SpanName       string
	Operation      string
	HasIdempotency bool
}

type TelemetryVerifier struct {
	pool *pgxpool.Pool
}

func NewTelemetryVerifier(pool *pgxpool.Pool) *TelemetryVerifier {
	return &TelemetryVerifier{pool: pool}
}

// IngestSpan records a runtime OpenTelemetry span for behavioral validation.
func (t *TelemetryVerifier) IngestSpan(ctx context.Context, event TelemetryEvent) error {
	_, err := t.pool.Exec(ctx, `
		INSERT INTO runtime_spans (tenant_id, workspace, trace_id, span_name, operation, has_idempotency)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, event.TenantID, event.Workspace, event.TraceID, event.SpanName, event.Operation, event.HasIdempotency)
	return err
}

// VerifyRuntimeClaim checks if telemetry spans substantiate runtime predicates (like IDEMPOTENT).
func (t *TelemetryVerifier) VerifyRuntimeClaim(ctx context.Context, tenantID uuid.UUID, workspace string, subject string) (bool, error) {
	var exists bool
	err := t.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM runtime_spans 
			WHERE tenant_id = $1 AND workspace = $2 AND (span_name ILIKE $3 OR operation ILIKE $3) AND has_idempotency = true
		)
	`, tenantID, workspace, "%"+subject+"%").Scan(&exists)
	return exists, err
}
