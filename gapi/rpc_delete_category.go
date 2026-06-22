package gapi

import (
	"context"
	"fmt"
	"os"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	cachepkg "github.com/MonitorAllen/nostalgia/internal/cache"
	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	"github.com/MonitorAllen/nostalgia/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.DeleteCategoryResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	arg := db.DeleteCategoryTxParams{
		ID:             req.GetId(),
		DeleteArticles: req.GetDeleteArticles(),
		AfterDelete: func() error {
			return server.taskDistributor.DistributeTaskDelayDeleteCacheDefault(ctx, key.CategoryAllKey)
		},
		AfterDeleteArticles: func(articleRefs []db.ListArticleResourceRefsByCategoryIDRow) error {
			cacheKeys := make([]string, 0, len(articleRefs)*2)

			for _, article := range articleRefs {
				path := fmt.Sprintf("./resources/%s/", article.ID.String())
				if err := os.RemoveAll(path); err != nil {
					return fmt.Errorf("failed to remove article resources: %w", err)
				}

				cacheKeys = append(cacheKeys, key.GetArticleIDKey(article.ID))
				if article.Slug.Valid {
					cacheKeys = append(cacheKeys, key.GetArticleSlugKey(article.Slug.String))
				}
			}

			articleCache := cachepkg.NewArticleCache(server.cache)
			if err := articleCache.InvalidateDetails(ctx, cacheKeys...); err != nil {
				return err
			}

			return articleCache.BumpListVersion(ctx, req.GetId())
		},
	}

	err = server.store.DeleteCategoryTx(ctx, arg)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete category: %v", err)
	}

	return &pb.DeleteCategoryResponse{}, nil
}
