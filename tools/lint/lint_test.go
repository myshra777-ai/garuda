// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package lint

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestLintRules walks the repo and reports every violation. This is
// the entry point: `go test ./tools/lint/` runs it. It is expected to
// pass with zero violations; if it fails, either the check is too
// narrow or the codebase has a real bug.
func TestLintRules(t *testing.T) {
	root := filepath.Join("..", "..")
	vs, err := Check(root)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(vs) == 0 {
		return
	}
	var b strings.Builder
	b.WriteString("\n")
	for _, v := range vs {
		b.WriteString(v.String())
		b.WriteString("\n\n")
	}
	t.Errorf("tools/lint: %d violation(s)%s", len(vs), b.String())
}

// TestSQLScopingFixtures asserts that the SQL scoping check fires on
// the shapes it is supposed to and stays silent on the shapes that
// look similar but are legitimate.
func TestSQLScopingFixtures(t *testing.T) {
	cases := []struct {
		file string
		want int
	}{
		{"testdata/sql_scoping/bad_tenant_only.go", 1},
		{"testdata/sql_scoping/bad_update_tenant_only.go", 1},
		{"testdata/sql_scoping/bad_insert_tenant_only.go", 1},
		{"testdata/sql_scoping/good_workspace_id.go", 0},
		{"testdata/sql_scoping/good_repository_id.go", 0},
		{"testdata/sql_scoping/good_pk_lookup.go", 0},
		{"testdata/sql_scoping/not_sql.go", 0},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			vs, err := checkFile(tc.file)
			if err != nil {
				t.Fatalf("checkFile: %v", err)
			}
			var sql []Violation
			for _, v := range vs {
				if v.Check == "sql-scoping" {
					sql = append(sql, v)
				}
			}
			if len(sql) != tc.want {
				t.Errorf("got %d sql-scoping violation(s), want %d", len(sql), tc.want)
				for _, v := range sql {
					t.Logf("  %s", v)
				}
			}
		})
	}
}

// TestVariadicFixtures asserts that the variadic check fires on the
// shape it targets and stays silent on legitimate variadic functions.
func TestVariadicFixtures(t *testing.T) {
	cases := []struct {
		file string
		want int
	}{
		{"testdata/variadic/bad_single_arg.go", 1},
		{"testdata/variadic/good_multi_arg.go", 0},
		{"testdata/variadic/good_range.go", 0},
		{"testdata/variadic/good_builtin.go", 0},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			vs, err := checkFile(tc.file)
			if err != nil {
				t.Fatalf("checkFile: %v", err)
			}
			var vv []Violation
			for _, v := range vs {
				if v.Check == "variadic-single-arg" {
					vv = append(vv, v)
				}
			}
			if len(vv) != tc.want {
				t.Errorf("got %d variadic violation(s), want %d", len(vv), tc.want)
				for _, v := range vv {
					t.Logf("  %s", v)
				}
			}
		})
	}
}

// TestTablesMatchMigrations asserts that WorkspaceScopedTables matches
// what the migrations declare. The migrations are the source of truth;
// a hand-edited map drifts.
func TestTablesMatchMigrations(t *testing.T) {
	migrations := filepath.Join("..", "..", "migrations")
	generated, err := parseMigrations(migrations)
	if err != nil {
		t.Fatalf("parseMigrations: %v", err)
	}
	for table, has := range generated {
		if has && !WorkspaceScopedTables[table] {
			t.Errorf("table %q has workspace_id in migrations but is missing from WorkspaceScopedTables", table)
		}
		if !has && WorkspaceScopedTables[table] {
			t.Errorf("table %q is in WorkspaceScopedTables but no migration declares workspace_id", table)
		}
	}
}

// TestAllowlistEntriesHaveReasons asserts that every allowlist entry
// carries a non-empty reason. An allowlist without reasons becomes a
// dumping ground.
func TestAllowlistEntriesHaveReasons(t *testing.T) {
	for i, e := range Allowlist {
		if strings.TrimSpace(e.Reason) == "" {
			t.Errorf("Allowlist[%d] (%s, %s, %s): empty reason", i, e.File, e.Check, e.Hash)
		}
	}
}
