package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/knowledge"
)

// HandleCheckDrift exposes the Knowledge Drift report over Model Context Protocol (MCP)
func HandleCheckDrift(ctx context.Context, pool *pgxpool.Pool, arguments json.RawMessage) (any, error) {
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	workspace := "default"

	evaluator := knowledge.NewEvaluator(pool)
	stats, undocumented, err := evaluator.EvaluateWorkspace(ctx, tenantID, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate workspace knowledge integrity: %w", err)
	}

	response := map[string]any{
		"workspace":            workspace,
		"documentation_health": stats,
		"undocumented_symbols": undocumented,
		"status":               "success",
	}

	if stats.Contradicted > 0 {
		response["gate_verdict"] = "BLOCKED: Architectural contradictions detected. Agents must resolve contract violations before proceeding."
	} else {
		response["gate_verdict"] = "PASSED: Code conforms to declared ADR specifications."
	}

	return response, nil
}

// HandleQueryClaims allows agents to retrieve active specifications for a given symbol.
func HandleQueryClaims(ctx context.Context, pool *pgxpool.Pool, arguments json.RawMessage) (any, error) {
	var params struct {
		Subject string `json:"subject"`
	}
	if len(arguments) > 0 {
		_ = json.Unmarshal(arguments, &params)
	}

	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	workspace := "default"

	query := `
		SELECT document_title, section_title, subject, modality, predicate, object, status, contradiction_reason
		FROM document_claims
		WHERE tenant_id = $1 AND workspace = $2
	`
	args := []any{tenantID, workspace}

	if params.Subject != "" {
		query += " AND (subject ILIKE $3 OR object ILIKE $3)"
		args = append(args, "%"+params.Subject+"%")
	}
	query += " ORDER BY line_start ASC"

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query claims: %w", err)
	}
	defer rows.Close()

	type ClaimView struct {
		Document      string `json:"document"`
		Section       string `json:"section"`
		Subject       string `json:"subject"`
		Modality      string `json:"modality"`
		Predicate     string `json:"predicate"`
		Object        string `json:"object"`
		Status        string `json:"status"`
		Contradiction string `json:"contradiction_reason,omitempty"`
	}

	var results []ClaimView
	for rows.Next() {
		var v ClaimView
		if err := rows.Scan(&v.Document, &v.Section, &v.Subject, &v.Modality, &v.Predicate, &v.Object, &v.Status, &v.Contradiction); err == nil {
			results = append(results, v)
		}
	}

	return results, nil
}
