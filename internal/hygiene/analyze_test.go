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

func TestAnalyzeEmptyInput(t *testing.T) {
	report := Analyze(nil, nil)

	if len(report.Findings) != 0 {
		t.Fatalf("findings = %d, want 0", len(report.Findings))
	}
	if report.EntityCount != 0 {
		t.Fatalf("entity count = %d, want 0", report.EntityCount)
	}
	if report.RelationCount != 0 {
		t.Fatalf("relationship count = %d, want 0", report.RelationCount)
	}
	if report.Summary != (Summary{}) {
		t.Fatalf("summary = %+v, want zero summary", report.Summary)
	}
}

func TestAnalyzeExcludesPackageAndFileCandidates(t *testing.T) {
	entities := []analyzer.Entity{
		{
			ID:   "package",
			Name: "pkg",
			Kind: analyzer.KindPackage,
		},
		{
			ID:   "file",
			Name: "pkg.go",
			Kind: analyzer.KindFile,
		},
	}

	report := Analyze(entities, nil)

	if report.Summary.StaticUnreferencedCandidates != 0 {
		t.Fatalf("static candidates = %d, want 0", report.Summary.StaticUnreferencedCandidates)
	}
}

func TestAnalyzeIgnoresEmptyDuplicateNames(t *testing.T) {
	entities := []analyzer.Entity{
		{
			ID:      "one",
			Name:    "",
			Package: "one",
			Kind:    analyzer.KindFunction,
		},
		{
			ID:      "two",
			Name:    "",
			Package: "two",
			Kind:    analyzer.KindFunction,
		},
	}

	report := Analyze(entities, nil)

	if report.Summary.DuplicateSymbolNames != 0 {
		t.Fatalf("duplicate findings = %d, want 0", report.Summary.DuplicateSymbolNames)
	}
}

func TestAnalyzeDuplicateNameIsAggregated(t *testing.T) {
	entities := []analyzer.Entity{
		{
			ID:      "one",
			Name:    "Shared",
			Package: "zeta",
			Kind:    analyzer.KindFunction,
		},
		{
			ID:      "two",
			Name:    "Shared",
			Package: "alpha",
			Kind:    analyzer.KindFunction,
		},
		{
			ID:      "three",
			Name:    "Shared",
			Package: "alpha",
			Kind:    analyzer.KindMethod,
		},
	}

	report := Analyze(entities, nil)

	if report.Summary.DuplicateSymbolNames != 1 {
		t.Fatalf("duplicate findings = %d, want 1", report.Summary.DuplicateSymbolNames)
	}

	var finding Finding
	for _, candidate := range report.Findings {
		if candidate.Kind == FindingDuplicateSymbolName {
			finding = candidate
			break
		}
	}

	if finding.Name != "Shared" {
		t.Fatalf("duplicate name = %q, want Shared", finding.Name)
	}
	if !reflect.DeepEqual(finding.Packages, []string{"alpha", "zeta"}) {
		t.Fatalf("packages = %v, want [alpha zeta]", finding.Packages)
	}
}

func TestAnalyzeIsIndependentOfInputOrder(t *testing.T) {
	firstEntities := []analyzer.Entity{
		{ID: "z", Name: "Shared", Kind: analyzer.KindFunction, Package: "zeta", File: "z.go"},
		{ID: "a", Name: "Shared", Kind: analyzer.KindFunction, Package: "alpha", File: "a.go"},
		{ID: "c", Name: "ContainsItems", Kind: analyzer.KindFunction, Package: "pkg", File: "c.go"},
	}

	secondEntities := []analyzer.Entity{
		firstEntities[2],
		firstEntities[0],
		firstEntities[1],
	}

	first := Analyze(firstEntities, nil)
	second := Analyze(secondEntities, nil)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("analysis depends on input order:\nfirst=%+v\nsecond=%+v", first, second)
	}
}
