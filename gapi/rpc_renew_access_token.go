package gapi

import (
	"context"
	"errors"
	"github.com/MonitorAllen/nostalgia/internal/cache"
	"time"

	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) RenewAccessToken(ctx context.Context, req *pb.RenewAccessTokenRequest) (*pb.RenewAccessTokenResponse, error) {
	refreshPayload, err := server.tokenMaker.VerifyAdminToken(req.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())

	}

	adminSessionKey := cache.GetAdminSessionKey(refreshPayload.AdminID)
	var adminSession AdminSession
	_, err = server.cache.Get(ctx, adminSessionKey, &adminSession)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, status.Error(codes.Unauthenticated, "session has expired")
		}
		return nil, status.Error(codes.Internal, "cannot fetch session")
	}

	// 检查session状态
	if adminSession.IsBlocked {
		return nil, status.Error(codes.PermissionDenied, "blocked session")
	}

	if adminSession.Payload.AdminID != refreshPayload.AdminID {
		return nil, status.Error(codes.Unauthenticated, "incorrect session user")
	}

	if adminSession.RefreshToken != req.RefreshToken {
		return nil, status.Error(codes.Unauthenticated, "mismatched session token")
	}

	if time.Now().After(refreshPayload.ExpireAt) {
		return nil, status.Error(codes.Unauthenticated, "session expired")
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateAdminToken(refreshPayload.AdminID, refreshPayload.Username, refreshPayload.RoleID, server.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	metadata := server.extractMetadata(ctx)

	adminSession.Payload = accessPayload
	adminSession.UserAgent = metadata.UserAgent
	adminSession.ClientIp = metadata.ClientIP

	err = server.cache.Set(ctx, adminSessionKey, adminSession, server.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &pb.RenewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: timestamppb.New(accessPayload.ExpireAt),
	}

	return resp, nil
}
