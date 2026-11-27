// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package impact

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/analyzer"
)

func paymentResults() (baseline, proposed *analyzer.Result) {
	baseline = &analyzer.Result{
		Entities: []analyzer.Entity{
			{
				ID: "pay.Charge", Kind: analyzer.KindStruct, Name: "payments.Charge", File: "charge.go",
				Fields: []analyzer.Field{{Name: "ID", Type: "string"}, {Name: "Amount", Type: "int64"}},
			},
			{
				ID: "pay.Gateway", Kind: analyzer.KindStruct, Name: "payments.Gateway", File: "gateway.go",
				Methods: []analyzer.Method{{Name: "Capture"}, {Name: "Refund"}},
			},
			{
				ID: "pay.client", Kind: analyzer.KindStruct, Name: "payments.client", File: "client.go",
				Methods: []analyzer.Method{{Name: "do"}, {Name: "retry"}},
			},
		},
		Relationships: []analyzer.Relationship{
			{From: "pay.Gateway", To: "pay.client", Type: "CALLS"},
		},
	}
	proposed = &analyzer.Result{
		Entities: []analyzer.Entity{
			{
				// Amount removed: breaking.
				ID: "pay.Charge", Kind: analyzer.KindStruct, Name: "payments.Charge", File: "charge.go",
				Fields: []analyzer.Field{{Name: "ID", Type: "string"}},
			},
			{
				// Exported type loses Refund: breaking.
				ID: "pay.Gateway", Kind: analyzer.KindStruct, Name: "payments.Gateway", File: "gateway.go",
				Methods: []analyzer.Method{{Name: "Capture"}},
			},
			{
				// Unexported type loses a method: not reported.
				ID: "pay.client", Kind: analyzer.KindStruct, Name: "payments.client", File: "client.go",
				Methods: []analyzer.Method{{Name: "do"}},
			},
		},
		// The Gateway -> client dependency is removed: a warning.
	}
	return baseline, proposed
}

func TestAnalyzeImpact_FindsBreakingChangesAndWarnings(t *testing.T) {
	baseline, proposed := paymentResults()
	report, err := AnalyzeImpact(context.Background(), nil, uuid.Nil, uuid.Nil, baseline, proposed)
	if err != nil {
		t.Fatalf("AnalyzeImpact: %v", err)
	}

	if len(report.BreakingChanges) != 2 {
		t.Fatalf("breaking changes = %d, want 2: %+v", len(report.BreakingChanges), report.BreakingChanges)
	}
	byEntity := map[string]BreakingChange{}
	for _, bc := range report.BreakingChanges {
		byEntity[bc.Entity] = bc
	}
	if bc := byEntity["payments.Charge"]; bc.ChangeType != "field_removed" || bc.Package != "payments" {
		t.Fatalf("Charge change = %+v, want field_removed in package payments", bc)
	}
	if bc := byEntity["payments.Gateway"]; bc.ChangeType != "method_removed" {
		t.Fatalf("Gateway change = %+v, want method_removed", bc)
	}
	if _, ok := byEntity["payments.client"]; ok {
		t.Fatal("a method removed from an unexported type must not be reported as breaking")
	}

	if len(report.Warnings) != 1 {
		t.Fatalf("warnings = %d, want 1: %+v", len(report.Warnings), report.Warnings)
	}
	w := report.Warnings[0]
	if w.Entity != "pay.Gateway" || !strings.Contains(w.Message, "pay.Gateway -> pay.client") || w.Confidence != 0.7 {
		t.Fatalf("unexpected warning: %+v", w)
	}

	if report.Summary.TotalChanges != 3 || report.Summary.BreakingCount != 2 || report.Summary.WarningCount != 1 {
		t.Fatalf("unexpected summary: %+v", report.Summary)
	}
	if len(report.EvidenceRoot) != 64 {
		t.Fatalf("evidence root should be a sha256 hex digest, got %q", report.EvidenceRoot)
	}
	if _, err := uuid.Parse(report.ID); err != nil {
		t.Fatalf("report ID is not a UUID: %q", report.ID)
	}
}

func TestAnalyzeImpact_EvidenceRootIsDeterministic(t *testing.T) {
	var roots []string
	var ids []string
	for i := 0; i < 10; i++ {
		baseline, proposed := paymentResults()
		report, err := AnalyzeImpact(context.Background(), nil, uuid.Nil, uuid.Nil, baseline, proposed)
		if err != nil {
			t.Fatalf("AnalyzeImpact: %v", err)
		}
		roots = append(roots, report.EvidenceRoot)
		ids = append(ids, report.ID)
	}
	for i := 1; i < len(roots); i++ {
		if roots[i] != roots[0] {
			t.Fatalf("run %d produced evidence root %s, run 0 produced %s: the same inputs must give the same root", i, roots[i], roots[0])
		}
	}
	if ids[0] == ids[1] {
		t.Fatal("each report should still get its own ID")
	}
}

func TestAnalyzeImpact_FindingsAreSorted(t *testing.T) {
	baseline, proposed := paymentResults()
	report, err := AnalyzeImpact(context.Background(), nil, uuid.Nil, uuid.Nil, baseline, proposed)
	if err != nil {
		t.Fatalf("AnalyzeImpact: %v", err)
	}
	if report.BreakingChanges[0].Entity != "payments.Charge" || report.BreakingChanges[1].Entity != "payments.Gateway" {
		t.Fatalf("breaking changes not in stable order: %+v", report.BreakingChanges)
	}
}

func TestAnalyzeImpact_NilInputsGiveEmptyReport(t *testing.T) {
	report, err := AnalyzeImpact(context.Background(), nil, uuid.Nil, uuid.Nil, nil, nil)
	if err != nil {
		t.Fatalf("AnalyzeImpact: %v", err)
	}
	if len(report.BreakingChanges) != 0 || len(report.Warnings) != 0 || report.Summary.TotalChanges != 0 {
		t.Fatalf("expected an empty report, got %+v", report)
	}
	if report.BreakingChanges == nil || report.Warnings == nil || report.AffectedRepos == nil {
		t.Fatal("slices should be empty, not nil, so they encode as [] in JSON")
	}
}

func TestAnalyzeImpact_UnchangedInputsHaveNoFindings(t *testing.T) {
	baseline, _ := paymentResults()
	same, _ := paymentResults()
	report, err := AnalyzeImpact(context.Background(), nil, uuid.Nil, uuid.Nil, baseline, same)
	if err != nil {
		t.Fatalf("AnalyzeImpact: %v", err)
	}
	if len(report.BreakingChanges) != 0 || len(report.Warnings) != 0 || report.Summary.TotalChanges != 0 {
		t.Fatalf("expected no findings for identical snapshots, got %+v", report)
	}
}

func TestAnalyzeEntityChange_AddedFieldsAreNotBreaking(t *testing.T) {
	ed := analyzer.EntityDiff{
		Name:       "payments.Charge",
		Status:     "modified",
		FieldsDiff: &analyzer.FieldsDiff{Added: []analyzer.Field{{Name: "Currency"}}},
	}
	if bc := analyzeEntityChange(ed, nil, nil); bc != nil {
		t.Fatalf("adding a field must not be breaking, got %+v", bc)
	}
}

func TestAnalyzeEntityChange_FieldRemovalTakesPriority(t *testing.T) {
	ed := analyzer.EntityDiff{
		Name:        "payments.Charge",
		EntityID:    "pay.Charge",
		Status:      "modified",
		Impact:      4,
		FieldsDiff:  &analyzer.FieldsDiff{Removed: []analyzer.Field{{Name: "Amount"}}},
		MethodsDiff: &analyzer.MethodsDiff{Removed: []analyzer.Method{{Name: "Total"}}},
	}
	bc := analyzeEntityChange(ed, nil, nil)
	if bc == nil || bc.ChangeType != "field_removed" || bc.ImpactCount != 4 {
		t.Fatalf("got %+v, want a field_removed change with impact 4", bc)
	}
	if len(bc.Evidence) != 1 || !strings.Contains(bc.Evidence[0], "Amount") {
		t.Fatalf("evidence should name the removed field: %v", bc.Evidence)
	}
	if bc.Mitigation == "" {
		t.Fatal("a breaking change should carry a mitigation")
	}
}

func TestExtractPackageFromEntity(t *testing.T) {
	cases := []struct {
		name, id, want string
	}{
		{"github.com/myshra777-ai/garuda/pkg/auth.User", "", "auth"},
		{"payments.Charge", "", "payments"},
		{"Charge", "", "main"},
		{"", "github.com/org/repo/billing.Invoice", "billing"},
		{"", "", "main"},
		{"pkg/auth.User.Method", "", "auth"},
	}
	for _, tc := range cases {
		if got := extractPackageFromEntity(tc.name, tc.id); got != tc.want {
			t.Errorf("extractPackageFromEntity(%q, %q) = %q, want %q", tc.name, tc.id, got, tc.want)
		}
	}
}

func TestIsEntityExported(t *testing.T) {
	cases := map[string]bool{
		"payments.Charge": true,
		"payments.charge": false,
		"Charge":          true,
		"charge":          false,
		"":                false,
		"pkg.Éclair":      true,
		"pkg._hidden":     false,
	}
	for name, want := range cases {
		if got := isEntityExported(name); got != want {
			t.Errorf("isEntityExported(%q) = %v, want %v", name, got, want)
		}
	}
}
