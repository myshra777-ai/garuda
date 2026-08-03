package types

import (
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

// AgentCheckpoint is the canonical DB entity for state persistence.
type AgentCheckpoint struct {
	ID             uuid.UUID        `json:"id"`
	TenantID       uuid.UUID        `json:"tenant_id"`
	AgentID        string           `json:"agent_id"`
	TaskID         *uuid.UUID       `json:"task_id,omitempty"`
	Status         CheckpointStatus `json:"status"`
	CheckpointData CheckpointData   `json:"checkpoint_data"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	ExpiresAt      *time.Time       `json:"expires_at,omitempty"`
}

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
