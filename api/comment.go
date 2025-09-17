package api

import (
	"database/sql"
	"errors"
	"fmt"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type createCommentRequest struct {
	ArticleID  uuid.UUID `json:"article_id" binding:"required"`
	Content    string    `json:"content" binding:"required,min=1"`
	ParentID   int64     `json:"parent_id"`
	FromUserID uuid.UUID `json:"from_user_id" binding:"required"`
	ToUserID   uuid.UUID `json:"to_user_id" binding:"required"`
}

type createCommentResponse struct {
	Comment Comment `json:"comment"`
}

func (server *Server) createComment(ctx *gin.Context) {
	var req createCommentRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if req.Content == "" {
		ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("content cant't be empty")))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	if req.FromUserID != authPayload.UserID {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("unauthorized user")))
		return
	}

	arg := db.CreateCommentParams{
		Content:    req.Content,
		ArticleID:  req.ArticleID,
		ParentID:   req.ParentID,
		FromUserID: authPayload.UserID,
		ToUserID:   req.ToUserID,
	}

	comment, err := server.store.CreateComment(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	toUser, err := server.store.GetUser(ctx, comment.ToUserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := createCommentResponse{Comment: Comment{
		ID:           comment.ID,
		Content:      comment.Content,
		ArticleID:    comment.ArticleID,
		ParentID:     comment.ParentID,
		Likes:        comment.Likes,
		FromUserID:   comment.FromUserID,
		ToUserID:     comment.ToUserID,
		CreatedAt:    comment.CreatedAt,
		DeletedAt:    comment.DeletedAt,
		FromUserName: authPayload.Username,
		ToUserName:   toUser.Username,
		Child:        []Comment{},
	}}

	ctx.JSON(http.StatusOK, resp)
}

type listCommentsByArticleIDRequest struct {
	ArticleID string `uri:"article_id" binding:"required,uuid"`
}

type Comment struct {
	ID           int64     `json:"id"`
	Content      string    `json:"content"`
	ArticleID    uuid.UUID `json:"article_id"`
	ParentID     int64     `json:"parent_id"`
	Likes        int32     `json:"likes"`
	FromUserID   uuid.UUID `json:"from_user_id"`
	ToUserID     uuid.UUID `json:"to_user_id"`
	CreatedAt    time.Time `json:"created_at"`
	DeletedAt    time.Time `json:"deleted_at"`
	FromUserName string    `json:"from_user_name"`
	ToUserName   string    `json:"to_user_name"`
	Child        []Comment `json:"child"`
}

type listCommentsByArticleIDResponse struct {
	Comments []Comment `json:"comments"`
}

func (server *Server) listCommentsByArticleID(ctx *gin.Context) {
	var req listCommentsByArticleIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	articleID, err := uuid.Parse(req.ArticleID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	comments, err := server.store.ListCommentsByArticleID(ctx, articleID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	commentTree := buildCommentTree(comments)

	ctx.JSON(http.StatusOK, listCommentsByArticleIDResponse{Comments: commentTree})
}

func buildCommentTree(rows []db.ListCommentsByArticleIDRow) []Comment {
	// 用来存储 id -> Comment 的映射
	commentMap := make(map[int64]Comment)

	// 初始化映射
	// 转换原始行数据为 Comment 结构
	for _, row := range rows {
		commentMap[row.ID] = Comment{
			ID:           row.ID,
			Content:      row.Content,
			ArticleID:    row.ArticleID,
			ParentID:     row.ParentID,
			Likes:        row.Likes,
			FromUserID:   row.FromUserID,
			ToUserID:     row.ToUserID,
			CreatedAt:    row.CreatedAt,
			DeletedAt:    row.DeletedAt,
			FromUserName: row.FromUserName.String,
			ToUserName:   row.ToUserName.String,
			Child:        []Comment{},
		}
	}

	// 2. 构造树形结构
	var rootComments []Comment
	for i := range rows {
		comment := rows[i]
		if comment.ParentID == 0 {
			// 根评论
			rootComments = append(rootComments, commentMap[comment.ID])
		} else if parent, exists := commentMap[int64(comment.ParentID)]; exists {
			// 如果父评论存在，将当前评论添加到父评论的 Child 列表中
			parent.Child = append(parent.Child, commentMap[comment.ID])
		}
	}

	return rootComments
}

type deleteCommentRequest struct {
	ID int64 `uri:"id" binding:"required"`
}

func (server *Server) deleteComment(ctx *gin.Context) {
	var req deleteCommentRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}

	comment, err := server.store.GetComment(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	if comment.FromUserID != authPayload.UserID {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("unauthorized user")))
		return
	}

	// 删除子评论
	err = server.store.DeleteChildComments(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, fmt.Errorf("delete child comment: %w", err))
		return
	}

	err = server.store.DeleteComment(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, fmt.Errorf("delete comment: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, nil)
}
