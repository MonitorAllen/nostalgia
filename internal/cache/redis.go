package cache

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisCache struct {
	rdb *redis.Client
}

// NewRedisCache Redis 实现
func NewRedisCache(config util.Config) *RedisCache {
	return &RedisCache{
		rdb: redis.NewClient(&redis.Options{
			Addr: config.RedisAddress,
			DB:   config.RedisCacheDB,
		}),
	}
}

func (r *RedisCache) Ping(ctx context.Context) error {
	return r.rdb.Ping(ctx).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string, dest any) (bool, error) {
	val, err := r.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.rdb.Set(ctx, key, bytes, ttl).Err()
}

func (r *RedisCache) Del(ctx context.Context, key string) error {
	return r.rdb.Del(ctx, key).Err()
}

func (r *RedisCache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return false, err
	}
	ok, err := r.rdb.SetNX(ctx, key, bytes, ttl).Result()
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (r *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return r.rdb.Incr(ctx, key).Result()
}

func (r *RedisCache) IsExpired(ctx context.Context, key string) (bool, error) {
	ttl, err := r.rdb.TTL(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return isExpiredTTL(ttl), nil
}

func (r *RedisCache) Close() error {
	return r.rdb.Close()
}

func isExpiredTTL(ttl time.Duration) bool {
	if ttl == time.Duration(-1) {
		return false
	}
	return ttl <= 0
}
