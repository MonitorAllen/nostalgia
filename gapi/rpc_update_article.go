package gapi

import (
	"context"
	"errors"
	"fmt"
	"github.com/MonitorAllen/nostalgia/internal/cache"
	"math"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

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

	dbArticle, err := server.store.GetArticle(ctx, articleId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "文章不存在")
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
			Cover: pgtype.Text{
				String: req.GetCover(),
				Valid:  req.Cover != nil,
			},
			Slug: pgtype.Text{
				String: req.GetSlug(),
				Valid:  *req.Slug != "",
			},
			CheckOutdated: pgtype.Bool{
				Bool:  req.GetCheckOutdated(),
				Valid: req.CheckOutdated != nil,
			},
			LastUpdated: pgtype.Timestamptz{
				Time:  time.Now(),
				Valid: true,
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
				if strings.Split(fileName, ".")[0] == "cover" {
					continue
				}
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

	arg.ReadTime = pgtype.Text{
		String: calculateReadTime(req.GetContent()),
		Valid:  true,
	}

	result, err := server.store.UpdateArticleTx(ctx, arg)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "article not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update article: %v", err)
	}

	// 删除缓存
	_ = server.cache.Del(ctx, cache.GetArticleIDKey(dbArticle.ID))
	if &dbArticle.Slug != nil {
		_ = server.cache.Del(ctx, cache.GetArticleSlugKey(dbArticle.Slug.String))
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

func calculateReadTime(htmlContent string) string {
	// 1. 去除 HTML 标签
	re := regexp.MustCompile(`<[^>]*>`)
	plainText := re.ReplaceAllString(htmlContent, "")

	// 2. 计算字数 (按 Rune 计算，支持中文)
	wordCount := utf8.RuneCountInString(plainText)

	// 3. 阅读速度：中文约 400 字/分钟，代码/英文混合可适当调整
	readSpeed := 400.0
	minutes := math.Ceil(float64(wordCount) / readSpeed)

	if minutes < 1 {
		return "1 分钟"
	}
	return fmt.Sprintf("%.0f 分钟", minutes)
}
