package router

import (
	"auction-house-lotTrio/internal/handler/auth"
	"auction-house-lotTrio/internal/handler/bid"
	lots2 "auction-house-lotTrio/internal/handler/lots"
	"auction-house-lotTrio/internal/handler/user"
	"auction-house-lotTrio/internal/handler/watchlist"
	"auction-house-lotTrio/internal/handler/wins"
	"auction-house-lotTrio/internal/middleware"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// nolint:funlen
func New(ctx context.Context, pool *pgxpool.Pool,
	authHandler auth.Handler, userHandler user.Handler,
	newMiddleware middleware.Middleware,
	lotHandler lots2.Handler, bidHandler bid.Handler, winsHandler wins.Handler,
	watchlistHandler watchlist.Handler) (*gin.Engine, error) {
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
		adminGroup.POST("/close-expired", lotHandler.CloseExpiredLot)
		adminGroup.GET("/lots", lotHandler.GetLotsForAdmin)
		adminGroup.PUT("/lots/:id/moderate", lotHandler.Moderate)
		adminGroup.GET("/stats", lotHandler.GetPlatformStats)
		adminGroup.GET("/export", lotHandler.ExportLots)
	}

	engine.Static("/uploads", "./uploads")

	lotsGroup := api.Group("/lots")
	lotsGroup.Use(newMiddleware.Auth())
	{
		lotsGroup.POST("/new", lotHandler.CreateLot)
		lotsGroup.GET("", lotHandler.GetAll)
		lotsGroup.GET("/:id", lotHandler.GetByID)
		lotsGroup.PUT("/:id", lotHandler.UpdateByID)
		lotsGroup.PUT("/:id/status", lotHandler.UpdateStatusByID)
		lotsGroup.DELETE("/:id", lotHandler.DeleteLots)
		lotsGroup.POST("/:id/photo", lotHandler.UploadPhoto)
		lotsGroup.GET("/:id/photo", lotHandler.GetPhoto)
		lotsGroup.GET("/mine", newMiddleware.Auth(), lotHandler.GetMine)

		lotsGroup.POST("/:id/bid", middleware.RequireRole("bidder"), bidHandler.PlaceBid)
		lotsGroup.GET("/:id/bids", bidHandler.GetBidsByLotID)

		lotsGroup.POST("/:id/watch", middleware.RequireRole("bidder"), watchlistHandler.Add)
		lotsGroup.DELETE("/:id/watch", middleware.RequireRole("bidder"), watchlistHandler.Delete)
	}

	api.GET("/bids", newMiddleware.Auth(), middleware.RequireRole("bidder"), bidHandler.GetBiddersBids)
	api.GET("/wins", newMiddleware.Auth(), middleware.RequireRole("bidder"), winsHandler.GetWins)
	api.GET("/watchlist", newMiddleware.Auth(), middleware.RequireRole("bidder"), watchlistHandler.GetWatchlist)

	return engine, nil
}
