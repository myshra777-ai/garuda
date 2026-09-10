// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package policy

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Evaluator runs a Policy against a workspace.
// It is stateless — every call opens its own DB queries.
type Evaluator struct {
	pool *pgxpool.Pool
}

func NewEvaluator(pool *pgxpool.Pool) *Evaluator {
	return &Evaluator{pool: pool}
}

// Evaluate runs one policy against a workspace.
// Returns the evaluation result (decision + evidence + reason).
// The evaluation is not persisted here — that's the engine's job.
func (e *Evaluator) Evaluate(
	ctx context.Context,
	p *Policy,
	tenantID, workspaceID uuid.UUID,
	subjectKind, subjectID, actor string,
) (*Evaluation, error) {

	ev := &Evaluation{
		ID:            uuid.New(),
		TenantID:      tenantID,
		WorkspaceID:   workspaceID,
		PolicyVersion: p.Version,
		SubjectKind:   subjectKind,
		SubjectID:     subjectID,
		Actor:         actor,
	}

	// Default: no predicate fired → ALLOW with reason
	ev.Decision = DecisionAllow
	ev.Reason = "no predicate matched"

	// All predicates must match (AND). First predicate that fails → stop.
	allMatched := true
	for i, pred := range p.When {
		matched, evidence, err := e.evaluatePredicate(ctx, p, &pred, tenantID, workspaceID)
		if err != nil {
			return nil, fmt.Errorf("predicate %d (%s): %w", i, pred.Type, err)
		}
		if !matched {
			allMatched = false
			break
		}
		ev.Evidence.MatchedPredicates = append(ev.Evidence.MatchedPredicates, pred.Type)
		ev.Evidence.ClaimIDs = append(ev.Evidence.ClaimIDs, evidence.ClaimIDs...)
		ev.Evidence.EntityIDs = append(ev.Evidence.EntityIDs, evidence.EntityIDs...)
		ev.Evidence.ContradictionIDs = append(ev.Evidence.ContradictionIDs, evidence.ContradictionIDs...)
		ev.Evidence.ReasoningNotes = append(ev.Evidence.ReasoningNotes, evidence.ReasoningNotes...)
	}

	if allMatched {
		ev.Decision = p.Then.Decision
		ev.Reason = strings.TrimSpace(p.Then.Reason)
	}

	return ev, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Predicate dispatch
// ─────────────────────────────────────────────────────────────────────────────

func (e *Evaluator) evaluatePredicate(
	ctx context.Context,
	p *Policy,
	pred *Predicate,
	tenantID, workspaceID uuid.UUID,
) (bool, Evidence, error) {

	switch pred.Type {
	case "entity_exists":
		return e.evalEntityExists(ctx, p, pred, tenantID, workspaceID)
	case "claim_exists":
		return e.evalClaimExists(ctx, p, pred, tenantID, workspaceID)
	case "contradiction_exists":
		return e.evalContradictionExists(ctx, p, pred, tenantID, workspaceID)
	case "verification_missing":
		return e.evalVerificationMissing(ctx, p, pred, tenantID, workspaceID)
	case "language_matches":
		return e.evalLanguageMatches(ctx, p, pred, tenantID, workspaceID)
	default:
		return false, Evidence{}, fmt.Errorf("unsupported predicate type: %s", pred.Type)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Predicate: entity_exists
//
// Fires when at least one entity matches the given filters.
//
// Params:
//   kind              (string, optional)  — entity kind, e.g. "struct"
//   name_pattern      (string, optional)  — regex against entity.name
//   exported          (bool,   optional)  — require is_exported = true
//   limit             (int,    optional)  — cap on evidence collected (default 10)
// ─────────────────────────────────────────────────────────────────────────────

func (e *Evaluator) evalEntityExists(
	ctx context.Context,
	p *Policy,
	pred *Predicate,
	tenantID, workspaceID uuid.UUID,
) (bool, Evidence, error) {

	kind := paramString(pred.Params, "kind")
	namePat := paramString(pred.Params, "name_pattern")
	exported := paramBool(pred.Params, "exported")
	limit := paramInt(pred.Params, "limit", 10)

	// Compile regex once
	if namePat != "" {
		if _, err := regexp.Compile(namePat); err != nil {
			return false, Evidence{}, fmt.Errorf("invalid name_pattern %q: %w", namePat, err)
		}
	}

	q := `
		SELECT id, name, kind, COALESCE(package, ''), COALESCE(file_path, '')
		FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2 AND kind != 'external'
	`
	args := []any{tenantID, workspaceID}
	argPos := 3

	if kind != "" {
		q += fmt.Sprintf(" AND kind = $%d", argPos)
		args = append(args, kind)
		argPos++
	}
	if exported {
		q += " AND is_exported = true"
	}
	if p.Scope.Package != "" {
		q += fmt.Sprintf(" AND package LIKE '%%' || $%d || '%%'", argPos)
		args = append(args, p.Scope.Package)
		argPos++
	}
	if namePat != "" {
		q += fmt.Sprintf(" AND name ~ $%d", argPos)
		args = append(args, namePat)
		argPos++
	}
	q += " LIMIT " + fmt.Sprint(limit)

	rows, err := e.pool.Query(ctx, q, args...)
	if err != nil {
		return false, Evidence{}, fmt.Errorf("entity_exists query: %w", err)
	}
	defer rows.Close()

	var ev Evidence
	for rows.Next() {
		var id uuid.UUID
		var name, kind, pkg, file string
		if err := rows.Scan(&id, &name, &kind, &pkg, &file); err != nil {
			continue
		}
		ev.EntityIDs = append(ev.EntityIDs, id)
		ev.ReasoningNotes = append(ev.ReasoningNotes,
			fmt.Sprintf("entity %s.%s (%s) at %s", pkg, name, kind, file))
	}

	if len(ev.EntityIDs) == 0 {
		return false, Evidence{}, nil
	}
	return true, ev, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Predicate: claim_exists
//
// Fires when at least one claim matches the given filters.
//
// Params:
//   claim_type         (string, optional) — e.g. "IMPORTS", "CALLS", "INHERITS"
//   from_name_pattern  (string, optional) — regex against the source entity name
//   to_name_pattern    (string, optional) — regex against the target entity name
//   limit              (int,    optional) — default 10
// ─────────────────────────────────────────────────────────────────────────────

func (e *Evaluator) evalClaimExists(
	ctx context.Context,
	p *Policy,
	pred *Predicate,
	tenantID, workspaceID uuid.UUID,
) (bool, Evidence, error) {

	claimType := paramString(pred.Params, "claim_type")
	fromPat := paramString(pred.Params, "from_name_pattern")
	toPat := paramString(pred.Params, "to_name_pattern")
	limit := paramInt(pred.Params, "limit", 10)

	q := `
		SELECT c.id, c.from_entity_id, c.to_entity_id,
		       e1.name AS from_name, COALESCE(e2.name, '') AS to_name
		FROM claims c
		JOIN entities e1 ON e1.id = c.from_entity_id
		LEFT JOIN entities e2 ON e2.id = c.to_entity_id
		WHERE c.tenant_id = $1 AND c.workspace_id = $2
	`
	args := []any{tenantID, workspaceID}
	argPos := 3

	if claimType != "" {
		q += fmt.Sprintf(" AND UPPER(c.claim_type) = UPPER($%d)", argPos)
		args = append(args, claimType)
		argPos++
	}
	if fromPat != "" {
		if _, err := regexp.Compile(fromPat); err != nil {
			return false, Evidence{}, fmt.Errorf("invalid from_name_pattern: %w", err)
		}
		q += fmt.Sprintf(" AND e1.name ~ $%d", argPos)
		args = append(args, fromPat)
		argPos++
	}
	if toPat != "" {
		if _, err := regexp.Compile(toPat); err != nil {
			return false, Evidence{}, fmt.Errorf("invalid to_name_pattern: %w", err)
		}
		q += fmt.Sprintf(" AND e2.name ~ $%d", argPos)
		args = append(args, toPat)
		argPos++
	}
	q += " LIMIT " + fmt.Sprint(limit)

	rows, err := e.pool.Query(ctx, q, args...)
	if err != nil {
		return false, Evidence{}, fmt.Errorf("claim_exists query: %w", err)
	}
	defer rows.Close()

	var ev Evidence
	for rows.Next() {
		var id, fromID uuid.UUID
		var toID *uuid.UUID
		var fromName, toName string
		if err := rows.Scan(&id, &fromID, &toID, &fromName, &toName); err != nil {
			continue
		}
		ev.ClaimIDs = append(ev.ClaimIDs, id)
		ev.EntityIDs = append(ev.EntityIDs, fromID)
		if toID != nil {
			ev.EntityIDs = append(ev.EntityIDs, *toID)
		}
		ev.ReasoningNotes = append(ev.ReasoningNotes,
			fmt.Sprintf("claim %s: %s -> %s", claimType, fromName, toName))
	}

	if len(ev.ClaimIDs) == 0 {
		return false, Evidence{}, nil
	}
	return true, ev, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Predicate: contradiction_exists
//
// Fires when a CONTRADICTED claim_verification exists in the workspace.
//
// Params:
//   limit (int, optional) — default 5
// ─────────────────────────────────────────────────────────────────────────────

func (e *Evaluator) evalContradictionExists(
	ctx context.Context,
	p *Policy,
	pred *Predicate,
	tenantID, workspaceID uuid.UUID,
) (bool, Evidence, error) {

	limit := paramInt(pred.Params, "limit", 5)

	rows, err := e.pool.Query(ctx, `
		SELECT cv.id, cv.source_entity_id,
		       COALESCE(cv.evidence_payload->>'raw_target', ''),
		       COALESCE(cv.reason, '')
		FROM claim_verifications cv
		WHERE cv.tenant_id = $1
		  AND cv.workspace_id = $2
		  AND cv.status = 'CONTRADICTED'
		LIMIT $3
	`, tenantID, workspaceID, limit)
	if err != nil {
		return false, Evidence{}, fmt.Errorf("contradiction_exists query: %w", err)
	}
	defer rows.Close()

	var ev Evidence
	for rows.Next() {
		var id, sourceID uuid.UUID
		var rawTarget, reason string
		if err := rows.Scan(&id, &sourceID, &rawTarget, &reason); err != nil {
			continue
		}
		ev.ContradictionIDs = append(ev.ContradictionIDs, id)
		ev.EntityIDs = append(ev.EntityIDs, sourceID)
		ev.ReasoningNotes = append(ev.ReasoningNotes,
			fmt.Sprintf("contradiction: source=%s target=%s reason=%s",
				sourceID, rawTarget, reason))
	}

	if len(ev.ContradictionIDs) == 0 {
		return false, Evidence{}, nil
	}
	return true, ev, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Predicate: verification_missing
//
// Fires when an entity exists but has no SUPPORTED runtime verification.
// Used by policies like "exported payment handlers must have runtime evidence".
//
// Params:
//   kind     (string, optional) — entity kind filter
//   exported (bool,   optional) — require is_exported = true
//   limit    (int,    optional) — default 10
// ─────────────────────────────────────────────────────────────────────────────

func (e *Evaluator) evalVerificationMissing(
	ctx context.Context,
	p *Policy,
	pred *Predicate,
	tenantID, workspaceID uuid.UUID,
) (bool, Evidence, error) {

	kind := paramString(pred.Params, "kind")
	exported := paramBool(pred.Params, "exported")
	limit := paramInt(pred.Params, "limit", 10)

	q := `
		SELECT e.id, e.name, e.kind, COALESCE(e.package, '')
		FROM entities e
		WHERE e.tenant_id = $1
		  AND e.workspace_id = $2
		  AND e.kind != 'external'
		  AND NOT EXISTS (
			SELECT 1 FROM claim_verifications cv
			WHERE cv.source_entity_id = e.id
			  AND cv.status = 'SUPPORTED'
		  )
	`
	args := []any{tenantID, workspaceID}
	argPos := 3

	if kind != "" {
		q += fmt.Sprintf(" AND e.kind = $%d", argPos)
		args = append(args, kind)
		argPos++
	}
	if exported {
		q += " AND e.is_exported = true"
	}
	q += " LIMIT " + fmt.Sprint(limit)

	rows, err := e.pool.Query(ctx, q, args...)
	if err != nil {
		return false, Evidence{}, fmt.Errorf("verification_missing query: %w", err)
	}
	defer rows.Close()

	var ev Evidence
	for rows.Next() {
		var id uuid.UUID
		var name, kind, pkg string
		if err := rows.Scan(&id, &name, &kind, &pkg); err != nil {
			continue
		}
		ev.EntityIDs = append(ev.EntityIDs, id)
		ev.ReasoningNotes = append(ev.ReasoningNotes,
			fmt.Sprintf("missing verification: %s.%s (%s)", pkg, name, kind))
	}

	if len(ev.EntityIDs) == 0 {
		return false, Evidence{}, nil
	}
	return true, ev, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Predicate: language_matches
//
// Fires when at least one entity in scope uses the given language.
// Always matches if the policy's Language field is "any".
//
// Params:
//   language (string, required) — e.g. "go", "python", "typescript"
// ─────────────────────────────────────────────────────────────────────────────

func (e *Evaluator) evalLanguageMatches(
	ctx context.Context,
	p *Policy,
	pred *Predicate,
	tenantID, workspaceID uuid.UUID,
) (bool, Evidence, error) {

	lang := paramString(pred.Params, "language")
	if lang == "" {
		lang = p.Language
	}
	if lang == "" || lang == "any" {
		return true, Evidence{}, nil
	}

	var count int
	err := e.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2
		  AND kind != 'external'
		  AND LOWER(language) = LOWER($3)
	`, tenantID, workspaceID, lang).Scan(&count)
	if err != nil {
		return false, Evidence{}, fmt.Errorf("language_matches query: %w", err)
	}
	if count == 0 {
		return false, Evidence{}, nil
	}
	ev := Evidence{
		ReasoningNotes: []string{
			fmt.Sprintf("workspace contains %d %s entities", count, lang),
		},
	}
	return true, ev, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Param helpers
// ─────────────────────────────────────────────────────────────────────────────

func paramString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func paramBool(m map[string]any, key string) bool {
	if m == nil {
		return false
	}
	v, ok := m[key]
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

func paramInt(m map[string]any, key string, def int) int {
	if m == nil {
		return def
	}
	v, ok := m[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return def
}
