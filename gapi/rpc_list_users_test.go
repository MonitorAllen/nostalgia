package gapi

import (
	"context"
	"testing"
	"time"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestListUsersNormalizesFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	store.EXPECT().
		ListAdminUsers(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, arg db.ListAdminUsersParams) ([]db.ListAdminUsersRow, error) {
			require.Equal(t, int32(20), arg.Limit)
			require.Equal(t, int32(0), arg.Offset)
			require.Equal(t, "all", arg.Status)
			require.True(t, arg.Q.Valid)
			require.Equal(t, "allen", arg.Q.String)
			return []db.ListAdminUsersRow{{ID: uuid.New(), Username: "allen", Role: util.Visitor}}, nil
		})
	store.EXPECT().
		CountAdminUsersByFilter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, arg db.CountAdminUsersByFilterParams) (int64, error) {
			require.Equal(t, "all", arg.Status)
			require.True(t, arg.Q.Valid)
			require.Equal(t, "allen", arg.Q.String)
			return 1, nil
		})

	resp, err := server.ListUsers(ctx, &pb.ListUsersRequest{
		Q:      " allen ",
		Status: "unknown",
		Page:   -1,
		Limit:  99,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), resp.GetCount())
	require.Len(t, resp.GetUsers(), 1)
	require.Equal(t, "allen", resp.GetUsers()[0].GetUsername())
}

func TestListUsersRejectsMissingAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)

	store.EXPECT().ListAdminUsers(gomock.Any(), gomock.Any()).Times(0)
	store.EXPECT().CountAdminUsersByFilter(gomock.Any(), gomock.Any()).Times(0)

	_, err := server.ListUsers(context.Background(), &pb.ListUsersRequest{})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, st.Code())
}
