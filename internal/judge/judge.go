package judge

import (
	"fmt"
	"strings"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

type Report struct {
	TotalChanges       int              `json:"total_changes"`
	BreakingCount      int              `json:"breaking_count"`
	DuplicationCount   int              `json:"duplication_count"`
	DeadCodeCount      int              `json:"dead_code_count"`
	BreakingChanges    []BreakingChange `json:"breaking_changes"`
	Duplications       []Duplication    `json:"duplications"`
	DeadCode           []DeadCode       `json:"dead_code"`
	StdLibAlternatives []StdLibAlt      `json:"stdlib_alternatives"`
	Recommendations    []string         `json:"recommendations"`
	Block              bool             `json:"block"`
	BlockReason        string           `json:"block_reason"`
	PassReason         string           `json:"pass_reason"`
}

type BreakingChange struct {
	Entity      string   `json:"entity"`
	ImpactCount int      `json:"impact_count"`
	Mitigation  string   `json:"mitigation"`
	Evidence    []string `json:"evidence"`
}

type Duplication struct {
	NewEntity      string `json:"new_entity"`
	ExistingEntity string `json:"existing_entity"`
	Recommendation string `json:"recommendation"`
}

type DeadCode struct {
	Entity string `json:"entity"`
	File   string `json:"file"`
}

type StdLibAlt struct {
	Entity      string `json:"entity"`
	Alternative string `json:"alternative"`
}

// Judge compares two results and produces a governance report.
func Judge(before, after *analyzer.Result) *Report {
	report := &Report{
		BreakingChanges:    []BreakingChange{},
		Duplications:       []Duplication{},
		DeadCode:           []DeadCode{},
		StdLibAlternatives: []StdLibAlt{},
		Recommendations:    []string{},
	}

	// 1. Diff entities
	diff := analyzer.Diff(before, after)

	// 2. Breaking changes: modified entities with high impact
	for _, ed := range diff.EntityDiffs {
		if ed.Status == "modified" && ed.Impact > 3 {
			report.BreakingChanges = append(report.BreakingChanges, BreakingChange{
				Entity:      ed.Name,
				ImpactCount: ed.Impact,
				Mitigation:  "Add a new method and deprecate the old one with a migration path.",
				Evidence:    []string{fmt.Sprintf("Used by %d references", ed.Impact)},
			})
			report.BreakingCount++
		}
	}

	// 3. Duplications: simplistic name-based matching
	for _, newEntity := range after.Entities {
		for _, oldEntity := range before.Entities {
			if newEntity.Kind == oldEntity.Kind && strings.Contains(newEntity.Name, oldEntity.Name) && newEntity.Name != oldEntity.Name {
				report.Duplications = append(report.Duplications, Duplication{
					NewEntity:      newEntity.Name,
					ExistingEntity: oldEntity.Name,
					Recommendation: fmt.Sprintf("Merge %s with %s", newEntity.Name, oldEntity.Name),
				})
				report.DuplicationCount++
				break
			}
		}
	}

	// 4. Dead code introduced: added entities with zero impact
	for _, ed := range diff.EntityDiffs {
		if ed.Status == "added" && ed.Impact == 0 {
			report.DeadCode = append(report.DeadCode, DeadCode{
				Entity: ed.Name,
				File:   ed.File,
			})
			report.DeadCodeCount++
		}
	}

	// 5. Standard library alternatives
	for _, entity := range after.Entities {
		nameLower := strings.ToLower(entity.Name)
		if strings.Contains(nameLower, "contains") {
			report.StdLibAlternatives = append(report.StdLibAlternatives, StdLibAlt{
				Entity:      entity.Name,
				Alternative: "slices.Contains",
			})
		} else if strings.Contains(nameLower, "sort") {
			report.StdLibAlternatives = append(report.StdLibAlternatives, StdLibAlt{
				Entity:      entity.Name,
				Alternative: "slices.Sort",
			})
		}
	}

	// 6. Recommendations
	if report.BreakingCount > 0 {
		report.Recommendations = append(report.Recommendations, "Rollback breaking changes or add migration path.")
	}
	if report.DuplicationCount > 0 {
		report.Recommendations = append(report.Recommendations, "Refactor to eliminate duplicates.")
	}
	if report.DeadCodeCount > 0 {
		report.Recommendations = append(report.Recommendations, "Remove dead code or add callers.")
	}
	if len(report.StdLibAlternatives) > 0 {
		report.Recommendations = append(report.Recommendations, "Replace custom implementations with standard library.")
	}
	if len(report.Recommendations) == 0 {
		report.Recommendations = append(report.Recommendations, "All good – no issues found.")
	}

	// 7. Block decision
	report.Block = report.BreakingCount > 0 || report.DeadCodeCount > 0
	if report.Block {
		report.BlockReason = "Breaking changes or dead code detected. Manual review required."
	} else {
		report.PassReason = "No blocking issues found."
	}
	report.TotalChanges = len(diff.EntityDiffs)

	return report
}
