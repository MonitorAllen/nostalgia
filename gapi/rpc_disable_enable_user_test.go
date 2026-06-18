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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDisableUserCallsTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)
	id := uuid.New()
	disabledAt := time.Now()

	store.EXPECT().
		DisableVisitorUserTx(gomock.Any(), db.DisableVisitorUserTxParams{
			ID:             id,
			DisabledReason: "spam",
		}).
		Times(1).
		Return(db.DisableVisitorUserTxResult{User: db.User{
			ID:             id,
			Username:       "visitor",
			Role:           util.Visitor,
			DisabledReason: "spam",
			DisabledAt:     pgtype.Timestamptz{Time: disabledAt, Valid: true},
		}}, nil)

	resp, err := server.DisableUser(ctx, &pb.DisableUserRequest{
		Id:     id.String(),
		Reason: " spam ",
	})

	require.NoError(t, err)
	require.Equal(t, id.String(), resp.GetUser().GetId())
	require.Equal(t, "spam", resp.GetUser().GetDisabledReason())
	require.NotNil(t, resp.GetUser().GetDisabledAt())
}

func TestDisableUserRejectsMissingAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)

	store.EXPECT().DisableVisitorUserTx(gomock.Any(), gomock.Any()).Times(0)

	_, err := server.DisableUser(context.Background(), &pb.DisableUserRequest{Id: uuid.NewString()})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, st.Code())
}

func TestDisableUserMapsNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	store.EXPECT().
		DisableVisitorUserTx(gomock.Any(), gomock.Any()).
		Times(1).
		Return(db.DisableVisitorUserTxResult{}, db.ErrRecordNotFound)

	_, err := server.DisableUser(ctx, &pb.DisableUserRequest{Id: uuid.NewString()})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestEnableUserCallsQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)
	id := uuid.New()

	store.EXPECT().
		EnableVisitorUser(gomock.Any(), id).
		Times(1).
		Return(db.User{ID: id, Username: "visitor", Role: util.Visitor}, nil)

	resp, err := server.EnableUser(ctx, &pb.EnableUserRequest{Id: id.String()})

	require.NoError(t, err)
	require.Equal(t, id.String(), resp.GetUser().GetId())
	require.Nil(t, resp.GetUser().GetDisabledAt())
}

func TestEnableUserRejectsMissingAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)

	store.EXPECT().EnableVisitorUser(gomock.Any(), gomock.Any()).Times(0)

	_, err := server.EnableUser(context.Background(), &pb.EnableUserRequest{Id: uuid.NewString()})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, st.Code())
}

func TestEnableUserMapsNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	store.EXPECT().
		EnableVisitorUser(gomock.Any(), gomock.Any()).
		Times(1).
		Return(db.User{}, db.ErrRecordNotFound)

	_, err := server.EnableUser(ctx, &pb.EnableUserRequest{Id: uuid.NewString()})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}
