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

	article := randomGetArticleRow(t, user.ID)

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
					GetArticle(gomock.Any(), gomock.Any()).
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
					GetArticle(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.GetArticleRow{}, db.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
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
				page:  1,
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
		ID:        articleID,
		Title:     util.RandomString(10),
		Summary:   util.RandomString(20),
		Content:   util.RandomString(30),
		IsPublish: false,
		Owner:     owner,
	}

	return article
}

func randomGetArticleRow(t *testing.T, owner uuid.UUID) db.GetArticleRow {
	articleID, err := uuid.NewRandom()
	require.NoError(t, err)

	article := db.GetArticleRow{
		ID:         articleID,
		Title:      util.RandomString(10),
		Summary:    util.RandomString(20),
		Content:    util.RandomString(30),
		IsPublish:  false,
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
