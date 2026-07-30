package watchlist

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/response"
	"auction-house-lotTrio/internal/service/watchlist"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	messageKey = "message"
)

type Handler interface {
	Add(c *gin.Context)
	Delete(c *gin.Context)
	GetWatchlist(c *gin.Context)
}

type handler struct {
	watchService watchlist.Service
	logger       *slog.Logger
}

func NewHandler(watchService watchlist.Service) Handler {
	return &handler{
		watchService: watchService,
		logger:       slog.With("module", "watchlist"),
	}
}

// Add godoc
//
//	@Summary		Добавить лот в список отслеживания
//	@Description	Добавляет лот в список по ID лота
//	@Tags			watchlist
//	@Security		BearerAuth
//	@Param			id	path		int	true	"ID лота"
//	@Success		201	{object}	map[string]string
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		409
//	@Failure		500
//	@Router			/lots/:id/watch [post]
func (h *handler) Add(c *gin.Context) {
	lotID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.RespondError(c, model.ErrInvalidID)
		return
	}

	userID, ok := c.Get("user_id")
	if !ok {
		response.RespondError(c, model.ErrUnauthorized)
		return
	}

	bidderID, ok := userID.(int64)
	if !ok {
		response.RespondError(c, model.ErrForbidden)
		return
	}

	if err := h.watchService.Add(c.Request.Context(), bidderID, lotID); err != nil {
		h.logger.Error("add watch", "err", err)
		response.RespondError(c, err)
		return
	}

	response.RespondJSON(c, http.StatusCreated, gin.H{
		messageKey: "lot added to watchlist",
	})
}

// Delete godoc
//
//	@Summary		Удалить лот из списка отслеживания
//	@Description	Удаляет лот из списка отслеживания участника по ID лота
//	@Tags			watchlist
//	@Security		BearerAuth
//	@Param			id	path		int	true	"ID лота"
//	@Success		200	{object}	map[string]string
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		500
//	@Router			/lots/:id/watch [delete]
func (h *handler) Delete(c *gin.Context) {
	lotID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.RespondError(c, model.ErrInvalidID)
		return
	}

	userID, ok := c.Get("user_id")
	if !ok {
		response.RespondError(c, model.ErrUnauthorized)
		return
	}

	bidderID, ok := userID.(int64)
	if !ok {
		response.RespondError(c, model.ErrForbidden)
		return
	}

	if err := h.watchService.Delete(c.Request.Context(), bidderID, lotID); err != nil {
		h.logger.Error("delete watch", "err", err)
		response.RespondError(c, err)
		return
	}

	response.RespondJSON(c, http.StatusOK, gin.H{
		messageKey: "lot removed from watchlist",
	})
}

// GetWatchlist godoc
//
//	@Summary		Получить список отслеживания
//	@Description	Возвращает список отслеживания участника с пагинацией
//	@Tags			watchlist
//	@Security		BearerAuth
//	@Produce		json
//	@Param			page	query		int	false	"Номер страницы"
//	@Param			limit	query		int	false	"Количество элементов на странице"
//	@Success		200	{object}	WatchlistResponse
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		500
//	@Router			/watchlist [get]
func (h *handler) GetWatchlist(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		response.RespondError(c, err)
		return
	}
	if page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "15"))
	if err != nil {
		response.RespondError(c, err)
		return
	}
	if limit < 1 || limit > 100 {
		limit = 15
	}

	userID, ok := c.Get("user_id")
	if !ok {
		response.RespondError(c, model.ErrUnauthorized)
		return
	}

	bidderID, ok := userID.(int64)
	if !ok {
		response.RespondError(c, model.ErrForbidden)
		return
	}
	items, total, err := h.watchService.GetWatchlist(
		c.Request.Context(),
		bidderID,
		page,
		limit,
	)
	if err != nil {
		h.logger.Error("get watchlist", "err", err)
		response.RespondError(c, err)
		return
	}

	totalPages := (total + limit - 1) / limit
	response.RespondJSON(c, http.StatusOK, gin.H{
		"items":      items,
		"page":       page,
		"limit":      limit,
		"total":      total,
		"totalPages": totalPages,
	})
}

type WatchlistResponse struct {
	Items      []model.WatchItem `json:"items"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	Total      int               `json:"total"`
	TotalPages int               `json:"totalPages"`
}
