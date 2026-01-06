package gapi

import (
	"context"
	"fmt"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/internal/cache"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/MonitorAllen/nostalgia/worker"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"testing"
	"time"
)

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor, cache cache.Cache) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store, taskDistributor, cache)
	require.NoError(t, err)

	return server
}

func newContextWithAdminBearerToken(t *testing.T, tokenMaker token.Maker, adminID int64, username string, roleID int64, duration time.Duration) context.Context {
	accessToken, _, err := tokenMaker.CreateAdminToken(adminID, username, roleID, duration)
	require.NoError(t, err)

	bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)
	md := metadata.MD{
		authorizationHeader: []string{
			bearerToken,
		},
	}

	return metadata.NewIncomingContext(context.Background(), md)
}
