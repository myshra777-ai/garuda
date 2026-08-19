// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadResult loads an AnalysisResult from a JSON file
func LoadResult(path string) (*Result, error) {
	data, err := os.ReadFile(path)
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
	if before == nil || after == nil {
		return &DiffReport{}
	}

	report := &DiffReport{
		StatsDiff:         StatsDiff{},
		FingerprintDiff:   FingerprintDiff{},
		EntityDiffs:       []EntityDiff{},
		RelationshipDiffs: []RelationshipDiff{},
		Summary:           DiffSummary{},
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
				File:     e.File,
				Status:   "added",
				Impact:   0,
			})
			report.Summary.Additions++
		}
	}

	// Removed entities
	for id, e := range beforeEntities {
		if _, exists := afterEntities[id]; !exists {
			impact := len(e.Fields) + len(e.Methods)
			report.EntityDiffs = append(report.EntityDiffs, EntityDiff{
				EntityID: id,
				Kind:     e.Kind,
				Name:     e.Name,
				File:     e.File,
				Status:   "removed",
				Impact:   impact,
			})
			report.Summary.Removals++
			// Removed entities are breaking changes
			if impact > 0 || e.Kind == KindStruct || e.Kind == KindInterface {
				report.Summary.BreakingChanges++
			}
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
					File:        beforeEntity.File,
					Status:      "modified",
					FieldsDiff:  fieldsDiff,
					MethodsDiff: methodsDiff,
					Impact:      impact,
				})
				report.Summary.Modified++

				// Classify breaking changes
				if fieldsDiff != nil {
					if len(fieldsDiff.Removed) > 0 {
						report.Summary.BreakingChanges++
					}
					if len(fieldsDiff.Modified) > 0 {
						report.Summary.Warnings++
					}
				}
				if methodsDiff != nil && len(methodsDiff.Removed) > 0 {
					report.Summary.BreakingChanges++
				}
			}
		}
	}

	// 4. Relationship diff
	beforeRels := make(map[string]Relationship)
	for _, r := range before.Relationships {
		key := fmt.Sprintf("%s|%s|%s", r.From, r.To, r.Type)
		beforeRels[key] = r
	}
	afterRels := make(map[string]Relationship)
	for _, r := range after.Relationships {
		key := fmt.Sprintf("%s|%s|%s", r.From, r.To, r.Type)
		afterRels[key] = r
	}

	for key, r := range afterRels {
		if _, exists := beforeRels[key]; !exists {
			report.RelationshipDiffs = append(report.RelationshipDiffs, RelationshipDiff{
				From:   r.From,
				To:     r.To,
				Type:   RelationshipType(r.Type),
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
				Type:   RelationshipType(r.Type),
				Status: "removed",
			})
			report.Summary.Removals++
			// Removed relationships can be breaking
			if r.Type == string(RelImplements) {
				report.Summary.BreakingChanges++
			}
		}
	}

	// 5. Detect signature changes (additional breaking change detection)
	for _, ed := range report.EntityDiffs {
		if ed.Status == "modified" && ed.Kind == KindFunction {
			beforeEntity := beforeEntities[ed.EntityID]
			afterEntity := afterEntities[ed.EntityID]
			if beforeEntity.Signature != afterEntity.Signature {
				// Signature change is breaking for exported functions
				if beforeEntity.Exported || afterEntity.Exported {
					report.Summary.BreakingChanges++
				}
			}
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

// fieldsEqual compares two fields for equality
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
	if a.Tag != b.Tag {
		return false
	}
	if a.Comment != b.Comment {
		return false
	}
	return true
}

// HasBreakingChanges returns true if the diff has any breaking changes
func (r *DiffReport) HasBreakingChanges() bool {
	return r.Summary.BreakingChanges > 0
}

// HasWarnings returns true if the diff has any warnings
func (r *DiffReport) HasWarnings() bool {
	return r.Summary.Warnings > 0
}
