package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type eqCreateUserWithRoleParamsMatcher struct {
	arg      db.CreateUserWithRoleParams
	password string
}

func (expected eqCreateUserWithRoleParamsMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(db.CreateUserWithRoleParams)
	if !ok {
		return false
	}

	if err := util.CheckPassword(expected.password, actualArg.HashedPassword); err != nil {
		return false
	}
	if actualArg.ID == uuid.Nil {
		return false
	}

	expected.arg.ID = actualArg.ID
	expected.arg.HashedPassword = actualArg.HashedPassword
	return reflect.DeepEqual(expected.arg, actualArg)
}

func (expected eqCreateUserWithRoleParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", expected.arg, expected.password)
}

func EqCreateUserWithRoleParams(arg db.CreateUserWithRoleParams, password string) gomock.Matcher {
	return eqCreateUserWithRoleParamsMatcher{arg: arg, password: password}
}

func TestSetupStatusAPI(t *testing.T) {
	testCases := []struct {
		name          string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "NoAdmin",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CountAdminUsers(gomock.Any()).
					Times(1).
					Return(int64(0), nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireSetupStatusBody(t, recorder.Body, false, true)
			},
		},
		{
			name: "ExistingAdmin",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CountAdminUsers(gomock.Any()).
					Times(1).
					Return(int64(1), nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireSetupStatusBody(t, recorder.Body, true, false)
			},
		},
		{
			name: "InternalError",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CountAdminUsers(gomock.Any()).
					Times(1).
					Return(int64(0), sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newSetupTestServer(t, store, "setup-token")
			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodGet, "/api/setup/status", nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateSetupAdminAPI(t *testing.T) {
	admin, password := randomUser(t)
	admin.Role = util.Admin
	admin.IsEmailVerified = true
	setupToken := "correct-setup-token"

	testCases := []struct {
		name          string
		setupToken    string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			setupToken: setupToken,
			body: gin.H{
				"setup_token": setupToken,
				"username":    admin.Username,
				"password":    password,
				"full_name":   admin.FullName,
				"email":       admin.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CountAdminUsers(gomock.Any()).
					Times(1).
					Return(int64(0), nil)

				arg := db.CreateUserWithRoleParams{
					Username:        admin.Username,
					FullName:        admin.FullName,
					Email:           admin.Email,
					IsEmailVerified: true,
					Role:            util.Admin,
				}
				store.EXPECT().
					CreateUserWithRole(gomock.Any(), EqCreateUserWithRoleParams(arg, password)).
					Times(1).
					Return(admin, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireSetupAdminBody(t, recorder.Body, admin, setupToken)
			},
		},
		{
			name:       "WrongSetupToken",
			setupToken: setupToken,
			body: gin.H{
				"setup_token": "wrong-setup-token",
				"username":    admin.Username,
				"password":    password,
				"full_name":   admin.FullName,
				"email":       admin.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CountAdminUsers(gomock.Any()).
					Times(1).
					Return(int64(0), nil)
				store.EXPECT().
					CreateUserWithRole(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				require.NotContains(t, recorder.Body.String(), "wrong-setup-token")
			},
		},
		{
			name:       "ExistingAdmin",
			setupToken: setupToken,
			body: gin.H{
				"setup_token": setupToken,
				"username":    admin.Username,
				"password":    password,
				"full_name":   admin.FullName,
				"email":       admin.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CountAdminUsers(gomock.Any()).
					Times(1).
					Return(int64(1), nil)
				store.EXPECT().
					CreateUserWithRole(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusConflict, recorder.Code)
			},
		},
		{
			name:       "MissingConfiguredSetupToken",
			setupToken: "",
			body: gin.H{
				"setup_token": "request-token",
				"username":    admin.Username,
				"password":    password,
				"full_name":   admin.FullName,
				"email":       admin.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CountAdminUsers(gomock.Any()).
					Times(1).
					Return(int64(0), nil)
				store.EXPECT().
					CreateUserWithRole(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.NotContains(t, recorder.Body.String(), "request-token")
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newSetupTestServer(t, store, tc.setupToken)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/api/setup/admin", bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func newSetupTestServer(t *testing.T, store db.Store, setupToken string) *Server {
	config := util.Config{
		Environment:           "development",
		TokenSymmetricKey:     util.RandomString(32),
		AccessTokenDuration:   time.Minute,
		RefreshTokenDuration:  time.Hour,
		SetupToken:            setupToken,
		UploadFileSizeLimit:   1024,
		UploadFileAllowedMime: []string{"image/png"},
	}

	server, err := NewServer(config, store, nil, nil)
	require.NoError(t, err)

	return server
}

func requireSetupStatusBody(t *testing.T, body *bytes.Buffer, initialized bool, setupAvailable bool) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var got map[string]bool
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)
	require.Equal(t, initialized, got["initialized"])
	require.Equal(t, setupAvailable, got["setup_available"])
	require.NotContains(t, string(data), "setup_token")
}

func requireSetupAdminBody(t *testing.T, body *bytes.Buffer, user db.User, setupToken string) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	require.NotContains(t, string(data), setupToken)
	require.NotContains(t, string(data), user.HashedPassword)

	var gotUser userResponse
	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)
	require.Equal(t, user.ID, gotUser.ID)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.True(t, gotUser.IsEmailVerified)
	require.Equal(t, util.Admin, gotUser.Role)
}
