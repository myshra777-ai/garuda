package analyzer

type DiffResult struct {
	AddedEntities    []Entity     `json:"added_entities"`
	RemovedEntities  []Entity     `json:"removed_entities"`
	ModifiedEntities []EntityDiff `json:"modified_entities"`
	IsBreaking       bool         `json:"is_breaking"`
}

type EntityDiff struct {
	EntityID       string   `json:"entity_id"`
	AddedFields    []Field  `json:"added_fields"`
	RemovedFields  []Field  `json:"removed_fields"`
	AddedMethods   []Method `json:"added_methods"`
	RemovedMethods []Method `json:"removed_methods"`
}

// Compare computes the semantic diff between two analysis snapshots
func Compare(base, target *AnalysisResult) *DiffResult {
	diff := &DiffResult{
		AddedEntities:    []Entity{},
		RemovedEntities:  []Entity{},
		ModifiedEntities: []EntityDiff{},
		IsBreaking:       false,
	}

	baseMap := make(map[string]Entity)
	for _, e := range base.Entities {
		baseMap[e.ID] = e
	}

	targetMap := make(map[string]Entity)
	for _, e := range target.Entities {
		targetMap[e.ID] = e
	}

	// 1. Detect Removed Entities (Breaking if exported)
	for id, baseEntity := range baseMap {
		if _, exists := targetMap[id]; !exists {
			diff.RemovedEntities = append(diff.RemovedEntities, baseEntity)
			if baseEntity.IsExported {
				diff.IsBreaking = true
			}
		}
	}

	// 2. Detect Added Entities
	for id, targetEntity := range targetMap {
		if _, exists := baseMap[id]; !exists {
			diff.AddedEntities = append(diff.AddedEntities, targetEntity)
		}
	}

	// 3. Detect Modified Structs & Interfaces
	for id, baseEntity := range baseMap {
		targetEntity, exists := targetMap[id]
		if !exists {
			continue
		}

		eDiff := EntityDiff{EntityID: id}

		// Field changes
		fBase := make(map[string]Field)
		for _, f := range baseEntity.Fields {
			fBase[f.Name] = f
		}
		fTarget := make(map[string]Field)
		for _, f := range targetEntity.Fields {
			fTarget[f.Name] = f
		}

		for name, field := range fBase {
			if _, ok := fTarget[name]; !ok {
				eDiff.RemovedFields = append(eDiff.RemovedFields, field)
				if baseEntity.IsExported {
					diff.IsBreaking = true
				}
			}
		}
		for name, field := range fTarget {
			if _, ok := fBase[name]; !ok {
				eDiff.AddedFields = append(eDiff.AddedFields, field)
			}
		}

		if len(eDiff.AddedFields) > 0 || len(eDiff.RemovedFields) > 0 || len(eDiff.AddedMethods) > 0 || len(eDiff.RemovedMethods) > 0 {
			diff.ModifiedEntities = append(diff.ModifiedEntities, eDiff)
		}
	}

	return diff
}
