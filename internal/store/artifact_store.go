package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/model"
)

// SaveArtifact inserts or updates (on conflict) an analysis artifact.
func (s *PostgresStore) SaveArtifact(ctx context.Context, artifact *model.AnalysisArtifact) error {
	// Marshal the model to JSON
	modelJSON, err := json.Marshal(artifact.Model)
	if err != nil {
		return fmt.Errorf("marshal model: %w", err)
	}

	query := `
        INSERT INTO analysis_artifacts (
            id, repository_id, commit_sha, analyzer_name, analyzer_version,
            source_fingerprint, semantic_fingerprint, schema_version, status,
            started_at, completed_at, model, error_summary, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW())
        ON CONFLICT (id) DO UPDATE
        SET status = EXCLUDED.status,
            completed_at = EXCLUDED.completed_at,
            model = EXCLUDED.model,
            error_summary = EXCLUDED.error_summary
    `
	_, err = s.pool.Exec(ctx, query,
		artifact.ID,
		artifact.RepositoryID,
		artifact.CommitSHA,
		artifact.AnalyzerName,
		artifact.AnalyzerVersion,
		artifact.SourceFingerprint,
		artifact.SemanticFingerprint,
		artifact.SchemaVersion,
		artifact.Status,
		artifact.StartedAt,
		artifact.CompletedAt,
		modelJSON,
		artifact.ErrorSummary,
	)
	if err != nil {
		return fmt.Errorf("insert artifact: %w", err)
	}
	return nil
}

// GetArtifact retrieves an artifact by ID.
func (s *PostgresStore) GetArtifact(ctx context.Context, id uuid.UUID) (*model.AnalysisArtifact, error) {
	query := `
        SELECT id, repository_id, commit_sha, analyzer_name, analyzer_version,
               source_fingerprint, semantic_fingerprint, schema_version, status,
               started_at, completed_at, model, error_summary, created_at
        FROM analysis_artifacts
        WHERE id = $1
    `
	var a model.AnalysisArtifact
	var modelJSON []byte
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.RepositoryID, &a.CommitSHA, &a.AnalyzerName, &a.AnalyzerVersion,
		&a.SourceFingerprint, &a.SemanticFingerprint, &a.SchemaVersion, &a.Status,
		&a.StartedAt, &a.CompletedAt, &modelJSON, &a.ErrorSummary, &a.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get artifact: %w", err)
	}
	// Unmarshal the model
	var snapshot model.RepositorySnapshot
	if err := json.Unmarshal(modelJSON, &snapshot); err != nil {
		return nil, fmt.Errorf("unmarshal model: %w", err)
	}
	a.Model = &snapshot
	return &a, nil
}

// GetLatestArtifact returns the most recent successful artifact for a repository and commit.
func (s *PostgresStore) GetLatestArtifact(ctx context.Context, repoID uuid.UUID, commitSHA string) (*model.AnalysisArtifact, error) {
	query := `
        SELECT id, repository_id, commit_sha, analyzer_name, analyzer_version,
               source_fingerprint, semantic_fingerprint, schema_version, status,
               started_at, completed_at, model, error_summary, created_at
        FROM analysis_artifacts
        WHERE repository_id = $1 AND commit_sha = $2 AND status = 'SUCCESS'
        ORDER BY created_at DESC
        LIMIT 1
    `
	var a model.AnalysisArtifact
	var modelJSON []byte
	err := s.pool.QueryRow(ctx, query, repoID, commitSHA).Scan(
		&a.ID, &a.RepositoryID, &a.CommitSHA, &a.AnalyzerName, &a.AnalyzerVersion,
		&a.SourceFingerprint, &a.SemanticFingerprint, &a.SchemaVersion, &a.Status,
		&a.StartedAt, &a.CompletedAt, &modelJSON, &a.ErrorSummary, &a.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get latest artifact: %w", err)
	}
	var snapshot model.RepositorySnapshot
	if err := json.Unmarshal(modelJSON, &snapshot); err != nil {
		return nil, fmt.Errorf("unmarshal model: %w", err)
	}
	a.Model = &snapshot
	return &a, nil
}

// ListArtifactsByRepo returns all artifacts for a repository, ordered by creation.
func (s *PostgresStore) ListArtifactsByRepo(ctx context.Context, repoID uuid.UUID) ([]*model.AnalysisArtifact, error) {
	query := `
        SELECT id, repository_id, commit_sha, analyzer_name, analyzer_version,
               source_fingerprint, semantic_fingerprint, schema_version, status,
               started_at, completed_at, model, error_summary, created_at
        FROM analysis_artifacts
        WHERE repository_id = $1
        ORDER BY created_at DESC
    `
	rows, err := s.pool.Query(ctx, query, repoID)
	if err != nil {
		return nil, fmt.Errorf("list artifacts: %w", err)
	}
	defer rows.Close()

	var artifacts []*model.AnalysisArtifact
	for rows.Next() {
		var a model.AnalysisArtifact
		var modelJSON []byte
		if err := rows.Scan(
			&a.ID, &a.RepositoryID, &a.CommitSHA, &a.AnalyzerName, &a.AnalyzerVersion,
			&a.SourceFingerprint, &a.SemanticFingerprint, &a.SchemaVersion, &a.Status,
			&a.StartedAt, &a.CompletedAt, &modelJSON, &a.ErrorSummary, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan artifact: %w", err)
		}
		var snapshot model.RepositorySnapshot
		if err := json.Unmarshal(modelJSON, &snapshot); err != nil {
			return nil, fmt.Errorf("unmarshal model: %w", err)
		}
		a.Model = &snapshot
		artifacts = append(artifacts, &a)
	}
	return artifacts, nil
}
