package cache

import (
	"context"

	"github.com/MonitorAllen/nostalgia/internal/cache/key"
)

type CategoryCache struct {
	cache Cache
}

func NewCategoryCache(cache Cache) *CategoryCache {
	return &CategoryCache{cache: cache}
}

func (c *CategoryCache) GetList(ctx context.Context, dest any) (bool, error) {
	if c == nil || c.cache == nil {
		return false, nil
	}
	return c.cache.Get(ctx, key.CategoryAllKey, dest)
}

func (c *CategoryCache) SetList(ctx context.Context, value any) error {
	if c == nil || c.cache == nil {
		return nil
	}
	return c.cache.Set(ctx, key.CategoryAllKey, value, WithJitter(CategoryListTTL))
}

func (c *CategoryCache) InvalidateList(ctx context.Context) error {
	if c == nil || c.cache == nil {
		return nil
	}
	return c.cache.Del(ctx, key.CategoryAllKey)
}
