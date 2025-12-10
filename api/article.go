package api

import (
	"errors"
	"fmt"
	"github.com/go-ego/gse"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var segmenter gse.Segmenter

func init() {
	segmenter.LoadDict()
}

type createArticleRequest struct {
	Title     string `json:"title" binding:"required"`
	Summary   string `json:"summary"`
	Content   string `json:"content"`
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
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, article)
}

type getArticleRequest struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type getArticleResponse struct {
	Article db.GetArticleRow `json:"article"`
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
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if !article.IsPublish {
		ctx.JSON(http.StatusForbidden, errorResponse(fmt.Errorf("访问受限")))
		return
	}

	ctx.JSON(http.StatusOK, getArticleResponse{Article: article})
}

type getArticleForUpdateRequest struct {
	ID    uuid.UUID `uri:"id" binding:"required,uuid"`
	Owner uuid.UUID `uri:"owner" binding:"required,uuid"`
}

func (server *Server) getArticleForUpdate(ctx *gin.Context) {
	var req getArticleForUpdateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	/*articleID, err := uuid.Parse(req.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}*/

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	if authPayload.UserID != req.Owner {
		ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("this article doesn't belong to the authenticated user")))
		return
	}

	article, err := server.store.GetArticleForUpdate(ctx, req.ID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, article)
}

type listArticleRequest struct {
	CategoryID int64 `form:"category_id" binding:"omitempty"`
	Page       int32 `form:"page" binding:"required,min=1"`
	Limit      int32 `form:"limit" binding:"required,min=1,max=50"`
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

	countArg := db.CountArticlesParams{
		IsPublish: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
	}

	arg.CategoryID = pgtype.Int8{
		Int64: req.CategoryID,
		Valid: req.CategoryID != 0,
	}
	countArg.CategoryID = pgtype.Int8{
		Int64: req.CategoryID,
		Valid: req.CategoryID != 0,
	}

	articles, err := server.store.ListArticles(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	countArticles, err := server.store.CountArticles(ctx, countArg)
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
		AfterUpdate: func(article db.Article) error {
			// 同步文章的文件列表，确保不会存在冗余文件
			contentFileNames := util.ExtractFileNames(article.Content)

			resourcePath := fmt.Sprintf("./resources/%s", article.ID.String())
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

	article, err := server.store.UpdateArticleTx(ctx, arg)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, article)
}

type incrementArticleLikesRequest struct {
	ID uuid.UUID `json:"id" binding:"required,uuid"`
}

func (server *Server) incrementArticleLikes(ctx *gin.Context) {
	var req incrementArticleLikesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// 验证唯一标识（已登陆用 userID，为登陆用 IP+UA）
	var userKey string
	authPayload, exists := ctx.Get(authorizationPayloadKey)
	if exists {
		user := authPayload.(*token.Payload)
		userKey = fmt.Sprintf("uid:%s", user.UserID.String())
	} else {
		ip := ctx.ClientIP()
		ua := ctx.Request.UserAgent()
		userKey = fmt.Sprintf("guest:%s:%s", ip, ua)
	}

	redisKey := fmt.Sprintf("articles:likes:%s:%s", req.ID, userKey)

	set, err := server.redisService.SetNX(ctx, redisKey, 1, time.Hour*12)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if set {
		err = server.store.IncrementArticleLikes(ctx, req.ID)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, errorResponse(err))
				return
			}

			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	} else {
		// 如果已经处理过则告知请求冲突，不需要+1
		ctx.JSON(http.StatusConflict, nil)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

type incrementArticleViewsRequest struct {
	ID uuid.UUID `json:"id" binding:"required,uuid"`
}

func (server *Server) incrementArticleViews(ctx *gin.Context) {
	var req incrementArticleViewsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// 验证唯一标识（已登陆用 userID，为登陆用 IP+UA）
	var userKey string
	authPayload, exists := ctx.Get(authorizationPayloadKey)
	if exists {
		user := authPayload.(*token.Payload)
		userKey = fmt.Sprintf("uid:%s", user.UserID.String())
	} else {
		ip := ctx.ClientIP()
		ua := ctx.Request.UserAgent()
		userKey = fmt.Sprintf("guest:%s:%s", ip, ua)
	}

	redisKey := fmt.Sprintf("articles:views:%s:%s", req.ID, userKey)

	set, err := server.redisService.SetNX(ctx, redisKey, 1, time.Hour*12)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if set {
		err = server.store.IncrementArticleViews(ctx, req.ID)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, errorResponse(err))
				return
			}

			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	} else {
		ctx.JSON(http.StatusConflict, nil)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

type searchArticlesRequest struct {
	Keyword string `form:"keyword" binding:"required"`
	Page    int32  `form:"page" binding:"required,min=0"`
	Limit   int32  `form:"limit" binding:"required,min=1,max=20"`
}

type searchArticlesResponse struct {
	Articles []db.SearchArticlesRow `json:"articles"`
	Count    int64                  `json:"count"`
}

func (server *Server) searchArticle(ctx *gin.Context) {
	var req searchArticlesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	userInput := req.Keyword

	segments := segmenter.Cut(userInput, true)

	var cleanSegments []string
	for _, s := range segments {
		if len(s) > 0 && s != " " {
			cleanSegments = append(cleanSegments, s)
		}
	}
	keyword := strings.Join(cleanSegments, " OR ")
	if keyword == "" {
		keyword = req.Keyword
	}

	arg := db.SearchArticlesParams{
		Limit:   req.Limit,
		Offset:  (req.Page - 1) * req.Limit,
		Keyword: keyword,
		IsPublish: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
	}
	searchArticlesRows, err := server.store.SearchArticles(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	countArg := db.CountSearchArticlesParams{
		Keyword: keyword,
		IsPublish: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
	}
	countSearchArticles, err := server.store.CountSearchArticles(ctx, countArg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, searchArticlesResponse{
		Articles: searchArticlesRows,
		Count:    countSearchArticles,
	})
}
