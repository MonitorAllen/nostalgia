package api

import (
	"errors"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/gin-gonic/gin"
	"net/http"
)

var AllCategoriesKey = "categories:all"

type listCategoriesRequest struct {
	Limit int32 `form:"limit" binding:"required"`
	Page  int32 `form:"page" binding:"required"`
}

type listCategoriesResponse struct {
	Categories []db.ListCategoriesCountArticlesRow `json:"categories"`
	Count      int64                               `json:"count"`
}

func (server *Server) listCategories(ctx *gin.Context) {
	var req listCategoriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListCategoriesCountArticlesParams{
		Limit:  req.Limit,
		Offset: (req.Page - 1) * req.Limit,
	}

	categories, err := server.store.ListCategoriesCountArticles(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	count, err := server.store.CountCategories(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := listCategoriesResponse{
		categories,
		count,
	}

	ctx.JSON(http.StatusOK, resp)
}

type getCategoryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type getCategoryResponse struct {
	Category db.Category `json:"category"`
}

func (server *Server) getCategory(ctx *gin.Context) {
	var req getCategoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	category, err := server.store.GetCategory(ctx, req.ID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := getCategoryResponse{category}
	ctx.JSON(http.StatusOK, resp)
}
