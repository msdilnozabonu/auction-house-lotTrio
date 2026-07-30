package seller

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/seller"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockRepo_GetSellerStats_Success(t *testing.T) {
	m := new(seller.MockRepo)
	ctx := context.Background()
	sellerID := int64(42)

	expected := model.SellerStats{
		CountSelle:      10,
		SumCurrentPrice: 5000.0,
		AVGCurrentPrice: 500.0,
		TopLots: []model.TopLot{
			{ID: 1, Title: "Lot 1", CurrentPrice: 1000},
			{ID: 2, Title: "Lot 2", CurrentPrice: 900},
		},
	}

	m.On("GetSellerStats", ctx, sellerID).Return(expected, nil)

	result, err := m.GetSellerStats(ctx, sellerID)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	m.AssertExpectations(t)
}

func TestMockRepo_GetSellerStats_Error(t *testing.T) {
	m := new(seller.MockRepo)
	ctx := context.Background()
	sellerID := int64(99)

	m.On("GetSellerStats", ctx, sellerID).
		Return(model.SellerStats{}, errors.New("seller stats: db error"))

	result, err := m.GetSellerStats(ctx, sellerID)

	assert.Error(t, err)
	assert.Equal(t, model.SellerStats{}, result)
	m.AssertExpectations(t)
}

func TestMock_CancelLot(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		id       int64
		sellerID int64
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "success",
			id:       1,
			sellerID: 42,
			mockErr:  nil,
			wantErr:  false,
		},
		{
			name:     "repo error",
			id:       2,
			sellerID: 42,
			mockErr:  errors.New("cancel lot: db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(Mock)
			m.On("CancelLot", ctx, tt.id, tt.sellerID).Return(tt.mockErr)

			err := m.CancelLot(ctx, tt.id, tt.sellerID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.mockErr.Error())
			} else {
				assert.NoError(t, err)
			}

			m.AssertExpectations(t)
		})
	}
}