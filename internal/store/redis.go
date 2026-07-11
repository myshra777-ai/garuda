package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// EnsureStreamGroup ensures that the Redis stream and consumer group exist idempotently.
func EnsureStreamGroup(ctx context.Context, rdb *redis.Client, stream, group string) error {
	_, err := rdb.XGroupCreate(ctx, stream, group, "0").Result()
	if err != nil {
		// If the error contains BUSYGROUP, the group already exists – we're safe.
		if err.Error() == "BUSYGROUP" || errors.Is(err, redis.Nil) {
			return nil
		}

		// If the stream doesn't exist yet, construct it cleanly with an initial token entry
		_, err = rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: stream,
			Values: map[string]interface{}{"init": "1"},
		}).Result()
		if err != nil {
			return fmt.Errorf("failed to create stream channel: %w", err)
		}

		// Re-attempt group initialization against the new active stream
		_, err = rdb.XGroupCreate(ctx, stream, group, "0").Result()
		if err != nil && err.Error() != "BUSYGROUP" {
			return fmt.Errorf("failed to bind consumer group: %w", err)
		}
	}
	return nil
}
