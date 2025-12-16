package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MonitorAllen/nostalgia/api"
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	_ "github.com/MonitorAllen/nostalgia/doc/statik"
	"github.com/MonitorAllen/nostalgia/gapi"
	"github.com/MonitorAllen/nostalgia/internal/service"
	"github.com/MonitorAllen/nostalgia/mail"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/MonitorAllen/nostalgia/util"
	"github.com/MonitorAllen/nostalgia/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	fs2 "github.com/rakyll/statik/fs"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal().Any("panic", r).Msg("unhandled panic")
		}
	}()

	// 加载配置
	config, err := util.LoadConfig("")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config:")
	}

	setupLogger(config)

	// 系统信号监听
	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	runDBMigration(config.MigrationURL, config.DBSource)

	connPool, err := pgxpool.New(ctx, config.DBSource)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db:")
	}

	store := db.NewStore(connPool)

	err = ensureDefaultUserExists(ctx, store, config)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create default user:")
	}

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	redisService := service.NewRedisService(config)

	waitGroup, ctx := errgroup.WithContext(ctx)

	runTaskProcessor(ctx, waitGroup, config, redisOpt, store)
	runGatewayServer(ctx, waitGroup, config, store, taskDistributor, redisService)
	runGrpcServer(ctx, waitGroup, config, store, taskDistributor, redisService)
	runGinServer(ctx, waitGroup, config, store, taskDistributor, redisService)

	err = waitGroup.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("error from wait group")
	}
}

func setupLogger(config util.Config) {
	// 开启调用者定位
	log.Logger = log.With().Caller().Logger()

	// 开发环境下，日志输出控制台
	if config.Environment == "development" {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out:        os.Stderr,
				TimeFormat: time.RFC3339,
			},
		)
	}
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance:")
	}

	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal().Err(err).Msg("failed to run migrate up:")
	}

	log.Info().Msg("db migrated successfully")
}

func runTaskProcessor(ctx context.Context, waitGroup *errgroup.Group, config util.Config, redisOpt asynq.RedisClientOpt, store db.Store) {
	mailer := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)
	log.Info().Msg("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown task processor")

		taskProcessor.Shutdown()
		log.Info().Msg("task processor is stopped")

		return nil
	})
}

func runGrpcServer(ctx context.Context, waitGroup *errgroup.Group, config util.Config, store db.Store, taskDistributor worker.TaskDistributor, redisService *service.RedisService) {
	server, err := gapi.NewServer(config, store, taskDistributor, redisService)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server:")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)

	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterNostalgiaServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	waitGroup.Go(func() error {
		log.Info().Msgf("start grpc server at %s", listener.Addr().String())
		err = grpcServer.Serve(listener)
		if err != nil {
			if errors.Is(err, grpc.ErrServerStopped) {
				return nil
			}
			log.Fatal().Err(err).Msg("grpc server failed to serve")
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()

		log.Info().Msg("graceful shutdown grpc server")

		grpcServer.GracefulStop()

		log.Info().Msg("grpc server stopped")

		return nil
	})
}

func runGatewayServer(ctx context.Context, waitGroup *errgroup.Group, config util.Config, store db.Store, taskDistributor worker.TaskDistributor, redisService *service.RedisService) {
	server, err := gapi.NewServer(config, store, taskDistributor, redisService)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server:")
	}

	grpcMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	err = pb.RegisterNostalgiaHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot register handler server: ")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs2.New()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create statik fs")
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	c := cors.New(cors.Options{
		AllowedOrigins: config.AllowedOrigins,
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Authorization",
			"Content-Type",
		},
		AllowCredentials: false,
	})
	handler := c.Handler(gapi.HttpLogger(mux))

	httpServer := &http.Server{
		Handler: handler,
		Addr:    config.GrpcGatewayAddress,
	}

	waitGroup.Go(func() error {
		log.Info().Msgf("start HTTP gateway server at %s", httpServer.Addr)
		err = httpServer.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			log.Error().Err(err).Msg("http server failed to serve")
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown http gateway server")
		err := httpServer.Shutdown(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("http server failed to shutdown")
			return err
		}
		log.Info().Msg("http gateway server stopped")
		return err
	})
}

func runGinServer(ctx context.Context, waitGroup *errgroup.Group, config util.Config, store db.Store, taskDistributor worker.TaskDistributor, redisService *service.RedisService) {
	server, err := api.NewServer(config, store, taskDistributor, redisService)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server:")
	}

	waitGroup.Go(func() error {
		log.Info().Msgf("start HTTP server at %s", config.HTTPServerAddress)
		err := server.Start(config.HTTPServerAddress, config)
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			log.Error().Err(err).Msg("HTTP server failed to serve")
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown HTTP server")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("graceful shutdown failed")
			return err
		}
		log.Info().Msg("All services is stopped")

		return nil
	})
}

func ensureDefaultUserExists(ctx context.Context, store db.Store, config util.Config) error {
	defaultUserID, err := uuid.Parse(config.DefaultUserID)
	if err != nil {
		return fmt.Errorf("failed to parse default user ID: %w", err)
	}

	_, err = store.GetUser(ctx, defaultUserID)
	if err == nil {
		log.Info().Msg("defult user already exists")
		return nil
	}

	if err != db.ErrRecordNotFound {
		return fmt.Errorf("failed to check default user: %w", err)
	}

	hashedPassword, err := util.HashPassword(config.DefaultUserPassword)
	if err != nil {
		return fmt.Errorf("failed to hash default user password: %w", err)
	}

	arg := db.CreateUserParams{
		ID:             defaultUserID,
		Username:       config.DefaultUsername,
		HashedPassword: hashedPassword,
		FullName:       config.DefaultUserFullname,
		Email:          config.DefaultUserEmail,
	}

	_, err = store.CreateUser(ctx, arg)
	if err != nil {
		return fmt.Errorf("failed to create default user: %w", err)
	}

	updateArg := db.UpdateUserParams{
		ID: arg.ID,
		IsEmailVerified: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
	}

	_, err = store.UpdateUser(ctx, updateArg)
	if err != nil {
		return fmt.Errorf("failed to active default user email: %w", err)
	}

	log.Info().Msgf("default user '%s' created", arg.Username)

	return nil
}
