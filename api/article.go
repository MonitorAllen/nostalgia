package api

import (
	"fmt"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"os"
)

type createArticleRequest struct {
	Title     string `json:"title" binding:"required"`
	Summary   string `json:"summary" binding:"required"`
	Content   string `json:"content" binding:"required"`
	IsPublish bool   `json:"is_publish"`
}

func (server *Server) createArticle(ctx *gin.Context) {
	var req createArticleRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	articleID, err := uuid.NewRandom()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateArticleParams{
		ID:        articleID,
		Title:     req.Title,
		Summary:   req.Summary,
		Content:   req.Content,
		IsPublish: req.IsPublish,
		Owner:     authPayload.UserID,
	}

	article, err := server.store.CreateArticle(ctx, arg)
	if err != nil {
		code, _ := db.ErrorCode(err)
		errCode := code
		switch errCode {
		case db.ForeignKeyViolation, db.UniqueViolation:
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, article)
}

type getArticleRequest struct {
	ID string `uri:"id" binding:"required,uuid"`
}

func (server *Server) getArticle(ctx *gin.Context) {
	var req getArticleRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	articleID, err := uuid.Parse(req.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	article, err := server.store.GetArticle(ctx, articleID)
	if err != nil {
		code, _ := db.ErrorCode(err)
		errCode := code
		switch errCode {
		case db.ForeignKeyViolation, db.UniqueViolation:
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, article)
}

type listArticleRequest struct {
	Page  int32 `form:"page" binding:"required,min=1"`
	Limit int32 `form:"limit" binding:"required,min=1,max=20"`
}

type listArticleResponse struct {
	Count    int64                `json:"count"`
	Articles []db.ListArticlesRow `json:"articles"`
}

func (server *Server) listArticle(ctx *gin.Context) {
	var req listArticleRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListArticlesParams{
		Limit:  req.Limit,
		Offset: (req.Page - 1) * req.Limit,
		IsPublish: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
	}

	articles, err := server.store.ListArticles(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	countArticles, err := server.store.CountArticles(ctx, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := listArticleResponse{
		Count:    countArticles,
		Articles: articles,
	}

	ctx.JSON(http.StatusOK, resp)
}

type updateArticleRequest struct {
	ID            uuid.UUID `json:"id" binding:"required"`
	Title         string    `json:"title"`
	Summary       string    `json:"summary"`
	Content       string    `json:"content"`
	IsPublish     *bool     `json:"is_publish"`
	Owner         uuid.UUID `json:"owner" binding:"required"`
	NeedSaveFiles []string  `json:"need_save_files"`
}

func (server *Server) updateArticle(ctx *gin.Context) {
	var req updateArticleRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	if authPayload.UserID != req.Owner {
		ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("this article doesn't belong to the authenticated user")))
		return
	}

	arg := db.UpdateArticleTxParams{
		UpdateArticleParams: db.UpdateArticleParams{
			ID: req.ID,
			Title: pgtype.Text{
				String: req.Title,
				Valid:  req.Title != "",
			},
			Summary: pgtype.Text{
				String: req.Summary,
				Valid:  req.Summary != "",
			},
			Content: pgtype.Text{
				String: req.Content,
				Valid:  req.Content != "",
			},
			IsPublish: pgtype.Bool{
				Bool:  *req.IsPublish,
				Valid: req.IsPublish != nil,
			},
		},
		AfterUpdate: func(articleID uuid.UUID, article db.Article, needSaveFiles []string) error {
			if len(needSaveFiles) == 0 {
				return nil
			}

			return nil
		},
	}

	article, err := server.store.UpdateArticleTx(ctx, arg, req.NeedSaveFiles)
	if err != nil {
		code, _ := db.ErrorCode(err)
		errCode := code
		switch errCode {
		case db.ForeignKeyViolation, db.UniqueViolation:
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, article)
}

func saveArticleFiles(articleID uuid.UUID, files []string) error {
	// 指定保存文件的目录
	dstDir := fmt.Sprintf("./public/resource/article/%s", articleID.String())

	// 创建目标目录（如果不存在）
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		err := os.MkdirAll(dstDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	err := util.DownloadFiles(files, dstDir)
	if err != nil {
		return err
	}

	return nil
}
