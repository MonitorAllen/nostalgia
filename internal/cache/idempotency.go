package cache

import (
	"context"

	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	"github.com/google/uuid"
)

type IdempotencyCache struct {
	cache Cache
}

func NewIdempotencyCache(cache Cache) *IdempotencyCache {
	return &IdempotencyCache{cache: cache}
}

func (i *IdempotencyCache) MarkArticleLikeByUser(ctx context.Context, articleID uuid.UUID, userID uuid.UUID) (bool, error) {
	if i == nil || i.cache == nil {
		return true, nil
	}
	return i.cache.SetNX(ctx, key.GetArticleLikeOnceUserIDKey(articleID, userID), 1, AuthenticatedLikeIdempotencyTTL)
}

func (i *IdempotencyCache) MarkArticleLikeByGuest(ctx context.Context, articleID uuid.UUID, ip string) (bool, error) {
	if i == nil || i.cache == nil {
		return true, nil
	}
	return i.cache.SetNX(ctx, key.GetArticleLikeOnceGuestKey(articleID, ip), 1, GuestLikeIdempotencyTTL)
}

func (i *IdempotencyCache) MarkArticleViewByUser(ctx context.Context, articleID uuid.UUID, userID uuid.UUID) (bool, error) {
	if i == nil || i.cache == nil {
		return true, nil
	}
	return i.cache.SetNX(ctx, key.GetArticleViewOnceUserIDKey(articleID, userID), 1, ArticleViewIdempotencyTTL)
}

func (i *IdempotencyCache) MarkArticleViewByGuest(ctx context.Context, articleID uuid.UUID, ip string) (bool, error) {
	if i == nil || i.cache == nil {
		return true, nil
	}
	return i.cache.SetNX(ctx, key.GetArticleViewOnceGuestKey(articleID, ip), 1, ArticleViewIdempotencyTTL)
}
