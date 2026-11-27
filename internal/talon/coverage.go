// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package talon

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Coverage verdicts. A guard verdict of ALLOW only means "nothing found";
// the coverage verdict says how much of the change Garuda could actually see.
const (
	CoverageComplete     = "COMPLETE"
	CoveragePartial      = "PARTIAL"
	CoverageInconclusive = "INCONCLUSIVE"
)

// EvidenceCoverage reports what the guard analyzed and what it could not.
// It is additive to GuardVerdict: existing fields keep their meaning.
type EvidenceCoverage struct {
	ChangedFilesTotal    int            `json:"changed_files_total"`
	ChangedFilesAnalyzed int            `json:"changed_files_analyzed"`
	LanguagesAnalyzed    []string       `json:"languages_analyzed"`
	UnresolvedFiles      []string       `json:"unresolved_files"`
	UnsupportedCodeFiles []string       `json:"unsupported_code_files"`
	NonCodeFiles         int            `json:"non_code_files"`
	GraphAvailable       bool           `json:"graph_available"`
	TouchedEntities      int            `json:"touched_entities"`
	ImpactedEntities     int            `json:"impacted_entities"`
	PoliciesEvaluated    int            `json:"policies_evaluated"`
	PolicyDecisions      map[string]int `json:"policy_decisions"`
	Verdict              string         `json:"verdict"`
	Gaps                 []string       `json:"gaps"`
}

// CoverageInputs is everything ComputeEvidenceCoverage needs, as plain
// values, so the computation has no database or git dependency.
type CoverageInputs struct {
	DiffError        error    // reading the diff failed: nothing can be claimed about the change
	ChangedFiles     []string // every file in the diff
	ResolvedFiles    []string // changed files that matched entities in the graph
	GraphAvailable   bool     // the diff was resolved against the graph and the impact index loaded
	TouchedEntities  int
	ImpactedEntities int
	PolicyDecisions  []string // one decision per evaluated policy
}

// analyzedLanguages maps the source extensions Garuda's analyzers index.
// Keep in sync with diff.isSupportedSourceFile.
var analyzedLanguages = map[string]string{
	".go":  "go",
	".py":  "python",
	".ts":  "typescript",
	".tsx": "typescript",
	".js":  "javascript",
	".jsx": "javascript",
}

// unanalyzedCodeExts are source languages Garuda does not index yet. A change
// in one of these is code the guard did not look at, so coverage is PARTIAL.
// Docs, config and data files are counted but do not lower coverage.
var unanalyzedCodeExts = map[string]bool{
	".java": true, ".kt": true, ".kts": true, ".scala": true, ".groovy": true,
	".rs": true, ".rb": true, ".php": true, ".cs": true, ".fs": true,
	".c": true, ".h": true, ".cc": true, ".cpp": true, ".cxx": true, ".hpp": true,
	".m": true, ".mm": true, ".swift": true, ".dart": true, ".lua": true,
	".ex": true, ".exs": true, ".erl": true, ".clj": true, ".hs": true,
	".sh": true, ".bash": true, ".ps1": true, ".sql": true, ".vue": true, ".svelte": true,
}

func normalizeChangedPath(p string) string {
	return filepath.ToSlash(strings.TrimPrefix(strings.TrimSpace(p), "./"))
}

func ComputeEvidenceCoverage(in CoverageInputs) *EvidenceCoverage {
	cov := &EvidenceCoverage{
		LanguagesAnalyzed:    []string{},
		UnresolvedFiles:      []string{},
		UnsupportedCodeFiles: []string{},
		GraphAvailable:       in.GraphAvailable,
		TouchedEntities:      in.TouchedEntities,
		ImpactedEntities:     in.ImpactedEntities,
		PolicyDecisions:      map[string]int{},
		Gaps:                 []string{},
	}

	resolved := make(map[string]bool, len(in.ResolvedFiles))
	for _, f := range in.ResolvedFiles {
		resolved[normalizeChangedPath(f)] = true
	}

	seen := make(map[string]bool, len(in.ChangedFiles))
	langs := map[string]bool{}
	sourceFiles := 0
	for _, raw := range in.ChangedFiles {
		f := normalizeChangedPath(raw)
		if f == "" || seen[f] {
			continue
		}
		seen[f] = true
		cov.ChangedFilesTotal++

		ext := strings.ToLower(filepath.Ext(f))
		lang, analyzable := analyzedLanguages[ext]
		switch {
		case analyzable:
			sourceFiles++
			if in.GraphAvailable && resolved[f] {
				cov.ChangedFilesAnalyzed++
				langs[lang] = true
			} else {
				cov.UnresolvedFiles = append(cov.UnresolvedFiles, f)
			}
		case unanalyzedCodeExts[ext]:
			cov.UnsupportedCodeFiles = append(cov.UnsupportedCodeFiles, f)
		default:
			cov.NonCodeFiles++
		}
	}
	for l := range langs {
		cov.LanguagesAnalyzed = append(cov.LanguagesAnalyzed, l)
	}
	sort.Strings(cov.LanguagesAnalyzed)
	sort.Strings(cov.UnresolvedFiles)
	sort.Strings(cov.UnsupportedCodeFiles)

	inconclusivePolicies := 0
	for _, d := range in.PolicyDecisions {
		d = strings.ToUpper(strings.TrimSpace(d))
		if d == "" {
			continue
		}
		cov.PoliciesEvaluated++
		cov.PolicyDecisions[d]++
		if d == CoverageInconclusive {
			inconclusivePolicies++
		}
	}

	// Gaps, most serious first.
	if in.DiffError != nil {
		cov.Gaps = append(cov.Gaps, fmt.Sprintf("the diff could not be read (%v), so the changed files are unknown", in.DiffError))
		cov.Verdict = CoverageInconclusive
		return cov
	}
	if !in.GraphAvailable && sourceFiles > 0 {
		cov.Gaps = append(cov.Gaps, "the code graph was not available, so call paths and blast radius were not checked; run `garuda index` and retry")
	}
	if in.GraphAvailable && len(cov.UnresolvedFiles) > 0 {
		cov.Gaps = append(cov.Gaps, fmt.Sprintf("%d changed source file(s) are not in the indexed graph (new or moved since the last index); re-index to analyze them", len(cov.UnresolvedFiles)))
	}
	if len(cov.UnsupportedCodeFiles) > 0 {
		cov.Gaps = append(cov.Gaps, fmt.Sprintf("%d changed file(s) are in languages Garuda does not analyze yet", len(cov.UnsupportedCodeFiles)))
	}
	if inconclusivePolicies > 0 {
		cov.Gaps = append(cov.Gaps, fmt.Sprintf("%d of %d policies were INCONCLUSIVE: the evidence they need was missing", inconclusivePolicies, cov.PoliciesEvaluated))
	}

	cov.Verdict = coverageVerdict(cov, sourceFiles, inconclusivePolicies)
	return cov
}

func coverageVerdict(cov *EvidenceCoverage, sourceFiles, inconclusivePolicies int) string {
	// Nothing Garuda can or should analyze changed: coverage is trivially complete.
	if sourceFiles == 0 && len(cov.UnsupportedCodeFiles) == 0 {
		if cov.PoliciesEvaluated > 0 && inconclusivePolicies == cov.PoliciesEvaluated {
			return CoverageInconclusive
		}
		if inconclusivePolicies > 0 {
			return CoveragePartial
		}
		return CoverageComplete
	}
	// Code changed but none of it could be analyzed.
	if cov.ChangedFilesAnalyzed == 0 {
		return CoverageInconclusive
	}
	if cov.PoliciesEvaluated > 0 && inconclusivePolicies == cov.PoliciesEvaluated {
		return CoverageInconclusive
	}
	if len(cov.UnresolvedFiles) > 0 || len(cov.UnsupportedCodeFiles) > 0 || inconclusivePolicies > 0 {
		return CoveragePartial
	}
	return CoverageComplete
}

// SummaryLines renders the coverage for terminal output: one headline and
// one line per gap.
func (c *EvidenceCoverage) SummaryLines() []string {
	if c == nil {
		return nil
	}
	head := fmt.Sprintf("%s (%d of %d changed file(s) analyzed", c.Verdict, c.ChangedFilesAnalyzed, c.ChangedFilesTotal)
	if len(c.LanguagesAnalyzed) > 0 {
		head += ": " + strings.Join(c.LanguagesAnalyzed, ", ")
	}
	head += ")"
	lines := []string{head}
	for _, g := range c.Gaps {
		lines = append(lines, "gap: "+g)
	}
	return lines
}
