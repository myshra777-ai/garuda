// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package impact

import (
	"path/filepath"
	"sort"
	"strings"
)

// DefaultImpactedLimit caps how many impacted entities are listed in
// machine-readable output, so a change to a hub symbol does not produce a
// multi-megabyte report.
const DefaultImpactedLimit = 200

// ImpactedSummary is one impacted entity in compact, report-friendly form.
// File is relative to the repository root, with forward slashes, so it can be
// matched against CODEOWNERS-style ownership rules.
type ImpactedSummary struct {
	Name       string        `json:"name"`
	Package    string        `json:"package,omitempty"`
	Kind       string        `json:"kind,omitempty"`
	File       string        `json:"file,omitempty"`
	Severity   SeverityLevel `json:"severity"`
	Depth      int           `json:"depth"`
	Confidence float64       `json:"confidence"`
}

// SeverityRank orders severities from LOW (1) to CRITICAL (4); unknown is 0.
func SeverityRank(s SeverityLevel) int {
	switch s {
	case SeverityCritical:
		return 4
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	default:
		return 0
	}
}

// SummarizeImpacted merges impacted entities from one or more blast-radius
// results into a deduplicated list. When an entity is reached more than once,
// the most severe, then shallowest, then most confident path wins. The list
// is ordered by severity, then depth, then name, and cut at limit (limit <= 0
// means DefaultImpactedLimit). The second result reports whether it was cut.
func SummarizeImpacted(results []*BlastRadiusResult, repoRoot string, limit int) ([]ImpactedSummary, bool) {
	if limit <= 0 {
		limit = DefaultImpactedLimit
	}
	best := make(map[string]ImpactedEntity)
	for _, r := range results {
		if r == nil {
			continue
		}
		for _, list := range r.BySeverity {
			for _, ie := range list {
				key := ie.EntityID
				if key == "" {
					key = ie.Package + "." + ie.Name + "@" + ie.FilePath
				}
				prev, seen := best[key]
				if !seen || betterImpact(ie, prev) {
					best[key] = ie
				}
			}
		}
	}

	out := make([]ImpactedSummary, 0, len(best))
	for _, ie := range best {
		out = append(out, ImpactedSummary{
			Name:       ie.Name,
			Package:    ie.Package,
			Kind:       ie.Kind,
			File:       relativeToRepo(ie.FilePath, repoRoot),
			Severity:   ie.Severity,
			Depth:      ie.Depth,
			Confidence: ie.CumulativeConf,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if ra, rb := SeverityRank(a.Severity), SeverityRank(b.Severity); ra != rb {
			return ra > rb
		}
		if a.Depth != b.Depth {
			return a.Depth < b.Depth
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.File < b.File
	})
	if len(out) > limit {
		return out[:limit], true
	}
	return out, false
}

func betterImpact(a, b ImpactedEntity) bool {
	if ra, rb := SeverityRank(a.Severity), SeverityRank(b.Severity); ra != rb {
		return ra > rb
	}
	if a.Depth != b.Depth {
		return a.Depth < b.Depth
	}
	return a.CumulativeConf > b.CumulativeConf
}

// relativeToRepo turns an indexed file path into a repository-relative,
// slash-separated path. Paths outside the repository are returned as they
// were (with forward slashes) rather than as "../" paths.
func relativeToRepo(p, repoRoot string) string {
	if p == "" {
		return ""
	}
	if repoRoot != "" && filepath.IsAbs(p) {
		if rel, err := filepath.Rel(repoRoot, p); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			p = rel
		}
	}
	return strings.TrimPrefix(filepath.ToSlash(p), "./")
}
