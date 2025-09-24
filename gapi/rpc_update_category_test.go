package gapi

import (
	"context"
	"database/sql"
	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
	"time"
)

func TestUpdateCategory(t *testing.T) {
	admin := randomAdmin(t)

	category := randomCategory()

	newName := util.RandomString(6)

	testCases := []struct {
		name          string
		req           *pb.UpdateCategoryRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.UpdateCategoryResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.UpdateCategoryRequest{
				Id:   category.ID,
				Name: newName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateCategoryParams{
					ID:   category.ID,
					Name: newName,
				}
				updateCategory := db.Category{
					ID:        category.ID,
					Name:      newName,
					IsSystem:  category.IsSystem,
					CreatedAt: category.CreatedAt,
					UpdatedAt: category.UpdatedAt,
				}
				store.EXPECT().
					UpdateCategory(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(updateCategory, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAdminBearerToken(t, tokenMaker, admin.ID, admin.Username, admin.RoleID, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateCategoryResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				updateCategory := res.GetCategory()
				require.Equal(t, newName, updateCategory.Name)
			},
		},
		{
			name: "NoAuthorization",
			req: &pb.UpdateCategoryRequest{
				Name: category.Name,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateCategory(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background() // 未带 token
			},
			checkResponse: func(t *testing.T, res *pb.UpdateCategoryResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "DuplicateName",
			req: &pb.UpdateCategoryRequest{
				Id:   category.ID,
				Name: category.Name,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateCategory(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Category{}, db.ErrUniqueViolation)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAdminBearerToken(t, tokenMaker, admin.ID, admin.Username, admin.RoleID, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateCategoryResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.AlreadyExists, st.Code())
			},
		},
		{
			name: "NotFound",
			req: &pb.UpdateCategoryRequest{
				Id:   category.ID,
				Name: category.Name,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateCategory(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Category{}, db.ErrRecordNotFound)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAdminBearerToken(t, tokenMaker, admin.ID, admin.Username, admin.RoleID, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateCategoryResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "DBError",
			req: &pb.UpdateCategoryRequest{
				Name: category.Name,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateCategory(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Category{}, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAdminBearerToken(t, tokenMaker, admin.ID, admin.Username, admin.RoleID, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateCategoryResponse, err error) {
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

			res, err := server.UpdateCategory(ctx, tc.req)

			tc.checkResponse(t, res, err)
		})
	}
}
