package cache

import (
	"context"
	"testing"
	"time"
)

func TestRedisCache_SetGet(t *testing.T) {
	// This test assumes Redis is running on localhost:6379
	// Skip if not available
	cache, err := NewRedisCache("localhost:6379", "", 0)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	key := "test:key"
	value := map[string]string{"foo": "bar"}

	// Set with TTL
	if err := cache.Set(ctx, key, value, 10*time.Second); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get
	var retrieved map[string]string
	found, err := cache.Get(ctx, key, &retrieved)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !found {
		t.Error("key not found")
	}
	if retrieved["foo"] != "bar" {
		t.Errorf("expected bar, got %s", retrieved["foo"])
	}

	// Delete
	if err := cache.Delete(ctx, key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	found, err = cache.Get(ctx, key, &retrieved)
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}
	if found {
		t.Error("key still exists after delete")
	}
}
