package types

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DiscoveredFact captures facts identified by execution agents during task processing.
type DiscoveredFact struct {
	ID         string    `json:"id"`
	Claim      string    `json:"claim"`
	Source     string    `json:"source"`
	Confidence float64   `json:"confidence"`
	ObservedAt time.Time `json:"observed_at"`
}

// CheckpointStatus represents the state lifecycle of an agent checkpoint.
type CheckpointStatus string

const (
	CheckpointStatusActive    CheckpointStatus = "active"
	CheckpointStatusHandedOff CheckpointStatus = "handed_off"
	CheckpointStatusResumed   CheckpointStatus = "resumed"
	CheckpointStatusExpired   CheckpointStatus = "expired"
)

// CheckpointData represents the Universal Intermediate Representation (IR).
type CheckpointData struct {
	ExecutionVersion int                    `json:"execution_version"`
	ProgressPercent  int                    `json:"progress_percent"`
	ActiveGoal       string                 `json:"active_goal"`
	ReasoningChain   []string               `json:"reasoning_chain"`
	FilesTouched     []string               `json:"files_touched"`
	DiscoveredFacts  []DiscoveredFact       `json:"discovered_facts"`
	DecisionIDs      []uuid.UUID            `json:"decision_ids"`
	PendingSteps     []string               `json:"pending_steps"`
	StateSnapshot    map[string]interface{} `json:"state_snapshot"`
}

// AgentCheckpoint is the canonical DB entity for structured state persistence.
type AgentCheckpoint struct {
	ID             uuid.UUID        `json:"id"`
	TenantID       uuid.UUID        `json:"tenant_id"`
	AgentID        string           `json:"agent_id"`
	CheckpointName string           `json:"checkpoint_name"`
	TaskID         *uuid.UUID       `json:"task_id,omitempty"`
	Status         CheckpointStatus `json:"status"`
	CheckpointData CheckpointData   `json:"checkpoint_data"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	ExpiresAt      *time.Time       `json:"expires_at,omitempty"`
}

// Checkpoint represents the raw JSON storage layer mapping for database queries.

// HandoffRequest parameters for transferring tasks across agents.
type HandoffRequest struct {
	CheckpointID uuid.UUID `json:"checkpoint_id"`
	FromAgentID  string    `json:"from_agent_id"`
	ToAgentID    string    `json:"to_agent_id"`
	HandoffNote  string    `json:"handoff_note,omitempty"`
}

// HandoffResponse result payload.
type HandoffResponse struct {
	Status       string          `json:"status"`
	CheckpointID uuid.UUID       `json:"checkpoint_id"`
	FromAgentID  string          `json:"from_agent_id"`
	ToAgentID    string          `json:"to_agent_id"`
	ContextState AgentCheckpoint `json:"context_state"`
}

// Checkpoint represents the raw database persistence struct.
type Checkpoint struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	TenantID       uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	AgentID        string          `json:"agent_id" db:"agent_id"`
	CheckpointName string          `json:"checkpoint_name" db:"checkpoint_name"`
	TaskID         *uuid.UUID      `json:"task_id,omitempty" db:"task_id"`
	CheckpointData json.RawMessage `json:"checkpoint_data" db:"checkpoint_data"`
	Status         string          `json:"status" db:"status"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	ExpiresAt      *time.Time      `json:"expires_at,omitempty" db:"expires_at"`
}

type CheckpointRepository interface {
	CreateCheckpoint(ctx context.Context, tenantID uuid.UUID, agentID string, name string, reason string, state json.RawMessage, merkleRoot string) (*AgentCheckpoint, error)
	GetLatestCheckpoint(ctx context.Context, tenantID uuid.UUID, agentID string) (*AgentCheckpoint, error)
}
