package gapi

import (
	"context"
	"database/sql"
	"testing"
	"time"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/MonitorAllen/nostalgia/util"
	mockwk "github.com/MonitorAllen/nostalgia/worker/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// eqUpdateCategoryTxParamsMatcher 自定义 Matcher
type eqUpdateCategoryTxParamsMatcher struct {
	expectedID   int64
	expectedName string
}

func (m eqUpdateCategoryTxParamsMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(db.UpdateCategoryTxParams)
	if !ok {
		return false
	}

	// 验证参数
	if actualArg.UpdateCategoryParams.ID != m.expectedID || actualArg.UpdateCategoryParams.Name != m.expectedName {
		return false
	}

	// 执行 AfterUpdate 回调（触发 taskDistributor 调用）
	return actualArg.AfterUpdate() == nil
}

func (m eqUpdateCategoryTxParamsMatcher) String() string {
	return "matches UpdateCategoryTxParams and executes AfterUpdate"
}

func EqUpdateCategoryTxParams(id int64, name string) gomock.Matcher {
	return eqUpdateCategoryTxParamsMatcher{expectedID: id, expectedName: name}
}

func TestUpdateCategory(t *testing.T) {
	admin := randomAdmin(t)

	category := randomCategory()

	newName := util.RandomString(6)

	testCases := []struct {
		name          string
		req           *pb.UpdateCategoryRequest
		buildStubs    func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.UpdateCategoryResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.UpdateCategoryRequest{
				Id:   category.ID,
				Name: newName,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				updateCategory := db.Category{
					ID:        category.ID,
					Name:      newName,
					IsSystem:  category.IsSystem,
					CreatedAt: category.CreatedAt,
					UpdatedAt: category.UpdatedAt,
				}

				// Mock TaskDistributor（会在 Matcher 执行 AfterUpdate 时被调用）
				taskDistributor.EXPECT().
					DistributeTaskDelayDeleteCacheDefault(gomock.Any(), gomock.Eq(key.CategoryAllKey)).
					Times(1).
					Return(nil)

				// 使用自定义 Matcher 执行 AfterUpdate 回调
				store.EXPECT().
					UpdateCategoryTx(gomock.Any(), EqUpdateCategoryTxParams(category.ID, newName)).
					Times(1).
					Return(db.UpdateCategoryTxResult{Category: updateCategory}, nil)
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
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					UpdateCategoryTx(gomock.Any(), gomock.Any()).
					Times(0)

				taskDistributor.EXPECT().
					DistributeTaskDelayDeleteCacheDefault(gomock.Any(), gomock.Any()).
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
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				taskDistributor.EXPECT().
					DistributeTaskDelayDeleteCacheDefault(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					UpdateCategoryTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UpdateCategoryTxResult{}, db.ErrUniqueViolation)
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
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				taskDistributor.EXPECT().
					DistributeTaskDelayDeleteCacheDefault(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					UpdateCategoryTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UpdateCategoryTxResult{}, db.ErrRecordNotFound)
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
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				taskDistributor.EXPECT().
					DistributeTaskDelayDeleteCacheDefault(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					UpdateCategoryTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UpdateCategoryTxResult{}, sql.ErrConnDone)
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
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			taskCtrl := gomock.NewController(t)
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)

			tc.buildStubs(store, taskDistributor)

			server := newTestServer(t, store, taskDistributor, nil)

			ctx := tc.buildContext(t, server.tokenMaker)

			res, err := server.UpdateCategory(ctx, tc.req)

			tc.checkResponse(t, res, err)
		})
	}
}
