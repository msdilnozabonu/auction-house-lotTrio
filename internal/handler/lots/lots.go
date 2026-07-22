package lots

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/response"
	"auction-house-lotTrio/internal/service/lots"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	CloseExpiredLot(c *gin.Context)
	CreateLot(c *gin.Context)
	GetAll(c *gin.Context)
	UpdateByID(c *gin.Context)
	DeleteLots(c *gin.Context)
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
const messageKey = "message"

// CreateLot     godoc
// @Summary      Создать лоты
// @Description  Создает лоты в базу
// @Tags         lots
// @Produce      json
// @Param		 input body lotsRequest true "Добавить лот"
// @Success      201
// @Failure      401
// @Failure      403
// @Router       /lots/new [post]
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
	c.JSON(http.StatusCreated, gin.H{messageKey: "Lot created successfully!"})
}

// GetAll  godoc
// @Summary Получение список лотов
// @Description Метод возрвщает список активных лотов
// @Tags         lots
// @Produce      json
// @Param		 input body lotsRequest true "Добавить лот"
// @Success      200
// @Failure      400
// @Router       /lots [get]
func (h *handler) GetAll(c *gin.Context) {
	items, err := h.lotsService.GetAll(c.Request.Context())
	if err != nil {
		h.logger.Error("get lots repository", "err", err)
		response.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

// UpdateByID  godoc
// @Summary Получение лот по ID
// @Description Получает лот по ID и меняет его значение
// @Tags         lots
// @Produce      json
// @Param		 input body updateLotRequest true "Изменит лот"
// @Success      200
// @Failure      401
// @Failure      403
// @Router       /lots/:id [put]
func (h *handler) UpdateByID(c *gin.Context) {
	var req updateLotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondError(c, err)
		return
	}

	id := c.Param("id")
	IDint, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		response.RespondError(c, err)
		return
	}

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


	err = h.lotsService.UpdateById(c.Request.Context(), model.Lots{
		ID:          IDint,
		SellerID:    sellerId,
		Title:       req.Title,
		Description: req.Description,
		StartPrice:  req.StartPrice,
		Photo:       req.Photo,
		EndAt:       req.EndsAt,
	})
	if err != nil {
		h.logger.Error("update lots repository", "err", err)
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{messageKey: "Lot updated successfully!"})
}


// DeleteLots  godoc
// @Summary Удаление лот по ID
// @Description Получает лот по ID и удаляет его
// @Tags         lots
// @Produce      json
// @Success      200
// @Failure      401
// @Failure      403
// @Router       /lots/:id [delete]
func (h *handler) DeleteLots(c *gin.Context) {
	id := c.Param("id")
	IDint, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		response.RespondError(c, err)
		return
	}

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

	err = h.lotsService.DeleteLots(c.Request.Context(), model.Lots{ID:IDint, SellerID: sellerId})
	if err != nil {
		h.logger.Error("delete lots repository", "err", err)
		response.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{messageKey: "Lot deleted successfully!"})
}


type lotsRequest struct {
	Title       string    `binding:"required" json:"title"`
	Description string    `binding:"required" json:"description"`
	StartPrice  float64   `binding:"required" json:"startPrice"`
	EndsAt      time.Time `binding:"required" json:"endsAt"`
	Photo       string    `binding:"required" json:"photo"`
	Status      string    `binding:"required" json:"status"`
}

type updateLotRequest struct {
	ID          int64     `json:"ID"`
	Title       string    `binding:"required" json:"title"`
	Description string    `binding:"required" json:"description"`
	StartPrice  float64   `binding:"required" json:"startPrice"`
	Photo       string    `binding:"required" json:"photo"`
	EndsAt      time.Time `binding:"required" json:"endsAt"`
}
