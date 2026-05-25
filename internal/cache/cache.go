package cache

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	if r == nil || r.client == nil {
		return "", redis.Nil
	}
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		// don't fail hard on cache issues
		log.Println("redis get error:", err)
	}
	return val, err
}

func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if r == nil || r.client == nil {
		return nil
	}
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		log.Println("redis set error:", err)
	}
	return err
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if r == nil || r.client == nil {
		return nil
	}
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		log.Println("redis del error:", err)
	}
	return err
}
