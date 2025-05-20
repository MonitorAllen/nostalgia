package gapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	mockservice "github.com/MonitorAllen/nostalgia/internal/service/mock"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"testing"
	"time"
)

func TestAdminInfo(t *testing.T) {
	admin := randomAdmin(t)
	payload := token.AdminPayload{
		ID:       uuid.New(),
		AdminID:  admin.ID,
		Username: admin.Username,
		RoleID:   admin.RoleID,
		IssuedAt: time.Now(),
		ExpireAt: time.Now().Add(time.Hour),
	}

	adminSession := AdminSession{
		Payload: &payload,
	}

	sessionBytes, err := json.Marshal(adminSession)
	require.NoError(t, err)

	testCases := []struct {
		name          string
		buildStubs    func(store *mockdb.MockStore, redisService *mockservice.MockRedis)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.AdminInfoResponse, err error)
	}{
		{
			name: "OK",
			buildStubs: func(store *mockdb.MockStore, redisService *mockservice.MockRedis) {
				key := adminSessionKey + strconv.FormatInt(payload.AdminID, 10)
				redisService.EXPECT().
					Get(key).
					Times(1).
					Return(string(sessionBytes), nil)

				store.EXPECT().
					GetAdmin(gomock.Any(), gomock.Eq(admin.Username)).
					Times(1).
					Return(admin, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAdminBearerToken(t, tokenMaker, admin.ID, admin.Username, admin.RoleID, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.AdminInfoResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				gotAdmin := res.GetAdmin()
				require.Equal(t, admin.Username, gotAdmin.Username)
				require.Equal(t, admin.IsActive, gotAdmin.IsActive)
			},
		},
		{
			name: "Unauthenticated",
			buildStubs: func(store *mockdb.MockStore, redisService *mockservice.MockRedis) {
				// 无需调用 redis/store
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background() // 未带 token
			},
			checkResponse: func(t *testing.T, res *pb.AdminInfoResponse, err error) {
				require.Error(t, err)
				require.Equal(t, codes.Unauthenticated, status.Code(err))
				require.Nil(t, res)
			},
		},
		{
			name: "RedisNotFound",
			buildStubs: func(store *mockdb.MockStore, redisService *mockservice.MockRedis) {
				redisService.EXPECT().
					Get(gomock.Any()).
					Times(1).
					Return("", errors.New("redis key not found"))
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAdminBearerToken(t, tokenMaker, admin.ID, admin.Username, admin.RoleID, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.AdminInfoResponse, err error) {
				require.Error(t, err)
				require.Equal(t, codes.Unauthenticated, status.Code(err))
				require.Nil(t, res)
			},
		},
		{
			name: "RedisUnmarshalError",
			buildStubs: func(store *mockdb.MockStore, redisService *mockservice.MockRedis) {
				redisService.EXPECT().
					Get(gomock.Any()).
					Times(1).
					Return("not-json", nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAdminBearerToken(t, tokenMaker, admin.ID, admin.Username, admin.RoleID, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.AdminInfoResponse, err error) {
				require.Error(t, err)
				require.Equal(t, codes.Internal, status.Code(err))
				require.Nil(t, res)
			},
		},
		{
			name: "DBError",
			buildStubs: func(store *mockdb.MockStore, redisService *mockservice.MockRedis) {
				redisService.EXPECT().
					Get(gomock.Any()).
					Times(1).
					Return(string(sessionBytes), nil)

				store.EXPECT().
					GetAdmin(gomock.Any(), gomock.Eq(admin.Username)).
					Times(1).
					Return(db.Admin{}, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithAdminBearerToken(t, tokenMaker, admin.ID, admin.Username, admin.RoleID, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.AdminInfoResponse, err error) {
				require.Error(t, err)
				require.Equal(t, codes.Internal, status.Code(err))
				require.Nil(t, res)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			store := mockdb.NewMockStore(ctrl)
			redisService := mockservice.NewMockRedis(ctrl)

			tc.buildStubs(store, redisService)

			server := newTestServer(t, store, nil, redisService)

			ctx := tc.buildContext(t, server.tokenMaker)

			res, err := server.AdminInfo(ctx, &pb.AdminInfoRequest{})

			tc.checkResponse(t, res, err)
		})
	}
}

func randomAdmin(t *testing.T) db.Admin {
	hashPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	admin := db.Admin{
		ID:             util.RandomInt(1, 1000),
		Username:       util.RandomOwner(),
		HashedPassword: hashPassword,
		IsActive:       true,
		RoleID:         util.RandomInt(1, 2),
	}

	return admin
}
