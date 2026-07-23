//nolint:wrapcheck
package lots

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

const (
	resultIndex = 0
	errorIndex  = 1

	totalIndex     = 1
	getAllErrIndex = 2
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) DeleteLots(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(resultIndex)
}

func (m *MockRepo) GetById(ctx context.Context, id int64) (*model.Lots, error) {
	args := m.Called(ctx, id)
	return args.Get(resultIndex).(*model.Lots), args.Error(errorIndex)
}

func (m *MockRepo) UpdateLot(ctx context.Context, l model.Lots) error {
	args := m.Called(ctx, l)
	return args.Error(resultIndex)
}

func (m *MockRepo) CreateLot(ctx context.Context, title,
	description, category string, startPrice float64, photo string,
	endsAt time.Time, status string, sellerID int64, currentPrice float64) error {
	args := m.Called(ctx, title, description, category, startPrice, photo, endsAt, status, sellerID, currentPrice)
	return args.Error(resultIndex)
}

func (m *MockRepo) GetAll(ctx context.Context, filter model.LotsFilter) ([]model.Lots, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(resultIndex).([]model.Lots), args.Int(totalIndex), args.Error(getAllErrIndex)
}

func (m *MockRepo) FindExpiredLot(ctx context.Context, now time.Time) ([]int64, error) {
	args := m.Called(ctx, now)
	return args.Get(resultIndex).([]int64), args.Error(errorIndex)
}

func (m *MockRepo) CloseLot(ctx context.Context, id int64) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(resultIndex), args.Error(errorIndex)
}

var _ Repo = (*MockRepo)(nil)
