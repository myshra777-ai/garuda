// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package repository

import (
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Repository struct {
	ID             uuid.UUID  `json:"id"`
	WorkspaceID    uuid.UUID  `json:"workspace_id"`
	Provider       string     `json:"provider"` // github, gitlab, etc.
	URL            string     `json:"url"`
	DefaultBranch  string     `json:"default_branch"`
	Language       string     `json:"language"`
	CurrentCommit  string     `json:"current_commit"`
	Enabled        bool       `json:"enabled"`
	AnalysisStatus string     `json:"analysis_status"` // pending, running, success, failed
	LastAnalyzedAt *time.Time `json:"last_analyzed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type Commit struct {
	RepositoryID uuid.UUID `json:"repository_id"`
	SHA          string    `json:"sha"`
	Author       string    `json:"author"`
	Message      string    `json:"message"`
	AnalyzedAt   time.Time `json:"analyzed_at"`
	AnalysisID   uuid.UUID `json:"analysis_id"` // link to analysis_artifacts
}
