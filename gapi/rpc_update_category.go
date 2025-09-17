package gapi

import (
	"context"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) UpdateCategory(ctx context.Context, res *pb.UpdateCategoryRequest) (*pb.UpdateCategoryResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	arg := db.UpdateCategoryParams{
		ID:   res.GetId(),
		Name: res.GetName(),
	}
	category, err := server.store.UpdateCategory(ctx, arg)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "Category not found")
		}
		if errors.Is(err, db.ErrUniqueViolation) {
			return nil, status.Error(codes.AlreadyExists, "Duplicate category name.")
		}
		return nil, status.Errorf(codes.Internal, "Error updating category: %v", err)
	}

	resp := &pb.UpdateCategoryResponse{
		Category: &pb.Category{
			ID:        category.ID,
			Name:      category.Name,
			CreatedAt: timestamppb.New(category.CreatedAt),
			UpdatedAt: timestamppb.New(category.UpdatedAt),
		},
	}

	return resp, nil
}
