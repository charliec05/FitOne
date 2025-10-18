package redisclient

import (
	"context"
	"fmt"
	"time"

	"fitonex/backend/internal/config"

	"github.com/redis/go-redis/v9"
)

// New creates a configured Redis client and performs a connectivity check.
func New(cfg *config.Config) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("redis parse url: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return client, nil
}
