package cache

import (
	"context"

	"github.com/MonitorAllen/nostalgia/internal/cache/key"
)

type ContributionCache struct {
	cache Cache
}

func NewContributionCache(cache Cache) *ContributionCache {
	return &ContributionCache{cache: cache}
}

func (c *ContributionCache) Get(ctx context.Context, dest any) (bool, error) {
	if c == nil || c.cache == nil {
		return false, nil
	}
	return c.cache.Get(ctx, key.GetUserContributionsKey(), dest)
}

func (c *ContributionCache) Set(ctx context.Context, value any) error {
	if c == nil || c.cache == nil {
		return nil
	}
	return c.cache.Set(ctx, key.GetUserContributionsKey(), value, WithJitter(ContributionsTTL))
}
