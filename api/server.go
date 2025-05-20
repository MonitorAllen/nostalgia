package api

import (
	"context"
	"fmt"
	"github.com/MonitorAllen/nostalgia/internal/service"
	"net/http"
	"time"

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
	redisService    service.Redis
}

// NewServer creates a new HTTPS server and setup routing
func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor, redisService *service.RedisService) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
		redisService:    redisService,
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

	router.POST("/api/users", server.createUser)
	router.POST("/api/users/login", server.loginUser)
	router.POST("/api/tokens/renew_access", server.renewAccessToken)
	router.GET("/api/users/verify_email", server.verifyEmail)
	router.GET("/api/users/contributions", server.contributions)

	router.GET("/api/articles/:id", server.getArticle)
	router.GET("/api/articles", server.listArticle)

	router.GET("/api/comments/:article_id", server.listCommentsByArticleID)

	authRoutes := router.Group("/api/").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/articles", server.createArticle)
	authRoutes.PUT("/articles", server.updateArticle)

	authRoutes.POST("/comments", server.createComment)
	authRoutes.DELETE("/comments/:id", server.deleteComment)

	authRoutes.POST("/upload/:id", server.uploadFile).Use(uploadFileMiddleware(server.config))

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
	//return server.router.Run(address)
}

func (server *Server) Shutdown(ctx context.Context) error {
	// 先关闭 HTTP 服务器，确保不再接收新的请求
	if err := server.server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("failed to shutdown HTTP server")
		return err
	}

	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := server.redisService.Close(); err != nil {
		log.Error().Err(err).Msg("failed to shutdown redis service")
	}

	return nil
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
