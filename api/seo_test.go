package api

import (
	"database/sql"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	mockcache "github.com/MonitorAllen/nostalgia/internal/cache/mock"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestRobotsTxtAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	redisCache := mockcache.NewMockCache(ctrl)
	server := newTestServer(t, store, nil, redisCache)
	server.config.Domain = "https://blog.example.com/"
	recorder := httptest.NewRecorder()

	request, err := http.NewRequest(http.MethodGet, "/robots.txt", nil)
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
	body := recorder.Body.String()
	require.Contains(t, body, "User-agent: *")
	require.Contains(t, body, "Disallow: /backend")
	require.Contains(t, body, "Disallow: /api")
	require.Contains(t, body, "Disallow: /v1")
	require.Contains(t, body, "Sitemap: https://blog.example.com/sitemap.xml")
}

func TestSitemapXMLAPI(t *testing.T) {
	ownerID := uuid.New()
	categoryRows := []db.ListPublishedCategorySitemapItemsRow{
		{ID: 1, UpdatedAt: time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC)},
		{ID: 2, UpdatedAt: time.Date(2026, 6, 2, 8, 0, 0, 0, time.UTC)},
	}
	articleRows := []db.ListPublishedArticleSitemapItemsRow{
		{
			ID:        uuid.MustParse("3b0daee2-05da-4346-a3f4-ba68a463bb28"),
			Slug:      pgtype.Text{String: "redis-cache-consistency", Valid: true},
			Owner:     ownerID,
			CreatedAt: time.Date(2026, 6, 3, 8, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 6, 4, 8, 0, 0, 0, time.UTC),
		},
		{
			ID:        uuid.MustParse("4b0daee2-05da-4346-a3f4-ba68a463bb28"),
			Slug:      pgtype.Text{},
			Owner:     ownerID,
			CreatedAt: time.Date(2026, 6, 5, 8, 0, 0, 0, time.UTC),
			UpdatedAt: time.Time{},
		},
	}

	testCases := []struct {
		name          string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListPublishedCategorySitemapItems(gomock.Any()).
					Times(1).
					Return(categoryRows, nil)
				store.EXPECT().
					ListPublishedArticleSitemapItems(gomock.Any()).
					Times(1).
					Return(articleRows, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				require.Equal(t, "application/xml; charset=utf-8", recorder.Header().Get("Content-Type"))
				body := recorder.Body.String()
				require.Contains(t, body, "<loc>https://blog.example.com/</loc>")
				require.Contains(t, body, "<loc>https://blog.example.com/category/1</loc>")
				require.Contains(t, body, "<loc>https://blog.example.com/category/2</loc>")
				require.Contains(t, body, "<loc>https://blog.example.com/article/redis-cache-consistency</loc>")
				require.Contains(t, body, "<loc>https://blog.example.com/article/4b0daee2-05da-4346-a3f4-ba68a463bb28</loc>")
				require.Contains(t, body, "<lastmod>2026-06-04</lastmod>")
				require.Contains(t, body, "<lastmod>2026-06-05</lastmod>")

				var urlSet sitemapURLSet
				require.NoError(t, xml.Unmarshal(recorder.Body.Bytes(), &urlSet))
				require.Len(t, urlSet.URLs, 5)
			},
		},
		{
			name: "StoreError",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListPublishedCategorySitemapItems(gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)
				store.EXPECT().ListPublishedArticleSitemapItems(gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.True(t, strings.Contains(recorder.Body.String(), "sql: connection is already closed"))
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			redisCache := mockcache.NewMockCache(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, nil, redisCache)
			server.config = util.Config{Domain: "https://blog.example.com/"}
			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodGet, "/sitemap.xml", nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
