package user

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/service/user"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_Me(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("success", func(t *testing.T) {
		m := new(user.MockService)
		h := New(m, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		c.Request = req
		c.Set("user_id", int64(42))
		expected := model.User{ID: 42, Login: "raykhona"}
		m.On("Me", mock.Anything, int64(42)).
			Return(expected, nil)
		h.Me(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "raykhona")
		m.AssertExpectations(t)
	})
	t.Run("unauthorized", func(t *testing.T) {
		m := new(user.MockService)
		h := New(m, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		h.Me(c)
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		c.Request = req
		require.Equal(t, http.StatusUnauthorized, w.Code)
		m.AssertExpectations(t)
	})
	t.Run("server error", func(t *testing.T) {
		m := new(user.MockService)
		h := New(m, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		c.Request = req
		c.Set("user_id", int64(42))
		m.On("Me", mock.Anything, int64(42)).
			Return(model.User{}, errors.New("db error"))
		h.Me(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
