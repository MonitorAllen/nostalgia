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

// eqCreateCategoryTxParamsMatcher 自定义 Matcher
// 在参数匹配时执行 AfterCreate 回调
type eqCreateCategoryTxParamsMatcher struct {
	expectedName string
}

func (m eqCreateCategoryTxParamsMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(db.CreateCategoryTxParams)
	if !ok {
		return false
	}

	if actualArg.Name != m.expectedName {
		return false
	}

	// 执行 AfterCreate 回调（触发 taskDistributor 调用）
	return actualArg.AfterCreate() == nil
}

func (m eqCreateCategoryTxParamsMatcher) String() string {
	return "matches CreateCategoryTxParams and executes AfterCreate"
}

func EqCreateCategoryTxParams(name string) gomock.Matcher {
	return eqCreateCategoryTxParamsMatcher{expectedName: name}
}

func TestCreateCategory(t *testing.T) {
	admin := randomAdmin(t)

	category := randomCategory()

	testCases := []struct {
		name          string
		req           *pb.CreateCategoryRequest
		buildStubs    func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.CreateCategoryResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateCategoryRequest{
				Name: category.Name,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				// Mock TaskDistributor
				taskDistributor.EXPECT().
					DistributeTaskDelayDeleteCacheDefault(gomock.Any(), gomock.Eq(key.CategoryAllKey)).
					Times(1).
					Return(nil)

				// 使用自定义 Matcher 执行 AfterCreate 回调
				store.EXPECT().
					CreateCategoryTx(gomock.Any(), EqCreateCategoryTxParams(category.Name)).
					Times(1).
					Return(db.CreateCategoryTxResult{Category: category}, nil)
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
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				// 无需调用 store
				store.EXPECT().CreateCategoryTx(gomock.Any(), gomock.Any()).Times(0)
				taskDistributor.EXPECT().DistributeTaskDelayDeleteCacheDefault(gomock.Any(), gomock.Any()).Times(0)
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
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				taskDistributor.EXPECT().DistributeTaskDelayDeleteCacheDefault(gomock.Any(), gomock.Any()).Times(0)

				store.EXPECT().
					CreateCategoryTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateCategoryTxResult{}, db.ErrUniqueViolation)
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
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				taskDistributor.EXPECT().DistributeTaskDelayDeleteCacheDefault(gomock.Any(), gomock.Any()).Times(0)

				store.EXPECT().
					CreateCategoryTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateCategoryTxResult{}, sql.ErrConnDone)
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
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			taskCtrl := gomock.NewController(t)
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)

			tc.buildStubs(store, taskDistributor)

			server := newTestServer(t, store, taskDistributor, nil)

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
