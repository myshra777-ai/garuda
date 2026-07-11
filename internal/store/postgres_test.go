package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/techtaytor/garuda/internal/types"
)

// TestMain runs migrations before any tests
func TestMain(m *testing.M) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		// Run migrations
		if err := Migrate(dbURL, "../../migrations"); err != nil {
			fmt.Printf("Failed to run migrations: %v\n", err)
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}

// cleanupDB removes all data from the test database
func cleanupDB(t *testing.T, dbURL string) {
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("failed to connect to database for cleanup: %v", err)
	}
	defer pool.Close()

	// Truncate all tables
	_, err = pool.Exec(context.Background(), `
		TRUNCATE TABLE audit_events, decisions RESTART IDENTITY CASCADE;
	`)
	if err != nil {
		t.Fatalf("failed to clean up database: %v", err)
	}
}

func TestPostgresStore_AppendAndGet(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	// Clean up before test
	cleanupDB(t, dbURL)

	// 1. Create store
	store, err := NewPostgresStore(dbURL)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// 2. Create a decision with DRAFT status
	rec := &registry.DecisionRecord{
		ID:         "D-001",
		Version:    1,
		Decision:   "Use PostgreSQL for financial records",
		Rationale:  "ACID compliance is required.",
		Scope:      registry.Scope{Domain: "infrastructure", System: "database"},
		Status:     registry.StatusDraft,
		Confidence: 0.9,
		Owner:      "alice@company.com",
		Approvers:  []string{"bob@company.com"},
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	// 3. Append
	if err := store.Append(rec, "alice@company.com"); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// 4. Get
	retrieved, err := store.Get("D-001")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved.Decision != "Use PostgreSQL for financial records" {
		t.Errorf("expected decision text to match, got %s", retrieved.Decision)
	}
	if retrieved.Status != registry.StatusDraft {
		t.Errorf("expected status DRAFT, got %s", retrieved.Status)
	}

	// 5. Update
	retrieved.Decision = "Use PostgreSQL for all financial data"
	if err := store.Update(retrieved, "alice@company.com"); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// 6. Verify update
	updated, err := store.Get("D-001")
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}
	if updated.Decision != "Use PostgreSQL for all financial data" {
		t.Errorf("expected updated decision text, got %s", updated.Decision)
	}
	if updated.Version != 2 {
		t.Errorf("expected version 2, got %d", updated.Version)
	}

	// 7. Transition to REVIEW
	if err := store.Transition("D-001", registry.StatusReview, "alice@company.com"); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}

	// 8. Verify transition
	transitioned, err := store.Get("D-001")
	if err != nil {
		t.Fatalf("Get after transition failed: %v", err)
	}
	if transitioned.Status != registry.StatusReview {
		t.Errorf("expected status REVIEW, got %s", transitioned.Status)
	}
}

func TestPostgresStore_ListByScope(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	// Clean up before test
	cleanupDB(t, dbURL)

	// 1. Create store
	store, err := NewPostgresStore(dbURL)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// 2. Create two decisions with different scopes
	rec1 := &registry.DecisionRecord{
		ID:         "D-001",
		Version:    1,
		Decision:   "Use PostgreSQL",
		Scope:      registry.Scope{Domain: "infrastructure", System: "database"},
		Status:     registry.StatusDraft,
		Confidence: 0.9,
		Owner:      "alice@company.com",
		Approvers:  []string{"bob@company.com"},
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	rec2 := &registry.DecisionRecord{
		ID:         "D-002",
		Version:    1,
		Decision:   "Use AWS for hosting",
		Scope:      registry.Scope{Domain: "infrastructure", System: "hosting"},
		Status:     registry.StatusDraft,
		Confidence: 0.9,
		Owner:      "alice@company.com",
		Approvers:  []string{"bob@company.com"},
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	// 3. Append both
	if err := store.Append(rec1, "alice@company.com"); err != nil {
		t.Fatalf("Append rec1 failed: %v", err)
	}
	if err := store.Append(rec2, "alice@company.com"); err != nil {
		t.Fatalf("Append rec2 failed: %v", err)
	}

	// 4. List by scope (database system)
	results, err := store.ListByScope(
		registry.Scope{System: "database"},
		[]registry.DecisionStatus{registry.StatusDraft},
	)
	if err != nil {
		t.Fatalf("ListByScope failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "D-001" {
		t.Errorf("expected D-001, got %s", results[0].ID)
	}

	// 5. List by scope (hosting system)
	results, err = store.ListByScope(
		registry.Scope{System: "hosting"},
		[]registry.DecisionStatus{registry.StatusDraft},
	)
	if err != nil {
		t.Fatalf("ListByScope failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "D-002" {
		t.Errorf("expected D-002, got %s", results[0].ID)
	}
}
