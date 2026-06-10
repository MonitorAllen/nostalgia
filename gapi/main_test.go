package gapi

import (
	"context"
	"fmt"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/internal/cache"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/MonitorAllen/nostalgia/worker"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"testing"
	"time"
)

type testStore struct {
	*mockdb.MockStore
}

func (store *testStore) CountAdminUsers(ctx context.Context) (int64, error) {
	return 0, nil
}

func (store *testStore) CreateUserWithRole(ctx context.Context, arg db.CreateUserWithRoleParams) (db.User, error) {
	return db.User{}, nil
}

func newGAPITestStore(store *mockdb.MockStore) db.Store {
	return &testStore{MockStore: store}
}

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor, cache cache.Cache) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store, taskDistributor, cache)
	require.NoError(t, err)

	return server
}

func newContextWithUserBearerToken(t *testing.T, tokenMaker token.Maker, userID uuid.UUID, username string, role string, duration time.Duration) context.Context {
	accessToken, _, err := tokenMaker.CreateToken(userID, username, role, duration)
	require.NoError(t, err)

	bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)
	md := metadata.MD{
		authorizationHeader: []string{
			bearerToken,
		},
	}

	return metadata.NewIncomingContext(context.Background(), md)
}

func newContextWithAdminBearerToken(t *testing.T, tokenMaker token.Maker, duration time.Duration) context.Context {
	return newContextWithUserBearerToken(t, tokenMaker, util.RandUserID(), util.RandomOwner(), util.Admin, duration)
}
