package gapi

import (
	"context"
	"testing"
	"time"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateArticleUsesAuthenticatedAdminAsOwner(t *testing.T) {
	adminID := util.RandUserID()
	title := "Owner bound draft"
	summary := "summary"
	content := "<p>content</p>"
	isPublish := false

	req := &pb.CreateArticleRequest{
		Title:     &title,
		Summary:   &summary,
		Content:   &content,
		IsPublish: &isPublish,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().
		CreateArticle(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(ctx context.Context, arg db.CreateArticleParams) (db.Article, error) {
			require.NotEqual(t, uuid.Nil, arg.ID)
			require.Equal(t, adminID, arg.Owner)
			require.Equal(t, title, arg.Title)
			require.Equal(t, summary, arg.Summary)
			require.Equal(t, content, arg.Content)
			require.Equal(t, isPublish, arg.IsPublish)
			require.Equal(t, int64(1), arg.CategoryID)

			return db.Article{
				ID:         arg.ID,
				Title:      arg.Title,
				Summary:    arg.Summary,
				Content:    arg.Content,
				IsPublish:  arg.IsPublish,
				Owner:      arg.Owner,
				CategoryID: arg.CategoryID,
				Cover:      arg.Cover,
			}, nil
		})

	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	ctx := newContextWithUserBearerToken(t, server.tokenMaker, adminID, "owner", util.Admin, time.Minute)

	resp, err := server.CreateArticle(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp.GetArticle())
	require.Equal(t, adminID.String(), resp.GetArticle().GetOwner())
}
