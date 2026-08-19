// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package harvester

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// Promoter handles the promotion of candidate harvested decisions.
type Promoter struct {
	store Store
}

// NewPromoter initializes a new Promoter instance.
func NewPromoter(store Store) *Promoter {
	return &Promoter{store: store}
}

// PromoteToDecision elevates a harvested entry into a validated, canonical decision.
func (p *Promoter) PromoteToDecision(ctx context.Context, hdID uuid.UUID, actor string) (*types.Decision, error) {
	hd, err := p.store.GetHarvestedDecision(ctx, hdID)
	if err != nil {
		return nil, err
	}
	if hd.HumanValidated {
		return nil, fmt.Errorf("already validated")
	}

	decision := &types.Decision{
		ID:         uuid.New(),
		Statement:  hd.ExtractedDecision,
		Status:     types.StatusDraft,
		Scope:      types.Scope{Domain: "harvested"},
		Owner:      actor,
		Confidence: hd.Confidence,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	// Use the store's SaveDecision
	if err := p.store.SaveDecision(ctx, decision); err != nil {
		return nil, err
	}

	hd.DecisionID = &decision.ID
	hd.HumanValidated = true
	hd.UpdatedAt = time.Now().UTC()
	if err := p.store.SaveHarvestedDecision(ctx, hd); err != nil {
		return nil, err
	}
	return decision, nil
}

// PromoteAllHighConfidence bulk promotes unvalidated decisions above a given threshold.
func (p *Promoter) PromoteAllHighConfidence(ctx context.Context, threshold float64, actor string) ([]*types.Decision, error) {
	decisions, err := p.store.ListHarvestedDecisions(ctx, "", false)
	if err != nil {
		return nil, fmt.Errorf("failed to list harvested decisions: %w", err)
	}

	var promoted []*types.Decision
	for _, hd := range decisions {
		if hd.Confidence >= threshold {
			d, err := p.PromoteToDecision(ctx, hd.ID, actor)
			if err != nil {
				continue
			}
			promoted = append(promoted, d)
		}
	}
	return promoted, nil
}
