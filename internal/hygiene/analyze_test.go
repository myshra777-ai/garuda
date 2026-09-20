// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package hygiene

import (
	"reflect"
	"testing"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

func TestAnalyzeClassifiesAdvisoryFindings(t *testing.T) {
	entities := []analyzer.Entity{
		{ID: "referenced", Name: "Referenced", Kind: analyzer.KindFunction, Package: "pkg", File: "pkg.go"},
		{ID: "candidate", Name: "Candidate", Kind: analyzer.KindFunction, Package: "pkg", File: "pkg.go"},
		{ID: "package", Name: "pkg", Kind: analyzer.KindPackage, Package: "pkg", File: ""},
		{ID: "contains", Name: "ContainsItems", Kind: analyzer.KindFunction, Package: "pkg", File: "pkg.go"},
		{ID: "duplicate-a", Name: "Shared", Kind: analyzer.KindFunction, Package: "one", File: "one.go"},
		{ID: "duplicate-b", Name: "Shared", Kind: analyzer.KindFunction, Package: "two", File: "two.go"},
	}
	relationships := []analyzer.Relationship{{From: "caller", To: "referenced", Type: string(analyzer.RelCalls)}}

	report := Analyze(entities, relationships)

	if report.EntityCount != len(entities) || report.RelationCount != len(relationships) {
		t.Fatalf("unexpected report counts: %+v", report)
	}
	if report.Summary.StaticUnreferencedCandidates == 0 {
		t.Fatal("expected static unreferenced candidate")
	}
	if report.Summary.DuplicateSymbolNames == 0 {
		t.Fatal("expected duplicate symbol-name candidate")
	}
	if report.Summary.DuplicateSymbolNames != 1 {
		t.Fatalf("expected one duplicate-name finding, got %d", report.Summary.DuplicateSymbolNames)
	}

	var duplicateFound bool
	for _, finding := range report.Findings {
		if finding.Kind != FindingDuplicateSymbolName {
			continue
		}
		duplicateFound = true
		if !reflect.DeepEqual(finding.Packages, []string{"one", "two"}) {
			t.Fatalf("unexpected duplicate packages: %v", finding.Packages)
		}
	}
	if !duplicateFound {
		t.Fatal("duplicate-name finding not found")
	}
	if report.Summary.StandardLibraryAlternatives == 0 {
		t.Fatal("expected standard-library alternative")
	}
	for _, finding := range report.Findings {
		if !finding.Advisory {
			t.Fatalf("finding is not advisory: %+v", finding)
		}
		if finding.Kind == FindingStaticUnreferencedCandidate && finding.Name == "Referenced" {
			t.Fatal("referenced entity was classified as unreferenced")
		}
	}
}

func TestAnalyzeIsDeterministic(t *testing.T) {
	entities := []analyzer.Entity{
		{ID: "b", Name: "BContains", Kind: analyzer.KindFunction, Package: "pkg", File: "b.go"},
		{ID: "a", Name: "A", Kind: analyzer.KindFunction, Package: "pkg", File: "a.go"},
	}
	relationships := []analyzer.Relationship(nil)
	first := Analyze(entities, relationships)
	second := Analyze(entities, relationships)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("analysis is not deterministic:\nfirst=%+v\nsecond=%+v", first, second)
	}
}

func TestAnalyzeDoesNotMutateInputs(t *testing.T) {
	entities := []analyzer.Entity{{ID: "a", Name: "A", Kind: analyzer.KindFunction, Package: "pkg"}}
	relationships := []analyzer.Relationship{{From: "x", To: "a", Type: string(analyzer.RelCalls)}}
	entitiesBefore := append([]analyzer.Entity(nil), entities...)
	relationshipsBefore := append([]analyzer.Relationship(nil), relationships...)

	_ = Analyze(entities, relationships)

	if !reflect.DeepEqual(entities, entitiesBefore) {
		t.Fatal("Analyze mutated entities")
	}
	if !reflect.DeepEqual(relationships, relationshipsBefore) {
		t.Fatal("Analyze mutated relationships")
	}
}
