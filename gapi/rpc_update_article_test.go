package gapi

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	mockcache "github.com/MonitorAllen/nostalgia/internal/cache/mock"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/MonitorAllen/nostalgia/util"
	mockwk "github.com/MonitorAllen/nostalgia/worker/mock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestUpdateArticleInvalidatesDetailAndListCaches(t *testing.T) {
	articleID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	oldSlug := "old-slug"
	newSlug := "new-slug"
	newCategoryID := int64(2)

	previousArticle := db.GetArticleRow{
		ID:         articleID,
		CategoryID: 1,
		Slug:       pgtype.Text{String: oldSlug, Valid: true},
	}
	updatedArticle := db.Article{
		ID:         articleID,
		Title:      "updated",
		Content:    "<p>updated</p>",
		IsPublish:  true,
		CategoryID: newCategoryID,
		Slug:       pgtype.Text{String: newSlug, Valid: true},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	taskDistributor := mockwk.NewMockTaskDistributor(ctrl)
	redisCache := mockcache.NewMockCache(ctrl)

	store.EXPECT().GetArticle(gomock.Any(), gomock.Eq(articleID)).Times(1).Return(previousArticle, nil)
	taskDistributor.EXPECT().
		DistributeTaskDelayDeleteCacheDefault(
			gomock.Any(),
			gomock.Eq(key.GetArticleIDKey(articleID)),
			gomock.Eq(key.GetArticleSlugKey(oldSlug)),
			gomock.Eq(key.GetArticleSlugKey(newSlug)),
			gomock.Eq(key.CategoryAllKey),
		).
		Times(1).
		Return(nil)
	redisCache.EXPECT().Incr(gomock.Any(), gomock.Eq(key.GetArticleListVersionKey(int64(0)))).Times(1).Return(int64(1), nil)
	redisCache.EXPECT().Incr(gomock.Any(), gomock.Eq(key.GetArticleListVersionKey(int64(1)))).Times(1).Return(int64(1), nil)
	redisCache.EXPECT().Incr(gomock.Any(), gomock.Eq(key.GetArticleListVersionKey(newCategoryID))).Times(1).Return(int64(1), nil)
	store.EXPECT().
		UpdateArticleTx(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(ctx context.Context, arg db.UpdateArticleTxParams) (db.UpdateArticleTxResult, error) {
			require.Equal(t, articleID, arg.ID)
			require.Equal(t, newCategoryID, arg.CategoryID.Int64)
			require.NoError(t, arg.AfterUpdate(updatedArticle))
			return db.UpdateArticleTxResult{Article: updatedArticle}, nil
		})

	server := newTestServer(t, newGAPITestStore(store), taskDistributor, redisCache)
	server.config.ResourcePath = t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(server.config.ResourcePath, "articles", articleID.String()), 0o755))
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)
	req := &pb.UpdateArticleRequest{
		Id:         articleID.String(),
		Slug:       &newSlug,
		CategoryId: &newCategoryID,
	}

	resp, err := server.UpdateArticle(ctx, req)
	require.NoError(t, err)
	require.Equal(t, articleID.String(), resp.GetArticle().GetId())
}

func TestUpdateArticleRejectsVisitorBeforeInvalidatingCache(t *testing.T) {
	articleID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	taskDistributor := mockwk.NewMockTaskDistributor(ctrl)
	redisCache := mockcache.NewMockCache(ctrl)

	store.EXPECT().GetArticle(gomock.Any(), gomock.Any()).Times(0)
	store.EXPECT().UpdateArticleTx(gomock.Any(), gomock.Any()).Times(0)
	taskDistributor.EXPECT().DistributeTaskDelayDeleteCacheDefault(gomock.Any(), gomock.Any()).Times(0)
	redisCache.EXPECT().Incr(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, newGAPITestStore(store), taskDistributor, redisCache)
	ctx := newContextWithUserBearerToken(t, server.tokenMaker, util.RandUserID(), "visitor", util.Visitor, time.Minute)
	req := &pb.UpdateArticleRequest{Id: articleID.String()}

	resp, err := server.UpdateArticle(ctx, req)
	require.Error(t, err)
	require.Nil(t, resp)
}
