package gapi

import (
	"context"
	"github.com/MonitorAllen/nostalgia/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) ListAllCategories(ctx context.Context, req *pb.ListAllCategoriesRequest) (*pb.ListAllCategoriesResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	categories, err := server.store.ListAllCategories(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list categories: %s", err)
	}

	resp := &pb.ListAllCategoriesResponse{
		Categories: convertCategories(categories),
	}

	return resp, nil
}
