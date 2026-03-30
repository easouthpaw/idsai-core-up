package cache

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient is a thin wrapper around *redis.Client with graceful degradation.
// If Redis becomes unavailable, operations log a warning and return as if the
// key was not found. This prevents a Redis outage from cascading into RBAC failures.
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new RedisClient and pings the server.
// If ping fails, a warning is logged but the client is still returned
// (graceful degradation — callers fall through to the database).
func NewRedisClient(addr, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		DialTimeout:  3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("[WARN] redis: ping failed (addr=%s): %v — RBAC cache disabled, falling through to DB", addr, err)
	} else {
		log.Printf("[INFO] redis: connected to %s, db=%d", addr, db)
	}

	return &RedisClient{client: rdb}
}

// Get retrieves a value by key. Returns ("", false, nil) on cache miss.
// On Redis error the miss is returned with a logged warning.
func (r *RedisClient) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		log.Printf("[WARN] redis.Get(%s): %v", key, err)
		return "", false, nil // graceful degradation
	}
	return val, true, nil
}

// Set stores a key with TTL. Errors are logged but not propagated.
func (r *RedisClient) Set(ctx context.Context, key, value string, ttl time.Duration) {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		log.Printf("[WARN] redis.Set(%s): %v", key, err)
	}
}

// Del deletes one or more keys. Errors are logged but not propagated.
func (r *RedisClient) Del(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}
	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		log.Printf("[WARN] redis.Del(%v): %v", keys, err)
	}
}

// DelByPattern deletes all keys matching a glob pattern (e.g. "rbac:user:<id>:*").
// Uses SCAN internally to avoid blocking Redis with KEYS in large databases.
func (r *RedisClient) DelByPattern(ctx context.Context, pattern string) {
	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			log.Printf("[WARN] redis.Scan(%s): %v", pattern, err)
			return
		}
		if len(keys) > 0 {
			r.Del(ctx, keys...)
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
}

// Close shuts down the Redis connection.
func (r *RedisClient) Close() error {
	return r.client.Close()
}
