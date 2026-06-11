package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestArticleCacheSkipsOutOfRangeListPages(t *testing.T) {
	fake := newFakeCache()
	articleCache := NewArticleCache(fake)
	params := ArticleListParams{Page: 6, Limit: 10}

	_, ok, err := articleCache.GetList(context.Background(), params)
	require.NoError(t, err)
	require.False(t, ok)

	err = articleCache.SetList(context.Background(), params, ArticleListPage{Count: 1, Articles: []db.ListArticlesRow{{ID: uuid.New()}}})
	require.NoError(t, err)
	require.Empty(t, fake.values)
}

func TestArticleCacheStoresListPagesWithVersionedKey(t *testing.T) {
	fake := newFakeCache()
	articleCache := NewArticleCache(fake)
	params := ArticleListParams{CategoryID: 7, Page: 2, Limit: 20}
	page := ArticleListPage{Count: 1, Articles: []db.ListArticlesRow{{ID: uuid.New(), Title: "cached"}}}

	err := fake.Set(context.Background(), key.GetArticleListVersionKey(7), int64(4), time.Hour)
	require.NoError(t, err)
	err = articleCache.SetList(context.Background(), params, page)
	require.NoError(t, err)

	cached, ok, err := articleCache.GetList(context.Background(), params)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, page.Count, cached.Count)
	require.Equal(t, page.Articles[0].Title, cached.Articles[0].Title)

	cacheKey := key.GetArticleListKey(4, 7, 2, 20)
	require.Contains(t, fake.values, cacheKey)
	require.GreaterOrEqual(t, fake.ttls[cacheKey], ArticleListTTL)
	require.LessOrEqual(t, fake.ttls[cacheKey], ArticleListTTL+ArticleListTTL/10)
}

func TestArticleCacheUsesShortTTLForEmptyListPages(t *testing.T) {
	fake := newFakeCache()
	articleCache := NewArticleCache(fake)
	params := ArticleListParams{Page: 1, Limit: 10}

	err := articleCache.SetList(context.Background(), params, ArticleListPage{Count: 0, Articles: []db.ListArticlesRow{}})
	require.NoError(t, err)

	cacheKey := key.GetArticleListKey(0, 0, 1, 10)
	require.GreaterOrEqual(t, fake.ttls[cacheKey], EmptyArticleListTTL)
	require.LessOrEqual(t, fake.ttls[cacheKey], EmptyArticleListTTL+EmptyArticleListTTL/10)
}

func TestArticleCacheBumpsAllAndCategoryListVersions(t *testing.T) {
	fake := newFakeCache()
	articleCache := NewArticleCache(fake)

	err := articleCache.BumpListVersion(context.Background(), 7, 7, 0, 8)
	require.NoError(t, err)

	require.Equal(t, int64(1), fake.increments[key.GetArticleListVersionKey(0)])
	require.Equal(t, int64(1), fake.increments[key.GetArticleListVersionKey(7)])
	require.Equal(t, int64(1), fake.increments[key.GetArticleListVersionKey(8)])
}

type fakeCache struct {
	values     map[string][]byte
	ttls       map[string]time.Duration
	increments map[string]int64
}

func newFakeCache() *fakeCache {
	return &fakeCache{
		values:     make(map[string][]byte),
		ttls:       make(map[string]time.Duration),
		increments: make(map[string]int64),
	}
}

func (f *fakeCache) Ping(context.Context) error {
	return nil
}

func (f *fakeCache) Get(_ context.Context, cacheKey string, dest any) (bool, error) {
	value, ok := f.values[cacheKey]
	if !ok {
		return false, nil
	}
	return true, json.Unmarshal(value, dest)
}

func (f *fakeCache) Set(_ context.Context, cacheKey string, value any, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	f.values[cacheKey] = bytes
	f.ttls[cacheKey] = ttl
	return nil
}

func (f *fakeCache) Del(_ context.Context, cacheKey string) error {
	delete(f.values, cacheKey)
	delete(f.ttls, cacheKey)
	return nil
}

func (f *fakeCache) SetNX(ctx context.Context, cacheKey string, value interface{}, ttl time.Duration) (bool, error) {
	if _, ok := f.values[cacheKey]; ok {
		return false, nil
	}
	return true, f.Set(ctx, cacheKey, value, ttl)
}

func (f *fakeCache) Incr(_ context.Context, cacheKey string) (int64, error) {
	f.increments[cacheKey]++
	value := f.increments[cacheKey]
	bytes, err := json.Marshal(value)
	if err != nil {
		return 0, err
	}
	f.values[cacheKey] = bytes
	return value, nil
}

func (f *fakeCache) IsExpired(context.Context, string) (bool, error) {
	return false, nil
}

func (f *fakeCache) Close() error {
	return nil
}
