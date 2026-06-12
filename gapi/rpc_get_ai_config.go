package gapi

import (
	"context"

	"github.com/MonitorAllen/nostalgia/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) GetAIConfig(ctx context.Context, _ *pb.GetAIConfigRequest) (*pb.GetAIConfigResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	cfg, err := server.resolveAIPolishConfig(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to load AI config")
	}

	return cfg.toResponse(), nil
}
