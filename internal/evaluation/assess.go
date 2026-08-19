// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package evaluation

import (
	"github.com/myshra777-ai/garuda/internal/analyzer"
)

// ChangeStatus represents the classification of a change.
type ChangeStatus string

const (
	StatusNecessary ChangeStatus = "NECESSARY"
	StatusRedundant ChangeStatus = "REDUNDANT"
	StatusConflict  ChangeStatus = "CONFLICT"
	StatusBreaking  ChangeStatus = "BREAKING"
	StatusUnknown   ChangeStatus = "UNKNOWN"
)

// Assessment is the result of evaluating a change.
type Assessment struct {
	Status           ChangeStatus `json:"status"`
	Confidence       float64      `json:"confidence"`
	AffectedEntities []string     `json:"affected_entities"`
	AffectedClaims   []string     `json:"affected_claims"`
	Evidence         []string     `json:"evidence"`
}

// AssessChange evaluates a change between two snapshots.
func AssessChange(baseline, proposed *analyzer.Result) (*Assessment, error) {
	report := analyzer.Diff(baseline, proposed)

	// 1. Breaking changes: modified entities with incoming references
	for _, ed := range report.EntityDiffs {
		if ed.Status == "modified" && ed.Impact > 0 {
			return &Assessment{
				Status:           StatusBreaking,
				Confidence:       0.95,
				AffectedEntities: []string{ed.Name},
				Evidence:         []string{"Entity modified, has incoming references"},
			}, nil
		}
	}

	// 2. Redundancy: added entity with same name as existing entity
	for _, ed := range report.EntityDiffs {
		if ed.Status == "added" {
			for _, existing := range baseline.Entities {
				if existing.Name == ed.Name {
					return &Assessment{
						Status:           StatusRedundant,
						Confidence:       0.85,
						AffectedEntities: []string{ed.Name, existing.Name},
						Evidence:         []string{"Entity added, but an entity with the same name already exists in the baseline"},
					}, nil
				}
			}
		}
	}

	// 3. Necessary: any change with new entities or relationships
	if len(report.EntityDiffs) > 0 || len(report.RelationshipDiffs) > 0 {
		return &Assessment{
			Status:     StatusNecessary,
			Confidence: 0.7,
			Evidence:   []string{"Change introduces new entities or relationships"},
		}, nil
	}

	// 4. Unknown: nothing changed
	return &Assessment{
		Status:     StatusUnknown,
		Confidence: 0.0,
		Evidence:   []string{"No changes detected"},
	}, nil
}
