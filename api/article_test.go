package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

type eqCreateArticleParamsMatcher struct {
	arg  db.CreateArticleParams
	user db.User
}

func (expected eqCreateArticleParamsMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(db.CreateArticleParams)
	if !ok {
		return false
	}

	expected.arg.ID = actualArg.ID
	if !reflect.DeepEqual(expected.arg, actualArg) {
		return false
	}

	return true
}

func (expected eqCreateArticleParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v", expected.arg)
}

func EqCreateArticleParams(arg db.CreateArticleParams, user db.User) gomock.Matcher {
	return eqCreateArticleParamsMatcher{arg, user}
}

func TestCreateArticleAPI(t *testing.T) {
	user, _ := randomUser(t)

	article := randomArticle(t, user.ID)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"title":      article.Title,
				"summary":    article.Summary,
				"content":    article.Content,
				"is_publish": article.IsPublish,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, user.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateArticleParams{
					Title:     article.Title,
					Summary:   article.Summary,
					Content:   article.Content,
					IsPublish: article.IsPublish,
					Owner:     article.Owner,
				}
				store.EXPECT().
					CreateArticle(gomock.Any(), EqCreateArticleParams(arg, user)).
					Times(1).
					Return(article, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchArticle(t, recorder.Body, article)
			},
		},
		{
			name: "BadRequest",
			body: gin.H{
				"summary":    article.Summary,
				"content":    article.Content,
				"is_publish": article.IsPublish,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, user.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateArticle(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"title":      article.Title,
				"summary":    article.Summary,
				"content":    article.Content,
				"is_publish": article.IsPublish,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, user.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateArticle(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Article{}, sql.ErrConnDone)
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

			// start test server and send request
			server := newTestServer(t, store, nil, nil)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/articles"

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetArticleAPI(t *testing.T) {
	user, _ := randomUser(t)

	article := randomGetArticleRow(t, user.ID, true)
	unpublishedArticle := randomGetArticleRow(t, user.ID, false)

	testCases := []struct {
		name          string
		articleID     string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			articleID: article.ID.String(),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetArticle(gomock.Any(), gomock.Eq(article.ID)).
					Times(1).
					Return(article, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchGetArticleRow(t, recorder.Body, article)
			},
		},
		{
			name:      "BadRequest",
			articleID: "not-uuid",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetArticle(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			articleID: article.ID.String(),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetArticle(gomock.Any(), gomock.Eq(article.ID)).
					Times(1).
					Return(db.GetArticleRow{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			articleID: article.ID.String(),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetArticle(gomock.Any(), gomock.Eq(article.ID)).
					Times(1).
					Return(db.GetArticleRow{}, db.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "Forbidden",
			articleID: unpublishedArticle.ID.String(),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetArticle(gomock.Any(), gomock.Eq(unpublishedArticle.ID)).
					Times(1).
					Return(unpublishedArticle, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
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

			// start test server and send request
			server := newTestServer(t, store, nil, nil)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/articles/%s", tc.articleID)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListArticleAPI(t *testing.T) {
	user, _ := randomUser(t)

	n := 5
	listArticlesRows := make([]db.ListArticlesRow, n)
	for i := 0; i < n; i++ {
		article := randomArticle(t, user.ID)
		listArticlesRows[i] = db.ListArticlesRow{
			ID:        article.ID,
			Title:     article.Title,
			Summary:   article.Summary,
			Views:     article.Views,
			Likes:     article.Likes,
			IsPublish: article.IsPublish,
			Owner:     article.Owner,
			CreatedAt: article.CreatedAt,
			UpdatedAt: article.UpdatedAt,
			DeletedAt: article.DeletedAt,
			Username: pgtype.Text{
				String: user.Username,
				Valid:  true,
			},
		}
	}

	type Query struct {
		page  int
		limit int
	}

	testCases := []struct {
		name          string
		query         Query
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				page:  1,
				limit: n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListArticlesParams{
					Limit:  int32(n),
					Offset: 0,
					IsPublish: pgtype.Bool{
						Bool:  true,
						Valid: true,
					},
				}
				store.EXPECT().
					ListArticles(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(listArticlesRows, nil)
				store.EXPECT().
					CountArticles(gomock.Any(), gomock.Any()).
					Times(1).
					Return(int64(n), nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchArticles(t, recorder.Body, listArticlesRows)
			},
		},
		{
			name: "BadRequest",
			query: Query{
				page:  0,
				limit: 40,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListArticles(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			query: Query{
				page:  1,
				limit: n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListArticles(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListArticlesRow{}, sql.ErrConnDone)
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

			// start test server and send request
			server := newTestServer(t, store, nil, nil)
			recorder := httptest.NewRecorder()

			url := "/api/articles"

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page", fmt.Sprintf("%d", tc.query.page))
			q.Add("limit", fmt.Sprintf("%d", tc.query.limit))
			request.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(t, recorder)
		})
	}
}

func TestSearchArticleAPI(t *testing.T) {
	user, _ := randomUser(t)
	n := 3
	searchArticlesRows := make([]db.SearchArticlesRow, n)
	for i := 0; i < n; i++ {
		article := randomArticle(t, user.ID)
		searchArticlesRows[i] = db.SearchArticlesRow{
			ID:        article.ID,
			Title:     article.Title,
			Summary:   article.Summary,
			Views:     article.Views,
			Likes:     article.Likes,
			IsPublish: article.IsPublish,
			Owner:     article.Owner,
			CreatedAt: article.CreatedAt,
			UpdatedAt: article.UpdatedAt,
			DeletedAt: article.DeletedAt,
			Username: pgtype.Text{
				String: user.Username,
				Valid:  true,
			},
		}
	}

	testCases := []struct {
		name          string
		req           searchArticlesRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			req: searchArticlesRequest{
				Keyword: "Go",
				Page:    1,
				Limit:   10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.SearchArticlesParams{
					Limit:   10,
					Offset:  0,
					Keyword: "go",
					IsPublish: pgtype.Bool{
						Bool:  true,
						Valid: true,
					},
				}
				store.EXPECT().SearchArticles(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(searchArticlesRows, nil)

				countArg := db.CountSearchArticlesParams{
					Keyword:   "go",
					IsPublish: pgtype.Bool{Bool: true, Valid: true},
				}
				store.EXPECT().CountSearchArticles(gomock.Any(), gomock.Eq(countArg)).
					Times(1).
					Return(int64(n), nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchSearchArticles(t, recorder.Body, searchArticlesRows)
			},
		},
		{
			name: "OK_EmptyResult",
			req: searchArticlesRequest{
				Keyword: "java",
				Page:    1,
				Limit:   10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				// 模拟返回空列表
				store.EXPECT().
					SearchArticles(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.SearchArticlesRow{}, nil)

				store.EXPECT().
					CountSearchArticles(gomock.Any(), gomock.Any()).
					Times(1).
					Return(int64(0), nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				// 验证返回的是空数组而不是 null，且 count 为 0
				data := recorder.Body.Bytes()
				var resp searchArticlesResponse
				err := json.Unmarshal(data, &resp)
				require.NoError(t, err)
				require.Equal(t, int64(0), resp.Count)
				require.Empty(t, resp.Articles)
			},
		},
		{
			name: "OK_WithSegmentation",
			req: searchArticlesRequest{
				Keyword: "Go 并发",
				Page:    1,
				Limit:   10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				// 修改这里：期望 "go OR 并发"
				arg := db.SearchArticlesParams{
					Limit:     10,
					Offset:    0,
					Keyword:   "go OR 并发", // <--- Go 改成小写
					IsPublish: pgtype.Bool{Bool: true, Valid: true},
				}

				store.EXPECT().
					SearchArticles(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(searchArticlesRows, nil)

				countArg := db.CountSearchArticlesParams{
					Keyword:   "go OR 并发", // <--- Go 改成小写
					IsPublish: pgtype.Bool{Bool: true, Valid: true},
				}
				store.EXPECT().
					CountSearchArticles(gomock.Any(), gomock.Eq(countArg)).
					Times(1).
					Return(int64(n), nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			req: searchArticlesRequest{
				Keyword: "Crash",
				Page:    1,
				Limit:   10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					SearchArticles(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.SearchArticlesRow{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidPageParam",
			req: searchArticlesRequest{
				Keyword: "Go",
				Page:    0, // 非法页码
				Limit:   10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().SearchArticles(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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

			testServer := newTestServer(t, store, nil, nil)
			recorder := httptest.NewRecorder()

			url := "/api/articles/search"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			q := request.URL.Query()
			q.Add("keyword", tc.req.Keyword)
			q.Add("page", fmt.Sprintf("%d", tc.req.Page))
			q.Add("limit", fmt.Sprintf("%d", tc.req.Limit))
			request.URL.RawQuery = q.Encode()

			testServer.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchSearchArticles(t *testing.T, body *bytes.Buffer, expectedRows []db.SearchArticlesRow) {
	data := body.Bytes()
	var resp searchArticlesResponse
	err := json.Unmarshal(data, &resp)
	require.NoError(t, err)

	require.Equal(t, int64(len(expectedRows)), resp.Count)
	require.Equal(t, len(expectedRows), len(resp.Articles))

	if len(expectedRows) > 0 {
		require.Equal(t, expectedRows[0].ID, resp.Articles[0].ID)
		require.Equal(t, expectedRows[0].Title, resp.Articles[0].Title)
	}
}

func requireBodyMatchArticles(t *testing.T, body *bytes.Buffer, listArticlesRow []db.ListArticlesRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotArticles listArticleResponse
	err = json.Unmarshal(data, &gotArticles)
	require.NoError(t, err)
	require.Equal(t, listArticlesRow, gotArticles.Articles)
}

func requireBodyMatchArticle(t *testing.T, body *bytes.Buffer, article db.Article) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotArticle db.Article
	err = json.Unmarshal(data, &gotArticle)
	require.NoError(t, err)

	require.Equal(t, article.ID.String(), gotArticle.ID.String())
	require.Equal(t, article.Title, gotArticle.Title)
	require.Equal(t, article.Content, gotArticle.Content)
	require.Equal(t, article.IsPublish, gotArticle.IsPublish)
	require.Equal(t, article.Owner, gotArticle.Owner)

}

func randomArticle(t *testing.T, owner uuid.UUID) db.Article {
	articleID, err := uuid.NewRandom()
	require.NoError(t, err)

	article := db.Article{
		ID:         articleID,
		Title:      util.RandomString(10),
		Summary:    util.RandomString(20),
		Content:    util.RandomString(30),
		IsPublish:  false,
		Owner:      owner,
		CategoryID: 1,
	}

	return article
}

func randomGetArticleRow(t *testing.T, owner uuid.UUID, isPublish bool) db.GetArticleRow {
	articleID, err := uuid.NewRandom()
	require.NoError(t, err)

	article := db.GetArticleRow{
		ID:         articleID,
		Title:      util.RandomString(10),
		Summary:    util.RandomString(20),
		Content:    util.RandomString(30),
		IsPublish:  isPublish,
		Owner:      owner,
		CategoryID: 1,
		CategoryName: pgtype.Text{
			String: util.RandomString(10),
			Valid:  true,
		},
	}

	return article
}

func requireBodyMatchGetArticleRow(t *testing.T, body *bytes.Buffer, article db.GetArticleRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var resp getArticleResponse
	err = json.Unmarshal(data, &resp)
	require.NoError(t, err)

	require.Equal(t, article.ID.String(), resp.Article.ID.String())
	require.Equal(t, article.Title, resp.Article.Title)
	require.Equal(t, article.Content, resp.Article.Content)
	require.Equal(t, article.IsPublish, resp.Article.IsPublish)
	require.Equal(t, article.Owner, resp.Article.Owner)
}
