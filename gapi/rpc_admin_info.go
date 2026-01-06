package gapi

import (
	"context"
	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	"github.com/MonitorAllen/nostalgia/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) AdminInfo(ctx context.Context, req *pb.AdminInfoRequest) (*pb.AdminInfoResponse, error) {
	accessPayload, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	expired, err := server.cache.IsExpired(ctx, key.GetAdminSessionKey(accessPayload.AdminID))
	if err != nil {
		return nil, status.Error(codes.Internal, "获取会话失败")
	}

	if expired {
		return nil, status.Errorf(codes.Unauthenticated, "会话过期，请重新登录")
	}

	getAdmin, err := server.store.GetAdmin(ctx, accessPayload.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot get admin: %v", err)
	}

	return &pb.AdminInfoResponse{
		Admin: convertAdmin(getAdmin),
	}, nil
}
