package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/analyzer"
)

// SaveEntities saves all entities and claims from an analysis result
func (s *PostgresStore) SaveEntities(
	ctx context.Context,
	tenantID, workspaceID, repoID, analysisID uuid.UUID,
	result *analyzer.Result,
) error {
	// 1. Insert entities
	for _, entity := range result.Entities {
		entityID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(
			tenantID.String()+workspaceID.String()+repoID.String()+entity.ID,
		))

		fieldsJSON, _ := json.Marshal(entity.Fields)
		methodsJSON, _ := json.Marshal(entity.Methods)

		_, err := s.pool.Exec(ctx, `
			INSERT INTO entities (
				id, tenant_id, workspace_id, repository_id, analysis_id,
				name, kind, package, file_path, commit_sha,
				signature, fields, methods, is_exported, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
			ON CONFLICT (tenant_id, workspace_id, repository_id, name, package) DO UPDATE
			SET analysis_id = EXCLUDED.analysis_id,
			    fields = EXCLUDED.fields,
			    methods = EXCLUDED.methods,
			    signature = EXCLUDED.signature,
			    is_exported = EXCLUDED.is_exported,
			    updated_at = NOW()
		`, entityID, tenantID, workspaceID, repoID, analysisID,
			entity.Name, entity.Kind, entity.Package, entity.File, "HEAD",
			entity.Signature, fieldsJSON, methodsJSON, entity.IsExported)
		if err != nil {
			return fmt.Errorf("failed to insert entity %s: %w", entity.Name, err)
		}
	}

	// 2. Insert claims (relationships)
	for _, rel := range result.Relationships {
		var fromID, toID uuid.UUID

		err := s.pool.QueryRow(ctx, `
			SELECT id FROM entities
			WHERE tenant_id = $1 AND workspace_id = $2 AND repository_id = $3
			AND name = $4 AND package = $5
		`, tenantID, workspaceID, repoID, rel.From, result.Package).Scan(&fromID)
		if err != nil {
			continue
		}

		err = s.pool.QueryRow(ctx, `
			SELECT id FROM entities
			WHERE tenant_id = $1 AND workspace_id = $2 AND repository_id = $3
			AND name = $4 AND package = $5
		`, tenantID, workspaceID, repoID, rel.To, result.Package).Scan(&toID)
		if err != nil {
			continue
		}

		_, err = s.pool.Exec(ctx, `
			INSERT INTO claims (
				id, tenant_id, workspace_id, repository_id, analysis_id,
				from_entity_id, to_entity_id, claim_type, confidence, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
			ON CONFLICT (tenant_id, from_entity_id, to_entity_id, claim_type) DO NOTHING
		`, uuid.New(), tenantID, workspaceID, repoID, analysisID,
			fromID, toID, rel.Type, 0.9)
		if err != nil {
			return fmt.Errorf("failed to insert claim: %w", err)
		}
	}

	return nil
}

// GetEntity retrieves an entity by name
func (s *PostgresStore) GetEntity(ctx context.Context, tenantID, workspaceID uuid.UUID, entityName string) (*analyzer.Entity, error) {
	var entity analyzer.Entity
	var fieldsJSON, methodsJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT name, kind, package, file_path, signature, fields, methods, is_exported
		FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2 AND name = $3
		ORDER BY created_at DESC LIMIT 1
	`, tenantID, workspaceID, entityName).Scan(
		&entity.Name, &entity.Kind, &entity.Package, &entity.File,
		&entity.Signature, &fieldsJSON, &methodsJSON, &entity.IsExported,
	)
	if err != nil {
		return nil, fmt.Errorf("entity not found: %w", err)
	}

	if len(fieldsJSON) > 0 {
		json.Unmarshal(fieldsJSON, &entity.Fields)
	}
	if len(methodsJSON) > 0 {
		json.Unmarshal(methodsJSON, &entity.Methods)
	}

	entity.ID = entity.Package + "." + entity.Name
	return &entity, nil
}

// GetEntityRelationships retrieves all relationships for an entity
func (s *PostgresStore) GetEntityRelationships(
	ctx context.Context,
	tenantID, workspaceID uuid.UUID,
	entityName, entityPackage string,
) (incoming, outgoing []analyzer.Relationship, err error) {
	var entityID uuid.UUID
	err = s.pool.QueryRow(ctx, `
		SELECT id FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2 AND name = $3 AND package = $4
	`, tenantID, workspaceID, entityName, entityPackage).Scan(&entityID)
	if err != nil {
		return nil, nil, fmt.Errorf("entity not found: %w", err)
	}

	// Outgoing relationships
	rows, err := s.pool.Query(ctx, `
		SELECT e.name, e.package, c.claim_type
		FROM claims c
		JOIN entities e ON e.id = c.to_entity_id
		WHERE c.tenant_id = $1 AND c.from_entity_id = $2
	`, tenantID, entityID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query outgoing: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rel analyzer.Relationship
		err := rows.Scan(&rel.To, &rel.Package, &rel.Type)
		if err != nil {
			continue
		}
		rel.From = entityName
		outgoing = append(outgoing, rel)
	}

	// Incoming relationships
	rows, err = s.pool.Query(ctx, `
		SELECT e.name, e.package, c.claim_type
		FROM claims c
		JOIN entities e ON e.id = c.from_entity_id
		WHERE c.tenant_id = $1 AND c.to_entity_id = $2
	`, tenantID, entityID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query incoming: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rel analyzer.Relationship
		err := rows.Scan(&rel.From, &rel.Package, &rel.Type)
		if err != nil {
			continue
		}
		rel.To = entityName
		incoming = append(incoming, rel)
	}

	return incoming, outgoing, nil
}

// ListEntities lists all entities in a workspace
func (s *PostgresStore) ListEntities(ctx context.Context, tenantID, workspaceID uuid.UUID) ([]analyzer.Entity, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT name, kind, package, file_path, is_exported
		FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2
		ORDER BY package, name
	`, tenantID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list entities: %w", err)
	}
	defer rows.Close()

	var entities []analyzer.Entity
	for rows.Next() {
		var e analyzer.Entity
		err := rows.Scan(&e.Name, &e.Kind, &e.Package, &e.File, &e.IsExported)
		if err != nil {
			continue
		}
		e.ID = e.Package + "." + e.Name
		entities = append(entities, e)
	}
	return entities, nil
}

// GetGraphData retrieves all nodes and edges for graph visualization
func (s *PostgresStore) GetGraphData(ctx context.Context, tenantID, workspaceID uuid.UUID) (nodes []map[string]interface{}, edges []map[string]interface{}, err error) {
	// Nodes
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, kind, package, file_path
		FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get nodes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var name, kind, pkg, file string
		err := rows.Scan(&id, &name, &kind, &pkg, &file)
		if err != nil {
			continue
		}
		nodes = append(nodes, map[string]interface{}{
			"id":      id.String(),
			"label":   name,
			"group":   kind,
			"package": pkg,
			"file":    file,
		})
	}

	// Edges
	rows, err = s.pool.Query(ctx, `
		SELECT from_entity_id, to_entity_id, claim_type
		FROM claims
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get edges: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var fromID, toID uuid.UUID
		var claimType string
		err := rows.Scan(&fromID, &toID, &claimType)
		if err != nil {
			continue
		}
		edges = append(edges, map[string]interface{}{
			"from": fromID.String(),
			"to":   toID.String(),
			"type": claimType,
		})
	}

	return nodes, edges, nil
}

// UpdateRepositorySyncStatus updates the sync status of a repository
// SaveSemanticGraph stores entities, claims, and observations from an analysis result
func (s *PostgresStore) SaveSemanticGraph(ctx context.Context, tenantID, workspaceID, repoID, analysisID uuid.UUID, result *analyzer.Result) error {
	// 1. Insert entities
	for _, entity := range result.Entities {
		entityID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(tenantID.String()+workspaceID.String()+repoID.String()+entity.ID))
		_, err := s.pool.Exec(ctx, `
			INSERT INTO entities (id, tenant_id, workspace_id, repository_id, analysis_id, name, kind, package, file_path, commit_sha, signature, fields, methods, is_exported, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
			ON CONFLICT (tenant_id, workspace_id, repository_id, name, package) DO UPDATE
			SET analysis_id = EXCLUDED.analysis_id,
			    fields = EXCLUDED.fields,
			    methods = EXCLUDED.methods,
			    signature = EXCLUDED.signature,
			    is_exported = EXCLUDED.is_exported
		`, entityID, tenantID, workspaceID, repoID, analysisID,
			entity.Name, entity.Kind, entity.Package, entity.File, "HEAD",
			entity.Signature, entity.Fields, entity.Methods, entity.IsExported)
		if err != nil {
			return fmt.Errorf("failed to insert entity %s: %w", entity.Name, err)
		}
	}

	// 2. Insert claims
	for _, rel := range result.Relationships {
		// Find the entity IDs
		var fromID, toID uuid.UUID
		err := s.pool.QueryRow(ctx, `
			SELECT id FROM entities WHERE tenant_id = $1 AND workspace_id = $2 AND repository_id = $3 AND name = $4 AND package = $5
		`, tenantID, workspaceID, repoID, rel.From, result.Package).Scan(&fromID)
		if err != nil {
			continue // Skip if entity not found
		}
		err = s.pool.QueryRow(ctx, `
			SELECT id FROM entities WHERE tenant_id = $1 AND workspace_id = $2 AND repository_id = $3 AND name = $4 AND package = $5
		`, tenantID, workspaceID, repoID, rel.To, result.Package).Scan(&toID)
		if err != nil {
			continue
		}

		_, err = s.pool.Exec(ctx, `
			INSERT INTO claims (id, tenant_id, workspace_id, repository_id, analysis_id, from_entity_id, to_entity_id, claim_type, confidence, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
			ON CONFLICT (tenant_id, from_entity_id, to_entity_id, claim_type) DO NOTHING
		`, uuid.New(), tenantID, workspaceID, repoID, analysisID,
			fromID, toID, rel.Type, 0.9)
		if err != nil {
			return fmt.Errorf("failed to insert claim: %w", err)
		}
	}
	return nil
}
