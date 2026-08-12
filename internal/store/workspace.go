package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Repository struct {
	ID             uuid.UUID  `json:"id"`
	WorkspaceID    uuid.UUID  `json:"workspace_id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	Provider       string     `json:"provider"`
	URL            string     `json:"url"`
	DefaultBranch  string     `json:"default_branch"`
	Language       string     `json:"language"`
	CurrentCommit  string     `json:"current_commit"`
	Enabled        bool       `json:"enabled"`
	AnalysisStatus string     `json:"analysis_status"`
	LastAnalyzedAt *time.Time `json:"last_analyzed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ErrWorkspaceExists is returned when attempting to create a workspace that already exists.
var ErrWorkspaceExists = errors.New("workspace with this name already exists")

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

	// Insert new workspace, or if name exists under tenant, fetch existing details
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

func (s *PostgresStore) AddRepository(ctx context.Context, workspaceID, provider, url, defaultBranch, language, tenantID string) (*Repository, error) {
	repoID := uuid.New()
	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace UUID: %w", err)
	}
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant UUID: %w", err)
	}

	// Verify workspace exists and belongs to tenant
	var count int
	err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM workspaces WHERE id = $1 AND tenant_id = $2`, workspaceUUID, tenantUUID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("check workspace: %w", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("workspace '%s' not found for tenant", workspaceID)
	}

	query := `
        INSERT INTO repositories (id, workspace_id, tenant_id, provider, url, default_branch, language, enabled, analysis_status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, true, 'pending', NOW(), NOW())
        ON CONFLICT (workspace_id, url) DO UPDATE SET
            provider = EXCLUDED.provider,
            tenant_id = EXCLUDED.tenant_id,
            default_branch = EXCLUDED.default_branch,
            language = EXCLUDED.language,
            enabled = EXCLUDED.enabled,
            analysis_status = EXCLUDED.analysis_status,
            updated_at = NOW()
        RETURNING id, workspace_id, tenant_id, provider, url, default_branch, COALESCE(language, ''), COALESCE(current_commit, ''), enabled, analysis_status, last_analyzed_at, created_at, updated_at
    `

	var repo Repository
	if err := s.pool.QueryRow(ctx, query, repoID, workspaceUUID, tenantUUID, provider, url, defaultBranch, language).Scan(
		&repo.ID, &repo.WorkspaceID, &repo.TenantID, &repo.Provider, &repo.URL,
		&repo.DefaultBranch, &repo.Language, &repo.CurrentCommit, &repo.Enabled,
		&repo.AnalysisStatus, &repo.LastAnalyzedAt, &repo.CreatedAt, &repo.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("insert repository: %w", err)
	}
	return &repo, nil
}

func (s *PostgresStore) ListRepositories(ctx context.Context, tenantIDStr, workspaceName string) ([]*Repository, error) {
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant UUID: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT r.id, r.workspace_id, r.tenant_id, r.provider, r.url, r.default_branch, 
		       COALESCE(r.language, 'go'), COALESCE(r.current_commit, ''), r.enabled, r.analysis_status, 
		       r.last_analyzed_at, r.created_at, r.updated_at
		FROM repositories r
		JOIN workspaces w ON r.workspace_id = w.id
		WHERE r.tenant_id = $1 AND w.name = $2
		ORDER BY r.url ASC
	`, tenantID, workspaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}
	defer rows.Close()

	var repos []*Repository
	for rows.Next() {
		var r Repository
		if err := rows.Scan(
			&r.ID, &r.WorkspaceID, &r.TenantID, &r.Provider, &r.URL, &r.DefaultBranch,
			&r.Language, &r.CurrentCommit, &r.Enabled, &r.AnalysisStatus,
			&r.LastAnalyzedAt, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan repository: %w", err)
		}
		repos = append(repos, &r)
	}
	return repos, nil
}
