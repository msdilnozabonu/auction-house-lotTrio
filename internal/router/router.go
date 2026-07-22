package router

import (
	"auction-house-lotTrio/internal/handler/auth"
	"auction-house-lotTrio/internal/handler/lots"
	"auction-house-lotTrio/internal/handler/user"
	"auction-house-lotTrio/internal/middleware"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func New(ctx context.Context, pool *pgxpool.Pool,
	authHandler auth.Handler, userHandler user.Handler,
	newMiddleware middleware.Middleware,
	lotsHandler lots.Handler) (*gin.Engine, error) {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	api := engine.Group("/api/v1")

	api.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	api.GET("/health", func(c *gin.Context) {
		if err := pool.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db unreachable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api.GET("/me", newMiddleware.Auth(), userHandler.Me)
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Registration)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.POST("/logout", newMiddleware.Auth(), authHandler.Logout)
	}

	adminGroup := api.Group("/admin")
	adminGroup.Use(newMiddleware.Auth(), middleware.RequireRole("admin"))
	{
		adminGroup.POST("/close-expired", lotsHandler.CloseExpiredLot)
	}

	lotsGroup := api.Group("/lots")
	{
		lotsGroup.POST("/new", newMiddleware.Auth(), lotsHandler.CreateLot)
		lotsGroup.GET("", lotsHandler.GetAll)
	}
	return engine, nil
}
