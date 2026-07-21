package lots

import (
	"auction-house-lotTrio/internal/response"
	"auction-house-lotTrio/internal/service/lots"
	"log/slog"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	CloseExpiredLot(c *gin.Context)
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
	response.RespondJSON(c, 200, gin.H{"closed": count})
}
