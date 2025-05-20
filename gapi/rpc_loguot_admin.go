package gapi

import (
	"context"
	"github.com/MonitorAllen/nostalgia/pb"
	"strconv"
)

func (server *Server) LogoutAdmin(ctx context.Context, req *pb.LogoutAdminRequest) (*pb.LogoutAdminResponse, error) {
	accessPayload, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	_ = server.redisService.Del(adminSessionKey + strconv.FormatInt(accessPayload.AdminID, 10))

	return &pb.LogoutAdminResponse{}, nil
}
