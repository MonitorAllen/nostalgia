package gapi

import (
	"context"
	"time"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateAdmin(ctx context.Context, req *pb.UpdateAdminRequest) (*pb.UpdateAdminResponse, error) {
	payload, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if payload.RoleID != 1 && payload.AdminID != req.GetId() {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	arg := db.UpdateAdminParams{
		ID: req.GetId(),
		Username: pgtype.Text{
			String: req.GetUsername(),
			Valid:  req.Username != nil,
		},
		RoleID: pgtype.Int8{
			Int64: req.GetRoleId(),
			Valid: req.RoleId != nil,
		},
		IsActive: pgtype.Bool{
			Bool:  req.GetIsActive(),
			Valid: req.IsActive != nil,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}

	if req.Password != nil {
		hashedPassword, err := util.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
		}
		arg.HashedPassword = pgtype.Text{
			String: hashedPassword,
			Valid:  true,
		}
	}

	updateAdmin, err := server.store.UpdateAdmin(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update admin: %v", err)
	}

	return &pb.UpdateAdminResponse{
		Admin: convertAdmin(updateAdmin),
	}, nil
}
