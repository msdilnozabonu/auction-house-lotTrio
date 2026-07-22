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


type Handler interface {
	CloseExpiredLot(c *gin.Context)
	CreateLot(c *gin.Context)
	GetAll(c *gin.Context)
}

type handler struct {
	lotsService lots.Service
	logger      *slog.Logger
}

func NewHandler(lotsService lots.Service) Handler {
	return &handler{
		lotsService: lotsService,
		logger:      slog.With("module", "lots")}
}

// CloseExpiredLot godoc
// @Summary      Закрыть просроченные лоты
// @Description  Ручной запуск закрытия лотов с истёкшим дедлайном
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]int
// @Failure      401
// @Failure      403
// @Router       /admin/close-expired [post]
func (h *handler) CloseExpiredLot(c *gin.Context) {
	count, err := h.lotsService.CloseExpiredLot(c.Request.Context())
	if err != nil {
		h.logger.Error("CloseExpiredLot: ", "error: ", err)
		response.RespondError(c, err)
		return
	}
	response.RespondJSON(c, http.StatusOK, gin.H{"closed": count})
}


func (h *handler) CreateLot(c *gin.Context) {
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
	err := h.lotsService.CreateLot(c.Request.Context(), req.Title, req.Description, req.StartPrice,
		req.Photo, req.EndsAt, req.Status, sellerId)
	if err != nil {
		response.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Lot created successfully!"})
}

func (h *handler) GetAll(c *gin.Context) {
	items, err := h.lotsService.GetAll(c.Request.Context())
	if err != nil {
		h.logger.Error("get lots repository", "err", err)
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
