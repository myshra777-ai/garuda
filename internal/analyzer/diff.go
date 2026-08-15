package analyzer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// DiffReport holds the complete diff result
type DiffReport struct {
	StatsDiff         StatsDiff          `json:"stats_diff"`
	FingerprintDiff   FingerprintDiff    `json:"fingerprint_diff"`
	EntityDiffs       []EntityDiff       `json:"entity_diffs"`
	RelationshipDiffs []RelationshipDiff `json:"relationship_diffs"`
	Summary           DiffSummary        `json:"summary"`
}

// StatsDiff shows numeric changes
type StatsDiff struct {
	Files       int `json:"files"`
	Packages    int `json:"packages"`
	Structs     int `json:"structs"`
	Interfaces  int `json:"interfaces"`
	Functions   int `json:"functions"`
	Imports     int `json:"imports"`
	TotalFields int `json:"total_fields"`
}

// FingerprintDiff shows if fingerprints match
type FingerprintDiff struct {
	Before string `json:"before"`
	After  string `json:"after"`
	Match  bool   `json:"match"`
}

// EntityDiff represents a change to an entity
type EntityDiff struct {
	EntityID    string       `json:"entity_id"`
	Kind        string       `json:"kind"`
	Name        string       `json:"name"`
	Status      string       `json:"status"` // added, removed, modified
	FieldsDiff  *FieldsDiff  `json:"fields_diff,omitempty"`
	MethodsDiff *MethodsDiff `json:"methods_diff,omitempty"`
	Impact      int          `json:"impact"` // number of references affected
}

// FieldsDiff shows field changes
type FieldsDiff struct {
	Added    []Field     `json:"added"`
	Removed  []Field     `json:"removed"`
	Modified []FieldDiff `json:"modified"`
}

// FieldDiff shows a field change (before/after)
type FieldDiff struct {
	Name   string `json:"name"`
	Before Field  `json:"before"`
	After  Field  `json:"after"`
}

// MethodsDiff shows method changes
type MethodsDiff struct {
	Added   []Method `json:"added"`
	Removed []Method `json:"removed"`
}

// RelationshipDiff represents a change to a relationship
type RelationshipDiff struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Type   string `json:"type"`
	Status string `json:"status"` // added, removed
}

// DiffSummary provides a high-level overview
type DiffSummary struct {
	BreakingChanges int `json:"breaking_changes"`
	Warnings        int `json:"warnings"`
	Additions       int `json:"additions"`
	Removals        int `json:"removals"`
	Modified        int `json:"modified"`
}

// LoadResult loads an AnalysisResult from a JSON file
func LoadResult(path string) (*Result, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	var result Result
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return &result, nil
}

// Diff compares two results and returns a DiffReport
func Diff(before, after *Result) *DiffReport {
	report := &DiffReport{
		EntityDiffs:       []EntityDiff{},
		RelationshipDiffs: []RelationshipDiff{},
		StatsDiff      StatsDiff            `json:"stats_diff"`
    FingerprintDiff FingerprintDiff      `json:"fingerprint_diff"`
	}

	// 1. Stats diff
	report.StatsDiff = StatsDiff{
		Files:       after.Stats.Files - before.Stats.Files,
		Packages:    after.Stats.Packages - before.Stats.Packages,
		Structs:     after.Stats.Structs - before.Stats.Structs,
		Interfaces:  after.Stats.Interfaces - before.Stats.Interfaces,
		Functions:   after.Stats.Functions - before.Stats.Functions,
		Imports:     after.Stats.Imports - before.Stats.Imports,
		TotalFields: after.Stats.TotalFields - before.Stats.TotalFields,
	}

	// 2. Fingerprint diff
	report.FingerprintDiff = FingerprintDiff{
		Before: before.Fingerprint,
		After:  after.Fingerprint,
		Match:  before.Fingerprint == after.Fingerprint,
	}

	// 3. Entity diff
	beforeEntities := make(map[string]Entity)
	for _, e := range before.Entities {
		beforeEntities[e.ID] = e
	}
	afterEntities := make(map[string]Entity)
	for _, e := range after.Entities {
		afterEntities[e.ID] = e
	}

	// Added entities
	for id, e := range afterEntities {
		if _, exists := beforeEntities[id]; !exists {
			report.EntityDiffs = append(report.EntityDiffs, EntityDiff{
				EntityID: id,
				Kind:     e.Kind,
				Name:     e.Name,
				Status:   "added",
				Impact:   0, // can compute later
			})
			report.Summary.Additions++
		}
	}

	// Removed entities
	for id, e := range beforeEntities {
		if _, exists := afterEntities[id]; !exists {
			report.EntityDiffs = append(report.EntityDiffs, EntityDiff{
				EntityID: id,
				Kind:     e.Kind,
				Name:     e.Name,
				Status:   "removed",
				Impact:   len(e.Fields) + len(e.Methods), // rough impact
			})
			report.Summary.Removals++
		}
	}

	// Modified entities
	for id, beforeEntity := range beforeEntities {
		if afterEntity, exists := afterEntities[id]; exists {
			fieldsDiff, methodsDiff, modified := diffEntity(beforeEntity, afterEntity)
			if modified {
				impact := 0
				if fieldsDiff != nil {
					impact += len(fieldsDiff.Added) + len(fieldsDiff.Removed) + len(fieldsDiff.Modified)
				}
				if methodsDiff != nil {
					impact += len(methodsDiff.Added) + len(methodsDiff.Removed)
				}
				report.EntityDiffs = append(report.EntityDiffs, EntityDiff{
					EntityID:    id,
					Kind:        beforeEntity.Kind,
					Name:        beforeEntity.Name,
					Status:      "modified",
					FieldsDiff:  fieldsDiff,
					MethodsDiff: methodsDiff,
					Impact:      impact,
				})
				report.Summary.Modified++
				// Breaking if fields removed or types changed
				if fieldsDiff != nil && len(fieldsDiff.Removed) > 0 {
					report.Summary.BreakingChanges++
				}
				if fieldsDiff != nil && len(fieldsDiff.Modified) > 0 {
					report.Summary.Warnings++
				}
			}
		}
	}

	// 4. Relationship diff
	beforeRels := make(map[string]Relationship)
	for _, r := range before.Relationships {
		key := r.From + "|" + r.To + "|" + r.Type
		beforeRels[key] = r
	}
	afterRels := make(map[string]Relationship)
	for _, r := range after.Relationships {
		key := r.From + "|" + r.To + "|" + r.Type
		afterRels[key] = r
	}

	for key, r := range afterRels {
		if _, exists := beforeRels[key]; !exists {
			report.RelationshipDiffs = append(report.RelationshipDiffs, RelationshipDiff{
				From:   r.From,
				To:     r.To,
				Type:   r.Type,
				Status: "added",
			})
			report.Summary.Additions++
		}
	}
	for key, r := range beforeRels {
		if _, exists := afterRels[key]; !exists {
			report.RelationshipDiffs = append(report.RelationshipDiffs, RelationshipDiff{
				From:   r.From,
				To:     r.To,
				Type:   r.Type,
				Status: "removed",
			})
			report.Summary.Removals++
		}
	}

	return report
}

// diffEntity compares two entities and returns field/method diffs and whether modified
func diffEntity(before, after Entity) (*FieldsDiff, *MethodsDiff, bool) {
	var fieldsDiff *FieldsDiff
	var methodsDiff *MethodsDiff
	modified := false

	// Field diff
	beforeFields := make(map[string]Field)
	for _, f := range before.Fields {
		beforeFields[f.Name] = f
	}
	afterFields := make(map[string]Field)
	for _, f := range after.Fields {
		afterFields[f.Name] = f
	}

	var addedFields []Field
	var removedFields []Field
	var modifiedFields []FieldDiff

	for name, f := range afterFields {
		if beforeField, exists := beforeFields[name]; !exists {
			addedFields = append(addedFields, f)
		} else if !fieldsEqual(beforeField, f) {
			modifiedFields = append(modifiedFields, FieldDiff{
				Name:   name,
				Before: beforeField,
				After:  f,
			})
		}
	}
	for name, f := range beforeFields {
		if _, exists := afterFields[name]; !exists {
			removedFields = append(removedFields, f)
		}
	}

	if len(addedFields) > 0 || len(removedFields) > 0 || len(modifiedFields) > 0 {
		fieldsDiff = &FieldsDiff{
			Added:    addedFields,
			Removed:  removedFields,
			Modified: modifiedFields,
		}
		modified = true
	}

	// Method diff
	beforeMethods := make(map[string]Method)
	for _, m := range before.Methods {
		beforeMethods[m.Name] = m
	}
	afterMethods := make(map[string]Method)
	for _, m := range after.Methods {
		afterMethods[m.Name] = m
	}

	var addedMethods []Method
	var removedMethods []Method
	for name, m := range afterMethods {
		if _, exists := beforeMethods[name]; !exists {
			addedMethods = append(addedMethods, m)
		}
	}
	for name, m := range beforeMethods {
		if _, exists := afterMethods[name]; !exists {
			removedMethods = append(removedMethods, m)
		}
	}
	if len(addedMethods) > 0 || len(removedMethods) > 0 {
		methodsDiff = &MethodsDiff{
			Added:   addedMethods,
			Removed: removedMethods,
		}
		modified = true
	}

	return fieldsDiff, methodsDiff, modified
}

// fieldsEqual compares two fields for equality (ignoring comments)
func fieldsEqual(a, b Field) bool {
	if a.Name != b.Name {
		return false
	}
	if a.Type != b.Type {
		return false
	}
	if a.JSONTag != b.JSONTag {
		return false
	}
	if a.DBTag != b.DBTag {
		return false
	}
	if a.ValidateTag != b.ValidateTag {
		return false
	}
	if a.IsPointer != b.IsPointer {
		return false
	}
	if a.IsSlice != b.IsSlice {
		return false
	}
	return true
}
