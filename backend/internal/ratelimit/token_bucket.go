package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Decision represents the outcome of a rate limit check.
type Decision struct {
	Allowed    bool
	RetryAfter time.Duration
	Remaining  float64
}

// Limiter defines the interface for rate limiters.
type Limiter interface {
	Allow(ctx context.Context, key string) (Decision, error)
}

// TokenBucket implements a Redis-backed token bucket limiter.
type TokenBucket struct {
	client   *redis.Client
	prefix   string
	rate     float64
	capacity float64
	interval time.Duration
	script   *redis.Script
}

// NewTokenBucket creates a new token bucket limiter.
func NewTokenBucket(client *redis.Client, prefix string, rate int, interval time.Duration) *TokenBucket {
	if rate <= 0 {
		panic("ratelimit: rate must be positive")
	}
	if interval <= 0 {
		panic("ratelimit: interval must be positive")
	}

	capacity := float64(rate)

	script := redis.NewScript(`
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local interval = tonumber(ARGV[3])
local now = tonumber(ARGV[4])

local data = redis.call("HMGET", key, "tokens", "timestamp")
local tokens = tonumber(data[1])
local timestamp = tonumber(data[2])

if tokens == nil then
	tokens = capacity
	timestamp = now
end

local delta = now - timestamp
if delta > 0 then
	local refill = delta * rate / interval
	tokens = math.min(capacity, tokens + refill)
	timestamp = now
end

if tokens < 1 then
	redis.call("HMSET", key, "tokens", tokens, "timestamp", timestamp)
	redis.call("PEXPIRE", key, math.ceil(interval))
	local deficit = 1 - tokens
	local wait = math.ceil(deficit * interval / rate)
	return {0, wait, tokens}
end

tokens = tokens - 1
redis.call("HMSET", key, "tokens", tokens, "timestamp", timestamp)
redis.call("PEXPIRE", key, math.ceil(interval))
return {1, 0, tokens}
`)

	return &TokenBucket{
		client:   client,
		prefix:   prefix,
		rate:     float64(rate),
		capacity: capacity,
		interval: interval,
		script:   script,
	}
}

// Allow attempts to consume a token for the provided key.
func (t *TokenBucket) Allow(ctx context.Context, key string) (Decision, error) {
	if t.client == nil {
		return Decision{}, fmt.Errorf("ratelimit: redis client is nil")
	}

	redisKey := fmt.Sprintf("%s:%s", t.prefix, key)
	now := time.Now().UnixMilli()

	args := []interface{}{
		t.rate,
		t.capacity,
		t.interval.Milliseconds(),
		now,
	}

	result, err := t.script.Run(ctx, t.client, []string{redisKey}, args...).Result()
	if err != nil {
		return Decision{}, fmt.Errorf("ratelimit: script execution failed: %w", err)
	}

	values, ok := result.([]interface{})
	if !ok || len(values) < 3 {
		return Decision{}, fmt.Errorf("ratelimit: unexpected script result %v", result)
	}

	allowed := values[0]
	retryRaw := values[1]
	remainingRaw := values[2]

	var decision Decision

	switch v := allowed.(type) {
	case int64:
		decision.Allowed = v == 1
	case float64:
		decision.Allowed = v == 1
	default:
		return Decision{}, fmt.Errorf("ratelimit: unexpected allowed type %T", allowed)
	}

	if retry, ok := retryRaw.(int64); ok {
		decision.RetryAfter = time.Duration(retry) * time.Millisecond
	} else if retryFloat, ok := retryRaw.(float64); ok {
		decision.RetryAfter = time.Duration(retryFloat) * time.Millisecond
	}

	switch remaining := remainingRaw.(type) {
	case int64:
		decision.Remaining = float64(remaining)
	case float64:
		decision.Remaining = remaining
	default:
		decision.Remaining = 0
	}

	return decision, nil
}
