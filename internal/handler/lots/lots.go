package lots

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/response"
	"auction-house-lotTrio/internal/service/lots"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type LotHandler interface {
	CreateLot(c *gin.Context)
	GetAll(c *gin.Context)
}

type lotsHandler struct {
	lotsService lots.LotService
}

func NewLotHandler(lotsService lots.LotService) LotHandler {
	return &lotsHandler{lotsService: lotsService}
}

func (l *lotsHandler) CreateLot(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.RespondError(c, model.ErrUnauthorized)
		return
	}
	sellerId, ok := userID.(int64)
	if !ok {
		response.RespondError(c, model.ErrForbidden)
		return
	}

	var req lotsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondError(c, err)
		return
	}
	err := l.lotsService.CreateLot(c.Request.Context(), req.Title, req.Description, req.StartPrice,
		req.Photo, req.EndsAt, req.Status, sellerId)
	if err != nil {
		response.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Lot created successfully!"})
}

func (l *lotsHandler) GetAll(c *gin.Context) {
	items, err := l.lotsService.GetAll(c.Request.Context())
	if err != nil {
		slog.Error("get lots repository", "err", err)
		response.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

type lotsRequest struct {
	Title       string    `binding:"required" json:"title"`
	Description string    `binding:"required" json:"description"`
	StartPrice  float64   `binding:"required" json:"startPrice"`
	EndsAt      time.Time `binding:"required" json:"endsAt"`
	Photo       string    `binding:"required" json:"photo"`
	Status      string    `binding:"required" json:"status"`
}
