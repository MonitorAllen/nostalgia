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

	post := randomArticle(t, user.ID)

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
				"title":      post.Title,
				"summary":    post.Summary,
				"content":    post.Content,
				"is_publish": post.IsPublish,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateArticleParams{
					Title:     post.Title,
					Summary:   post.Summary,
					Content:   post.Content,
					IsPublish: post.IsPublish,
					Owner:     post.Owner,
				}
				store.EXPECT().
					CreateArticle(gomock.Any(), EqCreateArticleParams(arg, user)).
					Times(1).
					Return(post, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchArticle(t, recorder.Body, post)
			},
		},
		{
			name: "BadRequest",
			body: gin.H{
				"summary":    post.Summary,
				"content":    post.Content,
				"is_publish": post.IsPublish,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Role, time.Minute)
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
				"title":      post.Title,
				"summary":    post.Summary,
				"content":    post.Content,
				"is_publish": post.IsPublish,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Role, time.Minute)
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
			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/posts"

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

	post := randomArticle(t, user.ID)

	testCases := []struct {
		name          string
		postID        string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			postID: post.ID.String(),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetArticle(gomock.Any(), gomock.Eq(post.ID)).
					Times(1).
					Return(post, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchArticle(t, recorder.Body, post)
			},
		},
		{
			name:   "BadRequest",
			postID: "not-uuid",
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
			name:   "InternalError",
			postID: post.ID.String(),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetArticle(gomock.Any(), gomock.Any()).
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
			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/posts/%s", tc.postID)

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
	posts := make([]db.Article, n)
	for i := 0; i < n; i++ {
		posts[i] = randomArticle(t, user.ID)
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
					Return(posts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchArticles(t, recorder.Body, posts)
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
					Return([]db.Article{}, sql.ErrConnDone)
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
			server := newTestServer(t, store, nil)
			recorder := httptest.NewRecorder()

			url := "/posts"

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

func requireBodyMatchArticles(t *testing.T, body *bytes.Buffer, posts []db.Article) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotArticles []db.Article
	err = json.Unmarshal(data, &gotArticles)
	require.NoError(t, err)
	require.Equal(t, posts, gotArticles)
}

func requireBodyMatchArticle(t *testing.T, body *bytes.Buffer, post db.Article) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotArticle db.Article
	err = json.Unmarshal(data, &gotArticle)
	require.NoError(t, err)

	require.Equal(t, post.ID.String(), gotArticle.ID.String())
	require.Equal(t, post.Title, gotArticle.Title)
	require.Equal(t, post.Content, gotArticle.Content)
	require.Equal(t, post.IsPublish, gotArticle.IsPublish)
	require.Equal(t, post.Owner, gotArticle.Owner)

}

func randomArticle(t *testing.T, owner uuid.UUID) db.Article {
	postID, err := uuid.NewRandom()
	require.NoError(t, err)

	post := db.Article{
		ID:        postID,
		Title:     util.RandomString(10),
		Summary:   util.RandomString(20),
		Content:   util.RandomString(30),
		IsPublish: false,
		Owner:     owner,
	}

	return post
}
