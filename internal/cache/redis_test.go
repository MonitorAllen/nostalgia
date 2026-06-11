package cache

import (
	"testing"
	"time"

	"github.com/MonitorAllen/nostalgia/util"
	"github.com/stretchr/testify/require"
)

func TestNewRedisCacheUsesConfiguredDatabase(t *testing.T) {
	redisCache := NewRedisCache(util.Config{
		RedisAddress: "localhost:6379",
		RedisCacheDB: 2,
	})
	defer redisCache.Close()

	require.Equal(t, 2, redisCache.rdb.Options().DB)
}

func TestIsExpiredTTLClassification(t *testing.T) {
	require.False(t, isExpiredTTL(-time.Nanosecond), "permanent keys should not be treated as expired")
	require.True(t, isExpiredTTL(-2*time.Nanosecond), "missing keys should be treated as expired")
	require.False(t, isExpiredTTL(time.Minute), "keys with positive ttl should not be expired")
	require.True(t, isExpiredTTL(0), "zero ttl should be treated as expired")
}
