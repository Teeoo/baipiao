package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

type Cache struct {
	client redis.UniversalClient
	ttl    time.Duration
}

const prefix = "gql:"

func NewCache(redisAddress string, ttl time.Duration) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})

	err := client.Ping().Err()
	if err != nil {
		return nil, fmt.Errorf("could not create cache: %w", err)
	}

	return &Cache{client: client, ttl: ttl}, nil
}

func (c *Cache) Add(ctx context.Context, key string, value interface{}) {
	c.client.Set(prefix+key, value, c.ttl)
}

func (c *Cache) Get(ctx context.Context, key string) (interface{}, bool) {
	s, err := c.client.Get(prefix + key).Result()
	if err != nil {
		return struct{}{}, false
	}
	return s, true
}
