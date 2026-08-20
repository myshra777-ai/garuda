// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// Role & Permission Bitmask
// ============================================================

type CapabilityFlags uint32

const (
	CapPropose   CapabilityFlags = 1 << 0
	CapExecute   CapabilityFlags = 1 << 1
	CapVeto      CapabilityFlags = 1 << 2
	CapSupersede CapabilityFlags = 1 << 3
	CapAudit     CapabilityFlags = 1 << 4
)

type AgentRole string

const (
	RoleArchitect AgentRole = "ARCHITECT"
	RoleEngineer  AgentRole = "ENGINEER"
	RoleAuditor   AgentRole = "AUDITOR"
	RoleCustom    AgentRole = "CUSTOM"
)

// AgentPermissions defines role capabilities.
type AgentPermissions struct {
	Capabilities  CapabilityFlags `json:"capabilities"`
	AllowedScopes []string        `json:"allowed_scopes"`
	MaxTokenCap   int64           `json:"max_token_cap"`
}

// DefaultRoleDefinitions maps roles to default permissions.
var DefaultRoleDefinitions = map[AgentRole]AgentPermissions{
	RoleArchitect: {
		AllowedScopes: []string{"infra/*", "security/*", "architecture/*"},
		MaxTokenCap:   100000,
	},
	RoleEngineer: {
		AllowedScopes: []string{"infra/*", "app/*", "database/*"},
		MaxTokenCap:   50000,
	},
	RoleAuditor: {
		AllowedScopes: []string{"*"},
		MaxTokenCap:   20000,
	},
}

// IsScopeAllowed checks if a scope matches any allowed pattern.
func (p *AgentPermissions) IsScopeAllowed(targetScope string) bool {
	for _, pattern := range p.AllowedScopes {
		if pattern == "*" || pattern == targetScope {
			return true
		}
		if len(pattern) > 2 && pattern[len(pattern)-2:] == ":*" {
			prefix := pattern[:len(pattern)-2]
			if len(targetScope) >= len(prefix) && targetScope[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}

// ============================================================
// Topology & Execution Models
// ============================================================

type TopologyStatus string

const (
	TopologyPending   TopologyStatus = "PENDING"
	TopologyActive    TopologyStatus = "ACTIVE"
	TopologyPaused    TopologyStatus = "PAUSED"
	TopologyCompleted TopologyStatus = "COMPLETED"
	TopologyFailed    TopologyStatus = "FAILED"
	TopologyVetoed    TopologyStatus = "VETOED"
)

// Topology represents the entire execution DAG.
type Topology struct {
	ID             uuid.UUID      `json:"id"`
	TenantID       uuid.UUID      `json:"tenant_id"`
	Goal           string         `json:"goal"`
	ScopeDomain    string         `json:"scope_domain"`
	ScopeSystem    string         `json:"scope_system"`
	Status         TopologyStatus `json:"status"`
	MaxTokenBudget int64          `json:"max_token_budget"`
	TokensConsumed int64          `json:"tokens_consumed"`
	Tasks          []*Task        `json:"tasks,omitempty"`
	Handoffs       []*Handoff     `json:"handoffs,omitempty"`
	MerkleRoot     string         `json:"merkle_root"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
}

// Task represents a single step in the topology.

type TaskStatus string

const (
	TaskPending     TaskStatus = "PENDING"
	TaskInProgress  TaskStatus = "IN_PROGRESS"
	TaskCompleted   TaskStatus = "COMPLETED"
	TaskQuarantined TaskStatus = "QUARANTINED"
	TaskFailed      TaskStatus = "FAILED"
)

// Handoff represents state transfer between tasks.
type Handoff struct {
	ID          uuid.UUID `json:"id"`
	TopologyID  uuid.UUID `json:"topology_id"`
	FromTaskID  uuid.UUID `json:"from_task_id"`
	ToTaskID    uuid.UUID `json:"to_task_id"`
	FromAgentID uuid.UUID `json:"from_agent_id"`
	ToAgentID   uuid.UUID `json:"to_agent_id"`
	StateDelta  string    `json:"state_delta"` // JSON compacted state
	Reason      string    `json:"reason"`
	MerkleProof string    `json:"merkle_proof"`
	CreatedAt   time.Time `json:"created_at"`
}

// TopologyTemplate allows saving reusable workflows.
type TopologyTemplate struct {
	ID          uuid.UUID      `json:"id"`
	TenantID    uuid.UUID      `json:"tenant_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	ScopeDomain string         `json:"scope_domain"`
	ScopeSystem string         `json:"scope_system"`
	Tasks       []TemplateTask `json:"tasks"`
	CreatedAt   time.Time      `json:"created_at"`
}

type TemplateTask struct {
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	RequiredRole AgentRole `json:"required_role"`
	Scope        string    `json:"scope"`
	TokenBudget  int64     `json:"token_budget"`
}

// TopologyAudit logs veto events.
type TopologyAudit struct {
	ID         uuid.UUID  `json:"id"`
	TopologyID uuid.UUID  `json:"topology_id"`
	Actor      string     `json:"actor"`
	Action     string     `json:"action"` // "veto", "pause", "resume", "rollback"
	Reason     string     `json:"reason"`
	TaskID     *uuid.UUID `json:"task_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
