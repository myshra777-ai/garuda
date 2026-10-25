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

// Evaluate runs one policy against a workspace without diff-scoping.
func (e *Evaluator) Evaluate(
	ctx context.Context,
	p *Policy,
	tenantID, workspaceID uuid.UUID,
	subjectKind, subjectID, actor string,
) (*Evaluation, error) {
	return e.EvaluateWithScope(ctx, p, tenantID, workspaceID, subjectKind, subjectID, actor, nil)
}

// EvaluateWithScope runs one policy against a workspace and optionally filters
// violations to net-new matches intersecting diffScope while tracking baseline
// violations separately.
func (e *Evaluator) EvaluateWithScope(
	ctx context.Context,
	p *Policy,
	tenantID, workspaceID uuid.UUID,
	subjectKind, subjectID, actor string,
	diffScope *DiffScopeFilter,
) (*Evaluation, error) {

	ev := &Evaluation{
		ID:               uuid.New(),
		TenantID:         tenantID,
		WorkspaceID:      workspaceID,
		PolicyExternalID: p.ID,
		PolicyTitle:      p.Title,
		PolicyVersion:    p.Version,
		SubjectKind:      subjectKind,
		SubjectID:        subjectID,
		Actor:            actor,
	}

	// Default: no predicate fired → ALLOW with reason
	ev.Decision = DecisionAllow
	ev.Reason = "no predicate matched"

	allMatched := true
	suppressedByDiffScope := false

	for i, pred := range p.When {
		matched, evidence, err := e.evaluatePredicate(ctx, p, &pred, tenantID, workspaceID)
		if err != nil {
			return nil, fmt.Errorf("predicate %d (%s): %w", i, pred.Type, err)
		}
		if !matched {
			allMatched = false
			break
		}

		// Apply diff-scoped net-new vs baseline partitioning when active
		if diffScope != nil && diffScope.OnlyNetNew && pred.Type != "language_matches" {
			filteredEv, netNewCount, baselineCount := filterEvidenceByDiffScope(pred.Type, evidence, diffScope)
			ev.NetNewViolationsCount += netNewCount
			ev.BaselineViolationsCount += baselineCount
			ev.Evidence.BaselineNotes = append(ev.Evidence.BaselineNotes, filteredEv.BaselineNotes...)

			if netNewCount == 0 {
				allMatched = false
				suppressedByDiffScope = baselineCount > 0
				break
			}
			evidence = filteredEv
		} else if pred.Type != "language_matches" {
			ev.NetNewViolationsCount += max(len(evidence.ReasoningNotes), 1)
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
	} else if suppressedByDiffScope {
		ev.Decision = DecisionAllow
		ev.Reason = fmt.Sprintf("0 net-new violations in diff (%d pre-existing baseline findings suppressed)", ev.BaselineViolationsCount)
	}

	return ev, nil
}

// filterEvidenceByDiffScope partitions matched evidence into net-new (touching
// diff entities/packages) vs pre-existing baseline findings.
func filterEvidenceByDiffScope(predType string, in Evidence, scope *DiffScopeFilter) (Evidence, int, int) {
	if scope == nil || !scope.OnlyNetNew {
		n := len(in.ReasoningNotes)
		if n == 0 && (len(in.EntityIDs) > 0 || len(in.ClaimIDs) > 0 || len(in.ContradictionIDs) > 0) {
			n = 1
		}
		return in, n, 0
	}

	// Check each reasoning note / entity pair
	var out Evidence
	netNew := 0
	baseline := 0

	// Build a quick lookup of whether any entity in `in.EntityIDs` matches the diff
	// For predicates that emit 2 entities per note (like claim_exists, transitive_claim_exists, layer_violation)
	// or 1 entity per note (entity_exists, verification_missing, cycle_exists, blast_radius_exceeds):
	notesCount := len(in.ReasoningNotes)
	if notesCount == 0 {
		for _, eid := range in.EntityIDs {
			if scope.MatchesEntity(eid, "") {
				netNew++
				out.EntityIDs = append(out.EntityIDs, eid)
			} else {
				baseline++
			}
		}
		return out, netNew, baseline
	}

	entitiesPerNote := 1
	if len(in.EntityIDs) >= notesCount*2 && (predType == "layer_violation" || predType == "claim_exists" || predType == "transitive_claim_exists") {
		entitiesPerNote = 2
	}

	for idx, note := range in.ReasoningNotes {
		noteMatches := false
		// Check associated entity IDs
		startEnt := idx * entitiesPerNote
		endEnt := startEnt + entitiesPerNote
		if endEnt > len(in.EntityIDs) {
			endEnt = len(in.EntityIDs)
		}
		for j := startEnt; j < endEnt; j++ {
			if scope.MatchesEntity(in.EntityIDs[j], "") {
				noteMatches = true
				break
			}
		}
		// Also check if any touched package name appears in the reasoning note
		if !noteMatches && len(scope.TouchedPackages) > 0 {
			for pkg := range scope.TouchedPackages {
				if pkg != "" && strings.Contains(note, pkg) {
					noteMatches = true
					break
				}
			}
		}

		if noteMatches {
			netNew++
			out.ReasoningNotes = append(out.ReasoningNotes, note)
			for j := startEnt; j < endEnt; j++ {
				out.EntityIDs = append(out.EntityIDs, in.EntityIDs[j])
			}
			if idx < len(in.ClaimIDs) {
				out.ClaimIDs = append(out.ClaimIDs, in.ClaimIDs[idx])
			}
			if idx < len(in.ContradictionIDs) {
				out.ContradictionIDs = append(out.ContradictionIDs, in.ContradictionIDs[idx])
			}
		} else {
			baseline++
			out.BaselineNotes = append(out.BaselineNotes, note)
		}
	}

	return out, netNew, baseline
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
	case "transitive_claim_exists":
		return e.evalTransitiveClaimExists(ctx, p, pred, tenantID, workspaceID)
	case "layer_violation":
		return e.evalLayerViolation(ctx, p, pred, tenantID, workspaceID)
	case "cycle_exists":
		return e.evalCycleExists(ctx, p, pred, tenantID, workspaceID)
	case "blast_radius_exceeds":
		return e.evalBlastRadiusExceeds(ctx, p, pred, tenantID, workspaceID)
	default:
		return false, Evidence{}, fmt.Errorf("unsupported predicate type: %s", pred.Type)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Predicate: entity_exists
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
	limit := paramInt(pred.Params, "limit", 25)

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
	if p.Scope.Pattern != "" {
		if _, err := regexp.Compile(p.Scope.Pattern); err != nil {
			return false, Evidence{}, fmt.Errorf("invalid scope.pattern %q: %w", p.Scope.Pattern, err)
		}
		q += fmt.Sprintf(" AND (COALESCE(package, '') ~ $%d OR name ~ $%d)", argPos, argPos)
		args = append(args, p.Scope.Pattern)
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
	limit := paramInt(pred.Params, "limit", 25)

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
	if p.Scope.Package != "" {
		q += fmt.Sprintf(" AND e1.package LIKE '%%' || $%d || '%%'", argPos)
		args = append(args, p.Scope.Package)
		argPos++
	}
	if p.Scope.Pattern != "" {
		if _, err := regexp.Compile(p.Scope.Pattern); err != nil {
			return false, Evidence{}, fmt.Errorf("invalid scope.pattern %q: %w", p.Scope.Pattern, err)
		}
		q += fmt.Sprintf(" AND (COALESCE(e1.package, '') ~ $%d OR e1.name ~ $%d)", argPos, argPos)
		args = append(args, p.Scope.Pattern)
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
		} else {
			ev.EntityIDs = append(ev.EntityIDs, fromID)
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
// ─────────────────────────────────────────────────────────────────────────────

func (e *Evaluator) evalContradictionExists(
	ctx context.Context,
	p *Policy,
	pred *Predicate,
	tenantID, workspaceID uuid.UUID,
) (bool, Evidence, error) {

	limit := paramInt(pred.Params, "limit", 25)

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
// ─────────────────────────────────────────────────────────────────────────────

func (e *Evaluator) evalVerificationMissing(
	ctx context.Context,
	p *Policy,
	pred *Predicate,
	tenantID, workspaceID uuid.UUID,
) (bool, Evidence, error) {

	kind := paramString(pred.Params, "kind")
	exported := paramBool(pred.Params, "exported")
	limit := paramInt(pred.Params, "limit", 25)

	q := `
		SELECT e.id, e.name, e.kind, COALESCE(e.package, '')
		FROM entities e
		WHERE e.tenant_id = $1
		  AND e.workspace_id = $2
		  AND e.kind != 'external'
		  AND NOT EXISTS (
			SELECT 1 FROM claim_verifications cv
			WHERE cv.source_entity_id = e.id
			  AND cv.workspace_id = e.workspace_id
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

// ─────────────────────────────────────────────────────────────────────────────
// Predicate: transitive_claim_exists
// ─────────────────────────────────────────────────────────────────────────────

func (e *Evaluator) evalTransitiveClaimExists(
	ctx context.Context,
	p *Policy,
	pred *Predicate,
	tenantID, workspaceID uuid.UUID,
) (bool, Evidence, error) {
	claimType := paramString(pred.Params, "claim_type")
	if claimType == "" {
		claimType = "CALLS"
	}
	fromPat := paramString(pred.Params, "from_name_pattern")
	toPat := paramString(pred.Params, "to_name_pattern")
	maxDepth := paramInt(pred.Params, "max_depth", 5)
	if maxDepth <= 0 || maxDepth > 10 {
		maxDepth = 5
	}
	limit := paramInt(pred.Params, "limit", 25)

	if fromPat != "" {
		if _, err := regexp.Compile(fromPat); err != nil {
			return false, Evidence{}, fmt.Errorf("invalid from_name_pattern: %w", err)
		}
	}
	if toPat != "" {
		if _, err := regexp.Compile(toPat); err != nil {
			return false, Evidence{}, fmt.Errorf("invalid to_name_pattern: %w", err)
		}
	}

	q := `
		WITH RECURSIVE call_path AS (
			SELECT
				c.id AS first_claim_id,
				c.from_entity_id AS root_id,
				e1.name AS root_name,
				c.to_entity_id AS current_id,
				COALESCE(e2.name, '') AS current_name,
				1 AS depth,
				ARRAY[c.from_entity_id, c.to_entity_id] AS visited,
				e1.name || ' -> ' || COALESCE(e2.name, '') AS trace_path
			FROM claims c
			JOIN entities e1 ON e1.id = c.from_entity_id
			LEFT JOIN entities e2 ON e2.id = c.to_entity_id
			WHERE c.tenant_id = $1
			  AND c.workspace_id = $2
			  AND UPPER(c.claim_type) = UPPER($3)
			  AND ($4 = '' OR e1.name ~ $4)
			  AND ($5 = '' OR COALESCE(e1.package, '') LIKE '%' || $5 || '%')

			UNION ALL

			SELECT
				cp.first_claim_id,
				cp.root_id,
				cp.root_name,
				c.to_entity_id,
				COALESCE(e_next.name, ''),
				cp.depth + 1,
				cp.visited || c.to_entity_id,
				cp.trace_path || ' -> ' || COALESCE(e_next.name, '')
			FROM call_path cp
			JOIN claims c ON c.from_entity_id = cp.current_id
			LEFT JOIN entities e_next ON e_next.id = c.to_entity_id
			WHERE c.tenant_id = $1
			  AND c.workspace_id = $2
			  AND UPPER(c.claim_type) = UPPER($3)
			  AND cp.depth < $6
			  AND NOT (c.to_entity_id = ANY(cp.visited))
		)
		SELECT first_claim_id, root_id, current_id, depth, trace_path
		FROM call_path
		WHERE ($7 = '' OR current_name ~ $7)
		ORDER BY depth ASC
		LIMIT $8
	`

	rows, err := e.pool.Query(ctx, q,
		tenantID, workspaceID, claimType, fromPat, p.Scope.Package, maxDepth, toPat, limit,
	)
	if err != nil {
		return false, Evidence{}, fmt.Errorf("transitive_claim_exists query: %w", err)
	}
	defer rows.Close()

	var ev Evidence
	for rows.Next() {
		var claimID, rootID uuid.UUID
		var targetID *uuid.UUID
		var depth int
		var tracePath string
		if err := rows.Scan(&claimID, &rootID, &targetID, &depth, &tracePath); err != nil {
			continue
		}
		ev.ClaimIDs = append(ev.ClaimIDs, claimID)
		ev.EntityIDs = append(ev.EntityIDs, rootID)
		if targetID != nil {
			ev.EntityIDs = append(ev.EntityIDs, *targetID)
		} else {
			ev.EntityIDs = append(ev.EntityIDs, rootID)
		}
		ev.ReasoningNotes = append(ev.ReasoningNotes,
			fmt.Sprintf("transitive %s (%d hops): %s", claimType, depth, tracePath))
	}

	if len(ev.ClaimIDs) == 0 {
		return false, Evidence{}, nil
	}
	return true, ev, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Predicate: layer_violation
// ─────────────────────────────────────────────────────────────────────────────

func (e *Evaluator) evalLayerViolation(
	ctx context.Context,
	_ *Policy,
	pred *Predicate,
	tenantID, workspaceID uuid.UUID,
) (bool, Evidence, error) {
	fromLayer := paramString(pred.Params, "from_layer")
	toLayer := paramString(pred.Params, "to_layer")
	limit := paramInt(pred.Params, "limit", 25)

	if fromLayer == "" || toLayer == "" {
		return false, Evidence{}, fmt.Errorf("layer_violation requires both from_layer and to_layer")
	}
	if _, err := regexp.Compile(fromLayer); err != nil {
		return false, Evidence{}, fmt.Errorf("invalid from_layer: %w", err)
	}
	if _, err := regexp.Compile(toLayer); err != nil {
		return false, Evidence{}, fmt.Errorf("invalid to_layer: %w", err)
	}

	q := `
		SELECT c.id, e1.id, e2.id, c.claim_type,
		       COALESCE(e1.package, ''), e1.name,
		       COALESCE(e2.package, ''), e2.name
		FROM claims c
		JOIN entities e1 ON e1.id = c.from_entity_id
		JOIN entities e2 ON e2.id = c.to_entity_id
		WHERE c.tenant_id = $1
		  AND c.workspace_id = $2
		  AND COALESCE(e1.package, '') ~ $3
		  AND COALESCE(e2.package, '') ~ $4
		  AND COALESCE(e1.package, '') != COALESCE(e2.package, '')
		LIMIT $5
	`
	rows, err := e.pool.Query(ctx, q, tenantID, workspaceID, fromLayer, toLayer, limit)
	if err != nil {
		return false, Evidence{}, fmt.Errorf("layer_violation query: %w", err)
	}
	defer rows.Close()

	var ev Evidence
	for rows.Next() {
		var claimID, fromID, toID uuid.UUID
		var cType, fromPkg, fromName, toPkg, toName string
		if err := rows.Scan(&claimID, &fromID, &toID, &cType, &fromPkg, &fromName, &toPkg, &toName); err != nil {
			continue
		}
		ev.ClaimIDs = append(ev.ClaimIDs, claimID)
		ev.EntityIDs = append(ev.EntityIDs, fromID, toID)
		ev.ReasoningNotes = append(ev.ReasoningNotes,
			fmt.Sprintf("layer violation (%s): %s.%s -> %s.%s", cType, fromPkg, fromName, toPkg, toName))
	}

	if len(ev.ClaimIDs) == 0 {
		return false, Evidence{}, nil
	}
	return true, ev, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Predicate: cycle_exists
// ─────────────────────────────────────────────────────────────────────────────

func (e *Evaluator) evalCycleExists(
	ctx context.Context,
	_ *Policy,
	pred *Predicate,
	tenantID, workspaceID uuid.UUID,
) (bool, Evidence, error) {
	maxDepth := paramInt(pred.Params, "max_depth", 5)
	if maxDepth <= 1 || maxDepth > 8 {
		maxDepth = 5
	}
	limit := paramInt(pred.Params, "limit", 10)

	q := `
		WITH pkg_edges AS (
			SELECT DISTINCT
				COALESCE(e1.package, '') AS from_pkg,
				COALESCE(e2.package, '') AS to_pkg,
				MIN(e1.id::text)::uuid AS sample_entity_id
			FROM claims c
			JOIN entities e1 ON e1.id = c.from_entity_id
			JOIN entities e2 ON e2.id = c.to_entity_id
			WHERE c.tenant_id = $1
			  AND c.workspace_id = $2
			  AND COALESCE(e1.package, '') != ''
			  AND COALESCE(e2.package, '') != ''
			  AND COALESCE(e1.package, '') != COALESCE(e2.package, '')
			GROUP BY COALESCE(e1.package, ''), COALESCE(e2.package, '')
		),
		RECURSIVE cycle_search AS (
			SELECT
				from_pkg AS start_pkg,
				to_pkg AS current_pkg,
				1 AS depth,
				ARRAY[from_pkg, to_pkg] AS path,
				from_pkg || ' -> ' || to_pkg AS cycle_str,
				sample_entity_id
			FROM pkg_edges

			UNION ALL

			SELECT
				cs.start_pkg,
				pe.to_pkg,
				cs.depth + 1,
				cs.path || pe.to_pkg,
				cs.cycle_str || ' -> ' || pe.to_pkg,
				cs.sample_entity_id
			FROM cycle_search cs
			JOIN pkg_edges pe ON pe.from_pkg = cs.current_pkg
			WHERE cs.depth < $3
			  AND (pe.to_pkg = cs.start_pkg OR NOT (pe.to_pkg = ANY(cs.path)))
		)
		SELECT sample_entity_id, cycle_str, depth
		FROM cycle_search
		WHERE current_pkg = start_pkg AND depth >= 2
		ORDER BY depth ASC
		LIMIT $4
	`
	rows, err := e.pool.Query(ctx, q, tenantID, workspaceID, maxDepth, limit)
	if err != nil {
		return false, Evidence{}, fmt.Errorf("cycle_exists query: %w", err)
	}
	defer rows.Close()

	var ev Evidence
	for rows.Next() {
		var entID uuid.UUID
		var cycleStr string
		var depth int
		if err := rows.Scan(&entID, &cycleStr, &depth); err != nil {
			continue
		}
		ev.EntityIDs = append(ev.EntityIDs, entID)
		ev.ReasoningNotes = append(ev.ReasoningNotes,
			fmt.Sprintf("dependency cycle (%d packages): %s", depth, cycleStr))
	}

	if len(ev.EntityIDs) == 0 {
		return false, Evidence{}, nil
	}
	return true, ev, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Predicate: blast_radius_exceeds
// ─────────────────────────────────────────────────────────────────────────────

func (e *Evaluator) evalBlastRadiusExceeds(
	ctx context.Context,
	p *Policy,
	pred *Predicate,
	tenantID, workspaceID uuid.UUID,
) (bool, Evidence, error) {
	maxCallers := paramInt(pred.Params, "max_callers", 10)
	namePat := paramString(pred.Params, "name_pattern")
	limit := paramInt(pred.Params, "limit", 25)

	if namePat != "" {
		if _, err := regexp.Compile(namePat); err != nil {
			return false, Evidence{}, fmt.Errorf("invalid name_pattern: %w", err)
		}
	}

	q := `
		SELECT e.id, e.name, COALESCE(e.package, ''), COUNT(DISTINCT c.from_entity_id) AS caller_count
		FROM entities e
		JOIN claims c ON c.to_entity_id = e.id
		WHERE e.tenant_id = $1
		  AND e.workspace_id = $2
		  AND ($3 = '' OR e.name ~ $3)
		  AND ($4 = '' OR COALESCE(e.package, '') LIKE '%' || $4 || '%')
		GROUP BY e.id, e.name, e.package
		HAVING COUNT(DISTINCT c.from_entity_id) > $5
		ORDER BY caller_count DESC
		LIMIT $6
	`
	rows, err := e.pool.Query(ctx, q, tenantID, workspaceID, namePat, p.Scope.Package, maxCallers, limit)
	if err != nil {
		return false, Evidence{}, fmt.Errorf("blast_radius_exceeds query: %w", err)
	}
	defer rows.Close()

	var ev Evidence
	for rows.Next() {
		var id uuid.UUID
		var name, pkg string
		var count int
		if err := rows.Scan(&id, &name, &pkg, &count); err != nil {
			continue
		}
		ev.EntityIDs = append(ev.EntityIDs, id)
		ev.ReasoningNotes = append(ev.ReasoningNotes,
			fmt.Sprintf("high blast radius: %s.%s has %d inbound dependents (threshold=%d)", pkg, name, count, maxCallers))
	}

	if len(ev.EntityIDs) == 0 {
		return false, Evidence{}, nil
	}
	return true, ev, nil
}
