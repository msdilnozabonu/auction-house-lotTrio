package auth

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/service/auth"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	Registration(c *gin.Context)
	Login(c *gin.Context)
}

type handler struct {
	authService auth.Service
}

func NewHandler(authService auth.Service) Handler {
	return &handler{authService: authService}
}

// Registration godoc
// @Summary      Регистрация пользователя
// @Description  Создаёт пользователя в базе
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body user true "Регистрация"
// @Success      201
// @Failure      400
// @Failure      500
// @Router       /auth/register [post]
func (h *handler) Registration(c *gin.Context) {
	var request user
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.Register(c.Request.Context(), request.Login, request.Pass, request.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully!"})
}

// Login godoc
// @Summary      Авторизация пользователя
// @Description  Проверяет логин и пароль, возвращает токен
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body user true "Логин и пароль"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /auth/login [post]
func (h *handler) Login(c *gin.Context) {
	var req user
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Login(c.Request.Context(), req.Login, req.Pass)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Логин или пароль неверный"})
			return
		}

		if errors.Is(err, model.ErrIncorrectPassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Логин или пароль неверный"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"authToken": user})
}

type user struct {
	Login string `json:"login"`
	Pass  string `json:"password"`
	Role  string `json:"role"`
}
