package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrate runs all SQL migration files in the migrations directory.
// It creates the migrations table if it doesn't exist and applies
// each migration file in order.
func Migrate(connString string, migrationsDir string) error {
	// Connect to the database
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// Create migrations table if it doesn't exist
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Read migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}

	// Sort files by name (assuming numeric prefix like 001_)
	// For simplicity, we'll just apply them in alphabetical order
	// In production, use a proper migration tool like goose or migrate

	for _, file := range files {
		name := filepath.Base(file)

		// Check if migration has already been applied
		var exists bool
		err := pool.QueryRow(context.Background(),
			"SELECT EXISTS(SELECT 1 FROM migrations WHERE name = $1)", name).
			Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", name, err)
		}
		if exists {
			continue
		}

		// Read and execute the migration
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", name, err)
		}

		// Split by semicolon to handle multiple statements
		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			_, err = pool.Exec(context.Background(), stmt)
			if err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", name, err)
			}
		}

		// Record the migration
		_, err = pool.Exec(context.Background(),
			"INSERT INTO migrations (name) VALUES ($1)", name)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", name, err)
		}

		fmt.Printf("Applied migration: %s\n", name)
	}

	return nil
}
