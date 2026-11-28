// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/myshra777-ai/garuda/migrations"
)

// migrationFile is one SQL migration, identified by its file name.
type migrationFile struct {
	Name    string
	Content []byte
}

// embeddedMigrations is the copy of migrations/ compiled into the binary.
// It is a variable so tests can substitute a small fixture.
var embeddedMigrations fs.FS = migrations.FS

// MigrationsDirEnv names the environment variable that points Migrate at a
// directory of .sql files instead of the embedded migrations.
const MigrationsDirEnv = "GARUDA_MIGRATIONS_DIR"

// loadMigrationFiles returns the migrations to apply, sorted by name.
//
// With dir empty, the migrations embedded at build time are used. That is
// the normal case. A non-empty dir is an explicit override and must contain
// .sql files. The second return value says which source was used.
func loadMigrationFiles(dir string) ([]migrationFile, string, error) {
	if dir != "" {
		paths, err := filepath.Glob(filepath.Join(dir, "*.sql"))
		if err != nil {
			return nil, "", fmt.Errorf("failed to list migration files: %w", err)
		}
		if len(paths) > 0 {
			sort.Strings(paths)
			files := make([]migrationFile, 0, len(paths))
			for _, p := range paths {
				content, err := os.ReadFile(p)
				if err != nil {
					return nil, "", fmt.Errorf("failed to read migration %s: %w", filepath.Base(p), err)
				}
				files = append(files, migrationFile{Name: filepath.Base(p), Content: content})
			}
			return files, dir, nil
		}
		return nil, "", fmt.Errorf("%s=%q has no .sql files", MigrationsDirEnv, dir)
	}

	names, err := fs.Glob(embeddedMigrations, "*.sql")
	if err != nil {
		return nil, "", fmt.Errorf("failed to list embedded migrations: %w", err)
	}
	sort.Strings(names)
	files := make([]migrationFile, 0, len(names))
	for _, n := range names {
		content, err := fs.ReadFile(embeddedMigrations, n)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read embedded migration %s: %w", n, err)
		}
		files = append(files, migrationFile{Name: n, Content: content})
	}
	return files, "embedded", nil
}
