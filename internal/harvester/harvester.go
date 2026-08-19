// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package harvester

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// HarvestedDecision represents a decision extracted from an unstructured source.
type HarvestedDecision struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`   // ✅ Added TenantID to HarvestedDecision Struct
	SourceType        string     `json:"source_type"` // slack, email, notion, jira
	SourceID          string     `json:"source_id"`   // channel_id, email_id, etc.
	SourceURL         string     `json:"source_url,omitempty"`
	RawText           string     `json:"raw_text"`
	ExtractedDecision string     `json:"extracted_decision"`
	Confidence        float64    `json:"confidence"` // 0.0 - 1.0
	HumanValidated    bool       `json:"human_validated"`
	DecisionID        *uuid.UUID `json:"decision_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// Harvester defines the interface for all source harvesters.
type Harvester interface {
	// Name returns the source type (e.g., "slack").
	Name() string

	// Harvest fetches messages/emails from the source and returns extracted decisions.
	Harvest(ctx context.Context, since time.Time) ([]*HarvestedDecision, error)

	// Watch starts a background goroutine that listens for new messages.
	Watch(ctx context.Context, since time.Time, callback func(*HarvestedDecision) error) error
}

// Store defines the persistence interface for harvested decisions.
// Store now includes SaveDecision
type Store interface {
	SaveHarvestedDecision(ctx context.Context, hd *HarvestedDecision) error
	GetHarvestedDecision(ctx context.Context, id uuid.UUID) (*HarvestedDecision, error)
	ListHarvestedDecisions(ctx context.Context, sourceType string, validatedOnly bool) ([]*HarvestedDecision, error)
	SaveDecision(ctx context.Context, d *types.Decision) error // Add this
}
