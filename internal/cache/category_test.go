package cache

import (
	"context"
	"testing"

	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	"github.com/stretchr/testify/require"
)

func TestCategoryCacheStoresListWithPolicyTTL(t *testing.T) {
	fake := newFakeCache()
	categoryCache := NewCategoryCache(fake)
	value := map[string]any{"count": 1}

	err := categoryCache.SetList(context.Background(), value)
	require.NoError(t, err)

	require.Contains(t, fake.values, key.CategoryAllKey)
	require.GreaterOrEqual(t, fake.ttls[key.CategoryAllKey], CategoryListTTL)
	require.LessOrEqual(t, fake.ttls[key.CategoryAllKey], CategoryListTTL+CategoryListTTL/10)
}

func TestCategoryCacheInvalidatesList(t *testing.T) {
	fake := newFakeCache()
	categoryCache := NewCategoryCache(fake)

	err := categoryCache.SetList(context.Background(), map[string]any{"count": 1})
	require.NoError(t, err)
	err = categoryCache.InvalidateList(context.Background())
	require.NoError(t, err)

	require.NotContains(t, fake.values, key.CategoryAllKey)
}
