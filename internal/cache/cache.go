package cache

import (
	"context"
	"time"
)

type Cache interface {
	Ping(ctx context.Context) error
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)
	IsExpired(ctx context.Context, key string) (bool, error)
	Close() error
}
