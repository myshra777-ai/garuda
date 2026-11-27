// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package talon

import (
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestCoverage_AllAnalyzedIsComplete(t *testing.T) {
	cov := ComputeEvidenceCoverage(CoverageInputs{
		ChangedFiles:     []string{"internal/pay/refund.go", "./web/app.tsx", "README.md"},
		ResolvedFiles:    []string{"internal/pay/refund.go", "web/app.tsx"},
		GraphAvailable:   true,
		TouchedEntities:  3,
		ImpactedEntities: 12,
		PolicyDecisions:  []string{"ALLOW", "WARN"},
	})
	if cov.Verdict != CoverageComplete {
		t.Fatalf("verdict = %s, want COMPLETE; gaps: %v", cov.Verdict, cov.Gaps)
	}
	if cov.ChangedFilesTotal != 3 || cov.ChangedFilesAnalyzed != 2 || cov.NonCodeFiles != 1 {
		t.Fatalf("counts wrong: %+v", cov)
	}
	if !reflect.DeepEqual(cov.LanguagesAnalyzed, []string{"go", "typescript"}) {
		t.Fatalf("languages = %v", cov.LanguagesAnalyzed)
	}
	if cov.PoliciesEvaluated != 2 || cov.PolicyDecisions["ALLOW"] != 1 || cov.PolicyDecisions["WARN"] != 1 {
		t.Fatalf("policy counts wrong: %+v", cov.PolicyDecisions)
	}
	if len(cov.Gaps) != 0 {
		t.Fatalf("a complete run should report no gaps: %v", cov.Gaps)
	}
}

func TestCoverage_FileMissingFromGraphIsPartial(t *testing.T) {
	cov := ComputeEvidenceCoverage(CoverageInputs{
		ChangedFiles:   []string{"internal/pay/refund.go", "internal/pay/new_flow.go"},
		ResolvedFiles:  []string{"internal/pay/refund.go"},
		GraphAvailable: true,
	})
	if cov.Verdict != CoveragePartial {
		t.Fatalf("verdict = %s, want PARTIAL", cov.Verdict)
	}
	if !reflect.DeepEqual(cov.UnresolvedFiles, []string{"internal/pay/new_flow.go"}) {
		t.Fatalf("unresolved = %v", cov.UnresolvedFiles)
	}
	if len(cov.Gaps) != 1 || !strings.Contains(cov.Gaps[0], "re-index") {
		t.Fatalf("gap should tell the user to re-index: %v", cov.Gaps)
	}
}

func TestCoverage_UnsupportedLanguageIsPartial(t *testing.T) {
	cov := ComputeEvidenceCoverage(CoverageInputs{
		ChangedFiles:   []string{"svc/main.go", "billing/Invoice.java", "db/001_init.sql"},
		ResolvedFiles:  []string{"svc/main.go"},
		GraphAvailable: true,
	})
	if cov.Verdict != CoveragePartial {
		t.Fatalf("verdict = %s, want PARTIAL", cov.Verdict)
	}
	if !reflect.DeepEqual(cov.UnsupportedCodeFiles, []string{"billing/Invoice.java", "db/001_init.sql"}) {
		t.Fatalf("unsupported = %v", cov.UnsupportedCodeFiles)
	}
}

func TestCoverage_NoGraphIsInconclusive(t *testing.T) {
	cov := ComputeEvidenceCoverage(CoverageInputs{
		ChangedFiles:    []string{"svc/main.go"},
		ResolvedFiles:   []string{"svc/main.go"}, // ignored without a graph
		GraphAvailable:  false,
		PolicyDecisions: []string{"ALLOW"},
	})
	if cov.Verdict != CoverageInconclusive {
		t.Fatalf("verdict = %s, want INCONCLUSIVE", cov.Verdict)
	}
	if cov.ChangedFilesAnalyzed != 0 || len(cov.UnresolvedFiles) != 1 {
		t.Fatalf("without a graph nothing counts as analyzed: %+v", cov)
	}
	if len(cov.Gaps) == 0 || !strings.Contains(cov.Gaps[0], "garuda index") {
		t.Fatalf("first gap should explain the missing graph: %v", cov.Gaps)
	}
}

func TestCoverage_OnlyUnsupportedCodeIsInconclusive(t *testing.T) {
	cov := ComputeEvidenceCoverage(CoverageInputs{
		ChangedFiles:   []string{"src/lib.rs"},
		GraphAvailable: true,
	})
	if cov.Verdict != CoverageInconclusive {
		t.Fatalf("verdict = %s, want INCONCLUSIVE: code changed and none of it was analyzed", cov.Verdict)
	}
}

func TestCoverage_InconclusivePolicies(t *testing.T) {
	base := CoverageInputs{
		ChangedFiles:   []string{"svc/main.go"},
		ResolvedFiles:  []string{"svc/main.go"},
		GraphAvailable: true,
	}

	some := base
	some.PolicyDecisions = []string{"ALLOW", "INCONCLUSIVE"}
	if v := ComputeEvidenceCoverage(some).Verdict; v != CoveragePartial {
		t.Fatalf("one inconclusive policy: verdict = %s, want PARTIAL", v)
	}

	all := base
	all.PolicyDecisions = []string{"INCONCLUSIVE", "inconclusive"}
	cov := ComputeEvidenceCoverage(all)
	if cov.Verdict != CoverageInconclusive {
		t.Fatalf("all policies inconclusive: verdict = %s, want INCONCLUSIVE", cov.Verdict)
	}
	if cov.PolicyDecisions["INCONCLUSIVE"] != 2 {
		t.Fatalf("decisions should be normalized to upper case: %v", cov.PolicyDecisions)
	}
}

func TestCoverage_DocsOnlyChangeIsComplete(t *testing.T) {
	cov := ComputeEvidenceCoverage(CoverageInputs{
		ChangedFiles:    []string{"docs/adr/0007.md", "config/app.yaml"},
		GraphAvailable:  false, // irrelevant: no code changed
		PolicyDecisions: []string{"ALLOW"},
	})
	if cov.Verdict != CoverageComplete || cov.NonCodeFiles != 2 || len(cov.Gaps) != 0 {
		t.Fatalf("docs-only change: %+v", cov)
	}
}

func TestCoverage_EmptyDiff(t *testing.T) {
	cov := ComputeEvidenceCoverage(CoverageInputs{GraphAvailable: true})
	if cov.Verdict != CoverageComplete || cov.ChangedFilesTotal != 0 {
		t.Fatalf("empty diff: %+v", cov)
	}
}

func TestCoverage_DiffErrorIsInconclusive(t *testing.T) {
	cov := ComputeEvidenceCoverage(CoverageInputs{
		DiffError:       errors.New("not a git repository"),
		GraphAvailable:  true,
		PolicyDecisions: []string{"ALLOW"},
	})
	if cov.Verdict != CoverageInconclusive {
		t.Fatalf("verdict = %s, want INCONCLUSIVE", cov.Verdict)
	}
	if len(cov.Gaps) != 1 || !strings.Contains(cov.Gaps[0], "not a git repository") {
		t.Fatalf("gap should carry the diff error: %v", cov.Gaps)
	}
}

func TestCoverage_DuplicatesAndPathForms(t *testing.T) {
	cov := ComputeEvidenceCoverage(CoverageInputs{
		ChangedFiles:   []string{"./a/x.go", "a/x.go", " a/x.go ", ""},
		ResolvedFiles:  []string{"./a/x.go"},
		GraphAvailable: true,
	})
	if cov.ChangedFilesTotal != 1 || cov.ChangedFilesAnalyzed != 1 {
		t.Fatalf("the same file in different spellings must count once: %+v", cov)
	}
}

func TestCoverage_ExtensionTableMatchesAnalyzers(t *testing.T) {
	// diff.isSupportedSourceFile decides which files reach the graph. If a
	// language is added there, it must be added to analyzedLanguages too.
	got := make([]string, 0, len(analyzedLanguages))
	for ext := range analyzedLanguages {
		got = append(got, ext)
		if unanalyzedCodeExts[ext] {
			t.Errorf("%s is listed as both analyzed and unanalyzed", ext)
		}
	}
	sort.Strings(got)
	want := []string{".go", ".js", ".jsx", ".py", ".ts", ".tsx"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("analyzed extensions = %v, want %v", got, want)
	}
}

func TestCoverage_JSONShape(t *testing.T) {
	// Contract: these keys are part of `check --json` and
	// `talon_verify_edit` output. Renaming one breaks CI integrations.
	raw, err := json.Marshal(ComputeEvidenceCoverage(CoverageInputs{}))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"changed_files_analyzed", "changed_files_total", "gaps", "graph_available",
		"impacted_entities", "languages_analyzed", "non_code_files", "policies_evaluated",
		"policy_decisions", "touched_entities", "unresolved_files", "unsupported_code_files", "verdict",
	}
	got := make([]string, 0, len(m))
	for k := range m {
		got = append(got, k)
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("evidence_coverage keys changed:\n got  %v\n want %v", got, want)
	}
	for _, k := range []string{"gaps", "languages_analyzed", "unresolved_files", "unsupported_code_files"} {
		if _, isList := m[k].([]any); !isList {
			t.Errorf("%s must encode as a JSON array, never null", k)
		}
	}
}

func TestCoverage_SummaryLines(t *testing.T) {
	var nilCov *EvidenceCoverage
	if nilCov.SummaryLines() != nil {
		t.Fatal("nil coverage should render nothing")
	}
	cov := ComputeEvidenceCoverage(CoverageInputs{
		ChangedFiles:   []string{"a.go", "b.go"},
		ResolvedFiles:  []string{"a.go"},
		GraphAvailable: true,
	})
	lines := cov.SummaryLines()
	if len(lines) != 2 || lines[0] != "PARTIAL (1 of 2 changed file(s) analyzed: go)" || !strings.HasPrefix(lines[1], "gap: ") {
		t.Fatalf("summary lines = %q", lines)
	}
}
