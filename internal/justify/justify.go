// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package justify

import (
	"fmt"
	"strings"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

type Justification struct {
	EntityName        string         `json:"entity_name"`
	File              string         `json:"file"`
	Necessity         Necessity      `json:"necessity"`
	Simplicity        Simplicity     `json:"simplicity"`
	StdLib            StdLib         `json:"std_lib"`
	Duplication       Duplication    `json:"duplication"`
	ContractImpact    ContractImpact `json:"contract_impact"`
	Evidence          []string       `json:"evidence"`
	Conclusion        string         `json:"conclusion"`
	OverallConfidence float64        `json:"overall_confidence"`
}

type Necessity struct {
	Verified    bool    `json:"verified"`
	Consumers   int     `json:"consumers"`
	LastChanged string  `json:"last_changed"`
	Confidence  float64 `json:"confidence"`
}

type Simplicity struct {
	Pass       bool   `json:"pass"`
	Lines      int    `json:"lines"`
	Suggestion string `json:"suggestion"`
}

type StdLib struct {
	Pass        bool    `json:"pass"`
	Alternative string  `json:"alternative"`
	Confidence  float64 `json:"confidence"`
}

type Duplication struct {
	Pass        bool    `json:"pass"`
	DuplicateOf string  `json:"duplicate_of"`
	Confidence  float64 `json:"confidence"`
}

type ContractImpact struct {
	Consumers  int      `json:"consumers"`
	IsBreaking bool     `json:"is_breaking"`
	Details    string   `json:"details"`
	Evidence   []string `json:"evidence"`
}

// Justify builds a justification for an entity using incoming/outgoing relationships.
func Justify(entity *analyzer.Entity, incoming, outgoing []analyzer.Relationship) *Justification {
	// Necessity
	consumerCount := len(incoming)
	necessity := Necessity{
		Verified:    consumerCount > 0,
		Consumers:   consumerCount,
		LastChanged: "unknown", // could be extended with git info
		Confidence:  0.9,
	}
	if consumerCount == 0 {
		necessity.Confidence = 0.1
	}

	// Simplicity (rough estimate)
	lines := len(entity.Fields)*3 + len(entity.Methods)*5 + 10
	simplicity := Simplicity{
		Pass:       lines < 50,
		Lines:      lines,
		Suggestion: "",
	}
	if lines > 100 {
		simplicity.Suggestion = "Consider refactoring into smaller functions."
	}

	// Standard library alternatives (simple pattern matching)
	stdLib := StdLib{
		Pass:        true,
		Alternative: "",
		Confidence:  0.7,
	}
	nameLower := strings.ToLower(entity.Name)
	if strings.Contains(nameLower, "contains") {
		stdLib.Alternative = "slices.Contains"
		stdLib.Pass = false
		stdLib.Confidence = 0.85
	} else if strings.Contains(nameLower, "sort") {
		stdLib.Alternative = "slices.Sort"
		stdLib.Pass = false
		stdLib.Confidence = 0.85
	}

	// Duplication (simplistic – same name in different packages)
	dup := Duplication{
		Pass:        true,
		DuplicateOf: "",
		Confidence:  0.9,
	}
	// In a real implementation, we would query the graph for similar entities.

	// Contract impact
	breaking := false
	impactDetails := ""
	evidence := []string{}
	for _, rel := range outgoing {
		evidence = append(evidence, fmt.Sprintf("depends on %s (%s)", rel.To, rel.Type))
	}
	if len(outgoing) > 5 {
		breaking = true
		impactDetails = "High number of outgoing dependencies may cause cascading failures"
	}
	contract := ContractImpact{
		Consumers:  len(outgoing),
		IsBreaking: breaking,
		Details:    impactDetails,
		Evidence:   evidence,
	}

	// Overall confidence: average
	avgConf := (necessity.Confidence + 0.9 + stdLib.Confidence + dup.Confidence + 0.8) / 5

	conclusion := "✅ Justified – this code is necessary and well-placed."
	if !necessity.Verified {
		conclusion = "⚠️ Dead code – no incoming references found. Consider removal."
	} else if !simplicity.Pass {
		conclusion = "⚠️ Code is complex – consider simplification."
	} else if !stdLib.Pass {
		conclusion = "⚠️ Standard library alternative available – consider using it."
	} else if !dup.Pass {
		conclusion = "⚠️ Duplicate logic found – consolidate."
	} else if breaking {
		conclusion = "⚠️ Potential breaking change – review impact."
	}

	return &Justification{
		EntityName:        entity.Name,
		File:              entity.File,
		Necessity:         necessity,
		Simplicity:        simplicity,
		StdLib:            stdLib,
		Duplication:       dup,
		ContractImpact:    contract,
		Evidence:          evidence,
		Conclusion:        conclusion,
		OverallConfidence: avgConf,
	}
}
