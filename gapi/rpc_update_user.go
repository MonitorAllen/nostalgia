package gapi

import (
	"context"
	"errors"
	"strings"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	fullName := strings.TrimSpace(req.GetFullName())
	email := strings.TrimSpace(req.GetEmail())
	if fullName == "" {
		return nil, status.Error(codes.InvalidArgument, "full name is required")
	}
	if email == "" || !strings.Contains(email, "@") {
		return nil, status.Error(codes.InvalidArgument, "valid email is required")
	}

	user, err := server.store.UpdateVisitorUser(ctx, db.UpdateVisitorUserParams{
		ID:              id,
		FullName:        fullName,
		Email:           email,
		IsEmailVerified: req.GetIsEmailVerified(),
	})
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if db.ErrorCode(err) == db.UniqueViolation {
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &pb.UpdateUserResponse{User: convertUser(user)}, nil
}
