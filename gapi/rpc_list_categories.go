package gapi

import (
	"context"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/MonitorAllen/nostalgia/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateListCategoriesRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	arg := db.ListCategoriesCountArticlesParams{
		Offset: (req.GetPage() - 1) * req.GetLimit(),
		Limit:  req.GetLimit(),
	}
	categories, err := server.store.ListCategoriesCountArticles(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list categories: %s", err)
	}

	countCategories, err := server.store.CountCategories(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to count categories: %s", err)
	}

	resp := &pb.ListCategoriesResponse{
		Categories: convertCategoriesCountArticleRow(categories),
		Count:      countCategories,
	}

	return resp, nil
}

func validateListCategoriesRequest(req *pb.ListCategoriesRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validator.ValidatePage(req.GetPage()); err != nil {
		violations = append(violations, fieldViolation("page", err))
	}
	if err := validator.ValidateLimit(req.GetLimit(), 0); err != nil {
		violations = append(violations, fieldViolation("limit", err))
	}

	return
}
