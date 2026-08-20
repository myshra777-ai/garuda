// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"time"

	"github.com/google/uuid"
)

// Policy represents a developer intent lock.
type Policy struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	Statement    string                 `json:"statement"`
	ScopeDomain  string                 `json:"scope_domain"`
	ScopeSystem  string                 `json:"scope_system"`
	Actor        string                 `json:"actor"`
	Status       string                 `json:"status"` // active, superseded, expired
	ValidFrom    time.Time              `json:"valid_from"`
	ValidTo      *time.Time             `json:"valid_to,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	SupersededBy *uuid.UUID             `json:"superseded_by,omitempty"`
	MerkleHash   string                 `json:"merkle_hash,omitempty"`
}

// PolicyViolation records when a policy is violated.
type PolicyViolation struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	PolicyID        uuid.UUID  `json:"policy_id"`
	Actor           string     `json:"actor"`
	AttemptedAction string     `json:"attempted_action"`
	DecisionID      *uuid.UUID `json:"decision_id,omitempty"`
	Reason          string     `json:"reason"`
	CreatedAt       time.Time  `json:"created_at"`
}
