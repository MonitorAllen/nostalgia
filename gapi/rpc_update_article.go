package gapi

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"time"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	arg := db.UpdateArticleTxParams{
		UpdateArticleParams: db.UpdateArticleParams{
			ID: articleId,
			Title: pgtype.Text{
				String: req.GetTitle(),
				Valid:  req.Title != nil,
			},
			Summary: pgtype.Text{
				String: req.GetSummary(),
				Valid:  req.Summary != nil,
			},
			Content: pgtype.Text{
				String: req.GetContent(),
				Valid:  req.Content != nil,
			},
			IsPublish: pgtype.Bool{
				Bool:  req.GetIsPublish(),
				Valid: req.IsPublish != nil,
			},
			CategoryID: pgtype.Int8{
				Int64: req.GetCategoryId(),
				Valid: req.CategoryId != nil,
			},
			UpdatedAt: pgtype.Timestamptz{
				Time:  time.Now(),
				Valid: true,
			},
		},
		AfterUpdate: func(article db.Article) error {
			// 同步文章的文件列表，确保不会存在冗余文件
			contentFileNames := util.ExtractFileNames(article.Content)

			resourcePath := fmt.Sprintf("%s/%s/%s", server.config.ResourcePath, "articles", article.ID.String())
			folderFiles, err := util.ListFiles(resourcePath)
			if err != nil {
				return err
			}

			for _, fileName := range folderFiles {
				if !slices.Contains(contentFileNames, fileName) {
					err := os.Remove(fmt.Sprintf("%s/%s", resourcePath, fileName))
					if err != nil {
						return err
					}
				}
			}

			return nil
		},
	}

	result, err := server.store.UpdateArticleTx(ctx, arg)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "article not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update article: %v", err)
	}

	resp := &pb.UpdateArticleResponse{
		Article: convertOnlyArticle(result.Article, false),
	}

	return resp, nil
}

func validateUpdateArticleRequest(req *pb.UpdateArticleRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if req.GetId() == "" {
		violations = append(violations, fieldViolation("id", fmt.Errorf("id is required")))
	}

	return
}
