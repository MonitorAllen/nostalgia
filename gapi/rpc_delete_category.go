package gapi

import (
	"context"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
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
		ID: req.GetId(),
		AfterDelete: func() error {
			return server.taskDistributor.DistributeTaskDelayDeleteCacheDefault(ctx, key.CategoryAllKey)
		},
	}

	err = server.store.DeleteCategoryTx(ctx, arg)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete category: %v", err)
	}

	return &pb.DeleteCategoryResponse{}, nil
}
