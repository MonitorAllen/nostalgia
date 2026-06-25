package gapi

import (
	"context"
	"testing"
	"time"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	mockcache "github.com/MonitorAllen/nostalgia/internal/cache/mock"
	"github.com/MonitorAllen/nostalgia/pb"
	mockwk "github.com/MonitorAllen/nostalgia/worker/mock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestDeleteArticleInvalidatesDetailAndListCaches(t *testing.T) {
	articleID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	slug := "deleted-slug"
	categoryID := int64(7)
	article := db.GetArticleRow{
		ID:         articleID,
		CategoryID: categoryID,
		Slug:       pgtype.Text{String: slug, Valid: true},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	taskDistributor := mockwk.NewMockTaskDistributor(ctrl)
	redisCache := mockcache.NewMockCache(ctrl)

	// authorizeAdmin checks the disabled-user flag; report "active"
	redisCache.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(false, nil).
		AnyTimes()

	store.EXPECT().GetArticle(gomock.Any(), gomock.Eq(articleID)).Times(1).Return(article, nil)
	taskDistributor.EXPECT().
		DistributeTaskDelayDeleteCacheDefault(
			gomock.Any(),
			gomock.Eq(key.GetArticleIDKey(articleID)),
			gomock.Eq(key.GetArticleSlugKey(slug)),
			gomock.Eq(key.CategoryAllKey),
		).
		Times(1).
		Return(nil)
	redisCache.EXPECT().Incr(gomock.Any(), gomock.Eq(key.GetArticleListVersionKey(int64(0)))).Times(1).Return(int64(1), nil)
	redisCache.EXPECT().Incr(gomock.Any(), gomock.Eq(key.GetArticleListVersionKey(categoryID))).Times(1).Return(int64(1), nil)
	store.EXPECT().
		DeleteArticleTx(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(ctx context.Context, arg db.DeleteArticleTxParams) error {
			require.Equal(t, articleID, arg.ID)
			return arg.AfterUpdate(articleID)
		})

	server := newTestServer(t, newGAPITestStore(store), taskDistributor, redisCache)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)
	resp, err := server.DeleteArticle(ctx, &pb.DeleteArticleRequest{Id: articleID.String()})

	require.NoError(t, err)
	require.NotNil(t, resp)
}
