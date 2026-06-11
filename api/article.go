package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	cachepkg "github.com/MonitorAllen/nostalgia/internal/cache"
	"github.com/MonitorAllen/nostalgia/internal/cache/key"
	"github.com/go-ego/gse"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var segmenter gse.Segmenter

var errArticleAccessRestricted = errors.New("访问受限")

func init() {
	segmenter.LoadDict()
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

	cacheKey := key.GetArticleIDKey(articleID)
	articleCache := cachepkg.NewArticleCache(server.cache)

	article, ok, err := articleCache.GetByID(ctx, articleID)
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Error().
			Err(err).
			Str("key", cacheKey).
			Str("module", "article").
			Str("action", "cache_get").
			Str("article_id", req.ID).
			Msg("根据 ID 获取文章缓存失败，降级为仅数据库")
	}

	if ok {
		ctx.JSON(http.StatusOK, getArticleResponse{article})
		return
	}

	value, err, _ := server.cacheLoadGroup.Do(cacheKey, func() (any, error) {
		article, err := server.store.GetArticle(ctx, articleID)
		if err != nil {
			return db.GetArticleRow{}, err
		}

		if !article.IsPublish {
			return db.GetArticleRow{}, errArticleAccessRestricted
		}

		if err := articleCache.SetByID(ctx, article); err != nil {
			log.Error().
				Err(err).
				Str("key", cacheKey).
				Str("module", "article").
				Str("action", "cache_set").
				Str("article_id", req.ID).
				Msg("根据 ID 设置文章缓存失败")
		}

		return article, nil
	})
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		if errors.Is(err, errArticleAccessRestricted) {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	article, ok = value.(db.GetArticleRow)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("unexpected article cache load result")))
		return
	}

	ctx.JSON(http.StatusOK, getArticleResponse{Article: article})
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

	articleCache := cachepkg.NewArticleCache(server.cache)
	cacheParams := cachepkg.ArticleListParams{
		CategoryID: req.CategoryID,
		Page:       req.Page,
		Limit:      req.Limit,
	}
	cachedPage, ok, err := articleCache.GetList(ctx, cacheParams)
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Error().
			Err(err).
			Str("module", "article").
			Str("action", "cache_get").
			Str("cache_namespace", "article_list").
			Msg("获取文章分页缓存失败，降级为仅数据库")
	}
	if ok {
		ctx.JSON(http.StatusOK, listArticleResponse{
			Count:    cachedPage.Count,
			Articles: cachedPage.Articles,
		})
		return
	}

	groupKey := fmt.Sprintf("cache:article:list:%d:%d:%d", req.CategoryID, req.Page, req.Limit)
	value, err, _ := server.cacheLoadGroup.Do(groupKey, func() (any, error) {
		articles, err := server.store.ListArticles(ctx, arg)
		if err != nil {
			return listArticleResponse{}, err
		}

		countArticles, err := server.store.CountArticles(ctx, countArg)
		if err != nil {
			return listArticleResponse{}, err
		}

		resp := listArticleResponse{
			Count:    countArticles,
			Articles: articles,
		}

		if err := articleCache.SetList(ctx, cacheParams, cachepkg.ArticleListPage(resp)); err != nil {
			log.Error().
				Err(err).
				Str("module", "article").
				Str("action", "cache_set").
				Str("cache_namespace", "article_list").
				Msg("设置文章分页缓存失败，降级为仅数据库")
		}

		return resp, nil
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp, ok := value.(listArticleResponse)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("unexpected article list cache load result")))
		return
	}

	ctx.JSON(http.StatusOK, resp)
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

	// 验证唯一标识（已登陆用 userID，未登陆用 IP）
	idempotencyCache := cachepkg.NewIdempotencyCache(server.cache)
	authPayload, exists := ctx.Get(authorizationPayloadKey)
	var ok bool
	var err error
	if exists {
		user := authPayload.(*token.Payload)
		ok, err = idempotencyCache.MarkArticleLikeByUser(ctx, req.ID, user.UserID)
	} else {
		ip := ctx.ClientIP()
		ok, err = idempotencyCache.MarkArticleLikeByGuest(ctx, req.ID, ip)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if !ok {
		// 如果已经处理过则告知请求冲突，不需要+1
		ctx.JSON(http.StatusConflict, nil)
		return
	}

	if err = server.store.IncrementArticleLikes(ctx, req.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, nil)
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

	// 验证唯一标识（已登陆用 userID，未登陆用 IP）
	idempotencyCache := cachepkg.NewIdempotencyCache(server.cache)
	authPayload, exists := ctx.Get(authorizationPayloadKey)
	var ok bool
	var err error
	if exists {
		user := authPayload.(*token.Payload)
		ok, err = idempotencyCache.MarkArticleViewByUser(ctx, req.ID, user.UserID)
	} else {
		ip := ctx.ClientIP()
		ok, err = idempotencyCache.MarkArticleViewByGuest(ctx, req.ID, ip)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if !ok {
		ctx.JSON(http.StatusConflict, nil)
		return
	}

	if err = server.store.IncrementArticleViews(ctx, req.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, nil)
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

type getArticleBySlugRequest struct {
	Slug string `uri:"slug" binding:"required,min=5"`
}

type getArticleBySlugResponse struct {
	Article db.GetArticleBySlugRow `json:"article"`
}

func (server *Server) getArticleBySlug(ctx *gin.Context) {
	var req getArticleBySlugRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	cacheKey := key.GetArticleSlugKey(req.Slug)
	articleCache := cachepkg.NewArticleCache(server.cache)

	article, ok, err := articleCache.GetBySlug(ctx, req.Slug)
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Error().
			Err(err).
			Str("key", cacheKey).
			Str("module", "article").
			Str("action", "cache_get").
			Str("article_slug", req.Slug).
			Msg("获取文章缓存失败，降级为仅数据库")
	}

	if ok {
		ctx.JSON(http.StatusOK, getArticleBySlugResponse{article})
		return
	}

	value, err, _ := server.cacheLoadGroup.Do(cacheKey, func() (any, error) {
		article, err := server.store.GetArticleBySlug(ctx, pgtype.Text{
			String: req.Slug,
			Valid:  true,
		})
		if err != nil {
			return db.GetArticleBySlugRow{}, err
		}

		if !article.IsPublish {
			return db.GetArticleBySlugRow{}, errArticleAccessRestricted
		}

		if err := articleCache.SetBySlug(ctx, req.Slug, article); err != nil {
			log.Error().
				Err(err).                      // 1. 记录错误堆栈
				Str("module", "article").      // 2. 模块：文章模块
				Str("action", "cache_set").    // 3. 动作：写入缓存
				Str("key", cacheKey).          // 4. 上下文：具体的 Redis Key
				Str("article_slug", req.Slug). // 5. 业务ID：关联的文章ID
				Msg("未能设置文章缓存，降级为仅数据库")        // 6. 消息：简明扼要
		}

		return article, nil
	})
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		if errors.Is(err, errArticleAccessRestricted) {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	article, ok = value.(db.GetArticleBySlugRow)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("unexpected article slug cache load result")))
		return
	}

	ctx.JSON(http.StatusOK, getArticleBySlugResponse{Article: article})
}
