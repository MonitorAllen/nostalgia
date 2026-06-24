package gapi

import (
	"context"
	"errors"
	"strings"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) DisableUser(ctx context.Context, req *pb.DisableUserRequest) (*pb.DisableUserResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	result, err := server.store.DisableVisitorUserTx(ctx, db.DisableVisitorUserTxParams{
		ID:             id,
		DisabledReason: strings.TrimSpace(req.GetReason()),
	})
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to disable user: %v", err)
	}

	// Mark user as disabled in cache so active tokens are rejected
	_ = server.cache.Set(ctx, key.GetUserDisabledKey(id.String()), true, 0)

	return &pb.DisableUserResponse{User: convertUser(result.User)}, nil
}
