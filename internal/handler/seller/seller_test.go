package seller

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/service/seller"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/seller/stats", nil).
		WithContext(context.Background())
	return c, w
}

func TestHandler_GetSellerStats_Success(t *testing.T) {
	mockService := new(seller.Mock)
	h := &handler{sellerService: mockService}

	c, w := setupTestContext()
	c.Set("user_id", int64(42))

	expected := model.SellerStats{
		CountSelle:      10,
		SumCurrentPrice: 5000,
		AVGCurrentPrice: 500,
		TopLots: []model.TopLot{
			{ID: 1, Title: "Lot 1", CurrentPrice: 1000},
		},
	}

	mockService.On("GetSellerStats", c.Request.Context(), int64(42)).Return(expected, nil)

	h.GetSellerStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var got model.SellerStats
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)

	mockService.AssertExpectations(t)
}

func TestHandler_GetSellerStats_NoUserID(t *testing.T) {
	mockService := new(seller.Mock)
	h := &handler{sellerService: mockService}

	c, w := setupTestContext()
	// user_id намеренно не устанавливаем

	h.GetSellerStats(c)

	assert.NotEqual(t, http.StatusOK, w.Code)
	mockService.AssertNotCalled(t, "GetSellerStats")
}

func TestHandler_GetSellerStats_InvalidUserIDType(t *testing.T) {
	mockService := new(seller.Mock)
	h := &handler{sellerService: mockService}

	c, w := setupTestContext()
	c.Set("user_id", "not-an-int64") // неверный тип

	h.GetSellerStats(c)

	assert.NotEqual(t, http.StatusOK, w.Code)
	mockService.AssertNotCalled(t, "GetSellerStats")
}

func TestHandler_GetSellerStats_ServiceError(t *testing.T) {
	mockService := new(seller.Mock)
	h := &handler{sellerService: mockService}

	c, w := setupTestContext()
	c.Set("user_id", int64(42))

	mockService.On("GetSellerStats", c.Request.Context(), int64(42)).
		Return(model.SellerStats{}, errors.New("db error"))

	h.GetSellerStats(c)

	assert.NotEqual(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func setupTestContextWithParam(idParam string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/seller/lots/"+idParam+"/cancel", nil).
		WithContext(context.Background())
	c.Params = gin.Params{{Key: "id", Value: idParam}}
	return c, w
}

func TestHandler_CanceledLot(t *testing.T) {
	tests := []struct {
		name           string
		setUserID      bool
		userIDValue    interface{}
		idParam        string
		mockErr        error
		expectCallMock bool
		wantStatus     int
	}{
		{
			name:           "success",
			setUserID:      true,
			userIDValue:    int64(42),
			idParam:        "1",
			mockErr:        nil,
			expectCallMock: true,
			wantStatus:     http.StatusOK,
		},
		{
			name:           "no user_id in context",
			setUserID:      false,
			idParam:        "1",
			expectCallMock: false,
			wantStatus:     http.StatusInternalServerError, // поправьте под реальный код RespondError
		},
		{
			name:           "invalid user_id type",
			setUserID:      true,
			userIDValue:    "not-an-int64",
			idParam:        "1",
			expectCallMock: false,
			wantStatus:     http.StatusInternalServerError, // поправьте под реальный код RespondError
		},
		{
			name:           "invalid id param",
			setUserID:      true,
			userIDValue:    int64(42),
			idParam:        "not-a-number",
			expectCallMock: false,
			wantStatus:     http.StatusInternalServerError, // поправьте под реальный код RespondError
		},
		{
			name:           "service returns error",
			setUserID:      true,
			userIDValue:    int64(42),
			idParam:        "1",
			mockErr:        errors.New("cancel lot: db error"),
			expectCallMock: true,
			wantStatus:     http.StatusInternalServerError, // поправьте под реальный код RespondError
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(seller.Mock)
			h := &handler{sellerService: mockService}

			c, w := setupTestContextWithParam(tt.idParam)
			if tt.setUserID {
				c.Set("user_id", tt.userIDValue)
			}

			if tt.expectCallMock {
				mockService.On("CancelLot", c.Request.Context(), int64(1), int64(42)).
					Return(tt.mockErr)
			}

			h.CanceledLot(c)

			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, http.StatusOK, w.Code)

				var got map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &got)
				assert.NoError(t, err)
				assert.Equal(t, "canceled", got["status"])
			} else {
				assert.NotEqual(t, http.StatusOK, w.Code)
			}

			if tt.expectCallMock {
				mockService.AssertExpectations(t)
			} else {
				mockService.AssertNotCalled(t, "CancelLot")
			}
		})
	}
}