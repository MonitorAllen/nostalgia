package gapi

import (
	"context"

	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateAIConfig(ctx context.Context, req *pb.UpdateAIConfigRequest) (*pb.GetAIConfigResponse, error) {
	payload, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}
	if server.store == nil {
		return nil, status.Error(codes.FailedPrecondition, "AI config storage is unavailable")
	}

	cfg, err := server.buildUpdatedAIConfig(ctx, req)
	if err != nil {
		return nil, err
	}

	saved, err := server.saveAIConfig(ctx, cfg, pgtype.UUID{Bytes: payload.UserID, Valid: true})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to save AI config")
	}

	return saved.toResponse(), nil
}
