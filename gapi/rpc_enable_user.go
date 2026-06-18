package gapi

import (
	"context"
	"errors"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) EnableUser(ctx context.Context, req *pb.EnableUserRequest) (*pb.EnableUserResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	user, err := server.store.EnableVisitorUser(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to enable user: %v", err)
	}

	return &pb.EnableUserResponse{User: convertUser(user)}, nil
}
