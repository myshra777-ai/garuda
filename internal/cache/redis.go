// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is a wrapper around redis.Client that provides
// typed Get/Set/Delete methods with TTL support.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache client.
// addr example: "localhost:6379"
// password: empty string for no auth
// db: 0 for default
func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     20,
		MinIdleConns: 5,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

// Close closes the Redis client gracefully.
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// Get retrieves a value from the cache and unmarshals it into dest.
// Returns true if the key exists, false if it does not.
// Returns error if unmarshaling fails or the connection fails.
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil // key does not exist
		}
		return false, fmt.Errorf("redis get failed: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false, fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}

	return true, nil
}

// Set stores a value in the cache with an optional TTL.
// If ttl is 0, the key has no expiration.
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
	}

	if ttl > 0 {
		return c.client.SetEx(ctx, key, string(data), ttl).Err()
	}
	return c.client.Set(ctx, key, string(data), 0).Err()
}

// Delete removes one or more keys from the cache.
func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

// DeletePattern removes all keys matching a pattern (e.g., "decision:*").
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to list keys for pattern %s: %w", pattern, err)
	}
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists in the cache.
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return count > 0, nil
}
