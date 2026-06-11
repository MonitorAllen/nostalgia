package key

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestArticleCacheKeysAreNamespaced(t *testing.T) {
	articleID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	require.Equal(t, "cache:article:id:11111111-1111-1111-1111-111111111111", GetArticleIDKey(articleID))
	require.Equal(t, "cache:article:slug:hello-cache", GetArticleSlugKey("hello-cache"))
	require.NotEqual(t, GetArticleIDKey(articleID), GetArticleSlugKey(articleID.String()))
}

func TestArticleListCacheKeysUseVersionAndBucket(t *testing.T) {
	require.Equal(t, "cache:article:list:version:all", GetArticleListVersionKey(0))
	require.Equal(t, "cache:article:list:version:category:7", GetArticleListVersionKey(7))

	require.Equal(t, "cache:article:list:v:3:category:all:page:1:limit:10", GetArticleListKey(3, 0, 1, 10))
	require.Equal(t, "cache:article:list:v:4:category:7:page:2:limit:20", GetArticleListKey(4, 7, 2, 20))
}

func TestArticleIdempotencyKeysIncludeActorType(t *testing.T) {
	articleID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	require.Equal(t, "idempotency:article:like:user:11111111-1111-1111-1111-111111111111:22222222-2222-2222-2222-222222222222", GetArticleLikeOnceUserIDKey(articleID, userID))
	require.Equal(t, "idempotency:article:like:guest:11111111-1111-1111-1111-111111111111:127.0.0.1", GetArticleLikeOnceGuestKey(articleID, "127.0.0.1"))
	require.Equal(t, "idempotency:article:view:user:11111111-1111-1111-1111-111111111111:22222222-2222-2222-2222-222222222222", GetArticleViewOnceUserIDKey(articleID, userID))
	require.Equal(t, "idempotency:article:view:guest:11111111-1111-1111-1111-111111111111:127.0.0.1", GetArticleViewOnceGuestKey(articleID, "127.0.0.1"))
}
