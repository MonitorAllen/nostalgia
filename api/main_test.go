package api

import (
	"context"
	"os"
	"testing"
	"time"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/internal/cache"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/MonitorAllen/nostalgia/worker"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// noopCache is a cache that stores nothing — Get always reports "not found".
// Used as a default for tests that don't exercise cache behaviour.
type noopCache struct{}

func (noopCache) Ping(context.Context) error                            { return nil }
func (noopCache) Get(context.Context, string, any) (bool, error)        { return false, nil }
func (noopCache) Set(context.Context, string, any, time.Duration) error { return nil }
func (noopCache) Del(context.Context, string) error                     { return nil }
func (noopCache) SetNX(context.Context, string, any, time.Duration) (bool, error) {
	return true, nil
}
func (noopCache) Incr(context.Context, string) (int64, error)     { return 0, nil }
func (noopCache) IsExpired(context.Context, string) (bool, error) { return false, nil }
func (noopCache) Close() error                                    { return nil }

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor, cache cache.Cache) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	if cache == nil {
		cache = noopCache{}
	}

	server, err := NewServer(config, store, taskDistributor, cache)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
