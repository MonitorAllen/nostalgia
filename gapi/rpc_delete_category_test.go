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
	mockwk "github.com/MonitorAllen/nostalgia/worker/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// eqDeleteCategoryTxParamsMatcher 自定义 Matcher
// 在参数匹配时执行 AfterDelete 回调
type eqDeleteCategoryTxParamsMatcher struct {
	expectedID int64
}

func (m eqDeleteCategoryTxParamsMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(db.DeleteCategoryTxParams)
	if !ok {
		return false
	}

	if actualArg.ID != m.expectedID {
		return false
	}

	// 执行 AfterDelete 回调（触发 taskDistributor 调用）
	return actualArg.AfterDelete() == nil
}

func (m eqDeleteCategoryTxParamsMatcher) String() string {
	return "matches DeleteCategoryTxParams and executes AfterDelete"
}

func EqDeleteCategoryTxParams(id int64) gomock.Matcher {
	return eqDeleteCategoryTxParamsMatcher{expectedID: id}
}

func TestDeleteCategory(t *testing.T) {
	admin := randomAdmin(t)

	category := randomCategory()

	testCases := []struct {
		name          string
		req           *pb.DeleteCategoryRequest
		buildStubs    func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.DeleteCategoryResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.DeleteCategoryRequest{
				Id: category.ID,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				// Mock TaskDistributor（会在 Matcher 执行 AfterDelete 时被调用）
				taskDistributor.EXPECT().
					DistributeTaskDelayDeleteCacheDefault(gomock.Any(), gomock.Eq(key.CategoryAllKey)).
					Times(1).
					Return(nil)

				// 使用自定义 Matcher 执行 AfterDelete 回调
				store.EXPECT().
					DeleteCategoryTx(gomock.Any(), EqDeleteCategoryTxParams(category.ID)).
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
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					DeleteCategoryTx(gomock.Any(), gomock.Any()).
					Times(0)

				taskDistributor.EXPECT().
					DistributeTaskDelayDeleteCacheDefault(gomock.Any(), gomock.Any()).
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
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					DeleteCategoryTx(gomock.Any(), gomock.Any()).
					Times(0)

				taskDistributor.EXPECT().
					DistributeTaskDelayDeleteCacheDefault(gomock.Any(), gomock.Any()).
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
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				taskDistributor.EXPECT().
					DistributeTaskDelayDeleteCacheDefault(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					DeleteCategoryTx(gomock.Any(), gomock.Any()).
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
			storeCtrl := gomock.NewController(t)
			t.Cleanup(storeCtrl.Finish)
			store := mockdb.NewMockStore(storeCtrl)

			taskCtrl := gomock.NewController(t)
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)

			tc.buildStubs(store, taskDistributor)

			server := newTestServer(t, store, taskDistributor, nil)

			ctx := tc.buildContext(t, server.tokenMaker)

			res, err := server.DeleteCategory(ctx, tc.req)

			tc.checkResponse(t, res, err)
		})
	}
}
