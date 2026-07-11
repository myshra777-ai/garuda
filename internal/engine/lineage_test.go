package engine

import (
	"testing"
	"time"

	"github.com/techtaytor/garuda/internal/lineage"
	"github.com/techtaytor/garuda/internal/types"
)

func TestRegisterDecision(t *testing.T) {
	reg := registry.NewRegistry(100)
	graph := lineage.NewGraph(100)
	engine := NewLineageEngine(reg, graph)

	// Create a decision with DRAFT status
	rec := &registry.DecisionRecord{
		ID:         "D-001",
		Decision:   "Use PostgreSQL for financial records",
		Scope:      registry.Scope{Domain: "infrastructure", System: "database"},
		Status:     registry.StatusDraft,
		Owner:      "alice@company.com",
		Approvers:  []string{"bob@company.com"},
		Confidence: 0.9,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	// Append to registry (works because status is DRAFT)
	if err := reg.Append(rec, "alice@company.com"); err != nil {
		t.Fatalf("failed to append to registry: %v", err)
	}

	// Transition through the legal path: DRAFT → REVIEW → APPROVED → CANONICAL
	if err := reg.Transition("D-001", registry.StatusReview, "alice@company.com"); err != nil {
		t.Fatalf("failed to transition to REVIEW: %v", err)
	}
	if err := reg.Transition("D-001", registry.StatusApproved, "alice@company.com"); err != nil {
		t.Fatalf("failed to transition to APPROVED: %v", err)
	}
	if err := reg.Transition("D-001", registry.StatusCanonical, "alice@company.com"); err != nil {
		t.Fatalf("failed to transition to CANONICAL: %v", err)
	}

	// Now register in lineage (works because status is CANONICAL)
	if err := engine.RegisterDecision("D-001"); err != nil {
		t.Fatalf("failed to register lineage: %v", err)
	}

	// Verify it was added
	lineage, err := engine.GetDecisionLineage("D-001")
	if err != nil {
		t.Fatalf("failed to get lineage: %v", err)
	}
	if lineage.DecisionID != "D-001" {
		t.Errorf("expected D-001, got %s", lineage.DecisionID)
	}
}
