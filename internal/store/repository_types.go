// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Repository represents a Git repository in a workspace
type Repository struct {
	ID             uuid.UUID  `json:"id"`
	WorkspaceID    uuid.UUID  `json:"workspace_id"`
	Provider       string     `json:"provider"`
	URL            string     `json:"url"`
	DefaultBranch  string     `json:"default_branch"`
	Language       string     `json:"language"`
	ModulePath     string     `json:"module_path"`
	Enabled        bool       `json:"enabled"`
	AnalysisStatus string     `json:"analysis_status"`
	CurrentCommit  *string    `json:"current_commit"`
	LastAnalyzedAt *time.Time `json:"last_analyzed_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CrossRepoEdge represents a relationship across repositories
type CrossRepoEdge struct {
	ID               uuid.UUID
	TenantID         uuid.UUID
	WorkspaceID      uuid.UUID
	FromRepoID       uuid.UUID
	ToRepoID         uuid.UUID
	FromEntityID     uuid.UUID
	ToEntityID       *uuid.UUID // nil if unresolved
	RelationshipType string
	Evidence         json.RawMessage
	Resolved         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
