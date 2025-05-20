package gapi

import (
	"context"
	"encoding/json"
	"errors"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/MonitorAllen/nostalgia/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
)

const (
	adminSessionKey = "admin:session:"
)

type AdminSession struct {
	Payload      *token.AdminPayload
	RefreshToken string
	UserAgent    string
	ClientIp     string
	IsBlocked    bool
}

func (server *Server) LoginAdmin(ctx context.Context, req *pb.LoginAdminRequest) (*pb.LoginAdminResponse, error) {
	violations := validateLoginAdminRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	admin, err := server.store.GetAdmin(ctx, req.GetUsername())
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not fount")
		}
		return nil, status.Errorf(codes.Internal, "failed to find user")
	}

	err = util.CheckPassword(req.GetPassword(), admin.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "incorrect password")
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateAdminToken(admin.ID, admin.Username, admin.RoleID, server.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to  create access token")
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateAdminToken(admin.ID, admin.Username, admin.RoleID, server.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token")
	}

	metadata := server.extractMetadata(ctx)

	// 使用 redis 管理 admin 会话
	adminSession := AdminSession{
		Payload:      accessPayload,
		RefreshToken: refreshToken,
		UserAgent:    metadata.UserAgent,
		ClientIp:     metadata.ClientIP,
		IsBlocked:    false,
	}

	bytesPayload, err := json.Marshal(adminSession)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal access payload")
	}

	err = server.redisService.Set(adminSessionKey+strconv.FormatInt(accessPayload.AdminID, 10), string(bytesPayload), server.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set session")
	}

	resp := &pb.LoginAdminResponse{
		Admin:                 convertAdmin(admin),
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpireAt),
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpireAt),
	}

	return resp, nil
}

func validateLoginAdminRequest(req *pb.LoginAdminRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if err := validator.ValidatePassword(req.GetPassword(), 3); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	return
}
