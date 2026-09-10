// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

// Package policy implements a declarative, deterministic enforcement layer
// on top of the semantic graph. Policies are inspectable YAML, evaluations
// are pure functions, and every decision is Merkle-anchored.
package policy

import (
	"time"

	"github.com/google/uuid"
)

// Decision is the outcome of evaluating a policy against a workspace.
type Decision string

const (
	DecisionAllow  Decision = "ALLOW"
	DecisionWarn   Decision = "WARN"
	DecisionReview Decision = "REVIEW"
	DecisionBlock  Decision = "BLOCK"
)

// Severity returns a numeric rank for decision precedence.
// When multiple policies fire, the highest-severity decision wins.
func (d Decision) Severity() int {
	switch d {
	case DecisionBlock:
		return 4
	case DecisionReview:
		return 3
	case DecisionWarn:
		return 2
	case DecisionAllow:
		return 1
	default:
		return 0
	}
}

// Policy is the parsed form of a policy YAML file.
type Policy struct {
	ID          string      `yaml:"id"          json:"id"`
	Version     string      `yaml:"version"     json:"version"`
	Title       string      `yaml:"title"       json:"title"`
	Description string      `yaml:"description" json:"description"`
	Priority    int         `yaml:"priority"    json:"priority"`  // higher = evaluated first
	Language    string      `yaml:"language"    json:"language"`  // "go", "python", "typescript", "any"
	Authority   string      `yaml:"authority"   json:"authority"` // who approved the policy
	Scope       Scope       `yaml:"scope"       json:"scope"`
	When        []Predicate `yaml:"when"        json:"when"` // all must match (AND)
	Then        Outcome     `yaml:"then"        json:"then"`
	ExpiresAt   *time.Time  `yaml:"expires_at,omitempty" json:"expires_at,omitempty"`
}

// Scope narrows the entities/claims the policy applies to.
// Empty fields mean "any".
type Scope struct {
	Domain  string `yaml:"domain,omitempty"   json:"domain,omitempty"`
	System  string `yaml:"system,omitempty"   json:"system,omitempty"`
	Package string `yaml:"package,omitempty"  json:"package,omitempty"`
	Pattern string `yaml:"pattern,omitempty"  json:"pattern,omitempty"` // regex on entity names
}

// Predicate is a single deterministic check.
// Supported Type values:
//
//   - entity_exists         — an entity matching params matches exists
//   - claim_exists          — a claim matching params exists
//   - contradiction_exists  — a CONTRADICTED claim exists in scope
//   - verification_missing  — an entity has no runtime observation
//   - language_matches      — the scope contains entities of a language
type Predicate struct {
	Type   string         `yaml:"type"   json:"type"`
	Params map[string]any `yaml:"params" json:"params"`
}

// Outcome is what happens when all predicates match.
type Outcome struct {
	Decision Decision `yaml:"decision" json:"decision"`
	Reason   string   `yaml:"reason"   json:"reason"`
}

// Evaluation is the immutable record of a policy run.
type Evaluation struct {
	ID                uuid.UUID `json:"id"`
	TenantID          uuid.UUID `json:"tenant_id"`
	WorkspaceID       uuid.UUID `json:"workspace_id"`
	PolicyID          uuid.UUID `json:"policy_id"`
	PolicyVersion     string    `json:"policy_version"`
	Decision          Decision  `json:"decision"`
	Reason            string    `json:"reason"`
	Evidence          Evidence  `json:"evidence"`
	EvaluatedAt       time.Time `json:"evaluated_at"`
	MerkleBlockHeight *int64    `json:"merkle_block_height,omitempty"`
	MerkleProof       []byte    `json:"merkle_proof,omitempty"`
	SubjectKind       string    `json:"subject_kind,omitempty"`
	SubjectID         string    `json:"subject_id,omitempty"`
	Actor             string    `json:"actor,omitempty"`
}

// Evidence is the set of graph objects that justified the decision.
// Empty evidence + non-ALLOW decision = the policy fired on absence (e.g., missing verification).
type Evidence struct {
	ClaimIDs          []uuid.UUID `json:"claim_ids,omitempty"`
	EntityIDs         []uuid.UUID `json:"entity_ids,omitempty"`
	ContradictionIDs  []uuid.UUID `json:"contradiction_ids,omitempty"`
	MatchedPredicates []string    `json:"matched_predicates"`
	SnapshotID        *uuid.UUID  `json:"snapshot_id,omitempty"`
	ReasoningNotes    []string    `json:"reasoning_notes,omitempty"`
}
