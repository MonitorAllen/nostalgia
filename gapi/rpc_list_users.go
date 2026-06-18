package gapi

import (
	"context"
	"strings"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func normalizeUserListPage(page int32) int32 {
	if page < 1 {
		return 1
	}
	return page
}

func normalizeUserListLimit(limit int32) int32 {
	switch limit {
	case 10, 20, 50:
		return limit
	default:
		return 20
	}
}

func normalizeUserStatus(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "enabled", "disabled":
		return normalized
	default:
		return "all"
	}
}

func optionalSearchText(value string) pgtype.Text {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: trimmed, Valid: true}
}

func (server *Server) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	page := normalizeUserListPage(req.GetPage())
	limit := normalizeUserListLimit(req.GetLimit())
	statusValue := normalizeUserStatus(req.GetStatus())
	q := optionalSearchText(req.GetQ())

	arg := db.ListAdminUsersParams{
		Limit:  limit,
		Offset: (page - 1) * limit,
		Status: statusValue,
		Q:      q,
	}

	users, err := server.store.ListAdminUsers(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	count, err := server.store.CountAdminUsersByFilter(ctx, db.CountAdminUsersByFilterParams{
		Status: statusValue,
		Q:      q,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count users: %v", err)
	}

	resp := &pb.ListUsersResponse{
		Users: make([]*pb.User, 0, len(users)),
		Count: count,
	}
	for _, user := range users {
		resp.Users = append(resp.Users, convertAdminUserRow(user))
	}

	return resp, nil
}
