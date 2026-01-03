package cache

import (
	"fmt"
	"github.com/google/uuid"
)

const (
	ArticleIDKey   = "cache:article:%s"
	ArticleSlugKey = "cache:article:%s"

	ArticleLikeOnceUserIDKey = "idempotency:article:like:%s:%s"
	ArticleViewOnceUserIDKey = "idempotency:article:view:%s:%s"
	ArticleLikeOnceGuestKey  = "idempotency:article:like:%s:%s"
	ArticleViewOnceGuestKey  = "idempotency:article:view:%s:%s"
)

func GetArticleIDKey(id uuid.UUID) string {
	return fmt.Sprintf(ArticleIDKey, id.String())
}

func GetArticleSlugKey(slug string) string {
	return fmt.Sprintf(ArticleSlugKey, slug)
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
