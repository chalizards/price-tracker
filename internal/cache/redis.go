package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(redisURL string) (*RedisClient, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	redisClient := &RedisClient{client: redis.NewClient(opts)}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx); err != nil {
		return nil, err
	}

	return redisClient, nil
}

func (redis *RedisClient) Ping(ctx context.Context) error {
	if err := redis.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}

	return nil
}

func (redis *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return redis.client.Get(ctx, key).Result()
}

func (redis *RedisClient) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return redis.client.Set(ctx, key, value, ttl).Err()
}

func (redis *RedisClient) Close() error {
	return redis.client.Close()
}
