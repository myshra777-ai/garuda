// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package impact

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/store"
)

// ent builds entity metadata for test graphs. The ID doubles as the name
// so traversal chains read naturally in assertions.
func ent(id, kind string) store.EntityMetadata {
	return store.EntityMetadata{ID: id, Name: id, Package: "pkg", Kind: kind, FilePath: id + ".go"}
}

// consumes records that consumer depends on target (consumer -> target),
// which is the direction ComputeBlastRadius walks backwards.
func consumes(idx *store.ImpactIndex, consumer, target store.EntityMetadata, claim string, conf float64) {
	idx.AddEdge(consumer, target, claim, "OBSERVATION", conf, store.Evidence{FilePath: consumer.FilePath})
}

func newIndex() *store.ImpactIndex {
	return store.NewImpactIndex(uuid.Nil)
}

// affected flattens a result into entityID -> ImpactedEntity.
func affected(t *testing.T, r *BlastRadiusResult) map[string]ImpactedEntity {
	t.Helper()
	out := map[string]ImpactedEntity{}
	for _, list := range r.BySeverity {
		for _, e := range list {
			if _, dup := out[e.EntityID]; dup {
				t.Fatalf("entity %s reported more than once", e.EntityID)
			}
			out[e.EntityID] = e
		}
	}
	if len(out) != r.TotalAffected {
		t.Fatalf("TotalAffected = %d, but %d distinct entities were reported", r.TotalAffected, len(out))
	}
	return out
}

func TestDefaultBlastRadiusConfig(t *testing.T) {
	cfg := DefaultBlastRadiusConfig()
	if cfg.MaxDepth != 3 || cfg.MinConfidence != 0.50 || !cfg.IncludeInferred {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestComputeBlastRadius_UnknownTargetReturnsEmptyResult(t *testing.T) {
	r := ComputeBlastRadius(newIndex(), "missing", DefaultBlastRadiusConfig())
	if r == nil {
		t.Fatal("expected a non-nil result")
	}
	if r.TargetEntityID != "missing" || r.TotalAffected != 0 || len(r.BySeverity) != 0 {
		t.Fatalf("expected an empty result for an unknown target, got %+v", r)
	}
}

func TestComputeBlastRadius_TargetWithNoConsumers(t *testing.T) {
	idx := newIndex()
	target := ent("Target", "FUNCTION")
	other := ent("Other", "FUNCTION")
	// Target depends on Other, so Target exists in the index but nothing consumes it.
	consumes(idx, target, other, "CALLS", 1.0)

	r := ComputeBlastRadius(idx, "Target", DefaultBlastRadiusConfig())
	if r.TotalAffected != 0 {
		t.Fatalf("expected no affected entities, got %d", r.TotalAffected)
	}
}

func TestComputeBlastRadius_TransitiveChainDepthAndConfidence(t *testing.T) {
	idx := newIndex()
	a, b, c := ent("A", "FUNCTION"), ent("B", "FUNCTION"), ent("C", "FUNCTION")
	consumes(idx, b, a, "CALLS", 0.9)
	consumes(idx, c, b, "CALLS", 0.8)

	got := affected(t, ComputeBlastRadius(idx, "A", DefaultBlastRadiusConfig()))
	if len(got) != 2 {
		t.Fatalf("expected B and C to be affected, got %v", got)
	}

	if got["B"].Depth != 1 || got["C"].Depth != 2 {
		t.Fatalf("depths: B=%d C=%d, want 1 and 2", got["B"].Depth, got["C"].Depth)
	}
	if !reflect.DeepEqual(got["C"].TraversalChain, []string{"A", "B", "C"}) {
		t.Fatalf("chain for C = %v, want [A B C]", got["C"].TraversalChain)
	}
	if got["B"].CumulativeConf != 0.9 {
		t.Fatalf("B confidence = %v, want 0.9", got["B"].CumulativeConf)
	}
	// 0.9 * 0.8 = 0.72, rounded to two decimals.
	if got["C"].CumulativeConf != 0.72 {
		t.Fatalf("C confidence = %v, want 0.72", got["C"].CumulativeConf)
	}
	if got["C"].EpistemicClass != "OBSERVATION" {
		t.Fatalf("epistemic class = %q, want OBSERVATION", got["C"].EpistemicClass)
	}
	if got["B"].Evidence.FilePath != "B.go" {
		t.Fatalf("evidence was not carried from the edge: %+v", got["B"].Evidence)
	}
}

func TestComputeBlastRadius_ChainsDoNotShareBackingArrays(t *testing.T) {
	// Two consumers of the same node must not overwrite each other's chain.
	idx := newIndex()
	a, b := ent("A", "FUNCTION"), ent("B", "FUNCTION")
	c1, c2 := ent("C1", "FUNCTION"), ent("C2", "FUNCTION")
	consumes(idx, b, a, "CALLS", 1.0)
	consumes(idx, c1, b, "CALLS", 1.0)
	consumes(idx, c2, b, "CALLS", 1.0)

	got := affected(t, ComputeBlastRadius(idx, "A", DefaultBlastRadiusConfig()))
	if !reflect.DeepEqual(got["C1"].TraversalChain, []string{"A", "B", "C1"}) {
		t.Fatalf("chain for C1 = %v", got["C1"].TraversalChain)
	}
	if !reflect.DeepEqual(got["C2"].TraversalChain, []string{"A", "B", "C2"}) {
		t.Fatalf("chain for C2 = %v", got["C2"].TraversalChain)
	}
}

func TestComputeBlastRadius_RespectsMaxDepth(t *testing.T) {
	idx := newIndex()
	a, b, c, d := ent("A", "FUNCTION"), ent("B", "FUNCTION"), ent("C", "FUNCTION"), ent("D", "FUNCTION")
	consumes(idx, b, a, "CALLS", 1.0)
	consumes(idx, c, b, "CALLS", 1.0)
	consumes(idx, d, c, "CALLS", 1.0)

	cfg := DefaultBlastRadiusConfig()
	cfg.MaxDepth = 2
	got := affected(t, ComputeBlastRadius(idx, "A", cfg))
	if _, ok := got["D"]; ok {
		t.Fatal("D is at depth 3 and must be excluded when MaxDepth is 2")
	}
	if len(got) != 2 {
		t.Fatalf("expected B and C only, got %d entities", len(got))
	}
}

func TestComputeBlastRadius_ZeroConfigFallsBackToDefaults(t *testing.T) {
	idx := newIndex()
	a, b, c, d, e := ent("A", "FUNCTION"), ent("B", "FUNCTION"), ent("C", "FUNCTION"), ent("D", "FUNCTION"), ent("E", "FUNCTION")
	consumes(idx, b, a, "CALLS", 1.0)
	consumes(idx, c, b, "CALLS", 1.0)
	consumes(idx, d, c, "CALLS", 1.0)
	consumes(idx, e, d, "CALLS", 1.0)
	weak := ent("Weak", "FUNCTION")
	consumes(idx, weak, a, "CALLS", 0.4)

	got := affected(t, ComputeBlastRadius(idx, "A", BlastRadiusConfig{}))
	if _, ok := got["E"]; ok {
		t.Fatal("a zero MaxDepth must fall back to 3, so E at depth 4 is excluded")
	}
	if _, ok := got["D"]; !ok {
		t.Fatal("D at depth 3 should be included with the default depth")
	}
	if _, ok := got["Weak"]; ok {
		t.Fatal("a zero MinConfidence must fall back to 0.50, so a 0.4 edge is excluded")
	}
}

func TestComputeBlastRadius_FiltersLowConfidenceEdges(t *testing.T) {
	idx := newIndex()
	a := ent("A", "FUNCTION")
	strong, weak, edge := ent("Strong", "FUNCTION"), ent("Weak", "FUNCTION"), ent("Edge", "FUNCTION")
	consumes(idx, strong, a, "CALLS", 0.9)
	consumes(idx, weak, a, "CALLS", 0.3)
	consumes(idx, edge, a, "CALLS", 0.5) // exactly at the threshold is kept

	got := affected(t, ComputeBlastRadius(idx, "A", DefaultBlastRadiusConfig()))
	if _, ok := got["Weak"]; ok {
		t.Fatal("an edge below MinConfidence must be ignored")
	}
	if _, ok := got["Edge"]; !ok {
		t.Fatal("an edge exactly at MinConfidence must be kept")
	}
	if _, ok := got["Strong"]; !ok {
		t.Fatal("a strong edge must be kept")
	}
}

func TestComputeBlastRadius_DiamondCountsEachEntityOnce(t *testing.T) {
	// A is consumed by B and C; both are consumed by D.
	idx := newIndex()
	a, b, c, d := ent("A", "FUNCTION"), ent("B", "FUNCTION"), ent("C", "FUNCTION"), ent("D", "FUNCTION")
	consumes(idx, b, a, "CALLS", 1.0)
	consumes(idx, c, a, "CALLS", 1.0)
	consumes(idx, d, b, "CALLS", 1.0)
	consumes(idx, d, c, "CALLS", 1.0)

	r := ComputeBlastRadius(idx, "A", DefaultBlastRadiusConfig())
	got := affected(t, r) // fails if D is reported twice
	if r.TotalAffected != 3 {
		t.Fatalf("TotalAffected = %d, want 3 (B, C, D)", r.TotalAffected)
	}
	if got["D"].Depth != 2 {
		t.Fatalf("D depth = %d, want 2", got["D"].Depth)
	}
}

func TestComputeBlastRadius_KeepsShortestPath(t *testing.T) {
	// D is reachable directly (depth 1) and through B (depth 2).
	idx := newIndex()
	a, b, d := ent("A", "FUNCTION"), ent("B", "FUNCTION"), ent("D", "FUNCTION")
	consumes(idx, b, a, "CALLS", 1.0)
	consumes(idx, d, b, "CALLS", 1.0)
	consumes(idx, d, a, "CALLS", 1.0)

	got := affected(t, ComputeBlastRadius(idx, "A", DefaultBlastRadiusConfig()))
	if got["D"].Depth != 1 {
		t.Fatalf("D depth = %d, want the shortest path of 1", got["D"].Depth)
	}
}

func TestComputeBlastRadius_CycleTerminatesAndExcludesTarget(t *testing.T) {
	idx := newIndex()
	a, b, c := ent("A", "FUNCTION"), ent("B", "FUNCTION"), ent("C", "FUNCTION")
	consumes(idx, b, a, "CALLS", 1.0)
	consumes(idx, c, b, "CALLS", 1.0)
	consumes(idx, a, c, "CALLS", 1.0) // closes the cycle back to the target

	cfg := DefaultBlastRadiusConfig()
	cfg.MaxDepth = 10
	got := affected(t, ComputeBlastRadius(idx, "A", cfg))
	if _, ok := got["A"]; ok {
		t.Fatal("the target must never be reported as affected by itself")
	}
	if len(got) != 2 {
		t.Fatalf("expected B and C, got %d entities", len(got))
	}
}

func TestComputeBlastRadius_CopiesConsumerMetadata(t *testing.T) {
	idx := newIndex()
	a, b := ent("A", "FUNCTION"), ent("B", "FUNCTION")
	consumes(idx, b, a, "CALLS", 1.0)
	got := affected(t, ComputeBlastRadius(idx, "A", DefaultBlastRadiusConfig()))
	if got["B"].Name != "B" || got["B"].FilePath != "B.go" || got["B"].Package != "pkg" {
		t.Fatalf("consumer metadata not copied: %+v", got["B"])
	}
}

func TestComputeBlastRadius_SeverityBuckets(t *testing.T) {
	idx := newIndex()
	target := ent("Target", "FUNCTION")
	route := ent("Route", "HTTP_ROUTE")
	caller := ent("Caller", "FUNCTION")
	test := ent("CallerTest", "TEST")
	consumes(idx, route, target, "CALLS", 1.0)
	consumes(idx, caller, target, "CALLS", 1.0)
	consumes(idx, test, target, "CALLS", 1.0)

	r := ComputeBlastRadius(idx, "Target", DefaultBlastRadiusConfig())
	got := affected(t, r)
	if got["Route"].Severity != SeverityCritical {
		t.Fatalf("route severity = %s, want CRITICAL", got["Route"].Severity)
	}
	if got["Caller"].Severity != SeverityHigh {
		t.Fatalf("caller severity = %s, want HIGH", got["Caller"].Severity)
	}
	if got["CallerTest"].Severity != SeverityLow {
		t.Fatalf("test severity = %s, want LOW", got["CallerTest"].Severity)
	}
	if len(r.BySeverity[SeverityCritical]) != 1 || len(r.BySeverity[SeverityHigh]) != 1 || len(r.BySeverity[SeverityLow]) != 1 {
		t.Fatalf("unexpected buckets: %+v", r.BySeverity)
	}
}

func TestClassifySeverity(t *testing.T) {
	cases := []struct {
		name  string
		kind  string
		claim string
		depth int
		class string
		conf  float64
		want  SeverityLevel
	}{
		{"direct http route", "HTTP_ROUTE", "CALLS", 1, "OBSERVATION", 1.0, SeverityCritical},
		{"direct api endpoint", "api_endpoint", "CALLS", 1, "OBSERVATION", 1.0, SeverityCritical},
		{"direct sql schema", "SQL_SCHEMA", "CALLS", 1, "OBSERVATION", 1.0, SeverityCritical},
		{"direct implementer", "STRUCT", "implements", 1, "OBSERVATION", 1.0, SeverityCritical},
		{"direct function caller", "FUNCTION", "CALLS", 1, "OBSERVATION", 1.0, SeverityHigh},
		{"direct test", "TEST", "CALLS", 1, "OBSERVATION", 1.0, SeverityLow},
		{"depth two route", "HTTP_ROUTE", "CALLS", 2, "OBSERVATION", 1.0, SeverityHigh},
		{"depth two function", "FUNCTION", "CALLS", 2, "OBSERVATION", 0.9, SeverityMedium},
		{"depth two low confidence", "FUNCTION", "CALLS", 2, "OBSERVATION", 0.69, SeverityLow},
		{"depth two at 0.70 threshold", "FUNCTION", "CALLS", 2, "OBSERVATION", 0.70, SeverityMedium},
		{"depth three", "FUNCTION", "CALLS", 3, "OBSERVATION", 1.0, SeverityLow},
		{"inferred direct route", "HTTP_ROUTE", "CALLS", 1, "INFERENCE", 1.0, SeverityMedium},
		{"inferred direct low confidence", "FUNCTION", "CALLS", 1, "INFERENCE", 0.6, SeverityLow},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			edge := store.ConsumerEdge{ClaimType: tc.claim}
			meta := store.EntityMetadata{Kind: tc.kind}
			if got := classifySeverity(edge, meta, tc.depth, tc.class, tc.conf); got != tc.want {
				t.Fatalf("classifySeverity = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestToImpactSummary(t *testing.T) {
	r := &BlastRadiusResult{
		TotalAffected: 7,
		BySeverity: map[SeverityLevel][]ImpactedEntity{
			SeverityCritical: make([]ImpactedEntity, 1),
			SeverityHigh:     make([]ImpactedEntity, 2),
			SeverityMedium:   make([]ImpactedEntity, 3),
			SeverityLow:      make([]ImpactedEntity, 1),
		},
	}
	s := r.ToImpactSummary()
	want := ImpactSummary{
		TotalChanges:  7,
		BreakingCount: 3,
		WarningCount:  4,
		TotalAffected: 7,
		Critical:      1,
		High:          2,
		Medium:        3,
		Low:           1,
	}
	if s != want {
		t.Fatalf("summary = %+v, want %+v", s, want)
	}
}

func TestToImpactSummary_Empty(t *testing.T) {
	r := &BlastRadiusResult{BySeverity: map[SeverityLevel][]ImpactedEntity{}}
	if s := r.ToImpactSummary(); s != (ImpactSummary{}) {
		t.Fatalf("expected a zero summary, got %+v", s)
	}
}

func TestSortImpactedEntities(t *testing.T) {
	in := []ImpactedEntity{
		{Name: "zeta", Severity: SeverityLow},
		{Name: "beta", Severity: SeverityHigh},
		{Name: "alpha", Severity: SeverityLow},
		{Name: "gamma", Severity: SeverityCritical},
		{Name: "delta", Severity: SeverityMedium},
		{Name: "alpha", Severity: SeverityHigh},
	}
	out := SortImpactedEntities(in)
	var got []string
	for _, e := range out {
		got = append(got, string(e.Severity)+":"+e.Name)
	}
	want := []string{"CRITICAL:gamma", "HIGH:alpha", "HIGH:beta", "MEDIUM:delta", "LOW:alpha", "LOW:zeta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}
