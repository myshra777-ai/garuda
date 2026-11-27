// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadMigrationFiles_PrefersDirectoryOnDisk(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"002_second.sql": "SELECT 2;",
		"001_first.sql":  "SELECT 1;",
		"README.md":      "not a migration",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "archive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "archive", "000_old.sql"), []byte("SELECT 0;"), 0o600); err != nil {
		t.Fatal(err)
	}

	files, source, err := loadMigrationFiles(dir)
	if err != nil {
		t.Fatalf("loadMigrationFiles: %v", err)
	}
	if source != dir {
		t.Fatalf("source = %q, want the directory on disk", source)
	}
	if len(files) != 2 || files[0].Name != "001_first.sql" || files[1].Name != "002_second.sql" {
		t.Fatalf("expected the two top-level .sql files in order, got %+v", files)
	}
	if string(files[0].Content) != "SELECT 1;" {
		t.Fatalf("content not read: %q", files[0].Content)
	}
}

func TestLoadMigrationFiles_FallsBackToEmbedded(t *testing.T) {
	saved := embeddedMigrations
	t.Cleanup(func() { embeddedMigrations = saved })
	embeddedMigrations = fstest.MapFS{
		"010_b.sql": {Data: []byte("SELECT 'b';")},
		"009_a.sql": {Data: []byte("SELECT 'a';")},
		"notes.txt": {Data: []byte("ignored")},
	}

	for _, dir := range []string{"", filepath.Join(t.TempDir(), "missing"), t.TempDir()} {
		files, source, err := loadMigrationFiles(dir)
		if err != nil {
			t.Fatalf("dir %q: %v", dir, err)
		}
		if source != "embedded" {
			t.Fatalf("dir %q: source = %q, want embedded", dir, source)
		}
		if len(files) != 2 || files[0].Name != "009_a.sql" || files[1].Name != "010_b.sql" {
			t.Fatalf("dir %q: unexpected files %+v", dir, files)
		}
	}
}

func TestEmbeddedMigrationsMatchRepository(t *testing.T) {
	// The embedded copy must contain exactly the top-level .sql files in
	// migrations/, so an installed binary applies the same schema as a
	// source checkout.
	repoDir := filepath.Join("..", "..", "migrations")
	onDisk, _ := filepath.Glob(filepath.Join(repoDir, "*.sql"))
	if len(onDisk) == 0 {
		t.Skip("migrations directory not found")
	}

	saved := embeddedMigrations
	t.Cleanup(func() { embeddedMigrations = saved })
	embedded, source, err := loadMigrationFiles("")
	if err != nil {
		t.Fatalf("loadMigrationFiles: %v", err)
	}
	if source != "embedded" {
		t.Fatalf("source = %q", source)
	}
	if len(embedded) != len(onDisk) {
		t.Fatalf("embedded %d migrations, repository has %d", len(embedded), len(onDisk))
	}
	for i, f := range embedded {
		if f.Name != filepath.Base(onDisk[i]) {
			t.Fatalf("migration %d: embedded %q, on disk %q", i, f.Name, filepath.Base(onDisk[i]))
		}
		want, err := os.ReadFile(onDisk[i])
		if err != nil {
			t.Fatal(err)
		}
		if string(f.Content) != string(want) {
			t.Fatalf("embedded %s differs from the file on disk", f.Name)
		}
		if !strings.HasSuffix(f.Name, ".sql") {
			t.Fatalf("non-SQL file embedded: %s", f.Name)
		}
	}
}
