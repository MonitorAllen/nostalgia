package gapi

import (
	"context"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateArticle(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleResponse, error) {
	_, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	title := "新建文章"
	if req.Title != nil {
		title = req.GetTitle()
	}

	isPublish := false
	if req.IsPublish != nil {
		isPublish = req.GetIsPublish()
	}

	aritcleID, err := uuid.NewRandom()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "生成文章ID失败: %v", err)
	}

	defaultUserID, err := uuid.Parse(server.config.DefaultUserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse default user ID: %v", err)
	}

	defaultCover := "/images/go.png"

	arg := db.CreateArticleParams{
		ID:         aritcleID,
		Title:      title,
		Summary:    req.GetSummary(),
		Content:    req.GetContent(),
		IsPublish:  isPublish,
		Owner:      defaultUserID,
		CategoryID: 1,
		Cover:      defaultCover,
	}

	article, err := server.store.CreateArticle(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create article: %v", err)
	}

	resp := &pb.CreateArticleResponse{
		Article: convertOnlyArticle(article, false),
	}

	return resp, nil
}
