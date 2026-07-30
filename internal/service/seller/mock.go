package seller

import (
	"auction-house-lotTrio/internal/model"
	"context"

	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
}

func (m *Mock) GetSellerStats(ctx context.Context, sellerID int64) (model.SellerStats, error) {
	args := m.Called(ctx, sellerID)
	return args.Get(0).(model.SellerStats), args.Error(1) //nolint:wrapcheck
}

var _ Service = (*Mock)(nil)
