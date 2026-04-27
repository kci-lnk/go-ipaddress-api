package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// GlobalRedisKeyPrefix is the fixed prefix for all Redis keys
	GlobalRedisKeyPrefix = "ipaddress:"
	CacheKeySuffix       = "cache:"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
	prefix string
}

func New(client *redis.Client, ttlSeconds int) *Cache {
	return &Cache{
		client: client,
		ttl:    time.Duration(ttlSeconds) * time.Second,
		prefix: GlobalRedisKeyPrefix + CacheKeySuffix,
	}
}

func (c *Cache) key(ip string) string {
	return fmt.Sprintf("%s%s", c.prefix, ip)
}

func (c *Cache) Get(ctx context.Context, ip string) ([]byte, error) {
	data, err := c.client.Get(ctx, c.key(ip)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

func (c *Cache) Set(ctx context.Context, ip string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(ip), jsonData, c.ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, ip string) error {
	return c.client.Del(ctx, c.key(ip)).Err()
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
