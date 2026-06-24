package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/MonitorAllen/nostalgia/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
)

func (server *Server) authorizeAdmin(ctx context.Context) (*token.Payload, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, "", fmt.Errorf("missing metadata")
	}

	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, "", fmt.Errorf("missing authorization header")
	}

	authHeader := values[0]
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return nil, "", fmt.Errorf("invalid authorization header format")
	}

	authType := strings.ToLower(fields[0])
	if authType != authorizationBearer {
		return nil, "", fmt.Errorf("unsupported authorization type: %s", authType)
	}

	accessToken := fields[1]
	payload, err := server.tokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, "", fmt.Errorf("invalid access token: %s", err)
	}
	if payload.Role != util.Admin {
		return nil, "", status.Error(codes.PermissionDenied, "admin role required")
	}

	// Check if user has been disabled
	var disabled bool
	if found, _ := server.cache.Get(ctx, key.GetUserDisabledKey(payload.UserID.String()), &disabled); found && disabled {
		return nil, "", status.Error(codes.PermissionDenied, "account disabled")
	}

	return payload, accessToken, nil
}
