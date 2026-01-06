package gapi

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/MonitorAllen/nostalgia/internal/cache/key"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) DeleteArticle(ctx context.Context, req *pb.DeleteArticleRequest) (*pb.DeleteArticleResponse, error) {
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

	_, err = server.store.GetArticle(ctx, articleId)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "article not found")
		}
		return nil, status.Error(codes.Internal, "failed to fetch article")
	}

	arg := db.DeleteArticleTxParams{
		ID: articleId,
		AfterUpdate: func(articleID uuid.UUID) error {
			path := fmt.Sprintf("./resources/%s/", articleID.String())

			err := os.RemoveAll(path)
			if err != nil {
				return fmt.Errorf("failed to remove article resources: %w", err)
			}

			return server.taskDistributor.DistributeTaskDelayDeleteCacheDefault(ctx, key.GetArticleIDKey(articleId))
		},
	}

	err = server.store.DeleteArticleTx(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot delete article: %v", err)
	}

	return &pb.DeleteArticleResponse{}, nil
}
