package gapi

import (
	"context"
	"testing"
	"time"

	mockdb "github.com/MonitorAllen/nostalgia/db/mock"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestListArticlesAppliesTitleFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, newGAPITestStore(store), nil, nil)
	ctx := newContextWithAdminBearerToken(t, server.tokenMaker, time.Minute)

	store.EXPECT().
		ListAllArticles(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, arg db.ListAllArticlesParams) ([]db.ListAllArticlesRow, error) {
			require.Equal(t, int32(12), arg.Limit)
			require.Equal(t, int32(12), arg.Offset)
			require.True(t, arg.Title.Valid)
			require.Equal(t, "redis", arg.Title.String)
			return []db.ListAllArticlesRow{
				{
					ID:    uuid.New(),
					Title: "Redis cache strategy",
				},
			}, nil
		})
	store.EXPECT().
		CountAllArticles(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, title pgtype.Text) (int64, error) {
			require.True(t, title.Valid)
			require.Equal(t, "redis", title.String)
			return 1, nil
		})

	resp, err := server.ListArticles(ctx, &pb.ListArticlesRequest{
		Page:  2,
		Limit: 12,
		Title: " redis ",
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), resp.GetCount())
	require.Len(t, resp.GetArticles(), 1)
	require.Equal(t, "Redis cache strategy", resp.GetArticles()[0].GetTitle())
}
