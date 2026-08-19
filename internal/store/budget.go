// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) ConsumeBudget(ctx context.Context, tenantID uuid.UUID, tokens int) error {
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := s.consumeBudgetTx(ctx, tenantID, tokens)
		if err == nil {
			return nil
		}
		// Retry on serialization conflict
		if strings.Contains(err.Error(), "could not serialize access") {
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
			continue
		}
		return err
	}
	return fmt.Errorf("budget update failed after %d retries", maxRetries)
}

func (s *PostgresStore) consumeBudgetTx(ctx context.Context, tenantID uuid.UUID, tokens int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Lock the budget row
	var remaining int64
	err = tx.QueryRow(ctx, `
		SELECT remaining_tokens FROM budgets
		WHERE tenant_id = $1
		FOR UPDATE
	`, tenantID).Scan(&remaining)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("budget row missing for tenant %s", tenantID)
		}
		return err
	}

	if int64(tokens) > remaining {
		return fmt.Errorf("insufficient budget: requested %d, remaining %d", tokens, remaining)
	}

	newRemaining := remaining - int64(tokens)

	// Update the budget
	_, err = tx.Exec(ctx, `
		UPDATE budgets SET remaining_tokens = $1
		WHERE tenant_id = $2
	`, newRemaining, tenantID)
	if err != nil {
		return err
	}

	// Log the consumption
	_, err = tx.Exec(ctx, `
		INSERT INTO budget_ledger (tenant_id, tokens_consumed, remaining_after, correlation_id)
		VALUES ($1, $2, $3, gen_random_uuid())
	`, tenantID, tokens, newRemaining)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
