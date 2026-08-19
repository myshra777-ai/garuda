// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Evidence defines the physical provenance backing a claim edge.
type Evidence struct {
	RepositoryID   uuid.UUID `json:"repository_id"`
	CommitSHA      string    `json:"commit_sha"`
	FilePath       string    `json:"file_path"`
	Symbol         string    `json:"symbol"`
	LineStart      int       `json:"line_start"`
	LineEnd        int       `json:"line_end"`
	ContentSnippet string    `json:"content_snippet"`
	ContentHash    string    `json:"content_hash"`
	Analyzer       string    `json:"analyzer"`
	AnalyzerVer    string    `json:"analyzer_version"`
	CapturedAt     time.Time `json:"captured_at"`
}

// EntityMetadata holds entity information for graph traversal
type EntityMetadata struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Package      string    `json:"package"`
	Kind         string    `json:"kind"`
	FilePath     string    `json:"file_path"`
	RepositoryID uuid.UUID `json:"repository_id"`
}

// ConsumerEdge represents an incoming dependency where TargetEntity is consumed by ConsumerEntityID.
type ConsumerEdge struct {
	ConsumerEntityID string   `json:"consumer_entity_id"`
	ClaimType        string   `json:"claim_type"`
	EpistemicClass   string   `json:"epistemic_class"`
	Confidence       float64  `json:"confidence"`
	Evidence         Evidence `json:"evidence"`
}

// DependencyEdge represents an outgoing dependency from SourceEntity to DependentEntityID.
type DependencyEdge struct {
	DependentEntityID string   `json:"dependent_entity_id"`
	ClaimType         string   `json:"claim_type"`
	EpistemicClass    string   `json:"epistemic_class"`
	Confidence        float64  `json:"confidence"`
	Evidence          Evidence `json:"evidence"`
}

// ImpactIndex holds O(1) in-memory adjacency lookups for workspace-level traversal.
type ImpactIndex struct {
	mu           sync.RWMutex
	workspaceID  uuid.UUID
	consumers    map[string][]ConsumerEdge   // TargetID -> Inbound consumers
	dependencies map[string][]DependencyEdge // SourceID -> Outbound dependencies
	entityLookup map[string]EntityMetadata
	builtAt      time.Time
}

// NewImpactIndex initializes an empty in-memory index.
func NewImpactIndex(workspaceID uuid.UUID) *ImpactIndex {
	return &ImpactIndex{
		workspaceID:  workspaceID,
		consumers:    make(map[string][]ConsumerEdge),
		dependencies: make(map[string][]DependencyEdge),
		entityLookup: make(map[string]EntityMetadata),
		builtAt:      time.Now().UTC(),
	}
}

// AddEdge registers an observed or inferred claim between two entities.
func (idx *ImpactIndex) AddEdge(source EntityMetadata, target EntityMetadata, claimType, epistemicClass string, conf float64, ev Evidence) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.entityLookup[source.ID] = source
	idx.entityLookup[target.ID] = target

	// Inbound link: target is consumed by source
	idx.consumers[target.ID] = append(idx.consumers[target.ID], ConsumerEdge{
		ConsumerEntityID: source.ID,
		ClaimType:        claimType,
		EpistemicClass:   epistemicClass,
		Confidence:       conf,
		Evidence:         ev,
	})

	// Outbound link: source depends on target
	idx.dependencies[source.ID] = append(idx.dependencies[source.ID], DependencyEdge{
		DependentEntityID: target.ID,
		ClaimType:         claimType,
		EpistemicClass:    epistemicClass,
		Confidence:        conf,
		Evidence:          ev,
	})
}

// GetConsumers returns all upstream callers/consumers of a given entity ID.
func (idx *ImpactIndex) GetConsumers(entityID string) []ConsumerEdge {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.consumers[entityID]
}

// GetEntityMeta retrieves metadata for a registered entity.
func (idx *ImpactIndex) GetEntityMeta(entityID string) (EntityMetadata, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	meta, exists := idx.entityLookup[entityID]
	return meta, exists
}

// BuildImpactIndex fetches workspace claims from PostgreSQL and projects the in-memory graph.
func (s *PostgresStore) BuildImpactIndex(ctx context.Context, workspaceID uuid.UUID) (*ImpactIndex, error) {
	idx := NewImpactIndex(workspaceID)

	// Fetch claims joined with source/target entity data
	query := `
		SELECT 
			c.from_entity_id, 
			se.name as source_name, 
			se.kind as source_kind, 
			se.package as source_package, 
			se.repository_id as source_repo,
			c.to_entity_id, 
			te.name as target_name, 
			te.kind as target_kind, 
			te.package as target_package, 
			te.repository_id as target_repo,
			c.claim_type, 
			c.confidence,
			te.file_path,
			te.commit_sha
		FROM claims c
		JOIN entities se ON c.from_entity_id = se.id
		JOIN entities te ON c.to_entity_id = te.id
		WHERE c.workspace_id = $1
	`
	rows, err := s.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query claims for impact index: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			srcID, tgtID     uuid.UUID
			srcName, tgtName string
			srcKind, tgtKind string
			srcPkg, tgtPkg   string
			srcRepo, tgtRepo uuid.UUID
			claimType        string
			confidence       float64
			filePath         string
			commitSHA        string
		)

		if err := rows.Scan(
			&srcID, &srcName, &srcKind, &srcPkg, &srcRepo,
			&tgtID, &tgtName, &tgtKind, &tgtPkg, &tgtRepo,
			&claimType, &confidence,
			&filePath, &commitSHA,
		); err != nil {
			return nil, fmt.Errorf("failed to scan claim row: %w", err)
		}

		// Default epistemic class to "OBSERVATION" since the column doesn't exist
		epistemicClass := "OBSERVATION"

		// Build evidence from entity data
		evidence := Evidence{
			RepositoryID: tgtRepo,
			CommitSHA:    commitSHA,
			FilePath:     filePath,
			Symbol:       tgtName,
			LineStart:    0,
			LineEnd:      0,
			ContentHash:  "",
			Analyzer:     "garuda",
			AnalyzerVer:  "1.0",
			CapturedAt:   time.Now().UTC(),
		}

		// Create source entity metadata
		srcMeta := EntityMetadata{
			ID:           srcID.String(),
			Name:         srcName,
			Package:      srcPkg,
			Kind:         srcKind,
			FilePath:     "",
			RepositoryID: srcRepo,
		}

		// Create target entity metadata
		tgtMeta := EntityMetadata{
			ID:           tgtID.String(),
			Name:         tgtName,
			Package:      tgtPkg,
			Kind:         tgtKind,
			FilePath:     filePath,
			RepositoryID: tgtRepo,
		}

		idx.AddEdge(srcMeta, tgtMeta, claimType, epistemicClass, confidence, evidence)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during claim row iteration: %w", err)
	}

	return idx, nil
}
