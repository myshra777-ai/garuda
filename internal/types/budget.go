package types

import (
	"time"

	"github.com/google/uuid"
)

// TenantBudget tracks remaining token and execution balances for a tenant.
type TenantBudget struct {
	TenantID           uuid.UUID `json:"tenant_id"`
	TokenBalance       int64     `json:"token_balance"`
	TokensConsumed     int64     `json:"tokens_consumed"`
	ExecutionLimit     int       `json:"execution_limit"`
	ExecutionsConsumed int       `json:"executions_consumed"`
	Status             string    `json:"status"` // active, throttled, exhausted
	MonthlyLimit       int64     `json:"monthly_limit"`
	LastResetAt        time.Time `json:"last_reset_at"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// BudgetConsumptionRequest captures a metering event submitted by an agent or gateway.
type BudgetConsumptionRequest struct {
	AgentID        string     `json:"agent_id"`
	TaskID         *uuid.UUID `json:"task_id,omitempty"`
	TokensUsed     int        `json:"tokens_used"`
	ExecutionsUsed int        `json:"executions_used"`
	Operation      string     `json:"operation"`
}

// BudgetConsumptionResponse reports whether consumption succeeded or exceeded quotas.
type BudgetConsumptionResponse struct {
	Allowed             bool          `json:"allowed"`
	RemainingTokens     int64         `json:"remaining_tokens"`
	RemainingExecutions int           `json:"remaining_executions"`
	Status              string        `json:"status"`
	Budget              *TenantBudget `json:"budget"`
}
