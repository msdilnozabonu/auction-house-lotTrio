package main

import (
	"auction-house-lotTrio/internal/handler/auth"
	bidhandler "auction-house-lotTrio/internal/handler/bid"
	lots2 "auction-house-lotTrio/internal/handler/lots"
	userhandler "auction-house-lotTrio/internal/handler/user"
	"auction-house-lotTrio/internal/middleware"
	bidrepo "auction-house-lotTrio/internal/repository/bid"
	"auction-house-lotTrio/internal/repository/lots"
	"auction-house-lotTrio/internal/repository/session"
	"auction-house-lotTrio/internal/repository/user"
	"auction-house-lotTrio/internal/router"
	auth2 "auction-house-lotTrio/internal/service/auth"
	bidservice "auction-house-lotTrio/internal/service/bid"
	lots3 "auction-house-lotTrio/internal/service/lots"
	userservice "auction-house-lotTrio/internal/service/user"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "auction-house-lotTrio/internal/docs"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	shutdownTime = 5
	scheduler    = 5 * time.Minute
)

func getLoggerLevel(level string) slog.Level {
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// @title						AuctionHouse API
// @version					1.0
// @host						localhost:9999
// @BasePath					/api/v1
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	if err := run(); err != nil {
		slog.Error("run", "err", err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found", "err", err)
	}
	logger := newLogger()
	ctx := context.Background()
	pool := newPool(ctx)

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	engine, err := buildRouter(ctx, sigCtx, pool, logger)
	if err != nil {
		slog.Error("create router", "err", err)
		return fmt.Errorf("create router: %w", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "9999"
	}
	addr := ":" + port
	srv := &http.Server{
		Addr:              addr,
		Handler:           engine,
		ReadHeaderTimeout: shutdownTime * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	slog.Info("server started", "port", port)

	<-sigCtx.Done()
	slog.Info("shutdown signal received")

	shCtx, cancel := context.WithTimeout(context.Background(), shutdownTime*time.Second)
	defer cancel()

	if err := srv.Shutdown(shCtx); err != nil {
		slog.Error("shutdown", "err", err)
	} else {
		slog.Info("server stopped")
	}

	pool.Close()
	slog.Info("shutdown")
	return nil
}
func newLogger() *slog.Logger {
	logLevel := getLoggerLevel(os.Getenv("LOG_LEVEL"))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)
	return logger
}

func newPool(ctx context.Context) *pgxpool.Pool {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		slog.Error("DB_DSN is not set")
		os.Exit(1)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("connect database", "err", err)
		os.Exit(1)
	}
	return pool
}

func buildRouter(ctx, schedulerCtx context.Context, pool *pgxpool.Pool, logger *slog.Logger) (*gin.Engine, error) {
	userRepo, err := user.New(pool)
	if err != nil {
		return nil, fmt.Errorf("create user repository: %w", err)
	}

	userService := userservice.New(userRepo)
	userHandler := userhandler.New(userService, logger)

	sessionRepo := session.New(pool)
	authService := auth2.NewService(userRepo, sessionRepo)
	authHandler := auth.NewHandler(authService)

	mw := middleware.NewMiddleware(authService)

	lotsRepo, err := lots.New(pool)
	if err != nil {
		slog.Error("create lots repository", "err", err)
		os.Exit(1)
	}
	lotsService := lots3.NewService(lotsRepo)
	lotsHandler := lots2.NewHandler(lotsService)
	lotsService.StartScheduler(schedulerCtx, scheduler)

	bidRepo, err := bidrepo.New(pool)
	if err != nil {
		slog.Error("create bid repository: %w", "err", err)
		os.Exit(1)
	}
	bidService := bidservice.NewService(bidRepo)
	bidHandler := bidhandler.NewHandler(bidService)

	engine, err := router.New(ctx, pool, authHandler, userHandler, mw, lotsHandler, bidHandler)

	if err != nil {
		return nil, fmt.Errorf("create router: %w", err)
	}

	return engine, nil
}
