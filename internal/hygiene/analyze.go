// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package hygiene

import (
	"fmt"
	"sort"
	"strings"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

// Analyze produces deterministic advisory findings from typed semantic data.
// It does not mutate the input slices or external state.
func Analyze(entities []analyzer.Entity, relationships []analyzer.Relationship) Report {
	report := Report{
		Findings:      make([]Finding, 0),
		EntityCount:   len(entities),
		RelationCount: len(relationships),
	}

	incoming := make(map[string]int, len(entities))
	for _, relationship := range relationships {
		if relationship.To != "" {
			incoming[relationship.To]++
		}
	}

	for _, entity := range entities {
		if incoming[entity.ID] != 0 || entity.Kind == analyzer.KindPackage || entity.Kind == analyzer.KindFile {
			continue
		}
		report.Findings = append(report.Findings, Finding{
			Kind:       FindingStaticUnreferencedCandidate,
			EntityID:   entity.ID,
			Name:       entity.Name,
			EntityKind: entity.Kind,
			Package:    entity.Package,
			File:       entity.File,
			LineStart:  entity.LineStart,
			LineEnd:    entity.LineEnd,
			Message:    fmt.Sprintf("%s has no incoming relationship in the indexed workspace graph", entity.Name),
			Reason:     "No incoming graph reference was found; this is not proof of dead code.",
			Advisory:   true,
		})
	}

	byName := make(map[string]map[string]struct{})
	for _, entity := range entities {
		if entity.Name == "" || entity.Package == "" {
			continue
		}
		if byName[entity.Name] == nil {
			byName[entity.Name] = make(map[string]struct{})
		}
		byName[entity.Name][entity.Package] = struct{}{}
	}
	for name, packages := range byName {
		if len(packages) < 2 {
			continue
		}

		packageNames := make([]string, 0, len(packages))
		for packageName := range packages {
			packageNames = append(packageNames, packageName)
		}
		sort.Strings(packageNames)

		report.Findings = append(report.Findings, Finding{
			Kind:     FindingDuplicateSymbolName,
			Name:     name,
			Packages: packageNames,
			Message:  fmt.Sprintf("%s appears in multiple packages: %s", name, strings.Join(packageNames, ", ")),
			Reason:   "Repeated symbol names do not prove duplicated implementation.",
			Advisory: true,
		})
	}

	for _, entity := range entities {
		alternative := ""
		switch {
		case strings.Contains(entity.Name, "Contains"):
			alternative = "slices.Contains"
		case strings.Contains(entity.Name, "Sort"):
			alternative = "slices.Sort"
		}
		if alternative == "" {
			continue
		}
		report.Findings = append(report.Findings, Finding{
			Kind:       FindingStandardLibraryAlternative,
			EntityID:   entity.ID,
			Name:       entity.Name,
			EntityKind: entity.Kind,
			Package:    entity.Package,
			File:       entity.File,
			LineStart:  entity.LineStart,
			LineEnd:    entity.LineEnd,
			Message:    fmt.Sprintf("%s may have a standard-library alternative: %s", entity.Name, alternative),
			Reason:     "Advisory naming heuristic; inspect implementation before changing code.",
			Advisory:   true,
		})
	}

	sortFindings(report.Findings)
	for _, finding := range report.Findings {
		switch finding.Kind {
		case FindingStaticUnreferencedCandidate:
			report.Summary.StaticUnreferencedCandidates++
		case FindingDuplicateSymbolName:
			report.Summary.DuplicateSymbolNames++
		case FindingStandardLibraryAlternative:
			report.Summary.StandardLibraryAlternatives++
		}
	}
	return report
}
