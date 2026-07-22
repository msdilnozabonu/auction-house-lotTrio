// nolint:wrapcheck
package lots

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
}

func (m *Mock) CreateLot(ctx context.Context, title,
	description string, startPrice float64, photo string,
	endsAt time.Time, status string, sellerID int64) error {
	args := m.Called(ctx, title, description, startPrice, photo, endsAt, status, sellerID)
	return args.Error(0)
}

func (m *Mock) GetAll(ctx context.Context) ([]model.Lots, error) {
	args := m.Called(ctx)
		return args.Get(0).([]model.Lots), args.Error(1)
}

func (m *Mock) CloseExpiredLot(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *Mock) StartScheduler(ctx context.Context, interval time.Duration) {
	m.Called(ctx, interval)
}

var _ Service = (*Mock)(nil)
