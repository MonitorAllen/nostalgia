package cache

import (
	"context"
	"testing"

	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	"github.com/stretchr/testify/require"
)

func TestContributionCacheStoresWithPolicyTTL(t *testing.T) {
	fake := newFakeCache()
	contributionCache := NewContributionCache(fake)
	value := map[string]any{"total": 42}

	err := contributionCache.Set(context.Background(), value)
	require.NoError(t, err)

	cacheKey := key.GetUserContributionsKey()
	require.Contains(t, fake.values, cacheKey)
	require.GreaterOrEqual(t, fake.ttls[cacheKey], ContributionsTTL)
	require.LessOrEqual(t, fake.ttls[cacheKey], ContributionsTTL+ContributionsTTL/10)
}
