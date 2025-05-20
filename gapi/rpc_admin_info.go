package gapi

import (
	"context"
	"encoding/json"
	"github.com/MonitorAllen/nostalgia/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
)

func (server *Server) AdminInfo(ctx context.Context, req *pb.AdminInfoRequest) (*pb.AdminInfoResponse, error) {
	accessPayload, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	adminSessionStr, err := server.redisService.Get(adminSessionKey + strconv.FormatInt(accessPayload.AdminID, 10))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	var adminSession AdminSession

	err = json.Unmarshal([]byte(adminSessionStr), &adminSession)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot get admin session: %v", err)
	}

	getAdmin, err := server.store.GetAdmin(ctx, adminSession.Payload.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot get admin: %v", err)
	}

	return &pb.AdminInfoResponse{
		Admin: convertAdmin(getAdmin),
	}, nil
}
