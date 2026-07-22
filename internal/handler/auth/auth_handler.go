package auth

import (
	"auction-house-lotTrio/internal/response"
	"auction-house-lotTrio/internal/service/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	errorKey = "error"
)

type Handler interface {
	Registration(c *gin.Context)
	Login(c *gin.Context)
	Refresh(c *gin.Context)
	Logout(c *gin.Context)
}

type handler struct {
	authService auth.Service
}

func NewHandler(authService auth.Service) Handler {
	return &handler{authService: authService}
}

// Registration godoc
//
//	@Summary		Регистрация пользователя
//	@Description	Создаёт пользователя в базу
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body	registerRequest	true	"Регистрация"
//	@Success		201
//	@Failure		400
//	@Failure		500
//	@Router			/auth/register [post]
func (h *handler) Registration(c *gin.Context) {
	var request registerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: err.Error()})
		return
	}

	err := h.authService.Register(c.Request.Context(), request.Login, request.Pass, request.Role)
	if err != nil {
		response.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully!"})
}

// Login godoc
//
//	@Summary		Авторизация пользователя
//	@Description	Проверяет логин и пароль, возвращает токен
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body	loginRequest	true	"Логин и пароль"
//	@Success		200
//	@Failure		400
//	@Failure		500
//	@Router			/auth/login [post]
func (h *handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: err.Error()})
		return
	}

	user, err := h.authService.Login(c.Request.Context(), req.Login, req.Pass)
	if err != nil {
		response.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

// Refresh godoc
//
//	@Summary		Обновление токенов
//	@Description	Ротирует refresh-токен и выдаёт новую пару токенов
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body	refreshRequest	true	"Refresh токен"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Router			/auth/refresh [post]
func (h *handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: err.Error()})
		return
	}

	tokens, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, tokens)
}

// Logout godoc
//
//	@Summary		Выход пользователя
//	@Description	Отзывает refresh-токен
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body	refreshRequest	true	"Refresh токен"
//	@Success		204
//	@Failure		400
//	@Failure		500
//	@Security		BearerAuth
//	@Router			/auth/logout [post]
func (h *handler) Logout(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: err.Error()})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type registerRequest struct {
	Login string `binding:"required" json:"login"`
	Pass  string `binding:"required" json:"password"`
	Role  string `json:"role"`
}

type loginRequest struct {
	Login string `binding:"required" json:"login"`
	Pass  string `binding:"required" json:"password"`
}

type refreshRequest struct {
	RefreshToken string `binding:"required" json:"refresh_token"`
}
