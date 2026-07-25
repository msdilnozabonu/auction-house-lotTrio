package lots

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/service/lots"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_CloseExpiredLot_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	m.On("CloseExpiredLot", mock.Anything).Return(3, nil)

	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/close-expired", nil)
	h.CloseExpiredLot(c)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any

	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, float64(3), resp["closed"])
	m.AssertExpectations(t)
}

func TestHandler_CloseExpiredLot_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	m.On("CloseExpiredLot", mock.Anything).Return(0, errors.New("database error"))
	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/close-expired", nil)
	h.CloseExpiredLot(c)
	require.Equal(t, http.StatusInternalServerError, w.Code)
	m.AssertExpectations(t)
}

func TestHandler_GetLotsForAdmin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	lot := []model.Lots{
		{ID: 1, Title: "Lot 1"},
		{ID: 2, Title: "Lot 2"},
	}
	m.On("FindLotsForAdmin", mock.Anything, mock.Anything).Return(lot, 1, nil)
	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/lots?status=live&page=1", nil)
	h.GetLotsForAdmin(c)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, float64(1), resp["total"])
	m.AssertExpectations(t)
}

func TestHandler_GetLotsForAdmin_WithPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	lot := []model.Lots{{ID: 1, Title: "Lot 1"}}
	m.On("FindLotsForAdmin", mock.Anything, mock.Anything).
		Return(lot, 1, nil)

	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/lots?page=2", nil)

	h.GetLotsForAdmin(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, float64(1), resp["total"])
	m.AssertExpectations(t)
}

func TestHandler_GetLotsForAdmin_WithSellerId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	lot := []model.Lots{{ID: 1, Title: "Lot 1"}}
	m.On("FindLotsForAdmin", mock.Anything, mock.Anything).
		Return(lot, 1, nil)

	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/lots?seller_id=2", nil)

	h.GetLotsForAdmin(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, float64(1), resp["total"])
	m.AssertExpectations(t)
}

func TestHandler_GetLotsForAdmin_WithLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	lot := []model.Lots{{ID: 1, Title: "Lot 1"}}
	m.On("FindLotsForAdmin", mock.Anything, mock.Anything).
		Return(lot, 1, nil)

	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/lots?page_size=2", nil)

	h.GetLotsForAdmin(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, float64(1), resp["total"])
	m.AssertExpectations(t)
}

func TestHandler_GetLotsForAdmin_WithDateField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	lot := []model.Lots{{ID: 1, Title: "Lot 1"}}
	m.On("FindLotsForAdmin", mock.Anything, mock.Anything).
		Return(lot, 1, nil)

	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet,
		"/admin/lots?date_field=ends_at&date_from=2026-01-02&date_to=2026-01-05", nil)

	h.GetLotsForAdmin(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, float64(1), resp["total"])
	m.AssertExpectations(t)
}

func TestHandler_GetLotsForAdmin_DateFrom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	lot := []model.Lots{{ID: 1, Title: "Lot 1"}}
	m.On("FindLotsForAdmin", mock.Anything, mock.Anything).
		Return(lot, 1, nil)

	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/lots?date_from=2006-01-02", nil)

	h.GetLotsForAdmin(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, float64(1), resp["total"])
	m.AssertExpectations(t)
}

func TestHandler_GetLotsForAdmin_DateTo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	lot := []model.Lots{{ID: 1, Title: "Lot 1"}}
	m.On("FindLotsForAdmin", mock.Anything, mock.Anything).
		Return(lot, 1, nil)

	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/lots?date_to=2006-01-02", nil)

	h.GetLotsForAdmin(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, float64(1), resp["total"])
	m.AssertExpectations(t)
}


func TestHandler_GetLotsForAdmin_InvalidSellerId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/lots?seller_id=invalid", nil)
	h.GetLotsForAdmin(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetLotsForAdmin_InvalidDateField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/lots?date_field=invalid", nil)
	h.GetLotsForAdmin(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetLotsForAdmin_InvalidDateFrom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/lots?date_from=invalid", nil)
	h.GetLotsForAdmin(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetLotsForAdmin_InvalidDateTo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/lots?date_to=invalid", nil)
	h.GetLotsForAdmin(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetLotsForAdmin_InvalidPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/lots?page=invalid", nil)
	h.GetLotsForAdmin(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetLotsForAdmin_InvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/lots?page_size=invalid", nil)
	h.GetLotsForAdmin(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetLotsForAdmin_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	m.On("FindLotsForAdmin", mock.Anything, mock.Anything).Return([]model.Lots{}, 0, errors.New("database error"))
	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/lots?status=live", nil)
	h.GetLotsForAdmin(c)
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
