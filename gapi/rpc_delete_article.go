package gapi

import (
	"context"
	"fmt"
	"os"

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

	arg := db.DeleteArticleTxParams{
		ID: articleId,
		AfterUpdate: func(articleID uuid.UUID) error {
			path := fmt.Sprintf("./resources/%s/", articleID.String())

			err := os.RemoveAll(path)

			return err
		},
	}

	err = server.store.DeleteArticleTx(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot delete article: %v", err)
	}

	return &pb.DeleteArticleResponse{}, nil
}
