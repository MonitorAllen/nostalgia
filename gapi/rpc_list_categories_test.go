package gapi

import (
	"context"
	"testing"
	"time"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestListCategoriesNormalizesPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	store.EXPECT().
		ListCategoriesCountArticles(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, arg db.ListCategoriesCountArticlesParams) ([]db.ListCategoriesCountArticlesRow, error) {
			require.Equal(t, int32(20), arg.Limit)
			require.Equal(t, int32(0), arg.Offset)
			return []db.ListCategoriesCountArticlesRow{{ID: 1, Name: "Go"}}, nil
		})
	store.EXPECT().
		CountCategories(gomock.Any()).
		Times(1).
		Return(int64(1), nil)

	resp, err := server.ListCategories(ctx, &pb.ListCategoriesRequest{Page: -1, Limit: 99})

	require.NoError(t, err)
	require.Equal(t, int64(1), resp.GetCount())
	require.Len(t, resp.GetCategories(), 1)
	require.Equal(t, "Go", resp.GetCategories()[0].GetName())
}

func TestListCategoriesRejectsMissingAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)

	store.EXPECT().ListCategoriesCountArticles(gomock.Any(), gomock.Any()).Times(0)
	store.EXPECT().CountCategories(gomock.Any()).Times(0)

	_, err := server.ListCategories(context.Background(), &pb.ListCategoriesRequest{})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, st.Code())
}
