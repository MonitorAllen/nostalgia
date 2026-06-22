package gapi

import (
	"context"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func normalizeCategoryListPage(page int32) int32 {
	if page < 1 {
		return 1
	}
	return page
}

func normalizeCategoryListLimit(limit int32) int32 {
	switch limit {
	case 10, 20, 50:
		return limit
	default:
		return 20
	}
}

func (server *Server) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	page := normalizeCategoryListPage(req.GetPage())
	limit := normalizeCategoryListLimit(req.GetLimit())
	categories, err := server.store.ListCategoriesCountArticles(ctx, db.ListCategoriesCountArticlesParams{
		Limit:  limit,
		Offset: (page - 1) * limit,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list categories: %v", err)
	}

	count, err := server.store.CountCategories(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count categories: %v", err)
	}

	return &pb.ListCategoriesResponse{
		Categories: convertCategoriesCountArticleRow(categories),
		Count:      count,
	}, nil
}
