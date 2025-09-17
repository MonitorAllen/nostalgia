package gapi

import (
	"context"
	"errors"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
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

	category, err := server.store.CreateCategory(ctx, res.GetName())
	if err != nil {
		if errors.Is(err, db.ErrUniqueViolation) {
			return nil, status.Errorf(codes.AlreadyExists, "Duplicate category name.")
		}
		return nil, status.Errorf(codes.Internal, "Could not create category: %v", err)
	}

	resp := &pb.CreateCategoryResponse{
		Category: &pb.Category{
			ID:        category.ID,
			Name:      category.Name,
			CreatedAt: timestamppb.New(category.CreatedAt),
			UpdatedAt: timestamppb.New(category.UpdatedAt),
		},
	}

	return resp, nil
}
