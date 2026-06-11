package cache

import (
	"context"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	"github.com/google/uuid"
)

const MaxCachedArticleListPage int32 = 5

type ArticleCache struct {
	cache Cache
}

type ArticleListParams struct {
	CategoryID int64
	Page       int32
	Limit      int32
}

type ArticleListPage struct {
	Count    int64                `json:"count"`
	Articles []db.ListArticlesRow `json:"articles"`
}

func NewArticleCache(cache Cache) *ArticleCache {
	return &ArticleCache{cache: cache}
}

func (a *ArticleCache) GetByID(ctx context.Context, id uuid.UUID) (db.GetArticleRow, bool, error) {
	var article db.GetArticleRow
	if a == nil || a.cache == nil {
		return article, false, nil
	}
	ok, err := a.cache.Get(ctx, key.GetArticleIDKey(id), &article)
	return article, ok, err
}

func (a *ArticleCache) SetByID(ctx context.Context, article db.GetArticleRow) error {
	if a == nil || a.cache == nil {
		return nil
	}
	return a.cache.Set(ctx, key.GetArticleIDKey(article.ID), article, WithJitter(ArticleDetailTTL))
}

func (a *ArticleCache) GetBySlug(ctx context.Context, slug string) (db.GetArticleBySlugRow, bool, error) {
	var article db.GetArticleBySlugRow
	if a == nil || a.cache == nil {
		return article, false, nil
	}
	ok, err := a.cache.Get(ctx, key.GetArticleSlugKey(slug), &article)
	return article, ok, err
}

func (a *ArticleCache) SetBySlug(ctx context.Context, slug string, article db.GetArticleBySlugRow) error {
	if a == nil || a.cache == nil {
		return nil
	}
	return a.cache.Set(ctx, key.GetArticleSlugKey(slug), article, WithJitter(ArticleDetailTTL))
}

func (a *ArticleCache) GetList(ctx context.Context, params ArticleListParams) (ArticleListPage, bool, error) {
	var page ArticleListPage
	if a == nil || a.cache == nil {
		return page, false, nil
	}
	if !shouldCacheArticleListPage(params.Page) {
		return page, false, nil
	}

	version, err := a.listVersion(ctx, params.CategoryID)
	if err != nil {
		return page, false, err
	}

	ok, err := a.cache.Get(ctx, key.GetArticleListKey(version, params.CategoryID, params.Page, params.Limit), &page)
	return page, ok, err
}

func (a *ArticleCache) SetList(ctx context.Context, params ArticleListParams, page ArticleListPage) error {
	if a == nil || a.cache == nil {
		return nil
	}
	if !shouldCacheArticleListPage(params.Page) {
		return nil
	}

	version, err := a.listVersion(ctx, params.CategoryID)
	if err != nil {
		return err
	}

	ttl := ArticleListTTL
	if len(page.Articles) == 0 {
		ttl = EmptyArticleListTTL
	}

	return a.cache.Set(ctx, key.GetArticleListKey(version, params.CategoryID, params.Page, params.Limit), page, WithJitter(ttl))
}

func (a *ArticleCache) BumpListVersion(ctx context.Context, categoryIDs ...int64) error {
	if a == nil || a.cache == nil {
		return nil
	}
	keys := map[string]struct{}{
		key.GetArticleListVersionKey(0): {},
	}

	for _, categoryID := range categoryIDs {
		if categoryID == 0 {
			continue
		}
		keys[key.GetArticleListVersionKey(categoryID)] = struct{}{}
	}

	for cacheKey := range keys {
		if _, err := a.cache.Incr(ctx, cacheKey); err != nil {
			return err
		}
	}

	return nil
}

func (a *ArticleCache) InvalidateDetails(ctx context.Context, keys ...string) error {
	if a == nil || a.cache == nil {
		return nil
	}
	for _, cacheKey := range keys {
		if cacheKey == "" {
			continue
		}
		if err := a.cache.Del(ctx, cacheKey); err != nil {
			return err
		}
	}
	return nil
}

func (a *ArticleCache) listVersion(ctx context.Context, categoryID int64) (int64, error) {
	var version int64
	if a == nil || a.cache == nil {
		return 0, nil
	}
	ok, err := a.cache.Get(ctx, key.GetArticleListVersionKey(categoryID), &version)
	if err != nil || !ok {
		return 0, err
	}
	return version, nil
}

func shouldCacheArticleListPage(page int32) bool {
	return page >= 1 && page <= MaxCachedArticleListPage
}
