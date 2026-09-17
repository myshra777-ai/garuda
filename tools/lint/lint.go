// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

// Package lint hosts Garuda's static checks. Each check is a function
// over the source tree; TestLintRules walks the repo and reports every
// violation. The package is run by `go test ./tools/lint/`, so it
// rides on the existing test runner — no separate binary, no separate
// CI step.
//
// The checks exist to prevent a specific pattern the tracker names
// "two sources of truth": two places that describe the same thing and
// diverge. The current checks catch the two shapes that can be
// detected syntactically:
//
//   - sql-scoping: a query against a workspace-scoped table that
//     filters by tenant_id but not workspace_id (or repository_id, or
//     a primary-key lookup).
//
//   - variadic-single-arg: a function that accepts ...T but only ever
//     reads args[0]. A caller that passes two options has the second
//     silently ignored.
//
// Categories that are not automated and the reasons they are not:
//
//   - tool-description-vs-handler: requires reading a handler's side
//     effects and comparing to a description string. Semantic, not
//     syntactic.
//
//   - DTO-field-declared-but-not-set: a field that is legitimately
//     zero is indistinguishable from one that was forgotten.
//
//   - filesystem-vs-DB-row existence: a runtime property, not a
//     syntactic one. Belongs in a test, not a linter.
package lint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Violation is one finding. File and Line locate it; Check names the
// rule that fired; Hash is a stable key for allowlisting (line numbers
// drift, literal content does not).
type Violation struct {
	File    string
	Line    int
	Check   string
	Literal string
	Fix     string
	Hash    string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s:%d\n  check: %s\n  hash:  %s\n  query: %s\n  fix:   %s",
		v.File, v.Line, v.Check, v.Hash, v.Literal, v.Fix)
}

// Check walks root and returns every violation. Directories named
// .git, vendor, node_modules, and testdata are skipped. Files ending
// in _test.go are skipped: test code deliberately constructs the
// shapes the checks look for.
func Check(root string) ([]Violation, error) {
	var violations []Violation
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "node_modules", "testdata":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		vs, err := checkFile(path)
		if err != nil {
			return err
		}
		violations = append(violations, vs...)
		return nil
	})
	return violations, err
}

// checkFile parses a single Go file and returns its violations. Kept
// unexported so the walker is the only public entry point; tests in
// this package call it directly against fixtures.
func checkFile(path string) ([]Violation, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	var vs []Violation
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.BasicLit:
			if node.Kind != token.STRING {
				return true
			}
			lit, err := strconv.Unquote(node.Value)
			if err != nil {
				return true
			}
			line := fset.Position(node.Pos()).Line
			vs = append(vs, sqlScopingViolations(path, line, lit)...)
		case *ast.FuncDecl:
			vs = append(vs, variadicViolations(path, fset, node)...)
		}
		return true
	})
	return vs, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Check 1: SQL scoping
// ─────────────────────────────────────────────────────────────────────────────

var (
	sqlVerbRe   = regexp.MustCompile(`(?i)^(SELECT|INSERT|UPDATE|DELETE|WITH)\b`)
	tableRefRe  = regexp.MustCompile(`(?i)\b(FROM|JOIN|UPDATE|INSERT\s+INTO|DELETE\s+FROM)\s+([a-z_][a-z0-9_]*)`)
	tenantRe    = regexp.MustCompile(`(?i)\btenant_id\b`)
	workspaceRe = regexp.MustCompile(`(?i)\bworkspace_id\b`)
	repoRe      = regexp.MustCompile(`(?i)\brepository_id\b`)
	pkLookupRe  = regexp.MustCompile(`(?i)\b(\w+\.)?id\s*=\s*\$`)
)

// sqlScopingViolations applies the SQL scoping check to one string
// literal. A literal fires if:
//
//  1. It begins with a SQL verb (the anti-false-positive guard against
//     log messages that mention SQL keywords).
//  2. It references at least one workspace-scoped table.
//  3. It mentions tenant_id — the query is trying to scope by tenant.
//  4. It mentions none of workspace_id, repository_id, or a
//     primary-key lookup (id = $N).
//
// The workspace_id check is a substring match, not an equality. That
// deliberately catches `workspace_id IN (SELECT id FROM workspaces …)`
// and other indirect scoping patterns that are legitimate.
func sqlScopingViolations(file string, line int, lit string) []Violation {
	trimmed := strings.TrimSpace(lit)
	if !sqlVerbRe.MatchString(trimmed) {
		return nil
	}
	matches := tableRefRe.FindAllStringSubmatch(lit, -1)
	if len(matches) == 0 {
		return nil
	}
	if !tenantRe.MatchString(lit) {
		return nil
	}
	if workspaceRe.MatchString(lit) {
		return nil
	}
	if repoRe.MatchString(lit) {
		return nil
	}
	if pkLookupRe.MatchString(lit) {
		return nil
	}

	hash := sha256Hex(lit)
	seen := make(map[string]bool)
	var vs []Violation
	for _, m := range matches {
		table := strings.ToLower(m[2])
		if !WorkspaceScopedTables[table] {
			continue
		}
		if seen[table] {
			continue
		}
		seen[table] = true
		if IsAllowed(file, "sql-scoping", hash) {
			continue
		}
		vs = append(vs, Violation{
			File:    file,
			Line:    line,
			Check:   "sql-scoping",
			Literal: truncate(firstLine(trimmed), 100),
			Fix: fmt.Sprintf(
				"add workspace_id (or repository_id) to the WHERE clause of the query against %s",
				table),
			Hash: hash,
		})
	}
	return vs
}

// ─────────────────────────────────────────────────────────────────────────────
// Check 2: variadic signature reads only the first argument
// ─────────────────────────────────────────────────────────────────────────────

// variadicViolations applies the variadic check to one function
// declaration. A function fires if:
//
//  1. Its last parameter is `...T` where T is a named, non-builtin
//     type (presumed to be a struct — the shape of a functional-option
//     argument).
//  2. Its body reads args[0].
//  3. Its body does NOT read args[i] for i > 0, range over args, or
//     otherwise treat the slice as a slice.
//
// The check is deliberately narrow. `...string`, `...any`, and
// `...interface{}` are legitimate variadic patterns and never fire.
// A function that does anything with the slice beyond reading index 0
// is legitimate and never fires.
func variadicViolations(file string, fset *token.FileSet, fn *ast.FuncDecl) []Violation {
	if fn.Type == nil || fn.Type.Params == nil || fn.Body == nil {
		return nil
	}
	var (
		varName string
		varElem string
	)
	for _, field := range fn.Type.Params.List {
		ell, ok := field.Type.(*ast.Ellipsis)
		if !ok {
			continue
		}
		if len(field.Names) == 0 {
			continue
		}
		ident, ok := ell.Elt.(*ast.Ident)
		if !ok {
			continue
		}
		varName = field.Names[0].Name
		varElem = ident.Name
	}
	if varName == "" {
		return nil
	}
	if isBuiltinType(varElem) {
		return nil
	}
	if varElem[0] < 'A' || varElem[0] > 'Z' {
		return nil
	}

	hasIndexZero := false
	legitimateUse := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.RangeStmt:
			if id, ok := node.X.(*ast.Ident); ok && id.Name == varName {
				legitimateUse = true
			}
		case *ast.IndexExpr:
			id, ok := node.X.(*ast.Ident)
			if !ok || id.Name != varName {
				return true
			}
			if lit, ok := node.Index.(*ast.BasicLit); ok && lit.Kind == token.INT && lit.Value == "0" {
				hasIndexZero = true
			} else {
				legitimateUse = true
			}
		}
		return true
	})
	if !hasIndexZero || legitimateUse {
		return nil
	}

	hash := sha256Hex(fn.Name.Name + "/" + varName + "/" + varElem)
	if IsAllowed(file, "variadic-single-arg", hash) {
		return nil
	}
	return []Violation{{
		File:  file,
		Line:  fset.Position(fn.Pos()).Line,
		Check: "variadic-single-arg",
		Literal: fmt.Sprintf("func %s(..., %s ...%s)",
			fn.Name.Name, varName, varElem),
		Fix: fmt.Sprintf(
			"change %s ...%s to %s %s; the body reads only the first element",
			varName, varElem, varName, varElem),
		Hash: hash,
	}}
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func isBuiltinType(name string) bool {
	switch name {
	case "string", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "bool", "byte", "rune", "error", "any":
		return true
	}
	return false
}

// ─────────────────────────────────────────────────────────────────────────────
// Migration parsing — used by TestTablesMatchMigrations
// ─────────────────────────────────────────────────────────────────────────────

var (
	createTableRe = regexp.MustCompile(`(?is)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([a-z_][a-z0-9_]*)\s*\(`)
	alterAddRe    = regexp.MustCompile(`(?is)ALTER\s+TABLE\s+(?:IF\s+EXISTS\s+)?([a-z_][a-z0-9_]*)\s+ADD\s+COLUMN\s+(?:IF\s+NOT\s+EXISTS\s+)?workspace_id\b`)
	alterDropRe   = regexp.MustCompile(`(?is)ALTER\s+TABLE\s+(?:IF\s+EXISTS\s+)?([a-z_][a-z0-9_]*)\s+DROP\s+COLUMN\s+(?:IF\s+EXISTS\s+)?workspace_id\b`)
)

// parseMigrations scans every .sql file in dir and returns the set of
// tables that declare workspace_id. It processes three statement
// shapes:
//
//   - CREATE TABLE [IF NOT EXISTS] <name> ( … ) — scan the body for
//     the workspace_id column.
//   - ALTER TABLE <name> ADD COLUMN [IF NOT EXISTS] workspace_id — set
//     the table as workspace-scoped.
//   - ALTER TABLE <name> DROP COLUMN [IF EXISTS] workspace_id — clear
//     it.
//
// Files are processed in lexicographic order. Duplicate filename
// prefixes (B3 in the tracker) do not affect the result because the
// parser reads file contents, not filenames.
func parseMigrations(dir string) (map[string]bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	out := make(map[string]bool)
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		body := string(data)

		for _, idx := range createTableRe.FindAllStringSubmatchIndex(body, -1) {
			if len(idx) < 4 {
				continue
			}
			table := strings.ToLower(body[idx[2]:idx[3]])
			openParen := idx[1] - 1
			tableBody, ok := extractParenBody(body, openParen)
			if !ok {
				continue
			}
			if workspaceRe.MatchString(tableBody) {
				out[table] = true
			} else if _, exists := out[table]; !exists {
				out[table] = false
			}
		}
		for _, m := range alterAddRe.FindAllStringSubmatch(body, -1) {
			out[strings.ToLower(m[1])] = true
		}
		for _, m := range alterDropRe.FindAllStringSubmatch(body, -1) {
			out[strings.ToLower(m[1])] = false
		}
	}
	return out, nil
}

// extractParenBody returns the substring inside the balanced parens
// starting at s[openParen]. SQL CREATE TABLE bodies contain nested
// parens (UNIQUE constraints, FOREIGN KEY clauses), so a regex cannot
// extract the body; a depth counter can.
func extractParenBody(s string, openParen int) (string, bool) {
	if openParen < 0 || openParen >= len(s) || s[openParen] != '(' {
		return "", false
	}
	depth := 0
	for i := openParen; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return s[openParen+1 : i], true
			}
		}
	}
	return "", false
}
