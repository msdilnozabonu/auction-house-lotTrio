package wins

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/service/wins"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_GetWins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("success", func(t *testing.T) {
		m := new(wins.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/wins", nil)
		c.Request = req
		c.Set("user_id", int64(1))
		expected := []model.Win{
			{LotID: 1,
				LotTitle:   "Vintage Painting",
				WinningBid: 2500.50,
				ClosedAt:   time.Date(2026, 7, 29, 13, 45, 0, 0, time.UTC),
			},
		}
		m.On("GetWins", mock.Anything, int64(1)).
			Return(expected, nil)
		h.GetWins(c)
		require.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()
		require.Contains(t, body, `"lot_id":1`)
		require.Contains(t, body, `"lot_title":"Vintage Painting"`)
		require.Contains(t, body, `"winning_bid":2500.5`)
		require.Contains(t, body, `"closed_at"`)
		m.AssertExpectations(t)
	})
	t.Run("unauthorized", func(t *testing.T) {
		m := new(wins.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/wins", nil)
		c.Request = req
		h.GetWins(c)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("service error", func(t *testing.T) {
		m := new(wins.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/wins", nil)
		c.Request = req
		c.Set("user_id", int64(1))
		m.On("GetWins", mock.Anything, int64(1)).
			Return([]model.Win(nil), errors.New("db error"))
		h.GetWins(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")
		m.AssertExpectations(t)
	})
}
