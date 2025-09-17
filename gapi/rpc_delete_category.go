package gapi

import (
	"context"
	"github.com/MonitorAllen/nostalgia/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.DeleteCategoryResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	err = server.store.DeleteCategoryTx(ctx, req.GetId())

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete category: %v", err)
	}

	return &pb.DeleteCategoryResponse{}, nil
}
