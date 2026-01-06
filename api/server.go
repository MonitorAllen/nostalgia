package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/MonitorAllen/nostalgia/internal/cache"

	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/token"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/MonitorAllen/nostalgia/worker"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	fs2 "github.com/rakyll/statik/fs"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
)

// Server serves HTTP requests for our banking service.
type Server struct {
	server          *http.Server
	config          util.Config
	store           db.Store
	tokenMaker      token.Maker
	router          *gin.Engine
	taskDistributor worker.TaskDistributor
	cache           cache.Cache
}

// NewServer creates a new HTTPS server and setup routing
func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor, cache cache.Cache) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
		cache:           cache,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("currency", validCurrency)
		if err != nil {
			log.Fatal().Err(err).Msg("register validator error:")
		}
	}

	// register router
	server.setupRouter()

	return server, nil
}

// 注册路由
func (server *Server) setupRouter() {
	if server.config.Environment != "development" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// /temp/upload 静态资源不受中间件影响
	router.Static("/temp/upload", "./temp/upload")
	router.Static("/resources/", "./resources")

	public := router.Group("/api")
	{
		public.POST("/users", server.createUser)
		public.POST("/users/login", server.loginUser)
		public.POST("/tokens/renew_access", server.renewAccessToken)
		public.GET("/users/verify_email", server.verifyEmail)
		public.GET("/users/contributions", server.contributions)

		public.GET("/articles/:id", server.getArticle)
		public.GET("/articles/slug/:slug", server.getArticleBySlug)
		public.GET("/articles", server.listArticle)
		public.PATCH("/articles/increment_likes", server.incrementArticleLikes)
		public.PATCH("/articles/increment_views", server.incrementArticleViews)
		public.GET("/articles/search", server.searchArticle)

		public.GET("/comments/:article_id", server.listCommentsByArticleID)

		public.GET("/categories", server.listCategories)
		public.GET("/categories/:id", server.getCategory)
	}

	authRoutes := router.Group("/api").Use(authMiddleware(server.tokenMaker))
	{
		authRoutes.POST("/comments", server.createComment)
		authRoutes.DELETE("/comments/:id", server.deleteComment)

		authRoutes.POST("/upload_file/", server.uploadFile).Use(uploadFileMiddleware(server.config))
	}

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string, config util.Config) error {
	mux := http.NewServeMux()
	mux.Handle("/", server.router)

	statikFS, err := fs2.New()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create statik fs")
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	c := cors.New(cors.Options{
		AllowedOrigins: config.AllowedOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders: []string{
			"Authorization",
			"Content-Type",
		},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)

	srv := &http.Server{
		Addr:    address,
		Handler: handler,
	}

	server.server = srv
	return srv.ListenAndServe()
}

func (server *Server) Shutdown(ctx context.Context) error {
	// 先关闭 HTTP 服务器，确保不再接收新的请求
	if err := server.server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("failed to shutdown HTTP server")
		return err
	}

	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := server.cache.Close(); err != nil {
		log.Error().Err(err).Msg("failed to shutdown cache service")
	}

	return nil
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
