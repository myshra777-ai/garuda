// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package hygiene

import (
	"sort"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

// FindingKind identifies the advisory hygiene category.
type FindingKind string

const (
	FindingStaticUnreferencedCandidate FindingKind = "STATIC_UNREFERENCED_CANDIDATE"
	FindingDuplicateSymbolName         FindingKind = "DUPLICATE_SYMBOL_NAME"
	FindingStandardLibraryAlternative  FindingKind = "STANDARD_LIBRARY_ALTERNATIVE"
)

// Finding is an advisory observation derived from the workspace semantic graph.
// It is not a persisted governance decision and does not prove dead code.
type Finding struct {
	Kind             FindingKind         `json:"kind"`
	EntityID         string              `json:"entity_id,omitempty"`
	Name             string              `json:"name"`
	EntityKind       analyzer.EntityKind `json:"entity_kind"`
	Package          string              `json:"package,omitempty"`
	Packages         []string            `json:"packages,omitempty"`
	File             string              `json:"file,omitempty"`
	LineStart        int                 `json:"line_start,omitempty"`
	LineEnd          int                 `json:"line_end,omitempty"`
	Message          string              `json:"message"`
	Reason           string              `json:"reason"`
	Confidence       *float64            `json:"confidence,omitempty"`
	ResolutionStatus string              `json:"resolution_status,omitempty"`
	ResolutionMethod string              `json:"resolution_method,omitempty"`
	EpistemicClass   string              `json:"epistemic_class,omitempty"`
	Advisory         bool                `json:"advisory"`
}

// Summary counts findings by advisory category.
type Summary struct {
	StaticUnreferencedCandidates int `json:"static_unreferenced_candidates"`
	DuplicateSymbolNames         int `json:"duplicate_symbol_names"`
	StandardLibraryAlternatives  int `json:"standard_library_alternatives"`
}

// Report is the deterministic, non-mutating result of a hygiene analysis.
type Report struct {
	Findings      []Finding `json:"findings"`
	Summary       Summary   `json:"summary"`
	EntityCount   int       `json:"entity_count"`
	RelationCount int       `json:"relationship_count"`
}

func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Kind != findings[j].Kind {
			return findings[i].Kind < findings[j].Kind
		}
		if findings[i].Name != findings[j].Name {
			return findings[i].Name < findings[j].Name
		}
		if findings[i].Package != findings[j].Package {
			return findings[i].Package < findings[j].Package
		}
		if findings[i].File != findings[j].File {
			return findings[i].File < findings[j].File
		}
		if findings[i].LineStart != findings[j].LineStart {
			return findings[i].LineStart < findings[j].LineStart
		}
		if findings[i].LineEnd != findings[j].LineEnd {
			return findings[i].LineEnd < findings[j].LineEnd
		}
		if findings[i].EntityID != findings[j].EntityID {
			return findings[i].EntityID < findings[j].EntityID
		}
		return findings[i].Message < findings[j].Message
	})
}
