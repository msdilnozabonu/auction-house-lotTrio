package bid

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/service/bid"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_PlaceBid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		m := new(bid.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"amount":100}`
		req := httptest.NewRequest(http.MethodPost, "/lots/1/bid", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("user_id", int64(42))

		m.On("PlaceBid", mock.Anything, int64(1), int64(42), float64(100)).Return(nil)

		h.PlaceBid(c)

		require.Equal(t, http.StatusCreated, w.Code)
		require.Contains(t, w.Body.String(), "bid placed successfully")
	})
	t.Run("service error", func(t *testing.T) {
		m := new(bid.MockService)
		h :=NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"amount":100}`
		req := httptest.NewRequest(http.MethodPost, "/lots/1/bid", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("user_id", int64(42))

		m.On("PlaceBid", mock.Anything, int64(1), int64(42), float64(100)).Return(errors.New("db error"))

		h.PlaceBid(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")
	})
}

func TestHandler_GetBiddersBids(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		m := new(bid.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/bids", nil)
		c.Request = req
		c.Set("user_id", int64(42))
		expected := []model.Bid{{BidderID: 1, Amount: 100}}
		m.On("GetBidderBids", mock.Anything, int64(42)).Return(expected, nil)
		h.GetBiddersBids(c)

		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "100")
	})

	t.Run("unauthorized", func(t *testing.T) {
		m := new(bid.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		h.GetBiddersBids(c)

		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("server error", func(t *testing.T) {
		m := new(bid.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/bids", nil)
		c.Request = req
		c.Set("user_id", int64(0))
		m.On("GetBidderBids", mock.Anything, int64(0)).
			Return(nil, errors.New("db error"))
		h.GetBiddersBids(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")
		m.AssertExpectations(t)
	})
}

func TestHandler_GetBidsByLotID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		m := new(bid.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/lots/1/bids", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("user_id", int64(99))
		expected := []model.Bid{{LotID: 1, Amount: 200}}
		m.On("GetBidsByLotID", mock.Anything, int64(1), int64(99)).Return(expected, nil)
		h.GetBidsByLotID(c)

		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "200")
	})

	t.Run("forbidden", func(t *testing.T) {
		m := new(bid.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("user_id", "not-an-int64")
		h.GetBidsByLotID(c)
		require.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("server error", func(t *testing.T) {
		m := new(bid.MockService)
		h := NewHandler(m)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/lots/1/bids", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("user_id", int64(0))
		m.On("GetBidsByLotID", mock.Anything, int64(1), int64(0)).
			Return(nil, errors.New("db error"))
		h.GetBidsByLotID(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")
		m.AssertExpectations(t)
	})
}