package gapi

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
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
		return nil, status.Errorf(codes.Internal, "failed to marshal access payload")

	}

	session, err := server.redisService.Get(adminSessionKey + strconv.FormatInt(refreshPayload.AdminID, 10))
	if err != nil {
		if errors.Is(redis.Nil, err) {
			return nil, status.Error(codes.Unauthenticated, "session has expired")
		}
		return nil, status.Error(codes.Internal, "cannot fetch session")
	}

	var adminSession AdminSession

	err = json.Unmarshal([]byte(session), &adminSession)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot unmarshal session")
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

	sessionBytes, err := json.Marshal(adminSession)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = server.redisService.Set(adminSessionKey+strconv.FormatInt(accessPayload.AdminID, 10), string(sessionBytes), server.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &pb.RenewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: timestamppb.New(accessPayload.ExpireAt),
	}

	return resp, nil
}
