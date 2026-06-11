package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	mockcache "github.com/MonitorAllen/nostalgia/internal/cache/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestHealthzAPI(t *testing.T) {
	server := newTestServer(t, nil, nil, nil)
	recorder := httptest.NewRecorder()

	request, err := http.NewRequest(http.MethodGet, "/healthz", nil)
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	requireHealthBody(t, recorder.Body.Bytes(), "ok")
}

func TestReadyzAPI(t *testing.T) {
	testCases := []struct {
		name          string
		buildStubs    func(store *mockdb.MockStore, redisCache *mockcache.MockCache)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Ready",
			buildStubs: func(store *mockdb.MockStore, redisCache *mockcache.MockCache) {
				store.EXPECT().Ping(gomock.Any()).Times(1).Return(nil)
				redisCache.EXPECT().Ping(gomock.Any()).Times(1).Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireReadyBody(t, recorder.Body.Bytes(), "ready", "ok", "ok")
			},
		},
		{
			name: "DatabaseUnavailable",
			buildStubs: func(store *mockdb.MockStore, redisCache *mockcache.MockCache) {
				store.EXPECT().Ping(gomock.Any()).Times(1).Return(sql.ErrConnDone)
				redisCache.EXPECT().Ping(gomock.Any()).Times(1).Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
				requireReadyBody(t, recorder.Body.Bytes(), "not_ready", "unavailable", "ok")
			},
		},
		{
			name: "RedisUnavailable",
			buildStubs: func(store *mockdb.MockStore, redisCache *mockcache.MockCache) {
				store.EXPECT().Ping(gomock.Any()).Times(1).Return(nil)
				redisCache.EXPECT().Ping(gomock.Any()).Times(1).Return(errors.New("redis unavailable"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
				requireReadyBody(t, recorder.Body.Bytes(), "not_ready", "ok", "unavailable")
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			cacheCtrl := gomock.NewController(t)
			defer cacheCtrl.Finish()

			store := mockdb.NewMockStore(storeCtrl)
			redisCache := mockcache.NewMockCache(cacheCtrl)
			tc.buildStubs(store, redisCache)

			server := newTestServer(t, store, nil, redisCache)
			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodGet, "/readyz", nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireHealthBody(t *testing.T, body []byte, status string) {
	var payload struct {
		Status string `json:"status"`
	}
	err := json.Unmarshal(body, &payload)
	require.NoError(t, err)
	require.Equal(t, status, payload.Status)
}

func requireReadyBody(t *testing.T, body []byte, status string, database string, redis string) {
	var payload struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	err := json.Unmarshal(body, &payload)
	require.NoError(t, err)
	require.Equal(t, status, payload.Status)
	require.Equal(t, database, payload.Checks["database"])
	require.Equal(t, redis, payload.Checks["redis"])
}
