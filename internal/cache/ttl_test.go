package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCacheTTLConstants(t *testing.T) {
	require.Equal(t, 24*time.Hour, ArticleDetailTTL)
	require.Equal(t, 15*time.Minute, ArticleListTTL)
	require.Equal(t, 5*time.Minute, EmptyArticleListTTL)
	require.Equal(t, 12*time.Hour, CategoryListTTL)
	require.Equal(t, 12*time.Hour, ContributionsTTL)
	require.Equal(t, 365*24*time.Hour, AuthenticatedLikeIdempotencyTTL)
	require.Equal(t, 7*24*time.Hour, GuestLikeIdempotencyTTL)
	require.Equal(t, 24*time.Hour, ArticleViewIdempotencyTTL)
}

func TestWithJitterKeepsTTLWithinTenPercent(t *testing.T) {
	base := time.Hour

	for i := 0; i < 100; i++ {
		jittered := WithJitter(base)
		require.GreaterOrEqual(t, jittered, base)
		require.LessOrEqual(t, jittered, base+base/10)
	}
}

func TestWithJitterLeavesNonPositiveTTLUnchanged(t *testing.T) {
	require.Equal(t, time.Duration(0), WithJitter(0))
	require.Equal(t, -time.Second, WithJitter(-time.Second))
}
