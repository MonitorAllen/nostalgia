package cache

import (
	"math/rand"
	"time"
)

const (
	ArticleDetailTTL                = 24 * time.Hour
	ArticleListTTL                  = 15 * time.Minute
	EmptyArticleListTTL             = 5 * time.Minute
	CategoryListTTL                 = 12 * time.Hour
	ContributionsTTL                = 12 * time.Hour
	AuthenticatedLikeIdempotencyTTL = 365 * 24 * time.Hour
	GuestLikeIdempotencyTTL         = 7 * 24 * time.Hour
	ArticleViewIdempotencyTTL       = 24 * time.Hour
)

func WithJitter(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return ttl
	}

	maxJitter := ttl / 10
	if maxJitter <= 0 {
		return ttl
	}

	return ttl + time.Duration(rand.Int63n(int64(maxJitter)+1))
}
