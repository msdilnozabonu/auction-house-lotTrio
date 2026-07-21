package middleware

import (
	"auction-house-lotTrio/internal/service/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Middleware interface {
	Auth() gin.HandlerFunc
}

type middleware struct {
	authService auth.Service
}

func NewMiddleware(authService auth.Service) Middleware {
	return &middleware{
		authService: authService}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		c.Next()
	}
}

func (m *middleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken := c.GetHeader("Authorization")
		if !strings.HasPrefix(authToken, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		token := strings.TrimPrefix(authToken, "Bearer ")
		uid, role, err := m.authService.ValidateAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("user_id", uid)
		c.Set("role", role)
		c.Next()
	}
}
