package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/analyzer"
)

// moduleCache caches module paths per workspace to avoid repeated DB queries
type moduleCache struct {
	mu      sync.RWMutex
	cache   map[string]map[string]uuid.UUID // workspaceID -> modulePath -> repoID
	expires map[string]time.Time
	ttl     time.Duration
}

func newModuleCache(ttl time.Duration) *moduleCache {
	return &moduleCache{
		cache:   make(map[string]map[string]uuid.UUID),
		expires: make(map[string]time.Time),
		ttl:     ttl,
	}
}

func (c *moduleCache) get(workspaceID string) (map[string]uuid.UUID, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if expiry, ok := c.expires[workspaceID]; ok && time.Now().Before(expiry) {
		return c.cache[workspaceID], true
	}
	return nil, false
}

func (c *moduleCache) set(workspaceID string, modules map[string]uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[workspaceID] = modules
	c.expires[workspaceID] = time.Now().Add(c.ttl)
}

func (c *moduleCache) invalidate(workspaceID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, workspaceID)
	delete(c.expires, workspaceID)
}

var crossRepoCache = newModuleCache(5 * time.Minute)

// getModuleMap returns a map of modulePath -> repoID for a workspace
func (s *PostgresStore) getModuleMap(ctx context.Context, workspaceID uuid.UUID) (map[string]uuid.UUID, error) {
	wsID := workspaceID.String()
	if modules, ok := crossRepoCache.get(wsID); ok {
		return modules, nil
	}

	rows, err := s.pool.Query(ctx, `
        SELECT id, module_path FROM repositories
        WHERE workspace_id = $1 AND module_path IS NOT NULL AND module_path != ''
    `, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query module paths: %w", err)
	}
	defer rows.Close()

	modules := make(map[string]uuid.UUID)
	for rows.Next() {
		var id uuid.UUID
		var mp string
		if err := rows.Scan(&id, &mp); err != nil {
			continue
		}
		mp = strings.TrimSuffix(mp, "/")
		modules[mp] = id
	}

	crossRepoCache.set(wsID, modules)
	return modules, nil
}

// findPackageEntityInRepo finds a package entity in a specific repository
func (s *PostgresStore) findPackageEntityInRepo(ctx context.Context, tenantID, repoID uuid.UUID, packagePath string) (*uuid.UUID, error) {
	var entityID uuid.UUID
	err := s.pool.QueryRow(ctx, `
        SELECT id FROM entities
        WHERE tenant_id = $1 AND repository_id = $2 AND package = $3 AND kind = 'package'
        LIMIT 1
    `, tenantID, repoID, packagePath).Scan(&entityID)
	if err != nil {
		return nil, err
	}
	return &entityID, nil
}

// SaveEntities stores entities from an analysis result with line numbers
func (s *PostgresStore) SaveEntities(
	ctx context.Context,
	tenantID, workspaceID, repoID, analysisID uuid.UUID,
	result *analyzer.Result,
) error {
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
                signature, fields, methods, is_exported,
                line, line_start, line_end,
                created_at, updated_at
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW(), NOW())
            ON CONFLICT (tenant_id, workspace_id, repository_id, name, package) DO UPDATE
            SET analysis_id = EXCLUDED.analysis_id,
                fields = EXCLUDED.fields,
                methods = EXCLUDED.methods,
                signature = EXCLUDED.signature,
                is_exported = EXCLUDED.is_exported,
                line = EXCLUDED.line,
                line_start = EXCLUDED.line_start,
                line_end = EXCLUDED.line_end,
                updated_at = NOW()
        `, entityID, tenantID, workspaceID, repoID, analysisID,
			entity.Name, string(entity.Kind), entity.Package, entity.File, "HEAD",
			entity.Signature, fieldsJSON, methodsJSON, entity.Exported,
			entity.Line, entity.LineStart, entity.LineEnd)
		if err != nil {
			return fmt.Errorf("failed to insert entity %s: %w", entity.Name, err)
		}
	}
	return nil
}

// SaveSemanticGraph stores entities and claims from an analysis result with cross-repo detection
func (s *PostgresStore) SaveSemanticGraph(
	ctx context.Context,
	tenantID, workspaceID, repoID, analysisID uuid.UUID,
	result *analyzer.Result,
) error {
	entityIDMap := make(map[string]uuid.UUID)

	// 1. Insert entities and populate map with the actual DB IDs
	for _, entity := range result.Entities {
		// First check if the entity already exists
		var existingID uuid.UUID
		err := s.pool.QueryRow(ctx, `
            SELECT id FROM entities
            WHERE tenant_id = $1 AND workspace_id = $2 AND repository_id = $3 
            AND name = $4 AND package = $5
        `, tenantID, workspaceID, repoID, entity.Name, entity.Package).Scan(&existingID)

		var entityID uuid.UUID
		if err == nil {
			// Entity exists, use existing ID
			entityID = existingID
			// Update it with new data including line numbers
			fieldsJSON, _ := json.Marshal(entity.Fields)
			methodsJSON, _ := json.Marshal(entity.Methods)
			_, err = s.pool.Exec(ctx, `
                UPDATE entities SET
                    analysis_id = $1,
                    kind = $2,
                    file_path = $3,
                    commit_sha = $4,
                    signature = $5,
                    fields = $6,
                    methods = $7,
                    is_exported = $8,
                    line = $9,
                    line_start = $10,
                    line_end = $11,
                    updated_at = NOW()
                WHERE id = $12
            `, analysisID, string(entity.Kind), entity.File, "HEAD",
				entity.Signature, fieldsJSON, methodsJSON, entity.Exported,
				entity.Line, entity.LineStart, entity.LineEnd, entityID)
			if err != nil {
				return fmt.Errorf("failed to update entity %s: %w", entity.Name, err)
			}
		} else {
			// Insert new entity
			entityID = uuid.NewSHA1(uuid.NameSpaceURL, []byte(
				tenantID.String()+workspaceID.String()+repoID.String()+entity.ID,
			))
			fieldsJSON, _ := json.Marshal(entity.Fields)
			methodsJSON, _ := json.Marshal(entity.Methods)

			_, err = s.pool.Exec(ctx, `
                INSERT INTO entities (
                    id, tenant_id, workspace_id, repository_id, analysis_id,
                    name, kind, package, file_path, commit_sha,
                    signature, fields, methods, is_exported,
                    line, line_start, line_end,
                    created_at, updated_at
                ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW(), NOW())
                ON CONFLICT (tenant_id, workspace_id, repository_id, name, package) DO UPDATE
                SET analysis_id = EXCLUDED.analysis_id,
                    fields = EXCLUDED.fields,
                    methods = EXCLUDED.methods,
                    signature = EXCLUDED.signature,
                    is_exported = EXCLUDED.is_exported,
                    line = EXCLUDED.line,
                    line_start = EXCLUDED.line_start,
                    line_end = EXCLUDED.line_end,
                    updated_at = NOW()
            `, entityID, tenantID, workspaceID, repoID, analysisID,
				entity.Name, string(entity.Kind), entity.Package, entity.File, "HEAD",
				entity.Signature, fieldsJSON, methodsJSON, entity.Exported,
				entity.Line, entity.LineStart, entity.LineEnd)
			if err != nil {
				return fmt.Errorf("failed to insert entity %s: %w", entity.Name, err)
			}
		}

		entityIDMap[entity.ID] = entityID
	}

	// 2. Insert local claims
	for _, rel := range result.Relationships {
		fromID, fromOk := entityIDMap[rel.From]
		toID, toOk := entityIDMap[rel.To]

		if !fromOk || !toOk {
			slog.Debug("Skipping claim: entity not in map",
				"from", rel.From, "fromOk", fromOk,
				"to", rel.To, "toOk", toOk)
			continue
		}

		_, err := s.pool.Exec(ctx, `
            INSERT INTO claims (
                id, tenant_id, workspace_id, repository_id, analysis_id,
                from_entity_id, to_entity_id, claim_type, confidence, created_at
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
            ON CONFLICT (tenant_id, from_entity_id, to_entity_id, claim_type) DO NOTHING
        `, uuid.New(), tenantID, workspaceID, repoID, analysisID,
			fromID, toID, rel.Type, rel.Confidence)
		if err != nil {
			return fmt.Errorf("failed to insert claim from %s to %s: %w", rel.From, rel.To, err)
		}
	}

	// 3. Detect cross-repo imports
	if err := s.detectCrossRepoImports(ctx, tenantID, workspaceID, repoID, analysisID, result); err != nil {
		slog.Warn("Cross-repo detection failed", "error", err)
	}

	return nil
}

// detectCrossRepoImports finds imports that cross repository boundaries
func (s *PostgresStore) detectCrossRepoImports(
	ctx context.Context,
	tenantID, workspaceID, repoID, analysisID uuid.UUID,
	result *analyzer.Result,
) error {
	// 1. Get module map (other repos in workspace)
	moduleMap, err := s.getModuleMap(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get module map: %w", err)
	}

	// 2. Get this repo's module path
	var thisModulePath string
	err = s.pool.QueryRow(ctx, `SELECT module_path FROM repositories WHERE id = $1`, repoID).Scan(&thisModulePath)
	if err != nil {
		slog.Warn("Repository missing module_path, skipping cross-repo detection", "repo_id", repoID)
		return nil
	}
	thisModulePath = strings.TrimSuffix(thisModulePath, "/")

	// 3. Get the file entity ID map for evidence linking
	fileEntityMap := make(map[string]uuid.UUID)
	rows, err := s.pool.Query(ctx, `
        SELECT id, file_path FROM entities
        WHERE tenant_id = $1 AND repository_id = $2 AND kind = 'file'
    `, tenantID, repoID)
	if err != nil {
		return fmt.Errorf("failed to query file entities: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var filePath string
		if err := rows.Scan(&id, &filePath); err != nil {
			continue
		}
		fileEntityMap[filePath] = id
	}

	// 4. Iterate over import relationships
	importedModules := make(map[string]bool)
	for _, rel := range result.Relationships {
		if rel.Type != string(analyzer.RelImports) {
			continue
		}

		importPath := rel.To
		if importPath == "" {
			continue
		}
		importPath = strings.TrimSuffix(importPath, "/")

		// Check if this is a local import
		if strings.HasPrefix(importPath, thisModulePath) {
			continue
		}

		// Find which repo this import belongs to
		var targetRepoID uuid.UUID
		var found bool

		for modPath, repoID := range moduleMap {
			modPath = strings.TrimSuffix(modPath, "/")
			if importPath == modPath || strings.HasPrefix(importPath, modPath+"/") {
				targetRepoID = repoID
				found = true
				break
			}
		}

		if !found {
			continue // External dependency, skip for MVP
		}

		// Deduplicate
		if importedModules[targetRepoID.String()+"_"+importPath] {
			continue
		}
		importedModules[targetRepoID.String()+"_"+importPath] = true

		// Find the from_entity_id (the file that contains the import)
		fromEntityID, ok := fileEntityMap[rel.Evidence.File]
		if !ok {
			var pkgID uuid.UUID
			err := s.pool.QueryRow(ctx, `
                SELECT id FROM entities
                WHERE tenant_id = $1 AND repository_id = $2 AND package = $3 AND kind = 'package'
            `, tenantID, repoID, rel.From).Scan(&pkgID)
			if err != nil {
				slog.Warn("Could not find from entity for cross-repo edge", "file", rel.Evidence.File)
				continue
			}
			fromEntityID = pkgID
		}

		// Try to find the target package entity
		targetPkgID, err := s.findPackageEntityInRepo(ctx, tenantID, targetRepoID, importPath)
		var toEntityID *uuid.UUID
		if err == nil && targetPkgID != nil {
			toEntityID = targetPkgID
		}

		evidenceJSON, err := json.Marshal(rel.Evidence)
		if err != nil {
			slog.Warn("Failed to marshal evidence", "error", err)
			continue
		}

		// Insert cross-repo edge
		if toEntityID != nil {
			_, err = s.pool.Exec(ctx, `
                INSERT INTO cross_repo_edges (
                    id, tenant_id, workspace_id, from_repo_id, to_repo_id,
                    from_entity_id, to_entity_id, relationship_type, evidence, resolved, created_at, updated_at
                ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true, NOW(), NOW())
                ON CONFLICT (tenant_id, from_repo_id, to_repo_id, from_entity_id, to_entity_id, relationship_type)
                DO NOTHING
            `, uuid.New(), tenantID, workspaceID, repoID, targetRepoID,
				fromEntityID, toEntityID, rel.Type, evidenceJSON)
			if err != nil {
				slog.Warn("Failed to insert cross-repo edge", "error", err)
				continue
			}
		} else {
			// Unresolved edge
			var existingID uuid.UUID
			err = s.pool.QueryRow(ctx, `
                SELECT id FROM cross_repo_edges
                WHERE tenant_id = $1 AND from_repo_id = $2 AND to_repo_id = $3
                AND from_entity_id = $4 AND to_entity_id IS NULL AND relationship_type = $5
                LIMIT 1
            `, tenantID, repoID, targetRepoID, fromEntityID, rel.Type).Scan(&existingID)
			if err == nil {
				continue // Already exists, skip
			}
			_, err = s.pool.Exec(ctx, `
                INSERT INTO cross_repo_edges (
                    id, tenant_id, workspace_id, from_repo_id, to_repo_id,
                    from_entity_id, to_entity_id, relationship_type, evidence, resolved, created_at, updated_at
                ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, false, NOW(), NOW())
            `, uuid.New(), tenantID, workspaceID, repoID, targetRepoID,
				fromEntityID, nil, rel.Type, evidenceJSON)
			if err != nil {
				slog.Warn("Failed to insert unresolved cross-repo edge", "error", err)
			}
		}
	}

	return nil
}

// GetEntity retrieves an entity by name and package within a workspace.
func (s *PostgresStore) GetEntity(ctx context.Context, tenantID, workspaceID uuid.UUID, name string) (*analyzer.Entity, error) {
	var entity analyzer.Entity
	var entityID string
	var kind string
	var pkg, file, signature string
	var exported bool
	var fieldsJSON, methodsJSON []byte
	var line, lineStart, lineEnd int

	err := s.pool.QueryRow(ctx, `
		SELECT id, name, kind, package, file_path, signature, fields, methods, is_exported,
		       line, line_start, line_end
		FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2 AND name = $3
	`, tenantID, workspaceID, name).Scan(
		&entityID, &entity.Name, &kind, &pkg, &file, &signature,
		&fieldsJSON, &methodsJSON, &exported,
		&line, &lineStart, &lineEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("entity %s not found: %w", name, err)
	}

	entity.ID = entityID
	entity.Kind = analyzer.EntityKind(kind)
	entity.Package = pkg
	entity.File = file
	entity.Signature = signature
	entity.Exported = exported
	entity.Line = line
	entity.LineStart = lineStart
	entity.LineEnd = lineEnd

	if len(fieldsJSON) > 0 {
		var fields []analyzer.Field
		if err := json.Unmarshal(fieldsJSON, &fields); err == nil {
			entity.Fields = fields
		}
	}
	if len(methodsJSON) > 0 {
		var methods []analyzer.Method
		if err := json.Unmarshal(methodsJSON, &methods); err == nil {
			entity.Methods = methods
		}
	}
	return &entity, nil
}

// GetEntityRelationships returns incoming and outgoing relationships for an entity.
func (s *PostgresStore) GetEntityRelationships(ctx context.Context, tenantID, workspaceID uuid.UUID, name, pkg string) ([]analyzer.Relationship, []analyzer.Relationship, error) {
	var entityID string
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2 AND name = $3 AND package = $4
	`, tenantID, workspaceID, name, pkg).Scan(&entityID)
	if err != nil {
		return nil, nil, fmt.Errorf("entity not found: %w", err)
	}

	var incoming []analyzer.Relationship
	var outgoing []analyzer.Relationship

	// Outgoing claims
	rows, err := s.pool.Query(ctx, `
		SELECT to_entity_id, claim_type
		FROM claims
		WHERE tenant_id = $1 AND workspace_id = $2 AND from_entity_id = $3
	`, tenantID, workspaceID, entityID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var to, typ string
			if err := rows.Scan(&to, &typ); err != nil {
				continue
			}
			outgoing = append(outgoing, analyzer.Relationship{
				To:   to,
				Type: typ, // ✅ CORRECT - typ is already a string
			})
		}
	}

	// Incoming claims
	rows2, err := s.pool.Query(ctx, `
		SELECT from_entity_id, claim_type
		FROM claims
		WHERE tenant_id = $1 AND workspace_id = $2 AND to_entity_id = $3
	`, tenantID, workspaceID, entityID)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var from, typ string
			if err := rows2.Scan(&from, &typ); err != nil {
				continue
			}
			incoming = append(incoming, analyzer.Relationship{
				From: from,
				Type: typ, // ✅ CORRECT - typ is already a string
			})
		}
	}

	return incoming, outgoing, nil
}

// ListEntities returns all entities in a workspace.
func (s *PostgresStore) ListEntities(ctx context.Context, tenantID, workspaceID uuid.UUID) ([]analyzer.Entity, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, kind, package, file_path, signature, fields, methods, is_exported,
		       line, line_start, line_end
		FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list entities: %w", err)
	}
	defer rows.Close()

	var entities []analyzer.Entity
	for rows.Next() {
		var id, name, kind, pkg, file, signature string
		var exported bool
		var fieldsJSON, methodsJSON []byte
		var line, lineStart, lineEnd int
		if err := rows.Scan(&id, &name, &kind, &pkg, &file, &signature,
			&fieldsJSON, &methodsJSON, &exported,
			&line, &lineStart, &lineEnd); err != nil {
			continue
		}
		entity := analyzer.Entity{
			ID:        id,
			Name:      name,
			Kind:      analyzer.EntityKind(kind),
			Package:   pkg,
			File:      file,
			Signature: signature,
			Exported:  exported,
			Line:      line,
			LineStart: lineStart,
			LineEnd:   lineEnd,
		}
		if len(fieldsJSON) > 0 {
			var fields []analyzer.Field
			if err := json.Unmarshal(fieldsJSON, &fields); err == nil {
				entity.Fields = fields
			}
		}
		if len(methodsJSON) > 0 {
			var methods []analyzer.Method
			if err := json.Unmarshal(methodsJSON, &methods); err == nil {
				entity.Methods = methods
			}
		}
		entities = append(entities, entity)
	}
	return entities, nil
}

// GetGraphData returns nodes and edges for graph visualisation.
func (s *PostgresStore) GetGraphData(ctx context.Context, tenantID, workspaceID uuid.UUID) ([]map[string]interface{}, []map[string]interface{}, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, kind, package, file_path, is_exported
		FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query entities: %w", err)
	}
	defer rows.Close()

	var nodes []map[string]interface{}
	for rows.Next() {
		var id, name, kind, pkg, file string
		var exported bool
		if err := rows.Scan(&id, &name, &kind, &pkg, &file, &exported); err != nil {
			continue
		}
		nodes = append(nodes, map[string]interface{}{
			"id":       id,
			"label":    name,
			"kind":     kind,
			"package":  pkg,
			"file":     file,
			"exported": exported,
		})
	}

	rows2, err := s.pool.Query(ctx, `
		SELECT from_entity_id, to_entity_id, claim_type
		FROM claims
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID)
	if err != nil {
		return nodes, nil, fmt.Errorf("failed to query claims: %w", err)
	}
	defer rows2.Close()

	var edges []map[string]interface{}
	for rows2.Next() {
		var from, to, typ string
		if err := rows2.Scan(&from, &to, &typ); err != nil {
			continue
		}
		edges = append(edges, map[string]interface{}{
			"from": from,
			"to":   to,
			"type": typ,
		})
	}

	return nodes, edges, nil
}
