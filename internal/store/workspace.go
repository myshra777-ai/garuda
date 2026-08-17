package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

var ErrWorkspaceExists = errors.New("workspace with this name already exists")

type Workspace struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// getTenantIDFromWorkspace retrieves the tenant ID for a given workspace.
func (s *PostgresStore) getTenantIDFromWorkspace(ctx context.Context, workspaceID uuid.UUID) (uuid.UUID, error) {
	var tenantID uuid.UUID
	err := s.pool.QueryRow(ctx, `SELECT tenant_id FROM workspaces WHERE id = $1`, workspaceID).Scan(&tenantID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get tenant ID for workspace: %w", err)
	}
	return tenantID, nil
}

// CreateWorkspace creates a new workspace or returns existing one.
func (s *PostgresStore) CreateWorkspace(ctx context.Context, tenantIDStr, name, description string) (*Workspace, error) {
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant UUID: %w", err)
	}

	var (
		id        uuid.UUID
		createdAt time.Time
		updatedAt time.Time
		inserted  bool
	)

	query := `
		WITH ins AS (
			INSERT INTO workspaces (id, tenant_id, name, description, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
			ON CONFLICT (tenant_id, name) DO NOTHING
			RETURNING id, created_at, updated_at, true AS inserted
		)
		SELECT id, created_at, updated_at, inserted FROM ins
		UNION ALL
		SELECT id, created_at, updated_at, false AS inserted 
		FROM workspaces 
		WHERE tenant_id = $2 AND name = $3
		LIMIT 1;
	`

	newID := uuid.New()
	err = s.pool.QueryRow(ctx, query, newID, tenantID, name, description).Scan(&id, &createdAt, &updatedAt, &inserted)
	if err != nil {
		return nil, fmt.Errorf("failed to execute create workspace query: %w", err)
	}

	ws := &Workspace{
		ID:          id,
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	if !inserted {
		return ws, ErrWorkspaceExists
	}

	return ws, nil
}

// ListWorkspaces lists all workspaces for a tenant.
func (s *PostgresStore) ListWorkspaces(ctx context.Context, tenantIDStr string) ([]*Workspace, error) {
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant UUID: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, tenant_id, name, COALESCE(description, ''), created_at, updated_at
		FROM workspaces WHERE tenant_id = $1 ORDER BY name ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*Workspace
	for rows.Next() {
		var w Workspace
		if err := rows.Scan(&w.ID, &w.TenantID, &w.Name, &w.Description, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, &w)
	}
	return workspaces, nil
}

// AddRepository adds a repository to a workspace.
func (s *PostgresStore) AddRepository(
	ctx context.Context,
	workspaceID uuid.UUID,
	provider, url, defaultBranch, language, modulePath string,
) (*Repository, error) {
	if _, err := s.getTenantIDFromWorkspace(ctx, workspaceID); err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	id := uuid.New()
	_, err := s.pool.Exec(ctx, `
        INSERT INTO repositories (
            id, workspace_id, provider, url, default_branch, language, module_path,
            enabled, analysis_status, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, true, 'pending', NOW(), NOW())
        ON CONFLICT (workspace_id, url) DO UPDATE
        SET provider = EXCLUDED.provider,
            default_branch = EXCLUDED.default_branch,
            language = EXCLUDED.language,
            module_path = EXCLUDED.module_path,
            updated_at = NOW()
    `, id, workspaceID, provider, url, defaultBranch, language, modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to add repository: %w", err)
	}

	return &Repository{
		ID:             id,
		WorkspaceID:    workspaceID,
		Provider:       provider,
		URL:            url,
		DefaultBranch:  defaultBranch,
		Language:       language,
		ModulePath:     modulePath,
		Enabled:        true,
		AnalysisStatus: "pending",
		CurrentCommit:  nil,
		LastAnalyzedAt: nil,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}, nil
}

// ListRepositories lists all repositories in a workspace.
func (s *PostgresStore) ListRepositories(ctx context.Context, workspaceID uuid.UUID) ([]*Repository, error) {
	var repos []*Repository
	query := `
		SELECT id, workspace_id, provider, url, default_branch, language,
		       current_commit, enabled, analysis_status, last_analyzed_at,
		       created_at, updated_at
		FROM repositories
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var repo Repository
		var currentCommit *string
		err := rows.Scan(
			&repo.ID, &repo.WorkspaceID, &repo.Provider, &repo.URL, &repo.DefaultBranch,
			&repo.Language, &currentCommit, &repo.Enabled, &repo.AnalysisStatus,
			&repo.LastAnalyzedAt, &repo.CreatedAt, &repo.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan repository: %w", err)
		}
		repo.CurrentCommit = currentCommit
		repos = append(repos, &repo)
	}
	return repos, nil
}

// UpdateRepositorySyncStatus updates the sync status and commit SHA of a repository.
func (s *PostgresStore) UpdateRepositorySyncStatus(ctx context.Context, tenantIDStr string, repoID uuid.UUID, commitSHA, status string) error {
	if _, err := uuid.Parse(tenantIDStr); err != nil {
		return fmt.Errorf("invalid tenant ID: %w", err)
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE repositories
		SET current_commit = $1, analysis_status = $2, last_analyzed_at = NOW()
		WHERE id = $3
	`, commitSHA, status, repoID)
	if err != nil {
		return fmt.Errorf("failed to update repository sync status: %w", err)
	}
	return nil
}

// UpdateRepositoryModulePath updates the module path of a repository.
func (s *PostgresStore) UpdateRepositoryModulePath(ctx context.Context, repoID uuid.UUID, modulePath string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE repositories SET module_path = $1 WHERE id = $2
	`, modulePath, repoID)
	if err != nil {
		return fmt.Errorf("failed to update module path: %w", err)
	}
	return nil
}

// SyncWorkspace executes a sync function for each repository in a workspace.
func (s *PostgresStore) SyncWorkspace(ctx context.Context, workspaceID uuid.UUID, syncFunc func(repo *Repository, path string) error) error {
	repos, err := s.ListRepositories(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}
	for _, repo := range repos {
		if !repo.Enabled {
			continue
		}
		tempDir := filepath.Join(os.TempDir(), "garuda-sync", workspaceID.String(), repo.ID.String())
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return fmt.Errorf("failed to create temp dir: %w", err)
		}
		if err := syncFunc(repo, tempDir); err != nil {
			// Log error but continue with other repos
			continue
		}
	}
	return nil
}
