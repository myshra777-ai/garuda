package store

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/techtaytor/garuda/internal/types"
)

func TestRefCountFlusherIntegration(t *testing.T) {
	// 1. Setup test databases
	// In a real CI environment, you'd use ephemeral containers.
	// For local testing, we assume PostgreSQL and Redis are running on standard ports.
	ctx := context.Background() // <-- Add this at the top of the test

	dbURL := "postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
	redisAddr := "localhost:6380"

	// 2. Create store and Redis client
	store, err := NewPostgresStore(dbURL)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer rdb.Close()

	// 3. Create a test task manifest with two content blocks
	taskID := uuid.New()
	customerID := uuid.New()

	// Create two content blocks
	block1Content := "function processRefund(amount) { if (amount > 500) { return 'requires approval'; } return 'approved'; }"
	block2Content := "policy: refunds over $500 require manager approval"

	block1Hash := sha256.Sum256([]byte(block1Content))
	block2Hash := sha256.Sum256([]byte(block2Content))

	blocks := []types.Block{
		{
			Hash:      block1Hash,
			Content:   block1Content,
			RefCount:  1,
			CreatedAt: time.Now(),
		},
		{
			Hash:      block2Hash,
			Content:   block2Content,
			RefCount:  1,
			CreatedAt: time.Now(),
		},
	}

	// 4. Ingest the blocks
	if err := store.IngestBlocks(ctx, blocks); err != nil {
		t.Fatalf("failed to ingest blocks: %v", err)
	}

	// 5. Create and save the task manifest
	manifest := &types.TaskManifest{
		TaskID:         taskID,
		CustomerID:     customerID,
		CredentialRef:  "vault://path/to/key",
		Title:          "Analyze refund policy",
		ScopeDomain:    "finance",
		ScopeSystem:    "refunds",
		Status:         types.StatusInProgress,
		ManifestBlocks: []types.BlockHash{block1Hash, block2Hash},
		NormalizedIR: types.UniversalContextState{
			Version:        1,
			Progress:       20,
			FilesTouched:   []string{"internal/refund/policy.go"},
			Reasoning:      "Found conflicting policies D-001 and D-002",
			RemainingSteps: []string{"Check D-003", "Verify compliance"},
			DiscoveredFacts: []types.DiscoveredFact{
				{Type: "decision", Source: "D-001", Target: "D-002"},
			},
		},
		IRVersion:   1,
		DecisionIDs: []uuid.UUID{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := store.SaveTaskManifest(ctx, manifest); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// 6. Emit reference changes (increment ref counts)
	refManager := NewRefCountManager(rdb)

	// Each block gets +1 (simulate an agent using the blocks)
	if err := refManager.EmitReferenceChange(ctx, block1Hash, 1); err != nil {
		t.Fatalf("failed to emit reference change for block1: %v", err)
	}
	if err := refManager.EmitReferenceChange(ctx, block2Hash, 1); err != nil {
		t.Fatalf("failed to emit reference change for block2: %v", err)
	}

	// 7. Run the flusher once (manually trigger a flush)
	// In production, this runs as a background goroutine.
	// For the test, we call flush directly.
	if err := store.flush(ctx, rdb); err != nil {
		t.Fatalf("failed to flush: %v", err)
	}

	// 8. Verify the block ref counts were updated
	// Query the cas_blocks table and check the ref counts.
	var refCount1, refCount2 int
	err = store.pool.QueryRow(ctx, "SELECT ref_count FROM cas_blocks WHERE block_hash = $1", block1Hash[:]).Scan(&refCount1)
	if err != nil {
		t.Fatalf("failed to query block1 ref count: %v", err)
	}
	err = store.pool.QueryRow(ctx, "SELECT ref_count FROM cas_blocks WHERE block_hash = $1", block2Hash[:]).Scan(&refCount2)
	if err != nil {
		t.Fatalf("failed to query block2 ref count: %v", err)
	}

	// Expected: 1 (initial) + 1 (emitted) = 2
	if refCount1 != 2 {
		t.Errorf("block1 ref_count expected 2, got %d", refCount1)
	}
	if refCount2 != 2 {
		t.Errorf("block2 ref_count expected 2, got %d", refCount2)
	}

	// 9. Verify the manifest can be retrieved
	// We'll need to add a GetTaskManifest method to the store.
	// For now, we just check that the manifest exists in the database.
	var exists bool
	err = store.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM task_manifests WHERE task_id = $1)", taskID).Scan(&exists)
	if err != nil {
		t.Fatalf("failed to check manifest existence: %v", err)
	}
	if !exists {
		t.Errorf("manifest not found in database")
	}
}
