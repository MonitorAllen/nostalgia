package gapi

import (
	"context"

	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) GetArticle(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}

	articleId, err := uuid.Parse(id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	getArticle, err := server.store.GetArticle(ctx, articleId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot get article: %v", err)
	}

	return &pb.GetArticleResponse{
		Article: convertArticleWithCategory(getArticle, req.GetNeedContent()),
	}, nil
}
