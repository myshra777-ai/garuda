// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package impact

import (
	"fmt"
	"path/filepath"
	"testing"
)

func impacted(id, name, file string, sev SeverityLevel, depth int, conf float64) ImpactedEntity {
	return ImpactedEntity{EntityID: id, Name: name, Package: "pay", Kind: "function", FilePath: file, Severity: sev, Depth: depth, CumulativeConf: conf}
}

func result(items ...ImpactedEntity) *BlastRadiusResult {
	r := &BlastRadiusResult{BySeverity: map[SeverityLevel][]ImpactedEntity{}}
	for _, it := range items {
		r.BySeverity[it.Severity] = append(r.BySeverity[it.Severity], it)
	}
	r.TotalAffected = len(items)
	return r
}

func TestSummarizeImpacted_DedupesAndOrders(t *testing.T) {
	root := filepath.FromSlash("/repo")
	abs := func(p string) string { return filepath.Join(root, filepath.FromSlash(p)) }

	a := result(
		impacted("1", "Refund", abs("svc/refund.go"), SeverityMedium, 3, 0.6),
		impacted("2", "Ledger", abs("ledger/ledger.go"), SeverityHigh, 2, 0.8),
	)
	b := result(
		// Entity 1 again, reached more severely through another touched symbol.
		impacted("1", "Refund", abs("svc/refund.go"), SeverityCritical, 1, 0.9),
		impacted("3", "Audit", "audit/log.go", SeverityLow, 4, 0.5),
	)

	got, truncated := SummarizeImpacted([]*BlastRadiusResult{a, nil, b}, root, 0)
	if truncated {
		t.Fatal("should not be truncated")
	}
	if len(got) != 3 {
		t.Fatalf("got %d entities, want 3 (deduplicated): %+v", len(got), got)
	}
	want := []struct {
		name, file string
		sev        SeverityLevel
		depth      int
	}{
		{"Refund", "svc/refund.go", SeverityCritical, 1},
		{"Ledger", "ledger/ledger.go", SeverityHigh, 2},
		{"Audit", "audit/log.go", SeverityLow, 4},
	}
	for i, w := range want {
		g := got[i]
		if g.Name != w.name || g.File != w.file || g.Severity != w.sev || g.Depth != w.depth {
			t.Errorf("entry %d = %+v, want %+v", i, g, w)
		}
	}
	if got[0].Confidence != 0.9 || got[0].Package != "pay" || got[0].Kind != "function" {
		t.Errorf("fields not carried over: %+v", got[0])
	}
}

func TestSummarizeImpacted_TieBreaks(t *testing.T) {
	r := result(
		impacted("1", "B", "b.go", SeverityHigh, 2, 0.5),
		impacted("2", "A", "a.go", SeverityHigh, 2, 0.5),
	)
	// Same entity reached twice at the same severity and depth: higher confidence wins.
	r2 := result(impacted("1", "B", "b.go", SeverityHigh, 2, 0.7))
	got, _ := SummarizeImpacted([]*BlastRadiusResult{r, r2}, "", 0)
	if got[0].Name != "A" || got[1].Name != "B" {
		t.Fatalf("equal severity and depth should sort by name: %+v", got)
	}
	if got[1].Confidence != 0.7 {
		t.Fatalf("the more confident path should win: %+v", got[1])
	}
}

func TestSummarizeImpacted_Limit(t *testing.T) {
	var items []ImpactedEntity
	for i := 0; i < 5; i++ {
		items = append(items, impacted(fmt.Sprint(i), fmt.Sprintf("E%d", i), "x.go", SeverityLow, 1, 0.5))
	}
	got, truncated := SummarizeImpacted([]*BlastRadiusResult{result(items...)}, "", 3)
	if len(got) != 3 || !truncated {
		t.Fatalf("len=%d truncated=%v, want 3 and true", len(got), truncated)
	}
	if got, truncated := SummarizeImpacted(nil, "", 3); len(got) != 0 || truncated {
		t.Fatalf("no results: %v %v", got, truncated)
	}
}

func TestSummarizeImpacted_MissingIDUsesNameAndFile(t *testing.T) {
	r := result(
		impacted("", "Same", "a.go", SeverityLow, 2, 0.5),
		impacted("", "Same", "a.go", SeverityMedium, 1, 0.5),
		impacted("", "Same", "b.go", SeverityLow, 1, 0.5),
	)
	got, _ := SummarizeImpacted([]*BlastRadiusResult{r}, "", 0)
	if len(got) != 2 {
		t.Fatalf("entities without IDs should dedupe by package, name and file: %+v", got)
	}
}

func TestRelativeToRepo(t *testing.T) {
	root := filepath.FromSlash("/repo")
	cases := map[string]string{
		filepath.FromSlash("/repo/a/b.go"): "a/b.go",
		filepath.FromSlash("/other/c.go"):  "/other/c.go",
		filepath.FromSlash("/repo2/c.go"):  "/repo2/c.go",
		"./rel/d.go":                       "rel/d.go",
		"":                                 "",
	}
	for in, want := range cases {
		if got := relativeToRepo(in, root); got != want {
			t.Errorf("relativeToRepo(%q) = %q, want %q", in, got, want)
		}
	}
	if got := relativeToRepo(filepath.FromSlash("/repo/a.go"), ""); got != "/repo/a.go" {
		t.Errorf("no root: %q", got)
	}
}

func TestSeverityRank(t *testing.T) {
	if !(SeverityRank(SeverityCritical) > SeverityRank(SeverityHigh) &&
		SeverityRank(SeverityHigh) > SeverityRank(SeverityMedium) &&
		SeverityRank(SeverityMedium) > SeverityRank(SeverityLow) &&
		SeverityRank(SeverityLow) > SeverityRank("BOGUS")) {
		t.Fatal("severity ranks out of order")
	}
}
