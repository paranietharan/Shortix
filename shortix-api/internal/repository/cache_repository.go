package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheRepository interface {
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

type cacheRepository struct {
	redis *redis.Client
}

func NewCacheRepository(redis *redis.Client) CacheRepository {
	return &cacheRepository{
		redis: redis,
	}
}

func (r *cacheRepository) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.redis.Set(ctx, key, value, ttl).Err()
}

func (r *cacheRepository) Get(ctx context.Context, key string) (string, error) {
	return r.redis.Get(ctx, key).Result()
}

func (r *cacheRepository) Delete(ctx context.Context, key string) error {
	return r.redis.Del(ctx, key).Err()
}
