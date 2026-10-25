// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

// Package diff parses unified git diffs (working tree, staged index, revision
// ranges, or explicit file lists) into file paths and modified line spans, and
// maps those line spans to exact AST entities and packages in the Garuda graph.
package diff

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/analyzer"
)

// LineRange represents a 1-indexed inclusive [Start, End] span of modified lines
// in the new version of a file.
type LineRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// Overlaps returns true if [r.Start, r.End] overlaps [lineStart, lineEnd].
func (r LineRange) Overlaps(lineStart, lineEnd int) bool {
	if lineEnd < lineStart {
		lineEnd = lineStart
	}
	return r.Start <= lineEnd && r.End >= lineStart
}

// FileDelta represents a modified file and its changed line spans.
type FileDelta struct {
	FilePath     string      `json:"file_path"`
	Status       string      `json:"status"` // "M", "A", "D", "U"
	AddedLines   int         `json:"added_lines"`
	DeletedLines int         `json:"deleted_lines"`
	Hunks        []LineRange `json:"hunks"`
}

// TouchedEntity represents an AST entity whose source span overlaps a diff hunk
// or resides in a modified file.
type TouchedEntity struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Kind      string    `json:"kind"`
	Package   string    `json:"package"`
	FilePath  string    `json:"file_path"`
	LineStart int       `json:"line_start"`
	LineEnd   int       `json:"line_end"`
	ExactHunk bool      `json:"exact_hunk"` // true if overlapped a modified line hunk directly
}

// TargetResolution holds the complete set of files, entities, and packages
// affected by a git diff or file list.
type TargetResolution struct {
	Mode            string          `json:"mode"` // "staged", "diff", "working_tree", "files"
	Revision        string          `json:"revision,omitempty"`
	Deltas          []FileDelta     `json:"deltas"`
	TouchedEntities []TouchedEntity `json:"touched_entities"`
	TouchedPackages []string        `json:"touched_packages"`
	TouchedFiles    []string        `json:"touched_files"`
}

var hunkHeaderRe = regexp.MustCompile(`^@@\s+-[0-9]+(?:,[0-9]+)?\s+\+([0-9]+)(?:,([0-9]+))?\s+@@`)

// ParseUnifiedDiff parses the output of `git diff --unified=0` into FileDeltas.
func ParseUnifiedDiff(raw string) []FileDelta {
	var deltas []FileDelta
	var current *FileDelta

	scanner := bufio.NewScanner(strings.NewReader(raw))
	scanner.Buffer(make([]byte, 1024*1024), 4*1024*1024)

	flush := func() {
		if current != nil && current.FilePath != "" {
			deltas = append(deltas, *current)
		}
		current = nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "diff --git "):
			flush()
			current = &FileDelta{Status: "M"}
		case strings.HasPrefix(line, "new file mode"):
			if current != nil {
				current.Status = "A"
			}
		case strings.HasPrefix(line, "deleted file mode"):
			if current != nil {
				current.Status = "D"
			}
		case strings.HasPrefix(line, "+++ "):
			if current == nil {
				current = &FileDelta{Status: "M"}
			}
			target := strings.TrimSpace(strings.TrimPrefix(line, "+++ "))
			if target == "/dev/null" {
				current.Status = "D"
				continue
			}
			target = strings.TrimPrefix(target, "b/")
			current.FilePath = filepath.ToSlash(target)
		case strings.HasPrefix(line, "--- "):
			if current != nil && current.FilePath == "" {
				src := strings.TrimSpace(strings.TrimPrefix(line, "--- "))
				if src != "/dev/null" {
					current.FilePath = filepath.ToSlash(strings.TrimPrefix(src, "a/"))
				}
			}
		case strings.HasPrefix(line, "@@ "):
			if current == nil {
				continue
			}
			if m := hunkHeaderRe.FindStringSubmatch(line); m != nil {
				start, _ := strconv.Atoi(m[1])
				count := 1
				if m[2] != "" {
					count, _ = strconv.Atoi(m[2])
				}
				if count <= 0 {
					// Pure deletion at line `start` — anchor a 1-line span at start
					if start <= 0 {
						start = 1
					}
					current.Hunks = append(current.Hunks, LineRange{Start: start, End: start})
				} else {
					current.Hunks = append(current.Hunks, LineRange{
						Start: start,
						End:   start + count - 1,
					})
				}
			}
		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			if current != nil {
				current.AddedLines++
			}
		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			if current != nil {
				current.DeletedLines++
			}
		}
	}
	flush()
	return deltas
}

// CollectGitDeltas runs git in repoDir to collect FileDeltas according to options:
//   - staged=true: `git diff --cached --unified=0`
//   - rev!="":     `git diff --unified=0 <rev>`
//   - files!=nil:  explicit file list (whole-file line ranges)
//   - default:     working tree (`git diff HEAD --unified=0` + untracked files)
func CollectGitDeltas(repoDir string, staged bool, rev string, explicitFiles []string) (string, []FileDelta, error) {
	if len(explicitFiles) > 0 {
		deltas := make([]FileDelta, 0, len(explicitFiles))
		for _, f := range explicitFiles {
			f = strings.TrimSpace(f)
			if f == "" {
				continue
			}
			if rel, err := filepath.Rel(repoDir, f); err == nil && !strings.HasPrefix(rel, "..") {
				f = rel
			}
			deltas = append(deltas, FileDelta{
				FilePath: filepath.ToSlash(f),
				Status:   "M",
				Hunks:    []LineRange{{Start: 1, End: 1_000_000}},
			})
		}
		return "files", deltas, nil
	}

	if staged {
		out, err := exec.Command("git", "-C", repoDir, "diff", "--cached", "--unified=0", "--no-color").Output()
		if err != nil {
			return "staged", nil, fmt.Errorf("git diff --cached: %w", err)
		}
		return "staged", ParseUnifiedDiff(string(out)), nil
	}

	if rev != "" {
		out, err := exec.Command("git", "-C", repoDir, "diff", "--unified=0", "--no-color", rev).Output()
		if err != nil {
			return "diff", nil, fmt.Errorf("git diff %s: %w", rev, err)
		}
		return "diff", ParseUnifiedDiff(string(out)), nil
	}

	// Default: working tree relative to HEAD (plus untracked source files)
	out, err := exec.Command("git", "-C", repoDir, "diff", "HEAD", "--unified=0", "--no-color").Output()
	if err != nil {
		// Fallback when repo has no commits yet
		out, _ = exec.Command("git", "-C", repoDir, "diff", "--unified=0", "--no-color").Output()
	}
	deltas := ParseUnifiedDiff(string(out))
	seen := make(map[string]bool, len(deltas))
	for _, d := range deltas {
		seen[d.FilePath] = true
	}

	// Also include untracked source files from `git ls-files --others --exclude-standard`
	if untrackedOut, uErr := exec.Command("git", "-C", repoDir, "ls-files", "--others", "--exclude-standard").Output(); uErr == nil {
		for _, line := range strings.Split(string(untrackedOut), "\n") {
			p := filepath.ToSlash(strings.TrimSpace(line))
			if p == "" || seen[p] || !isSupportedSourceFile(p) {
				continue
			}
			seen[p] = true
			deltas = append(deltas, FileDelta{
				FilePath: p,
				Status:   "A",
				Hunks:    []LineRange{{Start: 1, End: 1_000_000}},
			})
		}
	}

	return "working_tree", deltas, nil
}

// HasSourceFileDeltas returns true if at least one FileDelta touches a supported
// source code file (.go, .py, .ts, .tsx, .js, .jsx).
func HasSourceFileDeltas(deltas []FileDelta) bool {
	for _, d := range deltas {
		if isSupportedSourceFile(d.FilePath) {
			return true
		}
	}
	return false
}

func isSupportedSourceFile(p string) bool {
	ext := strings.ToLower(filepath.Ext(p))
	switch ext {
	case ".go", ".py", ".ts", ".tsx", ".js", ".jsx":
		return true
	default:
		return false
	}
}

// ResolveTouchedEntities maps FileDeltas to persisted entities and packages in PostgreSQL.
func ResolveTouchedEntities(
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID uuid.UUID,
	repoRoot string,
	mode, rev string,
	deltas []FileDelta,
) (*TargetResolution, error) {
	res := &TargetResolution{
		Mode:     mode,
		Revision: rev,
		Deltas:   deltas,
	}
	if len(deltas) == 0 {
		return res, nil
	}

	fileSet := make(map[string]FileDelta, len(deltas))
	for _, d := range deltas {
		clean := filepath.ToSlash(strings.TrimPrefix(d.FilePath, "./"))
		fileSet[clean] = d
		res.TouchedFiles = append(res.TouchedFiles, clean)
	}
	sort.Strings(res.TouchedFiles)

	rows, err := pool.Query(ctx, `
		SELECT id, name, kind, COALESCE(package, ''), COALESCE(file_path, ''),
		       COALESCE(line_start, COALESCE(line, 0)),
		       COALESCE(line_end, COALESCE(line, 0))
		FROM entities
		WHERE workspace_id = $1
		  AND kind != 'external'
		  AND file_path IS NOT NULL
		  AND file_path != ''
	`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("query workspace entities for diff mapping: %w", err)
	}
	defer rows.Close()

	pkgSet := make(map[string]bool)
	var exactMatches []TouchedEntity
	var fileMatches []TouchedEntity

	absRoot, _ := filepath.Abs(repoRoot)

	for rows.Next() {
		var id uuid.UUID
		var name, kind, pkg, fp string
		var lStart, lEnd int
		if err := rows.Scan(&id, &name, &kind, &pkg, &fp, &lStart, &lEnd); err != nil {
			continue
		}

		delta, matchedFile := matchFileDelta(fp, absRoot, fileSet)
		if !matchedFile {
			continue
		}
		if pkg != "" {
			pkgSet[pkg] = true
		}
		if lEnd < lStart {
			lEnd = lStart
		}

		overlapsHunk := false
		if len(delta.Hunks) == 0 {
			overlapsHunk = true
		} else {
			for _, h := range delta.Hunks {
				if h.Overlaps(lStart, lEnd) {
					overlapsHunk = true
					break
				}
			}
		}

		te := TouchedEntity{
			ID:        id,
			Name:      name,
			Kind:      kind,
			Package:   pkg,
			FilePath:  fp,
			LineStart: lStart,
			LineEnd:   lEnd,
			ExactHunk: overlapsHunk,
		}
		if overlapsHunk && kind != "package" && kind != "file" && kind != "module" {
			exactMatches = append(exactMatches, te)
		} else {
			fileMatches = append(fileMatches, te)
		}
	}

	// Prefer exact hunk-overlapping symbols; if a diff modified top-level imports
	// or comments outside any function span, fall back to all symbols in the touched file.
	if len(exactMatches) > 0 {
		res.TouchedEntities = exactMatches
	} else {
		res.TouchedEntities = fileMatches
	}

	sort.Slice(res.TouchedEntities, func(i, j int) bool {
		if res.TouchedEntities[i].FilePath != res.TouchedEntities[j].FilePath {
			return res.TouchedEntities[i].FilePath < res.TouchedEntities[j].FilePath
		}
		return res.TouchedEntities[i].LineStart < res.TouchedEntities[j].LineStart
	})

	for pkg := range pkgSet {
		res.TouchedPackages = append(res.TouchedPackages, pkg)
	}
	sort.Strings(res.TouchedPackages)

	return res, nil
}

// MatchAnalyzerEntities filters an in-memory analyzer.Result against FileDeltas.
func MatchAnalyzerEntities(entities []analyzer.Entity, repoRoot string, deltas []FileDelta) []analyzer.Entity {
	if len(deltas) == 0 {
		return nil
	}
	fileSet := make(map[string]FileDelta, len(deltas))
	for _, d := range deltas {
		clean := filepath.ToSlash(strings.TrimPrefix(d.FilePath, "./"))
		fileSet[clean] = d
	}
	absRoot, _ := filepath.Abs(repoRoot)

	var matched []analyzer.Entity
	for _, e := range entities {
		delta, ok := matchFileDelta(e.File, absRoot, fileSet)
		if !ok {
			continue
		}
		lStart := e.LineStart
		if lStart == 0 {
			lStart = e.Line
		}
		lEnd := e.LineEnd
		if lEnd < lStart {
			lEnd = lStart
		}
		if len(delta.Hunks) == 0 {
			matched = append(matched, e)
			continue
		}
		for _, h := range delta.Hunks {
			if h.Overlaps(lStart, lEnd) {
				matched = append(matched, e)
				break
			}
		}
	}
	return matched
}

func matchFileDelta(entityFilePath, absRoot string, fileSet map[string]FileDelta) (FileDelta, bool) {
	norm := filepath.ToSlash(entityFilePath)
	if absRoot != "" {
		if rel, err := filepath.Rel(absRoot, entityFilePath); err == nil && !strings.HasPrefix(rel, "..") {
			norm = filepath.ToSlash(rel)
		}
	}
	norm = strings.TrimPrefix(norm, "./")
	if d, ok := fileSet[norm]; ok {
		return d, true
	}
	// Suffix match when entityFilePath is absolute and repoRoot differs
	for relPath, d := range fileSet {
		if strings.HasSuffix(norm, "/"+relPath) || norm == relPath {
			return d, true
		}
	}
	return FileDelta{}, false
}
