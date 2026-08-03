package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// GetTenantBudget retrieves current balance state for a tenant, initializing defaults if missing.
func (s *PostgresStore) GetTenantBudget(ctx context.Context, tenantID uuid.UUID) (*types.TenantBudget, error) {
	query := `
		INSERT INTO tenant_budgets (tenant_id)
		VALUES ($1)
		ON CONFLICT (tenant_id) DO NOTHING;
	`
	_, _ = s.pool.Exec(ctx, query, tenantID)

	selectQuery := `
		SELECT tenant_id, token_balance, tokens_consumed, execution_limit, executions_consumed, status, monthly_limit, last_reset_at, created_at, updated_at
		FROM tenant_budgets
		WHERE tenant_id = $1;
	`

	var b types.TenantBudget
	err := s.pool.QueryRow(ctx, selectQuery, tenantID).Scan(
		&b.TenantID, &b.TokenBalance, &b.TokensConsumed, &b.ExecutionLimit,
		&b.ExecutionsConsumed, &b.Status, &b.MonthlyLimit, &b.LastResetAt, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tenant budget: %w", err)
	}
	return &b, nil
}

// PreflightCheckAndReserve verifies budget availability before compute begins.
func (s *PostgresStore) PreflightCheckAndReserve(ctx context.Context, tenantID uuid.UUID, estimatedTokens int) error {
	budget, err := s.GetTenantBudget(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("pre-flight budget check failed: %w", err)
	}

	if budget.Status == "exhausted" || budget.TokenBalance < int64(estimatedTokens) {
		return fmt.Errorf("insufficient token budget: required %d, available %d", estimatedTokens, budget.TokenBalance)
	}

	return nil
}

// ConsumeBudgetDeduct deducts tokens/executions atomically inside a database transaction.
func (s *PostgresStore) ConsumeBudgetDeduct(ctx context.Context, tenantID uuid.UUID, req types.BudgetConsumptionRequest) (*types.BudgetConsumptionResponse, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	initQuery := `
		INSERT INTO tenant_budgets (tenant_id) VALUES ($1)
		ON CONFLICT (tenant_id) DO NOTHING;
	`
	if _, err := tx.Exec(ctx, initQuery, tenantID); err != nil {
		return nil, fmt.Errorf("failed to initialize budget: %w", err)
	}

	var b types.TenantBudget
	lockQuery := `
		SELECT tenant_id, token_balance, tokens_consumed, execution_limit, executions_consumed, status, monthly_limit, last_reset_at, created_at, updated_at
		FROM tenant_budgets
		WHERE tenant_id = $1 FOR UPDATE;
	`
	err = tx.QueryRow(ctx, lockQuery, tenantID).Scan(
		&b.TenantID, &b.TokenBalance, &b.TokensConsumed, &b.ExecutionLimit, &b.ExecutionsConsumed, &b.Status,
		&b.MonthlyLimit, &b.LastResetAt, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to lock budget row: %w", err)
	}

	remainingTokens := b.TokenBalance - int64(req.TokensUsed)
	remainingExecs := b.ExecutionLimit - (b.ExecutionsConsumed + req.ExecutionsUsed)
	if remainingTokens < 0 || remainingExecs < 0 {
		_, _ = tx.Exec(ctx, `UPDATE tenant_budgets SET status = 'exhausted', updated_at = NOW() WHERE tenant_id = $1`, tenantID)
		_ = tx.Commit(ctx)

		return &types.BudgetConsumptionResponse{
			Allowed:             false,
			RemainingTokens:     b.TokenBalance,
			RemainingExecutions: b.ExecutionLimit - b.ExecutionsConsumed,
			Status:              "exhausted",
			Budget:              &b,
		}, nil
	}

	updateQuery := `
		UPDATE tenant_budgets
		SET token_balance = token_balance - $1,
		    tokens_consumed = tokens_consumed + $1,
		    executions_consumed = executions_consumed + $2,
		    updated_at = NOW()
		WHERE tenant_id = $3;
	`
	if _, err := tx.Exec(ctx, updateQuery, req.TokensUsed, req.ExecutionsUsed, tenantID); err != nil {
		return nil, fmt.Errorf("failed to update budget balances: %w", err)
	}

	ledgerQuery := `
		INSERT INTO budget_ledger (id, tenant_id, agent_id, task_id, tokens_used, executions_used, operation, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
	_, err = tx.Exec(ctx, ledgerQuery,
		uuid.New(), tenantID, req.AgentID, req.TaskID, req.TokensUsed, req.ExecutionsUsed, req.Operation, time.Now().UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to write budget ledger record: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit budget consumption: %w", err)
	}

	b.TokenBalance = remainingTokens
	b.ExecutionsConsumed += req.ExecutionsUsed
	return &types.BudgetConsumptionResponse{
		Allowed:             true,
		RemainingTokens:     remainingTokens,
		RemainingExecutions: remainingExecs,
		Status:              "active",
		Budget:              &b,
	}, nil
}
