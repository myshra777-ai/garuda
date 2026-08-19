// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package impact

import (
	"math"
	"sort"
	"strings"

	"github.com/myshra777-ai/garuda/internal/store"
)

type SeverityLevel string

const (
	SeverityCritical SeverityLevel = "CRITICAL"
	SeverityHigh     SeverityLevel = "HIGH"
	SeverityMedium   SeverityLevel = "MEDIUM"
	SeverityLow      SeverityLevel = "LOW"
)

// ImpactSummary provides high-level metrics

type BlastRadiusConfig struct {
	MaxDepth        int     `json:"max_depth"`
	MinConfidence   float64 `json:"min_confidence"`
	IncludeInferred bool    `json:"include_inferred"`
}

func DefaultBlastRadiusConfig() BlastRadiusConfig {
	return BlastRadiusConfig{
		MaxDepth:        3,
		MinConfidence:   0.50,
		IncludeInferred: true,
	}
}

type ImpactedEntity struct {
	EntityID       string         `json:"entity_id"`
	Name           string         `json:"name"`
	Package        string         `json:"package"`
	Kind           string         `json:"kind"`
	FilePath       string         `json:"file_path"`
	TraversalChain []string       `json:"traversal_chain"`
	Depth          int            `json:"depth"`
	CumulativeConf float64        `json:"cumulative_confidence"`
	EpistemicClass string         `json:"epistemic_class"` // "OBSERVATION" or "INFERENCE"
	Evidence       store.Evidence `json:"evidence"`
	Severity       SeverityLevel  `json:"severity"`
}

type BlastRadiusResult struct {
	TargetEntityID string                             `json:"target_entity_id"`
	TotalAffected  int                                `json:"total_affected"`
	BySeverity     map[SeverityLevel][]ImpactedEntity `json:"by_severity"`
}

// ToImpactSummary converts a BlastRadiusResult to an ImpactSummary
func (r *BlastRadiusResult) ToImpactSummary() ImpactSummary {
	summary := ImpactSummary{
		TotalChanges:  0,
		BreakingCount: 0,
		WarningCount:  0,
		ReposAffected: 0,
	}

	for sev, list := range r.BySeverity {
		switch sev {
		case SeverityCritical:
			summary.Critical = len(list)
			summary.BreakingCount += len(list)
		case SeverityHigh:
			summary.High = len(list)
			summary.BreakingCount += len(list)
		case SeverityMedium:
			summary.Medium = len(list)
			summary.WarningCount += len(list)
		case SeverityLow:
			summary.Low = len(list)
			summary.WarningCount += len(list)
		}
	}
	summary.TotalAffected = r.TotalAffected
	summary.TotalChanges = r.TotalAffected

	return summary
}

// ComputeBlastRadius executes a breadth-first search (BFS) over incoming consumer edges.
func ComputeBlastRadius(
	index *store.ImpactIndex,
	targetEntityID string,
	cfg BlastRadiusConfig,
) *BlastRadiusResult {
	if cfg.MaxDepth <= 0 {
		cfg.MaxDepth = 3
	}
	if cfg.MinConfidence <= 0 {
		cfg.MinConfidence = 0.50
	}

	result := &BlastRadiusResult{
		TargetEntityID: targetEntityID,
		BySeverity:     make(map[SeverityLevel][]ImpactedEntity),
	}

	targetMeta, exists := index.GetEntityMeta(targetEntityID)
	if !exists {
		return result
	}

	type queueNode struct {
		entityID       string
		depth          int
		cumulativeConf float64
		chain          []string
		hasInference   bool
	}

	visited := make(map[string]int) // entityID -> shortest depth visited
	queue := []queueNode{
		{
			entityID:       targetEntityID,
			depth:          0,
			cumulativeConf: 1.0,
			chain:          []string{targetMeta.Name},
			hasInference:   false,
		},
	}
	visited[targetEntityID] = 0

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr.depth >= cfg.MaxDepth {
			continue
		}

		consumers := index.GetConsumers(curr.entityID)
		for _, edge := range consumers {
			if edge.Confidence < cfg.MinConfidence {
				continue
			}
			// epistemic_class column doesn't exist in the database,
			// so we default to OBSERVATION
			// if !cfg.IncludeInferred && edge.EpistemicClass == "INFERENCE" {
			// 	continue
			// }

			nextDepth := curr.depth + 1
			if prevDepth, seen := visited[edge.ConsumerEntityID]; seen && prevDepth <= nextDepth {
				continue
			}
			visited[edge.ConsumerEntityID] = nextDepth

			consumerMeta, ok := index.GetEntityMeta(edge.ConsumerEntityID)
			if !ok {
				continue
			}

			cumulativeConf := curr.cumulativeConf * edge.Confidence
			// Since epistemic_class doesn't exist, default to OBSERVATION
			hasInference := curr.hasInference || false
			epistemicClass := "OBSERVATION"
			if hasInference {
				epistemicClass = "INFERENCE"
			}

			newChain := make([]string, len(curr.chain), len(curr.chain)+1)
			copy(newChain, curr.chain)
			newChain = append(newChain, consumerMeta.Name)

			severity := classifySeverity(edge, consumerMeta, nextDepth, epistemicClass, cumulativeConf)

			impacted := ImpactedEntity{
				EntityID:       consumerMeta.ID,
				Name:           consumerMeta.Name,
				Package:        consumerMeta.Package,
				Kind:           consumerMeta.Kind,
				FilePath:       consumerMeta.FilePath,
				TraversalChain: newChain,
				Depth:          nextDepth,
				CumulativeConf: math.Round(cumulativeConf*100) / 100,
				EpistemicClass: epistemicClass,
				Evidence:       edge.Evidence,
				Severity:       severity,
			}

			result.BySeverity[severity] = append(result.BySeverity[severity], impacted)
			result.TotalAffected++

			queue = append(queue, queueNode{
				entityID:       edge.ConsumerEntityID,
				depth:          nextDepth,
				cumulativeConf: cumulativeConf,
				chain:          newChain,
				hasInference:   hasInference,
			})
		}
	}

	return result
}

// classifySeverity calculates risk based on entity kind, epistemic class, depth, and edge type.
func classifySeverity(edge store.ConsumerEdge, meta store.EntityMetadata, depth int, epistemicClass string, conf float64) SeverityLevel {
	kind := strings.ToUpper(meta.Kind)
	claim := strings.ToUpper(edge.ClaimType)

	// Rule 1: Depth 1 direct public contract breach is CRITICAL
	if depth == 1 && epistemicClass == "OBSERVATION" {
		if kind == "HTTP_ROUTE" || kind == "API_ENDPOINT" || kind == "SQL_SCHEMA" || claim == "IMPLEMENTS" {
			return SeverityCritical
		}
		if kind != "TEST" {
			return SeverityHigh
		}
	}

	// Rule 2: Depth 2 direct observation on public routes is HIGH
	if depth == 2 && epistemicClass == "OBSERVATION" && (kind == "HTTP_ROUTE" || kind == "API_ENDPOINT") {
		return SeverityHigh
	}

	// Rule 3: Low confidence or deep transitive chains are LOW
	if conf < 0.70 || kind == "TEST" || depth >= 3 {
		return SeverityLow
	}

	return SeverityMedium
}

// SortImpactedEntities sorts impacted entities by severity (critical to low) and then by name
func SortImpactedEntities(entities []ImpactedEntity) []ImpactedEntity {
	sort.Slice(entities, func(i, j int) bool {
		severityRank := map[SeverityLevel]int{
			SeverityCritical: 0,
			SeverityHigh:     1,
			SeverityMedium:   2,
			SeverityLow:      3,
		}
		if severityRank[entities[i].Severity] != severityRank[entities[j].Severity] {
			return severityRank[entities[i].Severity] < severityRank[entities[j].Severity]
		}
		return entities[i].Name < entities[j].Name
	})
	return entities
}
