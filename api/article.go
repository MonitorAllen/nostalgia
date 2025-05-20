package api

import (
	"fmt"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
)

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

	for i := range articles {
		if articles[i].Tags[0] == "" {
			articles[i].Tags = []string{}
		}
	}

	countArticles, err := server.store.CountArticles(ctx, pgtype.Bool{Bool: true, Valid: true})
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
			contentFileNames := extractFileNames(article.Content)

			resourcePath := fmt.Sprintf("./resources/%s", article.ID.String())
			folderFiles, err := listFiles(resourcePath)
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

func listFiles(dirPath string) ([]string, error) {
	// 读取目录中的所有文件和子目录
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, entry := range entries {
		if !entry.IsDir() { // 判断是否为文件
			fileNames = append(fileNames, entry.Name())
		}
	}

	return fileNames, nil
}

func extractFileNames(content string) []string {
	// 定义一个正则表达式，匹配 URL 的基本结构
	urlRegex := `https?://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]`
	re := regexp.MustCompile(urlRegex)

	// 找出文章中的所有 URL
	matches := re.FindAllString(content, -1)

	var fileNames []string
	for _, url := range matches {
		// 从 URL 中提取文件名
		segments := strings.Split(url, "/")
		fileName := segments[len(segments)-1]

		// 排除没有文件名的情况（例如 URL 以 / 结尾）
		if fileName != "" {
			fileNames = append(fileNames, fileName)
		}
	}

	return fileNames
}
