package auth

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/service/auth"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("success", func(t *testing.T) {
		m := new(auth.MockRegister)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"login":"user1","password":"passpass"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		tokenPair := auth.TokenPair{AccessToken: "access123", RefreshToken: "refresh123"}
		m.On("Login", mock.Anything, "user1", "passpass").Return(tokenPair, nil)
		h := NewHandler(m)
		h.Login(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "access123")
		require.Contains(t, w.Body.String(), "refresh123")
	})
	t.Run("server error", func(t *testing.T) {
		m := new(auth.MockRegister)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"login":"user1","password":"passpass"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		m.On("Login", mock.Anything, "user1", "passpass").
			Return(auth.TokenPair{}, errors.New("db error"))
		h := NewHandler(m)
		h.Login(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")
		m.AssertExpectations(t)
	})
}
func TestHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("success", func(t *testing.T) {
		m := new(auth.MockRegister)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"login":"user1","password":"passpass","role":"user"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		m.On("Register", mock.Anything, "user1", "passpass", "user").Return(nil)

		h := NewHandler(m)
		h.Registration(c)

		require.Equal(t, http.StatusCreated, w.Code)
		require.Contains(t, w.Body.String(), "User created successfully")
	})
	t.Run("invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		m := new(auth.MockRegister)
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader("{bad json"))
		c.Request.Header.Set("Content-Type", "application/json")
		h := NewHandler(m)
		h.Registration(c)
		require.Equal(t, http.StatusBadRequest, w.Code)
		m.AssertNotCalled(t, "Register", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("server error", func(t *testing.T) {
		w := httptest.NewRecorder()
		m := new(auth.MockRegister)
		c, _ := gin.CreateTestContext(w)
		body := `{"login":"user1","password":"passpass","role":"user"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		m.On("Register", mock.Anything, "user1", "passpass", "user").Return(errors.New("db error"))
		h := NewHandler(m)
		h.Registration(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")
		m.AssertExpectations(t)
	})
}
func TestHandler_Refresh(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("success", func(t *testing.T) {
		m := new(auth.MockRegister)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"refresh_token":"refresh123"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		tokenPair := auth.TokenPair{AccessToken: "newAccess", RefreshToken: "newRefresh"}
		m.On("Refresh", mock.Anything, "refresh123").Return(tokenPair, nil)
		h := NewHandler(m)
		h.Refresh(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "newAccess")
		require.Contains(t, w.Body.String(), "newRefresh")
	})
	t.Run("server error", func(t *testing.T) {
		m := new(auth.MockRegister)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"refresh_token":"refresh123"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		m.On("Refresh", mock.Anything, "refresh123").Return(auth.TokenPair{}, errors.New("db error"))
		h := NewHandler(m)
		h.Refresh(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")
		m.AssertExpectations(t)
	})
}

func TestHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		m := new(auth.MockRegister)
		m.On("Logout", mock.Anything, "valid-refresh-token").Return(nil)
		h := NewHandler(m)
		body := `{"refresh_token": "valid-refresh-token"}`
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Logout(c)
		c.Writer.WriteHeaderNow()
		require.Equal(t, http.StatusNoContent, w.Code)
		require.Empty(t, w.Body.Bytes())
		m.AssertExpectations(t)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		m := new(auth.MockRegister)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader(`not-json`))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Logout(c)
		require.Equal(t, http.StatusBadRequest, w.Code)
		m.AssertNotCalled(t, "Logout", mock.Anything, mock.Anything)
	})
	t.Run("server error", func(t *testing.T) {
		m := new(auth.MockRegister)
		m.On("Logout", mock.Anything, "bad-token").Return(model.ErrInvalidToken)
		h := NewHandler(m)
		body := `{"refresh_token": "bad-token"}`
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Logout(c)
		require.Equal(t, http.StatusUnauthorized, w.Code)
		m.AssertExpectations(t)
	})
}
