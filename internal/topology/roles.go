// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package topology

import (
	"time"

	"github.com/google/uuid"
)

// AgentRole defines what an agent is permitted to do.
type AgentRole string

const (
	RoleArchitect AgentRole = "ARCHITECT"
	RoleEngineer  AgentRole = "ENGINEER"
	RoleAuditor   AgentRole = "AUDITOR"
)

// AgentPermissions defines the scope and power of an agent.
type AgentPermissions struct {
	AllowedScopes []string `json:"allowed_scopes"`
	CanSupersede  bool     `json:"can_supersede"`
	MaxTokenCap   int64    `json:"max_token_cap"`
	CanPropose    bool     `json:"can_propose"`
	CanExecute    bool     `json:"can_execute"`
	CanVeto       bool     `json:"can_veto"`
}

// RoleDefinition maps roles to permissions.
var RoleDefinitions = map[AgentRole]AgentPermissions{
	RoleArchitect: {
		AllowedScopes: []string{"infra/*", "security/*", "architecture/*"},
		CanSupersede:  true,
		MaxTokenCap:   100000,
		CanPropose:    true,
		CanExecute:    false,
		CanVeto:       false,
	},
	RoleEngineer: {
		AllowedScopes: []string{"infra/*", "app/*", "database/*"},
		CanSupersede:  false,
		MaxTokenCap:   50000,
		CanPropose:    true,
		CanExecute:    true,
		CanVeto:       false,
	},
	RoleAuditor: {
		AllowedScopes: []string{"*"},
		CanSupersede:  false,
		MaxTokenCap:   20000,
		CanPropose:    false,
		CanExecute:    false,
		CanVeto:       true,
	},
}

// Agent represents an active agent with a role.
type Agent struct {
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	Role        AgentRole        `json:"role"`
	Permissions AgentPermissions `json:"permissions"`
	CurrentTask *Task            `json:"current_task,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
}

// Task represents a unit of work with role assignment.
type Task struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	RequiredRole AgentRole `json:"required_role"`
	Scope        string    `json:"scope"`
	Status       string    `json:"status"` // pending, in_progress, completed
	TokenBudget  int64     `json:"token_budget"`
	CreatedAt    time.Time `json:"created_at"`
}

// Topology represents the execution graph.
type Topology struct {
	ID        uuid.UUID  `json:"id"`
	Goal      string     `json:"goal"`
	Agents    []*Agent   `json:"agents"`
	Tasks     []*Task    `json:"tasks"`
	Handoffs  []*Handoff `json:"handoffs"`
	CreatedAt time.Time  `json:"created_at"`
}

// Handoff represents a task transfer.
type Handoff struct {
	FromAgentID uuid.UUID `json:"from_agent_id"`
	ToAgentID   uuid.UUID `json:"to_agent_id"`
	TaskID      uuid.UUID `json:"task_id"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}
