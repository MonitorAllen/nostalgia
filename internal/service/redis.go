package service

import (
	"context"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisService struct {
	client *redis.Client
}

func NewRedisService(config util.Config) *RedisService {
	return &RedisService{
		client: redis.NewClient(&redis.Options{
			Addr: config.RedisAddress,
		}),
	}
}

func (r *RedisService) Get(key string) (string, error) {
	return r.client.Get(context.Background(), key).Result()
}

func (r *RedisService) Set(key string, value string, expiration time.Duration) error {
	return r.client.Set(context.Background(), key, value, expiration).Err()
}

func (r *RedisService) Del(key string) error {
	return r.client.Del(context.Background(), key).Err()
}

func (r *RedisService) Exists(key string) bool {
	return r.client.Exists(context.Background(), key).Val() == 1
}

func (r *RedisService) Close() error {
	return r.client.Close()
}

func (r *RedisService) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
