package gapi

import (
	"context"
	"errors"
	"time"

	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	resp := &pb.ListCategoriesResponse{}
	ok, err := server.cache.Get(ctx, key.CategoryAllKey, resp)
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Error().
			Err(err).
			Str("key", key.CategoryAllKey).
			Str("module", "category").
			Str("action", "cache_get").
			Msg("获取分类缓存失败，降级为仅数据库")
	}
	if ok {
		return resp, nil
	}

	categories, err := server.store.ListCategoriesCountArticles(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list categories: %s", err)
	}

	countCategories := len(categories)

	resp.Categories = convertCategoriesCountArticleRow(categories)
	resp.Count = int64(countCategories)

	err = server.cache.Set(ctx, key.CategoryAllKey, resp, 7*24*time.Hour)
	if err != nil {
		log.Error().
			Err(err).
			Str("key", key.CategoryAllKey).
			Str("module", "category").
			Str("action", "cache_set").
			Msg("设置分类缓存失败，降级为仅数据库")
	}

	return resp, nil
}
