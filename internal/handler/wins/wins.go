package wins

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/response"
	"auction-house-lotTrio/internal/service/wins"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	GetWins(c *gin.Context)
}

type handler struct {
	winsService wins.Service
	logger      *slog.Logger
}

func NewHandler(winsService wins.Service) Handler {
	return &handler{
		winsService: winsService,
		logger:      slog.With("module", "wins"),
	}
}

// GetWins godoc
//
//	@Summary		Получить выигранные лоты
//	@Description	Возвращает закрытые лоты, выигранные авторизованным участником
//	@Tags			wins
//	@Produce		json
//	@Success		200	{array}		model.Win
//	@Failure		401
//	@Failure		403
//	@Failure		500
//	@Security		BearerAuth
//	@Router			/wins [get]
func (h *handler) GetWins(c *gin.Context) {
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

	bidderWins, err := h.winsService.GetWins(c.Request.Context(), bidderID)
	if err != nil {
		h.logger.Error("get wins", "err", err)
		response.RespondError(c, err)
		return
	}

	response.RespondJSON(c, http.StatusOK, bidderWins)
}
