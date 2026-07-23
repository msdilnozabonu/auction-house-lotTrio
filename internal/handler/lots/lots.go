package lots

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/response"
	"auction-house-lotTrio/internal/service/lots"
	"errors"
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
	GetByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	UpdateStatusByID(c *gin.Context)
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
//
//	@Summary		Закрыть просроченные лоты
//	@Description	Ручной запуск закрытия лотов с истёкшим дедлайном
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]int
//	@Failure		401
//	@Failure		403
//	@Router			/admin/close-expired [post]
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
//
//	@Summary		Создать лоты
//	@Description	Создает лоты в базу
//	@Tags			lots
//	@Produce		json
//	@Param			input	body	lotsRequest	true	"Добавить лот"
//	@Success		201
//	@Failure		401
//	@Failure		403
//	@Security		BearerAuth
//	@Router			/lots/new [post]
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
	err := h.lotsService.CreateLot(c.Request.Context(), req.Title, req.Description, req.Category, req.StartPrice,
		req.Photo, req.EndsAt, req.Status, sellerId)
	if err != nil {
		response.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{messageKey: "Lot created successfully!"})
}

// GetAll  godoc
//
//	@Summary		Получение список лотов
//	@Description	Метод возрвщает список активных лотов
//	@Tags			lots
//	@Produce		json
//	@Param			search		query	string	false	"Поиск"
//	@Param			category	query	string	false	"Категория"
//	@Param			minPrice	query	number	false	"Минимальная цена"
//	@Param			maxPrice	query	number	false	"Максимальная цена"
//	@Param			page		query	int		false	"Номер страницы"
//	@Param			limit		query	int		false	"Количество элементов"
//	@Success		200
//	@Failure		400
//	@Security		BearerAuth
//	@Router			/lots [get]
func (h *handler) GetAll(c *gin.Context) {
	minPrice, _ := strconv.ParseFloat(c.Query("minPrice"), 64)
	maxPrice, _ := strconv.ParseFloat(c.Query("maxPrice"), 64)
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 15
	}

	filter := model.LotsFilter{
		Search:   c.Query("search"),
		Category: c.Query("category"),
		MinPrice: minPrice,
		MaxPrice: maxPrice,
		Page:     page,
		Limit:    limit,
	}
	items, total, err := h.lotsService.GetAll(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("get lots repository", "err", err)
		response.RespondError(c, err)
		return
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}
	c.JSON(http.StatusOK, pagResponse{Items: items, Total: total, Page: page, Limit: limit, TotalPages: totalPages})
}

// GetByID godoc
//
//	@Summary		Получение лота по ID
//	@Description	Returns a lot by its ID
//	@Tags			lots
//	@Produce		json
//	@Param			id	path	int	true	"Лот ID"
//	@Success		200
//	@Failure		400
//	@Failure		404
//	@Security		BearerAuth
//	@Router			/lots/:id [get]
func (h *handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	lotID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	role := c.GetString("role")
	var lot *model.Lots
	if role == "bidder" {
		lot, err = h.lotsService.GetByIDForBid(c.Request.Context(), model.Lots{ID: lotID})
	} else {
		lot, err = h.lotsService.GetByID(c.Request.Context(), model.Lots{ID: lotID})
	}
	if err != nil {
		h.logger.Error("get lot by id", "err", err)
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, lot)
}

// UpdateByID  godoc
//
//	@Summary		Получение лот по ID
//	@Description	Получает лот по ID и меняет его значение
//	@Tags			lots
//	@Produce		json
//	@Accept			json
//	@Param			input	body	updateLotRequest	true	"Изменит лот"
//	@Param			id		path	int					true	"ID лота"
//	@Success		200
//	@Failure		401
//	@Failure		403
//	@Security		BearerAuth
//	@Router			/lots/:id [put]
func (h *handler) UpdateByID(c *gin.Context) {
	var req updateLotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondError(c, err)
		return
	}

	id, sellerId, err := parseID(c)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	err = h.lotsService.UpdateById(c.Request.Context(), model.Lots{
		ID:          id,
		SellerID:    sellerId,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
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

// UpdateStatusByID  godoc
//
//	@Summary	Смена статуса
//	@Description	Меняет статус лота draft→live→closed->cancelled
//	@Tags			lots
//	@Produce		json
//	@Success		200
//	@Failure		401
//	@Security		BearerAuth
//	@Router			/lots/{id}/status [put]
func (h *handler) UpdateStatusByID(c *gin.Context) {
	id, sellerId, err := parseID(c)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondError(c, err)
		return
	}

	err = h.lotsService.UpdateStatus(c.Request.Context(), sellerId, id, req.Status)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{messageKey: "Status updated successfully!"})
}

// DeleteLots  godoc
//
//	@Summary		Удаление лот по ID
//	@Description	Получает лот по ID и удаляет его
//	@Tags			lots
//	@Produce		json
//	@Success		200
//	@Failure		401
//	@Failure		403
//	@Security		BearerAuth
//	@Router			/lots/:id [delete]
func (h *handler) DeleteLots(c *gin.Context) {
	id, sellerId, err := parseID(c)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	err = h.lotsService.DeleteLots(c.Request.Context(), model.Lots{ID: id, SellerID: sellerId})
	if err != nil {
		h.logger.Error("delete lots repository", "err", err)
		response.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{messageKey: "Lot deleted successfully!"})
}

func parseID(c *gin.Context) (int64, int64, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, 0, errors.New("invalid id")
	}

	userID, ok := c.Get("user_id")
	if !ok {
		return 0, 0, errors.New("user_id not found")
	}

	sellerID, ok := userID.(int64)
	if !ok {
		return 0, 0, errors.New("user_id has invalid type")
	}

	return id, sellerID, nil
}

type lotsRequest struct {
	Title       string    `binding:"required" json:"title"`
	Description string    `binding:"required" json:"description"`
	Category    string    `binding:"required" json:"category"`
	StartPrice  float64   `binding:"required" json:"startPrice"`
	EndsAt      time.Time `binding:"required" json:"endsAt"`
	Photo       string    `binding:"required" json:"photo"`
	Status      string    `binding:"required" json:"status"`
}

type updateLotRequest struct {
	ID          int64     `json:"ID"`
	Title       string    `binding:"required" json:"title"`
	Description string    `binding:"required" json:"description"`
	Category    string    `binding:"required" json:"category"`
	StartPrice  float64   `binding:"required" json:"startPrice"`
	Photo       string    `binding:"required" json:"photo"`
	EndsAt      time.Time `binding:"required" json:"endsAt"`
}

type pagResponse struct {
	Items      []model.Lots `json:"items"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
	Total      int          `json:"total"`
	TotalPages int          `json:"totalPages"`
}

type updateStatusRequest struct {
	Status string `json:"status"`
}
