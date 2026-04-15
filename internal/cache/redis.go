package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

// NewCache creates a new cache instance
func NewCache(addr string) *Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &Cache{client: rdb}
}

// Get retrieves a value from the cache
func (c *Cache) Get(ctx context.Context, key string, dest any) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Set sets a value in the cache
func (c *Cache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, ttl).Err()
}

// Flush flushes the cache
func (c *Cache) Flush(ctx context.Context) error {
	return c.client.FlushAll(ctx).Err()
}
