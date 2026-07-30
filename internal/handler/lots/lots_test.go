package lots

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/service/lots"
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

func TestHandler_Moderate_Approve(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	m.On("ModerateALot", mock.Anything, int64(1), true, "").Return(nil)
	h := NewHandler(m)
	body := `{"approve":true, "reason":""}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/admin/lots/1/moderate", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.Moderate(c)
	require.Equal(t, http.StatusOK, w.Code)
	m.AssertExpectations(t)
}

func TestHandler_Moderate_RejectWithReason(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	m.On("ModerateALot", mock.Anything, int64(1), false, "rejection reason").Return(nil)
	h := NewHandler(m)
	body := `{"approve":false, "reason":"rejection reason"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/admin/lots/1/moderate", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.Moderate(c)
	require.Equal(t, http.StatusOK, w.Code)
	m.AssertExpectations(t)
}

func TestHandler_Moderate_InvalidId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	h := NewHandler(m)
	body := `{"approve":true, "reason":""}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/admin/lots/invalid/moderate", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	h.Moderate(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Moderate_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	h := NewHandler(m)

	// Invalid JSON body
	body := `{"approve":true, "reason":` // malformed JSON
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/admin/lots/1/moderate", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.Moderate(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Moderate_InternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	m.On("ModerateALot", mock.Anything, int64(1), true, "").Return(errors.New("db error"))

	h := NewHandler(m)
	body := `{"approve":true, "reason":""}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/admin/lots/1/moderate", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.Moderate(c)
	require.Equal(t, http.StatusInternalServerError, w.Code)
	m.AssertExpectations(t)
}

func TestHandler_GetPlatformStats_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	m.On("GetPlatformStats", mock.Anything).Return(model.PlatformStats{}, nil)
	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/stats", nil)
	c.Request.Header.Set("Content-Type", "application/json")

	h.GetPlatformStats(c)
	require.Equal(t, http.StatusOK, w.Code)
	m.AssertExpectations(t)
}

func TestHandler_GetPlatformStats_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	m.On("GetPlatformStats", mock.Anything).Return(model.PlatformStats{}, errors.New("db error"))
	h := NewHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/stats", nil)
	h.GetPlatformStats(c)
	require.Equal(t, http.StatusInternalServerError, w.Code)
	m.AssertExpectations(t)
}

func TestHandler_ExportLots_JSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/export?format=json", nil)

	m := new(lots.Mock)
	expected := []model.LotsExport{
		{ID: 1, Title: "Lot A", Category: "Cars", Status: "live", SellerID: 10, Price: 1000.50, EndsAt: time.Now()},
	}
	m.On("ExportLots", mock.Anything, mock.Anything, mock.Anything).
		Return(expected, nil)

	h := NewHandler(m)
	h.ExportLots(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "attachment; filename=lots_export.json", w.Header().Get("Content-Disposition"))
	require.Contains(t, w.Body.String(), "Lot A")

	m.AssertExpectations(t)
}

func TestHandler_ExportLots_CSV(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/export?format=csv", nil)

	m := new(lots.Mock)
	expected := []model.LotsExport{
		{ID: 1, Title: "Lot A", Category: "Cars", Status: "live", SellerID: 10, Price: 1000.50, EndsAt: time.Now()},
	}
	m.On("ExportLots", mock.Anything, mock.Anything, mock.Anything).
		Return(expected, nil)

	h := NewHandler(m)
	h.ExportLots(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "attachment; filename=lots_export.csv", w.Header().Get("Content-Disposition"))
	require.Contains(t, w.Body.String(), "Lot A")

	m.AssertExpectations(t)
}

func TestHandler_ExportLots_DefaultJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/export", nil)

	m := new(lots.Mock)
	expected := []model.LotsExport{{ID: 1, Title: "Lot A"}}
	m.On("ExportLots", mock.Anything, mock.Anything, mock.Anything).
		Return(expected, nil)

	h := NewHandler(m)
	h.ExportLots(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "attachment; filename=lots_export.json", w.Header().Get("Content-Disposition"))
	require.Contains(t, w.Body.String(), "Lot A")
}

func TestHandler_ExportLots_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/export?format=incorrect", nil)

	h := NewHandler(m)
	h.ExportLots(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "invalid format")
}

func TestHandler_ExportLots_InvalidDateFrom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/export?date_from=bad-date", nil)

	h := NewHandler(m)
	h.ExportLots(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "invalid date_from")
}
func TestHandler_ExportLots_InvalidDateTo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/export?date_to=bad-date", nil)

	h := NewHandler(m)
	h.ExportLots(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "invalid date_to")
}

func TestHandler_ExportLots_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/export", nil)

	m := new(lots.Mock)
	m.On("ExportLots", mock.Anything, mock.Anything, mock.Anything).
		Return([]model.LotsExport{}, errors.New("db error"))

	h := NewHandler(m)
	h.ExportLots(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Contains(t, w.Body.String(), "internal server error")
}

func TestHandler_ExportLots_ValidDateFrom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/admin/export?date_from=2026-07-01", nil)

	m := new(lots.Mock)
	expected := []model.LotsExport{{ID: 1, Title: "Lot A"}}
	m.On("ExportLots", mock.Anything, mock.Anything, mock.Anything).
		Return(expected, nil)

	h := NewHandler(m)
	h.ExportLots(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "Lot A")
	m.AssertExpectations(t)
}

func TestHandler_ExportLots_ValidDateTo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/export?date_to=2026-07-28", nil)

	m := new(lots.Mock)
	expected := []model.LotsExport{{ID: 2, Title: "Lot B"}}
	m.On("ExportLots", mock.Anything, mock.Anything, mock.Anything).
		Return(expected, nil)

	h := NewHandler(m)
	h.ExportLots(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "Lot B")
	m.AssertExpectations(t)
}

func TestHandler_CreateLot_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("success", func(t *testing.T) {
		m := new(lots.Mock)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"title":"Test Lot","description":"Desc","category":"Cars",
                  "startPrice":1000.50,"photo":"img.png","endsAt":"2026-07-28T12:00:00Z","status":"live"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/new",
			strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_id", int64(1))
		m.On("CreateLot", mock.Anything, "Test Lot", "Desc", "Cars", 1000.50,
			"img.png", mock.AnythingOfType("time.Time"), "live", int64(1)).
			Return(nil)
		h := NewHandler(m)
		h.CreateLot(c)
		require.Equal(t, http.StatusCreated, w.Code)
		require.Contains(t, w.Body.String(), "Lot created successfully")
		m.AssertExpectations(t)
	})
	t.Run("unauthorized", func(t *testing.T) {
		m := new(lots.Mock)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/lots", nil)

		h := NewHandler(m)
		h.CreateLot(c)

		require.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		m := new(lots.Mock)
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/lots", strings.NewReader("{bad json"))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_id", int64(42))

		h := NewHandler(m)
		h.CreateLot(c)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"title":"Lot A","description":"Desc","category":"Cars","startPrice":1000,"photo":"img.png","endsAt":"2026-08-01T12:00:00Z","status":"live"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/lots", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_id", int64(42))

		m := new(lots.Mock)
		m.On("CreateLot", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("db error"))

		h := NewHandler(m)
		h.CreateLot(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")

		m.AssertExpectations(t)
	})
}

func TestHandler_GetAll(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		m := new(lots.Mock)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/lots?page=1&limit=2", nil)

		items := []model.Lots{
			{ID: 1, Title: "Lot 1"},
			{ID: 2, Title: "Lot 2"},
		}
		m.On("GetAll", mock.Anything, mock.AnythingOfType("model.LotsFilter")).
			Return(items, 2, nil)
		h := NewHandler(m)
		h.GetAll(c)
		require.Equal(t, http.StatusOK, w.Code)

		var resp pagResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		require.Equal(t, 2, resp.Total)
		require.Equal(t, 1, resp.Page)
		require.Equal(t, 2, resp.Limit)
		require.Equal(t, 1, resp.TotalPages)
		m.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		m := new(lots.Mock)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/lots?page=1&limit=2", nil)

		m.On("GetAll", mock.Anything, mock.AnythingOfType("model.LotsFilter")).
			Return([]model.Lots{}, 0, errors.New("db error"))

		h := NewHandler(m)
		h.GetAll(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")

		m.AssertExpectations(t)
	})
}

func TestHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success as bidder", func(t *testing.T) {
		m := new(lots.Mock)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("role", "bidder")
		c.Request = httptest.NewRequest(http.MethodGet, "/lots/1", nil)

		lot := &model.Lots{ID: 1, Title: "Lot 1"}
		m.On("GetByIDForBid", mock.Anything, model.Lots{ID: 1}).Return(lot, nil)
		h := NewHandler(m)
		h.GetByID(c)

		require.Equal(t, http.StatusOK, w.Code)

		var resp model.Lots
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		require.Equal(t, int64(1), resp.ID)
		require.Equal(t, "Lot 1", resp.Title)

		m.AssertExpectations(t)
	})

	t.Run("success as non-bidder", func(t *testing.T) {
		m := new(lots.Mock)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "2"}}
		c.Set("role", "seller")
		c.Request = httptest.NewRequest(http.MethodGet, "/lots/1", nil)
		lot := &model.Lots{ID: 2, Title: "Lot 2"}
		m.On("GetByID", mock.Anything, model.Lots{ID: 2}).Return(lot, nil)

		h := NewHandler(m)
		h.GetByID(c)
		require.Equal(t, http.StatusOK, w.Code)

		var resp model.Lots
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		require.Equal(t, int64(2), resp.ID)
		require.Equal(t, "Lot 2", resp.Title)
		m.AssertExpectations(t)
	})
	t.Run("service error", func(t *testing.T) {
		m := new(lots.Mock)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "3"}}
		c.Set("role", "seller")
		c.Request = httptest.NewRequest(http.MethodGet, "/lots/3", nil)
		m.On("GetByID", mock.Anything, model.Lots{ID: 3}).
			Return((*model.Lots)(nil), errors.New("db error"))
		h := NewHandler(m)
		h.GetByID(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")

		m.AssertExpectations(t)
	})
}

func TestHandler_UpdateByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		m := new(lots.Mock)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"title":"Updated Lot","description":"New Desc","category":"Cars",
                  "startPrice":2000,"photo":"img.png","endsAt":"2026-08-01T12:00:00Z"}`
		c.Request = httptest.NewRequest(http.MethodPut, "/lots/1", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", int64(1))

		m.On("UpdateById", mock.Anything, mock.AnythingOfType("model.Lots")).
			Return(nil)

		h := NewHandler(m)
		h.UpdateByID(c)

		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "Lot updated successfully")

		m.AssertExpectations(t)
	})
	t.Run("repository error", func(t *testing.T) {
		m := new(lots.Mock)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"title":"Updated Lot","description":"New Desc","category":"Cars",
                  "startPrice":2000,"photo":"img.png","endsAt":"2026-08-01T12:00:00Z"}`
		c.Request = httptest.NewRequest(http.MethodPut, "/lots/1", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", int64(1))

		m.On("UpdateById", mock.Anything, mock.AnythingOfType("model.Lots")).
			Return(errors.New("db error"))

		h := NewHandler(m)
		h.UpdateByID(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")
	})
}

func TestHandler_UpdateStatusByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("success", func(t *testing.T) {
		m := new(lots.Mock)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"status":"live"}`
		c.Request = httptest.NewRequest(http.MethodPut, "/lots/1/status", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", int64(1))
		m.On("UpdateStatus", mock.Anything, int64(1), int64(1), "live").Return(nil)
		h := NewHandler(m)
		h.UpdateStatusByID(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "Status updated successfully")
	})
	t.Run("repository error", func(t *testing.T) {
		m := new(lots.Mock)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"status":"live"}`
		c.Request = httptest.NewRequest(http.MethodPut, "/lots/1/status", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", int64(1))
		m.On("UpdateStatus", mock.Anything, int64(1), int64(1), "live").Return(errors.New("db error"))
		h := NewHandler(m)
		h.UpdateStatusByID(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")
	})
}

func TestHandler_DeleteLots(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("success", func(t *testing.T) {
		m := new(lots.Mock)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", int64(1))
		c.Request = httptest.NewRequest(http.MethodDelete, "/lots/1", nil)
		m.On("DeleteLots", mock.Anything, model.Lots{ID: 1, SellerID: 1}).Return(nil)

		h := NewHandler(m)
		h.DeleteLots(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "Lot deleted successfully")
	})
	t.Run("repo error", func(t *testing.T) {
		m := new(lots.Mock)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", int64(1))
		c.Request = httptest.NewRequest(http.MethodDelete, "/lots/1", nil)
		m.On("DeleteLots", mock.Anything, model.Lots{ID: 1, SellerID: 1}).
			Return(errors.New("db error"))
		h := NewHandler(m)
		h.DeleteLots(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")
	})
}
func TestHandler_UploadPhoto_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("photo", "test.png")
	require.NoError(t, err)
	_, err = part.Write([]byte{
		0x89, 0x50, 0x4E, 0x47,
		0x0D, 0x0A, 0x1A, 0x0A,
	})
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/lots/1/photo", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Set("user_id", int64(1))
	m.On("UploadPhotoById", mock.Anything, int64(1), int64(1),
		mock.AnythingOfType("string"), mock.AnythingOfType("int64"), mock.AnythingOfType("string")).
		Return(nil)
	h := NewHandler(m)
	h.UploadPhoto(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "file")
}
func TestHandler_GetPhoto(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("success", func(t *testing.T) {
		m := new(lots.Mock)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", int64(1))
		c.Request = httptest.NewRequest(http.MethodGet, "/lots/1/photo", nil)
		m.On("GetPhoto", mock.Anything, int64(1), int64(1)).
			Return("/uploads/images/test.png", nil)
		h := NewHandler(m)
		h.GetPhoto(c)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "test.png")
	})
	t.Run("repo error", func(t *testing.T) {
		m := new(lots.Mock)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", int64(1))
		c.Request = httptest.NewRequest(http.MethodGet, "/lots/1/photo", nil)
		m.On("GetPhoto", mock.Anything, int64(1), int64(1)).
			Return("", errors.New("db error"))
		h := NewHandler(m)
		h.GetPhoto(c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "db error")
	})
}
func TestHandler_GetMine_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", int64(1))
	c.Request = httptest.NewRequest(http.MethodGet, "/lots/mine", nil)
	lot := []model.Lots{
		{ID: 1, Title: "Lot 1"},
		{ID: 2, Title: "Lot 2"},
	}
	m.On("GetMineLots", mock.Anything, int64(1), mock.AnythingOfType("model.LotsFilter")).
		Return(lot, 2, nil)
	h := NewHandler(m)
	h.GetMine(c)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_CancelLot_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	h := NewHandler(m)
	m.On("CancelLot", mock.Anything, int64(2), "reason").Return(nil)
	body := `{"reason":"reason"}`
	req := httptest.NewRequest(http.MethodPut, "/admin/lots/2/cancel", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	h.CancelLot(c)
	require.Equal(t, http.StatusOK, w.Code)
	m.AssertExpectations(t)
}
func TestHandler_ReportLot_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	h := NewHandler(m)
	m.On("ReportLot", mock.Anything, int64(2), int64(4), "reason").Return(nil)
	body := `{"reason":"reason"}`
	req := httptest.NewRequest(http.MethodPut, "/admin/lots/2/report", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Set("user_id", int64(4))
	h.ReportLot(c)
	require.Equal(t, http.StatusOK, w.Code)
	m.AssertExpectations(t)
}
func TestHandler_ReportLot_DbError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	h := NewHandler(m)
	m.On("ReportLot", mock.Anything, int64(2), int64(4), "reason").Return(errors.New("db error"))
	body := `{"reason":"reason"}`
	req := httptest.NewRequest(http.MethodPut, "/admin/lots/2/report", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Set("user_id", int64(4))
	h.ReportLot(c)
	require.Equal(t, http.StatusInternalServerError, w.Code)
	m.AssertExpectations(t)

}
func TestHandler_GetReports_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := new(lots.Mock)
	h := NewHandler(m)
	expected := []model.ReportLot{
		{ID: 1, LotID: 2, ReporterID: 42, Reason: "reason"},
	}
	m.On("GetListOfReports", mock.Anything).Return(expected, nil)
	req := httptest.NewRequest(http.MethodGet, "/admin/reports", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	h.GetReports(c)
	require.Equal(t, http.StatusOK, w.Code)
	var got []model.ReportLot
	err := json.Unmarshal(w.Body.Bytes(), &got)
	require.NoError(t, err)
	require.Equal(t, expected, got)
	m.AssertExpectations(t)
}
