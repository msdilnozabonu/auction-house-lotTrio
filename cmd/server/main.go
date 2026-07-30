package main

import (
	"auction-house-lotTrio/internal/handler/auth"
	bidhandler "auction-house-lotTrio/internal/handler/bid"
	lots2 "auction-house-lotTrio/internal/handler/lots"
	"auction-house-lotTrio/internal/handler/seller"
	userhandler "auction-house-lotTrio/internal/handler/user"
	watchlisthandler "auction-house-lotTrio/internal/handler/watchlist"
	winshandler "auction-house-lotTrio/internal/handler/wins"
	"auction-house-lotTrio/internal/middleware"
	bidrepo "auction-house-lotTrio/internal/repository/bid"
	"auction-house-lotTrio/internal/repository/lots"
	seller2 "auction-house-lotTrio/internal/repository/seller"
	"auction-house-lotTrio/internal/repository/session"
	"auction-house-lotTrio/internal/repository/user"
	"auction-house-lotTrio/internal/repository/watchlist"
	winsrepo "auction-house-lotTrio/internal/repository/wins"
	"auction-house-lotTrio/internal/router"
	auth2 "auction-house-lotTrio/internal/service/auth"
	bidservice "auction-house-lotTrio/internal/service/bid"
	lots3 "auction-house-lotTrio/internal/service/lots"
	seller3 "auction-house-lotTrio/internal/service/seller"
	userservice "auction-house-lotTrio/internal/service/user"
	watchlistservice "auction-house-lotTrio/internal/service/watchlist"
	winsservice "auction-house-lotTrio/internal/service/wins"
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
	logger := newLogger()
	if err := godotenv.Load(); err != nil {
		logger.Warn(".env file not found", "err", err)
	}
	ctx := context.Background()
	pool := newPool(ctx)

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	engine, err := buildRouter(ctx, sigCtx, pool, logger)
	if err != nil {
		logger.Error("create router", "err", err)
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
			logger.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	logger.Info("server started", "port", port)

	<-sigCtx.Done()
	logger.Info("shutdown signal received")

	shCtx, cancel := context.WithTimeout(context.Background(), shutdownTime*time.Second)
	defer cancel()

	if err := srv.Shutdown(shCtx); err != nil {
		logger.Error("shutdown", "err", err)
	} else {
		logger.Info("server stopped")
	}

	pool.Close()
	logger.Info("shutdown")
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

//nolint:funlen
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
	bidService := bidservice.NewService(bidRepo, lotsRepo)
	bidHandler := bidhandler.NewHandler(bidService)

	winsRepo, err := winsrepo.New(pool)
	if err != nil {
		slog.Error("create wins repository", "err", err)
		os.Exit(1)
	}
	winsService := winsservice.NewService(winsRepo)
	winsHandler := winshandler.NewHandler(winsService)

	watchlistRepo, err := watchlist.New(pool)
	if err != nil {
		slog.Error("create watchlist repository", "err", err)
		os.Exit(1)
	}
	watchlistService := watchlistservice.NewService(watchlistRepo)
	watchlistHandler := watchlisthandler.NewHandler(watchlistService)

	sellerStatsRepo, err := seller2.New(pool)
	if err != nil {
		slog.Error("create seller stats repository", "err", err)
		os.Exit(1)
	}
	sellerStatsService := seller3.NewService(sellerStatsRepo, bidRepo, lotsRepo)
	sellerHandle := seller.NewHandler(sellerStatsService)

	engine, err := router.New(ctx, pool, authHandler, userHandler, mw, lotsHandler, bidHandler, winsHandler,
		watchlistHandler, sellerHandle)

	if err != nil {
		return nil, fmt.Errorf("create router: %w", err)
	}

	return engine, nil
}
