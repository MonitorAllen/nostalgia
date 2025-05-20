package gapi

import (
	"context"
	"fmt"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func (server *Server) UpdateArticle(ctx context.Context, req *pb.UpdateArticleRequest) (*pb.UpdateArticleResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateUpdateArticleRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	articleId, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	arg := db.UpdateArticleParams{
		Title: pgtype.Text{
			String: req.GetTitle(),
			Valid:  req.Title != nil,
		},
		Summary: pgtype.Text{
			String: req.GetSummary(),
			Valid:  req.Summary != nil,
		},
		IsPublish: pgtype.Bool{
			Bool:  req.GetIsPublish(),
			Valid: req.IsPublish != nil,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		ID: articleId,
	}

	article, err := server.store.UpdateArticle(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't update article: %v", err)
	}

	resp := &pb.UpdateArticleResponse{
		Article: convertOnlyArticle(article),
	}

	return resp, nil
}

func validateUpdateArticleRequest(req *pb.UpdateArticleRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if req.GetId() == "" {
		violations = append(violations, fieldViolation("id", fmt.Errorf("id is required")))
	}

	return
}
