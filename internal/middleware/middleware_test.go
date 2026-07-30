package middleware

import (
	"auction-house-lotTrio/internal/service/auth"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRequireRole_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("role", "admin")
	_ = RequireRole("admin")
	require.Equal(t, http.StatusOK, w.Code)
}
func TestRequireRole_Fail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequireRole("admin"))
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("role", "user")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}
func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("missing bearer token", func(t *testing.T) {
		m := new(auth.MockRegister)
		mw := NewMiddleware(m)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		mw.Auth()(c)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("valid token", func(t *testing.T) {
		m := new(auth.MockRegister)
		m.On("ValidateAccessToken", "goodtoken").
			Return(int64(42), "admin", nil)
		mw := NewMiddleware(m)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer goodtoken")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler := gin.New()
		handler.Use(mw.Auth())
		handler.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
		handler.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		m.AssertExpectations(t)
	})
	t.Run("invalid token", func(t *testing.T) {
		m := new(auth.MockRegister)
		m.On("ValidateAccessToken", "badtoken").
			Return(int64(1), "", errors.New("invalid token"))
		mw := NewMiddleware(m)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer badtoken")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler := gin.New()
		handler.Use(mw.Auth())
		handler.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
		handler.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
