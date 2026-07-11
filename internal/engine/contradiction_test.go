package engine

import (
	"strings"
	"testing"

	"github.com/techtaytor/garuda/internal/lineage"
	"github.com/techtaytor/garuda/internal/types"
)

func TestDirectConflictDetection(t *testing.T) {
	reg := registry.NewRegistry(100)
	graph := lineage.NewGraph(100)
	engine := NewContradictionEngine(reg, graph)

	// Create root reference canonical decision
	rec1 := &registry.DecisionRecord{
		ID:           "D-001",
		Decision:     "use postgresql for financial records",
		Scope:        registry.Scope{Domain: "infrastructure", System: "database"},
		Status:       registry.StatusDraft, // Must start as draft for Append invariant rules
		Owner:        "alice@company.com",
		Approvers:    []string{"bob@company.com"},
		Confidence:   0.9,
		Rationale:    "ACID compliance requirement.",
		Dependencies: []registry.DecisionID{},
	}

	// Commit genesis fact to registry layer safely
	if err := reg.Append(rec1, "alice@company.com"); err != nil {
		t.Fatalf("failed to setup test fixture: %v", err)
	}
	// Migrate status cleanly through valid state machine path
	if err := reg.Transition("D-001", registry.StatusReview, "alice@company.com"); err != nil {
		t.Fatalf("failed to cycle state to review: %v", err)
	}
	if err := reg.Transition("D-001", registry.StatusApproved, "bob@company.com"); err != nil {
		t.Fatalf("failed to cycle state to approved: %v", err)
	}
	if err := reg.Transition("D-001", registry.StatusCanonical, "bob@company.com"); err != nil {
		t.Fatalf("failed to finalize state to canonical: %v", err)
	}

	// Create competing candidate that explicitly clashes
	rec2 := &registry.DecisionRecord{
		ID:           "D-002",
		Decision:     "use mongodb for financial records",
		Scope:        registry.Scope{Domain: "infrastructure", System: "database"},
		Status:       registry.StatusDraft,
		Owner:        "alice@company.com",
		Approvers:    []string{"bob@company.com"},
		Confidence:   0.8,
		Rationale:    "NoSQL performance needs.",
		Dependencies: []registry.DecisionID{},
	}

	contradictions, err := engine.ValidateDecision(rec2)
	if err != nil {
		t.Fatalf("unexpected engine execution failure: %v", err)
	}
	if len(contradictions) == 0 {
		t.Error("security loop failure: expected direct conflict contradiction, detected none")
	}
}

func TestScopeOverlapDetection(t *testing.T) {
	reg := registry.NewRegistry(100)
	graph := lineage.NewGraph(100)
	engine := NewContradictionEngine(reg, graph)

	rec1 := &registry.DecisionRecord{
		ID:           "D-001",
		Decision:     "use postgresql database system layer",
		Scope:        registry.Scope{Domain: "infrastructure", System: "database"},
		Status:       registry.StatusDraft,
		Owner:        "alice@company.com",
		Approvers:    []string{"bob@company.com"},
		Confidence:   0.9,
		Rationale:    "Standard stack agreement.",
		Dependencies: []registry.DecisionID{},
	}

	if err := reg.Append(rec1, "alice@company.com"); err != nil {
		t.Fatalf("failed to setup base test fixture: %v", err)
	}
	if err := reg.Transition("D-001", registry.StatusReview, "alice@company.com"); err != nil {
		t.Fatalf("failed test migration: %v", err)
	}
	if err := reg.Transition("D-001", registry.StatusApproved, "bob@company.com"); err != nil {
		t.Fatalf("failed test migration: %v", err)
	}
	if err := reg.Transition("D-001", registry.StatusCanonical, "bob@company.com"); err != nil {
		t.Fatalf("failed test migration: %v", err)
	}

	rec2 := &registry.DecisionRecord{
		ID:           "D-002",
		Decision:     "use postgresql as well for matching scope data mapping",
		Scope:        registry.Scope{Domain: "infrastructure", System: "database"},
		Status:       registry.StatusDraft,
		Owner:        "alice@company.com",
		Approvers:    []string{"bob@company.com"},
		Confidence:   0.8,
		Rationale:    "Duplicate boundary declaration.",
		Dependencies: []registry.DecisionID{},
	}

	contradictions, err := engine.ValidateDecision(rec2)
	if err != nil {
		t.Fatalf("unexpected engine execution failure: %v", err)
	}
	if len(contradictions) == 0 {
		t.Error("governance leak: expected boundary scope overlap error block, detected none")
	}
}

func TestConstraintViolation(t *testing.T) {
	reg := registry.NewRegistry(100)
	graph := lineage.NewGraph(100)
	engine := NewContradictionEngine(reg, graph)

	engine.AddConstraint(Constraint{
		ID:          "no-mongodb",
		Description: "Strict organizational constraint rejecting MongoDB text components",
		Predicate: func(rec *registry.DecisionRecord) bool {
			return strings.Contains(normalize(rec.Decision), "mongodb")
		},
	})

	rec := &registry.DecisionRecord{
		ID:           "D-001",
		Decision:     "use mongodb for all backend data layers",
		Scope:        registry.Scope{Domain: "infrastructure"},
		Status:       registry.StatusDraft,
		Owner:        "alice@company.com",
		Approvers:    []string{"bob@company.com"},
		Confidence:   1.0,
		Rationale:    "Forced testing violation case.",
		Dependencies: []registry.DecisionID{},
	}

	contradictions, err := engine.ValidateDecision(rec)
	if err != nil {
		t.Fatalf("unexpected checking error: %v", err)
	}
	if len(contradictions) != 1 || contradictions[0].Type != ConstraintViolation {
		t.Errorf("policy failure: expected unique constraint violation frame, got: %v", contradictions)
	}
}
