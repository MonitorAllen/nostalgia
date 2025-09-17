package gapi

import (
	"context"
	"database/sql"
	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
	"time"
)

func TestDeleteCategory(t *testing.T) {
	admin := randomAdmin(t)

	category := randomCategory()

	testCases := []struct {
		name          string
		req           *pb.DeleteCategoryRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.DeleteCategoryResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.DeleteCategoryRequest{
				Id: category.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteCategoryTx(gomock.Any(), gomock.Eq(category.ID)).
					Times(1).
					Return(nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAdminBearerToken(t, tokenMaker, admin.ID, admin.Username, admin.RoleID, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.DeleteCategoryResponse, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "ExpiredToken",
			req: &pb.DeleteCategoryRequest{
				Id: category.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteCategoryTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAdminBearerToken(t, tokenMaker, admin.ID, admin.Username, admin.RoleID, -time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.DeleteCategoryResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "NoAuthorization",
			req: &pb.DeleteCategoryRequest{
				Id: category.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteCategoryTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background() // 未带 token
			},
			checkResponse: func(t *testing.T, res *pb.DeleteCategoryResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "DBError",
			req: &pb.DeleteCategoryRequest{
				Id: category.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteCategoryTx(gomock.Any(), gomock.Eq(category.ID)).
					Times(1).
					Return(sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAdminBearerToken(t, tokenMaker, admin.ID, admin.Username, admin.RoleID, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.DeleteCategoryResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			store := mockdb.NewMockStore(ctrl)

			tc.buildStubs(store)

			server := newTestServer(t, store, nil, nil)

			ctx := tc.buildContext(t, server.tokenMaker)

			res, err := server.DeleteCategory(ctx, tc.req)

			tc.checkResponse(t, res, err)
		})
	}
}
