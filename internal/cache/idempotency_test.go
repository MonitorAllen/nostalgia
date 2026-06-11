package cache

import (
	"context"
	"testing"

	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestIdempotencyCacheMarksArticleLikeByUserWithFiniteTTL(t *testing.T) {
	fake := newFakeCache()
	idempotencyCache := NewIdempotencyCache(fake)
	articleID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	ok, err := idempotencyCache.MarkArticleLikeByUser(context.Background(), articleID, userID)
	require.NoError(t, err)
	require.True(t, ok)

	cacheKey := key.GetArticleLikeOnceUserIDKey(articleID, userID)
	require.Equal(t, AuthenticatedLikeIdempotencyTTL, fake.ttls[cacheKey])
}

func TestIdempotencyCacheMarksArticleLikeByGuest(t *testing.T) {
	fake := newFakeCache()
	idempotencyCache := NewIdempotencyCache(fake)
	articleID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	ok, err := idempotencyCache.MarkArticleLikeByGuest(context.Background(), articleID, "127.0.0.1")
	require.NoError(t, err)
	require.True(t, ok)

	cacheKey := key.GetArticleLikeOnceGuestKey(articleID, "127.0.0.1")
	require.Equal(t, GuestLikeIdempotencyTTL, fake.ttls[cacheKey])
}

func TestIdempotencyCacheMarksArticleView(t *testing.T) {
	fake := newFakeCache()
	idempotencyCache := NewIdempotencyCache(fake)
	articleID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	ok, err := idempotencyCache.MarkArticleViewByUser(context.Background(), articleID, userID)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, ArticleViewIdempotencyTTL, fake.ttls[key.GetArticleViewOnceUserIDKey(articleID, userID)])

	ok, err = idempotencyCache.MarkArticleViewByGuest(context.Background(), articleID, "127.0.0.1")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, ArticleViewIdempotencyTTL, fake.ttls[key.GetArticleViewOnceGuestKey(articleID, "127.0.0.1")])
}
