package main

import (
	"auction-house-lotTrio/internal/handler/auth"

	lots3 "auction-house-lotTrio/internal/handler/lots"
	userhandler "auction-house-lotTrio/internal/handler/user"
	"auction-house-lotTrio/internal/middleware"

	lots2 "auction-house-lotTrio/internal/repository/lots"
	"auction-house-lotTrio/internal/repository/session"
	"auction-house-lotTrio/internal/repository/user"
	"auction-house-lotTrio/internal/router"
	auth2 "auction-house-lotTrio/internal/service/auth"
	"auction-house-lotTrio/internal/service/lots"
	userservice "auction-house-lotTrio/internal/service/user"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "auction-house-lotTrio/internal/docs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const shutdownTime = 5

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
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found", "err", err)
	}
	logLevel := getLoggerLevel(os.Getenv("LOG_LEVEL"))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	ctx := context.Background()

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

	userRepo, err := user.New(pool)
	if err != nil {
		slog.Error("create user repository", "err", err)
		os.Exit(1)
	}

	lotsRepo, err := lots2.New(pool)
	if err != nil {
		slog.Error("create lots repository", "err", err)
		os.Exit(1)
	}

	userService := userservice.New(userRepo)
	userHandler := userhandler.New(userService, logger)

	sessionRepo := session.New(pool)
	authService := auth2.NewService(userRepo, sessionRepo)
	authHandler := auth.NewHandler(authService)

	newMiddleware := middleware.NewMiddleware(authService)

	lotsService := lots.NewService(lotsRepo)
	lotsHandler := lots3.NewHandler(lotsService)

	lotsService.StartScheduler(ctx, 5*time.Minute)

	engine, err := router.New(ctx, pool, authHandler, userHandler, newMiddleware, lotsHandler)

	if err != nil {
		slog.Error("create router", "err", err)
		os.Exit(1)
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

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
}
