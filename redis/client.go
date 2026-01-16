package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/oceanmining/game-manager/config"
)

type Client struct {
	rdb *redis.Client
}

func NewClient(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

func (c *Client) ExampleCacheUser(ctx context.Context, userID int64, userData string, ttl time.Duration) error {
	key := fmt.Sprintf("user:%d", userID)
	return c.Set(ctx, key, userData, ttl)
}

func (c *Client) ExampleGetCachedUser(ctx context.Context, userID int64) (string, error) {
	key := fmt.Sprintf("user:%d", userID)
	return c.Get(ctx, key)
}
