package gapi

import (
	"context"
	"errors"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) ListArticles(ctx context.Context, req *pb.ListArticlesRequest) (*pb.ListArticlesResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	arg := db.ListAllArticlesParams{
		Limit:  req.GetLimit(),
		Offset: (req.GetPage() - 1) * req.GetLimit(),
	}

	articleList, err := server.store.ListAllArticles(ctx, arg)
	if err != nil && !errors.Is(err, db.ErrRecordNotFound) {
		return nil, status.Errorf(codes.Internal, "failed to find articles: %v", err)
	}

	countArticles, err := server.store.CountAllArticles(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count articles: %v", err)
	}

	resp := &pb.ListArticlesResponse{
		Articles: convertArticleList(articleList),
		Count:    &countArticles,
	}

	return resp, nil
}
