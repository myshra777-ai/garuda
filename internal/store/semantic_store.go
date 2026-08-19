// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/analyzer"
)

// moduleCache caches module paths per workspace to avoid repeated DB queries
type moduleCache struct {
	mu      sync.RWMutex
	cache   map[string]map[string]uuid.UUID
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

func (c *moduleCache) Invalidate(workspaceID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, workspaceID)
	delete(c.expires, workspaceID)
}

var crossRepoCache = newModuleCache(5 * time.Minute)

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

func (s *PostgresStore) findPackageEntityInRepo(ctx context.Context, tenantID, repoID uuid.UUID, packagePath string) (*uuid.UUID, error) {
	var entityID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM entities
		WHERE tenant_id = $1 AND repository_id = $2 
		  AND (package_path = $3 OR package = $3) 
		  AND kind = 'package'
		LIMIT 1
	`, tenantID, repoID, packagePath).Scan(&entityID)
	if err != nil {
		return nil, err
	}
	return &entityID, nil
}

func GenerateCanonicalEntityID(tenantID, workspaceID, repoID uuid.UUID, entity *analyzer.Entity) uuid.UUID {
	canonicalKey := fmt.Sprintf("%s|%s|%s|%s|%s",
		strings.TrimSpace(entity.ModulePath),
		strings.TrimSpace(entity.PackagePath),
		strings.TrimSpace(string(entity.Kind)),
		strings.TrimSpace(entity.ReceiverType),
		strings.TrimSpace(entity.Name),
	)
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(
		tenantID.String()+"/"+workspaceID.String()+"/"+repoID.String()+"/"+canonicalKey,
	))
}

func (s *PostgresStore) SaveEntities(
	ctx context.Context,
	tenantID, workspaceID, repoID, analysisID uuid.UUID,
	result *analyzer.Result,
	commitSHA string,
) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for SaveEntities: %w", err)
	}
	defer tx.Rollback(ctx)

	for i := range result.Entities {
		entity := &result.Entities[i]
		entityID := GenerateCanonicalEntityID(tenantID, workspaceID, repoID, entity)

		fieldsJSON, err := json.Marshal(entity.Fields)
		if err != nil || len(fieldsJSON) == 0 {
			fieldsJSON = []byte("[]")
		}
		methodsJSON, err := json.Marshal(entity.Methods)
		if err != nil || len(methodsJSON) == 0 {
			methodsJSON = []byte("[]")
		}

		line := entity.Line
		if line == 0 && entity.LineStart > 0 {
			line = entity.LineStart
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO entities (
				id, tenant_id, workspace_id, repository_id, analysis_id,
				name, kind, package, package_path, module_path, receiver_type,
				file_path, commit_sha, line, line_start, line_end,
				signature, fields, methods, is_exported,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9, $10, $11,
				$12, $13, $14, $15, $16,
				$17, $18, $19, $20,
				NOW(), NOW()
			)
			ON CONFLICT (id) DO UPDATE SET
				analysis_id   = EXCLUDED.analysis_id,
				name          = EXCLUDED.name,
				kind          = EXCLUDED.kind,
				package       = EXCLUDED.package,
				package_path  = EXCLUDED.package_path,
				module_path   = EXCLUDED.module_path,
				receiver_type = EXCLUDED.receiver_type,
				file_path     = EXCLUDED.file_path,
				commit_sha    = EXCLUDED.commit_sha,
				line          = EXCLUDED.line,
				line_start    = EXCLUDED.line_start,
				line_end      = EXCLUDED.line_end,
				signature     = EXCLUDED.signature,
				fields        = EXCLUDED.fields,
				methods       = EXCLUDED.methods,
				is_exported   = EXCLUDED.is_exported,
				updated_at    = NOW()
		`,
			entityID, tenantID, workspaceID, repoID, analysisID,
			entity.Name, string(entity.Kind), entity.Package,
			entity.PackagePath, entity.ModulePath, entity.ReceiverType,
			entity.File, commitSHA, line, entity.LineStart, entity.LineEnd,
			entity.Signature, fieldsJSON, methodsJSON, entity.Exported,
		)
		if err != nil {
			return fmt.Errorf("failed to upsert entity '%s.%s' (%s): %w", entity.Package, entity.Name, entity.Kind, err)
		}
	}

	return tx.Commit(ctx)
}

func (s *PostgresStore) SaveSemanticGraph(
	ctx context.Context,
	tenantID, workspaceID, repoID, analysisID uuid.UUID,
	result *analyzer.Result,
	commitSHA string,
) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin semantic graph transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	entityIDMap := make(map[string]uuid.UUID, len(result.Entities))

	for i := range result.Entities {
		entity := &result.Entities[i]
		entityID := GenerateCanonicalEntityID(tenantID, workspaceID, repoID, entity)

		fieldsJSON, err := json.Marshal(entity.Fields)
		if err != nil || len(fieldsJSON) == 0 {
			fieldsJSON = []byte("[]")
		}
		methodsJSON, err := json.Marshal(entity.Methods)
		if err != nil || len(methodsJSON) == 0 {
			methodsJSON = []byte("[]")
		}

		line := entity.Line
		if line == 0 && entity.LineStart > 0 {
			line = entity.LineStart
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO entities (
				id, tenant_id, workspace_id, repository_id, analysis_id,
				name, kind, package, package_path, module_path, receiver_type,
				file_path, commit_sha, line, line_start, line_end,
				signature, fields, methods, is_exported,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9, $10, $11,
				$12, $13, $14, $15, $16,
				$17, $18, $19, $20,
				NOW(), NOW()
			)
			ON CONFLICT (id) DO UPDATE SET
				analysis_id   = EXCLUDED.analysis_id,
				name          = EXCLUDED.name,
				kind          = EXCLUDED.kind,
				package       = EXCLUDED.package,
				package_path  = EXCLUDED.package_path,
				module_path   = EXCLUDED.module_path,
				receiver_type = EXCLUDED.receiver_type,
				file_path     = EXCLUDED.file_path,
				commit_sha    = EXCLUDED.commit_sha,
				line          = EXCLUDED.line,
				line_start    = EXCLUDED.line_start,
				line_end      = EXCLUDED.line_end,
				signature     = EXCLUDED.signature,
				fields        = EXCLUDED.fields,
				methods       = EXCLUDED.methods,
				is_exported   = EXCLUDED.is_exported,
				updated_at    = NOW()
		`,
			entityID, tenantID, workspaceID, repoID, analysisID,
			entity.Name, string(entity.Kind), entity.Package,
			entity.PackagePath, entity.ModulePath, entity.ReceiverType,
			entity.File, commitSHA, line, entity.LineStart, entity.LineEnd,
			entity.Signature, fieldsJSON, methodsJSON, entity.Exported,
		)
		if err != nil {
			return fmt.Errorf("failed to upsert entity %s (%s): %w", entity.Name, entity.Kind, err)
		}

		entityIDMap[entity.ID] = entityID
	}

	for _, rel := range result.Relationships {
		fromID, fromOk := entityIDMap[rel.From]
		toID, toOk := entityIDMap[rel.To]

		if !fromOk || !toOk {
			slog.Debug("Skipping claim edge: entity unresolved in snapshot",
				"from", rel.From, "fromOk", fromOk,
				"to", rel.To, "toOk", toOk)
			continue
		}

		epistemicClass := rel.EpistemicClass
		if epistemicClass == "" {
			epistemicClass = "OBSERVATION"
		}
		confidence := rel.Confidence
		if confidence <= 0 {
			confidence = 1.0
		}

		filePath := rel.Evidence.File
		lineStart := rel.Evidence.LineStart
		if lineStart == 0 {
			lineStart = rel.Evidence.Line
		}
		lineEnd := rel.Evidence.LineEnd
		if lineEnd == 0 {
			lineEnd = lineStart
		}

		claimID := uuid.New()
		_, err := tx.Exec(ctx, `
			INSERT INTO claims (
				id, tenant_id, workspace_id, repository_id, analysis_id,
				from_entity_id, to_entity_id, claim_type, epistemic_class, confidence,
				file_path, commit_sha, line_start, line_end,
				created_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9, $10,
				$11, $12, $13, $14,
				NOW()
			)
			ON CONFLICT (workspace_id, from_entity_id, to_entity_id, claim_type) DO UPDATE SET
				analysis_id     = EXCLUDED.analysis_id,
				epistemic_class = EXCLUDED.epistemic_class,
				confidence      = EXCLUDED.confidence,
				file_path       = EXCLUDED.file_path,
				commit_sha      = EXCLUDED.commit_sha,
				line_start      = EXCLUDED.line_start,
				line_end        = EXCLUDED.line_end
		`,
			claimID, tenantID, workspaceID, repoID, analysisID,
			fromID, toID, rel.Type, epistemicClass, confidence,
			filePath, commitSHA, lineStart, lineEnd,
		)
		if err != nil {
			return fmt.Errorf("failed to insert claim from %s to %s: %w", rel.From, rel.To, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit semantic graph transaction: %w", err)
	}

	if err := s.detectCrossRepoImports(ctx, tenantID, workspaceID, repoID, analysisID, result); err != nil {
		slog.Warn("Cross-repo dependency detection failed", "error", err)
	}

	return nil
}

func (s *PostgresStore) detectCrossRepoImports(
	ctx context.Context,
	tenantID, workspaceID, repoID, analysisID uuid.UUID,
	result *analyzer.Result,
) error {
	moduleMap, err := s.getModuleMap(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get module map: %w", err)
	}

	var thisModulePath string
	err = s.pool.QueryRow(ctx, `SELECT module_path FROM repositories WHERE id = $1`, repoID).Scan(&thisModulePath)
	if err != nil {
		slog.Warn("Repository missing module_path, skipping cross-repo detection", "repo_id", repoID)
		return nil
	}
	thisModulePath = strings.TrimSuffix(thisModulePath, "/")

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

	importedModules := make(map[string]bool)
	for _, rel := range result.Relationships {
		if rel.Type != string(analyzer.RelImports) {
			continue
		}

		importPath := strings.TrimSuffix(rel.To, "/")
		if importPath == "" || strings.HasPrefix(importPath, thisModulePath) {
			continue
		}

		var targetRepoID uuid.UUID
		var found bool

		for modPath, targetID := range moduleMap {
			modPath = strings.TrimSuffix(modPath, "/")
			if importPath == modPath || strings.HasPrefix(importPath, modPath+"/") {
				targetRepoID = targetID
				found = true
				break
			}
		}

		if !found {
			continue
		}

		dedupKey := targetRepoID.String() + "_" + importPath
		if importedModules[dedupKey] {
			continue
		}
		importedModules[dedupKey] = true

		fromEntityID, ok := fileEntityMap[rel.Evidence.File]
		if !ok {
			var pkgID uuid.UUID
			err := s.pool.QueryRow(ctx, `
				SELECT id FROM entities
				WHERE tenant_id = $1 AND repository_id = $2 
				  AND (package_path = $3 OR package = $3) 
				  AND kind = 'package'
				LIMIT 1
			`, tenantID, repoID, rel.From).Scan(&pkgID)
			if err != nil {
				slog.Debug("Could not find from entity for cross-repo edge", "file", rel.Evidence.File)
				continue
			}
			fromEntityID = pkgID
		}

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

		if toEntityID != nil {
			_, err = s.pool.Exec(ctx, `
				INSERT INTO cross_repo_edges (
					id, tenant_id, workspace_id, from_repo_id, to_repo_id,
					from_entity_id, to_entity_id, relationship_type, evidence, resolved, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true, NOW(), NOW())
				ON CONFLICT (tenant_id, from_repo_id, to_repo_id, from_entity_id, to_entity_id, relationship_type)
				DO UPDATE SET evidence = EXCLUDED.evidence, updated_at = NOW()
			`, uuid.New(), tenantID, workspaceID, repoID, targetRepoID,
				fromEntityID, toEntityID, rel.Type, evidenceJSON)
			if err != nil {
				slog.Warn("Failed to insert cross-repo edge", "error", err)
			}
		} else {
			var existingID uuid.UUID
			err = s.pool.QueryRow(ctx, `
				SELECT id FROM cross_repo_edges
				WHERE tenant_id = $1 AND from_repo_id = $2 AND to_repo_id = $3
				AND from_entity_id = $4 AND to_entity_id IS NULL AND relationship_type = $5
				LIMIT 1
			`, tenantID, repoID, targetRepoID, fromEntityID, rel.Type).Scan(&existingID)
			if err == nil {
				continue
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

func (s *PostgresStore) GetEntity(ctx context.Context, tenantID, workspaceID uuid.UUID, pkg, name string) (*analyzer.Entity, error) {
	var entity analyzer.Entity
	var entityID, kind, pkgName, pkgPath, modPath, recType, file, signature string
	var exported bool
	var fieldsJSON, methodsJSON []byte
	var line, lineStart, lineEnd int

	err := s.pool.QueryRow(ctx, `
		SELECT id, name, kind, package, package_path, module_path, receiver_type,
		       file_path, signature, fields, methods, is_exported,
		       line, line_start, line_end
		FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2 AND package = $3 AND name = $4
		LIMIT 1
	`, tenantID, workspaceID, pkg, name).Scan(
		&entityID, &entity.Name, &kind, &pkgName, &pkgPath, &modPath, &recType,
		&file, &signature, &fieldsJSON, &methodsJSON, &exported,
		&line, &lineStart, &lineEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("entity %s.%s not found: %w", pkg, name, err)
	}

	entity.ID = entityID
	entity.Kind = analyzer.EntityKind(kind)
	entity.Package = pkgName
	entity.PackagePath = pkgPath
	entity.ModulePath = modPath
	entity.ReceiverType = recType
	entity.File = file
	entity.Signature = signature
	entity.Exported = exported
	entity.Line = line
	entity.LineStart = lineStart
	entity.LineEnd = lineEnd

	if len(fieldsJSON) > 0 {
		_ = json.Unmarshal(fieldsJSON, &entity.Fields)
	}
	if len(methodsJSON) > 0 {
		_ = json.Unmarshal(methodsJSON, &entity.Methods)
	}
	return &entity, nil
}

func (s *PostgresStore) GetEntityRelationships(ctx context.Context, tenantID, workspaceID uuid.UUID, entityID string) ([]analyzer.Relationship, []analyzer.Relationship, error) {
	var incoming, outgoing []analyzer.Relationship

	rows, err := s.pool.Query(ctx, `
		SELECT to_entity_id, claim_type, confidence
		FROM claims
		WHERE tenant_id = $1 AND workspace_id = $2 AND from_entity_id = $3
	`, tenantID, workspaceID, entityID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var to, typ string
			var conf float64
			if err := rows.Scan(&to, &typ, &conf); err == nil {
				outgoing = append(outgoing, analyzer.Relationship{
					To:         to,
					Type:       typ,
					Confidence: conf,
				})
			}
		}
	}

	rows2, err := s.pool.Query(ctx, `
		SELECT from_entity_id, claim_type, confidence
		FROM claims
		WHERE tenant_id = $1 AND workspace_id = $2 AND to_entity_id = $3
	`, tenantID, workspaceID, entityID)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var from, typ string
			var conf float64
			if err := rows2.Scan(&from, &typ, &conf); err == nil {
				incoming = append(incoming, analyzer.Relationship{
					From:       from,
					Type:       typ,
					Confidence: conf,
				})
			}
		}
	}

	return incoming, outgoing, nil
}

func (s *PostgresStore) ListEntities(ctx context.Context, tenantID, workspaceID uuid.UUID) ([]analyzer.Entity, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, kind, package, package_path, module_path, receiver_type,
		       file_path, signature, fields, methods, is_exported,
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
		var id, name, kind, pkg, pkgPath, modPath, recType, file, signature string
		var exported bool
		var fieldsJSON, methodsJSON []byte
		var line, lineStart, lineEnd int

		if err := rows.Scan(
			&id, &name, &kind, &pkg, &pkgPath, &modPath, &recType,
			&file, &signature, &fieldsJSON, &methodsJSON, &exported,
			&line, &lineStart, &lineEnd,
		); err != nil {
			continue
		}

		entity := analyzer.Entity{
			ID:           id,
			Name:         name,
			Kind:         analyzer.EntityKind(kind),
			Package:      pkg,
			PackagePath:  pkgPath,
			ModulePath:   modPath,
			ReceiverType: recType,
			File:         file,
			Signature:    signature,
			Exported:     exported,
			Line:         line,
			LineStart:    lineStart,
			LineEnd:      lineEnd,
		}
		if len(fieldsJSON) > 0 {
			_ = json.Unmarshal(fieldsJSON, &entity.Fields)
		}
		if len(methodsJSON) > 0 {
			_ = json.Unmarshal(methodsJSON, &entity.Methods)
		}
		entities = append(entities, entity)
	}
	return entities, nil
}

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
