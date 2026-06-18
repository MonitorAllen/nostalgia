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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateUserTrimsEditableFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)
	id := uuid.New()

	store.EXPECT().
		UpdateVisitorUser(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, arg db.UpdateVisitorUserParams) (db.User, error) {
			require.Equal(t, id, arg.ID)
			require.Equal(t, "Visitor Allen", arg.FullName)
			require.Equal(t, "visitor@example.com", arg.Email)
			require.True(t, arg.IsEmailVerified)
			return db.User{
				ID:              id,
				Username:        "visitor",
				FullName:        arg.FullName,
				Email:           arg.Email,
				IsEmailVerified: arg.IsEmailVerified,
				Role:            util.Visitor,
			}, nil
		})

	resp, err := server.UpdateUser(ctx, &pb.UpdateUserRequest{
		Id:              id.String(),
		FullName:        " Visitor Allen ",
		Email:           " visitor@example.com ",
		IsEmailVerified: true,
	})

	require.NoError(t, err)
	require.Equal(t, id.String(), resp.GetUser().GetId())
	require.Equal(t, "Visitor Allen", resp.GetUser().GetFullName())
	require.Equal(t, "visitor@example.com", resp.GetUser().GetEmail())
	require.True(t, resp.GetUser().GetIsEmailVerified())
}

func TestUpdateUserRejectsMissingAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)

	store.EXPECT().UpdateVisitorUser(gomock.Any(), gomock.Any()).Times(0)

	_, err := server.UpdateUser(context.Background(), &pb.UpdateUserRequest{Id: uuid.NewString()})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, st.Code())
}

func TestUpdateUserMapsDuplicateEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	store.EXPECT().
		UpdateVisitorUser(gomock.Any(), gomock.Any()).
		Times(1).
		Return(db.User{}, &pgconn.PgError{Code: db.UniqueViolation})

	_, err := server.UpdateUser(ctx, &pb.UpdateUserRequest{
		Id:              uuid.NewString(),
		FullName:        "Visitor",
		Email:           "visitor@example.com",
		IsEmailVerified: true,
	})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.AlreadyExists, st.Code())
}

func TestUpdateUserMapsNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	store.EXPECT().
		UpdateVisitorUser(gomock.Any(), gomock.Any()).
		Times(1).
		Return(db.User{}, db.ErrRecordNotFound)

	_, err := server.UpdateUser(ctx, &pb.UpdateUserRequest{
		Id:              uuid.NewString(),
		FullName:        "Visitor",
		Email:           "visitor@example.com",
		IsEmailVerified: true,
	})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}
