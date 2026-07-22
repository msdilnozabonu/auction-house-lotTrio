//nolint:wrapcheck
package lots

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) DeleteLots(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepo) GetById(ctx context.Context, id int64) (*model.Lots, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Lots), args.Error(1)
}

func (m *MockRepo) UpdateLot(ctx context.Context, l model.Lots) error {
	args := m.Called(ctx, l)
	return args.Error(0)
}

func (m *MockRepo) CreateLot(ctx context.Context, title,
	description string, startPrice float64, photo string,
	endsAt time.Time, status string, sellerID int64, currentPrice float64) error {
	args := m.Called(ctx, title, description, startPrice, photo, endsAt, status, sellerID, currentPrice)
		return args.Error(0)
}

func (m *MockRepo) GetAll(ctx context.Context) ([]model.Lots, error) {
	args := m.Called(ctx)
		return args.Get(0).([]model.Lots), args.Error(1)
}

func (m *MockRepo) FindExpiredLot(ctx context.Context, now time.Time) ([]int64, error) {
	args := m.Called(ctx, now)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockRepo) CloseLot(ctx context.Context, id int64) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

var _ Repo = (*MockRepo)(nil)
