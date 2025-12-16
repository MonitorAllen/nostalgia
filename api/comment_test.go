package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateCommentAPI(t *testing.T) {
	user, _ := randomUser(t)
	article := randomArticle(t, user.ID, true)

	content := "大佬666！"

	sendCommentUser, _ := randomUser(t)

	comment := db.Comment{
		Content:    content,
		ArticleID:  article.ID,
		ParentID:   0,
		Likes:      0,
		FromUserID: sendCommentUser.ID,
		ToUserID:   article.Owner,
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"content":      comment.Content,
				"article_id":   comment.ArticleID,
				"parent_id":    comment.ParentID,
				"from_user_id": comment.FromUserID,
				"to_user_id":   comment.ToUserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, sendCommentUser.ID, sendCommentUser.Username, sendCommentUser.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateCommentParams{
					Content:    comment.Content,
					ArticleID:  comment.ArticleID,
					ParentID:   comment.ParentID,
					FromUserID: comment.FromUserID,
					ToUserID:   comment.ToUserID,
				}

				store.EXPECT().
					CreateComment(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(comment, nil)
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg.ToUserID)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchComment(t, recorder.Body, comment)
			},
		},
		{
			name: "BadRequest",
			body: gin.H{
				"content":      "",
				"article_id":   comment.ArticleID,
				"parent_id":    comment.ParentID,
				"from_user_id": comment.FromUserID,
				"to_user_id":   comment.ToUserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, sendCommentUser.ID, sendCommentUser.Username, sendCommentUser.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateComment(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			body: gin.H{
				"content":      comment.Content,
				"article_id":   comment.ArticleID,
				"parent_id":    comment.ParentID,
				"from_user_id": comment.FromUserID,
				"to_user_id":   comment.ToUserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, sendCommentUser.ID, sendCommentUser.Username, sendCommentUser.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateComment(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Comment{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			body: gin.H{
				"content":      comment.Content,
				"article_id":   comment.ArticleID,
				"parent_id":    comment.ParentID,
				"from_user_id": comment.FromUserID,
				"to_user_id":   comment.ToUserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, sendCommentUser.ID, sendCommentUser.Username, sendCommentUser.Role, -time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateComment(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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

			server := newTestServer(t, store, nil, nil)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/comments"

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchComment(t *testing.T, body *bytes.Buffer, comment db.Comment) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotComment createCommentResponse
	err = json.Unmarshal(data, &gotComment)
	require.NoError(t, err)

	require.NotEmpty(t, gotComment)

	require.Equal(t, comment.Content, gotComment.Comment.Content)
	require.Equal(t, comment.ArticleID, gotComment.Comment.ArticleID)
	require.Equal(t, comment.ParentID, gotComment.Comment.ParentID)
	require.Equal(t, comment.FromUserID, gotComment.Comment.FromUserID)
	require.Equal(t, comment.ToUserID, gotComment.Comment.ToUserID)
	require.Empty(t, gotComment.Comment.Child)
}
