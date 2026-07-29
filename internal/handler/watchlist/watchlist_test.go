package watchlist

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/service/watchlist"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_Add(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("success", func(t *testing.T) {
		m := new(watchlist.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodPost, "/lots/1/watch", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("user_id", int64(42))
		m.On("Add", mock.Anything, int64(42), int64(1)).Return(nil)
		h.Add(c)
		require.Equal(t, http.StatusCreated, w.Code)
		require.Contains(t, w.Body.String(), "lot added to watchlist")
		m.AssertExpectations(t)
	})
	t.Run("invalid id", func(t *testing.T) {
		m := new(watchlist.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodPost, "/lots/abc/watch", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "abc"}}
		c.Set("user_id", int64(42))
		h.Add(c)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("service error", func(t *testing.T) {
		m := new(watchlist.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodPost, "/lots/1/watch", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("user_id", int64(42))
		m.On("Add", mock.Anything, int64(42), int64(1)).Return(errors.New("db error"))
		h.Add(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")
		m.AssertExpectations(t)
	})
}

func TestHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("success", func(t *testing.T) {
		m := new(watchlist.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodDelete, "/lots/1/watch", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("user_id", int64(42))

		m.On("Delete", mock.Anything, int64(42), int64(1)).Return(nil)
		h.Delete(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "lot removed from watchlist")
		m.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		m := new(watchlist.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodDelete, "/lots/1/watch", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("user_id", int64(42))
		m.On("Delete", mock.Anything, int64(42), int64(1)).Return(errors.New("db error"))
		h.Delete(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")
		m.AssertExpectations(t)
	})
}

func TestHandler_GetWatchlist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		m := new(watchlist.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/watchlist?page=1&limit=10", nil)
		c.Request = req
		c.Set("user_id", int64(42))
		expected := []model.WatchItem{{LotID: 1}}
		m.On("GetWatchlist", mock.Anything, int64(42), 1, 10).
			Return(expected, 1, nil)

		h.GetWatchlist(c)

		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "totalPages")
		m.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		m := new(watchlist.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/watchlist", nil)
		c.Request = req

		h.GetWatchlist(c)

		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("service error", func(t *testing.T) {
		m := new(watchlist.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/watchlist", nil)
		c.Request = req
		c.Set("user_id", int64(42))
		m.On("GetWatchlist", mock.Anything, int64(42), 1, 15).
			Return([]model.WatchItem{}, 0, errors.New("db error"))
		h.GetWatchlist(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")
		m.AssertExpectations(t)
	})
}
