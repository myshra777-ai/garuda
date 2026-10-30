// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package hygiene

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

func formatEntityFileLine(file string, line int) string {
	clean := strings.TrimSpace(file)
	if clean == "" {
		return "unknown:1"
	}
	clean = filepath.ToSlash(clean)
	for _, prefix := range []string{"/home/rohit/garuda-validation/", "/home/rohit/garuda/", "/tmp/garuda/"} {
		if strings.HasPrefix(clean, prefix) {
			clean = strings.TrimPrefix(clean, prefix)
			break
		}
	}
	if line <= 0 {
		line = 1
	}
	return fmt.Sprintf("%s:%d", clean, line)
}

func isNoisySymbolName(name string) bool {
	trimmed := strings.TrimSpace(name)
	return trimmed == "" || strings.HasPrefix(trimmed, ".")
}

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
		if incoming[entity.ID] != 0 ||
			entity.Kind == analyzer.KindPackage ||
			entity.Kind == analyzer.KindFile ||
			string(entity.Kind) == "external" ||
			isNoisySymbolName(entity.Name) {
			continue
		}
		line := entity.LineStart
		if line <= 0 {
			line = entity.Line
		}
		fileLine := formatEntityFileLine(entity.File, line)
		report.Findings = append(report.Findings, Finding{
			Kind:       FindingStaticUnreferencedCandidate,
			EntityID:   entity.ID,
			Name:       entity.Name,
			EntityKind: entity.Kind,
			Package:    entity.Package,
			File:       entity.File,
			LineStart:  line,
			LineEnd:    entity.LineEnd,
			Message:    fmt.Sprintf("%s (%s) has no incoming relationship in the indexed workspace graph", entity.Name, fileLine),
			Reason:     "No incoming graph reference was found; this is not proof of dead code.",
			Advisory:   true,
		})
	}

	byName := make(map[string]map[string]struct{})
	byNameLoc := make(map[string]map[string]string)
	for _, entity := range entities {
		if isNoisySymbolName(entity.Name) || entity.Package == "" || string(entity.Kind) == "external" {
			continue
		}
		if byName[entity.Name] == nil {
			byName[entity.Name] = make(map[string]struct{})
			byNameLoc[entity.Name] = make(map[string]string)
		}
		byName[entity.Name][entity.Package] = struct{}{}
		line := entity.LineStart
		if line <= 0 {
			line = entity.Line
		}
		fl := formatEntityFileLine(entity.File, line)
		if prev, exists := byNameLoc[entity.Name][entity.Package]; !exists || fl < prev {
			byNameLoc[entity.Name][entity.Package] = fl
		}
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

		pkgWithLocs := make([]string, 0, len(packageNames))
		var firstFile string
		var firstLine int
		for i, pkgName := range packageNames {
			loc := byNameLoc[name][pkgName]
			pkgWithLocs = append(pkgWithLocs, fmt.Sprintf("%s (%s)", pkgName, loc))
			if i == 0 {
				firstFile = loc
				firstLine = 1
			}
		}

		report.Findings = append(report.Findings, Finding{
			Kind:      FindingDuplicateSymbolName,
			Name:      name,
			Packages:  packageNames,
			File:      firstFile,
			LineStart: firstLine,
			Message:   fmt.Sprintf("%s appears in multiple packages: %s", name, strings.Join(pkgWithLocs, ", ")),
			Reason:    "Repeated symbol names do not prove duplicated implementation.",
			Advisory:  true,
		})
	}

	for _, entity := range entities {
		if isNoisySymbolName(entity.Name) || string(entity.Kind) == "external" {
			continue
		}
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
		line := entity.LineStart
		if line <= 0 {
			line = entity.Line
		}
		fileLine := formatEntityFileLine(entity.File, line)
		report.Findings = append(report.Findings, Finding{
			Kind:       FindingStandardLibraryAlternative,
			EntityID:   entity.ID,
			Name:       entity.Name,
			EntityKind: entity.Kind,
			Package:    entity.Package,
			File:       entity.File,
			LineStart:  line,
			LineEnd:    entity.LineEnd,
			Message:    fmt.Sprintf("%s (%s) may have a standard-library alternative: %s", entity.Name, fileLine, alternative),
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

