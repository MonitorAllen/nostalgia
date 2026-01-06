package gapi

import (
	"context"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	"github.com/MonitorAllen/nostalgia/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) CreateCategory(ctx context.Context, res *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	arg := db.CreateCategoryTxParams{
		Name: res.GetName(),
		AfterCreate: func() error {
			return server.taskDistributor.DistributeTaskDelayDeleteCacheDefault(ctx, key.CategoryAllKey)
		},
	}

	result, err := server.store.CreateCategoryTx(ctx, arg)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			return nil, status.Errorf(codes.AlreadyExists, "Duplicate category name.")
		}
		return nil, status.Errorf(codes.Internal, "Failed to create category: %v", "please try again later or contact support")
	}

	resp := &pb.CreateCategoryResponse{
		Category: &pb.Category{
			Id:        result.Category.ID,
			Name:      result.Category.Name,
			CreatedAt: timestamppb.New(result.Category.CreatedAt),
			UpdatedAt: timestamppb.New(result.Category.UpdatedAt),
		},
	}

	return resp, nil
}
