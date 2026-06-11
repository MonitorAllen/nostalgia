package key

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	ArticleIDKey                  = "cache:article:id:%s"
	ArticleSlugKey                = "cache:article:slug:%s"
	ArticleCommentKey             = "cache:article:comment:%s"
	ArticleListVersionAllKey      = "cache:article:list:version:all"
	ArticleListVersionCategoryKey = "cache:article:list:version:category:%d"
	ArticleListKey                = "cache:article:list:v:%d:category:%s:page:%d:limit:%d"

	ArticleLikeOnceUserIDKey = "idempotency:article:like:user:%s:%s"
	ArticleViewOnceUserIDKey = "idempotency:article:view:user:%s:%s"
	ArticleLikeOnceGuestKey  = "idempotency:article:like:guest:%s:%s"
	ArticleViewOnceGuestKey  = "idempotency:article:view:guest:%s:%s"
)

func GetArticleIDKey(id uuid.UUID) string {
	return fmt.Sprintf(ArticleIDKey, id.String())
}

func GetArticleSlugKey(slug string) string {
	return fmt.Sprintf(ArticleSlugKey, slug)
}

func GetArticleListVersionKey(categoryID int64) string {
	if categoryID == 0 {
		return ArticleListVersionAllKey
	}
	return fmt.Sprintf(ArticleListVersionCategoryKey, categoryID)
}

func GetArticleListKey(version int64, categoryID int64, page int32, limit int32) string {
	return fmt.Sprintf(ArticleListKey, version, articleListCategoryBucket(categoryID), page, limit)
}

func articleListCategoryBucket(categoryID int64) string {
	if categoryID == 0 {
		return "all"
	}
	return fmt.Sprintf("%d", categoryID)
}

func GetArticleLikeOnceUserIDKey(articleID uuid.UUID, userID uuid.UUID) string {
	return fmt.Sprintf(ArticleLikeOnceUserIDKey, articleID.String(), userID.String())
}

func GetArticleViewOnceUserIDKey(articleID uuid.UUID, userID uuid.UUID) string {
	return fmt.Sprintf(ArticleViewOnceUserIDKey, articleID.String(), userID.String())
}

func GetArticleLikeOnceGuestKey(articleID uuid.UUID, ip string) string {
	return fmt.Sprintf(ArticleLikeOnceGuestKey, articleID.String(), ip)
}

func GetArticleViewOnceGuestKey(articleID uuid.UUID, ip string) string {
	return fmt.Sprintf(ArticleViewOnceGuestKey, articleID.String(), ip)
}
