package gapi

import (
	"context"
	"github.com/MonitorAllen/nostalgia/internal/cache"
	"github.com/MonitorAllen/nostalgia/pb"
)

func (server *Server) LogoutAdmin(ctx context.Context, req *pb.LogoutAdminRequest) (*pb.LogoutAdminResponse, error) {
	accessPayload, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	_ = server.cache.Del(ctx, cache.GetAdminSessionKey(accessPayload.AdminID))

	return &pb.LogoutAdminResponse{}, nil
}
