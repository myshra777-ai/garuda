package impact

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/store"
)

// ImpactReport represents the complete impact analysis result
type ImpactReport struct {
	ID              string           `json:"id"`
	BreakingChanges []BreakingChange `json:"breaking_changes"`
	Warnings        []Warning        `json:"warnings"`
	AffectedRepos   []string         `json:"affected_repos"`
	EvidenceRoot    string           `json:"evidence_root"`
	Summary         ImpactSummary    `json:"summary"`
}

// BreakingChange represents a change that will break consumers
type BreakingChange struct {
	Entity      string   `json:"entity"`
	Kind        string   `json:"kind"`
	Package     string   `json:"package"`
	ChangeType  string   `json:"change_type"` // "signature", "type", "removal", "schema"
	ImpactCount int      `json:"impact_count"`
	Consumers   []string `json:"consumers"`
	Evidence    []string `json:"evidence"`
	Mitigation  string   `json:"mitigation"`
}

// Warning represents a non-breaking but notable change
type Warning struct {
	Entity     string  `json:"entity"`
	Message    string  `json:"message"`
	Confidence float64 `json:"confidence"`
}

// ImpactSummary provides high-level metrics.
// These fields support both report consumers and CLI display logic.
type ImpactSummary struct {
	TotalChanges  int `json:"total_changes"`
	BreakingCount int `json:"breaking_count"`
	WarningCount  int `json:"warning_count"`
	ReposAffected int `json:"repos_affected"`
	TotalAffected int `json:"total_affected"`
	Critical      int `json:"critical"`
	High          int `json:"high"`
	Medium        int `json:"medium"`
	Low           int `json:"low"`
}

// AnalyzeImpact compares two snapshots and identifies breaking changes
func AnalyzeImpact(
	ctx context.Context,
	st *store.PostgresStore,
	tenantID, workspaceID uuid.UUID,
	baseline, proposed *analyzer.Result,
) (*ImpactReport, error) {
	report := &ImpactReport{
		ID:              uuid.New().String(),
		BreakingChanges: []BreakingChange{},
		Warnings:        []Warning{},
		AffectedRepos:   []string{},
	}

	// 1. Get diff
	diff := analyzer.Diff(baseline, proposed)

	// 2. Analyze each entity diff for breaking changes
	for _, ed := range diff.EntityDiffs {
		if ed.Status == "modified" {
			breaking := analyzeEntityChange(ed, baseline, proposed)
			if breaking != nil {
				report.BreakingChanges = append(report.BreakingChanges, *breaking)
			}
		}
	}

	// 3. Analyze relationship changes
	for _, rd := range diff.RelationshipDiffs {
		if rd.Status == "removed" {
			warning := analyzeRelationshipRemoval(rd, baseline)
			if warning != nil {
				report.Warnings = append(report.Warnings, *warning)
			}
		}
	}

	// 4. Calculate summary
	report.Summary = ImpactSummary{
		TotalChanges:  len(diff.EntityDiffs),
		BreakingCount: len(report.BreakingChanges),
		WarningCount:  len(report.Warnings),
		ReposAffected: len(report.AffectedRepos),
		TotalAffected: len(diff.EntityDiffs),
	}

	// 5. Generate evidence root (Merkle-like)
	evidenceData, _ := json.Marshal(report)
	hash := sha256.Sum256(evidenceData)
	report.EvidenceRoot = hex.EncodeToString(hash[:])

	return report, nil
}

// analyzeEntityChange determines if an entity change is breaking
func analyzeEntityChange(ed analyzer.EntityDiff, baseline, proposed *analyzer.Result) *BreakingChange {
	// 1. Check if fields were removed (breaking)
	if ed.FieldsDiff != nil && len(ed.FieldsDiff.Removed) > 0 {
		return &BreakingChange{
			Entity:      ed.Name,
			Kind:        string(ed.Kind),
			Package:     extractPackageFromEntity(ed.Name, ed.EntityID),
			ChangeType:  "field_removed",
			ImpactCount: ed.Impact,
			Consumers:   []string{"Multiple consumers may be affected"},
			Evidence:    []string{fmt.Sprintf("Field(s) removed: %v", ed.FieldsDiff.Removed)},
			Mitigation:  "Add a new field with a different name, deprecate the old one",
		}
	}

	// 2. Check if method signature changed (breaking for exported methods)
	if ed.MethodsDiff != nil && len(ed.MethodsDiff.Removed) > 0 && isEntityExported(ed.Name) {
		return &BreakingChange{
			Entity:      ed.Name,
			Kind:        string(ed.Kind),
			Package:     extractPackageFromEntity(ed.Name, ed.EntityID),
			ChangeType:  "method_removed",
			ImpactCount: ed.Impact,
			Consumers:   []string{"External callers may be broken"},
			Evidence:    []string{fmt.Sprintf("Method(s) removed: %v", ed.MethodsDiff.Removed)},
			Mitigation:  "Keep the method with a deprecation notice, add a new one",
		}
	}

	return nil
}

// analyzeRelationshipRemoval warns about removed dependencies
func analyzeRelationshipRemoval(rd analyzer.RelationshipDiff, baseline *analyzer.Result) *Warning {
	return &Warning{
		Entity:     rd.From,
		Message:    fmt.Sprintf("Removed dependency %s -> %s (%s)", rd.From, rd.To, rd.Type),
		Confidence: 0.7,
	}
}

// Helper: Extract package name from a qualified entity name or ID
// (e.g., "github.com/myshra777-ai/garuda/pkg/auth.User" -> "auth" or "pkg/auth")
func extractPackageFromEntity(name, entityID string) string {
	target := name
	if target == "" {
		target = entityID
	}
	lastSlash := strings.LastIndex(target, "/")
	if lastSlash != -1 {
		target = target[lastSlash+1:]
	}
	dotIdx := strings.Index(target, ".")
	if dotIdx != -1 {
		return target[:dotIdx]
	}
	return "main"
}

// Helper: Check if an entity identifier is exported (starts with uppercase rune)
func isEntityExported(name string) bool {
	cleanName := name
	if dotIdx := strings.LastIndex(cleanName, "."); dotIdx != -1 {
		cleanName = cleanName[dotIdx+1:]
	}
	r, _ := utf8.DecodeRuneInString(cleanName)
	return unicode.IsUpper(r)
}
