package user

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/response"
	userservice "auction-house-lotTrio/internal/service/user"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	Me(c *gin.Context)
}

type handler struct {
	userService userservice.Service
	logger      *slog.Logger
}

func New(userService userservice.Service, logger *slog.Logger) Handler {
	return &handler{
		userService: userService,
		logger:      logger,
	}
}

// Me godoc
//
//	@Summary		Информация о текущем пользователе
//	@Description	Возвращает данные авторизованного пользователя.
//	@Tags			user
//	@Produce		json
//	@Success		200
//	@Failure		401
//	@Failure		404
//	@Failure		500
//	@Security		BearerAuth
//	@Router			/me [get]
func (h *handler) Me(c *gin.Context) {
	v, ok := c.Get("user_id")
	if !ok {
		response.RespondError(c, model.ErrInvalidAuth)
		return
	}

	userID, ok := v.(int64)
	if !ok {
		response.RespondError(c, model.ErrInvalidAuth)
		return
	}

	user, err := h.userService.Me(c.Request.Context(), userID)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}
