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

func TestCreateCategory(t *testing.T) {
	admin := randomAdmin(t)

	category := randomCategory()

	testCases := []struct {
		name          string
		req           *pb.CreateCategoryRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.CreateCategoryResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateCategoryRequest{
				Name: category.Name,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateCategory(gomock.Any(), gomock.Eq(category.Name)).
					Times(1).
					Return(category, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAdminBearerToken(t, tokenMaker, admin.ID, admin.Username, admin.RoleID, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateCategoryResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				createCategory := res.GetCategory()
				require.NotEmpty(t, createCategory.Id)
				require.Equal(t, category.Name, createCategory.Name)
			},
		},
		{
			name: "Unauthenticated",
			req: &pb.CreateCategoryRequest{
				Name: category.Name,
			},
			buildStubs: func(store *mockdb.MockStore) {
				// 无需调用 store
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background() // 未带 token
			},
			checkResponse: func(t *testing.T, res *pb.CreateCategoryResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "DuplicateName",
			req: &pb.CreateCategoryRequest{
				Name: category.Name,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateCategory(gomock.Any(), gomock.Eq(category.Name)).
					Times(1).
					Return(db.Category{}, db.ErrUniqueViolation)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAdminBearerToken(t, tokenMaker, admin.ID, admin.Username, admin.RoleID, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateCategoryResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.AlreadyExists, st.Code())
			},
		},
		{
			name: "DBError",
			req: &pb.CreateCategoryRequest{
				Name: category.Name,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateCategory(gomock.Any(), gomock.Eq(category.Name)).
					Times(1).
					Return(db.Category{}, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAdminBearerToken(t, tokenMaker, admin.ID, admin.Username, admin.RoleID, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateCategoryResponse, err error) {
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

			res, err := server.CreateCategory(ctx, tc.req)

			tc.checkResponse(t, res, err)
		})
	}
}

func randomCategory() db.Category {
	return db.Category{
		ID:        util.RandomInt(1, 1000),
		Name:      util.RandomString(6),
		IsSystem:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
