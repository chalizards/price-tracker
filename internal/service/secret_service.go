package service

import (
	"context"
	"fmt"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/chalizards/price-tracker/internal/cache"
	"github.com/redis/go-redis/v9"
)

const secretCacheTTL = 1 * time.Hour

func GetGeminiSecret(ctx context.Context, redisClient *cache.RedisClient, secretName string) (string, error) {
	cached, err := redisClient.Get(ctx, "secret:"+secretName)
	if err == nil {
		return cached, nil
	}
	if err != redis.Nil {
		return "", fmt.Errorf("failed to read secret cache: %w", err)
	}

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secret manager client: %w", err)
	}
	defer client.Close()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName + "/versions/latest",
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %w", err)
	}

	value := string(result.Payload.Data)

	if cacheErr := redisClient.Set(ctx, "secret:"+secretName, value, secretCacheTTL); cacheErr != nil {
		return value, cacheErr
	}

	return value, nil
}
