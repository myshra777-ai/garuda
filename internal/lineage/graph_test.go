package lineage

import (
	"testing"

	"github.com/techtaytor/garuda/internal/types"
)

func mustID(s string) registry.DecisionID {
	id := registry.DecisionID(s)
	if err := id.Validate(); err != nil {
		panic(err)
	}
	return id
}

func TestGraphAddEdge_Valid(t *testing.T) {
	g := NewGraph(10)
	id1 := mustID("D-001")
	id2 := mustID("D-002")

	if err := g.AddEdge(id1, id2, EdgeDependsOn); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	children := g.GetChildren(id2)
	if len(children) != 1 || children[0] != id1 {
		t.Errorf("expected child D-001, got %v", children)
	}
}

func TestGraphAddEdge_CycleDetection(t *testing.T) {
	g := NewGraph(10)
	id1 := mustID("D-001")
	id2 := mustID("D-002")
	id3 := mustID("D-003")

	// 1→2, 2→3
	g.AddEdge(id1, id2, EdgeDependsOn)
	g.AddEdge(id2, id3, EdgeDependsOn)

	// 3→1 would create a cycle
	if err := g.AddEdge(id3, id1, EdgeDependsOn); err == nil {
		t.Error("expected cycle error, got nil")
	}
}

func TestGraphAddEdge_MaxNodes(t *testing.T) {
	g := NewGraph(3) // max 3 nodes
	id1 := mustID("D-001")
	id2 := mustID("D-002")
	id3 := mustID("D-003")
	id4 := mustID("D-004")

	// First edge: add 2 nodes → should work
	if err := g.AddEdge(id1, id2, EdgeDependsOn); err != nil {
		t.Fatalf("first edge failed: %v", err)
	}

	// Second edge: add 3rd node → should work
	if err := g.AddEdge(id2, id3, EdgeDependsOn); err != nil {
		t.Fatalf("second edge failed: %v", err)
	}

	// Third edge: would add 4th node → should fail
	err := g.AddEdge(id3, id4, EdgeDependsOn)
	if err == nil {
		t.Error("expected max nodes error, got nil")
	} else {
		// This is expected - log the error for debugging
		t.Logf("expected error: %v", err)
	}
}

func TestGraphImpactSet(t *testing.T) {
	g := NewGraph(10)
	id1 := mustID("D-001")
	id2 := mustID("D-002")
	id3 := mustID("D-003")

	// 1 depends on 2, 3 depends on 2
	g.AddEdge(id1, id2, EdgeDependsOn)
	g.AddEdge(id3, id2, EdgeDependsOn)

	impact := g.GetImpactSet(id2)
	if len(impact) != 2 {
		t.Errorf("expected 2 impacted, got %d", len(impact))
	}
}

func TestGraphRemoveDecision(t *testing.T) {
	g := NewGraph(10)
	id1 := mustID("D-001")
	id2 := mustID("D-002")

	g.AddEdge(id1, id2, EdgeDependsOn)

	if err := g.RemoveDecision(id1); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	// Edge should no longer exist
	children := g.GetChildren(id2)
	if len(children) != 0 {
		t.Errorf("expected no children, got %v", children)
	}
}

func TestGraphSupersedingChain(t *testing.T) {
	g := NewGraph(10)
	id1 := mustID("D-001")
	id2 := mustID("D-002")
	id3 := mustID("D-003")

	// 2 supersedes 1, 3 supersedes 2
	g.AddEdge(id2, id1, EdgeSupersedes)
	g.AddEdge(id3, id2, EdgeSupersedes)

	chain := g.GetSupersedingChain(id1)
	expected := []registry.DecisionID{id1, id2, id3}
	if len(chain) != len(expected) {
		t.Errorf("chain length mismatch: got %d, want %d", len(chain), len(expected))
	}
	for i, id := range chain {
		if id != expected[i] {
			t.Errorf("chain[%d] = %v, expected %v", i, id, expected[i])
		}
	}
}
