package api

import (
	"errors"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

var AllCategoriesKey = "categories:all"

type listCategoriesResponse struct {
	Categories []db.ListCategoriesCountArticlesRow `json:"categories"`
	Count      int64                               `json:"count"`
}

func (server *Server) listCategories(ctx *gin.Context) {
	var resp listCategoriesResponse
	ok, err := server.cache.Get(ctx, key.CategoryAllKey, &resp)
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Error().
			Err(err).
			Str("key", key.CategoryAllKey).
			Str("module", "category").
			Str("action", "cache_get").
			Msg("获取分类缓存失败，降级为仅数据库")
	}
	if ok {
		ctx.JSON(http.StatusOK, resp)
		return
	}

	categories, err := server.store.ListCategoriesCountArticles(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	count := len(categories)

	resp.Categories = categories
	resp.Count = int64(count)

	err = server.cache.Set(ctx, key.CategoryAllKey, resp, 7*24*time.Hour)
	if err != nil {
		log.Error().
			Err(err).
			Str("key", key.CategoryAllKey).
			Str("module", "category").
			Str("action", "cache_set").
			Msg("设置分类缓存失败，降级为仅数据库")
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
