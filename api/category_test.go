package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestListCategoriesAPIPaginates(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().
		ListCategoriesCountArticles(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, arg db.ListCategoriesCountArticlesParams) ([]db.ListCategoriesCountArticlesRow, error) {
			require.Equal(t, int32(10), arg.Limit)
			require.Equal(t, int32(10), arg.Offset)
			return []db.ListCategoriesCountArticlesRow{{ID: 1, Name: "Go"}}, nil
		})
	store.EXPECT().
		CountCategories(gomock.Any()).
		Times(1).
		Return(int64(21), nil)

	server := newTestServer(t, store, nil, nil)
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/api/categories?page=2&limit=10", nil)
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body listCategoriesResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, int64(21), body.Count)
	require.Len(t, body.Categories, 1)
	require.Equal(t, "Go", body.Categories[0].Name)
}

func TestListCategoriesAPIRejectsInvalidPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().ListCategoriesCountArticles(gomock.Any(), gomock.Any()).Times(0)
	store.EXPECT().CountCategories(gomock.Any()).Times(0)

	server := newTestServer(t, store, nil, nil)
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/api/categories?page=0&limit=10", nil)
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
